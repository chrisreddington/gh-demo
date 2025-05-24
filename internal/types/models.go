// Package types contains common type definitions used across the application.
// This package centralizes all data structures to avoid duplication and ensure consistency.
package types

// Issue represents an issue that can be created in a GitHub repository.
// It contains all the fields that can be specified when creating an issue via the GitHub API.
type Issue struct {
	NodeID    string   `json:"node_id,omitempty"`    // GitHub node ID for deletion operations
	Number    int      `json:"number,omitempty"`     // Issue number for identification
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels"`
	Assignees []string `json:"assignees"`
}

// Discussion represents a discussion that can be created in a GitHub repository.
// It contains all the fields that can be specified when creating a discussion via the GitHub API.
type Discussion struct {
	NodeID   string   `json:"node_id,omitempty"`   // GitHub node ID for deletion operations
	Number   int      `json:"number,omitempty"`    // Discussion number for identification
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Category string   `json:"category"`
	Labels   []string `json:"labels"`
}

// PullRequest represents a pull request that can be created in a GitHub repository.
// It contains all the fields that can be specified when creating a pull request via the GitHub API.
type PullRequest struct {
	NodeID    string   `json:"node_id,omitempty"`   // GitHub node ID for deletion operations
	Number    int      `json:"number,omitempty"`    // Pull request number for identification
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Head      string   `json:"head"`
	Base      string   `json:"base"`
	Labels    []string `json:"labels"`
	Assignees []string `json:"assignees"`
}

// Label represents a label that can be created in a GitHub repository.
// It contains all the fields that can be specified when creating a label via the GitHub API.
type Label struct {
	// Name is the display name for the label
	Name string `json:"name"`
	// Description is an optional description for the label
	Description string `json:"description,omitempty"`
	// Color is the hexadecimal color code for the label (without the # prefix)
	Color string `json:"color"`
}
