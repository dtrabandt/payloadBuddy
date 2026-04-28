package scenarios

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Validator validates scenario JSON against the defined schema rules.
type Validator struct {
	schemaVersion string
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{schemaVersion: "1.0.0"}
}

// ValidateJSON parses and validates raw JSON, returning the Scenario or an error.
func (sv *Validator) ValidateJSON(jsonData []byte) (*Scenario, error) {
	var scenario Scenario
	if err := json.Unmarshal(jsonData, &scenario); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v", err)
	}
	if err := sv.ValidateScenario(&scenario); err != nil {
		return nil, err
	}
	return &scenario, nil
}

// ValidateScenario validates a Scenario struct against all schema rules.
func (sv *Validator) ValidateScenario(scenario *Scenario) error {
	if scenario.ScenarioName == "" {
		return fmt.Errorf("scenario_name is required")
	}
	if len(scenario.ScenarioName) > 100 {
		return fmt.Errorf("scenario_name must be 100 characters or less")
	}
	if scenario.ScenarioType == "" {
		return fmt.Errorf("scenario_type is required")
	}
	validTypes := []string{"peak_hours", "maintenance", "network_issues", "database_load", "custom"}
	if !sv.isValidEnum(scenario.ScenarioType, validTypes) {
		return fmt.Errorf("scenario_type must be one of: %s", strings.Join(validTypes, ", "))
	}
	if scenario.BaseDelay == "" {
		return fmt.Errorf("base_delay is required")
	}
	if err := sv.validateDelayFormat(scenario.BaseDelay); err != nil {
		return fmt.Errorf("base_delay validation failed: %v", err)
	}
	if scenario.Description != "" && len(scenario.Description) > 500 {
		return fmt.Errorf("description must be 500 characters or less")
	}
	if scenario.DelayStrategy != "" {
		validStrategies := []string{"fixed", "random", "progressive", "burst"}
		if !sv.isValidEnum(scenario.DelayStrategy, validStrategies) {
			return fmt.Errorf("delay_strategy must be one of: %s", strings.Join(validStrategies, ", "))
		}
	}
	if scenario.SchemaVersion != "" {
		if err := sv.validateVersionFormat(scenario.SchemaVersion); err != nil {
			return fmt.Errorf("schema_version validation failed: %v", err)
		}
	}
	if scenario.ResponseLimits != nil {
		if err := sv.validateResponseLimits(scenario.ResponseLimits); err != nil {
			return fmt.Errorf("response_limits validation failed: %v", err)
		}
	}
	if scenario.ServiceNowConfig != nil {
		if err := sv.validateServiceNowConfig(scenario.ServiceNowConfig); err != nil {
			return fmt.Errorf("servicenow_config validation failed: %v", err)
		}
	}
	if scenario.ErrorInjection != nil {
		if err := sv.validateErrorInjection(scenario.ErrorInjection); err != nil {
			return fmt.Errorf("error_injection validation failed: %v", err)
		}
	}
	if scenario.PerfMonitoring != nil {
		if err := sv.validatePerformanceConfig(scenario.PerfMonitoring); err != nil {
			return fmt.Errorf("performance_monitoring validation failed: %v", err)
		}
	}
	if scenario.Metadata != nil {
		if err := sv.validateMetadata(scenario.Metadata); err != nil {
			return fmt.Errorf("metadata validation failed: %v", err)
		}
	}
	if scenario.ScenarioParams != nil {
		if err := sv.validateScenarioParameters(scenario.ScenarioParams); err != nil {
			return fmt.Errorf("scenario_parameters validation failed: %v", err)
		}
	}
	return nil
}

// ValidateScenarioFile validates a file and prints results; calls os.Exit on failure.
func (sv *Validator) ValidateScenarioFile(filePath string) {
	fmt.Printf("Validating scenario file: %s\n", filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("❌ Error: File does not exist: %s\n", filePath)
		os.Exit(1)
	}
	scenario, err := sv.ValidateScenarioFileContent(filePath)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	sv.printScenarioDetails(scenario)
}

// ValidateScenarioFileContent reads and validates a file, returning the Scenario or an error.
func (sv *Validator) ValidateScenarioFileContent(filePath string) (*Scenario, error) {
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: directory traversal not allowed")
	}
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", cleanPath)
	}
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	scenario, err := sv.ValidateJSON(content)
	if err != nil {
		return nil, fmt.Errorf("validation failed:\n%v", err)
	}
	return scenario, nil
}

func (sv *Validator) validateDelayFormat(delay string) error {
	durationPattern := regexp.MustCompile(`^(\d+(\.\d+)?(ns|us|μs|ms|s|m|h))|\d+$`)
	if !durationPattern.MatchString(delay) {
		return fmt.Errorf("invalid delay format: %s", delay)
	}
	_, err := ParseDelay(delay)
	return err
}

func (sv *Validator) validateVersionFormat(version string) error {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version) {
		return fmt.Errorf("invalid version format: %s (expected: x.y.z)", version)
	}
	return nil
}

func (sv *Validator) validateResponseLimits(limits *ResponseLimits) error {
	if limits.MaxCount != 0 && (limits.MaxCount < 1 || limits.MaxCount > 1000000) {
		return fmt.Errorf("max_count must be between 1 and 1000000")
	}
	if limits.DefaultCount < 0 || limits.DefaultCount > 1000000 {
		return fmt.Errorf("default_count must be between 0 and 1000000")
	}
	return nil
}

func (sv *Validator) validateServiceNowConfig(config *ServiceNowConfig) error {
	validRecordTypes := []string{"incident", "problem", "change_request", "catalog_task", "kb_knowledge", "sys_user"}
	for _, rt := range config.RecordTypes {
		if !sv.isValidEnum(rt, validRecordTypes) {
			return fmt.Errorf("invalid record_type: %s", rt)
		}
	}
	if config.SysIDFormat != "" {
		validFormats := []string{"standard", "uuid", "sequential"}
		if !sv.isValidEnum(config.SysIDFormat, validFormats) {
			return fmt.Errorf("sys_id_format must be one of: %s", strings.Join(validFormats, ", "))
		}
	}
	return nil
}

func (sv *Validator) validateErrorInjection(config *ErrorInjectionConfig) error {
	if config.ErrorRate < 0.0 || config.ErrorRate > 1.0 {
		return fmt.Errorf("error_rate must be between 0.0 and 1.0")
	}
	validErrorTypes := []string{"timeout", "authentication_failure", "server_error", "bad_request", "rate_limit", "connection_reset"}
	for _, et := range config.ErrorTypes {
		if !sv.isValidEnum(et, validErrorTypes) {
			return fmt.Errorf("invalid error_type: %s", et)
		}
	}
	if config.RecoveryDelay != "" {
		if err := sv.validateDelayFormat(config.RecoveryDelay); err != nil {
			return fmt.Errorf("recovery_delay validation failed: %v", err)
		}
	}
	if config.ConsecutiveErrorLimit < 1 || config.ConsecutiveErrorLimit > 10 {
		return fmt.Errorf("consecutive_error_limit must be between 1 and 10")
	}
	return nil
}

func (sv *Validator) validatePerformanceConfig(config *PerformanceConfig) error {
	if config.MetricsInterval < 1 || config.MetricsInterval > 10000 {
		return fmt.Errorf("metrics_interval must be between 1 and 10000")
	}
	return nil
}

func (sv *Validator) validateMetadata(metadata *ScenarioMetadata) error {
	if metadata.Version != "" {
		if err := sv.validateVersionFormat(metadata.Version); err != nil {
			return fmt.Errorf("version validation failed: %v", err)
		}
	}
	if metadata.CreatedDate != "" {
		if _, err := time.Parse("2006-01-02", metadata.CreatedDate); err != nil {
			return fmt.Errorf("created_date must be YYYY-MM-DD: %s", metadata.CreatedDate)
		}
	}
	if metadata.ModifiedDate != "" {
		if _, err := time.Parse("2006-01-02", metadata.ModifiedDate); err != nil {
			return fmt.Errorf("modified_date must be YYYY-MM-DD: %s", metadata.ModifiedDate)
		}
	}
	if metadata.Compatibility != nil {
		if metadata.Compatibility.MinPayloadBuddyVersion != "" {
			if err := sv.validateVersionFormat(metadata.Compatibility.MinPayloadBuddyVersion); err != nil {
				return fmt.Errorf("min_payloadbuddy_version validation failed: %v", err)
			}
		}
		for _, version := range metadata.Compatibility.TestedVersions {
			if err := sv.validateVersionFormat(version); err != nil {
				return fmt.Errorf("tested_versions contains invalid version: %v", err)
			}
		}
	}
	return nil
}

func (sv *Validator) validateScenarioParameters(params *ScenarioParameters) error {
	for key, value := range params.DelayOverrides {
		if !sv.isValidIdentifier(key) {
			return fmt.Errorf("invalid delay_override key: %s", key)
		}
		if err := sv.validateDelayFormat(value); err != nil {
			return fmt.Errorf("delay_override %s validation failed: %v", key, err)
		}
	}
	if params.TimingPatterns != nil {
		for _, interval := range params.TimingPatterns.Intervals {
			if interval < 1 {
				return fmt.Errorf("timing pattern intervals must be >= 1")
			}
		}
		for _, prob := range params.TimingPatterns.Probabilities {
			if prob < 0.0 || prob > 1.0 {
				return fmt.Errorf("timing pattern probabilities must be between 0.0 and 1.0")
			}
		}
	}
	return nil
}

func (sv *Validator) isValidEnum(value string, validValues []string) bool {
	for _, v := range validValues {
		if value == v {
			return true
		}
	}
	return false
}

func (sv *Validator) isValidIdentifier(identifier string) bool {
	return regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(identifier)
}

func (sv *Validator) printScenarioDetails(scenario *Scenario) {
	fmt.Printf("✅ Validation successful!\n\n")
	fmt.Printf("📋 Scenario Details:\n")
	fmt.Printf("   Name: %s\n", scenario.ScenarioName)
	fmt.Printf("   Type: %s\n", scenario.ScenarioType)
	fmt.Printf("   Base Delay: %s\n", scenario.BaseDelay)
	if scenario.DelayStrategy != "" {
		fmt.Printf("   Delay Strategy: %s\n", scenario.DelayStrategy)
	}
	if scenario.ServiceNowMode {
		fmt.Printf("   ServiceNow Mode: enabled\n")
	}
	if scenario.BatchSize > 0 {
		fmt.Printf("   Batch Size: %d\n", scenario.BatchSize)
	}
	if scenario.ResponseLimits != nil {
		if scenario.ResponseLimits.MaxCount > 0 {
			fmt.Printf("   Max Count: %d\n", scenario.ResponseLimits.MaxCount)
		}
		if scenario.ResponseLimits.DefaultCount > 0 {
			fmt.Printf("   Default Count: %d\n", scenario.ResponseLimits.DefaultCount)
		}
	}
	if scenario.Description != "" {
		fmt.Printf("   Description: %s\n", scenario.Description)
	}
	if scenario.Metadata != nil {
		if scenario.Metadata.Author != "" {
			fmt.Printf("   Author: %s\n", scenario.Metadata.Author)
		}
		if scenario.Metadata.Version != "" {
			fmt.Printf("   Version: %s\n", scenario.Metadata.Version)
		}
		if len(scenario.Metadata.Tags) > 0 {
			fmt.Printf("   Tags: %v\n", scenario.Metadata.Tags)
		}
	}
	fmt.Printf("\n🎯 Usage: Use this scenario with ?scenario=%s\n", scenario.ScenarioType)
	fmt.Printf("💡 Tip: Place this file in $HOME/.config/payloadBuddy/scenarios/ to make it available\n")
}
