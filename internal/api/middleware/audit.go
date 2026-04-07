package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

// AuditLog records all mutating API calls (POST/PUT/DELETE) to the audit_logs table.
func AuditLog(s store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // process request first

		// Only log mutating operations
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			return
		}

		merchantID, _ := c.Get("merchant_id")
		mid, _ := merchantID.(string)
		if mid == "" {
			return
		}

		log := &models.AuditLog{
			MerchantID:   mid,
			Action:       method,
			ResourceType: c.Request.URL.Path,
			ResourceID:   c.Param("id"),
			IPAddress:    c.ClientIP(),
		}
		s.CreateAuditLog(c.Request.Context(), log)
	}
}
