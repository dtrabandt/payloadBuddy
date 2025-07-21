package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
)

func TestStreamingPayloadHandler(t *testing.T) {
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
