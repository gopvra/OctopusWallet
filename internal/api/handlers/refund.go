package handlers

import (
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type RefundHandler struct {
	store store.Store
}

func NewRefundHandler(s store.Store) *RefundHandler {
	return &RefundHandler{store: s}
}

type CreateRefundRequest struct {
	PaymentID string `json:"payment_id" binding:"required"`
	ToAddress string `json:"to_address" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Reason    string `json:"reason"`
}

func (h *RefundHandler) CreateRefund(c *gin.Context) {
	var req CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID := c.GetString("merchant_id")

	if err := crypto.ValidateAmount(req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify payment belongs to merchant and is completed
	payment, err := h.store.GetPaymentByID(c.Request.Context(), req.PaymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if payment.MerchantID != merchantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "payment does not belong to this merchant"})
		return
	}
	if payment.Status != models.PaymentStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only refund completed payments"})
		return
	}

	// Validate refund amount does not exceed payment amount minus existing refunds
	existingTotal, err := h.store.GetRefundTotalByPayment(c.Request.Context(), req.PaymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing refunds"})
		return
	}
	existing := new(big.Int)
	existing.SetString(existingTotal, 10)
	refundAmt := new(big.Int)
	refundAmt.SetString(req.Amount, 10)
	received := new(big.Int)
	received.SetString(payment.AmountReceived, 10)
	if new(big.Int).Add(existing, refundAmt).Cmp(received) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refund total would exceed payment amount received"})
		return
	}

	if err := crypto.ValidateAddress(payment.Chain, req.ToAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund := &models.Refund{
		PaymentID:  req.PaymentID,
		MerchantID: merchantID,
		Chain:      payment.Chain,
		Token:      payment.Token,
		ToAddress:  req.ToAddress,
		Amount:     req.Amount,
		Reason:     req.Reason,
	}

	if err := h.store.CreateRefund(c.Request.Context(), refund); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refund"})
		return
	}

	c.JSON(http.StatusCreated, refund)
}

func (h *RefundHandler) GetRefund(c *gin.Context) {
	id := c.Param("id")
	merchantID := c.GetString("merchant_id")
	refund, err := h.store.GetRefundByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}
	if refund.MerchantID != merchantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}
	c.JSON(http.StatusOK, refund)
}

func (h *RefundHandler) ListRefundsByPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")
	merchantID := c.GetString("merchant_id")
	// Verify payment belongs to merchant
	payment, err := h.store.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil || payment.MerchantID != merchantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	refunds, err := h.store.GetRefundsByPayment(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list refunds"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"refunds": refunds})
}
