package hydrate

import (
	"context"
	"fmt"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// MockConfig allows configuration of the mock GitHubClient behavior
type MockConfig struct {
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

// ConfigurableMockGitHubClient provides a configurable mock implementation of GitHubClient
type ConfigurableMockGitHubClient struct {
	Config             MockConfig
	CreatedIssues      []types.Issue
	CreatedDiscussions []types.Discussion
	CreatedPRs         []types.PullRequest
	CreatedLabels      []string
	logger             common.Logger
}

func (m *ConfigurableMockGitHubClient) CreateIssue(ctx context.Context, issue types.Issue) error {
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

func (m *ConfigurableMockGitHubClient) CreateDiscussion(ctx context.Context, discussion types.Discussion) error {
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

func (m *ConfigurableMockGitHubClient) CreatePR(ctx context.Context, pullRequest types.PullRequest) error {
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

func (m *ConfigurableMockGitHubClient) ListLabels(ctx context.Context) ([]string, error) {
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

func (m *ConfigurableMockGitHubClient) CreateLabel(ctx context.Context, label types.Label) error {
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

func (m *ConfigurableMockGitHubClient) SetLogger(logger common.Logger) {
	m.logger = logger
}

// Helper functions to create common mock configurations

// NewSuccessfulMockGitHubClient creates a mock that succeeds for all operations
func NewSuccessfulMockGitHubClient(existingLabels ...string) *ConfigurableMockGitHubClient {
	labels := make(map[string]bool)
	for _, label := range existingLabels {
		labels[label] = true
	}

	return &ConfigurableMockGitHubClient{
		Config: MockConfig{
			ExistingLabels: labels,
		},
		CreatedIssues:      []types.Issue{},
		CreatedDiscussions: []types.Discussion{},
		CreatedPRs:         []types.PullRequest{},
		CreatedLabels:      []string{},
	}
}

// NewFailingMockGitHubClient creates a mock that fails for specified operations
func NewFailingMockGitHubClient(config MockConfig) *ConfigurableMockGitHubClient {
	if config.ExistingLabels == nil {
		config.ExistingLabels = make(map[string]bool)
	}

	return &ConfigurableMockGitHubClient{
		Config:             config,
		CreatedIssues:      []types.Issue{},
		CreatedDiscussions: []types.Discussion{},
		CreatedPRs:         []types.PullRequest{},
		CreatedLabels:      []string{},
	}
}
