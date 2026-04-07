package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.store.ListPayments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payments"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AdminPaymentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	payment, err := h.store.AdminGetPaymentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, payment)
}
