package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

func newPaginatedHandler() http.HandlerFunc {
	return NewPaginatedHandler(scenarios.NewManager())
}

func TestPaginatedPayloadHandler(t *testing.T) {
	handler := newPaginatedHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedItems  int
		checkMetadata  func(t *testing.T, metadata PaginationMetadata)
	}{
		{
			name:           "Default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedItems:  100,
			checkMetadata: func(t *testing.T, metadata PaginationMetadata) {
				if metadata.TotalCount != 10000 {
					t.Errorf("Expected total_count 10000, got %d", metadata.TotalCount)
				}
				if metadata.Limit != 100 {
					t.Errorf("Expected limit 100, got %d", metadata.Limit)
				}
				if metadata.Offset != 0 {
					t.Errorf("Expected offset 0, got %d", metadata.Offset)
				}
				if !metadata.HasMore {
					t.Error("Expected has_more to be true")
				}
				if metadata.NextOffset == nil || *metadata.NextOffset != 100 {
					t.Errorf("Expected next_offset 100, got %v", metadata.NextOffset)
				}
			},
		},
		{
			name:           "Limit/Offset pagination",
			queryParams:    "limit=50&offset=100",
			expectedStatus: http.StatusOK,
			expectedItems:  50,
			checkMetadata: func(t *testing.T, metadata PaginationMetadata) {
				if metadata.Limit != 50 {
					t.Errorf("Expected limit 50, got %d", metadata.Limit)
				}
				if metadata.Offset != 100 {
					t.Errorf("Expected offset 100, got %d", metadata.Offset)
				}
				if metadata.NextOffset == nil || *metadata.NextOffset != 150 {
					t.Errorf("Expected next_offset 150, got %v", metadata.NextOffset)
				}
			},
		},
		{
			name:           "Page/Size pagination",
			queryParams:    "page=2&size=25",
			expectedStatus: http.StatusOK,
			expectedItems:  25,
			checkMetadata: func(t *testing.T, metadata PaginationMetadata) {
				if metadata.Page != 2 {
					t.Errorf("Expected page 2, got %d", metadata.Page)
				}
				if metadata.Size != 25 {
					t.Errorf("Expected size 25, got %d", metadata.Size)
				}
				if metadata.NextPage == nil || *metadata.NextPage != 3 {
					t.Errorf("Expected next_page 3, got %v", metadata.NextPage)
				}
			},
		},
		{
			name:           "Last page no more data",
			queryParams:    "total=150&limit=100&offset=100",
			expectedStatus: http.StatusOK,
			expectedItems:  50,
			checkMetadata: func(t *testing.T, metadata PaginationMetadata) {
				if metadata.TotalCount != 150 {
					t.Errorf("Expected total_count 150, got %d", metadata.TotalCount)
				}
				if metadata.HasMore {
					t.Error("Expected has_more to be false")
				}
				if metadata.NextOffset != nil {
					t.Errorf("Expected next_offset to be nil, got %v", metadata.NextOffset)
				}
			},
		},
		{
			name:           "Beyond data range",
			queryParams:    "total=100&offset=200",
			expectedStatus: http.StatusOK,
			expectedItems:  0,
			checkMetadata: func(t *testing.T, metadata PaginationMetadata) {
				if metadata.HasMore {
					t.Error("Expected has_more to be false")
				}
			},
		},
		{
			name:           "ServiceNow mode enabled",
			queryParams:    "servicenow=true&limit=5",
			expectedStatus: http.StatusOK,
			expectedItems:  5,
			checkMetadata:  nil,
		},
		{
			name:           "Invalid total count",
			queryParams:    "total=2000000",
			expectedStatus: http.StatusBadRequest,
			expectedItems:  0,
		},
		{
			name:           "Zero total count",
			queryParams:    "total=0",
			expectedStatus: http.StatusBadRequest,
			expectedItems:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/paginated_payload?"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedStatus != http.StatusOK {
				return
			}

			var response PaginatedResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if len(response.Result) != tt.expectedItems {
				t.Errorf("Expected %d items, got %d", tt.expectedItems, len(response.Result))
			}
			if tt.checkMetadata != nil {
				tt.checkMetadata(t, response.Metadata)
			}

			if len(response.Result) > 0 {
				item := response.Result[0]
				if item.ID <= 0 {
					t.Error("Expected positive ID")
				}
				if item.Value == "" {
					t.Error("Expected non-empty value")
				}
				if item.Timestamp.IsZero() {
					t.Error("Expected non-zero timestamp")
				}
				if req.URL.Query().Get("servicenow") == "true" {
					if item.SysID == "" {
						t.Error("Expected sys_id in ServiceNow mode")
					}
					if item.Number == "" {
						t.Error("Expected number in ServiceNow mode")
					}
					if item.State == "" {
						t.Error("Expected state in ServiceNow mode")
					}
				}
			}
		})
	}
}

func TestPaginatedPayloadHandlerItemSequence(t *testing.T) {
	handler := newPaginatedHandler()

	req1 := httptest.NewRequest("GET", "/paginated_payload?total=250&limit=100&offset=0", nil)
	w1 := httptest.NewRecorder()
	handler(w1, req1)

	req2 := httptest.NewRequest("GET", "/paginated_payload?total=250&limit=100&offset=100", nil)
	w2 := httptest.NewRecorder()
	handler(w2, req2)

	var response1, response2 PaginatedResponse
	if err := json.NewDecoder(w1.Body).Decode(&response1); err != nil {
		t.Fatalf("Failed to decode first response: %v", err)
	}
	if err := json.NewDecoder(w2.Body).Decode(&response2); err != nil {
		t.Fatalf("Failed to decode second response: %v", err)
	}

	if len(response1.Result) == 0 || len(response2.Result) == 0 {
		t.Fatal("Expected items in both responses")
	}

	lastItemPage1 := response1.Result[len(response1.Result)-1]
	firstItemPage2 := response2.Result[0]
	if firstItemPage2.ID != lastItemPage1.ID+1 {
		t.Errorf("Expected page 2 to start with ID %d, got %d", lastItemPage1.ID+1, firstItemPage2.ID)
	}
}

func TestPaginatedPayloadHandlerConcurrency(t *testing.T) {
	handler := newPaginatedHandler()
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := range numRequests {
		go func(offset int) {
			req := httptest.NewRequest("GET", "/paginated_payload?limit=10&offset="+strconv.Itoa(offset*10), nil)
			w := httptest.NewRecorder()
			handler(w, req)

			var response PaginatedResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				results <- 0
				return
			}
			results <- len(response.Result)
		}(i)
	}

	for range numRequests {
		if itemCount := <-results; itemCount != 10 {
			t.Errorf("Expected 10 items, got %d", itemCount)
		}
	}
}

func TestPaginatedPayloadHandlerPerformance(t *testing.T) {
	handler := newPaginatedHandler()
	start := time.Now()

	req := httptest.NewRequest("GET", "/paginated_payload?limit=1000", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	duration := time.Since(start)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if duration > 100*time.Millisecond {
		t.Errorf("Handler took too long: %v", duration)
	}

	var response PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(response.Result) != 1000 {
		t.Errorf("Expected 1000 items, got %d", len(response.Result))
	}
}

func TestPaginatedPayloadHandlerDelayParameter(t *testing.T) {
	handler := newPaginatedHandler()
	start := time.Now()

	req := httptest.NewRequest("GET", "/paginated_payload?delay=50ms&limit=10", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	duration := time.Since(start)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if duration < 50*time.Millisecond {
		t.Errorf("Expected delay of at least 50ms, got %v", duration)
	}
}

func TestPaginatedPayloadHandlerHeaders(t *testing.T) {
	handler := newPaginatedHandler()

	req := httptest.NewRequest("GET", "/paginated_payload", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Expected Cache-Control no-cache, got %s", cc)
	}
}

func TestPaginationBoundaryConditions(t *testing.T) {
	handler := newPaginatedHandler()

	tests := []struct {
		name        string
		queryParams string
		expectError bool
	}{
		{"maximum limit", "limit=1000", false},
		{"limit exceeds max capped", "limit=1500", false},
		{"negative offset reset to 0", "offset=-10", false},
		{"zero page reset to 1", "page=0&size=50", false},
		{"negative page reset to 1", "page=-1&size=50", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/paginated_payload?"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			handler(w, req)

			if tt.expectError {
				if w.Code == http.StatusOK {
					t.Error("Expected error status, got 200")
				}
				return
			}
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
				return
			}
			var response PaginatedResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if response.Metadata.TotalCount <= 0 {
				t.Error("Expected positive total count")
			}
		})
	}
}

func TestPaginatedPayloadHandlerAuthentication(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewPaginatedHandler(sm)
	authCfg := auth.Setup(true, "testuser", "testpass")

	// Without credentials
	req := httptest.NewRequest("GET", "/paginated_payload?limit=10", nil)
	w := httptest.NewRecorder()
	authCfg.Middleware(handler)(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", w.Code)
	}

	// With correct credentials
	req = httptest.NewRequest("GET", "/paginated_payload?limit=10", nil)
	req.SetBasicAuth("testuser", "testpass")
	w = httptest.NewRecorder()
	authCfg.Middleware(handler)(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 with correct auth, got %d", w.Code)
	}
	var response PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	if len(response.Result) != 10 {
		t.Errorf("Expected 10 items, got %d", len(response.Result))
	}
}

func TestPaginatedPayloadHandlerScenarios(t *testing.T) {
	sm := scenarios.NewManager()
	handler := NewPaginatedHandler(sm)

	tests := []struct {
		name     string
		scenario string
	}{
		{"peak_hours", "peak_hours"},
		{"database_load", "database_load"},
		{"invalid scenario uses defaults", "invalid_scenario"},
		{"empty scenario uses defaults", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawURL string
			if tt.scenario != "" {
				rawURL = "/paginated_payload?scenario=" + tt.scenario + "&limit=5"
			} else {
				rawURL = "/paginated_payload?limit=5"
			}

			req := httptest.NewRequest("GET", rawURL, nil)
			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
				return
			}
			var response PaginatedResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
				return
			}
			if len(response.Result) != 5 {
				t.Errorf("Expected 5 items, got %d", len(response.Result))
			}
			if response.Metadata.TotalCount == 0 {
				t.Error("Expected non-zero total count in metadata")
			}
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		startID int
	}{
		{"zero start", 0},
		{"mid page", 50},
		{"large offset", 999900},
		{"small start", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := createCursor(tt.startID)
			gotID, gotLimit := parseCursor(cursor, 100)

			if gotID != tt.startID {
				t.Errorf("startID: got %d, want %d", gotID, tt.startID)
			}
			if gotLimit != 100 {
				t.Errorf("limit: got %d, want 100", gotLimit)
			}
		})
	}
}

func TestParseCursorInvalid(t *testing.T) {
	tests := []struct {
		name         string
		cursor       string
		defaultLimit int
	}{
		{"empty cursor", "", 50},
		{"garbage", "not-valid-base64!!!", 50},
		{"valid base64 invalid json", "dGhpcyBpcyBub3QganNvbg==", 75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotLimit := parseCursor(tt.cursor, tt.defaultLimit)
			if gotID != 0 {
				t.Errorf("Expected id=0 for invalid cursor, got %d", gotID)
			}
			if gotLimit != tt.defaultLimit {
				t.Errorf("Expected default limit %d, got %d", tt.defaultLimit, gotLimit)
			}
		})
	}
}

func BenchmarkPaginatedPayloadHandler(b *testing.B) {
	handler := newPaginatedHandler()
	req := httptest.NewRequest("GET", "/paginated_payload?limit=100", nil)
	b.ResetTimer()
	for range b.N {
		w := httptest.NewRecorder()
		handler(w, req)
	}
}

func BenchmarkPaginatedPayloadHandlerServiceNow(b *testing.B) {
	handler := newPaginatedHandler()
	req := httptest.NewRequest("GET", "/paginated_payload?servicenow=true&limit=100", nil)
	b.ResetTimer()
	for range b.N {
		w := httptest.NewRecorder()
		handler(w, req)
	}
}
