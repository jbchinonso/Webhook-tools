// Package validation wraps go-playground/validator so modules can bind and
// validate a request body in one call. Field errors are returned as a map
// suitable for the response envelope's error.details.
package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = newValidator()

func newValidator() *validator.Validate {
	val := validator.New(validator.WithRequiredStructEnabled())

	// Use the `json` tag name in error output rather than the Go field name.
	val.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	// Custom `password` tag. The field value is the plaintext; the rule set
	// lives in password.go so the policy is in one place.
	_ = val.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		ok, _ := checkPassword(fl.Field().String())
		return ok
	})

	return val
}

// Errors is the field-keyed map returned to clients under error.details.
type Errors map[string][]string

// Validate runs struct-tag validation and returns a map of field → messages.
// Returns nil if the struct is valid.
func Validate(s any) Errors {
	if err := v.Struct(s); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			return Errors{"_": {err.Error()}}
		}
		out := Errors{}
		for _, fe := range ve {
			out[fe.Field()] = append(out[fe.Field()], message(fe))
		}
		return out
	}
	return nil
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "password":
	_, reason := checkPassword(fe.Value().(string))
	if reason == "" {
		reason = "does not meet the password policy"
	}
	return reason
	default:
		return fmt.Sprintf("failed %s validation", fe.Tag())
	}
}