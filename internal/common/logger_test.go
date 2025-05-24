package common

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	// Test debug mode enabled
	debugLogger := NewLogger(true)
	if debugLogger == nil {
		t.Error("Expected logger to be created")
	}
	if !debugLogger.debug {
		t.Error("Expected debug mode to be enabled")
	}

	// Test debug mode disabled
	infoLogger := NewLogger(false)
	if infoLogger == nil {
		t.Error("Expected logger to be created")
	}
	if infoLogger.debug {
		t.Error("Expected debug mode to be disabled")
	}
}

func TestStandardLogger_Debug(t *testing.T) {
	// Test with debug enabled - should output
	debugLogger := NewLogger(true)
	// We can't easily capture stdout in tests, but we can ensure the method doesn't panic
	debugLogger.Debug("test debug message: %s", "value")

	// Test with debug disabled - should not output
	infoLogger := NewLogger(false)
	infoLogger.Debug("test debug message: %s", "value")
}

func TestStandardLogger_Info(t *testing.T) {
	// Test info logging (always outputs regardless of debug mode)
	debugLogger := NewLogger(true)
	debugLogger.Info("test info message: %s", "value")

	infoLogger := NewLogger(false)
	infoLogger.Info("test info message: %s", "value")
}

func TestStandardLogger_Interface(t *testing.T) {
	// Test that StandardLogger implements the Logger interface
	var logger Logger = NewLogger(true)
	if logger == nil {
		t.Error("Expected logger to implement Logger interface")
	}

	// Test that we can call interface methods
	logger.Debug("debug test")
	logger.Info("info test")
}
