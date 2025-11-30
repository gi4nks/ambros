// Package errors provides standardized error handling for the Ambros application.
package errors

import (
	"fmt"
)

// Error codes for the application
const (
	// Command related errors
	ErrCommandNotFound = "command_not_found"
	ErrCommandExists   = "command_exists"
	ErrInvalidCommand  = "invalid_command"
	ErrExecutionFailed = "execution_failed"

	// Repository related errors
	ErrRepositoryOpen  = "repository_open"
	ErrRepositoryClose = "repository_close"
	ErrRepositoryWrite = "repository_write"
	ErrRepositoryRead  = "repository_read"

	// Configuration related errors
	ErrConfigInvalid  = "config_invalid"
	ErrConfigNotFound = "config_not_found"

	// Scheduler related errors
	ErrScheduleInvalid  = "schedule_invalid"
	ErrScheduleConflict = "schedule_conflict"

	// API related errors
	ErrInvalidRequest = "invalid_request"
	ErrUnauthorized   = "unauthorized"
	ErrInternalServer = "internal_server"
	ErrNotFound       = "not_found"
)

// AppError represents an application-specific error
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewError creates a new AppError
func NewError(code string, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		return false
	}
	return appErr.Code == ErrCommandNotFound || appErr.Code == ErrConfigNotFound
}

// IsInvalidInput returns true if the error is related to invalid input
func IsInvalidInput(err error) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		return false
	}
	return appErr.Code == ErrInvalidCommand ||
		appErr.Code == ErrConfigInvalid ||
		appErr.Code == ErrScheduleInvalid ||
		appErr.Code == ErrInvalidRequest
}

// IsInternalError returns true if the error is an internal error
func IsInternalError(err error) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		return false
	}
	return appErr.Code == ErrRepositoryOpen ||
		appErr.Code == ErrRepositoryClose ||
		appErr.Code == ErrRepositoryWrite ||
		appErr.Code == ErrRepositoryRead ||
		appErr.Code == ErrInternalServer ||
		appErr.Code == ErrExecutionFailed
}
