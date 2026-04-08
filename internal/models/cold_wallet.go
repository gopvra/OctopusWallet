package models

import "time"

const (
	TransferTypeHotToCold = "hot_to_cold"
	TransferTypeColdToHot = "cold_to_hot"
)

type ColdWalletConfig struct {
	ID                  string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID          string    `gorm:"type:uuid;uniqueIndex:idx_cwc_merchant_chain;not null" json:"merchant_id"`
	Chain               string    `gorm:"uniqueIndex:idx_cwc_merchant_chain;not null" json:"chain"`
	ColdWalletAddress   string    `gorm:"not null" json:"cold_wallet_address"`
	HotWalletMaxBalance string    `gorm:"not null" json:"hot_wallet_max_balance"`
	Enabled             bool      `gorm:"default:false" json:"enabled"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type WalletTransfer struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID   string    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Chain        string    `gorm:"not null" json:"chain"`
	Token        string    `json:"token"`
	FromAddress  string    `gorm:"not null" json:"from_address"`
	ToAddress    string    `gorm:"not null" json:"to_address"`
	Amount       string    `gorm:"not null" json:"amount"`
	TransferType string    `gorm:"not null" json:"transfer_type"`
	Status       string    `gorm:"default:'pending';index" json:"status"`
	TxHash       *string   `gorm:"column:tx_hash" json:"tx_hash,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
