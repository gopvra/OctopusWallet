package models

import "time"

type Wallet struct {
	ID              string    `db:"id" json:"id"`
	MerchantID      string    `db:"merchant_id" json:"merchant_id"`
	Chain           string    `db:"chain" json:"chain"`
	Address         string    `db:"address" json:"address"`
	DerivationIndex int       `db:"derivation_index" json:"derivation_index"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}
