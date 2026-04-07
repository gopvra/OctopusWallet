package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListPayments(ctx context.Context, filter store.PaymentFilter) (*store.PaginatedResult[models.Payment], error) {
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
	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, filter.DateTo)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(address ILIKE $%d ESCAPE '\\' OR id::text ILIKE $%d ESCAPE '\\')", argIdx, argIdx))
		args = append(args, "%"+store.EscapeSearch(filter.Search)+"%")
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := s.db.GetContext(ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM payments %s", where), args...); err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "updated_at": true, "status": true, "chain": true, "amount_expected": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	query := fmt.Sprintf("SELECT * FROM payments %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sortCol, filter.Order, argIdx, argIdx+1)
	args = append(args, filter.PerPage, filter.Offset())

	var payments []models.Payment
	if err := s.db.SelectContext(ctx, &payments, query, args...); err != nil {
		return nil, err
	}
	if payments == nil {
		payments = []models.Payment{}
	}

	totalPages := (total + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Payment]{
		Data: payments,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: total, TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetPaymentByID(ctx context.Context, id string) (*models.Payment, error) {
	var p models.Payment
	err := s.db.GetContext(ctx, &p, "SELECT * FROM payments WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
