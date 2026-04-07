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

	// Sweep
	UpsertMerchantCollectionAddress(ctx context.Context, addr *models.MerchantCollectionAddress) error
	GetMerchantCollectionAddress(ctx context.Context, merchantID, chain string) (*models.MerchantCollectionAddress, error)
	GetMerchantCollectionAddresses(ctx context.Context, merchantID string) ([]models.MerchantCollectionAddress, error)
	CreateSweepTask(ctx context.Context, task *models.SweepTask) error
	GetPendingSweepTasks(ctx context.Context) ([]models.SweepTask, error)
	GetSweepTasksByMerchant(ctx context.Context, merchantID string) ([]models.SweepTask, error)
	UpdateSweepTaskStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error
	UpdateSweepTaskGasDeposit(ctx context.Context, id, gasDepositID string) error

	// Cold/Hot Wallet
	UpsertColdWalletConfig(ctx context.Context, cfg *models.ColdWalletConfig) error
	GetColdWalletConfig(ctx context.Context, merchantID, chain string) (*models.ColdWalletConfig, error)
	GetAllEnabledColdWalletConfigs(ctx context.Context) ([]models.ColdWalletConfig, error)
	GetColdWalletConfigsByMerchant(ctx context.Context, merchantID string) ([]models.ColdWalletConfig, error)
	CreateWalletTransfer(ctx context.Context, transfer *models.WalletTransfer) error
	GetPendingWalletTransfers(ctx context.Context) ([]models.WalletTransfer, error)
	GetWalletTransfersByMerchant(ctx context.Context, merchantID string) ([]models.WalletTransfer, error)
	UpdateWalletTransferStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error

	// Approval
	UpsertApprovalConfig(ctx context.Context, cfg *models.ApprovalConfig) error
	GetApprovalConfig(ctx context.Context, merchantID string) (*models.ApprovalConfig, error)
	CreatePayoutApproval(ctx context.Context, approval *models.PayoutApproval) error
	UpdatePayoutApprovalStatus(ctx context.Context, payoutID, approvalStatus string) error
	GetDailyPayoutTotal(ctx context.Context, merchantID, chain string) (string, error)
	IncrementDailyPayoutTotal(ctx context.Context, merchantID, chain, amount string) error

	// Gas
	CreateGasDeposit(ctx context.Context, deposit *models.GasDeposit) error
	GetPendingGasDeposits(ctx context.Context) ([]models.GasDeposit, error)
	UpdateGasDepositStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error
	GetGasDepositBySweepTask(ctx context.Context, sweepTaskID string) (*models.GasDeposit, error)
	CreateGasAlert(ctx context.Context, alert *models.GasAlert) error

	// Refunds
	CreateRefund(ctx context.Context, refund *models.Refund) error
	GetRefundByID(ctx context.Context, id string) (*models.Refund, error)
	GetRefundsByPayment(ctx context.Context, paymentID string) ([]models.Refund, error)
	GetRefundTotalByPayment(ctx context.Context, paymentID string) (string, error)
	GetPendingRefunds(ctx context.Context) ([]models.Refund, error)
	UpdateRefundStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error

	// Merchant Balances
	GetMerchantBalances(ctx context.Context, merchantID string) ([]models.MerchantBalance, error)
	UpdateMerchantBalance(ctx context.Context, merchantID, chain, token, deltaAvailable, deltaPending string) error

	// Supported Currencies
	GetSupportedCurrencies(ctx context.Context) ([]models.SupportedCurrency, error)
	GetSupportedCurrenciesByChain(ctx context.Context, chain string) ([]models.SupportedCurrency, error)

	// Batch Payouts
	CreateBatchPayout(ctx context.Context, batch *models.BatchPayout) error
	GetBatchPayoutByID(ctx context.Context, id string) (*models.BatchPayout, error)
	GetBatchPayoutsByMerchant(ctx context.Context, merchantID string) ([]models.BatchPayout, error)
	CreateBatchPayoutItem(ctx context.Context, item *models.BatchPayoutItem) error
	GetBatchPayoutItems(ctx context.Context, batchID string) ([]models.BatchPayoutItem, error)
	UpdateBatchPayoutStatus(ctx context.Context, id, status string, completed, failed int) error

	// IP Whitelist
	SetMerchantIPWhitelist(ctx context.Context, merchantID string, ips []string) error
	GetMerchantIPWhitelist(ctx context.Context, merchantID string) ([]string, error)

	// Pagination helpers
	GetPaymentsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]models.Payment, error)
	GetPayoutsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]models.Payout, error)

	// Payment Links
	CreatePaymentLink(ctx context.Context, link *models.PaymentLink) error
	GetPaymentLinkByID(ctx context.Context, id string) (*models.PaymentLink, error)
	GetPaymentLinksByMerchant(ctx context.Context, merchantID string) ([]models.PaymentLink, error)
	IncrementPaymentLinkUses(ctx context.Context, id string) error

	// Audit Logs
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	GetAuditLogs(ctx context.Context, merchantID string, limit, offset int) ([]models.AuditLog, error)

	// Export
	GetPaymentsByMerchantDateRange(ctx context.Context, merchantID, from, to string) ([]models.Payment, error)
	GetPayoutsByMerchantDateRange(ctx context.Context, merchantID, from, to string) ([]models.Payout, error)

	// Lifecycle
	Close() error
}
