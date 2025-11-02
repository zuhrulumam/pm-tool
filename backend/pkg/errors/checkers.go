package errors

import "net/http"

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode == http.StatusNotFound || appErr.Code == CodeNotFound
	}
	return false
}

// IsValidation checks if error is a validation error
func IsValidation(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode == http.StatusUnprocessableEntity || 
			appErr.Code == CodeValidationError
	}
	return false
}

// IsUnauthorized checks if error is an unauthorized error
func IsUnauthorized(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode == http.StatusUnauthorized || 
			appErr.Code == CodeUnauthorized
	}
	return false
}

// IsForbidden checks if error is a forbidden error
func IsForbidden(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode == http.StatusForbidden || 
			appErr.Code == CodeForbidden
	}
	return false
}

// IsConflict checks if error is a conflict error
func IsConflict(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode == http.StatusConflict || 
			appErr.Code == CodeConflict || 
			appErr.Code == CodeAlreadyExists
	}
	return false
}

// IsAlreadyExists checks if error is an already exists error
func IsAlreadyExists(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Code == CodeAlreadyExists
	}
	return false
}

// IsInvalidInput checks if error is an invalid input error
func IsInvalidInput(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Code == CodeInvalidInput
	}
	return false
}

// IsDatabaseError checks if error is a database error
func IsDatabaseError(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Code == CodeDatabaseError
	}
	return false
}

// IsCacheError checks if error is a cache error
func IsCacheError(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Code == CodeCacheError
	}
	return false
}

// IsExternalAPIError checks if error is an external API error
func IsExternalAPIError(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Code == CodeExternalAPI
	}
	return false
}

// GetStatusCode returns the HTTP status code for the error
func GetStatusCode(err error) int {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetCode returns the error code
func GetCode(err error) string {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Code
	}
	return "UNKNOWN_ERROR"
}

// GetMessage returns the error message
func GetMessage(err error) string {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Message
	}
	if err != nil {
		return err.Error()
	}
	return "unknown error"
}

// GetFields returns field-level errors
func GetFields(err error) map[string]string {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Fields
	}
	return nil
}
