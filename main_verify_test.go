package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyScenarioFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	
	// Test case 1: Valid scenario file
	validScenario := `{
		"schema_version": "1.0.0",
		"scenario_name": "Test Scenario",
		"scenario_type": "custom",
		"base_delay": "100ms",
		"delay_strategy": "fixed",
		"servicenow_mode": true,
		"description": "A test scenario for validation"
	}`
	
	validFile := filepath.Join(tempDir, "valid.json")
	err := os.WriteFile(validFile, []byte(validScenario), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid test file: %v", err)
	}
	
	// Test case 2: Invalid scenario file (missing required field)
	invalidScenario := `{
		"schema_version": "1.0.0",
		"scenario_type": "custom",
		"base_delay": "100ms"
	}`
	
	invalidFile := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte(invalidScenario), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}
	
	// Test case 3: Malformed JSON
	malformedJSON := `{
		"schema_version": "1.0.0",
		"scenario_name": "Test"
		"scenario_type": "custom"
	}`
	
	malformedFile := filepath.Join(tempDir, "malformed.json")
	err = os.WriteFile(malformedFile, []byte(malformedJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create malformed test file: %v", err)
	}
	
	tests := []struct {
		name        string
		filePath    string
		expectError bool
		description string
	}{
		{
			name:        "Valid scenario file",
			filePath:    validFile,
			expectError: false,
			description: "Should validate successfully",
		},
		{
			name:        "Invalid scenario file",
			filePath:    invalidFile,
			expectError: true,
			description: "Should fail validation due to missing scenario_name",
		},
		{
			name:        "Malformed JSON",
			filePath:    malformedFile,
			expectError: true,
			description: "Should fail due to invalid JSON syntax",
		},
		{
			name:        "Non-existent file",
			filePath:    filepath.Join(tempDir, "nonexistent.json"),
			expectError: true,
			description: "Should fail due to file not existing",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture the exit behavior by testing the validation logic directly
			// instead of calling verifyScenarioFile which calls os.Exit
			
			if tt.filePath == filepath.Join(tempDir, "nonexistent.json") {
				// Test file existence check
				if _, err := os.Stat(tt.filePath); !os.IsNotExist(err) {
					t.Errorf("Expected file to not exist, but it does")
				}
				return
			}
			
			// Read the file
			content, err := os.ReadFile(tt.filePath)
			if err != nil {
				if !tt.expectError {
					t.Errorf("Unexpected error reading file: %v", err)
				}
				return
			}
			
			// Test validation
			validator := NewScenarioValidator()
			scenario, err := validator.ValidateJSON(content)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
				if scenario == nil {
					t.Errorf("Expected scenario to be returned, but got nil")
				}
				if scenario != nil && scenario.ScenarioName != "Test Scenario" {
					t.Errorf("Expected scenario name 'Test Scenario', got '%s'", scenario.ScenarioName)
				}
			}
		})
	}
}

func TestVerifyFlagHandling(t *testing.T) {
	// Test the flag parsing logic by checking that paramVerify is properly defined
	if paramVerify == nil {
		t.Error("paramVerify flag should be defined")
	}
	
	// Test default value
	if *paramVerify != "" {
		t.Errorf("paramVerify default value should be empty, got '%s'", *paramVerify)
	}
}

func TestValidScenarioExamples(t *testing.T) {
	// Test validation of the built-in scenario examples
	validator := NewScenarioValidator()
	
	// Example scenarios that should be valid
	examples := []struct {
		name     string
		scenario string
	}{
		{
			name: "Minimal custom scenario",
			scenario: `{
				"schema_version": "1.0.0",
				"scenario_name": "Minimal Test",
				"scenario_type": "custom",
				"base_delay": "50ms"
			}`,
		},
		{
			name: "Full featured scenario",
			scenario: `{
				"schema_version": "1.0.0",
				"scenario_name": "Full Featured Test",
				"description": "A comprehensive test scenario",
				"scenario_type": "custom",
				"base_delay": "100ms",
				"delay_strategy": "progressive",
				"servicenow_mode": true,
				"batch_size": 50,
				"response_limits": {
					"max_count": 10000,
					"default_count": 1000
				},
				"servicenow_config": {
					"record_types": ["incident"],
					"state_rotation": ["New", "In Progress", "Closed"],
					"number_format": "INC%07d",
					"sys_id_format": "standard"
				},
				"metadata": {
					"author": "Test Author",
					"created_date": "2025-01-15",
					"version": "1.0.0",
					"tags": ["test", "example"]
				}
			}`,
		},
	}
	
	for _, example := range examples {
		t.Run(example.name, func(t *testing.T) {
			scenario, err := validator.ValidateJSON([]byte(example.scenario))
			if err != nil {
				t.Errorf("Example scenario should be valid, but got error: %v", err)
			}
			if scenario == nil {
				t.Error("Expected scenario to be returned, but got nil")
			}
		})
	}
}