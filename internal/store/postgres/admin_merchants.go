package postgres

import (
	"context"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListMerchants(ctx context.Context, filter store.MerchantFilter) (*store.PaginatedResult[models.Merchant], error) {
	filter.Normalize()

	query := s.db.WithContext(ctx).Model(&models.Merchant{})

	if filter.Search != "" {
		pattern := "%" + store.EscapeSearch(filter.Search) + "%"
		query = query.Where("(name ILIKE ? ESCAPE '\\' OR email ILIKE ? ESCAPE '\\')", pattern, pattern)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"name": true, "email": true, "created_at": true, "updated_at": true, "is_active": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	var merchants []models.Merchant
	if err := query.Order(sortCol + " " + filter.Order).
		Limit(filter.PerPage).Offset(filter.Offset()).
		Find(&merchants).Error; err != nil {
		return nil, err
	}
	if merchants == nil {
		merchants = []models.Merchant{}
	}

	totalPages := (int(total) + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Merchant]{
		Data: merchants,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: int(total), TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	var m models.Merchant
	if err := s.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) UpdateMerchant(ctx context.Context, id string, name, email, webhookURL string) error {
	return s.db.WithContext(ctx).Model(&models.Merchant{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        name,
			"email":       email,
			"webhook_url": webhookURL,
			"updated_at":  time.Now(),
		}).Error
}

func (s *Store) ToggleMerchantActive(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Exec(
		"UPDATE merchants SET is_active = NOT is_active, updated_at = now() WHERE id = ?", id).Error
}
