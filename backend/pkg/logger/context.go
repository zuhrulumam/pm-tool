package logger

import "context"

type contextKey string

const loggerContextKey contextKey = "logger"

// ToContext adds logger to context
func ToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

// FromContext extracts logger from context
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerContextKey).(Logger); ok {
		return logger
	}
	return globalLogger
}
