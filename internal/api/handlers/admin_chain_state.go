package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list chain states"})
		return
	}
	c.JSON(http.StatusOK, states)
}
