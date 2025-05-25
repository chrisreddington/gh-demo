package testutil

import (
	"context"
	"strings"
)

// GraphQLMockCategory represents a mock discussion category
type GraphQLMockCategory struct {
	ID   string
	Name string
}

// GraphQLMockResponse represents different types of GraphQL responses that can be configured
type GraphQLMockResponse struct {
	// Repository responses
	RepositoryID string
	Labels       []string
	Categories   []GraphQLMockCategory

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
	ErrorConfig
}

// GraphQLMockClient provides a configurable mock that can handle common GraphQL operations
type GraphQLMockClient struct {
	Responses map[string]*GraphQLMockResponse
	DoFunc    func(context.Context, string, map[string]interface{}, interface{}) error
}

// NewGraphQLMockClient creates a new GraphQL mock client
func NewGraphQLMockClient() *GraphQLMockClient {
	return &GraphQLMockClient{
		Responses: make(map[string]*GraphQLMockResponse),
	}
}

// Do implements the GraphQLClient interface
func (m *GraphQLMockClient) Do(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	// If custom DoFunc is provided, use it (for backward compatibility)
	if m.DoFunc != nil {
		return m.DoFunc(ctx, query, variables, response)
	}

	// Otherwise use the configurable response system
	return m.handleQuery(query, variables, response)
}

// handleQuery processes different GraphQL query patterns
func (m *GraphQLMockClient) handleQuery(query string, variables map[string]interface{}, response interface{}) error {
	// Handle GetRepositoryId query
	if strings.Contains(query, "GetRepositoryId") {
		return m.handleRepositoryQuery(response)
	}

	// Handle ListLabels query
	if strings.Contains(query, "labels") && strings.Contains(query, "nodes") {
		return m.handleLabelsQuery(response)
	}

	// Handle createLabel mutation
	if strings.Contains(query, "createLabel") {
		return m.handleCreateLabelMutation(variables, response)
	}

	// Handle createIssue mutation
	if strings.Contains(query, "createIssue") {
		return m.handleCreateIssueMutation(response)
	}

	// Handle createPullRequest mutation
	if strings.Contains(query, "createPullRequest") {
		return m.handleCreatePRMutation(response)
	}

	// Handle discussionCategories query
	if strings.Contains(query, "discussionCategories") {
		return m.handleDiscussionCategoriesQuery(response)
	}

	// Handle createDiscussion mutation
	if strings.Contains(query, "createDiscussion") {
		return m.handleCreateDiscussionMutation(response)
	}

	// Handle helper queries
	if strings.Contains(query, "GetLabelId") {
		return m.handleGetLabelIdQuery(response)
	}

	if strings.Contains(query, "GetUserId") {
		return m.handleGetUserIdQuery(response)
	}

	// Handle addLabelsToLabelable and addAssigneesToAssignable mutations
	if strings.Contains(query, "addLabelsToLabelable") {
		return m.handleAddLabelsToLabelableMutation(response)
	}

	if strings.Contains(query, "addAssigneesToAssignable") {
		return m.handleAddAssigneesToAssignableMutation(response)
	}

	// Default: return nil for unhandled queries
	return nil
}

func (m *GraphQLMockClient) handleRepositoryQuery(response interface{}) error {
	resp := response.(*struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	})
	if mockResp, exists := m.Responses["repository"]; exists {
		if err := mockResp.GetErrorOrDefault("repository query failed"); err != nil {
			return NewMockError(err.Error())
		}
		resp.Repository.ID = mockResp.RepositoryID
	} else {
		resp.Repository.ID = "default-repo-id"
	}
	return nil
}

func (m *GraphQLMockClient) handleLabelsQuery(response interface{}) error {
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
		if err := mockResp.GetErrorOrDefault("labels query failed"); err != nil {
			return NewMockError(err.Error())
		}
		for _, label := range mockResp.Labels {
			resp.Repository.Labels.Nodes = append(resp.Repository.Labels.Nodes, struct {
				Name string `json:"name"`
			}{Name: label})
		}
	}
	return nil
}

func (m *GraphQLMockClient) handleCreateLabelMutation(variables map[string]interface{}, response interface{}) error {
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
		if err := mockResp.GetErrorOrDefault("create label failed"); err != nil {
			return NewMockError(err.Error())
		}
		resp.CreateLabel.Label.ID = mockResp.LabelID
	} else {
		resp.CreateLabel.Label.ID = "default-label-id"
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

func (m *GraphQLMockClient) handleCreateIssueMutation(response interface{}) error {
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
		if err := mockResp.GetErrorOrDefault("create issue failed"); err != nil {
			return NewMockError(err.Error())
		}
		resp.CreateIssue.Issue.ID = mockResp.IssueID
		resp.CreateIssue.Issue.Number = mockResp.IssueNumber
	} else {
		resp.CreateIssue.Issue.ID = "default-issue-id"
		resp.CreateIssue.Issue.Number = 1
	}
	resp.CreateIssue.Issue.Title = "Test Issue"
	resp.CreateIssue.Issue.URL = "https://github.com/owner/repo/issues/1"
	return nil
}

func (m *GraphQLMockClient) handleCreatePRMutation(response interface{}) error {
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
		if err := mockResp.GetErrorOrDefault("create pull request failed"); err != nil {
			return NewMockError(err.Error())
		}
		resp.CreatePullRequest.PullRequest.ID = mockResp.PRID
		resp.CreatePullRequest.PullRequest.Number = mockResp.PRNumber
	} else {
		resp.CreatePullRequest.PullRequest.ID = "default-pr-id"
		resp.CreatePullRequest.PullRequest.Number = 1
	}
	resp.CreatePullRequest.PullRequest.Title = "Test PR"
	resp.CreatePullRequest.PullRequest.URL = "https://github.com/owner/repo/pull/1"
	return nil
}

func (m *GraphQLMockClient) handleDiscussionCategoriesQuery(response interface{}) error {
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
		if err := mockResp.GetErrorOrDefault("discussion categories query failed"); err != nil {
			return NewMockError(err.Error())
		}
		resp.Repository.ID = mockResp.RepositoryID
		for _, cat := range mockResp.Categories {
			resp.Repository.Categories.Nodes = append(resp.Repository.Categories.Nodes, struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{ID: cat.ID, Name: cat.Name})
		}
	} else {
		resp.Repository.ID = "default-repo-id"
		resp.Repository.Categories.Nodes = []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			{ID: "default-cat-id", Name: "General"},
		}
	}
	return nil
}

func (m *GraphQLMockClient) handleCreateDiscussionMutation(response interface{}) error {
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
		if err := mockResp.GetErrorOrDefault("create discussion failed"); err != nil {
			return NewMockError(err.Error())
		}
		resp.CreateDiscussion.Discussion.ID = mockResp.DiscussionID
		resp.CreateDiscussion.Discussion.Number = mockResp.DiscussionNumber
	} else {
		resp.CreateDiscussion.Discussion.ID = "default-discussion-id"
		resp.CreateDiscussion.Discussion.Number = 1
	}
	resp.CreateDiscussion.Discussion.Title = "Test Discussion"
	resp.CreateDiscussion.Discussion.URL = "https://github.com/owner/repo/discussions/1"
	return nil
}

func (m *GraphQLMockClient) handleGetLabelIdQuery(response interface{}) error {
	resp := response.(*struct {
		Repository struct {
			Label struct {
				ID string `json:"id"`
			} `json:"label"`
		} `json:"repository"`
	})
	resp.Repository.Label.ID = "default-label-id"
	return nil
}

func (m *GraphQLMockClient) handleGetUserIdQuery(response interface{}) error {
	resp := response.(*struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	})
	if mockResp, exists := m.Responses["user"]; exists {
		resp.User.ID = mockResp.UserID
	} else {
		resp.User.ID = "default-user-id"
	}
	return nil
}

func (m *GraphQLMockClient) handleAddLabelsToLabelableMutation(response interface{}) error {
	resp := response.(*struct {
		AddLabelsToLabelable struct {
			ClientMutationID string `json:"clientMutationId"`
		} `json:"addLabelsToLabelable"`
	})
	resp.AddLabelsToLabelable.ClientMutationID = "mutation-id-123"
	return nil
}

func (m *GraphQLMockClient) handleAddAssigneesToAssignableMutation(response interface{}) error {
	resp := response.(*struct {
		AddAssigneesToAssignable struct {
			ClientMutationID string `json:"clientMutationId"`
		} `json:"addAssigneesToAssignable"`
	})
	resp.AddAssigneesToAssignable.ClientMutationID = "mutation-id-456"
	return nil
}

// NewDefaultGraphQLMock creates a GraphQL mock with sensible defaults for successful operations
func NewDefaultGraphQLMock() *GraphQLMockClient {
	client := NewGraphQLMockClient()
	builder := NewResponseBuilder()
	
	client.Responses = map[string]*GraphQLMockResponse{
		"repository": {
			RepositoryID: builder.DefaultRepositoryID,
		},
		"labels": {
			Labels: []string{"bug", "enhancement", "documentation"},
		},
		"createLabel": {
			LabelID: builder.DefaultLabelID,
		},
		"createIssue": {
			IssueID:     builder.DefaultIssueID,
			IssueNumber: builder.DefaultIssueNumber,
		},
		"createPR": {
			PRID:     builder.DefaultPRID,
			PRNumber: builder.DefaultPRNumber,
		},
		"discussionCategories": {
			RepositoryID: builder.DefaultRepositoryID,
			Categories: []GraphQLMockCategory{
				{ID: "cat-id-123", Name: "General"},
				{ID: "cat-id-456", Name: "Q&A"},
			},
		},
		"createDiscussion": {
			DiscussionID:     builder.DefaultDiscussionID,
			DiscussionNumber: builder.DefaultDiscussionNumber,
		},
		"user": {
			UserID: builder.DefaultUserID,
		},
	}
	
	return client
}

// NewErrorGraphQLMock creates a GraphQL mock that returns errors for specified operations
func NewErrorGraphQLMock(errorOperations map[string]string) *GraphQLMockClient {
	client := NewGraphQLMockClient()
	
	for op, errMsg := range errorOperations {
		client.Responses[op] = &GraphQLMockResponse{
			ErrorConfig: ErrorConfig{
				ShouldError:  true,
				ErrorMessage: errMsg,
			},
		}
	}
	
	return client
}