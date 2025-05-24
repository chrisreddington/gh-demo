// Package hydrate provides preservation logic for selective cleanup operations.
package hydrate

import (
	"regexp"

	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// ShouldPreserveIssue checks if an issue should be preserved based on the configuration.
func ShouldPreserveIssue(pc *config.PreserveConfig, issue types.Issue) bool {
	// Check by ID
	for _, id := range pc.Issues.PreserveByID {
		if issue.NodeID == id {
			return true
		}
	}

	// Check by title (exact match and regex)
	for _, pattern := range pc.Issues.PreserveByTitle {
		if isMatchOrRegex(issue.Title, pattern) {
			return true
		}
	}

	// Check by labels
	for _, preserveLabel := range pc.Issues.PreserveByLabel {
		for _, issueLabel := range issue.Labels {
			if issueLabel == preserveLabel {
				return true
			}
		}
	}

	return false
}

// ShouldPreserveDiscussion checks if a discussion should be preserved based on the configuration.
func ShouldPreserveDiscussion(pc *config.PreserveConfig, discussion types.Discussion) bool {
	// Check by ID
	for _, id := range pc.Discussions.PreserveByID {
		if discussion.NodeID == id {
			return true
		}
	}

	// Check by title (exact match and regex)
	for _, pattern := range pc.Discussions.PreserveByTitle {
		if isMatchOrRegex(discussion.Title, pattern) {
			return true
		}
	}

	// Check by category
	for _, category := range pc.Discussions.PreserveByCategory {
		if discussion.Category == category {
			return true
		}
	}

	return false
}

// ShouldPreservePR checks if a pull request should be preserved based on the configuration.
func ShouldPreservePR(pc *config.PreserveConfig, pr types.PullRequest) bool {
	// Check by ID
	for _, id := range pc.PullRequests.PreserveByID {
		if pr.NodeID == id {
			return true
		}
	}

	// Check by title (exact match and regex)
	for _, pattern := range pc.PullRequests.PreserveByTitle {
		if isMatchOrRegex(pr.Title, pattern) {
			return true
		}
	}

	// Check by labels
	for _, preserveLabel := range pc.PullRequests.PreserveByLabel {
		for _, prLabel := range pr.Labels {
			if prLabel == preserveLabel {
				return true
			}
		}
	}

	return false
}

// ShouldPreserveLabel checks if a label should be preserved based on the configuration.
func ShouldPreserveLabel(pc *config.PreserveConfig, labelName string) bool {
	for _, name := range pc.Labels.PreserveByName {
		if labelName == name {
			return true
		}
	}
	return false
}

// isMatchOrRegex checks if a string matches either exactly or as a regex pattern.
// It first tries exact match, then regex if the pattern starts with '^' or contains regex special chars.
func isMatchOrRegex(value, pattern string) bool {
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