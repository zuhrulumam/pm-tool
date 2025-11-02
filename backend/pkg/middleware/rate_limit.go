package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware implements rate limiting using Redis
func RateLimitMiddleware(redisClient *redis.Client, requestsPerSecond int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		key := fmt.Sprintf("rate_limit:%s", c.ClientIP())

		// Get current count
		val, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if val >= requestsPerSecond {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Second)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		c.Next()
	}
}
