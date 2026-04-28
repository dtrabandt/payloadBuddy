package scenarios

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

//go:embed embedded/*.json
var embeddedScenarios embed.FS

// Manager loads and provides access to scenario configurations.
type Manager struct {
	scenarios map[string]*Scenario
	userPath  string
	validator *Validator
}

// NewManager creates a Manager, loading embedded scenarios first and then
// any user-defined scenarios from $HOME/.config/payloadBuddy/scenarios/.
func NewManager() *Manager {
	sm := &Manager{
		scenarios: make(map[string]*Scenario),
		userPath:  scenarioPath(),
		validator: NewValidator(),
	}
	sm.loadEmbedded()
	sm.loadUser()
	return sm
}

// GetScenario retrieves a scenario by type, returning nil when not found.
func (sm *Manager) GetScenario(scenarioType string) *Scenario {
	return sm.scenarios[scenarioType]
}

// ListScenarios returns all available scenario types in sorted order.
func (sm *Manager) ListScenarios() []string {
	types := make([]string, 0, len(sm.scenarios))
	for t := range sm.scenarios {
		types = append(types, t)
	}
	slices.Sort(types)
	return types
}

// GetScenarioDelay calculates the delay and strategy for a scenario at itemIndex.
func (sm *Manager) GetScenarioDelay(scenarioType string, itemIndex int) (time.Duration, DelayStrategy) {
	scenario := sm.GetScenario(scenarioType)
	if scenario == nil {
		return 10 * time.Millisecond, FixedDelay
	}

	baseDelay, err := ParseDelay(scenario.BaseDelay)
	if err != nil {
		baseDelay = 10 * time.Millisecond
	}
	strategy := ParseDelayStrategy(scenario.DelayStrategy)

	switch scenario.ScenarioType {
	case "peak_hours":
		return 200 * time.Millisecond, FixedDelay
	case "maintenance":
		if itemIndex%500 == 0 {
			return 2 * time.Second, FixedDelay
		}
		return 500 * time.Millisecond, FixedDelay
	case "network_issues":
		return baseDelay, RandomDelay
	case "database_load":
		degradation := time.Duration(itemIndex/100) * 10 * time.Millisecond
		return baseDelay + degradation, FixedDelay
	default:
		return baseDelay, strategy
	}
}

// GetScenarioConfig returns runtime configuration values for a scenario.
func (sm *Manager) GetScenarioConfig(scenarioType string) (batchSize int, serviceNowMode bool, maxCount int, defaultCount int) {
	scenario := sm.GetScenario(scenarioType)
	if scenario == nil {
		return 100, false, 1000000, 10000
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

// ParseDelay converts a delay string (e.g. "100ms" or "100") to time.Duration.
func ParseDelay(delayStr string) (time.Duration, error) {
	if duration, err := time.ParseDuration(delayStr); err == nil {
		return duration, nil
	}
	if ms, err := strconv.Atoi(delayStr); err == nil {
		return time.Duration(ms) * time.Millisecond, nil
	}
	return 0, fmt.Errorf("invalid delay format: %s", delayStr)
}

// ParseDelayStrategy converts a strategy name to a DelayStrategy constant.
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

// scenarioPath returns the OS-appropriate user scenario directory, creating it if absent.
func scenarioPath() string {
	basePath := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		basePath = os.Getenv("USERPROFILE")
	}
	p := filepath.Join(basePath, ".config", "payloadBuddy", "scenarios")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.MkdirAll(p, 0750); err != nil {
			slog.Warn("failed to create scenario directory", "path", p, "err", err)
		} else {
			slog.Info("created scenario directory", "path", p)
		}
	}
	return p
}

func (sm *Manager) loadEmbedded() {
	entries, err := embeddedScenarios.ReadDir("embedded")
	if err != nil {
		slog.Warn("failed to read embedded scenarios", "err", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || entry.Name() == "scenario_schema_v1.0.0.json" {
			continue
		}
		content, err := embeddedScenarios.ReadFile(filepath.Join("embedded", entry.Name()))
		if err != nil {
			slog.Warn("failed to read embedded scenario", "file", entry.Name(), "err", err)
			continue
		}
		scenario, err := sm.validator.ValidateJSON(content)
		if err != nil {
			slog.Warn("validation failed for embedded scenario", "file", entry.Name(), "err", err)
			continue
		}
		if !isCompatible(scenario) {
			slog.Warn("embedded scenario not compatible with current version", "name", scenario.ScenarioName)
			continue
		}
		sm.scenarios[scenario.ScenarioType] = scenario
		slog.Info("loaded embedded scenario", "name", scenario.ScenarioName, "type", scenario.ScenarioType)
	}
}

func (sm *Manager) loadUser() {
	if _, err := os.Stat(sm.userPath); os.IsNotExist(err) {
		return
	}
	err := filepath.WalkDir(sm.userPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		cleanPath := filepath.Clean(path)
		userPathAbs, _ := filepath.Abs(sm.userPath)
		pathAbs, _ := filepath.Abs(cleanPath)
		if !strings.HasPrefix(pathAbs, userPathAbs) {
			slog.Warn("skipping file outside user directory", "path", path)
			return nil
		}
		content, err := os.ReadFile(cleanPath)
		if err != nil {
			slog.Warn("failed to read user scenario", "path", cleanPath, "err", err)
			return nil
		}
		scenario, err := sm.validator.ValidateJSON(content)
		if err != nil {
			slog.Warn("validation failed for user scenario", "path", path, "err", err)
			return nil
		}
		if !isCompatible(scenario) {
			slog.Warn("user scenario not compatible with current version", "name", scenario.ScenarioName)
			return nil
		}
		if existing, exists := sm.scenarios[scenario.ScenarioType]; exists {
			slog.Info("user scenario overriding embedded scenario",
				"name", scenario.ScenarioName, "type", scenario.ScenarioType, "replaced", existing.ScenarioName)
		}
		sm.scenarios[scenario.ScenarioType] = scenario
		slog.Info("loaded user scenario", "name", scenario.ScenarioName, "type", scenario.ScenarioType)
		return nil
	})
	if err != nil {
		slog.Warn("error scanning user scenarios", "err", err)
	}
}

func isCompatible(scenario *Scenario) bool {
	if scenario.Metadata == nil || scenario.Metadata.Compatibility == nil {
		return true
	}
	// If no minimum version is specified, the scenario is compatible with all versions.
	return scenario.Metadata.Compatibility.MinPayloadBuddyVersion == ""
}
