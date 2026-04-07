package coldwallet

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
	"github.com/octopuswallet/octopuswallet/pkg/crypto"
)

type Service struct {
	store    store.Store
	registry *chain.Registry
	webhook  *webhook.Service
	seed     []byte
	configs  map[string]config.ChainConfig
}

func NewService(s store.Store, registry *chain.Registry, wh *webhook.Service, seed []byte, configs map[string]config.ChainConfig) *Service {
	return &Service{store: s, registry: registry, webhook: wh, seed: seed, configs: configs}
}

func (s *Service) Start(ctx context.Context) {
	slog.Info("cold wallet service started")

	checkTicker := time.NewTicker(60 * time.Second)
	transferTicker := time.NewTicker(15 * time.Second)
	defer checkTicker.Stop()
	defer transferTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("cold wallet service stopped")
			return
		case <-checkTicker.C:
			s.checkHotWalletBalances(ctx)
		case <-transferTicker.C:
			s.processPendingTransfers(ctx)
		}
	}
}

func (s *Service) checkHotWalletBalances(ctx context.Context) {
	configs, err := s.store.GetAllEnabledColdWalletConfigs(ctx)
	if err != nil {
		return
	}

	for _, cfg := range configs {
		s.checkBalance(ctx, &cfg)
	}
}

func (s *Service) checkBalance(ctx context.Context, cfg *models.ColdWalletConfig) {
	if cfg.ColdWalletAddress == "" || cfg.HotWalletMaxBalance == "0" {
		return
	}

	chainImpl, err := s.registry.Get(cfg.Chain)
	if err != nil {
		return
	}

	// Get merchant's collection address (hot wallet) or first wallet
	collAddr, err := s.store.GetMerchantCollectionAddress(ctx, cfg.MerchantID, cfg.Chain)
	if err != nil {
		return
	}

	balance, err := chainImpl.GetBalance(ctx, collAddr.Address, "")
	if err != nil {
		return
	}

	balanceBig := new(big.Int)
	balanceBig.SetString(balance, 10)
	maxBalance := new(big.Int)
	maxBalance.SetString(cfg.HotWalletMaxBalance, 10)

	if balanceBig.Cmp(maxBalance) <= 0 {
		return
	}

	// Transfer excess to cold wallet
	excess := new(big.Int).Sub(balanceBig, maxBalance)

	transfer := &models.WalletTransfer{
		MerchantID:   cfg.MerchantID,
		Chain:        cfg.Chain,
		Token:        "",
		FromAddress:  collAddr.Address,
		ToAddress:    cfg.ColdWalletAddress,
		Amount:       excess.String(),
		TransferType: models.TransferTypeHotToCold,
	}
	if err := s.store.CreateWalletTransfer(ctx, transfer); err != nil {
		slog.Error("failed to create cold transfer", "merchant_id", cfg.MerchantID, "error", err)
		return
	}
	slog.Info("cold transfer created", "merchant_id", cfg.MerchantID, "amount", excess.String())
}

func (s *Service) processPendingTransfers(ctx context.Context) {
	transfers, err := s.store.GetPendingWalletTransfers(ctx)
	if err != nil {
		return
	}

	for _, transfer := range transfers {
		if transfer.TransferType == models.TransferTypeHotToCold {
			s.processHotToCold(ctx, &transfer)
		}
		// cold_to_hot requires manual cold wallet signing — skip auto-processing
	}
}

func (s *Service) processHotToCold(ctx context.Context, transfer *models.WalletTransfer) {
	chainImpl, err := s.registry.Get(transfer.Chain)
	if err != nil {
		return
	}

	// Find wallet for from_address to get derivation index
	wallet, err := s.store.GetWalletByAddress(ctx, transfer.Chain, transfer.FromAddress)
	if err != nil {
		errMsg := "wallet not found for address"
		s.store.UpdateWalletTransferStatus(ctx, transfer.ID, models.PayoutStatusFailed, nil, &errMsg)
		return
	}

	merchantIndex := crypto.MerchantIDToIndex(transfer.MerchantID)
	privKey, err := chainImpl.DerivePrivateKey(s.seed, merchantIndex, uint32(wallet.DerivationIndex))
	if err != nil {
		errMsg := "failed to derive key: " + err.Error()
		s.store.UpdateWalletTransferStatus(ctx, transfer.ID, models.PayoutStatusFailed, nil, &errMsg)
		return
	}
	defer crypto.ZeroBytes(privKey)

	s.store.UpdateWalletTransferStatus(ctx, transfer.ID, "processing", nil, nil)

	txHash, err := chainImpl.SendTransaction(ctx, chain.SendRequest{
		FromAddress: transfer.FromAddress,
		ToAddress:   transfer.ToAddress,
		Amount:      transfer.Amount,
		Token:       transfer.Token,
		PrivateKey:  privKey,
	})
	if err != nil {
		errMsg := "transfer tx failed: " + err.Error()
		s.store.UpdateWalletTransferStatus(ctx, transfer.ID, models.PayoutStatusFailed, nil, &errMsg)
		s.sendTransferWebhook(ctx, transfer, models.PayoutStatusFailed, "", errMsg)
		return
	}

	s.store.UpdateWalletTransferStatus(ctx, transfer.ID, models.PayoutStatusCompleted, &txHash, nil)
	s.sendTransferWebhook(ctx, transfer, models.PayoutStatusCompleted, txHash, "")
	slog.Info("cold transfer completed", "transfer_id", transfer.ID, "tx_hash", txHash)
}

func (s *Service) sendTransferWebhook(ctx context.Context, transfer *models.WalletTransfer, status, txHash, errMsg string) {
	merchant, err := s.store.GetMerchantByID(ctx, transfer.MerchantID)
	if err != nil || merchant.WebhookURL == "" {
		return
	}
	var eventType webhook.EventType
	if status == models.PayoutStatusCompleted {
		eventType = webhook.EventTransferCompleted
	} else {
		eventType = webhook.EventTransferFailed
	}
	data := webhook.TransferEventData{
		TransferID:   transfer.ID,
		Chain:        transfer.Chain,
		TransferType: transfer.TransferType,
		FromAddress:  transfer.FromAddress,
		ToAddress:    transfer.ToAddress,
		Amount:       transfer.Amount,
		TxHash:       txHash,
		Status:       status,
		Error:        errMsg,
	}
	go s.webhook.Send(ctx, merchant.WebhookURL, merchant.APIKeyHash, eventType, data)
}

