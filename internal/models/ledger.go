package models

import "time"

type MerchantBalance struct {
	ID         string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID string    `gorm:"type:uuid;uniqueIndex:idx_mb_merchant_chain_token;not null" json:"merchant_id"`
	Chain      string    `gorm:"uniqueIndex:idx_mb_merchant_chain_token;not null" json:"chain"`
	Token      string    `gorm:"uniqueIndex:idx_mb_merchant_chain_token;not null" json:"token"`
	Available  string    `gorm:"default:'0'" json:"available"`
	Pending    string    `gorm:"default:'0'" json:"pending"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type SupportedCurrency struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Chain        string    `gorm:"not null;index" json:"chain"`
	Symbol       string    `gorm:"not null" json:"symbol"`
	Name         string    `gorm:"not null" json:"name"`
	TokenAddress string    `json:"token_address"`
	Decimals     int       `gorm:"not null" json:"decimals"`
	IsNative     bool      `gorm:"default:false" json:"is_native"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	MinAmount    string    `gorm:"default:'0'" json:"min_amount"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type BatchPayout struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID     string    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Chain          string    `gorm:"not null" json:"chain"`
	Token          string    `json:"token"`
	TotalAmount    string    `gorm:"not null" json:"total_amount"`
	TotalCount     int       `gorm:"not null" json:"total_count"`
	CompletedCount int       `gorm:"default:0" json:"completed_count"`
	FailedCount    int       `gorm:"default:0" json:"failed_count"`
	Status         string    `gorm:"default:'pending';index" json:"status"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type BatchPayoutItem struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	BatchID      string    `gorm:"type:uuid;index;not null" json:"batch_id"`
	PayoutID     *string   `gorm:"type:uuid" json:"payout_id,omitempty"`
	ToAddress    string    `gorm:"not null" json:"to_address"`
	Amount       string    `gorm:"not null" json:"amount"`
	Status       string    `gorm:"default:'pending'" json:"status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type MerchantIPWhitelist struct {
	ID         string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MerchantID string `gorm:"type:uuid;index;not null" json:"merchant_id"`
	IPAddress  string `gorm:"column:ip_address;not null" json:"ip_address"`
}

func (MerchantIPWhitelist) TableName() string {
	return "merchant_ip_whitelist"
}
