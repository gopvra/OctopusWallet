package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

type Store struct {
	db *gorm.DB
}

func New(databaseURL string, maxOpenConns int) (*Store, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying db: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DB returns the underlying gorm.DB for admin store operations.
func (s *Store) DB() *gorm.DB {
	return s.db
}

// --- Merchants ---

func (s *Store) CreateMerchant(ctx context.Context, m *models.Merchant) error {
	return s.db.WithContext(ctx).Create(m).Error
}

func (s *Store) GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	var m models.Merchant
	err := s.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) GetMerchantByAPIKeyHash(ctx context.Context, hash string) (*models.Merchant, error) {
	var m models.Merchant
	err := s.db.WithContext(ctx).Where("api_key_hash = ? AND is_active = true", hash).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// --- Wallets ---

func (s *Store) CreateWallet(ctx context.Context, w *models.Wallet) error {
	return s.db.WithContext(ctx).Create(w).Error
}

func (s *Store) GetWalletsByMerchant(ctx context.Context, merchantID string) ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("chain, derivation_index").Find(&wallets).Error
	return wallets, err
}

func (s *Store) GetWalletByAddress(ctx context.Context, chain, address string) (*models.Wallet, error) {
	var w models.Wallet
	err := s.db.WithContext(ctx).Where("chain = ? AND address = ?", chain, address).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) GetAllWatchAddresses(ctx context.Context) (map[string]map[string]struct{}, error) {
	var results []struct {
		Chain   string
		Address string
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT w.chain, w.address FROM wallets w
		INNER JOIN payments p ON p.address = w.address AND p.chain = w.chain
		WHERE p.status IN ('pending', 'confirming')
	`).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[string]struct{})
	for _, r := range results {
		if out[r.Chain] == nil {
			out[r.Chain] = make(map[string]struct{})
		}
		out[r.Chain][r.Address] = struct{}{}
	}
	return out, nil
}

func (s *Store) GetNextDerivationIndex(ctx context.Context, merchantID, chain string) (int, error) {
	var maxIdx sql.NullInt64
	err := s.db.WithContext(ctx).Model(&models.Wallet{}).
		Where("merchant_id = ? AND chain = ?", merchantID, chain).
		Select("MAX(derivation_index)").Scan(&maxIdx).Error
	if err != nil || !maxIdx.Valid {
		return 0, err
	}
	return int(maxIdx.Int64) + 1, nil
}

// --- Payments ---

func (s *Store) CreatePayment(ctx context.Context, p *models.Payment) error {
	return s.db.WithContext(ctx).Create(p).Error
}

func (s *Store) GetPaymentByID(ctx context.Context, id string) (*models.Payment, error) {
	var p models.Payment
	err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetPendingPayments(ctx context.Context) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.WithContext(ctx).Where("status IN ?", []string{"pending", "confirming"}).Order("created_at").Find(&payments).Error
	return payments, err
}

func (s *Store) GetPaymentByAddress(ctx context.Context, chain, address string) (*models.Payment, error) {
	var p models.Payment
	err := s.db.WithContext(ctx).Where("chain = ? AND address = ? AND status IN ?", chain, address, []string{"pending", "confirming"}).
		Order("created_at DESC").First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) UpdatePaymentStatus(ctx context.Context, id, status string, txHash *string, confirmations int) error {
	updates := map[string]interface{}{
		"status":        status,
		"confirmations": confirmations,
		"updated_at":    time.Now(),
	}
	if txHash != nil {
		updates["tx_hash"] = *txHash
	}
	return s.db.WithContext(ctx).Model(&models.Payment{}).
		Where("id = ? AND status NOT IN ?", id, []string{"completed", "expired"}).
		Updates(updates).Error
}

// --- Payouts ---

func (s *Store) CreatePayout(ctx context.Context, p *models.Payout) error {
	return s.db.WithContext(ctx).Create(p).Error
}

func (s *Store) GetPayoutByID(ctx context.Context, id string) (*models.Payout, error) {
	var p models.Payout
	err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetPendingPayouts(ctx context.Context) ([]models.Payout, error) {
	var payouts []models.Payout
	err := s.db.WithContext(ctx).Raw(`
		UPDATE payouts SET status = 'processing', updated_at = now()
		WHERE id IN (
			SELECT id FROM payouts
			WHERE status = 'pending' AND (approval_status = '' OR approval_status = 'approved')
			ORDER BY created_at LIMIT 10
			FOR UPDATE SKIP LOCKED
		) RETURNING *
	`).Scan(&payouts).Error
	return payouts, err
}

func (s *Store) UpdatePayoutStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	return s.db.WithContext(ctx).Model(&models.Payout{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        status,
			"tx_hash":       txHash,
			"error_message": errMsg,
			"updated_at":    time.Now(),
		}).Error
}

// --- Chain State ---

func (s *Store) GetLastScannedBlock(ctx context.Context, chain string) (uint64, error) {
	var cs models.ChainState
	err := s.db.WithContext(ctx).First(&cs, "chain = ?", chain).Error
	if err != nil {
		return 0, nil // default to 0 if not found
	}
	return cs.LastScannedBlock, nil
}

func (s *Store) SetLastScannedBlock(ctx context.Context, chain string, height uint64) error {
	cs := models.ChainState{Chain: chain, LastScannedBlock: height}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chain"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_scanned_block", "updated_at"}),
	}).Create(&cs).Error
}

// --- Sweep ---

func (s *Store) UpsertMerchantCollectionAddress(ctx context.Context, addr *models.MerchantCollectionAddress) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "merchant_id"}, {Name: "chain"}},
		DoUpdates: clause.AssignmentColumns([]string{"address", "sweep_threshold", "is_active", "updated_at"}),
	}).Create(addr).Error
}

func (s *Store) GetMerchantCollectionAddress(ctx context.Context, merchantID, chain string) (*models.MerchantCollectionAddress, error) {
	var addr models.MerchantCollectionAddress
	err := s.db.WithContext(ctx).Where("merchant_id = ? AND chain = ? AND is_active = true", merchantID, chain).First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (s *Store) GetMerchantCollectionAddresses(ctx context.Context, merchantID string) ([]models.MerchantCollectionAddress, error) {
	var addrs []models.MerchantCollectionAddress
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&addrs).Error
	return addrs, err
}

func (s *Store) CreateSweepTask(ctx context.Context, task *models.SweepTask) error {
	return s.db.WithContext(ctx).Create(task).Error
}

func (s *Store) GetPendingSweepTasks(ctx context.Context) ([]models.SweepTask, error) {
	var tasks []models.SweepTask
	// Atomic claim: pending tasks
	err := s.db.WithContext(ctx).Raw(`
		UPDATE sweep_tasks SET status = 'processing', updated_at = now()
		WHERE id IN (
			SELECT id FROM sweep_tasks WHERE status = 'pending'
			ORDER BY created_at LIMIT 20
			FOR UPDATE SKIP LOCKED
		) RETURNING *
	`).Scan(&tasks).Error
	if err != nil {
		return nil, err
	}
	// Also fetch gas_needed tasks
	var gasTasks []models.SweepTask
	s.db.WithContext(ctx).Raw(`
		SELECT * FROM sweep_tasks WHERE id IN (
			SELECT id FROM sweep_tasks WHERE status = 'gas_needed'
			ORDER BY created_at LIMIT 20
			FOR UPDATE SKIP LOCKED
		)
	`).Scan(&gasTasks)
	return append(tasks, gasTasks...), nil
}

func (s *Store) GetSweepTasksByMerchant(ctx context.Context, merchantID string) ([]models.SweepTask, error) {
	var tasks []models.SweepTask
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (s *Store) UpdateSweepTaskStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	return s.db.WithContext(ctx).Model(&models.SweepTask{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "tx_hash": txHash, "error_message": errMsg, "updated_at": time.Now()}).Error
}

func (s *Store) UpdateSweepTaskGasDeposit(ctx context.Context, id, gasDepositID string) error {
	return s.db.WithContext(ctx).Model(&models.SweepTask{}).Where("id = ?", id).
		Updates(map[string]interface{}{"gas_deposit_id": gasDepositID, "status": "gas_needed", "updated_at": time.Now()}).Error
}

// --- Cold/Hot Wallet ---

func (s *Store) UpsertColdWalletConfig(ctx context.Context, cfg *models.ColdWalletConfig) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "merchant_id"}, {Name: "chain"}},
		DoUpdates: clause.AssignmentColumns([]string{"cold_wallet_address", "hot_wallet_max_balance", "enabled", "updated_at"}),
	}).Create(cfg).Error
}

func (s *Store) GetColdWalletConfig(ctx context.Context, merchantID, chain string) (*models.ColdWalletConfig, error) {
	var cfg models.ColdWalletConfig
	err := s.db.WithContext(ctx).Where("merchant_id = ? AND chain = ?", merchantID, chain).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) GetAllEnabledColdWalletConfigs(ctx context.Context) ([]models.ColdWalletConfig, error) {
	var configs []models.ColdWalletConfig
	err := s.db.WithContext(ctx).Where("enabled = true").Find(&configs).Error
	return configs, err
}

func (s *Store) GetColdWalletConfigsByMerchant(ctx context.Context, merchantID string) ([]models.ColdWalletConfig, error) {
	var configs []models.ColdWalletConfig
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&configs).Error
	return configs, err
}

func (s *Store) CreateWalletTransfer(ctx context.Context, t *models.WalletTransfer) error {
	return s.db.WithContext(ctx).Create(t).Error
}

func (s *Store) GetPendingWalletTransfers(ctx context.Context) ([]models.WalletTransfer, error) {
	var transfers []models.WalletTransfer
	err := s.db.WithContext(ctx).Raw(`
		UPDATE wallet_transfers SET status = 'processing', updated_at = now()
		WHERE id IN (
			SELECT id FROM wallet_transfers WHERE status = 'pending'
			ORDER BY created_at LIMIT 10
			FOR UPDATE SKIP LOCKED
		) RETURNING *
	`).Scan(&transfers).Error
	return transfers, err
}

func (s *Store) GetWalletTransfersByMerchant(ctx context.Context, merchantID string) ([]models.WalletTransfer, error) {
	var transfers []models.WalletTransfer
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("created_at DESC").Find(&transfers).Error
	return transfers, err
}

func (s *Store) UpdateWalletTransferStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	return s.db.WithContext(ctx).Model(&models.WalletTransfer{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "tx_hash": txHash, "error_message": errMsg, "updated_at": time.Now()}).Error
}

// --- Approval ---

func (s *Store) UpsertApprovalConfig(ctx context.Context, cfg *models.ApprovalConfig) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "merchant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"approval_threshold", "single_tx_limit", "daily_limit", "auto_release", "enabled", "updated_at"}),
	}).Create(cfg).Error
}

func (s *Store) GetApprovalConfig(ctx context.Context, merchantID string) (*models.ApprovalConfig, error) {
	var cfg models.ApprovalConfig
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) CreatePayoutApproval(ctx context.Context, a *models.PayoutApproval) error {
	return s.db.WithContext(ctx).Create(a).Error
}

func (s *Store) UpdatePayoutApprovalStatus(ctx context.Context, payoutID, approvalStatus string) error {
	return s.db.WithContext(ctx).Model(&models.Payout{}).Where("id = ?", payoutID).
		Updates(map[string]interface{}{"approval_status": approvalStatus, "updated_at": time.Now()}).Error
}

func (s *Store) GetDailyPayoutTotal(ctx context.Context, merchantID, chain string) (string, error) {
	var total models.PayoutDailyTotal
	err := s.db.WithContext(ctx).Where("merchant_id = ? AND chain = ? AND date = CURRENT_DATE", merchantID, chain).First(&total).Error
	if err != nil {
		return "0", nil
	}
	return total.TotalAmount, nil
}

func (s *Store) IncrementDailyPayoutTotal(ctx context.Context, merchantID, chain, amount string) error {
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO payout_daily_totals (merchant_id, chain, date, total_amount)
		VALUES (?, ?, CURRENT_DATE, ?)
		ON CONFLICT (merchant_id, chain, date) DO UPDATE
		SET total_amount = (CAST(payout_daily_totals.total_amount AS NUMERIC) + CAST(? AS NUMERIC))::TEXT,
		    updated_at = now()
	`, merchantID, chain, amount, amount).Error
}

// --- Gas ---

func (s *Store) CreateGasDeposit(ctx context.Context, d *models.GasDeposit) error {
	return s.db.WithContext(ctx).Create(d).Error
}

func (s *Store) GetPendingGasDeposits(ctx context.Context) ([]models.GasDeposit, error) {
	var deposits []models.GasDeposit
	err := s.db.WithContext(ctx).Raw(`
		UPDATE gas_deposits SET status = 'processing', updated_at = now()
		WHERE id IN (
			SELECT id FROM gas_deposits WHERE status = 'pending'
			ORDER BY created_at LIMIT 10
			FOR UPDATE SKIP LOCKED
		) RETURNING *
	`).Scan(&deposits).Error
	return deposits, err
}

func (s *Store) UpdateGasDepositStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	return s.db.WithContext(ctx).Model(&models.GasDeposit{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "tx_hash": txHash, "error_message": errMsg, "updated_at": time.Now()}).Error
}

func (s *Store) GetGasDepositBySweepTask(ctx context.Context, sweepTaskID string) (*models.GasDeposit, error) {
	var d models.GasDeposit
	err := s.db.WithContext(ctx).Where("sweep_task_id = ?", sweepTaskID).Order("created_at DESC").First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) CreateGasAlert(ctx context.Context, a *models.GasAlert) error {
	return s.db.WithContext(ctx).Create(a).Error
}

// --- Refunds ---

func (s *Store) CreateRefund(ctx context.Context, r *models.Refund) error {
	return s.db.WithContext(ctx).Create(r).Error
}

func (s *Store) GetRefundByID(ctx context.Context, id string) (*models.Refund, error) {
	var r models.Refund
	err := s.db.WithContext(ctx).First(&r, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) GetRefundsByPayment(ctx context.Context, paymentID string) ([]models.Refund, error) {
	var refunds []models.Refund
	err := s.db.WithContext(ctx).Where("payment_id = ?", paymentID).Order("created_at DESC").Find(&refunds).Error
	return refunds, err
}

func (s *Store) GetRefundTotalByPayment(ctx context.Context, paymentID string) (string, error) {
	var total string
	err := s.db.WithContext(ctx).Raw(
		"SELECT COALESCE(SUM(amount::numeric), 0)::text FROM refunds WHERE payment_id = ? AND status != 'failed'", paymentID,
	).Scan(&total).Error
	return total, err
}

func (s *Store) GetPendingRefunds(ctx context.Context) ([]models.Refund, error) {
	var refunds []models.Refund
	err := s.db.WithContext(ctx).Raw(`
		UPDATE refunds SET status = 'processing', updated_at = now()
		WHERE id IN (
			SELECT id FROM refunds WHERE status = 'pending'
			ORDER BY created_at LIMIT 10
			FOR UPDATE SKIP LOCKED
		) RETURNING *
	`).Scan(&refunds).Error
	return refunds, err
}

func (s *Store) UpdateRefundStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	return s.db.WithContext(ctx).Model(&models.Refund{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "tx_hash": txHash, "error_message": errMsg, "updated_at": time.Now()}).Error
}

// --- Merchant Balances ---

func (s *Store) GetMerchantBalances(ctx context.Context, merchantID string) ([]models.MerchantBalance, error) {
	var balances []models.MerchantBalance
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&balances).Error
	return balances, err
}

func (s *Store) UpdateMerchantBalance(ctx context.Context, merchantID, chain, token, deltaAvailable, deltaPending string) error {
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO merchant_balances (merchant_id, chain, token, available, pending)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (merchant_id, chain, token) DO UPDATE
		SET available = (CAST(merchant_balances.available AS NUMERIC) + CAST(? AS NUMERIC))::TEXT,
		    pending = (CAST(merchant_balances.pending AS NUMERIC) + CAST(? AS NUMERIC))::TEXT,
		    updated_at = now()
	`, merchantID, chain, token, deltaAvailable, deltaPending, deltaAvailable, deltaPending).Error
}

// --- Supported Currencies ---

func (s *Store) GetSupportedCurrencies(ctx context.Context) ([]models.SupportedCurrency, error) {
	var currencies []models.SupportedCurrency
	err := s.db.WithContext(ctx).Where("is_active = true").Order("chain, symbol").Find(&currencies).Error
	return currencies, err
}

func (s *Store) GetSupportedCurrenciesByChain(ctx context.Context, chain string) ([]models.SupportedCurrency, error) {
	var currencies []models.SupportedCurrency
	err := s.db.WithContext(ctx).Where("chain = ? AND is_active = true", chain).Order("symbol").Find(&currencies).Error
	return currencies, err
}

// --- Batch Payouts ---

func (s *Store) CreateBatchPayout(ctx context.Context, b *models.BatchPayout) error {
	return s.db.WithContext(ctx).Create(b).Error
}

func (s *Store) GetBatchPayoutByID(ctx context.Context, id string) (*models.BatchPayout, error) {
	var b models.BatchPayout
	err := s.db.WithContext(ctx).First(&b, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) GetBatchPayoutsByMerchant(ctx context.Context, merchantID string) ([]models.BatchPayout, error) {
	var batches []models.BatchPayout
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("created_at DESC").Find(&batches).Error
	return batches, err
}

func (s *Store) CreateBatchPayoutItem(ctx context.Context, item *models.BatchPayoutItem) error {
	return s.db.WithContext(ctx).Create(item).Error
}

func (s *Store) GetBatchPayoutItems(ctx context.Context, batchID string) ([]models.BatchPayoutItem, error) {
	var items []models.BatchPayoutItem
	err := s.db.WithContext(ctx).Where("batch_id = ?", batchID).Order("created_at").Find(&items).Error
	return items, err
}

func (s *Store) UpdateBatchPayoutStatus(ctx context.Context, id, status string, completed, failed int) error {
	return s.db.WithContext(ctx).Model(&models.BatchPayout{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "completed_count": completed, "failed_count": failed, "updated_at": time.Now()}).Error
}

// --- IP Whitelist ---

func (s *Store) SetMerchantIPWhitelist(ctx context.Context, merchantID string, ips []string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("merchant_id = ?", merchantID).Delete(&models.MerchantIPWhitelist{}).Error; err != nil {
			return err
		}
		for _, ip := range ips {
			entry := models.MerchantIPWhitelist{MerchantID: merchantID, IPAddress: ip}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) GetMerchantIPWhitelist(ctx context.Context, merchantID string) ([]string, error) {
	var entries []models.MerchantIPWhitelist
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&entries).Error
	if err != nil {
		return nil, err
	}
	ips := make([]string, len(entries))
	for i, e := range entries {
		ips[i] = e.IPAddress
	}
	return ips, nil
}

// --- Pagination ---

func (s *Store) GetPaymentsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error
	return payments, err
}

func (s *Store) GetPayoutsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]models.Payout, error) {
	var payouts []models.Payout
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&payouts).Error
	return payouts, err
}

// --- Payment Links ---

func (s *Store) CreatePaymentLink(ctx context.Context, link *models.PaymentLink) error {
	return s.db.WithContext(ctx).Create(link).Error
}

func (s *Store) GetPaymentLinkByID(ctx context.Context, id string) (*models.PaymentLink, error) {
	var link models.PaymentLink
	err := s.db.WithContext(ctx).Where("id = ? AND is_active = true", id).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *Store) GetPaymentLinksByMerchant(ctx context.Context, merchantID string) ([]models.PaymentLink, error) {
	var links []models.PaymentLink
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("created_at DESC").Find(&links).Error
	return links, err
}

func (s *Store) IncrementPaymentLinkUses(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Exec("UPDATE payment_links SET uses_count = uses_count + 1 WHERE id = ?", id).Error
}

// --- Audit Logs ---

func (s *Store) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return s.db.WithContext(ctx).Create(log).Error
}

func (s *Store) GetAuditLogs(ctx context.Context, merchantID string, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, err
}

// --- Export ---

func (s *Store) GetPaymentsByMerchantDateRange(ctx context.Context, merchantID, from, to string) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.WithContext(ctx).Where("merchant_id = ? AND created_at >= ?::timestamptz AND created_at <= ?::timestamptz", merchantID, from, to).
		Order("created_at DESC").Find(&payments).Error
	return payments, err
}

func (s *Store) GetPayoutsByMerchantDateRange(ctx context.Context, merchantID, from, to string) ([]models.Payout, error) {
	var payouts []models.Payout
	err := s.db.WithContext(ctx).Where("merchant_id = ? AND created_at >= ?::timestamptz AND created_at <= ?::timestamptz", merchantID, from, to).
		Order("created_at DESC").Find(&payouts).Error
	return payouts, err
}
