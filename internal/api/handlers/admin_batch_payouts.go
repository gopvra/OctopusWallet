package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.store.ListBatchPayouts(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list batch payouts"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AdminBatchPayoutHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	batch, err := h.store.AdminGetBatchPayoutByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "batch payout not found"})
		return
	}

	items, _ := h.store.AdminGetBatchPayoutItems(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"batch": batch,
		"items": items,
	})
}
