package githubapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chrisreddington/gh-demo/internal/testutil"
	"github.com/chrisreddington/gh-demo/internal/types"
)

// Tests for GHClient
func TestNewGHClientWithClients(t *testing.T) {
	mockGQL := &testutil.SimpleMockGraphQLClient{}

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
	mockGQL := &testutil.SimpleMockGraphQLClient{}

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
	testIssue := testutil.DataFactory.CreateTestIssue("Test Issue")
	testIssue.Labels = []string{"bug"}
	testIssue.Assignees = []string{"testuser"}
	_, err = client.CreateIssue(context.Background(), testIssue)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test CreatePR
	testPR := testutil.DataFactory.CreateTestPR("Test PR")
	testPR.Labels = []string{"enhancement"}
	testPR.Assignees = []string{"testuser"}
	_, err = client.CreatePR(context.Background(), testPR)
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

	_, err := client.CreateDiscussion(context.Background(), types.Discussion{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreateDiscussion_CategoryNotFound(t *testing.T) {
	gqlClient := &testutil.SimpleMockGraphQLClient{
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

	_, err := client.CreateDiscussion(context.Background(), types.Discussion{
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
	gqlClient := &testutil.SimpleMockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			return fmt.Errorf("GraphQL error: network timeout")
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	_, err := client.CreateDiscussion(context.Background(), types.Discussion{
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
				_, err := client.CreateDiscussion(context.Background(), types.Discussion{
					Title:    "Test Discussion",
					Body:     "This is a test discussion",
					Category: "General",
				})
				return err
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

	_, err := client.CreatePR(context.Background(), types.PullRequest{
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
				_, err := client.CreateIssue(context.Background(), types.Issue{
					Title:     "Test Issue",
					Body:      "Test description",
					Labels:    []string{"bug"},
					Assignees: []string{"testuser"},
				})
				return err
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
				_, err := client.CreatePR(context.Background(), types.PullRequest{
					Title:     "Test PR",
					Body:      "Test description",
					Head:      "feature-branch",
					Base:      "main",
					Labels:    []string{"enhancement"},
					Assignees: []string{"testuser"},
				})
				return err
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
				_, err := client.CreateDiscussion(context.Background(), types.Discussion{
					Title:    "Test Discussion",
					Body:     "Test body",
					Category: "General",
					Labels:   []string{"question"},
				})
				return err
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
	mockGQL := &testutil.SimpleMockGraphQLClient{}

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
	mockGQL := &testutil.SimpleMockGraphQLClient{}

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
	gqlClient := &testutil.SimpleMockGraphQLClient{
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

	_, err := client.CreateDiscussion(context.Background(), types.Discussion{
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
	gqlClient := &testutil.SimpleMockGraphQLClient{
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
	_, err := client.CreateDiscussion(context.Background(), types.Discussion{
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
	gqlClient := &testutil.SimpleMockGraphQLClient{
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
	_, err := client.CreateDiscussion(context.Background(), types.Discussion{
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
	gqlClient := &testutil.SimpleMockGraphQLClient{
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
	_, err := client.CreatePR(context.Background(), types.PullRequest{
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
	_, err = client.CreatePR(context.Background(), types.PullRequest{
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
	_, err = client.CreatePR(context.Background(), types.PullRequest{
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
	gqlClient := &testutil.SimpleMockGraphQLClient{
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

	_, err := client.CreatePR(context.Background(), types.PullRequest{
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
	gqlClient := &testutil.SimpleMockGraphQLClient{
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

	_, err := client.CreatePR(context.Background(), types.PullRequest{
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
	gqlClient := &testutil.SimpleMockGraphQLClient{
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

	_, err := client.CreatePR(context.Background(), types.PullRequest{
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
	client, err := NewGHClient(context.Background(), "testowner", "testrepo")
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

	gqlClient := &testutil.SimpleMockGraphQLClient{
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

	_, err = client.CreateIssue(ctx, types.Issue{
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

func TestDeleteDiscussion(t *testing.T) {
	mockGQL := &testutil.SimpleMockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			// Check the mutation is correct
			if !strings.Contains(query, "deleteDiscussion") {
				t.Error("Expected deleteDiscussion mutation")
			}

			// Check variables
			discussionID, ok := variables["discussionId"].(string)
			if !ok || discussionID != "test-node-id" {
				t.Errorf("Expected discussionId 'test-node-id', got %v", variables["discussionId"])
			}

			// Mock response
			resp := response.(*struct {
				DeleteDiscussion struct {
					Discussion struct {
						ID    string `json:"id"`
						Title string `json:"title"`
					} `json:"discussion"`
				} `json:"deleteDiscussion"`
			})

			resp.DeleteDiscussion.Discussion.ID = "test-node-id"
			resp.DeleteDiscussion.Discussion.Title = "Test Discussion"
			return nil
		},
	}

	client := &GHClient{
		gqlClient: mockGQL,
		Owner:     "testowner",
		Repo:      "testrepo",
		logger:    &MockLogger{},
	}

	err := client.DeleteDiscussion(context.Background(), "test-node-id")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDeleteDiscussion_ValidationErrors(t *testing.T) {
	client := &GHClient{
		gqlClient: &testutil.SimpleMockGraphQLClient{},
		Owner:     "testowner",
		Repo:      "testrepo",
		logger:    &MockLogger{},
	}

	// Test empty node ID
	err := client.DeleteDiscussion(context.Background(), "")
	if err == nil {
		t.Error("Expected validation error for empty node ID")
	}

	// Test whitespace only node ID
	err = client.DeleteDiscussion(context.Background(), "   ")
	if err == nil {
		t.Error("Expected validation error for whitespace-only node ID")
	}

	// Test nil GraphQL client
	nilClient := &GHClient{
		gqlClient: nil,
		Owner:     "testowner",
		Repo:      "testrepo",
		logger:    &MockLogger{},
	}

	err = nilClient.DeleteDiscussion(context.Background(), "test-node-id")
	if err == nil {
		t.Error("Expected validation error for nil GraphQL client")
	}
}

func TestDeleteDiscussion_GraphQLError(t *testing.T) {
	mockGQL := &testutil.SimpleMockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			return errors.New("GraphQL error")
		},
	}

	client := &GHClient{
		gqlClient: mockGQL,
		Owner:     "testowner",
		Repo:      "testrepo",
		logger:    &MockLogger{},
	}

	err := client.DeleteDiscussion(context.Background(), "test-node-id")
	if err == nil {
		t.Error("Expected an error from GraphQL client")
	}
}

// TestListIssues tests the ListIssues function
func TestListIssues(t *testing.T) {
	tests := []struct {
		name            string
		setupMockClient func() *testutil.SimpleMockGraphQLClient
		expectError     bool
		expectedCount   int
		errorText       string
	}{
		{
			name: "successful list with multiple pages",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				callCount := 0
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						callCount++
						resp := response.(*struct {
							Repository struct {
								Issues struct {
									Nodes []struct {
										ID     string `json:"id"`
										Number int    `json:"number"`
										Title  string `json:"title"`
										Body   string `json:"body"`
										Labels struct {
											Nodes []struct {
												Name string `json:"name"`
											} `json:"nodes"`
										} `json:"labels"`
									} `json:"nodes"`
									PageInfo struct {
										HasNextPage bool    `json:"hasNextPage"`
										EndCursor   *string `json:"endCursor"`
									} `json:"pageInfo"`
								} `json:"issues"`
							} `json:"repository"`
						})

						if callCount == 1 {
							// First page
							cursor := "cursor1"
							resp.Repository.Issues.Nodes = []struct {
								ID     string `json:"id"`
								Number int    `json:"number"`
								Title  string `json:"title"`
								Body   string `json:"body"`
								Labels struct {
									Nodes []struct {
										Name string `json:"name"`
									} `json:"nodes"`
								} `json:"labels"`
							}{
								{
									ID:     "issue1",
									Number: 1,
									Title:  "Issue 1",
									Body:   "Body 1",
									Labels: struct {
										Nodes []struct {
											Name string `json:"name"`
										} `json:"nodes"`
									}{
										Nodes: []struct {
											Name string `json:"name"`
										}{{Name: "bug"}},
									},
								},
							}
							resp.Repository.Issues.PageInfo.HasNextPage = true
							resp.Repository.Issues.PageInfo.EndCursor = &cursor
						} else {
							// Second page
							resp.Repository.Issues.Nodes = []struct {
								ID     string `json:"id"`
								Number int    `json:"number"`
								Title  string `json:"title"`
								Body   string `json:"body"`
								Labels struct {
									Nodes []struct {
										Name string `json:"name"`
									} `json:"nodes"`
								} `json:"labels"`
							}{
								{
									ID:     "issue2",
									Number: 2,
									Title:  "Issue 2",
									Body:   "Body 2",
									Labels: struct {
										Nodes []struct {
											Name string `json:"name"`
										} `json:"nodes"`
									}{
										Nodes: []struct {
											Name string `json:"name"`
										}{{Name: "enhancement"}},
									},
								},
							}
							resp.Repository.Issues.PageInfo.HasNextPage = false
							resp.Repository.Issues.PageInfo.EndCursor = nil
						}
						return nil
					},
				}
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "empty repository",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						resp := response.(*struct {
							Repository struct {
								Issues struct {
									Nodes []struct {
										ID     string `json:"id"`
										Number int    `json:"number"`
										Title  string `json:"title"`
										Body   string `json:"body"`
										Labels struct {
											Nodes []struct {
												Name string `json:"name"`
											} `json:"nodes"`
										} `json:"labels"`
									} `json:"nodes"`
									PageInfo struct {
										HasNextPage bool    `json:"hasNextPage"`
										EndCursor   *string `json:"endCursor"`
									} `json:"pageInfo"`
								} `json:"issues"`
							} `json:"repository"`
						})

						resp.Repository.Issues.Nodes = []struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							Title  string `json:"title"`
							Body   string `json:"body"`
							Labels struct {
								Nodes []struct {
									Name string `json:"name"`
								} `json:"nodes"`
							} `json:"labels"`
						}{}
						resp.Repository.Issues.PageInfo.HasNextPage = false
						resp.Repository.Issues.PageInfo.EndCursor = nil
						return nil
					},
				}
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name: "graphql client error",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						return fmt.Errorf("network error")
					},
				}
			},
			expectError: true,
			errorText:   "failed to fetch issues",
		},
		{
			name: "nil client validation",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{} // Return empty client to test validation
			},
			expectError: true,
			errorText:   "GraphQL client is not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GHClient{
				Owner:     "testowner",
				Repo:      "testrepo",
				gqlClient: tt.setupMockClient(),
				logger:    &MockLogger{},
			}

			// Special case for nil client test
			if tt.name == "nil client validation" {
				client.gqlClient = nil
			}

			issues, err := client.ListIssues(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(issues) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d", tt.expectedCount, len(issues))
			}

			// Validate issue structure if any returned
			if len(issues) > 0 {
				issue := issues[0]
				if issue.NodeID == "" {
					t.Error("Expected non-empty NodeID")
				}
				if issue.Title == "" {
					t.Error("Expected non-empty Title")
				}
			}
		})
	}
}

// TestListDiscussions tests the ListDiscussions function
func TestListDiscussions(t *testing.T) {
	tests := []struct {
		name            string
		setupMockClient func() *testutil.SimpleMockGraphQLClient
		expectError     bool
		expectedCount   int
		errorText       string
	}{
		{
			name: "successful list",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						resp := response.(*struct {
							Repository struct {
								Discussions struct {
									Nodes []struct {
										ID       string `json:"id"`
										Number   int    `json:"number"`
										Title    string `json:"title"`
										Body     string `json:"body"`
										Category struct {
											Name string `json:"name"`
										} `json:"category"`
									} `json:"nodes"`
									PageInfo struct {
										HasNextPage bool    `json:"hasNextPage"`
										EndCursor   *string `json:"endCursor"`
									} `json:"pageInfo"`
								} `json:"discussions"`
							} `json:"repository"`
						})

						resp.Repository.Discussions.Nodes = []struct {
							ID       string `json:"id"`
							Number   int    `json:"number"`
							Title    string `json:"title"`
							Body     string `json:"body"`
							Category struct {
								Name string `json:"name"`
							} `json:"category"`
						}{
							{
								ID:     "discussion1",
								Number: 1,
								Title:  "Discussion 1",
								Body:   "Body 1",
								Category: struct {
									Name string `json:"name"`
								}{Name: "General"},
							},
						}
						resp.Repository.Discussions.PageInfo.HasNextPage = false
						resp.Repository.Discussions.PageInfo.EndCursor = nil
						return nil
					},
				}
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "graphql error",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						return fmt.Errorf("api error")
					},
				}
			},
			expectError: true,
			errorText:   "failed to fetch discussions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GHClient{
				Owner:     "testowner",
				Repo:      "testrepo",
				gqlClient: tt.setupMockClient(),
				logger:    &MockLogger{},
			}

			discussions, err := client.ListDiscussions(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(discussions) != tt.expectedCount {
				t.Errorf("Expected %d discussions, got %d", tt.expectedCount, len(discussions))
			}
		})
	}
}

// TestListPRs tests the ListPRs function
func TestListPRs(t *testing.T) {
	tests := []struct {
		name            string
		setupMockClient func() *testutil.SimpleMockGraphQLClient
		expectError     bool
		expectedCount   int
		errorText       string
	}{
		{
			name: "successful list",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						resp := response.(*struct {
							Repository struct {
								PullRequests struct {
									Nodes []struct {
										ID          string `json:"id"`
										Number      int    `json:"number"`
										Title       string `json:"title"`
										Body        string `json:"body"`
										HeadRefName string `json:"headRefName"`
										BaseRefName string `json:"baseRefName"`
										Labels      struct {
											Nodes []struct {
												Name string `json:"name"`
											} `json:"nodes"`
										} `json:"labels"`
									} `json:"nodes"`
									PageInfo struct {
										HasNextPage bool    `json:"hasNextPage"`
										EndCursor   *string `json:"endCursor"`
									} `json:"pageInfo"`
								} `json:"pullRequests"`
							} `json:"repository"`
						})

						resp.Repository.PullRequests.Nodes = []struct {
							ID          string `json:"id"`
							Number      int    `json:"number"`
							Title       string `json:"title"`
							Body        string `json:"body"`
							HeadRefName string `json:"headRefName"`
							BaseRefName string `json:"baseRefName"`
							Labels      struct {
								Nodes []struct {
									Name string `json:"name"`
								} `json:"nodes"`
							} `json:"labels"`
						}{
							{
								ID:          "pr1",
								Number:      1,
								Title:       "PR 1",
								Body:        "Body 1",
								HeadRefName: "feature",
								BaseRefName: "main",
								Labels: struct {
									Nodes []struct {
										Name string `json:"name"`
									} `json:"nodes"`
								}{
									Nodes: []struct {
										Name string `json:"name"`
									}{{Name: "feature"}},
								},
							},
						}
						resp.Repository.PullRequests.PageInfo.HasNextPage = false
						resp.Repository.PullRequests.PageInfo.EndCursor = nil
						return nil
					},
				}
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "graphql error",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						return fmt.Errorf("api error")
					},
				}
			},
			expectError: true,
			errorText:   "failed to fetch pull requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GHClient{
				Owner:     "testowner",
				Repo:      "testrepo",
				gqlClient: tt.setupMockClient(),
				logger:    &MockLogger{},
			}

			prs, err := client.ListPRs(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(prs) != tt.expectedCount {
				t.Errorf("Expected %d PRs, got %d", tt.expectedCount, len(prs))
			}
		})
	}
}

// TestDeleteIssue tests the DeleteIssue function
func TestDeleteIssue(t *testing.T) {
	tests := []struct {
		name            string
		nodeID          string
		setupMockClient func() *testutil.SimpleMockGraphQLClient
		expectError     bool
		errorText       string
	}{
		{
			name:   "successful deletion",
			nodeID: "issue-node-123",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						if !strings.Contains(query, "closeIssue") {
							t.Error("Expected closeIssue mutation")
						}

						// Check variables
						issueID, ok := variables["issueId"].(string)
						if !ok || issueID != "issue-node-123" {
							t.Errorf("Expected issueId 'issue-node-123', got %v", variables["issueId"])
						}

						// Mock response
						resp := response.(*struct {
							CloseIssue struct {
								Issue struct {
									ID    string `json:"id"`
									State string `json:"state"`
								} `json:"issue"`
							} `json:"closeIssue"`
						})

						resp.CloseIssue.Issue.ID = "issue-node-123"
						resp.CloseIssue.Issue.State = "CLOSED"
						return nil
					},
				}
			},
			expectError: false,
		},
		{
			name:   "empty node ID",
			nodeID: "",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{}
			},
			expectError: true,
			errorText:   "node ID cannot be empty",
		},
		{
			name:   "graphql error",
			nodeID: "issue-node-123",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						return fmt.Errorf("api error")
					},
				}
			},
			expectError: true,
			errorText:   "failed to close issue",
		},
		{
			name:   "issue not properly closed",
			nodeID: "issue-node-123",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						// Mock response with issue still open
						resp := response.(*struct {
							CloseIssue struct {
								Issue struct {
									ID    string `json:"id"`
									State string `json:"state"`
								} `json:"issue"`
							} `json:"closeIssue"`
						})

						resp.CloseIssue.Issue.ID = "issue-node-123"
						resp.CloseIssue.Issue.State = "OPEN" // Still open
						return nil
					},
				}
			},
			expectError: true,
			errorText:   "issue was not properly closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GHClient{
				Owner:     "testowner",
				Repo:      "testrepo",
				gqlClient: tt.setupMockClient(),
				logger:    &MockLogger{},
			}

			err := client.DeleteIssue(context.Background(), tt.nodeID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestDeletePR tests the DeletePR function
func TestDeletePR(t *testing.T) {
	tests := []struct {
		name            string
		nodeID          string
		setupMockClient func() *testutil.SimpleMockGraphQLClient
		expectError     bool
		errorText       string
	}{
		{
			name:   "successful deletion",
			nodeID: "pr-node-123",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						if !strings.Contains(query, "closePullRequest") {
							t.Error("Expected closePullRequest mutation")
						}

						// Check variables
						prID, ok := variables["pullRequestId"].(string)
						if !ok || prID != "pr-node-123" {
							t.Errorf("Expected pullRequestId 'pr-node-123', got %v", variables["pullRequestId"])
						}

						// Mock response
						resp := response.(*struct {
							ClosePullRequest struct {
								PullRequest struct {
									ID    string `json:"id"`
									State string `json:"state"`
								} `json:"pullRequest"`
							} `json:"closePullRequest"`
						})

						resp.ClosePullRequest.PullRequest.ID = "pr-node-123"
						resp.ClosePullRequest.PullRequest.State = "CLOSED"
						return nil
					},
				}
			},
			expectError: false,
		},
		{
			name:   "empty node ID",
			nodeID: "",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{}
			},
			expectError: true,
			errorText:   "node ID cannot be empty",
		},
		{
			name:   "graphql error",
			nodeID: "pr-node-123",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						return fmt.Errorf("api error")
					},
				}
			},
			expectError: true,
			errorText:   "failed to close pull request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GHClient{
				Owner:     "testowner",
				Repo:      "testrepo",
				gqlClient: tt.setupMockClient(),
				logger:    &MockLogger{},
			}

			err := client.DeletePR(context.Background(), tt.nodeID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestDeleteLabel tests the DeleteLabel function
func TestDeleteLabel(t *testing.T) {
	tests := []struct {
		name            string
		labelName       string
		setupMockClient func() *testutil.SimpleMockGraphQLClient
		expectError     bool
		errorText       string
	}{
		{
			name:      "successful deletion",
			labelName: "test-label",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						if strings.Contains(query, "repository(owner:") && strings.Contains(query, "label(name:") {
							// First query: get label ID by name
							resp := response.(*struct {
								Repository struct {
									Label struct {
										ID string `json:"id"`
									} `json:"label"`
								} `json:"repository"`
							})
							resp.Repository.Label.ID = "label-id-123"
							return nil
						} else if strings.Contains(query, "deleteLabel") {
							// Second query: delete label mutation
							// Check variables
							labelID, ok := variables["labelId"].(string)
							if !ok || labelID != "label-id-123" {
								t.Errorf("Expected labelId 'label-id-123', got %v", variables["labelId"])
							}

							// Mock response
							resp := response.(*struct {
								DeleteLabel struct {
									ClientMutationID string `json:"clientMutationId"`
								} `json:"deleteLabel"`
							})

							resp.DeleteLabel.ClientMutationID = "test-mutation"
							return nil
						}
						return fmt.Errorf("unexpected query: %s", query)
					},
				}
			},
			expectError: false,
		},
		{
			name:      "empty label name",
			labelName: "",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{}
			},
			expectError: true,
			errorText:   "label name cannot be empty",
		},
		{
			name:      "graphql error",
			labelName: "test-label",
			setupMockClient: func() *testutil.SimpleMockGraphQLClient {
				return &testutil.SimpleMockGraphQLClient{
					DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
						return fmt.Errorf("api error")
					},
				}
			},
			expectError: true,
			errorText:   "failed to find label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GHClient{
				Owner:     "testowner",
				Repo:      "testrepo",
				gqlClient: tt.setupMockClient(),
				logger:    &MockLogger{},
			}

			err := client.DeleteLabel(context.Background(), tt.labelName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
