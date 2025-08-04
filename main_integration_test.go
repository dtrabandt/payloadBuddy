package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Expected output constants for consistent testing
const (
	ExpectedUsageText    = "Usage of"
	ExpectedPortText     = "-port"
	ExpectedVerifyText   = "-verify"
	ExpectedAuthText     = "-auth"
	ExpectedValidateText = "Validating scenario file:"
	ExpectedErrorText    = "❌ Error: File does not exist:"
	ExpectedHTTPText     = "on http://localhost:8080"
)

// Helper function to build test binary - shared across integration tests
func buildTestBinary(t *testing.T) string {
	tempDir := t.TempDir()
	testBinary := filepath.Join(tempDir, "payloadBuddy-test")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	return testBinary
}

// Helper function to run command and get output
func runCommandWithOutput(binary string, args ...string) (string, error) {
	cmd := exec.Command(binary, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper function to check if output contains all expected strings
func checkOutputContains(t *testing.T, output string, expected []string, testName string) {
	for _, expectedStr := range expected {
		if !strings.Contains(output, expectedStr) {
			t.Errorf("%s: Expected output to contain %q, but it was missing from:\n%s", testName, expectedStr, output)
		}
	}
}

func TestMain_StartupOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testBinary := buildTestBinary(t)

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name: "verify_nonexistent_file",
			args: []string{"-verify", "nonexistent.json"},
			expected: []string{
				ExpectedValidateText + " nonexistent.json",
				ExpectedErrorText,
			},
		},
		{
			name: "help_flag_output",
			args: []string{"-h"},
			expected: []string{
				ExpectedUsageText,
				ExpectedAuthText,
				ExpectedPortText,
				ExpectedVerifyText,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := runCommandWithOutput(testBinary, tt.args...)
			checkOutputContains(t, output, tt.expected, tt.name)
		})
	}
}

func TestMain_HelpOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testBinary := buildTestBinary(t)

	output, _ := runCommandWithOutput(testBinary, "-h")

	// Help should show the usage information and all flags
	expectedInHelp := []string{
		ExpectedUsageText,
		ExpectedAuthText,
		ExpectedPortText,
		ExpectedVerifyText,
	}

	checkOutputContains(t, output, expectedInHelp, "help_output")
}

func TestMain_PortHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testBinary := buildTestBinary(t)

	tests := []struct {
		name         string
		portArg      string
		expectedPort string
		description  string
	}{
		{
			name:         "invalid_port_fallback",
			portArg:      "-port=invalid",
			expectedPort: ExpectedHTTPText,
			description:  "Invalid port should fallback to default 8080",
		},
		{
			name:         "out_of_range_port_fallback",
			portArg:      "-port=70000",
			expectedPort: ExpectedHTTPText,
			description:  "Out of range port should fallback to default 8080",
		},
		{
			name:         "negative_port_fallback",
			portArg:      "-port=-1",
			expectedPort: ExpectedHTTPText,
			description:  "Negative port should fallback to default 8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These will try to start the server but fail quickly due to port conflicts
			output, _ := runCommandWithOutput(testBinary, tt.portArg)

			if !strings.Contains(output, tt.expectedPort) {
				t.Errorf("%s: Expected output to contain %q, but got:\n%s", tt.description, tt.expectedPort, output)
			}
		})
	}
}

func TestMain_ComprehensiveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testBinary := buildTestBinary(t)

	// Comprehensive integration test covering multiple scenarios
	tests := []struct {
		name        string
		args        []string
		expected    []string
		description string
	}{
		{
			name: "flag_parsing_help",
			args: []string{"-h"},
			expected: []string{
				ExpectedUsageText,
				ExpectedPortText,
				ExpectedVerifyText,
				ExpectedAuthText,
			},
			description: "Help flag should show all available flags and usage",
		},
		{
			name: "verify_integration",
			args: []string{"-verify", "missing.json"},
			expected: []string{
				ExpectedValidateText,
				ExpectedErrorText,
				"missing.json",
			},
			description: "Verify flag should validate scenario files and show errors",
		},
		{
			name: "port_validation",
			args: []string{"-port=abc123"},
			expected: []string{
				ExpectedHTTPText, // Should fallback to 8080
			},
			description: "Invalid ports should fallback to default 8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := runCommandWithOutput(testBinary, tt.args...)
			checkOutputContains(t, output, tt.expected, tt.description)
		})
	}
}
