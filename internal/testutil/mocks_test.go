package testutil

import (
	"context"
	"testing"

	"github.com/chrisreddington/gh-demo/internal/types"
)

func TestMockGitHubClient(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, client *MockGitHubClient)
	}{
		{
			name: "creates new client with empty state",
			test: func(t *testing.T, client *MockGitHubClient) {
				if len(client.ExistingLabels) != 0 {
					t.Errorf("Expected empty ExistingLabels, got %d items", len(client.ExistingLabels))
				}
				if len(client.CreatedLabels) != 0 {
					t.Errorf("Expected empty CreatedLabels, got %d items", len(client.CreatedLabels))
				}
			},
		},
		{
			name: "creates issues successfully",
			test: func(t *testing.T, client *MockGitHubClient) {
				issue := types.Issue{Title: "Test Issue", Body: "Test body"}
				err := client.CreateIssue(context.Background(), issue)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(client.CreatedIssues) != 1 {
					t.Errorf("Expected 1 created issue, got %d", len(client.CreatedIssues))
				}
			},
		},
		{
			name: "fails issues when configured",
			test: func(t *testing.T, client *MockGitHubClient) {
				client.FailMethods["CreateIssue"] = true
				issue := types.Issue{Title: "Test Issue", Body: "Test body"}
				err := client.CreateIssue(context.Background(), issue)
				if err == nil {
					t.Error("Expected error when CreateIssue is configured to fail")
				}
			},
		},
		{
			name: "creates and lists labels",
			test: func(t *testing.T, client *MockGitHubClient) {
				label := types.Label{Name: "test", Color: "ff0000", Description: "Test label"}
				err := client.CreateLabel(context.Background(), label)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				labels, err := client.ListLabels(context.Background())
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if len(labels) != 1 || labels[0] != "test" {
					t.Errorf("Expected [test], got %v", labels)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockGitHubClient()
			tt.test(t, client)
		})
	}
}

func TestMockGraphQLClient(t *testing.T) {
	client := &MockGraphQLClient{}

	// Test that it doesn't panic with any GraphQL operation
	err := client.Do(context.Background(), "query { repository(owner: \"test\", name: \"test\") { id } }", nil, &struct{}{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockLogger(t *testing.T) {
	logger := &MockLogger{}

	logger.Debug("test debug: %s", "value")
	if logger.LastMessage != "test debug: value" {
		t.Errorf("Expected 'test debug: value', got '%s'", logger.LastMessage)
	}
	if !logger.DebugCalled {
		t.Error("Expected DebugCalled to be true")
	}

	logger.Info("test info: %s", "value")
	if logger.LastMessage != "test info: value" {
		t.Errorf("Expected 'test info: value', got '%s'", logger.LastMessage)
	}
	if !logger.InfoCalled {
		t.Error("Expected InfoCalled to be true")
	}
}
