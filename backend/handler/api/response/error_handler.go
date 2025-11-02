package response

import (
	"os"
	
	"github.com/gin-gonic/gin"
	apperr "github.com/zuhrulumam/pm-tool/pkg/errors"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code"`
	Details interface{}       `json:"details,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// HandleError handles AppError and returns appropriate JSON response
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	
	// Try to extract AppError
	appErr, ok := apperr.GetAppError(err)
	if !ok {
		// Not an AppError - treat as internal error
		c.JSON(500, ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}
	
	// Build response from AppError
	resp := ErrorResponse{
		Error: appErr.Message,
		Code:  appErr.Code,
	}
	
	// Include internal details only in dev environment
	env := os.Getenv("APP_ENV")
	if env == "development" || env == "dev" {
		if appErr.Internal != nil {
			resp.Details = appErr.StackTrace()
		}
	}
	
	// Include field errors if validation error
	if appErr.Fields != nil && len(appErr.Fields) > 0 {
		resp.Fields = appErr.Fields
	}
	
	c.JSON(appErr.StatusCode, resp)
}
