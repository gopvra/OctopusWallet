package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/auth"
	"github.com/octopuswallet/octopuswallet/internal/models"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			R.Abort(c, errcode.ErrUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			R.Abort(c, errcode.ErrUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(secret, parts[1])
		if err != nil {
			R.Abort(c, errcode.ErrAdminTokenInvalid)
			return
		}

		// Reject refresh tokens used as access tokens
		if claims.Issuer != "octopus-admin" {
			R.Abort(c, errcode.ErrAdminTokenInvalid)
			return
		}

		c.Set("admin_user_id", claims.UserID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("admin_role")
		if role != "super_admin" {
			R.Abort(c, errcode.ErrAdminInsufficientRole)
			return
		}
		c.Next()
	}
}

func RequirePermission(perm models.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("admin_role")
		if !models.HasPermission(role, perm) {
			R.Abort(c, errcode.ErrAdminInsufficientRole)
			return
		}
		c.Next()
	}
}
