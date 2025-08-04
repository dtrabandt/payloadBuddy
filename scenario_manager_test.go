package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewScenarioManager(t *testing.T) {
	sm := NewScenarioManager()
	if sm == nil {
		t.Fatal("NewScenarioManager returned nil")
	}

	if sm.scenarios == nil {
		t.Error("scenarios map not initialized")
	}

	if sm.validator == nil {
		t.Error("validator not initialized")
	}
}

func TestLoadEmbeddedScenarios(t *testing.T) {
	sm := NewScenarioManager()

	// Check that embedded scenarios are loaded
	expectedScenarios := []string{"peak_hours", "maintenance", "network_issues", "database_load"}

	for _, scenarioType := range expectedScenarios {
		scenario := sm.GetScenario(scenarioType)
		if scenario == nil {
			t.Errorf("Expected embedded scenario %s not found", scenarioType)
			continue
		}

		if scenario.ScenarioType != scenarioType {
			t.Errorf("Scenario type mismatch: expected %s, got %s", scenarioType, scenario.ScenarioType)
		}

		if scenario.ScenarioName == "" {
			t.Errorf("Scenario %s has empty name", scenarioType)
		}

		if scenario.BaseDelay == "" {
			t.Errorf("Scenario %s has empty base_delay", scenarioType)
		}
	}
}

func TestGetScenario(t *testing.T) {
	sm := NewScenarioManager()

	// Test existing scenario
	scenario := sm.GetScenario("peak_hours")
	if scenario == nil {
		t.Error("Failed to get peak_hours scenario")
	}

	// Test non-existent scenario
	scenario = sm.GetScenario("non_existent")
	if scenario != nil {
		t.Error("Expected nil for non-existent scenario")
	}
}

func TestListScenarios(t *testing.T) {
	sm := NewScenarioManager()
	scenarios := sm.ListScenarios()

	if len(scenarios) == 0 {
		t.Error("No scenarios found")
	}

	// Check that we have the expected embedded scenarios
	expectedScenarios := map[string]bool{
		"peak_hours":     false,
		"maintenance":    false,
		"network_issues": false,
		"database_load":  false,
	}

	for _, scenarioType := range scenarios {
		if _, exists := expectedScenarios[scenarioType]; exists {
			expectedScenarios[scenarioType] = true
		}
	}

	for scenarioType, found := range expectedScenarios {
		if !found {
			t.Errorf("Expected scenario %s not found in list", scenarioType)
		}
	}
}

func TestGetScenarioDelay(t *testing.T) {
	sm := NewScenarioManager()

	// Test peak_hours scenario
	delay, strategy := sm.GetScenarioDelay("peak_hours", 0)
	expectedDelay := 200 * time.Millisecond
	if delay != expectedDelay {
		t.Errorf("Peak hours delay: expected %v, got %v", expectedDelay, delay)
	}
	if strategy != FixedDelay {
		t.Errorf("Peak hours strategy: expected FixedDelay, got %v", strategy)
	}

	// Test maintenance scenario with spike
	delay, strategy = sm.GetScenarioDelay("maintenance", 500)
	expectedDelay = 2 * time.Second
	if delay != expectedDelay {
		t.Errorf("Maintenance spike delay: expected %v, got %v", expectedDelay, delay)
	}

	// Test maintenance scenario without spike
	delay, strategy = sm.GetScenarioDelay("maintenance", 100)
	expectedDelay = 500 * time.Millisecond
	if delay != expectedDelay {
		t.Errorf("Maintenance normal delay: expected %v, got %v", expectedDelay, delay)
	}

	// Test database_load scenario progression
	delay1, _ := sm.GetScenarioDelay("database_load", 0)
	delay2, _ := sm.GetScenarioDelay("database_load", 100)
	if delay2 <= delay1 {
		t.Error("Database load delay should increase over time")
	}

	// Test non-existent scenario
	delay, strategy = sm.GetScenarioDelay("non_existent", 0)
	if delay != 10*time.Millisecond || strategy != FixedDelay {
		t.Errorf("Non-existent scenario should return defaults: got delay=%v, strategy=%v", delay, strategy)
	}
}

func TestGetScenarioConfig(t *testing.T) {
	sm := NewScenarioManager()

	// Test peak_hours scenario config
	batchSize, serviceNowMode, maxCount, defaultCount := sm.GetScenarioConfig("peak_hours")

	if batchSize == 0 {
		t.Error("Batch size should not be 0")
	}

	if maxCount == 0 {
		t.Error("Max count should not be 0")
	}

	if defaultCount == 0 {
		t.Error("Default count should not be 0")
	}

	// ServiceNow scenarios should enable ServiceNow mode
	if !serviceNowMode {
		t.Error("Peak hours scenario should have ServiceNow mode enabled")
	}

	// Test non-existent scenario
	batchSize, serviceNowMode, maxCount, defaultCount = sm.GetScenarioConfig("non_existent")
	if batchSize != 100 || serviceNowMode != false || maxCount != 1000000 || defaultCount != 10000 {
		t.Error("Non-existent scenario should return default values")
	}
}

func TestParseDelay(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"100ms", 100 * time.Millisecond, false},
		{"1s", 1 * time.Second, false},
		{"500", 500 * time.Millisecond, false},
		{"1.5s", 1500 * time.Millisecond, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		result, err := ParseDelay(tc.input)

		if tc.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("Input %s: expected %v, got %v", tc.input, tc.expected, result)
			}
		}
	}
}

func TestParseDelayStrategy(t *testing.T) {
	testCases := []struct {
		input    string
		expected DelayStrategy
	}{
		{"fixed", FixedDelay},
		{"FIXED", FixedDelay},
		{"random", RandomDelay},
		{"progressive", ProgressiveDelay},
		{"burst", BurstDelay},
		{"unknown", FixedDelay}, // default
		{"", FixedDelay},        // default
	}

	for _, tc := range testCases {
		result := ParseDelayStrategy(tc.input)
		if result != tc.expected {
			t.Errorf("Input %s: expected %v, got %v", tc.input, tc.expected, result)
		}
	}
}

func TestUserScenarioOverride(t *testing.T) {
	// Create a temporary directory for test scenarios
	tempDir := t.TempDir()

	// Create a custom scenario that overrides peak_hours
	customScenario := Scenario{
		SchemaVersion:  "1.0.0",
		ScenarioName:   "Custom Peak Hours",
		ScenarioType:   "peak_hours",
		BaseDelay:      "300ms",
		DelayStrategy:  "fixed",
		ServiceNowMode: false,
		BatchSize:      50,
	}

	scenarioJSON, err := json.Marshal(customScenario)
	if err != nil {
		t.Fatalf("Failed to marshal test scenario: %v", err)
	}

	scenarioFile := filepath.Join(tempDir, "custom_peak_hours.json")
	err = os.WriteFile(scenarioFile, scenarioJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test scenario file: %v", err)
	}

	// Create a scenario manager with custom user path
	sm := &ScenarioManager{
		scenarios: make(map[string]*Scenario),
		userPath:  tempDir,
		validator: NewScenarioValidator(),
	}

	// Load embedded scenarios first
	sm.loadEmbeddedScenarios()
	originalScenario := sm.GetScenario("peak_hours")
	if originalScenario == nil {
		t.Fatal("Failed to load embedded peak_hours scenario")
	}

	// Load user scenarios (should override)
	sm.loadUserScenarios()
	overriddenScenario := sm.GetScenario("peak_hours")
	if overriddenScenario == nil {
		t.Fatal("Failed to load overridden peak_hours scenario")
	}

	// Verify the scenario was overridden
	if overriddenScenario.ScenarioName != "Custom Peak Hours" {
		t.Errorf("Expected overridden scenario name 'Custom Peak Hours', got '%s'", overriddenScenario.ScenarioName)
	}

	if overriddenScenario.BaseDelay != "300ms" {
		t.Errorf("Expected overridden base delay '300ms', got '%s'", overriddenScenario.BaseDelay)
	}
}
