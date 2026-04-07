package postgres

import (
	"context"
	"database/sql"
	"fmt"

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
	_, err := s.db.ExecContext(ctx,
		"UPDATE payments SET status = $1, tx_hash = $2, confirmations = $3, updated_at = now() WHERE id = $4",
		status, txHash, confirmations, id)
	return err
}

// --- Payouts ---

func (s *Store) CreatePayout(ctx context.Context, p *models.Payout) error {
	query := `INSERT INTO payouts (merchant_id, chain, token, to_address, amount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query,
		p.MerchantID, p.Chain, p.Token, p.ToAddress, p.Amount).
		Scan(&p.ID, &p.Status, &p.CreatedAt, &p.UpdatedAt)
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
	var payouts []models.Payout
	err := s.db.SelectContext(ctx, &payouts,
		"SELECT * FROM payouts WHERE status = 'pending' ORDER BY created_at")
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
