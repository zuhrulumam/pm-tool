package errors

import "fmt"

// Domain error codes
const (
	CodeValidationError = "VALIDATION_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeAlreadyExists   = "ALREADY_EXISTS"
	CodeInvalidInput    = "INVALID_INPUT"
	CodeDependency      = "DEPENDENCY_ERROR"
	CodeDatabaseError   = "DATABASE_ERROR"
	CodeCacheError      = "CACHE_ERROR"
	CodeQueueError      = "QUEUE_ERROR"
	CodeExternalAPI     = "EXTERNAL_API_ERROR"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeConflict        = "CONFLICT"
	CodeInternal        = "INTERNAL_ERROR"
)

// Domain-specific error constructors with stack traces

// ValidationError creates a validation error (400)
func ValidationError(message string) *AppError {
	return New(CodeValidationError, message, 400)
}

// NotFoundError creates a not found error (404)
func NotFoundError(entity string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", entity), 404)
}

// AlreadyExistsError creates an already exists error (409)
func AlreadyExistsError(field string) *AppError {
	return New(CodeAlreadyExists, fmt.Sprintf("%s already exists", field), 409)
}

// InvalidInputError creates an invalid input error (400)
func InvalidInputError(message string) *AppError {
	return New(CodeInvalidInput, message, 400)
}

// DatabaseError wraps a database error (500)
func DatabaseError(err error, operation string) *AppError {
	return Wrap(err, CodeDatabaseError, 
		fmt.Sprintf("database error during %s", operation), 500)
}

// CacheError wraps a cache error (500)
func CacheError(err error, operation string) *AppError {
	return Wrap(err, CodeCacheError, 
		fmt.Sprintf("cache error during %s", operation), 500)
}

// QueueError wraps a queue error (500)
func QueueError(err error, operation string) *AppError {
	return Wrap(err, CodeQueueError, 
		fmt.Sprintf("queue error during %s", operation), 500)
}

// ExternalAPIError wraps an external API error (502)
func ExternalAPIError(err error, service string) *AppError {
	return Wrap(err, CodeExternalAPI, 
		fmt.Sprintf("external API error: %s", service), 502)
}

// DependencyError wraps a dependency error (500)
func DependencyError(err error, dependency string) *AppError {
	return Wrap(err, CodeDependency, 
		fmt.Sprintf("dependency error: %s", dependency), 500)
}

// UnauthorizedError creates an unauthorized error (401)
func UnauthorizedError(message string) *AppError {
	if message == "" {
		message = "unauthorized access"
	}
	return New(CodeUnauthorized, message, 401)
}

// ForbiddenError creates a forbidden error (403)
func ForbiddenError(message string) *AppError {
	if message == "" {
		message = "access forbidden"
	}
	return New(CodeForbidden, message, 403)
}

// ConflictError creates a conflict error (409)
func ConflictError(message string) *AppError {
	return New(CodeConflict, message, 409)
}

// InternalError wraps an internal error (500)
func InternalError(err error, message string) *AppError {
	if message == "" {
		message = "internal server error"
	}
	return Wrap(err, CodeInternal, message, 500)
}
