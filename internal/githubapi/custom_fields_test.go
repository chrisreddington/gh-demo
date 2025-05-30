package githubapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/chrisreddington/gh-demo/internal/types"
)

// TestCreateProjectV2Field_BasicTypes tests field creation for basic field types (text, number, date)
func TestCreateProjectV2Field_BasicTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		wantType  string
	}{
		{
			name:      "text field",
			fieldType: "text",
			wantType:  "TEXT",
		},
		{
			name:      "number field",
			fieldType: "number", 
			wantType:  "NUMBER",
		},
		{
			name:      "date field",
			fieldType: "date",
			wantType:  "DATE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &ConfigurableMockGraphQLClient{
				DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
					// Verify the mutation uses the correct GraphQL schema
					if !strings.Contains(query, "createProjectV2Field") {
						t.Errorf("Expected createProjectV2Field mutation, got: %s", query)
					}
					
					// Verify variables contain correct data type
					if dataType, ok := variables["dataType"]; !ok || dataType != tt.wantType {
						t.Errorf("Expected dataType %s, got %v", tt.wantType, dataType)
					}
					
					// Mock successful response
					mockResponse := map[string]interface{}{
						"createProjectV2Field": map[string]interface{}{
							"projectV2Field": map[string]interface{}{
								"id":       "field_123",
								"name":     "Test Field",
								"dataType": tt.wantType,
							},
						},
					}
					
					respBytes, _ := json.Marshal(mockResponse)
					return json.Unmarshal(respBytes, response)
				},
			}

			client := createTestClientWithGraphQL(mockClient)
			
			field := types.ProjectV2Field{
				Name: "Test Field",
				Type: tt.fieldType,
			}

			err := client.createProjectV2Field(context.Background(), "project_123", field)
			if err != nil {
				t.Errorf("createProjectV2Field() error = %v", err)
			}
		})
	}
}

// TestCreateProjectV2Field_InvalidType tests validation for unsupported field types
func TestCreateProjectV2Field_InvalidType(t *testing.T) {
	mockClient := &ConfigurableMockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			t.Error("GraphQL client should not be called for invalid field type")
			return nil
		},
	}

	client := createTestClientWithGraphQL(mockClient)
	
	field := types.ProjectV2Field{
		Name: "Test Field",
		Type: "invalid_type",
	}

	err := client.createProjectV2Field(context.Background(), "project_123", field)
	if err == nil {
		t.Error("Expected error for invalid field type, got nil")
	}
	
	if !strings.Contains(err.Error(), "unsupported field type") {
		t.Errorf("Expected unsupported field type error, got: %v", err)
	}
}

// TestCreateProjectV2SingleSelectField tests single select field creation with proper option structure
func TestCreateProjectV2SingleSelectField(t *testing.T) {
	mockClient := &ConfigurableMockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			// Verify the mutation uses the correct single select schema
			if !strings.Contains(query, "createProjectV2Field") {
				t.Errorf("Expected createProjectV2Field mutation, got: %s", query)
			}
			
			if !strings.Contains(query, "singleSelectOptions") {
				t.Errorf("Expected singleSelectOptions in mutation, got: %s", query)
			}
			
			// Verify options structure
			options, ok := variables["options"].([]map[string]interface{})
			if !ok {
				t.Errorf("Expected options as []map[string]interface{}, got %T", variables["options"])
			}
			
			if len(options) != 2 {
				t.Errorf("Expected 2 options, got %d", len(options))
			}
			
			// Verify first option has required fields
			option1 := options[0]
			if option1["name"] != "High" {
				t.Errorf("Expected option name 'High', got %v", option1["name"])
			}
			if option1["description"] != "High priority" {
				t.Errorf("Expected option description 'High priority', got %v", option1["description"])
			}
			if option1["color"] != "RED" {
				t.Errorf("Expected option color 'RED', got %v", option1["color"])
			}
			
			// Verify second option with fallback description
			option2 := options[1]
			if option2["name"] != "Low" {
				t.Errorf("Expected option name 'Low', got %v", option2["name"])
			}
			if option2["description"] != "Low" { // Should fallback to name
				t.Errorf("Expected option description 'Low' (fallback), got %v", option2["description"])
			}
			if option2["color"] != "GRAY" { // Should default to GRAY
				t.Errorf("Expected option color 'GRAY' (default), got %v", option2["color"])
			}
			
			// Mock successful response
			mockResponse := map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id":       "field_123",
						"name":     "Priority",
						"dataType": "SINGLE_SELECT",
					},
				},
			}
			
			respBytes, _ := json.Marshal(mockResponse)
			return json.Unmarshal(respBytes, response)
		},
	}

	client := createTestClientWithGraphQL(mockClient)
	
	field := types.ProjectV2Field{
		Name: "Priority",
		Type: "single_select",
		Options: []types.ProjectV2FieldOption{
			{
				Name:        "High",
				Description: "High priority",
				Color:       "red", // Should be converted to RED
			},
			{
				Name:        "Low",
				Description: "", // Should fallback to name
				Color:       "",  // Should default to GRAY
			},
		},
	}

	err := client.createProjectV2SingleSelectField(context.Background(), "project_123", field)
	if err != nil {
		t.Errorf("createProjectV2SingleSelectField() error = %v", err)
	}
}

// TestCreateProjectV2SingleSelectField_NoOptions tests validation for single select without options
func TestCreateProjectV2SingleSelectField_NoOptions(t *testing.T) {
	mockClient := &ConfigurableMockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			t.Error("GraphQL client should not be called for single select field without options")
			return nil
		},
	}

	client := createTestClientWithGraphQL(mockClient)
	
	field := types.ProjectV2Field{
		Name:    "Priority",
		Type:    "single_select",
		Options: []types.ProjectV2FieldOption{}, // Empty options
	}

	err := client.createProjectV2SingleSelectField(context.Background(), "project_123", field)
	if err == nil {
		t.Error("Expected error for single select field without options, got nil")
	}
	
	if !strings.Contains(err.Error(), "must have at least one option") {
		t.Errorf("Expected validation error for missing options, got: %v", err)
	}
}

// TestCreateProjectV2Field_RoutesToSingleSelect tests that single_select type routes to the correct method
func TestCreateProjectV2Field_RoutesToSingleSelect(t *testing.T) {
	singleSelectCalled := false
	
	mockClient := &ConfigurableMockGraphQLClient{
		DoFunc: func(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
			// This should be the single select mutation
			if strings.Contains(query, "singleSelectOptions") {
				singleSelectCalled = true
			}
			
			// Mock successful response
			mockResponse := map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id":       "field_123",
						"name":     "Priority",
						"dataType": "SINGLE_SELECT",
					},
				},
			}
			
			respBytes, _ := json.Marshal(mockResponse)
			return json.Unmarshal(respBytes, response)
		},
	}

	client := createTestClientWithGraphQL(mockClient)
	
	field := types.ProjectV2Field{
		Name: "Priority",
		Type: "single_select",
		Options: []types.ProjectV2FieldOption{
			{Name: "High", Description: "High priority", Color: "red"},
		},
	}

	err := client.createProjectV2Field(context.Background(), "project_123", field)
	if err != nil {
		t.Errorf("createProjectV2Field() error = %v", err)
	}
	
	if !singleSelectCalled {
		t.Error("Expected single select mutation to be called for single_select field type")
	}
}

// createTestClientWithGraphQL creates a test client with the provided GraphQL client
func createTestClientWithGraphQL(gqlClient GraphQLClient) *GHClient {
	client, _ := NewGHClientWithClients("test-owner", "test-repo", gqlClient)
	return client
}
