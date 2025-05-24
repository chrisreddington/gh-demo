package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/errors"
	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/hydrate"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

// repositoryInfo holds the resolved owner and repository information
type repositoryInfo struct {
	Owner string
	Repo  string
}

// resolveRepositoryInfo validates and resolves the repository owner and name.
// It tries to get missing values from the current git context if available.
func resolveRepositoryInfo(owner, repo string) (*repositoryInfo, error) {
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
		return nil, fmt.Errorf("--owner and --repo are required (or run inside a GitHub repo)")
	}
	
	return &repositoryInfo{
		Owner: resolvedOwner,
		Repo:  resolvedRepo,
	}, nil
}

// configurationPaths holds the paths to all configuration files
type configurationPaths struct {
	Issues       string
	Discussions  string
	PullRequests string
}

// buildConfigurationPaths constructs the full paths to configuration files.
func buildConfigurationPaths(configPath string) (*configurationPaths, error) {
	root, err := hydrate.FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("could not find project root: %v", err)
	}
	
	return &configurationPaths{
		Issues:       filepath.Join(root, configPath, "issues.json"),
		Discussions:  filepath.Join(root, configPath, "discussions.json"),
		PullRequests: filepath.Join(root, configPath, "prs.json"),
	}, nil
}

// createGitHubClient creates and configures a GitHub API client.
func createGitHubClient(repoInfo *repositoryInfo, debug bool) (githubapi.GitHubClient, error) {
	client, err := githubapi.NewGHClient(repoInfo.Owner, repoInfo.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}
	
	// Set logger for debug mode
	if debug {
		client.SetLogger(common.NewLogger(debug))
	}
	
	return client, nil
}

// handleHydrationResult processes the result of the hydration operation.
// It handles both complete failures and partial failures with appropriate user feedback.
func handleHydrationResult(err error) error {
	if err != nil {
		// Check if this is a partial failure using proper error type detection
		if errors.IsPartialFailure(err) {
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

// executeHydrate contains the core hydration logic separated from CLI concerns
// executeHydrate performs the hydration operation with the given parameters.
// It validates required parameters, resolves git context if needed, and orchestrates the hydration process.
func executeHydrate(ctx context.Context, owner, repo, configPath string, issues, discussions, pullRequests, debug bool) error {
	// Resolve repository information
	repoInfo, err := resolveRepositoryInfo(owner, repo)
	if err != nil {
		return err
	}
	
	// Build configuration file paths
	paths, err := buildConfigurationPaths(configPath)
	if err != nil {
		return err
	}
	
	// Create and configure GitHub client
	client, err := createGitHubClient(repoInfo, debug)
	if err != nil {
		return err
	}
	
	// Perform hydration
	err = hydrate.HydrateWithLabels(client, paths.Issues, paths.Discussions, paths.PullRequests, issues, discussions, pullRequests, debug)
	
	// Handle the result
	return handleHydrationResult(err)
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
			ctx := cmd.Context()
			err := executeHydrate(ctx, owner, repo, configPath, issues, discussions, pullRequests, debug)
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
