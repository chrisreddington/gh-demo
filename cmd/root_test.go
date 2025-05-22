package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecute(t *testing.T) {
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

func TestExecuteWithHelp(t *testing.T) {
	// Test Execute with help flag to avoid actual execution
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Capture stdout 
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Set args to show help
	os.Args = []string{"gh-demo", "--help"}
	
	// Create a separate root command for testing to avoid modifying global state
	testRootCmd := &cobra.Command{
		Use:   "gh-demo",
		Short: "GitHub Demo CLI Extension",
	}
	testRootCmd.AddCommand(NewHydrateCmd())
	
	// Execute with help should not cause issues
	testRootCmd.SetArgs([]string{"--help"})
	err := testRootCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = old
	
	// Read captured output
	out, _ := io.ReadAll(r)
	output := string(out)
	
	// Help should not return an error
	if err != nil {
		t.Errorf("Help command should not return error: %v", err)
	}
	
	// Output should contain usage information
	if len(output) == 0 {
		t.Error("Help command should produce output")
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
	w.Close()
	os.Stderr = originalStderr
	
	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	// Help command should not return an error
	if err != nil {
		t.Errorf("Help command should not return error: %v", err)
	}
}