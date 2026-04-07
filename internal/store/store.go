package store

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

type Store interface {
	// Merchants
	CreateMerchant(ctx context.Context, merchant *models.Merchant) error
	GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error)
	GetMerchantByAPIKeyHash(ctx context.Context, hash string) (*models.Merchant, error)

	// Wallets
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
	GetWalletsByMerchant(ctx context.Context, merchantID string) ([]models.Wallet, error)
	GetWalletByAddress(ctx context.Context, chain, address string) (*models.Wallet, error)
	GetAllWatchAddresses(ctx context.Context) (map[string]map[string]struct{}, error) // chain -> set of addresses
	GetNextDerivationIndex(ctx context.Context, merchantID, chain string) (int, error)

	// Payments
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPaymentByID(ctx context.Context, id string) (*models.Payment, error)
	GetPendingPayments(ctx context.Context) ([]models.Payment, error)
	GetPaymentByAddress(ctx context.Context, chain, address string) (*models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, id, status string, txHash *string, confirmations int) error

	// Payouts
	CreatePayout(ctx context.Context, payout *models.Payout) error
	GetPayoutByID(ctx context.Context, id string) (*models.Payout, error)
	GetPendingPayouts(ctx context.Context) ([]models.Payout, error)
	UpdatePayoutStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error

	// Chain state
	GetLastScannedBlock(ctx context.Context, chain string) (uint64, error)
	SetLastScannedBlock(ctx context.Context, chain string, height uint64) error

	// Lifecycle
	Close() error
}
