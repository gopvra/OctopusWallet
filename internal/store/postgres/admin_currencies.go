package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

func (s *Store) ListAllCurrencies(ctx context.Context) ([]models.SupportedCurrency, error) {
	var currencies []models.SupportedCurrency
	err := s.db.SelectContext(ctx, &currencies, "SELECT * FROM supported_currencies ORDER BY chain, symbol")
	if currencies == nil {
		currencies = []models.SupportedCurrency{}
	}
	return currencies, err
}
