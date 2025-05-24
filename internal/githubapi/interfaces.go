package githubapi

import (
	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// GitHubClient defines the interface for all GitHub API operations needed by the hydration process.
// This interface enables easy mocking for tests and ensures consistent API across different implementations.
// All methods should return appropriate errors when operations fail.
type GitHubClient interface {
	// ListLabels retrieves all existing labels from the repository
	ListLabels() ([]string, error)
	// CreateLabel creates a new label in the repository using the provided label data
	CreateLabel(label types.Label) error
	// CreateIssue creates a new issue in the repository using the provided issue data
	CreateIssue(issue types.Issue) error
	// CreateDiscussion creates a new discussion in the repository using the provided discussion data
	CreateDiscussion(discussion types.Discussion) error
	// CreatePR creates a new pull request in the repository using the provided pull request data
	CreatePR(pullRequest types.PullRequest) error
	// SetLogger sets the logger for debug output during API operations
	SetLogger(logger common.Logger)
}
