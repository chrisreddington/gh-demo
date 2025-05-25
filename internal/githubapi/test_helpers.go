package githubapi

import (
	"context"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/testutil"
)

// MockResponse represents different types of GraphQL responses that can be configured
type MockResponse struct {
	// Repository responses
	RepositoryID string
	Labels       []string
	Categories   []MockCategory

	// Mutation responses
	LabelID          string
	IssueID          string
	IssueNumber      int
	PRID             string
	PRNumber         int
	DiscussionID     string
	DiscussionNumber int
	UserID           string

	// Error simulation
	ShouldError  bool
	ErrorMessage string
}

type MockCategory struct {
	ID   string
	Name string
}

// ConfigurableMockGraphQLClient provides a configurable mock that can handle common GraphQL operations
type ConfigurableMockGraphQLClient struct {
	Responses map[string]*MockResponse
	DoFunc    func(context.Context, string, map[string]interface{}, interface{}) error
}

func (m *ConfigurableMockGraphQLClient) Do(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	// If custom DoFunc is provided, use it (for backward compatibility)
	if m.DoFunc != nil {
		return m.DoFunc(ctx, query, variables, response)
	}

	// Otherwise use the configurable response system
	return m.handleQuery(query, variables, response)
}

func (m *ConfigurableMockGraphQLClient) handleQuery(query string, variables map[string]interface{}, response interface{}) error {
	// Handle GetRepositoryId query
	if strings.Contains(query, "GetRepositoryId") {
		resp := response.(*struct {
			Repository struct {
				ID string `json:"id"`
			} `json:"repository"`
		})
		if mockResp, exists := m.Responses["repository"]; exists {
			if mockResp.ShouldError {
				return testutil.NewMockError(mockResp.ErrorMessage)
			}
			resp.Repository.ID = mockResp.RepositoryID
		} else {
			resp.Repository.ID = testutil.DefaultValues.RepositoryID
		}
		return nil
	}

	// Handle ListLabels query
	if strings.Contains(query, "labels") && strings.Contains(query, "nodes") {
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
		if mockResp, exists := m.Responses["labels"]; exists {
			if mockResp.ShouldError {
				return testutil.NewMockError(mockResp.ErrorMessage)
			}
			for _, label := range mockResp.Labels {
				resp.Repository.Labels.Nodes = append(resp.Repository.Labels.Nodes, struct {
					Name string `json:"name"`
				}{Name: label})
			}
		}
		return nil
	}

	// Handle createLabel mutation
	if strings.Contains(query, "createLabel") {
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
		if mockResp, exists := m.Responses["createLabel"]; exists {
			if mockResp.ShouldError {
				return testutil.NewMockError(mockResp.ErrorMessage)
			}
			resp.CreateLabel.Label.ID = mockResp.LabelID
		} else {
			resp.CreateLabel.Label.ID = testutil.DefaultValues.LabelID
		}
		// Copy label data from variables
		if input, ok := variables["input"].(map[string]interface{}); ok {
			if name, ok := input["name"].(string); ok {
				resp.CreateLabel.Label.Name = name
			}
			if color, ok := input["color"].(string); ok {
				resp.CreateLabel.Label.Color = color
			}
			if desc, ok := input["description"].(string); ok {
				resp.CreateLabel.Label.Description = desc
			}
		}
		return nil
	}

	// Handle createIssue mutation
	if strings.Contains(query, "createIssue") {
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
		if mockResp, exists := m.Responses["createIssue"]; exists {
			if mockResp.ShouldError {
				return testutil.NewMockError(mockResp.ErrorMessage)
			}
			resp.CreateIssue.Issue.ID = mockResp.IssueID
			resp.CreateIssue.Issue.Number = mockResp.IssueNumber
		} else {
			resp.CreateIssue.Issue.ID = testutil.DefaultValues.IssueID
			resp.CreateIssue.Issue.Number = testutil.DefaultValues.IssueNumber
		}
		resp.CreateIssue.Issue.Title = "Test Issue"
		resp.CreateIssue.Issue.URL = "https://github.com/owner/repo/issues/1"
		return nil
	}

	// Handle createPullRequest mutation
	if strings.Contains(query, "createPullRequest") {
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
		if mockResp, exists := m.Responses["createPR"]; exists {
			if mockResp.ShouldError {
				return testutil.NewMockError(mockResp.ErrorMessage)
			}
			resp.CreatePullRequest.PullRequest.ID = mockResp.PRID
			resp.CreatePullRequest.PullRequest.Number = mockResp.PRNumber
		} else {
			resp.CreatePullRequest.PullRequest.ID = testutil.DefaultValues.PRID
			resp.CreatePullRequest.PullRequest.Number = testutil.DefaultValues.PRNumber
		}
		resp.CreatePullRequest.PullRequest.Title = "Test PR"
		resp.CreatePullRequest.PullRequest.URL = "https://github.com/owner/repo/pull/1"
		return nil
	}

	// Handle discussionCategories query
	if strings.Contains(query, "discussionCategories") {
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
		if mockResp, exists := m.Responses["discussionCategories"]; exists {
			if mockResp.ShouldError {
				return testutil.NewMockError(mockResp.ErrorMessage)
			}
			resp.Repository.ID = mockResp.RepositoryID
			for _, cat := range mockResp.Categories {
				resp.Repository.Categories.Nodes = append(resp.Repository.Categories.Nodes, struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}{ID: cat.ID, Name: cat.Name})
			}
		} else {
			resp.Repository.ID = testutil.DefaultValues.RepositoryID
			resp.Repository.Categories.Nodes = []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				{ID: "default-cat-id", Name: "General"},
			}
		}
		return nil
	}

	// Handle createDiscussion mutation
	if strings.Contains(query, "createDiscussion") {
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
		if mockResp, exists := m.Responses["createDiscussion"]; exists {
			if mockResp.ShouldError {
				return testutil.NewMockError(mockResp.ErrorMessage)
			}
			resp.CreateDiscussion.Discussion.ID = mockResp.DiscussionID
			resp.CreateDiscussion.Discussion.Number = mockResp.DiscussionNumber
		} else {
			resp.CreateDiscussion.Discussion.ID = testutil.DefaultValues.DiscussionID
			resp.CreateDiscussion.Discussion.Number = testutil.DefaultValues.DiscussionNumber
		}
		resp.CreateDiscussion.Discussion.Title = "Test Discussion"
		resp.CreateDiscussion.Discussion.URL = "https://github.com/owner/repo/discussions/1"
		return nil
	}

	// Handle GetLabelId, GetUserId, and other helper queries with default responses
	if strings.Contains(query, "GetLabelId") {
		resp := response.(*struct {
			Repository struct {
				Label struct {
					ID string `json:"id"`
				} `json:"label"`
			} `json:"repository"`
		})
		resp.Repository.Label.ID = testutil.DefaultValues.LabelID
		return nil
	}

	if strings.Contains(query, "GetUserId") {
		resp := response.(*struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		})
		if mockResp, exists := m.Responses["user"]; exists {
			resp.User.ID = mockResp.UserID
		} else {
			resp.User.ID = testutil.DefaultValues.UserID
		}
		return nil
	}

	// Handle addLabelsToLabelable and addAssigneesToAssignable mutations
	if strings.Contains(query, "addLabelsToLabelable") {
		resp := response.(*struct {
			AddLabelsToLabelable struct {
				ClientMutationID string `json:"clientMutationId"`
			} `json:"addLabelsToLabelable"`
		})
		resp.AddLabelsToLabelable.ClientMutationID = "mutation-id-123"
		return nil
	}

	if strings.Contains(query, "addAssigneesToAssignable") {
		resp := response.(*struct {
			AddAssigneesToAssignable struct {
				ClientMutationID string `json:"clientMutationId"`
			} `json:"addAssigneesToAssignable"`
		})
		resp.AddAssigneesToAssignable.ClientMutationID = "mutation-id-456"
		return nil
	}

	// Default: return nil for unhandled queries
	return nil
}

// Helper functions to create common mock configurations

// NewDefaultMockGraphQL creates a mock with sensible defaults for successful operations
func NewDefaultMockGraphQL() *ConfigurableMockGraphQLClient {
	return &ConfigurableMockGraphQLClient{
		Responses: map[string]*MockResponse{
			"repository": {
				RepositoryID: testutil.DefaultValues.RepositoryID,
			},
			"labels": {
				Labels: []string{"bug", "enhancement", "documentation"},
			},
			"createLabel": {
				LabelID: testutil.DefaultValues.LabelID,
			},
			"createIssue": {
				IssueID:     testutil.DefaultValues.IssueID,
				IssueNumber: testutil.DefaultValues.IssueNumber,
			},
			"createPR": {
				PRID:     testutil.DefaultValues.PRID,
				PRNumber: testutil.DefaultValues.PRNumber,
			},
			"discussionCategories": {
				RepositoryID: testutil.DefaultValues.RepositoryID,
				Categories: []MockCategory{
					{ID: "cat-id-123", Name: "General"},
					{ID: "cat-id-456", Name: "Q&A"},
				},
			},
			"createDiscussion": {
				DiscussionID:     testutil.DefaultValues.DiscussionID,
				DiscussionNumber: testutil.DefaultValues.DiscussionNumber,
			},
			"user": {
				UserID: testutil.DefaultValues.UserID,
			},
		},
	}
}

// NewErrorMockGraphQL creates a mock that returns errors for specified operations
func NewErrorMockGraphQL(errorOperations map[string]string) *ConfigurableMockGraphQLClient {
	responses := map[string]*MockResponse{}

	for op, errMsg := range errorOperations {
		responses[op] = &MockResponse{
			ShouldError:  true,
			ErrorMessage: errMsg,
		}
	}

	return &ConfigurableMockGraphQLClient{Responses: responses}
}

// CreateTestClient creates a GHClient with the provided mock GraphQL client
func CreateTestClient(mockGQL GraphQLClient) *GHClient {
	return &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: mockGQL,
	}
}
