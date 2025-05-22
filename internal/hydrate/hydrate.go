package hydrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chrisreddington/gh-demo/internal/githubapi"
)

// Logger handles debug and info logging
type Logger struct {
	debug bool
}

// NewLogger creates a new logger with the specified debug mode
func NewLogger(debug bool) *Logger {
	return &Logger{debug: debug}
}

// Debug logs a message only when debug mode is enabled
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Info logs a message always
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// SectionSummary holds statistics for a section
type SectionSummary struct {
	Name     string
	Total    int
	Success  int
	Failures int
	Errors   []string
}

// HydrateWithLabels loads content, collects all labels, and ensures labels exist before hydration.
// It continues processing even if individual items fail, collecting all errors and reporting them at the end.
func HydrateWithLabels(client githubapi.GitHubClient, issuesPath, discussionsPath, prsPath string, includeIssues, includeDiscussions, includePRs, debug bool) error {
	logger := NewLogger(debug)
	
	issues, discussions, prs, err := HydrateFromFiles(issuesPath, discussionsPath, prsPath, includeIssues, includeDiscussions, includePRs)
	if err != nil {
		return err
	}
	
	// Collect and ensure all labels exist before creating content
	labels := CollectLabels(issues, discussions, prs)
	labelSummary := &SectionSummary{Name: "Labels", Total: len(labels)}
	
	logger.Debug("Found %d unique labels to ensure exist: %v", len(labels), labels)
	
	if err := EnsureLabelsExist(client, labels, logger, labelSummary); err != nil {
		return fmt.Errorf("failed to ensure labels exist: %w", err)
	}
	
	// Report label summary
	logger.Info("Labels: %d total, %d successful, %d failed", labelSummary.Total, labelSummary.Success, labelSummary.Failures)
	
	var allErrors []string
	
	// Create issues
	if includeIssues {
		issueSummary := &SectionSummary{Name: "Issues", Total: len(issues)}
		logger.Debug("Creating %d issues", len(issues))
		
		for i, issue := range issues {
			issueInput := githubapi.IssueInput{
				Title:     issue.Title,
				Body:      issue.Body,
				Labels:    issue.Labels,
				Assignees: issue.Assignees,
			}
			if err := client.CreateIssue(issueInput); err != nil {
				errorMsg := fmt.Sprintf("Issue %d (%s): %v", i+1, issue.Title, err)
				allErrors = append(allErrors, errorMsg)
				issueSummary.Errors = append(issueSummary.Errors, errorMsg)
				issueSummary.Failures++
				logger.Debug("Failed to create issue '%s': %v", issue.Title, err)
			} else {
				issueSummary.Success++
				logger.Debug("Successfully created issue '%s'", issue.Title)
			}
		}
		
		logger.Info("Issues: %d total, %d successful, %d failed", issueSummary.Total, issueSummary.Success, issueSummary.Failures)
	}
	
	// Create discussions
	if includeDiscussions {
		discussionSummary := &SectionSummary{Name: "Discussions", Total: len(discussions)}
		logger.Debug("Creating %d discussions", len(discussions))
		
		for i, discussion := range discussions {
			discussionInput := githubapi.DiscussionInput{
				Title:    discussion.Title,
				Body:     discussion.Body,
				Category: discussion.Category,
				Labels:   discussion.Labels,
			}
			if err := client.CreateDiscussion(discussionInput); err != nil {
				errorMsg := fmt.Sprintf("Discussion %d (%s): %v", i+1, discussion.Title, err)
				allErrors = append(allErrors, errorMsg)
				discussionSummary.Errors = append(discussionSummary.Errors, errorMsg)
				discussionSummary.Failures++
				logger.Debug("Failed to create discussion '%s': %v", discussion.Title, err)
			} else {
				discussionSummary.Success++
				logger.Debug("Successfully created discussion '%s'", discussion.Title)
			}
		}
		
		logger.Info("Discussions: %d total, %d successful, %d failed", discussionSummary.Total, discussionSummary.Success, discussionSummary.Failures)
	}
	
	// Create pull requests
	if includePRs {
		prSummary := &SectionSummary{Name: "Pull Requests", Total: len(prs)}
		logger.Debug("Creating %d pull requests", len(prs))
		
		for i, pr := range prs {
			prInput := githubapi.PRInput{
				Title:     pr.Title,
				Body:      pr.Body,
				Head:      pr.Head,
				Base:      pr.Base,
				Labels:    pr.Labels,
				Assignees: pr.Assignees,
			}
			if err := client.CreatePR(prInput); err != nil {
				errorMsg := fmt.Sprintf("Pull Request %d (%s): %v", i+1, pr.Title, err)
				allErrors = append(allErrors, errorMsg)
				prSummary.Errors = append(prSummary.Errors, errorMsg)
				prSummary.Failures++
				logger.Debug("Failed to create pull request '%s': %v", pr.Title, err)
			} else {
				prSummary.Success++
				logger.Debug("Successfully created pull request '%s'", pr.Title)
			}
		}
		
		logger.Info("Pull Requests: %d total, %d successful, %d failed", prSummary.Total, prSummary.Success, prSummary.Failures)
	}
	
	// If any errors occurred, return them as a combined error but don't fail completely
	if len(allErrors) > 0 {
		return fmt.Errorf("some items failed to create:\n  - %s", strings.Join(allErrors, "\n  - "))
	}
	
	return nil
}

type Issue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels"`
	Assignees []string `json:"assignees"`
}

type Discussion struct {
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Category string   `json:"category"`
	Labels   []string `json:"labels"`
}

type PullRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Head      string   `json:"head"`
	Base      string   `json:"base"`
	Labels    []string `json:"labels"`
	Assignees []string `json:"assignees"`
}

// EnsureLabelsExist checks if each label exists in the repo, and creates it if not.
func EnsureLabelsExist(client githubapi.GitHubClient, labels []string, logger *Logger, summary *SectionSummary) error {
	if len(labels) == 0 {
		return nil
	}
	
	logger.Debug("Fetching existing labels from repository")
	existing, err := client.ListLabels()
	if err != nil {
		return err
	}
	
	existSet := make(map[string]struct{}, len(existing))
	for _, l := range existing {
		existSet[l] = struct{}{}
	}
	
	logger.Debug("Found %d existing labels in repository", len(existing))
	
	for _, l := range labels {
		if _, ok := existSet[l]; !ok {
			logger.Debug("Creating missing label '%s'", l)
			if err := client.CreateLabel(l); err != nil {
				errorMsg := fmt.Sprintf("Label '%s': %v", l, err)
				summary.Errors = append(summary.Errors, errorMsg)
				summary.Failures++
				logger.Debug("Failed to create label '%s': %v", l, err)
			} else {
				summary.Success++
				logger.Debug("Successfully created label '%s'", l)
			}
		} else {
			summary.Success++
			logger.Debug("Label '%s' already exists", l)
		}
	}
	
	return nil
}

func HydrateFromFiles(issuesPath, discussionsPath, prsPath string, includeIssues, includeDiscussions, includePRs bool) ([]Issue, []Discussion, []PullRequest, error) {
	var issues []Issue
	var discussions []Discussion
	var prs []PullRequest

	if includeIssues {
		data, err := os.ReadFile(issuesPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &issues); err != nil {
			return nil, nil, nil, err
		}
	}

	if includeDiscussions {
		data, err := os.ReadFile(discussionsPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &discussions); err != nil {
			return nil, nil, nil, err
		}
	}

	if includePRs {
		data, err := os.ReadFile(prsPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := json.Unmarshal(data, &prs); err != nil {
			return nil, nil, nil, err
		}
	}

	return issues, discussions, prs, nil
}

// CollectLabels returns a deduplicated list of all labels used in issues, discussions, and PRs.
func CollectLabels(issues []Issue, discussions []Discussion, prs []PullRequest) []string {
	labelSet := make(map[string]struct{})
	for _, i := range issues {
		for _, l := range i.Labels {
			labelSet[l] = struct{}{}
		}
	}
	for _, d := range discussions {
		for _, l := range d.Labels {
			labelSet[l] = struct{}{}
		}
	}
	for _, p := range prs {
		for _, l := range p.Labels {
			labelSet[l] = struct{}{}
		}
	}
	out := make([]string, 0, len(labelSet))
	for l := range labelSet {
		out = append(out, l)
	}
	return out
}

// FindProjectRoot traverses up from the current file to find the directory containing go.mod
func FindProjectRoot() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
