package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/google/uuid"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminPayoutHandler struct {
	store store.AdminStore
}

func NewAdminPayoutHandler(s store.AdminStore) *AdminPayoutHandler {
	return &AdminPayoutHandler{store: s}
}

func (h *AdminPayoutHandler) List(c *gin.Context) {
	var filter store.PayoutFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	result, err := h.store.ListPayouts(c.Request.Context(), filter)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, result)
}

func (h *AdminPayoutHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	payout, err := h.store.AdminGetPayoutByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	R.OK(c, payout)
}
