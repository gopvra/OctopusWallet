package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/google/uuid"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminBatchPayoutHandler struct {
	store store.AdminStore
}

func NewAdminBatchPayoutHandler(s store.AdminStore) *AdminBatchPayoutHandler {
	return &AdminBatchPayoutHandler{store: s}
}

func (h *AdminBatchPayoutHandler) List(c *gin.Context) {
	var filter store.BatchPayoutFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	result, err := h.store.ListBatchPayouts(c.Request.Context(), filter)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, result)
}

func (h *AdminBatchPayoutHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	batch, err := h.store.AdminGetBatchPayoutByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}

	items, _ := h.store.AdminGetBatchPayoutItems(c.Request.Context(), id)

	R.OK(c, gin.H{
		"batch": batch,
		"items": items,
	})
}
