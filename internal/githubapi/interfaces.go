package githubapi

import (
	"context"
)

// IssueInput represents the input for creating an issue
type IssueInput struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

// DiscussionInput represents the input for creating a discussion
type DiscussionInput struct {
	Title        string `json:"title"`
	Body         string `json:"body"`
	CategoryID   string `json:"categoryId"`
	RepositoryID string `json:"repositoryId"`
}

// PullRequestInput represents the input for creating a pull request
type PullRequestInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Draft bool   `json:"draft,omitempty"`
}

// LabelInput represents the input for creating a label
type LabelInput struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

// IssueClient defines the interface for working with GitHub issues
type IssueClient interface {
	CreateIssue(ctx context.Context, owner, repo string, input *IssueInput) (string, error)
}

// DiscussionClient defines the interface for working with GitHub discussions
type DiscussionClient interface {
	CreateDiscussion(ctx context.Context, input *DiscussionInput) (string, error)
	GetDiscussionCategories(ctx context.Context, owner, repo string) ([]DiscussionCategory, error)
}

// PullRequestClient defines the interface for working with GitHub pull requests
type PullRequestClient interface {
	CreatePullRequest(ctx context.Context, owner, repo string, input *PullRequestInput) (string, error)
}

// LabelClient defines the interface for working with GitHub labels
type LabelClient interface {
	CreateLabel(ctx context.Context, owner, repo string, input *LabelInput) (string, error)
}

// RepositoryClient defines the interface for working with GitHub repositories
type RepositoryClient interface {
	GetRepositoryID(ctx context.Context, owner, repo string) (string, error)
}

// GitHubClient combines all GitHub API client interfaces
type GitHubClient interface {
	IssueClient
	DiscussionClient
	PullRequestClient
	LabelClient
	RepositoryClient
}

// DiscussionCategory represents a GitHub discussion category
type DiscussionCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
