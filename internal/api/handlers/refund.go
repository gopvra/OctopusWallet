package handlers

import (
	"math/big"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	merchantID := c.GetString("merchant_id")

	if err := crypto.ValidateAmount(req.Amount); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	// Verify payment belongs to merchant and is completed
	payment, err := h.store.GetPaymentByID(c.Request.Context(), req.PaymentID)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	if payment.MerchantID != merchantID {
		R.Fail(c, errcode.ErrForbidden)
		return
	}
	if payment.Status != models.PaymentStatusCompleted {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	// Validate refund amount does not exceed payment amount minus existing refunds
	existingTotal, err := h.store.GetRefundTotalByPayment(c.Request.Context(), req.PaymentID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	existing := new(big.Int)
	existing.SetString(existingTotal, 10)
	refundAmt := new(big.Int)
	refundAmt.SetString(req.Amount, 10)
	received := new(big.Int)
	received.SetString(payment.AmountReceived, 10)
	if new(big.Int).Add(existing, refundAmt).Cmp(received) > 0 {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	if err := crypto.ValidateAddress(payment.Chain, req.ToAddress); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
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
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, refund)
}

func (h *RefundHandler) GetRefund(c *gin.Context) {
	id := c.Param("id")
	merchantID := c.GetString("merchant_id")
	refund, err := h.store.GetRefundByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	if refund.MerchantID != merchantID {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	R.OK(c, refund)
}

func (h *RefundHandler) ListRefundsByPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")
	merchantID := c.GetString("merchant_id")
	// Verify payment belongs to merchant
	payment, err := h.store.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil || payment.MerchantID != merchantID {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	refunds, err := h.store.GetRefundsByPayment(c.Request.Context(), paymentID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"refunds": refunds})
}
