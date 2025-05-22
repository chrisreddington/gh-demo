package hydrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
)

// Hydrator manages the hydration process
type Hydrator struct {
	client githubapi.GitHubClient
}

// NewHydrator creates a new Hydrator instance
func NewHydrator(client githubapi.GitHubClient) *Hydrator {
	return &Hydrator{
		client: client,
	}
}

// findConfigFile attempts to find a hydration configuration file
func findConfigFile(path string) (string, error) {
	// Check if path is a file
	fileInfo, err := os.Stat(path)
	if err == nil && !fileInfo.IsDir() {
		return path, nil
	}

	// If path is a directory, look for hydrate.json
	if err == nil && fileInfo.IsDir() {
		configPath := filepath.Join(path, "hydrate.json")
		_, err := os.Stat(configPath)
		if err == nil {
			return configPath, nil
		}
	}

	// Default to using the current directory
	configPath := "hydrate.json"
	_, err = os.Stat(configPath)
	if err == nil {
		return configPath, nil
	}

	return "", fmt.Errorf("configuration file not found")
}

// Hydrate processes a configuration file and creates the specified GitHub resources
func (h *Hydrator) Hydrate(ctx context.Context, configPath, owner, repo string) error {
	// Find the configuration file
	filePath, err := findConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("unable to find configuration file: %w", err)
	}

	// Read the configuration file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read configuration file: %w", err)
	}

	// Parse the configuration
	var schema Schema
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return fmt.Errorf("unable to parse configuration file: %w", err)
	}

	// Process labels first so they're available for issues and PRs
	if len(schema.Labels) > 0 {
		if err := h.processLabels(ctx, owner, repo, schema.Labels); err != nil {
			return fmt.Errorf("error processing labels: %w", err)
		}
	}

	// Process issues
	if len(schema.Issues) > 0 {
		if err := h.processIssues(ctx, owner, repo, schema.Issues); err != nil {
			return fmt.Errorf("error processing issues: %w", err)
		}
	}

	// Process discussions
	if len(schema.Discussions) > 0 {
		if err := h.processDiscussions(ctx, owner, repo, schema.Discussions); err != nil {
			return fmt.Errorf("error processing discussions: %w", err)
		}
	}

	// Process pull requests
	if len(schema.PRs) > 0 {
		if err := h.processPullRequests(ctx, owner, repo, schema.PRs); err != nil {
			return fmt.Errorf("error processing pull requests: %w", err)
		}
	}

	return nil
}

// processLabels creates labels specified in the configuration
func (h *Hydrator) processLabels(ctx context.Context, owner, repo string, labels []Label) error {
	for _, label := range labels {
		input := &githubapi.LabelInput{
			Name:        label.Name,
			Color:       label.Color,
			Description: label.Description,
		}

		_, err := h.client.CreateLabel(ctx, owner, repo, input)
		if err != nil {
			// Ignore errors for duplicate labels
			// TODO: Add better handling for API errors
			fmt.Printf("Warning: Could not create label %s: %v\n", label.Name, err)
			continue
		}

		fmt.Printf("Created label: %s\n", label.Name)

		// Add a small delay to avoid rate limits
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// processIssues creates issues specified in the configuration
func (h *Hydrator) processIssues(ctx context.Context, owner, repo string, issues []Issue) error {
	for _, issue := range issues {
		input := &githubapi.IssueInput{
			Title:     issue.Title,
			Body:      issue.Body,
			Labels:    issue.Labels,
			Assignees: issue.Assignees,
		}

		url, err := h.client.CreateIssue(ctx, owner, repo, input)
		if err != nil {
			return fmt.Errorf("failed to create issue '%s': %w", issue.Title, err)
		}

		fmt.Printf("Created issue: %s (%s)\n", issue.Title, url)

		// Add a small delay to avoid rate limits
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

// processDiscussions creates discussions specified in the configuration
func (h *Hydrator) processDiscussions(ctx context.Context, owner, repo string, discussions []Discussion) error {
	// Get repository ID
	repoID, err := h.client.GetRepositoryID(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get repository ID: %w", err)
	}

	// Get discussion categories
	categories, err := h.client.GetDiscussionCategories(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get discussion categories: %w", err)
	}

	// Map category names to IDs
	categoryMap := make(map[string]string)
	for _, category := range categories {
		categoryMap[category.Name] = category.ID
	}

	for _, discussion := range discussions {
		// Look up category ID
		categoryID, ok := categoryMap[discussion.Category]
		if !ok {
			return fmt.Errorf("discussion category '%s' not found", discussion.Category)
		}

		input := &githubapi.DiscussionInput{
			Title:        discussion.Title,
			Body:         discussion.Body,
			CategoryID:   categoryID,
			RepositoryID: repoID,
		}

		url, err := h.client.CreateDiscussion(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to create discussion '%s': %w", discussion.Title, err)
		}

		fmt.Printf("Created discussion: %s (%s)\n", discussion.Title, url)

		// Add a small delay to avoid rate limits
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

// processPullRequests creates pull requests specified in the configuration
func (h *Hydrator) processPullRequests(ctx context.Context, owner, repo string, prs []PR) error {
	for _, pr := range prs {
		input := &githubapi.PullRequestInput{
			Title: pr.Title,
			Body:  pr.Body,
			Head:  pr.Head,
			Base:  pr.Base,
			Draft: pr.Draft,
		}

		url, err := h.client.CreatePullRequest(ctx, owner, repo, input)
		if err != nil {
			return fmt.Errorf("failed to create pull request '%s': %w", pr.Title, err)
		}

		fmt.Printf("Created pull request: %s (%s)\n", pr.Title, url)

		// Add a small delay to avoid rate limits
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}
