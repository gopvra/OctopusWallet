package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/api/middleware"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type BalanceHandler struct {
	store       store.Store
	ipWhitelist *middleware.IPWhitelist
}

func NewBalanceHandler(s store.Store, ipw *middleware.IPWhitelist) *BalanceHandler {
	return &BalanceHandler{store: s, ipWhitelist: ipw}
}

func (h *BalanceHandler) GetBalances(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	balances, err := h.store.GetMerchantBalances(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"balances": balances})
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
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"payments": payments, "limit": limit, "offset": offset})
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
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"payouts": payouts, "limit": limit, "offset": offset})
}

type SetIPWhitelistRequest struct {
	IPs []string `json:"ips" binding:"required"`
}

func (h *BalanceHandler) SetIPWhitelist(c *gin.Context) {
	var req SetIPWhitelistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	merchantID := c.GetString("merchant_id")
	if err := h.store.SetMerchantIPWhitelist(c.Request.Context(), merchantID, req.IPs); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	// Sync middleware so IP restriction takes effect immediately
	if h.ipWhitelist != nil {
		h.ipWhitelist.SetWhitelist(merchantID, req.IPs)
	}
	R.OK(c, gin.H{"message": "IP whitelist updated", "ips": req.IPs})
}

func (h *BalanceHandler) GetIPWhitelist(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	ips, err := h.store.GetMerchantIPWhitelist(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"ips": ips})
}
