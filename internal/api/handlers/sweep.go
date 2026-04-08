package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
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
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	merchantID := c.GetString("merchant_id")

	if err := crypto.ValidateAddress(req.Chain, req.Address); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	if req.SweepThreshold != "" {
		if err := crypto.ValidateAmountOrZero(req.SweepThreshold); err != nil {
			R.FailMsg(c, errcode.ErrBadRequest, "sweep_threshold: "+err.Error())
			return
		}
	}

	addr := &models.MerchantCollectionAddress{
		MerchantID:     merchantID,
		Chain:          req.Chain,
		Address:        req.Address,
		SweepThreshold: req.SweepThreshold,
		IsActive:       true,
	}
	if err := h.store.UpsertMerchantCollectionAddress(c.Request.Context(), addr); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"message": "collection address saved"})
}

func (h *SweepHandler) GetCollectionAddresses(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	addrs, err := h.store.GetMerchantCollectionAddresses(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"collection_addresses": addrs})
}

func (h *SweepHandler) ListTasks(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	tasks, err := h.store.GetSweepTasksByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"sweep_tasks": tasks})
}
