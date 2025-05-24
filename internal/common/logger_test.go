package common

import (
	"testing"
)

// TestNewLogger tests logger constructor with different configurations
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		debugMode   bool
		expectedDebug bool
	}{
		{
			name:        "debug mode enabled",
			debugMode:   true,
			expectedDebug: true,
		},
		{
			name:        "debug mode disabled", 
			debugMode:   false,
			expectedDebug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.debugMode)
			if logger.debug != tt.expectedDebug {
				t.Errorf("Expected debug mode to be %v, got %v", tt.expectedDebug, logger.debug)
			}

			// Test that logger implements the Logger interface
			var _ Logger = logger
		})
	}
}

// TestStandardLogger_Methods tests logger method behavior with different configurations
func TestStandardLogger_Methods(t *testing.T) {
	tests := []struct {
		name      string
		debugMode bool
		testFunc  func(t *testing.T, logger Logger)
	}{
		{
			name:      "debug method with debug enabled",
			debugMode: true,
			testFunc: func(t *testing.T, logger Logger) {
				// We can't easily capture stdout in tests, but we can ensure the method doesn't panic
				logger.Debug("test debug message: %s", "value")
			},
		},
		{
			name:      "debug method with debug disabled",
			debugMode: false,
			testFunc: func(t *testing.T, logger Logger) {
				// Should not output but also shouldn't panic
				logger.Debug("test debug message: %s", "value")
			},
		},
		{
			name:      "info method with debug enabled",
			debugMode: true,
			testFunc: func(t *testing.T, logger Logger) {
				// Info logging always outputs regardless of debug mode
				logger.Info("test info message: %s", "value")
			},
		},
		{
			name:      "info method with debug disabled",
			debugMode: false,
			testFunc: func(t *testing.T, logger Logger) {
				// Info logging always outputs regardless of debug mode
				logger.Info("test info message: %s", "value")
			},
		},
		{
			name:      "interface compliance",
			debugMode: true,
			testFunc: func(t *testing.T, logger Logger) {
				// Test that we can call interface methods
				logger.Debug("debug test")
				logger.Info("info test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.debugMode)
			tt.testFunc(t, logger)
		})
	}
}
