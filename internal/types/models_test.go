package types

import (
	"testing"
)

// TestTypeStructures tests that all types have the expected field structures and values
func TestTypeStructures(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Issue type structure",
			testFunc: func(t *testing.T) {
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
			},
		},
		{
			name: "Discussion type structure",
			testFunc: func(t *testing.T) {
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
			},
		},
		{
			name: "PullRequest type structure",
			testFunc: func(t *testing.T) {
				pullRequest := PullRequest{
					Title:     "Test PR",
					Body:      "Test Body",
					Head:      "feature-branch",
					Base:      "main",
					Labels:    []string{"enhancement"},
					Assignees: []string{"reviewer1"},
				}

				if pullRequest.Title != "Test PR" {
					t.Errorf("Expected Title to be 'Test PR', got %s", pullRequest.Title)
				}
				if pullRequest.Head != "feature-branch" {
					t.Errorf("Expected Head to be 'feature-branch', got %s", pullRequest.Head)
				}
				if pullRequest.Base != "main" {
					t.Errorf("Expected Base to be 'main', got %s", pullRequest.Base)
				}
				if len(pullRequest.Labels) != 1 {
					t.Errorf("Expected 1 label, got %d", len(pullRequest.Labels))
				}
				if len(pullRequest.Assignees) != 1 {
					t.Errorf("Expected 1 assignee, got %d", len(pullRequest.Assignees))
				}
			},
		},
		{
			name: "Label type structure",
			testFunc: func(t *testing.T) {
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
