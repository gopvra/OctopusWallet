package postgres

import (
	"context"

	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

func (s *Store) ListWallets(ctx context.Context, filter store.WalletFilter) (*store.PaginatedResult[models.Wallet], error) {
	filter.Normalize()

	query := s.db.WithContext(ctx).Model(&models.Wallet{})

	if filter.Chain != "" {
		query = query.Where("chain = ?", filter.Chain)
	}
	if filter.MerchantID != "" {
		query = query.Where("merchant_id = ?", filter.MerchantID)
	}
	if filter.Search != "" {
		pattern := "%" + store.EscapeSearch(filter.Search) + "%"
		query = query.Where("address ILIKE ? ESCAPE '\\'", pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "created_at"
	allowed := map[string]bool{"created_at": true, "chain": true, "address": true}
	if allowed[filter.Sort] {
		sortCol = filter.Sort
	}

	var wallets []models.Wallet
	if err := query.Order(sortCol + " " + filter.Order).
		Limit(filter.PerPage).Offset(filter.Offset()).
		Find(&wallets).Error; err != nil {
		return nil, err
	}
	if wallets == nil {
		wallets = []models.Wallet{}
	}

	totalPages := (int(total) + filter.PerPage - 1) / filter.PerPage
	return &store.PaginatedResult[models.Wallet]{
		Data: wallets,
		Meta: store.Meta{Page: filter.Page, PerPage: filter.PerPage, Total: int(total), TotalPages: totalPages},
	}, nil
}
