package models

import "time"

const (
	ApprovalStatusNone            = ""
	ApprovalStatusPendingApproval = "pending_approval"
	ApprovalStatusApproved        = "approved"
	ApprovalStatusRejected        = "rejected"
)

type ApprovalConfig struct {
	ID                string    `db:"id" json:"id"`
	MerchantID        string    `db:"merchant_id" json:"merchant_id"`
	ApprovalThreshold string    `db:"approval_threshold" json:"approval_threshold"`
	SingleTxLimit     string    `db:"single_tx_limit" json:"single_tx_limit"`
	DailyLimit        string    `db:"daily_limit" json:"daily_limit"`
	AutoRelease       bool      `db:"auto_release" json:"auto_release"`
	Enabled           bool      `db:"enabled" json:"enabled"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

type PayoutApproval struct {
	ID           string    `db:"id" json:"id"`
	PayoutID     string    `db:"payout_id" json:"payout_id"`
	MerchantID   string    `db:"merchant_id" json:"merchant_id"`
	Action       string    `db:"action" json:"action"`
	ApproverID   string    `db:"approver_id" json:"approver_id"`
	ApproverNote string    `db:"approver_note" json:"approver_note"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
