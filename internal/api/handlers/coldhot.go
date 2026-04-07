package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type ColdHotHandler struct {
	store store.Store
}

func NewColdHotHandler(s store.Store) *ColdHotHandler {
	return &ColdHotHandler{store: s}
}

type SetColdWalletConfigRequest struct {
	Chain               string `json:"chain" binding:"required"`
	ColdWalletAddress   string `json:"cold_wallet_address" binding:"required"`
	HotWalletMaxBalance string `json:"hot_wallet_max_balance" binding:"required"`
	Enabled             bool   `json:"enabled"`
}

func (h *ColdHotHandler) SetConfig(c *gin.Context) {
	var req SetColdWalletConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	merchantID := c.GetString("merchant_id")

	if err := crypto.ValidateAddress(req.Chain, req.ColdWalletAddress); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, "cold_wallet_address: "+err.Error())
		return
	}
	if err := crypto.ValidateAmountOrZero(req.HotWalletMaxBalance); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, "hot_wallet_max_balance: "+err.Error())
		return
	}

	cfg := &models.ColdWalletConfig{
		MerchantID:          merchantID,
		Chain:               req.Chain,
		ColdWalletAddress:   req.ColdWalletAddress,
		HotWalletMaxBalance: req.HotWalletMaxBalance,
		Enabled:             req.Enabled,
	}
	if err := h.store.UpsertColdWalletConfig(c.Request.Context(), cfg); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"message": "cold wallet config saved"})
}

func (h *ColdHotHandler) GetConfig(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	configs, err := h.store.GetColdWalletConfigsByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"configs": configs})
}

func (h *ColdHotHandler) ListTransfers(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	transfers, err := h.store.GetWalletTransfersByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"transfers": transfers})
}
