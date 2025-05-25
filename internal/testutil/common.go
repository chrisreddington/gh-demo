package testutil

import (
	"fmt"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// ErrorConfig represents common error simulation configuration
type ErrorConfig struct {
	ShouldError  bool
	ErrorMessage string
}

// MockError implements error interface for testing
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}

// NewMockError creates a new mock error with the given message
func NewMockError(message string) *MockError {
	return &MockError{Message: message}
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

// ResponseBuilder provides common patterns for building mock responses
type ResponseBuilder struct {
	DefaultIssueID          string
	DefaultIssueNumber      int
	DefaultPRID             string
	DefaultPRNumber         int
	DefaultDiscussionID     string
	DefaultDiscussionNumber int
	DefaultLabelID          string
	DefaultRepositoryID     string
	DefaultUserID           string
}

// NewResponseBuilder creates a ResponseBuilder with sensible defaults
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		DefaultIssueID:          "issue-id-123",
		DefaultIssueNumber:      1,
		DefaultPRID:             "pr-id-123",
		DefaultPRNumber:         1,
		DefaultDiscussionID:     "discussion-id-123",
		DefaultDiscussionNumber: 1,
		DefaultLabelID:          "label-id-123",
		DefaultRepositoryID:     "repo-id-123",
		DefaultUserID:           "user-id-789",
	}
}

// TestDataFactory provides common test data creation patterns
type TestDataFactory struct{}

// NewTestDataFactory creates a new TestDataFactory
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

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

// MockLogger provides a simple mock logger for testing
type MockLogger struct {
	LastMessage string
	LastArgs    []interface{}
	DebugCalls  []string
	InfoCalls   []string
	ErrorCalls  []string
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		DebugCalls: make([]string, 0),
		InfoCalls:  make([]string, 0),
		ErrorCalls: make([]string, 0),
	}
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.LastArgs = args
	m.DebugCalls = append(m.DebugCalls, m.LastMessage)
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.LastArgs = args
	m.InfoCalls = append(m.InfoCalls, m.LastMessage)
}

func (m *MockLogger) Error(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.LastArgs = args
	m.ErrorCalls = append(m.ErrorCalls, m.LastMessage)
}

// Verify MockLogger implements common.Logger interface
var _ common.Logger = (*MockLogger)(nil)