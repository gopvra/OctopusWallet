package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListPayments(ctx context.Context, filter store.PaymentFilter) (*store.PaginatedResult[models.Payment], error) {
	filter.Normalize()

	query := s.db.WithContext(ctx).Model(&models.Payment{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Chain != "" {
		query = query.Where("chain = ?", filter.Chain)
	}
	if filter.MerchantID != "" {
		query = query.Where("merchant_id = ?", filter.MerchantID)
	}
	if filter.DateFrom != "" {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		query = query.Where("created_at <= ?", filter.DateTo)
	}
	if filter.Search != "" {
		pattern := "%" + store.EscapeSearch(filter.Search) + "%"
		query = query.Where("(address ILIKE ? ESCAPE '\\' OR id::text ILIKE ? ESCAPE '\\')", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "updated_at": true, "status": true, "chain": true, "amount_expected": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	var payments []models.Payment
	if err := query.Order(sortCol + " " + filter.Order).
		Limit(filter.PerPage).Offset(filter.Offset()).
		Find(&payments).Error; err != nil {
		return nil, err
	}
	if payments == nil {
		payments = []models.Payment{}
	}

	totalPages := (int(total) + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Payment]{
		Data: payments,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: int(total), TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetPaymentByID(ctx context.Context, id string) (*models.Payment, error) {
	var p models.Payment
	if err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
