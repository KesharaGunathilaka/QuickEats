package httputil

import (
	"fmt"
	"net/http"
)

// AppError is the standard application error that maps to the JSON error envelope.
type AppError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Status  int         `json:"-"`
	Details interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Standard error constructors
func ErrValidation(msg string, details interface{}) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: msg, Status: http.StatusBadRequest, Details: details}
}

func ErrUnauthenticated(msg string) *AppError {
	return &AppError{Code: "UNAUTHENTICATED", Message: msg, Status: http.StatusUnauthorized}
}

func ErrForbidden(msg string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: msg, Status: http.StatusForbidden}
}

func ErrNotFound(resource string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: fmt.Sprintf("%s not found", resource), Status: http.StatusNotFound}
}

func ErrConflict(msg string) *AppError {
	return &AppError{Code: "CONFLICT", Message: msg, Status: http.StatusConflict}
}

func ErrBusinessRule(code, msg string) *AppError {
	return &AppError{Code: code, Message: msg, Status: http.StatusUnprocessableEntity}
}

func ErrInternal(msg string) *AppError {
	return &AppError{Code: "INTERNAL_ERROR", Message: msg, Status: http.StatusInternalServerError}
}

func ErrServiceUnavailable(msg string) *AppError {
	return &AppError{Code: "SERVICE_UNAVAILABLE", Message: msg, Status: http.StatusServiceUnavailable}
}
