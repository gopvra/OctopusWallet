package errcode

// Lang represents a supported language.
type Lang string

const (
	LangEN Lang = "en"
	LangZH Lang = "zh"
)

// messages maps error codes to localized messages.
// Add new languages by adding a new Lang key to the inner map.
var messages = map[Code]map[Lang]string{
	OK: {LangEN: "ok", LangZH: "成功"},

	// General
	ErrBadRequest:       {LangEN: "invalid request parameters", LangZH: "请求参数无效"},
	ErrUnauthorized:     {LangEN: "authentication required", LangZH: "需要认证"},
	ErrForbidden:        {LangEN: "access denied", LangZH: "拒绝访问"},
	ErrNotFound:         {LangEN: "resource not found", LangZH: "资源未找到"},
	ErrRateLimited:      {LangEN: "rate limit exceeded, please try later", LangZH: "请求过于频繁，请稍后再试"},
	ErrIdempotencyHit:   {LangEN: "duplicate request (idempotency)", LangZH: "重复请求（幂等性）"},
	ErrInternalServer:   {LangEN: "internal server error", LangZH: "服务器内部错误"},
	ErrInvalidSignature: {LangEN: "invalid request signature", LangZH: "请求签名无效"},
	ErrTimestampExpired: {LangEN: "request timestamp expired", LangZH: "请求时间戳已过期"},
	ErrIPNotWhitelisted: {LangEN: "IP address not whitelisted", LangZH: "IP 地址不在白名单中"},
	ErrInvalidUUID:      {LangEN: "invalid ID format", LangZH: "ID 格式无效"},
	ErrInvalidAmount:    {LangEN: "invalid amount", LangZH: "金额无效"},
	ErrInvalidAddress:   {LangEN: "invalid blockchain address", LangZH: "区块链地址无效"},
	ErrInvalidURL:       {LangEN: "invalid URL format", LangZH: "URL 格式无效"},
	ErrUnsupportedChain: {LangEN: "unsupported blockchain", LangZH: "不支持的区块链"},

	// Payment
	ErrPaymentNotFound:      {LangEN: "payment not found", LangZH: "支付订单未找到"},
	ErrPaymentExpired:       {LangEN: "payment has expired", LangZH: "支付订单已过期"},
	ErrPaymentNotCompleted:  {LangEN: "payment is not completed", LangZH: "支付订单未完成"},
	ErrPaymentCreateFailed:  {LangEN: "failed to create payment", LangZH: "创建支付订单失败"},
	ErrPaymentAlreadyExists: {LangEN: "payment already exists", LangZH: "支付订单已存在"},
	ErrPaymentLinkNotFound:  {LangEN: "payment link not found", LangZH: "支付链接未找到"},
	ErrPaymentLinkInactive:  {LangEN: "payment link is inactive", LangZH: "支付链接已停用"},

	// Payout
	ErrPayoutNotFound:           {LangEN: "payout not found", LangZH: "打款订单未找到"},
	ErrPayoutCreateFailed:       {LangEN: "failed to create payout", LangZH: "创建打款订单失败"},
	ErrPayoutNotPendingApproval: {LangEN: "payout is not pending approval", LangZH: "打款订单不在待审批状态"},
	ErrPayoutNotOwnedByMerchant: {LangEN: "payout does not belong to this merchant", LangZH: "打款订单不属于此商户"},
	ErrPayoutExceedsSingleLimit: {LangEN: "amount exceeds single transaction limit", LangZH: "金额超过单笔交易限额"},
	ErrPayoutExceedsDailyLimit:  {LangEN: "amount would exceed daily limit", LangZH: "金额将超过每日限额"},
	ErrPayoutApprovalFailed:     {LangEN: "failed to approve payout", LangZH: "审批打款失败"},
	ErrPayoutRejectFailed:       {LangEN: "failed to reject payout", LangZH: "拒绝打款失败"},
	ErrBatchCreateFailed:        {LangEN: "failed to create batch payout", LangZH: "创建批量打款失败"},
	ErrBatchNotFound:            {LangEN: "batch payout not found", LangZH: "批量打款未找到"},
	ErrBatchPartialFailure:      {LangEN: "some batch items failed to create", LangZH: "部分批量打款项创建失败"},

	// Refund
	ErrRefundNotFound:       {LangEN: "refund not found", LangZH: "退款未找到"},
	ErrRefundCreateFailed:   {LangEN: "failed to create refund", LangZH: "创建退款失败"},
	ErrRefundExceedsPayment: {LangEN: "refund total would exceed payment amount", LangZH: "退款总额将超过支付金额"},
	ErrRefundOnlyCompleted:  {LangEN: "can only refund completed payments", LangZH: "只能退款已完成的支付"},

	// Merchant / Wallet / Config
	ErrMerchantNotFound:       {LangEN: "merchant not found", LangZH: "商户未找到"},
	ErrMerchantCreateFailed:   {LangEN: "failed to create merchant", LangZH: "创建商户失败"},
	ErrMerchantInactive:       {LangEN: "merchant is inactive", LangZH: "商户已停用"},
	ErrWalletCreateFailed:     {LangEN: "failed to create wallet", LangZH: "创建钱包失败"},
	ErrDerivationIndexFailed:  {LangEN: "failed to get derivation index", LangZH: "获取派生索引失败"},
	ErrDeriveAddressFailed:    {LangEN: "failed to derive address", LangZH: "派生地址失败"},
	ErrSweepConfigFailed:      {LangEN: "failed to save sweep config", LangZH: "保存归集配置失败"},
	ErrColdWalletConfigFailed: {LangEN: "failed to save cold wallet config", LangZH: "保存冷钱包配置失败"},
	ErrApprovalConfigInvalid:  {LangEN: "invalid approval configuration", LangZH: "审批配置无效"},
	ErrApprovalConfigFailed:   {LangEN: "failed to save approval config", LangZH: "保存审批配置失败"},
	ErrWebhookURLInvalid:      {LangEN: "invalid webhook URL", LangZH: "Webhook URL 无效"},
	ErrWebhookURLPrivate:      {LangEN: "webhook URL must not point to private addresses", LangZH: "Webhook URL 不能指向内网地址"},

	// Admin
	ErrAdminLoginFailed:      {LangEN: "invalid username or password", LangZH: "用户名或密码错误"},
	ErrAdminTokenInvalid:     {LangEN: "invalid or expired token", LangZH: "令牌无效或已过期"},
	ErrAdminTokenExpired:     {LangEN: "token has expired", LangZH: "令牌已过期"},
	ErrAdminUserNotFound:     {LangEN: "admin user not found", LangZH: "管理员用户未找到"},
	ErrAdminUserCreateFailed: {LangEN: "failed to create admin user", LangZH: "创建管理员用户失败"},
	ErrAdminUserUpdateFailed: {LangEN: "failed to update admin user", LangZH: "更新管理员用户失败"},
	ErrAdminUserDeleteFailed: {LangEN: "failed to delete admin user", LangZH: "删除管理员用户失败"},
	ErrAdminCannotDeleteSelf: {LangEN: "cannot delete yourself", LangZH: "不能删除自己"},
	ErrAdminLastSuperAdmin:   {LangEN: "cannot delete the last super admin", LangZH: "不能删除最后一个超级管理员"},
	ErrAdminPasswordTooShort: {LangEN: "password must be 12-128 characters", LangZH: "密码长度必须在 12-128 个字符之间"},
	ErrAdminInsufficientRole: {LangEN: "insufficient permissions", LangZH: "权限不足"},
	ErrAdminRefreshFailed:    {LangEN: "failed to refresh token", LangZH: "刷新令牌失败"},
	ErrAdminUserDeactivated:  {LangEN: "user account is deactivated", LangZH: "用户账号已停用"},
}

// Msg returns the localized message for a code. Falls back to English if not found.
func (c Code) Msg(lang Lang) string {
	if m, ok := messages[c]; ok {
		if msg, ok := m[lang]; ok {
			return msg
		}
		if msg, ok := m[LangEN]; ok {
			return msg
		}
	}
	return "unknown error"
}

// Int returns the integer value.
func (c Code) Int() int {
	return int(c)
}
