package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.store.ListRefunds(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list refunds"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AdminRefundHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	refund, err := h.store.AdminGetRefundByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}
	c.JSON(http.StatusOK, refund)
}
