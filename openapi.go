package main

// OpenAPI 3.1.1 data structures for specification generation

// OpenAPISpec represents the complete OpenAPI 3.1.1 specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers,omitempty"`
	Paths      map[string]OpenAPIPath `json:"paths"`
	Components *OpenAPIComponents     `json:"components,omitempty"`
}

// OpenAPIInfo contains API metadata
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPIServer represents a server configuration
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// OpenAPIPath represents all operations for a specific path
type OpenAPIPath struct {
	Get    *OpenAPIOperation `json:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty"`
}

// OpenAPIOperation represents a single API operation
type OpenAPIOperation struct {
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
	Tags        []string                   `json:"tags,omitempty"`
	Security    []map[string][]string      `json:"security,omitempty"`
}

// OpenAPIParameter represents a parameter in the API
type OpenAPIParameter struct {
	Name        string         `json:"name"`
	In          string         `json:"in"` // "query", "header", "path", "cookie"
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      *OpenAPISchema `json:"schema,omitempty"`
	Example     interface{}    `json:"example,omitempty"`
}

// OpenAPIResponse represents a response from an API operation
type OpenAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty"`
}

// OpenAPIMediaType represents a media type (e.g., application/json)
type OpenAPIMediaType struct {
	Schema  *OpenAPISchema `json:"schema,omitempty"`
	Example interface{}    `json:"example,omitempty"`
}

// OpenAPISchema represents a data schema
type OpenAPISchema struct {
	Type        string                    `json:"type,omitempty"`
	Format      string                    `json:"format,omitempty"`
	Items       *OpenAPISchema            `json:"items,omitempty"`
	Properties  map[string]*OpenAPISchema `json:"properties,omitempty"`
	Required    []string                  `json:"required,omitempty"`
	Example     interface{}               `json:"example,omitempty"`
	Description string                    `json:"description,omitempty"`
	Minimum     *int                      `json:"minimum,omitempty"`
	Maximum     *int                      `json:"maximum,omitempty"`
	Enum        []interface{}             `json:"enum,omitempty"`
}

// OpenAPISecurityScheme represents a security scheme definition
type OpenAPISecurityScheme struct {
	Type   string `json:"type"`
	Scheme string `json:"scheme,omitempty"`
}

// OpenAPIComponents contains reusable components
type OpenAPIComponents struct {
	Schemas         map[string]*OpenAPISchema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*OpenAPISecurityScheme `json:"securitySchemes,omitempty"`
}

// OpenAPIPathSpec represents the specification contribution from a single plugin
type OpenAPIPathSpec struct {
	Path      string                    `json:"path"`
	Operation OpenAPIPath               `json:"operation"`
	Schemas   map[string]*OpenAPISchema `json:"schemas,omitempty"`
}
