package postgres

import (
	"context"
	"database/sql"

	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) GetDashboardStats(ctx context.Context) (*store.DashboardStats, error) {
	stats := &store.DashboardStats{}

	queries := []struct {
		query string
		dest  interface{}
	}{
		{"SELECT COUNT(*) FROM merchants", &stats.TotalMerchants},
		{"SELECT COUNT(*) FROM merchants WHERE is_active = true", &stats.ActiveMerchants},
		{"SELECT COUNT(*) FROM payments", &stats.TotalPayments},
		{"SELECT COUNT(*) FROM payouts", &stats.TotalPayouts},
		{"SELECT COUNT(*) FROM payments WHERE status IN ('pending', 'confirming')", &stats.PendingPayments},
		{"SELECT COUNT(*) FROM payouts WHERE status = 'pending'", &stats.PendingPayouts},
	}

	for _, q := range queries {
		if err := s.db.GetContext(ctx, q.dest, q.query); err != nil {
			return nil, err
		}
	}

	var volume sql.NullString
	err := s.db.GetContext(ctx, &volume,
		"SELECT COALESCE(SUM(amount_received::numeric), 0)::text FROM payments WHERE status = 'completed'")
	if err != nil {
		return nil, err
	}
	if volume.Valid {
		stats.TotalVolume = volume.String
	} else {
		stats.TotalVolume = "0"
	}

	return stats, nil
}

func (s *Store) GetVolumeChart(ctx context.Context, days int) ([]store.VolumePoint, error) {
	query := `
		SELECT
			date_trunc('day', created_at)::date::text AS date,
			COUNT(*) AS count,
			COALESCE(SUM(amount_received::numeric), 0)::text AS volume
		FROM payments
		WHERE created_at >= now() - make_interval(days => $1)
		GROUP BY date_trunc('day', created_at)::date
		ORDER BY date`

	var points []store.VolumePoint
	err := s.db.SelectContext(ctx, &points, query, days)
	if points == nil {
		points = []store.VolumePoint{}
	}
	return points, err
}

func (s *Store) GetChainDistribution(ctx context.Context) ([]store.ChainDistribution, error) {
	query := `
		SELECT
			chain,
			COUNT(*) AS count,
			COALESCE(SUM(amount_received::numeric), 0)::text AS volume
		FROM payments
		GROUP BY chain
		ORDER BY count DESC`

	var dist []store.ChainDistribution
	err := s.db.SelectContext(ctx, &dist, query)
	if dist == nil {
		dist = []store.ChainDistribution{}
	}
	return dist, err
}

func (s *Store) GetRecentActivity(ctx context.Context, limit int) ([]store.RecentActivity, error) {
	query := `
		(SELECT id, 'payment' AS type, chain, amount_expected AS amount, status, created_at
		 FROM payments ORDER BY created_at DESC LIMIT $1)
		UNION ALL
		(SELECT id, 'payout' AS type, chain, amount, status, created_at
		 FROM payouts ORDER BY created_at DESC LIMIT $1)
		ORDER BY created_at DESC LIMIT $1`

	var activity []store.RecentActivity
	err := s.db.SelectContext(ctx, &activity, query, limit)
	if activity == nil {
		activity = []store.RecentActivity{}
	}
	return activity, err
}
