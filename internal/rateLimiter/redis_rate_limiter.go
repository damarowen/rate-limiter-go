package rateLimiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client         *redis.Client
	defaultLimit   int
	defaultWindow  time.Duration
	premiumConfigs map[string]*PremiumConfig
}

type PremiumConfig struct {
	limit  int
	window time.Duration
}

func NewRedisRateLimiter(client *redis.Client, defaultLimit int, defaultWindow time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:         client,
		defaultLimit:   defaultLimit,
		defaultWindow:  defaultWindow,
		premiumConfigs: make(map[string]*PremiumConfig),
	}
}

func (rrl *RedisRateLimiter) SetPremiumClient(key string, limit int, window time.Duration) {
	rrl.premiumConfigs[key] = &PremiumConfig{
		limit:  limit,
		window: window,
	}
}

func (rrl *RedisRateLimiter) Allow(key string) bool {
	ctx := context.Background()

	// Get config for this key
	limit := rrl.defaultLimit
	window := rrl.defaultWindow

	if config, exists := rrl.premiumConfigs[key]; exists {
		limit = config.limit
		window = config.window
	}

	// Redis key format: rate_limit:{api_key}
	redisKey := fmt.Sprintf("rate_limit:%s", key)

	getKeys, err := rrl.GetCurrentCount(key)

	if err != nil {
		fmt.Println("Error getting current count for key", key, ":", err)
		return false
	}

	if getKeys > int64(limit) {
		return false
	}

	// Increment counter
	count, err := rrl.client.Incr(ctx, redisKey).Result()
	if err != nil {
		// On error, deny request (fail-safe)
		return false
	}

	// Set expiry on first request
	if count == 1 {
		err := rrl.client.Expire(ctx, redisKey, window).Err()
		if err != nil {
			return false
		}
	}

	// Check if within limit
	return count <= int64(limit)
}

func (rrl *RedisRateLimiter) Reset(key string) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	rrl.client.Del(ctx, redisKey)
}

// GetCurrentCount returns current request count (for monitoring)
func (rrl *RedisRateLimiter) GetCurrentCount(key string) (int64, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	get, err := rrl.client.Get(ctx, redisKey).Int64()

	if errors.Is(err, redis.Nil) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}
	return get, nil
}

// GetTTL returns remaining time until window reset
func (rrl *RedisRateLimiter) GetTTL(key string) (time.Duration, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	return rrl.client.TTL(ctx, redisKey).Result()
}
