package githubapi

import (
	"strings"
	"testing"
)

// TestGraphQLMutationSyntax tests that all GraphQL mutations have valid syntax
func TestGraphQLMutationSyntax(t *testing.T) {
	mutations := []struct {
		name     string
		mutation string
	}{
		{
			name:     "createIssueMutation",
			mutation: createIssueMutation,
		},
		{
			name:     "createDiscussionMutation",
			mutation: createDiscussionMutation,
		},
		{
			name:     "createPullRequestMutation",
			mutation: createPullRequestMutation,
		},
		{
			name:     "createLabelMutation",
			mutation: createLabelMutation,
		},
		{
			name:     "addLabelsToLabelableMutation",
			mutation: addLabelsToLabelableMutation,
		},
		{
			name:     "addAssigneesToAssignableMutation",
			mutation: addAssigneesToAssignableMutation,
		},
		{
			name:     "deleteDiscussionMutation",
			mutation: deleteDiscussionMutation,
		},
		{
			name:     "deleteIssueMutation",
			mutation: deleteIssueMutation,
		},
		{
			name:     "deletePullRequestMutation",
			mutation: deletePullRequestMutation,
		},
		{
			name:     "deleteLabelMutation",
			mutation: deleteLabelMutation,
		},
	}

	for _, tt := range mutations {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mutation == "" {
				t.Error("Mutation should not be empty")
				return
			}

			// Basic GraphQL syntax validation
			if !strings.Contains(tt.mutation, "mutation") {
				t.Error("Mutation should contain 'mutation' keyword")
			}

			// Check for proper opening and closing braces
			openBraces := strings.Count(tt.mutation, "{")
			closeBraces := strings.Count(tt.mutation, "}")
			if openBraces != closeBraces {
				t.Errorf("Unmatched braces in mutation: %d open, %d close", openBraces, closeBraces)
			}

			// Check for variable declarations (should start with $)
			if strings.Contains(tt.mutation, "mutation") && strings.Contains(tt.mutation, "(") {
				// If mutation has parameters, they should use $ syntax
				if strings.Contains(tt.mutation, ":") && !strings.Contains(tt.mutation, "$") {
					t.Error("Mutation parameters should use $ variable syntax")
				}
			}
		})
	}
}

// TestGraphQLQuerySyntax tests that all GraphQL queries have valid syntax
func TestGraphQLQuerySyntax(t *testing.T) {
	queries := []struct {
		name  string
		query string
	}{
		{
			name:  "listLabelsQuery",
			query: listLabelsQuery,
		},
		{
			name:  "repositoryWithDiscussionCategoriesQuery",
			query: repositoryWithDiscussionCategoriesQuery,
		},
		{
			name:  "getLabelByNameQuery",
			query: getLabelByNameQuery,
		},
		{
			name:  "listIssuesQuery",
			query: listIssuesQuery,
		},
		{
			name:  "listDiscussionsQuery",
			query: listDiscussionsQuery,
		},
		{
			name:  "listPullRequestsQuery",
			query: listPullRequestsQuery,
		},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			if tt.query == "" {
				t.Error("Query should not be empty")
				return
			}

			// Basic GraphQL syntax validation
			if !strings.Contains(tt.query, "query") {
				t.Error("Query should contain 'query' keyword")
			}

			// Check for proper opening and closing braces
			openBraces := strings.Count(tt.query, "{")
			closeBraces := strings.Count(tt.query, "}")
			if openBraces != closeBraces {
				t.Errorf("Unmatched braces in query: %d open, %d close", openBraces, closeBraces)
			}

			// Check for variable declarations (should start with $)
			if strings.Contains(tt.query, "query") && strings.Contains(tt.query, "(") {
				// If query has parameters, they should use $ syntax
				if strings.Contains(tt.query, ":") && !strings.Contains(tt.query, "$") {
					t.Error("Query parameters should use $ variable syntax")
				}
			}
		})
	}
}

// TestMutationVariableSubstitution tests that mutations properly define variables
func TestMutationVariableSubstitution(t *testing.T) {
	tests := []struct {
		name            string
		mutation        string
		expectedVars    []string
		mustContainVars []string
	}{
		{
			name:         "createIssueMutation variables",
			mutation:     createIssueMutation,
			expectedVars: []string{"$repositoryId", "$title", "$body"},
		},
		{
			name:         "createDiscussionMutation variables",
			mutation:     createDiscussionMutation,
			expectedVars: []string{"$input"},
		},
		{
			name:         "createPullRequestMutation variables",
			mutation:     createPullRequestMutation,
			expectedVars: []string{"$repositoryId", "$title", "$body", "$headRefName", "$baseRefName"},
		},
		{
			name:         "deleteIssueMutation variables",
			mutation:     deleteIssueMutation,
			expectedVars: []string{"$issueId"},
		},
		{
			name:         "deleteLabelMutation variables",
			mutation:     deleteLabelMutation,
			expectedVars: []string{"$labelId"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expectedVar := range tt.expectedVars {
				if !strings.Contains(tt.mutation, expectedVar) {
					t.Errorf("Mutation should contain variable %s", expectedVar)
				}
			}

			for _, mustContainVar := range tt.mustContainVars {
				if !strings.Contains(tt.mutation, mustContainVar) {
					t.Errorf("Mutation must contain variable %s", mustContainVar)
				}
			}
		})
	}
}

// TestQueryVariableSubstitution tests that queries properly define variables
func TestQueryVariableSubstitution(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		expectedVars []string
	}{
		{
			name:         "listLabelsQuery variables",
			query:        listLabelsQuery,
			expectedVars: []string{"$owner", "$name"},
		},
		{
			name:         "repositoryWithDiscussionCategoriesQuery variables",
			query:        repositoryWithDiscussionCategoriesQuery,
			expectedVars: []string{"$owner", "$name"},
		},
		{
			name:         "getLabelByNameQuery variables",
			query:        getLabelByNameQuery,
			expectedVars: []string{"$owner", "$name", "$labelName"},
		},
		{
			name:         "listIssuesQuery variables",
			query:        listIssuesQuery,
			expectedVars: []string{"$owner", "$name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expectedVar := range tt.expectedVars {
				if !strings.Contains(tt.query, expectedVar) {
					t.Errorf("Query should contain variable %s", expectedVar)
				}
			}
		})
	}
}

// TestMutationCompleteness tests that mutations return appropriate fields
func TestMutationCompleteness(t *testing.T) {
	tests := []struct {
		name               string
		mutation           string
		expectedReturnType string
		expectedFields     []string
	}{
		{
			name:               "createIssueMutation return fields",
			mutation:           createIssueMutation,
			expectedReturnType: "createIssue",
			expectedFields:     []string{"issue", "id", "number", "title", "url"},
		},
		{
			name:               "createDiscussionMutation return fields",
			mutation:           createDiscussionMutation,
			expectedReturnType: "createDiscussion",
			expectedFields:     []string{"discussion", "id", "number", "title", "url"},
		},
		{
			name:               "createPullRequestMutation return fields",
			mutation:           createPullRequestMutation,
			expectedReturnType: "createPullRequest",
			expectedFields:     []string{"pullRequest", "id", "number", "title", "url"},
		},
		{
			name:               "deleteIssueMutation return fields",
			mutation:           deleteIssueMutation,
			expectedReturnType: "closeIssue",
			expectedFields:     []string{"issue", "id", "state"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.mutation, tt.expectedReturnType) {
				t.Errorf("Mutation should contain return type %s", tt.expectedReturnType)
			}

			for _, field := range tt.expectedFields {
				if !strings.Contains(tt.mutation, field) {
					t.Errorf("Mutation should return field %s", field)
				}
			}
		})
	}
}

// TestGraphQLConstantsExist tests that all required GraphQL constants are defined
func TestGraphQLConstantsExist(t *testing.T) {
	constants := []struct {
		name  string
		value string
	}{
		{"createIssueMutation", createIssueMutation},
		{"createDiscussionMutation", createDiscussionMutation},
		{"createPullRequestMutation", createPullRequestMutation},
		{"createLabelMutation", createLabelMutation},
		{"addLabelsToLabelableMutation", addLabelsToLabelableMutation},
		{"addAssigneesToAssignableMutation", addAssigneesToAssignableMutation},
		{"deleteDiscussionMutation", deleteDiscussionMutation},
		{"deleteIssueMutation", deleteIssueMutation},
		{"deletePullRequestMutation", deletePullRequestMutation},
		{"deleteLabelMutation", deleteLabelMutation},
		{"listLabelsQuery", listLabelsQuery},
		{"repositoryWithDiscussionCategoriesQuery", repositoryWithDiscussionCategoriesQuery},
		{"getLabelByNameQuery", getLabelByNameQuery},
		{"listIssuesQuery", listIssuesQuery},
		{"listDiscussionsQuery", listDiscussionsQuery},
		{"listPullRequestsQuery", listPullRequestsQuery},
	}

	for _, constant := range constants {
		t.Run(constant.name, func(t *testing.T) {
			if constant.value == "" {
				t.Errorf("GraphQL constant %s should not be empty", constant.name)
			}

			if len(constant.value) < 10 {
				t.Errorf("GraphQL constant %s seems too short (length: %d)", constant.name, len(constant.value))
			}
		})
	}
}
