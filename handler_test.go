package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDataHandler_ReturnsCorrectJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	w := httptest.NewRecorder()

	DataHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Status = %d; want 200", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}

	var items []Item
	err := json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		t.Fatalf("JSON decode failed: %v", err)
	}
	if len(items) != 100000 {
		t.Errorf("Array length = %d; want 100000", len(items))
	}
	// Optional: Spot-check Inhalt
	if items[0].ID != 0 || items[99999].ID != 99999 {
		t.Errorf("IDs falsch: got %d...%d", items[0].ID, items[99999].ID)
	}
}
