package errors

import (
	"fmt"

	"github.com/palantir/stacktrace"
)

// AppError represents an application error with context and stack trace
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	StatusCode int               `json:"-"`
	Internal   error             `json:"-"`
	Fields     map[string]string `json:"fields,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// StackTrace returns the stack trace if internal error has one
func (e *AppError) StackTrace() string {
	if e.Internal != nil {
		return fmt.Sprintf("%+v", e.Internal)
	}
	return ""
}

// New creates a new AppError with stack trace
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   stacktrace.NewError(message),
	}
}

// Wrap wraps an existing error with context and stack trace
func Wrap(err error, code, message string, statusCode int) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   stacktrace.Propagate(err, message),
	}
}

// Propagate wraps error while preserving the original stack trace
func Propagate(err error, code, message string, statusCode int) *AppError {
	if err == nil {
		return nil
	}

	// If already an AppError, update and return
	if appErr, ok := err.(*AppError); ok {
		appErr.Message = message
		return appErr
	}

	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   stacktrace.Propagate(err, message),
	}
}

// WithFields adds field-level errors (for validation)
func (e *AppError) WithFields(fields map[string]string) *AppError {
	e.Fields = fields
	return e
}

// WithField adds a single field error
func (e *AppError) WithField(field, message string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[field] = message
	return e
}

// IsAppError checks if error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from error
func GetAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// RootCause returns the root cause of the error
func RootCause(err error) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		return stacktrace.RootCause(appErr.Internal)
	}

	return stacktrace.RootCause(err)
}
