package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/hydrate"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

// NewHydrateCmd returns the Cobra command for repository hydration
func NewHydrateCmd() *cobra.Command {
	var owner, repo string
	var issues, discussions, prs bool
	var debug bool

	cmd := &cobra.Command{
		Use:   "hydrate",
		Short: "Hydrate a repository with demo issues, discussions, and pull requests",
		Run: func(cmd *cobra.Command, args []string) {
			resolvedOwner := owner
			resolvedRepo := repo
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
				fmt.Fprintln(os.Stderr, "--owner and --repo are required (or run inside a GitHub repo)")
				os.Exit(1)
			}

			root, err := hydrate.FindProjectRoot()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not find project root: %v\n", err)
				os.Exit(1)
			}
			issuesPath := filepath.Join(root, ".github", "demos", "issues.json")
			discussionsPath := filepath.Join(root, ".github", "demos", "discussions.json")
			prsPath := filepath.Join(root, ".github", "demos", "prs.json")

			client := githubapi.NewGHClient(resolvedOwner, resolvedRepo)
			err = hydrate.HydrateWithLabels(client, issuesPath, discussionsPath, prsPath, issues, discussions, prs, debug)
			if err != nil {
				// Check if this is a partial failure (some items succeeded, some failed)
				if strings.Contains(err.Error(), "some items failed to create:") {
					fmt.Fprintf(os.Stderr, "Hydration completed with some failures:\n%v\n", err)
					fmt.Println("Repository hydration completed with some failures. Check the errors above for details.")
				} else {
					// Complete failure
					fmt.Fprintf(os.Stderr, "Hydration failed: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Println("Repository hydrated successfully!")
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
