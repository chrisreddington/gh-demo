package hydrate

// Schema represents the structure of the hydrate configuration file
type Schema struct {
	Issues      []Issue      `json:"issues,omitempty"`
	Discussions []Discussion `json:"discussions,omitempty"`
	PRs         []PR         `json:"prs,omitempty"`
	Labels      []Label      `json:"labels,omitempty"`
}

// Issue represents a GitHub issue to be created
type Issue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

// Discussion represents a GitHub discussion to be created
type Discussion struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Category string `json:"category"`
}

// PR represents a GitHub pull request to be created
type PR struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Draft bool   `json:"draft,omitempty"`
}

// Label represents a GitHub label to be created
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}
