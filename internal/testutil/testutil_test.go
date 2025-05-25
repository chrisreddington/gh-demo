package testutil

import (
	"context"
	"testing"
)

func TestErrorConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      ErrorConfig
		defaultMsg  string
		expectError bool
		expectedMsg string
	}{
		{
			name:        "no error configured",
			config:      ErrorConfig{ShouldError: false},
			defaultMsg:  "default error",
			expectError: false,
		},
		{
			name:        "error with custom message",
			config:      ErrorConfig{ShouldError: true, ErrorMessage: "custom error"},
			defaultMsg:  "default error",
			expectError: true,
			expectedMsg: "custom error",
		},
		{
			name:        "error with default message",
			config:      ErrorConfig{ShouldError: true},
			defaultMsg:  "default error",
			expectError: true,
			expectedMsg: "default error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.GetErrorOrDefault(tt.defaultMsg)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if err.Error() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestTestDataFactory(t *testing.T) {
	factory := NewTestDataFactory()

	t.Run("CreateTestIssue", func(t *testing.T) {
		issue := factory.CreateTestIssue("Custom Issue")
		if issue.Title != "Custom Issue" {
			t.Errorf("Expected title 'Custom Issue', got '%s'", issue.Title)
		}
		if issue.Body == "" {
			t.Error("Expected non-empty body")
		}
		if len(issue.Labels) == 0 {
			t.Error("Expected at least one label")
		}

		// Test with empty title (should use default)
		defaultIssue := factory.CreateTestIssue("")
		if defaultIssue.Title != "Test Issue" {
			t.Errorf("Expected default title 'Test Issue', got '%s'", defaultIssue.Title)
		}
	})

	t.Run("CreateTestPR", func(t *testing.T) {
		pr := factory.CreateTestPR("Custom PR")
		if pr.Title != "Custom PR" {
			t.Errorf("Expected title 'Custom PR', got '%s'", pr.Title)
		}
		if pr.Head == "" || pr.Base == "" {
			t.Error("Expected non-empty head and base branches")
		}
	})

	t.Run("CreateTestDiscussion", func(t *testing.T) {
		discussion := factory.CreateTestDiscussion("Custom Discussion")
		if discussion.Title != "Custom Discussion" {
			t.Errorf("Expected title 'Custom Discussion', got '%s'", discussion.Title)
		}
		if discussion.Category == "" {
			t.Error("Expected non-empty category")
		}
	})

	t.Run("CreateTestLabel", func(t *testing.T) {
		label := factory.CreateTestLabel("custom-label")
		if label.Name != "custom-label" {
			t.Errorf("Expected name 'custom-label', got '%s'", label.Name)
		}
		if label.Color == "" {
			t.Error("Expected non-empty color")
		}
	})
}

func TestMockLogger(t *testing.T) {
	logger := NewMockLogger()

	logger.Debug("debug message: %s", "value")
	logger.Info("info message: %s", "value")
	logger.Error("error message: %s", "value")

	if len(logger.DebugCalls) != 1 {
		t.Errorf("Expected 1 debug call, got %d", len(logger.DebugCalls))
	}
	if len(logger.InfoCalls) != 1 {
		t.Errorf("Expected 1 info call, got %d", len(logger.InfoCalls))
	}
	if len(logger.ErrorCalls) != 1 {
		t.Errorf("Expected 1 error call, got %d", len(logger.ErrorCalls))
	}

	if logger.LastMessage != "error message: value" {
		t.Errorf("Expected last message to be 'error message: value', got '%s'", logger.LastMessage)
	}
}

func TestGitHubClientMock(t *testing.T) {
	mock := NewSuccessfulGitHubClientMock("existing-label")
	factory := NewTestDataFactory()
	ctx := context.Background()

	t.Run("CreateIssue", func(t *testing.T) {
		issue := factory.CreateTestIssue("Test Issue")
		err := mock.CreateIssue(ctx, issue)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(mock.CreatedIssues) != 1 {
			t.Errorf("Expected 1 created issue, got %d", len(mock.CreatedIssues))
		}
	})

	t.Run("CreatePR", func(t *testing.T) {
		pr := factory.CreateTestPR("Test PR")
		err := mock.CreatePR(ctx, pr)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(mock.CreatedPRs) != 1 {
			t.Errorf("Expected 1 created PR, got %d", len(mock.CreatedPRs))
		}
	})

	t.Run("CreateDiscussion", func(t *testing.T) {
		discussion := factory.CreateTestDiscussion("Test Discussion")
		err := mock.CreateDiscussion(ctx, discussion)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(mock.CreatedDiscussions) != 1 {
			t.Errorf("Expected 1 created discussion, got %d", len(mock.CreatedDiscussions))
		}
	})

	t.Run("ListLabels", func(t *testing.T) {
		labels, err := mock.ListLabels(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(labels) != 1 || labels[0] != "existing-label" {
			t.Errorf("Expected ['existing-label'], got %v", labels)
		}
	})

	t.Run("CreateLabel", func(t *testing.T) {
		label := factory.CreateTestLabel("new-label")
		err := mock.CreateLabel(ctx, label)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(mock.CreatedLabels) != 1 {
			t.Errorf("Expected 1 created label, got %d", len(mock.CreatedLabels))
		}
	})
}

func TestGitHubClientMock_FailingScenarios(t *testing.T) {
	config := GitHubClientMockConfig{
		FailIssues:     true,
		IssueErrorMsg:  "custom issue error",
		FailPRs:        true,
		FailDiscussions: true,
	}
	mock := NewFailingGitHubClientMock(config)
	factory := NewTestDataFactory()
	ctx := context.Background()

	t.Run("CreateIssue_Fails", func(t *testing.T) {
		issue := factory.CreateTestIssue("Test Issue")
		err := mock.CreateIssue(ctx, issue)
		if err == nil {
			t.Error("Expected error but got nil")
		}
		if err.Error() != "custom issue error" {
			t.Errorf("Expected 'custom issue error', got '%s'", err.Error())
		}
	})

	t.Run("CreatePR_Fails", func(t *testing.T) {
		pr := factory.CreateTestPR("Test PR")
		err := mock.CreatePR(ctx, pr)
		if err == nil {
			t.Error("Expected error but got nil")
		}
		// Should use default error message since PRErrorMsg is empty
		if err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	})
}

func TestGraphQLMockClient(t *testing.T) {
	mock := NewDefaultGraphQLMock()
	ctx := context.Background()

	t.Run("HandleRepositoryQuery", func(t *testing.T) {
		response := &struct {
			Repository struct {
				ID string `json:"id"`
			} `json:"repository"`
		}{}

		err := mock.Do(ctx, "GetRepositoryId", nil, response)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if response.Repository.ID != "repo-id-123" {
			t.Errorf("Expected 'repo-id-123', got '%s'", response.Repository.ID)
		}
	})

	t.Run("HandleCreateIssueQuery", func(t *testing.T) {
		response := &struct {
			CreateIssue struct {
				Issue struct {
					ID     string `json:"id"`
					Number int    `json:"number"`
					Title  string `json:"title"`
					URL    string `json:"url"`
				} `json:"issue"`
			} `json:"createIssue"`
		}{}

		err := mock.Do(ctx, "createIssue", nil, response)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if response.CreateIssue.Issue.ID != "issue-id-123" {
			t.Errorf("Expected 'issue-id-123', got '%s'", response.CreateIssue.Issue.ID)
		}
	})
}

func TestGraphQLMockClient_ErrorScenarios(t *testing.T) {
	mock := NewErrorGraphQLMock(map[string]string{
		"repository": "unauthorized access",
	})
	ctx := context.Background()

	t.Run("HandleRepositoryQuery_Error", func(t *testing.T) {
		response := &struct {
			Repository struct {
				ID string `json:"id"`
			} `json:"repository"`
		}{}

		err := mock.Do(ctx, "GetRepositoryId", nil, response)
		if err == nil {
			t.Error("Expected error but got nil")
		}
		if err.Error() != "unauthorized access" {
			t.Errorf("Expected 'unauthorized access', got '%s'", err.Error())
		}
	})
}