package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-demo",
	Short: "A tool to automate GitHub repository hydration",
	Long: `A GitHub CLI extension that automates repository hydration tasks, 
such as creating issues, discussions, pull requests, and labels 
using the GitHub API.`,
}

// Mock os.Exit for testing
var osExit = os.Exit

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		osExit(1)
	}
}
