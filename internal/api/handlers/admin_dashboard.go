package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminDashboardHandler struct {
	store store.AdminStore
}

func NewAdminDashboardHandler(s store.AdminStore) *AdminDashboardHandler {
	return &AdminDashboardHandler{store: s}
}

func (h *AdminDashboardHandler) Stats(c *gin.Context) {
	stats, err := h.store.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *AdminDashboardHandler) VolumeChart(c *gin.Context) {
	days := 7
	if d, err := strconv.Atoi(c.DefaultQuery("days", "7")); err == nil && d > 0 && d <= 365 {
		days = d
	}

	points, err := h.store.GetVolumeChart(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get volume chart"})
		return
	}
	c.JSON(http.StatusOK, points)
}

func (h *AdminDashboardHandler) ChainDistribution(c *gin.Context) {
	dist, err := h.store.GetChainDistribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chain distribution"})
		return
	}
	c.JSON(http.StatusOK, dist)
}

func (h *AdminDashboardHandler) RecentActivity(c *gin.Context) {
	limit := 20
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	activity, err := h.store.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get recent activity"})
		return
	}
	c.JSON(http.StatusOK, activity)
}
