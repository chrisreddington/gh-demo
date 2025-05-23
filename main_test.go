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

// TestRunFunction tests the run() function
func TestRunFunction(t *testing.T) {
	// Test that run() function can be called and returns an error
	// when Execute() would fail (no arguments provided)
	
	// Save original args and restore afterwards
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Set args to show help (which should not return an error)
	os.Args = []string{"gh-demo", "--help"}
	
	err := run()
	// Help command should not return an error
	if err != nil {
		t.Errorf("run() with --help should not return error, got: %v", err)
	}
}

// TestRunFunctionWithInvalidArgs tests run() with invalid arguments
func TestRunFunctionWithInvalidArgs(t *testing.T) {
	// Save original args and restore afterwards
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Set invalid args that should cause an error
	os.Args = []string{"gh-demo", "invalid-command"}
	
	err := run()
	// Invalid command should return an error
	if err == nil {
		t.Error("run() with invalid command should return an error")
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

// TestMainFunctionDirect tests main() function directly with success case
func TestMainFunctionDirect(t *testing.T) {
	// This test actually calls main() directly for the success case where it doesn't call os.Exit()
	// We use the --help flag which should not cause os.Exit(1)
	
	// Save original args and restore afterwards
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Set args to show help (should exit with code 0, which doesn't call os.Exit in our implementation)
	os.Args = []string{"gh-demo", "--help"}
	
	// Capture any panic from os.Exit() and recover
	defer func() {
		if r := recover(); r != nil {
			// If os.Exit was called, we'll get here
			// Help should not cause os.Exit, so this would be unexpected
			t.Errorf("main() with --help should not call os.Exit(), but got panic: %v", r)
		}
	}()
	
	// Call main() directly - this will execute the help command
	main()
	
	// If we get here, main() completed successfully without calling os.Exit()
}

// TestMainFunctionIntegration tests main() indirectly through subprocess
func TestMainFunctionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Test main() function by running the program as a subprocess with --help
	cmd := exec.Command("go", "run", "main.go", "--help")
	cmd.Dir = "/home/runner/work/gh-demo/gh-demo"
	
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
	cmd.Dir = "/home/runner/work/gh-demo/gh-demo"
	
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

// TestMainFunctionErrorPath tests main() error handling by mocking
func TestMainFunctionErrorPath(t *testing.T) {
	// We can't easily test main() directly due to os.Exit(), but we can test
	// the error path by testing the run() function that main() calls
	
	// Save original args and restore afterwards
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Test with an invalid subcommand to trigger error in run()
	os.Args = []string{"gh-demo", "nonexistent-command"}
	
	err := run()
	if err == nil {
		t.Error("run() should return error for invalid command")
	}
	
	// Verify error message
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected 'unknown command' error, got: %v", err)
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
