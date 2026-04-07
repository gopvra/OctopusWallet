package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.store.ListPayouts(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payouts"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AdminPayoutHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	payout, err := h.store.AdminGetPayoutByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payout not found"})
		return
	}
	c.JSON(http.StatusOK, payout)
}
