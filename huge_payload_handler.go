package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Item represents a single object in the JSON payload returned by the /payload endpoint.
type Item struct {
	ID   int    `json:"id"`   // Unique identifier for the item
	Name string `json:"name"` // Name of the item (static "Object" in this example)
}

// HugePayloadHandler handles HTTP GET requests to the /payload endpoint.
//
// It generates a slice of 10000 Item objects and returns them as a JSON array.
// This endpoint is primarily used for testing REST client implementations and
// observing behavior when consuming very large JSON responses.
func HugePayloadHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header so clients interpret the response as JSON.
	w.Header().Set("Content-Type", "application/json")

	// Parse count parameter, default to 10000
	count := 10000
	if val := r.URL.Query().Get("count"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 && parsed <= 1000000 {
			count = parsed
		}
	}

	// Preallocate a slice of Item with 'count' elements.
	data := make([]Item, count)

	// Populate each Item in the slice with an ID and a static name.
	for i := 1; i <= count; i++ {
		data[i-1] = Item{
			ID:   i,
			Name: "Object " + strconv.Itoa(i),
		}
	}

	// Encode the slice as JSON and write it to the response writer.
	// If encoding fails, an HTTP 500 error is sent.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode payload", http.StatusInternalServerError)
	}
}
