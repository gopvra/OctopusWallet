package models

import "time"

const (
	TransferTypeHotToCold = "hot_to_cold"
	TransferTypeColdToHot = "cold_to_hot"
)

type ColdWalletConfig struct {
	ID                  string    `db:"id" json:"id"`
	MerchantID          string    `db:"merchant_id" json:"merchant_id"`
	Chain               string    `db:"chain" json:"chain"`
	ColdWalletAddress   string    `db:"cold_wallet_address" json:"cold_wallet_address"`
	HotWalletMaxBalance string    `db:"hot_wallet_max_balance" json:"hot_wallet_max_balance"`
	Enabled             bool      `db:"enabled" json:"enabled"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

type WalletTransfer struct {
	ID           string    `db:"id" json:"id"`
	MerchantID   string    `db:"merchant_id" json:"merchant_id"`
	Chain        string    `db:"chain" json:"chain"`
	Token        string    `db:"token" json:"token"`
	FromAddress  string    `db:"from_address" json:"from_address"`
	ToAddress    string    `db:"to_address" json:"to_address"`
	Amount       string    `db:"amount" json:"amount"`
	TransferType string    `db:"transfer_type" json:"transfer_type"`
	Status       string    `db:"status" json:"status"`
	TxHash       *string   `db:"tx_hash" json:"tx_hash,omitempty"`
	ErrorMessage *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
