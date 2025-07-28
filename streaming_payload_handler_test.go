package main

import (
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
	req := httptest.NewRequest("GET", "/stream_payload", nil)
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
	req := httptest.NewRequest("GET", "/stream_payload", nil)
	w := httptest.NewRecorder()
	
	basicAuthMiddleware(StreamingPayloadHandler)(w, req)
	resp := w.Result()
	
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without auth, got %d", resp.StatusCode)
	}
	
	// Test with wrong credentials
	req = createStreamAuthRequest("GET", "/stream_payload", "wrong", "credentials")
	w = httptest.NewRecorder()
	
	basicAuthMiddleware(StreamingPayloadHandler)(w, req)
	resp = w.Result()
	
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with wrong auth, got %d", resp.StatusCode)
	}
	
	// Test with correct credentials
	req = createStreamAuthRequest("GET", "/stream_payload", "streamuser", "streampass")
	w = httptest.NewRecorder()
	
	basicAuthMiddleware(StreamingPayloadHandler)(w, req)
	resp = w.Result()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with correct auth, got %d", resp.StatusCode)
	}
}
