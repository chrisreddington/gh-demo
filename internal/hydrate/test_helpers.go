package hydrate

import (
	"context"
	"fmt"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/errors"
	"github.com/chrisreddington/gh-demo/internal/testutil"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// MockConfig allows configuration of the mock GitHubClient behavior
type MockConfig struct {
	ExistingLabels          map[string]bool
	Issues                  testutil.ErrorConfig
	PRs                     testutil.ErrorConfig
	Discussions             testutil.ErrorConfig
	ListLabels              testutil.ErrorConfig
	CreateLabel             testutil.ErrorConfig
	FailProjectCreation     bool
	FailProjectItemAddition bool
	FailProjectRetrieval    bool
	FailProjectFieldConfiguration bool
	FailProjectDescriptionUpdate bool
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

func (m *ConfigurableMockGitHubClient) CreateIssue(ctx context.Context, issue types.Issue) (*types.CreatedItemInfo, error) {
	if err := m.Config.Issues.GetErrorOrDefault(fmt.Sprintf("simulated issue creation failure for: %s", issue.Title)); err != nil {
		return nil, err
	}
	m.CreatedIssues = append(m.CreatedIssues, issue)
	return &types.CreatedItemInfo{
		NodeID: fmt.Sprintf("mock-issue-id-%d", len(m.CreatedIssues)),
		Title:  issue.Title,
		Type:   "issue",
		Number: len(m.CreatedIssues),
		URL:    fmt.Sprintf("https://github.com/owner/repo/issues/%d", len(m.CreatedIssues)),
	}, nil
}

func (m *ConfigurableMockGitHubClient) CreateDiscussion(ctx context.Context, discussion types.Discussion) (*types.CreatedItemInfo, error) {
	if err := m.Config.Discussions.GetErrorOrDefault(fmt.Sprintf("simulated discussion creation failure for: %s", discussion.Title)); err != nil {
		return nil, err
	}
	m.CreatedDiscussions = append(m.CreatedDiscussions, discussion)
	return &types.CreatedItemInfo{
		NodeID: fmt.Sprintf("mock-discussion-id-%d", len(m.CreatedDiscussions)),
		Title:  discussion.Title,
		Type:   "discussion",
		Number: len(m.CreatedDiscussions),
		URL:    fmt.Sprintf("https://github.com/owner/repo/discussions/%d", len(m.CreatedDiscussions)),
	}, nil
}

func (m *ConfigurableMockGitHubClient) CreatePR(ctx context.Context, pullRequest types.PullRequest) (*types.CreatedItemInfo, error) {
	if err := m.Config.PRs.GetErrorOrDefault(fmt.Sprintf("simulated PR creation failure for: %s (head: %s, base: %s)", pullRequest.Title, pullRequest.Head, pullRequest.Base)); err != nil {
		return nil, err
	}
	m.CreatedPRs = append(m.CreatedPRs, pullRequest)
	return &types.CreatedItemInfo{
		NodeID: fmt.Sprintf("mock-pr-id-%d", len(m.CreatedPRs)),
		Title:  pullRequest.Title,
		Type:   "pull_request",
		Number: len(m.CreatedPRs),
		URL:    fmt.Sprintf("https://github.com/owner/repo/pull/%d", len(m.CreatedPRs)),
	}, nil
}

func (m *ConfigurableMockGitHubClient) ListLabels(ctx context.Context) ([]string, error) {
	if err := m.Config.ListLabels.GetErrorOrDefault("simulated list labels failure"); err != nil {
		return nil, err
	}
	labels := make([]string, 0, len(m.Config.ExistingLabels))
	for l := range m.Config.ExistingLabels {
		labels = append(labels, l)
	}
	return labels, nil
}

func (m *ConfigurableMockGitHubClient) CreateLabel(ctx context.Context, label types.Label) error {
	if err := m.Config.CreateLabel.GetErrorOrDefault(fmt.Sprintf("simulated create label failure for: %s", label.Name)); err != nil {
		return err
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

// Listing operations for cleanup
func (m *ConfigurableMockGitHubClient) ListIssues(ctx context.Context) ([]types.Issue, error) {
	// For testing, return created issues
	return m.CreatedIssues, nil
}

func (m *ConfigurableMockGitHubClient) ListDiscussions(ctx context.Context) ([]types.Discussion, error) {
	// For testing, return created discussions
	return m.CreatedDiscussions, nil
}

func (m *ConfigurableMockGitHubClient) ListPRs(ctx context.Context) ([]types.PullRequest, error) {
	// For testing, return created PRs
	return m.CreatedPRs, nil
}

// Deletion operations for cleanup
func (m *ConfigurableMockGitHubClient) DeleteIssue(ctx context.Context, nodeID string) error {
	// For testing, just remove from created issues if found
	for i, issue := range m.CreatedIssues {
		if issue.NodeID == nodeID {
			m.CreatedIssues = append(m.CreatedIssues[:i], m.CreatedIssues[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *ConfigurableMockGitHubClient) DeleteDiscussion(ctx context.Context, nodeID string) error {
	// For testing, just remove from created discussions if found
	for i, discussion := range m.CreatedDiscussions {
		if discussion.NodeID == nodeID {
			m.CreatedDiscussions = append(m.CreatedDiscussions[:i], m.CreatedDiscussions[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *ConfigurableMockGitHubClient) DeletePR(ctx context.Context, nodeID string) error {
	// For testing, just remove from created PRs if found
	for i, pullRequest := range m.CreatedPRs {
		if pullRequest.NodeID == nodeID {
			m.CreatedPRs = append(m.CreatedPRs[:i], m.CreatedPRs[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *ConfigurableMockGitHubClient) DeleteLabel(ctx context.Context, name string) error {
	// For testing, just remove from existing labels
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

// ProjectV2 operations for testing compatibility

func (m *ConfigurableMockGitHubClient) CreateProjectV2(ctx context.Context, config types.ProjectV2Configuration) (*types.ProjectV2, error) {
	if m.Config.FailProjectCreation {
		return nil, errors.ProjectError("create_project", "mock project creation failure", fmt.Errorf("mock error"))
	}

	// Return a mock project for testing
	return &types.ProjectV2{
		NodeID:      "test-project-node-id",
		ID:          "test-project-id",
		Number:      1,
		Title:       config.Title,
		Description: config.Description,
		Visibility:  config.Visibility,
		URL:         "https://github.com/users/test/projects/1",
	}, nil
}

func (m *ConfigurableMockGitHubClient) AddItemToProjectV2(ctx context.Context, projectID, itemNodeID string) error {
	if m.Config.FailProjectItemAddition {
		return errors.ProjectError("add_item_to_project", "mock project item addition failure", fmt.Errorf("mock error"))
	}

	// For testing, just return success
	return nil
}

func (m *ConfigurableMockGitHubClient) GetProjectV2(ctx context.Context, projectID string) (*types.ProjectV2, error) {
	if m.Config.FailProjectRetrieval {
		return nil, errors.ProjectError("get_project", "mock project retrieval failure", fmt.Errorf("mock error"))
	}

	// Return a mock project for testing
	return &types.ProjectV2{
		NodeID:      projectID,
		ID:          projectID,
		Number:      1,
		Title:       "Test Project",
		Description: "Test project description",
		Visibility:  "private",
		URL:         "https://github.com/users/test/projects/1",
	}, nil
}

// ConfigureProjectV2Fields mock implementation for project field configuration
func (m *ConfigurableMockGitHubClient) ConfigureProjectV2Fields(ctx context.Context, projectID string, fields []types.ProjectV2Field) error {
	if m.Config.FailProjectFieldConfiguration {
		return errors.ProjectError("configure_project_fields", "mock project field configuration failure", fmt.Errorf("mock error"))
	}

	// For testing, just return success
	return nil
}

// UpdateProjectV2Description mock implementation for project description updates
func (m *ConfigurableMockGitHubClient) UpdateProjectV2Description(ctx context.Context, projectID, description string) error {
	if m.Config.FailProjectDescriptionUpdate {
		return errors.ProjectError("update_project_description", "mock project description update failure", fmt.Errorf("mock error"))
	}

	// For testing, just return success
	return nil
}

// Helper functions to create common mock configurations

// NewSuccessfulMockGitHubClient creates a mock that succeeds for all operations
func NewSuccessfulMockGitHubClient(existingLabels ...string) *ConfigurableMockGitHubClient {
	return &ConfigurableMockGitHubClient{
		Config: MockConfig{
			ExistingLabels: testutil.Factory.CreateLabelMap(existingLabels...),
		},
		CreatedIssues:      testutil.EmptyCollections.Issues,
		CreatedDiscussions: testutil.EmptyCollections.Discussions,
		CreatedPRs:         testutil.EmptyCollections.PRs,
		CreatedLabels:      testutil.EmptyCollections.Labels,
	}
}

// NewFailingMockGitHubClient creates a mock that fails for specified operations
func NewFailingMockGitHubClient(config MockConfig) *ConfigurableMockGitHubClient {
	if config.ExistingLabels == nil {
		config.ExistingLabels = testutil.Factory.CreateLabelMap()
	}

	return &ConfigurableMockGitHubClient{
		Config:             config,
		CreatedIssues:      testutil.EmptyCollections.Issues,
		CreatedDiscussions: testutil.EmptyCollections.Discussions,
		CreatedPRs:         testutil.EmptyCollections.PRs,
		CreatedLabels:      testutil.EmptyCollections.Labels,
	}
}
