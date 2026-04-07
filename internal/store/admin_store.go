package store

import (
	"context"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/models"
)

type PaginationParams struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Sort    string `form:"sort"`
	Order   string `form:"order"`
	Search  string `form:"search"`
}

func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 20
	}
	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc"
	}
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

type PaginatedResult[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type MerchantFilter struct {
	PaginationParams
	IsActive *bool `form:"is_active"`
}

type PaymentFilter struct {
	PaginationParams
	Status     string `form:"status"`
	Chain      string `form:"chain"`
	MerchantID string `form:"merchant_id"`
	DateFrom   string `form:"date_from"`
	DateTo     string `form:"date_to"`
}

type PayoutFilter struct {
	PaginationParams
	Status     string `form:"status"`
	Chain      string `form:"chain"`
	MerchantID string `form:"merchant_id"`
	DateFrom   string `form:"date_from"`
	DateTo     string `form:"date_to"`
}

type WalletFilter struct {
	PaginationParams
	Chain      string `form:"chain"`
	MerchantID string `form:"merchant_id"`
}

type DashboardStats struct {
	TotalMerchants  int    `json:"total_merchants"`
	ActiveMerchants int    `json:"active_merchants"`
	TotalPayments   int    `json:"total_payments"`
	TotalPayouts    int    `json:"total_payouts"`
	TotalVolume     string `json:"total_volume"`
	PendingPayments int    `json:"pending_payments"`
	PendingPayouts  int    `json:"pending_payouts"`
}

type VolumePoint struct {
	Date   string `json:"date" db:"date"`
	Count  int    `json:"count" db:"count"`
	Volume string `json:"volume" db:"volume"`
}

type ChainDistribution struct {
	Chain  string `json:"chain" db:"chain"`
	Count  int    `json:"count" db:"count"`
	Volume string `json:"volume" db:"volume"`
}

type RecentActivity struct {
	ID        string    `json:"id" db:"id"`
	Type      string    `json:"type"`
	Chain     string    `json:"chain" db:"chain"`
	Amount    string    `json:"amount" db:"amount"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AdminStore interface {
	// Admin Users
	CreateAdminUser(ctx context.Context, user *models.AdminUser) error
	GetAdminUserByID(ctx context.Context, id string) (*models.AdminUser, error)
	GetAdminUserByUsername(ctx context.Context, username string) (*models.AdminUser, error)
	ListAdminUsers(ctx context.Context) ([]models.AdminUser, error)
	UpdateAdminUser(ctx context.Context, user *models.AdminUser) error
	DeleteAdminUser(ctx context.Context, id string) error
	CountAdminUsers(ctx context.Context) (int, error)

	// Merchants
	ListMerchants(ctx context.Context, filter MerchantFilter) (*PaginatedResult[models.Merchant], error)
	AdminGetMerchantByID(ctx context.Context, id string) (*models.Merchant, error)
	UpdateMerchant(ctx context.Context, id string, name, email, webhookURL string) error
	ToggleMerchantActive(ctx context.Context, id string) error

	// Payments
	ListPayments(ctx context.Context, filter PaymentFilter) (*PaginatedResult[models.Payment], error)
	AdminGetPaymentByID(ctx context.Context, id string) (*models.Payment, error)

	// Payouts
	ListPayouts(ctx context.Context, filter PayoutFilter) (*PaginatedResult[models.Payout], error)
	AdminGetPayoutByID(ctx context.Context, id string) (*models.Payout, error)

	// Wallets
	ListWallets(ctx context.Context, filter WalletFilter) (*PaginatedResult[models.Wallet], error)

	// Chain State
	ListChainStates(ctx context.Context) ([]models.ChainState, error)

	// Dashboard
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetVolumeChart(ctx context.Context, days int) ([]VolumePoint, error)
	GetChainDistribution(ctx context.Context) ([]ChainDistribution, error)
	GetRecentActivity(ctx context.Context, limit int) ([]RecentActivity, error)
}
