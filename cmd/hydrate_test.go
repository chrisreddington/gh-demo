package cmd

import (
	"os"
	"testing"
)

func TestHydrateCmd_MissingArguments(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test case with missing arguments
	os.Args = []string{"gh-demo", "hydrate"}

	// Save the original exit function
	origExit := osExit
	defer func() { osExit = origExit }()
	
	// Replace the exit function
	osExit = func(code int) {
		// Don't actually exit the test
		panic("exit")
	}

	// Run the function and recover from the panic
	defer func() {
		if r := recover(); r != nil {
			if r != "exit" {
				t.Errorf("Unexpected panic: %v", r)
			}
		}
	}()

	hydrateCmd()

	// The function should panic with "exit" before reaching here
	t.Error("Expected command to exit, but it did not")
}