package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

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

func createTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", filename, err)
	}
	return filePath
}

func buildVerifyTestBinary(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	testBinary := filepath.Join(tempDir, "payloadBuddy-verify-test")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build test binary: %v\n%s", err, out)
	}
	return testBinary
}

func TestVerifyScenarioFile_ValidationLogic(t *testing.T) {
	tempDir := t.TempDir()

	validFile := createTestFile(t, tempDir, "valid.json", validScenarioTemplate)
	invalidFile := createTestFile(t, tempDir, "invalid.json", invalidScenarioTemplate)
	malformedFile := createTestFile(t, tempDir, "malformed.json", malformedJSONTemplate)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{"valid scenario file", validFile, false},
		{"invalid scenario missing name", invalidFile, true},
		{"malformed JSON", malformedFile, true},
		{"non-existent file", filepath.Join(tempDir, "nonexistent.json"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.filePath); os.IsNotExist(err) {
				if !tt.expectError {
					t.Errorf("File should exist: %s", tt.filePath)
				}
				return
			}

			content, err := os.ReadFile(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			validator := scenarios.NewValidator()
			scenario, err := validator.ValidateJSON(content)

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
				if scenario == nil {
					t.Error("Expected scenario to be returned, but got nil")
				}
				if scenario != nil && scenario.ScenarioName != "Test Scenario" {
					t.Errorf("Expected scenario name 'Test Scenario', got %q", scenario.ScenarioName)
				}
			}
		})
	}
}

func TestVerifyFlagHandling(t *testing.T) {
	if paramVerify == nil {
		t.Error("paramVerify flag should be defined")
	}
	if *paramVerify != "" {
		t.Errorf("paramVerify default value should be empty, got %q", *paramVerify)
	}
}

func TestVerifyScenarioFile_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	validFile := createTestFile(t, tempDir, "valid.json", validScenarioTemplate)
	invalidFile := createTestFile(t, tempDir, "invalid.json", invalidScenarioTemplate)
	testBinary := buildVerifyTestBinary(t)

	tests := []struct {
		name             string
		filePath         string
		expectError      bool
		expectedInOutput string
	}{
		{"valid scenario file", validFile, false, "✅ Validation successful!"},
		{"invalid scenario file", invalidFile, true, "❌ validation failed:"},
		{"non-existent file", filepath.Join(tempDir, "nonexistent.json"), true, "❌ Error: File does not exist:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(testBinary, "-verify", tt.filePath)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected command to fail, but it succeeded. Output: %s", outputStr)
				}
				if !strings.Contains(outputStr, tt.expectedInOutput) {
					t.Errorf("Expected output to contain %q, got: %s", tt.expectedInOutput, outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Expected command to succeed, got error: %v. Output: %s", err, outputStr)
				}
				if !strings.Contains(outputStr, tt.expectedInOutput) {
					t.Errorf("Expected output to contain %q, got: %s", tt.expectedInOutput, outputStr)
				}
				if !strings.Contains(outputStr, "📋 Scenario Details:") {
					t.Errorf("Expected scenario details in output, got: %s", outputStr)
				}
				if !strings.Contains(outputStr, "Name: Test Scenario") {
					t.Errorf("Expected scenario name in output, got: %s", outputStr)
				}
			}
		})
	}
}

func TestVerifyScenarioFile_OutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	testFile := createTestFile(t, tempDir, "comprehensive.json", comprehensiveScenarioTemplate)
	testBinary := buildVerifyTestBinary(t)

	cmd := exec.Command(testBinary, "-verify", testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Verification command failed: %v. Output: %s", err, string(output))
	}

	outputStr := string(output)
	for _, expected := range []string{
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
	} {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain %q, got: %s", expected, outputStr)
		}
	}
}

func TestValidScenarioExamples(t *testing.T) {
	validator := scenarios.NewValidator()

	examples := []struct {
		name     string
		scenario string
	}{
		{"valid scenario template", validScenarioTemplate},
		{"comprehensive scenario template", comprehensiveScenarioTemplate},
		{"minimal custom scenario", `{
			"schema_version": "1.0.0",
			"scenario_name": "Minimal Test",
			"scenario_type": "custom",
			"base_delay": "50ms"
		}`},
	}

	for _, example := range examples {
		t.Run(example.name, func(t *testing.T) {
			scenario, err := validator.ValidateJSON([]byte(example.scenario))
			if err != nil {
				t.Errorf("Example should be valid, got error: %v", err)
			}
			if scenario == nil {
				t.Error("Expected scenario to be returned, got nil")
			}
		})
	}
}
