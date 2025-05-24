package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/hydrate"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

// executeHydrate contains the core hydration logic separated from CLI concerns
// executeHydrate performs the hydration operation with the given parameters.
// It validates required parameters, resolves git context if needed, and orchestrates the hydration process.
func executeHydrate(owner, repo, configPath string, issues, discussions, pullRequests, debug bool) error {
	resolvedOwner := strings.TrimSpace(owner)
	resolvedRepo := strings.TrimSpace(repo)
	if resolvedOwner == "" || resolvedRepo == "" {
		// Try to get from current git context
		repoCtx, err := repository.Current()
		if err == nil {
			if resolvedOwner == "" {
				resolvedOwner = repoCtx.Owner
			}
			if resolvedRepo == "" {
				resolvedRepo = repoCtx.Name
			}
		}
	}
	if resolvedOwner == "" || resolvedRepo == "" {
		return fmt.Errorf("--owner and --repo are required (or run inside a GitHub repo)")
	}

	root, err := hydrate.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %v", err)
	}
	issuesPath := filepath.Join(root, configPath, "issues.json")
	discussionsPath := filepath.Join(root, configPath, "discussions.json")
	pullRequestsPath := filepath.Join(root, configPath, "prs.json")

	client, err := githubapi.NewGHClient(resolvedOwner, resolvedRepo)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}
	// Set logger for debug mode
	if debug {
		client.SetLogger(common.NewLogger(debug))
	}
	err = hydrate.HydrateWithLabels(client, issuesPath, discussionsPath, pullRequestsPath, issues, discussions, pullRequests, debug)
	if err != nil {
		// Check if this is a partial failure (some items succeeded, some failed)
		if strings.Contains(err.Error(), "some items failed to create:") {
			fmt.Fprintf(os.Stderr, "Hydration completed with some failures:\n%v\n", err)
			fmt.Println("Repository hydration completed with some failures. Check the errors above for details.")
			return nil // Partial failures are considered success for CLI purposes
		} else {
			// Complete failure
			return fmt.Errorf("hydration failed: %v", err)
		}
	} else {
		fmt.Println("Repository hydrated successfully!")
	}
	return nil
}

// NewHydrateCmd returns the Cobra command for repository hydration
func NewHydrateCmd() *cobra.Command {
	var owner, repo, configPath string
	var issues, discussions, pullRequests bool
	var debug bool

	cmd := &cobra.Command{
		Use:   "hydrate",
		Short: "Hydrate a repository with demo issues, discussions, and pull requests",
		Run: func(cmd *cobra.Command, args []string) {
			err := executeHydrate(owner, repo, configPath, issues, discussions, pullRequests, debug)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (required)")
	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name (required)")
	cmd.Flags().StringVar(&configPath, "config-path", config.DefaultConfigPath, "Path to configuration files relative to project root")
	cmd.Flags().BoolVar(&issues, "issues", true, "Include issues")
	cmd.Flags().BoolVar(&discussions, "discussions", true, "Include discussions")
	cmd.Flags().BoolVar(&pullRequests, "prs", true, "Include pull requests")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode for detailed logging")

	return cmd
}
