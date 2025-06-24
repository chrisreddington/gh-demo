// Package types contains common type definitions used across the application.
// This package centralizes all data structures to avoid duplication and ensure consistency.
package types

// Issue represents an issue that can be created in a GitHub repository.
// It contains all the fields that can be specified when creating an issue via the GitHub API.
type Issue struct {
	NodeID    string   `json:"node_id,omitempty"` // GitHub node ID for deletion operations
	Number    int      `json:"number,omitempty"`  // Issue number for identification
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels"`
	Assignees []string `json:"assignees"`
}

// Discussion represents a discussion that can be created in a GitHub repository.
// It contains all the fields that can be specified when creating a discussion via the GitHub API.
type Discussion struct {
	NodeID   string   `json:"node_id,omitempty"` // GitHub node ID for deletion operations
	Number   int      `json:"number,omitempty"`  // Discussion number for identification
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Category string   `json:"category"`
	Labels   []string `json:"labels"`
}

// PullRequest represents a pull request that can be created in a GitHub repository.
// It contains all the fields that can be specified when creating a pull request via the GitHub API.
type PullRequest struct {
	NodeID    string   `json:"node_id,omitempty"` // GitHub node ID for deletion operations
	Number    int      `json:"number,omitempty"`  // Pull request number for identification
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

// ProjectV2 represents a GitHub ProjectV2 that can be created for organizing repository content.
// It contains all the fields that can be specified when creating a project via the GitHub API.
type ProjectV2 struct {
	NodeID      string `json:"node_id,omitempty"`     // GitHub node ID for project operations
	ID          string `json:"id,omitempty"`          // Project ID for API operations
	Number      int    `json:"number,omitempty"`      // Project number for identification
	Title       string `json:"title"`                 // Project title
	Description string `json:"description,omitempty"` // Project description
	Visibility  string `json:"visibility,omitempty"`  // Project visibility (private/public)
	URL         string `json:"url,omitempty"`         // Project URL
}

// ProjectV2Configuration defines the configuration for creating a ProjectV2.
// It provides options for customizing project creation with sensible defaults.
type ProjectV2Configuration struct {
	Title       string                  `json:"title"`                 // Project title (required)
	Description string                  `json:"description,omitempty"` // Project description
	Visibility  string                  `json:"visibility,omitempty"`  // Project visibility (private/public, defaults to private)
	Fields      []ProjectV2Field        `json:"fields,omitempty"`      // Custom project fields
	Views       []ProjectV2View         `json:"views,omitempty"`       // Project views/layouts
	Templates   []ProjectV2ItemTemplate `json:"templates,omitempty"`   // Item templates for different content types
}

// ProjectV2Field represents a custom field that can be added to a project.
type ProjectV2Field struct {
	Name        string                 `json:"name"`                  // Field name
	Type        string                 `json:"type"`                  // Field type (text, number, date, single_select, etc.)
	Description string                 `json:"description,omitempty"` // Field description
	Options     []ProjectV2FieldOption `json:"options,omitempty"`     // Options for select fields
}

// ProjectV2FieldOption represents an option for select-type project fields.
type ProjectV2FieldOption struct {
	Name        string `json:"name"`                  // Option name
	Description string `json:"description,omitempty"` // Option description
	Color       string `json:"color,omitempty"`       // Option color
}

// ProjectV2View represents a view/layout configuration for a project.
type ProjectV2View struct {
	Name        string                `json:"name"`                  // View name
	Description string                `json:"description,omitempty"` // View description
	Layout      string                `json:"layout,omitempty"`      // View layout (table, board, etc.)
	Fields      []string              `json:"fields,omitempty"`      // Fields to display in view
	Filters     []ProjectV2ViewFilter `json:"filters,omitempty"`     // View filters
	GroupBy     string                `json:"group_by,omitempty"`    // Field to group by
	SortBy      []ProjectV2ViewSort   `json:"sort_by,omitempty"`     // Sort configuration
}

// ProjectV2ViewFilter represents a filter for project views.
type ProjectV2ViewFilter struct {
	Field    string      `json:"field"`    // Field to filter on
	Operator string      `json:"operator"` // Filter operator (equals, contains, etc.)
	Value    interface{} `json:"value"`    // Filter value
}

// ProjectV2ViewSort represents sort configuration for project views.
type ProjectV2ViewSort struct {
	Field     string `json:"field"`               // Field to sort by
	Direction string `json:"direction,omitempty"` // Sort direction (asc/desc)
}

// ProjectV2ItemTemplate defines default field values for different content types.
type ProjectV2ItemTemplate struct {
	ContentType string                 `json:"content_type"`           // Content type (issue, pull_request, discussion)
	FieldValues map[string]interface{} `json:"field_values,omitempty"` // Default field values
	Description string                 `json:"description,omitempty"`  // Template description
}

// CreatedItemInfo represents information about a successfully created GitHub item.
type CreatedItemInfo struct {
	NodeID string // The GitHub node ID of the created item
	Title  string // The title of the created item
	Type   string // The type of item (issue, discussion, pull_request)
	Number int    // The GitHub number of the created item
	URL    string // The URL to the created item
}
