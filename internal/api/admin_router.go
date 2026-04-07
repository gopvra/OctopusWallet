package api

import (
	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/api/handlers"
	"github.com/octopuswallet/octopuswallet/internal/api/middleware"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"golang.org/x/time/rate"
)

func SetupAdminRoutes(r *gin.Engine, adminStore store.AdminStore, jwtSecret string, allowedOrigins []string) {
	authHandler := handlers.NewAdminAuthHandler(adminStore, jwtSecret)
	dashboardHandler := handlers.NewAdminDashboardHandler(adminStore)
	merchantHandler := handlers.NewAdminMerchantHandler(adminStore)
	paymentHandler := handlers.NewAdminPaymentHandler(adminStore)
	payoutHandler := handlers.NewAdminPayoutHandler(adminStore)
	walletHandler := handlers.NewAdminWalletHandler(adminStore)
	chainStateHandler := handlers.NewAdminChainStateHandler(adminStore)
	refundHandler := handlers.NewAdminRefundHandler(adminStore)
	batchPayoutHandler := handlers.NewAdminBatchPayoutHandler(adminStore)
	balanceHandler := handlers.NewAdminBalanceHandler(adminStore)
	currencyHandler := handlers.NewAdminCurrencyHandler(adminStore)
	adminUserHandler := handlers.NewAdminUserHandler(adminStore)

	admin := r.Group("/api/admin/v1")
	admin.Use(middleware.CORS(allowedOrigins))
	admin.Use(middleware.SecurityHeaders())

	// Public admin endpoints (with strict rate limiting on login)
	loginRL := middleware.NewRateLimiter(rate.Limit(5.0/60.0), 5) // 5 per minute
	admin.POST("/auth/login", loginRL.Middleware(), authHandler.Login)
	admin.POST("/auth/refresh", authHandler.Refresh)

	// Authenticated admin endpoints
	protected := admin.Group("")
	protected.Use(middleware.JWTAuth(jwtSecret))
	{
		// Auth
		protected.GET("/auth/me", authHandler.Me)

		// Dashboard
		protected.GET("/dashboard/stats", dashboardHandler.Stats)
		protected.GET("/dashboard/volume-chart", dashboardHandler.VolumeChart)
		protected.GET("/dashboard/chain-distribution", dashboardHandler.ChainDistribution)
		protected.GET("/dashboard/recent-activity", dashboardHandler.RecentActivity)

		// Merchants
		protected.GET("/merchants", merchantHandler.List)
		protected.GET("/merchants/:id", merchantHandler.GetByID)
		protected.PUT("/merchants/:id", merchantHandler.Update)
		protected.PATCH("/merchants/:id/toggle-active", merchantHandler.ToggleActive)

		// Payments
		protected.GET("/payments", paymentHandler.List)
		protected.GET("/payments/:id", paymentHandler.GetByID)

		// Payouts
		protected.GET("/payouts", payoutHandler.List)
		protected.GET("/payouts/:id", payoutHandler.GetByID)

		// Wallets
		protected.GET("/wallets", walletHandler.List)

		// Refunds
		protected.GET("/refunds", refundHandler.List)
		protected.GET("/refunds/:id", refundHandler.GetByID)

		// Batch Payouts
		protected.GET("/batch-payouts", batchPayoutHandler.List)
		protected.GET("/batch-payouts/:id", batchPayoutHandler.GetByID)

		// Balances
		protected.GET("/balances", balanceHandler.List)

		// Currencies
		protected.GET("/currencies", currencyHandler.List)

		// Chain State
		protected.GET("/chain-state", chainStateHandler.List)

		// Admin Users (super_admin only)
		adminUsers := protected.Group("")
		adminUsers.Use(middleware.RequireSuperAdmin())
		{
			adminUsers.GET("/admin-users", adminUserHandler.List)
			adminUsers.POST("/admin-users", adminUserHandler.Create)
			adminUsers.PUT("/admin-users/:id", adminUserHandler.Update)
			adminUsers.DELETE("/admin-users/:id", adminUserHandler.Delete)
		}
	}
}
