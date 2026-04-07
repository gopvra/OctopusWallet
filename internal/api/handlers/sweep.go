package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type SweepHandler struct {
	store store.Store
}

func NewSweepHandler(s store.Store) *SweepHandler {
	return &SweepHandler{store: s}
}

type SetCollectionAddressRequest struct {
	Chain          string `json:"chain" binding:"required"`
	Address        string `json:"address" binding:"required"`
	SweepThreshold string `json:"sweep_threshold"`
}

func (h *SweepHandler) SetCollectionAddress(c *gin.Context) {
	var req SetCollectionAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	merchantID := c.GetString("merchant_id")
	addr := &models.MerchantCollectionAddress{
		MerchantID:     merchantID,
		Chain:          req.Chain,
		Address:        req.Address,
		SweepThreshold: req.SweepThreshold,
		IsActive:       true,
	}
	if err := h.store.UpsertMerchantCollectionAddress(c.Request.Context(), addr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save collection address"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "collection address saved"})
}

func (h *SweepHandler) GetCollectionAddresses(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	addrs, err := h.store.GetMerchantCollectionAddresses(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get collection addresses"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"collection_addresses": addrs})
}

func (h *SweepHandler) ListTasks(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	tasks, err := h.store.GetSweepTasksByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sweep tasks"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sweep_tasks": tasks})
}
