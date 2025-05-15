package cmd

import (
	   "os"
	   "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	   Use:   "gh-demo",
	   Short: "GitHub Demo CLI Extension",
}

func Execute() {
	   if err := rootCmd.Execute(); err != nil {
			   os.Exit(1)
	   }
}

func init() {
	   rootCmd.AddCommand(NewHydrateCmd())
}
