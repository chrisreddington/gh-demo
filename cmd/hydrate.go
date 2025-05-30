package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
func resolveRepositoryInfo(ctx context.Context, owner, repo string) (*repositoryInfo, error) {
	// Check if context is cancelled before operations
	if err := ctx.Err(); err != nil {
		return nil, err
	}

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
		return nil, errors.ValidationError("validate_repository", "--owner and --repo are required (or run inside a GitHub repo)")
	}

	return &repositoryInfo{
		Owner: resolvedOwner,
		Repo:  resolvedRepo,
	}, nil
}

// createGitHubClient creates and configures a GitHub API client.
func createGitHubClient(ctx context.Context, repoInfo *repositoryInfo, logger common.Logger) (githubapi.GitHubClient, error) {
	client, err := githubapi.NewGHClient(ctx, repoInfo.Owner, repoInfo.Repo)
	if err != nil {
		return nil, errors.APIError("create_client", "failed to create GitHub client", err)
	}

	// Set logger for debug output
	client.SetLogger(logger)

	return client, nil
}

// handleHydrationResult processes the result of the hydration operation.
// It handles both complete failures and partial failures with appropriate user feedback.
func handleHydrationResult(ctx context.Context, err error, logger common.Logger) error {
	if err != nil {
		// Check if this is a partial failure using proper error type detection
		if errors.IsPartialFailure(err) {
			logger.Info("Repository hydration completed with some failures")
			// Check if context is cancelled before I/O operation
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			fmt.Fprintf(os.Stderr, "Hydration completed with some failures:\n%v\n", err)
			return nil // Partial failures are considered success for CLI purposes
		} else {
			// Complete failure
			return errors.APIError("hydrate_repository", "hydration failed", err)
		}
	} else {
		logger.Info("Repository hydrated successfully")
	}
	return nil
}

// CleanupFlags holds all cleanup-related command line flags
type CleanupFlags struct {
	Clean            bool
	CleanIssues      bool
	CleanDiscussions bool
	CleanPRs         bool
	CleanLabels      bool
	DryRun           bool
	PreserveConfig   string
}

// ProjectFlags holds all project-related command line flags
type ProjectFlags struct {
	CreateProject        bool
	ProjectConfig        string
	FailOnProjectError   bool
}

// executeHydrate contains the core hydration logic separated from CLI concerns
// executeHydrate performs the hydration operation with the given parameters.
// It validates required parameters, resolves git context if needed, and orchestrates the hydration process.
func executeHydrate(ctx context.Context, owner, repo, configPath string, issues, discussions, pullRequests, debug bool, cleanupFlags CleanupFlags, projectFlags ProjectFlags) error {
	// Create logger for operations
	logger := common.NewLogger(debug) // Use debug flag for logger

	// Resolve repository information
	repoInfo, err := resolveRepositoryInfo(ctx, owner, repo)
	if err != nil {
		return err
	}

	// Find project root
	root, err := hydrate.FindProjectRoot(ctx)
	if err != nil {
		return errors.FileError("find_project_root", "could not find project root", err)
	}

	// Create configuration object
	cfg := config.NewConfigurationWithRoot(ctx, root, configPath)

	// Create and configure GitHub client
	client, err := createGitHubClient(ctx, repoInfo, logger)
	if err != nil {
		return err
	}

	// Perform cleanup if requested
	if shouldPerformCleanup(ctx, cleanupFlags) {
		err := performCleanup(ctx, client, cleanupFlags, cfg, logger)
		if err != nil {
			// Log cleanup error but continue with hydration unless it's a critical failure
			logger.Info("Cleanup encountered errors but continuing with hydration: %v", err)
		}
	}

	// Perform hydration with project support
	if projectFlags.CreateProject {
		err = hydrate.HydrateWithProject(ctx, client, cfg, issues, discussions, pullRequests, logger, cleanupFlags.DryRun, true, projectFlags.ProjectConfig)
	} else {
		err = hydrate.HydrateWithLabels(ctx, client, cfg, issues, discussions, pullRequests, logger, cleanupFlags.DryRun)
	}

	// Handle project-specific errors
	if err != nil && projectFlags.CreateProject {
		// Check if this is a project-related error
		if errors.IsLayer(err, "project") && projectFlags.FailOnProjectError {
			return err
		} else if errors.IsLayer(err, "project") {
			// Log project error but continue if FailOnProjectError is false
			logger.Info("Project creation failed but continuing with standard hydration: %v", err)
			// Retry with standard hydration
			err = hydrate.HydrateWithLabels(ctx, client, cfg, issues, discussions, pullRequests, logger, cleanupFlags.DryRun)
		}
	}

	// Handle the result
	return handleHydrationResult(ctx, err, logger)
}

// shouldPerformCleanup determines if any cleanup operations should be performed
func shouldPerformCleanup(ctx context.Context, flags CleanupFlags) bool {
	return flags.Clean || flags.CleanIssues || flags.CleanDiscussions || flags.CleanPRs || flags.CleanLabels
}

// performCleanup executes cleanup operations based on flags
func performCleanup(ctx context.Context, client githubapi.GitHubClient, flags CleanupFlags, cfg *config.Configuration, logger common.Logger) error {
	// Load preserve configuration
	preserveConfigPath := flags.PreserveConfig
	if preserveConfigPath == "" {
		preserveConfigPath = cfg.PreservePath
	}

	preserveConfig, err := config.LoadPreserveConfig(ctx, preserveConfigPath)
	if err != nil {
		return errors.FileError("load_preserve_config", "failed to load preserve configuration", err)
	}

	// Create cleanup options
	cleanupOptions := hydrate.CleanupOptions{
		CleanIssues:      flags.Clean || flags.CleanIssues,
		CleanDiscussions: flags.Clean || flags.CleanDiscussions,
		CleanPRs:         flags.Clean || flags.CleanPRs,
		CleanLabels:      flags.Clean || flags.CleanLabels,
		DryRun:           flags.DryRun,
		PreserveConfig:   preserveConfig,
	}

	// Perform cleanup
	summary, err := hydrate.CleanupBeforeHydration(ctx, client, cleanupOptions, logger)
	if summary != nil {
		// Log cleanup summary
		logger.Info("Cleanup completed: %d issues cleaned, %d discussions cleaned, %d PRs cleaned, %d labels cleaned",
			summary.IssuesDeleted, summary.DiscussionsDeleted, summary.PRsDeleted, summary.LabelsDeleted)
	}

	return err
}

// NewHydrateCmd returns the Cobra command for repository hydration
func NewHydrateCmd() *cobra.Command {
	var owner, repo, configPath string
	var issues, discussions, pullRequests bool
	var debug bool

	// Cleanup flags
	var cleanupFlags CleanupFlags
	
	// Project flags
	var projectFlags ProjectFlags

	cmd := &cobra.Command{
		Use:   "hydrate",
		Short: "Hydrate a repository with demo issues, discussions, and pull requests",
		Long: `Hydrate a repository with demo issues, discussions, and pull requests.

Cleanup flags allow you to clean existing objects before hydrating:
  --clean: Clean all object types (issues, discussions, PRs, labels)
  --clean-issues: Clean only issues
  --clean-discussions: Clean only discussions
  --clean-prs: Clean only pull requests
  --clean-labels: Clean only labels
  --dry-run: Preview what would be created and deleted without actually performing operations
  --preserve-config: Path to preserve configuration file (default: .github/demos/preserve.json)

Project flags allow you to create and organize content in a GitHub Project:
  --create-project: Create a ProjectV2 and associate all created content with it
  --project-config: Path to project configuration file (default: .github/demos/project-config.json)
  --fail-on-project-error: Fail entire operation if project creation fails (default: continue with standard hydration)`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create context with cancellation for Ctrl+C
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			err := executeHydrate(ctx, owner, repo, configPath, issues, discussions, pullRequests, debug, cleanupFlags, projectFlags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Setup command line flags
	setupHydrateCmdFlags(cmd, &owner, &repo, &configPath, &issues, &discussions, &pullRequests, &debug, &cleanupFlags, &projectFlags)

	return cmd
}

// setupHydrateCmdFlags configures all command line flags for the hydrate command.
// This separates flag configuration from command creation for better maintainability.
func setupHydrateCmdFlags(cmd *cobra.Command, owner, repo, configPath *string, issues, discussions, pullRequests, debug *bool, cleanupFlags *CleanupFlags, projectFlags *ProjectFlags) {
	// Repository flags
	cmd.Flags().StringVar(owner, "owner", "", "GitHub repository owner (required)")
	cmd.Flags().StringVar(repo, "repo", "", "GitHub repository name (required)")
	cmd.Flags().StringVar(configPath, "config-path", config.DefaultConfigPath, "Path to configuration files relative to project root")

	// Content type flags
	cmd.Flags().BoolVar(issues, "issues", true, "Include issues")
	cmd.Flags().BoolVar(discussions, "discussions", true, "Include discussions")
	cmd.Flags().BoolVar(pullRequests, "prs", true, "Include pull requests")

	// Debug flag
	cmd.Flags().BoolVar(debug, "debug", false, "Enable debug mode for detailed logging")

	// Cleanup flags
	cmd.Flags().BoolVar(&cleanupFlags.Clean, "clean", false, "Clean all existing objects before hydrating")
	cmd.Flags().BoolVar(&cleanupFlags.CleanIssues, "clean-issues", false, "Clean existing issues before hydrating")
	cmd.Flags().BoolVar(&cleanupFlags.CleanDiscussions, "clean-discussions", false, "Clean existing discussions before hydrating")
	cmd.Flags().BoolVar(&cleanupFlags.CleanPRs, "clean-prs", false, "Clean existing pull requests before hydrating")
	cmd.Flags().BoolVar(&cleanupFlags.CleanLabels, "clean-labels", false, "Clean existing labels before hydrating")
	cmd.Flags().BoolVar(&cleanupFlags.DryRun, "dry-run", false, "Preview what would be created and deleted without actually performing operations")
	cmd.Flags().StringVar(&cleanupFlags.PreserveConfig, "preserve-config", "", "Path to preserve configuration file (default: .github/demos/preserve.json)")

	// Project flags
	cmd.Flags().BoolVar(&projectFlags.CreateProject, "create-project", false, "Create a ProjectV2 and associate all created content with it")
	cmd.Flags().StringVar(&projectFlags.ProjectConfig, "project-config", "", "Path to project configuration file (default: .github/demos/project-config.json)")
	cmd.Flags().BoolVar(&projectFlags.FailOnProjectError, "fail-on-project-error", false, "Fail entire operation if project creation fails (default: continue with standard hydration)")
}
