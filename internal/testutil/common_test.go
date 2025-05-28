package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/chrisreddington/gh-demo/internal/common"
)

func TestErrorConfig_GetErrorOrDefault(t *testing.T) {
	tests := []struct {
		name        string
		config      ErrorConfig
		defaultMsg  string
		expectError bool
		expectedMsg string
	}{
		{
			name: "should not error when ShouldError is false",
			config: ErrorConfig{
				ShouldError:  false,
				ErrorMessage: "custom error",
			},
			defaultMsg:  "default error",
			expectError: false,
		},
		{
			name: "should return custom error message when set",
			config: ErrorConfig{
				ShouldError:  true,
				ErrorMessage: "custom error message",
			},
			defaultMsg:  "default error",
			expectError: true,
			expectedMsg: "custom error message",
		},
		{
			name: "should return default message when custom message is empty",
			config: ErrorConfig{
				ShouldError:  true,
				ErrorMessage: "",
			},
			defaultMsg:  "default error message",
			expectError: true,
			expectedMsg: "default error message",
		},
		{
			name: "should handle empty default message",
			config: ErrorConfig{
				ShouldError:  true,
				ErrorMessage: "",
			},
			defaultMsg:  "",
			expectError: true,
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.GetErrorOrDefault(tt.defaultMsg)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if err.Error() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestTestDataFactory_CreateTestIssue(t *testing.T) {
	factory := &TestDataFactory{}

	tests := []struct {
		name          string
		title         string
		expectedTitle string
	}{
		{
			name:          "creates issue with provided title",
			title:         "Custom Issue Title",
			expectedTitle: "Custom Issue Title",
		},
		{
			name:          "creates issue with default title when empty",
			title:         "",
			expectedTitle: "Test Issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := factory.CreateTestIssue(tt.title)

			if issue.Title != tt.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expectedTitle, issue.Title)
			}
			if issue.Body != "Test issue body" {
				t.Errorf("Expected body 'Test issue body', got '%s'", issue.Body)
			}
			if len(issue.Labels) != 1 || issue.Labels[0] != "test" {
				t.Errorf("Expected labels ['test'], got %v", issue.Labels)
			}
			if len(issue.Assignees) != 0 {
				t.Errorf("Expected empty assignees, got %v", issue.Assignees)
			}
		})
	}
}

func TestTestDataFactory_CreateTestPR(t *testing.T) {
	factory := &TestDataFactory{}

	tests := []struct {
		name          string
		title         string
		expectedTitle string
	}{
		{
			name:          "creates PR with provided title",
			title:         "Custom PR Title",
			expectedTitle: "Custom PR Title",
		},
		{
			name:          "creates PR with default title when empty",
			title:         "",
			expectedTitle: "Test PR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := factory.CreateTestPR(tt.title)

			if pr.Title != tt.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expectedTitle, pr.Title)
			}
			if pr.Body != "Test PR body" {
				t.Errorf("Expected body 'Test PR body', got '%s'", pr.Body)
			}
			if pr.Head != "feature-branch" {
				t.Errorf("Expected head 'feature-branch', got '%s'", pr.Head)
			}
			if pr.Base != "main" {
				t.Errorf("Expected base 'main', got '%s'", pr.Base)
			}
			if len(pr.Labels) != 1 || pr.Labels[0] != "test" {
				t.Errorf("Expected labels ['test'], got %v", pr.Labels)
			}
			if len(pr.Assignees) != 0 {
				t.Errorf("Expected empty assignees, got %v", pr.Assignees)
			}
		})
	}
}

func TestTestDataFactory_CreateTestDiscussion(t *testing.T) {
	factory := &TestDataFactory{}

	tests := []struct {
		name          string
		title         string
		expectedTitle string
	}{
		{
			name:          "creates discussion with provided title",
			title:         "Custom Discussion Title",
			expectedTitle: "Custom Discussion Title",
		},
		{
			name:          "creates discussion with default title when empty",
			title:         "",
			expectedTitle: "Test Discussion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discussion := factory.CreateTestDiscussion(tt.title)

			if discussion.Title != tt.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expectedTitle, discussion.Title)
			}
			if discussion.Body != "Test discussion body" {
				t.Errorf("Expected body 'Test discussion body', got '%s'", discussion.Body)
			}
			if discussion.Category != "General" {
				t.Errorf("Expected category 'General', got '%s'", discussion.Category)
			}
			if len(discussion.Labels) != 1 || discussion.Labels[0] != "test" {
				t.Errorf("Expected labels ['test'], got %v", discussion.Labels)
			}
		})
	}
}

func TestTestDataFactory_CreateTestLabel(t *testing.T) {
	factory := &TestDataFactory{}

	tests := []struct {
		name         string
		labelName    string
		expectedName string
	}{
		{
			name:         "creates label with provided name",
			labelName:    "custom-label",
			expectedName: "custom-label",
		},
		{
			name:         "creates label with default name when empty",
			labelName:    "",
			expectedName: "test-label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := factory.CreateTestLabel(tt.labelName)

			if label.Name != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, label.Name)
			}
			if label.Color != "FF0000" {
				t.Errorf("Expected color 'FF0000', got '%s'", label.Color)
			}
			if label.Description != "Test label description" {
				t.Errorf("Expected description 'Test label description', got '%s'", label.Description)
			}
		})
	}
}

func TestMockError(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedString string
	}{
		{
			name:           "returns correct error message",
			message:        "test error message",
			expectedString: "test error message",
		},
		{
			name:           "handles empty message",
			message:        "",
			expectedString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &MockError{Message: tt.message}
			if err.Error() != tt.expectedString {
				t.Errorf("Expected error string '%s', got '%s'", tt.expectedString, err.Error())
			}
		})
	}
}

func TestNewMockError(t *testing.T) {
	message := "test error from factory"
	err := NewMockError(message)

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if err.Error() != message {
		t.Errorf("Expected error message '%s', got '%s'", message, err.Error())
	}

	// Verify it's actually a MockError
	mockErr, ok := err.(*MockError)
	if !ok {
		t.Error("Expected error to be of type *MockError")
	}
	if mockErr.Message != message {
		t.Errorf("Expected MockError.Message '%s', got '%s'", message, mockErr.Message)
	}
}

func TestMockLogger(t *testing.T) {
	logger := &MockLogger{}

	// Verify it implements the common.Logger interface
	var _ common.Logger = logger

	t.Run("Debug method", func(t *testing.T) {
		logger.Debug("debug message: %s", "test")

		expectedMessage := "debug message: test"
		if logger.LastMessage != expectedMessage {
			t.Errorf("Expected LastMessage '%s', got '%s'", expectedMessage, logger.LastMessage)
		}
		if len(logger.DebugCalls) != 1 {
			t.Errorf("Expected 1 debug call, got %d", len(logger.DebugCalls))
		}
		if logger.DebugCalls[0] != expectedMessage {
			t.Errorf("Expected DebugCalls[0] '%s', got '%s'", expectedMessage, logger.DebugCalls[0])
		}
	})

	t.Run("Info method", func(t *testing.T) {
		logger.Info("info message: %s", "value")

		expectedMessage := "info message: value"
		if logger.LastMessage != expectedMessage {
			t.Errorf("Expected LastMessage '%s', got '%s'", expectedMessage, logger.LastMessage)
		}
		if len(logger.InfoCalls) != 1 {
			t.Errorf("Expected 1 info call, got %d", len(logger.InfoCalls))
		}
		if logger.InfoCalls[0] != expectedMessage {
			t.Errorf("Expected InfoCalls[0] '%s', got '%s'", expectedMessage, logger.InfoCalls[0])
		}
	})

	t.Run("Error method", func(t *testing.T) {
		logger.Error("error message: %d", 404)

		expectedMessage := "error message: 404"
		if logger.LastMessage != expectedMessage {
			t.Errorf("Expected LastMessage '%s', got '%s'", expectedMessage, logger.LastMessage)
		}
		if len(logger.ErrorCalls) != 1 {
			t.Errorf("Expected 1 error call, got %d", len(logger.ErrorCalls))
		}
		if logger.ErrorCalls[0] != expectedMessage {
			t.Errorf("Expected ErrorCalls[0] '%s', got '%s'", expectedMessage, logger.ErrorCalls[0])
		}
	})

	t.Run("multiple calls accumulate", func(t *testing.T) {
		// Reset logger
		logger = &MockLogger{}

		logger.Debug("debug 1")
		logger.Debug("debug 2")
		logger.Info("info 1")
		logger.Error("error 1")

		if len(logger.DebugCalls) != 2 {
			t.Errorf("Expected 2 debug calls, got %d", len(logger.DebugCalls))
		}
		if len(logger.InfoCalls) != 1 {
			t.Errorf("Expected 1 info call, got %d", len(logger.InfoCalls))
		}
		if len(logger.ErrorCalls) != 1 {
			t.Errorf("Expected 1 error call, got %d", len(logger.ErrorCalls))
		}
		if logger.LastMessage != "error 1" {
			t.Errorf("Expected LastMessage 'error 1', got '%s'", logger.LastMessage)
		}
	})
}

func TestMockFactory_CreateErrorConfigMap(t *testing.T) {
	factory := &MockFactory{}

	tests := []struct {
		name            string
		errorOperations map[string]string
		expectedCount   int
	}{
		{
			name: "creates config map with multiple operations",
			errorOperations: map[string]string{
				"create_issue": "issue creation failed",
				"create_pr":    "pr creation failed",
			},
			expectedCount: 2,
		},
		{
			name:            "handles empty map",
			errorOperations: map[string]string{},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := factory.CreateErrorConfigMap(tt.errorOperations)

			if len(configs) != tt.expectedCount {
				t.Errorf("Expected %d configs, got %d", tt.expectedCount, len(configs))
			}

			for operation, expectedMsg := range tt.errorOperations {
				config, exists := configs[operation]
				if !exists {
					t.Errorf("Expected config for operation '%s' to exist", operation)
					continue
				}
				if !config.ShouldError {
					t.Errorf("Expected ShouldError to be true for operation '%s'", operation)
				}
				if config.ErrorMessage != expectedMsg {
					t.Errorf("Expected ErrorMessage '%s' for operation '%s', got '%s'", expectedMsg, operation, config.ErrorMessage)
				}
			}
		})
	}
}

func TestMockFactory_CreateSuccessfulErrorConfig(t *testing.T) {
	factory := &MockFactory{}

	config := factory.CreateSuccessfulErrorConfig()

	if config.ShouldError {
		t.Error("Expected ShouldError to be false")
	}
	if config.ErrorMessage != "" {
		t.Errorf("Expected ErrorMessage to be empty, got '%s'", config.ErrorMessage)
	}
}

func TestMockFactory_CreateLabelMap(t *testing.T) {
	factory := &MockFactory{}

	tests := []struct {
		name          string
		labels        []string
		expectedCount int
	}{
		{
			name:          "creates map with multiple labels",
			labels:        []string{"bug", "feature", "enhancement"},
			expectedCount: 3,
		},
		{
			name:          "handles single label",
			labels:        []string{"bug"},
			expectedCount: 1,
		},
		{
			name:          "handles no labels",
			labels:        []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labelMap := factory.CreateLabelMap(tt.labels...)

			if len(labelMap) != tt.expectedCount {
				t.Errorf("Expected %d labels, got %d", tt.expectedCount, len(labelMap))
			}

			for _, label := range tt.labels {
				if !labelMap[label] {
					t.Errorf("Expected label '%s' to be true in map", label)
				}
			}
		})
	}
}

func TestMockFactory_CreateTestCollections(t *testing.T) {
	factory := &MockFactory{}

	tests := []struct {
		name           string
		numIssues      int
		numPRs         int
		numDiscussions int
	}{
		{
			name:           "creates collections with specified counts",
			numIssues:      3,
			numPRs:         2,
			numDiscussions: 1,
		},
		{
			name:           "handles zero counts",
			numIssues:      0,
			numPRs:         0,
			numDiscussions: 0,
		},
		{
			name:           "handles single items",
			numIssues:      1,
			numPRs:         1,
			numDiscussions: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues, prs, discussions := factory.CreateTestCollections(tt.numIssues, tt.numPRs, tt.numDiscussions)

			if len(issues) != tt.numIssues {
				t.Errorf("Expected %d issues, got %d", tt.numIssues, len(issues))
			}
			if len(prs) != tt.numPRs {
				t.Errorf("Expected %d PRs, got %d", tt.numPRs, len(prs))
			}
			if len(discussions) != tt.numDiscussions {
				t.Errorf("Expected %d discussions, got %d", tt.numDiscussions, len(discussions))
			}

			// Verify content of created items
			for i, issue := range issues {
				expectedTitle := fmt.Sprintf("Test Issue %d", i+1)
				if issue.Title != expectedTitle {
					t.Errorf("Expected issue %d title '%s', got '%s'", i, expectedTitle, issue.Title)
				}
			}

			for i, pr := range prs {
				expectedTitle := fmt.Sprintf("Test PR %d", i+1)
				if pr.Title != expectedTitle {
					t.Errorf("Expected PR %d title '%s', got '%s'", i, expectedTitle, pr.Title)
				}
			}

			for i, discussion := range discussions {
				expectedTitle := fmt.Sprintf("Test Discussion %d", i+1)
				if discussion.Title != expectedTitle {
					t.Errorf("Expected discussion %d title '%s', got '%s'", i, expectedTitle, discussion.Title)
				}
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	// Test that DefaultValues struct has expected values
	if DefaultValues.IssueID != "issue-id-123" {
		t.Errorf("Expected IssueID 'issue-id-123', got '%s'", DefaultValues.IssueID)
	}
	if DefaultValues.IssueNumber != 1 {
		t.Errorf("Expected IssueNumber 1, got %d", DefaultValues.IssueNumber)
	}
	if DefaultValues.PRID != "pr-id-123" {
		t.Errorf("Expected PRID 'pr-id-123', got '%s'", DefaultValues.PRID)
	}
	if DefaultValues.PRNumber != 1 {
		t.Errorf("Expected PRNumber 1, got %d", DefaultValues.PRNumber)
	}
	if DefaultValues.DiscussionID != "discussion-id-123" {
		t.Errorf("Expected DiscussionID 'discussion-id-123', got '%s'", DefaultValues.DiscussionID)
	}
	if DefaultValues.DiscussionNumber != 1 {
		t.Errorf("Expected DiscussionNumber 1, got %d", DefaultValues.DiscussionNumber)
	}
	if DefaultValues.LabelID != "label-id-123" {
		t.Errorf("Expected LabelID 'label-id-123', got '%s'", DefaultValues.LabelID)
	}
	if DefaultValues.RepositoryID != "repo-id-123" {
		t.Errorf("Expected RepositoryID 'repo-id-123', got '%s'", DefaultValues.RepositoryID)
	}
	if DefaultValues.UserID != "user-id-789" {
		t.Errorf("Expected UserID 'user-id-789', got '%s'", DefaultValues.UserID)
	}
}

func TestEmptyCollections(t *testing.T) {
	// Test that EmptyCollections struct has expected empty slices
	if len(EmptyCollections.Issues) != 0 {
		t.Errorf("Expected empty Issues slice, got length %d", len(EmptyCollections.Issues))
	}
	if len(EmptyCollections.Discussions) != 0 {
		t.Errorf("Expected empty Discussions slice, got length %d", len(EmptyCollections.Discussions))
	}
	if len(EmptyCollections.PRs) != 0 {
		t.Errorf("Expected empty PRs slice, got length %d", len(EmptyCollections.PRs))
	}
	if len(EmptyCollections.Labels) != 0 {
		t.Errorf("Expected empty Labels slice, got length %d", len(EmptyCollections.Labels))
	}
}

func TestGlobalInstances(t *testing.T) {
	// Test that global instances are properly initialized
	if Factory == nil {
		t.Error("Expected Factory to be initialized")
	}
	if DataFactory == nil {
		t.Error("Expected DataFactory to be initialized")
	}

	// Test that they are of the correct types
	_, factoryOk := interface{}(Factory).(*MockFactory)
	if !factoryOk {
		t.Error("Expected Factory to be of type *MockFactory")
	}
	_, dataFactoryOk := interface{}(DataFactory).(*TestDataFactory)
	if !dataFactoryOk {
		t.Error("Expected DataFactory to be of type *TestDataFactory")
	}
}

func TestSimpleMockGraphQLClient(t *testing.T) {
	t.Run("default behavior returns no error", func(t *testing.T) {
		client := &SimpleMockGraphQLClient{}
		ctx := context.Background()
		err := client.Do(ctx, "query", map[string]interface{}{}, nil)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("custom DoFunc is called", func(t *testing.T) {
		var capturedQuery string
		var capturedVariables map[string]interface{}

		client := &SimpleMockGraphQLClient{
			DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
				capturedQuery = query
				capturedVariables = variables
				return NewMockError("custom error")
			},
		}

		ctx := context.Background()
		testQuery := "test query"
		testVariables := map[string]interface{}{"key": "value"}

		err := client.Do(ctx, testQuery, testVariables, nil)

		if err == nil {
			t.Error("Expected error from custom DoFunc")
		}
		if err.Error() != "custom error" {
			t.Errorf("Expected error 'custom error', got '%s'", err.Error())
		}
		if capturedQuery != testQuery {
			t.Errorf("Expected query '%s', got '%s'", testQuery, capturedQuery)
		}
		if capturedVariables["key"] != "value" {
			t.Errorf("Expected variables to contain key='value', got %v", capturedVariables)
		}
	})
}
