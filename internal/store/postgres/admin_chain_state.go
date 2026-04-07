package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

func (s *Store) ListChainStates(ctx context.Context) ([]models.ChainState, error) {
	var states []models.ChainState
	err := s.db.SelectContext(ctx, &states, "SELECT * FROM chain_state ORDER BY chain")
	if states == nil {
		states = []models.ChainState{}
	}
	return states, err
}
