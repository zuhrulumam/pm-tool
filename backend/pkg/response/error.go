package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/zuhrulumam/pm-tool/pkg/errors"
)

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Error sends an error response
func Errorf(c *gin.Context, err error) {
	// Check if it's an AppError
	if appErr, ok := apperrors.GetAppError(err); ok {
		c.JSON(appErr.StatusCode, Response{
			Success: false,
			Error: &Error{
				Code:    appErr.Code,
				Message: appErr.Message,
				Fields:  appErr.Fields,
			},
		})
		return
	}

	// Generic error
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &Error{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
		},
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &Error{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &Error{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &Error{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &Error{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// ValidationError sends a 422 Unprocessable Entity response
func ValidationError(c *gin.Context, message string, fields map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error: &Error{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Fields:  fields,
		},
	})
}
