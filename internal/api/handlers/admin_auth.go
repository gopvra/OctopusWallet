package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/auth"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash is used for constant-time login to prevent username enumeration
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

type AdminAuthHandler struct {
	store     store.AdminStore
	jwtSecret string
}

func NewAdminAuthHandler(s store.AdminStore, jwtSecret string) *AdminAuthHandler {
	return &AdminAuthHandler{store: s, jwtSecret: jwtSecret}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	user, err := h.store.GetAdminUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		// Constant-time: run bcrypt even for non-existent users to prevent timing attacks
		bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		R.Fail(c, errcode.ErrUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		R.Fail(c, errcode.ErrUnauthorized)
		return
	}

	tokens, err := auth.GenerateTokenPair(h.jwtSecret, user.ID, user.Username, user.Role)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, gin.H{
		"user":  user,
		"token": tokens,
	})
}

func (h *AdminAuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	claims, err := auth.ValidateToken(h.jwtSecret, req.RefreshToken)
	if err != nil {
		R.Fail(c, errcode.ErrUnauthorized)
		return
	}

	// Ensure this is a refresh token, not an access token
	if claims.Issuer != "octopus-admin-refresh" {
		R.Fail(c, errcode.ErrUnauthorized)
		return
	}

	// Verify user still exists and is active
	user, err := h.store.GetAdminUserByID(c.Request.Context(), claims.UserID)
	if err != nil || !user.IsActive {
		R.Fail(c, errcode.ErrUnauthorized)
		return
	}

	tokens, err := auth.GenerateTokenPair(h.jwtSecret, user.ID, user.Username, user.Role)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, tokens)
}

func (h *AdminAuthHandler) Me(c *gin.Context) {
	userID := c.GetString("admin_user_id")
	user, err := h.store.GetAdminUserByID(c.Request.Context(), userID)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	R.OK(c, user)
}
