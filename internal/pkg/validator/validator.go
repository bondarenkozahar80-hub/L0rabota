package validator

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	v := validator.New()
	// Регистрация кастомных валидаций
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) ValidateStruct(s interface{}) error {
	if err := cv.validator.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				return fmt.Errorf("field %s: %s", fieldError.Field(), getValidationMessage(fieldError))
			}
		}
		return err
	}
	return nil
}

func getValidationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", fieldError.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fieldError.Param())
	default:
		return fmt.Sprintf("failed %s validation", fieldError.Tag())
	}
}
