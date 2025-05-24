package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// TestExecute tests Execute function behavior with different argument scenarios
func TestExecute(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args and restore afterwards
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			err := Execute()

			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}

// TestExecuteStructure tests the structure rather than execution
func TestExecuteStructure(t *testing.T) {
	// Test that Execute function exists and can be called
	// We test the structure rather than actual execution to avoid os.Exit calls

	if rootCmd.Use != "gh-demo" {
		t.Error("Root command should be properly configured")
	}

	// Check that hydrate command is added
	var hydrateFound bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "hydrate" {
			hydrateFound = true
			break
		}
	}

	if !hydrateFound {
		t.Error("Hydrate command should be added to root")
	}
}

func TestRootCmd(t *testing.T) {
	if rootCmd.Use != "gh-demo" {
		t.Errorf("Expected Use to be 'gh-demo', got %s", rootCmd.Use)
	}

	if rootCmd.Short != "GitHub Demo CLI Extension" {
		t.Errorf("Expected Short to be 'GitHub Demo CLI Extension', got %s", rootCmd.Short)
	}
}

func TestInit(t *testing.T) {
	// Test that init function properly adds the hydrate command
	var found bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "hydrate" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected hydrate command to be added to root command")
	}
}

func TestRootCmdHasHydrateSubcommand(t *testing.T) {
	hydrateCmd := &cobra.Command{
		Use: "hydrate",
	}

	// Check if hydrate command is in the list of commands
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == hydrateCmd.Use {
			found = true
			break
		}
	}

	if !found {
		t.Error("Hydrate command should be added to root command")
	}
}

func TestRootCmdExecution(t *testing.T) {
	// Test root command structure and basic functionality
	cmd := &cobra.Command{
		Use:   "gh-demo",
		Short: "GitHub Demo CLI Extension",
	}

	// Add hydrate command
	cmd.AddCommand(NewHydrateCmd())

	// Test that we can set args and get subcommands
	cmd.SetArgs([]string{"hydrate", "--help"})

	// Capture stderr to avoid test output pollution
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := cmd.Execute()

	// Restore stderr
	if closeErr := w.Close(); closeErr != nil {
		t.Errorf("Failed to close stderr writer: %v", closeErr)
	}
	os.Stderr = originalStderr

	// Read captured output
	var buf bytes.Buffer
	if _, copyErr := io.Copy(&buf, r); copyErr != nil {
		t.Errorf("Failed to copy captured output: %v", copyErr)
	}

	// Help command should not return an error
	if err != nil {
		t.Errorf("Help command should not return error: %v", err)
	}
}
