package githubapi

type IssueInput struct {
	Title     string
	Body      string
	Labels    []string
	Assignees []string
}

type DiscussionInput struct {
	Title    string
	Body     string
	Category string
	Labels   []string
}

type PRInput struct {
	Title     string
	Body      string
	Head      string
	Base      string
	Labels    []string
	Assignees []string
}
