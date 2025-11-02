package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithContext(ctx context.Context) Logger
	WithFields(fields ...Field) Logger
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// zerologLogger implements Logger interface using zerolog
type zerologLogger struct {
	logger zerolog.Logger
}

// Config for logger
type Config struct {
	Level      string
	Output     io.Writer
	PrettyPrint bool
}

// New creates a new logger instance
func New(config Config) Logger {
	// Parse log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Set output
	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	// Pretty print for development
	if config.PrettyPrint {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	zlog := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &zerologLogger{logger: zlog}
}

// Default creates a logger with default config
func Default() Logger {
	return New(Config{
		Level:       "info",
		Output:      os.Stdout,
		PrettyPrint: false,
	})
}

func (l *zerologLogger) Debug(msg string, fields ...Field) {
	l.log(l.logger.Debug(), msg, fields...)
}

func (l *zerologLogger) Info(msg string, fields ...Field) {
	l.log(l.logger.Info(), msg, fields...)
}

func (l *zerologLogger) Warn(msg string, fields ...Field) {
	l.log(l.logger.Warn(), msg, fields...)
}

func (l *zerologLogger) Error(msg string, fields ...Field) {
	l.log(l.logger.Error(), msg, fields...)
}

func (l *zerologLogger) Fatal(msg string, fields ...Field) {
	l.log(l.logger.Fatal(), msg, fields...)
}

func (l *zerologLogger) WithContext(ctx context.Context) Logger {
	return &zerologLogger{
		logger: l.logger.With().Logger(),
	}
}

func (l *zerologLogger) WithFields(fields ...Field) Logger {
	ctx := l.logger.With()
	for _, f := range fields {
		ctx = ctx.Interface(f.Key, f.Value)
	}
	return &zerologLogger{
		logger: ctx.Logger(),
	}
}

func (l *zerologLogger) log(event *zerolog.Event, msg string, fields ...Field) {
	for _, f := range fields {
		event = event.Interface(f.Key, f.Value)
	}
	event.Msg(msg)
}

// Global logger instance
var globalLogger Logger = Default()

// SetGlobal sets the global logger instance
func SetGlobal(logger Logger) {
	globalLogger = logger
}

// Global convenience functions
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	globalLogger.Fatal(msg, fields...)
}
