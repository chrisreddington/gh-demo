package errors

import (
	"fmt"
	"strings"
	"testing"
)

// TestLayeredError_Error tests the Error() method formatting
func TestLayeredError_Error(t *testing.T) {
	tests := []struct {
		name     string
		layer    string
		operation string
		message  string
		cause    error
		expected string
	}{
		{
			name:      "error without cause",
			layer:     "validation",
			operation: "validate_flags",
			message:   "owner cannot be empty",
			cause:     nil,
			expected:  "[validation:validate_flags] owner cannot be empty",
		},
		{
			name:      "error with cause",
			layer:     "api",
			operation: "create_issue",
			message:   "failed to create GitHub issue",
			cause:     fmt.Errorf("HTTP 422: Validation Failed"),
			expected:  "[api:create_issue] failed to create GitHub issue: HTTP 422: Validation Failed",
		},
		{
			name:      "file error with cause",
			layer:     "file",
			operation: "read_labels",
			message:   "failed to read labels file",
			cause:     fmt.Errorf("no such file or directory"),
			expected:  "[file:read_labels] failed to read labels file: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewLayeredError(tt.layer, tt.operation, tt.message, tt.cause)
			if err.Error() != tt.expected {
				t.Errorf("LayeredError.Error() = %q, want %q", err.Error(), tt.expected)
			}
		})
	}
}

// TestLayeredError_Unwrap tests the Unwrap() method
func TestLayeredError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewLayeredError("api", "test", "test message", cause)
	
	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("LayeredError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	errNoCause := NewLayeredError("api", "test", "test message", nil)
	if unwrapped := errNoCause.Unwrap(); unwrapped != nil {
		t.Errorf("LayeredError.Unwrap() with no cause = %v, want nil", unwrapped)
	}
}

// TestLayeredError_WithContext tests the WithContext() method
func TestLayeredError_WithContext(t *testing.T) {
	err := NewLayeredError("file", "read_config", "file not found", nil)
	
	// Add context
	err.WithContext("path", "/path/to/config.json")
	err.WithContext("expected_format", "JSON")
	
	if err.Context["path"] != "/path/to/config.json" {
		t.Errorf("Expected context path = '/path/to/config.json', got %q", err.Context["path"])
	}
	if err.Context["expected_format"] != "JSON" {
		t.Errorf("Expected context expected_format = 'JSON', got %q", err.Context["expected_format"])
	}
}

// TestConvenienceFunctions tests the layer-specific error creation functions
func TestConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() error
		layer    string
		operation string
	}{
		{
			name:      "APIError",
			fn:        func() error { return APIError("create_issue", "failed to create issue", fmt.Errorf("HTTP 422")) },
			layer:     "api",
			operation: "create_issue",
		},
		{
			name:      "ValidationError",
			fn:        func() error { return ValidationError("validate_flags", "owner cannot be empty") },
			layer:     "validation",
			operation: "validate_flags",
		},
		{
			name:      "FileError",
			fn:        func() error { return FileError("read_file", "file not found", fmt.Errorf("no such file")) },
			layer:     "file",
			operation: "read_file",
		},
		{
			name:      "ConfigError",
			fn:        func() error { return ConfigError("parse_config", "invalid JSON", fmt.Errorf("unexpected token")) },
			layer:     "config",
			operation: "parse_config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			layeredErr, ok := IsLayeredError(err)
			if !ok {
				t.Fatalf("Expected LayeredError, got %T", err)
			}
			if layeredErr.Layer != tt.layer {
				t.Errorf("Expected layer %q, got %q", tt.layer, layeredErr.Layer)
			}
			if layeredErr.Operation != tt.operation {
				t.Errorf("Expected operation %q, got %q", tt.operation, layeredErr.Operation)
			}
		})
	}
}

// TestIsLayeredError tests error detection
func TestIsLayeredError(t *testing.T) {
	layeredErr := NewLayeredError("api", "test", "test message", nil)
	regularErr := fmt.Errorf("regular error")

	// Test with LayeredError
	if detected, ok := IsLayeredError(layeredErr); !ok || detected != layeredErr {
		t.Error("IsLayeredError should detect LayeredError")
	}

	// Test with regular error
	if _, ok := IsLayeredError(regularErr); ok {
		t.Error("IsLayeredError should not detect regular error")
	}

	// Test with nil
	if _, ok := IsLayeredError(nil); ok {
		t.Error("IsLayeredError should not detect nil error")
	}
}

// TestIsLayer tests layer-specific error detection
func TestIsLayer(t *testing.T) {
	apiErr := APIError("test", "test message", nil)
	validationErr := ValidationError("test", "test message")
	regularErr := fmt.Errorf("regular error")

	if !IsLayer(apiErr, "api") {
		t.Error("IsLayer should detect api layer")
	}
	if IsLayer(apiErr, "validation") {
		t.Error("IsLayer should not detect wrong layer")
	}
	if !IsLayer(validationErr, "validation") {
		t.Error("IsLayer should detect validation layer")
	}
	if IsLayer(regularErr, "api") {
		t.Error("IsLayer should not detect layer in regular error")
	}
}

// TestIsOperation tests operation-specific error detection
func TestIsOperation(t *testing.T) {
	err1 := APIError("create_issue", "test message", nil)
	err2 := ValidationError("validate_flags", "test message")
	regularErr := fmt.Errorf("regular error")

	if !IsOperation(err1, "create_issue") {
		t.Error("IsOperation should detect create_issue operation")
	}
	if IsOperation(err1, "validate_flags") {
		t.Error("IsOperation should not detect wrong operation")
	}
	if !IsOperation(err2, "validate_flags") {
		t.Error("IsOperation should detect validate_flags operation")
	}
	if IsOperation(regularErr, "create_issue") {
		t.Error("IsOperation should not detect operation in regular error")
	}
}

// TestErrorCollector tests the ErrorCollector functionality
func TestErrorCollector(t *testing.T) {
	collector := NewErrorCollector("test_operation")

	// Test empty collector
	if err := collector.Result(); err != nil {
		t.Errorf("Empty collector should return nil, got %v", err)
	}

	// Test single error
	singleErr := fmt.Errorf("single error")
	collector.Add(singleErr)
	if err := collector.Result(); err != singleErr {
		t.Errorf("Single error collector should return the error, got %v", err)
	}

	// Test multiple errors
	collector.Add(fmt.Errorf("second error"))
	result := collector.Result()
	if result == nil {
		t.Error("Multiple errors should not return nil")
	}
	if !IsPartialFailure(result) {
		t.Error("Multiple errors should return PartialFailureError")
	}

	// Check that the error message contains both errors
	errMsg := result.Error()
	if !strings.Contains(errMsg, "single error") || !strings.Contains(errMsg, "second error") {
		t.Errorf("Partial failure should contain all error messages, got: %s", errMsg)
	}
}

// TestErrorCollector_IgnoreNil tests that nil errors are ignored
func TestErrorCollector_IgnoreNil(t *testing.T) {
	collector := NewErrorCollector("test_operation")
	
	collector.Add(nil)
	collector.Add(fmt.Errorf("real error"))
	collector.Add(nil)
	
	result := collector.Result()
	if result == nil {
		t.Error("Should return the non-nil error")
	}
	
	// Should return single error, not partial failure
	if IsPartialFailure(result) {
		t.Error("Should return single error when only one non-nil error exists")
	}
	
	if result.Error() != "real error" {
		t.Errorf("Expected 'real error', got %q", result.Error())
	}
}

// TestPartialFailureError_Compatibility tests that existing PartialFailureError still works
func TestPartialFailureError_Compatibility(t *testing.T) {
	errors := []string{"error 1", "error 2", "error 3"}
	err := NewPartialFailureError(errors)
	
	if !IsPartialFailure(err) {
		t.Error("IsPartialFailure should detect PartialFailureError")
	}
	
	errMsg := err.Error()
	if !strings.Contains(errMsg, "some items failed to create") {
		t.Error("PartialFailureError should contain expected message")
	}
	
	for _, errStr := range errors {
		if !strings.Contains(errMsg, errStr) {
			t.Errorf("PartialFailureError should contain error: %s", errStr)
		}
	}
}