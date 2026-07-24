package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/dtrabandt/payloadBuddy/internal/handlers"
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

// TestNewHTTPServer_Timeouts covers CODE_REVIEW #4: a 30s WriteTimeout truncated every
// stream longer than that into an HTTP 200 with malformed JSON.
func TestNewHTTPServer_Timeouts(t *testing.T) {
	server := newHTTPServer("8080", http.NewServeMux())

	if server.WriteTimeout != 0 {
		t.Errorf("WriteTimeout must stay disabled so streams are not truncated, got %v", server.WriteTimeout)
	}
	if server.ReadHeaderTimeout != 10*time.Second {
		t.Errorf("Expected ReadHeaderTimeout 10s, got %v", server.ReadHeaderTimeout)
	}
	if server.IdleTimeout != 120*time.Second {
		t.Errorf("Expected IdleTimeout 120s, got %v", server.IdleTimeout)
	}
	if server.Addr != ":8080" {
		t.Errorf("Expected addr :8080, got %s", server.Addr)
	}
}

// TestStreamTerminatesUnderRealServer covers CODE_REVIEW #4 end to end. Every other
// streaming test uses httptest.NewRecorder(), which has no write deadline — which is
// exactly why a truncated stream was invisible to CI. This one runs a real
// http.Server configured by newHTTPServer and streams for longer than the write
// deadline used to be.
func TestStreamTerminatesUnderRealServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long streaming test in short mode")
	}

	sm := scenarios.NewManager()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stream_payload", handlers.StreamingPayloadPlugin{SM: sm}.Handler())

	ts := httptest.NewUnstartedServer(mux)
	configured := newHTTPServer("0", mux)
	ts.Config.ReadHeaderTimeout = configured.ReadHeaderTimeout
	ts.Config.ReadTimeout = configured.ReadTimeout
	ts.Config.WriteTimeout = configured.WriteTimeout
	ts.Config.IdleTimeout = configured.IdleTimeout
	ts.Start()
	defer ts.Close()

	// 4000 items at 10ms is ~40s of streaming — past the 30s deadline that used to
	// cut the response off after roughly 2,900 items.
	const wantItems = 4000
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	url := fmt.Sprintf("%s/stream_payload?count=%d&delay=10ms", ts.URL, wantItems)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Failed to build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read the stream: %v", err)
	}
	if !strings.HasSuffix(strings.TrimSpace(string(body)), "]") {
		t.Error("Stream was truncated: body does not end with a closing ']'")
	}

	var items []map[string]any
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatalf("Stream is not valid JSON: %v", err)
	}
	if len(items) != wantItems {
		t.Errorf("Expected %d items, got %d", wantItems, len(items))
	}
}
