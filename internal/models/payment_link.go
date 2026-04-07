package models

import "time"

type PaymentLink struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID  string    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Chain       string    `gorm:"not null" json:"chain"`
	Token       string    `json:"token"`
	Amount      string    `gorm:"not null" json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	RedirectURL string    `json:"redirect_url"`
	IsReusable  bool      `gorm:"default:false" json:"is_reusable"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	UsesCount   int       `gorm:"default:0" json:"uses_count"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type AuditLog struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID   string    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Action       string    `gorm:"not null" json:"action"`
	ResourceType string    `gorm:"not null" json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}
