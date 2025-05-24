package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-demo",
	Short: "GitHub Demo CLI Extension",
}

// Execute runs the root command and returns any error that occurs during execution.
// This is the main entry point for the CLI application and should be called from main().
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(NewHydrateCmd())
}
