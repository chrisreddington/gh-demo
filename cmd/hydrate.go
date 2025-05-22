package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/hydrate"
)

// hydrateCmd executes the hydrate subcommand
func hydrateCmd() {
	// Check for required arguments: owner and repo
	if len(os.Args) < 4 {
		fmt.Println("Usage: gh-demo hydrate <owner> <repo> [config-path]")
		fmt.Println("Example: gh-demo hydrate octocat my-repo ./hydrate.json")
		os.Exit(1)
	}

	// Parse arguments
	owner := os.Args[2]
	repo := os.Args[3]

	// Optional config path
	configPath := ""
	if len(os.Args) > 4 {
		configPath = os.Args[4]
	}

	// Create GitHub API client
	client, err := githubapi.NewGitHubClient()
	if err != nil {
		fmt.Printf("Error initializing GitHub client: %v\n", err)
		os.Exit(1)
	}

	// Create hydrator
	hydrator := hydrate.NewHydrator(client)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run the hydration process
	fmt.Printf("Hydrating repository %s/%s...\n", owner, repo)
	if err := hydrator.Hydrate(ctx, configPath, owner, repo); err != nil {
		fmt.Printf("Error during hydration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Repository hydration completed successfully!")
}
