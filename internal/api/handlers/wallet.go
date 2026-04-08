package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"wallets": wallets})
}
