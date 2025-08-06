package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestPaginatedPayloadHandler(t *testing.T) {
	// Disable auth for tests
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

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
			expectedItems:  100, // default limit
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
			name:           "Last page - no more data",
			queryParams:    "total=150&limit=100&offset=100",
			expectedStatus: http.StatusOK,
			expectedItems:  50, // Only 50 items left
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
			checkMetadata: func(t *testing.T, metadata PaginationMetadata) {
				// Will be checked in items validation
			},
		},
		{
			name:           "Invalid total count",
			queryParams:    "total=2000000", // Exceeds max
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

			PaginatedPayloadHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus != http.StatusOK {
				return // Skip further checks for error cases
			}

			var response PaginatedResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if len(response.Result) != tt.expectedItems {
				t.Errorf("Expected %d items, got %d", tt.expectedItems, len(response.Result))
			}

			// Check metadata if provided
			if tt.checkMetadata != nil {
				tt.checkMetadata(t, response.Metadata)
			}

			// Validate item structure
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

				// Check ServiceNow fields if enabled
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
	// Disable auth for tests
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	// Test that items are correctly sequenced across pages
	req1 := httptest.NewRequest("GET", "/paginated_payload?total=250&limit=100&offset=0", nil)
	w1 := httptest.NewRecorder()
	PaginatedPayloadHandler(w1, req1)

	req2 := httptest.NewRequest("GET", "/paginated_payload?total=250&limit=100&offset=100", nil)
	w2 := httptest.NewRecorder()
	PaginatedPayloadHandler(w2, req2)

	var response1, response2 PaginatedResponse
	json.NewDecoder(w1.Body).Decode(&response1)
	json.NewDecoder(w2.Body).Decode(&response2)

	// Check that second page starts where first page ended
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
	// Disable auth for tests
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	// Test concurrent requests
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(offset int) {
			req := httptest.NewRequest("GET", "/paginated_payload?limit=10&offset="+strconv.Itoa(offset*10), nil)
			w := httptest.NewRecorder()
			PaginatedPayloadHandler(w, req)

			var response PaginatedResponse
			json.NewDecoder(w.Body).Decode(&response)
			results <- len(response.Result)
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		itemCount := <-results
		if itemCount != 10 {
			t.Errorf("Expected 10 items, got %d", itemCount)
		}
	}
}

func TestPaginatedPayloadHandlerPerformance(t *testing.T) {
	// Disable auth for tests
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	start := time.Now()

	req := httptest.NewRequest("GET", "/paginated_payload?limit=1000", nil)
	w := httptest.NewRecorder()
	PaginatedPayloadHandler(w, req)

	duration := time.Since(start)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should complete within reasonable time (adjust threshold as needed)
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
	// Disable auth for tests
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	start := time.Now()

	req := httptest.NewRequest("GET", "/paginated_payload?delay=50ms&limit=10", nil)
	w := httptest.NewRecorder()
	PaginatedPayloadHandler(w, req)

	duration := time.Since(start)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should have delayed for at least 50ms
	if duration < 50*time.Millisecond {
		t.Errorf("Expected delay of at least 50ms, got %v", duration)
	}
}

func TestPaginatedPayloadHandlerHeaders(t *testing.T) {
	// Disable auth for tests
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	req := httptest.NewRequest("GET", "/paginated_payload", nil)
	w := httptest.NewRecorder()

	PaginatedPayloadHandler(w, req)

	// Check headers
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	if cacheControl := w.Header().Get("Cache-Control"); cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control no-cache, got %s", cacheControl)
	}
}

func TestPaginationBoundaryConditions(t *testing.T) {
	// Disable auth for tests
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	tests := []struct {
		name        string
		queryParams string
		expectError bool
	}{
		{
			name:        "Maximum limit",
			queryParams: "limit=1000",
			expectError: false,
		},
		{
			name:        "Limit exceeds maximum - should be capped",
			queryParams: "limit=1500",
			expectError: false,
		},
		{
			name:        "Negative offset - should be reset to 0",
			queryParams: "offset=-10",
			expectError: false,
		},
		{
			name:        "Zero page - should be reset to 1",
			queryParams: "page=0&size=50",
			expectError: false,
		},
		{
			name:        "Negative page - should be reset to 1",
			queryParams: "page=-1&size=50",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/paginated_payload?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			PaginatedPayloadHandler(w, req)

			if tt.expectError {
				if w.Code == http.StatusOK {
					t.Error("Expected error status, got 200")
				}
			} else {
				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200, got %d", w.Code)
				}

				var response PaginatedResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Should have valid response structure
				if response.Metadata.TotalCount <= 0 {
					t.Error("Expected positive total count")
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkPaginatedPayloadHandler(b *testing.B) {
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	req := httptest.NewRequest("GET", "/paginated_payload?limit=100", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		PaginatedPayloadHandler(w, req)
	}
}

func BenchmarkPaginatedPayloadHandlerLargeLimit(b *testing.B) {
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	req := httptest.NewRequest("GET", "/paginated_payload?limit=1000", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		PaginatedPayloadHandler(w, req)
	}
}

func BenchmarkPaginatedPayloadHandlerServiceNow(b *testing.B) {
	originalAuth := *enableAuth
	*enableAuth = false
	defer func() { *enableAuth = originalAuth }()

	req := httptest.NewRequest("GET", "/paginated_payload?servicenow=true&limit=100", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		PaginatedPayloadHandler(w, req)
	}
}

func TestPaginatedPayloadHandlerAuthentication(t *testing.T) {
	// Save original auth state
	originalAuth := *enableAuth
	originalUsername := authUsername
	originalPassword := authPassword
	defer func() {
		*enableAuth = originalAuth
		authUsername = originalUsername
		authPassword = originalPassword
	}()

	// Enable auth for testing
	*enableAuth = true
	authUsername = "testuser"
	authPassword = "testpass"

	tests := []struct {
		name           string
		useAuth        bool
		expectedStatus int
		description    string
	}{
		{
			name:           "Without authentication should return 401",
			useAuth:        false,
			expectedStatus: http.StatusUnauthorized,
			description:    "Paginated endpoint should require authentication when auth is enabled",
		},
		{
			name:           "With correct authentication should return 200",
			useAuth:        true,
			expectedStatus: http.StatusOK,
			description:    "Paginated endpoint should work with correct authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/paginated_payload?limit=10", nil)

			// Add authentication if required
			if tt.useAuth {
				req.SetBasicAuth(authUsername, authPassword)
			}

			w := httptest.NewRecorder()

			// Use middleware wrapper for authentication testing
			handler := basicAuthMiddleware(PaginatedPayloadHandler)
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.description, tt.expectedStatus, w.Code)
			}

			// Additional checks for successful authentication
			if tt.useAuth && w.Code == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected content-type application/json, got %s", contentType)
				}

				// Verify it's valid JSON response
				var response PaginatedResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				if len(response.Result) != 10 {
					t.Errorf("Expected 10 items in result, got %d", len(response.Result))
				}

				if response.Metadata.TotalCount == 0 {
					t.Error("Expected non-zero total count in metadata")
				}
			}
		})
	}
}
