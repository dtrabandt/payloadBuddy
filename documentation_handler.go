package main

import (
	"encoding/json"
	"net/http"
)

// DocumentationPlugin implements PayloadPlugin for OpenAPI documentation
type DocumentationPlugin struct{}

// Path returns the HTTP path for the OpenAPI JSON endpoint
func (d DocumentationPlugin) Path() string { 
	return "/openapi.json" 
}

// Handler returns the handler function for the OpenAPI JSON endpoint
func (d DocumentationPlugin) Handler() http.HandlerFunc { 
	return OpenAPIHandler 
}

// OpenAPISpec returns the OpenAPI specification for the documentation endpoint itself
func (d DocumentationPlugin) OpenAPISpec() OpenAPIPathSpec {
	return OpenAPIPathSpec{
		Path: "/openapi.json",
		Operation: OpenAPIPath{
			Get: &OpenAPIOperation{
				Summary:     "Get OpenAPI specification",
				Description: "Returns the complete OpenAPI 3.1.1 specification for all available endpoints",
				Tags:        []string{"documentation"},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "OpenAPI 3.1.1 specification",
						Content: map[string]OpenAPIMediaType{
							"application/json": {
								Schema: &OpenAPISchema{
									Type:        "object",
									Description: "OpenAPI 3.1.1 specification document",
								},
							},
						},
					},
				},
			},
		},
	}
}

// SwaggerUIPlugin implements PayloadPlugin for Swagger UI
type SwaggerUIPlugin struct{}

// Path returns the HTTP path for the Swagger UI endpoint
func (s SwaggerUIPlugin) Path() string { 
	return "/swagger" 
}

// Handler returns the handler function for the Swagger UI endpoint
func (s SwaggerUIPlugin) Handler() http.HandlerFunc { 
	return SwaggerUIHandler 
}

// OpenAPISpec returns the OpenAPI specification for the Swagger UI endpoint
func (s SwaggerUIPlugin) OpenAPISpec() OpenAPIPathSpec {
	return OpenAPIPathSpec{
		Path: "/swagger",
		Operation: OpenAPIPath{
			Get: &OpenAPIOperation{
				Summary:     "Swagger UI",
				Description: "Interactive API documentation using Swagger UI",
				Tags:        []string{"documentation"},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "Swagger UI HTML page",
						Content: map[string]OpenAPIMediaType{
							"text/html": {
								Schema: &OpenAPISchema{
									Type:        "string",
									Description: "HTML page with Swagger UI",
								},
							},
						},
					},
				},
			},
		},
	}
}

// OpenAPIHandler generates and serves the complete OpenAPI 3.1.1 specification
func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create the base OpenAPI specification
	spec := OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: OpenAPIInfo{
			Title:       "PayloadBuddy API",
			Description: "A REST API server for testing with large and streaming JSON payloads, specifically designed for ServiceNow integration testing",
			Version:     "1.0.0",
		},
		Servers: []OpenAPIServer{
			{
				URL:         "http://localhost:8080",
				Description: "Development server",
			},
		},
		Paths:      make(map[string]OpenAPIPath),
		Components: &OpenAPIComponents{
			Schemas: make(map[string]*OpenAPISchema),
		},
	}

	// Collect specifications from all plugins
	for _, plugin := range plugins {
		pathSpec := plugin.OpenAPISpec()
		
		// Add the path operation
		spec.Paths[pathSpec.Path] = pathSpec.Operation
		
		// Merge schemas
		if pathSpec.Schemas != nil {
			for name, schema := range pathSpec.Schemas {
				spec.Components.Schemas[name] = schema
			}
		}
	}

	// Add authentication security scheme if authentication is enabled
	if *enableAuth {
		if spec.Components == nil {
			spec.Components = &OpenAPIComponents{}
		}
		if spec.Components.SecuritySchemes == nil {
			spec.Components.SecuritySchemes = make(map[string]*OpenAPISecurityScheme)
		}
		
		// Add Basic Auth security scheme
		spec.Components.SecuritySchemes["BasicAuth"] = &OpenAPISecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}

		// Add security requirements to each operation
		for path, pathItem := range spec.Paths {
			if pathItem.Get != nil {
				// Create a copy of the operation to avoid modifying the original
				newGet := *pathItem.Get
				// Add security requirement
				newGet.Security = []map[string][]string{
					{"BasicAuth": {}},
				}
				// Update description to document auth requirement
				if newGet.Description != "" {
					newGet.Description += "\n\nRequires HTTP Basic Authentication when server is started with -auth flag."
				} else {
					newGet.Description = "Requires HTTP Basic Authentication when server is started with -auth flag."
				}
				pathItem.Get = &newGet
				spec.Paths[path] = pathItem
			}
		}
	}

	// Encode and send the specification
	if err := json.NewEncoder(w).Encode(spec); err != nil {
		http.Error(w, "Failed to encode OpenAPI specification", http.StatusInternalServerError)
	}
}

// SwaggerUIHandler serves the Swagger UI HTML interface
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>PayloadBuddy API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	_, _ = w.Write([]byte(html))
}

// Register documentation plugins in init function
func init() {
	registerPlugin(DocumentationPlugin{})
	registerPlugin(SwaggerUIPlugin{})
}