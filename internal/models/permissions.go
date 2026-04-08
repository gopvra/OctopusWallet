package models

type Permission string

const (
	// Dashboard
	PermDashboardView Permission = "dashboard:view"

	// Merchants
	PermMerchantList   Permission = "merchant:list"
	PermMerchantView   Permission = "merchant:view"
	PermMerchantUpdate Permission = "merchant:update"
	PermMerchantToggle Permission = "merchant:toggle"

	// Payments
	PermPaymentList Permission = "payment:list"
	PermPaymentView Permission = "payment:view"

	// Payouts
	PermPayoutList Permission = "payout:list"
	PermPayoutView Permission = "payout:view"

	// Refunds
	PermRefundList Permission = "refund:list"
	PermRefundView Permission = "refund:view"

	// Batch Payouts
	PermBatchPayoutList Permission = "batch_payout:list"
	PermBatchPayoutView Permission = "batch_payout:view"

	// Wallets
	PermWalletList Permission = "wallet:list"

	// Balances
	PermBalanceList Permission = "balance:list"

	// Currencies
	PermCurrencyList Permission = "currency:list"

	// Chain State
	PermChainStateList Permission = "chain_state:list"

	// Admin Users
	PermAdminUserList   Permission = "admin_user:list"
	PermAdminUserCreate Permission = "admin_user:create"
	PermAdminUserUpdate Permission = "admin_user:update"
	PermAdminUserDelete Permission = "admin_user:delete"

	// Settings
	PermSettingsView   Permission = "settings:view"
	PermSettingsUpdate Permission = "settings:update"
)

// RolePermissions defines what each role can do.
// super_admin has ALL permissions.
// admin has read-only access to most things.
// viewer has dashboard + basic read access only.
var RolePermissions = map[string][]Permission{
	RoleSuperAdmin: {
		// super_admin gets everything
		PermDashboardView,
		PermMerchantList, PermMerchantView, PermMerchantUpdate, PermMerchantToggle,
		PermPaymentList, PermPaymentView,
		PermPayoutList, PermPayoutView,
		PermRefundList, PermRefundView,
		PermBatchPayoutList, PermBatchPayoutView,
		PermWalletList,
		PermBalanceList,
		PermCurrencyList,
		PermChainStateList,
		PermAdminUserList, PermAdminUserCreate, PermAdminUserUpdate, PermAdminUserDelete,
		PermSettingsView, PermSettingsUpdate,
	},
	RoleAdmin: {
		// admin: read-only on most resources, no user management, no merchant toggle
		PermDashboardView,
		PermMerchantList, PermMerchantView,
		PermPaymentList, PermPaymentView,
		PermPayoutList, PermPayoutView,
		PermRefundList, PermRefundView,
		PermBatchPayoutList, PermBatchPayoutView,
		PermWalletList,
		PermBalanceList,
		PermCurrencyList,
		PermChainStateList,
		PermSettingsView,
	},
	RoleViewer: {
		// viewer: dashboard + basic read
		PermDashboardView,
		PermPaymentList,
		PermPayoutList,
		PermMerchantList,
	},
}

const RoleViewer = "viewer"

// HasPermission checks if a role has a specific permission.
func HasPermission(role string, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
