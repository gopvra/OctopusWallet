package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListAllMerchantBalances(ctx context.Context, filter store.BalanceFilter) ([]models.MerchantBalance, error) {
	query := s.db.WithContext(ctx).Model(&models.MerchantBalance{})

	if filter.MerchantID != "" {
		query = query.Where("merchant_id = ?", filter.MerchantID)
	}
	if filter.Chain != "" {
		query = query.Where("chain = ?", filter.Chain)
	}

	var balances []models.MerchantBalance
	err := query.Order("merchant_id, chain").Find(&balances).Error
	if balances == nil {
		balances = []models.MerchantBalance{}
	}
	return balances, err
}
