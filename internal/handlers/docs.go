package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/dtrabandt/payloadBuddy/internal/openapi"
)

// DocumentationPlugin implements PayloadPlugin for the /openapi.json endpoint.
// Call Init before registering to wire in the full plugin list and auth config.
type DocumentationPlugin struct {
	plugins []PayloadPlugin
	authCfg *auth.Config
}

func (DocumentationPlugin) Path() string { return "/openapi.json" }

func (d DocumentationPlugin) Handler() http.HandlerFunc {
	return NewOpenAPIHandler(d.plugins, d.authCfg)
}

// Init wires the documentation plugin with the full plugin list and auth config.
// Must be called before Handler() is used.
func (d *DocumentationPlugin) Init(plugins []PayloadPlugin, authCfg *auth.Config) {
	d.plugins = plugins
	d.authCfg = authCfg
}

func (DocumentationPlugin) OpenAPISpec() openapi.PathSpec {
	return openapi.PathSpec{
		Path: "/openapi.json",
		Operation: openapi.OpenAPIPath{
			Get: &openapi.OpenAPIOperation{
				Summary:     "Get OpenAPI specification",
				Description: "Returns the complete OpenAPI 3.1.0 specification for all available endpoints",
				Tags:        []string{"documentation"},
				Responses:   map[string]openapi.OpenAPIResponse{"200": {Description: "OpenAPI 3.1.0 specification"}},
			},
		},
	}
}

// SwaggerUIPlugin implements PayloadPlugin for the /swagger endpoint.
type SwaggerUIPlugin struct{}

func (SwaggerUIPlugin) Path() string              { return "/swagger" }
func (SwaggerUIPlugin) Handler() http.HandlerFunc { return SwaggerUIHandler }

func (SwaggerUIPlugin) OpenAPISpec() openapi.PathSpec {
	return openapi.PathSpec{
		Path: "/swagger",
		Operation: openapi.OpenAPIPath{
			Get: &openapi.OpenAPIOperation{
				Summary:     "Swagger UI",
				Description: "Interactive API documentation using Swagger UI",
				Tags:        []string{"documentation"},
				Responses:   map[string]openapi.OpenAPIResponse{"200": {Description: "Swagger UI HTML page"}},
			},
		},
	}
}

// NewOpenAPIHandler builds the OpenAPI specification from all registered plugins.
func NewOpenAPIHandler(plugins []PayloadPlugin, authCfg *auth.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		spec := openapi.OpenAPISpec{
			OpenAPI: "3.1.0",
			Info: openapi.OpenAPIInfo{
				Title:       "PayloadBuddy API",
				Description: "A REST API server for testing with large and streaming JSON payloads, specifically designed for ServiceNow integration testing",
				Version:     "1.0.0",
			},
			Servers: []openapi.OpenAPIServer{{URL: "http://localhost:8080", Description: "Development server"}},
			Paths:   make(map[string]openapi.OpenAPIPath),
			Components: &openapi.OpenAPIComponents{
				Schemas: make(map[string]*openapi.OpenAPISchema),
			},
		}

		for _, plugin := range plugins {
			pathSpec := plugin.OpenAPISpec()
			spec.Paths[pathSpec.Path] = pathSpec.Operation
			for name, schema := range pathSpec.Schemas {
				spec.Components.Schemas[name] = schema
			}
		}

		if authCfg != nil && authCfg.Enabled {
			if spec.Components.SecuritySchemes == nil {
				spec.Components.SecuritySchemes = make(map[string]*openapi.OpenAPISecurityScheme)
			}
			spec.Components.SecuritySchemes["BasicAuth"] = &openapi.OpenAPISecurityScheme{Type: "http", Scheme: "basic"}
			for path, pathItem := range spec.Paths {
				if pathItem.Get != nil {
					op := *pathItem.Get
					op.Security = []map[string][]string{{"BasicAuth": {}}}
					if op.Description != "" {
						op.Description += "\n\nRequires HTTP Basic Authentication when server is started with -auth flag."
					} else {
						op.Description = "Requires HTTP Basic Authentication when server is started with -auth flag."
					}
					pathItem.Get = &op
					spec.Paths[path] = pathItem
				}
			}
		}

		if err := json.NewEncoder(w).Encode(spec); err != nil {
			http.Error(w, "Failed to encode OpenAPI specification", http.StatusInternalServerError)
		}
	}
}

// SwaggerUIHandler serves the Swagger UI HTML interface.
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>PayloadBuddy API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css"
          integrity="sha384-ZJ2d83jl4Lvr6GKYzXpvQUmu+8us6T5frIryNHoLuypLK61jUnnCWZWyyrnifLda"
          crossorigin="anonymous" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"
            integrity="sha384-yrdF3mlUytUBwQyEVFAdwuUKEC9Qqrf+IUCgFgho4O5O6irf77pMjv36FN4eTpQD"
            crossorigin="anonymous"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"
            integrity="sha384-azzkurII4f+bjmZvm3hWhj7JezshyXtwobwneRyWCCIksK61Xi0Ry3xA2am9/TWp"
            crossorigin="anonymous"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
                plugins: [SwaggerUIBundle.plugins.DownloadUrl],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
	_, _ = w.Write([]byte(html))
}
