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