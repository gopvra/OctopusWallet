package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListBatchPayouts(ctx context.Context, filter store.BatchPayoutFilter) (*store.PaginatedResult[models.BatchPayout], error) {
	filter.Normalize()

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Chain != "" {
		conditions = append(conditions, fmt.Sprintf("chain = $%d", argIdx))
		args = append(args, filter.Chain)
		argIdx++
	}
	if filter.MerchantID != "" {
		conditions = append(conditions, fmt.Sprintf("merchant_id = $%d", argIdx))
		args = append(args, filter.MerchantID)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := s.db.GetContext(ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM batch_payouts %s", where), args...); err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "status": true, "chain": true, "total_amount": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	query := fmt.Sprintf("SELECT * FROM batch_payouts %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sortCol, filter.Order, argIdx, argIdx+1)
	args = append(args, filter.PerPage, filter.Offset())

	var batches []models.BatchPayout
	if err := s.db.SelectContext(ctx, &batches, query, args...); err != nil {
		return nil, err
	}
	if batches == nil {
		batches = []models.BatchPayout{}
	}

	totalPages := (total + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.BatchPayout]{
		Data: batches,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: total, TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetBatchPayoutByID(ctx context.Context, id string) (*models.BatchPayout, error) {
	var b models.BatchPayout
	err := s.db.GetContext(ctx, &b, "SELECT * FROM batch_payouts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) AdminGetBatchPayoutItems(ctx context.Context, batchID string) ([]models.BatchPayoutItem, error) {
	var items []models.BatchPayoutItem
	err := s.db.SelectContext(ctx, &items, "SELECT * FROM batch_payout_items WHERE batch_id = $1 ORDER BY created_at", batchID)
	if items == nil {
		items = []models.BatchPayoutItem{}
	}
	return items, err
}
