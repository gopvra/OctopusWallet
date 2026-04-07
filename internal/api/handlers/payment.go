package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
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
	Chain  string `json:"chain" binding:"required"`
	Amount string `json:"amount" binding:"required"`
	Token  string `json:"token"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID := c.GetString("merchant_id")

	// Get chain implementation
	chainImpl, err := h.registry.Get(req.Chain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported chain: " + req.Chain})
		return
	}

	// Get next derivation index for this merchant on this chain
	nextIndex, err := h.store.GetNextDerivationIndex(c.Request.Context(), merchantID, req.Chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get derivation index"})
		return
	}

	// Derive a fresh address
	// merchantIndex is derived from first 4 bytes of merchant UUID for deterministic mapping
	merchantIndex := merchantIDToIndex(merchantID)
	address, err := chainImpl.DeriveAddress(h.seed, merchantIndex, uint32(nextIndex))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to derive address"})
		return
	}

	// Create wallet record
	wallet := &models.Wallet{
		MerchantID:      merchantID,
		Chain:           req.Chain,
		Address:         address,
		DerivationIndex: nextIndex,
	}
	if err := h.store.CreateWallet(c.Request.Context(), wallet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wallet"})
		return
	}

	// Create payment with 30 minute expiry
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}

	c.JSON(http.StatusCreated, CreatePaymentResponse{
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
	payment, err := h.store.GetPaymentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, payment)
}

// merchantIDToIndex converts a UUID merchant ID to a uint32 index for HD derivation.
func merchantIDToIndex(id string) uint32 {
	var sum uint32
	for _, b := range []byte(id) {
		sum = sum*31 + uint32(b)
	}
	return sum & 0x7FFFFFFF // ensure non-negative
}
