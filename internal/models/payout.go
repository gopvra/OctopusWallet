package models

import "time"

const (
	PayoutStatusPending    = "pending"
	PayoutStatusProcessing = "processing"
	PayoutStatusCompleted  = "completed"
	PayoutStatusFailed     = "failed"
	PayoutStatusRejected   = "rejected"
)

type Payout struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID     string    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Chain          string    `gorm:"not null" json:"chain"`
	Token          string    `json:"token"`
	ToAddress      string    `gorm:"not null" json:"to_address"`
	Amount         string    `gorm:"not null" json:"amount"`
	Status         string    `gorm:"default:'pending';index" json:"status"`
	ApprovalStatus string    `gorm:"column:approval_status;default:''" json:"approval_status"`
	TxHash         *string   `gorm:"column:tx_hash" json:"tx_hash,omitempty"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
