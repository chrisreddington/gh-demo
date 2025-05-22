package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/hydrate"
	"github.com/spf13/cobra"
)

var (
	owner      string
	repo       string
	configPath string
)

func init() {
	// Create the hydrate command
	hydrateCmd := &cobra.Command{
		Use:   "hydrate <owner> <repo> [config-path]",
		Short: "Hydrate a GitHub repository with issues, discussions, PRs, and labels",
		Long: `Hydrate a GitHub repository with predefined resources based on a configuration file.
This command creates issues, discussions, pull requests, and labels in the specified repository.`,
		Example: "  gh-demo hydrate octocat my-repo ./hydrate.json",
		Args:    cobra.MinimumNArgs(2),
		RunE:    runHydrateCmd,
	}

	// Add the root command with hydrate as a subcommand
	rootCmd.AddCommand(hydrateCmd)
}

// runHydrateCmd executes the hydrate subcommand
func runHydrateCmd(cmd *cobra.Command, args []string) error {
	// Parse arguments
	owner = args[0]
	repo = args[1]

	// Optional config path
	configPath = ""
	if len(args) > 2 {
		configPath = args[2]
	}

	// Create GitHub API client
	client, err := githubapi.NewGitHubClient()
	if err != nil {
		return fmt.Errorf("error initializing GitHub client: %v", err)
	}

	// Create hydrator
	hydrator := hydrate.NewHydrator(client)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run the hydration process
	fmt.Printf("Hydrating repository %s/%s...\n", owner, repo)
	if err := hydrator.Hydrate(ctx, configPath, owner, repo); err != nil {
		return fmt.Errorf("error during hydration: %v", err)
	}

	fmt.Println("Repository hydration completed successfully!")
	return nil
}
