// Package redis provides Redis-based caching functionality.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"telegram-api/internal/domain"

	"github.com/redis/go-redis/v9"
)

// Key prefixes for organization.
const (
	PrefixSession    = "session:"    // User sessions (JWT blacklist)
	PrefixRateLimit  = "rate:"       // Rate limiting
	PrefixTelegram   = "tg:session:" // Temporary Telegram session data
	PrefixVerifyCode = "verify:"     // Temporary verification codes
	PrefixCache      = "cache:"      // General cache
)

// CacheRepository implements domain.CacheRepository using Redis.
// Single Responsibility: Only handles cache operations.
type CacheRepository struct {
	client *redis.Client
}

// NewCacheRepository creates a new instance of the cache repository.
func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{client: client}
}

// Set stores a string value with TTL.
func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return fmt.Errorf("unsupported type for Set: use SetJSON for objects")
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	err := r.client.Set(ctx, key, strValue, ttl).Err()
	if err != nil {
		return wrapRedisError(err, "set")
	}
	return nil
}

// Get gets a string value.
func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil // Key does not exist, return empty
		}
		return "", wrapRedisError(err, "get")
	}
	return val, nil
}

// Delete deletes one or more keys.
func (r *CacheRepository) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return wrapRedisError(err, "delete")
	}
	return nil
}

// Exists checks if a key exists.
func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, wrapRedisError(err, "exists")
	}
	return count > 0, nil
}

// SetJSON stores an object as JSON with TTL.
func (r *CacheRepository) SetJSON(ctx context.Context, key string, value interface{}, ttlSeconds int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error serializing JSON: %w", err)
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return wrapRedisError(err, "setJSON")
	}
	return nil
}

// GetJSON gets a JSON object and deserializes it.
func (r *CacheRepository) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.ErrCache // Key does not exist
		}
		return wrapRedisError(err, "getJSON")
	}

	if err := json.Unmarshal(val, dest); err != nil {
		return fmt.Errorf("error deserializing JSON: %w", err)
	}
	return nil
}

// IncrementRateLimit increments a counter for rate limiting.
// Returns the new counter value.
func (r *CacheRepository) IncrementRateLimit(ctx context.Context, key string, windowSeconds int) (int64, error) {
	pipe := r.client.Pipeline()

	// Increment counter.
	incr := pipe.Incr(ctx, key)

	// Set TTL only if first time (new key).
	pipe.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, wrapRedisError(err, "incrementRateLimit")
	}

	return incr.Val(), nil
}

// GetRateLimitCount gets the current rate limiting counter.
func (r *CacheRepository) GetRateLimitCount(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, wrapRedisError(err, "getRateLimitCount")
	}
	return val, nil
}

// SetWithNX saves only if key does not exist (useful for locks).
func (r *CacheRepository) SetWithNX(ctx context.Context, key string, value interface{}, ttlSeconds int) (bool, error) {
	ttl := time.Duration(ttlSeconds) * time.Second
	// Use SetArgs with NX option for go-redis v9 compatibility.
	result, err := r.client.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Result()
	if err != nil {
		return false, wrapRedisError(err, "setNX")
	}
	// Redis returns "OK" if key was set, empty string if key already existed.
	return result == "OK", nil
}

// GetTTL gets the remaining time to live of a key.
func (r *CacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, wrapRedisError(err, "getTTL")
	}
	return ttl, nil
}

// ScanKeys searches keys by pattern (use with caution in production).
func (r *CacheRepository) ScanKeys(ctx context.Context, pattern string, count int64) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var batch []string
		var err error
		batch, cursor, err = r.client.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, wrapRedisError(err, "scan")
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// BlacklistToken adds a token to blacklist (for logout).
func (r *CacheRepository) BlacklistToken(ctx context.Context, tokenID string, ttlSeconds int) error {
	key := PrefixSession + "blacklist:" + tokenID
	return r.Set(ctx, key, "1", ttlSeconds)
}

// IsTokenBlacklisted checks if a token is in blacklist.
func (r *CacheRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := PrefixSession + "blacklist:" + tokenID
	return r.Exists(ctx, key)
}

// StoreTelegramCode temporarily stores a verification code.
func (r *CacheRepository) StoreTelegramCode(
	ctx context.Context,
	sessionID string,
	codeHash string,
	ttlSeconds int,
) error {
	key := PrefixVerifyCode + sessionID
	return r.Set(ctx, key, codeHash, ttlSeconds)
}

// GetTelegramCode gets verification code hash.
func (r *CacheRepository) GetTelegramCode(ctx context.Context, sessionID string) (string, error) {
	key := PrefixVerifyCode + sessionID
	return r.Get(ctx, key)
}

// Health checks Redis connectivity.
func (r *CacheRepository) Health(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		return wrapRedisError(err, "health check")
	}
	return nil
}

// wrapRedisError wraps Redis errors.
func wrapRedisError(err error, operation string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w (redis: %v)", operation, domain.ErrCache, err)
}

// Compile-time verification.
var _ domain.CacheRepository = (*CacheRepository)(nil)
