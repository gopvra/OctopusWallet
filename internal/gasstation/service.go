package gasstation

import (
	"context"
	"encoding/hex"
	"log/slog"
	"math/big"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
)

type Service struct {
	store    store.Store
	registry *chain.Registry
	webhook  *webhook.Service
	config   config.GasStationConfig
}

func NewService(s store.Store, registry *chain.Registry, wh *webhook.Service, cfg config.GasStationConfig) *Service {
	return &Service{store: s, registry: registry, webhook: wh, config: cfg}
}

// RequestGasDeposit creates a gas deposit request for a sweep task.
func (s *Service) RequestGasDeposit(ctx context.Context, chainName, toAddress string, sweepTaskID string) error {
	chainCfg, ok := s.config.Chains[chainName]
	if !ok {
		return nil
	}

	deposit := &models.GasDeposit{
		Chain:       chainName,
		ToAddress:   toAddress,
		Amount:      chainCfg.GasAmount,
		SweepTaskID: &sweepTaskID,
	}
	if err := s.store.CreateGasDeposit(ctx, deposit); err != nil {
		return err
	}

	// Link gas deposit to sweep task
	return s.store.UpdateSweepTaskGasDeposit(ctx, sweepTaskID, deposit.ID)
}

func (s *Service) Start(ctx context.Context) {
	slog.Info("gas station service started")

	depositTicker := time.NewTicker(15 * time.Second)
	balanceTicker := time.NewTicker(5 * time.Minute)
	defer depositTicker.Stop()
	defer balanceTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("gas station service stopped")
			return
		case <-depositTicker.C:
			s.processDeposits(ctx)
		case <-balanceTicker.C:
			s.checkBalances(ctx)
		}
	}
}

func (s *Service) processDeposits(ctx context.Context) {
	deposits, err := s.store.GetPendingGasDeposits(ctx)
	if err != nil {
		slog.Error("failed to get pending gas deposits", "error", err)
		return
	}

	for _, deposit := range deposits {
		s.processDeposit(ctx, &deposit)
	}
}

func (s *Service) processDeposit(ctx context.Context, deposit *models.GasDeposit) {
	chainCfg, ok := s.config.Chains[deposit.Chain]
	if !ok || chainCfg.StationPrivateKey == "" {
		errMsg := "gas station not configured for chain " + deposit.Chain
		s.store.UpdateGasDepositStatus(ctx, deposit.ID, models.GasDepositStatusFailed, nil, &errMsg)
		return
	}

	chainImpl, err := s.registry.Get(deposit.Chain)
	if err != nil {
		return
	}

	// Update to processing
	s.store.UpdateGasDepositStatus(ctx, deposit.ID, models.GasDepositStatusProcessing, nil, nil)

	// Send gas from station
	privKeyBytes, err := hexToBytes(chainCfg.StationPrivateKey)
	if err != nil {
		errMsg := "invalid gas station private key: " + err.Error()
		s.store.UpdateGasDepositStatus(ctx, deposit.ID, models.GasDepositStatusFailed, nil, &errMsg)
		return
	}
	defer zeroBytes(privKeyBytes)
	txHash, err := chainImpl.SendTransaction(ctx, chain.SendRequest{
		FromAddress: chainCfg.StationAddress,
		ToAddress:   deposit.ToAddress,
		Amount:      deposit.Amount,
		Token:       "", // always native
		PrivateKey:  privKeyBytes,
	})
	if err != nil {
		errMsg := "gas deposit tx failed: " + err.Error()
		s.store.UpdateGasDepositStatus(ctx, deposit.ID, models.GasDepositStatusFailed, nil, &errMsg)
		slog.Error("gas deposit failed", "deposit_id", deposit.ID, "error", err)
		return
	}

	s.store.UpdateGasDepositStatus(ctx, deposit.ID, models.GasDepositStatusCompleted, &txHash, nil)

	// If linked to a sweep task, set it back to pending so sweep service retries
	if deposit.SweepTaskID != nil {
		s.store.UpdateSweepTaskStatus(ctx, *deposit.SweepTaskID, models.SweepStatusPending, nil, nil)
	}

	slog.Info("gas deposited", "chain", deposit.Chain, "to", deposit.ToAddress, "tx", txHash)
}

func (s *Service) checkBalances(ctx context.Context) {
	for chainName, chainCfg := range s.config.Chains {
		if chainCfg.StationAddress == "" || chainCfg.LowBalanceAlert == "" {
			continue
		}

		chainImpl, err := s.registry.Get(chainName)
		if err != nil {
			continue
		}

		balance, err := chainImpl.GetBalance(ctx, chainCfg.StationAddress, "")
		if err != nil {
			continue
		}

		balanceBig := new(big.Int)
		balanceBig.SetString(balance, 10)
		threshold := new(big.Int)
		threshold.SetString(chainCfg.LowBalanceAlert, 10)

		if balanceBig.Cmp(threshold) < 0 {
			alert := &models.GasAlert{
				Chain:          chainName,
				StationAddress: chainCfg.StationAddress,
				Balance:        balance,
				Threshold:      chainCfg.LowBalanceAlert,
			}
			s.store.CreateGasAlert(ctx, alert)
			slog.Warn("gas station low balance", "chain", chainName, "balance", balance, "threshold", chainCfg.LowBalanceAlert)
		}
	}
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func hexToBytes(hexStr string) ([]byte, error) {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	return hex.DecodeString(hexStr)
}
