package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/openapi"
	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

// StreamItem represents a single object in the streamed JSON payload.
type StreamItem struct {
	ID        int       `json:"id"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	SysID     string    `json:"sys_id,omitempty"`
	Number    string    `json:"number,omitempty"`
	State     string    `json:"state,omitempty"`
}

// StreamingPayloadPlugin implements PayloadPlugin for the /stream_payload endpoint.
type StreamingPayloadPlugin struct {
	SM *scenarios.Manager
}

func (StreamingPayloadPlugin) Path() string { return "/stream_payload" }

func (s StreamingPayloadPlugin) Handler() http.HandlerFunc {
	return NewStreamingHandler(s.SM)
}

// NewStreamingHandler returns an http.HandlerFunc backed by the given scenario manager.
func NewStreamingHandler(sm *scenarios.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		scenario := strings.ToLower(r.URL.Query().Get("scenario"))

		var defaultCount, maxCount, defaultBatchSize int
		var defaultSNMode bool
		if sm != nil && scenario != "" {
			defaultBatchSize, defaultSNMode, maxCount, defaultCount = sm.GetScenarioConfig(scenario)
		} else {
			defaultCount = 10000
			maxCount = 1000000
			defaultBatchSize = 100
		}

		count := getIntParam(r, "count", defaultCount)
		baseDelay := getDurationParam(r, "delay", 10*time.Millisecond)
		strategy := getDelayStrategy(r)
		batchSize := getIntParam(r, "batch_size", defaultBatchSize)

		snMode := defaultSNMode
		if v := r.URL.Query().Get("servicenow"); v != "" {
			snMode = v == "true"
		}

		if count <= 0 || count > maxCount {
			http.Error(w, fmt.Sprintf("Count must be between 1 and %d", maxCount), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		if _, err := w.Write([]byte("[\n")); err != nil {
			return
		}
		flusher.Flush()

		for i := range count {
			select {
			case <-ctx.Done():
				_, _ = w.Write([]byte("\n]")) // best-effort close; client may already be gone
				return
			default:
			}

			var item StreamItem
			if snMode {
				item = StreamItem{
					ID:        i,
					Value:     fmt.Sprintf("ServiceNow Record %d", i),
					Timestamp: time.Now(),
					SysID:     generateSysID(),
					Number:    fmt.Sprintf("INC%07d", i),
					State:     []string{"New", "In Progress", "Resolved", "Closed"}[i%4],
				}
			} else {
				item = StreamItem{ID: i, Value: fmt.Sprintf("streamed data %d", i), Timestamp: time.Now()}
			}

			data, err := json.Marshal(item)
			if err != nil {
				http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
				return
			}

			if i > 0 {
				if _, err := w.Write([]byte(",\n")); err != nil {
					return
				}
			}
			if _, err := w.Write(data); err != nil {
				return
			}

			if err := applyDelay(ctx, strategy, baseDelay, scenario, i, sm); err != nil {
				_, _ = w.Write([]byte("\n]")) // best-effort close; client may already be gone
				return
			}

			if i%batchSize == 0 {
				flusher.Flush()
			}
		}

		_, _ = w.Write([]byte("\n]")) // best-effort close; connection may drop on large payloads
		flusher.Flush()
	}
}

// applyDelay sleeps for the appropriate duration based on strategy, scenario, and item index.
// Returns a non-nil error only when the context is canceled.
func applyDelay(ctx context.Context, strategy scenarios.DelayStrategy, baseDelay time.Duration, scenario string, itemIndex int, sm *scenarios.Manager) error {
	var delay time.Duration

	if sm != nil && scenario != "" {
		// The scenario manager already resolves the effective delay; the strategy
		// classification it returns is only consulted in the no-scenario branch below,
		// so it is intentionally discarded here.
		calculatedDelay, _ := sm.GetScenarioDelay(scenario, itemIndex)
		if scenario == "network_issues" {
			randFloat, err := secureRandFloat32()
			if err != nil || randFloat >= 0.1 {
				delay = calculatedDelay
			} else {
				randInt, err := secureRandIntn(3000)
				if err != nil {
					delay = calculatedDelay
				} else {
					delay = time.Duration(randInt) * time.Millisecond
				}
			}
		} else {
			delay = calculatedDelay
		}
	} else {
		switch scenario {
		case "peak_hours":
			delay = 200 * time.Millisecond
		case "maintenance":
			if itemIndex%500 == 0 {
				delay = 2 * time.Second
			} else {
				delay = 500 * time.Millisecond
			}
		case "network_issues":
			randFloat, err := secureRandFloat32()
			if err != nil || randFloat >= 0.1 {
				delay = baseDelay
			} else {
				randInt, err := secureRandIntn(3000)
				if err != nil {
					delay = baseDelay
				} else {
					delay = time.Duration(randInt) * time.Millisecond
				}
			}
		case "database_load":
			delay = baseDelay + time.Duration(itemIndex/100)*10*time.Millisecond
		default:
			switch strategy {
			case scenarios.NoDelay:
				return nil
			case scenarios.FixedDelay:
				delay = baseDelay
			case scenarios.RandomDelay:
				randInt64, err := secureRandInt63n(int64(baseDelay * 2))
				if err != nil {
					delay = baseDelay
				} else {
					delay = time.Duration(randInt64)
				}
			case scenarios.ProgressiveDelay:
				delay = baseDelay * time.Duration(itemIndex/1000+1)
			case scenarios.BurstDelay:
				if itemIndex%100 == 0 && itemIndex > 0 {
					delay = baseDelay * 10
				} else {
					delay = baseDelay / 10
				}
			}
		}
	}

	if delay <= 0 {
		return nil
	}
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (StreamingPayloadPlugin) OpenAPISpec() openapi.PathSpec {
	return openapi.PathSpec{
		Path: "/stream_payload",
		Operation: openapi.OpenAPIPath{
			Get: &openapi.OpenAPIOperation{
				Summary:     "Get streaming JSON payload",
				Description: "Returns a real-time JSON stream with configurable delays and ServiceNow-specific scenarios",
				Tags:        []string{"streaming"},
				Parameters: []openapi.OpenAPIParameter{
					{Name: "count", In: "query", Description: "Number of objects to stream (default: 100, max: 100000)", Schema: &openapi.OpenAPISchema{Type: "integer", Minimum: &[]int{1}[0], Maximum: &[]int{100000}[0], Example: 100}},
					{Name: "delay", In: "query", Description: "Base delay between items (e.g., '100ms', '1s', or milliseconds)", Schema: &openapi.OpenAPISchema{Type: "string", Example: "100ms"}},
					{Name: "strategy", In: "query", Description: "Delay strategy: fixed, random, progressive, burst", Schema: &openapi.OpenAPISchema{Type: "string", Enum: []interface{}{"fixed", "random", "progressive", "burst"}, Example: "fixed"}},
					{Name: "scenario", In: "query", Description: "ServiceNow simulation scenario", Schema: &openapi.OpenAPISchema{Type: "string", Enum: []any{"peak_hours", "maintenance", "network_issues", "database_load"}, Example: "peak_hours"}},
					{Name: "batch_size", In: "query", Description: "Items per flush batch (default: 10)", Schema: &openapi.OpenAPISchema{Type: "integer", Minimum: &[]int{1}[0], Example: 10}},
					{Name: "servicenow", In: "query", Description: "Enable ServiceNow-style record format", Schema: &openapi.OpenAPISchema{Type: "boolean", Example: false}},
				},
				Responses: map[string]openapi.OpenAPIResponse{
					"200": {Description: "Successful streaming response with JSON array"},
					"500": {Description: "Internal server error"},
				},
			},
		},
		Schemas: map[string]*openapi.OpenAPISchema{
			"StreamItem": {
				Type: "object",
				Properties: map[string]*openapi.OpenAPISchema{
					"id":        {Type: "integer", Description: "Unique identifier"},
					"value":     {Type: "string", Description: "Value or description"},
					"timestamp": {Type: "string", Format: "date-time", Description: "Generation timestamp"},
					"sys_id":    {Type: "string", Description: "ServiceNow system ID (optional)"},
					"number":    {Type: "string", Description: "ServiceNow ticket number (optional)"},
					"state":     {Type: "string", Description: "ServiceNow state (optional)"},
				},
				Required: []string{"id", "value", "timestamp"},
			},
		},
	}
}
