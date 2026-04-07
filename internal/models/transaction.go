package models

import "time"

type Transaction struct {
	ID            string    `db:"id" json:"id"`
	Chain         string    `db:"chain" json:"chain"`
	TxHash        string    `db:"tx_hash" json:"tx_hash"`
	FromAddress   string    `db:"from_address" json:"from_address"`
	ToAddress     string    `db:"to_address" json:"to_address"`
	Amount        string    `db:"amount" json:"amount"`
	Token         string    `db:"token" json:"token"`
	BlockHeight   uint64    `db:"block_height" json:"block_height"`
	Confirmations uint64    `db:"confirmations" json:"confirmations"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
