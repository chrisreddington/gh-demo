package githubapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chrisreddington/gh-demo/internal/types"
)

// MockGraphQLClient implements the GraphQLClient interface for testing
type MockGraphQLClient struct {
	DoFunc func(context.Context, string, map[string]interface{}, interface{}) error
}

func (m *MockGraphQLClient) Do(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	if m.DoFunc != nil {
		return m.DoFunc(ctx, query, variables, response)
	}
	return nil
}

// Tests for GHClient
func TestNewGHClientWithClients(t *testing.T) {
	mockGQL := &MockGraphQLClient{}

	client, err := NewGHClientWithClients("testowner", "testrepo", mockGQL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	if client.Owner != "testowner" {
		t.Errorf("Expected owner to be 'testowner', got '%s'", client.Owner)
	}
	if client.Repo != "testrepo" {
		t.Errorf("Expected repo to be 'testrepo', got '%s'", client.Repo)
	}
}

func TestNewGHClientWithClients_ValidationErrors(t *testing.T) {
	mockGQL := &MockGraphQLClient{}

	tests := []struct {
		name        string
		owner       string
		repo        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty owner",
			owner:       "",
			repo:        "testrepo",
			expectError: true,
			errorMsg:    "owner cannot be empty",
		},
		{
			name:        "whitespace only owner",
			owner:       "   ",
			repo:        "testrepo",
			expectError: true,
			errorMsg:    "owner cannot be empty",
		},
		{
			name:        "empty repo",
			owner:       "testowner",
			repo:        "",
			expectError: true,
			errorMsg:    "repo cannot be empty",
		},
		{
			name:        "whitespace only repo",
			owner:       "testowner",
			repo:        "   ",
			expectError: true,
			errorMsg:    "repo cannot be empty",
		},
		{
			name:        "valid trimmed parameters",
			owner:       "  testowner  ",
			repo:        "  testrepo  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGHClientWithClients(tt.owner, tt.repo, mockGQL)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if client == nil {
					t.Error("Expected client to be created")
					return
				}
				// Check that parameters are properly trimmed
				if client.Owner != strings.TrimSpace(tt.owner) {
					t.Errorf("Expected owner to be trimmed to '%s', got '%s'", strings.TrimSpace(tt.owner), client.Owner)
				}
				if client.Repo != strings.TrimSpace(tt.repo) {
					t.Errorf("Expected repo to be trimmed to '%s', got '%s'", strings.TrimSpace(tt.repo), client.Repo)
				}
			}
		})
	}
}

func TestGHClientWithMockClients(t *testing.T) {
	client := CreateTestClient(NewDefaultMockGraphQL())

	// Test CreateLabel
	err := client.CreateLabel(context.Background(), types.Label{
		Name:        "test-label",
		Description: "A test label",
		Color:       "ff0000",
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test CreateIssue
	err = client.CreateIssue(context.Background(), types.Issue{
		Title:     "Test Issue",
		Body:      "This is a test issue",
		Labels:    []string{"bug"},
		Assignees: []string{"testuser"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test CreatePR
	err = client.CreatePR(context.Background(), types.PullRequest{
		Title:     "Test PR",
		Body:      "This is a test PR",
		Head:      "feature-branch",
		Base:      "main",
		Labels:    []string{"enhancement"},
		Assignees: []string{"testuser"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// GraphQL tests
func TestListLabels(t *testing.T) {
	client := CreateTestClient(NewDefaultMockGraphQL())

	labels, err := client.ListLabels(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}
	expected := []string{"bug", "enhancement", "documentation"}
	for i, label := range labels {
		if label != expected[i] {
			t.Errorf("Expected label %s at position %d, got %s", expected[i], i, label)
		}
	}
}

func TestCreateDiscussion(t *testing.T) {
	client := CreateTestClient(NewDefaultMockGraphQL())

	err := client.CreateDiscussion(context.Background(), types.Discussion{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreateDiscussion_CategoryNotFound(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			if strings.Contains(query, "discussionCategories") {
				// Repository info query with different categories
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
					{ID: "cat-id-123", Name: "General"},
					{ID: "cat-id-456", Name: "Q&A"},
				}
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreateDiscussion(context.Background(), types.Discussion{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "NonExistent",
	})
	if err == nil {
		t.Error("Expected an error for non-existent category")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error to contain 'not found', got: %v", err)
	}
}

func TestCreateDiscussion_GraphQLError(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			return fmt.Errorf("GraphQL error: network timeout")
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreateDiscussion(context.Background(), types.Discussion{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
	})
	if err == nil {
		t.Error("Expected an error from GraphQL client")
	}
}

func TestGHClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupClient   func() *GHClient
		testFunc      func(*GHClient) error
		expectError   bool
		errorContains string
	}{
		{
			name: "ListLabels GraphQL error",
			setupClient: func() *GHClient {
				return CreateTestClient(NewErrorMockGraphQL(map[string]string{
					"labels": "GraphQL error: unauthorized",
				}))
			},
			testFunc: func(client *GHClient) error {
				_, err := client.ListLabels(context.Background())
				return err
			},
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			name: "ListLabels nil client",
			setupClient: func() *GHClient {
				return &GHClient{
					Owner:     "testowner",
					Repo:      "testrepo",
					gqlClient: nil,
				}
			},
			testFunc: func(client *GHClient) error {
				_, err := client.ListLabels(context.Background())
				return err
			},
			expectError:   true,
			errorContains: "GraphQL client is not initialized",
		},
		{
			name: "CreateDiscussion nil client",
			setupClient: func() *GHClient {
				return &GHClient{
					Owner:     "testowner",
					Repo:      "testrepo",
					gqlClient: nil,
				}
			},
			testFunc: func(client *GHClient) error {
				return client.CreateDiscussion(context.Background(), types.Discussion{
					Title:    "Test Discussion",
					Body:     "This is a test discussion",
					Category: "General",
				})
			},
			expectError:   true,
			errorContains: "not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			err := tt.testFunc(client)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected an error for %s", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
			}
		})
	}
}

func TestCreatePR(t *testing.T) {
	client := CreateTestClient(NewDefaultMockGraphQL())

	err := client.CreatePR(context.Background(), types.PullRequest{
		Title: "Test PR",
		Body:  "This is a test PR",
		Head:  "feature-branch",
		Base:  "main",
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestGHClient_GraphQLOperations provides comprehensive testing of GraphQL operations with table-driven approach
func TestGHClient_GraphQLOperations(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		setupMock   func() *ConfigurableMockGraphQLClient
		testFunc    func(*GHClient) error
		expectError bool
	}{
		{
			name:      "CreateLabel success",
			operation: "createLabel",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			testFunc: func(client *GHClient) error {
				return client.CreateLabel(context.Background(), types.Label{
					Name:        "feature",
					Description: "New feature",
					Color:       "00ff00",
				})
			},
			expectError: false,
		},
		{
			name:      "CreateIssue success",
			operation: "createIssue",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			testFunc: func(client *GHClient) error {
				return client.CreateIssue(context.Background(), types.Issue{
					Title:     "Test Issue",
					Body:      "Test description",
					Labels:    []string{"bug"},
					Assignees: []string{"testuser"},
				})
			},
			expectError: false,
		},
		{
			name:      "CreatePR success",
			operation: "createPR",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			testFunc: func(client *GHClient) error {
				return client.CreatePR(context.Background(), types.PullRequest{
					Title:     "Test PR",
					Body:      "Test description",
					Head:      "feature-branch",
					Base:      "main",
					Labels:    []string{"enhancement"},
					Assignees: []string{"testuser"},
				})
			},
			expectError: false,
		},
		{
			name:      "CreateDiscussion success",
			operation: "createDiscussion",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			testFunc: func(client *GHClient) error {
				return client.CreateDiscussion(context.Background(), types.Discussion{
					Title:    "Test Discussion",
					Body:     "Test body",
					Category: "General",
					Labels:   []string{"question"},
				})
			},
			expectError: false,
		},
		{
			name:      "ListLabels success",
			operation: "listLabels",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			testFunc: func(client *GHClient) error {
				labels, err := client.ListLabels(context.Background())
				if err != nil {
					return err
				}
				if len(labels) != 3 {
					return fmt.Errorf("expected 3 labels, got %d", len(labels))
				}
				return nil
			},
			expectError: false,
		},
		{
			name:      "CreateLabel error",
			operation: "createLabel",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewErrorMockGraphQL(map[string]string{
					"createLabel": "label creation failed",
				})
			},
			testFunc: func(client *GHClient) error {
				return client.CreateLabel(context.Background(), types.Label{
					Name: "test",
				})
			},
			expectError: true,
		},
		{
			name:      "ListLabels error",
			operation: "listLabels",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewErrorMockGraphQL(map[string]string{
					"labels": "access denied",
				})
			},
			testFunc: func(client *GHClient) error {
				_, err := client.ListLabels(context.Background())
				return err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()
			client := CreateTestClient(mockClient)

			err := tt.testFunc(client)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s but got none", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
			}
		})
	}
}

func TestSetLogger(t *testing.T) {
	mockGQL := &MockGraphQLClient{}

	client, err := NewGHClientWithClients("testowner", "testrepo", mockGQL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	logger := &MockLogger{}
	client.SetLogger(logger)

	if client.logger != logger {
		t.Error("Expected logger to be set")
	}
}

type MockLogger struct {
	lastMessage string
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.lastMessage = fmt.Sprintf(format, args...)
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.lastMessage = fmt.Sprintf(format, args...)
}

func TestDebugLog(t *testing.T) {
	mockGQL := &MockGraphQLClient{}

	client, err := NewGHClientWithClients("testowner", "testrepo", mockGQL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	mockLogger := &MockLogger{}
	client.SetLogger(mockLogger)

	client.debugLog("test message: %s", "value")

	if mockLogger.lastMessage != "test message: value" {
		t.Errorf("Expected 'test message: value', got '%s'", mockLogger.lastMessage)
	}
}

// TestCreateDiscussionWithLabels tests the addLabelToDiscussion function through CreateDiscussion
func TestCreateDiscussionWithLabels(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			if strings.Contains(query, "discussionCategories") {
				// Repository info query
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
					{ID: "cat-id-123", Name: "General"},
				}
			} else if strings.Contains(query, "createDiscussion") {
				// Create discussion mutation
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
				resp.CreateDiscussion.Discussion.ID = "disc-id-123"
				resp.CreateDiscussion.Discussion.Number = 1
				resp.CreateDiscussion.Discussion.Title = "Test Discussion"
				resp.CreateDiscussion.Discussion.URL = "https://github.com/testowner/testrepo/discussions/1"
			} else if strings.Contains(query, "label(name:") {
				// Label query for addLabelToDiscussion
				resp := response.(*struct {
					Repository struct {
						Label struct {
							ID string `json:"id"`
						} `json:"label"`
					} `json:"repository"`
				})
				resp.Repository.Label.ID = "label-id-123"
			} else if strings.Contains(query, "addLabelsToLabelable") {
				// Add label mutation
				resp := response.(*struct {
					AddLabelsToLabelable struct {
						ClientMutationID string `json:"clientMutationId"`
					} `json:"addLabelsToLabelable"`
				})
				resp.AddLabelsToLabelable.ClientMutationID = "mutation-id-123"
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreateDiscussion(context.Background(), types.Discussion{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
		Labels:   []string{"bug", "enhancement"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestAddLabelToDiscussion_LabelNotFound tests error handling in addLabelToDiscussion
func TestAddLabelToDiscussion_LabelNotFound(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			if strings.Contains(query, "discussionCategories") {
				// Repository info query
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
					{ID: "cat-id-123", Name: "General"},
				}
			} else if strings.Contains(query, "createDiscussion") {
				// Create discussion mutation
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
				resp.CreateDiscussion.Discussion.ID = "disc-id-123"
				resp.CreateDiscussion.Discussion.Number = 1
				resp.CreateDiscussion.Discussion.Title = "Test Discussion"
				resp.CreateDiscussion.Discussion.URL = "https://github.com/testowner/testrepo/discussions/1"
			} else if strings.Contains(query, "label(name:") {
				// Label query that returns empty (label not found)
				resp := response.(*struct {
					Repository struct {
						Label struct {
							ID string `json:"id"`
						} `json:"label"`
					} `json:"repository"`
				})
				resp.Repository.Label.ID = "" // Empty ID means label not found
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	// This should still succeed, but the label addition will fail silently
	err := client.CreateDiscussion(context.Background(), types.Discussion{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
		Labels:   []string{"nonexistent-label"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestAddLabelToDiscussion_GraphQLError tests GraphQL error handling
func TestAddLabelToDiscussion_GraphQLError(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			if strings.Contains(query, "discussionCategories") {
				// Repository info query
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
					{ID: "cat-id-123", Name: "General"},
				}
			} else if strings.Contains(query, "createDiscussion") {
				// Create discussion mutation
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
				resp.CreateDiscussion.Discussion.ID = "disc-id-123"
				resp.CreateDiscussion.Discussion.Number = 1
				resp.CreateDiscussion.Discussion.Title = "Test Discussion"
				resp.CreateDiscussion.Discussion.URL = "https://github.com/testowner/testrepo/discussions/1"
			} else if strings.Contains(query, "label(name:") {
				// Return error for label query
				return fmt.Errorf("GraphQL error: label query failed")
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	// This should still succeed overall, but label addition will fail
	err := client.CreateDiscussion(context.Background(), types.Discussion{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
		Labels:   []string{"test-label"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestCreatePR_ValidationErrors tests CreatePR validation error paths
func TestCreatePR_ValidationErrors(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			// This should not be called for validation errors
			t.Error("GraphQL client should not be called for validation errors")
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	// Test empty head branch
	err := client.CreatePR(context.Background(), types.PullRequest{
		Title: "Test PR",
		Body:  "Test body",
		Head:  "", // Empty head should cause error
		Base:  "main",
	})
	if err == nil {
		t.Error("Expected error for empty head branch")
	}
	if !strings.Contains(err.Error(), "head branch cannot be empty") {
		t.Errorf("Expected 'head branch cannot be empty' error, got: %v", err)
	}

	// Test empty base branch
	err = client.CreatePR(context.Background(), types.PullRequest{
		Title: "Test PR",
		Body:  "Test body",
		Head:  "feature",
		Base:  "", // Empty base should cause error
	})
	if err == nil {
		t.Error("Expected error for empty base branch")
	}
	if !strings.Contains(err.Error(), "base branch cannot be empty") {
		t.Errorf("Expected 'base branch cannot be empty' error, got: %v", err)
	}

	// Test head and base are the same
	err = client.CreatePR(context.Background(), types.PullRequest{
		Title: "Test PR",
		Body:  "Test body",
		Head:  "main",
		Base:  "main", // Same as head should cause error
	})
	if err == nil {
		t.Error("Expected error for same head and base branches")
	}
	if !strings.Contains(err.Error(), "head and base branches cannot be the same") {
		t.Errorf("Expected 'head and base branches cannot be the same' error, got: %v", err)
	}
}

// TestCreatePR_WithLabelsAndAssignees tests CreatePR with labels and assignees
func TestCreatePR_WithLabelsAndAssignees(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			if strings.Contains(query, "GetRepositoryId") {
				// Repository ID query
				resp := response.(*struct {
					Repository struct {
						ID string `json:"id"`
					} `json:"repository"`
				})
				resp.Repository.ID = "repo-id-123"
				return nil
			} else if strings.Contains(query, "GetLabelId") {
				// Label ID query
				resp := response.(*struct {
					Repository struct {
						Label struct {
							ID string `json:"id"`
						} `json:"label"`
					} `json:"repository"`
				})
				resp.Repository.Label.ID = "label-id-456"
				return nil
			} else if strings.Contains(query, "GetUserId") {
				// User ID query
				resp := response.(*struct {
					User struct {
						ID string `json:"id"`
					} `json:"user"`
				})
				resp.User.ID = "user-id-789"
				return nil
			} else if strings.Contains(query, "createPullRequest") {
				// Create pull request mutation
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
				resp.CreatePullRequest.PullRequest.URL = "https://github.com/test/test/pull/1"
				return nil
			} else if strings.Contains(query, "addLabelsToLabelable") || strings.Contains(query, "addAssigneesToAssignable") {
				// Add labels/assignees mutations
				return nil
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreatePR(context.Background(), types.PullRequest{
		Title:     "Test PR",
		Body:      "Test body",
		Head:      "feature",
		Base:      "main",
		Labels:    []string{"bug", "enhancement"},
		Assignees: []string{"testuser"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestCreatePR_LabelsAssigneesFailure tests CreatePR when labels/assignees update fails
func TestCreatePR_LabelsAssigneesFailure(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			if strings.Contains(query, "GetRepositoryId") {
				// Repository ID query
				resp := response.(*struct {
					Repository struct {
						ID string `json:"id"`
					} `json:"repository"`
				})
				resp.Repository.ID = "repo-id-123"
				return nil
			} else if strings.Contains(query, "GetLabelId") {
				// Label ID query
				resp := response.(*struct {
					Repository struct {
						Label struct {
							ID string `json:"id"`
						} `json:"label"`
					} `json:"repository"`
				})
				resp.Repository.Label.ID = "label-id-456"
				return nil
			} else if strings.Contains(query, "GetUserId") {
				// User ID query
				resp := response.(*struct {
					User struct {
						ID string `json:"id"`
					} `json:"user"`
				})
				resp.User.ID = "user-id-789"
				return nil
			} else if strings.Contains(query, "createPullRequest") {
				// Create pull request mutation succeeds
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
				resp.CreatePullRequest.PullRequest.URL = "https://github.com/test/test/pull/1"
				return nil
			} else if strings.Contains(query, "addLabelsToLabelable") || strings.Contains(query, "addAssigneesToAssignable") {
				// Add labels/assignees mutations fail
				return fmt.Errorf("failed to update labels/assignees")
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreatePR(context.Background(), types.PullRequest{
		Title:     "Test PR",
		Body:      "Test body",
		Head:      "feature",
		Base:      "main",
		Labels:    []string{"bug"},
		Assignees: []string{"testuser"},
	})
	if err == nil {
		t.Error("Expected error when labels/assignees update fails")
	}
	if !strings.Contains(err.Error(), "failed to add labels/assignees") {
		t.Errorf("Expected 'failed to add labels/assignees' error, got: %v", err)
	}
}

// TestCreatePR_RequestFailure tests CreatePR when the initial request fails
func TestCreatePR_RequestFailure(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			if strings.Contains(query, "GetRepositoryId") {
				// Repository ID query
				resp := response.(*struct {
					Repository struct {
						ID string `json:"id"`
					} `json:"repository"`
				})
				resp.Repository.ID = "repo-id-123"
				return nil
			} else if strings.Contains(query, "createPullRequest") {
				// Create pull request mutation fails
				return fmt.Errorf("server error")
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreatePR(context.Background(), types.PullRequest{
		Title: "Test PR",
		Body:  "Test body",
		Head:  "feature",
		Base:  "main",
	})
	if err == nil {
		t.Error("Expected error when request fails")
	}
	if !strings.Contains(err.Error(), "failed to create pull request") {
		t.Errorf("Expected 'failed to create pull request' error, got: %v", err)
	}
}

// TestNewGHClient_Integration tests the real GitHub client creation
// This test requires authentication and should be skipped in CI without credentials
func TestNewGHClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test will only pass if GitHub credentials are available
	client, err := NewGHClient("testowner", "testrepo")
	if err != nil {
		// If error contains authentication message, skip the test
		if strings.Contains(err.Error(), "authentication token not found") {
			t.Skip("Skipping integration test: GitHub authentication not available")
		}
		t.Fatalf("Failed to create client: %v", err)
	}
	if client.Owner != "testowner" {
		t.Errorf("Expected owner to be 'testowner', got '%s'", client.Owner)
	}
	if client.Repo != "testrepo" {
		t.Errorf("Expected repo to be 'testrepo', got '%s'", client.Repo)
	}
}

// TestCreateIssue_ContextTimeout tests that context timeout is handled correctly
func TestCreateIssue_ContextTimeout(t *testing.T) {
	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(2 * time.Millisecond)

	gqlClient := &MockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			// Check if context is already cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// This should not be reached due to timeout
				return nil
			}
		},
	}

	client, err := NewGHClientWithClients("testowner", "testrepo", gqlClient)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.CreateIssue(ctx, types.Issue{
		Title: "Test Issue",
		Body:  "Test body",
	})

	if err == nil {
		t.Error("Expected context timeout error, got nil")
		return
	}

	// Check if the error message is user-friendly for context timeout
	errStr := err.Error()
	if !strings.Contains(errStr, "timed out") && !strings.Contains(errStr, "cancelled") {
		t.Errorf("Expected user-friendly timeout message, got: %v", err)
	}
}

// TestGraphQLClientWrapper_ContextCancellation tests that long-running operations can be cancelled
func TestGraphQLClientWrapper_ContextCancellation(t *testing.T) {
	// Create a mock underlying client that doesn't support context (like go-gh)
	slowUnderlyingClient := &slowClient{}

	// Wrap it with our graphQLClientWrapper
	wrapper := &graphQLClientWrapper{client: slowUnderlyingClient}

	// Create context that will cancel after 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := wrapper.Do(ctx, "test query", nil, nil)
	duration := time.Since(start)

	// Should return an error due to context timeout
	if err == nil {
		t.Error("Expected context timeout error, got nil")
		return
	}

	// Should complete quickly (around 100ms), not after 2 seconds
	// This will fail with current implementation since it can't cancel the underlying call
	if duration > 500*time.Millisecond {
		t.Errorf("Operation took too long (%v), context cancellation may not be working", duration)
	}

	// Should be a context error
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context error, got: %v", err)
	}
}

// slowClient simulates the go-gh client that doesn't support context
type slowClient struct{}

func (s *slowClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	// Simulate a long operation that can't be cancelled (like go-gh client)
	time.Sleep(2 * time.Second)
	return nil
}
