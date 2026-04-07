package models

import "time"

type Wallet struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID      string    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Chain           string    `gorm:"not null" json:"chain"`
	Address         string    `gorm:"not null;index" json:"address"`
	DerivationIndex int       `gorm:"not null" json:"derivation_index"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}
