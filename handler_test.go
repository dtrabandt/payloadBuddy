// handler_test.go

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

// TestPayloadHandler_JSONSchema validates the /payload endpoint response structure against a JSON schema.
func TestPayloadHandler_JSONSchema(t *testing.T) {
	fmt.Println("[TestPayloadHandler_JSONSchema] Starting test for /payload endpoint")
	req := httptest.NewRequest(http.MethodGet, "/payload", nil)
	w := httptest.NewRecorder()

	PayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[TestPayloadHandler_JSONSchema] Failed to read response body: %v\n", err)
		t.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Printf("[TestPayloadHandler_JSONSchema] Response body size: %d bytes\n", len(bodyBytes))

	schema := `{
        "type": "array",
        "items": {
            "type": "object",
            "properties": {
                "id": {"type": "integer"},
                "name": {"type": "string"}
            },
            "required": ["id", "name"]
        }
    }`

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(bodyBytes)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		fmt.Printf("[TestPayloadHandler_JSONSchema] Schema validation failed: %v\n", err)
		t.Fatalf("Schema validation failed: %v", err)
	}
	if !result.Valid() {
		fmt.Println("[TestPayloadHandler_JSONSchema] Schema validation errors:")
		for _, err := range result.Errors() {
			fmt.Printf("  %s\n", err)
			t.Errorf("Schema error: %s", err)
		}
	} else {
		fmt.Println("[TestPayloadHandler_JSONSchema] Schema validation passed.")
	}
}

// TestPayloadHandler_ResponseLength checks that the /payload endpoint returns exactly 100,000 items.
func TestPayloadHandler_ResponseLength(t *testing.T) {
	fmt.Println("[TestPayloadHandler_ResponseLength] Starting length test for /payload endpoint")
	req := httptest.NewRequest(http.MethodGet, "/payload", nil)
	w := httptest.NewRecorder()

	PayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var payload []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		fmt.Printf("[TestPayloadHandler_ResponseLength] Failed to decode JSON: %v\n", err)
		t.Fatalf("Failed to decode JSON: %v", err)
	}
	fmt.Printf("[TestPayloadHandler_ResponseLength] Decoded payload length: %d\n", len(payload))
	if len(payload) != 100000 {
		t.Errorf("Expected payload length 100000, got %d", len(payload))
	} else {
		fmt.Println("[TestPayloadHandler_ResponseLength] Payload length is correct.")
	}
}
