package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestVerifyScenarioFile_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	// Build the test binary to use for subprocess testing
	testBinary := filepath.Join(tempDir, "payloadBuddy-test")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	tests := []struct {
		name             string
		filePath         string
		expectError      bool
		expectedInOutput string
		description      string
	}{
		{
			name:             "Valid scenario file",
			filePath:         validFile,
			expectError:      false,
			expectedInOutput: "✅ Validation successful!",
			description:      "Should validate successfully and show success message",
		},
		{
			name:             "Invalid scenario file",
			filePath:         invalidFile,
			expectError:      true,
			expectedInOutput: "❌ Validation failed:",
			description:      "Should fail validation due to missing scenario_name",
		},
		{
			name:             "Non-existent file",
			filePath:         filepath.Join(tempDir, "nonexistent.json"),
			expectError:      true,
			expectedInOutput: "❌ Error: File does not exist:",
			description:      "Should fail due to file not existing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the binary with -verify flag as a subprocess
			cmd := exec.Command(testBinary, "-verify", tt.filePath)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected command to fail, but it succeeded. Output: %s", outputStr)
				}
				if !strings.Contains(outputStr, tt.expectedInOutput) {
					t.Errorf("Expected output to contain %q, but got: %s", tt.expectedInOutput, outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Expected command to succeed, but it failed with error: %v. Output: %s", err, outputStr)
				}
				if !strings.Contains(outputStr, tt.expectedInOutput) {
					t.Errorf("Expected output to contain %q, but got: %s", tt.expectedInOutput, outputStr)
				}
				// For valid files, also check that scenario details are displayed
				if !strings.Contains(outputStr, "📋 Scenario Details:") {
					t.Errorf("Expected output to contain scenario details, but got: %s", outputStr)
				}
				if !strings.Contains(outputStr, "Name: Test Scenario") {
					t.Errorf("Expected output to contain scenario name, but got: %s", outputStr)
				}
			}
		})
	}
}

func TestVerifyScenarioFile_OutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a comprehensive test scenario to verify all output fields
	tempDir := t.TempDir()

	comprehensiveScenario := `{
		"schema_version": "1.0.0",
		"scenario_name": "Comprehensive Test Scenario",
		"description": "A comprehensive test scenario with all fields",
		"scenario_type": "custom",
		"base_delay": "150ms",
		"delay_strategy": "progressive",
		"servicenow_mode": true,
		"batch_size": 25,
		"response_limits": {
			"max_count": 5000,
			"default_count": 500
		},
		"metadata": {
			"author": "Test Author",
			"version": "1.2.3",
			"tags": ["test", "comprehensive", "validation"]
		}
	}`

	testFile := filepath.Join(tempDir, "comprehensive.json")
	err := os.WriteFile(testFile, []byte(comprehensiveScenario), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Build test binary
	testBinary := filepath.Join(tempDir, "payloadBuddy-test")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	// Run verification
	cmd = exec.Command(testBinary, "-verify", testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Verification command failed: %v. Output: %s", err, string(output))
	}

	outputStr := string(output)

	// Check for all expected output components
	expectedOutputs := []string{
		fmt.Sprintf("Validating scenario file: %s", testFile),
		"✅ Validation successful!",
		"📋 Scenario Details:",
		"Name: Comprehensive Test Scenario",
		"Type: custom",
		"Base Delay: 150ms",
		"Delay Strategy: progressive",
		"ServiceNow Mode: enabled",
		"Batch Size: 25",
		"Max Count: 5000",
		"Default Count: 500",
		"Description: A comprehensive test scenario with all fields",
		"Author: Test Author",
		"Version: 1.2.3",
		"Tags: [test comprehensive validation]",
		"🎯 Usage: Use this scenario with ?scenario=custom",
		"💡 Tip: Place this file in $HOME/.config/payloadBuddy/scenarios/ to make it available",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain %q, but it was missing from: %s", expected, outputStr)
		}
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
