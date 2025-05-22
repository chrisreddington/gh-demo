package githubapi

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockClient implements the interface required by GQLClient
type MockGQLClient struct {
	QueryFunc   func(string, interface{}, map[string]interface{}) error
	MutateFunc  func(string, interface{}, map[string]interface{}) error
}

func (m *MockGQLClient) Query(name string, query interface{}, variables map[string]interface{}) error {
	if m.QueryFunc != nil {
		return m.QueryFunc(name, query, variables)
	}
	return nil
}

func (m *MockGQLClient) Mutate(name string, query interface{}, variables map[string]interface{}) error {
	if m.MutateFunc != nil {
		return m.MutateFunc(name, query, variables)
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
	gqlClient := &MockGQLClient{
		QueryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return nil
		},
		MutateFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return nil
		},
	}

	restClient := &MockRESTClient{
		RequestFunc: func(method string, path string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		},
	}

	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: &RESTClient{client: restClient},
		gqlClient:  &GQLClient{client: gqlClient},
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

func TestListLabels(t *testing.T) {
	mockGQLClient := &MockGQLClient{
		QueryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			// Type assertion to get the query struct
			q, ok := query.(*struct {
				Repository struct {
					Labels struct {
						Nodes []struct {
							Name string
						}
						PageInfo struct {
							HasNextPage bool
							EndCursor   string
						}
					} `graphql:"labels(first: 100)"`
				} `graphql:"repository(owner: $owner, name: $name)"`
			})
			if !ok {
				t.Fatalf("query type assertion failed")
			}

			// Populate the query with test data
			q.Repository.Labels.Nodes = []struct{ Name string }{
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
		gqlClient: &GQLClient{client: mockGQLClient},
	}

	labels, err := client.ListLabels()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}

	expectedLabels := []string{"bug", "enhancement", "documentation"}
	for i, label := range labels {
		if label != expectedLabels[i] {
			t.Errorf("Expected label %s, got %s", expectedLabels[i], label)
		}
	}
}

func TestCreateDiscussion(t *testing.T) {
	mockGQLClient := &MockGQLClient{
		QueryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			// Type assertion to get the query struct
			q, ok := query.(*struct {
				Repository struct {
					ID         string
					Categories struct {
						Nodes []struct {
							ID   string
							Name string
						}
					} `graphql:"discussionCategories(first: 50)"`
				} `graphql:"repository(owner: $owner, name: $name)"`
			})
			if !ok {
				t.Fatalf("query type assertion failed")
			}

			// Populate the query with test data
			q.Repository.ID = "R_1234"
			q.Repository.Categories.Nodes = []struct {
				ID   string
				Name string
			}{
				{ID: "C_1", Name: "General"},
				{ID: "C_2", Name: "Ideas"},
				{ID: "C_3", Name: "Q&A"},
			}

			return nil
		},
		MutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			// Type assertion to get the mutation struct
			m, ok := mutation.(*struct {
				CreateDiscussion struct {
					Discussion struct {
						ID string
					}
				} `graphql:"createDiscussion(input: $input)"`
			})
			if !ok {
				t.Fatalf("mutation type assertion failed")
			}

			// Populate the mutation response
			m.CreateDiscussion.Discussion.ID = "D_1234"

			return nil
		},
	}

	client := &GHClient{
		Owner:     "testowner",
		Repo:      "testrepo",
		gqlClient: &GQLClient{client: mockGQLClient},
	}

	err := client.CreateDiscussion(DiscussionInput{
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		Category: "General",
		Labels:   []string{"discussion"},
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRESTClientRequest(t *testing.T) {
	mockRESTClient := &MockRESTClient{
		RequestFunc: func(method, path string, body io.Reader) (*http.Response, error) {
			if method != "POST" {
				t.Errorf("Expected method POST, got %s", method)
			}
			if path != "test/path" {
				t.Errorf("Expected path test/path, got %s", path)
			}
			return &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		},
	}

	client := &RESTClient{client: mockRESTClient}
	err := client.Request("POST", "test/path", map[string]string{"key": "value"}, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreatePR(t *testing.T) {
	// First test: successful creation with assignees and labels
	mockRESTClient := &MockRESTClient{
		RequestFunc: func(method, path string, body io.Reader) (*http.Response, error) {
			if path == "repos/testowner/testrepo/pulls" {
				return &http.Response{
					StatusCode: 201,
					Body:       io.NopCloser(strings.NewReader(`{"number": 42}`)),
				}, nil
			} else if path == "repos/testowner/testrepo/issues/42" {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{}`)),
				}, nil
			}
			return &http.Response{StatusCode: 404}, nil
		},
	}

	client := &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: &RESTClient{client: mockRESTClient},
	}

	err := client.CreatePR(PRInput{
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

	// Second test: PR creation without labels/assignees
	mockRESTClient = &MockRESTClient{
		RequestFunc: func(method, path string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		},
	}

	client = &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: &RESTClient{client: mockRESTClient},
	}

	err = client.CreatePR(PRInput{
		Title: "Simple PR",
		Body:  "Simple PR without labels or assignees",
		Head:  "feature-branch",
		Base:  "main",
	})

	if err != nil {
		t.Errorf("Expected no error for simple PR, got %v", err)
	}

	// Third test: error case
	mockRESTClient = &MockRESTClient{
		RequestFunc: func(method, path string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: 422,
				Body:       io.NopCloser(strings.NewReader(`{"message": "Validation failed"}`)),
			}, nil
		},
	}

	client = &GHClient{
		Owner:      "testowner",
		Repo:       "testrepo",
		restClient: &RESTClient{client: mockRESTClient},
	}

	err = client.CreatePR(PRInput{
		Title: "Invalid PR",
		Body:  "This PR should fail",
		Head:  "feature-branch",
		Base:  "main",
	})

	if err == nil {
		t.Error("Expected error for invalid PR, got nil")
	}
}