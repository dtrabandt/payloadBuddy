package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// helper function to create authenticated request (shared with huge_payload tests)
func createStreamAuthRequest(method, path string, username, password string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if username != "" && password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		req.Header.Set("Authorization", "Basic "+auth)
	}
	return req
}

func TestStreamingPayloadHandler_Basic(t *testing.T) {
	*enableAuth = false
	req := httptest.NewRequest("GET", "/stream_payload?count=3&delay=1ms", nil)
	w := httptest.NewRecorder()

	StreamingPayloadHandler(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	if te := resp.Header.Get("Transfer-Encoding"); !strings.Contains(te, "chunked") {
		t.Errorf("Expected Transfer-Encoding chunked, got %s", te)
	}

	body := w.Body.String()
	if !strings.HasPrefix(body, "[") || !strings.HasSuffix(body, "]") {
		t.Errorf("Expected body to be a JSON array, got %s", body[:50])
	}
}

func TestStreamingPayloadHandler_WithParameters(t *testing.T) {
	*enableAuth = false
	// Test with count parameter
	req := httptest.NewRequest("GET", "/stream_payload?count=5", nil)
	w := httptest.NewRecorder()

	StreamingPayloadHandler(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()

	// Parse JSON to verify count
	var items []StreamItem
	if err := json.Unmarshal([]byte(body), &items); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
}

func TestStreamingPayloadHandler_ServiceNowMode(t *testing.T) {
	*enableAuth = false
	req := httptest.NewRequest("GET", "/stream_payload?count=3&servicenow=true", nil)
	w := httptest.NewRecorder()

	StreamingPayloadHandler(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()

	// Parse JSON to verify ServiceNow fields
	var items []StreamItem
	if err := json.Unmarshal([]byte(body), &items); err != nil {
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
	*enableAuth = false
	// Test with count too high
	req := httptest.NewRequest("GET", "/stream_payload?count=2000000", nil)
	w := httptest.NewRecorder()

	StreamingPayloadHandler(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid count, got %d", resp.StatusCode)
	}
}

func TestStreamingPayloadHandler_DelayParameter(t *testing.T) {
	*enableAuth = false
	start := time.Now()

	req := httptest.NewRequest("GET", "/stream_payload?count=3&delay=20ms", nil)
	w := httptest.NewRecorder()

	StreamingPayloadHandler(w, req)

	elapsed := time.Since(start)

	// Should take at least 60ms (3 items with 20ms delay each, minus some)
	if elapsed < 55*time.Millisecond {
		t.Errorf("Expected delay to be applied, took only %v", elapsed)
	}
}

func TestStreamingPayloadHandler_Scenarios(t *testing.T) {
	scenarios := []string{"peak_hours", "maintenance", "network_issues", "database_load"}

	for _, scenario := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			params := url.Values{}
			params.Add("count", "5")
			params.Add("scenario", scenario)

			*enableAuth = false
			req := httptest.NewRequest("GET", "/stream_payload?"+params.Encode(), nil)
			w := httptest.NewRecorder()

			StreamingPayloadHandler(w, req)
			resp := w.Result()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for scenario %s, got %d", scenario, resp.StatusCode)
			}

			body := w.Body.String()
			if !strings.HasPrefix(body, "[") || !strings.HasSuffix(body, "]") {
				t.Errorf("Expected valid JSON array for scenario %s", scenario)
			}
		})
	}
}

func TestStreamingPayloadHandler_DelayStrategies(t *testing.T) {
	strategies := []string{"fixed", "random", "progressive", "burst"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			params := url.Values{}
			params.Add("count", "5")
			params.Add("delay", "1ms")
			params.Add("strategy", strategy)

			*enableAuth = false
			req := httptest.NewRequest("GET", "/stream_payload?"+params.Encode(), nil)
			w := httptest.NewRecorder()

			StreamingPayloadHandler(w, req)
			resp := w.Result()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for strategy %s, got %d", strategy, resp.StatusCode)
			}
		})
	}
}

// TestStreamingPayloadHandler_AuthenticationRequired tests that authentication is required when enabled.
func TestStreamingPayloadHandler_AuthenticationRequired(t *testing.T) {
	*enableAuth = true
	authUsername = "streamuser"
	authPassword = "streampass"

	// Test without credentials
	req := httptest.NewRequest("GET", "/stream_payload?count=1&delay=1ms", nil)
	w := httptest.NewRecorder()

	basicAuthMiddleware(StreamingPayloadHandler)(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without auth, got %d", resp.StatusCode)
	}

	// Test with wrong credentials
	req = createStreamAuthRequest("GET", "/stream_payload?count=1&delay=1ms", "wrong", "credentials")
	w = httptest.NewRecorder()

	basicAuthMiddleware(StreamingPayloadHandler)(w, req)
	resp = w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with wrong auth, got %d", resp.StatusCode)
	}

	// Test with correct credentials
	req = createStreamAuthRequest("GET", "/stream_payload?count=1&delay=1ms", "streamuser", "streampass")
	w = httptest.NewRecorder()

	basicAuthMiddleware(StreamingPayloadHandler)(w, req)
	resp = w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with correct auth, got %d", resp.StatusCode)
	}
}

func TestGetDurationParam(t *testing.T) {
	tests := []struct {
		name         string
		paramValue   string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			name:         "empty parameter uses default",
			paramValue:   "",
			defaultValue: 100 * time.Millisecond,
			expected:     100 * time.Millisecond,
		},
		{
			name:         "valid duration string",
			paramValue:   "250ms",
			defaultValue: 100 * time.Millisecond,
			expected:     250 * time.Millisecond,
		},
		{
			name:         "valid duration seconds",
			paramValue:   "2s",
			defaultValue: 100 * time.Millisecond,
			expected:     2 * time.Second,
		},
		{
			name:         "milliseconds as integer",
			paramValue:   "500",
			defaultValue: 100 * time.Millisecond,
			expected:     500 * time.Millisecond,
		},
		{
			name:         "invalid format uses default",
			paramValue:   "invalid",
			defaultValue: 200 * time.Millisecond,
			expected:     200 * time.Millisecond,
		},
		{
			name:         "negative number as milliseconds",
			paramValue:   "-100",
			defaultValue: 50 * time.Millisecond,
			expected:     -100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with the parameter
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
		{
			name:         "empty parameter uses default",
			paramValue:   "",
			defaultValue: 1000,
			expected:     1000,
		},
		{
			name:         "valid integer",
			paramValue:   "5000",
			defaultValue: 1000,
			expected:     5000,
		},
		{
			name:         "zero value",
			paramValue:   "0",
			defaultValue: 1000,
			expected:     0,
		},
		{
			name:         "negative value",
			paramValue:   "-100",
			defaultValue: 1000,
			expected:     -100,
		},
		{
			name:         "invalid format uses default",
			paramValue:   "invalid",
			defaultValue: 2000,
			expected:     2000,
		},
		{
			name:         "float format uses default",
			paramValue:   "123.45",
			defaultValue: 500,
			expected:     500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with the parameter
			req := httptest.NewRequest("GET", "/?count="+tt.paramValue, nil)

			result := getIntParam(req, "count", tt.defaultValue)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestApplyDelay_EdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		strategy  DelayStrategy
		baseDelay time.Duration
		scenario  string
		itemIndex int
		expectErr bool
	}{
		{
			name:      "no delay strategy",
			strategy:  NoDelay,
			baseDelay: 100 * time.Millisecond,
			scenario:  "",
			itemIndex: 0,
			expectErr: false,
		},
		{
			name:      "zero base delay",
			strategy:  FixedDelay,
			baseDelay: 0,
			scenario:  "",
			itemIndex: 0,
			expectErr: false,
		},
		{
			name:      "maintenance spike trigger",
			strategy:  FixedDelay,
			baseDelay: 10 * time.Millisecond,
			scenario:  "maintenance",
			itemIndex: 500, // Should trigger spike
			expectErr: false,
		},
		{
			name:      "maintenance no spike",
			strategy:  FixedDelay,
			baseDelay: 10 * time.Millisecond,
			scenario:  "maintenance",
			itemIndex: 499, // Should not trigger spike
			expectErr: false,
		},
		{
			name:      "database load progression",
			strategy:  FixedDelay,
			baseDelay: 5 * time.Millisecond,
			scenario:  "database_load",
			itemIndex: 1000,
			expectErr: false,
		},
		{
			name:      "burst strategy spike",
			strategy:  BurstDelay,
			baseDelay: 10 * time.Millisecond,
			scenario:  "",
			itemIndex: 100, // Should trigger long pause
			expectErr: false,
		},
		{
			name:      "burst strategy normal",
			strategy:  BurstDelay,
			baseDelay: 10 * time.Millisecond,
			scenario:  "",
			itemIndex: 99, // Should not trigger long pause
			expectErr: false,
		},
		{
			name:      "progressive delay",
			strategy:  ProgressiveDelay,
			baseDelay: 5 * time.Millisecond,
			scenario:  "",
			itemIndex: 2000,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := applyDelay(ctx, tt.strategy, tt.baseDelay, tt.scenario, tt.itemIndex)
			elapsed := time.Since(start)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For NoDelay or zero baseDelay, should complete quickly
			if tt.strategy == NoDelay || tt.baseDelay == 0 {
				if elapsed > 10*time.Millisecond {
					t.Errorf("Expected quick completion for no delay, took %v", elapsed)
				}
			}
		})
	}
}

func TestApplyDelay_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	err := applyDelay(ctx, FixedDelay, 100*time.Millisecond, "", 0)

	if err == nil {
		t.Error("Expected context cancellation error")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestApplyDelay_NetworkIssuesScenario(t *testing.T) {
	// Test network_issues scenario multiple times to hit the random 10% chance
	ctx := context.Background()

	hitLongDelay := false
	hitShortDelay := false

	// Run many iterations to increase chance of hitting both paths
	for i := 0; i < 100; i++ {
		start := time.Now()
		err := applyDelay(ctx, FixedDelay, 1*time.Millisecond, "network_issues", i)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Check if we hit a long delay (network spike) or short delay
		if elapsed > 50*time.Millisecond {
			hitLongDelay = true
		} else {
			hitShortDelay = true
		}

		// If we've hit both, we can break early
		if hitLongDelay && hitShortDelay {
			break
		}
	}

	// We should have hit at least the short delay path
	if !hitShortDelay {
		t.Error("Expected to hit short delay path in network_issues scenario")
	}

	// Note: We might not hit the long delay due to randomness, but that's okay
	// The important thing is we're testing the code path
}

// Test additional streaming handler edge cases
func TestStreamingPayloadHandler_EdgeCases(t *testing.T) {
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

			StreamingPayloadHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", tt.url, w.Code)
			}
		})
	}
}

// Test parameter parsing edge cases
func TestParameterParsing_EdgeCases(t *testing.T) {
	t.Run("getDurationParam_boundaries", func(t *testing.T) {
		tests := []struct {
			value    string
			expected time.Duration
		}{
			{"0", 0},
			{"-50", -50 * time.Millisecond},
			{"999999999", 999999999 * time.Millisecond},
		}

		for _, tt := range tests {
			req := httptest.NewRequest("GET", "/?param="+tt.value, nil)
			result := getDurationParam(req, "param", 100*time.Millisecond)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		}
	})
}
