package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

func (s *Store) CreateAdminUser(ctx context.Context, user *models.AdminUser) error {
	query := `INSERT INTO admin_users (username, email, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_active, created_at, updated_at`
	return s.db.QueryRowxContext(ctx, query, user.Username, user.Email, user.Password, user.Role).
		Scan(&user.ID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
}

func (s *Store) GetAdminUserByID(ctx context.Context, id string) (*models.AdminUser, error) {
	var user models.AdminUser
	err := s.db.GetContext(ctx, &user, "SELECT * FROM admin_users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetAdminUserByUsername(ctx context.Context, username string) (*models.AdminUser, error) {
	var user models.AdminUser
	err := s.db.GetContext(ctx, &user, "SELECT * FROM admin_users WHERE username = $1 AND is_active = true", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) ListAdminUsers(ctx context.Context) ([]models.AdminUser, error) {
	var users []models.AdminUser
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM admin_users ORDER BY created_at DESC")
	return users, err
}

func (s *Store) UpdateAdminUser(ctx context.Context, user *models.AdminUser) error {
	query := `UPDATE admin_users SET username = $1, email = $2, role = $3, is_active = $4, updated_at = now()
		WHERE id = $5`
	_, err := s.db.ExecContext(ctx, query, user.Username, user.Email, user.Role, user.IsActive, user.ID)
	return err
}

func (s *Store) DeleteAdminUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM admin_users WHERE id = $1", id)
	return err
}

func (s *Store) CountAdminUsers(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM admin_users")
	return count, err
}
