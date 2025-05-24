package common

// Logger interface defines the contract for debug and info logging across all packages.
// This interface enables consistent logging behavior throughout the application and
// facilitates testing by allowing mock logger implementations.
type Logger interface {
	// Debug logs a debug message with printf-style formatting.
	// Debug messages are only shown when debug mode is enabled.
	Debug(format string, args ...interface{})
	// Info logs an informational message with printf-style formatting.
	// Info messages are always shown regardless of debug mode.
	Info(format string, args ...interface{})
}
