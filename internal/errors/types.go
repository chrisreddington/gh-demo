// Package errors provides custom error types for better error handling throughout the application.
// This improves error detection and handling compared to string matching.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// PartialFailureError represents an error where some operations succeeded and some failed.
// This allows callers to distinguish between complete failures and partial failures.
type PartialFailureError struct {
	Errors []string // Individual error messages for failed operations
}

// Error implements the error interface.
func (e *PartialFailureError) Error() string {
	return fmt.Sprintf("some items failed to create:\n  - %s", strings.Join(e.Errors, "\n  - "))
}

// NewPartialFailureError creates a new PartialFailureError with the given error messages.
func NewPartialFailureError(errors []string) *PartialFailureError {
	return &PartialFailureError{Errors: errors}
}

// IsPartialFailure checks if an error is a PartialFailureError.
func IsPartialFailure(err error) bool {
	_, ok := err.(*PartialFailureError)
	return ok
}

// LayeredError provides a structured approach to error handling with layers and operations.
// This allows for easy categorization and handling of errors by their source and type.
type LayeredError struct {
	Layer     string            // "api", "validation", "file", "config", "network"
	Operation string            // "create_issue", "read_file", "parse_json", etc.
	Message   string            // Human-readable error message
	Cause     error             // Underlying error that caused this error
	Context   map[string]string // Optional context (file paths, HTTP codes, etc.)
}

// Error implements the error interface with a consistent format.
func (e *LayeredError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", e.Layer, e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Layer, e.Operation, e.Message)
}

// Unwrap implements the errors.Unwrap interface for error chains.
func (e *LayeredError) Unwrap() error {
	return e.Cause
}

// NewLayeredError creates a new LayeredError with the specified parameters.
func NewLayeredError(layer, operation, message string, cause error) *LayeredError {
	return &LayeredError{
		Layer:     layer,
		Operation: operation,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]string),
	}
}

// WithContext adds context information to the error and returns the modified error.
func (e *LayeredError) WithContext(key, value string) *LayeredError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// Layer-specific convenience functions for common error types

// APIError creates a new LayeredError for API-related operations.
func APIError(operation, message string, cause error) error {
	return NewLayeredError("api", operation, message, cause)
}

// ValidationError creates a new LayeredError for validation failures.
func ValidationError(operation, message string) error {
	return NewLayeredError("validation", operation, message, nil)
}

// FileError creates a new LayeredError for file operations.
func FileError(operation, message string, cause error) error {
	return NewLayeredError("file", operation, message, cause)
}

// ConfigError creates a new LayeredError for configuration-related issues.
func ConfigError(operation, message string, cause error) error {
	return NewLayeredError("config", operation, message, cause)
}

// Error checking and detection functions

// IsLayeredError checks if an error is a LayeredError and returns it if so.
func IsLayeredError(err error) (*LayeredError, bool) {
	var layeredErr *LayeredError
	if errors.As(err, &layeredErr) {
		return layeredErr, true
	}
	return nil, false
}

// IsLayer checks if an error is a LayeredError with the specified layer.
func IsLayer(err error, layer string) bool {
	if layeredErr, ok := IsLayeredError(err); ok {
		return layeredErr.Layer == layer
	}
	return false
}

// IsOperation checks if an error is a LayeredError with the specified operation.
func IsOperation(err error, operation string) bool {
	if layeredErr, ok := IsLayeredError(err); ok {
		return layeredErr.Operation == operation
	}
	return false
}

// ErrorCollector provides a simple way to collect multiple errors and return them appropriately.
type ErrorCollector struct {
	errors    []error
	operation string
}

// NewErrorCollector creates a new ErrorCollector for the specified operation.
func NewErrorCollector(operation string) *ErrorCollector {
	return &ErrorCollector{operation: operation}
}

// Add appends an error to the collection if it's not nil.
func (ec *ErrorCollector) Add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
}

// Result returns nil if no errors were collected, the single error if only one was collected,
// or a PartialFailureError if multiple errors were collected.
func (ec *ErrorCollector) Result() error {
	if len(ec.errors) == 0 {
		return nil
	}
	if len(ec.errors) == 1 {
		return ec.errors[0]
	}

	// Convert to partial failure with error strings
	errorStrings := make([]string, len(ec.errors))
	for i, err := range ec.errors {
		errorStrings[i] = err.Error()
	}
	return NewPartialFailureError(errorStrings)
}
