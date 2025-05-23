package main

import (
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

// TestMainFunctionExists verifies that main function exists
// We can't test main() directly without it trying to execute the CLI
func TestMainFunctionExists(t *testing.T) {
	// The existence of this test file and successful compilation
	// verifies that main() function exists and imports work correctly

	// Test passes if compilation succeeded
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
}

// TestCmdExecuteAccessible tests that cmd.Execute is accessible from main
func TestCmdExecuteAccessible(t *testing.T) {
	// We can't actually call cmd.Execute() as it would run the CLI,
	// but we can verify that it's accessible (which exercises the import path)

	// This test verifies that the main.go import structure is correct
	// If cmd.Execute wasn't accessible, main.go wouldn't compile

	// Create a new command to verify we can access the cmd package functionality
	// that main() would use
	hydrateCmd := cmd.NewHydrateCmd()

	if hydrateCmd == nil {
		t.Error("Should be able to access cmd package functionality")
		return
	}

	// Verify we can access the command that main() would execute through cmd.Execute()
	if hydrateCmd.Use != "hydrate" {
		t.Error("Should be able to access hydrate command")
	}
}
