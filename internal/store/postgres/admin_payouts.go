package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListPayouts(ctx context.Context, filter store.PayoutFilter) (*store.PaginatedResult[models.Payout], error) {
	filter.Normalize()

	query := s.db.WithContext(ctx).Model(&models.Payout{})

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
		query = query.Where("(to_address ILIKE ? ESCAPE '\\' OR id::text ILIKE ? ESCAPE '\\')", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "updated_at": true, "status": true, "chain": true, "amount": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	var payouts []models.Payout
	if err := query.Order(sortCol + " " + filter.Order).
		Limit(filter.PerPage).Offset(filter.Offset()).
		Find(&payouts).Error; err != nil {
		return nil, err
	}
	if payouts == nil {
		payouts = []models.Payout{}
	}

	totalPages := (int(total) + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Payout]{
		Data: payouts,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: int(total), TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetPayoutByID(ctx context.Context, id string) (*models.Payout, error) {
	var p models.Payout
	if err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
