package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfiguration(t *testing.T) {
	basePath := "test/config"
	config := NewConfiguration(context.Background(), basePath)

	if config.BasePath != basePath {
		t.Errorf("Expected BasePath %s, got %s", basePath, config.BasePath)
	}

	expectedIssuesPath := filepath.Join(basePath, IssuesFilename)
	if config.IssuesPath != expectedIssuesPath {
		t.Errorf("Expected IssuesPath %s, got %s", expectedIssuesPath, config.IssuesPath)
	}

	expectedDiscussionsPath := filepath.Join(basePath, DiscussionsFilename)
	if config.DiscussionsPath != expectedDiscussionsPath {
		t.Errorf("Expected DiscussionsPath %s, got %s", expectedDiscussionsPath, config.DiscussionsPath)
	}

	expectedPRsPath := filepath.Join(basePath, PullRequestsFilename)
	if config.PullRequestsPath != expectedPRsPath {
		t.Errorf("Expected PullRequestsPath %s, got %s", expectedPRsPath, config.PullRequestsPath)
	}

	expectedLabelsPath := filepath.Join(basePath, LabelsFilename)
	if config.LabelsPath != expectedLabelsPath {
		t.Errorf("Expected LabelsPath %s, got %s", expectedLabelsPath, config.LabelsPath)
	}

	expectedPreservePath := filepath.Join(basePath, PreserveFilename)
	if config.PreservePath != expectedPreservePath {
		t.Errorf("Expected PreservePath %s, got %s", expectedPreservePath, config.PreservePath)
	}
}

func TestNewConfigurationWithRoot(t *testing.T) {
	projectRoot := "/project/root"
	basePath := "config/demo"

	config := NewConfigurationWithRoot(context.Background(), projectRoot, basePath)

	expectedBasePath := filepath.Join(projectRoot, basePath)
	if config.BasePath != expectedBasePath {
		t.Errorf("Expected BasePath %s, got %s", expectedBasePath, config.BasePath)
	}

	expectedIssuesPath := filepath.Join(expectedBasePath, IssuesFilename)
	if config.IssuesPath != expectedIssuesPath {
		t.Errorf("Expected IssuesPath %s, got %s", expectedIssuesPath, config.IssuesPath)
	}

	expectedPreservePath := filepath.Join(expectedBasePath, PreserveFilename)
	if config.PreservePath != expectedPreservePath {
		t.Errorf("Expected PreservePath %s, got %s", expectedPreservePath, config.PreservePath)
	}
}

func TestConfigurationConstants(t *testing.T) {
	// Test that file name constants are set correctly
	if IssuesFilename != "issues.json" {
		t.Errorf("Expected IssuesFilename 'issues.json', got %s", IssuesFilename)
	}
	if DiscussionsFilename != "discussions.json" {
		t.Errorf("Expected DiscussionsFilename 'discussions.json', got %s", DiscussionsFilename)
	}
	if PullRequestsFilename != "prs.json" {
		t.Errorf("Expected PullRequestsFilename 'prs.json', got %s", PullRequestsFilename)
	}
	if LabelsFilename != "labels.json" {
		t.Errorf("Expected LabelsFilename 'labels.json', got %s", LabelsFilename)
	}
	if PreserveFilename != "preserve.json" {
		t.Errorf("Expected PreserveFilename 'preserve.json', got %s", PreserveFilename)
	}
}

// TestLoadPreserveConfig tests loading preserve configuration from file
func TestLoadPreserveConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(t *testing.T) string
		expectError bool
		errorText   string
		validate    func(t *testing.T, config *PreserveConfig)
	}{
		{
			name: "non-existent file returns empty config",
			setupFile: func(t *testing.T) string {
				return "/non/existent/path/preserve.json"
			},
			expectError: false,
			validate: func(t *testing.T, config *PreserveConfig) {
				if config == nil {
					t.Error("Expected non-nil config for non-existent file")
					return
				}
				// Should return empty config
				if len(config.Issues.PreserveByTitle) != 0 {
					t.Error("Expected empty preserve config for non-existent file")
				}
			},
		},
		{
			name: "valid preserve config file",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "preserve.json")

				preserveConfig := PreserveConfig{
					Issues: struct {
						PreserveByTitle []string `json:"preserve_by_title,omitempty"`
						PreserveByLabel []string `json:"preserve_by_label,omitempty"`
						PreserveByID    []string `json:"preserve_by_id,omitempty"`
					}{
						PreserveByTitle: []string{"Important Issue", "^Release.*"},
						PreserveByLabel: []string{"permanent", "keep"},
						PreserveByID:    []string{"node123", "node456"},
					},
					Discussions: struct {
						PreserveByTitle    []string `json:"preserve_by_title,omitempty"`
						PreserveByCategory []string `json:"preserve_by_category,omitempty"`
						PreserveByID       []string `json:"preserve_by_id,omitempty"`
					}{
						PreserveByTitle:    []string{"Important Discussion"},
						PreserveByCategory: []string{"Announcements"},
						PreserveByID:       []string{"disc123"},
					},
					PullRequests: struct {
						PreserveByTitle []string `json:"preserve_by_title,omitempty"`
						PreserveByLabel []string `json:"preserve_by_label,omitempty"`
						PreserveByID    []string `json:"preserve_by_id,omitempty"`
					}{
						PreserveByTitle: []string{"Critical PR"},
						PreserveByLabel: []string{"hotfix"},
						PreserveByID:    []string{"pr123"},
					},
					Labels: struct {
						PreserveByName []string `json:"preserve_by_name,omitempty"`
					}{
						PreserveByName: []string{"bug", "enhancement"},
					},
				}

				data, err := json.Marshal(preserveConfig)
				if err != nil {
					t.Fatalf("Failed to marshal preserve config: %v", err)
				}

				if err := os.WriteFile(configPath, data, 0644); err != nil {
					t.Fatalf("Failed to write preserve config file: %v", err)
				}

				return configPath
			},
			expectError: false,
			validate: func(t *testing.T, config *PreserveConfig) {
				if config == nil {
					t.Error("Expected non-nil config")
					return
				}

				// Validate Issues config
				if len(config.Issues.PreserveByTitle) != 2 {
					t.Errorf("Expected 2 preserve by title entries, got %d", len(config.Issues.PreserveByTitle))
				}
				if config.Issues.PreserveByTitle[0] != "Important Issue" {
					t.Errorf("Expected 'Important Issue', got %s", config.Issues.PreserveByTitle[0])
				}
				if len(config.Issues.PreserveByLabel) != 2 {
					t.Errorf("Expected 2 preserve by label entries, got %d", len(config.Issues.PreserveByLabel))
				}
				if len(config.Issues.PreserveByID) != 2 {
					t.Errorf("Expected 2 preserve by ID entries, got %d", len(config.Issues.PreserveByID))
				}

				// Validate Discussions config
				if len(config.Discussions.PreserveByTitle) != 1 {
					t.Errorf("Expected 1 discussion preserve by title entry, got %d", len(config.Discussions.PreserveByTitle))
				}
				if len(config.Discussions.PreserveByCategory) != 1 {
					t.Errorf("Expected 1 discussion preserve by category entry, got %d", len(config.Discussions.PreserveByCategory))
				}

				// Validate PullRequests config
				if len(config.PullRequests.PreserveByTitle) != 1 {
					t.Errorf("Expected 1 PR preserve by title entry, got %d", len(config.PullRequests.PreserveByTitle))
				}

				// Validate Labels config
				if len(config.Labels.PreserveByName) != 2 {
					t.Errorf("Expected 2 label preserve by name entries, got %d", len(config.Labels.PreserveByName))
				}
			},
		},
		{
			name: "invalid JSON file",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "invalid.json")

				if err := os.WriteFile(configPath, []byte(`{"invalid": json}`), 0644); err != nil {
					t.Fatalf("Failed to write invalid JSON file: %v", err)
				}

				return configPath
			},
			expectError: true,
			errorText:   "parse_preserve_config",
		},
		{
			name: "empty JSON file",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "empty.json")

				if err := os.WriteFile(configPath, []byte(`{}`), 0644); err != nil {
					t.Fatalf("Failed to write empty JSON file: %v", err)
				}

				return configPath
			},
			expectError: false,
			validate: func(t *testing.T, config *PreserveConfig) {
				if config == nil {
					t.Error("Expected non-nil config for empty JSON")
					return
				}
				// Empty JSON should result in empty config
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			configPath := tt.setupFile(t)

			config, err := LoadPreserveConfig(ctx, configPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestLoadPreserveConfig_ContextCancellation tests context cancellation handling
func TestLoadPreserveConfig_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config, err := LoadPreserveConfig(ctx, "/any/path")

	if err == nil {
		t.Error("Expected context cancellation error")
		return
	}

	if config != nil {
		t.Error("Expected nil config on context cancellation")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}
