package api

import (
	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/api/handlers"
	"github.com/octopuswallet/octopuswallet/internal/api/middleware"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"golang.org/x/time/rate"
)

func NewRouter(s store.Store, registry *chain.Registry, seed []byte, adminStore store.AdminStore, adminJWTSecret string, adminAllowedOrigins []string) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"chains": registry.Names(),
		})
	})

	// Rate limiter: 100 requests per second, burst 200
	rl := middleware.NewRateLimiter(rate.Limit(100), 200)

	merchantHandler := handlers.NewMerchantHandler(s)
	paymentHandler := handlers.NewPaymentHandler(s, registry, seed)
	payoutHandler := handlers.NewPayoutHandler(s, registry)
	walletHandler := handlers.NewWalletHandler(s)

	v1 := r.Group("/api/v1")
	v1.Use(rl.Middleware())

	// Public endpoints
	v1.POST("/merchants/register", merchantHandler.Register)

	// Authenticated endpoints
	auth := v1.Group("")
	auth.Use(middleware.APIKeyAuth(s))
	{
		auth.GET("/merchants/profile", merchantHandler.GetProfile)
		auth.POST("/payments/create", paymentHandler.CreatePayment)
		auth.GET("/payments/:id", paymentHandler.GetPayment)
		auth.POST("/payouts/create", payoutHandler.CreatePayout)
		auth.GET("/payouts/:id", payoutHandler.GetPayout)
		auth.GET("/wallets", walletHandler.List)
	}

	// Admin routes
	if adminStore != nil {
		SetupAdminRoutes(r, adminStore, adminJWTSecret, adminAllowedOrigins)
	}

	return r
}
