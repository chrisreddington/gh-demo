package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestHydrateCmd_MissingArguments(t *testing.T) {
	// Create a command for testing
	cmd := &cobra.Command{
		Use:  "hydrate",
		RunE: runHydrateCmd,
		Args: cobra.MinimumNArgs(2),
	}

	// Set args to insufficient number
	cmd.SetArgs([]string{})

	// Create an output buffer to capture error message
	buf := new(bytes.Buffer)
	cmd.SetErr(buf)

	// Run the command and capture the error
	err := cmd.Execute()

	// The command should return an error with insufficient args
	if err == nil {
		t.Error("Expected command to return an error with insufficient arguments, but it did not")
	}

	// Check if the error message contains information about required arguments
	if buf.Len() == 0 {
		t.Error("Expected an error message but got none")
	}
}
