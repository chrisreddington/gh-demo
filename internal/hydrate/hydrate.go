// Package hydrate provides functionality for hydrating GitHub repositories with demo content.
// It handles the creation of issues, discussions, and pull requests based on JSON configuration files,
// ensuring that all required labels exist before creating content.
package hydrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/common"
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
// It continues processing even if individual items fail, collecting all errors and reporting them at the end.
func HydrateWithLabels(client githubapi.GitHubClient, issuesPath, discussionsPath, pullRequestsPath string, includeIssues, includeDiscussions, includePullRequests, debug bool) error {
	logger := common.NewLogger(debug)

	issues, discussions, pullRequests, err := HydrateFromFiles(issuesPath, discussionsPath, pullRequestsPath, includeIssues, includeDiscussions, includePullRequests)
	if err != nil {
		return err
	}

	// Collect and ensure all labels exist before creating content
	labels := CollectLabels(issues, discussions, pullRequests)
	labelSummary := &SectionSummary{Name: "Labels", Total: len(labels)}

	logger.Debug("Found %d unique labels to ensure exist: %v", len(labels), labels)

	if err := EnsureLabelsExist(client, labels, logger, labelSummary); err != nil {
		return fmt.Errorf("failed to ensure labels exist: %w", err)
	}

	// Report label summary
	logger.Info("Labels: %d total, %d successful, %d failed", labelSummary.Total, labelSummary.Success, labelSummary.Failures)

	var allErrors []string

	// Create issues
	if includeIssues {
		issueSummary := &SectionSummary{Name: "Issues", Total: len(issues)}
		logger.Debug("Creating %d issues", len(issues))

		for i, issue := range issues {
			if err := client.CreateIssue(issue); err != nil {
				errorMsg := fmt.Sprintf("Issue %d (%s): %v", i+1, issue.Title, err)
				allErrors = append(allErrors, errorMsg)
				issueSummary.Errors = append(issueSummary.Errors, errorMsg)
				issueSummary.Failures++
				logger.Debug("Failed to create issue '%s': %v", issue.Title, err)
			} else {
				issueSummary.Success++
				logger.Debug("Successfully created issue '%s'", issue.Title)
			}
		}
		logger.Info("Issues: %d total, %d successful, %d failed", issueSummary.Total, issueSummary.Success, issueSummary.Failures)
	}

	// Create discussions
	if includeDiscussions {
		discussionSummary := &SectionSummary{Name: "Discussions", Total: len(discussions)}
		logger.Debug("Creating %d discussions", len(discussions))

		for i, discussion := range discussions {
			if err := client.CreateDiscussion(discussion); err != nil {
				errorMsg := fmt.Sprintf("Discussion %d (%s): %v", i+1, discussion.Title, err)
				allErrors = append(allErrors, errorMsg)
				discussionSummary.Errors = append(discussionSummary.Errors, errorMsg)
				discussionSummary.Failures++
				logger.Debug("Failed to create discussion '%s': %v", discussion.Title, err)
			} else {
				discussionSummary.Success++
				logger.Debug("Successfully created discussion '%s'", discussion.Title)
			}
		}
		logger.Info("Discussions: %d total, %d successful, %d failed", discussionSummary.Total, discussionSummary.Success, discussionSummary.Failures)
	}

	// Create pull requests
	if includePullRequests {
		pullRequestSummary := &SectionSummary{Name: "Pull Requests", Total: len(pullRequests)}
		logger.Debug("Creating %d pull requests", len(pullRequests))

		for i, pullRequest := range pullRequests {
			if err := client.CreatePR(pullRequest); err != nil {
				errorMsg := fmt.Sprintf("Pull Request %d (%s): %v", i+1, pullRequest.Title, err)
				allErrors = append(allErrors, errorMsg)
				pullRequestSummary.Errors = append(pullRequestSummary.Errors, errorMsg)
				pullRequestSummary.Failures++
				logger.Debug("Failed to create pull request '%s': %v", pullRequest.Title, err)
			} else {
				pullRequestSummary.Success++
				logger.Debug("Successfully created pull request '%s'", pullRequest.Title)
			}
		}
		logger.Info("Pull Requests: %d total, %d successful, %d failed", pullRequestSummary.Total, pullRequestSummary.Success, pullRequestSummary.Failures)
	}

	// If any errors occurred, return them as a combined error but don't fail completely
	if len(allErrors) > 0 {
		return fmt.Errorf("some items failed to create:\n  - %s", strings.Join(allErrors, "\n  - "))
	}

	return nil
}

// EnsureLabelsExist checks if each label exists in the repo, and creates it if not.
func EnsureLabelsExist(client githubapi.GitHubClient, labels []string, logger common.Logger, summary *SectionSummary) error {
	if len(labels) == 0 {
		return nil
	}

	logger.Debug("Fetching existing labels from repository")
	existing, err := client.ListLabels()
	if err != nil {
		return err
	}

	existSet := make(map[string]struct{}, len(existing))
	for _, l := range existing {
		existSet[l] = struct{}{}
	}

	logger.Debug("Found %d existing labels in repository", len(existing))

	for _, l := range labels {
		if _, ok := existSet[l]; !ok {
			logger.Debug("Creating missing label '%s'", l)
			if err := client.CreateLabel(l); err != nil {
				errorMsg := fmt.Sprintf("Label '%s': %v", l, err)
				summary.Errors = append(summary.Errors, errorMsg)
				summary.Failures++
				logger.Debug("Failed to create label '%s': %v", l, err)
			} else {
				summary.Success++
				logger.Debug("Successfully created label '%s'", l)
			}
		} else {
			summary.Success++
			logger.Debug("Label '%s' already exists", l)
		}
	}

	return nil
}

// HydrateFromFiles loads issues, discussions, and pull requests from their respective JSON files.
// It only loads files for content types that are included (enabled by the respective boolean flags).
func HydrateFromFiles(issuesPath, discussionsPath, pullRequestsPath string, includeIssues, includeDiscussions, includePullRequests bool) ([]types.Issue, []types.Discussion, []types.PullRequest, error) {
	var issues []types.Issue
	var discussions []types.Discussion
	var pullRequests []types.PullRequest

	if includeIssues {
		data, err := os.ReadFile(issuesPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &issues); err != nil {
			return nil, nil, nil, err
		}
	}

	if includeDiscussions {
		data, err := os.ReadFile(discussionsPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &discussions); err != nil {
			return nil, nil, nil, err
		}
	}

	if includePullRequests {
		data, err := os.ReadFile(pullRequestsPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &pullRequests); err != nil {
			return nil, nil, nil, err
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

// FindProjectRoot traverses up from the current file to find the directory containing go.mod
func FindProjectRoot() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
