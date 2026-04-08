package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps go-redis and provides typed caching, idempotency, and rate limiting.
type Client struct {
	rdb *redis.Client
}

// New creates a new Redis cache client.
func New(addr, password string, db int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// --- Generic Cache ---

func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Exists(ctx context.Context, key string) bool {
	n, _ := c.rdb.Exists(ctx, key).Result()
	return n > 0
}

// --- Idempotency ---

type IdempotencyEntry struct {
	Status int    `json:"status"`
	Body   []byte `json:"body"`
}

// SetIdempotent stores an idempotency response. Returns false if key already exists.
func (c *Client) SetIdempotent(ctx context.Context, key string, entry *IdempotencyEntry, ttl time.Duration) (bool, error) {
	data, _ := json.Marshal(entry)
	ok, err := c.rdb.SetNX(ctx, "idem:"+key, data, ttl).Result()
	return ok, err
}

// GetIdempotent retrieves a cached idempotency response.
func (c *Client) GetIdempotent(ctx context.Context, key string) (*IdempotencyEntry, error) {
	data, err := c.rdb.Get(ctx, "idem:"+key).Bytes()
	if err != nil {
		return nil, err
	}
	var entry IdempotencyEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// --- Rate Limiting (sliding window) ---

// RateLimit checks if a key has exceeded the limit within the window.
// Returns (allowed bool, remaining int, error).
func (c *Client) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()
	redisKey := "rl:" + key

	pipe := c.rdb.Pipeline()
	// Remove old entries
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))
	// Count current entries
	countCmd := pipe.ZCard(ctx, redisKey)
	// Add current request
	pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(now), Member: now})
	// Set expiry on the key
	pipe.Expire(ctx, redisKey, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return true, limit, err // fail open
	}

	count := int(countCmd.Val())
	if count >= limit {
		return false, 0, nil
	}
	return true, limit - count, nil
}

// --- Dashboard Cache ---

func (c *Client) CacheKey(parts ...string) string {
	key := "cache"
	for _, p := range parts {
		key += ":" + p
	}
	return key
}
