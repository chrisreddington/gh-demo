package testutil

import (
	"context"
	"fmt"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// ErrorConfig represents common error simulation configuration
type ErrorConfig struct {
	ShouldError  bool
	ErrorMessage string
}

// GetErrorOrDefault returns the configured error message or a default message
func (c *ErrorConfig) GetErrorOrDefault(defaultMsg string) error {
	if !c.ShouldError {
		return nil
	}
	msg := c.ErrorMessage
	if msg == "" {
		msg = defaultMsg
	}
	return fmt.Errorf("%s", msg)
}

// TestDataFactory provides common test data creation patterns
type TestDataFactory struct{}

// CreateTestIssue creates a test issue with default values
func (f *TestDataFactory) CreateTestIssue(title string) types.Issue {
	if title == "" {
		title = "Test Issue"
	}
	return types.Issue{
		Title:     title,
		Body:      "Test issue body",
		Labels:    []string{"test"},
		Assignees: []string{},
	}
}

// CreateTestPR creates a test pull request with default values
func (f *TestDataFactory) CreateTestPR(title string) types.PullRequest {
	if title == "" {
		title = "Test PR"
	}
	return types.PullRequest{
		Title:     title,
		Body:      "Test PR body",
		Head:      "feature-branch",
		Base:      "main",
		Labels:    []string{"test"},
		Assignees: []string{},
	}
}

// CreateTestDiscussion creates a test discussion with default values
func (f *TestDataFactory) CreateTestDiscussion(title string) types.Discussion {
	if title == "" {
		title = "Test Discussion"
	}
	return types.Discussion{
		Title:    title,
		Body:     "Test discussion body",
		Category: "General",
		Labels:   []string{"test"},
	}
}

// CreateTestLabel creates a test label with default values
func (f *TestDataFactory) CreateTestLabel(name string) types.Label {
	if name == "" {
		name = "test-label"
	}
	return types.Label{
		Name:        name,
		Color:       "FF0000",
		Description: "Test label description",
	}
}

// MockError provides a simple error implementation for testing
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}

// NewMockError creates a new MockError with the given message
func NewMockError(message string) error {
	return &MockError{Message: message}
}

// MockLogger provides a simple mock logger for testing
type MockLogger struct {
	LastMessage string
	DebugCalls  []string
	InfoCalls   []string
	ErrorCalls  []string
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.DebugCalls = append(m.DebugCalls, m.LastMessage)
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.InfoCalls = append(m.InfoCalls, m.LastMessage)
}

func (m *MockLogger) Error(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.ErrorCalls = append(m.ErrorCalls, m.LastMessage)
}

// Verify MockLogger implements common.Logger interface
var _ common.Logger = (*MockLogger)(nil)

// MockFactory provides common patterns for creating test mocks
type MockFactory struct{}

// CreateErrorConfigMap creates a map of ErrorConfig for multiple operations
func (f *MockFactory) CreateErrorConfigMap(errorOperations map[string]string) map[string]ErrorConfig {
	configs := make(map[string]ErrorConfig)
	for operation, errorMessage := range errorOperations {
		configs[operation] = ErrorConfig{
			ShouldError:  true,
			ErrorMessage: errorMessage,
		}
	}
	return configs
}

// CreateSuccessfulErrorConfig creates an ErrorConfig that doesn't trigger errors
func (f *MockFactory) CreateSuccessfulErrorConfig() ErrorConfig {
	return ErrorConfig{
		ShouldError:  false,
		ErrorMessage: "",
	}
}

// CreateLabelMap creates a map of labels with the specified existing state
func (f *MockFactory) CreateLabelMap(labels ...string) map[string]bool {
	labelMap := make(map[string]bool)
	for _, label := range labels {
		labelMap[label] = true
	}
	return labelMap
}

// CreateTestCollections creates test collections with the specified number of items
func (f *MockFactory) CreateTestCollections(numIssues, numPRs, numDiscussions int) ([]types.Issue, []types.PullRequest, []types.Discussion) {
	issues := make([]types.Issue, numIssues)
	for i := 0; i < numIssues; i++ {
		issues[i] = DataFactory.CreateTestIssue(fmt.Sprintf("Test Issue %d", i+1))
	}

	prs := make([]types.PullRequest, numPRs)
	for i := 0; i < numPRs; i++ {
		prs[i] = DataFactory.CreateTestPR(fmt.Sprintf("Test PR %d", i+1))
	}

	discussions := make([]types.Discussion, numDiscussions)
	for i := 0; i < numDiscussions; i++ {
		discussions[i] = DataFactory.CreateTestDiscussion(fmt.Sprintf("Test Discussion %d", i+1))
	}

	return issues, prs, discussions
}

// DefaultValues provides common default values used across different mock implementations
var DefaultValues = struct {
	IssueID          string
	IssueNumber      int
	PRID             string
	PRNumber         int
	DiscussionID     string
	DiscussionNumber int
	LabelID          string
	RepositoryID     string
	UserID           string
}{
	IssueID:          "issue-id-123",
	IssueNumber:      1,
	PRID:             "pr-id-123",
	PRNumber:         1,
	DiscussionID:     "discussion-id-123",
	DiscussionNumber: 1,
	LabelID:          "label-id-123",
	RepositoryID:     "repo-id-123",
	UserID:           "user-id-789",
}

// EmptyCollections provides empty slices for initializing mock collections
var EmptyCollections = struct {
	Issues      []types.Issue
	Discussions []types.Discussion
	PRs         []types.PullRequest
	Labels      []string
}{
	Issues:      []types.Issue{},
	Discussions: []types.Discussion{},
	PRs:         []types.PullRequest{},
	Labels:      []string{},
}

// Factory provides a global instance of MockFactory for convenience
var Factory = &MockFactory{}

// DataFactory provides a global instance of TestDataFactory for convenience  
var DataFactory = &TestDataFactory{}

// SimpleMockGraphQLClient provides a basic mock for GraphQL operations
type SimpleMockGraphQLClient struct {
	DoFunc func(context.Context, string, map[string]interface{}, interface{}) error
}

func (m *SimpleMockGraphQLClient) Do(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	if m.DoFunc != nil {
		return m.DoFunc(ctx, query, variables, response)
	}
	return nil
}