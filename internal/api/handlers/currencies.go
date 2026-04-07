package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type CurrencyHandler struct {
	store    store.Store
	registry *chain.Registry
}

func NewCurrencyHandler(s store.Store, registry *chain.Registry) *CurrencyHandler {
	return &CurrencyHandler{store: s, registry: registry}
}

func (h *CurrencyHandler) ListCurrencies(c *gin.Context) {
	chainFilter := c.Query("chain")
	var err error
	if chainFilter != "" {
		currencies, err := h.store.GetSupportedCurrenciesByChain(c.Request.Context(), chainFilter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list currencies"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"currencies": currencies})
		return
	}
	currencies, err := h.store.GetSupportedCurrencies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list currencies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"currencies": currencies})
}

// GetExchangeRate returns estimated fee/rate for a chain (placeholder for rate API)
func (h *CurrencyHandler) GetExchangeRate(c *gin.Context) {
	chainName := c.Query("chain")
	if chainName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chain parameter required"})
		return
	}

	chainImpl, err := h.registry.Get(chainName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported chain"})
		return
	}

	fee, err := chainImpl.EstimateFee(c.Request.Context(), chain.SendRequest{Amount: "0"})
	if err != nil {
		fee = "unknown"
	}

	c.JSON(http.StatusOK, gin.H{
		"chain":          chainName,
		"native_symbol":  chainImpl.NativeSymbol(),
		"estimated_fee":  fee,
	})
}
