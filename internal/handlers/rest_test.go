package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/xeipuuv/gojsonschema"
)

func TestRestPayloadHandler_JSONSchema(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/rest_payload", nil)
	w := httptest.NewRecorder()
	RestPayloadHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	schema := `{
		"type": "array",
		"items": {
			"type": "object",
			"properties": {
				"id":   {"type": "integer"},
				"name": {"type": "string"}
			},
			"required": ["id", "name"]
		}
	}`

	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(schema),
		gojsonschema.NewBytesLoader(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Schema validation failed: %v", err)
	}
	for _, e := range result.Errors() {
		t.Errorf("Schema error: %s", e)
	}
}

func TestRestPayloadHandler_ResponseLength(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/rest_payload", nil)
	w := httptest.NewRecorder()
	RestPayloadHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	var payload []Item
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}
	if len(payload) != 10000 {
		t.Errorf("Expected payload length 10000, got %d", len(payload))
	}
}

func TestRestPayloadHandler_PayloadContent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/rest_payload", nil)
	w := httptest.NewRecorder()
	RestPayloadHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	var payload []Item
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}
	if payload[0].ID != 1 {
		t.Errorf("First item ID: expected 1, got %d", payload[0].ID)
	}
}

func TestRestPayloadHandler_CountParameter(t *testing.T) {
	req := httptest.NewRequest("GET", "/rest_payload?count=5", nil)
	w := httptest.NewRecorder()
	RestPayloadHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	var items []Item
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
}

func TestRestPayloadHandler_AuthenticationRequired(t *testing.T) {
	authCfg := auth.Setup(true, "testuser", "testpass")

	// Without credentials
	req := httptest.NewRequest("GET", "/rest_payload", nil)
	w := httptest.NewRecorder()
	authCfg.Middleware(RestPayloadHandler)(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", w.Code)
	}

	// Wrong credentials
	req = httptest.NewRequest("GET", "/rest_payload", nil)
	req.SetBasicAuth("wrong", "credentials")
	w = httptest.NewRecorder()
	authCfg.Middleware(RestPayloadHandler)(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 with wrong auth, got %d", w.Code)
	}

	// Correct credentials
	req = httptest.NewRequest("GET", "/rest_payload", nil)
	req.SetBasicAuth("testuser", "testpass")
	w = httptest.NewRecorder()
	authCfg.Middleware(RestPayloadHandler)(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 with correct auth, got %d", w.Code)
	}
}

func TestRestPayloadPlugin_OpenAPISpec(t *testing.T) {
	plugin := RestPayloadPlugin{}
	spec := plugin.OpenAPISpec()

	if spec.Path != "/rest_payload" {
		t.Errorf("Wrong path: got %v want /rest_payload", spec.Path)
	}
	if spec.Operation.Get == nil {
		t.Fatal("Missing GET operation")
	}
	if len(spec.Operation.Get.Parameters) == 0 {
		t.Error("Missing parameters")
	}

	var countParam *interface{}
	for _, param := range spec.Operation.Get.Parameters {
		if param.Name == "count" {
			p := interface{}(param)
			countParam = &p
			if param.In != "query" {
				t.Errorf("Wrong parameter location: got %v want query", param.In)
			}
			if param.Required {
				t.Error("Count parameter should not be required")
			}
			break
		}
	}
	if countParam == nil {
		t.Fatal("Missing count parameter")
	}

	if _, exists := spec.Operation.Get.Responses["200"]; !exists {
		t.Error("Missing 200 response")
	}
	if _, exists := spec.Operation.Get.Responses["500"]; !exists {
		t.Error("Missing 500 response")
	}
	if _, exists := spec.Schemas["Item"]; !exists {
		t.Error("Missing Item schema")
	}
}
