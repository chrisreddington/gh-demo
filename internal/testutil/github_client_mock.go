package testutil

import (
	"context"
	"fmt"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// GitHubClientMockConfig allows configuration of the mock GitHubClient behavior
type GitHubClientMockConfig struct {
	ExistingLabels      map[string]bool
	FailIssues          bool
	FailPRs             bool
	FailDiscussions     bool
	FailListLabels      bool
	FailCreateLabel     bool
	IssueErrorMsg       string
	PRErrorMsg          string
	DiscussionErrorMsg  string
	ListLabelsErrorMsg  string
	CreateLabelErrorMsg string
}

// GitHubClientMock provides a configurable mock implementation of GitHubClient
type GitHubClientMock struct {
	Config             GitHubClientMockConfig
	CreatedIssues      []types.Issue
	CreatedDiscussions []types.Discussion
	CreatedPRs         []types.PullRequest
	CreatedLabels      []string
	logger             common.Logger
}

// NewGitHubClientMock creates a new GitHub client mock
func NewGitHubClientMock() *GitHubClientMock {
	return &GitHubClientMock{
		Config: GitHubClientMockConfig{
			ExistingLabels: make(map[string]bool),
		},
		CreatedIssues:      []types.Issue{},
		CreatedDiscussions: []types.Discussion{},
		CreatedPRs:         []types.PullRequest{},
		CreatedLabels:      []string{},
	}
}

func (m *GitHubClientMock) CreateIssue(ctx context.Context, issue types.Issue) error {
	if m.Config.FailIssues {
		msg := m.Config.IssueErrorMsg
		if msg == "" {
			msg = fmt.Sprintf("simulated issue creation failure for: %s", issue.Title)
		}
		return fmt.Errorf("%s", msg)
	}
	m.CreatedIssues = append(m.CreatedIssues, issue)
	return nil
}

func (m *GitHubClientMock) CreateDiscussion(ctx context.Context, discussion types.Discussion) error {
	if m.Config.FailDiscussions {
		msg := m.Config.DiscussionErrorMsg
		if msg == "" {
			msg = fmt.Sprintf("simulated discussion creation failure for: %s", discussion.Title)
		}
		return fmt.Errorf("%s", msg)
	}
	m.CreatedDiscussions = append(m.CreatedDiscussions, discussion)
	return nil
}

func (m *GitHubClientMock) CreatePR(ctx context.Context, pullRequest types.PullRequest) error {
	if m.Config.FailPRs {
		msg := m.Config.PRErrorMsg
		if msg == "" {
			msg = fmt.Sprintf("simulated PR creation failure for: %s (head: %s, base: %s)", pullRequest.Title, pullRequest.Head, pullRequest.Base)
		}
		return fmt.Errorf("%s", msg)
	}
	m.CreatedPRs = append(m.CreatedPRs, pullRequest)
	return nil
}

func (m *GitHubClientMock) ListLabels(ctx context.Context) ([]string, error) {
	if m.Config.FailListLabels {
		msg := m.Config.ListLabelsErrorMsg
		if msg == "" {
			msg = "simulated list labels failure"
		}
		return nil, fmt.Errorf("%s", msg)
	}
	labels := make([]string, 0, len(m.Config.ExistingLabels))
	for l := range m.Config.ExistingLabels {
		labels = append(labels, l)
	}
	return labels, nil
}

func (m *GitHubClientMock) CreateLabel(ctx context.Context, label types.Label) error {
	if m.Config.FailCreateLabel {
		msg := m.Config.CreateLabelErrorMsg
		if msg == "" {
			msg = fmt.Sprintf("simulated create label failure for: %s", label.Name)
		}
		return fmt.Errorf("%s", msg)
	}
	m.CreatedLabels = append(m.CreatedLabels, label.Name)
	if m.Config.ExistingLabels == nil {
		m.Config.ExistingLabels = make(map[string]bool)
	}
	m.Config.ExistingLabels[label.Name] = true
	return nil
}

func (m *GitHubClientMock) SetLogger(logger common.Logger) {
	m.logger = logger
}

// Listing operations for cleanup
func (m *GitHubClientMock) ListIssues(ctx context.Context) ([]types.Issue, error) {
	return m.CreatedIssues, nil
}

func (m *GitHubClientMock) ListDiscussions(ctx context.Context) ([]types.Discussion, error) {
	return m.CreatedDiscussions, nil
}

func (m *GitHubClientMock) ListPRs(ctx context.Context) ([]types.PullRequest, error) {
	return m.CreatedPRs, nil
}

// Deletion operations for cleanup
func (m *GitHubClientMock) DeleteIssue(ctx context.Context, nodeID string) error {
	for i, issue := range m.CreatedIssues {
		if issue.NodeID == nodeID {
			m.CreatedIssues = append(m.CreatedIssues[:i], m.CreatedIssues[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *GitHubClientMock) DeleteDiscussion(ctx context.Context, nodeID string) error {
	for i, discussion := range m.CreatedDiscussions {
		if discussion.NodeID == nodeID {
			m.CreatedDiscussions = append(m.CreatedDiscussions[:i], m.CreatedDiscussions[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *GitHubClientMock) DeletePR(ctx context.Context, nodeID string) error {
	for i, pullRequest := range m.CreatedPRs {
		if pullRequest.NodeID == nodeID {
			m.CreatedPRs = append(m.CreatedPRs[:i], m.CreatedPRs[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *GitHubClientMock) DeleteLabel(ctx context.Context, name string) error {
	if m.Config.ExistingLabels != nil {
		delete(m.Config.ExistingLabels, name)
	}
	for i, label := range m.CreatedLabels {
		if label == name {
			m.CreatedLabels = append(m.CreatedLabels[:i], m.CreatedLabels[i+1:]...)
			break
		}
	}
	return nil
}

// NewSuccessfulGitHubClientMock creates a mock that succeeds for all operations
func NewSuccessfulGitHubClientMock(existingLabels ...string) *GitHubClientMock {
	labels := make(map[string]bool)
	for _, label := range existingLabels {
		labels[label] = true
	}

	return &GitHubClientMock{
		Config: GitHubClientMockConfig{
			ExistingLabels: labels,
		},
		CreatedIssues:      []types.Issue{},
		CreatedDiscussions: []types.Discussion{},
		CreatedPRs:         []types.PullRequest{},
		CreatedLabels:      []string{},
	}
}

// NewFailingGitHubClientMock creates a mock that fails for specified operations
func NewFailingGitHubClientMock(config GitHubClientMockConfig) *GitHubClientMock {
	if config.ExistingLabels == nil {
		config.ExistingLabels = make(map[string]bool)
	}

	return &GitHubClientMock{
		Config:             config,
		CreatedIssues:      []types.Issue{},
		CreatedDiscussions: []types.Discussion{},
		CreatedPRs:         []types.PullRequest{},
		CreatedLabels:      []string{},
	}
}