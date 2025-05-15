package cmd

import (
	"fmt"
	"os"
)

// List of available subcommands
var subcommands = []string{"hydrate"}

func Execute() {
	if len(os.Args) < 2 {
		fmt.Println("Available subcommands:")
		for _, cmd := range subcommands {
			fmt.Printf("  %s\n", cmd)
		}
		os.Exit(0)
	}

	switch os.Args[1] {
	case "hydrate":
		hydrateCmd()
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[1])
		fmt.Println("Available subcommands:")
		for _, cmd := range subcommands {
			fmt.Printf("  %s\n", cmd)
		}
		os.Exit(1)
	}
}
