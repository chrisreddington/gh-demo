// Package config provides application configuration constants and default values.
// This centralizes magic strings and configuration defaults to improve maintainability.
package config

import "time"

const (
	// DefaultConfigPath is the default directory for configuration files relative to project root
	DefaultConfigPath = ".github/demos"

	// DefaultLabelColor is the default color for auto-generated labels
	DefaultLabelColor = "ededed"

	// APITimeout is the default timeout for GitHub API operations
	APITimeout = 30 * time.Second

	// FileOperationTimeout is the timeout for file I/O operations
	FileOperationTimeout = 10 * time.Second
)
