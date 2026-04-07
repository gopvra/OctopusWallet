package sweep

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/gasstation"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type Service struct {
	store      store.Store
	registry   *chain.Registry
	webhook    *webhook.Service
	gasStation *gasstation.Service
	seed       []byte
	configs    map[string]config.ChainConfig
}

func NewService(s store.Store, registry *chain.Registry, wh *webhook.Service, gs *gasstation.Service, seed []byte, configs map[string]config.ChainConfig) *Service {
	return &Service{
		store:      s,
		registry:   registry,
		webhook:    wh,
		gasStation: gs,
		seed:       seed,
		configs:    configs,
	}
}

func (s *Service) Start(ctx context.Context) {
	slog.Info("sweep service started")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("sweep service stopped")
			return
		case <-ticker.C:
			s.processPendingSweeps(ctx)
		}
	}
}

// OnPaymentCompleted is called by the monitor service when a payment is completed.
func (s *Service) OnPaymentCompleted(ctx context.Context, payment *models.Payment) {
	collAddr, err := s.store.GetMerchantCollectionAddress(ctx, payment.MerchantID, payment.Chain)
	if err != nil || !collAddr.IsActive {
		return
	}

	// Check threshold
	received := new(big.Int)
	received.SetString(payment.AmountReceived, 10)
	threshold := new(big.Int)
	threshold.SetString(collAddr.SweepThreshold, 10)

	if threshold.Sign() > 0 && received.Cmp(threshold) < 0 {
		return // below threshold
	}

	task := &models.SweepTask{
		MerchantID:  payment.MerchantID,
		PaymentID:   &payment.ID,
		Chain:       payment.Chain,
		Token:       payment.Token,
		FromAddress: payment.Address,
		ToAddress:   collAddr.Address,
		Amount:      payment.AmountReceived,
	}
	if err := s.store.CreateSweepTask(ctx, task); err != nil {
		slog.Error("failed to create sweep task", "payment_id", payment.ID, "error", err)
		return
	}
	slog.Info("sweep task created", "task_id", task.ID, "payment_id", payment.ID)
}

func (s *Service) processPendingSweeps(ctx context.Context) {
	tasks, err := s.store.GetPendingSweepTasks(ctx)
	if err != nil {
		slog.Error("failed to get pending sweep tasks", "error", err)
		return
	}

	for _, task := range tasks {
		if task.Status == models.SweepStatusGasNeeded {
			s.checkGasDeposit(ctx, &task)
		} else {
			s.processSweep(ctx, &task)
		}
	}
}

func (s *Service) processSweep(ctx context.Context, task *models.SweepTask) {
	chainImpl, err := s.registry.Get(task.Chain)
	if err != nil {
		return
	}

	// For token sweeps, check if the address has enough gas
	if task.Token != "" {
		gasBalance, err := chainImpl.GetBalance(ctx, task.FromAddress, "")
		if err == nil {
			gasBig := new(big.Int)
			gasBig.SetString(gasBalance, 10)
			if gasBig.Sign() == 0 {
				// Need gas deposit
				if err := s.gasStation.RequestGasDeposit(ctx, task.Chain, task.FromAddress, task.ID); err != nil {
					slog.Error("failed to request gas deposit", "task_id", task.ID, "error", err)
				}
				s.store.UpdateSweepTaskStatus(ctx, task.ID, models.SweepStatusGasNeeded, nil, nil)
				return
			}
		}
	}

	// Find the wallet to get derivation index
	wallet, err := s.store.GetWalletByAddress(ctx, task.Chain, task.FromAddress)
	if err != nil {
		errMsg := "wallet not found for address"
		s.store.UpdateSweepTaskStatus(ctx, task.ID, models.SweepStatusFailed, nil, &errMsg)
		return
	}

	merchantIndex := merchantIDToIndex(task.MerchantID)
	privKey, err := chainImpl.DerivePrivateKey(s.seed, merchantIndex, uint32(wallet.DerivationIndex))
	if err != nil {
		errMsg := "failed to derive key: " + err.Error()
		s.store.UpdateSweepTaskStatus(ctx, task.ID, models.SweepStatusFailed, nil, &errMsg)
		return
	}

	defer crypto.ZeroBytes(privKey)

	s.store.UpdateSweepTaskStatus(ctx, task.ID, models.SweepStatusProcessing, nil, nil)

	txHash, err := chainImpl.SendTransaction(ctx, chain.SendRequest{
		FromAddress: task.FromAddress,
		ToAddress:   task.ToAddress,
		Amount:      task.Amount,
		Token:       task.Token,
		PrivateKey:  privKey,
	})
	if err != nil {
		errMsg := "sweep tx failed: " + err.Error()
		s.store.UpdateSweepTaskStatus(ctx, task.ID, models.SweepStatusFailed, nil, &errMsg)
		s.sendSweepWebhook(ctx, task, models.SweepStatusFailed, "", errMsg)
		return
	}

	s.store.UpdateSweepTaskStatus(ctx, task.ID, models.SweepStatusCompleted, &txHash, nil)
	s.sendSweepWebhook(ctx, task, models.SweepStatusCompleted, txHash, "")
	slog.Info("sweep completed", "task_id", task.ID, "tx_hash", txHash)
}

func (s *Service) checkGasDeposit(ctx context.Context, task *models.SweepTask) {
	if task.GasDepositID == nil {
		return
	}
	deposit, err := s.store.GetGasDepositBySweepTask(ctx, task.ID)
	if err != nil {
		return
	}
	if deposit.Status == models.GasDepositStatusCompleted {
		// Gas deposited, retry sweep
		s.store.UpdateSweepTaskStatus(ctx, task.ID, models.SweepStatusPending, nil, nil)
	}
}

func (s *Service) sendSweepWebhook(ctx context.Context, task *models.SweepTask, status, txHash, errMsg string) {
	merchant, err := s.store.GetMerchantByID(ctx, task.MerchantID)
	if err != nil || merchant.WebhookURL == "" {
		return
	}
	var eventType webhook.EventType
	if status == models.SweepStatusCompleted {
		eventType = webhook.EventSweepCompleted
	} else {
		eventType = webhook.EventSweepFailed
	}
	paymentID := ""
	if task.PaymentID != nil {
		paymentID = *task.PaymentID
	}
	data := webhook.SweepEventData{
		SweepTaskID: task.ID,
		PaymentID:   paymentID,
		Chain:       task.Chain,
		FromAddress: task.FromAddress,
		ToAddress:   task.ToAddress,
		Amount:      task.Amount,
		TxHash:      txHash,
		Status:      status,
		Error:       errMsg,
	}
	go s.webhook.Send(ctx, merchant.WebhookURL, merchant.APIKeyHash, eventType, data)
}

func merchantIDToIndex(id string) uint32 {
	var sum uint32
	for _, b := range []byte(id) {
		sum = sum*31 + uint32(b)
	}
	return sum & 0x7FFFFFFF
}
