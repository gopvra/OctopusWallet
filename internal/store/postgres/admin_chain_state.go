package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

func (s *Store) ListChainStates(ctx context.Context) ([]models.ChainState, error) {
	var states []models.ChainState
	err := s.db.WithContext(ctx).Order("chain").Find(&states).Error
	if states == nil {
		states = []models.ChainState{}
	}
	return states, err
}
