package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListRefunds(ctx context.Context, filter store.RefundFilter) (*store.PaginatedResult[models.Refund], error) {
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
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(to_address ILIKE $%d ESCAPE '\\' OR id::text ILIKE $%d ESCAPE '\\')", argIdx, argIdx))
		args = append(args, "%"+store.EscapeSearch(filter.Search)+"%")
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := s.db.GetContext(ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM refunds %s", where), args...); err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "status": true, "chain": true, "amount": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	query := fmt.Sprintf("SELECT * FROM refunds %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sortCol, filter.Order, argIdx, argIdx+1)
	args = append(args, filter.PerPage, filter.Offset())

	var refunds []models.Refund
	if err := s.db.SelectContext(ctx, &refunds, query, args...); err != nil {
		return nil, err
	}
	if refunds == nil {
		refunds = []models.Refund{}
	}

	totalPages := (total + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Refund]{
		Data: refunds,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: total, TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetRefundByID(ctx context.Context, id string) (*models.Refund, error) {
	var r models.Refund
	err := s.db.GetContext(ctx, &r, "SELECT * FROM refunds WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
