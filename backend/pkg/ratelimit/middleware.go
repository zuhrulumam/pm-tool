package ratelimit

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	ctxpkg "github.com/zuhrulumam/pm-tool/pkg/context"
	"github.com/zuhrulumam/pm-tool/pkg/response"
)

// Middleware creates a rate limiting middleware
func Middleware(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP address as key (or use user ID if authenticated)
		key := c.ClientIP()
		
		// Try to get user ID from context for authenticated requests
		if userID, ok := ctxpkg.GetUserID(c.Request.Context()); ok {
			key = fmt.Sprintf("user:%s", userID)
		}

		// Check rate limit
		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			response.Errorf(c, err)
			c.Abort()
			return
		}

		if !allowed {
			response.Errorf(c, &RateLimitError{})
			c.Abort()
			return
		}

		// Get remaining requests
		remaining, _ := limiter.GetRemaining(c.Request.Context(), key)
		
		// Set rate limit headers
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// PerUser creates a rate limiter for authenticated users
func PerUser(requestsPerMinute int) gin.HandlerFunc {
	// This would need access to Redis client from config
	// For now, return a placeholder
	return func(c *gin.Context) {
		c.Next()
	}
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}

func (e *RateLimitError) StatusCode() int {
	return http.StatusTooManyRequests
}
