package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type MerchantHandler struct {
	store store.Store
}

func NewMerchantHandler(s store.Store) *MerchantHandler {
	return &MerchantHandler{store: s}
}

type RegisterRequest struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	WebhookURL string `json:"webhook_url"`
}

type RegisterResponse struct {
	Merchant *models.Merchant `json:"merchant"`
	APIKey   string           `json:"api_key"`
}

func (h *MerchantHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := crypto.GenerateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate api key"})
		return
	}

	merchant := &models.Merchant{
		Name:       req.Name,
		Email:      req.Email,
		APIKeyHash: crypto.HashAPIKey(apiKey),
		WebhookURL: req.WebhookURL,
	}

	if err := h.store.CreateMerchant(c.Request.Context(), merchant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create merchant"})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		Merchant: merchant,
		APIKey:   apiKey,
	})
}

func (h *MerchantHandler) GetProfile(c *gin.Context) {
	merchant, _ := c.Get("merchant")
	c.JSON(http.StatusOK, merchant)
}
