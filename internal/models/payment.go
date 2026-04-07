package models

import "time"

const (
	PaymentStatusPending    = "pending"
	PaymentStatusConfirming = "confirming"
	PaymentStatusCompleted  = "completed"
	PaymentStatusExpired    = "expired"
)

type Payment struct {
	ID             string     `db:"id" json:"id"`
	MerchantID     string     `db:"merchant_id" json:"merchant_id"`
	Chain          string     `db:"chain" json:"chain"`
	Token          string     `db:"token" json:"token"`
	AmountExpected string     `db:"amount_expected" json:"amount_expected"`
	AmountReceived string     `db:"amount_received" json:"amount_received"`
	Address        string     `db:"address" json:"address"`
	Status         string     `db:"status" json:"status"`
	TxHash         *string    `db:"tx_hash" json:"tx_hash,omitempty"`
	Confirmations  int        `db:"confirmations" json:"confirmations"`
	ExpiresAt      *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}
