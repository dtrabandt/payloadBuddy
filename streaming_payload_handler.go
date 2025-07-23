package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// StreamItem represents a single object in the streamed JSON payload
type StreamItem struct {
	ID        int       `json:"id"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	SysID     string    `json:"sys_id,omitempty"` // ServiceNow style
	Number    string    `json:"number,omitempty"` // ServiceNow ticket number
	State     string    `json:"state,omitempty"`  // ServiceNow state
}

// DelayStrategy defines different delay patterns
type DelayStrategy int

const (
	NoDelay DelayStrategy = iota
	FixedDelay
	RandomDelay
	ProgressiveDelay
	BurstDelay
)

// Helper function to parse duration parameters
func getDurationParam(r *http.Request, param string, defaultValue time.Duration) time.Duration {
	val := r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}

	// Try parsing as duration first (e.g., "100ms", "1s")
	if duration, err := time.ParseDuration(val); err == nil {
		return duration
	}

	// Fallback: parse as milliseconds
	if ms, err := strconv.Atoi(val); err == nil {
		return time.Duration(ms) * time.Millisecond
	}

	return defaultValue
}

// Helper function to parse integer parameters
func getIntParam(r *http.Request, param string, defaultValue int) int {
	val := r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}

	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}

	return defaultValue
}

// Helper function to parse delay strategy
func getDelayStrategy(r *http.Request) DelayStrategy {
	strategy := strings.ToLower(r.URL.Query().Get("strategy"))
	switch strategy {
	case "fixed":
		return FixedDelay
	case "random":
		return RandomDelay
	case "progressive":
		return ProgressiveDelay
	case "burst":
		return BurstDelay
	default:
		return FixedDelay
	}
}

// Helper function to generate ServiceNow-style sys_id
func generateSysID() string {
	chars := "abcdef0123456789"
	result := make([]byte, 32)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// Helper function to apply delay based on strategy and scenario
func applyDelay(ctx context.Context, strategy DelayStrategy, baseDelay time.Duration, scenario string, itemIndex int) error {
	var delay time.Duration

	// ServiceNow scenario-based delays
	switch scenario {
	case "peak_hours":
		delay = 200 * time.Millisecond
	case "maintenance":
		if itemIndex%500 == 0 {
			delay = 2 * time.Second // Maintenance spike
		} else {
			delay = 500 * time.Millisecond
		}
	case "network_issues":
		if rand.Float32() < 0.1 { // 10% chance of network spike
			delay = time.Duration(rand.Intn(3000)) * time.Millisecond
		} else {
			delay = baseDelay
		}
	case "database_load":
		dbLoadDelay := time.Duration(itemIndex/100) * 10 * time.Millisecond
		delay = baseDelay + dbLoadDelay
	default:
		// Apply strategy-based delay
		switch strategy {
		case NoDelay:
			return nil
		case FixedDelay:
			delay = baseDelay
		case RandomDelay:
			delay = time.Duration(rand.Int63n(int64(baseDelay * 2)))
		case ProgressiveDelay:
			delay = baseDelay * time.Duration(itemIndex/1000+1)
		case BurstDelay:
			if itemIndex%100 == 0 && itemIndex > 0 {
				delay = baseDelay * 10 // Long pause after burst
			} else {
				delay = baseDelay / 10 // Short pause between items
			}
		}
	}

	if delay <= 0 {
		return nil
	}

	// Context-aware delay
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// StreamingPayloadHandler streams large JSON data in chunks with configurable delays
//
// Query Parameters:
//   - count: Number of items to stream (default: 10000)
//   - delay: Base delay between items (e.g., "100ms", "1s", or milliseconds as integer)
//   - strategy: Delay strategy ("fixed", "random", "progressive", "burst")
//   - scenario: ServiceNow scenarios ("peak_hours", "maintenance", "network_issues", "database_load")
//   - batch_size: Items per flush batch (default: 100)
//   - servicenow: Generate ServiceNow-style fields (default: false)
//
// Examples:
//   - /stream?count=1000&delay=100ms&strategy=random
//   - /stream?scenario=peak_hours&servicenow=true
//   - /stream?delay=50ms&strategy=progressive&batch_size=50
func StreamingPayloadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse parameters
	count := getIntParam(r, "count", 10000)
	baseDelay := getDurationParam(r, "delay", 10)
	strategy := getDelayStrategy(r)
	scenario := strings.ToLower(r.URL.Query().Get("scenario"))
	batchSize := getIntParam(r, "batch_size", 100)
	serviceNowMode := r.URL.Query().Get("servicenow") == "true"

	// Validate parameters
	if count <= 0 || count > 1000000 { // Reasonable limits
		http.Error(w, "Count must be between 1 and 1,000,000", http.StatusBadRequest)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache")

	// Get flusher for real-time streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Start JSON array
	w.Write([]byte("[\n"))
	flusher.Flush()

	// Stream items
	for i := 0; i < count; i++ {
		// Check for client cancellation
		select {
		case <-ctx.Done():
			// Client disconnected, clean exit
			w.Write([]byte("\n]"))
			return
		default:
		}

		// Create item
		var item StreamItem
		if serviceNowMode {
			item = StreamItem{
				ID:        i,
				Value:     fmt.Sprintf("ServiceNow Record %d", i),
				Timestamp: time.Now(),
				SysID:     generateSysID(),
				Number:    fmt.Sprintf("INC%07d", i),
				State:     []string{"New", "In Progress", "Resolved", "Closed"}[i%4],
			}
		} else {
			item = StreamItem{
				ID:        i,
				Value:     fmt.Sprintf("streamed data %d", i),
				Timestamp: time.Now(),
			}
		}

		// Marshal item
		data, err := json.Marshal(item)
		if err != nil {
			http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
			return
		}

		// Write separator for items after the first
		if i > 0 {
			w.Write([]byte(",\n"))
		}

		// Write item
		w.Write(data)

		// Apply delay
		if err := applyDelay(ctx, strategy, baseDelay, scenario, i); err != nil {
			// Context cancelled during delay
			w.Write([]byte("\n]"))
			return
		}

		// Flush in batches
		if i%batchSize == 0 {
			flusher.Flush()
		}
	}

	// Close JSON array
	w.Write([]byte("\n]"))
	flusher.Flush()
}
