package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/octopuswallet/octopuswallet/internal/api/handlers"
	"github.com/octopuswallet/octopuswallet/internal/api/middleware"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"golang.org/x/time/rate"
)

func SetupAdminRoutes(r *gin.Engine, adminStore store.AdminStore, jwtSecret string, allowedOrigins []string, wa *webauthn.WebAuthn) {
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
	webauthnHandler := handlers.NewAdminWebAuthnHandler(adminStore, wa, jwtSecret)

	admin := r.Group("/api/admin/v1")
	admin.Use(middleware.CORS(allowedOrigins))
	admin.Use(middleware.SecurityHeaders())

	// Public admin endpoints (with strict rate limiting)
	loginRL := middleware.NewRateLimiter(rate.Limit(5.0/60.0), 5)
	refreshRL := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10)
	admin.POST("/auth/login", loginRL.Middleware(), authHandler.Login)
	admin.POST("/auth/refresh", refreshRL.Middleware(), authHandler.Refresh)

	// WebAuthn public endpoints (passkey login — no JWT needed)
	admin.POST("/auth/webauthn/login/begin", loginRL.Middleware(), webauthnHandler.BeginLogin)
	admin.POST("/auth/webauthn/login/finish", loginRL.Middleware(), webauthnHandler.FinishLogin)

	// Authenticated admin endpoints
	protected := admin.Group("")
	protected.Use(middleware.JWTAuth(jwtSecret))
	{
		protected.GET("/auth/me", authHandler.Me)

		// WebAuthn registration (requires authentication)
		protected.POST("/auth/webauthn/register/begin", webauthnHandler.BeginRegistration)
		protected.POST("/auth/webauthn/register/finish", webauthnHandler.FinishRegistration)
		protected.GET("/auth/webauthn/credentials", webauthnHandler.ListCredentials)
		protected.DELETE("/auth/webauthn/credentials/:id", webauthnHandler.DeleteCredential)

		// Dashboard
		protected.GET("/dashboard/stats", middleware.RequirePermission(models.PermDashboardView), dashboardHandler.Stats)
		protected.GET("/dashboard/volume-chart", middleware.RequirePermission(models.PermDashboardView), dashboardHandler.VolumeChart)
		protected.GET("/dashboard/chain-distribution", middleware.RequirePermission(models.PermDashboardView), dashboardHandler.ChainDistribution)
		protected.GET("/dashboard/recent-activity", middleware.RequirePermission(models.PermDashboardView), dashboardHandler.RecentActivity)

		// Merchants
		protected.GET("/merchants", middleware.RequirePermission(models.PermMerchantList), merchantHandler.List)
		protected.GET("/merchants/:id", middleware.RequirePermission(models.PermMerchantView), merchantHandler.GetByID)
		protected.PUT("/merchants/:id", middleware.RequirePermission(models.PermMerchantUpdate), merchantHandler.Update)
		protected.PATCH("/merchants/:id/toggle-active", middleware.RequirePermission(models.PermMerchantToggle), merchantHandler.ToggleActive)

		// Payments
		protected.GET("/payments", middleware.RequirePermission(models.PermPaymentList), paymentHandler.List)
		protected.GET("/payments/:id", middleware.RequirePermission(models.PermPaymentView), paymentHandler.GetByID)

		// Payouts
		protected.GET("/payouts", middleware.RequirePermission(models.PermPayoutList), payoutHandler.List)
		protected.GET("/payouts/:id", middleware.RequirePermission(models.PermPayoutView), payoutHandler.GetByID)

		// Wallets
		protected.GET("/wallets", middleware.RequirePermission(models.PermWalletList), walletHandler.List)

		// Refunds
		protected.GET("/refunds", middleware.RequirePermission(models.PermRefundList), refundHandler.List)
		protected.GET("/refunds/:id", middleware.RequirePermission(models.PermRefundView), refundHandler.GetByID)

		// Batch Payouts
		protected.GET("/batch-payouts", middleware.RequirePermission(models.PermBatchPayoutList), batchPayoutHandler.List)
		protected.GET("/batch-payouts/:id", middleware.RequirePermission(models.PermBatchPayoutView), batchPayoutHandler.GetByID)

		// Balances
		protected.GET("/balances", middleware.RequirePermission(models.PermBalanceList), balanceHandler.List)

		// Currencies
		protected.GET("/currencies", middleware.RequirePermission(models.PermCurrencyList), currencyHandler.List)

		// Chain State
		protected.GET("/chain-state", middleware.RequirePermission(models.PermChainStateList), chainStateHandler.List)

		// Admin Users
		protected.GET("/admin-users", middleware.RequirePermission(models.PermAdminUserList), adminUserHandler.List)
		protected.POST("/admin-users", middleware.RequirePermission(models.PermAdminUserCreate), adminUserHandler.Create)
		protected.PUT("/admin-users/:id", middleware.RequirePermission(models.PermAdminUserUpdate), adminUserHandler.Update)
		protected.DELETE("/admin-users/:id", middleware.RequirePermission(models.PermAdminUserDelete), adminUserHandler.Delete)
	}
}
