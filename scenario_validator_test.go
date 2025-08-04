package main

import (
	"testing"
)

func TestScenarioValidator(t *testing.T) {
	validator := NewScenarioValidator()

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
	validator := NewScenarioValidator()

	// Test missing scenario_name
	scenario := Scenario{
		ScenarioType: "custom",
		BaseDelay:    "100ms",
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !contains(err.Error(), "scenario_name is required") {
		t.Errorf("Expected scenario_name required error, got: %v", err)
	}

	// Test missing scenario_type
	scenario = Scenario{
		ScenarioName: "Test",
		BaseDelay:    "100ms",
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !contains(err.Error(), "scenario_type is required") {
		t.Errorf("Expected scenario_type required error, got: %v", err)
	}

	// Test missing base_delay
	scenario = Scenario{
		ScenarioName: "Test",
		ScenarioType: "custom",
	}
	err = validator.ValidateScenario(&scenario)
	if err == nil || !contains(err.Error(), "base_delay is required") {
		t.Errorf("Expected base_delay required error, got: %v", err)
	}
}

func TestScenarioValidatorEnums(t *testing.T) {
	validator := NewScenarioValidator()

	// Test invalid scenario_type
	scenario := Scenario{
		ScenarioName: "Test",
		ScenarioType: "invalid_type",
		BaseDelay:    "100ms",
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !contains(err.Error(), "scenario_type must be one of") {
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
	if err == nil || !contains(err.Error(), "delay_strategy must be one of") {
		t.Errorf("Expected delay_strategy enum error, got: %v", err)
	}
}

func TestScenarioValidatorDelayFormat(t *testing.T) {
	validator := NewScenarioValidator()

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
	validator := NewScenarioValidator()

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
	if err == nil || !contains(err.Error(), "max_count must be between") {
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
	if err == nil || !contains(err.Error(), "default_count must be between 0 and 1000000") {
		t.Errorf("Expected default_count validation error, got: %v", err)
	}
}

func TestScenarioValidatorServiceNowConfig(t *testing.T) {
	validator := NewScenarioValidator()

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
	if err == nil || !contains(err.Error(), "invalid record_type") {
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
	if err == nil || !contains(err.Error(), "sys_id_format must be one of") {
		t.Errorf("Expected sys_id_format validation error, got: %v", err)
	}
}

func TestScenarioValidatorVersionFormat(t *testing.T) {
	validator := NewScenarioValidator()

	// Test invalid schema_version
	scenario := Scenario{
		ScenarioName:  "Test",
		ScenarioType:  "custom",
		BaseDelay:     "100ms",
		SchemaVersion: "invalid.version",
	}
	err := validator.ValidateScenario(&scenario)
	if err == nil || !contains(err.Error(), "schema_version validation failed") {
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
	if err == nil || !contains(err.Error(), "version validation failed") {
		t.Errorf("Expected metadata version validation error, got: %v", err)
	}
}

func TestScenarioValidatorDateFormat(t *testing.T) {
	validator := NewScenarioValidator()

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
	if err == nil || !contains(err.Error(), "created_date validation failed") {
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
	validator := NewScenarioValidator()

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
	validator := NewScenarioValidator()

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
	if err == nil || !contains(err.Error(), "error_rate must be between") {
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
	if err == nil || !contains(err.Error(), "invalid error_type") {
		t.Errorf("Expected error_type validation error, got: %v", err)
	}
}

func TestPerformanceConfigValidation(t *testing.T) {
	validator := NewScenarioValidator()

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
	if err == nil || !contains(err.Error(), "metrics_interval must be between") {
		t.Errorf("Expected metrics_interval validation error, got: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
