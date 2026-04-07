package refund

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
	return &Service{store: s, registry: registry, webhook: wh, seed: seed, configs: configs}
}

func (s *Service) Start(ctx context.Context) {
	slog.Info("refund service started")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("refund service stopped")
			return
		case <-ticker.C:
			s.processPendingRefunds(ctx)
		}
	}
}

func (s *Service) processPendingRefunds(ctx context.Context) {
	refunds, err := s.store.GetPendingRefunds(ctx)
	if err != nil {
		slog.Error("failed to get pending refunds", "error", err)
		return
	}
	for _, r := range refunds {
		s.processRefund(ctx, &r)
	}
}

func (s *Service) processRefund(ctx context.Context, r *models.Refund) {
	chainImpl, err := s.registry.Get(r.Chain)
	if err != nil {
		errMsg := "unsupported chain"
		s.store.UpdateRefundStatus(ctx, r.ID, models.RefundStatusFailed, nil, &errMsg)
		return
	}

	// Find wallet to derive private key
	wallet, err := s.store.GetWalletByAddress(ctx, r.Chain, r.ToAddress)
	if wallet == nil {
		// Refund goes to an external address — use the payment's receive address as source
		payment, err := s.store.GetPaymentByID(ctx, r.PaymentID)
		if err != nil {
			errMsg := "payment not found"
			s.store.UpdateRefundStatus(ctx, r.ID, models.RefundStatusFailed, nil, &errMsg)
			return
		}
		wallet, err = s.store.GetWalletByAddress(ctx, r.Chain, payment.Address)
		if err != nil {
			errMsg := "source wallet not found"
			s.store.UpdateRefundStatus(ctx, r.ID, models.RefundStatusFailed, nil, &errMsg)
			return
		}
	}

	merchantIndex := crypto.MerchantIDToIndex(r.MerchantID)
	privKey, err := chainImpl.DerivePrivateKey(s.seed, merchantIndex, uint32(wallet.DerivationIndex))
	if err != nil {
		errMsg := "failed to derive key: " + err.Error()
		s.store.UpdateRefundStatus(ctx, r.ID, models.RefundStatusFailed, nil, &errMsg)
		return
	}
	defer crypto.ZeroBytes(privKey)

	txHash, err := chainImpl.SendTransaction(ctx, chain.SendRequest{
		FromAddress: wallet.Address,
		ToAddress:   r.ToAddress,
		Amount:      r.Amount,
		Token:       r.Token,
		PrivateKey:  privKey,
	})
	if err != nil {
		errMsg := "refund tx failed: " + err.Error()
		s.store.UpdateRefundStatus(ctx, r.ID, models.RefundStatusFailed, nil, &errMsg)
		s.sendRefundWebhook(ctx, r, models.RefundStatusFailed, "", errMsg)
		return
	}

	s.store.UpdateRefundStatus(ctx, r.ID, models.RefundStatusCompleted, &txHash, nil)
	// Deduct from merchant balance
	s.store.UpdateMerchantBalance(ctx, r.MerchantID, r.Chain, r.Token, "-"+r.Amount, "0")
	s.sendRefundWebhook(ctx, r, models.RefundStatusCompleted, txHash, "")
	slog.Info("refund completed", "refund_id", r.ID, "tx_hash", txHash)
}

func (s *Service) sendRefundWebhook(ctx context.Context, r *models.Refund, status, txHash, errMsg string) {
	merchant, err := s.store.GetMerchantByID(ctx, r.MerchantID)
	if err != nil || merchant.WebhookURL == "" {
		return
	}
	var eventType webhook.EventType
	if status == models.RefundStatusCompleted {
		eventType = webhook.EventRefundCompleted
	} else {
		eventType = webhook.EventRefundFailed
	}
	data := webhook.PayoutEventData{
		PayoutID:  r.ID,
		Chain:     r.Chain,
		ToAddress: r.ToAddress,
		Amount:    r.Amount,
		TxHash:    txHash,
		Status:    status,
		Error:     errMsg,
	}
	go s.webhook.Send(ctx, merchant.WebhookURL, merchant.APIKeyHash, eventType, data)
}
