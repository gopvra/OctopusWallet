package models

import "time"

type Merchant struct {
	ID         string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name       string    `gorm:"not null" json:"name"`
	Email      string    `gorm:"uniqueIndex;not null" json:"email"`
	APIKeyHash string    `gorm:"column:api_key_hash;not null" json:"-"`
	WebhookURL string    `gorm:"column:webhook_url" json:"webhook_url"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
