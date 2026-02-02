package errors

import (
	"errors"
	"fmt"
)

// Standard error types for better error handling and HTTP status mapping

// ValidationError indicates that input validation failed
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation error on field '%s': %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// NotFoundError indicates that a requested resource was not found
type NotFoundError struct {
	Resource string
	ID       interface{}
	Err      error
}

func (e *NotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s not found (ID: %v): %v", e.Resource, e.ID, e.Err)
	}
	return fmt.Sprintf("%s not found (ID: %v)", e.Resource, e.ID)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(resource string, id interface{}, err error) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
		Err:      err,
	}
}

// ConflictError indicates that the operation conflicts with existing state
type ConflictError struct {
	Resource string
	Message  string
	Err      error
}

func (e *ConflictError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("conflict on %s: %s: %v", e.Resource, e.Message, e.Err)
	}
	return fmt.Sprintf("conflict on %s: %s", e.Resource, e.Message)
}

func (e *ConflictError) Unwrap() error {
	return e.Err
}

func NewConflictError(resource, message string, err error) *ConflictError {
	return &ConflictError{
		Resource: resource,
		Message:  message,
		Err:      err,
	}
}

// TimeoutError indicates that an operation exceeded its time limit
type TimeoutError struct {
	Operation string
	Duration  string
	Err       error
}

func (e *TimeoutError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("timeout during %s after %s: %v", e.Operation, e.Duration, e.Err)
	}
	return fmt.Sprintf("timeout during %s after %s", e.Operation, e.Duration)
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

func NewTimeoutError(operation, duration string, err error) *TimeoutError {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration,
		Err:       err,
	}
}

// AuthorizationError indicates that the user is not authorized
type AuthorizationError struct {
	User      string
	Resource  string
	Operation string
	Err       error
}

func (e *AuthorizationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("user '%s' not authorized to %s on %s: %v", e.User, e.Operation, e.Resource, e.Err)
	}
	return fmt.Sprintf("user '%s' not authorized to %s on %s", e.User, e.Operation, e.Resource)
}

func (e *AuthorizationError) Unwrap() error {
	return e.Err
}

func NewAuthorizationError(user, resource, operation string, err error) *AuthorizationError {
	return &AuthorizationError{
		User:      user,
		Resource:  resource,
		Operation: operation,
		Err:       err,
	}
}

// Helper functions to check error types

func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

func IsNotFoundError(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

func IsConflictError(err error) bool {
	var ce *ConflictError
	return errors.As(err, &ce)
}

func IsTimeoutError(err error) bool {
	var te *TimeoutError
	return errors.As(err, &te)
}

func IsAuthorizationError(err error) bool {
	var ae *AuthorizationError
	return errors.As(err, &ae)
}

// GetHTTPStatus returns the appropriate HTTP status code for an error
func GetHTTPStatus(err error) int {
	switch {
	case IsValidationError(err):
		return 400 // Bad Request
	case IsNotFoundError(err):
		return 404 // Not Found
	case IsConflictError(err):
		return 409 // Conflict
	case IsTimeoutError(err):
		return 408 // Request Timeout
	case IsAuthorizationError(err):
		return 403 // Forbidden
	default:
		return 500 // Internal Server Error
	}
}
