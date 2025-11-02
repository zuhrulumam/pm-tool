package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// registerCustomValidators registers custom validation functions
func registerCustomValidators(v *validator.Validate) {
	v.RegisterValidation("phone", validatePhone)
}

// validatePhone validates phone numbers (E.164 format)
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Let 'required' tag handle empty values
	}
	
	// E.164 format: +[country code][number]
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}
