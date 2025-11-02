package telemetry

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// InitLogger initializes zerolog with trace context
func InitLogger(level, format string) zerolog.Logger {
	var output io.Writer = os.Stdout

	// Set log level
	logLevel := zerolog.InfoLevel
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Use console writer for non-JSON format
	if format != "json" {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	return zerolog.New(output).With().Timestamp().Logger()
}

// LoggerWithTraceContext adds trace context to logger
func LoggerWithTraceContext(ctx context.Context, logger zerolog.Logger) zerolog.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return logger
	}

	spanCtx := span.SpanContext()
	return logger.With().
		Str("trace_id", spanCtx.TraceID().String()).
		Str("span_id", spanCtx.SpanID().String()).
		Logger()
}
