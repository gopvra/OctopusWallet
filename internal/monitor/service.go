package monitor

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
)

// OnPaymentCompletedFunc is called when a payment transitions to completed.
type OnPaymentCompletedFunc func(ctx context.Context, payment *models.Payment)

type Service struct {
	store              store.Store
	registry           *chain.Registry
	webhook            *webhook.Service
	configs            map[string]config.ChainConfig
	onPaymentCompleted OnPaymentCompletedFunc
}

func NewService(s store.Store, registry *chain.Registry, wh *webhook.Service, configs map[string]config.ChainConfig) *Service {
	return &Service{
		store:    s,
		registry: registry,
		webhook:  wh,
		configs:  configs,
	}
}

// SetOnPaymentCompleted registers a callback for payment completion (used by sweep service).
func (s *Service) SetOnPaymentCompleted(fn OnPaymentCompletedFunc) {
	s.onPaymentCompleted = fn
}

// Start launches a monitoring goroutine for each registered chain.
func (s *Service) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for _, c := range s.registry.All() {
		chainCfg, ok := s.configs[c.Name()]
		if !ok || !chainCfg.Enabled {
			continue
		}

		wg.Add(1)
		go func(c chain.Chain, cfg config.ChainConfig) {
			defer wg.Done()
			s.monitorChain(ctx, c, cfg)
		}(c, chainCfg)
	}

	// Payment expiry checker
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.expirePayments(ctx)
	}()

	wg.Wait()
}

func (s *Service) monitorChain(ctx context.Context, c chain.Chain, cfg config.ChainConfig) {
	slog.Info("starting chain monitor", "chain", c.Name(), "poll_interval", cfg.PollInterval)

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping chain monitor", "chain", c.Name())
			return
		case <-ticker.C:
			if err := s.scanChain(ctx, c, cfg); err != nil {
				slog.Error("chain scan error", "chain", c.Name(), "error", err)
			}
		}
	}
}

func (s *Service) scanChain(ctx context.Context, c chain.Chain, cfg config.ChainConfig) error {
	// Get current block height
	currentHeight, err := c.GetCurrentBlockHeight(ctx)
	if err != nil {
		return err
	}

	// Get last scanned block
	lastScanned, err := s.store.GetLastScannedBlock(ctx, c.Name())
	if err != nil {
		return err
	}

	if lastScanned == 0 {
		// Start from current block minus a small buffer
		lastScanned = currentHeight - 1
		if err := s.store.SetLastScannedBlock(ctx, c.Name(), lastScanned); err != nil {
			return err
		}
	}

	// Load watch addresses for this chain
	allAddresses, err := s.store.GetAllWatchAddresses(ctx)
	if err != nil {
		return err
	}
	watchAddresses := allAddresses[c.Name()]
	if len(watchAddresses) == 0 {
		// No addresses to watch, just update the last scanned block
		if currentHeight > lastScanned {
			return s.store.SetLastScannedBlock(ctx, c.Name(), currentHeight)
		}
		return nil
	}

	// Scan each block from last scanned + 1 to current
	for height := lastScanned + 1; height <= currentHeight; height++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		txs, err := c.ScanBlockForPayments(ctx, height, watchAddresses)
		if err != nil {
			slog.Error("block scan error", "chain", c.Name(), "block", height, "error", err)
			break
		}

		for _, tx := range txs {
			s.processIncomingTx(ctx, c, cfg, tx)
		}

		if err := s.store.SetLastScannedBlock(ctx, c.Name(), height); err != nil {
			return err
		}
	}

	// Check confirmations for existing confirming payments
	s.updateConfirmations(ctx, c, cfg)

	return nil
}

func (s *Service) processIncomingTx(ctx context.Context, c chain.Chain, cfg config.ChainConfig, tx chain.IncomingTx) {
	payment, err := s.store.GetPaymentByAddress(ctx, c.Name(), tx.ToAddress)
	if err != nil {
		return // no matching payment
	}

	if payment.Status != models.PaymentStatusPending {
		return
	}

	txHash := tx.TxHash
	status := models.PaymentStatusConfirming

	// Check if already has enough confirmations
	confirmations, err := c.GetTransactionConfirmations(ctx, tx.TxHash)
	if err == nil && confirmations >= cfg.ConfirmationsRequired {
		status = models.PaymentStatusCompleted
	}

	if err := s.store.UpdatePaymentStatus(ctx, payment.ID, status, &txHash, int(confirmations)); err != nil {
		slog.Error("failed to update payment", "payment_id", payment.ID, "error", err)
		return
	}

	// Send webhook
	s.sendPaymentWebhook(ctx, payment, status, txHash, int(confirmations))

	// On completion: update merchant balance + trigger auto-sweep
	if status == models.PaymentStatusCompleted {
		payment.AmountReceived = tx.Amount
		s.store.UpdateMerchantBalance(ctx, payment.MerchantID, payment.Chain, payment.Token, tx.Amount, "0")
		if s.onPaymentCompleted != nil {
			s.onPaymentCompleted(ctx, payment)
		}
	}
}

func (s *Service) updateConfirmations(ctx context.Context, c chain.Chain, cfg config.ChainConfig) {
	payments, err := s.store.GetPendingPayments(ctx)
	if err != nil {
		return
	}

	for _, payment := range payments {
		if payment.Chain != c.Name() || payment.Status != models.PaymentStatusConfirming || payment.TxHash == nil {
			continue
		}

		confirmations, err := c.GetTransactionConfirmations(ctx, *payment.TxHash)
		if err != nil {
			continue
		}

		if confirmations >= cfg.ConfirmationsRequired {
			if err := s.store.UpdatePaymentStatus(ctx, payment.ID, models.PaymentStatusCompleted, payment.TxHash, int(confirmations)); err != nil {
				slog.Error("failed to complete payment", "payment_id", payment.ID, "error", err)
				continue
			}
			s.sendPaymentWebhook(ctx, &payment, models.PaymentStatusCompleted, *payment.TxHash, int(confirmations))
			s.store.UpdateMerchantBalance(ctx, payment.MerchantID, payment.Chain, payment.Token, payment.AmountReceived, "0")
			if s.onPaymentCompleted != nil {
				s.onPaymentCompleted(ctx, &payment)
			}
		} else {
			// Update confirmation count
			if err := s.store.UpdatePaymentStatus(ctx, payment.ID, models.PaymentStatusConfirming, payment.TxHash, int(confirmations)); err != nil {
				slog.Error("failed to update confirmations", "payment_id", payment.ID, "error", err)
			}
		}
	}
}

func (s *Service) sendPaymentWebhook(ctx context.Context, payment *models.Payment, status, txHash string, confirmations int) {
	merchant, err := s.store.GetMerchantByID(ctx, payment.MerchantID)
	if err != nil || merchant.WebhookURL == "" {
		return
	}

	var eventType webhook.EventType
	switch status {
	case models.PaymentStatusConfirming:
		eventType = webhook.EventPaymentConfirming
	case models.PaymentStatusCompleted:
		eventType = webhook.EventPaymentCompleted
	case models.PaymentStatusExpired:
		eventType = webhook.EventPaymentExpired
	default:
		return
	}

	data := webhook.PaymentEventData{
		PaymentID:      payment.ID,
		Chain:          payment.Chain,
		Address:        payment.Address,
		AmountExpected: payment.AmountExpected,
		AmountReceived: payment.AmountReceived,
		TxHash:         txHash,
		Confirmations:  confirmations,
		Status:         status,
	}

	go func() {
		if err := s.webhook.Send(ctx, merchant.WebhookURL, merchant.APIKeyHash, eventType, data); err != nil {
			slog.Error("webhook delivery failed", "merchant_id", merchant.ID, "event", eventType, "error", err)
		}
	}()
}

func (s *Service) expirePayments(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			payments, err := s.store.GetPendingPayments(ctx)
			if err != nil {
				continue
			}
			now := time.Now()
			for _, p := range payments {
				if p.Status == models.PaymentStatusPending && p.ExpiresAt != nil && now.After(*p.ExpiresAt) {
					if err := s.store.UpdatePaymentStatus(ctx, p.ID, models.PaymentStatusExpired, nil, 0); err != nil {
						slog.Error("failed to expire payment", "payment_id", p.ID, "error", err)
						continue
					}
					s.sendPaymentWebhook(ctx, &p, models.PaymentStatusExpired, "", 0)
				}
			}
		}
	}
}
