package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/google/uuid"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminRefundHandler struct {
	store store.AdminStore
}

func NewAdminRefundHandler(s store.AdminStore) *AdminRefundHandler {
	return &AdminRefundHandler{store: s}
}

func (h *AdminRefundHandler) List(c *gin.Context) {
	var filter store.RefundFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	result, err := h.store.ListRefunds(c.Request.Context(), filter)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, result)
}

func (h *AdminRefundHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	refund, err := h.store.AdminGetRefundByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	R.OK(c, refund)
}
