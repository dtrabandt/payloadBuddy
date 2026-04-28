package scenarios

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	sm := NewManager()
	if sm == nil {
		t.Fatal("NewManager returned nil")
	}
	if sm.scenarios == nil {
		t.Error("scenarios map not initialized")
	}
	if sm.validator == nil {
		t.Error("validator not initialized")
	}
}

func TestLoadEmbeddedScenarios(t *testing.T) {
	sm := NewManager()
	for _, st := range []string{"peak_hours", "maintenance", "network_issues", "database_load"} {
		s := sm.GetScenario(st)
		if s == nil {
			t.Errorf("embedded scenario %s not found", st)
			continue
		}
		if s.ScenarioType != st {
			t.Errorf("scenario type mismatch: expected %s, got %s", st, s.ScenarioType)
		}
		if s.ScenarioName == "" {
			t.Errorf("scenario %s has empty name", st)
		}
		if s.BaseDelay == "" {
			t.Errorf("scenario %s has empty base_delay", st)
		}
	}
}

func TestGetScenario(t *testing.T) {
	sm := NewManager()
	if sm.GetScenario("peak_hours") == nil {
		t.Error("failed to get peak_hours scenario")
	}
	if sm.GetScenario("non_existent") != nil {
		t.Error("expected nil for non-existent scenario")
	}
}

func TestListScenarios(t *testing.T) {
	sm := NewManager()
	list := sm.ListScenarios()
	if len(list) == 0 {
		t.Fatal("no scenarios found")
	}
	expected := map[string]bool{"peak_hours": false, "maintenance": false, "network_issues": false, "database_load": false}
	for _, st := range list {
		expected[st] = true
	}
	for st, found := range expected {
		if !found {
			t.Errorf("expected scenario %s not in list", st)
		}
	}
}

func TestGetScenarioDelay(t *testing.T) {
	sm := NewManager()

	delay, strategy := sm.GetScenarioDelay("peak_hours", 0)
	if delay != 200*time.Millisecond {
		t.Errorf("peak_hours delay: expected 200ms, got %v", delay)
	}
	if strategy != FixedDelay {
		t.Errorf("peak_hours strategy: expected FixedDelay, got %v", strategy)
	}

	if d, _ := sm.GetScenarioDelay("maintenance", 500); d != 2*time.Second {
		t.Errorf("maintenance spike: expected 2s, got %v", d)
	}
	if d, _ := sm.GetScenarioDelay("maintenance", 100); d != 500*time.Millisecond {
		t.Errorf("maintenance normal: expected 500ms, got %v", d)
	}

	d1, _ := sm.GetScenarioDelay("database_load", 0)
	d2, _ := sm.GetScenarioDelay("database_load", 100)
	if d2 <= d1 {
		t.Error("database_load delay should increase over time")
	}

	d, st := sm.GetScenarioDelay("non_existent", 0)
	if d != 10*time.Millisecond || st != FixedDelay {
		t.Errorf("non-existent scenario defaults wrong: delay=%v, strategy=%v", d, st)
	}
}

func TestGetScenarioConfig(t *testing.T) {
	sm := NewManager()

	batchSize, snMode, maxCount, defaultCount := sm.GetScenarioConfig("peak_hours")
	if batchSize == 0 {
		t.Error("batch size should not be 0")
	}
	if maxCount == 0 {
		t.Error("max count should not be 0")
	}
	if defaultCount == 0 {
		t.Error("default count should not be 0")
	}
	if !snMode {
		t.Error("peak_hours should have ServiceNow mode enabled")
	}

	batchSize, snMode, maxCount, defaultCount = sm.GetScenarioConfig("non_existent")
	if batchSize != 100 || snMode || maxCount != 1000000 || defaultCount != 10000 {
		t.Error("non-existent scenario should return defaults")
	}
}

func TestParseDelay(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"100ms", 100 * time.Millisecond, false},
		{"1s", time.Second, false},
		{"500", 500 * time.Millisecond, false},
		{"1.5s", 1500 * time.Millisecond, false},
		{"invalid", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		d, err := ParseDelay(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseDelay(%q): expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseDelay(%q): unexpected error: %v", tt.input, err)
			}
			if d != tt.expected {
				t.Errorf("ParseDelay(%q): expected %v, got %v", tt.input, tt.expected, d)
			}
		}
	}
}

func TestParseDelayStrategy(t *testing.T) {
	tests := []struct {
		input    string
		expected DelayStrategy
	}{
		{"fixed", FixedDelay},
		{"FIXED", FixedDelay},
		{"random", RandomDelay},
		{"progressive", ProgressiveDelay},
		{"burst", BurstDelay},
		{"unknown", FixedDelay},
		{"", FixedDelay},
	}
	for _, tt := range tests {
		if got := ParseDelayStrategy(tt.input); got != tt.expected {
			t.Errorf("ParseDelayStrategy(%q): expected %v, got %v", tt.input, tt.expected, got)
		}
	}
}

func TestUserScenarioOverride(t *testing.T) {
	tempDir := t.TempDir()

	custom := Scenario{
		SchemaVersion:  "1.0.0",
		ScenarioName:   "Custom Peak Hours",
		ScenarioType:   "peak_hours",
		BaseDelay:      "300ms",
		DelayStrategy:  "fixed",
		ServiceNowMode: false,
		BatchSize:      50,
	}
	data, _ := json.Marshal(custom)
	if err := os.WriteFile(filepath.Join(tempDir, "custom.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	sm := &Manager{
		scenarios: make(map[string]*Scenario),
		userPath:  tempDir,
		validator: NewValidator(),
	}
	sm.loadEmbedded()
	if sm.GetScenario("peak_hours") == nil {
		t.Fatal("embedded peak_hours not found")
	}
	sm.loadUser()

	overridden := sm.GetScenario("peak_hours")
	if overridden == nil {
		t.Fatal("overridden peak_hours scenario not found")
	}
	if overridden.ScenarioName != "Custom Peak Hours" {
		t.Errorf("expected 'Custom Peak Hours', got %q", overridden.ScenarioName)
	}
	if overridden.BaseDelay != "300ms" {
		t.Errorf("expected base_delay '300ms', got %q", overridden.BaseDelay)
	}
}
