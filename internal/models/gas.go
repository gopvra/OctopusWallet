package models

import "time"

const (
	GasDepositStatusPending    = "pending"
	GasDepositStatusProcessing = "processing"
	GasDepositStatusCompleted  = "completed"
	GasDepositStatusFailed     = "failed"
)

type GasDeposit struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Chain        string    `gorm:"not null" json:"chain"`
	ToAddress    string    `gorm:"not null" json:"to_address"`
	Amount       string    `gorm:"not null" json:"amount"`
	Status       string    `gorm:"default:'pending';index" json:"status"`
	TxHash       *string   `gorm:"column:tx_hash" json:"tx_hash,omitempty"`
	SweepTaskID  *string   `gorm:"type:uuid" json:"sweep_task_id,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type GasAlert struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Chain          string    `gorm:"not null" json:"chain"`
	StationAddress string    `gorm:"not null" json:"station_address"`
	Balance        string    `gorm:"not null" json:"balance"`
	Threshold      string    `gorm:"not null" json:"threshold"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}
