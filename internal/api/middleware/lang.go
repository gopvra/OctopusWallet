package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
)

// LangMiddleware extracts the preferred language from the request.
// Priority: ?lang= query param > Accept-Language header > default (en).
func LangMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := errcode.LangEN

		// 1. Query parameter
		if q := c.Query("lang"); q != "" {
			lang = parseLang(q)
		} else if al := c.GetHeader("Accept-Language"); al != "" {
			// 2. Accept-Language header (simplified — takes first match)
			lang = parseAcceptLanguage(al)
		}

		c.Set("lang", lang)
		c.Next()
	}
}

func parseLang(s string) errcode.Lang {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "zh"):
		return errcode.LangZH
	default:
		return errcode.LangEN
	}
}

func parseAcceptLanguage(header string) errcode.Lang {
	// Parse first language tag from Accept-Language
	// e.g. "zh-CN,zh;q=0.9,en;q=0.8" → zh
	parts := strings.Split(header, ",")
	if len(parts) > 0 {
		tag := strings.TrimSpace(strings.Split(parts[0], ";")[0])
		return parseLang(tag)
	}
	return errcode.LangEN
}
