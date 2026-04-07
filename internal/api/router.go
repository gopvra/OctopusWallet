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

func NewRouter(s store.Store, registry *chain.Registry, seed []byte, wh *webhook.Service, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "chains": registry.Names()})
	})

	// Security middleware
	rl := middleware.NewRateLimiter(rate.Limit(100), 200)
	idempotency := middleware.NewIdempotencyStore()

	merchantHandler := handlers.NewMerchantHandler(s)
	paymentHandler := handlers.NewPaymentHandler(s, registry, seed)
	payoutHandler := handlers.NewPayoutHandler(s, registry, wh)
	walletHandler := handlers.NewWalletHandler(s)
	approvalHandler := handlers.NewApprovalHandler(s, wh)
	sweepHandler := handlers.NewSweepHandler(s)
	coldHotHandler := handlers.NewColdHotHandler(s)
	gasHandler := handlers.NewGasStationHandler(registry, cfg.GasStation)

	v1 := r.Group("/api/v1")
	v1.Use(rl.Middleware())

	// Public endpoints
	v1.POST("/merchants/register", merchantHandler.Register)

	// Authenticated endpoints
	auth := v1.Group("")
	auth.Use(middleware.APIKeyAuth(s))
	auth.Use(middleware.RequestHMAC()) // optional request signature validation
	{
		auth.GET("/merchants/profile", merchantHandler.GetProfile)

		// Payments (with idempotency)
		auth.POST("/payments/create", idempotency.Middleware(), paymentHandler.CreatePayment)
		auth.GET("/payments/:id", paymentHandler.GetPayment)

		// Payouts (with idempotency + approval workflow)
		auth.POST("/payouts/create", idempotency.Middleware(), payoutHandler.CreatePayout)
		auth.GET("/payouts/:id", payoutHandler.GetPayout)

		// Approval
		auth.POST("/approval/config", approvalHandler.SetConfig)
		auth.GET("/approval/config", approvalHandler.GetConfig)
		auth.POST("/payouts/:id/approve", approvalHandler.ApprovePayout)
		auth.POST("/payouts/:id/reject", approvalHandler.RejectPayout)

		// Wallets
		auth.GET("/wallets", walletHandler.List)

		// Sweep / Auto-collection
		auth.POST("/sweep/collection-address", sweepHandler.SetCollectionAddress)
		auth.GET("/sweep/collection-address", sweepHandler.GetCollectionAddresses)
		auth.GET("/sweep/tasks", sweepHandler.ListTasks)

		// Cold/Hot wallet
		auth.POST("/cold-wallet/config", coldHotHandler.SetConfig)
		auth.GET("/cold-wallet/config", coldHotHandler.GetConfig)
		auth.GET("/cold-wallet/transfers", coldHotHandler.ListTransfers)

		// Gas station
		auth.GET("/gas/status", gasHandler.GetStatus)
	}

	return r
}
