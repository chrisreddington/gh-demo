// Package hydrate provides preservation logic for selective cleanup operations.
package hydrate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"

	"github.com/chrisreddington/gh-demo/internal/errors"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// PreserveConfig defines the configuration for objects to preserve during cleanup.
// It supports multiple criteria for each object type including exact matches and regex patterns.
type PreserveConfig struct {
	Issues struct {
		PreserveByTitle []string `json:"preserve_by_title,omitempty"`
		PreserveByLabel []string `json:"preserve_by_label,omitempty"`
		PreserveByID    []string `json:"preserve_by_id,omitempty"`
	} `json:"issues,omitempty"`

	Discussions struct {
		PreserveByTitle    []string `json:"preserve_by_title,omitempty"`
		PreserveByCategory []string `json:"preserve_by_category,omitempty"`
		PreserveByID       []string `json:"preserve_by_id,omitempty"`
	} `json:"discussions,omitempty"`

	PullRequests struct {
		PreserveByTitle []string `json:"preserve_by_title,omitempty"`
		PreserveByLabel []string `json:"preserve_by_label,omitempty"`
		PreserveByID    []string `json:"preserve_by_id,omitempty"`
	} `json:"pull_requests,omitempty"`

	Labels struct {
		PreserveByName []string `json:"preserve_by_name,omitempty"`
	} `json:"labels,omitempty"`
}

// LoadPreserveConfig loads the preserve configuration from the specified file path.
// If the file doesn't exist, it returns an empty configuration (preserve nothing).
func LoadPreserveConfig(ctx context.Context, filePath string) (*PreserveConfig, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &PreserveConfig{}, nil
	}

	// Read file contents
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.FileError("read_preserve_config", "failed to read preserve configuration file", err)
	}

	// Parse JSON
	var config PreserveConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.FileError("parse_preserve_config", "failed to parse preserve configuration JSON", err)
	}

	return &config, nil
}

// ShouldPreserveIssue checks if an issue should be preserved based on the configuration.
func (pc *PreserveConfig) ShouldPreserveIssue(issue types.Issue) bool {
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
func (pc *PreserveConfig) ShouldPreserveDiscussion(discussion types.Discussion) bool {
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
func (pc *PreserveConfig) ShouldPreservePR(pr types.PullRequest) bool {
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
func (pc *PreserveConfig) ShouldPreserveLabel(labelName string) bool {
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

// DefaultPreserveConfigPath returns the default path for the preserve configuration file.
func DefaultPreserveConfigPath(configRoot string) string {
	return filepath.Join(configRoot, "preserve.json")
}