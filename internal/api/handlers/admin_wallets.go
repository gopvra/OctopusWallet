package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	result, err := h.store.ListWallets(c.Request.Context(), filter)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, result)
}
