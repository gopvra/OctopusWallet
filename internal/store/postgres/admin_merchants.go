package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListMerchants(ctx context.Context, filter store.MerchantFilter) (*store.PaginatedResult[models.Merchant], error) {
	filter.Normalize()

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d ESCAPE '\\' OR email ILIKE $%d ESCAPE '\\')", argIdx, argIdx))
		args = append(args, "%"+store.EscapeSearch(filter.Search)+"%")
		argIdx++
	}
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM merchants %s", where)
	if err := s.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	// Allowed sort columns
	sortCol := "created_at"
	allowed := map[string]bool{"name": true, "email": true, "created_at": true, "updated_at": true, "is_active": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	query := fmt.Sprintf("SELECT * FROM merchants %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sortCol, filter.Order, argIdx, argIdx+1)
	args = append(args, filter.PerPage, filter.Offset())

	var merchants []models.Merchant
	if err := s.db.SelectContext(ctx, &merchants, query, args...); err != nil {
		return nil, err
	}
	if merchants == nil {
		merchants = []models.Merchant{}
	}

	totalPages := (total + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Merchant]{
		Data: merchants,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: total, TotalPages: totalPages},
	}, nil
}

func (s *Store) AdminGetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	var m models.Merchant
	err := s.db.GetContext(ctx, &m, "SELECT * FROM merchants WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) UpdateMerchant(ctx context.Context, id string, name, email, webhookURL string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE merchants SET name = $1, email = $2, webhook_url = $3, updated_at = now() WHERE id = $4",
		name, email, webhookURL, id)
	return err
}

func (s *Store) ToggleMerchantActive(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE merchants SET is_active = NOT is_active, updated_at = now() WHERE id = $1", id)
	return err
}
