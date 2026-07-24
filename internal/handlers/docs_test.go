package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/dtrabandt/payloadBuddy/internal/openapi"
	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

// buildOpenAPIHandler wires up all plugins and returns an OpenAPI handler for testing.
func buildOpenAPIHandler(authEnabled bool) http.HandlerFunc {
	sm := scenarios.NewManager()
	var authCfg *auth.Config
	if authEnabled {
		authCfg = auth.Setup(true, "u", "p")
	} else {
		authCfg = auth.Setup(false, "", "")
	}
	docPlugin := &DocumentationPlugin{}
	allPlugins := []PayloadPlugin{
		RestPayloadPlugin{},
		StreamingPayloadPlugin{SM: sm},
		PaginatedPayloadPlugin{SM: sm},
		docPlugin,
		SwaggerUIPlugin{},
	}
	docPlugin.Init(allPlugins, authCfg)
	return NewOpenAPIHandler(allPlugins, authCfg)
}

func TestOpenAPIHandler_JSONResponse(t *testing.T) {
	handler := buildOpenAPIHandler(false)

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}
	if cors := rr.Header().Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Expected CORS header *, got %s", cors)
	}

	var spec openapi.OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("Wrong OpenAPI version: got %v want 3.1.0", spec.OpenAPI)
	}
	if spec.Info.Title != "PayloadBuddy API" {
		t.Errorf("Wrong API title: got %v", spec.Info.Title)
	}
	if spec.Info.Version != "1.0.0" {
		t.Errorf("Wrong API version: got %v", spec.Info.Version)
	}

	for _, path := range []string{"/rest_payload", "/stream_payload", "/openapi.json", "/swagger"} {
		if _, exists := spec.Paths[path]; !exists {
			t.Errorf("Missing path in OpenAPI spec: %s", path)
		}
	}
}

func TestOpenAPIHandler_PathsAndSchemas(t *testing.T) {
	handler := buildOpenAPIHandler(false)

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var spec openapi.OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

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

	if spec.Components == nil || spec.Components.Schemas == nil {
		t.Fatal("Missing components or schemas")
	}
	for _, schemaName := range []string{"Item", "StreamItem"} {
		if _, ok := spec.Components.Schemas[schemaName]; !ok {
			t.Errorf("Missing schema: %s", schemaName)
		}
	}

	itemSchema, exists := spec.Components.Schemas["Item"]
	if !exists {
		t.Fatal("Missing Item schema")
	}
	if itemSchema.Type != "object" {
		t.Errorf("Wrong Item schema type: got %v want object", itemSchema.Type)
	}
	for _, prop := range []string{"id", "name"} {
		if _, exists := itemSchema.Properties[prop]; !exists {
			t.Errorf("Missing property %s in Item schema", prop)
		}
	}
}

func TestOpenAPIHandler_WithAuthentication(t *testing.T) {
	handler := buildOpenAPIHandler(true)

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var spec openapi.OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

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

func TestOpenAPIHandler_SecuritySchemeWhenAuthEnabled(t *testing.T) {
	handler := buildOpenAPIHandler(true)

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var spec openapi.OpenAPISpec
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

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

	for _, endpoint := range []string{"/rest_payload", "/stream_payload"} {
		path, exists := spec.Paths[endpoint]
		if !exists {
			t.Errorf("Missing endpoint: %s", endpoint)
			continue
		}
		if path.Get == nil || len(path.Get.Security) == 0 {
			t.Errorf("Missing security requirements for %s", endpoint)
			continue
		}
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
}

func TestSwaggerUIHandler_HTMLResponse(t *testing.T) {
	req := httptest.NewRequest("GET", "/swagger", nil)
	rr := httptest.NewRecorder()
	SwaggerUIHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html; charset=utf-8, got %s", ct)
	}
	body := rr.Body.String()
	for _, element := range []string{
		"<!DOCTYPE html>",
		"PayloadBuddy API Documentation",
		"swagger-ui-dist",
		"SwaggerUIBundle",
		"url: '/openapi.json'",
		"dom_id: '#swagger-ui'",
	} {
		if !strings.Contains(body, element) {
			t.Errorf("Missing required element in Swagger UI HTML: %s", element)
		}
	}
}

func TestDocumentationPlugin_Interface(t *testing.T) {
	sm := scenarios.NewManager()
	authCfg := auth.Setup(false, "", "")
	docPlugin := &DocumentationPlugin{}
	allPlugins := []PayloadPlugin{docPlugin}
	docPlugin.Init(allPlugins, authCfg)
	_ = sm

	if path := docPlugin.Path(); path != "/openapi.json" {
		t.Errorf("Wrong path: got %v want /openapi.json", path)
	}
	if handler := docPlugin.Handler(); handler == nil {
		t.Error("Handler should not be nil")
	}
	spec := docPlugin.OpenAPISpec()
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

	if path := plugin.Path(); path != "/swagger" {
		t.Errorf("Wrong path: got %v want /swagger", path)
	}
	if handler := plugin.Handler(); handler == nil {
		t.Error("Handler should not be nil")
	}
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
