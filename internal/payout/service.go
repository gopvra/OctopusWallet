package payout

import (
	"context"
	"log/slog"
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
	return &Service{
		store:    s,
		registry: registry,
		webhook:  wh,
		seed:     seed,
		configs:  configs,
	}
}

// Start processes pending payouts in a loop.
func (s *Service) Start(ctx context.Context) {
	slog.Info("payout service started")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("payout service stopped")
			return
		case <-ticker.C:
			s.processPendingPayouts(ctx)
		}
	}
}

func (s *Service) processPendingPayouts(ctx context.Context) {
	payouts, err := s.store.GetPendingPayouts(ctx)
	if err != nil {
		slog.Error("failed to get pending payouts", "error", err)
		return
	}

	for _, payout := range payouts {
		s.processPayout(ctx, &payout)
	}
}

func (s *Service) processPayout(ctx context.Context, payout *models.Payout) {
	chainImpl, err := s.registry.Get(payout.Chain)
	if err != nil {
		errMsg := "unsupported chain: " + payout.Chain
		s.store.UpdatePayoutStatus(ctx, payout.ID, models.PayoutStatusFailed, nil, &errMsg)
		return
	}

	// Find the merchant's wallet on this chain to get the from address and private key
	wallets, err := s.store.GetWalletsByMerchant(ctx, payout.MerchantID)
	if err != nil || len(wallets) == 0 {
		errMsg := "no wallet found for merchant on chain " + payout.Chain
		s.store.UpdatePayoutStatus(ctx, payout.ID, models.PayoutStatusFailed, nil, &errMsg)
		return
	}

	// Use the first wallet for this chain as the source
	var sourceWallet *models.Wallet
	for _, w := range wallets {
		if w.Chain == payout.Chain {
			sourceWallet = &w
			break
		}
	}
	if sourceWallet == nil {
		errMsg := "no wallet found for chain " + payout.Chain
		s.store.UpdatePayoutStatus(ctx, payout.ID, models.PayoutStatusFailed, nil, &errMsg)
		return
	}

	// Derive private key
	merchantIndex := crypto.MerchantIDToIndex(payout.MerchantID)
	privKey, err := chainImpl.DerivePrivateKey(s.seed, merchantIndex, uint32(sourceWallet.DerivationIndex))
	if err != nil {
		errMsg := "failed to derive private key: " + err.Error()
		s.store.UpdatePayoutStatus(ctx, payout.ID, models.PayoutStatusFailed, nil, &errMsg)
		return
	}
	defer crypto.ZeroBytes(privKey) // zero private key after use

	// Send transaction
	txHash, err := chainImpl.SendTransaction(ctx, chain.SendRequest{
		FromAddress: sourceWallet.Address,
		ToAddress:   payout.ToAddress,
		Amount:      payout.Amount,
		Token:       payout.Token,
		PrivateKey:  privKey,
	})
	if err != nil {
		errMsg := "transaction failed: " + err.Error()
		s.store.UpdatePayoutStatus(ctx, payout.ID, models.PayoutStatusFailed, nil, &errMsg)
		s.sendPayoutWebhook(ctx, payout, models.PayoutStatusFailed, "", errMsg)
		return
	}

	s.store.UpdatePayoutStatus(ctx, payout.ID, models.PayoutStatusCompleted, &txHash, nil)
	s.store.UpdateMerchantBalance(ctx, payout.MerchantID, payout.Chain, payout.Token, "-"+payout.Amount, "0")
	s.sendPayoutWebhook(ctx, payout, models.PayoutStatusCompleted, txHash, "")
	slog.Info("payout completed", "payout_id", payout.ID, "tx_hash", txHash)
}

func (s *Service) sendPayoutWebhook(ctx context.Context, payout *models.Payout, status, txHash, errMsg string) {
	merchant, err := s.store.GetMerchantByID(ctx, payout.MerchantID)
	if err != nil || merchant.WebhookURL == "" {
		return
	}

	var eventType webhook.EventType
	if status == models.PayoutStatusCompleted {
		eventType = webhook.EventPayoutCompleted
	} else {
		eventType = webhook.EventPayoutFailed
	}

	data := webhook.PayoutEventData{
		PayoutID:  payout.ID,
		Chain:     payout.Chain,
		ToAddress: payout.ToAddress,
		Amount:    payout.Amount,
		TxHash:    txHash,
		Status:    status,
		Error:     errMsg,
	}

	go func() {
		if err := s.webhook.Send(ctx, merchant.WebhookURL, merchant.APIKeyHash, eventType, data); err != nil {
			slog.Error("payout webhook failed", "payout_id", payout.ID, "error", err)
		}
	}()
}

