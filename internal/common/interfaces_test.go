package common

import (
	"testing"
)

// MockTestLogger implements the Logger interface for testing
type MockTestLogger struct {
	debugCalls []string
	infoCalls  []string
}

// Debug implements Logger interface
func (m *MockTestLogger) Debug(format string, args ...interface{}) {
	m.debugCalls = append(m.debugCalls, format)
}

// Info implements Logger interface  
func (m *MockTestLogger) Info(format string, args ...interface{}) {
	m.infoCalls = append(m.infoCalls, format)
}

// TestLoggerInterface tests that our Logger interface contract is working
func TestLoggerInterface(t *testing.T) {
	// Test that our mock implements the interface
	var logger Logger = &MockTestLogger{}
	
	logger.Debug("debug message: %s", "test")
	logger.Info("info message: %d", 42)
	
	// Verify calls were recorded
	mock := logger.(*MockTestLogger)
	if len(mock.debugCalls) != 1 {
		t.Errorf("Expected 1 debug call, got %d", len(mock.debugCalls))
	}
	if len(mock.infoCalls) != 1 {
		t.Errorf("Expected 1 info call, got %d", len(mock.infoCalls))
	}
	
	// Verify message formats
	if mock.debugCalls[0] != "debug message: %s" {
		t.Errorf("Expected debug format 'debug message: %%s', got %s", mock.debugCalls[0])
	}
	if mock.infoCalls[0] != "info message: %d" {
		t.Errorf("Expected info format 'info message: %%d', got %s", mock.infoCalls[0])
	}
}

// TestNewLoggerInterface tests that NewLogger returns a valid Logger implementation
func TestNewLoggerInterface(t *testing.T) {
	tests := []struct {
		name      string
		debugMode bool
	}{
		{
			name:      "debug mode enabled interface test",
			debugMode: true,
		},
		{
			name:      "debug mode disabled interface test",
			debugMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.debugMode)
			if logger == nil {
				t.Error("NewLogger should return non-nil logger")
				return
			}
			
			// Test that it implements the Logger interface
			var _ Logger = logger
			
			// Test that we can call methods without panic
			logger.Debug("test debug")
			logger.Info("test info")
		})
	}
}

// TestLoggerInterfaceCompliance tests that different logger implementations comply with interface
func TestLoggerInterfaceCompliance(t *testing.T) {
	loggers := []struct {
		name   string
		logger Logger
	}{
		{
			name:   "MockTestLogger",
			logger: &MockTestLogger{},
		},
		{
			name:   "NewLogger(false) interface test",
			logger: NewLogger(false),
		},
		{
			name:   "NewLogger(true) interface test",
			logger: NewLogger(true),
		},
	}

	for _, tt := range loggers {
		t.Run(tt.name, func(t *testing.T) {
			if tt.logger == nil {
				t.Error("Logger should not be nil")
				return
			}
			
			// Test that calling methods doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Logger methods should not panic, got: %v", r)
				}
			}()
			
			tt.logger.Debug("test debug message")
			tt.logger.Info("test info message")
		})
	}
}