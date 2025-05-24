package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/chrisreddington/gh-demo/cmd"
)

// TestMainPackageStructure tests that the main package is properly structured
func TestMainPackageStructure(t *testing.T) {
	// Test that we can create a hydrate command (this exercises the import)
	hydrateCmd := cmd.NewHydrateCmd()
	if hydrateCmd == nil {
		t.Error("Expected to be able to create hydrate command")
		return
	}

	if hydrateCmd.Use != "hydrate" {
		t.Errorf("Expected hydrate command to have Use 'hydrate', got %s", hydrateCmd.Use)
	}
}

// TestMainImports tests that all required imports are working
func TestMainImports(t *testing.T) {
	// Test that cmd.Execute function exists and is callable
	// We can't call it directly as it would execute the CLI, but we can verify
	// that the function exists by checking the cmd package

	// This test passes if the imports compile successfully
	// The fact that we can reference cmd package means imports are working

	// Create a command to verify the package is properly imported
	hydrateCmd := cmd.NewHydrateCmd()
	if hydrateCmd.Use != "hydrate" {
		t.Error("Failed to import and use cmd package correctly")
	}
}

// TestRunFunction tests the run() function with different argument scenarios
func TestRunFunction(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "help command success",
			args:        []string{"gh-demo", "--help"},
			expectError: false,
			description: "Help command should not return an error",
		},
		{
			name:        "invalid command error",
			args:        []string{"gh-demo", "invalid-command"},
			expectError: true,
			description: "Invalid command should return an error",
		},
		{
			name:        "nonexistent command error",
			args:        []string{"gh-demo", "nonexistent-command"},
			expectError: true,
			description: "Nonexistent command should return error with unknown command message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args and restore afterwards
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			err := run()

			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}

			// For the nonexistent command case, verify error message
			if tt.name == "nonexistent command error" && err != nil {
				if !strings.Contains(err.Error(), "unknown command") {
					t.Errorf("Expected 'unknown command' error, got: %v", err)
				}
			}
		})
	}
}

// TestMainFunctionDirect tests main() function scenarios with different arguments
func TestMainFunctionDirect(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldPanic bool
		description string
	}{
		{
			name:        "main with help success",
			args:        []string{"gh-demo", "--help"},
			shouldPanic: false,
			description: "Main with help should not call os.Exit()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args and restore afterwards
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			// Capture any panic from os.Exit() and recover
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("%s: main() should not call os.Exit(), but got panic: %v", tt.description, r)
					}
				}
			}()

			// Call main() directly
			main()

			// If we get here, main() completed successfully without calling os.Exit()
		})
	}
}

// TestMainFunctionExists verifies that main function exists
func TestMainFunctionExists(t *testing.T) {
	// Test that main function exists by verifying we can build the program
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// The existence of this test and successful compilation verifies main() exists
	// This is a structural test to ensure the main function is present
}

// TestMainFunctionIntegration tests main() indirectly through subprocess
func TestMainFunctionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test main() function by running the program as a subprocess with --help
	cmd := exec.Command("go", "run", "main.go", "--help")
	// Use current working directory instead of hard-coded GitHub Actions path

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("main() execution with --help failed: %v\nOutput: %s", err, output)
		return
	}

	// Verify the help output contains expected content
	outputStr := string(output)
	if !strings.Contains(outputStr, "GitHub Demo CLI Extension") {
		t.Error("main() should show GitHub Demo CLI Extension in help output")
	}
}

// TestMainFunctionWithError tests main() error handling through subprocess
func TestMainFunctionWithErrorSubprocess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test main() function by running the program as a subprocess with invalid command
	cmd := exec.Command("go", "run", "main.go", "invalid-command")
	// Use current working directory instead of hard-coded GitHub Actions path

	output, err := cmd.CombinedOutput()
	// This should fail (exit code 1) due to invalid command
	if err == nil {
		t.Error("main() with invalid command should exit with error")
		return
	}

	// Verify we get the expected error output
	outputStr := string(output)
	if !strings.Contains(outputStr, "Error:") || !strings.Contains(outputStr, "unknown command") {
		t.Errorf("main() should show error for invalid input, got: %s", outputStr)
	}
}

// TestCmdExecuteAccessible tests that cmd.Execute is accessible from main
func TestCmdExecuteAccessible(t *testing.T) {
	// We can verify that cmd.Execute returns the expected type
	// by calling it with help arguments

	// Save original args and restore afterwards
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set args to show help
	os.Args = []string{"gh-demo", "--help"}

	// Call cmd.Execute() directly (which is what run() does)
	err := cmd.Execute()

	// Help should not return an error
	if err != nil {
		t.Errorf("cmd.Execute() with --help should not return error, got: %v", err)
	}
}
