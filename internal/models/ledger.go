package models

import "time"

type MerchantBalance struct {
	ID         string    `db:"id" json:"id"`
	MerchantID string    `db:"merchant_id" json:"merchant_id"`
	Chain      string    `db:"chain" json:"chain"`
	Token      string    `db:"token" json:"token"`
	Available  string    `db:"available" json:"available"`
	Pending    string    `db:"pending" json:"pending"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type SupportedCurrency struct {
	ID           string    `db:"id" json:"id"`
	Chain        string    `db:"chain" json:"chain"`
	Symbol       string    `db:"symbol" json:"symbol"`
	Name         string    `db:"name" json:"name"`
	TokenAddress string    `db:"token_address" json:"token_address"`
	Decimals     int       `db:"decimals" json:"decimals"`
	IsNative     bool      `db:"is_native" json:"is_native"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	MinAmount    string    `db:"min_amount" json:"min_amount"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type BatchPayout struct {
	ID             string    `db:"id" json:"id"`
	MerchantID     string    `db:"merchant_id" json:"merchant_id"`
	Chain          string    `db:"chain" json:"chain"`
	Token          string    `db:"token" json:"token"`
	TotalAmount    string    `db:"total_amount" json:"total_amount"`
	TotalCount     int       `db:"total_count" json:"total_count"`
	CompletedCount int       `db:"completed_count" json:"completed_count"`
	FailedCount    int       `db:"failed_count" json:"failed_count"`
	Status         string    `db:"status" json:"status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type BatchPayoutItem struct {
	ID           string    `db:"id" json:"id"`
	BatchID      string    `db:"batch_id" json:"batch_id"`
	PayoutID     *string   `db:"payout_id" json:"payout_id,omitempty"`
	ToAddress    string    `db:"to_address" json:"to_address"`
	Amount       string    `db:"amount" json:"amount"`
	Status       string    `db:"status" json:"status"`
	ErrorMessage *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
