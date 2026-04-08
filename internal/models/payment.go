package models

import "time"

const (
	PaymentStatusPending    = "pending"
	PaymentStatusConfirming = "confirming"
	PaymentStatusCompleted  = "completed"
	PaymentStatusExpired    = "expired"
)

type Payment struct {
	ID             string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID     string     `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Chain          string     `gorm:"not null" json:"chain"`
	Token          string     `json:"token"`
	AmountExpected string     `gorm:"not null" json:"amount_expected"`
	AmountReceived string     `gorm:"default:'0'" json:"amount_received"`
	Address        string     `gorm:"not null;index" json:"address"`
	Status         string     `gorm:"default:'pending';index" json:"status"`
	Currency       string     `json:"currency"`
	Description    string     `json:"description"`
	OrderID        string     `gorm:"column:order_id" json:"order_id,omitempty"`
	RedirectURL    string     `gorm:"column:redirect_url" json:"redirect_url,omitempty"`
	TxHash         *string    `gorm:"column:tx_hash" json:"tx_hash,omitempty"`
	Confirmations  int        `gorm:"default:0" json:"confirmations"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
