package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type PaymentLinkHandler struct {
	store    store.Store
	registry *chain.Registry
}

func NewPaymentLinkHandler(s store.Store, r *chain.Registry) *PaymentLinkHandler {
	return &PaymentLinkHandler{store: s, registry: r}
}

type CreatePaymentLinkRequest struct {
	Chain       string `json:"chain" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
	Token       string `json:"token"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	RedirectURL string `json:"redirect_url"`
	IsReusable  bool   `json:"is_reusable"`
}

func (h *PaymentLinkHandler) Create(c *gin.Context) {
	var req CreatePaymentLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}
	if _, err := h.registry.Get(req.Chain); err != nil {
		R.Fail(c, errcode.ErrBadRequest)
		return
	}
	if err := crypto.ValidateAmount(req.Amount); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	merchantID := c.GetString("merchant_id")
	link := &models.PaymentLink{
		MerchantID:  merchantID,
		Chain:       req.Chain,
		Token:       req.Token,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		RedirectURL: req.RedirectURL,
		IsReusable:  req.IsReusable,
	}
	if err := h.store.CreatePaymentLink(c.Request.Context(), link); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{
		"payment_link": link,
		"url":          "/pay/link/" + link.ID,
	})
}

func (h *PaymentLinkHandler) List(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	links, err := h.store.GetPaymentLinksByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"payment_links": links})
}

// GetPublic is a public endpoint — no auth required. Returns link info for checkout.
func (h *PaymentLinkHandler) GetPublic(c *gin.Context) {
	id := c.Param("id")
	link, err := h.store.GetPaymentLinkByID(c.Request.Context(), id)
	if err != nil {
		R.Fail(c, errcode.ErrNotFound)
		return
	}
	if !link.IsReusable && link.UsesCount > 0 {
		R.Fail(c, errcode.ErrPaymentLinkInactive)
		return
	}
	R.OK(c, link)
}
