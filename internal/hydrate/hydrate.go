package hydrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
)

// HydrateWithLabels loads content, collects all labels, and ensures labels exist before hydration.
// It continues processing even if individual items fail, collecting all errors and reporting them at the end.
func HydrateWithLabels(client githubapi.GitHubClient, issuesPath, discussionsPath, prsPath string, includeIssues, includeDiscussions, includePRs bool) error {
	issues, discussions, prs, err := HydrateFromFiles(issuesPath, discussionsPath, prsPath, includeIssues, includeDiscussions, includePRs)
	if err != nil {
		return err
	}
	
	// Collect and ensure all labels exist before creating content
	labels := CollectLabels(issues, discussions, prs)
	if err := EnsureLabelsExist(client, labels); err != nil {
		return fmt.Errorf("failed to ensure labels exist: %w", err)
	}
	
	var errors []string
	
	// Create issues
	if includeIssues {
		for i, issue := range issues {
			issueInput := githubapi.IssueInput{
				Title:     issue.Title,
				Body:      issue.Body,
				Labels:    issue.Labels,
				Assignees: issue.Assignees,
			}
			if err := client.CreateIssue(issueInput); err != nil {
				errors = append(errors, fmt.Sprintf("Issue %d (%s): %v", i+1, issue.Title, err))
			}
		}
	}
	
	// Create discussions
	if includeDiscussions {
		for i, discussion := range discussions {
			discussionInput := githubapi.DiscussionInput{
				Title:    discussion.Title,
				Body:     discussion.Body,
				Category: discussion.Category,
				Labels:   discussion.Labels,
			}
			if err := client.CreateDiscussion(discussionInput); err != nil {
				errors = append(errors, fmt.Sprintf("Discussion %d (%s): %v", i+1, discussion.Title, err))
			}
		}
	}
	
	// Create pull requests
	if includePRs {
		for i, pr := range prs {
			prInput := githubapi.PRInput{
				Title:     pr.Title,
				Body:      pr.Body,
				Head:      pr.Head,
				Base:      pr.Base,
				Labels:    pr.Labels,
				Assignees: pr.Assignees,
			}
			if err := client.CreatePR(prInput); err != nil {
				errors = append(errors, fmt.Sprintf("Pull Request %d (%s): %v", i+1, pr.Title, err))
			}
		}
	}
	
	// If any errors occurred, return them as a combined error but don't fail completely
	if len(errors) > 0 {
		return fmt.Errorf("some items failed to create:\n  - %s", strings.Join(errors, "\n  - "))
	}
	
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

// EnsureLabelsExist checks if each label exists in the repo, and creates it if not.
func EnsureLabelsExist(client githubapi.GitHubClient, labels []string) error {
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

// FindProjectRoot traverses up from the current file to find the directory containing go.mod
func FindProjectRoot() (string, error) {
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
