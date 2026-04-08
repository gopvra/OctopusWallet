package models

import "time"

// WebAuthnCredential stores a user's registered passkey/biometric credential.
type WebAuthnCredential struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID          string    `gorm:"type:uuid;index;not null" json:"user_id"`
	CredentialID    []byte    `gorm:"not null;uniqueIndex" json:"-"`
	PublicKey       []byte    `gorm:"not null" json:"-"`
	AttestationType string    `json:"attestation_type"`
	AAGUID          []byte    `json:"-"`
	SignCount       uint32    `gorm:"default:0" json:"-"`
	Transport       string    `json:"transport"`
	DisplayName     string    `json:"display_name"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}
