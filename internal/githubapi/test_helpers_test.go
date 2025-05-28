package githubapi

import (
	"strings"
	"testing"

	"github.com/chrisreddington/gh-demo/internal/testutil"
)

// TestConfigurableMockGraphQLClient_HandlerFunctions tests the individual handler functions
func TestConfigurableMockGraphQLClient_HandlerFunctions(t *testing.T) {
	tests := []struct {
		name        string
		handlerFunc func(*ConfigurableMockGraphQLClient) func(interface{}) error
		setupMock   func() *ConfigurableMockGraphQLClient
		expectError bool
	}{
		{
			name: "handleRepositoryQuery success",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handleRepositoryQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			expectError: false,
		},
		{
			name: "handleRepositoryQuery error",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handleRepositoryQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewErrorMockGraphQL(map[string]string{
					"repository": "repository access denied",
				})
			},
			expectError: true,
		},
		{
			name: "handleLabelsQuery success",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handleLabelsQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			expectError: false,
		},
		{
			name: "handleLabelsQuery error",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handleLabelsQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewErrorMockGraphQL(map[string]string{
					"labels": "labels access denied",
				})
			},
			expectError: true,
		},
		{
			name: "handleIssueCreationQuery success",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handleIssueCreationQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			expectError: false,
		},
		{
			name: "handleIssueCreationQuery error",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handleIssueCreationQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewErrorMockGraphQL(map[string]string{
					"createIssue": "issue creation failed",
				})
			},
			expectError: true,
		},
		{
			name: "handlePullRequestCreationQuery success",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handlePullRequestCreationQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			expectError: false,
		},
		{
			name: "handlePullRequestCreationQuery error",
			handlerFunc: func(client *ConfigurableMockGraphQLClient) func(interface{}) error {
				return client.handlePullRequestCreationQuery
			},
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewErrorMockGraphQL(map[string]string{
					"createPR": "PR creation failed",
				})
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupMock()
			handler := tt.handlerFunc(client)

			// Create appropriate response struct based on handler type
			var response interface{}

			// Extract operation type from test name for cleaner switch logic
			operationType := ""
			switch {
			case strings.Contains(tt.name, "handleRepositoryQuery"):
				operationType = "repository"
			case strings.Contains(tt.name, "handleLabelsQuery"):
				operationType = "labels"
			case strings.Contains(tt.name, "handleIssueCreationQuery"):
				operationType = "issue"
			case strings.Contains(tt.name, "handlePullRequestCreationQuery"):
				operationType = "pullRequest"
			}

			switch operationType {
			case "repository":
				response = &struct {
					Repository struct {
						ID string `json:"id"`
					} `json:"repository"`
				}{}
			case "labels":
				response = &struct {
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
				}{}
			case "issue":
				response = &struct {
					CreateIssue struct {
						Issue struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							Title  string `json:"title"`
							URL    string `json:"url"`
						} `json:"issue"`
					} `json:"createIssue"`
				}{}
			case "pullRequest":
				response = &struct {
					CreatePullRequest struct {
						PullRequest struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							Title  string `json:"title"`
							URL    string `json:"url"`
						} `json:"pullRequest"`
					} `json:"createPullRequest"`
				}{}
			}

			err := handler(response)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Validate that non-error cases set default values
			if !tt.expectError && err == nil {
				switch resp := response.(type) {
				case *struct {
					Repository struct {
						ID string `json:"id"`
					} `json:"repository"`
				}:
					if resp.Repository.ID != testutil.DefaultValues.RepositoryID {
						t.Errorf("Expected repository ID %s, got %s", testutil.DefaultValues.RepositoryID, resp.Repository.ID)
					}
				case *struct {
					CreateIssue struct {
						Issue struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							Title  string `json:"title"`
							URL    string `json:"url"`
						} `json:"issue"`
					} `json:"createIssue"`
				}:
					if resp.CreateIssue.Issue.ID != testutil.DefaultValues.IssueID {
						t.Errorf("Expected issue ID %s, got %s", testutil.DefaultValues.IssueID, resp.CreateIssue.Issue.ID)
					}
					if resp.CreateIssue.Issue.Title != "Test Issue" {
						t.Errorf("Expected issue title 'Test Issue', got %s", resp.CreateIssue.Issue.Title)
					}
				case *struct {
					CreatePullRequest struct {
						PullRequest struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							Title  string `json:"title"`
							URL    string `json:"url"`
						} `json:"pullRequest"`
					} `json:"createPullRequest"`
				}:
					if resp.CreatePullRequest.PullRequest.ID != testutil.DefaultValues.PRID {
						t.Errorf("Expected PR ID %s, got %s", testutil.DefaultValues.PRID, resp.CreatePullRequest.PullRequest.ID)
					}
					if resp.CreatePullRequest.PullRequest.Title != "Test PR" {
						t.Errorf("Expected PR title 'Test PR', got %s", resp.CreatePullRequest.PullRequest.Title)
					}
				}
			}
		})
	}
}

// TestHandleLabelCreationQuery tests the label creation handler with variables
func TestHandleLabelCreationQuery(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *ConfigurableMockGraphQLClient
		variables   map[string]interface{}
		expectError bool
	}{
		{
			name: "success with input variables",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewDefaultMockGraphQL()
			},
			variables: map[string]interface{}{
				"input": map[string]interface{}{
					"name":        "test-label",
					"color":       "ff0000",
					"description": "Test description",
				},
			},
			expectError: false,
		},
		{
			name: "error response",
			setupMock: func() *ConfigurableMockGraphQLClient {
				return NewErrorMockGraphQL(map[string]string{
					"createLabel": "label creation failed",
				})
			},
			variables:   map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupMock()
			response := &struct {
				CreateLabel struct {
					Label struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						Color       string `json:"color"`
						Description string `json:"description"`
					} `json:"label"`
				} `json:"createLabel"`
			}{}

			err := client.handleLabelCreationQuery(tt.variables, response)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Validate variable copying for success case
			if !tt.expectError && err == nil {
				if input, ok := tt.variables["input"].(map[string]interface{}); ok {
					if name, ok := input["name"].(string); ok && response.CreateLabel.Label.Name != name {
						t.Errorf("Expected label name %s, got %s", name, response.CreateLabel.Label.Name)
					}
					if color, ok := input["color"].(string); ok && response.CreateLabel.Label.Color != color {
						t.Errorf("Expected label color %s, got %s", color, response.CreateLabel.Label.Color)
					}
					if desc, ok := input["description"].(string); ok && response.CreateLabel.Label.Description != desc {
						t.Errorf("Expected label description %s, got %s", desc, response.CreateLabel.Label.Description)
					}
				}
			}
		})
	}
}
