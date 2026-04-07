package handlers

import (

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/config"
)

type GasStationHandler struct {
	registry  *chain.Registry
	gasConfig config.GasStationConfig
}

func NewGasStationHandler(registry *chain.Registry, gasConfig config.GasStationConfig) *GasStationHandler {
	return &GasStationHandler{registry: registry, gasConfig: gasConfig}
}

type GasStatus struct {
	Chain   string `json:"chain"`
	Address string `json:"address"`
	Balance string `json:"balance"`
}

func (h *GasStationHandler) GetStatus(c *gin.Context) {
	var statuses []GasStatus
	for chainName, chainCfg := range h.gasConfig.Chains {
		if chainCfg.StationAddress == "" {
			continue
		}
		chainImpl, err := h.registry.Get(chainName)
		if err != nil {
			continue
		}
		balance, err := chainImpl.GetBalance(c.Request.Context(), chainCfg.StationAddress, "")
		if err != nil {
			balance = "error"
		}
		statuses = append(statuses, GasStatus{
			Chain:   chainName,
			Address: chainCfg.StationAddress,
			Balance: balance,
		})
	}
	R.OK(c, gin.H{"gas_stations": statuses})
}
