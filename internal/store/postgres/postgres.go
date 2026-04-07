package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/octopuswallet/octopuswallet/internal/models"
)

type Store struct {
	db *sqlx.DB
}

func New(databaseURL string, maxOpenConns int) (*Store, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// --- Merchants ---

func (s *Store) CreateMerchant(ctx context.Context, m *models.Merchant) error {
	query := `INSERT INTO merchants (name, email, api_key_hash, webhook_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, is_active`
	return s.db.QueryRowxContext(ctx, query, m.Name, m.Email, m.APIKeyHash, m.WebhookURL).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt, &m.IsActive)
}

func (s *Store) GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	var m models.Merchant
	err := s.db.GetContext(ctx, &m, "SELECT * FROM merchants WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) GetMerchantByAPIKeyHash(ctx context.Context, hash string) (*models.Merchant, error) {
	var m models.Merchant
	err := s.db.GetContext(ctx, &m, "SELECT * FROM merchants WHERE api_key_hash = $1 AND is_active = true", hash)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// --- Wallets ---

func (s *Store) CreateWallet(ctx context.Context, w *models.Wallet) error {
	query := `INSERT INTO wallets (merchant_id, chain, address, derivation_index)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return s.db.QueryRowxContext(ctx, query, w.MerchantID, w.Chain, w.Address, w.DerivationIndex).
		Scan(&w.ID, &w.CreatedAt)
}

func (s *Store) GetWalletsByMerchant(ctx context.Context, merchantID string) ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := s.db.SelectContext(ctx, &wallets,
		"SELECT * FROM wallets WHERE merchant_id = $1 ORDER BY chain, derivation_index", merchantID)
	return wallets, err
}

func (s *Store) GetWalletByAddress(ctx context.Context, chain, address string) (*models.Wallet, error) {
	var w models.Wallet
	err := s.db.GetContext(ctx, &w, "SELECT * FROM wallets WHERE chain = $1 AND address = $2", chain, address)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) GetAllWatchAddresses(ctx context.Context) (map[string]map[string]struct{}, error) {
	type row struct {
		Chain   string `db:"chain"`
		Address string `db:"address"`
	}
	var rows []row
	err := s.db.SelectContext(ctx, &rows,
		`SELECT w.chain, w.address FROM wallets w
		 INNER JOIN payments p ON p.address = w.address AND p.chain = w.chain
		 WHERE p.status IN ('pending', 'confirming')`)
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]struct{})
	for _, r := range rows {
		if result[r.Chain] == nil {
			result[r.Chain] = make(map[string]struct{})
		}
		result[r.Chain][r.Address] = struct{}{}
	}
	return result, nil
}

func (s *Store) GetNextDerivationIndex(ctx context.Context, merchantID, chain string) (int, error) {
	var idx sql.NullInt64
	err := s.db.GetContext(ctx, &idx,
		"SELECT MAX(derivation_index) FROM wallets WHERE merchant_id = $1 AND chain = $2",
		merchantID, chain)
	if err != nil {
		return 0, err
	}
	if !idx.Valid {
		return 0, nil
	}
	return int(idx.Int64) + 1, nil
}

// --- Payments ---

func (s *Store) CreatePayment(ctx context.Context, p *models.Payment) error {
	query := `INSERT INTO payments (merchant_id, chain, token, amount_expected, address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status, amount_received, confirmations, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		p.MerchantID, p.Chain, p.Token, p.AmountExpected, p.Address, p.ExpiresAt).
		Scan(&p.ID, &p.Status, &p.AmountReceived, &p.Confirmations, &p.CreatedAt, &p.UpdatedAt)
}

func (s *Store) GetPaymentByID(ctx context.Context, id string) (*models.Payment, error) {
	var p models.Payment
	err := s.db.GetContext(ctx, &p, "SELECT * FROM payments WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetPendingPayments(ctx context.Context) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.SelectContext(ctx, &payments,
		"SELECT * FROM payments WHERE status IN ('pending', 'confirming') ORDER BY created_at")
	return payments, err
}

func (s *Store) GetPaymentByAddress(ctx context.Context, chain, address string) (*models.Payment, error) {
	var p models.Payment
	err := s.db.GetContext(ctx, &p,
		"SELECT * FROM payments WHERE chain = $1 AND address = $2 AND status IN ('pending', 'confirming') ORDER BY created_at DESC LIMIT 1",
		chain, address)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) UpdatePaymentStatus(ctx context.Context, id, status string, txHash *string, confirmations int) error {
	// Use conditional update to prevent race conditions:
	// Only update if current status allows the transition
	result, err := s.db.ExecContext(ctx,
		`UPDATE payments SET status = $1, tx_hash = COALESCE($2, tx_hash), confirmations = $3, updated_at = now()
		 WHERE id = $4 AND status NOT IN ('completed', 'expired')`,
		status, txHash, confirmations, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("payment %s already in terminal state", id)
	}
	return nil
}

// --- Payouts ---

func (s *Store) CreatePayout(ctx context.Context, p *models.Payout) error {
	query := `INSERT INTO payouts (merchant_id, chain, token, to_address, amount, approval_status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status, approval_status, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		p.MerchantID, p.Chain, p.Token, p.ToAddress, p.Amount, p.ApprovalStatus).
		Scan(&p.ID, &p.Status, &p.ApprovalStatus, &p.CreatedAt, &p.UpdatedAt)
}

func (s *Store) GetPayoutByID(ctx context.Context, id string) (*models.Payout, error) {
	var p models.Payout
	err := s.db.GetContext(ctx, &p, "SELECT * FROM payouts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetPendingPayouts(ctx context.Context) ([]models.Payout, error) {
	// Atomically claim pending payouts to prevent double-processing
	var payouts []models.Payout
	err := s.db.SelectContext(ctx, &payouts,
		`UPDATE payouts SET status = 'processing', updated_at = now()
		 WHERE id IN (
		   SELECT id FROM payouts
		   WHERE status = 'pending' AND (approval_status = '' OR approval_status = 'approved')
		   ORDER BY created_at LIMIT 10 FOR UPDATE SKIP LOCKED
		 ) RETURNING *`)
	return payouts, err
}

func (s *Store) UpdatePayoutStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE payouts SET status = $1, tx_hash = $2, error_message = $3, updated_at = now() WHERE id = $4",
		status, txHash, errMsg, id)
	return err
}

// --- Chain State ---

func (s *Store) GetLastScannedBlock(ctx context.Context, chain string) (uint64, error) {
	var height uint64
	err := s.db.GetContext(ctx, &height,
		"SELECT last_scanned_block FROM chain_state WHERE chain = $1", chain)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return height, err
}

func (s *Store) SetLastScannedBlock(ctx context.Context, chain string, height uint64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chain_state (chain, last_scanned_block, updated_at) VALUES ($1, $2, now())
		 ON CONFLICT (chain) DO UPDATE SET last_scanned_block = $2, updated_at = now()`,
		chain, height)
	return err
}

// --- Sweep ---

func (s *Store) UpsertMerchantCollectionAddress(ctx context.Context, addr *models.MerchantCollectionAddress) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO merchant_collection_addresses (merchant_id, chain, address, sweep_threshold, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (merchant_id, chain) DO UPDATE
		 SET address = $3, sweep_threshold = $4, is_active = $5, updated_at = now()`,
		addr.MerchantID, addr.Chain, addr.Address, addr.SweepThreshold, addr.IsActive)
	return err
}

func (s *Store) GetMerchantCollectionAddress(ctx context.Context, merchantID, chain string) (*models.MerchantCollectionAddress, error) {
	var addr models.MerchantCollectionAddress
	err := s.db.GetContext(ctx, &addr,
		"SELECT * FROM merchant_collection_addresses WHERE merchant_id = $1 AND chain = $2 AND is_active = true",
		merchantID, chain)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (s *Store) GetMerchantCollectionAddresses(ctx context.Context, merchantID string) ([]models.MerchantCollectionAddress, error) {
	var addrs []models.MerchantCollectionAddress
	err := s.db.SelectContext(ctx, &addrs,
		"SELECT * FROM merchant_collection_addresses WHERE merchant_id = $1", merchantID)
	return addrs, err
}

func (s *Store) CreateSweepTask(ctx context.Context, task *models.SweepTask) error {
	query := `INSERT INTO sweep_tasks (merchant_id, payment_id, chain, token, from_address, to_address, amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, status, retry_count, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		task.MerchantID, task.PaymentID, task.Chain, task.Token, task.FromAddress, task.ToAddress, task.Amount).
		Scan(&task.ID, &task.Status, &task.RetryCount, &task.CreatedAt, &task.UpdatedAt)
}

func (s *Store) GetPendingSweepTasks(ctx context.Context) ([]models.SweepTask, error) {
	// Atomically claim pending sweep tasks (not gas_needed — those need different handling)
	var pending []models.SweepTask
	err := s.db.SelectContext(ctx, &pending,
		`UPDATE sweep_tasks SET status = 'processing', updated_at = now()
		 WHERE id IN (
		   SELECT id FROM sweep_tasks
		   WHERE status = 'pending'
		   ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED
		 ) RETURNING *`)
	if err != nil {
		return nil, err
	}

	// Gas-needed tasks: select with lock but don't change status (checkGasDeposit handles transition)
	var gasNeeded []models.SweepTask
	err = s.db.SelectContext(ctx, &gasNeeded,
		`SELECT * FROM sweep_tasks
		 WHERE id IN (
		   SELECT id FROM sweep_tasks
		   WHERE status = 'gas_needed'
		   ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED
		 )`)
	if err != nil {
		return pending, nil
	}

	return append(pending, gasNeeded...), nil
}

func (s *Store) GetSweepTasksByMerchant(ctx context.Context, merchantID string) ([]models.SweepTask, error) {
	var tasks []models.SweepTask
	err := s.db.SelectContext(ctx, &tasks,
		"SELECT * FROM sweep_tasks WHERE merchant_id = $1 ORDER BY created_at DESC", merchantID)
	return tasks, err
}

func (s *Store) UpdateSweepTaskStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE sweep_tasks SET status = $1, tx_hash = $2, error_message = $3, updated_at = now() WHERE id = $4",
		status, txHash, errMsg, id)
	return err
}

func (s *Store) UpdateSweepTaskGasDeposit(ctx context.Context, id, gasDepositID string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE sweep_tasks SET gas_deposit_id = $1, status = 'gas_needed', updated_at = now() WHERE id = $2",
		gasDepositID, id)
	return err
}

// --- Cold/Hot Wallet ---

func (s *Store) UpsertColdWalletConfig(ctx context.Context, cfg *models.ColdWalletConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO cold_wallet_configs (merchant_id, chain, cold_wallet_address, hot_wallet_max_balance, enabled)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (merchant_id, chain) DO UPDATE
		 SET cold_wallet_address = $3, hot_wallet_max_balance = $4, enabled = $5, updated_at = now()`,
		cfg.MerchantID, cfg.Chain, cfg.ColdWalletAddress, cfg.HotWalletMaxBalance, cfg.Enabled)
	return err
}

func (s *Store) GetColdWalletConfig(ctx context.Context, merchantID, chain string) (*models.ColdWalletConfig, error) {
	var cfg models.ColdWalletConfig
	err := s.db.GetContext(ctx, &cfg,
		"SELECT * FROM cold_wallet_configs WHERE merchant_id = $1 AND chain = $2", merchantID, chain)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) GetAllEnabledColdWalletConfigs(ctx context.Context) ([]models.ColdWalletConfig, error) {
	var configs []models.ColdWalletConfig
	err := s.db.SelectContext(ctx, &configs,
		"SELECT * FROM cold_wallet_configs WHERE enabled = true")
	return configs, err
}

func (s *Store) GetColdWalletConfigsByMerchant(ctx context.Context, merchantID string) ([]models.ColdWalletConfig, error) {
	var configs []models.ColdWalletConfig
	err := s.db.SelectContext(ctx, &configs,
		"SELECT * FROM cold_wallet_configs WHERE merchant_id = $1", merchantID)
	return configs, err
}

func (s *Store) CreateWalletTransfer(ctx context.Context, transfer *models.WalletTransfer) error {
	query := `INSERT INTO wallet_transfers (merchant_id, chain, token, from_address, to_address, amount, transfer_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, status, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		transfer.MerchantID, transfer.Chain, transfer.Token, transfer.FromAddress, transfer.ToAddress, transfer.Amount, transfer.TransferType).
		Scan(&transfer.ID, &transfer.Status, &transfer.CreatedAt, &transfer.UpdatedAt)
}

func (s *Store) GetPendingWalletTransfers(ctx context.Context) ([]models.WalletTransfer, error) {
	var transfers []models.WalletTransfer
	err := s.db.SelectContext(ctx, &transfers,
		`UPDATE wallet_transfers SET status = 'processing', updated_at = now()
		 WHERE id IN (
		   SELECT id FROM wallet_transfers WHERE status = 'pending'
		   ORDER BY created_at LIMIT 10 FOR UPDATE SKIP LOCKED
		 ) RETURNING *`)
	return transfers, err
}

func (s *Store) GetWalletTransfersByMerchant(ctx context.Context, merchantID string) ([]models.WalletTransfer, error) {
	var transfers []models.WalletTransfer
	err := s.db.SelectContext(ctx, &transfers,
		"SELECT * FROM wallet_transfers WHERE merchant_id = $1 ORDER BY created_at DESC", merchantID)
	return transfers, err
}

func (s *Store) UpdateWalletTransferStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE wallet_transfers SET status = $1, tx_hash = $2, error_message = $3, updated_at = now() WHERE id = $4",
		status, txHash, errMsg, id)
	return err
}

// --- Approval ---

func (s *Store) UpsertApprovalConfig(ctx context.Context, cfg *models.ApprovalConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO approval_configs (merchant_id, approval_threshold, single_tx_limit, daily_limit, auto_release, enabled)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (merchant_id) DO UPDATE
		 SET approval_threshold = $2, single_tx_limit = $3, daily_limit = $4, auto_release = $5, enabled = $6, updated_at = now()`,
		cfg.MerchantID, cfg.ApprovalThreshold, cfg.SingleTxLimit, cfg.DailyLimit, cfg.AutoRelease, cfg.Enabled)
	return err
}

func (s *Store) GetApprovalConfig(ctx context.Context, merchantID string) (*models.ApprovalConfig, error) {
	var cfg models.ApprovalConfig
	err := s.db.GetContext(ctx, &cfg,
		"SELECT * FROM approval_configs WHERE merchant_id = $1", merchantID)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) CreatePayoutApproval(ctx context.Context, approval *models.PayoutApproval) error {
	query := `INSERT INTO payout_approvals (payout_id, merchant_id, action, approver_id, approver_note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return s.db.QueryRowxContext(ctx, query,
		approval.PayoutID, approval.MerchantID, approval.Action, approval.ApproverID, approval.ApproverNote).
		Scan(&approval.ID, &approval.CreatedAt)
}

func (s *Store) UpdatePayoutApprovalStatus(ctx context.Context, payoutID, approvalStatus string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE payouts SET approval_status = $1, updated_at = now() WHERE id = $2",
		approvalStatus, payoutID)
	return err
}

func (s *Store) GetDailyPayoutTotal(ctx context.Context, merchantID, chain string) (string, error) {
	var total sql.NullString
	err := s.db.GetContext(ctx, &total,
		"SELECT total_amount FROM payout_daily_totals WHERE merchant_id = $1 AND chain = $2 AND date = CURRENT_DATE",
		merchantID, chain)
	if err != nil || !total.Valid {
		return "0", nil
	}
	return total.String, nil
}

func (s *Store) IncrementDailyPayoutTotal(ctx context.Context, merchantID, chain, amount string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO payout_daily_totals (merchant_id, chain, date, total_amount)
		 VALUES ($1, $2, CURRENT_DATE, $3)
		 ON CONFLICT (merchant_id, chain, date) DO UPDATE
		 SET total_amount = (CAST(payout_daily_totals.total_amount AS NUMERIC) + CAST($3 AS NUMERIC))::TEXT,
		     updated_at = now()`,
		merchantID, chain, amount)
	return err
}

// --- Gas ---

func (s *Store) CreateGasDeposit(ctx context.Context, deposit *models.GasDeposit) error {
	query := `INSERT INTO gas_deposits (chain, to_address, amount, sweep_task_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		deposit.Chain, deposit.ToAddress, deposit.Amount, deposit.SweepTaskID).
		Scan(&deposit.ID, &deposit.Status, &deposit.CreatedAt, &deposit.UpdatedAt)
}

func (s *Store) GetPendingGasDeposits(ctx context.Context) ([]models.GasDeposit, error) {
	var deposits []models.GasDeposit
	err := s.db.SelectContext(ctx, &deposits,
		`UPDATE gas_deposits SET status = 'processing', updated_at = now()
		 WHERE id IN (
		   SELECT id FROM gas_deposits WHERE status = 'pending'
		   ORDER BY created_at LIMIT 10 FOR UPDATE SKIP LOCKED
		 ) RETURNING *`)
	return deposits, err
}

func (s *Store) UpdateGasDepositStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE gas_deposits SET status = $1, tx_hash = $2, error_message = $3, updated_at = now() WHERE id = $4",
		status, txHash, errMsg, id)
	return err
}

func (s *Store) GetGasDepositBySweepTask(ctx context.Context, sweepTaskID string) (*models.GasDeposit, error) {
	var deposit models.GasDeposit
	err := s.db.GetContext(ctx, &deposit,
		"SELECT * FROM gas_deposits WHERE sweep_task_id = $1 ORDER BY created_at DESC LIMIT 1", sweepTaskID)
	if err != nil {
		return nil, err
	}
	return &deposit, nil
}

func (s *Store) CreateGasAlert(ctx context.Context, alert *models.GasAlert) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO gas_alerts (chain, station_address, balance, threshold) VALUES ($1, $2, $3, $4)",
		alert.Chain, alert.StationAddress, alert.Balance, alert.Threshold)
	return err
}

// --- Refunds ---

func (s *Store) CreateRefund(ctx context.Context, r *models.Refund) error {
	query := `INSERT INTO refunds (payment_id, merchant_id, chain, token, to_address, amount, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, status, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		r.PaymentID, r.MerchantID, r.Chain, r.Token, r.ToAddress, r.Amount, r.Reason).
		Scan(&r.ID, &r.Status, &r.CreatedAt, &r.UpdatedAt)
}

func (s *Store) GetRefundByID(ctx context.Context, id string) (*models.Refund, error) {
	var r models.Refund
	err := s.db.GetContext(ctx, &r, "SELECT * FROM refunds WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) GetRefundsByPayment(ctx context.Context, paymentID string) ([]models.Refund, error) {
	var refunds []models.Refund
	err := s.db.SelectContext(ctx, &refunds,
		"SELECT * FROM refunds WHERE payment_id = $1 ORDER BY created_at DESC", paymentID)
	return refunds, err
}

func (s *Store) GetRefundTotalByPayment(ctx context.Context, paymentID string) (string, error) {
	var total string
	err := s.db.GetContext(ctx, &total,
		"SELECT COALESCE(SUM(amount::numeric), 0)::text FROM refunds WHERE payment_id = $1 AND status != 'failed'", paymentID)
	return total, err
}

func (s *Store) GetPendingRefunds(ctx context.Context) ([]models.Refund, error) {
	var refunds []models.Refund
	err := s.db.SelectContext(ctx, &refunds,
		`UPDATE refunds SET status = 'processing', updated_at = now()
		 WHERE id IN (SELECT id FROM refunds WHERE status = 'pending' ORDER BY created_at LIMIT 10 FOR UPDATE SKIP LOCKED)
		 RETURNING *`)
	return refunds, err
}

func (s *Store) UpdateRefundStatus(ctx context.Context, id, status string, txHash *string, errMsg *string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE refunds SET status = $1, tx_hash = $2, error_message = $3, updated_at = now() WHERE id = $4",
		status, txHash, errMsg, id)
	return err
}

// --- Merchant Balances ---

func (s *Store) GetMerchantBalances(ctx context.Context, merchantID string) ([]models.MerchantBalance, error) {
	var balances []models.MerchantBalance
	err := s.db.SelectContext(ctx, &balances,
		"SELECT * FROM merchant_balances WHERE merchant_id = $1", merchantID)
	return balances, err
}

func (s *Store) UpdateMerchantBalance(ctx context.Context, merchantID, chain, token, deltaAvailable, deltaPending string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO merchant_balances (merchant_id, chain, token, available, pending)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (merchant_id, chain, token) DO UPDATE
		 SET available = (CAST(merchant_balances.available AS NUMERIC) + CAST($4 AS NUMERIC))::TEXT,
		     pending = (CAST(merchant_balances.pending AS NUMERIC) + CAST($5 AS NUMERIC))::TEXT,
		     updated_at = now()`,
		merchantID, chain, token, deltaAvailable, deltaPending)
	return err
}

// --- Supported Currencies ---

func (s *Store) GetSupportedCurrencies(ctx context.Context) ([]models.SupportedCurrency, error) {
	var currencies []models.SupportedCurrency
	err := s.db.SelectContext(ctx, &currencies,
		"SELECT * FROM supported_currencies WHERE is_active = true ORDER BY chain, symbol")
	return currencies, err
}

func (s *Store) GetSupportedCurrenciesByChain(ctx context.Context, chain string) ([]models.SupportedCurrency, error) {
	var currencies []models.SupportedCurrency
	err := s.db.SelectContext(ctx, &currencies,
		"SELECT * FROM supported_currencies WHERE chain = $1 AND is_active = true ORDER BY symbol", chain)
	return currencies, err
}

// --- Batch Payouts ---

func (s *Store) CreateBatchPayout(ctx context.Context, b *models.BatchPayout) error {
	query := `INSERT INTO batch_payouts (merchant_id, chain, token, total_amount, total_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, completed_count, failed_count, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		b.MerchantID, b.Chain, b.Token, b.TotalAmount, b.TotalCount).
		Scan(&b.ID, &b.Status, &b.CompletedCount, &b.FailedCount, &b.CreatedAt, &b.UpdatedAt)
}

func (s *Store) GetBatchPayoutByID(ctx context.Context, id string) (*models.BatchPayout, error) {
	var b models.BatchPayout
	err := s.db.GetContext(ctx, &b, "SELECT * FROM batch_payouts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) GetBatchPayoutsByMerchant(ctx context.Context, merchantID string) ([]models.BatchPayout, error) {
	var batches []models.BatchPayout
	err := s.db.SelectContext(ctx, &batches,
		"SELECT * FROM batch_payouts WHERE merchant_id = $1 ORDER BY created_at DESC", merchantID)
	return batches, err
}

func (s *Store) CreateBatchPayoutItem(ctx context.Context, item *models.BatchPayoutItem) error {
	query := `INSERT INTO batch_payout_items (batch_id, to_address, amount)
		VALUES ($1, $2, $3) RETURNING id, status, created_at`
	return s.db.QueryRowxContext(ctx, query, item.BatchID, item.ToAddress, item.Amount).
		Scan(&item.ID, &item.Status, &item.CreatedAt)
}

func (s *Store) GetBatchPayoutItems(ctx context.Context, batchID string) ([]models.BatchPayoutItem, error) {
	var items []models.BatchPayoutItem
	err := s.db.SelectContext(ctx, &items,
		"SELECT * FROM batch_payout_items WHERE batch_id = $1 ORDER BY created_at", batchID)
	return items, err
}

func (s *Store) UpdateBatchPayoutStatus(ctx context.Context, id, status string, completed, failed int) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE batch_payouts SET status = $1, completed_count = $2, failed_count = $3, updated_at = now() WHERE id = $4",
		status, completed, failed, id)
	return err
}

// --- IP Whitelist ---

func (s *Store) SetMerchantIPWhitelist(ctx context.Context, merchantID string, ips []string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM merchant_ip_whitelist WHERE merchant_id = $1", merchantID)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO merchant_ip_whitelist (merchant_id, ip_address) VALUES ($1, $2)", merchantID, ip)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetMerchantIPWhitelist(ctx context.Context, merchantID string) ([]string, error) {
	var ips []string
	err := s.db.SelectContext(ctx, &ips,
		"SELECT ip_address FROM merchant_ip_whitelist WHERE merchant_id = $1", merchantID)
	return ips, err
}

// --- Pagination ---

func (s *Store) GetPaymentsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.SelectContext(ctx, &payments,
		"SELECT * FROM payments WHERE merchant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		merchantID, limit, offset)
	return payments, err
}

func (s *Store) GetPayoutsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]models.Payout, error) {
	var payouts []models.Payout
	err := s.db.SelectContext(ctx, &payouts,
		"SELECT * FROM payouts WHERE merchant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		merchantID, limit, offset)
	return payouts, err
}

// --- Payment Links ---

func (s *Store) CreatePaymentLink(ctx context.Context, link *models.PaymentLink) error {
	query := `INSERT INTO payment_links (merchant_id, chain, token, amount, currency, description, redirect_url, is_reusable)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, is_active, uses_count, created_at`
	return s.db.QueryRowxContext(ctx, query,
		link.MerchantID, link.Chain, link.Token, link.Amount, link.Currency, link.Description, link.RedirectURL, link.IsReusable).
		Scan(&link.ID, &link.IsActive, &link.UsesCount, &link.CreatedAt)
}

func (s *Store) GetPaymentLinkByID(ctx context.Context, id string) (*models.PaymentLink, error) {
	var link models.PaymentLink
	err := s.db.GetContext(ctx, &link, "SELECT * FROM payment_links WHERE id = $1 AND is_active = true", id)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *Store) GetPaymentLinksByMerchant(ctx context.Context, merchantID string) ([]models.PaymentLink, error) {
	var links []models.PaymentLink
	err := s.db.SelectContext(ctx, &links,
		"SELECT * FROM payment_links WHERE merchant_id = $1 ORDER BY created_at DESC", merchantID)
	return links, err
}

func (s *Store) IncrementPaymentLinkUses(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE payment_links SET uses_count = uses_count + 1 WHERE id = $1", id)
	return err
}

// --- Audit Logs ---

func (s *Store) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (merchant_id, action, resource_type, resource_id, ip_address)
		 VALUES ($1, $2, $3, $4, $5)`,
		log.MerchantID, log.Action, log.ResourceType, log.ResourceID, log.IPAddress)
	return err
}

func (s *Store) GetAuditLogs(ctx context.Context, merchantID string, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.SelectContext(ctx, &logs,
		"SELECT * FROM audit_logs WHERE merchant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		merchantID, limit, offset)
	return logs, err
}

// --- Export ---

func (s *Store) GetPaymentsByMerchantDateRange(ctx context.Context, merchantID, from, to string) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.SelectContext(ctx, &payments,
		`SELECT * FROM payments WHERE merchant_id = $1 AND created_at >= $2::timestamptz AND created_at <= $3::timestamptz
		 ORDER BY created_at DESC`, merchantID, from, to)
	return payments, err
}

func (s *Store) GetPayoutsByMerchantDateRange(ctx context.Context, merchantID, from, to string) ([]models.Payout, error) {
	var payouts []models.Payout
	err := s.db.SelectContext(ctx, &payouts,
		`SELECT * FROM payouts WHERE merchant_id = $1 AND created_at >= $2::timestamptz AND created_at <= $3::timestamptz
		 ORDER BY created_at DESC`, merchantID, from, to)
	return payouts, err
}
