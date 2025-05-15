package internal

import (
	"encoding/json"
	"os"
	"testing"
)

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

func TestReadIssuesJSON(t *testing.T) {
	data, err := os.ReadFile("../../gh-demo/.github/demos/issues.json")
	if err != nil {
		t.Fatalf("failed to read issues.json: %v", err)
	}
	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		t.Fatalf("failed to unmarshal issues.json: %v", err)
	}
	if len(issues) == 0 {
		t.Error("expected at least one issue in issues.json")
	}
}

func TestReadDiscussionsJSON(t *testing.T) {
	data, err := os.ReadFile("../../gh-demo/.github/demos/discussions.json")
	if err != nil {
		t.Fatalf("failed to read discussions.json: %v", err)
	}
	var discussions []Discussion
	if err := json.Unmarshal(data, &discussions); err != nil {
		t.Fatalf("failed to unmarshal discussions.json: %v", err)
	}
	if len(discussions) == 0 {
		t.Error("expected at least one discussion in discussions.json")
	}
}

func TestReadPRsJSON(t *testing.T) {
	data, err := os.ReadFile("../../gh-demo/.github/demos/prs.json")
	if err != nil {
		t.Fatalf("failed to read prs.json: %v", err)
	}
	var prs []PullRequest
	if err := json.Unmarshal(data, &prs); err != nil {
		t.Fatalf("failed to unmarshal prs.json: %v", err)
	}
	if len(prs) == 0 {
		t.Error("expected at least one PR in prs.json")
	}
}
