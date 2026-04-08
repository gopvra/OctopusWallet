package handlers

import (
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type PaymentHandler struct {
	store    store.Store
	registry *chain.Registry
	seed     []byte
}

func NewPaymentHandler(s store.Store, registry *chain.Registry, seed []byte) *PaymentHandler {
	return &PaymentHandler{store: s, registry: registry, seed: seed}
}

type CreatePaymentRequest struct {
	Chain       string `json:"chain" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
	Token       string `json:"token"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	OrderID     string `json:"order_id"`
	RedirectURL string `json:"redirect_url"`
}

type CreatePaymentResponse struct {
	ID      string     `json:"id"`
	Chain   string     `json:"chain"`
	Address string     `json:"address"`
	Amount  string     `json:"amount"`
	Token   string     `json:"token"`
	Status  string     `json:"status"`
	Expires *time.Time `json:"expires_at,omitempty"`
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	merchantID := c.GetString("merchant_id")

	if req.RedirectURL != "" {
		u, err := url.Parse(req.RedirectURL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
			R.Fail(c, errcode.ErrInvalidURL)
			return
		}
	}

	if err := crypto.ValidateAmount(req.Amount); err != nil {
		R.Fail(c, errcode.ErrInvalidAmount)
		return
	}

	chainImpl, err := h.registry.Get(req.Chain)
	if err != nil {
		R.Fail(c, errcode.ErrUnsupportedChain)
		return
	}

	nextIndex, err := h.store.GetNextDerivationIndex(c.Request.Context(), merchantID, req.Chain)
	if err != nil {
		R.Fail(c, errcode.ErrDerivationIndexFailed)
		return
	}

	merchantIndex := crypto.MerchantIDToIndex(merchantID)
	address, err := chainImpl.DeriveAddress(h.seed, merchantIndex, uint32(nextIndex))
	if err != nil {
		R.Fail(c, errcode.ErrDeriveAddressFailed)
		return
	}

	wallet := &models.Wallet{
		MerchantID:      merchantID,
		Chain:           req.Chain,
		Address:         address,
		DerivationIndex: nextIndex,
	}
	if err := h.store.CreateWallet(c.Request.Context(), wallet); err != nil {
		R.Fail(c, errcode.ErrWalletCreateFailed)
		return
	}

	expiresAt := time.Now().Add(30 * time.Minute)
	payment := &models.Payment{
		MerchantID:     merchantID,
		Chain:          req.Chain,
		Token:          req.Token,
		AmountExpected: req.Amount,
		Address:        address,
		ExpiresAt:      &expiresAt,
	}
	if err := h.store.CreatePayment(c.Request.Context(), payment); err != nil {
		R.Fail(c, errcode.ErrPaymentCreateFailed)
		return
	}

	R.OK(c, CreatePaymentResponse{
		ID:      payment.ID,
		Chain:   req.Chain,
		Address: address,
		Amount:  req.Amount,
		Token:   req.Token,
		Status:  payment.Status,
		Expires: payment.ExpiresAt,
	})
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	merchantID := c.GetString("merchant_id")
	payment, err := h.store.GetPaymentByID(c.Request.Context(), id)
	if err != nil || payment.MerchantID != merchantID {
		R.Fail(c, errcode.ErrPaymentNotFound)
		return
	}
	R.OK(c, payment)
}
