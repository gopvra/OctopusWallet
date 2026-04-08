package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/google/uuid"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

type AdminUserHandler struct {
	store store.AdminStore
}

func NewAdminUserHandler(s store.AdminStore) *AdminUserHandler {
	return &AdminUserHandler{store: s}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	users, err := h.store.ListAdminUsers(c.Request.Context())
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, users)
}

type CreateAdminUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12,max=128"`
	Role     string `json:"role" binding:"required,oneof=admin super_admin viewer"`
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	var req CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	user := &models.AdminUser{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hash),
		Role:     req.Role,
	}

	if err := h.store.CreateAdminUser(c.Request.Context(), user); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, user)
}

type UpdateAdminUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,oneof=admin super_admin viewer"`
	IsActive bool   `json:"is_active"`
	Password string `json:"password,omitempty"`
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	var req UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	user, err := h.store.GetAdminUserByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}

	user.Username = req.Username
	user.Email = req.Email
	user.Role = req.Role
	user.IsActive = req.IsActive

	if req.Password != "" {
		if len(req.Password) < 12 || len(req.Password) > 128 {
			R.Fail(c, errcode.ErrBadRequest)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			R.Fail(c, errcode.ErrInternalServer)
			return
		}
		user.Password = string(hash)
	}

	if err := h.store.UpdateAdminUser(c.Request.Context(), user); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, user)
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	currentUserID := c.GetString("admin_user_id")

	// Prevent deleting yourself
	if id == currentUserID {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	// Prevent deleting last super_admin
	target, err := h.store.GetAdminUserByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	if target.Role == models.RoleSuperAdmin {
		users, _ := h.store.ListAdminUsers(c.Request.Context())
		superCount := 0
		for _, u := range users {
			if u.Role == models.RoleSuperAdmin && u.IsActive {
				superCount++
			}
		}
		if superCount <= 1 {
			R.Fail(c, errcode.ErrBadRequest)
			return
		}
	}

	if err := h.store.DeleteAdminUser(c.Request.Context(), id); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"message": "admin user deleted"})
}
