package config

import (
	"context"
	"path/filepath"
	"testing"
)

func TestNewConfiguration(t *testing.T) {
	basePath := "test/config"
	config := NewConfiguration(basePath)

	if config.BasePath != basePath {
		t.Errorf("Expected BasePath %s, got %s", basePath, config.BasePath)
	}

	expectedIssuesPath := filepath.Join(basePath, IssuesFilename)
	if config.IssuesPath != expectedIssuesPath {
		t.Errorf("Expected IssuesPath %s, got %s", expectedIssuesPath, config.IssuesPath)
	}

	expectedDiscussionsPath := filepath.Join(basePath, DiscussionsFilename)
	if config.DiscussionsPath != expectedDiscussionsPath {
		t.Errorf("Expected DiscussionsPath %s, got %s", expectedDiscussionsPath, config.DiscussionsPath)
	}

	expectedPRsPath := filepath.Join(basePath, PullRequestsFilename)
	if config.PullRequestsPath != expectedPRsPath {
		t.Errorf("Expected PullRequestsPath %s, got %s", expectedPRsPath, config.PullRequestsPath)
	}

	expectedLabelsPath := filepath.Join(basePath, LabelsFilename)
	if config.LabelsPath != expectedLabelsPath {
		t.Errorf("Expected LabelsPath %s, got %s", expectedLabelsPath, config.LabelsPath)
	}
}

func TestNewConfigurationWithRoot(t *testing.T) {
	projectRoot := "/project/root"
	basePath := "config/demo"

	config := NewConfigurationWithRoot(context.Background(), projectRoot, basePath)

	expectedBasePath := filepath.Join(projectRoot, basePath)
	if config.BasePath != expectedBasePath {
		t.Errorf("Expected BasePath %s, got %s", expectedBasePath, config.BasePath)
	}

	expectedIssuesPath := filepath.Join(expectedBasePath, IssuesFilename)
	if config.IssuesPath != expectedIssuesPath {
		t.Errorf("Expected IssuesPath %s, got %s", expectedIssuesPath, config.IssuesPath)
	}
}

func TestConfigurationConstants(t *testing.T) {
	// Test that file name constants are set correctly
	if IssuesFilename != "issues.json" {
		t.Errorf("Expected IssuesFilename 'issues.json', got %s", IssuesFilename)
	}
	if DiscussionsFilename != "discussions.json" {
		t.Errorf("Expected DiscussionsFilename 'discussions.json', got %s", DiscussionsFilename)
	}
	if PullRequestsFilename != "prs.json" {
		t.Errorf("Expected PullRequestsFilename 'prs.json', got %s", PullRequestsFilename)
	}
	if LabelsFilename != "labels.json" {
		t.Errorf("Expected LabelsFilename 'labels.json', got %s", LabelsFilename)
	}
}
