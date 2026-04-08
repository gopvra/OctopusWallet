package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/google/uuid"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminPaymentHandler struct {
	store store.AdminStore
}

func NewAdminPaymentHandler(s store.AdminStore) *AdminPaymentHandler {
	return &AdminPaymentHandler{store: s}
}

func (h *AdminPaymentHandler) List(c *gin.Context) {
	var filter store.PaymentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	result, err := h.store.ListPayments(c.Request.Context(), filter)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, result)
}

func (h *AdminPaymentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	payment, err := h.store.AdminGetPaymentByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	R.OK(c, payment)
}
