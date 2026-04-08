package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListBatchPayouts(ctx context.Context, filter store.BatchPayoutFilter) (*store.PaginatedResult[models.BatchPayout], error) {
	filter.Normalize()

	query := s.db.WithContext(ctx).Model(&models.BatchPayout{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Chain != "" {
		query = query.Where("chain = ?", filter.Chain)
	}
	if filter.MerchantID != "" {
		query = query.Where("merchant_id = ?", filter.MerchantID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "status": true, "chain": true, "total_amount": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	var batches []models.BatchPayout
	if err := query.Order(sortCol + " " + filter.Order).
		Limit(filter.PerPage).Offset(filter.Offset()).
		Find(&batches).Error; err != nil {
		return nil, err
	}
	if batches == nil {
		batches = []models.BatchPayout{}
	}

	totalPages := (int(total) + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.BatchPayout]{
		Data: batches,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: int(total), TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetBatchPayoutByID(ctx context.Context, id string) (*models.BatchPayout, error) {
	var b models.BatchPayout
	if err := s.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) AdminGetBatchPayoutItems(ctx context.Context, batchID string) ([]models.BatchPayoutItem, error) {
	var items []models.BatchPayoutItem
	err := s.db.WithContext(ctx).Where("batch_id = ?", batchID).Order("created_at").Find(&items).Error
	if items == nil {
		items = []models.BatchPayoutItem{}
	}
	return items, err
}
