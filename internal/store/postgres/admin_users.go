package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

func (s *Store) CreateAdminUser(ctx context.Context, user *models.AdminUser) error {
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *Store) GetAdminUserByID(ctx context.Context, id string) (*models.AdminUser, error) {
	var user models.AdminUser
	if err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetAdminUserByUsername(ctx context.Context, username string) (*models.AdminUser, error) {
	var user models.AdminUser
	if err := s.db.WithContext(ctx).Where("username = ? AND is_active = true", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) ListAdminUsers(ctx context.Context) ([]models.AdminUser, error) {
	var users []models.AdminUser
	err := s.db.WithContext(ctx).Order("created_at DESC").Find(&users).Error
	return users, err
}

func (s *Store) UpdateAdminUser(ctx context.Context, user *models.AdminUser) error {
	return s.db.WithContext(ctx).Save(user).Error
}

func (s *Store) DeleteAdminUser(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.AdminUser{}, "id = ?", id).Error
}

func (s *Store) CountAdminUsers(ctx context.Context) (int, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.AdminUser{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
