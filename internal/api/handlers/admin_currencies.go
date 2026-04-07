package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminCurrencyHandler struct {
	store store.AdminStore
}

func NewAdminCurrencyHandler(s store.AdminStore) *AdminCurrencyHandler {
	return &AdminCurrencyHandler{store: s}
}

func (h *AdminCurrencyHandler) List(c *gin.Context) {
	currencies, err := h.store.ListAllCurrencies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list currencies"})
		return
	}
	c.JSON(http.StatusOK, currencies)
}
