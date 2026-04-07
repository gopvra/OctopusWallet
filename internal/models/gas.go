package models

import "time"

const (
	GasDepositStatusPending    = "pending"
	GasDepositStatusProcessing = "processing"
	GasDepositStatusCompleted  = "completed"
	GasDepositStatusFailed     = "failed"
)

type GasDeposit struct {
	ID           string    `db:"id" json:"id"`
	Chain        string    `db:"chain" json:"chain"`
	ToAddress    string    `db:"to_address" json:"to_address"`
	Amount       string    `db:"amount" json:"amount"`
	Status       string    `db:"status" json:"status"`
	TxHash       *string   `db:"tx_hash" json:"tx_hash,omitempty"`
	SweepTaskID  *string   `db:"sweep_task_id" json:"sweep_task_id,omitempty"`
	ErrorMessage *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type GasAlert struct {
	ID             string    `db:"id" json:"id"`
	Chain          string    `db:"chain" json:"chain"`
	StationAddress string    `db:"station_address" json:"station_address"`
	Balance        string    `db:"balance" json:"balance"`
	Threshold      string    `db:"threshold" json:"threshold"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
