package crypto

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var (
	ErrInvalidWebhookURL = errors.New("invalid webhook URL: must be http or https")
	ErrPrivateWebhookURL = errors.New("webhook URL must not point to private/internal addresses")
)

// ValidateWebhookURL checks that a webhook URL is safe to call.
// Rejects internal IPs (SSRF prevention), non-http(s) schemes, and empty hosts.
func ValidateWebhookURL(rawURL string) error {
	if rawURL == "" {
		return nil // empty is allowed (no webhook configured)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidWebhookURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidWebhookURL
	}

	host := u.Hostname()
	if host == "" {
		return ErrInvalidWebhookURL
	}

	// Block localhost variants
	lower := strings.ToLower(host)
	if lower == "localhost" || lower == "127.0.0.1" || lower == "::1" || lower == "0.0.0.0" {
		return ErrPrivateWebhookURL
	}

	// Resolve and check all IPs
	ips, err := net.LookupHost(host)
	if err != nil {
		// Can't resolve — allow (might be valid later, or DNS not available at config time)
		return nil
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return ErrPrivateWebhookURL
		}
	}

	return nil
}
