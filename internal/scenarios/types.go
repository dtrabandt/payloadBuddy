package scenarios

// DelayStrategy defines the pattern used to space items during streaming.
type DelayStrategy int

const (
	NoDelay DelayStrategy = iota
	FixedDelay
	RandomDelay
	ProgressiveDelay
	BurstDelay
)

// Scenario represents a complete scenario configuration.
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

// ResponseLimits defines response count limits.
type ResponseLimits struct {
	MaxCount     int `json:"max_count,omitempty"`
	DefaultCount int `json:"default_count,omitempty"`
}

// ScenarioParameters contains flexible scenario-specific parameters.
type ScenarioParameters struct {
	DelayOverrides   map[string]string      `json:"delay_overrides,omitempty"`
	TimingPatterns   *TimingPatterns        `json:"timing_patterns,omitempty"`
	SimulationConfig map[string]interface{} `json:"simulation_config,omitempty"`
}

// TimingPatterns defines custom timing patterns.
type TimingPatterns struct {
	Intervals     []int                  `json:"intervals,omitempty"`
	Probabilities []float64              `json:"probabilities,omitempty"`
	Thresholds    map[string]interface{} `json:"thresholds,omitempty"`
}

// ServiceNowConfig contains ServiceNow-specific configuration.
type ServiceNowConfig struct {
	RecordTypes         []string               `json:"record_types,omitempty"`
	StateRotation       []string               `json:"state_rotation,omitempty"`
	NumberFormat        string                 `json:"number_format,omitempty"`
	SysIDFormat         string                 `json:"sys_id_format,omitempty"`
	CustomFields        map[string][]string    `json:"custom_fields,omitempty"`
	TableSpecificConfig map[string]interface{} `json:"table_specific_config,omitempty"`
}

// ErrorInjectionConfig defines error injection parameters.
type ErrorInjectionConfig struct {
	Enabled               bool     `json:"enabled,omitempty"`
	ErrorRate             float64  `json:"error_rate,omitempty"`
	ErrorTypes            []string `json:"error_types,omitempty"`
	RecoveryDelay         string   `json:"recovery_delay,omitempty"`
	ConsecutiveErrorLimit int      `json:"consecutive_error_limit,omitempty"`
}

// PerformanceConfig defines performance monitoring settings.
type PerformanceConfig struct {
	Enabled           bool `json:"enabled,omitempty"`
	MetricsInterval   int  `json:"metrics_interval,omitempty"`
	MemoryTracking    bool `json:"memory_tracking,omitempty"`
	CheckpointLogging bool `json:"checkpoint_logging,omitempty"`
}

// ScenarioMetadata contains scenario metadata.
type ScenarioMetadata struct {
	Author        string             `json:"author,omitempty"`
	CreatedDate   string             `json:"created_date,omitempty"`
	ModifiedDate  string             `json:"modified_date,omitempty"`
	Version       string             `json:"version,omitempty"`
	Project       string             `json:"project,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	Compatibility *CompatibilityInfo `json:"compatibility,omitempty"`
}

// CompatibilityInfo defines version compatibility.
type CompatibilityInfo struct {
	MinPayloadBuddyVersion string   `json:"min_payloadbuddy_version,omitempty"`
	TestedVersions         []string `json:"tested_versions,omitempty"`
}
