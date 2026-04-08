package models

import "time"

const (
	SweepStatusPending   = "pending"
	SweepStatusGasNeeded = "gas_needed"
	SweepStatusCompleted = "completed"
	SweepStatusFailed    = "failed"
)

type MerchantCollectionAddress struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID     string    `gorm:"type:uuid;uniqueIndex:idx_mca_merchant_chain;not null" json:"merchant_id"`
	Chain          string    `gorm:"uniqueIndex:idx_mca_merchant_chain;not null" json:"chain"`
	Address        string    `gorm:"not null" json:"address"`
	SweepThreshold string    `gorm:"default:'0'" json:"sweep_threshold"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type SweepTask struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID   string    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	PaymentID    *string   `gorm:"type:uuid" json:"payment_id,omitempty"`
	Chain        string    `gorm:"not null" json:"chain"`
	Token        string    `json:"token"`
	FromAddress  string    `gorm:"not null" json:"from_address"`
	ToAddress    string    `gorm:"not null" json:"to_address"`
	Amount       string    `gorm:"not null" json:"amount"`
	Status       string    `gorm:"default:'pending';index" json:"status"`
	TxHash       *string   `gorm:"column:tx_hash" json:"tx_hash,omitempty"`
	GasDepositID *string   `gorm:"type:uuid" json:"gas_deposit_id,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	RetryCount   int       `gorm:"default:0" json:"retry_count"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
