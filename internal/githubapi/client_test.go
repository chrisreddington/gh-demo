package githubapi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockGraphQLClient implements the GraphQLClient interface for testing
type MockGraphQLClient struct {
	DoFunc func(string, map[string]interface{}, interface{}) error
}

func (m *MockGraphQLClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	if m.DoFunc != nil {
		return m.DoFunc(query, variables, response)
	}
	return nil
}

// MockRESTClient implements the interface required by RESTClient
type MockRESTClient struct {
	RequestFunc func(string, string, io.Reader) (*http.Response, error)
}

func (m *MockRESTClient) Request(method string, path string, body io.Reader) (*http.Response, error) {
	if m.RequestFunc != nil {
		return m.RequestFunc(method, path, body)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("{}")),
	}, nil
}

// Tests for GHClient
func TestNewGHClient(t *testing.T) {
	client := NewGHClient("testowner", "testrepo")
	if client.Owner != "testowner" {
		t.Errorf("Expected owner to be 'testowner', got '%s'", client.Owner)
	}
	if client.Repo != "testrepo" {
		t.Errorf("Expected repo to be 'testrepo', got '%s'", client.Repo)
	}
}

func TestGHClientWithMockClients(t *testing.T) {
	restClient := &MockRESTClient{
		RequestFunc: func(method string, path string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		},
	}

	gqlClient := &MockGraphQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// Mock successful responses for GraphQL queries
			return nil
		},
	}

	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: &RESTClient{client: restClient},
		gqlClient:  gqlClient,
	}

	// Test CreateLabel
	err := client.CreateLabel("test-label")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test CreateIssue
	err = client.CreateIssue(IssueInput{
		Title:     "Test Issue",
		Body:      "This is a test issue",
		Labels:    []string{"bug"},
		Assignees: []string{"testuser"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test CreatePR
	err = client.CreatePR(PRInput{
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
	gqlClient := &MockGraphQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// Mock successful labels response
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
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	labels, err := client.ListLabels()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}
	if labels[0] != "bug" || labels[1] != "enhancement" || labels[2] != "documentation" {
		t.Errorf("Expected labels [bug, enhancement, documentation], got %v", labels)
	}
}

func TestCreateDiscussion(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			// Handle both repository query and create discussion mutation
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
					{ID: "cat-id-456", Name: "Q&A"},
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
			}
			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreateDiscussion(DiscussionInput{
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
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
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

	err := client.CreateDiscussion(DiscussionInput{
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
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			return fmt.Errorf("GraphQL error: network timeout")
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	err := client.CreateDiscussion(DiscussionInput{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
	})
	if err == nil {
		t.Error("Expected an error from GraphQL client")
	}
}

func TestListLabelsError(t *testing.T) {
	gqlClient := &MockGraphQLClient{
		DoFunc: func(query string, variables map[string]interface{}, response interface{}) error {
			return fmt.Errorf("GraphQL error: unauthorized")
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: gqlClient,
	}

	_, err := client.ListLabels()
	if err == nil {
		t.Error("Expected an error from GraphQL client")
	}
}

func TestCreateDiscussionError(t *testing.T) {
	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: nil, // This will cause an error
	}

	err := client.CreateDiscussion(DiscussionInput{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
	})
	if err == nil {
		t.Error("Expected an error when GraphQL client is nil")
	}
}

func TestListLabelsNilClient(t *testing.T) {
	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: nil, // This will cause an error
	}

	_, err := client.ListLabels()
	if err == nil {
		t.Error("Expected an error when GraphQL client is nil")
	}
	if !strings.Contains(err.Error(), "GraphQL client is not initialized") {
		t.Errorf("Expected error to mention GraphQL client not initialized, got: %v", err)
	}
}

// REST client tests that should work
func TestRESTClientRequest(t *testing.T) {
	restClient := &MockRESTClient{
		RequestFunc: func(method string, path string, body io.Reader) (*http.Response, error) {
			if method != "GET" {
				t.Errorf("Expected method GET, got %s", method)
			}
			if path != "test/path" {
				t.Errorf("Expected path test/path, got %s", path)
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"test": "data"}`)),
			}, nil
		},
	}

	wrappedClient := &RESTClient{client: restClient}
	var response map[string]interface{}
	err := wrappedClient.Request("GET", "test/path", nil, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response to be populated")
	}
}

func TestCreatePR(t *testing.T) {
	restClient := &MockRESTClient{
		RequestFunc: func(method string, path string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(strings.NewReader(`{"number": 1, "html_url": "https://github.com/test/test/pull/1"}`)),
			}, nil
		},
	}

	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: &RESTClient{client: restClient},
		gqlClient:  nil, // PR creation uses REST API
	}

	err := client.CreatePR(PRInput{
		Title: "Test PR",
		Body:  "This is a test PR",
		Head:  "feature-branch",
		Base:  "main",
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreatePRValidation(t *testing.T) {
	restClient := &MockRESTClient{
		RequestFunc: func(method string, path string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: 422,
				Body:       io.NopCloser(strings.NewReader(`{"message": "Validation Failed", "errors": [{"message": "No commits between main and feature-branch"}]}`)),
			}, nil
		},
	}

	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: &RESTClient{client: restClient},
		gqlClient:  nil,
	}

	err := client.CreatePR(PRInput{
		Title: "Test PR",
		Body:  "This is a test PR",
		Head:  "feature-branch",
		Base:  "main",
	})
	if err == nil {
		t.Error("Expected an error for validation failure")
	}
}

func TestSetLogger(t *testing.T) {
	client := NewGHClient("testowner", "testrepo")
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
	client := NewGHClient("testowner", "testrepo")
	mockLogger := &MockLogger{}
	client.SetLogger(mockLogger)

	client.debugLog("test message: %s", "value")

	if mockLogger.lastMessage != "test message: value" {
		t.Errorf("Expected 'test message: value', got '%s'", mockLogger.lastMessage)
	}
}

func TestCreateLabelError(t *testing.T) {
	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: nil, // This will cause an error
		gqlClient:  nil,
	}

	err := client.CreateLabel("test-label")
	if err == nil {
		t.Error("Expected an error when REST client is nil")
	}
}

func TestCreateIssueError(t *testing.T) {
	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: nil, // This will cause an error
		gqlClient:  nil,
	}

	err := client.CreateIssue(IssueInput{Title: "Test", Body: "Test"})
	if err == nil {
		t.Error("Expected an error when REST client is nil")
	}
}

func TestCreatePRError(t *testing.T) {
	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: nil, // This will cause an error
		gqlClient:  nil,
	}

	err := client.CreatePR(PRInput{Title: "Test", Head: "feature", Base: "main"})
	if err == nil {
		t.Error("Expected an error when REST client is nil")
	}
}
