/*
Package githubapi provides GitHub API client functionality for the gh-demo CLI extension.

This package implements GraphQL-only operations for interacting with the GitHub API, providing
efficient and type-safe access to GitHub resources.

Client Creation Patterns:
1. NewGHClient() - Creates a real GitHub client using go-gh authentication
2. NewGHClientWithClients() - Accepts injected GraphQL clients for testing with mocks

All GitHub operations (creating issues, discussions, pull requests, and managing labels) use
GraphQL mutations and queries for consistent performance and functionality.

Testing Strategy:
- Unit tests use NewGHClientWithClients() with mock GraphQL clients
- Integration tests use NewGHClient() and skip when authentication is unavailable
- CI runs tests in short mode to skip integration tests by default

GraphQL Operations:
- CreateLabel: Uses createLabel mutation
- CreateIssue: Uses createIssue mutation with labels and assignees
- CreatePR: Uses createPullRequest mutation with labels and assignees
- ListLabels: Uses GraphQL query for efficient label retrieval
- CreateDiscussion: Uses GraphQL for discussions and label management
*/

package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/common"
	"github.com/chrisreddington/gh-demo/internal/config"
	"github.com/chrisreddington/gh-demo/internal/errors"
	"github.com/chrisreddington/gh-demo/internal/types"
	"github.com/cli/go-gh/v2/pkg/api"
)

// GraphQLClient interface for testability
type GraphQLClient interface {
	Do(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error
}

// graphQLClientWrapper wraps the go-gh GraphQL client to implement our interface
type graphQLClientWrapper struct {
	client interface {
		Do(string, map[string]interface{}, interface{}) error
	}
}

func (w *graphQLClientWrapper) Do(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	// Check if context is already cancelled before making the request
	if err := ctx.Err(); err != nil {
		return err
	}

	// Since go-gh GraphQL client doesn't support context directly,
	// we implement context handling by running the operation in a goroutine
	// and using select to handle cancellation/timeout
	type result struct {
		err error
	}

	resultChan := make(chan result, 1)

	go func() {
		err := w.client.Do(query, variables, response)
		resultChan <- result{err: err}
	}()

	select {
	case <-ctx.Done():
		// Context was cancelled or timed out
		return ctx.Err()
	case res := <-resultChan:
		// Operation completed
		return res.err
	}
}

// GHClient is the main client for all GitHub API operations
type GHClient struct {
	Owner     string
	Repo      string
	gqlClient GraphQLClient
	logger    common.Logger
}

// NewGHClient creates a new GitHub API client for the specified owner and repository.
// It initializes the GraphQL client using the go-gh library and validates that
// the owner and repo parameters are not empty. The client uses GraphQL exclusively
// for all GitHub operations including creating issues, discussions, pull requests, and managing labels.
func NewGHClient(ctx context.Context, owner, repo string) (*GHClient, error) {
	// Check if context is cancelled before operations
	if err := ctx.Err(); err != nil {
		return nil, err
	}

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

	return &GHClient{
		Owner:     strings.TrimSpace(owner),
		Repo:      strings.TrimSpace(repo),
		gqlClient: &graphQLClientWrapper{client: gqlClient},
		logger:    nil, // Will be set when SetLogger is called
	}, nil
}

// NewGHClientWithClients creates a new GitHub API client with provided GraphQL client for testing.
// This constructor allows dependency injection of mock GraphQL clients for unit testing while
// maintaining the same validation and initialization logic as NewGHClient.
func NewGHClientWithClients(owner, repo string, gqlClient GraphQLClient) (*GHClient, error) {
	if strings.TrimSpace(owner) == "" {
		return nil, errors.ValidationError("validate_client_params", "owner cannot be empty")
	}
	if strings.TrimSpace(repo) == "" {
		return nil, errors.ValidationError("validate_client_params", "repo cannot be empty")
	}

	return &GHClient{
		Owner:     strings.TrimSpace(owner),
		Repo:      strings.TrimSpace(repo),
		gqlClient: gqlClient,
		logger:    nil, // Will be set when SetLogger is called
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

// Label operations
func (c *GHClient) ListLabels(ctx context.Context) ([]string, error) {
	if c.gqlClient == nil {
		return nil, errors.ValidationError("validate_client", "GraphQL client is not initialized")
	}

	c.debugLog("Fetching labels from repository %s/%s", c.Owner, c.Repo)

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

	// Create timeout context for API call
	apiCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
	defer cancel()

	err := c.gqlClient.Do(apiCtx, listLabelsQuery, variables, &response)
	if err != nil {
		c.debugLog("Failed to fetch labels: %v", err)
		if errors.IsContextError(err) {
			return nil, errors.ContextError("list_labels", err)
		}
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
// It validates that the GraphQL client is initialized and creates the label with
// the specified name, description, and color using GraphQL mutations.
func (c *GHClient) CreateLabel(ctx context.Context, label types.Label) error {
	if c.gqlClient == nil {
		return errors.ValidationError("validate_client", "GraphQL client is not initialized")
	}

	c.debugLog("Creating label '%s' (color: %s) in repository %s/%s", label.Name, label.Color, c.Owner, c.Repo)

	// First, get the repository ID
	var repoResponse struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}

	repoVariables := map[string]interface{}{
		"owner": c.Owner,
		"name":  c.Repo,
	}

	// Create timeout context for repository query
	repoCtx, repoCancel := context.WithTimeout(ctx, config.APITimeout)
	defer repoCancel()

	err := c.gqlClient.Do(repoCtx, getRepositoryIdQuery, repoVariables, &repoResponse)
	if err != nil {
		c.debugLog("Failed to fetch repository ID for label creation: %v", err)
		if errors.IsContextError(err) {
			return errors.ContextError("get_repository_id", err)
		}
		return errors.APIError("get_repository_id", "failed to fetch repository ID", err)
	}

	if repoResponse.Repository.ID == "" {
		return errors.ValidationError("validate_repository", "repository not found")
	}

	// Create the label using GraphQL mutation
	var mutationResponse struct {
		CreateLabel struct {
			Label struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Color       string `json:"color"`
				Description string `json:"description"`
			} `json:"label"`
		} `json:"createLabel"`
	}

	mutationVariables := map[string]interface{}{
		"repositoryId": repoResponse.Repository.ID,
		"name":         label.Name,
		"color":        label.Color,
		"description":  label.Description,
	}

	// Create timeout context for label creation
	createCtx, createCancel := context.WithTimeout(ctx, config.APITimeout)
	defer createCancel()

	err = c.gqlClient.Do(createCtx, createLabelMutation, mutationVariables, &mutationResponse)
	if err != nil {
		c.debugLog("Failed to create label '%s': %v", label.Name, err)
		if errors.IsContextError(err) {
			return errors.ContextError("create_label", err)
		}
		layeredErr := errors.NewLayeredError("api", "create_label", "failed to create GitHub label", err)
		return layeredErr.WithContext("name", label.Name).WithContext("color", label.Color)
	}

	// Verify label was created
	if mutationResponse.CreateLabel.Label.ID == "" {
		c.debugLog("Label creation for '%s' failed - no Label ID returned", label.Name)
		err := errors.APIError("create_label", "label creation failed - no Label ID returned from GitHub API", nil)
		return errors.WithContextSafe(err, "name", label.Name)
	}

	c.debugLog("Successfully created label '%s' with color '%s'", label.Name, label.Color)
	return nil
}

// resolveLabelIDs resolves label names to their corresponding IDs
func (c *GHClient) resolveLabelIDs(ctx context.Context, labelNames []string) ([]string, error) {
	if len(labelNames) == 0 {
		return nil, nil
	}

	labelIDs := make([]string, 0, len(labelNames))

	for _, labelName := range labelNames {
		var labelResponse struct {
			Repository struct {
				Label struct {
					ID string `json:"id"`
				} `json:"label"`
			} `json:"repository"`
		}

		labelVariables := map[string]interface{}{
			"owner":     c.Owner,
			"name":      c.Repo,
			"labelName": labelName,
		}

		// Create timeout context for the label query
		labelCtx, labelCancel := context.WithTimeout(ctx, config.APITimeout)
		defer labelCancel()

		err := c.gqlClient.Do(labelCtx, getLabelIdQuery, labelVariables, &labelResponse)
		if err != nil {
			c.debugLog("Failed to find label '%s': %v", labelName, err)
			// Continue with other labels even if one fails
			continue
		}

		if labelResponse.Repository.Label.ID != "" {
			labelIDs = append(labelIDs, labelResponse.Repository.Label.ID)
			c.debugLog("Resolved label '%s' to ID: %s", labelName, labelResponse.Repository.Label.ID)
		} else {
			c.debugLog("Label '%s' not found in repository", labelName)
		}
	}

	return labelIDs, nil
}

// resolveUserIDs resolves user logins to their corresponding IDs
func (c *GHClient) resolveUserIDs(ctx context.Context, userLogins []string) ([]string, error) {
	if len(userLogins) == 0 {
		return nil, nil
	}

	userIDs := make([]string, 0, len(userLogins))

	for _, login := range userLogins {
		var userResponse struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		}

		userVariables := map[string]interface{}{
			"login": login,
		}

		// Create timeout context for the user query
		userCtx, userCancel := context.WithTimeout(ctx, config.APITimeout)
		defer userCancel()

		err := c.gqlClient.Do(userCtx, getUserIdQuery, userVariables, &userResponse)
		if err != nil {
			c.debugLog("Failed to find user '%s': %v", login, err)
			// Continue with other users even if one fails
			continue
		}

		if userResponse.User.ID != "" {
			userIDs = append(userIDs, userResponse.User.ID)
			c.debugLog("Resolved user '%s' to ID: %s", login, userResponse.User.ID)
		} else {
			c.debugLog("User '%s' not found", login)
		}
	}

	return userIDs, nil
}

// CreateIssue creates a new issue in the repository using the provided issue data.
// It validates that the GraphQL client is initialized and creates the issue with
// the specified title, body, labels, and assignees using GraphQL mutations.
func (c *GHClient) CreateIssue(ctx context.Context, issue types.Issue) error {
	if c.gqlClient == nil {
		return errors.ValidationError("validate_client", "GraphQL client is not initialized")
	}

	c.debugLog("Creating issue '%s' in repository %s/%s", issue.Title, c.Owner, c.Repo)

	// First, get the repository ID
	var repoResponse struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}

	repoVariables := map[string]interface{}{
		"owner": c.Owner,
		"name":  c.Repo,
	}

	// Create timeout context for repository query
	repoCtx, repoCancel := context.WithTimeout(ctx, config.APITimeout)
	defer repoCancel()

	err := c.gqlClient.Do(repoCtx, getRepositoryIdQuery, repoVariables, &repoResponse)
	if err != nil {
		c.debugLog("Failed to fetch repository ID for issue creation: %v", err)
		if errors.IsContextError(err) {
			return errors.ContextError("get_repository_id", err)
		}
		return errors.APIError("get_repository_id", "failed to fetch repository ID", err)
	}

	if repoResponse.Repository.ID == "" {
		return errors.ValidationError("validate_repository", "repository not found")
	}

	// Resolve label names to IDs
	labelIDs, err := c.resolveLabelIDs(ctx, issue.Labels)
	if err != nil {
		c.debugLog("Failed to resolve label IDs: %v", err)
		return errors.APIError("resolve_labels", "failed to resolve label IDs", err)
	}

	// Resolve assignee logins to IDs
	assigneeIDs, err := c.resolveUserIDs(ctx, issue.Assignees)
	if err != nil {
		c.debugLog("Failed to resolve assignee IDs: %v", err)
		return errors.APIError("resolve_assignees", "failed to resolve assignee IDs", err)
	}

	// Create the issue using GraphQL mutation
	var mutationResponse struct {
		CreateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
				URL    string `json:"url"`
			} `json:"issue"`
		} `json:"createIssue"`
	}

	mutationVariables := map[string]interface{}{
		"repositoryId": repoResponse.Repository.ID,
		"title":        issue.Title,
		"body":         issue.Body,
		"labelIds":     labelIDs,
		"assigneeIds":  assigneeIDs,
	}

	// Create timeout context for issue creation
	createCtx, createCancel := context.WithTimeout(ctx, config.APITimeout)
	defer createCancel()

	err = c.gqlClient.Do(createCtx, createIssueMutation, mutationVariables, &mutationResponse)
	if err != nil {
		c.debugLog("Failed to create issue '%s': %v", issue.Title, err)
		if errors.IsContextError(err) {
			return errors.ContextError("create_issue", err)
		}
		layeredErr := errors.NewLayeredError("api", "create_issue", "failed to create GitHub issue", err)
		return layeredErr.WithContext("title", issue.Title)
	}

	// Verify issue was created
	if mutationResponse.CreateIssue.Issue.ID == "" {
		c.debugLog("Issue creation for '%s' failed - no Issue ID returned", issue.Title)
		err := errors.APIError("create_issue", "issue creation failed - no Issue ID returned from GitHub API", nil)
		return errors.WithContextSafe(err, "title", issue.Title)
	}

	c.debugLog("Successfully created issue '%s' (Number: %d, URL: %s)",
		issue.Title, mutationResponse.CreateIssue.Issue.Number, mutationResponse.CreateIssue.Issue.URL)
	return nil
}

// CreateDiscussion creates a new discussion in the repository using the provided discussion data.
// It uses GraphQL to create the discussion with the specified title, body, category, and labels.
// The method automatically finds the correct category ID and adds labels after creation.
func (c *GHClient) CreateDiscussion(ctx context.Context, discussion types.Discussion) error {
	if c.gqlClient == nil {
		return errors.ValidationError("validate_client", "GraphQL client is not initialized")
	}

	c.debugLog("Creating discussion '%s' in repository %s/%s", discussion.Title, c.Owner, c.Repo)

	// First, get the repository ID and discussion categories

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

	// Create timeout context for API call
	apiCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
	defer cancel()

	err := c.gqlClient.Do(apiCtx, repositoryWithDiscussionCategoriesQuery, repoVariables, &repoResponse)
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
		err := errors.ValidationError("validate_discussion_category", fmt.Sprintf("discussion category '%s' not found in available categories", discussion.Category))
		err = errors.WithContextSafe(err, "requested_category", discussion.Category)
		return errors.WithContextSafe(err, "available_categories", fmt.Sprintf("%v", availableCategories))
	}

	c.debugLog("Found matching category ID for '%s': %s (actual: '%s')",
		discussion.Category, categoryID, matchedCategory)

	// Create the discussion

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

	// Create timeout context for the creation mutation
	createCtx, createCancel := context.WithTimeout(ctx, config.APITimeout)
	defer createCancel()

	err = c.gqlClient.Do(createCtx, createDiscussionMutation, mutationVariables, &mutationResponse)
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
		err := errors.APIError("create_discussion", "discussion creation failed - no Discussion ID returned from GitHub API", nil)
		return errors.WithContextSafe(err, "title", discussion.Title)
	}

	discussionID := mutationResponse.CreateDiscussion.Discussion.ID
	discussionURL := mutationResponse.CreateDiscussion.Discussion.URL
	c.debugLog("Discussion created with ID: %s, URL: %s", discussionID, discussionURL)

	// Add labels if specified
	if len(discussion.Labels) > 0 && mutationResponse.CreateDiscussion.Discussion.ID != "" {
		c.debugLog("Adding %d labels to discussion '%s'", len(discussion.Labels), discussion.Title)

		// Add labels to the discussion using the discussion ID
		for _, label := range discussion.Labels {
			err := c.addLabelToDiscussion(ctx, mutationResponse.CreateDiscussion.Discussion.ID, label)
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
func (c *GHClient) addLabelToDiscussion(ctx context.Context, discussionID, labelName string) error {
	// First, find the label ID for the label name

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

	// Create timeout context for the label query
	labelCtx, labelCancel := context.WithTimeout(ctx, config.APITimeout)
	defer labelCancel()

	err := c.gqlClient.Do(labelCtx, labelByNameQuery, labelVariables, &labelResponse)
	if err != nil {
		return errors.APIError("find_label", fmt.Sprintf("failed to find label '%s'", labelName), err)
	}

	if labelResponse.Repository.Label.ID == "" {
		err := errors.ValidationError("validate_label", fmt.Sprintf("label '%s' not found in repository", labelName))
		return errors.WithContextSafe(err, "label_name", labelName)
	}

	// Add the label to the discussion

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

	// Create timeout context for the add label mutation
	addLabelCtx, addLabelCancel := context.WithTimeout(ctx, config.APITimeout)
	defer addLabelCancel()

	err = c.gqlClient.Do(addLabelCtx, addLabelsToLabelableMutation, labelMutationVariables, &labelMutationResponse)
	if err != nil {
		return errors.APIError("add_label_to_discussion", fmt.Sprintf("failed to add label '%s' to discussion", labelName), err)
	}

	return nil
}

// addLabelsAndAssigneesToPR adds labels and assignees to an existing pull request using its ID
func (c *GHClient) addLabelsAndAssigneesToPR(ctx context.Context, prID string, labelNames []string, assigneeLogins []string) error {
	if len(labelNames) == 0 && len(assigneeLogins) == 0 {
		return nil // Nothing to add
	}

	// Resolve label names to IDs
	labelIDs, err := c.resolveLabelIDs(ctx, labelNames)
	if err != nil {
		c.debugLog("Failed to resolve label IDs for PR: %v", err)
		return errors.APIError("resolve_labels", "failed to resolve label IDs", err)
	}

	// Resolve assignee logins to IDs
	assigneeIDs, err := c.resolveUserIDs(ctx, assigneeLogins)
	if err != nil {
		c.debugLog("Failed to resolve assignee IDs for PR: %v", err)
		return errors.APIError("resolve_assignees", "failed to resolve assignee IDs", err)
	}

	// Only proceed if we have labels or assignees to add
	if len(labelIDs) == 0 && len(assigneeIDs) == 0 {
		c.debugLog("No valid labels or assignees to add to PR")
		return nil
	}

	// Add labels if we have any
	if len(labelIDs) > 0 {

		var labelResponse struct {
			AddLabelsToLabelable struct {
				ClientMutationID string `json:"clientMutationId"`
			} `json:"addLabelsToLabelable"`
		}

		labelVariables := map[string]interface{}{
			"labelableId": prID,
			"labelIds":    labelIDs,
		}

		labelCtx, labelCancel := context.WithTimeout(ctx, config.APITimeout)
		defer labelCancel()

		err = c.gqlClient.Do(labelCtx, addLabelsToLabelableMutationWithParams, labelVariables, &labelResponse)
		if err != nil {
			c.debugLog("Failed to add labels to PR: %v", err)
			return errors.APIError("add_labels_to_pr", "failed to add labels to pull request", err)
		}
	}

	// Add assignees if we have any
	if len(assigneeIDs) > 0 {

		var assigneeResponse struct {
			AddAssigneesToAssignable struct {
				ClientMutationID string `json:"clientMutationId"`
			} `json:"addAssigneesToAssignable"`
		}

		assigneeVariables := map[string]interface{}{
			"assignableId": prID,
			"assigneeIds":  assigneeIDs,
		}

		assigneeCtx, assigneeCancel := context.WithTimeout(ctx, config.APITimeout)
		defer assigneeCancel()

		err = c.gqlClient.Do(assigneeCtx, addAssigneesToAssignableMutation, assigneeVariables, &assigneeResponse)
		if err != nil {
			c.debugLog("Failed to add assignees to PR: %v", err)
			return errors.APIError("add_assignees_to_pr", "failed to add assignees to pull request", err)
		}
	}

	return nil
}

// CreatePR creates a new pull request in the repository using the provided pull request data.
// It validates the head and base branches, creates the PR via GraphQL API, and adds labels/assignees if specified.
func (c *GHClient) CreatePR(ctx context.Context, pullRequest types.PullRequest) error {
	if c.gqlClient == nil {
		return errors.ValidationError("validate_client", "GraphQL client is not initialized")
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

	// First, get the repository ID
	var repoResponse struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}

	repoVariables := map[string]interface{}{
		"owner": c.Owner,
		"name":  c.Repo,
	}

	// Create timeout context for repository query
	repoCtx, repoCancel := context.WithTimeout(ctx, config.APITimeout)
	defer repoCancel()

	err := c.gqlClient.Do(repoCtx, getRepositoryIdQuery, repoVariables, &repoResponse)
	if err != nil {
		c.debugLog("Failed to fetch repository ID for PR creation: %v", err)
		if errors.IsContextError(err) {
			return errors.ContextError("get_repository_id", err)
		}
		return errors.APIError("get_repository_id", "failed to fetch repository ID", err)
	}

	if repoResponse.Repository.ID == "" {
		return errors.ValidationError("validate_repository", "repository not found")
	}

	// Create the pull request using GraphQL mutation
	var mutationResponse struct {
		CreatePullRequest struct {
			PullRequest struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
				URL    string `json:"url"`
			} `json:"pullRequest"`
		} `json:"createPullRequest"`
	}

	mutationVariables := map[string]interface{}{
		"repositoryId": repoResponse.Repository.ID,
		"title":        pullRequest.Title,
		"body":         pullRequest.Body,
		"headRefName":  pullRequest.Head,
		"baseRefName":  pullRequest.Base,
	}

	// Create timeout context for PR creation
	createCtx, createCancel := context.WithTimeout(ctx, config.APITimeout)
	defer createCancel()

	err = c.gqlClient.Do(createCtx, createPullRequestMutation, mutationVariables, &mutationResponse)
	if err != nil {
		c.debugLog("Failed to create pull request '%s': %v", pullRequest.Title, err)
		if errors.IsContextError(err) {
			return errors.ContextError("create_pull_request", err)
		}
		err = errors.APIError("create_pull_request", "failed to create pull request", err)
		err = errors.WithContextSafe(err, "title", pullRequest.Title)
		err = errors.WithContextSafe(err, "head", pullRequest.Head)
		return errors.WithContextSafe(err, "base", pullRequest.Base)
	}

	// Verify PR was created
	if mutationResponse.CreatePullRequest.PullRequest.ID == "" {
		c.debugLog("PR creation for '%s' failed - no PR ID returned", pullRequest.Title)
		err := errors.APIError("create_pull_request", "pull request creation failed - no PR ID returned from GitHub API", nil)
		return errors.WithContextSafe(err, "title", pullRequest.Title)
	}

	prID := mutationResponse.CreatePullRequest.PullRequest.ID
	c.debugLog("Successfully created pull request '%s' (Number: %d, URL: %s)",
		pullRequest.Title, mutationResponse.CreatePullRequest.PullRequest.Number, mutationResponse.CreatePullRequest.PullRequest.URL)

	// Add labels and assignees if specified
	if len(pullRequest.Labels) > 0 || len(pullRequest.Assignees) > 0 {
		c.debugLog("Adding labels/assignees to PR '%s'", pullRequest.Title)
		err := c.addLabelsAndAssigneesToPR(ctx, prID, pullRequest.Labels, pullRequest.Assignees)
		if err != nil {
			c.debugLog("Failed to add labels/assignees to PR '%s': %v", pullRequest.Title, err)
			err = errors.APIError("add_pr_labels_assignees", "created PR but failed to add labels/assignees", err)
			return errors.WithContextSafe(err, "title", pullRequest.Title)
		}
	}

	c.debugLog("Successfully created pull request '%s'", pullRequest.Title)
	return nil
}

// Listing operations for cleanup

// ListIssues retrieves all existing issues from the repository
func (c *GHClient) ListIssues(ctx context.Context) ([]types.Issue, error) {
	if c.gqlClient == nil {
		return nil, errors.ValidationError("list_issues", "GraphQL client is not initialized")
	}

	c.debugLog("Fetching issues from repository %s/%s", c.Owner, c.Repo)

	var allIssues []types.Issue
	var cursor *string

	for {
		var response struct {
			Repository struct {
				Issues struct {
					Nodes []struct {
						ID     string `json:"id"`
						Number int    `json:"number"`
						Title  string `json:"title"`
						Body   string `json:"body"`
						Labels struct {
							Nodes []struct {
								Name string `json:"name"`
							} `json:"nodes"`
						} `json:"labels"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool    `json:"hasNextPage"`
						EndCursor   *string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"issues"`
			} `json:"repository"`
		}

		variables := map[string]interface{}{
			"owner": c.Owner,
			"name":  c.Repo,
			"first": 100,
		}
		if cursor != nil {
			variables["after"] = *cursor
		}

		// Create timeout context for API call
		apiCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
		defer cancel()

		err := c.gqlClient.Do(apiCtx, listIssuesQuery, variables, &response)
		if err != nil {
			c.debugLog("Failed to fetch issues: %v", err)
			if errors.IsContextError(err) {
				return nil, errors.ContextError("list_issues", err)
			}
			return nil, errors.APIError("list_issues", "failed to fetch issues", err)
		}

		// Convert GraphQL response to types.Issue
		for _, issue := range response.Repository.Issues.Nodes {
			labels := make([]string, 0, len(issue.Labels.Nodes))
			for _, label := range issue.Labels.Nodes {
				labels = append(labels, label.Name)
			}

			allIssues = append(allIssues, types.Issue{
				NodeID: issue.ID,
				Number: issue.Number,
				Title:  issue.Title,
				Body:   issue.Body,
				Labels: labels,
			})
		}

		// Check if we need to fetch more pages
		if !response.Repository.Issues.PageInfo.HasNextPage {
			break
		}
		cursor = response.Repository.Issues.PageInfo.EndCursor
	}

	c.debugLog("Successfully fetched %d issues", len(allIssues))
	return allIssues, nil
}

// ListDiscussions retrieves all existing discussions from the repository
func (c *GHClient) ListDiscussions(ctx context.Context) ([]types.Discussion, error) {
	if c.gqlClient == nil {
		return nil, errors.ValidationError("list_discussions", "GraphQL client is not initialized")
	}

	c.debugLog("Fetching discussions from repository %s/%s", c.Owner, c.Repo)

	var allDiscussions []types.Discussion
	var cursor *string

	for {
		var response struct {
			Repository struct {
				Discussions struct {
					Nodes []struct {
						ID       string `json:"id"`
						Number   int    `json:"number"`
						Title    string `json:"title"`
						Body     string `json:"body"`
						Category struct {
							Name string `json:"name"`
						} `json:"category"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool    `json:"hasNextPage"`
						EndCursor   *string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"discussions"`
			} `json:"repository"`
		}

		variables := map[string]interface{}{
			"owner": c.Owner,
			"name":  c.Repo,
			"first": 100,
		}
		if cursor != nil {
			variables["after"] = *cursor
		}

		// Create timeout context for API call
		apiCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
		defer cancel()

		err := c.gqlClient.Do(apiCtx, listDiscussionsQuery, variables, &response)
		if err != nil {
			c.debugLog("Failed to fetch discussions: %v", err)
			if errors.IsContextError(err) {
				return nil, errors.ContextError("list_discussions", err)
			}
			return nil, errors.APIError("list_discussions", "failed to fetch discussions", err)
		}

		// Convert GraphQL response to types.Discussion
		for _, discussion := range response.Repository.Discussions.Nodes {
			allDiscussions = append(allDiscussions, types.Discussion{
				NodeID:   discussion.ID,
				Number:   discussion.Number,
				Title:    discussion.Title,
				Body:     discussion.Body,
				Category: discussion.Category.Name,
			})
		}

		// Check if we need to fetch more pages
		if !response.Repository.Discussions.PageInfo.HasNextPage {
			break
		}
		cursor = response.Repository.Discussions.PageInfo.EndCursor
	}

	c.debugLog("Successfully fetched %d discussions", len(allDiscussions))
	return allDiscussions, nil
}

// ListPRs retrieves all existing pull requests from the repository
func (c *GHClient) ListPRs(ctx context.Context) ([]types.PullRequest, error) {
	if c.gqlClient == nil {
		return nil, errors.ValidationError("list_prs", "GraphQL client is not initialized")
	}

	c.debugLog("Fetching pull requests from repository %s/%s", c.Owner, c.Repo)

	var allPRs []types.PullRequest
	var cursor *string

	for {
		var response struct {
			Repository struct {
				PullRequests struct {
					Nodes []struct {
						ID          string `json:"id"`
						Number      int    `json:"number"`
						Title       string `json:"title"`
						Body        string `json:"body"`
						HeadRefName string `json:"headRefName"`
						BaseRefName string `json:"baseRefName"`
						Labels      struct {
							Nodes []struct {
								Name string `json:"name"`
							} `json:"nodes"`
						} `json:"labels"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool    `json:"hasNextPage"`
						EndCursor   *string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"pullRequests"`
			} `json:"repository"`
		}

		variables := map[string]interface{}{
			"owner": c.Owner,
			"name":  c.Repo,
			"first": 100,
		}
		if cursor != nil {
			variables["after"] = *cursor
		}

		// Create timeout context for API call
		apiCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
		defer cancel()

		err := c.gqlClient.Do(apiCtx, listPullRequestsQuery, variables, &response)
		if err != nil {
			c.debugLog("Failed to fetch pull requests: %v", err)
			if errors.IsContextError(err) {
				return nil, errors.ContextError("list_prs", err)
			}
			return nil, errors.APIError("list_prs", "failed to fetch pull requests", err)
		}

		// Convert GraphQL response to types.PullRequest
		for _, pr := range response.Repository.PullRequests.Nodes {
			labels := make([]string, 0, len(pr.Labels.Nodes))
			for _, label := range pr.Labels.Nodes {
				labels = append(labels, label.Name)
			}

			allPRs = append(allPRs, types.PullRequest{
				NodeID: pr.ID,
				Number: pr.Number,
				Title:  pr.Title,
				Body:   pr.Body,
				Head:   pr.HeadRefName,
				Base:   pr.BaseRefName,
				Labels: labels,
			})
		}

		// Check if we need to fetch more pages
		if !response.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}
		cursor = response.Repository.PullRequests.PageInfo.EndCursor
	}

	c.debugLog("Successfully fetched %d pull requests", len(allPRs))
	return allPRs, nil
}

// Deletion operations for cleanup

// DeleteIssue deletes an issue by its node ID
func (c *GHClient) DeleteIssue(ctx context.Context, nodeID string) error {
	if c.gqlClient == nil {
		return errors.ValidationError("delete_issue", "GraphQL client is not initialized")
	}

	if strings.TrimSpace(nodeID) == "" {
		return errors.ValidationError("delete_issue", "node ID cannot be empty")
	}

	c.debugLog("Closing issue with nodeID: %s in repository %s/%s", nodeID, c.Owner, c.Repo)

	var response struct {
		CloseIssue struct {
			Issue struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"issue"`
		} `json:"closeIssue"`
	}

	variables := map[string]interface{}{
		"issueId": nodeID,
	}

	// Create timeout context for API call
	apiCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
	defer cancel()

	err := c.gqlClient.Do(apiCtx, deleteIssueMutation, variables, &response)
	if err != nil {
		c.debugLog("Failed to close issue %s: %v", nodeID, err)
		if errors.IsContextError(err) {
			return errors.ContextError("delete_issue", err)
		}
		err = errors.APIError("delete_issue", "failed to close issue", err)
		return errors.WithContextSafe(err, "node_id", nodeID)
	}

	// Verify the issue was closed
	if response.CloseIssue.Issue.State != "CLOSED" {
		c.debugLog("Issue %s was not properly closed - state: %s", nodeID, response.CloseIssue.Issue.State)
		err := errors.APIError("delete_issue", "issue was not properly closed", nil)
		err = errors.WithContextSafe(err, "node_id", nodeID)
		return errors.WithContextSafe(err, "state", response.CloseIssue.Issue.State)
	}

	c.debugLog("Successfully closed issue %s", nodeID)
	return nil
}

// DeleteDiscussion deletes a discussion by its node ID using the GraphQL deleteDiscussion mutation
func (c *GHClient) DeleteDiscussion(ctx context.Context, nodeID string) error {
	if c.gqlClient == nil {
		return errors.ValidationError("delete_discussion", "GraphQL client is not initialized")
	}

	if strings.TrimSpace(nodeID) == "" {
		return errors.ValidationError("delete_discussion", "node ID cannot be empty")
	}

	c.debugLog("Deleting discussion with nodeID: %s in repository %s/%s", nodeID, c.Owner, c.Repo)

	mutationVariables := map[string]interface{}{
		"discussionId": nodeID,
	}

	var mutationResponse struct {
		DeleteDiscussion struct {
			Discussion struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"discussion"`
		} `json:"deleteDiscussion"`
	}

	deleteCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
	defer cancel()

	err := c.gqlClient.Do(deleteCtx, deleteDiscussionMutation, mutationVariables, &mutationResponse)
	if err != nil {
		c.debugLog("Failed to delete discussion with nodeID %s: %v", nodeID, err)
		err = errors.APIError("delete_discussion", "failed to delete discussion via GraphQL", err)
		return errors.WithContextSafe(err, "node_id", nodeID)
	}

	c.debugLog("Successfully deleted discussion '%s' (ID: %s)",
		mutationResponse.DeleteDiscussion.Discussion.Title,
		mutationResponse.DeleteDiscussion.Discussion.ID)

	return nil
}

// DeletePR deletes a pull request by its node ID
func (c *GHClient) DeletePR(ctx context.Context, nodeID string) error {
	if c.gqlClient == nil {
		return errors.ValidationError("delete_pr", "GraphQL client is not initialized")
	}

	if strings.TrimSpace(nodeID) == "" {
		return errors.ValidationError("delete_pr", "node ID cannot be empty")
	}

	c.debugLog("Closing pull request with nodeID: %s in repository %s/%s", nodeID, c.Owner, c.Repo)

	var response struct {
		ClosePullRequest struct {
			PullRequest struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"pullRequest"`
		} `json:"closePullRequest"`
	}

	variables := map[string]interface{}{
		"pullRequestId": nodeID,
	}

	// Create timeout context for API call
	apiCtx, cancel := context.WithTimeout(ctx, config.APITimeout)
	defer cancel()

	err := c.gqlClient.Do(apiCtx, deletePullRequestMutation, variables, &response)
	if err != nil {
		c.debugLog("Failed to close pull request %s: %v", nodeID, err)
		if errors.IsContextError(err) {
			return errors.ContextError("delete_pr", err)
		}
		err = errors.APIError("delete_pr", "failed to close pull request", err)
		return errors.WithContextSafe(err, "node_id", nodeID)
	}

	// Verify the pull request was closed
	if response.ClosePullRequest.PullRequest.State != "CLOSED" {
		c.debugLog("Pull request %s was not properly closed - state: %s", nodeID, response.ClosePullRequest.PullRequest.State)
		err := errors.APIError("delete_pr", "pull request was not properly closed", nil)
		err = errors.WithContextSafe(err, "node_id", nodeID)
		return errors.WithContextSafe(err, "state", response.ClosePullRequest.PullRequest.State)
	}

	c.debugLog("Successfully closed pull request %s", nodeID)
	return nil
}

// DeleteLabel deletes a label by its name
func (c *GHClient) DeleteLabel(ctx context.Context, name string) error {
	if c.gqlClient == nil {
		return errors.ValidationError("delete_label", "GraphQL client is not initialized")
	}

	if strings.TrimSpace(name) == "" {
		return errors.ValidationError("delete_label", "label name cannot be empty")
	}

	c.debugLog("Deleting label '%s' from repository %s/%s", name, c.Owner, c.Repo)

	// First, get the label ID by name
	var labelResponse struct {
		Repository struct {
			Label struct {
				ID string `json:"id"`
			} `json:"label"`
		} `json:"repository"`
	}

	labelVariables := map[string]interface{}{
		"owner":     c.Owner,
		"name":      c.Repo,
		"labelName": name,
	}

	// Create timeout context for the label query
	labelCtx, labelCancel := context.WithTimeout(ctx, config.APITimeout)
	defer labelCancel()

	err := c.gqlClient.Do(labelCtx, getLabelByNameQuery, labelVariables, &labelResponse)
	if err != nil {
		c.debugLog("Failed to find label '%s': %v", name, err)
		if errors.IsContextError(err) {
			return errors.ContextError("find_label", err)
		}
		err = errors.APIError("find_label", fmt.Sprintf("failed to find label '%s'", name), err)
		return errors.WithContextSafe(err, "label_name", name)
	}

	if labelResponse.Repository.Label.ID == "" {
		c.debugLog("Label '%s' not found in repository", name)
		err := errors.ValidationError("validate_label", fmt.Sprintf("label '%s' not found in repository", name))
		return errors.WithContextSafe(err, "label_name", name)
	}

	// Delete the label using its ID
	var deleteResponse struct {
		DeleteLabel struct {
			ClientMutationID string `json:"clientMutationId"`
		} `json:"deleteLabel"`
	}

	deleteVariables := map[string]interface{}{
		"labelId": labelResponse.Repository.Label.ID,
	}

	// Create timeout context for the delete mutation
	deleteCtx, deleteCancel := context.WithTimeout(ctx, config.APITimeout)
	defer deleteCancel()

	err = c.gqlClient.Do(deleteCtx, deleteLabelMutation, deleteVariables, &deleteResponse)
	if err != nil {
		c.debugLog("Failed to delete label '%s': %v", name, err)
		if errors.IsContextError(err) {
			return errors.ContextError("delete_label", err)
		}
		err = errors.APIError("delete_label", fmt.Sprintf("failed to delete label '%s'", name), err)
		err = errors.WithContextSafe(err, "label_name", name)
		return errors.WithContextSafe(err, "label_id", labelResponse.Repository.Label.ID)
	}

	c.debugLog("Successfully deleted label '%s'", name)
	return nil
}
