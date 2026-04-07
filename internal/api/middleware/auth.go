package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

func APIKeyAuth(s store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if key == "" {
			R.Abort(c, errcode.ErrUnauthorized)
			return
		}

		hash := crypto.HashAPIKey(key)
		merchant, err := s.GetMerchantByAPIKeyHash(c.Request.Context(), hash)
		if err != nil {
			R.Abort(c, errcode.ErrUnauthorized)
			return
		}

		c.Set("merchant", merchant)
		c.Set("merchant_id", merchant.ID)
		c.Next()
	}
}
