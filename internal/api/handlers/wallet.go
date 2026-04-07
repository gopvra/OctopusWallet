package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type WalletHandler struct {
	store store.Store
}

func NewWalletHandler(s store.Store) *WalletHandler {
	return &WalletHandler{store: s}
}

func (h *WalletHandler) List(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	wallets, err := h.store.GetWalletsByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list wallets"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallets": wallets})
}
