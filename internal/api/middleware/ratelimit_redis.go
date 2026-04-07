package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/cache"
)

// RedisRateLimiter provides rate limiting backed by Redis (sliding window).
type RedisRateLimiter struct {
	redis  *cache.Client
	limit  int
	window time.Duration
}

func NewRedisRateLimiter(r *cache.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{redis: r, limit: limit, window: window}
}

func (rl *RedisRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if id, exists := c.Get("merchant_id"); exists {
			key = id.(string)
		}

		allowed, _, err := rl.redis.RateLimit(c.Request.Context(), key, rl.limit, rl.window)
		if err != nil {
			// Fail open on Redis error
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
