package githubapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
	logger     Logger
}

// RESTClient wraps the REST client for testability
type RESTClient struct {
	client interface {
		Request(string, string, io.Reader) (*http.Response, error)
	}
}

func NewGHClient(owner, repo string) *GHClient {
	// Create GraphQL and REST clients using go-gh
	gqlClient, err := api.DefaultGraphQLClient()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize GraphQL client: %v\n", err)
	}

	restRawClient, err := api.DefaultRESTClient()
	var restClient *RESTClient
	if err == nil {
		restClient = &RESTClient{client: restRawClient}
	} else {
		fmt.Printf("Warning: Failed to initialize REST client: %v\n", err)
	}

	return &GHClient{
		Owner:      owner,
		Repo:       repo,
		gqlClient:  gqlClient,
		restClient: restClient,
		logger:     nil, // Will be set when SetLogger is called
	}
}

// SetLogger sets the logger for debug output
func (c *GHClient) SetLogger(logger Logger) {
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
	defer resp.Body.Close()

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
					return fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, apiError.Message, apiError.Errors[0].Message)
				}
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiError.Message)
			}
			// If not parseable as JSON, return raw body
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		}
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

// Label operations
func (c *GHClient) ListLabels() ([]string, error) {
	if c.gqlClient == nil {
		return nil, fmt.Errorf("GraphQL client is not initialized")
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
		return nil, fmt.Errorf("failed to fetch labels: %w", err)
	}

	labels := make([]string, 0, len(response.Repository.Labels.Nodes))
	for _, label := range response.Repository.Labels.Nodes {
		labels = append(labels, label.Name)
	}

	c.debugLog("Successfully fetched %d labels", len(labels))
	return labels, nil
}

func (c *GHClient) CreateLabel(label string) error {
	if c.restClient == nil {
		return fmt.Errorf("REST client is not initialized")
	}

	c.debugLog("Creating label '%s' in repository %s/%s", label, c.Owner, c.Repo)

	// Using the REST API for label creation as it's simpler than GraphQL for this case
	path := fmt.Sprintf("repos/%s/%s/labels", c.Owner, c.Repo)
	payload := map[string]interface{}{
		"name":        label,
		"description": "Label created by gh-demo hydration tool",
		"color":       "ededed", // Light gray color as default
	}

	err := c.restClient.Request("POST", path, payload, nil)
	if err != nil {
		c.debugLog("Failed to create label '%s': %v", label, err)
		return err
	}

	c.debugLog("Successfully created label '%s'", label)
	return nil
}

// Issue operations
func (c *GHClient) CreateIssue(issue IssueInput) error {
	if c.restClient == nil {
		return fmt.Errorf("REST client is not initialized")
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
		return err
	}

	c.debugLog("Successfully created issue '%s'", issue.Title)
	return nil
}

// Discussion operations
func (c *GHClient) CreateDiscussion(disc DiscussionInput) error {
	if c.gqlClient == nil {
		return fmt.Errorf("GraphQL client is not initialized")
	}

	c.debugLog("Creating discussion '%s' in repository %s/%s", disc.Title, c.Owner, c.Repo)

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
		return fmt.Errorf("failed to fetch repository info: %w", err)
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
		c.debugLog("Comparing category '%s' with requested '%s'", category.Name, disc.Category)
		if strings.EqualFold(category.Name, disc.Category) {
			categoryID = category.ID
			matchedCategory = category.Name
			break
		}
	}

	if categoryID == "" {
		c.debugLog("Discussion category '%s' not found in available categories: %v",
			disc.Category, availableCategories)
		return fmt.Errorf("discussion category '%s' not found in available categories: %v",
			disc.Category, availableCategories)
	}

	c.debugLog("Found matching category ID for '%s': %s (actual: '%s')",
		disc.Category, categoryID, matchedCategory)

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
			"title":        disc.Title,
			"body":         disc.Body,
		},
	}

	// Debug: Log the exact variables being sent to GitHub
	inputData, _ := json.MarshalIndent(mutationVariables, "", "  ")
	c.debugLog("Mutation input: %s", string(inputData))

	err = c.gqlClient.Do(createMutation, mutationVariables, &mutationResponse)
	if err != nil {
		c.debugLog("Failed to create discussion '%s': %v", disc.Title, err)
		return fmt.Errorf("failed to create discussion: %w", err)
	}

	// Debug: Log what we got back from GitHub
	c.debugLog("GitHub response - Discussion.ID: '%s', Number: %d, Title: '%s', URL: '%s'",
		mutationResponse.CreateDiscussion.Discussion.ID,
		mutationResponse.CreateDiscussion.Discussion.Number,
		mutationResponse.CreateDiscussion.Discussion.Title,
		mutationResponse.CreateDiscussion.Discussion.URL)

	// Verify discussion was created by checking for a valid ID and URL
	if mutationResponse.CreateDiscussion.Discussion.ID == "" {
		c.debugLog("Discussion creation for '%s' failed - no Discussion ID returned", disc.Title)
		return fmt.Errorf("discussion creation for '%s' failed - no Discussion ID returned from GitHub API", disc.Title)
	}

	discussionID := mutationResponse.CreateDiscussion.Discussion.ID
	discussionURL := mutationResponse.CreateDiscussion.Discussion.URL
	c.debugLog("Discussion created with ID: %s, URL: %s", discussionID, discussionURL)

	// Add labels if specified
	if len(disc.Labels) > 0 && mutationResponse.CreateDiscussion.Discussion.ID != "" {
		c.debugLog("Adding %d labels to discussion '%s'", len(disc.Labels), disc.Title)

		// Add labels to the discussion using the discussion ID
		for _, label := range disc.Labels {
			err := c.addLabelToDiscussion(mutationResponse.CreateDiscussion.Discussion.ID, label)
			if err != nil {
				c.debugLog("Failed to add label '%s' to discussion: %v", label, err)
				// Continue with other labels even if one fails
			} else {
				c.debugLog("Successfully added label '%s' to discussion", label)
			}
		}
	}

	c.debugLog("Successfully created discussion '%s' (URL: %s)", disc.Title, discussionURL)
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
		return fmt.Errorf("failed to find label '%s': %w", labelName, err)
	}

	if labelResponse.Repository.Label.ID == "" {
		return fmt.Errorf("label '%s' not found in repository", labelName)
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
		return fmt.Errorf("failed to add label '%s' to discussion: %w", labelName, err)
	}

	return nil
}

// PR operations
func (c *GHClient) CreatePR(pr PRInput) error {
	if c.restClient == nil {
		return fmt.Errorf("REST client is not initialized")
	}

	c.debugLog("Creating pull request '%s' in repository %s/%s (head: %s, base: %s)", pr.Title, c.Owner, c.Repo, pr.Head, pr.Base)

	// Basic validation
	if pr.Head == "" {
		c.debugLog("PR head branch is empty")
		return fmt.Errorf("PR head branch cannot be empty")
	}
	if pr.Base == "" {
		c.debugLog("PR base branch is empty")
		return fmt.Errorf("PR base branch cannot be empty")
	}
	if pr.Head == pr.Base {
		c.debugLog("PR head and base branches are the same: %s", pr.Head)
		return fmt.Errorf("PR head and base branches cannot be the same (%s)", pr.Head)
	}

	path := fmt.Sprintf("repos/%s/%s/pulls", c.Owner, c.Repo)
	payload := map[string]interface{}{
		"title": pr.Title,
		"body":  pr.Body,
		"head":  pr.Head,
		"base":  pr.Base,
	}

	var response map[string]interface{}
	err := c.restClient.Request("POST", path, payload, &response)
	if err != nil {
		c.debugLog("Failed to create pull request '%s': %v", pr.Title, err)
		return fmt.Errorf("failed to create pull request '%s' (head: %s, base: %s): %w", pr.Title, pr.Head, pr.Base, err)
	}

	// If the PR was created successfully and has labels/assignees, add them
	if prNumber, ok := response["number"].(float64); ok && (len(pr.Labels) > 0 || len(pr.Assignees) > 0) {
		c.debugLog("Adding labels/assignees to PR '%s'", pr.Title)
		issuePayload := map[string]interface{}{
			"labels":    pr.Labels,
			"assignees": pr.Assignees,
		}

		issuePath := fmt.Sprintf("repos/%s/%s/issues/%d", c.Owner, c.Repo, int(prNumber))
		if err := c.restClient.Request("PATCH", issuePath, issuePayload, nil); err != nil {
			c.debugLog("Failed to add labels/assignees to PR '%s': %v", pr.Title, err)
			return fmt.Errorf("created PR '%s' but failed to add labels/assignees: %w", pr.Title, err)
		}
	}

	c.debugLog("Successfully created pull request '%s'", pr.Title)
	return nil
}
