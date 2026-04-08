package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
			R.Fail(c, errcode.ErrInternalServer)
			return
		}
		R.OK(c, gin.H{"currencies": currencies})
		return
	}
	currencies, err := h.store.GetSupportedCurrencies(c.Request.Context())
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"currencies": currencies})
}

// GetExchangeRate returns estimated fee/rate for a chain (placeholder for rate API)
func (h *CurrencyHandler) GetExchangeRate(c *gin.Context) {
	chainName := c.Query("chain")
	if chainName == "" {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	chainImpl, err := h.registry.Get(chainName)
	if err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}

	fee, err := chainImpl.EstimateFee(c.Request.Context(), chain.SendRequest{Amount: "0"})
	if err != nil {
		fee = "unknown"
	}

	R.OK(c, gin.H{
		"chain":          chainName,
		"native_symbol":  chainImpl.NativeSymbol(),
		"estimated_fee":  fee,
	})
}
