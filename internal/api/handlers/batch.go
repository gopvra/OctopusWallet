package handlers

import (
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type BatchPayoutHandler struct {
	store    store.Store
	registry *chain.Registry
}

func NewBatchPayoutHandler(s store.Store, registry *chain.Registry) *BatchPayoutHandler {
	return &BatchPayoutHandler{store: s, registry: registry}
}

type BatchPayoutItemReq struct {
	ToAddress string `json:"to_address" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
}

type CreateBatchPayoutRequest struct {
	Chain string             `json:"chain" binding:"required"`
	Token string             `json:"token"`
	Items []BatchPayoutItemReq `json:"items" binding:"required,min=1,max=100"`
}

func (h *BatchPayoutHandler) CreateBatchPayout(c *gin.Context) {
	var req CreateBatchPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID := c.GetString("merchant_id")

	if _, err := h.registry.Get(req.Chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported chain"})
		return
	}

	// Validate all items
	totalAmount := new(big.Int)
	for i, item := range req.Items {
		if err := crypto.ValidateAmount(item.Amount); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "item": i})
			return
		}
		if err := crypto.ValidateAddress(req.Chain, item.ToAddress); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "item": i})
			return
		}
		amt := new(big.Int)
		amt.SetString(item.Amount, 10)
		totalAmount.Add(totalAmount, amt)
	}

	batch := &models.BatchPayout{
		MerchantID:  merchantID,
		Chain:       req.Chain,
		Token:       req.Token,
		TotalAmount: totalAmount.String(),
		TotalCount:  len(req.Items),
	}

	if err := h.store.CreateBatchPayout(c.Request.Context(), batch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create batch"})
		return
	}

	// Create individual items with approval check per item
	var failedItems int
	for _, item := range req.Items {
		batchItem := &models.BatchPayoutItem{
			BatchID:   batch.ID,
			ToAddress: item.ToAddress,
			Amount:    item.Amount,
		}
		if err := h.store.CreateBatchPayoutItem(c.Request.Context(), batchItem); err != nil {
			failedItems++
			continue
		}

		// Check approval limits per item (same rules as single payout)
		approvalStatus, rejectReason, _ := CheckPayoutLimits(c, h.store, merchantID, req.Chain, item.Amount)
		if rejectReason != "" {
			errMsg := rejectReason
			batchItem.ErrorMessage = &errMsg
			failedItems++
			continue
		}

		payout := &models.Payout{
			MerchantID:     merchantID,
			Chain:          req.Chain,
			Token:          req.Token,
			ToAddress:      item.ToAddress,
			Amount:         item.Amount,
			ApprovalStatus: approvalStatus,
		}
		if err := h.store.CreatePayout(c.Request.Context(), payout); err != nil {
			failedItems++
			continue
		}
		h.store.IncrementDailyPayoutTotal(c.Request.Context(), merchantID, req.Chain, item.Amount)
	}

	c.JSON(http.StatusCreated, batch)
}

func (h *BatchPayoutHandler) GetBatchPayout(c *gin.Context) {
	id := c.Param("id")
	merchantID := c.GetString("merchant_id")
	batch, err := h.store.GetBatchPayoutByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "batch not found"})
		return
	}
	if batch.MerchantID != merchantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "batch not found"})
		return
	}
	items, _ := h.store.GetBatchPayoutItems(c.Request.Context(), id)
	c.JSON(http.StatusOK, gin.H{"batch": batch, "items": items})
}

func (h *BatchPayoutHandler) ListBatchPayouts(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	batches, err := h.store.GetBatchPayoutsByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list batches"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"batches": batches})
}
