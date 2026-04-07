package errcode

// Code represents a business error code.
type Code int

// Success
const OK Code = 0

// 1xxxx — General / Auth / Validation
const (
	ErrBadRequest       Code = 10001
	ErrUnauthorized     Code = 10002
	ErrForbidden        Code = 10003
	ErrNotFound         Code = 10004
	ErrRateLimited      Code = 10005
	ErrIdempotencyHit   Code = 10006
	ErrInternalServer   Code = 10007
	ErrInvalidSignature Code = 10008
	ErrTimestampExpired Code = 10009
	ErrIPNotWhitelisted Code = 10010
	ErrInvalidUUID      Code = 10011
	ErrInvalidAmount    Code = 10012
	ErrInvalidAddress   Code = 10013
	ErrInvalidURL       Code = 10014
	ErrUnsupportedChain Code = 10015
)

// 2xxxx — Payment
const (
	ErrPaymentNotFound      Code = 20001
	ErrPaymentExpired       Code = 20002
	ErrPaymentNotCompleted  Code = 20003
	ErrPaymentCreateFailed  Code = 20004
	ErrPaymentAlreadyExists Code = 20005
	ErrPaymentLinkNotFound  Code = 20006
	ErrPaymentLinkInactive  Code = 20007
)

// 3xxxx — Payout
const (
	ErrPayoutNotFound           Code = 30001
	ErrPayoutCreateFailed       Code = 30002
	ErrPayoutNotPendingApproval Code = 30003
	ErrPayoutNotOwnedByMerchant Code = 30004
	ErrPayoutExceedsSingleLimit Code = 30005
	ErrPayoutExceedsDailyLimit  Code = 30006
	ErrPayoutApprovalFailed     Code = 30007
	ErrPayoutRejectFailed       Code = 30008
	ErrBatchCreateFailed        Code = 30009
	ErrBatchNotFound            Code = 30010
	ErrBatchPartialFailure      Code = 30011
)

// 4xxxx — Refund
const (
	ErrRefundNotFound       Code = 40001
	ErrRefundCreateFailed   Code = 40002
	ErrRefundExceedsPayment Code = 40003
	ErrRefundOnlyCompleted  Code = 40004
)

// 5xxxx — Merchant / Wallet / Sweep / Config
const (
	ErrMerchantNotFound       Code = 50001
	ErrMerchantCreateFailed   Code = 50002
	ErrMerchantInactive       Code = 50003
	ErrWalletCreateFailed     Code = 50004
	ErrDerivationIndexFailed  Code = 50005
	ErrDeriveAddressFailed    Code = 50006
	ErrSweepConfigFailed      Code = 50007
	ErrColdWalletConfigFailed Code = 50008
	ErrApprovalConfigInvalid  Code = 50009
	ErrApprovalConfigFailed   Code = 50010
	ErrWebhookURLInvalid      Code = 50011
	ErrWebhookURLPrivate      Code = 50012
)

// 6xxxx — Admin
const (
	ErrAdminLoginFailed         Code = 60001
	ErrAdminTokenInvalid        Code = 60002
	ErrAdminTokenExpired        Code = 60003
	ErrAdminUserNotFound        Code = 60004
	ErrAdminUserCreateFailed    Code = 60005
	ErrAdminUserUpdateFailed    Code = 60006
	ErrAdminUserDeleteFailed    Code = 60007
	ErrAdminCannotDeleteSelf    Code = 60008
	ErrAdminLastSuperAdmin      Code = 60009
	ErrAdminPasswordTooShort    Code = 60010
	ErrAdminInsufficientRole    Code = 60011
	ErrAdminRefreshFailed       Code = 60012
	ErrAdminUserDeactivated     Code = 60013
)
