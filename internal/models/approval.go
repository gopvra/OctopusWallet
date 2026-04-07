package models

import "time"

const (
	ApprovalStatusNone            = ""
	ApprovalStatusPendingApproval = "pending_approval"
	ApprovalStatusApproved        = "approved"
	ApprovalStatusRejected        = "rejected"
)

type ApprovalConfig struct {
	ID                string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID        string    `gorm:"type:uuid;uniqueIndex;not null" json:"merchant_id"`
	ApprovalThreshold string    `gorm:"not null" json:"approval_threshold"`
	SingleTxLimit     string    `json:"single_tx_limit"`
	DailyLimit        string    `json:"daily_limit"`
	AutoRelease       bool      `gorm:"default:false" json:"auto_release"`
	Enabled           bool      `gorm:"default:false" json:"enabled"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type PayoutApproval struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	PayoutID     string    `gorm:"type:uuid;index;not null" json:"payout_id"`
	MerchantID   string    `gorm:"type:uuid;not null" json:"merchant_id"`
	Action       string    `gorm:"not null" json:"action"`
	ApproverID   string    `gorm:"not null" json:"approver_id"`
	ApproverNote string    `json:"approver_note"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type PayoutDailyTotal struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID  string    `gorm:"type:uuid;uniqueIndex:idx_pdt_merchant_chain_date;not null" json:"merchant_id"`
	Chain       string    `gorm:"uniqueIndex:idx_pdt_merchant_chain_date;not null" json:"chain"`
	Date        string    `gorm:"uniqueIndex:idx_pdt_merchant_chain_date;not null" json:"date"`
	TotalAmount string    `gorm:"not null;default:'0'" json:"total_amount"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PayoutDailyTotal) TableName() string {
	return "payout_daily_totals"
}
