package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	otelredis "github.com/redis/go-redis/extra/redisotel/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Config holds Redis configuration
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	
	// Telemetry configuration
	ServiceName     string
	DetailedMetrics bool
	
}

// ✅ FIX: Client always wraps redis.Client for consistent interface
// Works with or without telemetry
type Client struct {
	*redis.Client
	metrics *RedisMetrics
	tracer  trace.Tracer
	config  Config
}

// ✅ FIX: Always returns *Client (not *redis.Client)
// NewClient creates a new Redis client with comprehensive telemetry
func NewClient(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Add OpenTelemetry instrumentation
	// Basic tracing - traces all Redis operations
	if err := otelredis.InstrumentTracing(rdb); err != nil {
		return nil, fmt.Errorf("failed to instrument Redis tracing: %w", err)
	}

	// Basic metrics - collects Redis default metrics
	if err := otelredis.InstrumentMetrics(rdb); err != nil {
		return nil, fmt.Errorf("failed to instrument Redis metrics: %w", err)
	}

	// Initialize custom metrics collector
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "redis-client"
	}
	metrics := NewRedisMetrics(serviceName)

	// Get tracer for custom spans
	tracer := otel.Tracer("redis-client")

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Record initial connection
	metrics.RecordConnection(true)

	return &Client{
		Client:  rdb,
		metrics: metrics,
		tracer:  tracer,
		config:  cfg,
	}, nil
	
}


// ✅ Enhanced telemetry-aware methods below
// These override the embedded redis.Client methods for detailed tracking

// Get wraps redis.Client.Get with custom telemetry
func (c *Client) Get(ctx context.Context, key string) *redis.StringCmd {
	start := time.Now()
	
	// Create custom span for detailed tracking
	ctx, span := c.tracer.Start(ctx, "redis.get",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	// Add attributes
	span.SetAttributes(
		attribute.String("redis.command", "GET"),
		attribute.String("redis.key", c.sanitizeKey(key)),
		attribute.Int("redis.db", c.config.DB),
	)
	
	// Execute command
	cmd := c.Client.Get(ctx, key)
	
	// Record metrics
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("GET", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("GET", status, duration)
	
	// Track cache hit/miss
	if cmd.Err() == redis.Nil {
		c.metrics.RecordCacheHit("GET", false)
	} else if cmd.Err() == nil {
		c.metrics.RecordCacheHit("GET", true)
		
	}
	
	return cmd
}

// Set wraps redis.Client.Set with custom telemetry
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	start := time.Now()
	
	ctx, span := c.tracer.Start(ctx, "redis.set",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	span.SetAttributes(
		attribute.String("redis.command", "SET"),
		attribute.String("redis.key", c.sanitizeKey(key)),
		attribute.Int("redis.db", c.config.DB),
		attribute.String("redis.expiration", expiration.String()),
	)
	
	cmd := c.Client.Set(ctx, key, value, expiration)
	
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("SET", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("SET", status, duration)
	
	
	
	return cmd
}

// Del wraps redis.Client.Del with custom telemetry
func (c *Client) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	start := time.Now()
	
	ctx, span := c.tracer.Start(ctx, "redis.del",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	span.SetAttributes(
		attribute.String("redis.command", "DEL"),
		attribute.Int("redis.key_count", len(keys)),
		attribute.Int("redis.db", c.config.DB),
	)
	
	cmd := c.Client.Del(ctx, keys...)
	
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("DEL", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("DEL", status, duration)
	
	return cmd
}

// Exists wraps redis.Client.Exists with custom telemetry
func (c *Client) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	start := time.Now()
	
	ctx, span := c.tracer.Start(ctx, "redis.exists",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	span.SetAttributes(
		attribute.String("redis.command", "EXISTS"),
		attribute.Int("redis.key_count", len(keys)),
		attribute.Int("redis.db", c.config.DB),
	)
	
	cmd := c.Client.Exists(ctx, keys...)
	
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("EXISTS", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("EXISTS", status, duration)
	
	return cmd
}

// HGet wraps redis.Client.HGet with custom telemetry
func (c *Client) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	start := time.Now()
	
	ctx, span := c.tracer.Start(ctx, "redis.hget",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	span.SetAttributes(
		attribute.String("redis.command", "HGET"),
		attribute.String("redis.key", c.sanitizeKey(key)),
		attribute.String("redis.field", field),
		attribute.Int("redis.db", c.config.DB),
	)
	
	cmd := c.Client.HGet(ctx, key, field)
	
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil && cmd.Err() != redis.Nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("HGET", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("HGET", status, duration)
	
	// Track cache hit/miss
	if cmd.Err() == redis.Nil {
		c.metrics.RecordCacheHit("HGET", false)
	} else if cmd.Err() == nil {
		c.metrics.RecordCacheHit("HGET", true)
	}
	
	return cmd
}

// HSet wraps redis.Client.HSet with custom telemetry
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	start := time.Now()
	
	ctx, span := c.tracer.Start(ctx, "redis.hset",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	span.SetAttributes(
		attribute.String("redis.command", "HSET"),
		attribute.String("redis.key", c.sanitizeKey(key)),
		attribute.Int("redis.field_count", len(values)/2),
		attribute.Int("redis.db", c.config.DB),
	)
	
	cmd := c.Client.HSet(ctx, key, values...)
	
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("HSET", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("HSET", status, duration)
	
	return cmd
}

// Incr wraps redis.Client.Incr with custom telemetry
func (c *Client) Incr(ctx context.Context, key string) *redis.IntCmd {
	start := time.Now()
	
	ctx, span := c.tracer.Start(ctx, "redis.incr",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	span.SetAttributes(
		attribute.String("redis.command", "INCR"),
		attribute.String("redis.key", c.sanitizeKey(key)),
		attribute.Int("redis.db", c.config.DB),
	)
	
	cmd := c.Client.Incr(ctx, key)
	
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("INCR", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("INCR", status, duration)
	
	return cmd
}

// Expire wraps redis.Client.Expire with custom telemetry
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	start := time.Now()
	
	ctx, span := c.tracer.Start(ctx, "redis.expire",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	
	span.SetAttributes(
		attribute.String("redis.command", "EXPIRE"),
		attribute.String("redis.key", c.sanitizeKey(key)),
		attribute.String("redis.expiration", expiration.String()),
		attribute.Int("redis.db", c.config.DB),
	)
	
	cmd := c.Client.Expire(ctx, key, expiration)
	
	duration := time.Since(start).Seconds()
	status := "success"
	if cmd.Err() != nil {
		status = "error"
		span.RecordError(cmd.Err())
		c.metrics.RecordError("EXPIRE", categorizeRedisError(cmd.Err()))
	}
	
	c.metrics.RecordOperation("EXPIRE", status, duration)
	
	return cmd
}

// GetConnectionPoolStats returns current connection pool statistics
func (c *Client) GetConnectionPoolStats() *redis.PoolStats {
	stats := c.Client.PoolStats()
	
	// Record connection pool metrics
	c.metrics.RecordConnectionPool(
		stats.Hits,
		stats.Misses,
		stats.Timeouts,
		uint32(stats.TotalConns),
		uint32(stats.IdleConns),
		uint32(stats.StaleConns),
	)
	
	return stats
}

// Close closes the Redis client and records final metrics
func (c *Client) Close() error {
	c.metrics.RecordConnection(false)
	return c.Client.Close()
}

// sanitizeKey sanitizes sensitive data from keys for tracing
func (c *Client) sanitizeKey(key string) string {
	// Only include key pattern if detailed metrics is disabled
	if !c.config.DetailedMetrics {
		// Return generic pattern or hash
		return "***"
	}
	return key
}

// categorizeRedisError categorizes Redis errors for better metrics
func categorizeRedisError(err error) string {
	if err == nil {
		return "none"
	}
	
	switch err {
	case redis.Nil:
		return "not_found"
	case context.DeadlineExceeded:
		return "timeout"
	case context.Canceled:
		return "canceled"
	default:
		errStr := err.Error()
		switch {
		case contains(errStr, "connection"):
			return "connection_error"
		case contains(errStr, "timeout"):
			return "timeout"
		case contains(errStr, "auth"):
			return "auth_error"
		case contains(errStr, "readonly"):
			return "readonly_error"
		default:
			return "unknown_error"
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr)*2))
}

