package handlers

import (
	"math/big"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	merchantID := c.GetString("merchant_id")

	if _, err := h.registry.Get(req.Chain); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	// Validate all items
	totalAmount := new(big.Int)
	for i, item := range req.Items {
		if err := crypto.ValidateAmount(item.Amount); err != nil {
			R.FailData(c, errcode.ErrBadRequest, gin.H{"detail": err.Error(), "item": i})
			return
		}
		if err := crypto.ValidateAddress(req.Chain, item.ToAddress); err != nil {
			R.FailData(c, errcode.ErrBadRequest, gin.H{"detail": err.Error(), "item": i})
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
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	// Create individual items (each will become a payout)
	var failedItems []int
	for i, item := range req.Items {
		batchItem := &models.BatchPayoutItem{
			BatchID:   batch.ID,
			ToAddress: item.ToAddress,
			Amount:    item.Amount,
		}
		if err := h.store.CreateBatchPayoutItem(c.Request.Context(), batchItem); err != nil {
			failedItems = append(failedItems, i)
			continue
		}

		payout := &models.Payout{
			MerchantID: merchantID,
			Chain:      req.Chain,
			Token:      req.Token,
			ToAddress:  item.ToAddress,
			Amount:     item.Amount,
		}
		if err := h.store.CreatePayout(c.Request.Context(), payout); err != nil {
			failedItems = append(failedItems, i)
		}
	}

	resp := gin.H{"batch": batch}
	if len(failedItems) > 0 {
		resp["failed_items"] = failedItems
		resp["warning"] = "some items failed to create"
	}
	R.OK(c, resp)
}

func (h *BatchPayoutHandler) GetBatchPayout(c *gin.Context) {
	id := c.Param("id")
	merchantID := c.GetString("merchant_id")
	batch, err := h.store.GetBatchPayoutByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	if batch.MerchantID != merchantID {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	items, _ := h.store.GetBatchPayoutItems(c.Request.Context(), id)
	R.OK(c, gin.H{"batch": batch, "items": items})
}

func (h *BatchPayoutHandler) ListBatchPayouts(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	batches, err := h.store.GetBatchPayoutsByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"batches": batches})
}
