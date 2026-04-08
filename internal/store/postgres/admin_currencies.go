package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

func (s *Store) ListAllCurrencies(ctx context.Context) ([]models.SupportedCurrency, error) {
	var currencies []models.SupportedCurrency
	err := s.db.WithContext(ctx).Order("chain, symbol").Find(&currencies).Error
	if currencies == nil {
		currencies = []models.SupportedCurrency{}
	}
	return currencies, err
}
