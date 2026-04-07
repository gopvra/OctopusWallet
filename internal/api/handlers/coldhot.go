package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	merchantID := c.GetString("merchant_id")
	cfg := &models.ColdWalletConfig{
		MerchantID:          merchantID,
		Chain:               req.Chain,
		ColdWalletAddress:   req.ColdWalletAddress,
		HotWalletMaxBalance: req.HotWalletMaxBalance,
		Enabled:             req.Enabled,
	}
	if err := h.store.UpsertColdWalletConfig(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cold wallet config saved"})
}

func (h *ColdHotHandler) GetConfig(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	configs, err := h.store.GetColdWalletConfigsByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get configs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func (h *ColdHotHandler) ListTransfers(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	transfers, err := h.store.GetWalletTransfersByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transfers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transfers": transfers})
}
