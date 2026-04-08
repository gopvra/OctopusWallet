package errcode

// Lang represents a supported language.
type Lang string

const (
	LangEN Lang = "en"
	LangZH Lang = "zh"
	LangJA Lang = "ja"
	LangKO Lang = "ko"
	LangES Lang = "es"
)

// messages maps error codes to localized messages.
var messages = map[Code]map[Lang]string{
	OK: {LangEN: "ok", LangZH: "成功", LangJA: "成功", LangKO: "성공", LangES: "éxito"},

	// General
	ErrBadRequest:       {LangEN: "invalid request parameters", LangZH: "请求参数无效", LangJA: "リクエストパラメータが無効です", LangKO: "잘못된 요청 매개변수", LangES: "parámetros de solicitud inválidos"},
	ErrUnauthorized:     {LangEN: "authentication required", LangZH: "需要认证", LangJA: "認証が必要です", LangKO: "인증이 필요합니다", LangES: "autenticación requerida"},
	ErrForbidden:        {LangEN: "access denied", LangZH: "拒绝访问", LangJA: "アクセス拒否", LangKO: "접근 거부", LangES: "acceso denegado"},
	ErrNotFound:         {LangEN: "resource not found", LangZH: "资源未找到", LangJA: "リソースが見つかりません", LangKO: "리소스를 찾을 수 없습니다", LangES: "recurso no encontrado"},
	ErrRateLimited:      {LangEN: "rate limit exceeded, please try later", LangZH: "请求过于频繁，请稍后再试", LangJA: "リクエスト制限を超えました。後でお試しください", LangKO: "요청 한도 초과, 나중에 다시 시도해주세요", LangES: "límite de solicitudes excedido, intente más tarde"},
	ErrIdempotencyHit:   {LangEN: "duplicate request (idempotency)", LangZH: "重复请求（幂等性）", LangJA: "重複リクエスト（冪等性）", LangKO: "중복 요청 (멱등성)", LangES: "solicitud duplicada (idempotencia)"},
	ErrInternalServer:   {LangEN: "internal server error", LangZH: "服务器内部错误", LangJA: "サーバー内部エラー", LangKO: "내부 서버 오류", LangES: "error interno del servidor"},
	ErrInvalidSignature: {LangEN: "invalid request signature", LangZH: "请求签名无效", LangJA: "リクエスト署名が無効です", LangKO: "잘못된 요청 서명", LangES: "firma de solicitud inválida"},
	ErrTimestampExpired: {LangEN: "request timestamp expired", LangZH: "请求时间戳已过期", LangJA: "リクエストのタイムスタンプが期限切れです", LangKO: "요청 타임스탬프 만료", LangES: "marca de tiempo de solicitud expirada"},
	ErrIPNotWhitelisted: {LangEN: "IP address not whitelisted", LangZH: "IP 地址不在白名单中", LangJA: "IPアドレスがホワイトリストにありません", LangKO: "IP 주소가 화이트리스트에 없습니다", LangES: "dirección IP no en lista blanca"},
	ErrInvalidUUID:      {LangEN: "invalid ID format", LangZH: "ID 格式无效", LangJA: "ID形式が無効です", LangKO: "잘못된 ID 형식", LangES: "formato de ID inválido"},
	ErrInvalidAmount:    {LangEN: "invalid amount", LangZH: "金额无效", LangJA: "金額が無効です", LangKO: "잘못된 금액", LangES: "monto inválido"},
	ErrInvalidAddress:   {LangEN: "invalid blockchain address", LangZH: "区块链地址无效", LangJA: "ブロックチェーンアドレスが無効です", LangKO: "잘못된 블록체인 주소", LangES: "dirección blockchain inválida"},
	ErrInvalidURL:       {LangEN: "invalid URL format", LangZH: "URL 格式无效", LangJA: "URL形式が無効です", LangKO: "잘못된 URL 형식", LangES: "formato de URL inválido"},
	ErrUnsupportedChain: {LangEN: "unsupported blockchain", LangZH: "不支持的区块链", LangJA: "サポートされていないブロックチェーン", LangKO: "지원되지 않는 블록체인", LangES: "blockchain no soportada"},

	// Payment
	ErrPaymentNotFound:      {LangEN: "payment not found", LangZH: "支付订单未找到", LangJA: "決済が見つかりません", LangKO: "결제를 찾을 수 없습니다", LangES: "pago no encontrado"},
	ErrPaymentExpired:       {LangEN: "payment has expired", LangZH: "支付订单已过期", LangJA: "決済の有効期限が切れました", LangKO: "결제가 만료되었습니다", LangES: "el pago ha expirado"},
	ErrPaymentNotCompleted:  {LangEN: "payment is not completed", LangZH: "支付订单未完成", LangJA: "決済が完了していません", LangKO: "결제가 완료되지 않았습니다", LangES: "el pago no está completado"},
	ErrPaymentCreateFailed:  {LangEN: "failed to create payment", LangZH: "创建支付订单失败", LangJA: "決済の作成に失敗しました", LangKO: "결제 생성에 실패했습니다", LangES: "error al crear el pago"},
	ErrPaymentAlreadyExists: {LangEN: "payment already exists", LangZH: "支付订单已存在", LangJA: "決済はすでに存在します", LangKO: "결제가 이미 존재합니다", LangES: "el pago ya existe"},
	ErrPaymentLinkNotFound:  {LangEN: "payment link not found", LangZH: "支付链接未找到", LangJA: "決済リンクが見つかりません", LangKO: "결제 링크를 찾을 수 없습니다", LangES: "enlace de pago no encontrado"},
	ErrPaymentLinkInactive:  {LangEN: "payment link is inactive", LangZH: "支付链接已停用", LangJA: "決済リンクは無効です", LangKO: "결제 링크가 비활성화되었습니다", LangES: "enlace de pago inactivo"},

	// Payout
	ErrPayoutNotFound:           {LangEN: "payout not found", LangZH: "打款订单未找到", LangJA: "送金が見つかりません", LangKO: "출금을 찾을 수 없습니다", LangES: "pago no encontrado"},
	ErrPayoutCreateFailed:       {LangEN: "failed to create payout", LangZH: "创建打款订单失败", LangJA: "送金の作成に失敗しました", LangKO: "출금 생성에 실패했습니다", LangES: "error al crear el pago"},
	ErrPayoutNotPendingApproval: {LangEN: "payout is not pending approval", LangZH: "打款订单不在待审批状态", LangJA: "送金は承認待ちではありません", LangKO: "출금이 승인 대기 상태가 아닙니다", LangES: "el pago no está pendiente de aprobación"},
	ErrPayoutNotOwnedByMerchant: {LangEN: "payout does not belong to this merchant", LangZH: "打款订单不属于此商户", LangJA: "この送金はこの加盟店に属していません", LangKO: "이 출금은 이 가맹점에 속하지 않습니다", LangES: "el pago no pertenece a este comerciante"},
	ErrPayoutExceedsSingleLimit: {LangEN: "amount exceeds single transaction limit", LangZH: "金额超过单笔交易限额", LangJA: "金額が1回の取引限度額を超えています", LangKO: "금액이 단일 거래 한도를 초과합니다", LangES: "el monto excede el límite por transacción"},
	ErrPayoutExceedsDailyLimit:  {LangEN: "amount would exceed daily limit", LangZH: "金额将超过每日限额", LangJA: "金額が1日の限度額を超えます", LangKO: "금액이 일일 한도를 초과합니다", LangES: "el monto excedería el límite diario"},
	ErrPayoutApprovalFailed:     {LangEN: "failed to approve payout", LangZH: "审批打款失败", LangJA: "送金の承認に失敗しました", LangKO: "출금 승인에 실패했습니다", LangES: "error al aprobar el pago"},
	ErrPayoutRejectFailed:       {LangEN: "failed to reject payout", LangZH: "拒绝打款失败", LangJA: "送金の拒否に失敗しました", LangKO: "출금 거절에 실패했습니다", LangES: "error al rechazar el pago"},
	ErrBatchCreateFailed:        {LangEN: "failed to create batch payout", LangZH: "创建批量打款失败", LangJA: "一括送金の作成に失敗しました", LangKO: "일괄 출금 생성에 실패했습니다", LangES: "error al crear pago por lotes"},
	ErrBatchNotFound:            {LangEN: "batch payout not found", LangZH: "批量打款未找到", LangJA: "一括送金が見つかりません", LangKO: "일괄 출금을 찾을 수 없습니다", LangES: "pago por lotes no encontrado"},
	ErrBatchPartialFailure:      {LangEN: "some batch items failed to create", LangZH: "部分批量打款项创建失败", LangJA: "一部の一括送金項目の作成に失敗しました", LangKO: "일부 일괄 항목 생성에 실패했습니다", LangES: "algunos elementos del lote fallaron"},

	// Refund
	ErrRefundNotFound:       {LangEN: "refund not found", LangZH: "退款未找到", LangJA: "返金が見つかりません", LangKO: "환불을 찾을 수 없습니다", LangES: "reembolso no encontrado"},
	ErrRefundCreateFailed:   {LangEN: "failed to create refund", LangZH: "创建退款失败", LangJA: "返金の作成に失敗しました", LangKO: "환불 생성에 실패했습니다", LangES: "error al crear el reembolso"},
	ErrRefundExceedsPayment: {LangEN: "refund total would exceed payment amount", LangZH: "退款总额将超过支付金额", LangJA: "返金合計が決済額を超えます", LangKO: "환불 총액이 결제 금액을 초과합니다", LangES: "el total del reembolso excedería el monto del pago"},
	ErrRefundOnlyCompleted:  {LangEN: "can only refund completed payments", LangZH: "只能退款已完成的支付", LangJA: "完了した決済のみ返金可能です", LangKO: "완료된 결제만 환불할 수 있습니다", LangES: "solo se pueden reembolsar pagos completados"},

	// Merchant / Wallet / Config
	ErrMerchantNotFound:       {LangEN: "merchant not found", LangZH: "商户未找到", LangJA: "加盟店が見つかりません", LangKO: "가맹점을 찾을 수 없습니다", LangES: "comerciante no encontrado"},
	ErrMerchantCreateFailed:   {LangEN: "failed to create merchant", LangZH: "创建商户失败", LangJA: "加盟店の作成に失敗しました", LangKO: "가맹점 생성에 실패했습니다", LangES: "error al crear el comerciante"},
	ErrMerchantInactive:       {LangEN: "merchant is inactive", LangZH: "商户已停用", LangJA: "加盟店は無効です", LangKO: "가맹점이 비활성화되었습니다", LangES: "comerciante inactivo"},
	ErrWalletCreateFailed:     {LangEN: "failed to create wallet", LangZH: "创建钱包失败", LangJA: "ウォレットの作成に失敗しました", LangKO: "지갑 생성에 실패했습니다", LangES: "error al crear la billetera"},
	ErrDerivationIndexFailed:  {LangEN: "failed to get derivation index", LangZH: "获取派生索引失败", LangJA: "派生インデックスの取得に失敗しました", LangKO: "파생 인덱스 조회에 실패했습니다", LangES: "error al obtener el índice de derivación"},
	ErrDeriveAddressFailed:    {LangEN: "failed to derive address", LangZH: "派生地址失败", LangJA: "アドレスの派生に失敗しました", LangKO: "주소 파생에 실패했습니다", LangES: "error al derivar la dirección"},
	ErrSweepConfigFailed:      {LangEN: "failed to save sweep config", LangZH: "保存归集配置失败", LangJA: "スイープ設定の保存に失敗しました", LangKO: "스윕 설정 저장에 실패했습니다", LangES: "error al guardar la configuración de barrido"},
	ErrColdWalletConfigFailed: {LangEN: "failed to save cold wallet config", LangZH: "保存冷钱包配置失败", LangJA: "コールドウォレット設定の保存に失敗しました", LangKO: "콜드월렛 설정 저장에 실패했습니다", LangES: "error al guardar la configuración de billetera fría"},
	ErrApprovalConfigInvalid:  {LangEN: "invalid approval configuration", LangZH: "审批配置无效", LangJA: "承認設定が無効です", LangKO: "잘못된 승인 설정", LangES: "configuración de aprobación inválida"},
	ErrApprovalConfigFailed:   {LangEN: "failed to save approval config", LangZH: "保存审批配置失败", LangJA: "承認設定の保存に失敗しました", LangKO: "승인 설정 저장에 실패했습니다", LangES: "error al guardar la configuración de aprobación"},
	ErrWebhookURLInvalid:      {LangEN: "invalid webhook URL", LangZH: "Webhook URL 无效", LangJA: "Webhook URLが無効です", LangKO: "잘못된 Webhook URL", LangES: "URL de webhook inválida"},
	ErrWebhookURLPrivate:      {LangEN: "webhook URL must not point to private addresses", LangZH: "Webhook URL 不能指向内网地址", LangJA: "Webhook URLはプライベートアドレスを指してはいけません", LangKO: "Webhook URL은 내부 주소를 가리킬 수 없습니다", LangES: "la URL del webhook no debe apuntar a direcciones privadas"},

	// Admin
	ErrAdminLoginFailed:      {LangEN: "invalid username or password", LangZH: "用户名或密码错误", LangJA: "ユーザー名またはパスワードが無効です", LangKO: "잘못된 사용자 이름 또는 비밀번호", LangES: "nombre de usuario o contraseña inválidos"},
	ErrAdminTokenInvalid:     {LangEN: "invalid or expired token", LangZH: "令牌无效或已过期", LangJA: "トークンが無効または期限切れです", LangKO: "유효하지 않거나 만료된 토큰", LangES: "token inválido o expirado"},
	ErrAdminTokenExpired:     {LangEN: "token has expired", LangZH: "令牌已过期", LangJA: "トークンの有効期限が切れました", LangKO: "토큰이 만료되었습니다", LangES: "el token ha expirado"},
	ErrAdminUserNotFound:     {LangEN: "admin user not found", LangZH: "管理员用户未找到", LangJA: "管理者ユーザーが見つかりません", LangKO: "관리자 사용자를 찾을 수 없습니다", LangES: "usuario administrador no encontrado"},
	ErrAdminUserCreateFailed: {LangEN: "failed to create admin user", LangZH: "创建管理员用户失败", LangJA: "管理者ユーザーの作成に失敗しました", LangKO: "관리자 사용자 생성에 실패했습니다", LangES: "error al crear el usuario administrador"},
	ErrAdminUserUpdateFailed: {LangEN: "failed to update admin user", LangZH: "更新管理员用户失败", LangJA: "管理者ユーザーの更新に失敗しました", LangKO: "관리자 사용자 업데이트에 실패했습니다", LangES: "error al actualizar el usuario administrador"},
	ErrAdminUserDeleteFailed: {LangEN: "failed to delete admin user", LangZH: "删除管理员用户失败", LangJA: "管理者ユーザーの削除に失敗しました", LangKO: "관리자 사용자 삭제에 실패했습니다", LangES: "error al eliminar el usuario administrador"},
	ErrAdminCannotDeleteSelf: {LangEN: "cannot delete yourself", LangZH: "不能删除自己", LangJA: "自分自身を削除できません", LangKO: "자기 자신을 삭제할 수 없습니다", LangES: "no puede eliminarse a sí mismo"},
	ErrAdminLastSuperAdmin:   {LangEN: "cannot delete the last super admin", LangZH: "不能删除最后一个超级管理员", LangJA: "最後のスーパー管理者は削除できません", LangKO: "마지막 슈퍼 관리자를 삭제할 수 없습니다", LangES: "no se puede eliminar al último super administrador"},
	ErrAdminPasswordTooShort: {LangEN: "password must be 12-128 characters", LangZH: "密码长度必须在 12-128 个字符之间", LangJA: "パスワードは12〜128文字である必要があります", LangKO: "비밀번호는 12-128자여야 합니다", LangES: "la contraseña debe tener entre 12 y 128 caracteres"},
	ErrAdminInsufficientRole: {LangEN: "insufficient permissions", LangZH: "权限不足", LangJA: "権限が不足しています", LangKO: "권한이 부족합니다", LangES: "permisos insuficientes"},
	ErrAdminRefreshFailed:    {LangEN: "failed to refresh token", LangZH: "刷新令牌失败", LangJA: "トークンの更新に失敗しました", LangKO: "토큰 갱신에 실패했습니다", LangES: "error al renovar el token"},
	ErrAdminUserDeactivated:  {LangEN: "user account is deactivated", LangZH: "用户账号已停用", LangJA: "ユーザーアカウントは無効化されています", LangKO: "사용자 계정이 비활성화되었습니다", LangES: "la cuenta de usuario está desactivada"},
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
