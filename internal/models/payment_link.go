package models

import "time"

type PaymentLink struct {
	ID          string    `db:"id" json:"id"`
	MerchantID  string    `db:"merchant_id" json:"merchant_id"`
	Chain       string    `db:"chain" json:"chain"`
	Token       string    `db:"token" json:"token"`
	Amount      string    `db:"amount" json:"amount"`
	Currency    string    `db:"currency" json:"currency"`
	Description string    `db:"description" json:"description"`
	RedirectURL string    `db:"redirect_url" json:"redirect_url"`
	IsReusable  bool      `db:"is_reusable" json:"is_reusable"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	UsesCount   int       `db:"uses_count" json:"uses_count"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type AuditLog struct {
	ID           string    `db:"id" json:"id"`
	MerchantID   string    `db:"merchant_id" json:"merchant_id"`
	Action       string    `db:"action" json:"action"`
	ResourceType string    `db:"resource_type" json:"resource_type"`
	ResourceID   string    `db:"resource_id" json:"resource_id"`
	IPAddress    string    `db:"ip_address" json:"ip_address"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
