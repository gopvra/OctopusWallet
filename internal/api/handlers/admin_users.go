package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list admin users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

type CreateAdminUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=admin super_admin"`
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	var req CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &models.AdminUser{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hash),
		Role:     req.Role,
	}

	if err := h.store.CreateAdminUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

type UpdateAdminUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,oneof=admin super_admin"`
	IsActive bool   `json:"is_active"`
	Password string `json:"password,omitempty"`
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.store.GetAdminUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin user not found"})
		return
	}

	user.Username = req.Username
	user.Email = req.Email
	user.Role = req.Role
	user.IsActive = req.IsActive

	if req.Password != "" {
		if len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = string(hash)
	}

	if err := h.store.UpdateAdminUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update admin user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	currentUserID := c.GetString("admin_user_id")

	// Prevent deleting yourself
	if id == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	if err := h.store.DeleteAdminUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete admin user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "admin user deleted"})
}
