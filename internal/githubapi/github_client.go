package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// GitHubClientImpl implements the GitHubClient interface
type GitHubClientImpl struct {
	client interface {
		Do(string, map[string]interface{}, interface{}) error
	}
	rest interface {
		Get(path string, response interface{}) error
		Post(path string, body io.Reader, response interface{}) error
		Patch(path string, body io.Reader, response interface{}) error
		Delete(path string, response interface{}) error
	}
}

// NewGitHubClient creates a new GitHubClient using the go-gh library
func NewGitHubClient() (*GitHubClientImpl, error) {
	// Initialize GraphQL client for GitHub API
	gqlClient, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Initialize REST client for GitHub API
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w", err)
	}

	return &GitHubClientImpl{
		client: gqlClient,
		rest:   restClient,
	}, nil
}

// CreateIssue creates a new issue in the specified repository
func (c *GitHubClientImpl) CreateIssue(ctx context.Context, owner, repo string, input *IssueInput) (string, error) {
	// Build the GraphQL mutation
	mutation := `
	mutation CreateIssue($input: CreateIssueInput!) {
		createIssue(input: $input) {
			issue {
				id
				number
				url
			}
		}
	}
	`

	// Create the variables for the mutation
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": "",
			"title":        input.Title,
			"body":         input.Body,
			"labelIds":     []string{},
		},
	}

	// Get the repository ID
	repoID, err := c.GetRepositoryID(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get repository ID: %w", err)
	}
	variables["input"].(map[string]interface{})["repositoryId"] = repoID

	// Process labels if provided
	if len(input.Labels) > 0 {
		// Convert label names to IDs
		labelIDs := []string{}
		// For simplicity in this implementation, we're not converting label names to IDs
		// In a full implementation, we would need to query for label IDs
		variables["input"].(map[string]interface{})["labelIds"] = labelIDs
	}

	// Process assignees if provided
	if len(input.Assignees) > 0 {
		assigneeIDs := []string{}
		// For simplicity, we're not converting assignee usernames to IDs
		// In a full implementation, we would need to query for user IDs
		variables["input"].(map[string]interface{})["assigneeIds"] = assigneeIDs
	}

	// Execute the mutation
	var response struct {
		CreateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				URL    string `json:"url"`
			} `json:"issue"`
		} `json:"createIssue"`
	}

	err = c.client.Do(mutation, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to create issue: %w", err)
	}

	return response.CreateIssue.Issue.URL, nil
}

// CreateDiscussion creates a new discussion in the specified repository
func (c *GitHubClientImpl) CreateDiscussion(ctx context.Context, input *DiscussionInput) (string, error) {
	// Build the GraphQL mutation
	mutation := `
	mutation CreateDiscussion($input: CreateDiscussionInput!) {
		createDiscussion(input: $input) {
			discussion {
				id
				url
			}
		}
	}
	`

	// Create the variables for the mutation
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": input.RepositoryID,
			"categoryId":   input.CategoryID,
			"title":        input.Title,
			"body":         input.Body,
		},
	}

	// Execute the mutation
	var response struct {
		CreateDiscussion struct {
			Discussion struct {
				ID  string `json:"id"`
				URL string `json:"url"`
			} `json:"discussion"`
		} `json:"createDiscussion"`
	}

	err := c.client.Do(mutation, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to create discussion: %w", err)
	}

	return response.CreateDiscussion.Discussion.URL, nil
}

// GetDiscussionCategories retrieves the available discussion categories in a repository
func (c *GitHubClientImpl) GetDiscussionCategories(ctx context.Context, owner, repo string) ([]DiscussionCategory, error) {
	// Build the GraphQL query
	query := `
	query GetDiscussionCategories($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			discussionCategories(first: 10) {
				nodes {
					id
					name
				}
			}
		}
	}
	`

	// Create the variables for the query
	variables := map[string]interface{}{
		"owner": owner,
		"name":  repo,
	}

	// Execute the query
	var response struct {
		Repository struct {
			DiscussionCategories struct {
				Nodes []DiscussionCategory `json:"nodes"`
			} `json:"discussionCategories"`
		} `json:"repository"`
	}

	err := c.client.Do(query, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get discussion categories: %w", err)
	}

	return response.Repository.DiscussionCategories.Nodes, nil
}

// CreatePullRequest creates a new pull request in the specified repository
func (c *GitHubClientImpl) CreatePullRequest(ctx context.Context, owner, repo string, input *PullRequestInput) (string, error) {
	// Build the GraphQL mutation
	mutation := `
	mutation CreatePullRequest($input: CreatePullRequestInput!) {
		createPullRequest(input: $input) {
			pullRequest {
				id
				number
				url
			}
		}
	}
	`

	// Create the variables for the mutation
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": "",
			"baseRefName":  input.Base,
			"headRefName":  input.Head,
			"title":        input.Title,
			"body":         input.Body,
			"draft":        input.Draft,
		},
	}

	// Get the repository ID
	repoID, err := c.GetRepositoryID(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get repository ID: %w", err)
	}
	variables["input"].(map[string]interface{})["repositoryId"] = repoID

	// Execute the mutation
	var response struct {
		CreatePullRequest struct {
			PullRequest struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				URL    string `json:"url"`
			} `json:"pullRequest"`
		} `json:"createPullRequest"`
	}

	err = c.client.Do(mutation, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	return response.CreatePullRequest.PullRequest.URL, nil
}

// CreateLabel creates a new label in the specified repository
func (c *GitHubClientImpl) CreateLabel(ctx context.Context, owner, repo string, input *LabelInput) (string, error) {
	// Use REST API for label creation
	path := fmt.Sprintf("repos/%s/%s/labels", owner, repo)
	payload := map[string]string{
		"name":        input.Name,
		"color":       input.Color,
		"description": input.Description,
	}

	// Convert payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal label payload: %w", err)
	}

	// Execute the POST request
	var response map[string]interface{}
	err = c.rest.Post(path, strings.NewReader(string(body)), &response)
	if err != nil {
		return "", fmt.Errorf("failed to create label: %w", err)
	}

	// Extract label name from response
	if name, ok := response["name"].(string); ok {
		return name, nil
	}

	return "", fmt.Errorf("unexpected response format when creating label")
}

// GetRepositoryID retrieves the ID of a repository
func (c *GitHubClientImpl) GetRepositoryID(ctx context.Context, owner, repo string) (string, error) {
	// Build the GraphQL query
	query := `
	query GetRepositoryID($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			id
		}
	}
	`

	// Create the variables for the query
	variables := map[string]interface{}{
		"owner": owner,
		"name":  repo,
	}

	// Execute the query
	var response struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}

	err := c.client.Do(query, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get repository ID: %w", err)
	}

	return response.Repository.ID, nil
}
