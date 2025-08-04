package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Test scenario templates for reuse
var (
	validScenarioTemplate = `{
		"schema_version": "1.0.0",
		"scenario_name": "Test Scenario",
		"scenario_type": "custom",
		"base_delay": "100ms",
		"delay_strategy": "fixed",
		"servicenow_mode": true,
		"description": "A test scenario for validation"
	}`

	invalidScenarioTemplate = `{
		"schema_version": "1.0.0",
		"scenario_type": "custom",
		"base_delay": "100ms"
	}`

	malformedJSONTemplate = `{
		"schema_version": "1.0.0",
		"scenario_name": "Test"
		"scenario_type": "custom"
	}`

	comprehensiveScenarioTemplate = `{
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
)

// Helper function to create test files
func createTestFile(t *testing.T, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", filename, err)
	}
	return filePath
}

// Helper function to build test binary (duplicated from main_integration_test.go)
func buildVerifyTestBinary(t *testing.T) string {
	tempDir := t.TempDir()
	testBinary := filepath.Join(tempDir, "payloadBuddy-verify-test")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	return testBinary
}

func TestVerifyScenarioFile_ValidationLogic(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create test files using helper function
	validFile := createTestFile(t, tempDir, "valid.json", validScenarioTemplate)
	invalidFile := createTestFile(t, tempDir, "invalid.json", invalidScenarioTemplate) 
	malformedFile := createTestFile(t, tempDir, "malformed.json", malformedJSONTemplate)

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

	// Create test files using helper function and templates
	validFile := createTestFile(t, tempDir, "valid.json", validScenarioTemplate)
	invalidFile := createTestFile(t, tempDir, "invalid.json", invalidScenarioTemplate)

	// Build the test binary to use for subprocess testing
	testBinary := buildVerifyTestBinary(t)

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
			expectedInOutput: "❌ validation failed:",
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

	// Create test file using helper function and template
	testFile := createTestFile(t, tempDir, "comprehensive.json", comprehensiveScenarioTemplate)

	// Build test binary
	testBinary := buildVerifyTestBinary(t)

	// Run verification
	cmd := exec.Command(testBinary, "-verify", testFile)
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

	// Use templates and add additional examples
	examples := []struct {
		name     string
		scenario string
	}{
		{
			name:     "Valid scenario template",
			scenario: validScenarioTemplate,
		},
		{
			name:     "Comprehensive scenario template", 
			scenario: comprehensiveScenarioTemplate,
		},
		{
			name: "Minimal custom scenario",
			scenario: `{
				"schema_version": "1.0.0",
				"scenario_name": "Minimal Test",
				"scenario_type": "custom",
				"base_delay": "50ms"
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
