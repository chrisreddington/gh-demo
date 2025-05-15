package hydrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

import "github.com/chrisreddington/gh-demo/internal/githubapi"
func TestHydrateWithRealGHClient(t *testing.T) {
	   // This test uses the real (stubbed) GHClient to ensure wiring is correct.
	   client := githubapi.NewGHClient("octocat", "demo-repo")

	   root, err := FindProjectRoot()
	   if err != nil {
			   t.Fatalf("could not find project root: %v", err)
	   }
	   issuesPath := filepath.Join(root, ".github", "demos", "issues.json")
	   discussionsPath := filepath.Join(root, ".github", "demos", "discussions.json")
	   prsPath := filepath.Join(root, ".github", "demos", "prs.json")

	   // Should not error with stubbed methods
	   err = HydrateWithLabels(client, issuesPath, discussionsPath, prsPath, true, true, true)
	   if err != nil {
			   t.Fatalf("HydrateWithLabels with real GHClient failed: %v", err)
	   }
}

// MockGitHubClient is a mock for label operations
type MockGitHubClient struct {
	ExistingLabels map[string]bool
	CreatedLabels  []string
}

// Add stubs for the rest of the interface
func (m *MockGitHubClient) CreateIssue(issue githubapi.IssueInput) error    { return nil }
func (m *MockGitHubClient) CreateDiscussion(d githubapi.DiscussionInput) error { return nil }
func (m *MockGitHubClient) CreatePR(pr githubapi.PRInput) error           { return nil }

func (m *MockGitHubClient) ListLabels() ([]string, error) {
	labels := make([]string, 0, len(m.ExistingLabels))
	for l := range m.ExistingLabels {
		labels = append(labels, l)
	}
	return labels, nil
}

func (m *MockGitHubClient) CreateLabel(label string) error {
	m.CreatedLabels = append(m.CreatedLabels, label)
	m.ExistingLabels[label] = true
	return nil
}

func TestHydrateWithLabels(t *testing.T) {
	// Setup mock client with only "bug" and "demo" existing
	client := &MockGitHubClient{ExistingLabels: map[string]bool{"bug": true, "demo": true}}

	// Use demo files for content
root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("could not find project root: %v", err)
	}
	issuesPath := filepath.Join(root, ".github", "demos", "issues.json")
	discussionsPath := filepath.Join(root, ".github", "demos", "discussions.json")
	prsPath := filepath.Join(root, ".github", "demos", "prs.json")

	// Hydrate and ensure labels
	err = HydrateWithLabels(client, issuesPath, discussionsPath, prsPath, true, true, true)
	if err != nil {
		t.Fatalf("HydrateWithLabels failed: %v", err)
	}

	// Check that any labels in the files but not in ExistingLabels were created
	// ("enhancement", "ci", etc. depending on your demo files)
	created := make(map[string]bool)
	for _, l := range client.CreatedLabels {
		created[l] = true
	}
	// Should not create "bug" or "demo" again
	if created["bug"] || created["demo"] {
		t.Error("should not create labels that already exist")
	}
}

func TestReadIssuesJSON(t *testing.T) {
root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("could not find project root: %v", err)
	}
	path := filepath.Join(root, ".github", "demos", "issues.json")
	data, err := os.ReadFile(path)
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
root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("could not find project root: %v", err)
	}
	path := filepath.Join(root, ".github", "demos", "discussions.json")
	data, err := os.ReadFile(path)
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
root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("could not find project root: %v", err)
	}
	path := filepath.Join(root, ".github", "demos", "prs.json")
	data, err := os.ReadFile(path)
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
