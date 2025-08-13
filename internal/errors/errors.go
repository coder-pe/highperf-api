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

package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents a structured API error
type APIError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	StatusCode int         `json:"-"`
}

func (e APIError) Error() string {
	return e.Message
}

// Common error codes
const (
	CodeInternal        = "INTERNAL_ERROR"
	CodeBadRequest      = "BAD_REQUEST"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeValidation      = "VALIDATION_ERROR"
	CodeTooManyRequests = "TOO_MANY_REQUESTS"
	CodeTimeout         = "TIMEOUT"
)

// Predefined errors
var (
	ErrInternal        = NewAPIError(CodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrBadRequest      = NewAPIError(CodeBadRequest, "Bad request", http.StatusBadRequest)
	ErrUnauthorized    = NewAPIError(CodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden       = NewAPIError(CodeForbidden, "Forbidden", http.StatusForbidden)
	ErrNotFound        = NewAPIError(CodeNotFound, "Resource not found", http.StatusNotFound)
	ErrConflict        = NewAPIError(CodeConflict, "Resource conflict", http.StatusConflict)
	ErrTooManyRequests = NewAPIError(CodeTooManyRequests, "Too many requests", http.StatusTooManyRequests)
	ErrTimeout         = NewAPIError(CodeTimeout, "Request timeout", http.StatusGatewayTimeout)
)

// NewAPIError creates a new API error
func NewAPIError(code, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to an API error
func (e *APIError) WithDetails(details interface{}) *APIError {
	return &APIError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
		StatusCode: e.StatusCode,
	}
}

// WithMessage overrides the error message
func (e *APIError) WithMessage(message string) *APIError {
	return &APIError{
		Code:       e.Code,
		Message:    message,
		Details:    e.Details,
		StatusCode: e.StatusCode,
	}
}

// ValidationError creates a validation error with details
func ValidationError(details interface{}) *APIError {
	return &APIError{
		Code:       CodeValidation,
		Message:    "Validation failed",
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// UserNotFoundError creates a user not found error
func UserNotFoundError(userID interface{}) *APIError {
	return &APIError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("User not found"),
		Details:    map[string]interface{}{"user_id": userID},
		StatusCode: http.StatusNotFound,
	}
}

// UserAlreadyExistsError creates a user already exists error
func UserAlreadyExistsError(email string) *APIError {
	return &APIError{
		Code:       CodeConflict,
		Message:    "User already exists",
		Details:    map[string]interface{}{"email": email},
		StatusCode: http.StatusConflict,
	}
}

// InvalidCredentialsError creates an invalid credentials error
func InvalidCredentialsError() *APIError {
	return &APIError{
		Code:       CodeUnauthorized,
		Message:    "Invalid credentials",
		StatusCode: http.StatusUnauthorized,
	}
}

// InsufficientPermissionsError creates an insufficient permissions error
func InsufficientPermissionsError(permission string) *APIError {
	return &APIError{
		Code:       CodeForbidden,
		Message:    "Insufficient permissions",
		Details:    map[string]interface{}{"required_permission": permission},
		StatusCode: http.StatusForbidden,
	}
}

// FromError converts a standard error to an APIError
func FromError(err error) *APIError {
	if err == nil {
		return nil
	}

	// If it's already an APIError, return it
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	// Check for common repository errors
	switch {
	case errors.Is(err, errors.New("user not found")):
		return ErrNotFound.WithMessage("User not found")
	case errors.Is(err, errors.New("user already exists")):
		return ErrConflict.WithMessage("User already exists")
	default:
		// Return generic internal error for unknown errors
		return ErrInternal
	}
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *APIError) *ErrorResponse {
	return &ErrorResponse{
		Error: *err,
	}
}

// HealthError represents a health check error
type HealthError struct {
	Component string `json:"component"`
	Error     string `json:"error"`
	Status    string `json:"status"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string         `json:"status"`
	Version string         `json:"version,omitempty"`
	Checks  []*HealthError `json:"checks,omitempty"`
}

// NewHealthResponse creates a new health response
func NewHealthResponse(status, version string, checks []*HealthError) *HealthResponse {
	return &HealthResponse{
		Status:  status,
		Version: version,
		Checks:  checks,
	}
}