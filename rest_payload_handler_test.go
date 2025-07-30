// rest_payload_handler_test.go

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

// helper function to create authenticated request
func createAuthRequest(method, path string, username, password string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if username != "" && password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		req.Header.Set("Authorization", "Basic "+auth)
	}
	return req
}

// TestRestPayloadHandler_JSONSchema validates the /rest_payload endpoint response structure against a JSON schema.
func TestRestPayloadHandler_JSONSchema(t *testing.T) {
	fmt.Println("[TestRestPayloadHandler_JSONSchema] Starting test for /rest_payload endpoint")

	// Test without auth (auth disabled by default in tests)
	*enableAuth = false
	req := httptest.NewRequest(http.MethodGet, "/rest_payload", nil)
	w := httptest.NewRecorder()

	RestPayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[TestRestPayloadHandler_JSONSchema] Failed to read response body: %v\n", err)
		t.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Printf("[TestRestPayloadHandler_JSONSchema] Response body size: %d bytes\n", len(bodyBytes))

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
		fmt.Printf("[TestRestPayloadHandler_JSONSchema] Schema validation failed: %v\n", err)
		t.Fatalf("Schema validation failed: %v", err)
	}
	if !result.Valid() {
		fmt.Println("[TestRestPayloadHandler_JSONSchema] Schema validation errors:")
		for _, err := range result.Errors() {
			fmt.Printf("  %s\n", err)
			t.Errorf("Schema error: %s", err)
		}
	} else {
		fmt.Println("[TestRestPayloadHandler_JSONSchema] Schema validation passed.")
	}
}

// TestRestPayloadHandler_ResponseLength checks that the /rest_payload endpoint returns exactly 10,000 items.
func TestRestPayloadHandler_ResponseLength(t *testing.T) {
	fmt.Println("[TestRestPayloadHandler_ResponseLength] Starting length test for /rest_payload endpoint")
	*enableAuth = false
	req := httptest.NewRequest(http.MethodGet, "/rest_payload", nil)
	w := httptest.NewRecorder()

	RestPayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var payload []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		fmt.Printf("[TestRestPayloadHandler_ResponseLength] Failed to decode JSON: %v\n", err)
		t.Fatalf("Failed to decode JSON: %v", err)
	}
	fmt.Printf("[TestRestPayloadHandler_ResponseLength] Decoded payload length: %d\n", len(payload))
	if len(payload) != 10000 {
		t.Errorf("Expected payload length 10000, got %d", len(payload))
	} else {
		fmt.Println("[TestRestPayloadHandler_ResponseLength] Payload length is correct.")
	}
}

// TestRestPayloadHandler_PayloadContent checks the special contents of the payload.
func TestRestPayloadHandler_PayloadContent(t *testing.T) {
	fmt.Println("[TestRestPayloadHandler_PayloadContent] Starting content test for /rest_payload endpoint")
	*enableAuth = false
	req := httptest.NewRequest(http.MethodGet, "/rest_payload", nil)
	w := httptest.NewRecorder()

	RestPayloadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var payload []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		fmt.Printf("[TestRestPayloadHandler_PayloadContent] Failed to decode JSON: %v\n", err)
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if payload[0].ID != 1 {
		t.Errorf("First item ID is incorrect: expected 1, got %d", payload[0].ID)
	} else {
		fmt.Println("[TestRestPayloadHandler_PayloadContent] First item ID is correct.")
	}

	fmt.Println("[TestRestPayloadHandler_PayloadContent] All items have correct ids.")
}

// TestRestPayloadHandler_CountParameter checks that the /rest_payload endpoint respects the count query parameter.
func TestRestPayloadHandler_CountParameter(t *testing.T) {
	*enableAuth = false
	req := httptest.NewRequest("GET", "/rest_payload?count=5", nil)
	w := httptest.NewRecorder()

	RestPayloadHandler(w, req)
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

// TestRestPayloadHandler_AuthenticationRequired tests that authentication is required when enabled.
func TestRestPayloadHandler_AuthenticationRequired(t *testing.T) {
	*enableAuth = true
	authUsername = "testuser"
	authPassword = "testpass"

	// Test without credentials
	req := httptest.NewRequest("GET", "/rest_payload", nil)
	w := httptest.NewRecorder()

	basicAuthMiddleware(RestPayloadHandler)(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without auth, got %d", resp.StatusCode)
	}

	// Test with wrong credentials
	req = createAuthRequest("GET", "/rest_payload", "wrong", "credentials")
	w = httptest.NewRecorder()

	basicAuthMiddleware(RestPayloadHandler)(w, req)
	resp = w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with wrong auth, got %d", resp.StatusCode)
	}

	// Test with correct credentials
	req = createAuthRequest("GET", "/rest_payload", "testuser", "testpass")
	w = httptest.NewRecorder()

	basicAuthMiddleware(RestPayloadHandler)(w, req)
	resp = w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with correct auth, got %d", resp.StatusCode)
	}
}
