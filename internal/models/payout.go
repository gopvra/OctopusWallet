package models

import "time"

const (
	PayoutStatusPending    = "pending"
	PayoutStatusProcessing = "processing"
	PayoutStatusCompleted  = "completed"
	PayoutStatusFailed     = "failed"
)

type Payout struct {
	ID           string    `db:"id" json:"id"`
	MerchantID   string    `db:"merchant_id" json:"merchant_id"`
	Chain        string    `db:"chain" json:"chain"`
	Token        string    `db:"token" json:"token"`
	ToAddress    string    `db:"to_address" json:"to_address"`
	Amount       string    `db:"amount" json:"amount"`
	Status       string    `db:"status" json:"status"`
	TxHash       *string   `db:"tx_hash" json:"tx_hash,omitempty"`
	ErrorMessage *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
