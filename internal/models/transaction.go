package models

import "time"

type Transaction struct {
	ID            string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Chain         string    `gorm:"not null;index" json:"chain"`
	TxHash        string    `gorm:"column:tx_hash;not null;index" json:"tx_hash"`
	FromAddress   string    `gorm:"not null" json:"from_address"`
	ToAddress     string    `gorm:"not null" json:"to_address"`
	Amount        string    `gorm:"not null" json:"amount"`
	Token         string    `json:"token"`
	BlockHeight   uint64    `gorm:"not null" json:"block_height"`
	Confirmations uint64    `gorm:"default:0" json:"confirmations"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}
