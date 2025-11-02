package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	apperrors "github.com/zuhrulumam/pm-tool/pkg/errors"
)

// Validator interface
type Validator interface {
	Validate(v interface{}) error
	ValidateStruct(v interface{}) map[string]string
	RegisterCustom(tag string, fn validator.Func)
}

// validatorImpl implements Validator interface
type validatorImpl struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() Validator {
	v := validator.New()

	// Register custom tag name func to use json tag names in errors
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators(v)

	return &validatorImpl{validate: v}
}

// Validate validates a struct and returns an AppError if validation fails
func (v *validatorImpl) Validate(value interface{}) error {
	if err := v.validate.Struct(value); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fields := make(map[string]string)
			for _, e := range validationErrors {
				fields[e.Field()] = msgForTag(e)
			}
			return apperrors.Validation("Validation failed", fields)
		}
		return apperrors.BadRequest(err.Error())
	}
	return nil
}

// ValidateStruct validates a struct and returns field errors
func (v *validatorImpl) ValidateStruct(value interface{}) map[string]string {
	fields := make(map[string]string)
	
	if err := v.validate.Struct(value); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				fields[e.Field()] = msgForTag(e)
			}
		}
	}
	
	return fields
}

// RegisterCustom registers a custom validator
func (v *validatorImpl) RegisterCustom(tag string, fn validator.Func) {
	v.validate.RegisterValidation(tag, fn)
}

// msgForTag returns a user-friendly message for a validation tag
func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// Global validator instance
var globalValidator Validator = New()

// SetGlobal sets the global validator instance
func SetGlobal(v Validator) {
	globalValidator = v
}

// Validate validates using the global validator
func Validate(v interface{}) error {
	return globalValidator.Validate(v)
}

// ValidateStruct validates using the global validator
func ValidateStruct(v interface{}) map[string]string {
	return globalValidator.ValidateStruct(v)
}
