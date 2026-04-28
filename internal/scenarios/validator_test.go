package scenarios

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScenarioValidator(t *testing.T) {
	validator := NewValidator()

	// Test valid scenario
	validScenario := Scenario{
		SchemaVersion:  "1.0.0",
		ScenarioName:   "Test Scenario",
		Description:    "A test scenario for validation",
		ScenarioType:   "custom",
		BaseDelay:      "100ms",
		DelayStrategy:  "fixed",
		ServiceNowMode: true,
		BatchSize:      50,
		ResponseLimits: &ResponseLimits{
			MaxCount:     50000,
			DefaultCount: 5000,
		},
		ServiceNowConfig: &ServiceNowConfig{
			RecordTypes:   []string{"incident", "problem"},
			StateRotation: []string{"New", "In Progress", "Resolved"},
			NumberFormat:  "INC%07d",
			SysIDFormat:   "standard",
		},
		Metadata: &ScenarioMetadata{
			Author:      "Test Author",
			CreatedDate: "2025-01-15",
			Version:     "1.0.0",
			Tags:        []string{"test", "validation"},
			Compatibility: &CompatibilityInfo{
				MinPayloadBuddyVersion: "1.0.0",
				TestedVersions:         []string{"1.0.0", "1.1.0"},
			},
		},
	}

	err := validator.ValidateScenario(&validScenario)
	if err != nil {
		t.Errorf("Valid scenario failed validation: %v", err)
	}
}

func TestScenarioValidatorRequiredFields(t *testing.T) {
	validator := NewValidator()

	// Test missing scenario_name
	scenario := Scenario{
		ScenarioType: "custom",
		BaseDelay:    "100ms",
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "scenario_name is required") {
		t.Errorf("Expected scenario_name required error, got: %v", err)
	}

	// Test missing scenario_type
	scenario = Scenario{
		ScenarioName: "Test",
		BaseDelay:    "100ms",
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "scenario_type is required") {
		t.Errorf("Expected scenario_type required error, got: %v", err)
	}

	// Test missing base_delay
	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "base_delay is required") {
		t.Errorf("Expected base_delay required error, got: %v", err)
	}
}

func TestScenarioValidatorEnums(t *testing.T) {
	validator := NewValidator()

	// Test invalid scenario_type
	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "invalid_type",
		BaseDelay:    "100ms",
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "scenario_type must be one of") {
		t.Errorf("Expected scenario_type enum error, got: %v", err)
	}

	// Test invalid delay_strategy
	scenario = Scenario{
		ScenarioName:  "Test",
		ScenarioType:  "custom",
		BaseDelay:     "100ms",
		DelayStrategy: "invalid_strategy",
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "delay_strategy must be one of") {
		t.Errorf("Expected delay_strategy enum error, got: %v", err)
	}
}

func TestScenarioValidatorDelayFormat(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		delay     string
		shouldErr bool
	}{
		{"100ms", false},
		{"1s", false},
		{"500", false},
		{"1.5s", false},
		{"invalid", true},
		{"", true},
		{"100xyz", true},
	}

	for _, tc := range testCases {
		scenario := Scenario{
			ScenarioName: "Test",
			ScenarioType: "custom",
			BaseDelay:    tc.delay,
		}

		err := validator.ValidateScenario(&scenario)
		if tc.shouldErr && err == nil {
			t.Errorf("Expected error for delay %s, but validation passed", tc.delay)
		}
		if !tc.shouldErr && err != nil {
			t.Errorf("Unexpected error for delay %s: %v", tc.delay, err)
		}
	}
}

func TestScenarioValidatorResponseLimits(t *testing.T) {
	validator := NewValidator()

	// Test invalid max_count
	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ResponseLimits: &ResponseLimits{
			MaxCount: 2000000, // exceeds limit
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "max_count must be between") {
		t.Errorf("Expected max_count validation error, got: %v", err)
	}

	// Test invalid default_count (negative value)
	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ResponseLimits: &ResponseLimits{
			DefaultCount: -1, // below minimum
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "default_count must be between 0 and 1000000") {
		t.Errorf("Expected default_count validation error, got: %v", err)
	}
}

func TestScenarioValidatorServiceNowConfig(t *testing.T) {
	validator := NewValidator()

	// Test invalid record_type
	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ServiceNowConfig: &ServiceNowConfig{
			RecordTypes: []string{"invalid_record_type"},
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "invalid record_type") {
		t.Errorf("Expected record_type validation error, got: %v", err)
	}

	// Test invalid sys_id_format
	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ServiceNowConfig: &ServiceNowConfig{
			SysIDFormat: "invalid_format",
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "sys_id_format must be one of") {
		t.Errorf("Expected sys_id_format validation error, got: %v", err)
	}
}

func TestScenarioValidatorVersionFormat(t *testing.T) {
	validator := NewValidator()

	// Test invalid schema_version
	scenario := Scenario{
		ScenarioName:  "Test",
		ScenarioType:  "custom",
		BaseDelay:     "100ms",
		SchemaVersion: "invalid.version",
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "schema_version validation failed") {
		t.Errorf("Expected schema_version validation error, got: %v", err)
	}

	// Test invalid metadata version
	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		Metadata: &ScenarioMetadata{
			Version: "invalid.version",
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "version validation failed") {
		t.Errorf("Expected metadata version validation error, got: %v", err)
	}
}

func TestScenarioValidatorDateFormat(t *testing.T) {
	validator := NewValidator()

	// Test invalid created_date
	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		Metadata: &ScenarioMetadata{
			CreatedDate: "invalid-date",
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "created_date") {
		t.Errorf("Expected created_date validation error, got: %v", err)
	}

	// Test valid date
	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		Metadata: &ScenarioMetadata{
			CreatedDate: "2025-01-15",
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err != nil {
		t.Errorf("Valid date should not cause validation error: %v", err)
	}
}

func TestValidateJSON(t *testing.T) {
	validator := NewValidator()

	// Test valid JSON
	validJSON := `{
		"scenario_name": "Test Scenario",
		"scenario_type": "custom",
		"base_delay": "100ms"
	}`

	scenario, err := validator.ValidateJSON([]byte(validJSON))
	if err != nil {
		t.Errorf("Valid JSON failed validation: %v", err)
	}
	if scenario == nil {
		t.Error("ValidateJSON returned nil scenario for valid input")
	}

	// Test invalid JSON
	invalidJSON := `{
		"scenario_name": "Test",
		"invalid_json": 
	}`

	_, err = validator.ValidateJSON([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected JSON parsing error for invalid JSON")
	}

	// Test JSON that fails validation
	invalidScenarioJSON := `{
		"scenario_name": "",
		"scenario_type": "custom",
		"base_delay": "100ms"
	}`

	_, err = validator.ValidateJSON([]byte(invalidScenarioJSON))
	if err == nil {
		t.Error("Expected validation error for empty scenario_name")
	}
}

func TestErrorInjectionValidation(t *testing.T) {
	validator := NewValidator()

	// Test invalid error_rate
	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ErrorInjection: &ErrorInjectionConfig{
			Enabled:   true,
			ErrorRate: 1.5, // exceeds maximum
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "error_rate must be between") {
		t.Errorf("Expected error_rate validation error, got: %v", err)
	}

	// Test invalid error_type
	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ErrorInjection: &ErrorInjectionConfig{
			Enabled:    true,
			ErrorRate:  0.1,
			ErrorTypes: []string{"invalid_error_type"},
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "invalid error_type") {
		t.Errorf("Expected error_type validation error, got: %v", err)
	}
}

func TestPerformanceConfigValidation(t *testing.T) {
	validator := NewValidator()

	// Test invalid metrics_interval
	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		PerfMonitoring: &PerformanceConfig{
			Enabled:         true,
			MetricsInterval: 15000, // exceeds maximum
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "metrics_interval must be between") {
		t.Errorf("Expected metrics_interval validation error, got: %v", err)
	}
}

// Test the refactored ValidateScenarioFileContent function
func TestValidateScenarioFileContent(t *testing.T) {
	validator := NewValidator()

	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Test data
	validScenario := `{
		"schema_version": "1.0.0",
		"scenario_name": "Test Scenario",
		"scenario_type": "custom",
		"base_delay": "100ms",
		"description": "A test scenario"
	}`

	invalidScenario := `{
		"schema_version": "1.0.0",
		"scenario_type": "custom",
		"base_delay": "100ms"
	}`

	malformedJSON := `{
		"schema_version": "1.0.0",
		"scenario_name": "Test"
		"scenario_type": "custom"
	}`

	tests := []struct {
		name        string
		content     string
		filename    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_scenario",
			content:     validScenario,
			filename:    "valid.json",
			expectError: false,
		},
		{
			name:        "invalid_scenario_missing_name",
			content:     invalidScenario,
			filename:    "invalid.json",
			expectError: true,
			errorMsg:    "scenario_name is required",
		},
		{
			name:        "malformed_json",
			content:     malformedJSON,
			filename:    "malformed.json",
			expectError: true,
			errorMsg:    "JSON parsing failed",
		},
		{
			name:        "nonexistent_file",
			content:     "", // No file will be created
			filename:    "nonexistent.json",
			expectError: true,
			errorMsg:    "file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.filename)

			// Create file only if content is provided
			if tt.content != "" {
				err := os.WriteFile(filePath, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Test ValidateScenarioFileContent
			scenario, err := validator.ValidateScenarioFileContent(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, but got none", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				if scenario != nil {
					t.Error("Expected nil scenario on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if scenario == nil {
					t.Error("Expected scenario to be returned")
				} else {
					if scenario.ScenarioName != "Test Scenario" {
						t.Errorf("Expected scenario name 'Test Scenario', got %q", scenario.ScenarioName)
					}
					if scenario.ScenarioType != "custom" {
						t.Errorf("Expected scenario type 'custom', got %q", scenario.ScenarioType)
					}
				}
			}
		})
	}
}

func TestValidateScenarioFile_Success(t *testing.T) {
	validator := NewValidator()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "valid.json")
	content := `{
		"schema_version": "1.0.0",
		"scenario_name": "Test",
		"scenario_type": "custom",
		"base_delay": "100ms"
	}`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Success path: ValidateScenarioFile prints details and returns without calling os.Exit.
	validator.ValidateScenarioFile(filePath)
}

func TestValidateScenarioParameters_Coverage(t *testing.T) {
	validator := NewValidator()

	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ScenarioParams: &ScenarioParameters{
			DelayOverrides: map[string]string{"invalid key": "100ms"},
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "invalid delay_override key") {
		t.Errorf("Expected delay_override key error, got: %v", err)
	}

	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ScenarioParams: &ScenarioParameters{
			DelayOverrides: map[string]string{"validKey": "invalid"},
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "delay_override") {
		t.Errorf("Expected delay_override value error, got: %v", err)
	}

	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ScenarioParams: &ScenarioParameters{
			TimingPatterns: &TimingPatterns{Intervals: []int{0}},
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "timing pattern intervals") {
		t.Errorf("Expected timing interval error, got: %v", err)
	}

	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ScenarioParams: &ScenarioParameters{
			TimingPatterns: &TimingPatterns{Probabilities: []float64{1.5}},
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "timing pattern probabilities") {
		t.Errorf("Expected timing probability error, got: %v", err)
	}
}

func TestValidateMetadata_CompatibilityVersions(t *testing.T) {
	validator := NewValidator()

	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		Metadata: &ScenarioMetadata{
			Compatibility: &CompatibilityInfo{
				MinPayloadBuddyVersion: "invalid.version",
			},
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "min_payloadbuddy_version") {
		t.Errorf("Expected min_payloadbuddy_version error, got: %v", err)
	}

	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		Metadata: &ScenarioMetadata{
			Compatibility: &CompatibilityInfo{
				TestedVersions: []string{"invalid.version"},
			},
		},
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "tested_version") {
		t.Errorf("Expected tested_version error, got: %v", err)
	}
}

func TestValidatePerformanceConfig_ValidInterval(t *testing.T) {
	validator := NewValidator()

	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		PerfMonitoring: &PerformanceConfig{
			Enabled:         true,
			MetricsInterval: 1000,
		},
	}
	if err := validator.ValidateScenario(&scenario); err != nil {
		t.Errorf("Valid metrics_interval should not error, got: %v", err)
	}
}

func TestValidateScenarioFile_Comprehensive(t *testing.T) {
	validator := NewValidator()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "full.json")
	// All optional fields populated to exercise all printScenarioDetails branches.
	content := `{
		"schema_version": "1.0.0",
		"scenario_name": "Full Test",
		"scenario_type": "custom",
		"base_delay": "200ms",
		"delay_strategy": "progressive",
		"servicenow_mode": true,
		"batch_size": 50,
		"description": "Full scenario description",
		"response_limits": {
			"max_count": 5000,
			"default_count": 500
		},
		"metadata": {
			"author": "Test Author",
			"version": "1.2.3",
			"tags": ["a", "b"]
		}
	}`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	validator.ValidateScenarioFile(filePath)
}

func TestErrorInjectionValidation_NegativeRate(t *testing.T) {
	validator := NewValidator()

	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
		BaseDelay:    "100ms",
		ErrorInjection: &ErrorInjectionConfig{
			Enabled:   true,
			ErrorRate: -0.1,
		},
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !strings.Contains(err.Error(), "error_rate must be between") {
		t.Errorf("Expected error_rate error for negative value, got: %v", err)
	}
}

func TestGetScenarioDelay_CustomType(t *testing.T) {
	sm := NewManager()
	sm.scenarios["mytype"] = &Scenario{
		ScenarioType:  "mytype",
		BaseDelay:     "50ms",
		DelayStrategy: "fixed",
	}

	delay, strategy := sm.GetScenarioDelay("mytype", 0)
	if delay != 50*time.Millisecond {
		t.Errorf("Expected 50ms, got %v", delay)
	}
	if strategy != FixedDelay {
		t.Errorf("Expected FixedDelay, got %v", strategy)
	}
}
