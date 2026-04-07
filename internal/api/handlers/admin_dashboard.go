package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
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
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, stats)
}

func (h *AdminDashboardHandler) VolumeChart(c *gin.Context) {
	days := 7
	if d, err := strconv.Atoi(c.DefaultQuery("days", "7")); err == nil && d > 0 && d <= 365 {
		days = d
	}

	points, err := h.store.GetVolumeChart(c.Request.Context(), days)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, points)
}

func (h *AdminDashboardHandler) ChainDistribution(c *gin.Context) {
	dist, err := h.store.GetChainDistribution(c.Request.Context())
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, dist)
}

func (h *AdminDashboardHandler) RecentActivity(c *gin.Context) {
	limit := 20
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	activity, err := h.store.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, activity)
}
