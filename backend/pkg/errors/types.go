package errors

import "net/http"

// HTTP error constructors for handler layer

// BadRequest creates a 400 error
func BadRequest(message string) *AppError {
	return New("BAD_REQUEST", message, http.StatusBadRequest)
}

// Unauthorized creates a 401 error
func Unauthorized(message string) *AppError {
	return New("UNAUTHORIZED", message, http.StatusUnauthorized)
}

// Forbidden creates a 403 error
func Forbidden(message string) *AppError {
	return New("FORBIDDEN", message, http.StatusForbidden)
}

// NotFound creates a 404 error
func NotFound(message string) *AppError {
	return New("NOT_FOUND", message, http.StatusNotFound)
}

// Conflict creates a 409 error
func Conflict(message string) *AppError {
	return New("CONFLICT", message, http.StatusConflict)
}

// Validation creates a 422 error with field errors
func Validation(message string, fields map[string]string) *AppError {
	return New("VALIDATION_ERROR", message, http.StatusUnprocessableEntity).WithFields(fields)
}

// Internal creates a 500 error
func Internal(err error) *AppError {
	return Wrap(err, "INTERNAL_ERROR", "An internal error occurred", http.StatusInternalServerError)
}

// ServiceUnavailable creates a 503 error
func ServiceUnavailable(message string) *AppError {
	return New("SERVICE_UNAVAILABLE", message, http.StatusServiceUnavailable)
}

// TooManyRequests creates a 429 error
func TooManyRequests(message string) *AppError {
	return New("TOO_MANY_REQUESTS", message, http.StatusTooManyRequests)
}

// GatewayTimeout creates a 504 error
func GatewayTimeout(message string) *AppError {
	return New("GATEWAY_TIMEOUT", message, http.StatusGatewayTimeout)
}
