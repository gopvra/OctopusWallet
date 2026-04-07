package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type AdminMerchantHandler struct {
	store store.AdminStore
}

func NewAdminMerchantHandler(s store.AdminStore) *AdminMerchantHandler {
	return &AdminMerchantHandler{store: s}
}

func (h *AdminMerchantHandler) List(c *gin.Context) {
	var filter store.MerchantFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.store.ListMerchants(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list merchants"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AdminMerchantHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	merchant, err := h.store.AdminGetMerchantByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "merchant not found"})
		return
	}
	c.JSON(http.StatusOK, merchant)
}

func (h *AdminMerchantHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name       string `json:"name" binding:"required"`
		Email      string `json:"email" binding:"required,email"`
		WebhookURL string `json:"webhook_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := crypto.ValidateWebhookURL(req.WebhookURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.UpdateMerchant(c.Request.Context(), id, req.Name, req.Email, req.WebhookURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update merchant"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "merchant updated"})
}

func (h *AdminMerchantHandler) ToggleActive(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.store.ToggleMerchantActive(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle merchant"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "merchant status toggled"})
}
