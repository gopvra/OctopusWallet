package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

func (s *Store) CreateWebAuthnCredential(ctx context.Context, cred *models.WebAuthnCredential) error {
	return s.db.WithContext(ctx).Create(cred).Error
}

func (s *Store) GetWebAuthnCredentialsByUserID(ctx context.Context, userID string) ([]models.WebAuthnCredential, error) {
	var creds []models.WebAuthnCredential
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&creds).Error
	return creds, err
}

func (s *Store) GetWebAuthnCredentialByCredentialID(ctx context.Context, credentialID []byte) (*models.WebAuthnCredential, error) {
	var cred models.WebAuthnCredential
	err := s.db.WithContext(ctx).Where("credential_id = ?", credentialID).First(&cred).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (s *Store) UpdateWebAuthnCredentialSignCount(ctx context.Context, credentialID []byte, signCount uint32) error {
	return s.db.WithContext(ctx).Model(&models.WebAuthnCredential{}).
		Where("credential_id = ?", credentialID).
		Update("sign_count", signCount).Error
}

func (s *Store) DeleteWebAuthnCredential(ctx context.Context, id, userID string) error {
	return s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.WebAuthnCredential{}).Error
}
