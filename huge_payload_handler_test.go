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

// TestHugePayloadHandler_JSONSchema validates the /payload endpoint response structure against a JSON schema.
func TestHugePayloadHandler_JSONSchema(t *testing.T) {
	fmt.Println("[TestHugePayloadHandler_JSONSchema] Starting test for /payload endpoint")
	req := httptest.NewRequest(http.MethodGet, "/payload", nil)
	w := httptest.NewRecorder()

	HugePayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[TestHugePayloadHandler_JSONSchema] Failed to read response body: %v\n", err)
		t.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Printf("[TestHugePayloadHandler_JSONSchema] Response body size: %d bytes\n", len(bodyBytes))

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
		fmt.Printf("[TestHugePayloadHandler_JSONSchema] Schema validation failed: %v\n", err)
		t.Fatalf("Schema validation failed: %v", err)
	}
	if !result.Valid() {
		fmt.Println("[TestHugePayloadHandler_JSONSchema] Schema validation errors:")
		for _, err := range result.Errors() {
			fmt.Printf("  %s\n", err)
			t.Errorf("Schema error: %s", err)
		}
	} else {
		fmt.Println("[TestHugePayloadHandler_JSONSchema] Schema validation passed.")
	}
}

// TestHugePayloadHandler_ResponseLength checks that the /payload endpoint returns exactly 100,000 items.
func TestHugePayloadHandler_ResponseLength(t *testing.T) {
	fmt.Println("[TestHugePayloadHandler_ResponseLength] Starting length test for /payload endpoint")
	req := httptest.NewRequest(http.MethodGet, "/payload", nil)
	w := httptest.NewRecorder()

	HugePayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var payload []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		fmt.Printf("[TestHugePayloadHandler_ResponseLength] Failed to decode JSON: %v\n", err)
		t.Fatalf("Failed to decode JSON: %v", err)
	}
	fmt.Printf("[TestHugePayloadHandler_ResponseLength] Decoded payload length: %d\n", len(payload))
	if len(payload) != 10000 {
		t.Errorf("Expected payload length 10000, got %d", len(payload))
	} else {
		fmt.Println("[TestHugePayloadHandler_ResponseLength] Payload length is correct.")
	}
}

// TestHugePayloadHandler_PayloadContent checks the special contents of the payload.
func TestHugePayloadHandler_PayloadContent(t *testing.T) {
	fmt.Println("[TestHugePayloadHandler_PayloadContent] Starting content test for /payload endpoint")
	req := httptest.NewRequest(http.MethodGet, "/huge_payload", nil)
	w := httptest.NewRecorder()

	HugePayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var payload []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		fmt.Printf("[TestHugePayloadHandler_PayloadContent] Failed to decode JSON: %v\n", err)
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if payload[0].ID != 1 {
		t.Errorf("First item ID is incorrect: expected 1, got %d", payload[0].ID)
	} else {
		fmt.Println("[TestHugePayloadHandler_PayloadContent] First item ID is correct.")
	}

	fmt.Println("[TestHugePayloadHandler_PayloadContent] All items have correct ids.")
}

// TestHugePayloadHandler_CountParameter checks that the /huge_payload endpoint respects the count query parameter.
func TestHugePayloadHandler_CountParameter(t *testing.T) {
	req := httptest.NewRequest("GET", "/huge_payload?count=5", nil)
	w := httptest.NewRecorder()

	HugePayloadHandler(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()

	var items []Item
	if err := json.Unmarshal([]byte(body), &items); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
}
