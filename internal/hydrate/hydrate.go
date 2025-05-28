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
	"strings"

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

// CleanupOptions defines the options for cleanup operations
type CleanupOptions struct {
	CleanIssues      bool
	CleanDiscussions bool
	CleanPRs         bool
	CleanLabels      bool
	DryRun           bool
	PreserveConfig   *config.PreserveConfig
}

// CleanupSummary holds statistics for cleanup operations
type CleanupSummary struct {
	IssuesDeleted        int
	IssuesPreserved      int
	DiscussionsDeleted   int
	DiscussionsPreserved int
	PRsDeleted           int
	PRsPreserved         int
	LabelsDeleted        int
	LabelsPreserved      int
	Errors               []string
}

// handleListError creates and returns error for list operation failures
func handleListError(err error, operation, itemType string) []string {
	wrappedErr := errors.WrapWithOperation(err, "cleanup", operation, fmt.Sprintf("failed to list %s", itemType))
	return []string{wrappedErr.Error()}
}

// handleDeleteError creates, collects and logs error for delete operation failures
func handleDeleteError(err error, collector *errors.ErrorCollector, logger common.Logger, itemType, title, nodeID string) {
	wrappedErr := errors.WrapWithOperation(err, "cleanup", fmt.Sprintf("delete_%s", itemType), fmt.Sprintf("failed to delete %s", itemType))
	wrappedErr = errors.WithContextSafe(wrappedErr, "title", title)
	wrappedErr = errors.WithContextSafe(wrappedErr, "node_id", nodeID)
	collector.Add(wrappedErr)
	logger.Info("Failed to delete %s '%s': %v", itemType, title, err)
}

// handleLabelDeleteError creates, collects and logs error for label delete operation failures
func handleLabelDeleteError(err error, collector *errors.ErrorCollector, logger common.Logger, labelName string) {
	wrappedErr := errors.WrapWithOperation(err, "cleanup", "delete_label", "failed to delete label")
	wrappedErr = errors.WithContextSafe(wrappedErr, "label_name", labelName)
	collector.Add(wrappedErr)
	logger.Info("Failed to delete label '%s': %v", labelName, err)
}

// convertErrorsToStringSlice converts collected errors to string slice for backward compatibility
func convertErrorsToStringSlice(collector *errors.ErrorCollector) []string {
	if result := collector.Result(); result != nil {
		if partialErr, ok := result.(*errors.PartialFailureError); ok {
			return partialErr.Errors
		}
		return []string{result.Error()}
	}
	return nil
}

// HydrateWithLabels loads content, collects all labels, and ensures labels exist before hydration.
// It supports both explicit label definitions from labels.json and auto-generated labels with defaults.
// It continues processing even if individual items fail, collecting all errors and reporting them at the end.
func HydrateWithLabels(ctx context.Context, client githubapi.GitHubClient, cfg *config.Configuration, includeIssues, includeDiscussions, includePullRequests bool, logger common.Logger, dryRun bool) error {

	if dryRun {
		logger.Info("Starting hydration operations (dry-run: true)")
	}

	issues, discussions, pullRequests, err := HydrateFromConfiguration(ctx, cfg, includeIssues, includeDiscussions, includePullRequests)
	if err != nil {
		return errors.ConfigError("load_config_files", "failed to load configuration files", err)
	}

	// Try to read explicit label definitions from labels.json
	explicitLabels, err := ReadLabelsJSON(ctx, cfg.LabelsPath)
	if err != nil {
		err = errors.WrapWithOperation(err, "config", "read_labels_config", "failed to read labels configuration")
		return errors.WithContextSafe(err, "path", cfg.LabelsPath)
	}

	// Collect label names referenced in content
	referencedLabelNames := CollectLabels(ctx, issues, discussions, pullRequests)

	// Prepare the final list of labels to ensure exist
	labelsToEnsure := prepareLabelsToEnsure(ctx, explicitLabels, referencedLabelNames)

	labelSummary := &SectionSummary{Name: "Labels", Total: len(labelsToEnsure)}

	if len(explicitLabels) > 0 {
		logger.Debug("Found %d explicit label definitions from %s", len(explicitLabels), cfg.LabelsPath)
	}
	logger.Debug("Found %d total labels to ensure exist", len(labelsToEnsure))

	if err := EnsureDefinedLabelsExist(ctx, client, labelsToEnsure, logger, labelSummary, dryRun); err != nil {
		return errors.APIError("ensure_labels", "failed to ensure labels exist", err)
	}

	// Report label summary
	logger.Info("Labels: %d total, %d successful, %d failed", labelSummary.Total, labelSummary.Success, labelSummary.Failures)

	// Create issues, discussions, and pull requests
	if err := createRepositoryContent(ctx, client, issues, discussions, pullRequests, includeIssues, includeDiscussions, includePullRequests, logger, dryRun); err != nil {
		return err
	}

	return nil
}

// createRepositoryContent orchestrates the creation of all content types.
// This function handles the creation of issues, discussions, and pull requests
// and collects any errors that occur during the process.
func createRepositoryContent(ctx context.Context, client githubapi.GitHubClient, issues []types.Issue, discussions []types.Discussion, pullRequests []types.PullRequest, includeIssues, includeDiscussions, includePullRequests bool, logger common.Logger, dryRun bool) error {
	var allErrors []string

	// Create issues, discussions, and pull requests
	if includeIssues {
		issueErrors, err := createIssues(ctx, client, issues, logger, dryRun)
		if err != nil {
			return err
		}
		if len(issueErrors) > 0 {
			allErrors = append(allErrors, issueErrors...)
		}
	}

	if includeDiscussions {
		discussionErrors, err := createDiscussions(ctx, client, discussions, logger, dryRun)
		if err != nil {
			return err
		}
		if len(discussionErrors) > 0 {
			allErrors = append(allErrors, discussionErrors...)
		}
	}

	if includePullRequests {
		prErrors, err := createPullRequests(ctx, client, pullRequests, logger, dryRun)
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
func prepareLabelsToEnsure(ctx context.Context, explicitLabels []types.Label, referencedLabelNames []string) []types.Label {
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

// createItems is a generic function for creating GitHub objects (issues, discussions, PRs).
// It eliminates code duplication between the specific creation functions.
func createItems[T any](
	ctx context.Context,
	client githubapi.GitHubClient,
	items []T,
	itemType string,
	createFunc func(context.Context, T) error,
	getTitleFunc func(T) string,
	logger common.Logger,
	dryRun bool,
) ([]string, error) {
	if len(items) == 0 {
		return nil, nil
	}

	var errors []string
	summary := &SectionSummary{Name: itemType, Total: len(items)}
	logger.Debug("Creating %d %s", len(items), strings.ToLower(itemType))

	for i, item := range items {
		// Check for cancellation before each item creation
		if err := ctx.Err(); err != nil {
			return errors, err
		}

		title := getTitleFunc(item)
		if dryRun {
			logger.Info("Would create %s: %s", strings.ToLower(itemType[:len(itemType)-1]), title)
			summary.Success++
		} else {
			if err := createFunc(ctx, item); err != nil {
				errorMsg := common.FormatCreationError(itemType[:len(itemType)-1], title, i, err)
				errors = append(errors, errorMsg)
				summary.Errors = append(summary.Errors, errorMsg)
				summary.Failures++
				logger.Debug("Failed to create %s '%s': %v", strings.ToLower(itemType[:len(itemType)-1]), title, err)
			} else {
				summary.Success++
				logger.Debug("Successfully created %s '%s'", strings.ToLower(itemType[:len(itemType)-1]), title)
			}
		}
	}
	logger.Info("%s: %d total, %d successful, %d failed", itemType, summary.Total, summary.Success, summary.Failures)
	return errors, nil
}

// createIssues creates all issues and collects any errors that occur.
// It returns a slice of error messages for any issues that failed to create.
func createIssues(ctx context.Context, client githubapi.GitHubClient, issues []types.Issue, logger common.Logger, dryRun bool) ([]string, error) {
	return createItems(
		ctx, client, issues, "Issues",
		client.CreateIssue,
		func(issue types.Issue) string { return issue.Title },
		logger, dryRun,
	)
}

// createDiscussions creates all discussions and collects any errors that occur.
// It returns a slice of error messages for any discussions that failed to create.
func createDiscussions(ctx context.Context, client githubapi.GitHubClient, discussions []types.Discussion, logger common.Logger, dryRun bool) ([]string, error) {
	return createItems(
		ctx, client, discussions, "Discussions",
		client.CreateDiscussion,
		func(discussion types.Discussion) string { return discussion.Title },
		logger, dryRun,
	)
}

// createPullRequests creates all pull requests and collects any errors that occur.
// It returns a slice of error messages for any pull requests that failed to create.
func createPullRequests(ctx context.Context, client githubapi.GitHubClient, pullRequests []types.PullRequest, logger common.Logger, dryRun bool) ([]string, error) {
	return createItems(
		ctx, client, pullRequests, "Pull Requests",
		client.CreatePR,
		func(pr types.PullRequest) string { return pr.Title },
		logger, dryRun,
	)
}

// EnsureDefinedLabelsExist creates any missing labels in the repository.
// It checks which labels already exist and only creates those that are missing.
// This function works with full Label objects that include color and description.
func EnsureDefinedLabelsExist(ctx context.Context, client githubapi.GitHubClient, labels []types.Label, logger common.Logger, summary *SectionSummary, dryRun bool) error {
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
			if dryRun {
				logger.Info("Would create label: %s (color: %s)", label.Name, label.Color)
				summary.Success++
			} else {
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

// CleanupBeforeHydration performs cleanup operations before hydration
func CleanupBeforeHydration(ctx context.Context, client githubapi.GitHubClient, options CleanupOptions, logger common.Logger) (*CleanupSummary, error) {
	summary := &CleanupSummary{
		Errors: make([]string, 0),
	}

	logger.Info("Starting cleanup operations (dry-run: %v)", options.DryRun)

	var allErrors []string

	// Clean issues
	if options.CleanIssues {
		issueErrors := cleanupIssues(ctx, client, options, summary, logger)
		if len(issueErrors) > 0 {
			allErrors = append(allErrors, issueErrors...)
		}
	}

	// Clean discussions
	if options.CleanDiscussions {
		discussionErrors := cleanupDiscussions(ctx, client, options, summary, logger)
		if len(discussionErrors) > 0 {
			allErrors = append(allErrors, discussionErrors...)
		}
	}

	// Clean pull requests
	if options.CleanPRs {
		prErrors := cleanupPRs(ctx, client, options, summary, logger)
		if len(prErrors) > 0 {
			allErrors = append(allErrors, prErrors...)
		}
	}

	// Clean labels
	if options.CleanLabels {
		labelErrors := cleanupLabels(ctx, client, options, summary, logger)
		if len(labelErrors) > 0 {
			allErrors = append(allErrors, labelErrors...)
		}
	}

	summary.Errors = allErrors

	// Log summary
	logger.Info("Cleanup summary: Issues(%d deleted, %d preserved), Discussions(%d deleted, %d preserved), PRs(%d deleted, %d preserved), Labels(%d deleted, %d preserved)",
		summary.IssuesDeleted, summary.IssuesPreserved,
		summary.DiscussionsDeleted, summary.DiscussionsPreserved,
		summary.PRsDeleted, summary.PRsPreserved,
		summary.LabelsDeleted, summary.LabelsPreserved)

	if len(allErrors) > 0 {
		logger.Info("Cleanup completed with %d errors", len(allErrors))
		// Return partial failure error if there were errors
		return summary, errors.NewPartialFailureError(allErrors)
	}

	logger.Info("Cleanup completed successfully")
	return summary, nil
}

// cleanupItems is a generic function for cleaning up GitHub objects.
// It eliminates code duplication between the specific cleanup functions.
func cleanupItems[T any](
	ctx context.Context,
	client githubapi.GitHubClient,
	options CleanupOptions,
	summary *CleanupSummary,
	logger common.Logger,
	itemType string,
	listFunc func(context.Context) ([]T, error),
	preserveFunc func(context.Context, *config.PreserveConfig, T) bool,
	deleteFunc func(context.Context, string) error,
	getTitleFunc func(T) string,
	getNodeIDFunc func(T) string,
	updatePreservedCount func(*CleanupSummary),
	updateDeletedCount func(*CleanupSummary),
) []string {
	operationName := common.FormatOperationContext("cleanup", itemType)
	collector := errors.NewErrorCollector(operationName)

	items, err := listFunc(ctx)
	if err != nil {
		return handleListError(err, common.FormatOperationContext("list", itemType), strings.ToLower(itemType))
	}

	logger.Debug("Found %d %s to evaluate for cleanup", len(items), strings.ToLower(itemType))

	for _, item := range items {
		title := getTitleFunc(item)
		if options.PreserveConfig != nil && preserveFunc(ctx, options.PreserveConfig, item) {
			updatePreservedCount(summary)
			logger.Debug("Preserving %s: %s", strings.ToLower(itemType[:len(itemType)-1]), title)
			continue
		}

		if options.DryRun {
			logger.Info("Would delete %s: %s", strings.ToLower(itemType[:len(itemType)-1]), title)
		} else {
			logger.Debug("Deleting %s: %s", strings.ToLower(itemType[:len(itemType)-1]), title)
			nodeID := getNodeIDFunc(item)
			if err := deleteFunc(ctx, nodeID); err != nil {
				handleDeleteError(err, collector, logger, strings.ToLower(itemType[:len(itemType)-1]), title, nodeID)
				continue
			}
		}
		updateDeletedCount(summary)
	}

	return convertErrorsToStringSlice(collector)
}

// cleanupIssues handles cleanup of issues
func cleanupIssues(ctx context.Context, client githubapi.GitHubClient, options CleanupOptions, summary *CleanupSummary, logger common.Logger) []string {
	return cleanupItems(
		ctx, client, options, summary, logger, "Issues",
		client.ListIssues,
		ShouldPreserveIssue,
		client.DeleteIssue,
		func(issue types.Issue) string { return issue.Title },
		func(issue types.Issue) string { return issue.NodeID },
		func(s *CleanupSummary) { s.IssuesPreserved++ },
		func(s *CleanupSummary) { s.IssuesDeleted++ },
	)
}

// cleanupDiscussions handles cleanup of discussions
func cleanupDiscussions(ctx context.Context, client githubapi.GitHubClient, options CleanupOptions, summary *CleanupSummary, logger common.Logger) []string {
	return cleanupItems(
		ctx, client, options, summary, logger, "Discussions",
		client.ListDiscussions,
		ShouldPreserveDiscussion,
		client.DeleteDiscussion,
		func(discussion types.Discussion) string { return discussion.Title },
		func(discussion types.Discussion) string { return discussion.NodeID },
		func(s *CleanupSummary) { s.DiscussionsPreserved++ },
		func(s *CleanupSummary) { s.DiscussionsDeleted++ },
	)
}

// cleanupPRs handles cleanup of pull requests
func cleanupPRs(ctx context.Context, client githubapi.GitHubClient, options CleanupOptions, summary *CleanupSummary, logger common.Logger) []string {
	return cleanupItems(
		ctx, client, options, summary, logger, "Pull Requests",
		client.ListPRs,
		ShouldPreservePR,
		client.DeletePR,
		func(pr types.PullRequest) string { return pr.Title },
		func(pr types.PullRequest) string { return pr.NodeID },
		func(s *CleanupSummary) { s.PRsPreserved++ },
		func(s *CleanupSummary) { s.PRsDeleted++ },
	)
}

// cleanupLabels handles cleanup of labels
func cleanupLabels(ctx context.Context, client githubapi.GitHubClient, options CleanupOptions, summary *CleanupSummary, logger common.Logger) []string {
	collector := errors.NewErrorCollector("cleanup_labels")

	labelNames, err := client.ListLabels(ctx)
	if err != nil {
		return handleListError(err, "list_labels", "labels")
	}

	logger.Debug("Found %d labels to evaluate for cleanup", len(labelNames))

	for _, labelName := range labelNames {
		if options.PreserveConfig != nil && ShouldPreserveLabel(ctx, options.PreserveConfig, labelName) {
			summary.LabelsPreserved++
			logger.Debug("Preserving label: %s", labelName)
			continue
		}

		if options.DryRun {
			logger.Info("Would delete label: %s", labelName)
		} else {
			logger.Debug("Deleting label: %s", labelName)
			if err := client.DeleteLabel(ctx, labelName); err != nil {
				handleLabelDeleteError(err, collector, logger, labelName)
				continue
			}
		}
		summary.LabelsDeleted++
	}

	return convertErrorsToStringSlice(collector)
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
			err = errors.WrapWithOperation(err, "file", "read_issues", "failed to read issues file")
			return nil, nil, nil, errors.WithContextSafe(err, "path", issuesPath)
		}
		if err := json.Unmarshal(data, &issues); err != nil {
			err = errors.WrapWithOperation(err, "file", "parse_issues", "failed to parse issues file")
			return nil, nil, nil, errors.WithContextSafe(err, "path", issuesPath)
		}
	}

	if includeDiscussions {
		// Check for cancellation before reading discussions file
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}

		data, err := os.ReadFile(discussionsPath)
		if err != nil {
			err = errors.WrapWithOperation(err, "file", "read_discussions", "failed to read discussions file")
			return nil, nil, nil, errors.WithContextSafe(err, "path", discussionsPath)
		}
		if err := json.Unmarshal(data, &discussions); err != nil {
			err = errors.WrapWithOperation(err, "file", "parse_discussions", "failed to parse discussions file")
			return nil, nil, nil, errors.WithContextSafe(err, "path", discussionsPath)
		}
	}

	if includePullRequests {
		// Check for cancellation before reading pull requests file
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		data, err := os.ReadFile(pullRequestsPath)
		if err != nil {
			err = errors.WrapWithOperation(err, "file", "read_pull_requests", "failed to read pull requests file")
			return nil, nil, nil, errors.WithContextSafe(err, "path", pullRequestsPath)
		}
		if err := json.Unmarshal(data, &pullRequests); err != nil {
			err = errors.WrapWithOperation(err, "file", "parse_pull_requests", "failed to parse pull requests file")
			return nil, nil, nil, errors.WithContextSafe(err, "path", pullRequestsPath)
		}
	}

	return issues, discussions, pullRequests, nil
}

// CollectLabels returns a deduplicated list of all labels used in issues, discussions, and PRs.
// CollectLabels returns a deduplicated list of all labels used in issues, discussions, and pull requests.
func CollectLabels(ctx context.Context, issues []types.Issue, discussions []types.Discussion, pullRequests []types.PullRequest) []string {
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
		err = errors.WrapWithOperation(err, "file", "read_labels", "failed to read labels file")
		return nil, errors.WithContextSafe(err, "path", labelsPath)
	}

	var labels []types.Label
	if err := json.Unmarshal(content, &labels); err != nil {
		err = errors.WrapWithOperation(err, "file", "parse_labels", "invalid JSON in labels file")
		return nil, errors.WithContextSafe(err, "path", labelsPath)
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
