package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Embedded scenarios - included at compile time
//
//go:embed scenarios/*.json
var embeddedScenarios embed.FS

// Scenario represents a complete scenario configuration
type Scenario struct {
	SchemaVersion    string                `json:"schema_version"`
	ScenarioName     string                `json:"scenario_name"`
	Description      string                `json:"description,omitempty"`
	ScenarioType     string                `json:"scenario_type"`
	BaseDelay        string                `json:"base_delay"`
	DelayStrategy    string                `json:"delay_strategy,omitempty"`
	ServiceNowMode   bool                  `json:"servicenow_mode,omitempty"`
	BatchSize        int                   `json:"batch_size,omitempty"`
	ResponseLimits   *ResponseLimits       `json:"response_limits,omitempty"`
	ScenarioParams   *ScenarioParameters   `json:"scenario_parameters,omitempty"`
	ServiceNowConfig *ServiceNowConfig     `json:"servicenow_config,omitempty"`
	ErrorInjection   *ErrorInjectionConfig `json:"error_injection,omitempty"`
	PerfMonitoring   *PerformanceConfig    `json:"performance_monitoring,omitempty"`
	Metadata         *ScenarioMetadata     `json:"metadata,omitempty"`
}

// ResponseLimits defines response count limits
type ResponseLimits struct {
	MaxCount     int `json:"max_count,omitempty"`
	DefaultCount int `json:"default_count,omitempty"`
}

// ScenarioParameters contains flexible scenario-specific parameters
type ScenarioParameters struct {
	DelayOverrides   map[string]string      `json:"delay_overrides,omitempty"`
	TimingPatterns   *TimingPatterns        `json:"timing_patterns,omitempty"`
	SimulationConfig map[string]interface{} `json:"simulation_config,omitempty"`
}

// TimingPatterns defines custom timing patterns
type TimingPatterns struct {
	Intervals     []int                  `json:"intervals,omitempty"`
	Probabilities []float64              `json:"probabilities,omitempty"`
	Thresholds    map[string]interface{} `json:"thresholds,omitempty"`
}

// ServiceNowConfig contains ServiceNow-specific configuration
type ServiceNowConfig struct {
	RecordTypes         []string               `json:"record_types,omitempty"`
	StateRotation       []string               `json:"state_rotation,omitempty"`
	NumberFormat        string                 `json:"number_format,omitempty"`
	SysIDFormat         string                 `json:"sys_id_format,omitempty"`
	CustomFields        map[string][]string    `json:"custom_fields,omitempty"`
	TableSpecificConfig map[string]interface{} `json:"table_specific_config,omitempty"`
}

// ErrorInjectionConfig defines error injection parameters
type ErrorInjectionConfig struct {
	Enabled               bool     `json:"enabled,omitempty"`
	ErrorRate             float64  `json:"error_rate,omitempty"`
	ErrorTypes            []string `json:"error_types,omitempty"`
	RecoveryDelay         string   `json:"recovery_delay,omitempty"`
	ConsecutiveErrorLimit int      `json:"consecutive_error_limit,omitempty"`
}

// PerformanceConfig defines performance monitoring settings
type PerformanceConfig struct {
	Enabled           bool `json:"enabled,omitempty"`
	MetricsInterval   int  `json:"metrics_interval,omitempty"`
	MemoryTracking    bool `json:"memory_tracking,omitempty"`
	CheckpointLogging bool `json:"checkpoint_logging,omitempty"`
}

// ScenarioMetadata contains scenario metadata
type ScenarioMetadata struct {
	Author        string             `json:"author,omitempty"`
	CreatedDate   string             `json:"created_date,omitempty"`
	ModifiedDate  string             `json:"modified_date,omitempty"`
	Version       string             `json:"version,omitempty"`
	Project       string             `json:"project,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	Compatibility *CompatibilityInfo `json:"compatibility,omitempty"`
}

// CompatibilityInfo defines version compatibility
type CompatibilityInfo struct {
	MinPayloadBuddyVersion string   `json:"min_payloadbuddy_version,omitempty"`
	TestedVersions         []string `json:"tested_versions,omitempty"`
}

// ScenarioManager manages loading and accessing scenarios
type ScenarioManager struct {
	scenarios map[string]*Scenario
	userPath  string
	validator *ScenarioValidator
}

// NewScenarioManager creates a new scenario manager
func NewScenarioManager() *ScenarioManager {
	sm := &ScenarioManager{
		scenarios: make(map[string]*Scenario),
		userPath:  getScenarioPath(),
		validator: NewScenarioValidator(),
	}

	// Load scenarios in order: embedded first, then user scenarios
	sm.loadEmbeddedScenarios()
	sm.loadUserScenarios()

	return sm
}

// getScenarioPath returns the user scenario directory path
func getScenarioPath() string {
	var basePath string
	switch runtime.GOOS {
	case "windows":
		basePath = os.Getenv("USERPROFILE")
	default:
		basePath = os.Getenv("HOME")
	}

	scenarioPath := filepath.Join(basePath, ".config", "payloadBuddy", "scenarios")

	// Create directory if it doesn't exist
	if _, err := os.Stat(scenarioPath); os.IsNotExist(err) {
		if err := os.MkdirAll(scenarioPath, os.ModePerm); err != nil {
			log.Printf("Warning: Failed to create scenario directory %s: %v", scenarioPath, err)
		} else {
			log.Printf("Created scenario directory: %s", scenarioPath)
		}
	}

	return scenarioPath
}

// loadEmbeddedScenarios loads scenarios embedded in the binary
func (sm *ScenarioManager) loadEmbeddedScenarios() {
	entries, err := embeddedScenarios.ReadDir("scenarios")
	if err != nil {
		log.Printf("Warning: Failed to read embedded scenarios: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && entry.Name() != "scenario_schema_v1.0.0.json" {
			content, err := embeddedScenarios.ReadFile(filepath.Join("scenarios", entry.Name()))
			if err != nil {
				log.Printf("Warning: Failed to read embedded scenario %s: %v", entry.Name(), err)
				continue
			}

			// Validate and parse scenario
			scenario, err := sm.validator.ValidateJSON(content)
			if err != nil {
				log.Printf("Warning: Validation failed for embedded scenario %s: %v", entry.Name(), err)
				continue
			}

			// Validate compatibility
			if !sm.isCompatible(scenario) {
				log.Printf("Warning: Embedded scenario %s is not compatible with current version", scenario.ScenarioName)
				continue
			}

			sm.scenarios[scenario.ScenarioType] = scenario
			log.Printf("Loaded embedded scenario: %s (%s)", scenario.ScenarioName, scenario.ScenarioType)
		}
	}
}

// loadUserScenarios loads user-defined scenarios from the config directory
func (sm *ScenarioManager) loadUserScenarios() {
	if _, err := os.Stat(sm.userPath); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to load
		return
	}

	err := filepath.WalkDir(sm.userPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Warning: Failed to read user scenario %s: %v", path, err)
				return nil // Continue with next file
			}

			// Validate and parse scenario
			scenario, err := sm.validator.ValidateJSON(content)
			if err != nil {
				log.Printf("Warning: Validation failed for user scenario %s: %v", path, err)
				return nil // Continue with next file
			}

			// Validate compatibility
			if !sm.isCompatible(scenario) {
				log.Printf("Warning: User scenario %s is not compatible with current version", scenario.ScenarioName)
				return nil
			}

			// User scenarios override embedded ones with same scenario_type
			if existing, exists := sm.scenarios[scenario.ScenarioType]; exists {
				log.Printf("User scenario %s (%s) overriding embedded scenario %s",
					scenario.ScenarioName, scenario.ScenarioType, existing.ScenarioName)
			}

			sm.scenarios[scenario.ScenarioType] = scenario
			log.Printf("Loaded user scenario: %s (%s)", scenario.ScenarioName, scenario.ScenarioType)
		}

		return nil
	})

	if err != nil {
		log.Printf("Warning: Error scanning user scenarios: %v", err)
	}
}

// isCompatible checks if a scenario is compatible with the current version
func (sm *ScenarioManager) isCompatible(scenario *Scenario) bool {
	if scenario.Metadata == nil || scenario.Metadata.Compatibility == nil {
		// No compatibility info, assume compatible
		return true
	}

	minVersion := scenario.Metadata.Compatibility.MinPayloadBuddyVersion
	if minVersion == "" {
		return true
	}

	// For now, we'll implement a simple version check
	// In a real implementation, you'd parse and compare semantic versions
	// This is a placeholder that accepts all versions
	return true
}

// GetScenario retrieves a scenario by type
func (sm *ScenarioManager) GetScenario(scenarioType string) *Scenario {
	return sm.scenarios[scenarioType]
}

// ListScenarios returns all available scenario types
func (sm *ScenarioManager) ListScenarios() []string {
	var types []string
	for scenarioType := range sm.scenarios {
		types = append(types, scenarioType)
	}
	return types
}

// ParseDelay converts a delay string to time.Duration
func ParseDelay(delayStr string) (time.Duration, error) {
	// Try parsing as duration first (e.g., "100ms", "1s")
	if duration, err := time.ParseDuration(delayStr); err == nil {
		return duration, nil
	}

	// Fallback: parse as milliseconds
	if ms, err := strconv.Atoi(delayStr); err == nil {
		return time.Duration(ms) * time.Millisecond, nil
	}

	return 0, fmt.Errorf("invalid delay format: %s", delayStr)
}

// ParseDelayStrategy converts a strategy string to DelayStrategy
func ParseDelayStrategy(strategy string) DelayStrategy {
	switch strings.ToLower(strategy) {
	case "fixed":
		return FixedDelay
	case "random":
		return RandomDelay
	case "progressive":
		return ProgressiveDelay
	case "burst":
		return BurstDelay
	default:
		return FixedDelay
	}
}

// GetScenarioDelay calculates delay for a scenario at a specific item index
func (sm *ScenarioManager) GetScenarioDelay(scenarioType string, itemIndex int) (time.Duration, DelayStrategy) {
	scenario := sm.GetScenario(scenarioType)
	if scenario == nil {
		// Return default values if scenario not found
		return 10 * time.Millisecond, FixedDelay
	}

	baseDelay, err := ParseDelay(scenario.BaseDelay)
	if err != nil {
		baseDelay = 10 * time.Millisecond
	}

	strategy := ParseDelayStrategy(scenario.DelayStrategy)

	// Apply scenario-specific delay modifications
	switch scenario.ScenarioType {
	case "peak_hours":
		return 200 * time.Millisecond, FixedDelay
	case "maintenance":
		if itemIndex%500 == 0 {
			return 2 * time.Second, FixedDelay // Maintenance spike
		}
		return 500 * time.Millisecond, FixedDelay
	case "network_issues":
		// This will be handled by the caller using random logic
		return baseDelay, RandomDelay
	case "database_load":
		// Progressive degradation: baseDelay + (itemIndex/100 * 10ms)
		degradation := time.Duration(itemIndex/100) * 10 * time.Millisecond
		return baseDelay + degradation, FixedDelay
	default:
		return baseDelay, strategy
	}
}

// GetScenarioConfig returns configuration values for a scenario
func (sm *ScenarioManager) GetScenarioConfig(scenarioType string) (batchSize int, serviceNowMode bool, maxCount int, defaultCount int) {
	scenario := sm.GetScenario(scenarioType)
	if scenario == nil {
		return 100, false, 1000000, 10000 // Default values
	}

	batchSize = 100
	if scenario.BatchSize > 0 {
		batchSize = scenario.BatchSize
	}

	serviceNowMode = scenario.ServiceNowMode

	maxCount = 1000000
	defaultCount = 10000
	if scenario.ResponseLimits != nil {
		if scenario.ResponseLimits.MaxCount > 0 {
			maxCount = scenario.ResponseLimits.MaxCount
		}
		if scenario.ResponseLimits.DefaultCount > 0 {
			defaultCount = scenario.ResponseLimits.DefaultCount
		}
	}

	return
}
