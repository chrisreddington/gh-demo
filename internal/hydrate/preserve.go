// Package hydrate provides preservation logic for selective cleanup operations.
package hydrate

import (
	"context"
	"regexp"

	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// checkPreservationByID checks if an item should be preserved based on its ID
func checkPreservationByID(nodeID string, preserveByID []string) bool {
	for _, id := range preserveByID {
		if nodeID == id {
			return true
		}
	}
	return false
}

// checkPreservationByTitle checks if an item should be preserved based on its title patterns
func checkPreservationByTitle(ctx context.Context, title string, preserveByTitle []string) bool {
	for _, pattern := range preserveByTitle {
		if isMatchOrRegex(ctx, title, pattern) {
			return true
		}
	}
	return false
}

// checkPreservationByLabels checks if an item should be preserved based on its labels
func checkPreservationByLabels(itemLabels []string, preserveByLabel []string) bool {
	for _, preserveLabel := range preserveByLabel {
		for _, itemLabel := range itemLabels {
			if itemLabel == preserveLabel {
				return true
			}
		}
	}
	return false
}

// checkPreservationByCategory checks if a discussion should be preserved based on its category
func checkPreservationByCategory(category string, preserveByCategory []string) bool {
	for _, preserveCategory := range preserveByCategory {
		if category == preserveCategory {
			return true
		}
	}
	return false
}

// checkPreservationByName checks if an item should be preserved based on its name
func checkPreservationByName(name string, preserveByName []string) bool {
	for _, preserveName := range preserveByName {
		if name == preserveName {
			return true
		}
	}
	return false
}

// ShouldPreserveIssue checks if an issue should be preserved based on the configuration.
func ShouldPreserveIssue(ctx context.Context, preserveConfig *config.PreserveConfig, issue types.Issue) bool {
	return checkPreservationByID(issue.NodeID, preserveConfig.Issues.PreserveByID) ||
		checkPreservationByTitle(ctx, issue.Title, preserveConfig.Issues.PreserveByTitle) ||
		checkPreservationByLabels(issue.Labels, preserveConfig.Issues.PreserveByLabel)
}

// ShouldPreserveDiscussion checks if a discussion should be preserved based on the configuration.
func ShouldPreserveDiscussion(ctx context.Context, preserveConfig *config.PreserveConfig, discussion types.Discussion) bool {
	return checkPreservationByID(discussion.NodeID, preserveConfig.Discussions.PreserveByID) ||
		checkPreservationByTitle(ctx, discussion.Title, preserveConfig.Discussions.PreserveByTitle) ||
		checkPreservationByCategory(discussion.Category, preserveConfig.Discussions.PreserveByCategory)
}

// ShouldPreservePR checks if a pull request should be preserved based on the configuration.
func ShouldPreservePR(ctx context.Context, preserveConfig *config.PreserveConfig, pullRequest types.PullRequest) bool {
	return checkPreservationByID(pullRequest.NodeID, preserveConfig.PullRequests.PreserveByID) ||
		checkPreservationByTitle(ctx, pullRequest.Title, preserveConfig.PullRequests.PreserveByTitle) ||
		checkPreservationByLabels(pullRequest.Labels, preserveConfig.PullRequests.PreserveByLabel)
}

// ShouldPreserveLabel checks if a label should be preserved based on the configuration.
func ShouldPreserveLabel(ctx context.Context, preserveConfig *config.PreserveConfig, labelName string) bool {
	return checkPreservationByName(labelName, preserveConfig.Labels.PreserveByName)
}

// isMatchOrRegex checks if a string matches either exactly or as a regex pattern.
// It first tries exact match, then regex if the pattern starts with '^' or contains regex special chars.
func isMatchOrRegex(ctx context.Context, value, pattern string) bool {
	// Try exact match first
	if value == pattern {
		return true
	}

	// Try regex match if pattern looks like regex
	if len(pattern) > 0 && (pattern[0] == '^' || regexp.QuoteMeta(pattern) != pattern) {
		if regex, err := regexp.Compile(pattern); err == nil {
			return regex.MatchString(value)
		}
	}

	return false
}
