package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// RequestValidator validates request body using struct tags
func RequestValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validation happens in ShouldBindJSON, but we can add custom logic here
		c.Next()
	}
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// FormatValidationError formats validator errors into readable messages
func FormatValidationError(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = field + " must be a valid email"
			case "min":
				errors[field] = field + " must be at least " + e.Param()
			case "max":
				errors[field] = field + " must be at most " + e.Param()
			default:
				errors[field] = field + " validation failed on " + e.Tag()
			}
		}
	}
	
	return errors
}
