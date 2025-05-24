// Package common provides shared utilities and interfaces used across the application.
// This includes logging interfaces and implementations for debug and informational output.
package common

import (
	"fmt"
)

// StandardLogger is a concrete implementation of the Logger interface.
// It provides debug and info logging capabilities with configurable debug mode.
type StandardLogger struct {
	debug bool // Whether debug messages should be printed
}

// NewLogger creates a new logger with the specified debug mode.
// When debug is true, debug messages will be printed to stderr with [DEBUG] prefix.
// Info messages are always printed to stdout.
func NewLogger(debug bool) *StandardLogger {
	return &StandardLogger{debug: debug}
}

// Debug logs a message only when debug mode is enabled
func (l *StandardLogger) Debug(format string, args ...interface{}) {
	if l.debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Info logs a message always
func (l *StandardLogger) Info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
