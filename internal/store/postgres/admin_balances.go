package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListAllMerchantBalances(ctx context.Context, filter store.BalanceFilter) ([]models.MerchantBalance, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.MerchantID != "" {
		conditions = append(conditions, fmt.Sprintf("merchant_id = $%d", argIdx))
		args = append(args, filter.MerchantID)
		argIdx++
	}
	if filter.Chain != "" {
		conditions = append(conditions, fmt.Sprintf("chain = $%d", argIdx))
		args = append(args, filter.Chain)
		argIdx++
	}

	query := "SELECT * FROM merchant_balances"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY merchant_id, chain"

	var balances []models.MerchantBalance
	err := s.db.SelectContext(ctx, &balances, query, args...)
	if balances == nil {
		balances = []models.MerchantBalance{}
	}
	return balances, err
}
