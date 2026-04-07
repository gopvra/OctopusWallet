package models

import "time"

type Merchant struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Email      string    `db:"email" json:"email"`
	APIKeyHash string    `db:"api_key_hash" json:"-"`
	WebhookURL string    `db:"webhook_url" json:"webhook_url"`
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
