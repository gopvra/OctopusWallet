package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AuditLogHandler struct {
	store store.Store
}

func NewAuditLogHandler(s store.Store) *AuditLogHandler {
	return &AuditLogHandler{store: s}
}

func (h *AuditLogHandler) List(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	logs, err := h.store.GetAuditLogs(c.Request.Context(), merchantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"audit_logs": logs, "limit": limit, "offset": offset})
}
