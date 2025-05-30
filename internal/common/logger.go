// Package common provides shared utilities and interfaces used across the application.
// This includes logging interfaces and implementations for debug and informational output.
package common

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// StandardLogger is a concrete implementation of the Logger interface.
// It provides debug and info logging capabilities with configurable debug mode.
type StandardLogger struct {
	debug     bool   // Whether debug messages should be printed
	requestID string // Request ID for tracing operations
}

// GenerateRequestID generates a simple request ID for operation tracing.
func GenerateRequestID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("req_%d", r.Intn(100000))
}

// NewLogger creates a new logger with the specified debug mode.
// When debug is true, debug messages will be printed to stderr with [DEBUG] prefix.
// Info messages are always printed to stdout.
func NewLogger(debug bool) *StandardLogger {
	return &StandardLogger{
		debug:     debug,
		requestID: GenerateRequestID(),
	}
}

// Debug logs a message only when debug mode is enabled
func (l *StandardLogger) Debug(format string, args ...interface{}) {
	if l.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] [%s] "+format+"\n", append([]interface{}{l.requestID}, args...)...)
	}
}

// Info logs a message always
func (l *StandardLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[%s] "+format+"\n", append([]interface{}{l.requestID}, args...)...)
}

// FormatCreationError creates a standardized error message for failed creation operations.
// This ensures consistent error formatting across different object types.
func FormatCreationError(itemType, title string, index int, err error) string {
	return fmt.Sprintf("%s %d (%s): %v", itemType, index+1, title, err)
}

// FormatOperationContext creates a standardized context string for operations.
// This provides consistent operation naming across the application.
func FormatOperationContext(operation, itemType string) string {
	return fmt.Sprintf("%s_%s", operation, strings.ToLower(strings.ReplaceAll(itemType, " ", "_")))
}
