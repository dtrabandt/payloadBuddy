package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
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

// NOTE: TestMain_PortHandling was removed because:
// - Port validation logic is already thoroughly tested in unit tests (main_test.go)
// - Integration tests were redundant and caused hanging when trying to bind to busy ports
// - Integration tests should focus on end-to-end behavior, not duplicate unit test coverage

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
		// NOTE: port_validation test removed - redundant with unit tests and causes hanging
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := runCommandWithOutput(testBinary, tt.args...)
			checkOutputContains(t, output, tt.expected, tt.description)
		})
	}
}
