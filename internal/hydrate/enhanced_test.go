package hydrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
)

// Enhanced MockGitHubClient that tracks created content
type EnhancedMockGitHubClient struct {
	ExistingLabels  map[string]bool
	CreatedLabels   []string
	CreatedIssues   []githubapi.IssueInput
	CreatedDiscs    []githubapi.DiscussionInput
	CreatedPRs      []githubapi.PRInput
}

func (m *EnhancedMockGitHubClient) ListLabels() ([]string, error) {
	labels := make([]string, 0, len(m.ExistingLabels))
	for l := range m.ExistingLabels {
		labels = append(labels, l)
	}
	return labels, nil
}

func (m *EnhancedMockGitHubClient) CreateLabel(label string) error {
	m.CreatedLabels = append(m.CreatedLabels, label)
	m.ExistingLabels[label] = true
	return nil
}

func (m *EnhancedMockGitHubClient) CreateIssue(issue githubapi.IssueInput) error {
	m.CreatedIssues = append(m.CreatedIssues, issue)
	return nil
}

func (m *EnhancedMockGitHubClient) CreateDiscussion(disc githubapi.DiscussionInput) error {
	m.CreatedDiscs = append(m.CreatedDiscs, disc)
	return nil
}

func (m *EnhancedMockGitHubClient) CreatePR(pr githubapi.PRInput) error {
	m.CreatedPRs = append(m.CreatedPRs, pr)
	return nil
}

func (m *EnhancedMockGitHubClient) SetLogger(logger githubapi.Logger) {
	// Mock implementation - does nothing
}

func TestHydrateFromFileErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hydrate-error-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existent files
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.json")
	_, _, _, err = HydrateFromFiles(nonExistentPath, nonExistentPath, nonExistentPath, true, false, false)
	if err == nil {
		t.Errorf("Expected error for nonexistent file, got nil")
	}

	// Test with invalid JSON for issues
	invalidIssuesPath := filepath.Join(tmpDir, "invalid-issues.json")
	if err := os.WriteFile(invalidIssuesPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("failed to write invalid issues file: %v", err)
	}

	_, _, _, err = HydrateFromFiles(invalidIssuesPath, nonExistentPath, nonExistentPath, true, false, false)
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got nil")
	}

	// Test with invalid JSON for discussions
	invalidDiscussionsPath := filepath.Join(tmpDir, "invalid-discussions.json")
	if err := os.WriteFile(invalidDiscussionsPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("failed to write invalid discussions file: %v", err)
	}

	_, _, _, err = HydrateFromFiles(nonExistentPath, invalidDiscussionsPath, nonExistentPath, false, true, false)
	if err == nil {
		t.Errorf("Expected error for invalid discussions JSON, got nil")
	}

	// Test with invalid JSON for PRs
	invalidPRsPath := filepath.Join(tmpDir, "invalid-prs.json")
	if err := os.WriteFile(invalidPRsPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("failed to write invalid PRs file: %v", err)
	}

	_, _, _, err = HydrateFromFiles(nonExistentPath, nonExistentPath, invalidPRsPath, false, false, true)
	if err == nil {
		t.Errorf("Expected error for invalid PRs JSON, got nil")
	}
}

func TestEnhancedHydrateWithLabels(t *testing.T) {
	// Setup enhanced mock client
	client := &EnhancedMockGitHubClient{
		ExistingLabels: map[string]bool{"bug": true, "demo": true},
		CreatedLabels:  []string{},
		CreatedIssues:  []githubapi.IssueInput{},
		CreatedDiscs:   []githubapi.DiscussionInput{},
		CreatedPRs:     []githubapi.PRInput{},
	}

	// Create temporary test files
	tmpDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test issues file
	issuesContent := `[
		{
			"title": "Test Issue",
			"body": "This is a test issue",
			"labels": ["bug", "enhancement"],
			"assignees": ["octocat"]
		}
	]`
	issuesPath := filepath.Join(tmpDir, "issues.json")
	if err := os.WriteFile(issuesPath, []byte(issuesContent), 0644); err != nil {
		t.Fatalf("failed to write issues file: %v", err)
	}

	// Create test discussions file
	discussionsContent := `[
		{
			"title": "Test Discussion",
			"body": "This is a test discussion",
			"category": "General",
			"labels": ["discussion", "question"]
		}
	]`
	discussionsPath := filepath.Join(tmpDir, "discussions.json")
	if err := os.WriteFile(discussionsPath, []byte(discussionsContent), 0644); err != nil {
		t.Fatalf("failed to write discussions file: %v", err)
	}

	// Create test PRs file
	prsContent := `[
		{
			"title": "Test PR",
			"body": "This is a test PR",
			"head": "feature-branch",
			"base": "main",
			"labels": ["enhancement", "ready-for-review"],
			"assignees": ["octocat"]
		}
	]`
	prsPath := filepath.Join(tmpDir, "prs.json")
	if err := os.WriteFile(prsPath, []byte(prsContent), 0644); err != nil {
		t.Fatalf("failed to write PRs file: %v", err)
	}

	// Test hydration with all content types
	err = HydrateWithLabels(client, issuesPath, discussionsPath, prsPath, true, true, true, false)
	if err != nil {
		t.Fatalf("HydrateWithLabels failed: %v", err)
	}

	// Verify labels were created correctly
	expectedLabels := []string{"enhancement", "discussion", "question", "ready-for-review"}
	created := make(map[string]bool)
	for _, l := range client.CreatedLabels {
		created[l] = true
	}
	
	// Should not create "bug" again as it already exists
	if created["bug"] {
		t.Error("should not create 'bug' label that already exists")
	}
	
	// Should have created all expected labels
	for _, label := range expectedLabels {
		if !created[label] {
			t.Errorf("expected label '%s' to be created, but it wasn't", label)
		}
	}

	// Verify issues were created correctly
	if len(client.CreatedIssues) != 1 {
		t.Errorf("expected 1 issue to be created, got %d", len(client.CreatedIssues))
	} else {
		issue := client.CreatedIssues[0]
		if issue.Title != "Test Issue" {
			t.Errorf("expected issue title 'Test Issue', got '%s'", issue.Title)
		}
	}

	// Verify discussions were created correctly
	if len(client.CreatedDiscs) != 1 {
		t.Errorf("expected 1 discussion to be created, got %d", len(client.CreatedDiscs))
	} else {
		disc := client.CreatedDiscs[0]
		if disc.Title != "Test Discussion" {
			t.Errorf("expected discussion title 'Test Discussion', got '%s'", disc.Title)
		}
		if disc.Category != "General" {
			t.Errorf("expected discussion category 'General', got '%s'", disc.Category)
		}
	}

	// Verify PRs were created correctly
	if len(client.CreatedPRs) != 1 {
		t.Errorf("expected 1 PR to be created, got %d", len(client.CreatedPRs))
	} else {
		pr := client.CreatedPRs[0]
		if pr.Title != "Test PR" {
			t.Errorf("expected PR title 'Test PR', got '%s'", pr.Title)
		}
		if pr.Head != "feature-branch" {
			t.Errorf("expected PR head 'feature-branch', got '%s'", pr.Head)
		}
		if pr.Base != "main" {
			t.Errorf("expected PR base 'main', got '%s'", pr.Base)
		}
	}

	// Test selective hydration (only issues)
	client = &EnhancedMockGitHubClient{
		ExistingLabels: map[string]bool{"bug": true, "demo": true},
		CreatedLabels:  []string{},
		CreatedIssues:  []githubapi.IssueInput{},
		CreatedDiscs:   []githubapi.DiscussionInput{},
		CreatedPRs:     []githubapi.PRInput{},
	}
	
	err = HydrateWithLabels(client, issuesPath, discussionsPath, prsPath, true, false, false, false)
	if err != nil {
		t.Fatalf("HydrateWithLabels with only issues failed: %v", err)
	}
	
	if len(client.CreatedIssues) != 1 {
		t.Errorf("expected 1 issue to be created, got %d", len(client.CreatedIssues))
	}
	
	if len(client.CreatedDiscs) != 0 {
		t.Errorf("expected 0 discussions to be created, got %d", len(client.CreatedDiscs))
	}
	
	if len(client.CreatedPRs) != 0 {
		t.Errorf("expected 0 PRs to be created, got %d", len(client.CreatedPRs))
	}
}