package main

import (
	"encoding/json"
	"net/http"
)

// Item represents a single object in the JSON payload returned by the /payload endpoint.
type Item struct {
	ID   int    `json:"id"`   // Unique identifier for the item
	Name string `json:"name"` // Name of the item (static "Object" in this example)
}

// PayloadHandler handles HTTP GET requests to the /payload endpoint.
//
// It generates a slice of 100,000 Item objects and returns them as a JSON array.
// This endpoint is primarily used for testing REST client implementations and
// observing behavior when consuming very large JSON responses.
func PayloadHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header so clients interpret the response as JSON.
	w.Header().Set("Content-Type", "application/json")

	// Preallocate a slice of Item with 100,000 elements.
	data := make([]Item, 100000)

	// Populate each Item in the slice with an ID and a static name.
	for i := 0; i < 100000; i++ {
		data[i] = Item{
			ID:   i,
			Name: "Object",
		}
	}

	// Encode the slice as JSON and write it to the response writer.
	// If encoding fails, an HTTP 500 error is sent.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// It's a good idea to log errors in a real application.
		http.Error(w, "Failed to encode payload", http.StatusInternalServerError)
	}
}
