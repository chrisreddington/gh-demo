package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/hydrate"
	"github.com/spf13/cobra"
)

// NewHydrateCmd returns the Cobra command for repository hydration
func NewHydrateCmd() *cobra.Command {
	var owner, repo string
	var issues, discussions, prs bool

	cmd := &cobra.Command{
		Use:   "hydrate",
		Short: "Hydrate a repository with demo issues, discussions, and pull requests",
		Run: func(cmd *cobra.Command, args []string) {
			if owner == "" || repo == "" {
				fmt.Fprintln(os.Stderr, "--owner and --repo are required")
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

			client := githubapi.NewGHClient(owner, repo)
			err = hydrate.HydrateWithLabels(client, issuesPath, discussionsPath, prsPath, issues, discussions, prs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Hydration failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Repository hydrated successfully!")
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (required)")
	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name (required)")
	cmd.Flags().BoolVar(&issues, "issues", true, "Include issues")
	cmd.Flags().BoolVar(&discussions, "discussions", true, "Include discussions")
	cmd.Flags().BoolVar(&prs, "prs", true, "Include pull requests")

	return cmd
}
