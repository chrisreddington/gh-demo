package errors

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// TestLayeredError_Error tests the Error() method formatting
func TestLayeredError_Error(t *testing.T) {
	tests := []struct {
		name      string
		layer     string
		operation string
		message   string
		cause     error
		expected  string
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
	err = err.WithContext("path", "/path/to/config.json")
	err = err.WithContext("expected_format", "JSON")

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
		name      string
		fn        func() error
		layer     string
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

// TestIsContextError tests context error detection
func TestIsContextError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "regular error",
			err:      fmt.Errorf("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsContextError(tt.err)
			if result != tt.expected {
				t.Errorf("IsContextError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestContextError tests context error wrapping
func TestContextError(t *testing.T) {
	tests := []struct {
		name          string
		operation     string
		cause         error
		expectedLayer string
		expectedOp    string
		expectedMsg   string
	}{
		{
			name:          "context canceled",
			operation:     "create_issue",
			cause:         context.Canceled,
			expectedLayer: "context",
			expectedOp:    "create_issue",
			expectedMsg:   "operation was cancelled (interrupted by user)",
		},
		{
			name:          "context deadline exceeded",
			operation:     "fetch_labels",
			cause:         context.DeadlineExceeded,
			expectedLayer: "context",
			expectedOp:    "fetch_labels",
			expectedMsg:   "operation timed out (exceeded 30 second limit)",
		},
		{
			name:      "non-context error",
			operation: "test_op",
			cause:     fmt.Errorf("regular error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContextError(tt.operation, tt.cause)

			if tt.expectedLayer == "" {
				// Non-context error should be returned as-is
				if result != tt.cause {
					t.Errorf("ContextError() should return original error for non-context errors, got %v", result)
				}
				return
			}

			// Context error should be wrapped in LayeredError
			layeredErr, ok := IsLayeredError(result)
			if !ok {
				t.Errorf("ContextError() should return LayeredError for context errors, got %T", result)
				return
			}

			if layeredErr.Layer != tt.expectedLayer {
				t.Errorf("ContextError() layer = %v, want %v", layeredErr.Layer, tt.expectedLayer)
			}

			if layeredErr.Operation != tt.expectedOp {
				t.Errorf("ContextError() operation = %v, want %v", layeredErr.Operation, tt.expectedOp)
			}

			if layeredErr.Message != tt.expectedMsg {
				t.Errorf("ContextError() message = %v, want %v", layeredErr.Message, tt.expectedMsg)
			}

			if layeredErr.Cause != tt.cause {
				t.Errorf("ContextError() cause = %v, want %v", layeredErr.Cause, tt.cause)
			}
		})
	}
}

// TestAsLayeredError tests safe type assertion helper
func TestAsLayeredError(t *testing.T) {
	layeredErr := APIError("test_op", "test message", nil)
	regularErr := fmt.Errorf("regular error")

	// Test with LayeredError
	result := AsLayeredError(layeredErr)
	if result == nil {
		t.Error("AsLayeredError should return LayeredError for valid LayeredError")
		return
	}
	if result.Layer != "api" {
		t.Errorf("Expected layer 'api', got %s", result.Layer)
	}

	// Test with regular error
	result = AsLayeredError(regularErr)
	if result != nil {
		t.Error("AsLayeredError should return nil for non-LayeredError")
	}

	// Test with nil error
	result = AsLayeredError(nil)
	if result != nil {
		t.Error("AsLayeredError should return nil for nil error")
	}
}

// TestWithContextSafe tests safe context addition helper
func TestWithContextSafe(t *testing.T) {
	layeredErr := APIError("test_op", "test message", nil)
	regularErr := fmt.Errorf("regular error")

	// Test with LayeredError
	result := WithContextSafe(layeredErr, "key", "value")
	if layeredResult := AsLayeredError(result); layeredResult != nil {
		if layeredResult.Context["key"] != "value" {
			t.Errorf("Expected context key=value, got %s", layeredResult.Context["key"])
		}
	} else {
		t.Error("WithContextSafe should preserve LayeredError")
	}

	// Test with regular error (should return unchanged)
	result = WithContextSafe(regularErr, "key", "value")
	if result != regularErr {
		t.Error("WithContextSafe should return original error for non-LayeredError")
	}

	// Test with nil error
	result = WithContextSafe(nil, "key", "value")
	if result != nil {
		t.Error("WithContextSafe should return nil for nil error")
	}
}

// TestWrapWithOperation tests error wrapping helper
func TestWrapWithOperation(t *testing.T) {
	originalErr := fmt.Errorf("original error")

	// Test with regular error
	result := WrapWithOperation(originalErr, "test_layer", "test_op", "test message")
	layeredResult := AsLayeredError(result)
	if layeredResult == nil {
		t.Error("WrapWithOperation should create LayeredError")
		return
	}
	if layeredResult.Layer != "test_layer" {
		t.Errorf("Expected layer 'test_layer', got %s", layeredResult.Layer)
	}
	if layeredResult.Operation != "test_op" {
		t.Errorf("Expected operation 'test_op', got %s", layeredResult.Operation)
	}
	if layeredResult.Cause != originalErr {
		t.Errorf("Expected cause to be original error")
	}

	// Test with LayeredError (should wrap it)
	layeredErr := APIError("inner_op", "inner message", originalErr)
	result = WrapWithOperation(layeredErr, "outer_layer", "outer_op", "outer message")
	outerResult := AsLayeredError(result)
	if outerResult == nil {
		t.Error("WrapWithOperation should create outer LayeredError")
		return
	}
	if outerResult.Layer != "outer_layer" {
		t.Errorf("Expected outer layer 'outer_layer', got %s", outerResult.Layer)
	}
	if outerResult.Cause != layeredErr {
		t.Error("Expected cause to be the inner LayeredError")
	}

	// Test with nil error
	result = WrapWithOperation(nil, "test_layer", "test_op", "test message")
	if result != nil {
		t.Error("WrapWithOperation should return nil for nil error")
	}
}
