package models

import "time"

const (
	RefundStatusPending    = "pending"
	RefundStatusProcessing = "processing"
	RefundStatusCompleted  = "completed"
	RefundStatusFailed     = "failed"
)

type Refund struct {
	ID           string    `db:"id" json:"id"`
	PaymentID    string    `db:"payment_id" json:"payment_id"`
	MerchantID   string    `db:"merchant_id" json:"merchant_id"`
	Chain        string    `db:"chain" json:"chain"`
	Token        string    `db:"token" json:"token"`
	ToAddress    string    `db:"to_address" json:"to_address"`
	Amount       string    `db:"amount" json:"amount"`
	Status       string    `db:"status" json:"status"`
	TxHash       *string   `db:"tx_hash" json:"tx_hash,omitempty"`
	Reason       string    `db:"reason" json:"reason"`
	ErrorMessage *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
