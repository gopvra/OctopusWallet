package api

import (
	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/api/handlers"
	"github.com/octopuswallet/octopuswallet/internal/api/middleware"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
	"golang.org/x/time/rate"
)

func NewRouter(s store.Store, registry *chain.Registry, seed []byte, wh *webhook.Service, cfg *config.Config, hub *Hub, adminStore store.AdminStore) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "chains": registry.Names()})
	})

	// WebSocket for real-time payment status
	r.GET("/ws/payments/:id", HandleWebSocket(hub, cfg.Admin.AllowedOrigins))

	// Serve frontend static files (if built)
	r.Static("/static", "./web/dist/assets")
	r.StaticFile("/", "./web/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		// SPA fallback — serve index.html for non-API routes
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] != "/api" && c.Request.URL.Path[:3] != "/ws" {
			c.File("./web/dist/index.html")
			return
		}
		c.JSON(404, gin.H{"error": "not found"})
	})

	rl := middleware.NewRateLimiter(rate.Limit(100), 200)
	idempotency := middleware.NewIdempotencyStore()
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
	v1.Use(rl.Middleware())

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
		// Merchant
		auth.GET("/merchants/profile", merchantHandler.GetProfile)

		// Payments / Invoices (with idempotency)
		auth.POST("/payments/create", idempotency.Middleware(), paymentHandler.CreatePayment)
		auth.GET("/payments/:id", paymentHandler.GetPayment)
		auth.GET("/payments", balanceHandler.ListPayments) // paginated list

		// Refunds
		auth.POST("/refunds/create", idempotency.Middleware(), refundHandler.CreateRefund)
		auth.GET("/refunds/:id", refundHandler.GetRefund)
		auth.GET("/payments/:payment_id/refunds", refundHandler.ListRefundsByPayment)

		// Payouts (with idempotency + approval)
		auth.POST("/payouts/create", idempotency.Middleware(), payoutHandler.CreatePayout)
		auth.GET("/payouts/:id", payoutHandler.GetPayout)
		auth.GET("/payouts", balanceHandler.ListPayouts) // paginated list

		// Batch Payouts
		auth.POST("/payouts/batch", idempotency.Middleware(), batchHandler.CreateBatchPayout)
		auth.GET("/payouts/batch/:id", batchHandler.GetBatchPayout)
		auth.GET("/payouts/batches", batchHandler.ListBatchPayouts)

		// Approval
		auth.POST("/approval/config", approvalHandler.SetConfig)
		auth.GET("/approval/config", approvalHandler.GetConfig)
		auth.POST("/payouts/:id/approve", approvalHandler.ApprovePayout)
		auth.POST("/payouts/:id/reject", approvalHandler.RejectPayout)

		// Balance / Ledger
		auth.GET("/balances", balanceHandler.GetBalances)

		// Wallets
		auth.GET("/wallets", walletHandler.List)

		// Sweep
		auth.POST("/sweep/collection-address", sweepHandler.SetCollectionAddress)
		auth.GET("/sweep/collection-address", sweepHandler.GetCollectionAddresses)
		auth.GET("/sweep/tasks", sweepHandler.ListTasks)

		// Cold/Hot wallet
		auth.POST("/cold-wallet/config", coldHotHandler.SetConfig)
		auth.GET("/cold-wallet/config", coldHotHandler.GetConfig)
		auth.GET("/cold-wallet/transfers", coldHotHandler.ListTransfers)

		// Gas station
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

	// Admin routes
	if adminStore != nil {
		SetupAdminRoutes(r, adminStore, cfg.Admin.JWTSecret, cfg.Admin.AllowedOrigins)
	}

	return r
}
