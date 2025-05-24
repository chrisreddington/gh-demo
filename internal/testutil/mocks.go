package testutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// MockGraphQLClient provides a simplified mock for GraphQL operations
type MockGraphQLClient struct {
	Responses map[string]interface{}
	Errors    map[string]error
}

func (m *MockGraphQLClient) Do(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	// Check for error first
	for queryType, err := range m.Errors {
		if strings.Contains(query, queryType) {
			return err
		}
	}

	// Handle responses using a simplified approach
	if strings.Contains(query, "GetRepositoryId") {
		resp := response.(*struct {
			Repository struct {
				ID string `json:"id"`
			} `json:"repository"`
		})
		resp.Repository.ID = "repo-id-123"
	} else if strings.Contains(query, "createLabel") {
		resp := response.(*struct {
			CreateLabel struct {
				Label struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Color       string `json:"color"`
					Description string `json:"description"`
				} `json:"label"`
			} `json:"createLabel"`
		})
		resp.CreateLabel.Label.ID = "label-id-123"
		resp.CreateLabel.Label.Name = "test-label"
		resp.CreateLabel.Label.Color = "ff0000"
		resp.CreateLabel.Label.Description = "A test label"
	} else if strings.Contains(query, "createIssue") {
		resp := response.(*struct {
			CreateIssue struct {
				Issue struct {
					ID     string `json:"id"`
					Number int    `json:"number"`
					Title  string `json:"title"`
					URL    string `json:"url"`
				} `json:"issue"`
			} `json:"createIssue"`
		})
		resp.CreateIssue.Issue.ID = "issue-id-123"
		resp.CreateIssue.Issue.Number = 1
		resp.CreateIssue.Issue.Title = "Test Issue"
		resp.CreateIssue.Issue.URL = "https://github.com/owner/repo/issues/1"
	} else if strings.Contains(query, "createPullRequest") {
		resp := response.(*struct {
			CreatePullRequest struct {
				PullRequest struct {
					ID     string `json:"id"`
					Number int    `json:"number"`
					Title  string `json:"title"`
					URL    string `json:"url"`
				} `json:"pullRequest"`
			} `json:"createPullRequest"`
		})
		resp.CreatePullRequest.PullRequest.ID = "pr-id-123"
		resp.CreatePullRequest.PullRequest.Number = 1
		resp.CreatePullRequest.PullRequest.Title = "Test PR"
		resp.CreatePullRequest.PullRequest.URL = "https://github.com/owner/repo/pull/1"
	} else if strings.Contains(query, "discussionCategories") {
		// Repository discussion categories query
		resp := response.(*struct {
			Repository struct {
				ID         string `json:"id"`
				Categories struct {
					Nodes []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"discussionCategories"`
			} `json:"repository"`
		})
		resp.Repository.ID = "repo-id-123"
		resp.Repository.Categories.Nodes = []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			{ID: "category-id-1", Name: "General"},
			{ID: "category-id-2", Name: "Q&A"},
		}
	} else if strings.Contains(query, "createDiscussion") {
		resp := response.(*struct {
			CreateDiscussion struct {
				Discussion struct {
					ID     string `json:"id"`
					Number int    `json:"number"`
					Title  string `json:"title"`
					URL    string `json:"url"`
				} `json:"discussion"`
			} `json:"createDiscussion"`
		})
		resp.CreateDiscussion.Discussion.ID = "discussion-id-123"
		resp.CreateDiscussion.Discussion.Number = 1
		resp.CreateDiscussion.Discussion.Title = "Test Discussion"
		resp.CreateDiscussion.Discussion.URL = "https://github.com/owner/repo/discussions/1"
	} else if strings.Contains(query, "labels") {
		// List labels query
		resp := response.(*struct {
			Repository struct {
				Labels struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"labels"`
			} `json:"repository"`
		})
		resp.Repository.Labels.Nodes = []struct {
			Name string `json:"name"`
		}{
			{Name: "bug"},
			{Name: "enhancement"},
			{Name: "documentation"},
		}
	} else if strings.Contains(query, "GetLabelId") {
		// Label ID query for issue/PR creation
		resp := response.(*struct {
			Repository struct {
				Label struct {
					ID string `json:"id"`
				} `json:"label"`
			} `json:"repository"`
		})
		resp.Repository.Label.ID = "label-id-456"
	} else if strings.Contains(query, "GetUserId") {
		// User ID query for assignees
		resp := response.(*struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		})
		resp.User.ID = "user-id-789"
	} else if strings.Contains(query, "addLabelsToLabelable") {
		// Add labels mutation
		resp := response.(*struct {
			AddLabelsToLabelable struct {
				ClientMutationID string `json:"clientMutationId"`
			} `json:"addLabelsToLabelable"`
		})
		resp.AddLabelsToLabelable.ClientMutationID = "mutation-id-123"
	} else if strings.Contains(query, "addAssigneesToAssignable") {
		// Add assignees mutation
		resp := response.(*struct {
			AddAssigneesToAssignable struct {
				ClientMutationID string `json:"clientMutationId"`
			} `json:"addAssigneesToAssignable"`
		})
		resp.AddAssigneesToAssignable.ClientMutationID = "mutation-id-456"
	}

	return nil
}

// MockGitHubClient provides a consolidated mock for GitHub API operations
type MockGitHubClient struct {
	ExistingLabels     map[string]bool
	CreatedLabels      []string
	CreatedIssues      []types.Issue
	CreatedDiscussions []types.Discussion
	CreatedPRs         []types.PullRequest
	FailMethods        map[string]bool // Method name -> should fail
	logger             common.Logger
}

func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		ExistingLabels: make(map[string]bool),
		FailMethods:    make(map[string]bool),
	}
}

func (m *MockGitHubClient) CreateIssue(ctx context.Context, issue types.Issue) error {
	if m.FailMethods["CreateIssue"] {
		return fmt.Errorf("simulated issue creation failure for: %s", issue.Title)
	}
	m.CreatedIssues = append(m.CreatedIssues, issue)
	return nil
}

func (m *MockGitHubClient) CreateDiscussion(ctx context.Context, discussion types.Discussion) error {
	if m.FailMethods["CreateDiscussion"] {
		return fmt.Errorf("simulated discussion creation failure for: %s", discussion.Title)
	}
	m.CreatedDiscussions = append(m.CreatedDiscussions, discussion)
	return nil
}

func (m *MockGitHubClient) CreatePR(ctx context.Context, pullRequest types.PullRequest) error {
	if m.FailMethods["CreatePR"] {
		return fmt.Errorf("simulated PR creation failure for: %s (head: %s, base: %s)", pullRequest.Title, pullRequest.Head, pullRequest.Base)
	}
	m.CreatedPRs = append(m.CreatedPRs, pullRequest)
	return nil
}

func (m *MockGitHubClient) ListLabels(ctx context.Context) ([]string, error) {
	if m.FailMethods["ListLabels"] {
		return nil, fmt.Errorf("simulated list labels failure")
	}
	labels := make([]string, 0, len(m.ExistingLabels))
	for l := range m.ExistingLabels {
		labels = append(labels, l)
	}
	return labels, nil
}

func (m *MockGitHubClient) CreateLabel(ctx context.Context, label types.Label) error {
	if m.FailMethods["CreateLabel"] {
		return fmt.Errorf("simulated label creation failure for: %s", label.Name)
	}
	m.CreatedLabels = append(m.CreatedLabels, label.Name)
	m.ExistingLabels[label.Name] = true
	return nil
}

func (m *MockGitHubClient) SetLogger(logger common.Logger) {
	m.logger = logger
}

// MockLogger provides a simple mock logger for testing
type MockLogger struct {
	LastMessage string
	DebugCalled bool
	InfoCalled  bool
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.DebugCalled = true
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.LastMessage = fmt.Sprintf(format, args...)
	m.InfoCalled = true
}