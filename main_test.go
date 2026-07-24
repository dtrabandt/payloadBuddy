package main

import (
	"testing"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

func TestSetupPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "8080"},
		{"9090", "9090"},
		{"invalid_port", "8080"},
		{"70000", "8080"},
		{"0", "8080"},
		{"-1", "8080"},
		{"65535", "65535"},
		{"65536", "8080"},
		{"1", "1"},
	}

	for _, tt := range tests {
		if got := setupPort(tt.input); got != tt.expected {
			t.Errorf("setupPort(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSetupPort_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty_string", "", "8080"},
		{"valid_mid_range", "9090", "9090"},
		{"valid_low", "80", "80"},
		{"valid_high", "65000", "65000"},
		{"port_minimum", "1", "1"},
		{"port_maximum", "65535", "65535"},
		{"port_zero", "0", "8080"},
		{"port_negative", "-1", "8080"},
		{"port_too_high", "65536", "8080"},
		{"invalid_string", "invalid", "8080"},
		{"mixed_string", "80abc", "8080"},
		{"float_string", "80.5", "8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setupPort(tt.input); got != tt.expected {
				t.Errorf("setupPort(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestScenarioUsageContext(t *testing.T) {
	tests := []struct {
		scenario string
		wantHint bool
	}{
		{"peak_hours", true},
		{"maintenance", true},
		{"network_issues", true},
		{"database_load", true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		got := scenarioUsageContext(tt.scenario)
		if tt.wantHint && got == "" {
			t.Errorf("scenarioUsageContext(%q) returned empty, want non-empty", tt.scenario)
		}
		if !tt.wantHint && got != "" {
			t.Errorf("scenarioUsageContext(%q) = %q, want empty", tt.scenario, got)
		}
	}
}

func TestPrintServiceNowScenarios_NoPanic(t *testing.T) {
	sm := scenarios.NewManager()
	printServiceNowScenarios(sm)

	scenarios := sm.ListScenarios()
	if len(scenarios) == 0 {
		t.Error("Expected at least some scenarios to be available")
	}
	for _, scenarioType := range scenarios {
		if sm.GetScenario(scenarioType) == nil {
			t.Errorf("GetScenario(%q) returned nil", scenarioType)
		}
	}
}

func TestPrintUsageExamples_NoPanic(t *testing.T) {
	sm := scenarios.NewManager()
	authCfg := auth.Setup(false, "", "")
	printUsageExamples("8080", sm, authCfg)
	printUsageExamples("9999", sm, authCfg)
}

func TestPrintStartupInfo_NoPanic(t *testing.T) {
	sm := scenarios.NewManager()
	authCfg := auth.Setup(false, "", "")
	printStartupInfo("8080", sm, authCfg)
}

func TestRegisterPluginsAndStart_PortLogic(t *testing.T) {
	tests := []struct {
		name         string
		portParam    string
		expectedPort string
	}{
		{"default_port", "8080", "8080"},
		{"custom_valid_port", "9999", "9999"},
		{"invalid_port_fallback", "invalid", "8080"},
		{"out_of_range_fallback", "70000", "8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*paramPort = tt.portParam
			if got := setupPort(*paramPort); got != tt.expectedPort {
				t.Errorf("Expected port %s, got %s", tt.expectedPort, got)
			}
		})
	}
	*paramPort = "8080" // restore default
}

func TestStartHTTPServer_Configuration(t *testing.T) {
	t.Skip("startHTTPServer calls ListenAndServe which blocks — covered by integration tests")
}
