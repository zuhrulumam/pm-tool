package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter interface
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	Reset(ctx context.Context, key string) error
	GetRemaining(ctx context.Context, key string) (int64, error)
}

// Options for rate limiter
type Options struct {
	RequestsPerWindow int           // Max requests per window
	Window            time.Duration // Time window
	KeyPrefix         string        // Prefix for Redis keys
}

// redisRateLimiter implements RateLimiter using Redis
type redisRateLimiter struct {
	client  *redis.Client
	options Options
}

// New creates a new RateLimiter
func New(client *redis.Client, options Options) RateLimiter {
	if options.KeyPrefix == "" {
		options.KeyPrefix = "ratelimit"
	}
	if options.Window == 0 {
		options.Window = time.Minute
	}
	if options.RequestsPerWindow == 0 {
		options.RequestsPerWindow = 100
	}

	return &redisRateLimiter{
		client:  client,
		options: options,
	}
}

// Allow checks if a request is allowed
func (rl *redisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := rl.makeKey(key)
	
	// Increment counter
	count, err := rl.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	// Set expiration on first request
	if count == 1 {
		rl.client.Expire(ctx, redisKey, rl.options.Window)
	}

	// Check if limit exceeded
	return count <= int64(rl.options.RequestsPerWindow), nil
}

// Reset resets the rate limit for a key
func (rl *redisRateLimiter) Reset(ctx context.Context, key string) error {
	redisKey := rl.makeKey(key)
	return rl.client.Del(ctx, redisKey).Err()
}

// GetRemaining returns the number of remaining requests
func (rl *redisRateLimiter) GetRemaining(ctx context.Context, key string) (int64, error) {
	redisKey := rl.makeKey(key)
	
	count, err := rl.client.Get(ctx, redisKey).Int64()
	if err != nil {
		if err == redis.Nil {
			return int64(rl.options.RequestsPerWindow), nil
		}
		return 0, err
	}

	remaining := int64(rl.options.RequestsPerWindow) - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

func (rl *redisRateLimiter) makeKey(key string) string {
	return fmt.Sprintf("%s:%s", rl.options.KeyPrefix, key)
}
