package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type PayoutHandler struct {
	store    store.Store
	registry *chain.Registry
}

func NewPayoutHandler(s store.Store, registry *chain.Registry) *PayoutHandler {
	return &PayoutHandler{store: s, registry: registry}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported chain: " + req.Chain})
		return
	}

	payout := &models.Payout{
		MerchantID: merchantID,
		Chain:      req.Chain,
		Token:      req.Token,
		ToAddress:  req.ToAddress,
		Amount:     req.Amount,
	}

	if err := h.store.CreatePayout(c.Request.Context(), payout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payout"})
		return
	}

	c.JSON(http.StatusCreated, payout)
}

func (h *PayoutHandler) GetPayout(c *gin.Context) {
	id := c.Param("id")
	payout, err := h.store.GetPayoutByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payout not found"})
		return
	}
	c.JSON(http.StatusOK, payout)
}
