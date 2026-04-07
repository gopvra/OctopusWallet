package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type PayoutHandler struct {
	store    store.Store
	registry *chain.Registry
	webhook  *webhook.Service
}

func NewPayoutHandler(s store.Store, registry *chain.Registry, wh *webhook.Service) *PayoutHandler {
	return &PayoutHandler{store: s, registry: registry, webhook: wh}
}

type CreatePayoutRequest struct {
	Chain     string `json:"chain" binding:"required"`
	ToAddress string `json:"to_address" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Token     string `json:"token"`
}

func (h *PayoutHandler) CreatePayout(c *gin.Context) {
	var req CreatePayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID := c.GetString("merchant_id")

	// Validate chain exists
	if _, err := h.registry.Get(req.Chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported chain"})
		return
	}

	// Validate amount is positive integer
	if err := crypto.ValidateAmount(req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate address format
	if err := crypto.ValidateAddress(req.Chain, req.ToAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check approval config, limits, and determine approval status
	approvalStatus, rejectReason, err := CheckPayoutLimits(c, h.store, merchantID, req.Chain, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check payout limits"})
		return
	}
	if rejectReason != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": rejectReason})
		return
	}

	payout := &models.Payout{
		MerchantID:     merchantID,
		Chain:          req.Chain,
		Token:          req.Token,
		ToAddress:      req.ToAddress,
		Amount:         req.Amount,
		ApprovalStatus: approvalStatus,
	}

	if err := h.store.CreatePayout(c.Request.Context(), payout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payout"})
		return
	}

	// Record daily total for limit enforcement
	h.store.IncrementDailyPayoutTotal(c.Request.Context(), merchantID, req.Chain, req.Amount)

	// Send webhook
	if approvalStatus == models.ApprovalStatusPendingApproval {
		h.sendPendingApprovalWebhook(c, payout)
	}

	c.JSON(http.StatusCreated, payout)
}

func (h *PayoutHandler) GetPayout(c *gin.Context) {
	id := c.Param("id")
	merchantID := c.GetString("merchant_id")
	payout, err := h.store.GetPayoutByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payout not found"})
		return
	}
	if payout.MerchantID != merchantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "payout not found"})
		return
	}
	c.JSON(http.StatusOK, payout)
}

func (h *PayoutHandler) sendPendingApprovalWebhook(c *gin.Context, payout *models.Payout) {
	merchant, err := h.store.GetMerchantByID(c.Request.Context(), payout.MerchantID)
	if err != nil || merchant.WebhookURL == "" {
		return
	}
	data := webhook.ApprovalEventData{
		PayoutID: payout.ID,
		Chain:    payout.Chain,
		Amount:   payout.Amount,
		Status:   models.ApprovalStatusPendingApproval,
	}
	go h.webhook.Send(c.Request.Context(), merchant.WebhookURL, merchant.APIKeyHash, webhook.EventPayoutPendingApproval, data)
}
