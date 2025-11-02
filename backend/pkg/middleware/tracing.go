package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware adds OpenTelemetry tracing
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path,
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// Set context with span
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Record response status
		span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))

		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "HTTP error")
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}
