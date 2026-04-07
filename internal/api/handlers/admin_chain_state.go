package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type AdminChainStateHandler struct {
	store store.AdminStore
}

func NewAdminChainStateHandler(s store.AdminStore) *AdminChainStateHandler {
	return &AdminChainStateHandler{store: s}
}

func (h *AdminChainStateHandler) List(c *gin.Context) {
	states, err := h.store.ListChainStates(c.Request.Context())
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, states)
}
