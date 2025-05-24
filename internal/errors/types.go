// Package errors provides custom error types for better error handling throughout the application.
// This improves error detection and handling compared to string matching.
package errors

import (
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