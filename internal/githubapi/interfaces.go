package githubapi

// Logger interface for debug and info logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
}

type GitHubClient interface {
	ListLabels() ([]string, error)
	CreateLabel(label string) error
	CreateIssue(issue IssueInput) error
	CreateDiscussion(disc DiscussionInput) error
	CreatePR(pr PRInput) error
	SetLogger(logger Logger)
}
