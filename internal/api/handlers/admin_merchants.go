package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/google/uuid"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type AdminMerchantHandler struct {
	store store.AdminStore
}

func NewAdminMerchantHandler(s store.AdminStore) *AdminMerchantHandler {
	return &AdminMerchantHandler{store: s}
}

func (h *AdminMerchantHandler) List(c *gin.Context) {
	var filter store.MerchantFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	result, err := h.store.ListMerchants(c.Request.Context(), filter)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, result)
}

func (h *AdminMerchantHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	merchant, err := h.store.AdminGetMerchantByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	R.OK(c, merchant)
}

func (h *AdminMerchantHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	var req struct {
		Name       string `json:"name" binding:"required"`
		Email      string `json:"email" binding:"required,email"`
		WebhookURL string `json:"webhook_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := crypto.ValidateWebhookURL(req.WebhookURL); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	if err := h.store.UpdateMerchant(c.Request.Context(), id, req.Name, req.Email, req.WebhookURL); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"message": "merchant updated"})
}

func (h *AdminMerchantHandler) ToggleActive(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	if err := h.store.ToggleMerchantActive(c.Request.Context(), id); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"message": "merchant status toggled"})
}
