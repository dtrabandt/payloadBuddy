package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

func TestStreamingPayloadHandler_Basic(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	req := httptest.NewRequest("GET", "/stream_payload?count=3&delay=1ms", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}
	body := w.Body.String()
	if !strings.HasPrefix(body, "[") || !strings.HasSuffix(strings.TrimSpace(body), "]") {
		t.Errorf("Expected body to be a JSON array, got %s", body[:min(50, len(body))])
	}
}

func TestStreamingPayloadHandler_WithParameters(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	req := httptest.NewRequest("GET", "/stream_payload?count=5", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	var items []StreamItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
}

func TestStreamingPayloadHandler_ServiceNowMode(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	req := httptest.NewRequest("GET", "/stream_payload?count=3&servicenow=true", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	var items []StreamItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if len(items) > 0 {
		item := items[0]
		if item.SysID == "" {
			t.Error("Expected SysID to be set in ServiceNow mode")
		}
		if item.Number == "" {
			t.Error("Expected Number to be set in ServiceNow mode")
		}
		if item.State == "" {
			t.Error("Expected State to be set in ServiceNow mode")
		}
		if !strings.Contains(item.Number, "INC") {
			t.Errorf("Expected incident number format, got %s", item.Number)
		}
	}
}

func TestStreamingPayloadHandler_InvalidCount(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	req := httptest.NewRequest("GET", "/stream_payload?count=2000000", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid count, got %d", w.Code)
	}
}

func TestStreamingPayloadHandler_DelayParameter(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	start := time.Now()
	req := httptest.NewRequest("GET", "/stream_payload?count=3&delay=20ms", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	elapsed := time.Since(start)

	if elapsed < 55*time.Millisecond {
		t.Errorf("Expected delay to be applied, took only %v", elapsed)
	}
}

func TestStreamingPayloadHandler_Scenarios(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	for _, scenario := range []string{"peak_hours", "maintenance", "network_issues", "database_load"} {
		t.Run(scenario, func(t *testing.T) {
			params := url.Values{}
			params.Add("count", "2")
			params.Add("scenario", scenario)

			req := httptest.NewRequest("GET", "/stream_payload?"+params.Encode(), nil)
			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for scenario %s, got %d", scenario, w.Code)
			}
			body := w.Body.String()
			if !strings.HasPrefix(body, "[") || !strings.HasSuffix(strings.TrimSpace(body), "]") {
				t.Errorf("Expected valid JSON array for scenario %s", scenario)
			}
		})
	}
}

func TestStreamingPayloadHandler_DelayStrategies(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	for _, strategy := range []string{"fixed", "random", "progressive", "burst"} {
		t.Run(strategy, func(t *testing.T) {
			params := url.Values{}
			params.Add("count", "3")
			params.Add("delay", "1ms")
			params.Add("strategy", strategy)

			req := httptest.NewRequest("GET", "/stream_payload?"+params.Encode(), nil)
			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for strategy %s, got %d", strategy, w.Code)
			}
		})
	}
}

func TestStreamingPayloadHandler_AuthenticationRequired(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)
	authCfg := auth.Setup(true, "streamuser", "streampass")

	// Without credentials
	req := httptest.NewRequest("GET", "/stream_payload?count=1&delay=1ms", nil)
	w := httptest.NewRecorder()
	authCfg.Middleware(handler)(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", w.Code)
	}

	// Wrong credentials
	req = httptest.NewRequest("GET", "/stream_payload?count=1&delay=1ms", nil)
	req.SetBasicAuth("wrong", "credentials")
	w = httptest.NewRecorder()
	authCfg.Middleware(handler)(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 with wrong auth, got %d", w.Code)
	}

	// Correct credentials
	req = httptest.NewRequest("GET", "/stream_payload?count=1&delay=1ms", nil)
	req.SetBasicAuth("streamuser", "streampass")
	w = httptest.NewRecorder()
	authCfg.Middleware(handler)(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 with correct auth, got %d", w.Code)
	}
}

func TestStreamingPayloadHandler_EdgeCases(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewStreamingHandler(sm)

	tests := []struct {
		name string
		url  string
	}{
		{"with_batch_size", "/stream_payload?count=3&batch_size=1&delay=1ms"},
		{"with_servicenow_false", "/stream_payload?count=2&servicenow=false&delay=1ms"},
		{"strategy_combinations", "/stream_payload?count=1&strategy=fixed&delay=1ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", tt.url, w.Code)
			}
		})
	}
}

func TestGetDurationParam(t *testing.T) {
	tests := []struct {
		name         string
		paramValue   string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"empty uses default", "", 100 * time.Millisecond, 100 * time.Millisecond},
		{"valid ms string", "250ms", 100 * time.Millisecond, 250 * time.Millisecond},
		{"valid seconds", "2s", 100 * time.Millisecond, 2 * time.Second},
		{"integer as ms", "500", 100 * time.Millisecond, 500 * time.Millisecond},
		{"invalid uses default", "invalid", 200 * time.Millisecond, 200 * time.Millisecond},
		{"negative integer", "-100", 50 * time.Millisecond, -100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?delay="+tt.paramValue, nil)
			result := getDurationParam(req, "delay", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name         string
		paramValue   string
		defaultValue int
		expected     int
	}{
		{"empty uses default", "", 1000, 1000},
		{"valid integer", "5000", 1000, 5000},
		{"zero value", "0", 1000, 0},
		{"negative value", "-100", 1000, -100},
		{"invalid uses default", "invalid", 2000, 2000},
		{"float uses default", "123.45", 500, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?count="+tt.paramValue, nil)
			result := getIntParam(req, "count", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestApplyDelay_NoDelay(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := applyDelay(ctx, scenarios.NoDelay, 100*time.Millisecond, "", 0, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Errorf("NoDelay took too long: %v", elapsed)
	}
}

func TestApplyDelay_ZeroBaseDelay(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := applyDelay(ctx, scenarios.FixedDelay, 0, "", 0, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Errorf("Zero base delay took too long: %v", elapsed)
	}
}

func TestApplyDelay_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := applyDelay(ctx, scenarios.FixedDelay, 100*time.Millisecond, "", 0, nil)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestApplyDelay_StrategiesNoError(t *testing.T) {
	ctx := context.Background()
	sm := scenarios.NewManager()

	tests := []struct {
		name      string
		strategy  scenarios.DelayStrategy
		baseDelay time.Duration
		scenario  string
		itemIndex int
	}{
		{"fixed_delay", scenarios.FixedDelay, 1 * time.Millisecond, "", 0},
		{"random_delay", scenarios.RandomDelay, 1 * time.Millisecond, "", 0},
		{"progressive_item_0", scenarios.ProgressiveDelay, 1 * time.Millisecond, "", 0},
		{"burst_no_spike", scenarios.BurstDelay, 1 * time.Millisecond, "", 1},
		{"database_load_sm", scenarios.FixedDelay, 0, "database_load", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useSM *scenarios.Manager
			if tt.scenario != "" {
				useSM = sm
			}
			if err := applyDelay(ctx, tt.strategy, tt.baseDelay, tt.scenario, tt.itemIndex, useSM); err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestApplyDelay_NetworkIssuesScenario(t *testing.T) {
	ctx := context.Background()
	sm := scenarios.NewManager()

	// A single network_issues delay must complete without error.
	if err := applyDelay(ctx, scenarios.FixedDelay, 1*time.Millisecond, "network_issues", 0, sm); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestStreamingPayloadPlugin_OpenAPISpec(t *testing.T) {
	sm := scenarios.NewManager()
	plugin := StreamingPayloadPlugin{SM: sm}
	spec := plugin.OpenAPISpec()

	if spec.Path != "/stream_payload" {
		t.Errorf("Wrong path: got %v want /stream_payload", spec.Path)
	}
	if spec.Operation.Get == nil {
		t.Fatal("Missing GET operation")
	}

	expectedParams := []string{"count", "delay", "strategy", "scenario", "batch_size", "servicenow"}
	paramNames := make(map[string]bool)
	for _, param := range spec.Operation.Get.Parameters {
		paramNames[param.Name] = true
	}
	for _, name := range expectedParams {
		if !paramNames[name] {
			t.Errorf("Missing parameter: %s", name)
		}
	}

	if _, exists := spec.Schemas["StreamItem"]; !exists {
		t.Error("Missing StreamItem schema")
	}
	streamItemSchema := spec.Schemas["StreamItem"]
	for _, prop := range []string{"id", "value", "timestamp", "sys_id", "number", "state"} {
		if _, exists := streamItemSchema.Properties[prop]; !exists {
			t.Errorf("Missing property %s in StreamItem schema", prop)
		}
	}
}
