package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminWalletHandler struct {
	store store.AdminStore
}

func NewAdminWalletHandler(s store.AdminStore) *AdminWalletHandler {
	return &AdminWalletHandler{store: s}
}

func (h *AdminWalletHandler) List(c *gin.Context) {
	var filter store.WalletFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.store.ListWallets(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list wallets"})
		return
	}
	c.JSON(http.StatusOK, result)
}
