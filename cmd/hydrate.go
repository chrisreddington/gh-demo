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
		Use:   "hydrate [owner] [repo] [config-path]",
		Short: "Hydrate a GitHub repository with issues, discussions, PRs, and labels",
		Long: `Hydrate a GitHub repository with predefined resources based on a configuration file.
This command creates issues, discussions, pull requests, and labels in the specified repository.

If owner and repo are not specified, the current repository will be used.
If config-path is not specified, it will look for hydrate.json in .github/demos/ then in the current directory.`,
		Example: "  gh-demo hydrate octocat my-repo ./hydrate.json\n  gh-demo hydrate # uses current repo and default config",
		Args:    cobra.MaximumNArgs(3),
		RunE:    runHydrateCmd,
	}

	// Add the root command with hydrate as a subcommand
	rootCmd.AddCommand(hydrateCmd)
}

// runHydrateCmd executes the hydrate subcommand
func runHydrateCmd(cmd *cobra.Command, args []string) error {
	// Create GitHub API client
	client, err := githubapi.NewGitHubClient()
	if err != nil {
		return fmt.Errorf("error initializing GitHub client: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Parse arguments
	owner = ""
	repo = ""
	configPath = ""

	// If arguments are provided, use them; otherwise use current repository
	switch len(args) {
	case 3:
		configPath = args[2]
		fallthrough
	case 2:
		owner = args[0]
		repo = args[1]
	case 1:
		configPath = args[0]
	}

	// If owner or repo not specified, get from current repository
	if owner == "" || repo == "" {
		currentOwner, currentRepo, err := client.GetCurrentRepository(ctx)
		if err != nil {
			return fmt.Errorf("error getting current repository: %v", err)
		}
		if owner == "" {
			owner = currentOwner
		}
		if repo == "" {
			repo = currentRepo
		}
	}

	// Create hydrator
	hydrator := hydrate.NewHydrator(client)

	// Run the hydration process
	fmt.Printf("Hydrating repository %s/%s...\n", owner, repo)
	if err := hydrator.Hydrate(ctx, configPath, owner, repo); err != nil {
		return fmt.Errorf("error during hydration: %v", err)
	}

	fmt.Println("Repository hydration completed successfully!")
	return nil
}
