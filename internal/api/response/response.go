package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
)

// Response is the unified API response envelope.
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// getLang extracts the language from gin context (set by LangMiddleware).
func getLang(c *gin.Context) errcode.Lang {
	if lang, exists := c.Get("lang"); exists {
		if l, ok := lang.(errcode.Lang); ok {
			return l
		}
	}
	return errcode.LangEN
}

// OK sends a success response with data.
func OK(c *gin.Context, data interface{}) {
	lang := getLang(c)
	c.JSON(http.StatusOK, Response{
		Code: errcode.OK.Int(),
		Msg:  errcode.OK.Msg(lang),
		Data: data,
	})
}

// OKMsg sends a success response with a custom message.
func OKMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: errcode.OK.Int(),
		Msg:  msg,
	})
}

// Fail sends an error response with an error code.
func Fail(c *gin.Context, code errcode.Code) {
	lang := getLang(c)
	httpStatus := codeToHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code: code.Int(),
		Msg:  code.Msg(lang),
	})
}

// FailData sends an error response with an error code and extra data.
func FailData(c *gin.Context, code errcode.Code, data interface{}) {
	lang := getLang(c)
	httpStatus := codeToHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code: code.Int(),
		Msg:  code.Msg(lang),
		Data: data,
	})
}

// FailMsg sends an error response with a custom message (for validation errors).
func FailMsg(c *gin.Context, code errcode.Code, msg string) {
	httpStatus := codeToHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code: code.Int(),
		Msg:  msg,
	})
}

// Abort sends an error response and aborts the middleware chain.
// Use in middleware (auth, rate limit, etc.) instead of Fail.
func Abort(c *gin.Context, code errcode.Code) {
	lang := getLang(c)
	httpStatus := codeToHTTPStatus(code)
	c.AbortWithStatusJSON(httpStatus, Response{
		Code: code.Int(),
		Msg:  code.Msg(lang),
	})
}

// AbortMsg sends an error response with custom message and aborts.
func AbortMsg(c *gin.Context, code errcode.Code, msg string) {
	httpStatus := codeToHTTPStatus(code)
	c.AbortWithStatusJSON(httpStatus, Response{
		Code: code.Int(),
		Msg:  msg,
	})
}

// codeToHTTPStatus maps error code ranges to HTTP status codes.
func codeToHTTPStatus(code errcode.Code) int {
	switch {
	case code == errcode.OK:
		return http.StatusOK
	case code == errcode.ErrUnauthorized, code == errcode.ErrAdminTokenInvalid,
		code == errcode.ErrAdminTokenExpired, code == errcode.ErrInvalidSignature:
		return http.StatusUnauthorized
	case code == errcode.ErrForbidden, code == errcode.ErrAdminInsufficientRole,
		code == errcode.ErrIPNotWhitelisted:
		return http.StatusForbidden
	case code == errcode.ErrRateLimited:
		return http.StatusTooManyRequests
	case code == errcode.ErrNotFound, code == errcode.ErrPaymentNotFound,
		code == errcode.ErrPayoutNotFound, code == errcode.ErrRefundNotFound,
		code == errcode.ErrMerchantNotFound, code == errcode.ErrAdminUserNotFound,
		code == errcode.ErrBatchNotFound, code == errcode.ErrPaymentLinkNotFound:
		return http.StatusNotFound
	case code.Int() >= 10001 && code.Int() <= 10015:
		return http.StatusBadRequest
	default:
		// 2xxxx-6xxxx business errors default to 200 with error code in body
		return http.StatusOK
	}
}
