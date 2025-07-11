// Package config provides application configuration constants and default values.
// This centralizes magic strings and configuration defaults to improve maintainability.
package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/chrisreddington/gh-demo/internal/errors"
	"github.com/chrisreddington/gh-demo/internal/types"
)

const (
	// DefaultConfigPath is the default directory for configuration files relative to project root
	DefaultConfigPath = ".github/demos"

	// DefaultLabelColor is the default color for auto-generated labels
	DefaultLabelColor = "ededed"

	// APITimeout is the default timeout for GitHub API operations
	APITimeout = 30 * time.Second

	// FileOperationTimeout is the timeout for file I/O operations
	FileOperationTimeout = 10 * time.Second

	// ProjectV2 defaults
	DefaultProjectVisibility = "private"
	DefaultProjectTitle      = "Repository Hydration Project"

	// Configuration file names
	IssuesFilename        = "issues.json"
	DiscussionsFilename   = "discussions.json"
	PullRequestsFilename  = "prs.json"
	LabelsFilename        = "labels.json"
	PreserveFilename      = "preserve.json"
	ProjectConfigFilename = "project-config.json"
)

// Configuration holds all configuration paths and provides validation.
// It standardizes the configuration pattern across the application.
type Configuration struct {
	// BasePath is the directory containing configuration files
	BasePath string

	// Computed file paths
	IssuesPath        string
	DiscussionsPath   string
	PullRequestsPath  string
	LabelsPath        string
	PreservePath      string
	ProjectConfigPath string
}

// NewConfiguration creates a new configuration with the given base path.
// The base path should be relative to the project root.
func NewConfiguration(ctx context.Context, basePath string) *Configuration {
	return &Configuration{
		BasePath:          basePath,
		IssuesPath:        filepath.Join(basePath, IssuesFilename),
		DiscussionsPath:   filepath.Join(basePath, DiscussionsFilename),
		PullRequestsPath:  filepath.Join(basePath, PullRequestsFilename),
		LabelsPath:        filepath.Join(basePath, LabelsFilename),
		PreservePath:      filepath.Join(basePath, PreserveFilename),
		ProjectConfigPath: filepath.Join(basePath, ProjectConfigFilename),
	}
}

// NewConfigurationWithRoot creates a new configuration with absolute paths
// resolved from the project root and base path.
func NewConfigurationWithRoot(ctx context.Context, projectRoot, basePath string) *Configuration {
	absoluteBasePath := filepath.Join(projectRoot, basePath)
	return &Configuration{
		BasePath:          absoluteBasePath,
		IssuesPath:        filepath.Join(absoluteBasePath, IssuesFilename),
		DiscussionsPath:   filepath.Join(absoluteBasePath, DiscussionsFilename),
		PullRequestsPath:  filepath.Join(absoluteBasePath, PullRequestsFilename),
		LabelsPath:        filepath.Join(absoluteBasePath, LabelsFilename),
		PreservePath:      filepath.Join(absoluteBasePath, PreserveFilename),
		ProjectConfigPath: filepath.Join(absoluteBasePath, ProjectConfigFilename),
	}
}

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
	// Check if context is cancelled before performing file operations
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &PreserveConfig{}, nil
	}

	// Check context again before reading file
	if err := ctx.Err(); err != nil {
		return nil, err
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

// LoadProjectConfiguration loads project configuration from the specified file path.
// If the file doesn't exist, it returns a default configuration.
// This provides a consistent way to load project settings across the application.
func LoadProjectConfiguration(ctx context.Context, filePath string) (*types.ProjectV2Configuration, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return default configuration
		return GetDefaultProjectConfiguration(), nil
	}

	// Check context before file operation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Read file contents
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.FileError("read_project_config", "failed to read project configuration file", err)
	}

	// Parse JSON
	var config types.ProjectV2Configuration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.FileError("parse_project_config", "failed to parse project configuration JSON", err)
	}

	// Apply defaults for missing values
	if config.Visibility == "" {
		config.Visibility = DefaultProjectVisibility
	}
	if config.Title == "" {
		config.Title = DefaultProjectTitle
	}

	return &config, nil
}

// GetDefaultProjectConfiguration returns a default project configuration
// with sensible defaults for repository hydration projects.
func GetDefaultProjectConfiguration() *types.ProjectV2Configuration {
	return &types.ProjectV2Configuration{
		Title:       DefaultProjectTitle,
		Description: "Project for organizing repository hydration content including issues, discussions, and pull requests.",
		Visibility:  DefaultProjectVisibility,
		Fields: []types.ProjectV2Field{
			{
				Name:        "Priority",
				Type:        "single_select",
				Description: "Content priority level",
				Options: []types.ProjectV2FieldOption{
					{Name: "High", Color: "d73a4a"},
					{Name: "Medium", Color: "fbca04"},
					{Name: "Low", Color: "0e8a16"},
				},
			},
			{
				Name:        "Status",
				Type:        "single_select",
				Description: "Content status",
				Options: []types.ProjectV2FieldOption{
					{Name: "To Do", Color: "ededed"},
					{Name: "In Progress", Color: "fbca04"},
					{Name: "Done", Color: "0e8a16"},
				},
			},
		},
		Views: []types.ProjectV2View{
			{
				Name:        "All Items",
				Description: "View of all project items",
				Layout:      "table",
				Fields:      []string{"title", "assignees", "status", "priority"},
			},
			{
				Name:        "By Status",
				Description: "Items grouped by status",
				Layout:      "board",
				GroupBy:     "status",
				Fields:      []string{"title", "assignees", "priority"},
			},
		},
	}
}

// CreateTimeoutContext creates a context with the default API timeout.
// This provides a consistent timeout pattern across all API operations.
func CreateTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, APITimeout)
}

// CreateFileOperationContext creates a context with the default file operation timeout.
// This provides a consistent timeout pattern for file I/O operations.
func CreateFileOperationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, FileOperationTimeout)
}

// CreateCustomTimeoutContext creates a context with a custom timeout duration.
// Use this when the default timeouts are not appropriate for specific operations.
func CreateCustomTimeoutContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
