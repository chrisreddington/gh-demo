// Package hydrate provides preservation logic for selective cleanup operations.
package hydrate

import (
	"context"
	"regexp"

	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// ShouldPreserveIssue checks if an issue should be preserved based on the configuration.
func ShouldPreserveIssue(ctx context.Context, preserveConfig *config.PreserveConfig, issue types.Issue) bool {
	// Check by ID
	for _, id := range preserveConfig.Issues.PreserveByID {
		if issue.NodeID == id {
			return true
		}
	}

	// Check by title (exact match and regex)
	for _, pattern := range preserveConfig.Issues.PreserveByTitle {
		if isMatchOrRegex(ctx, issue.Title, pattern) {
			return true
		}
	}

	// Check by labels
	for _, preserveLabel := range preserveConfig.Issues.PreserveByLabel {
		for _, issueLabel := range issue.Labels {
			if issueLabel == preserveLabel {
				return true
			}
		}
	}

	return false
}

// ShouldPreserveDiscussion checks if a discussion should be preserved based on the configuration.
func ShouldPreserveDiscussion(ctx context.Context, preserveConfig *config.PreserveConfig, discussion types.Discussion) bool {
	// Check by ID
	for _, id := range preserveConfig.Discussions.PreserveByID {
		if discussion.NodeID == id {
			return true
		}
	}

	// Check by title (exact match and regex)
	for _, pattern := range preserveConfig.Discussions.PreserveByTitle {
		if isMatchOrRegex(ctx, discussion.Title, pattern) {
			return true
		}
	}

	// Check by category
	for _, category := range preserveConfig.Discussions.PreserveByCategory {
		if discussion.Category == category {
			return true
		}
	}

	return false
}

// ShouldPreservePR checks if a pull request should be preserved based on the configuration.
func ShouldPreservePR(ctx context.Context, preserveConfig *config.PreserveConfig, pullRequest types.PullRequest) bool {
	// Check by ID
	for _, id := range preserveConfig.PullRequests.PreserveByID {
		if pullRequest.NodeID == id {
			return true
		}
	}

	// Check by title (exact match and regex)
	for _, pattern := range preserveConfig.PullRequests.PreserveByTitle {
		if isMatchOrRegex(ctx, pullRequest.Title, pattern) {
			return true
		}
	}

	// Check by labels
	for _, preserveLabel := range preserveConfig.PullRequests.PreserveByLabel {
		for _, prLabel := range pullRequest.Labels {
			if prLabel == preserveLabel {
				return true
			}
		}
	}

	return false
}

// ShouldPreserveLabel checks if a label should be preserved based on the configuration.
func ShouldPreserveLabel(ctx context.Context, preserveConfig *config.PreserveConfig, labelName string) bool {
	for _, name := range preserveConfig.Labels.PreserveByName {
		if labelName == name {
			return true
		}
	}
	return false
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
