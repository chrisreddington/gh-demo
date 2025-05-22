package hydrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
)

// MockGitHubClient is a mock implementation of the GitHubClient interface
type MockGitHubClient struct {
	CreateIssueFunc            func(ctx context.Context, owner, repo string, input *githubapi.IssueInput) (string, error)
	CreateDiscussionFunc       func(ctx context.Context, input *githubapi.DiscussionInput) (string, error)
	GetDiscussionCategoriesFunc func(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error)
	CreatePullRequestFunc      func(ctx context.Context, owner, repo string, input *githubapi.PullRequestInput) (string, error)
	CreateLabelFunc            func(ctx context.Context, owner, repo string, input *githubapi.LabelInput) (string, error)
	GetRepositoryIDFunc        func(ctx context.Context, owner, repo string) (string, error)
}

func (m *MockGitHubClient) CreateIssue(ctx context.Context, owner, repo string, input *githubapi.IssueInput) (string, error) {
	return m.CreateIssueFunc(ctx, owner, repo, input)
}

func (m *MockGitHubClient) CreateDiscussion(ctx context.Context, input *githubapi.DiscussionInput) (string, error) {
	return m.CreateDiscussionFunc(ctx, input)
}

func (m *MockGitHubClient) GetDiscussionCategories(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error) {
	return m.GetDiscussionCategoriesFunc(ctx, owner, repo)
}

func (m *MockGitHubClient) CreatePullRequest(ctx context.Context, owner, repo string, input *githubapi.PullRequestInput) (string, error) {
	return m.CreatePullRequestFunc(ctx, owner, repo, input)
}

func (m *MockGitHubClient) CreateLabel(ctx context.Context, owner, repo string, input *githubapi.LabelInput) (string, error) {
	return m.CreateLabelFunc(ctx, owner, repo, input)
}

func (m *MockGitHubClient) GetRepositoryID(ctx context.Context, owner, repo string) (string, error) {
	return m.GetRepositoryIDFunc(ctx, owner, repo)
}

// TestHydrator_ProcessLabels tests the processLabels method
func TestHydrator_ProcessLabels(t *testing.T) {
	// Create a mock client that counts label creations
	labelCreationCount := 0
	mockClient := &MockGitHubClient{
		CreateLabelFunc: func(ctx context.Context, owner, repo string, input *githubapi.LabelInput) (string, error) {
			labelCreationCount++
			return input.Name, nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test labels
	labels := []Label{
		{Name: "bug", Color: "ff0000", Description: "This is a bug"},
		{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
	}

	// Process the labels
	err := hydrator.processLabels(context.Background(), "testowner", "testrepo", labels)
	if err != nil {
		t.Errorf("processLabels returned an error: %v", err)
	}

	// Check that the correct number of labels were created
	if labelCreationCount != 2 {
		t.Errorf("Expected 2 label creations, got %d", labelCreationCount)
	}
}

// TestHydrator_ProcessLabels_Error tests error handling in the processLabels method
func TestHydrator_ProcessLabels_Error(t *testing.T) {
	// Create a mock client that returns an error for the second label
	mockClient := &MockGitHubClient{
		CreateLabelFunc: func(ctx context.Context, owner, repo string, input *githubapi.LabelInput) (string, error) {
			if input.Name == "enhancement" {
				return "", fmt.Errorf("API error")
			}
			return input.Name, nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test labels
	labels := []Label{
		{Name: "bug", Color: "ff0000", Description: "This is a bug"},
		{Name: "enhancement", Color: "00ff00", Description: "This is an enhancement"},
	}

	// Process the labels - should continue despite error
	err := hydrator.processLabels(context.Background(), "testowner", "testrepo", labels)
	if err != nil {
		t.Errorf("processLabels should not return an error when CreateLabel fails: %v", err)
	}
}

// TestHydrator_ProcessIssues tests the processIssues method
func TestHydrator_ProcessIssues(t *testing.T) {
	// Create a mock client that counts issue creations
	issueCreationCount := 0
	mockClient := &MockGitHubClient{
		CreateIssueFunc: func(ctx context.Context, owner, repo string, input *githubapi.IssueInput) (string, error) {
			issueCreationCount++
			return "https://github.com/testowner/testrepo/issues/1", nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test issues
	issues := []Issue{
		{Title: "Issue 1", Body: "This is issue 1", Labels: []string{"bug"}},
		{Title: "Issue 2", Body: "This is issue 2", Assignees: []string{"testuser"}},
	}

	// Process the issues
	err := hydrator.processIssues(context.Background(), "testowner", "testrepo", issues)
	if err != nil {
		t.Errorf("processIssues returned an error: %v", err)
	}

	// Check that the correct number of issues were created
	if issueCreationCount != 2 {
		t.Errorf("Expected 2 issue creations, got %d", issueCreationCount)
	}
}

// TestHydrator_ProcessIssues_Error tests error handling in the processIssues method
func TestHydrator_ProcessIssues_Error(t *testing.T) {
	// Create a mock client that returns an error
	mockClient := &MockGitHubClient{
		CreateIssueFunc: func(ctx context.Context, owner, repo string, input *githubapi.IssueInput) (string, error) {
			return "", fmt.Errorf("API error")
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test issues
	issues := []Issue{
		{Title: "Issue 1", Body: "This is issue 1"},
	}

	// Process the issues
	err := hydrator.processIssues(context.Background(), "testowner", "testrepo", issues)
	if err == nil {
		t.Error("processIssues should return an error when CreateIssue fails")
	}
}

// TestHydrator_ProcessDiscussions tests the processDiscussions method
func TestHydrator_ProcessDiscussions(t *testing.T) {
	// Create a mock client
	discussionCreationCount := 0
	mockClient := &MockGitHubClient{
		GetRepositoryIDFunc: func(ctx context.Context, owner, repo string) (string, error) {
			return "R_123456", nil
		},
		GetDiscussionCategoriesFunc: func(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error) {
			return []githubapi.DiscussionCategory{
				{ID: "CAT_1", Name: "General"},
				{ID: "CAT_2", Name: "Ideas"},
			}, nil
		},
		CreateDiscussionFunc: func(ctx context.Context, input *githubapi.DiscussionInput) (string, error) {
			discussionCreationCount++
			return "https://github.com/testowner/testrepo/discussions/1", nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test discussions
	discussions := []Discussion{
		{Title: "Discussion 1", Body: "This is discussion 1", Category: "General"},
		{Title: "Discussion 2", Body: "This is discussion 2", Category: "Ideas"},
	}

	// Process the discussions
	err := hydrator.processDiscussions(context.Background(), "testowner", "testrepo", discussions)
	if err != nil {
		t.Errorf("processDiscussions returned an error: %v", err)
	}

	// Check that the correct number of discussions were created
	if discussionCreationCount != 2 {
		t.Errorf("Expected 2 discussion creations, got %d", discussionCreationCount)
	}
}

// TestHydrator_ProcessDiscussions_GetRepoIDError tests error handling when GetRepositoryID fails
func TestHydrator_ProcessDiscussions_GetRepoIDError(t *testing.T) {
	// Create a mock client that returns an error for GetRepositoryID
	mockClient := &MockGitHubClient{
		GetRepositoryIDFunc: func(ctx context.Context, owner, repo string) (string, error) {
			return "", fmt.Errorf("API error")
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test discussions
	discussions := []Discussion{
		{Title: "Discussion 1", Body: "This is discussion 1", Category: "General"},
	}

	// Process the discussions
	err := hydrator.processDiscussions(context.Background(), "testowner", "testrepo", discussions)
	if err == nil {
		t.Error("processDiscussions should return an error when GetRepositoryID fails")
	}
}

// TestHydrator_ProcessDiscussions_GetCategoriesError tests error handling when GetDiscussionCategories fails
func TestHydrator_ProcessDiscussions_GetCategoriesError(t *testing.T) {
	// Create a mock client that returns an error for GetDiscussionCategories
	mockClient := &MockGitHubClient{
		GetRepositoryIDFunc: func(ctx context.Context, owner, repo string) (string, error) {
			return "R_123456", nil
		},
		GetDiscussionCategoriesFunc: func(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test discussions
	discussions := []Discussion{
		{Title: "Discussion 1", Body: "This is discussion 1", Category: "General"},
	}

	// Process the discussions
	err := hydrator.processDiscussions(context.Background(), "testowner", "testrepo", discussions)
	if err == nil {
		t.Error("processDiscussions should return an error when GetDiscussionCategories fails")
	}
}

// TestHydrator_ProcessDiscussions_CategoryNotFound tests error handling when a category is not found
func TestHydrator_ProcessDiscussions_CategoryNotFound(t *testing.T) {
	// Create a mock client that doesn't include the requested category
	mockClient := &MockGitHubClient{
		GetRepositoryIDFunc: func(ctx context.Context, owner, repo string) (string, error) {
			return "R_123456", nil
		},
		GetDiscussionCategoriesFunc: func(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error) {
			return []githubapi.DiscussionCategory{
				{ID: "CAT_1", Name: "General"},
			}, nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test discussions with a non-existent category
	discussions := []Discussion{
		{Title: "Discussion 1", Body: "This is discussion 1", Category: "NonExistent"},
	}

	// Process the discussions
	err := hydrator.processDiscussions(context.Background(), "testowner", "testrepo", discussions)
	if err == nil {
		t.Error("processDiscussions should return an error when category is not found")
	}
}

// TestHydrator_ProcessDiscussions_CreateError tests error handling when CreateDiscussion fails
func TestHydrator_ProcessDiscussions_CreateError(t *testing.T) {
	// Create a mock client that returns an error for CreateDiscussion
	mockClient := &MockGitHubClient{
		GetRepositoryIDFunc: func(ctx context.Context, owner, repo string) (string, error) {
			return "R_123456", nil
		},
		GetDiscussionCategoriesFunc: func(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error) {
			return []githubapi.DiscussionCategory{
				{ID: "CAT_1", Name: "General"},
			}, nil
		},
		CreateDiscussionFunc: func(ctx context.Context, input *githubapi.DiscussionInput) (string, error) {
			return "", fmt.Errorf("API error")
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test discussions
	discussions := []Discussion{
		{Title: "Discussion 1", Body: "This is discussion 1", Category: "General"},
	}

	// Process the discussions
	err := hydrator.processDiscussions(context.Background(), "testowner", "testrepo", discussions)
	if err == nil {
		t.Error("processDiscussions should return an error when CreateDiscussion fails")
	}
}

// TestHydrator_ProcessPullRequests tests the processPullRequests method
func TestHydrator_ProcessPullRequests(t *testing.T) {
	// Create a mock client that counts PR creations
	prCreationCount := 0
	mockClient := &MockGitHubClient{
		CreatePullRequestFunc: func(ctx context.Context, owner, repo string, input *githubapi.PullRequestInput) (string, error) {
			prCreationCount++
			return "https://github.com/testowner/testrepo/pull/1", nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test PRs
	prs := []PR{
		{Title: "PR 1", Body: "This is PR 1", Base: "main", Head: "feature-1"},
		{Title: "PR 2", Body: "This is PR 2", Base: "main", Head: "feature-2", Draft: true},
	}

	// Process the PRs
	err := hydrator.processPullRequests(context.Background(), "testowner", "testrepo", prs)
	if err != nil {
		t.Errorf("processPullRequests returned an error: %v", err)
	}

	// Check that the correct number of PRs were created
	if prCreationCount != 2 {
		t.Errorf("Expected 2 PR creations, got %d", prCreationCount)
	}
}

// TestHydrator_ProcessPullRequests_Error tests error handling in the processPullRequests method
func TestHydrator_ProcessPullRequests_Error(t *testing.T) {
	// Create a mock client that returns an error
	mockClient := &MockGitHubClient{
		CreatePullRequestFunc: func(ctx context.Context, owner, repo string, input *githubapi.PullRequestInput) (string, error) {
			return "", fmt.Errorf("API error")
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Create test PRs
	prs := []PR{
		{Title: "PR 1", Body: "This is PR 1", Base: "main", Head: "feature-1"},
	}

	// Process the PRs
	err := hydrator.processPullRequests(context.Background(), "testowner", "testrepo", prs)
	if err == nil {
		t.Error("processPullRequests should return an error when CreatePullRequest fails")
	}
}

// TestHydrator_Hydrate tests the Hydrate method
func TestHydrator_Hydrate(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test configuration file
	configContent := `{
		"labels": [
			{
				"name": "bug",
				"color": "ff0000",
				"description": "This is a bug"
			}
		],
		"issues": [
			{
				"title": "Test Issue",
				"body": "This is a test issue",
				"labels": ["bug"]
			}
		]
	}`
	configPath := filepath.Join(tempDir, "hydrate.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create a mock client that counts API calls
	labelCreationCount := 0
	issueCreationCount := 0
	mockClient := &MockGitHubClient{
		CreateLabelFunc: func(ctx context.Context, owner, repo string, input *githubapi.LabelInput) (string, error) {
			labelCreationCount++
			return input.Name, nil
		},
		CreateIssueFunc: func(ctx context.Context, owner, repo string, input *githubapi.IssueInput) (string, error) {
			issueCreationCount++
			return "https://github.com/testowner/testrepo/issues/1", nil
		},
		GetRepositoryIDFunc: func(ctx context.Context, owner, repo string) (string, error) {
			return "R_123456", nil
		},
		GetDiscussionCategoriesFunc: func(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error) {
			return []githubapi.DiscussionCategory{
				{ID: "DC_123456", Name: "General"},
			}, nil
		},
		CreateDiscussionFunc: func(ctx context.Context, input *githubapi.DiscussionInput) (string, error) {
			return "https://github.com/testowner/testrepo/discussions/1", nil
		},
		CreatePullRequestFunc: func(ctx context.Context, owner, repo string, input *githubapi.PullRequestInput) (string, error) {
			return "https://github.com/testowner/testrepo/pull/1", nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Run the hydration process
	err = hydrator.Hydrate(context.Background(), configPath, "testowner", "testrepo")
	if err != nil {
		t.Errorf("Hydrate returned an error: %v", err)
	}

	// Check that the correct number of resources were created
	if labelCreationCount != 1 {
		t.Errorf("Expected 1 label creation, got %d", labelCreationCount)
	}
	if issueCreationCount != 1 {
		t.Errorf("Expected 1 issue creation, got %d", issueCreationCount)
	}
}

// TestHydrator_Hydrate_CompleteConfig tests the Hydrate method with all resource types
func TestHydrator_Hydrate_CompleteConfig(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test configuration file with all resource types
	configContent := `{
		"labels": [
			{
				"name": "bug",
				"color": "ff0000",
				"description": "This is a bug"
			}
		],
		"issues": [
			{
				"title": "Test Issue",
				"body": "This is a test issue",
				"labels": ["bug"]
			}
		],
		"discussions": [
			{
				"title": "Test Discussion",
				"body": "This is a test discussion",
				"category": "General"
			}
		],
		"prs": [
			{
				"title": "Test PR",
				"body": "This is a test PR",
				"base": "main",
				"head": "feature"
			}
		]
	}`
	configPath := filepath.Join(tempDir, "hydrate.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Track creation counts
	labelCount := 0
	issueCount := 0
	discussionCount := 0
	prCount := 0

	// Create a mock client that tracks all API calls
	mockClient := &MockGitHubClient{
		CreateLabelFunc: func(ctx context.Context, owner, repo string, input *githubapi.LabelInput) (string, error) {
			labelCount++
			return input.Name, nil
		},
		CreateIssueFunc: func(ctx context.Context, owner, repo string, input *githubapi.IssueInput) (string, error) {
			issueCount++
			return "https://github.com/testowner/testrepo/issues/1", nil
		},
		GetRepositoryIDFunc: func(ctx context.Context, owner, repo string) (string, error) {
			return "R_123456", nil
		},
		GetDiscussionCategoriesFunc: func(ctx context.Context, owner, repo string) ([]githubapi.DiscussionCategory, error) {
			return []githubapi.DiscussionCategory{
				{ID: "DC_123456", Name: "General"},
			}, nil
		},
		CreateDiscussionFunc: func(ctx context.Context, input *githubapi.DiscussionInput) (string, error) {
			discussionCount++
			return "https://github.com/testowner/testrepo/discussions/1", nil
		},
		CreatePullRequestFunc: func(ctx context.Context, owner, repo string, input *githubapi.PullRequestInput) (string, error) {
			prCount++
			return "https://github.com/testowner/testrepo/pull/1", nil
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Run the hydration process
	err = hydrator.Hydrate(context.Background(), configPath, "testowner", "testrepo")
	if err != nil {
		t.Errorf("Hydrate returned an error: %v", err)
	}

	// Check that all resource types were created
	if labelCount != 1 {
		t.Errorf("Expected 1 label creation, got %d", labelCount)
	}
	if issueCount != 1 {
		t.Errorf("Expected 1 issue creation, got %d", issueCount)
	}
	if discussionCount != 1 {
		t.Errorf("Expected 1 discussion creation, got %d", discussionCount)
	}
	if prCount != 1 {
		t.Errorf("Expected 1 PR creation, got %d", prCount)
	}
}

// TestHydrator_Hydrate_FileNotFound tests error handling when the config file is not found
func TestHydrator_Hydrate_FileNotFound(t *testing.T) {
	// Create a mock client
	mockClient := &MockGitHubClient{}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Run the hydration process with a non-existent file
	err := hydrator.Hydrate(context.Background(), "/path/to/nonexistent/file", "testowner", "testrepo")
	if err == nil {
		t.Error("Hydrate should return an error when the config file is not found")
	}
}

// TestHydrator_Hydrate_InvalidJSON tests error handling when the config file contains invalid JSON
func TestHydrator_Hydrate_InvalidJSON(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an invalid JSON file
	invalidContent := `{ "labels": [ { "name": "bug" }`
	configPath := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create a mock client
	mockClient := &MockGitHubClient{}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Run the hydration process with the invalid JSON file
	err = hydrator.Hydrate(context.Background(), configPath, "testowner", "testrepo")
	if err == nil {
		t.Error("Hydrate should return an error when the config file contains invalid JSON")
	}
}

// TestHydrator_Hydrate_ProcessError tests error handling when a process function returns an error
func TestHydrator_Hydrate_ProcessError(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a config file with issues
	configContent := `{
		"issues": [
			{
				"title": "Test Issue",
				"body": "This is a test issue"
			}
		]
	}`
	configPath := filepath.Join(tempDir, "hydrate.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create a mock client that returns an error for CreateIssue
	mockClient := &MockGitHubClient{
		CreateIssueFunc: func(ctx context.Context, owner, repo string, input *githubapi.IssueInput) (string, error) {
			return "", errors.New("API error")
		},
	}

	// Create a hydrator with the mock client
	hydrator := NewHydrator(mockClient)

	// Run the hydration process
	err = hydrator.Hydrate(context.Background(), configPath, "testowner", "testrepo")
	if err == nil {
		t.Error("Hydrate should return an error when a process function returns an error")
	}
}

// TestFindConfigFile tests the findConfigFile function
func TestFindConfigFile(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "hydrate-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "hydrate.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test with direct file path
	foundPath, err := findConfigFile(configPath)
	if err != nil {
		t.Errorf("findConfigFile returned an error with direct path: %v", err)
	}
	if foundPath != configPath {
		t.Errorf("Expected found path to be '%s', got '%s'", configPath, foundPath)
	}

	// Test with directory path
	foundPath, err = findConfigFile(tempDir)
	if err != nil {
		t.Errorf("findConfigFile returned an error with directory path: %v", err)
	}
	if foundPath != configPath {
		t.Errorf("Expected found path to be '%s', got '%s'", configPath, foundPath)
	}
}