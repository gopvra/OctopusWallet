package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	httpClient *http.Client
	maxRetries int
	backoff    time.Duration
}

func NewService(timeout time.Duration, maxRetries int, backoff time.Duration) *Service {
	return &Service{
		httpClient: &http.Client{Timeout: timeout},
		maxRetries: maxRetries,
		backoff:    backoff,
	}
}

// Send dispatches a webhook event to the merchant's URL with HMAC-SHA256 signature.
func (s *Service) Send(ctx context.Context, webhookURL, apiKeyHash string, eventType EventType, data any) error {
	if webhookURL == "" {
		return nil
	}

	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal webhook event: %w", err)
	}

	// Generate HMAC-SHA256 signature using the merchant's API key hash as secret
	mac := hmac.New(sha256.New, []byte(apiKeyHash))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	// Retry with exponential backoff
	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(s.backoff * time.Duration(1<<uint(attempt-1))):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create webhook request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-ID", event.ID)
		req.Header.Set("X-Webhook-Timestamp", event.Timestamp.Format(time.RFC3339))

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = err
			slog.Warn("webhook delivery failed", "url", webhookURL, "attempt", attempt+1, "error", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			slog.Info("webhook delivered", "url", webhookURL, "event", eventType, "id", event.ID)
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		slog.Warn("webhook delivery failed", "url", webhookURL, "status", resp.StatusCode, "attempt", attempt+1)
	}

	return fmt.Errorf("webhook delivery failed after %d attempts: %w", s.maxRetries+1, lastErr)
}
