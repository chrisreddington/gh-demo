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

	// Configuration file names
	IssuesFilename       = "issues.json"
	DiscussionsFilename  = "discussions.json"
	PullRequestsFilename = "prs.json"
	LabelsFilename       = "labels.json"
	PreserveFilename     = "preserve.json"
)

// Configuration holds all configuration paths and provides validation.
// It standardizes the configuration pattern across the application.
type Configuration struct {
	// BasePath is the directory containing configuration files
	BasePath string

	// Computed file paths
	IssuesPath       string
	DiscussionsPath  string
	PullRequestsPath string
	LabelsPath       string
	PreservePath     string
}

// NewConfiguration creates a new configuration with the given base path.
// The base path should be relative to the project root.
func NewConfiguration(ctx context.Context, basePath string) *Configuration {
	return &Configuration{
		BasePath:         basePath,
		IssuesPath:       filepath.Join(basePath, IssuesFilename),
		DiscussionsPath:  filepath.Join(basePath, DiscussionsFilename),
		PullRequestsPath: filepath.Join(basePath, PullRequestsFilename),
		LabelsPath:       filepath.Join(basePath, LabelsFilename),
		PreservePath:     filepath.Join(basePath, PreserveFilename),
	}
}

// NewConfigurationWithRoot creates a new configuration with absolute paths
// resolved from the project root and base path.
func NewConfigurationWithRoot(ctx context.Context, projectRoot, basePath string) *Configuration {
	absoluteBasePath := filepath.Join(projectRoot, basePath)
	return &Configuration{
		BasePath:         absoluteBasePath,
		IssuesPath:       filepath.Join(absoluteBasePath, IssuesFilename),
		DiscussionsPath:  filepath.Join(absoluteBasePath, DiscussionsFilename),
		PullRequestsPath: filepath.Join(absoluteBasePath, PullRequestsFilename),
		LabelsPath:       filepath.Join(absoluteBasePath, LabelsFilename),
		PreservePath:     filepath.Join(absoluteBasePath, PreserveFilename),
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
