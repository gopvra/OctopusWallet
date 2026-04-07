package models

import "time"

const (
	SweepStatusPending    = "pending"
	SweepStatusGasNeeded  = "gas_needed"
	SweepStatusProcessing = "processing"
	SweepStatusCompleted  = "completed"
	SweepStatusFailed     = "failed"
)

type MerchantCollectionAddress struct {
	ID             string    `db:"id" json:"id"`
	MerchantID     string    `db:"merchant_id" json:"merchant_id"`
	Chain          string    `db:"chain" json:"chain"`
	Address        string    `db:"address" json:"address"`
	SweepThreshold string    `db:"sweep_threshold" json:"sweep_threshold"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type SweepTask struct {
	ID           string    `db:"id" json:"id"`
	MerchantID   string    `db:"merchant_id" json:"merchant_id"`
	PaymentID    *string   `db:"payment_id" json:"payment_id,omitempty"`
	Chain        string    `db:"chain" json:"chain"`
	Token        string    `db:"token" json:"token"`
	FromAddress  string    `db:"from_address" json:"from_address"`
	ToAddress    string    `db:"to_address" json:"to_address"`
	Amount       string    `db:"amount" json:"amount"`
	Status       string    `db:"status" json:"status"`
	TxHash       *string   `db:"tx_hash" json:"tx_hash,omitempty"`
	GasDepositID *string   `db:"gas_deposit_id" json:"gas_deposit_id,omitempty"`
	ErrorMessage *string   `db:"error_message" json:"error_message,omitempty"`
	RetryCount   int       `db:"retry_count" json:"retry_count"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
