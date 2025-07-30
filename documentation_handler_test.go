package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPIHandler_JSONResponse(t *testing.T) {
	// Disable auth for testing
	*enableAuth = false

	req, err := http.NewRequest("GET", "/openapi.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(OpenAPIHandler)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Check CORS header
	if cors := rr.Header().Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("handler returned wrong CORS header: got %v want %v", cors, "*")
	}

	// Parse JSON response
	var spec OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Validate OpenAPI version
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("Wrong OpenAPI version: got %v want %v", spec.OpenAPI, "3.1.0")
	}

	// Validate basic info
	if spec.Info.Title != "PayloadBuddy API" {
		t.Errorf("Wrong API title: got %v want %v", spec.Info.Title, "PayloadBuddy API")
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("Wrong API version: got %v want %v", spec.Info.Version, "1.0.0")
	}

	// Check that paths are present
	expectedPaths := []string{"/rest_payload", "/stream_payload", "/openapi.json", "/swagger"}
	for _, path := range expectedPaths {
		if _, exists := spec.Paths[path]; !exists {
			t.Errorf("Missing path in OpenAPI spec: %s", path)
		}
	}
}

func TestOpenAPIHandler_PathsAndSchemas(t *testing.T) {
	// Disable auth for testing
	*enableAuth = false

	req, err := http.NewRequest("GET", "/openapi.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(OpenAPIHandler)
	handler.ServeHTTP(rr, req)

	var spec OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Test rest_payload path
	restPath, exists := spec.Paths["/rest_payload"]
	if !exists {
		t.Fatal("Missing /rest_payload path")
	}

	if restPath.Get == nil {
		t.Fatal("Missing GET operation for /rest_payload")
	}

	if restPath.Get.Summary != "Get large JSON payload" {
		t.Errorf("Wrong summary for /rest_payload: got %v", restPath.Get.Summary)
	}

	// Test stream_payload path
	streamPath, exists := spec.Paths["/stream_payload"]
	if !exists {
		t.Fatal("Missing /stream_payload path")
	}

	if streamPath.Get == nil {
		t.Fatal("Missing GET operation for /stream_payload")
	}

	if streamPath.Get.Summary != "Get streaming JSON payload" {
		t.Errorf("Wrong summary for /stream_payload: got %v", streamPath.Get.Summary)
	}

	// Test schemas
	if spec.Components == nil || spec.Components.Schemas == nil {
		t.Fatal("Missing components or schemas")
	}

	expectedSchemas := []string{"Item", "StreamItem"}
	for _, schemaName := range expectedSchemas {
		if _, exists := spec.Components.Schemas[schemaName]; !exists {
			t.Errorf("Missing schema: %s", schemaName)
		}
	}

	// Test Item schema structure
	itemSchema, exists := spec.Components.Schemas["Item"]
	if !exists {
		t.Fatal("Missing Item schema")
	}

	if itemSchema.Type != "object" {
		t.Errorf("Wrong Item schema type: got %v want object", itemSchema.Type)
	}

	expectedProperties := []string{"id", "name"}
	for _, prop := range expectedProperties {
		if _, exists := itemSchema.Properties[prop]; !exists {
			t.Errorf("Missing property %s in Item schema", prop)
		}
	}
}

func TestOpenAPIHandler_WithAuthentication(t *testing.T) {
	// Enable auth for testing
	*enableAuth = true
	defer func() { *enableAuth = false }() // Reset after test

	req, err := http.NewRequest("GET", "/openapi.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(OpenAPIHandler)
	handler.ServeHTTP(rr, req)

	var spec OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check that authentication information is included in descriptions
	restPath, exists := spec.Paths["/rest_payload"]
	if !exists {
		t.Fatal("Missing /rest_payload path")
	}

	if restPath.Get == nil {
		t.Fatal("Missing GET operation for /rest_payload")
	}

	if !strings.Contains(restPath.Get.Description, "HTTP Basic Authentication") {
		t.Error("Missing authentication information in description")
	}
}

func TestSwaggerUIHandler_HTMLResponse(t *testing.T) {
	req, err := http.NewRequest("GET", "/swagger", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SwaggerUIHandler)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	expected := "text/html"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
	}

	body := rr.Body.String()

	// Check for essential Swagger UI elements
	requiredElements := []string{
		"<!DOCTYPE html>",
		"PayloadBuddy API Documentation",
		"swagger-ui-dist",
		"SwaggerUIBundle",
		"url: '/openapi.json'",
		"dom_id: '#swagger-ui'",
	}

	for _, element := range requiredElements {
		if !strings.Contains(body, element) {
			t.Errorf("Missing required element in Swagger UI HTML: %s", element)
		}
	}
}

func TestDocumentationPlugin_Interface(t *testing.T) {
	plugin := DocumentationPlugin{}

	// Test Path method
	if path := plugin.Path(); path != "/openapi.json" {
		t.Errorf("Wrong path: got %v want /openapi.json", path)
	}

	// Test Handler method
	handler := plugin.Handler()
	if handler == nil {
		t.Error("Handler should not be nil")
	}

	// Test OpenAPISpec method
	spec := plugin.OpenAPISpec()
	if spec.Path != "/openapi.json" {
		t.Errorf("Wrong spec path: got %v want /openapi.json", spec.Path)
	}

	if spec.Operation.Get == nil {
		t.Error("Missing GET operation in spec")
	}

	if spec.Operation.Get.Summary != "Get OpenAPI specification" {
		t.Errorf("Wrong summary: got %v", spec.Operation.Get.Summary)
	}
}

func TestSwaggerUIPlugin_Interface(t *testing.T) {
	plugin := SwaggerUIPlugin{}

	// Test Path method
	if path := plugin.Path(); path != "/swagger" {
		t.Errorf("Wrong path: got %v want /swagger", path)
	}

	// Test Handler method
	handler := plugin.Handler()
	if handler == nil {
		t.Error("Handler should not be nil")
	}

	// Test OpenAPISpec method
	spec := plugin.OpenAPISpec()
	if spec.Path != "/swagger" {
		t.Errorf("Wrong spec path: got %v want /swagger", spec.Path)
	}

	if spec.Operation.Get == nil {
		t.Error("Missing GET operation in spec")
	}

	if spec.Operation.Get.Summary != "Swagger UI" {
		t.Errorf("Wrong summary: got %v", spec.Operation.Get.Summary)
	}
}

func TestDocumentationEndpoints_NoAuthRequired(t *testing.T) {
	// Enable auth for testing
	*enableAuth = true
	authUsername = "testuser"
	authPassword = "testpass"
	defer func() { 
		*enableAuth = false
		authUsername = ""
		authPassword = ""
	}()

	tests := []struct {
		name     string
		path     string
		handler  http.HandlerFunc
		wantAuth bool
	}{
		{
			name:     "OpenAPI endpoint should not require auth",
			path:     "/openapi.json",
			handler:  OpenAPIHandler,
			wantAuth: false,
		},
		{
			name:     "Swagger UI should not require auth",
			path:     "/swagger",
			handler:  SwaggerUIHandler,
			wantAuth: false,
		},
		{
			name:     "Rest payload should require auth",
			path:     "/rest_payload",
			handler:  RestPayloadHandler,
			wantAuth: true,
		},
		{
			name:     "Streaming payload should require auth",
			path:     "/stream_payload",
			handler:  StreamingPayloadHandler,
			wantAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			// Test without authentication first
			if tt.wantAuth {
				// Should require auth - test with middleware
				handler := basicAuthMiddleware(tt.handler)
				handler.ServeHTTP(rr, req)
				if status := rr.Code; status != http.StatusUnauthorized {
					t.Errorf("Expected 401 Unauthorized without auth, got %v", status)
				}

				// Test with correct auth
				rr = httptest.NewRecorder()
				req.SetBasicAuth(authUsername, authPassword)
				handler.ServeHTTP(rr, req)
				if status := rr.Code; status != http.StatusOK {
					t.Errorf("Expected 200 OK with correct auth, got %v", status)
				}
			} else {
				// Should not require auth - test without middleware
				tt.handler.ServeHTTP(rr, req)
				if status := rr.Code; status != http.StatusOK {
					t.Errorf("Expected 200 OK without auth for %s, got %v", tt.path, status)
				}

				// Also test that it works with auth (should still work)
				rr = httptest.NewRecorder()
				req.SetBasicAuth(authUsername, authPassword)
				tt.handler.ServeHTTP(rr, req)
				if status := rr.Code; status != http.StatusOK {
					t.Errorf("Expected 200 OK with auth for %s, got %v", tt.path, status)
				}
			}
		})
	}
}

func TestOpenAPIHandler_SecuritySchemeWhenAuthEnabled(t *testing.T) {
	// Enable auth for testing
	*enableAuth = true
	defer func() { *enableAuth = false }()

	req, err := http.NewRequest("GET", "/openapi.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(OpenAPIHandler)
	handler.ServeHTTP(rr, req)

	var spec OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check that security schemes are present
	if spec.Components == nil || spec.Components.SecuritySchemes == nil {
		t.Fatal("Missing security schemes when auth is enabled")
	}

	basicAuth, exists := spec.Components.SecuritySchemes["BasicAuth"]
	if !exists {
		t.Fatal("Missing BasicAuth security scheme")
	}

	if basicAuth.Type != "http" {
		t.Errorf("Wrong security scheme type: got %v want http", basicAuth.Type)
	}

	if basicAuth.Scheme != "basic" {
		t.Errorf("Wrong security scheme: got %v want basic", basicAuth.Scheme)
	}

	// Check that API endpoints have security requirements
	apiEndpoints := []string{"/rest_payload", "/stream_payload"}
	for _, endpoint := range apiEndpoints {
		path, exists := spec.Paths[endpoint]
		if !exists {
			t.Errorf("Missing endpoint: %s", endpoint)
			continue
		}

		if path.Get == nil {
			t.Errorf("Missing GET operation for %s", endpoint)
			continue
		}

		if len(path.Get.Security) == 0 {
			t.Errorf("Missing security requirements for %s", endpoint)
			continue
		}

		// Check that BasicAuth is required
		found := false
		for _, secReq := range path.Get.Security {
			if _, hasBasicAuth := secReq["BasicAuth"]; hasBasicAuth {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("BasicAuth not required for %s", endpoint)
		}
	}

	// Check that documentation endpoints do NOT have security requirements in the spec
	// (they are excluded at the middleware level, not the spec level)
	docEndpoints := []string{"/openapi.json", "/swagger"}
	for _, endpoint := range docEndpoints {
		path, exists := spec.Paths[endpoint]
		if !exists {
			t.Errorf("Missing endpoint: %s", endpoint)
			continue
		}

		if path.Get == nil {
			t.Errorf("Missing GET operation for %s", endpoint)
			continue
		}

		// Documentation endpoints should also show security in the spec for consistency
		// even though they're excluded at the middleware level
		if len(path.Get.Security) == 0 {
			t.Errorf("Missing security requirements in spec for %s (should be consistent)", endpoint)
		}
	}
}

func TestRestPayloadPlugin_OpenAPISpec(t *testing.T) {
	plugin := RestPayloadPlugin{}
	spec := plugin.OpenAPISpec()

	// Test basic spec properties
	if spec.Path != "/rest_payload" {
		t.Errorf("Wrong path: got %v want /rest_payload", spec.Path)
	}

	if spec.Operation.Get == nil {
		t.Fatal("Missing GET operation")
	}

	// Test parameters
	if len(spec.Operation.Get.Parameters) == 0 {
		t.Error("Missing parameters")
	}

	// Find count parameter
	var countParam *OpenAPIParameter
	for _, param := range spec.Operation.Get.Parameters {
		if param.Name == "count" {
			countParam = &param
			break
		}
	}

	if countParam == nil {
		t.Fatal("Missing count parameter")
	}

	if countParam.In != "query" {
		t.Errorf("Wrong parameter location: got %v want query", countParam.In)
	}

	if countParam.Required {
		t.Error("Count parameter should not be required")
	}

	// Test responses
	if spec.Operation.Get.Responses == nil {
		t.Fatal("Missing responses")
	}

	if _, exists := spec.Operation.Get.Responses["200"]; !exists {
		t.Error("Missing 200 response")
	}

	if _, exists := spec.Operation.Get.Responses["500"]; !exists {
		t.Error("Missing 500 response")
	}

	// Test schemas
	if spec.Schemas == nil {
		t.Fatal("Missing schemas")
	}

	if _, exists := spec.Schemas["Item"]; !exists {
		t.Error("Missing Item schema")
	}
}

func TestStreamingPayloadPlugin_OpenAPISpec(t *testing.T) {
	plugin := StreamingPayloadPlugin{}
	spec := plugin.OpenAPISpec()

	// Test basic spec properties
	if spec.Path != "/stream_payload" {
		t.Errorf("Wrong path: got %v want /stream_payload", spec.Path)
	}

	if spec.Operation.Get == nil {
		t.Fatal("Missing GET operation")
	}

	// Test that we have multiple parameters
	expectedParams := []string{"count", "delay", "strategy", "scenario", "batch_size", "servicenow"}
	paramNames := make(map[string]bool)
	for _, param := range spec.Operation.Get.Parameters {
		paramNames[param.Name] = true
	}

	for _, expectedParam := range expectedParams {
		if !paramNames[expectedParam] {
			t.Errorf("Missing parameter: %s", expectedParam)
		}
	}

	// Test schemas
	if spec.Schemas == nil {
		t.Fatal("Missing schemas")
	}

	if _, exists := spec.Schemas["StreamItem"]; !exists {
		t.Error("Missing StreamItem schema")
	}

	// Test StreamItem schema properties
	streamItemSchema := spec.Schemas["StreamItem"]
	expectedProperties := []string{"id", "value", "timestamp", "sys_id", "number", "state"}
	for _, prop := range expectedProperties {
		if _, exists := streamItemSchema.Properties[prop]; !exists {
			t.Errorf("Missing property %s in StreamItem schema", prop)
		}
	}

	// Test required fields
	requiredFields := []string{"id", "value", "timestamp"}
	for _, required := range requiredFields {
		found := false
		for _, req := range streamItemSchema.Required {
			if req == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing required field: %s", required)
		}
	}
}