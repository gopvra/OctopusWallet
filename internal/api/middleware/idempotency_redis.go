package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/cache"
)

// RedisIdempotencyStore provides idempotency checking backed by Redis.
type RedisIdempotencyStore struct {
	redis *cache.Client
}

func NewRedisIdempotencyStore(r *cache.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{redis: r}
}

func (s *RedisIdempotencyStore) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		merchantID, _ := c.Get("merchant_id")
		scopedKey := hashKey(key, merchantID)

		// Check cache
		entry, err := s.redis.GetIdempotent(c.Request.Context(), scopedKey)
		if err == nil && entry != nil {
			c.Data(entry.Status, "application/json", entry.Body)
			c.Abort()
			return
		}

		writer := &responseCapture{ResponseWriter: c.Writer}
		c.Writer = writer

		c.Next()

		// Cache response
		newEntry := &cache.IdempotencyEntry{
			Status: writer.status,
			Body:   writer.body,
		}
		s.redis.SetIdempotent(c.Request.Context(), scopedKey, newEntry, 24*time.Hour)
	}
}
