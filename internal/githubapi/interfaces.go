package githubapi

import "github.com/chrisreddington/gh-demo/internal/common"

type GitHubClient interface {
	ListLabels() ([]string, error)
	CreateLabel(label string) error
	CreateIssue(issue IssueInput) error
	CreateDiscussion(disc DiscussionInput) error
	CreatePR(pr PRInput) error
	SetLogger(logger common.Logger)
}
