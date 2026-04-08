package api

import (
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/octopuswallet/octopuswallet/internal/api/handlers"
	"github.com/octopuswallet/octopuswallet/internal/api/middleware"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/cache"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
	"golang.org/x/time/rate"
)

func NewRouter(s store.Store, registry *chain.Registry, seed []byte, wh *webhook.Service, cfg *config.Config, hub *Hub, adminStore store.AdminStore, redis *cache.Client) *gin.Engine {
	r := gin.Default()

	// Global middleware: language detection for i18n responses
	r.Use(middleware.LangMiddleware())

	r.GET("/health", func(c *gin.Context) {
		R.OK(c, gin.H{"status": "ok", "chains": registry.Names()})
	})

	// WebSocket for real-time payment status
	r.GET("/ws/payments/:id", HandleWebSocket(hub, cfg.Admin.AllowedOrigins))

	// Serve frontend static files (if built)
	r.Static("/static", "./web/dist/assets")
	r.StaticFile("/", "./web/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] != "/api" && c.Request.URL.Path[:3] != "/ws" {
			c.File("./web/dist/index.html")
			return
		}
		c.JSON(404, gin.H{"error": "not found"})
	})

	// Rate limiter: use Redis if available, else in-memory
	var rlMiddleware gin.HandlerFunc
	if redis != nil {
		rl := middleware.NewRedisRateLimiter(redis, 100, time.Minute)
		rlMiddleware = rl.Middleware()
	} else {
		rl := middleware.NewRateLimiter(rate.Limit(100), 200)
		rlMiddleware = rl.Middleware()
	}

	// Idempotency: use Redis if available, else in-memory
	var idemMiddleware gin.HandlerFunc
	if redis != nil {
		idem := middleware.NewRedisIdempotencyStore(redis)
		idemMiddleware = idem.Middleware()
	} else {
		idem := middleware.NewIdempotencyStore()
		idemMiddleware = idem.Middleware()
	}

	ipWhitelist := middleware.NewIPWhitelist()

	merchantHandler := handlers.NewMerchantHandler(s)
	paymentHandler := handlers.NewPaymentHandler(s, registry, seed)
	payoutHandler := handlers.NewPayoutHandler(s, registry, wh)
	walletHandler := handlers.NewWalletHandler(s)
	approvalHandler := handlers.NewApprovalHandler(s, wh)
	sweepHandler := handlers.NewSweepHandler(s)
	coldHotHandler := handlers.NewColdHotHandler(s)
	gasHandler := handlers.NewGasStationHandler(registry, cfg.GasStation)
	refundHandler := handlers.NewRefundHandler(s)
	currencyHandler := handlers.NewCurrencyHandler(s, registry)
	batchHandler := handlers.NewBatchPayoutHandler(s, registry)
	balanceHandler := handlers.NewBalanceHandler(s, ipWhitelist)
	paymentLinkHandler := handlers.NewPaymentLinkHandler(s, registry)
	exportHandler := handlers.NewExportHandler(s)
	auditLogHandler := handlers.NewAuditLogHandler(s)

	v1 := r.Group("/api/v1")
	v1.Use(rlMiddleware)

	// Public endpoints
	v1.POST("/merchants/register", merchantHandler.Register)
	v1.GET("/currencies", currencyHandler.ListCurrencies)
	v1.GET("/rates", currencyHandler.GetExchangeRate)
	v1.GET("/payment-links/:id", paymentLinkHandler.GetPublic)

	// Authenticated endpoints
	auth := v1.Group("")
	auth.Use(middleware.APIKeyAuth(s))
	auth.Use(ipWhitelist.Middleware())
	auth.Use(middleware.RequestHMAC())
	auth.Use(middleware.AuditLog(s))
	{
		auth.GET("/merchants/profile", merchantHandler.GetProfile)

		auth.POST("/payments/create", idemMiddleware, paymentHandler.CreatePayment)
		auth.GET("/payments/:id", paymentHandler.GetPayment)
		auth.GET("/payments", balanceHandler.ListPayments)

		auth.POST("/refunds/create", idemMiddleware, refundHandler.CreateRefund)
		auth.GET("/refunds/:id", refundHandler.GetRefund)
		auth.GET("/payments/:payment_id/refunds", refundHandler.ListRefundsByPayment)

		auth.POST("/payouts/create", idemMiddleware, payoutHandler.CreatePayout)
		auth.GET("/payouts/:id", payoutHandler.GetPayout)
		auth.GET("/payouts", balanceHandler.ListPayouts)

		auth.POST("/payouts/batch", idemMiddleware, batchHandler.CreateBatchPayout)
		auth.GET("/payouts/batch/:id", batchHandler.GetBatchPayout)
		auth.GET("/payouts/batches", batchHandler.ListBatchPayouts)

		auth.POST("/approval/config", approvalHandler.SetConfig)
		auth.GET("/approval/config", approvalHandler.GetConfig)
		auth.POST("/payouts/:id/approve", approvalHandler.ApprovePayout)
		auth.POST("/payouts/:id/reject", approvalHandler.RejectPayout)

		auth.GET("/balances", balanceHandler.GetBalances)
		auth.GET("/wallets", walletHandler.List)

		auth.POST("/sweep/collection-address", sweepHandler.SetCollectionAddress)
		auth.GET("/sweep/collection-address", sweepHandler.GetCollectionAddresses)
		auth.GET("/sweep/tasks", sweepHandler.ListTasks)

		auth.POST("/cold-wallet/config", coldHotHandler.SetConfig)
		auth.GET("/cold-wallet/config", coldHotHandler.GetConfig)
		auth.GET("/cold-wallet/transfers", coldHotHandler.ListTransfers)

		auth.GET("/gas/status", gasHandler.GetStatus)

		// Payment Links
		auth.POST("/payment-links", paymentLinkHandler.Create)
		auth.GET("/payment-links", paymentLinkHandler.List)

		// Export (CSV/JSON)
		auth.GET("/export/payments", exportHandler.ExportPayments)
		auth.GET("/export/payouts", exportHandler.ExportPayouts)

		// Audit Logs
		auth.GET("/audit-logs", auditLogHandler.List)

		// IP Whitelist
		auth.POST("/security/ip-whitelist", balanceHandler.SetIPWhitelist)
		auth.GET("/security/ip-whitelist", balanceHandler.GetIPWhitelist)
	}

	if adminStore != nil {
		// Initialize WebAuthn
		wa, err := webauthn.New(&webauthn.Config{
			RPDisplayName: "OctopusWallet Admin",
			RPID:          cfg.Admin.WebAuthnRPID,
			RPOrigins:     []string{cfg.Admin.WebAuthnOrigin},
		})
		if err != nil {
			slog.Warn("webauthn initialization failed, passkey login disabled", "error", err)
		}
		SetupAdminRoutes(r, adminStore, cfg.Admin.JWTSecret, cfg.Admin.AllowedOrigins, wa)
	}

	return r
}
