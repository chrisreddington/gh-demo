package types

import (
	"testing"
)

// TestIssueTypeStructure tests that the Issue type has the expected fields
func TestIssueTypeStructure(t *testing.T) {
	issue := Issue{
		Title:     "Test Issue",
		Body:      "Test Body",
		Labels:    []string{"bug", "feature"},
		Assignees: []string{"user1"},
	}

	if issue.Title != "Test Issue" {
		t.Errorf("Expected Title to be 'Test Issue', got %s", issue.Title)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(issue.Labels))
	}
	if len(issue.Assignees) != 1 {
		t.Errorf("Expected 1 assignee, got %d", len(issue.Assignees))
	}
}

// TestDiscussionTypeStructure tests that the Discussion type has the expected fields
func TestDiscussionTypeStructure(t *testing.T) {
	discussion := Discussion{
		Title:    "Test Discussion",
		Body:     "Test Body",
		Category: "Ideas",
		Labels:   []string{"question"},
	}

	if discussion.Title != "Test Discussion" {
		t.Errorf("Expected Title to be 'Test Discussion', got %s", discussion.Title)
	}
	if discussion.Category != "Ideas" {
		t.Errorf("Expected Category to be 'Ideas', got %s", discussion.Category)
	}
	if len(discussion.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(discussion.Labels))
	}
}

// TestPullRequestTypeStructure tests that the PullRequest type has the expected fields
func TestPullRequestTypeStructure(t *testing.T) {
	pr := PullRequest{
		Title:     "Test PR",
		Body:      "Test Body",
		Head:      "feature-branch",
		Base:      "main",
		Labels:    []string{"enhancement"},
		Assignees: []string{"reviewer1"},
	}

	if pr.Title != "Test PR" {
		t.Errorf("Expected Title to be 'Test PR', got %s", pr.Title)
	}
	if pr.Head != "feature-branch" {
		t.Errorf("Expected Head to be 'feature-branch', got %s", pr.Head)
	}
	if pr.Base != "main" {
		t.Errorf("Expected Base to be 'main', got %s", pr.Base)
	}
	if len(pr.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(pr.Labels))
	}
	if len(pr.Assignees) != 1 {
		t.Errorf("Expected 1 assignee, got %d", len(pr.Assignees))
	}
}

// TestLabelTypeStructure tests that the Label type has the expected fields
func TestLabelTypeStructure(t *testing.T) {
	label := Label{
		Name:        "bug",
		Description: "Something isn't working",
		Color:       "d73a4a",
	}

	if label.Name != "bug" {
		t.Errorf("Expected Name to be 'bug', got %s", label.Name)
	}
	if label.Description != "Something isn't working" {
		t.Errorf("Expected Description to be 'Something isn't working', got %s", label.Description)
	}
	if label.Color != "d73a4a" {
		t.Errorf("Expected Color to be 'd73a4a', got %s", label.Color)
	}
}
