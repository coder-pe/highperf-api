/*
 * Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Use JSON tag names in validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	_ = validate.RegisterValidation("password", validatePassword)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   any    `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}
	
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// Validate validates a struct and returns formatted validation errors
func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var errs []ValidationError
		
		for _, err := range validationErrors {
			errs = append(errs, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   err.Value(),
				Message: getErrorMessage(err),
			})
		}
		
		return ValidationErrors{Errors: errs}
	}

	return fmt.Errorf("validation failed: %w", err)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// getErrorMessage returns a human-readable error message for validation errors
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "password":
		return fmt.Sprintf("%s must contain at least 8 characters with uppercase, lowercase, number and special character", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// validatePassword is a custom validator for password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}
	
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' || 
			 char == '%' || char == '^' || char == '&' || char == '*' ||
			 char == '(' || char == ')' || char == '-' || char == '_' ||
			 char == '+' || char == '=' || char == '{' || char == '}' ||
			 char == '[' || char == ']' || char == '|' || char == '\\' ||
			 char == ':' || char == ';' || char == '"' || char == '\'' ||
			 char == '<' || char == '>' || char == ',' || char == '.' ||
			 char == '?' || char == '/' || char == '`' || char == '~':
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}