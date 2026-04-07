package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListRefunds(ctx context.Context, filter store.RefundFilter) (*store.PaginatedResult[models.Refund], error) {
	filter.Normalize()

	query := s.db.WithContext(ctx).Model(&models.Refund{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Chain != "" {
		query = query.Where("chain = ?", filter.Chain)
	}
	if filter.MerchantID != "" {
		query = query.Where("merchant_id = ?", filter.MerchantID)
	}
	if filter.Search != "" {
		pattern := "%" + store.EscapeSearch(filter.Search) + "%"
		query = query.Where("(to_address ILIKE ? ESCAPE '\\' OR id::text ILIKE ? ESCAPE '\\')", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "status": true, "chain": true, "amount": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	var refunds []models.Refund
	if err := query.Order(sortCol + " " + filter.Order).
		Limit(filter.PerPage).Offset(filter.Offset()).
		Find(&refunds).Error; err != nil {
		return nil, err
	}
	if refunds == nil {
		refunds = []models.Refund{}
	}

	totalPages := (int(total) + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Refund]{
		Data: refunds,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: int(total), TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetRefundByID(ctx context.Context, id string) (*models.Refund, error) {
	var r models.Refund
	if err := s.db.WithContext(ctx).First(&r, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}
