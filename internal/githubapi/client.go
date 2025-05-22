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
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

// Label operations
func (c *GHClient) ListLabels() ([]string, error) {
	if c.gqlClient == nil {
		return nil, fmt.Errorf("GraphQL client is not initialized")
	}

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
		return nil, fmt.Errorf("failed to fetch labels: %w", err)
	}

	labels := make([]string, 0, len(query.Repository.Labels.Nodes))
	for _, label := range query.Repository.Labels.Nodes {
		labels = append(labels, label.Name)
	}

	return labels, nil
}

func (c *GHClient) CreateLabel(label string) error {
	if c.restClient == nil {
		return fmt.Errorf("REST client is not initialized")
	}

	// Using the REST API for label creation as it's simpler than GraphQL for this case
	path := fmt.Sprintf("repos/%s/%s/labels", c.Owner, c.Repo)
	payload := map[string]interface{}{
		"name":        label,
		"description": fmt.Sprintf("Label created by gh-demo hydration tool"),
		"color":       "ededed", // Light gray color as default
	}

	return c.restClient.Request("POST", path, payload, nil)
}

// Issue operations
func (c *GHClient) CreateIssue(issue IssueInput) error {
	if c.restClient == nil {
		return fmt.Errorf("REST client is not initialized")
	}

	path := fmt.Sprintf("repos/%s/%s/issues", c.Owner, c.Repo)
	payload := map[string]interface{}{
		"title":     issue.Title,
		"body":      issue.Body,
		"labels":    issue.Labels,
		"assignees": issue.Assignees,
	}

	return c.restClient.Request("POST", path, payload, nil)
}

// Discussion operations
func (c *GHClient) CreateDiscussion(disc DiscussionInput) error {
	if c.gqlClient == nil {
		return fmt.Errorf("GraphQL client is not initialized")
	}

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
		return fmt.Errorf("failed to create discussion: %w", err)
	}

	// Add labels if specified
	if len(disc.Labels) > 0 && mutation.CreateDiscussion.Discussion.ID != "" {
		// Note: Adding labels to discussions might require additional API calls
		// This would be implemented in a full solution
	}

	return nil
}

// PR operations
func (c *GHClient) CreatePR(pr PRInput) error {
	if c.restClient == nil {
		return fmt.Errorf("REST client is not initialized")
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
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	// If the PR was created successfully and has labels/assignees, add them
	if prNumber, ok := response["number"].(float64); ok && (len(pr.Labels) > 0 || len(pr.Assignees) > 0) {
		issuePayload := map[string]interface{}{
			"labels":    pr.Labels,
			"assignees": pr.Assignees,
		}

		issuePath := fmt.Sprintf("repos/%s/%s/issues/%d", c.Owner, c.Repo, int(prNumber))
		if err := c.restClient.Request("PATCH", issuePath, issuePayload, nil); err != nil {
			return fmt.Errorf("created PR but failed to add labels/assignees: %w", err)
		}
	}

	return nil
}
