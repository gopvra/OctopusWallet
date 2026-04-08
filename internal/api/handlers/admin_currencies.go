package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, currencies)
}
