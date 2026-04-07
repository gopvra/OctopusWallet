package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type BalanceHandler struct {
	store store.Store
}

func NewBalanceHandler(s store.Store) *BalanceHandler {
	return &BalanceHandler{store: s}
}

func (h *BalanceHandler) GetBalances(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	balances, err := h.store.GetMerchantBalances(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balances"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balances": balances})
}

// ListPayments with pagination
func (h *BalanceHandler) ListPayments(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	payments, err := h.store.GetPaymentsByMerchant(c.Request.Context(), merchantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payments": payments, "limit": limit, "offset": offset})
}

// ListPayouts with pagination
func (h *BalanceHandler) ListPayouts(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	payouts, err := h.store.GetPayoutsByMerchant(c.Request.Context(), merchantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payouts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payouts": payouts, "limit": limit, "offset": offset})
}

type SetIPWhitelistRequest struct {
	IPs []string `json:"ips" binding:"required"`
}

func (h *BalanceHandler) SetIPWhitelist(c *gin.Context) {
	var req SetIPWhitelistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	merchantID := c.GetString("merchant_id")
	if err := h.store.SetMerchantIPWhitelist(c.Request.Context(), merchantID, req.IPs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set IP whitelist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP whitelist updated", "ips": req.IPs})
}

func (h *BalanceHandler) GetIPWhitelist(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	ips, err := h.store.GetMerchantIPWhitelist(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get IP whitelist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ips": ips})
}
