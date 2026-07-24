package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dtrabandt/payloadBuddy/internal/openapi"
)

const (
	defaultRestCount = 10000
	maxRestCount     = 1000000
)

// Item represents a single object in the REST payload response.
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RestPayloadPlugin implements PayloadPlugin for the /rest_payload endpoint.
type RestPayloadPlugin struct{}

func (RestPayloadPlugin) Path() string              { return "/rest_payload" }
func (RestPayloadPlugin) Handler() http.HandlerFunc { return RestPayloadHandler }

// RestPayloadHandler handles GET /rest_payload — returns a large JSON array.
func RestPayloadHandler(w http.ResponseWriter, r *http.Request) {
	q := newQueryParser(r)
	count := q.Int("count", defaultRestCount)
	if err := q.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if count <= 0 || count > maxRestCount {
		http.Error(w, fmt.Sprintf("Count must be between 1 and %d", maxRestCount), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data := make([]Item, count)
	for i := 1; i <= count; i++ {
		data[i-1] = Item{ID: i, Name: "Object " + strconv.Itoa(i)}
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode payload", http.StatusInternalServerError)
	}
}

func (RestPayloadPlugin) OpenAPISpec() openapi.PathSpec {
	return openapi.PathSpec{
		Path: "/rest_payload",
		Operation: openapi.OpenAPIPath{
			Get: &openapi.OpenAPIOperation{
				Summary:     "Get large JSON payload",
				Description: "Returns a configurable number of JSON objects for testing REST client implementations",
				Tags:        []string{"payload"},
				Parameters: []openapi.OpenAPIParameter{
					{
						Name:        "count",
						In:          "query",
						Description: "Number of objects to return (default: 10000, max: 1000000)",
						Required:    false,
						Schema: &openapi.OpenAPISchema{
							Type:    "integer",
							Minimum: &[]int{1}[0],
							Maximum: &[]int{1000000}[0],
							Example: 10000,
						},
					},
				},
				Responses: map[string]openapi.OpenAPIResponse{
					"200": {
						Description: "Successful response with JSON array",
						Content: map[string]openapi.OpenAPIMediaType{
							"application/json": {
								Schema: &openapi.OpenAPISchema{
									Type: "array",
									Items: &openapi.OpenAPISchema{
										Type: "object",
										Properties: map[string]*openapi.OpenAPISchema{
											"id":   {Type: "integer", Description: "Unique identifier", Example: 1},
											"name": {Type: "string", Description: "Name of the item", Example: "Object 1"},
										},
										Required: []string{"id", "name"},
									},
								},
								Example: []Item{{ID: 1, Name: "Object 1"}, {ID: 2, Name: "Object 2"}},
							},
						},
					},
					"400": {
						Description: "Bad request — invalid count parameter",
						Content: map[string]openapi.OpenAPIMediaType{
							"text/plain": {Schema: &openapi.OpenAPISchema{Type: "string", Example: "Count must be between 1 and 1000000"}},
						},
					},
					"500": {
						Description: "Internal server error",
						Content: map[string]openapi.OpenAPIMediaType{
							"text/plain": {Schema: &openapi.OpenAPISchema{Type: "string", Example: "Failed to encode payload"}},
						},
					},
				},
			},
		},
		Schemas: map[string]*openapi.OpenAPISchema{
			"Item": {
				Type: "object",
				Properties: map[string]*openapi.OpenAPISchema{
					"id":   {Type: "integer", Description: "Unique identifier"},
					"name": {Type: "string", Description: "Name of the item"},
				},
				Required: []string{"id", "name"},
			},
		},
	}
}
