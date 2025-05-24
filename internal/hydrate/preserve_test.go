package hydrate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// TestLoadPreserveConfig tests loading preserve configuration from file
func TestLoadPreserveConfig(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "preserve_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		setupConfig    func(string) string // returns file path
		expectError    bool
		expectedConfig *config.PreserveConfig
	}{
		{
			name: "non-existent file returns empty config",
			setupConfig: func(dir string) string {
				return filepath.Join(dir, "nonexistent.json")
			},
			expectError:    false,
			expectedConfig: &config.PreserveConfig{},
		},
		{
			name: "valid config file",
			setupConfig: func(dir string) string {
				preserveConfig := config.PreserveConfig{}
				preserveConfig.Issues.PreserveByTitle = []string{"Important Issue"}
				preserveConfig.Issues.PreserveByLabel = []string{"permanent"}
				preserveConfig.Labels.PreserveByName = []string{"bug", "enhancement"}

				data, _ := json.Marshal(preserveConfig)
				path := filepath.Join(dir, "preserve.json")
				os.WriteFile(path, data, 0644)
				return path
			},
			expectError: false,
			expectedConfig: &config.PreserveConfig{
				Issues: struct {
					PreserveByTitle []string `json:"preserve_by_title,omitempty"`
					PreserveByLabel []string `json:"preserve_by_label,omitempty"`
					PreserveByID    []string `json:"preserve_by_id,omitempty"`
				}{
					PreserveByTitle: []string{"Important Issue"},
					PreserveByLabel: []string{"permanent"},
				},
				Labels: struct {
					PreserveByName []string `json:"preserve_by_name,omitempty"`
				}{
					PreserveByName: []string{"bug", "enhancement"},
				},
			},
		},
		{
			name: "invalid JSON returns error",
			setupConfig: func(dir string) string {
				path := filepath.Join(dir, "invalid.json")
				os.WriteFile(path, []byte("{invalid json"), 0644)
				return path
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupConfig(tempDir)
			
			config, err := config.LoadPreserveConfig(context.Background(), filePath)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if !tt.expectError {
				// Compare relevant fields
				if len(config.Issues.PreserveByTitle) != len(tt.expectedConfig.Issues.PreserveByTitle) {
					t.Errorf("Expected %d PreserveByTitle items, got %d", 
						len(tt.expectedConfig.Issues.PreserveByTitle), len(config.Issues.PreserveByTitle))
				}
				if len(config.Labels.PreserveByName) != len(tt.expectedConfig.Labels.PreserveByName) {
					t.Errorf("Expected %d PreserveByName items, got %d", 
						len(tt.expectedConfig.Labels.PreserveByName), len(config.Labels.PreserveByName))
				}
			}
		})
	}
}

// TestShouldPreserveIssue tests issue preservation logic
func TestShouldPreserveIssue(t *testing.T) {
	preserveConfig := &config.PreserveConfig{}
	preserveConfig.Issues.PreserveByTitle = []string{"Important Issue", "^Release.*"}
	preserveConfig.Issues.PreserveByLabel = []string{"permanent", "keep"}
	preserveConfig.Issues.PreserveByID = []string{"node123", "node456"}

	tests := []struct {
		name     string
		issue    types.Issue
		expected bool
	}{
		{
			name: "preserve by exact title",
			issue: types.Issue{
				Title: "Important Issue",
			},
			expected: true,
		},
		{
			name: "preserve by regex title",
			issue: types.Issue{
				Title: "Release v1.0.0",
			},
			expected: true,
		},
		{
			name: "preserve by label",
			issue: types.Issue{
				Title:  "Some Issue",
				Labels: []string{"bug", "permanent"},
			},
			expected: true,
		},
		{
			name: "preserve by node ID",
			issue: types.Issue{
				NodeID: "node123",
				Title:  "Any Issue",
			},
			expected: true,
		},
		{
			name: "do not preserve",
			issue: types.Issue{
				Title:  "Regular Issue",
				Labels: []string{"bug"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldPreserveIssue(preserveConfig, tt.issue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for issue: %s", tt.expected, result, tt.issue.Title)
			}
		})
	}
}

// TestShouldPreserveDiscussion tests discussion preservation logic
func TestShouldPreserveDiscussion(t *testing.T) {
	preserveConfig := &config.PreserveConfig{}
	preserveConfig.Discussions.PreserveByTitle = []string{"Welcome Discussion"}
	preserveConfig.Discussions.PreserveByCategory = []string{"Announcements"}
	preserveConfig.Discussions.PreserveByID = []string{"disc123"}

	tests := []struct {
		name       string
		discussion types.Discussion
		expected   bool
	}{
		{
			name: "preserve by title",
			discussion: types.Discussion{
				Title: "Welcome Discussion",
			},
			expected: true,
		},
		{
			name: "preserve by category",
			discussion: types.Discussion{
				Title:    "Some Announcement",
				Category: "Announcements",
			},
			expected: true,
		},
		{
			name: "preserve by node ID",
			discussion: types.Discussion{
				NodeID: "disc123",
				Title:  "Any Discussion",
			},
			expected: true,
		},
		{
			name: "do not preserve",
			discussion: types.Discussion{
				Title:    "Regular Discussion",
				Category: "General",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldPreserveDiscussion(preserveConfig, tt.discussion)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for discussion: %s", tt.expected, result, tt.discussion.Title)
			}
		})
	}
}

// TestShouldPreservePR tests pull request preservation logic
func TestShouldPreservePR(t *testing.T) {
	preserveConfig := &config.PreserveConfig{}
	preserveConfig.PullRequests.PreserveByTitle = []string{"^Release.*"}
	preserveConfig.PullRequests.PreserveByLabel = []string{"release"}
	preserveConfig.PullRequests.PreserveByID = []string{"pr123"}

	tests := []struct {
		name        string
		pullRequest types.PullRequest
		expected    bool
	}{
		{
			name: "preserve by regex title",
			pullRequest: types.PullRequest{
				Title: "Release v2.0.0",
			},
			expected: true,
		},
		{
			name: "preserve by label",
			pullRequest: types.PullRequest{
				Title:  "Some PR",
				Labels: []string{"feature", "release"},
			},
			expected: true,
		},
		{
			name: "preserve by node ID",
			pullRequest: types.PullRequest{
				NodeID: "pr123",
				Title:  "Any PR",
			},
			expected: true,
		},
		{
			name: "do not preserve",
			pullRequest: types.PullRequest{
				Title:  "Regular PR",
				Labels: []string{"feature"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldPreservePR(preserveConfig, tt.pullRequest)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for PR: %s", tt.expected, result, tt.pullRequest.Title)
			}
		})
	}
}

// TestShouldPreserveLabel tests label preservation logic
func TestShouldPreserveLabel(t *testing.T) {
	preserveConfig := &config.PreserveConfig{}
	preserveConfig.Labels.PreserveByName = []string{"bug", "enhancement", "help wanted"}

	tests := []struct {
		name      string
		labelName string
		expected  bool
	}{
		{
			name:      "preserve existing label",
			labelName: "bug",
			expected:  true,
		},
		{
			name:      "preserve another existing label",
			labelName: "enhancement",
			expected:  true,
		},
		{
			name:      "do not preserve unlisted label",
			labelName: "feature",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldPreserveLabel(preserveConfig, tt.labelName)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for label: %s", tt.expected, result, tt.labelName)
			}
		})
	}
}

// TestIsMatchOrRegex tests the regex matching helper function
func TestIsMatchOrRegex(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			value:    "hello world",
			pattern:  "hello world",
			expected: true,
		},
		{
			name:     "regex match with ^",
			value:    "Release v1.0.0",
			pattern:  "^Release.*",
			expected: true,
		},
		{
			name:     "regex match with .*",
			value:    "test string here",
			pattern:  "test.*here",
			expected: true,
		},
		{
			name:     "no match",
			value:    "hello world",
			pattern:  "goodbye",
			expected: false,
		},
		{
			name:     "invalid regex falls back to exact match",
			value:    "test[",
			pattern:  "test[",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMatchOrRegex(tt.value, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for value '%s' and pattern '%s'", 
					tt.expected, result, tt.value, tt.pattern)
			}
		})
	}
}