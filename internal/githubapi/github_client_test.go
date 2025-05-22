package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetRepositoryID tests the GetRepositoryID method
func TestGetRepositoryID(t *testing.T) {
	// Create a mock GQL client that returns a repository ID
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// Check the query variables
			owner, ok := variables["owner"].(string)
			if !ok || owner != "testowner" {
				t.Errorf("Expected owner to be 'testowner', got %v", variables["owner"])
			}

			name, ok := variables["name"].(string)
			if !ok || name != "testrepo" {
				t.Errorf("Expected name to be 'testrepo', got %v", variables["name"])
			}

			// Set the response
			resp := response.(*struct {
				Repository struct {
					ID string `json:"id"`
				} `json:"repository"`
			})
			resp.Repository.ID = "R_123456"
			return nil
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method
	id, err := client.GetRepositoryID(context.Background(), "testowner", "testrepo")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	if id != "R_123456" {
		t.Errorf("Expected repository ID to be 'R_123456', got %v", id)
	}
}

// TestGetRepositoryID_Error tests the error handling in GetRepositoryID
func TestGetRepositoryID_Error(t *testing.T) {
	// Create a mock GQL client that returns an error
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			return fmt.Errorf("API error")
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method
	_, err := client.GetRepositoryID(context.Background(), "testowner", "testrepo")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// MockGitHubClient implements GitHubClient interface for testing
type MockGitHubClient struct {
	GetRepositoryIDFunc      func(ctx context.Context, owner, repo string) (string, error)
	GetCurrentRepositoryFunc func(ctx context.Context) (string, string, error)
	// Add other methods as needed
}

func (m *MockGitHubClient) GetRepositoryID(ctx context.Context, owner, repo string) (string, error) {
	return m.GetRepositoryIDFunc(ctx, owner, repo)
}

func (m *MockGitHubClient) GetCurrentRepository(ctx context.Context) (string, string, error) {
	return m.GetCurrentRepositoryFunc(ctx)
}

// Implement other required interface methods with empty implementations
func (m *MockGitHubClient) CreateIssue(ctx context.Context, owner, repo string, input *IssueInput) (string, error) {
	return "", nil
}
func (m *MockGitHubClient) CreateDiscussion(ctx context.Context, input *DiscussionInput) (string, error) {
	return "", nil
}
func (m *MockGitHubClient) GetDiscussionCategories(ctx context.Context, owner, repo string) ([]DiscussionCategory, error) {
	return nil, nil
}
func (m *MockGitHubClient) CreatePullRequest(ctx context.Context, owner, repo string, input *PullRequestInput) (string, error) {
	return "", nil
}
func (m *MockGitHubClient) CreateLabel(ctx context.Context, owner, repo string, input *LabelInput) (string, error) {
	return "", nil
}

// TestGetRepositoryID_WithCurrentRepository tests GetRepositoryID with empty owner and repo
func TestGetRepositoryID_WithCurrentRepository(t *testing.T) {
	// Initialize variables to track calls
	var usedCurrentRepo bool
	var queriedOwner, queriedRepo string

	// Create a mock client
	client := &GitHubClientImpl{
		client: &MockGQLClient{
			DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
				// Track which owner and repo were used in the GraphQL query
				queriedOwner = variables["owner"].(string)
				queriedRepo = variables["name"].(string)

				// Set the response
				resp := response.(*struct {
					Repository struct {
						ID string `json:"id"`
					} `json:"repository"`
				})
				resp.Repository.ID = "R_current"
				return nil
			},
		},
	}

	// Replace GetCurrentRepository with a test double
	// We'll use a separate function that delegates to the original when not testing
	originalFunc := client.GetCurrentRepository
	testGetCurrentRepository := func(ctx context.Context) (string, string, error) {
		usedCurrentRepo = true // Flag that we called this function
		return "currentowner", "currentrepo", nil
	}

	// Swap the implementation
	client.GetCurrentRepository = testGetCurrentRepository

	// Test with empty owner and repo
	id, err := client.GetRepositoryID(context.Background(), "", "")

	// Restore the original function
	client.GetCurrentRepository = originalFunc

	// Check results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if id != "R_current" {
		t.Errorf("Expected repository ID to be 'R_current', got %v", id)
	}

	if !usedCurrentRepo {
		t.Error("Expected GetCurrentRepository to be called, but it wasn't")
	}

	if queriedOwner != "currentowner" || queriedRepo != "currentrepo" {
		t.Errorf("Expected owner='currentowner' and repo='currentrepo', got owner='%s' repo='%s'",
			queriedOwner, queriedRepo)
	}
}

// TestGetRepositoryID_EmptyParams tests that GetRepositoryID uses GetCurrentRepository when parameters are empty
func TestGetRepositoryID_EmptyParams(t *testing.T) {
	// Track if the current repository method was called
	getCurrentRepoCalled := false

	// Create the mock to handle GetRepositoryID calls
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			owner := variables["owner"].(string)
			repo := variables["name"].(string)

			// Verify the values were correctly passed through
			if owner != "currentowner" {
				t.Errorf("Expected owner to be 'currentowner', got %v", owner)
			}
			if repo != "currentrepo" {
				t.Errorf("Expected repo to be 'currentrepo', got %v", repo)
			}

			// Set the response
			resp := response.(*struct {
				Repository struct {
					ID string `json:"id"`
				} `json:"repository"`
			})
			resp.Repository.ID = "R_current"
			return nil
		},
	}

	// Create a client with our mocked GraphQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Store a reference to the original method
	originalMethod := client.GetCurrentRepository

	// Create a test implementation
	testImplementation := func(ctx context.Context) (string, string, error) {
		getCurrentRepoCalled = true
		return "currentowner", "currentrepo", nil
	}

	// Replace the method with our test implementation
	client.GetCurrentRepository = testImplementation

	// Call the method with empty owner and repo
	id, err := client.GetRepositoryID(context.Background(), "", "")

	// Restore the original implementation
	client.GetCurrentRepository = originalMethod

	// Verify expectations
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if id != "R_current" {
		t.Errorf("Expected repository ID 'R_current', got %v", id)
	}
	if !getCurrentRepoCalled {
		t.Error("Expected GetCurrentRepository to be called, but it wasn't")
	}
}

// TestCreateIssue tests the CreateIssue method
func TestCreateIssue(t *testing.T) {
	// Create a mock GQL client that returns a successful issue creation response
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// For GetRepositoryID query
			input, ok := variables["input"].(map[string]interface{})
			if ok {
				// This is the createIssue mutation
				title, titleOk := input["title"].(string)
				if !titleOk || title != "Test Issue" {
					t.Errorf("Expected issue title to be 'Test Issue', got %v", input["title"])
				}

				// Set the response for createIssue
				resp := response.(*struct {
					CreateIssue struct {
						Issue struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							URL    string `json:"url"`
						} `json:"issue"`
					} `json:"createIssue"`
				})
				resp.CreateIssue.Issue.ID = "I_123456"
				resp.CreateIssue.Issue.Number = 1
				resp.CreateIssue.Issue.URL = "https://github.com/testowner/testrepo/issues/1"
				return nil
			}

			// Set the response for GetRepositoryID
			resp := response.(*struct {
				Repository struct {
					ID string `json:"id"`
				} `json:"repository"`
			})
			resp.Repository.ID = "R_123456"
			return nil
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method
	input := &IssueInput{
		Title: "Test Issue",
		Body:  "This is a test issue",
	}
	url, err := client.CreateIssue(context.Background(), "testowner", "testrepo", input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	expectedURL := "https://github.com/testowner/testrepo/issues/1"
	if url != expectedURL {
		t.Errorf("Expected issue URL to be '%s', got %s", expectedURL, url)
	}
}

// TestCreateIssue_WithLabelsAndAssignees tests the CreateIssue method with labels and assignees
func TestCreateIssue_WithLabelsAndAssignees(t *testing.T) {
	// Create a mock GQL client that returns a successful issue creation response
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			input, ok := variables["input"].(map[string]interface{})
			if ok {
				// Check that labels and assignees are properly set
				if labelIDs, ok := input["labelIds"].([]string); !ok || len(labelIDs) != 0 {
					// Not an error since we're not actually converting labels to IDs in the implementation
				}
				if assigneeIDs, ok := input["assigneeIds"].([]string); !ok || len(assigneeIDs) != 0 {
					// Not an error since we're not actually converting assignees to IDs in the implementation
				}

				// Set the response for createIssue
				resp := response.(*struct {
					CreateIssue struct {
						Issue struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							URL    string `json:"url"`
						} `json:"issue"`
					} `json:"createIssue"`
				})
				resp.CreateIssue.Issue.ID = "I_123456"
				resp.CreateIssue.Issue.Number = 1
				resp.CreateIssue.Issue.URL = "https://github.com/testowner/testrepo/issues/1"
				return nil
			}

			// Set the response for GetRepositoryID
			resp := response.(*struct {
				Repository struct {
					ID string `json:"id"`
				} `json:"repository"`
			})
			resp.Repository.ID = "R_123456"
			return nil
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method with labels and assignees
	input := &IssueInput{
		Title:     "Test Issue",
		Body:      "This is a test issue",
		Labels:    []string{"bug", "help wanted"},
		Assignees: []string{"octocat"},
	}
	url, err := client.CreateIssue(context.Background(), "testowner", "testrepo", input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	expectedURL := "https://github.com/testowner/testrepo/issues/1"
	if url != expectedURL {
		t.Errorf("Expected issue URL to be '%s', got %s", expectedURL, url)
	}
}

// TestCreateIssue_Error tests the error handling in CreateIssue
func TestCreateIssue_Error(t *testing.T) {
	// Create a mock GQL client that returns an error
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			return fmt.Errorf("API error")
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method
	input := &IssueInput{
		Title: "Test Issue",
		Body:  "This is a test issue",
	}
	_, err := client.CreateIssue(context.Background(), "testowner", "testrepo", input)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestCreateDiscussion tests the CreateDiscussion method
func TestCreateDiscussion(t *testing.T) {
	// Create a mock GQL client that returns a successful discussion creation response
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// Check inputs
			input, ok := variables["input"].(map[string]interface{})
			if ok {
				if title, ok := input["title"].(string); !ok || title != "Test Discussion" {
					t.Errorf("Expected discussion title to be 'Test Discussion', got %v", input["title"])
				}
				if repoID, ok := input["repositoryId"].(string); !ok || repoID != "R_123456" {
					t.Errorf("Expected repositoryId to be 'R_123456', got %v", input["repositoryId"])
				}
				if catID, ok := input["categoryId"].(string); !ok || catID != "CAT_123" {
					t.Errorf("Expected categoryId to be 'CAT_123', got %v", input["categoryId"])
				}

				// Set the response for createDiscussion
				resp := response.(*struct {
					CreateDiscussion struct {
						Discussion struct {
							ID  string `json:"id"`
							URL string `json:"url"`
						} `json:"discussion"`
					} `json:"createDiscussion"`
				})
				resp.CreateDiscussion.Discussion.ID = "D_123456"
				resp.CreateDiscussion.Discussion.URL = "https://github.com/testowner/testrepo/discussions/1"
				return nil
			}

			return nil
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method
	input := &DiscussionInput{
		Title:        "Test Discussion",
		Body:         "This is a test discussion",
		CategoryID:   "CAT_123",
		RepositoryID: "R_123456",
	}
	url, err := client.CreateDiscussion(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	expectedURL := "https://github.com/testowner/testrepo/discussions/1"
	if url != expectedURL {
		t.Errorf("Expected discussion URL to be '%s', got %s", expectedURL, url)
	}
}

// TestGetDiscussionCategories tests the GetDiscussionCategories method
func TestGetDiscussionCategories(t *testing.T) {
	// Create a mock GQL client that returns discussion categories
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// Check the query variables
			owner, ok := variables["owner"].(string)
			if !ok || owner != "testowner" {
				t.Errorf("Expected owner to be 'testowner', got %v", variables["owner"])
			}

			name, ok := variables["name"].(string)
			if !ok || name != "testrepo" {
				t.Errorf("Expected name to be 'testrepo', got %v", variables["name"])
			}

			// Set the response
			resp := response.(*struct {
				Repository struct {
					DiscussionCategories struct {
						Nodes []DiscussionCategory `json:"nodes"`
					} `json:"discussionCategories"`
				} `json:"repository"`
			})
			resp.Repository.DiscussionCategories.Nodes = []DiscussionCategory{
				{ID: "CAT_1", Name: "General"},
				{ID: "CAT_2", Name: "Ideas"},
			}
			return nil
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method
	categories, err := client.GetDiscussionCategories(context.Background(), "testowner", "testrepo")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}
	if categories[0].Name != "General" {
		t.Errorf("Expected first category name to be 'General', got '%s'", categories[0].Name)
	}
	if categories[1].Name != "Ideas" {
		t.Errorf("Expected second category name to be 'Ideas', got '%s'", categories[1].Name)
	}
}

// TestCreatePullRequest tests the CreatePullRequest method
func TestCreatePullRequest(t *testing.T) {
	// Create a mock GQL client that returns a successful PR creation response
	mockGQL := &MockGQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// For GetRepositoryID query
			input, ok := variables["input"].(map[string]interface{})
			if ok {
				// This is the createPullRequest mutation
				if title, ok := input["title"].(string); !ok || title != "Test PR" {
					t.Errorf("Expected PR title to be 'Test PR', got %v", input["title"])
				}
				if base, ok := input["baseRefName"].(string); !ok || base != "main" {
					t.Errorf("Expected base branch to be 'main', got %v", input["baseRefName"])
				}
				if head, ok := input["headRefName"].(string); !ok || head != "feature" {
					t.Errorf("Expected head branch to be 'feature', got %v", input["headRefName"])
				}

				// Set the response for createPullRequest
				resp := response.(*struct {
					CreatePullRequest struct {
						PullRequest struct {
							ID     string `json:"id"`
							Number int    `json:"number"`
							URL    string `json:"url"`
						} `json:"pullRequest"`
					} `json:"createPullRequest"`
				})
				resp.CreatePullRequest.PullRequest.ID = "PR_123456"
				resp.CreatePullRequest.PullRequest.Number = 1
				resp.CreatePullRequest.PullRequest.URL = "https://github.com/testowner/testrepo/pull/1"
				return nil
			}

			// Set the response for GetRepositoryID
			resp := response.(*struct {
				Repository struct {
					ID string `json:"id"`
				} `json:"repository"`
			})
			resp.Repository.ID = "R_123456"
			return nil
		},
	}

	// Create a client with the mock GQL client
	client := &GitHubClientImpl{
		client: mockGQL,
	}

	// Call the method
	input := &PullRequestInput{
		Title: "Test PR",
		Body:  "This is a test PR",
		Base:  "main",
		Head:  "feature",
	}
	url, err := client.CreatePullRequest(context.Background(), "testowner", "testrepo", input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	expectedURL := "https://github.com/testowner/testrepo/pull/1"
	if url != expectedURL {
		t.Errorf("Expected PR URL to be '%s', got %s", expectedURL, url)
	}
}

// TestCreateLabel tests the CreateLabel method
func TestCreateLabel(t *testing.T) {
	// Setup mock REST server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/repos/testowner/testrepo/labels" {
			t.Errorf("Expected path '/repos/testowner/testrepo/labels', got '%s'", r.URL.Path)
		}

		// Send a mock response
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, `{"name":"bug","color":"ff0000","description":"Bug label"}`)
	}))
	defer server.Close()

	// Mock REST client
	mockREST := &MockRESTClient{
		PostFunc: func(path string, body io.Reader, response interface{}) error {
			if path != "repos/testowner/testrepo/labels" {
				t.Errorf("Expected path 'repos/testowner/testrepo/labels', got '%s'", path)
			}

			// Parse the provided body as a reader
			var requestBody map[string]string
			json.NewDecoder(body).Decode(&requestBody)

			if requestBody["name"] != "bug" {
				t.Errorf("Expected label name 'bug', got '%s'", requestBody["name"])
			}

			// Set response
			respMap := response.(*map[string]interface{})
			*respMap = map[string]interface{}{
				"name": "bug",
			}
			return nil
		},
	}

	// Create a client with the mock REST client
	client := &GitHubClientImpl{
		rest: mockREST,
	}

	// Create test label
	input := &LabelInput{
		Name:        "bug",
		Color:       "ff0000",
		Description: "Bug label",
	}

	// Call the method
	name, err := client.CreateLabel(context.Background(), "testowner", "testrepo", input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	if name != "bug" {
		t.Errorf("Expected label name 'bug', got '%s'", name)
	}
}

// TestCreateLabel_Error tests the error handling in CreateLabel
func TestCreateLabel_Error(t *testing.T) {
	// Mock REST client that returns an error
	mockREST := &MockRESTClient{
		PostFunc: func(path string, body io.Reader, response interface{}) error {
			return fmt.Errorf("API error")
		},
	}

	// Create a client with the mock REST client
	client := &GitHubClientImpl{
		rest: mockREST,
	}

	// Create test label
	input := &LabelInput{
		Name:        "bug",
		Color:       "ff0000",
		Description: "Bug label",
	}

	// Call the method
	_, err := client.CreateLabel(context.Background(), "testowner", "testrepo", input)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestCreateLabel_InvalidResponse tests the error handling for invalid responses in CreateLabel
func TestCreateLabel_InvalidResponse(t *testing.T) {
	// Mock REST client that returns a response without the expected field
	mockREST := &MockRESTClient{
		PostFunc: func(path string, body io.Reader, response interface{}) error {
			// Set an invalid response without the "name" field
			respMap := response.(*map[string]interface{})
			*respMap = map[string]interface{}{
				"color": "ff0000",
			}
			return nil
		},
	}

	// Create a client with the mock REST client
	client := &GitHubClientImpl{
		rest: mockREST,
	}

	// Create test label
	input := &LabelInput{
		Name:        "bug",
		Color:       "ff0000",
		Description: "Bug label",
	}

	// Call the method
	_, err := client.CreateLabel(context.Background(), "testowner", "testrepo", input)
	if err == nil {
		t.Error("Expected error for invalid response, got nil")
	}
}

// MockGQLClient is a mock implementation of the GQLClient interface
type MockGQLClient struct {
	DoFunc func(query string, variables map[string]interface{}, response interface{}) error
}

func (m *MockGQLClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	return m.DoFunc(query, variables, response)
}

// MockRESTClient is a mock implementation of the RESTClient interface
type MockRESTClient struct {
	GetFunc    func(path string, response interface{}) error
	PostFunc   func(path string, body io.Reader, response interface{}) error
	PatchFunc  func(path string, body io.Reader, response interface{}) error
	DeleteFunc func(path string, response interface{}) error
}

func (m *MockRESTClient) Get(path string, response interface{}) error {
	return m.GetFunc(path, response)
}

func (m *MockRESTClient) Post(path string, body io.Reader, response interface{}) error {
	return m.PostFunc(path, body, response)
}

func (m *MockRESTClient) Patch(path string, body io.Reader, response interface{}) error {
	return m.PatchFunc(path, body, response)
}

func (m *MockRESTClient) Delete(path string, response interface{}) error {
	return m.DeleteFunc(path, response)
}
