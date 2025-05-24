package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/hydrate"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

// executeHydrate contains the core hydration logic separated from CLI concerns
func executeHydrate(owner, repo string, issues, discussions, prs, debug bool) error {
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
	issuesPath := filepath.Join(root, ".github", "demos", "issues.json")
	discussionsPath := filepath.Join(root, ".github", "demos", "discussions.json")
	prsPath := filepath.Join(root, ".github", "demos", "prs.json")

	client, err := githubapi.NewGHClient(resolvedOwner, resolvedRepo)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}
	// Set logger for debug mode
	if debug {
		client.SetLogger(common.NewLogger(debug))
	}
	err = hydrate.HydrateWithLabels(client, issuesPath, discussionsPath, prsPath, issues, discussions, prs, debug)
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
	var owner, repo string
	var issues, discussions, prs bool
	var debug bool

	cmd := &cobra.Command{
		Use:   "hydrate",
		Short: "Hydrate a repository with demo issues, discussions, and pull requests",
		Run: func(cmd *cobra.Command, args []string) {
			err := executeHydrate(owner, repo, issues, discussions, prs, debug)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (required)")
	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name (required)")
	cmd.Flags().BoolVar(&issues, "issues", true, "Include issues")
	cmd.Flags().BoolVar(&discussions, "discussions", true, "Include discussions")
	cmd.Flags().BoolVar(&prs, "prs", true, "Include pull requests")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode for detailed logging")

	return cmd
}
