package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// IPWhitelist restricts access to specified IPs per merchant.
// If no whitelist is configured for a merchant, all IPs are allowed.
type IPWhitelist struct {
	mu        sync.RWMutex
	whitelist map[string]map[string]struct{} // merchant_id -> set of allowed IPs/CIDRs
}

func NewIPWhitelist() *IPWhitelist {
	return &IPWhitelist{
		whitelist: make(map[string]map[string]struct{}),
	}
}

func (w *IPWhitelist) SetWhitelist(merchantID string, ips []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(ips) == 0 {
		delete(w.whitelist, merchantID)
		return
	}
	set := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		set[strings.TrimSpace(ip)] = struct{}{}
	}
	w.whitelist[merchantID] = set
}

func (w *IPWhitelist) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID, exists := c.Get("merchant_id")
		if !exists {
			c.Next()
			return
		}

		w.mu.RLock()
		allowed, hasWhitelist := w.whitelist[merchantID.(string)]
		w.mu.RUnlock()

		if !hasWhitelist {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		if _, ok := allowed[clientIP]; ok {
			c.Next()
			return
		}

		// Check CIDR matches
		ip := net.ParseIP(clientIP)
		if ip != nil {
			for cidr := range allowed {
				if strings.Contains(cidr, "/") {
					_, network, err := net.ParseCIDR(cidr)
					if err == nil && network.Contains(ip) {
						c.Next()
						return
					}
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP not whitelisted"})
	}
}
