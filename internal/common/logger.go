package common

import (
	"fmt"
)

// StandardLogger is a concrete implementation of the Logger interface
type StandardLogger struct {
	debug bool
}

// NewLogger creates a new logger with the specified debug mode
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
