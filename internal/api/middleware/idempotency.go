package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const maxIdempotencyEntries = 10000

type IdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]idempotencyEntry
}

type idempotencyEntry struct {
	status   int
	body     []byte
	expireAt time.Time
}

func NewIdempotencyStore() *IdempotencyStore {
	s := &IdempotencyStore{entries: make(map[string]idempotencyEntry)}
	go s.cleanup()
	return s
}

func (s *IdempotencyStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.entries {
			if now.After(v.expireAt) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *IdempotencyStore) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		merchantID, _ := c.Get("merchant_id")
		scopedKey := hashKey(key, merchantID)

		s.mu.RLock()
		entry, exists := s.entries[scopedKey]
		s.mu.RUnlock()

		if exists && time.Now().Before(entry.expireAt) {
			c.Data(entry.status, "application/json", entry.body)
			c.Abort()
			return
		}

		// Reject if store is full (DoS protection)
		s.mu.RLock()
		size := len(s.entries)
		s.mu.RUnlock()
		if size >= maxIdempotencyEntries {
			c.Next()
			return
		}

		writer := &responseCapture{ResponseWriter: c.Writer}
		c.Writer = writer

		c.Next()

		s.mu.Lock()
		s.entries[scopedKey] = idempotencyEntry{
			status:   writer.status,
			body:     writer.body,
			expireAt: time.Now().Add(24 * time.Hour),
		}
		s.mu.Unlock()
	}
}

func hashKey(key string, merchantID interface{}) string {
	h := sha256.New()
	h.Write([]byte(key))
	if mid, ok := merchantID.(string); ok {
		h.Write([]byte(mid))
	}
	return hex.EncodeToString(h.Sum(nil))
}

type responseCapture struct {
	gin.ResponseWriter
	status int
	body   []byte
}

func (w *responseCapture) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseCapture) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseCapture) WriteString(s string) (int, error) {
	w.body = append(w.body, []byte(s)...)
	return w.ResponseWriter.WriteString(s)
}

// RequestHMAC validates that incoming requests are signed with HMAC-SHA256.
// Signature is computed over: method + path + body + timestamp using the API key hash.
// Signature is optional — only enforced when X-Request-Signature header is present.
func RequestHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		sig := c.GetHeader("X-Request-Signature")
		if sig == "" {
			c.Next()
			return
		}

		timestamp := c.GetHeader("X-Request-Timestamp")
		if timestamp == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Request-Timestamp"})
			return
		}

		ts, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp format"})
			return
		}
		if time.Since(ts).Abs() > 5*time.Minute {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "request timestamp expired"})
			return
		}

		// Read body for signature verification
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Get merchant's API key hash for HMAC computation
		merchant, exists := c.Get("merchant")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		// Compute expected HMAC: method + path + body + timestamp
		type apiKeyHolder interface {
			GetAPIKeyHash() string
		}
		m, ok := merchant.(apiKeyHolder)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "merchant context missing for request signing"})
			return
		}
		secret := m.GetAPIKeyHash()

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(c.Request.Method))
		mac.Write([]byte(c.Request.URL.Path))
		mac.Write(body)
		mac.Write([]byte(timestamp))
		expectedSig := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid request signature"})
			return
		}

		c.Next()
	}
}
