package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// StreamItem represents a single object in the streamed JSON payload returned by the /stream_payload endpoint.
type StreamItem struct {
	ID    int    `json:"id"`    // Unique identifier for the item
	Value string `json:"value"` // Value of the item (dynamic in this example)
}

// StreamingPayloadHandler streams large JSON data in chunks for testing REST clients.
//
// It generates and streams 10,000 StreamItem objects as a JSON array, writing each item
// directly to the response in a chunked fashion. This endpoint is useful for testing
// clients that need to process large datasets without loading the entire response into memory.
func StreamingPayloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")

	w.Write([]byte("[\n"))
	encoder := json.NewEncoder(w)
	for i := 0; i < 10000; i++ {
		item := StreamItem{
			ID:    i,
			Value: fmt.Sprintf("streamed data %d", i),
		}
		if err := encoder.Encode(item); err != nil {
			// Optionally log error in real applications
			break
		}
		if i < 9999 {
			w.Write([]byte(",\n"))
		}
	}
	w.Write([]byte("]"))
}
