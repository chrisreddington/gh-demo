// Package config provides application configuration constants and default values.
// This centralizes magic strings and configuration defaults to improve maintainability.
package config

import (
	"context"
	"path/filepath"
	"time"
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
}

// NewConfiguration creates a new configuration with the given base path.
// The base path should be relative to the project root.
func NewConfiguration(basePath string) *Configuration {
	return &Configuration{
		BasePath:         basePath,
		IssuesPath:       filepath.Join(basePath, IssuesFilename),
		DiscussionsPath:  filepath.Join(basePath, DiscussionsFilename),
		PullRequestsPath: filepath.Join(basePath, PullRequestsFilename),
		LabelsPath:       filepath.Join(basePath, LabelsFilename),
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
	}
}
