/*
Package githubapi provides GitHub API client functionality for the gh-demo CLI extension.

This package implements two patterns for client creation:
1. NewGHClient() - Creates a real GitHub client using go-gh authentication
2. NewGHClientWithClients() - Accepts injected clients for testing with mocks

The dependency injection pattern allows for comprehensive unit testing without requiring
GitHub authentication, while still maintaining integration test capabilities when
credentials are available.

Testing Strategy:
- Unit tests use NewGHClientWithClients() with mock clients
- Integration tests use NewGHClient() and skip when authentication is unavailable
- CI runs tests in short mode to skip integration tests by default
*/

package githubapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/errors"
	"github.com/chrisreddington/gh-demo/internal/types"
	"github.com/cli/go-gh/v2/pkg/api"
)

// GraphQLClient interface for testability
type GraphQLClient interface {
	Do(query string, variables map[string]interface{}, response interface{}) error
}

// GHClient is the main client for all GitHub API operations
type GHClient struct {
	Owner      string
	Repo       string
	gqlClient  GraphQLClient
	restClient *RESTClient
	logger     common.Logger
}

// RESTClient wraps the REST client for testability
type RESTClient struct {
	client interface {
		Request(string, string, io.Reader) (*http.Response, error)
	}
}

// NewGHClient creates a new GitHub API client for the specified owner and repository.
// It initializes both GraphQL and REST clients using the go-gh library and validates that
// the owner and repo parameters are not empty. The client is ready to perform operations
// like creating issues, discussions, pull requests, and managing labels.
func NewGHClient(owner, repo string) (*GHClient, error) {
	if strings.TrimSpace(owner) == "" {
		return nil, errors.ValidationError("validate_client_params", "owner cannot be empty")
	}
	if strings.TrimSpace(repo) == "" {
		return nil, errors.ValidationError("validate_client_params", "repo cannot be empty")
	}

	// Create GraphQL client using go-gh
	gqlClient, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, errors.APIError("create_graphql_client", "failed to initialize GraphQL client", err)
	}

	// Create REST client using go-gh
	restRawClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, errors.APIError("create_rest_client", "failed to initialize REST client", err)
	}

	restClient := &RESTClient{client: restRawClient}

	return &GHClient{
		Owner:      strings.TrimSpace(owner),
		Repo:       strings.TrimSpace(repo),
		gqlClient:  gqlClient,
		restClient: restClient,
		logger:     nil, // Will be set when SetLogger is called
	}, nil
}

// NewGHClientWithClients creates a new GitHub API client with provided clients for testing.
// This constructor allows dependency injection of mock clients for unit testing while
// maintaining the same validation and initialization logic as NewGHClient.
func NewGHClientWithClients(owner, repo string, gqlClient GraphQLClient, restClient *RESTClient) (*GHClient, error) {
	if strings.TrimSpace(owner) == "" {
		return nil, errors.ValidationError("validate_client_params", "owner cannot be empty")
	}
	if strings.TrimSpace(repo) == "" {
		return nil, errors.ValidationError("validate_client_params", "repo cannot be empty")
	}

	return &GHClient{
		Owner:      strings.TrimSpace(owner),
		Repo:       strings.TrimSpace(repo),
		gqlClient:  gqlClient,
		restClient: restClient,
		logger:     nil, // Will be set when SetLogger is called
	}, nil
}

// SetLogger sets the logger for debug output
func (c *GHClient) SetLogger(logger common.Logger) {
	c.logger = logger
}

// debugLog logs a debug message if logger is available
func (c *GHClient) debugLog(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.Debug(format, args...)
	}
}

// Request makes an HTTP request to the REST API
func (c *RESTClient) Request(method string, path string, body interface{}, response interface{}) error {
	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	resp, err := c.client.Request(method, path, requestBody)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't fail the operation
			// In a real application, you'd use a proper logger here
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if response != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return json.NewDecoder(resp.Body).Decode(response)
	}

	if resp.StatusCode >= 400 {
		// Try to read the error response body for more details
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr == nil && len(bodyBytes) > 0 {
			// Try to parse as GitHub API error format
			var apiError struct {
				Message string `json:"message"`
				Errors  []struct {
					Field   string `json:"field"`
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"errors"`
			}
			if jsonErr := json.Unmarshal(bodyBytes, &apiError); jsonErr == nil && apiError.Message != "" {
				if len(apiError.Errors) > 0 {
					return errors.APIError("http_request", fmt.Sprintf("HTTP %d: %s - %s", resp.StatusCode, apiError.Message, apiError.Errors[0].Message), nil)
				}
				return errors.APIError("http_request", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, apiError.Message), nil)
			}
			// If not parseable as JSON, return raw body
			return errors.APIError("http_request", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)), nil)
		}
		return errors.APIError("http_request", fmt.Sprintf("API request failed with status %d", resp.StatusCode), nil)
	}

	return nil
}

// Label operations
func (c *GHClient) ListLabels() ([]string, error) {
	if c.gqlClient == nil {
		return nil, errors.ValidationError("validate_client", "GraphQL client is not initialized")
	}

	c.debugLog("Fetching labels from repository %s/%s", c.Owner, c.Repo)

	labelsQuery := `
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				labels(first: 100) {
					nodes {
						name
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`

	var response struct {
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
	}

	variables := map[string]interface{}{
		"owner": c.Owner,
		"name":  c.Repo,
	}

	err := c.gqlClient.Do(labelsQuery, variables, &response)
	if err != nil {
		c.debugLog("Failed to fetch labels: %v", err)
		return nil, errors.APIError("list_labels", "failed to fetch labels", err)
	}

	labels := make([]string, 0, len(response.Repository.Labels.Nodes))
	for _, label := range response.Repository.Labels.Nodes {
		labels = append(labels, label.Name)
	}

	c.debugLog("Successfully fetched %d labels", len(labels))
	return labels, nil
}

// CreateLabel creates a new label in the repository using the provided label data.
// It validates that the REST client is initialized and creates the label with
// the specified name, description, and color.
func (c *GHClient) CreateLabel(label types.Label) error {
	if c.restClient == nil {
		return errors.ValidationError("create_label", "REST client is not initialized")
	}

	c.debugLog("Creating label '%s' (color: %s) in repository %s/%s", label.Name, label.Color, c.Owner, c.Repo)

	// Using the REST API for label creation as it's simpler than GraphQL for this case
	path := fmt.Sprintf("repos/%s/%s/labels", c.Owner, c.Repo)
	payload := map[string]interface{}{
		"name":  label.Name,
		"color": label.Color,
	}

	// Add description if provided
	if label.Description != "" {
		payload["description"] = label.Description
	}

	err := c.restClient.Request("POST", path, payload, nil)
	if err != nil {
		c.debugLog("Failed to create label '%s': %v", label.Name, err)
		layeredErr := errors.NewLayeredError("api", "create_label", "failed to create GitHub label", err)
		return layeredErr.WithContext("name", label.Name).WithContext("color", label.Color)
	}

	c.debugLog("Successfully created label '%s' with color '%s'", label.Name, label.Color)
	return nil
}

// CreateIssue creates a new issue in the repository using the provided issue data.
// It validates that the REST client is initialized and creates the issue with
// the specified title, body, labels, and assignees.
func (c *GHClient) CreateIssue(issue types.Issue) error {
	if c.restClient == nil {
		return errors.ValidationError("create_issue", "REST client is not initialized")
	}

	c.debugLog("Creating issue '%s' in repository %s/%s", issue.Title, c.Owner, c.Repo)

	path := fmt.Sprintf("repos/%s/%s/issues", c.Owner, c.Repo)
	payload := map[string]interface{}{
		"title":     issue.Title,
		"body":      issue.Body,
		"labels":    issue.Labels,
		"assignees": issue.Assignees,
	}

	err := c.restClient.Request("POST", path, payload, nil)
	if err != nil {
		c.debugLog("Failed to create issue '%s': %v", issue.Title, err)
		layeredErr := errors.NewLayeredError("api", "create_issue", "failed to create GitHub issue", err)
		return layeredErr.WithContext("title", issue.Title)
	}

	c.debugLog("Successfully created issue '%s'", issue.Title)
	return nil
}

// CreateDiscussion creates a new discussion in the repository using the provided discussion data.
// It uses GraphQL to create the discussion with the specified title, body, category, and labels.
// The method automatically finds the correct category ID and adds labels after creation.
func (c *GHClient) CreateDiscussion(discussion types.Discussion) error {
	if c.gqlClient == nil {
		return errors.ValidationError("validate_client", "GraphQL client is not initialized")
	}

	c.debugLog("Creating discussion '%s' in repository %s/%s", discussion.Title, c.Owner, c.Repo)

	// First, get the repository ID and discussion categories
	repoQuery := `
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				id
				discussionCategories(first: 50) {
					nodes {
						id
						name
					}
				}
			}
		}
	`

	var repoResponse struct {
		Repository struct {
			ID         string `json:"id"`
			Categories struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"discussionCategories"`
		} `json:"repository"`
	}

	repoVariables := map[string]interface{}{
		"owner": c.Owner,
		"name":  c.Repo,
	}

	err := c.gqlClient.Do(repoQuery, repoVariables, &repoResponse)
	if err != nil {
		c.debugLog("Failed to fetch repository info for discussion: %v", err)
		return errors.APIError("fetch_repository_info", "failed to fetch repository info", err)
	}

	// Get available categories for debugging
	availableCategories := make([]string, 0, len(repoResponse.Repository.Categories.Nodes))
	for _, cat := range repoResponse.Repository.Categories.Nodes {
		availableCategories = append(availableCategories, cat.Name)
	}
	c.debugLog("Available discussion categories: %v", availableCategories)

	// Find the category ID that matches the requested category name
	var categoryID string
	var matchedCategory string
	for _, category := range repoResponse.Repository.Categories.Nodes {
		c.debugLog("Comparing category '%s' with requested '%s'", category.Name, discussion.Category)
		if strings.EqualFold(category.Name, discussion.Category) {
			categoryID = category.ID
			matchedCategory = category.Name
			break
		}
	}

	if categoryID == "" {
		c.debugLog("Discussion category '%s' not found in available categories: %v",
			discussion.Category, availableCategories)
		layeredErr := errors.ValidationError("validate_discussion_category", fmt.Sprintf("discussion category '%s' not found in available categories", discussion.Category))
		return layeredErr.(*errors.LayeredError).WithContext("requested_category", discussion.Category).WithContext("available_categories", fmt.Sprintf("%v", availableCategories))
	}

	c.debugLog("Found matching category ID for '%s': %s (actual: '%s')",
		discussion.Category, categoryID, matchedCategory)

	// Create the discussion
	createMutation := `
		mutation($input: CreateDiscussionInput!) {
			createDiscussion(input: $input) {
				discussion {
					id
					number
					title
					url
				}
			}
		}
	`

	var mutationResponse struct {
		CreateDiscussion struct {
			Discussion struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
				URL    string `json:"url"`
			} `json:"discussion"`
		} `json:"createDiscussion"`
	}

	mutationVariables := map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": repoResponse.Repository.ID,
			"categoryId":   categoryID,
			"title":        discussion.Title,
			"body":         discussion.Body,
		},
	}

	// Debug: Log the exact variables being sent to GitHub
	inputData, _ := json.MarshalIndent(mutationVariables, "", "  ")
	c.debugLog("Mutation input: %s", string(inputData))

	err = c.gqlClient.Do(createMutation, mutationVariables, &mutationResponse)
	if err != nil {
		c.debugLog("Failed to create discussion '%s': %v", discussion.Title, err)
		return errors.APIError("create_discussion", "failed to create discussion", err)
	}

	// Debug: Log what we got back from GitHub
	c.debugLog("GitHub response - Discussion.ID: '%s', Number: %d, Title: '%s', URL: '%s'",
		mutationResponse.CreateDiscussion.Discussion.ID,
		mutationResponse.CreateDiscussion.Discussion.Number,
		mutationResponse.CreateDiscussion.Discussion.Title,
		mutationResponse.CreateDiscussion.Discussion.URL)

	// Verify discussion was created by checking for a valid ID and URL
	if mutationResponse.CreateDiscussion.Discussion.ID == "" {
		c.debugLog("Discussion creation for '%s' failed - no Discussion ID returned", discussion.Title)
		layeredErr := errors.APIError("create_discussion", "discussion creation failed - no Discussion ID returned from GitHub API", nil)
		return layeredErr.(*errors.LayeredError).WithContext("title", discussion.Title)
	}

	discussionID := mutationResponse.CreateDiscussion.Discussion.ID
	discussionURL := mutationResponse.CreateDiscussion.Discussion.URL
	c.debugLog("Discussion created with ID: %s, URL: %s", discussionID, discussionURL)

	// Add labels if specified
	if len(discussion.Labels) > 0 && mutationResponse.CreateDiscussion.Discussion.ID != "" {
		c.debugLog("Adding %d labels to discussion '%s'", len(discussion.Labels), discussion.Title)

		// Add labels to the discussion using the discussion ID
		for _, label := range discussion.Labels {
			err := c.addLabelToDiscussion(mutationResponse.CreateDiscussion.Discussion.ID, label)
			if err != nil {
				c.debugLog("Failed to add label '%s' to discussion: %v", label, err)
				// Continue with other labels even if one fails
			} else {
				c.debugLog("Successfully added label '%s' to discussion", label)
			}
		}
	}

	c.debugLog("Successfully created discussion '%s' (URL: %s)", discussion.Title, discussionURL)
	return nil
}

// addLabelToDiscussion is a helper method to add a label to a discussion
func (c *GHClient) addLabelToDiscussion(discussionID, labelName string) error {
	// First, find the label ID for the label name
	labelQuery := `
		query($owner: String!, $repo: String!, $name: String!) {
			repository(owner: $owner, name: $repo) {
				label(name: $name) {
					id
				}
			}
		}
	`

	var labelResponse struct {
		Repository struct {
			Label struct {
				ID string `json:"id"`
			} `json:"label"`
		} `json:"repository"`
	}

	labelVariables := map[string]interface{}{
		"owner": c.Owner,
		"repo":  c.Repo,
		"name":  labelName,
	}

	err := c.gqlClient.Do(labelQuery, labelVariables, &labelResponse)
	if err != nil {
		return errors.APIError("find_label", fmt.Sprintf("failed to find label '%s'", labelName), err)
	}

	if labelResponse.Repository.Label.ID == "" {
		layeredErr := errors.ValidationError("validate_label", fmt.Sprintf("label '%s' not found in repository", labelName))
		return layeredErr.(*errors.LayeredError).WithContext("label_name", labelName)
	}

	// Add the label to the discussion
	addLabelMutation := `
		mutation($input: AddLabelsToLabelableInput!) {
			addLabelsToLabelable(input: $input) {
				clientMutationId
			}
		}
	`

	var labelMutationResponse struct {
		AddLabelsToLabelable struct {
			ClientMutationID string `json:"clientMutationId"`
		} `json:"addLabelsToLabelable"`
	}

	labelMutationVariables := map[string]interface{}{
		"input": map[string]interface{}{
			"labelableId": discussionID,
			"labelIds":    []string{labelResponse.Repository.Label.ID},
		},
	}

	err = c.gqlClient.Do(addLabelMutation, labelMutationVariables, &labelMutationResponse)
	if err != nil {
		return errors.APIError("add_label_to_discussion", fmt.Sprintf("failed to add label '%s' to discussion", labelName), err)
	}

	return nil
}

// CreatePR creates a new pull request in the repository using the provided pull request data.
// It validates the head and base branches, creates the PR via REST API, and adds labels/assignees if specified.
func (c *GHClient) CreatePR(pullRequest types.PullRequest) error {
	if c.restClient == nil {
		return errors.ValidationError("validate_client", "REST client is not initialized")
	}

	c.debugLog("Creating pull request '%s' in repository %s/%s (head: %s, base: %s)", pullRequest.Title, c.Owner, c.Repo, pullRequest.Head, pullRequest.Base)

	// Basic validation
	if pullRequest.Head == "" {
		c.debugLog("PR head branch is empty")
		return errors.ValidationError("validate_pr", "head branch cannot be empty")
	}
	if pullRequest.Base == "" {
		c.debugLog("PR base branch is empty")
		return errors.ValidationError("validate_pr", "base branch cannot be empty")
	}
	if pullRequest.Head == pullRequest.Base {
		c.debugLog("PR head and base branches are the same: %s", pullRequest.Head)
		return errors.ValidationError("validate_pr", fmt.Sprintf("head and base branches cannot be the same (%s)", pullRequest.Head))
	}

	path := fmt.Sprintf("repos/%s/%s/pulls", c.Owner, c.Repo)
	payload := map[string]interface{}{
		"title": pullRequest.Title,
		"body":  pullRequest.Body,
		"head":  pullRequest.Head,
		"base":  pullRequest.Base,
	}

	var response map[string]interface{}
	err := c.restClient.Request("POST", path, payload, &response)
	if err != nil {
		c.debugLog("Failed to create pull request '%s': %v", pullRequest.Title, err)
		layeredErr := errors.APIError("create_pull_request", "failed to create pull request", err)
		return layeredErr.(*errors.LayeredError).WithContext("title", pullRequest.Title).WithContext("head", pullRequest.Head).WithContext("base", pullRequest.Base)
	}

	// If the PR was created successfully and has labels/assignees, add them
	if prNumber, ok := response["number"].(float64); ok && (len(pullRequest.Labels) > 0 || len(pullRequest.Assignees) > 0) {
		c.debugLog("Adding labels/assignees to PR '%s'", pullRequest.Title)
		issuePayload := map[string]interface{}{
			"labels":    pullRequest.Labels,
			"assignees": pullRequest.Assignees,
		}

		issuePath := fmt.Sprintf("repos/%s/%s/issues/%d", c.Owner, c.Repo, int(prNumber))
		if err := c.restClient.Request("PATCH", issuePath, issuePayload, nil); err != nil {
			c.debugLog("Failed to add labels/assignees to PR '%s': %v", pullRequest.Title, err)
			layeredErr := errors.APIError("add_pr_labels_assignees", "created PR but failed to add labels/assignees", err)
			return layeredErr.(*errors.LayeredError).WithContext("title", pullRequest.Title)
		}
	}

	c.debugLog("Successfully created pull request '%s'", pullRequest.Title)
	return nil
}
