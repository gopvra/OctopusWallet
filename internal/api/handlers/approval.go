package handlers

import (
	"database/sql"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
)

type ApprovalHandler struct {
	store   store.Store
	webhook *webhook.Service
}

func NewApprovalHandler(s store.Store, wh *webhook.Service) *ApprovalHandler {
	return &ApprovalHandler{store: s, webhook: wh}
}

type SetApprovalConfigRequest struct {
	ApprovalThreshold string `json:"approval_threshold" binding:"required"`
	SingleTxLimit     string `json:"single_tx_limit"`
	DailyLimit        string `json:"daily_limit"`
	AutoRelease       bool   `json:"auto_release"`
	Enabled           bool   `json:"enabled"`
}

func (h *ApprovalHandler) SetConfig(c *gin.Context) {
	var req SetApprovalConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID := c.GetString("merchant_id")
	cfg := &models.ApprovalConfig{
		MerchantID:        merchantID,
		ApprovalThreshold: req.ApprovalThreshold,
		SingleTxLimit:     req.SingleTxLimit,
		DailyLimit:        req.DailyLimit,
		AutoRelease:       req.AutoRelease,
		Enabled:           req.Enabled,
	}

	if err := h.store.UpsertApprovalConfig(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "approval config saved"})
}

func (h *ApprovalHandler) GetConfig(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	cfg, err := h.store.GetApprovalConfig(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"config": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

type ApproveRejectRequest struct {
	ApproverID string `json:"approver_id" binding:"required"`
	Note       string `json:"note"`
}

func (h *ApprovalHandler) ApprovePayout(c *gin.Context) {
	payoutID := c.Param("id")
	var req ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID := c.GetString("merchant_id")
	payout, err := h.store.GetPayoutByID(c.Request.Context(), payoutID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payout not found"})
		return
	}
	if payout.MerchantID != merchantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "payout does not belong to this merchant"})
		return
	}
	if payout.ApprovalStatus != models.ApprovalStatusPendingApproval {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payout is not pending approval"})
		return
	}

	// Create approval record
	approval := &models.PayoutApproval{
		PayoutID:     payoutID,
		MerchantID:   merchantID,
		Action:       "approved",
		ApproverID:   req.ApproverID,
		ApproverNote: req.Note,
	}
	if err := h.store.CreatePayoutApproval(c.Request.Context(), approval); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create approval"})
		return
	}

	// Update payout approval status
	if err := h.store.UpdatePayoutApprovalStatus(c.Request.Context(), payoutID, models.ApprovalStatusApproved); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payout"})
		return
	}

	// Send webhook
	h.sendApprovalWebhook(c, payout, models.ApprovalStatusApproved, req.ApproverID)

	c.JSON(http.StatusOK, gin.H{"message": "payout approved", "payout_id": payoutID})
}

func (h *ApprovalHandler) RejectPayout(c *gin.Context) {
	payoutID := c.Param("id")
	var req ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID := c.GetString("merchant_id")
	payout, err := h.store.GetPayoutByID(c.Request.Context(), payoutID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payout not found"})
		return
	}
	if payout.MerchantID != merchantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "payout does not belong to this merchant"})
		return
	}
	if payout.ApprovalStatus != models.ApprovalStatusPendingApproval {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payout is not pending approval"})
		return
	}

	// Create rejection record
	approval := &models.PayoutApproval{
		PayoutID:     payoutID,
		MerchantID:   merchantID,
		Action:       "rejected",
		ApproverID:   req.ApproverID,
		ApproverNote: req.Note,
	}
	h.store.CreatePayoutApproval(c.Request.Context(), approval)

	// Update payout
	h.store.UpdatePayoutApprovalStatus(c.Request.Context(), payoutID, models.ApprovalStatusRejected)
	errMsg := "rejected by " + req.ApproverID
	h.store.UpdatePayoutStatus(c.Request.Context(), payoutID, models.PayoutStatusRejected, nil, &errMsg)

	h.sendApprovalWebhook(c, payout, models.ApprovalStatusRejected, req.ApproverID)

	c.JSON(http.StatusOK, gin.H{"message": "payout rejected", "payout_id": payoutID})
}

func (h *ApprovalHandler) sendApprovalWebhook(c *gin.Context, payout *models.Payout, status, approver string) {
	merchant, err := h.store.GetMerchantByID(c.Request.Context(), payout.MerchantID)
	if err != nil || merchant.WebhookURL == "" {
		return
	}
	var eventType webhook.EventType
	if status == models.ApprovalStatusApproved {
		eventType = webhook.EventPayoutApproved
	} else {
		eventType = webhook.EventPayoutRejected
	}
	data := webhook.ApprovalEventData{
		PayoutID: payout.ID,
		Chain:    payout.Chain,
		Amount:   payout.Amount,
		Status:   status,
		Approver: approver,
	}
	go h.webhook.Send(c.Request.Context(), merchant.WebhookURL, merchant.APIKeyHash, eventType, data)
}

// CheckPayoutLimits checks approval config and returns the approval_status for a new payout.
// Returns: approvalStatus, error message (if rejected), error
func CheckPayoutLimits(ctx *gin.Context, s store.Store, merchantID, chain, amount string) (string, string, error) {
	cfg, err := s.GetApprovalConfig(ctx.Request.Context(), merchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ApprovalStatusNone, "", nil // no config = auto release
		}
		return "", "", err
	}
	if !cfg.Enabled {
		return models.ApprovalStatusNone, "", nil
	}

	amountBig := new(big.Int)
	amountBig.SetString(amount, 10)

	// Check single tx limit
	if cfg.SingleTxLimit != "0" && cfg.SingleTxLimit != "" {
		limit := new(big.Int)
		limit.SetString(cfg.SingleTxLimit, 10)
		if amountBig.Cmp(limit) > 0 {
			return "", "amount exceeds single transaction limit", nil
		}
	}

	// Check daily limit
	if cfg.DailyLimit != "0" && cfg.DailyLimit != "" {
		dailyTotal, err := s.GetDailyPayoutTotal(ctx.Request.Context(), merchantID, chain)
		if err != nil {
			return "", "", err
		}
		totalBig := new(big.Int)
		totalBig.SetString(dailyTotal, 10)
		newTotal := new(big.Int).Add(totalBig, amountBig)
		limit := new(big.Int)
		limit.SetString(cfg.DailyLimit, 10)
		if newTotal.Cmp(limit) > 0 {
			return "", "amount would exceed daily limit", nil
		}
	}

	// Determine approval status
	if cfg.AutoRelease {
		threshold := new(big.Int)
		threshold.SetString(cfg.ApprovalThreshold, 10)
		if threshold.Sign() > 0 && amountBig.Cmp(threshold) < 0 {
			return models.ApprovalStatusNone, "", nil // auto release
		}
	}

	return models.ApprovalStatusPendingApproval, "", nil
}
