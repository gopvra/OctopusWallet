package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := crypto.ValidateWebhookURL(req.WebhookURL); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	apiKey, err := crypto.GenerateAPIKey()
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	merchant := &models.Merchant{
		Name:       req.Name,
		Email:      req.Email,
		APIKeyHash: crypto.HashAPIKey(apiKey),
		WebhookURL: req.WebhookURL,
	}

	if err := h.store.CreateMerchant(c.Request.Context(), merchant); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, RegisterResponse{
		Merchant: merchant,
		APIKey:   apiKey,
	})
}

func (h *MerchantHandler) GetProfile(c *gin.Context) {
	merchant, _ := c.Get("merchant")
	R.OK(c, merchant)
}
