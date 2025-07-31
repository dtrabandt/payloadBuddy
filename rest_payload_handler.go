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

// RestPayloadHandler handles HTTP GET requests to the /payload endpoint.
//
// It generates a slice of 10000 Item objects and returns them as a JSON array.
// This endpoint is primarily used for testing REST client implementations and
// observing behavior when consuming very large JSON responses.
func RestPayloadHandler(w http.ResponseWriter, r *http.Request) {
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

// OpenAPISpec returns the OpenAPI specification for the rest payload endpoint
func (h RestPayloadPlugin) OpenAPISpec() OpenAPIPathSpec {
	return OpenAPIPathSpec{
		Path: "/rest_payload",
		Operation: OpenAPIPath{
			Get: &OpenAPIOperation{
				Summary:     "Get large JSON payload",
				Description: "Returns a configurable number of JSON objects for testing REST client implementations",
				Tags:        []string{"payload"},
				Parameters: []OpenAPIParameter{
					{
						Name:        "count",
						In:          "query",
						Description: "Number of objects to return (default: 10000, max: 1000000)",
						Required:    false,
						Schema: &OpenAPISchema{
							Type:    "integer",
							Minimum: &[]int{1}[0],
							Maximum: &[]int{1000000}[0],
							Example: 10000,
						},
					},
				},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "Successful response with JSON array",
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
											"name": {
												Type:        "string",
												Description: "Name of the item",
												Example:     "Object 1",
											},
										},
										Required: []string{"id", "name"},
									},
								},
								Example: []Item{
									{ID: 1, Name: "Object 1"},
									{ID: 2, Name: "Object 2"},
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
									Example: "Failed to encode payload",
								},
							},
						},
					},
				},
			},
		},
		Schemas: map[string]*OpenAPISchema{
			"Item": {
				Type: "object",
				Properties: map[string]*OpenAPISchema{
					"id": {
						Type:        "integer",
						Description: "Unique identifier for the item",
					},
					"name": {
						Type:        "string",
						Description: "Name of the item",
					},
				},
				Required: []string{"id", "name"},
			},
		},
	}
}
