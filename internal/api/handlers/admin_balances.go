package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminBalanceHandler struct {
	store store.AdminStore
}

func NewAdminBalanceHandler(s store.AdminStore) *AdminBalanceHandler {
	return &AdminBalanceHandler{store: s}
}

func (h *AdminBalanceHandler) List(c *gin.Context) {
	var filter store.BalanceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	balances, err := h.store.ListAllMerchantBalances(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list balances"})
		return
	}
	c.JSON(http.StatusOK, balances)
}
