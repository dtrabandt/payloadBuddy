package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
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

// secureRandFloat32 generates a cryptographically secure random float32 between 0 and 1
func secureRandFloat32() (float32, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		return 0, err
	}
	return float32(n.Int64()) / float32(1<<24), nil
}

// secureRandIntn generates a cryptographically secure random int between 0 and n
func secureRandIntn(n int) (int, error) {
	bigN, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(bigN.Int64()), nil
}

// secureRandInt63n generates a cryptographically secure random int64 between 0 and n
func secureRandInt63n(n int64) (int64, error) {
	bigN, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		return 0, err
	}
	return bigN.Int64(), nil
}

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
		randIdx, err := secureRandIntn(len(chars))
		if err != nil {
			// Fallback to deterministic pattern if crypto/rand fails
			result[i] = chars[i%len(chars)]
		} else {
			result[i] = chars[randIdx]
		}
	}
	return string(result)
}

// Helper function to apply delay based on strategy and scenario
func applyDelay(ctx context.Context, strategy DelayStrategy, baseDelay time.Duration, scenario string, itemIndex int) error {
	var delay time.Duration

	// Check if we have a scenario configured
	if scenarioManager != nil && scenario != "" {
		calculatedDelay, calculatedStrategy := scenarioManager.GetScenarioDelay(scenario, itemIndex)

		// For network_issues scenario, we still need to apply random logic
		if scenario == "network_issues" {
			randFloat, err := secureRandFloat32()
			if err != nil {
				delay = calculatedDelay
			} else if randFloat < 0.1 { // 10% chance of network spike
				randInt, err := secureRandIntn(3000)
				if err != nil {
					delay = calculatedDelay
				} else {
					delay = time.Duration(randInt) * time.Millisecond
				}
			} else {
				delay = calculatedDelay
			}
		} else {
			delay = calculatedDelay
			strategy = calculatedStrategy
		}
	} else {
		// Fallback to legacy hardcoded scenario logic for backward compatibility
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
			randFloat, err := secureRandFloat32()
			if err != nil {
				delay = baseDelay
			} else if randFloat < 0.1 { // 10% chance of network spike
				randInt, err := secureRandIntn(3000)
				if err != nil {
					delay = baseDelay
				} else {
					delay = time.Duration(randInt) * time.Millisecond
				}
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
				randInt64, err := secureRandInt63n(int64(baseDelay * 2))
				if err != nil {
					delay = baseDelay // Fallback to fixed delay if crypto/rand fails
				} else {
					delay = time.Duration(randInt64)
				}
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
	}

	// Apply strategy-based modifications if not handled by scenario
	if scenario == "" || (scenarioManager == nil) {
		switch strategy {
		case NoDelay:
			return nil
		case FixedDelay:
			// delay already set
		case RandomDelay:
			randInt64, err := secureRandInt63n(int64(baseDelay * 2))
			if err != nil {
				delay = baseDelay // Fallback to fixed delay if crypto/rand fails
			} else {
				delay = time.Duration(randInt64)
			}
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

	// Parse basic parameters
	scenario := strings.ToLower(r.URL.Query().Get("scenario"))

	// Get scenario-based defaults if scenario manager is available and scenario is specified
	var defaultCount, maxCount, defaultBatchSize int
	var defaultServiceNowMode bool
	if scenarioManager != nil && scenario != "" {
		defaultBatchSize, defaultServiceNowMode, maxCount, defaultCount = scenarioManager.GetScenarioConfig(scenario)
	} else {
		// Use hardcoded defaults for backward compatibility
		defaultCount = 10000
		maxCount = 1000000
		defaultBatchSize = 100
		defaultServiceNowMode = false
	}

	// Parse parameters with scenario-aware defaults
	count := getIntParam(r, "count", defaultCount)
	baseDelay := getDurationParam(r, "delay", 10*time.Millisecond)
	strategy := getDelayStrategy(r)
	batchSize := getIntParam(r, "batch_size", defaultBatchSize)

	// ServiceNow mode: use scenario default unless explicitly overridden
	serviceNowMode := defaultServiceNowMode
	if serviceNowParam := r.URL.Query().Get("servicenow"); serviceNowParam != "" {
		serviceNowMode = serviceNowParam == "true"
	}

	// Validate parameters
	if count <= 0 || count > maxCount {
		http.Error(w, fmt.Sprintf("Count must be between 1 and %d", maxCount), http.StatusBadRequest)
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
	if _, err := w.Write([]byte("[\n")); err != nil {
		return
	}
	flusher.Flush()

	// Stream items
	for i := 0; i < count; i++ {
		// Check for client cancellation
		select {
		case <-ctx.Done():
			// Client disconnected, clean exit
			_, _ = w.Write([]byte("\n]"))
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
			if _, err := w.Write([]byte(",\n")); err != nil {
				return
			}
		}

		// Write item
		if _, err := w.Write(data); err != nil {
			return
		}

		// Apply delay
		if err := applyDelay(ctx, strategy, baseDelay, scenario, i); err != nil {
			// Context cancelled during delay
			_, _ = w.Write([]byte("\n]"))
			return
		}

		// Flush in batches
		if i%batchSize == 0 {
			flusher.Flush()
		}
	}

	// Close JSON array
	_, _ = w.Write([]byte("\n]"))
	flusher.Flush()
}

// OpenAPISpec returns the OpenAPI specification for the streaming payload endpoint
func (s StreamingPayloadPlugin) OpenAPISpec() OpenAPIPathSpec {
	return OpenAPIPathSpec{
		Path: "/stream_payload",
		Operation: OpenAPIPath{
			Get: &OpenAPIOperation{
				Summary:     "Get streaming JSON payload",
				Description: "Returns a real-time JSON stream with configurable delays and ServiceNow-specific scenarios",
				Tags:        []string{"streaming"},
				Parameters: []OpenAPIParameter{
					{
						Name:        "count",
						In:          "query",
						Description: "Number of objects to stream (default: 100, max: 100000)",
						Required:    false,
						Schema: &OpenAPISchema{
							Type:    "integer",
							Minimum: &[]int{1}[0],
							Maximum: &[]int{100000}[0],
							Example: 100,
						},
					},
					{
						Name:        "delay",
						In:          "query",
						Description: "Base delay between items (e.g., '100ms', '1s', or just milliseconds)",
						Required:    false,
						Schema: &OpenAPISchema{
							Type:    "string",
							Example: "100ms",
						},
					},
					{
						Name:        "strategy",
						In:          "query",
						Description: "Delay strategy: 'fixed' = consistent delay, 'random' = random delay up to 2x base, 'progressive' = increasing delay over time, 'burst' = short delays with periodic long pauses",
						Required:    false,
						Schema: &OpenAPISchema{
							Type:    "string",
							Enum:    []interface{}{"fixed", "random", "progressive", "burst"},
							Example: "fixed",
						},
					},
					{
						Name:        "scenario",
						In:          "query",
						Description: "ServiceNow simulation scenario: 'peak_hours' = 200ms delays, 'maintenance' = 500ms with 2s spikes every 500 items, 'network_issues' = random spikes up to 3s (10% chance), 'database_load' = progressively increasing delays",
						Required:    false,
						Schema: &OpenAPISchema{
							Type:    "string",
							Enum:    []interface{}{"peak_hours", "maintenance", "network_issues", "database_load"},
							Example: "peak_hours",
						},
					},
					{
						Name:        "batch_size",
						In:          "query",
						Description: "Number of items to send before flushing (default: 10)",
						Required:    false,
						Schema: &OpenAPISchema{
							Type:    "integer",
							Minimum: &[]int{1}[0],
							Example: 10,
						},
					},
					{
						Name:        "servicenow",
						In:          "query",
						Description: "Enable ServiceNow-style record format",
						Required:    false,
						Schema: &OpenAPISchema{
							Type:    "boolean",
							Example: false,
						},
					},
				},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "Successful streaming response with JSON array",
						Content: map[string]OpenAPIMediaType{
							"application/json": {
								Schema: &OpenAPISchema{
									Type: "array",
									Items: &OpenAPISchema{
										Type: "object",
										Properties: map[string]*OpenAPISchema{
											"id": {
												Type:        "integer",
												Description: "Unique identifier for the item",
												Example:     1,
											},
											"value": {
												Type:        "string",
												Description: "Value or description of the item",
												Example:     "streamed data 1",
											},
											"timestamp": {
												Type:        "string",
												Format:      "date-time",
												Description: "Timestamp when the item was generated",
											},
											"sys_id": {
												Type:        "string",
												Description: "ServiceNow system ID (when ServiceNow mode is enabled)",
												Example:     "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
											},
											"number": {
												Type:        "string",
												Description: "ServiceNow ticket number (when ServiceNow mode is enabled)",
												Example:     "INC0000001",
											},
											"state": {
												Type:        "string",
												Description: "ServiceNow state (when ServiceNow mode is enabled)",
												Example:     "New",
											},
										},
										Required: []string{"id", "value", "timestamp"},
									},
								},
								Example: []StreamItem{
									{
										ID:        1,
										Value:     "streamed data 1",
										Timestamp: time.Now(),
									},
									{
										ID:        2,
										Value:     "ServiceNow Record 2",
										Timestamp: time.Now(),
										SysID:     "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
										Number:    "INC0000002",
										State:     "In Progress",
									},
								},
							},
						},
					},
					"500": {
						Description: "Internal server error",
						Content: map[string]OpenAPIMediaType{
							"text/plain": {
								Schema: &OpenAPISchema{
									Type:    "string",
									Example: "JSON encoding failed",
								},
							},
						},
					},
				},
			},
		},
		Schemas: map[string]*OpenAPISchema{
			"StreamItem": {
				Type: "object",
				Properties: map[string]*OpenAPISchema{
					"id": {
						Type:        "integer",
						Description: "Unique identifier for the item",
					},
					"value": {
						Type:        "string",
						Description: "Value or description of the item",
					},
					"timestamp": {
						Type:        "string",
						Format:      "date-time",
						Description: "Timestamp when the item was generated",
					},
					"sys_id": {
						Type:        "string",
						Description: "ServiceNow system ID (optional)",
					},
					"number": {
						Type:        "string",
						Description: "ServiceNow ticket number (optional)",
					},
					"state": {
						Type:        "string",
						Description: "ServiceNow state (optional)",
					},
				},
				Required: []string{"id", "value", "timestamp"},
			},
		},
	}
}
