package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	balances, err := h.store.ListAllMerchantBalances(c.Request.Context(), filter)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, balances)
}
