// Package hydrate provides functionality for hydrating GitHub repositories with demo content.
// It handles the creation of issues, discussions, and pull requests based on JSON configuration files,
// ensuring that all required labels exist before creating content.
package hydrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/errors"
	"github.com/chrisreddington/gh-demo/internal/githubapi"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// SectionSummary holds statistics for a hydration section (labels, issues, discussions, pull requests).
// It tracks the total number of items processed, successful operations, failures, and detailed error messages.
type SectionSummary struct {
	Name     string   // Name of the section (e.g., "Issues", "Labels")
	Total    int      // Total number of items to process
	Success  int      // Number of successful operations
	Failures int      // Number of failed operations
	Errors   []string // Detailed error messages for failed operations
}

// HydrateWithLabels loads content, collects all labels, and ensures labels exist before hydration.
// It supports both explicit label definitions from labels.json and auto-generated labels with defaults.
// It continues processing even if individual items fail, collecting all errors and reporting them at the end.
func HydrateWithLabels(ctx context.Context, client githubapi.GitHubClient, cfg *config.Configuration, includeIssues, includeDiscussions, includePullRequests, debug bool) error {
	logger := common.NewLogger(debug)

	issues, discussions, pullRequests, err := HydrateFromConfiguration(ctx, cfg, includeIssues, includeDiscussions, includePullRequests)
	if err != nil {
		return errors.ConfigError("load_config_files", "failed to load configuration files", err)
	}

	// Try to read explicit label definitions from labels.json
	explicitLabels, err := ReadLabelsJSON(ctx, cfg.LabelsPath)
	if err != nil {
		layeredErr := errors.NewLayeredError("config", "read_labels_config", "failed to read labels configuration", err)
		return layeredErr.WithContext("path", cfg.LabelsPath)
	}

	// Collect label names referenced in content
	referencedLabelNames := CollectLabels(issues, discussions, pullRequests)

	// Prepare the final list of labels to ensure exist
	labelsToEnsure := prepareLabelsToEnsure(explicitLabels, referencedLabelNames)

	labelSummary := &SectionSummary{Name: "Labels", Total: len(labelsToEnsure)}

	if len(explicitLabels) > 0 {
		logger.Debug("Found %d explicit label definitions from %s", len(explicitLabels), cfg.LabelsPath)
	}
	logger.Debug("Found %d total labels to ensure exist", len(labelsToEnsure))

	if err := EnsureDefinedLabelsExist(ctx, client, labelsToEnsure, logger, labelSummary); err != nil {
		return errors.APIError("ensure_labels", "failed to ensure labels exist", err)
	}

	// Report label summary
	logger.Info("Labels: %d total, %d successful, %d failed", labelSummary.Total, labelSummary.Success, labelSummary.Failures)

	var allErrors []string

	// Create issues, discussions, and pull requests
	if includeIssues {
		issueErrors, err := createIssues(ctx, client, issues, logger)
		if err != nil {
			return err
		}
		if len(issueErrors) > 0 {
			allErrors = append(allErrors, issueErrors...)
		}
	}

	if includeDiscussions {
		discussionErrors, err := createDiscussions(ctx, client, discussions, logger)
		if err != nil {
			return err
		}
		if len(discussionErrors) > 0 {
			allErrors = append(allErrors, discussionErrors...)
		}
	}

	if includePullRequests {
		prErrors, err := createPullRequests(ctx, client, pullRequests, logger)
		if err != nil {
			return err
		}
		if len(prErrors) > 0 {
			allErrors = append(allErrors, prErrors...)
		}
	}

	// If any errors occurred, return them as a combined error but don't fail completely
	if len(allErrors) > 0 {
		return errors.NewPartialFailureError(allErrors)
	}

	return nil
}

// prepareLabelsToEnsure builds the final list of labels that need to be ensured to exist.
// It combines explicit labels from labels.json with auto-generated labels for any referenced labels.
func prepareLabelsToEnsure(explicitLabels []types.Label, referencedLabelNames []string) []types.Label {
	// Create a map of explicit labels by name for quick lookup
	explicitLabelMap := make(map[string]types.Label)
	for _, label := range explicitLabels {
		explicitLabelMap[label.Name] = label
	}

	// Build final list of labels to ensure exist
	var labelsToEnsure []types.Label

	// Add all explicit labels from labels.json
	labelsToEnsure = append(labelsToEnsure, explicitLabels...)

	// Add any referenced labels that aren't explicitly defined (with defaults)
	for _, labelName := range referencedLabelNames {
		if _, exists := explicitLabelMap[labelName]; !exists {
			// Create a default label for any referenced label not explicitly defined
			defaultLabel := types.Label{
				Name:        labelName,
				Description: "Label created by gh-demo hydration tool",
				Color:       config.DefaultLabelColor, // Light gray default color
			}
			labelsToEnsure = append(labelsToEnsure, defaultLabel)
		}
	}

	return labelsToEnsure
}

// createIssues creates all issues and collects any errors that occur.
// It returns a slice of error messages for any issues that failed to create.
func createIssues(ctx context.Context, client githubapi.GitHubClient, issues []types.Issue, logger common.Logger) ([]string, error) {
	if len(issues) == 0 {
		return nil, nil
	}

	var errors []string
	issueSummary := &SectionSummary{Name: "Issues", Total: len(issues)}
	logger.Debug("Creating %d issues", len(issues))

	for i, issue := range issues {
		// Check for cancellation before each issue creation
		if err := ctx.Err(); err != nil {
			return errors, err
		}

		if err := client.CreateIssue(ctx, issue); err != nil {
			errorMsg := fmt.Sprintf("Issue %d (%s): %v", i+1, issue.Title, err)
			errors = append(errors, errorMsg)
			issueSummary.Errors = append(issueSummary.Errors, errorMsg)
			issueSummary.Failures++
			logger.Debug("Failed to create issue '%s': %v", issue.Title, err)
		} else {
			issueSummary.Success++
			logger.Debug("Successfully created issue '%s'", issue.Title)
		}
	}
	logger.Info("Issues: %d total, %d successful, %d failed", issueSummary.Total, issueSummary.Success, issueSummary.Failures)
	return errors, nil
}

// createDiscussions creates all discussions and collects any errors that occur.
// It returns a slice of error messages for any discussions that failed to create.
func createDiscussions(ctx context.Context, client githubapi.GitHubClient, discussions []types.Discussion, logger common.Logger) ([]string, error) {
	if len(discussions) == 0 {
		return nil, nil
	}

	var errors []string
	discussionSummary := &SectionSummary{Name: "Discussions", Total: len(discussions)}
	logger.Debug("Creating %d discussions", len(discussions))

	for i, discussion := range discussions {
		// Check for cancellation before each discussion creation
		if err := ctx.Err(); err != nil {
			return errors, err
		}

		if err := client.CreateDiscussion(ctx, discussion); err != nil {
			errorMsg := fmt.Sprintf("Discussion %d (%s): %v", i+1, discussion.Title, err)
			errors = append(errors, errorMsg)
			discussionSummary.Errors = append(discussionSummary.Errors, errorMsg)
			discussionSummary.Failures++
			logger.Debug("Failed to create discussion '%s': %v", discussion.Title, err)
		} else {
			discussionSummary.Success++
			logger.Debug("Successfully created discussion '%s'", discussion.Title)
		}
	}
	logger.Info("Discussions: %d total, %d successful, %d failed", discussionSummary.Total, discussionSummary.Success, discussionSummary.Failures)
	return errors, nil
}

// createPullRequests creates all pull requests and collects any errors that occur.
// It returns a slice of error messages for any pull requests that failed to create.
func createPullRequests(ctx context.Context, client githubapi.GitHubClient, pullRequests []types.PullRequest, logger common.Logger) ([]string, error) {
	if len(pullRequests) == 0 {
		return nil, nil
	}

	var errors []string
	pullRequestSummary := &SectionSummary{Name: "Pull Requests", Total: len(pullRequests)}
	logger.Debug("Creating %d pull requests", len(pullRequests))

	for i, pullRequest := range pullRequests {
		// Check for cancellation before each pull request creation
		if err := ctx.Err(); err != nil {
			return errors, err
		}

		if err := client.CreatePR(ctx, pullRequest); err != nil {
			errorMsg := fmt.Sprintf("Pull Request %d (%s): %v", i+1, pullRequest.Title, err)
			errors = append(errors, errorMsg)
			pullRequestSummary.Errors = append(pullRequestSummary.Errors, errorMsg)
			pullRequestSummary.Failures++
			logger.Debug("Failed to create pull request '%s': %v", pullRequest.Title, err)
		} else {
			pullRequestSummary.Success++
			logger.Debug("Successfully created pull request '%s'", pullRequest.Title)
		}
	}
	logger.Info("Pull Requests: %d total, %d successful, %d failed", pullRequestSummary.Total, pullRequestSummary.Success, pullRequestSummary.Failures)
	return errors, nil
}

// EnsureDefinedLabelsExist creates any missing labels in the repository.
// It checks which labels already exist and only creates those that are missing.
// This function works with full Label objects that include color and description.
func EnsureDefinedLabelsExist(ctx context.Context, client githubapi.GitHubClient, labels []types.Label, logger common.Logger, summary *SectionSummary) error {
	if len(labels) == 0 {
		return nil
	}

	logger.Debug("Fetching existing labels from repository")
	existing, err := client.ListLabels(ctx)
	if err != nil {
		return err
	}

	existSet := make(map[string]struct{}, len(existing))
	for _, l := range existing {
		existSet[l] = struct{}{}
	}

	logger.Debug("Found %d existing labels in repository", len(existing))

	for _, label := range labels {
		// Check for cancellation before each label creation
		if err := ctx.Err(); err != nil {
			return err
		}

		if _, ok := existSet[label.Name]; !ok {
			logger.Debug("Creating missing label '%s' (color: %s)", label.Name, label.Color)

			if err := client.CreateLabel(ctx, label); err != nil {
				errorMsg := fmt.Sprintf("Label '%s': %v", label.Name, err)
				summary.Errors = append(summary.Errors, errorMsg)
				summary.Failures++
				logger.Debug("Failed to create label '%s': %v", label.Name, err)
			} else {
				summary.Success++
				logger.Debug("Successfully created label '%s' with color '%s'", label.Name, label.Color)
			}
		} else {
			summary.Success++
			logger.Debug("Label '%s' already exists", label.Name)
		}
	}

	return nil
}

// HydrateFromConfiguration loads issues, discussions, and pull requests from their respective JSON files
// using a Configuration object. It only loads files for content types that are included.
func HydrateFromConfiguration(ctx context.Context, cfg *config.Configuration, includeIssues, includeDiscussions, includePullRequests bool) ([]types.Issue, []types.Discussion, []types.PullRequest, error) {
	return HydrateFromFiles(ctx, cfg.IssuesPath, cfg.DiscussionsPath, cfg.PullRequestsPath, includeIssues, includeDiscussions, includePullRequests)
}

// HydrateWithLabelsFromPaths is a backward compatibility wrapper for tests.
// New code should use HydrateWithLabels with Configuration instead.
func HydrateWithLabelsFromPaths(ctx context.Context, client githubapi.GitHubClient, issuesPath, discussionsPath, pullRequestsPath string, includeIssues, includeDiscussions, includePullRequests, debug bool) error {
	// Create a configuration object from the individual paths
	// Extract the base directory from the issues path
	basePath := filepath.Dir(issuesPath)
	cfg := config.NewConfiguration(basePath)

	// Override the computed paths with the actual provided paths
	cfg.IssuesPath = issuesPath
	cfg.DiscussionsPath = discussionsPath
	cfg.PullRequestsPath = pullRequestsPath
	cfg.LabelsPath = filepath.Join(basePath, config.LabelsFilename)

	return HydrateWithLabels(ctx, client, cfg, includeIssues, includeDiscussions, includePullRequests, debug)
}

// HydrateFromFiles loads issues, discussions, and pull requests from their respective JSON files.
// It only loads files for content types that are included (enabled by the respective boolean flags).
func HydrateFromFiles(ctx context.Context, issuesPath, discussionsPath, pullRequestsPath string, includeIssues, includeDiscussions, includePullRequests bool) ([]types.Issue, []types.Discussion, []types.PullRequest, error) {
	var issues []types.Issue
	var discussions []types.Discussion
	var pullRequests []types.PullRequest

	if includeIssues {
		// Check for cancellation before reading issues file
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}

		data, err := os.ReadFile(issuesPath)
		if err != nil {
			layeredErr := errors.NewLayeredError("file", "read_issues", "failed to read issues file", err)
			return nil, nil, nil, layeredErr.WithContext("path", issuesPath)
		}
		if err := json.Unmarshal(data, &issues); err != nil {
			layeredErr := errors.NewLayeredError("file", "parse_issues", "failed to parse issues file", err)
			return nil, nil, nil, layeredErr.WithContext("path", issuesPath)
		}
	}

	if includeDiscussions {
		// Check for cancellation before reading discussions file
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}

		data, err := os.ReadFile(discussionsPath)
		if err != nil {
			layeredErr := errors.NewLayeredError("file", "read_discussions", "failed to read discussions file", err)
			return nil, nil, nil, layeredErr.WithContext("path", discussionsPath)
		}
		if err := json.Unmarshal(data, &discussions); err != nil {
			layeredErr := errors.NewLayeredError("file", "parse_discussions", "failed to parse discussions file", err)
			return nil, nil, nil, layeredErr.WithContext("path", discussionsPath)
		}
	}

	if includePullRequests {
		// Check for cancellation before reading pull requests file
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		data, err := os.ReadFile(pullRequestsPath)
		if err != nil {
			layeredErr := errors.NewLayeredError("file", "read_pull_requests", "failed to read pull requests file", err)
			return nil, nil, nil, layeredErr.WithContext("path", pullRequestsPath)
		}
		if err := json.Unmarshal(data, &pullRequests); err != nil {
			layeredErr := errors.NewLayeredError("file", "parse_pull_requests", "failed to parse pull requests file", err)
			return nil, nil, nil, layeredErr.WithContext("path", pullRequestsPath)
		}
	}

	return issues, discussions, pullRequests, nil
}

// CollectLabels returns a deduplicated list of all labels used in issues, discussions, and PRs.
// CollectLabels returns a deduplicated list of all labels used in issues, discussions, and pull requests.
func CollectLabels(issues []types.Issue, discussions []types.Discussion, pullRequests []types.PullRequest) []string {
	labelSet := make(map[string]struct{})
	for _, issue := range issues {
		for _, label := range issue.Labels {
			labelSet[label] = struct{}{}
		}
	}
	for _, discussion := range discussions {
		for _, label := range discussion.Labels {
			labelSet[label] = struct{}{}
		}
	}
	for _, pullRequest := range pullRequests {
		for _, label := range pullRequest.Labels {
			labelSet[label] = struct{}{}
		}
	}
	labels := make([]string, 0, len(labelSet))
	for label := range labelSet {
		labels = append(labels, label)
	}
	return labels
}

// ReadLabelsJSON reads label definitions from a JSON file.
// This allows users to define labels with specific colors and descriptions.
// Returns an empty slice if the file doesn't exist (not an error condition).
func ReadLabelsJSON(ctx context.Context, labelsPath string) ([]types.Label, error) {
	// Check for cancellation before starting file operations
	if err := ctx.Err(); err != nil {
		return nil, errors.ContextError("read_labels", err)
	}

	if _, err := os.Stat(labelsPath); os.IsNotExist(err) {
		// File doesn't exist, return empty slice (not an error)
		return []types.Label{}, nil
	}

	// Check for cancellation before reading file
	if err := ctx.Err(); err != nil {
		return nil, errors.ContextError("read_labels", err)
	}

	content, err := os.ReadFile(labelsPath)
	if err != nil {
		layeredErr := errors.NewLayeredError("file", "read_labels", "failed to read labels file", err)
		return nil, layeredErr.WithContext("path", labelsPath)
	}

	var labels []types.Label
	if err := json.Unmarshal(content, &labels); err != nil {
		layeredErr := errors.NewLayeredError("file", "parse_labels", "invalid JSON in labels file", err)
		return nil, layeredErr.WithContext("path", labelsPath)
	}

	return labels, nil
}

// FindProjectRoot traverses up from the current file to find the directory containing go.mod
func FindProjectRoot(ctx context.Context) (string, error) {
	// Check for cancellation before starting directory traversal
	if err := ctx.Err(); err != nil {
		return "", errors.ContextError("find_project_root", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		// Check for cancellation during each iteration
		if err := ctx.Err(); err != nil {
			return "", errors.ContextError("find_project_root", err)
		}

		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.FileError("find_project_root", "could not find project root (no go.mod found)", os.ErrNotExist)
		}
		dir = parent
	}
}
