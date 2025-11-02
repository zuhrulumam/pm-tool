package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		log.Info().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("duration", duration).
			Str("client_ip", c.ClientIP()).
			Str("request_id", c.GetString("request_id")).
			Msg("HTTP request")
	}
}
