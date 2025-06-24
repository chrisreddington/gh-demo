package githubapi

import (
	"context"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// GitHubClient defines the interface for all GitHub API operations needed by the hydration process.
// This interface enables easy mocking for tests and ensures consistent API across different implementations.
// All methods should return appropriate errors when operations fail.
type GitHubClient interface {
	// Creation operations
	// ListLabels retrieves all existing labels from the repository
	ListLabels(ctx context.Context) ([]string, error)
	// CreateLabel creates a new label in the repository using the provided label data
	CreateLabel(ctx context.Context, label types.Label) error
	// CreateIssue creates a new issue and returns detailed information about the created item
	CreateIssue(ctx context.Context, issue types.Issue) (*types.CreatedItemInfo, error)
	// CreateDiscussion creates a new discussion and returns detailed information about the created item
	CreateDiscussion(ctx context.Context, discussion types.Discussion) (*types.CreatedItemInfo, error)
	// CreatePR creates a new pull request and returns detailed information about the created item
	CreatePR(ctx context.Context, pullRequest types.PullRequest) (*types.CreatedItemInfo, error)

	// Listing operations for cleanup
	// ListIssues retrieves all existing issues from the repository
	ListIssues(ctx context.Context) ([]types.Issue, error)
	// ListDiscussions retrieves all existing discussions from the repository
	ListDiscussions(ctx context.Context) ([]types.Discussion, error)
	// ListPRs retrieves all existing pull requests from the repository
	ListPRs(ctx context.Context) ([]types.PullRequest, error)

	// Deletion operations for cleanup
	// DeleteIssue deletes an issue by its node ID
	DeleteIssue(ctx context.Context, nodeID string) error
	// DeleteDiscussion deletes a discussion by its node ID
	DeleteDiscussion(ctx context.Context, nodeID string) error
	// DeletePR deletes a pull request by its node ID
	DeletePR(ctx context.Context, nodeID string) error
	// DeleteLabel deletes a label by its name
	DeleteLabel(ctx context.Context, name string) error

	// ProjectV2 operations
	// CreateProjectV2 creates a new ProjectV2 for the repository owner
	CreateProjectV2(ctx context.Context, config types.ProjectV2Configuration) (*types.ProjectV2, error)
	// ConfigureProjectV2Fields creates custom fields for a ProjectV2 based on the configuration
	ConfigureProjectV2Fields(ctx context.Context, projectID string, fields []types.ProjectV2Field) error
	// UpdateProjectV2Description updates the description of an existing ProjectV2
	UpdateProjectV2Description(ctx context.Context, projectID, description string) error
	// AddItemToProjectV2 adds an item (issue, PR, discussion) to a ProjectV2
	AddItemToProjectV2(ctx context.Context, projectID, itemNodeID string) error
	// GetProjectV2 retrieves project information by ID
	GetProjectV2(ctx context.Context, projectID string) (*types.ProjectV2, error)

	// SetLogger sets the logger for debug output during API operations
	SetLogger(logger common.Logger)
}
