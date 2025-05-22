package githubapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

// GHClient is the main client for all GitHub API operations
type GHClient struct {
	Owner      string
	Repo       string
	gqlClient  *GQLClient
	restClient *RESTClient
	logger     Logger
}

// GQLClient wraps the GraphQL client for testability
type GQLClient struct {
	client interface {
		Query(string, interface{}, map[string]interface{}) error
		Mutate(string, interface{}, map[string]interface{}) error
	}
}

// RESTClient wraps the REST client for testability
type RESTClient struct {
	client interface {
		Request(string, string, io.Reader) (*http.Response, error)
	}
}

func NewGHClient(owner, repo string) *GHClient {
	// Create GraphQL and REST clients using go-gh
	var gqlClient *GQLClient
	var restClient *RESTClient

	gqlRawClient, err := api.DefaultGraphQLClient()
	if err == nil {
		gqlClient = &GQLClient{client: gqlRawClient}
	} else {
		fmt.Printf("Warning: Failed to initialize GraphQL client: %v\n", err)
	}

	restRawClient, err := api.DefaultRESTClient()
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

// Query executes a GraphQL query
func (c *GQLClient) Query(name string, query interface{}, variables map[string]interface{}) error {
	return c.client.Query(name, query, variables)
}

// Mutate executes a GraphQL mutation
func (c *GQLClient) Mutate(name string, query interface{}, variables map[string]interface{}) error {
	return c.client.Mutate(name, query, variables)
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

	var query struct {
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
	}

	variables := map[string]interface{}{
		"owner": graphql.String(c.Owner),
		"name":  graphql.String(c.Repo),
	}

	err := c.gqlClient.Query("RepositoryLabels", &query, variables)
	if err != nil {
		c.debugLog("Failed to fetch labels: %v", err)
		return nil, fmt.Errorf("failed to fetch labels: %w", err)
	}

	labels := make([]string, 0, len(query.Repository.Labels.Nodes))
	for _, label := range query.Repository.Labels.Nodes {
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
		"description": fmt.Sprintf("Label created by gh-demo hydration tool"),
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

	// First, get the repository ID and category ID
	var repoQuery struct {
		Repository struct {
			ID         string
			Categories struct {
				Nodes []struct {
					ID   string
					Name string
				}
			} `graphql:"discussionCategories(first: 50)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	repoVariables := map[string]interface{}{
		"owner": graphql.String(c.Owner),
		"name":  graphql.String(c.Repo),
	}

	err := c.gqlClient.Query("RepositoryInfo", &repoQuery, repoVariables)
	if err != nil {
		c.debugLog("Failed to fetch repository info for discussion: %v", err)
		return fmt.Errorf("failed to fetch repository info: %w", err)
	}

	// Find the category ID that matches the requested category name
	var categoryID string
	for _, category := range repoQuery.Repository.Categories.Nodes {
		if strings.EqualFold(category.Name, disc.Category) {
			categoryID = category.ID
			break
		}
	}

	if categoryID == "" {
		c.debugLog("Discussion category '%s' not found", disc.Category)
		return fmt.Errorf("discussion category '%s' not found", disc.Category)
	}

	// Create the discussion
	var mutation struct {
		CreateDiscussion struct {
			Discussion struct {
				ID string
			}
		} `graphql:"createDiscussion(input: $input)"`
	}

	mutationVariables := map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": repoQuery.Repository.ID,
			"categoryId":   categoryID,
			"title":        disc.Title,
			"body":         disc.Body,
		},
	}

	err = c.gqlClient.Mutate("CreateDiscussion", &mutation, mutationVariables)
	if err != nil {
		c.debugLog("Failed to create discussion '%s': %v", disc.Title, err)
		return fmt.Errorf("failed to create discussion: %w", err)
	}

	// Add labels if specified
	if len(disc.Labels) > 0 && mutation.CreateDiscussion.Discussion.ID != "" {
		// Add labels to the discussion using the discussion ID
		for _, label := range disc.Labels {
			var labelMutation struct {
				AddLabelsToLabelable struct {
					Labelable struct {
						ID string
					}
				} `graphql:"addLabelsToLabelable(input: $input)"`
			}
			
			// First, we need to find the label ID for the label name
			var labelQuery struct {
				Repository struct {
					Label struct {
						ID string
					} `graphql:"label(name: $name)"`
				} `graphql:"repository(owner: $owner, name: $repo)"`
			}
			
			labelVariables := map[string]interface{}{
				"owner": graphql.String(c.Owner),
				"repo":  graphql.String(c.Repo),
				"name":  graphql.String(label),
			}
			
			err = c.gqlClient.Query("GetLabelID", &labelQuery, labelVariables)
			if err != nil {
				c.debugLog("Failed to find label '%s' for discussion: %v", label, err)
				continue // Skip this label if not found
			}
			
			if labelQuery.Repository.Label.ID == "" {
				c.debugLog("Label '%s' not found in repository", label)
				continue
			}
			
			labelMutationVariables := map[string]interface{}{
				"input": map[string]interface{}{
					"labelableId": mutation.CreateDiscussion.Discussion.ID,
					"labelIds":    []string{labelQuery.Repository.Label.ID},
				},
			}
			
			err = c.gqlClient.Mutate("AddLabelToDiscussion", &labelMutation, labelMutationVariables)
			if err != nil {
				c.debugLog("Failed to add label '%s' to discussion: %v", label, err)
				// Continue with other labels even if one fails
			} else {
				c.debugLog("Successfully added label '%s' to discussion", label)
			}
		}
	}

	c.debugLog("Successfully created discussion '%s'", disc.Title)
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
