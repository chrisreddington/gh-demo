package hydrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// HydrateWithLabels loads content, collects all labels, and ensures labels exist before hydration.
func HydrateWithLabels(client GitHubLabelClient, issuesPath, discussionsPath, prsPath string, includeIssues, includeDiscussions, includePRs bool) error {
	issues, discussions, prs, err := HydrateFromFiles(issuesPath, discussionsPath, prsPath, includeIssues, includeDiscussions, includePRs)
	if err != nil {
		return err
	}
	labels := CollectLabels(issues, discussions, prs)
	if err := EnsureLabelsExist(client, labels); err != nil {
		return err
	}
	// (In a real implementation, would proceed to create issues/discussions/PRs here)
	return nil
}

type Issue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels"`
	Assignees []string `json:"assignees"`
}

type Discussion struct {
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Category string   `json:"category"`
	Labels   []string `json:"labels"`
}

type PullRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Head      string   `json:"head"`
	Base      string   `json:"base"`
	Labels    []string `json:"labels"`
	Assignees []string `json:"assignees"`
}

// GitHubLabelClient defines the interface for label operations (for mocking/testing)
type GitHubLabelClient interface {
	ListLabels() ([]string, error)
	CreateLabel(label string) error
}

// EnsureLabelsExist checks if each label exists in the repo, and creates it if not.
func EnsureLabelsExist(client GitHubLabelClient, labels []string) error {
	existing, err := client.ListLabels()
	if err != nil {
		return err
	}
	existSet := make(map[string]struct{}, len(existing))
	for _, l := range existing {
		existSet[l] = struct{}{}
	}
	for _, l := range labels {
		if _, ok := existSet[l]; !ok {
			if err := client.CreateLabel(l); err != nil {
				return err
			}
		}
	}
	return nil
}

func HydrateFromFiles(issuesPath, discussionsPath, prsPath string, includeIssues, includeDiscussions, includePRs bool) ([]Issue, []Discussion, []PullRequest, error) {
	var issues []Issue
	var discussions []Discussion
	var prs []PullRequest

	if includeIssues {
		data, err := os.ReadFile(issuesPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &issues); err != nil {
			return nil, nil, nil, err
		}
	}

	if includeDiscussions {
		data, err := os.ReadFile(discussionsPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &discussions); err != nil {
			return nil, nil, nil, err
		}
	}

	if includePRs {
		data, err := os.ReadFile(prsPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &prs); err != nil {
			return nil, nil, nil, err
		}
	}

	return issues, discussions, prs, nil
}

// CollectLabels returns a deduplicated list of all labels used in issues, discussions, and PRs.
func CollectLabels(issues []Issue, discussions []Discussion, prs []PullRequest) []string {
	labelSet := make(map[string]struct{})
	for _, i := range issues {
		for _, l := range i.Labels {
			labelSet[l] = struct{}{}
		}
	}
	for _, d := range discussions {
		for _, l := range d.Labels {
			labelSet[l] = struct{}{}
		}
	}
	for _, p := range prs {
		for _, l := range p.Labels {
			labelSet[l] = struct{}{}
		}
	}
	out := make([]string, 0, len(labelSet))
	for l := range labelSet {
		out = append(out, l)
	}
	return out
}

// findProjectRoot traverses up from the current file to find the directory containing go.mod
func findProjectRoot() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
