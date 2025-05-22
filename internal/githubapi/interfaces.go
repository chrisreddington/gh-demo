package githubapi

type GitHubClient interface {
	ListLabels() ([]string, error)
	CreateLabel(label string) error
	CreateIssue(issue IssueInput) error
	CreateDiscussion(disc DiscussionInput) error
	CreatePR(pr PRInput) error
}
