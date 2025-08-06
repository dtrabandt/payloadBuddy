package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PaginatedItem represents a single object in a paginated response
type PaginatedItem struct {
	ID        int       `json:"id"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	SysID     string    `json:"sys_id,omitempty"` // ServiceNow style
	Number    string    `json:"number,omitempty"` // ServiceNow ticket number
	State     string    `json:"state,omitempty"`  // ServiceNow state
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page,omitempty"`   // For page/size pagination
	Size       int     `json:"size,omitempty"`   // For page/size pagination
	Limit      int     `json:"limit,omitempty"`  // For limit/offset pagination
	Offset     int     `json:"offset,omitempty"` // For limit/offset pagination
	HasMore    bool    `json:"has_more"`
	NextOffset *int    `json:"next_offset,omitempty"` // For limit/offset pagination
	NextPage   *int    `json:"next_page,omitempty"`   // For page/size pagination
	NextCursor *string `json:"next_cursor,omitempty"` // For cursor-based pagination
}

// PaginatedResponse represents the complete paginated API response
type PaginatedResponse struct {
	Result   []PaginatedItem    `json:"result"`
	Metadata PaginationMetadata `json:"metadata"`
}

// PaginatedPayloadHandler handles paginated REST API responses
//
// Query Parameters:
//   - total: Total number of items available (default: 10000, scenario-configurable)
//   - limit: Number of items per page for limit/offset pagination (default: 100, scenario-configurable)
//   - offset: Starting position for limit/offset pagination (default: 0)
//   - page: Page number for page/size pagination (default: 1)
//   - size: Items per page for page/size pagination (default: 100, scenario-configurable)
//   - cursor: Cursor token for cursor-based pagination
//   - servicenow: Generate ServiceNow-style fields (default: false, scenario-configurable)
//   - delay: Delay before response (e.g., "100ms", "1s")
//   - scenario: ServiceNow scenarios ("peak_hours", "maintenance", "network_issues", "database_load")
//
// Pagination Types:
//   - Limit/Offset: Use 'limit' and 'offset' parameters
//   - Page/Size: Use 'page' and 'size' parameters
//   - Cursor: Use 'cursor' parameter
//
// Examples:
//   - /paginated_payload?limit=50&offset=100
//   - /paginated_payload?page=2&size=25&servicenow=true
//   - /paginated_payload?cursor=eyJpZCI6MTAwfQ%3D%3D
//   - /paginated_payload?scenario=peak_hours&servicenow=true
//   - /paginated_payload?scenario=database_load&limit=25
func PaginatedPayloadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse scenario parameter
	scenario := strings.ToLower(r.URL.Query().Get("scenario"))

	// Get scenario-based defaults if scenario manager is available and scenario is specified
	var defaultCount, maxCount, defaultBatchSize int
	var defaultServiceNowMode bool
	if scenarioManager != nil && scenario != "" {
		defaultBatchSize, defaultServiceNowMode, maxCount, defaultCount = scenarioManager.GetScenarioConfig(scenario)
	} else {
		// Use hardcoded defaults for backward compatibility
		defaultCount = 10000
		maxCount = 1000000
		defaultBatchSize = 100
		defaultServiceNowMode = false
	}

	// Parse parameters with scenario-aware defaults
	totalCount := getIntParam(r, "total", defaultCount)
	limit := getIntParam(r, "limit", defaultBatchSize)
	offset := getIntParam(r, "offset", 0)
	page := getIntParam(r, "page", 1)
	size := getIntParam(r, "size", defaultBatchSize)
	cursor := r.URL.Query().Get("cursor")

	// ServiceNow mode: use scenario default unless explicitly overridden
	serviceNowMode := defaultServiceNowMode
	if serviceNowParam := r.URL.Query().Get("servicenow"); serviceNowParam != "" {
		serviceNowMode = serviceNowParam == "true"
	}

	delay := getDurationParam(r, "delay", 0)

	// Validate parameters
	if totalCount <= 0 || totalCount > maxCount {
		http.Error(w, fmt.Sprintf("Total count must be between 1 and %d", maxCount), http.StatusBadRequest)
		return
	}

	// Apply scenario-based delay if specified
	if scenario != "" && scenarioManager != nil {
		// For pagination, use item index 0 to get base scenario delay
		scenarioDelay, _ := scenarioManager.GetScenarioDelay(scenario, 0)
		if scenarioDelay > 0 {
			time.Sleep(scenarioDelay)
		}
	} else if delay > 0 {
		// Apply custom delay if specified (simulates API processing time)
		time.Sleep(delay)
	}

	// Determine pagination type and calculate parameters
	var startIndex, pageSize int
	var paginationType string

	if cursor != "" {
		// Cursor-based pagination
		paginationType = "cursor"
		startIndex, pageSize = parseCursor(cursor, limit)
	} else if r.URL.Query().Has("page") || r.URL.Query().Has("size") {
		// Page/size pagination
		paginationType = "page"
		if page < 1 {
			page = 1
		}
		if size <= 0 || size > 1000 {
			size = 100
		}
		startIndex = (page - 1) * size
		pageSize = size
	} else {
		// Limit/offset pagination (default)
		paginationType = "offset"
		if offset < 0 {
			offset = 0
		}
		if limit <= 0 || limit > 1000 {
			limit = 100
		}
		startIndex = offset
		pageSize = limit
	}

	// Validate bounds
	if startIndex >= totalCount {
		// Return empty page if offset/page is beyond data
		response := PaginatedResponse{
			Result:   []PaginatedItem{},
			Metadata: createPaginationMetadata(paginationType, totalCount, startIndex, pageSize, page, size, limit, offset, false),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Calculate end index and actual items to return
	endIndex := min(startIndex+pageSize, totalCount)
	actualSize := endIndex - startIndex

	// Generate items for this page
	items := make([]PaginatedItem, actualSize)
	for i := range actualSize {
		itemID := startIndex + i + 1 // 1-based IDs
		var item PaginatedItem

		if serviceNowMode {
			item = PaginatedItem{
				ID:        itemID,
				Value:     fmt.Sprintf("ServiceNow Record %d", itemID),
				Timestamp: time.Now(),
				SysID:     generateSysID(),
				Number:    fmt.Sprintf("INC%07d", itemID),
				State:     []string{"New", "In Progress", "Resolved", "Closed"}[itemID%4],
			}
		} else {
			item = PaginatedItem{
				ID:        itemID,
				Value:     fmt.Sprintf("Item %d", itemID),
				Timestamp: time.Now(),
			}
		}
		items[i] = item
	}

	// Determine if there are more pages
	hasMore := endIndex < totalCount

	// Create response
	response := PaginatedResponse{
		Result:   items,
		Metadata: createPaginationMetadata(paginationType, totalCount, startIndex, pageSize, page, size, limit, offset, hasMore),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// createPaginationMetadata creates appropriate metadata based on pagination type
func createPaginationMetadata(paginationType string, totalCount, startIndex, pageSize, page, size, limit, offset int, hasMore bool) PaginationMetadata {
	metadata := PaginationMetadata{
		TotalCount: totalCount,
		HasMore:    hasMore,
	}

	switch paginationType {
	case "page":
		metadata.Page = page
		metadata.Size = size
		if hasMore {
			nextPage := page + 1
			metadata.NextPage = &nextPage
		}
	case "cursor":
		metadata.Limit = pageSize
		if hasMore {
			nextCursor := createCursor(startIndex + pageSize)
			metadata.NextCursor = &nextCursor
		}
	default: // offset
		metadata.Limit = limit
		metadata.Offset = offset
		if hasMore {
			nextOffset := offset + limit
			metadata.NextOffset = &nextOffset
		}
	}

	return metadata
}

// parseCursor decodes a cursor token to extract starting position
func parseCursor(cursor string, defaultLimit int) (int, int) {
	// Simple base64 encoded JSON cursor: {"id":100,"limit":50}
	// For production, use more secure/complex cursor implementation
	decoded, err := base64Decode(cursor)
	if err != nil {
		return 0, defaultLimit
	}

	var cursorData struct {
		ID    int `json:"id"`
		Limit int `json:"limit"`
	}

	if err := json.Unmarshal([]byte(decoded), &cursorData); err != nil {
		return 0, defaultLimit
	}

	limit := cursorData.Limit
	if limit <= 0 || limit > 1000 {
		limit = defaultLimit
	}

	return cursorData.ID, limit
}

// createCursor creates a cursor token for the given starting position
func createCursor(startID int) string {
	cursorData := struct {
		ID    int `json:"id"`
		Limit int `json:"limit"`
	}{
		ID:    startID,
		Limit: 100, // Default limit for cursor pagination
	}

	data, _ := json.Marshal(cursorData)
	return base64Encode(string(data))
}

// Simple base64 encoding/decoding helpers
func base64Encode(data string) string {
	// Simple implementation - in production, use encoding/base64
	return fmt.Sprintf("cursor_%d", len(data)) // Simplified for demo
}

func base64Decode(cursor string) (string, error) {
	// Simple implementation - in production, use encoding/base64
	if cursor == "" {
		return "", fmt.Errorf("empty cursor")
	}
	return "{\"id\":0,\"limit\":100}", nil // Simplified for demo
}

// Plugin registration
type PaginatedPayloadPlugin struct{}

func (p PaginatedPayloadPlugin) Path() string {
	return "/paginated_payload"
}

func (p PaginatedPayloadPlugin) Handler() http.HandlerFunc {
	return PaginatedPayloadHandler
}

func init() {
	registerPlugin(PaginatedPayloadPlugin{})
}

// OpenAPISpec returns the OpenAPI specification for the paginated payload endpoint
func (p PaginatedPayloadPlugin) OpenAPISpec() OpenAPIPathSpec {
	return OpenAPIPathSpec{
		Path:      "/paginated_payload",
		Operation: p.buildOpenAPIOperation(),
		Schemas:   p.buildOpenAPISchemas(),
	}
}

// buildOpenAPIOperation creates the OpenAPI operation specification
func (p PaginatedPayloadPlugin) buildOpenAPIOperation() OpenAPIPath {
	return OpenAPIPath{
		Get: &OpenAPIOperation{
			Summary:     "Get paginated JSON payload",
			Description: "Returns paginated JSON data supporting limit/offset, page/size, and cursor-based pagination patterns commonly used with ServiceNow Data Stream actions",
			Tags:        []string{"pagination", "servicenow"},
			Parameters:  p.buildOpenAPIParameters(),
			Responses:   p.buildOpenAPIResponses(),
		},
	}
}

// buildOpenAPIParameters creates the parameter specifications
func (p PaginatedPayloadPlugin) buildOpenAPIParameters() []OpenAPIParameter {
	return []OpenAPIParameter{
		{
			Name:        "total",
			In:          "query",
			Description: "Total number of items available across all pages (default: 10000, max: 1000000)",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "integer",
				Minimum: &[]int{1}[0],
				Maximum: &[]int{1000000}[0],
				Example: 10000,
			},
		},
		{
			Name:        "limit",
			In:          "query",
			Description: "Number of items per page for limit/offset pagination (default: 100, max: 1000)",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "integer",
				Minimum: &[]int{1}[0],
				Maximum: &[]int{1000}[0],
				Example: 100,
			},
		},
		{
			Name:        "offset",
			In:          "query",
			Description: "Starting position for limit/offset pagination (default: 0)",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "integer",
				Minimum: &[]int{0}[0],
				Example: 0,
			},
		},
		{
			Name:        "page",
			In:          "query",
			Description: "Page number for page/size pagination (default: 1)",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "integer",
				Minimum: &[]int{1}[0],
				Example: 1,
			},
		},
		{
			Name:        "size",
			In:          "query",
			Description: "Items per page for page/size pagination (default: 100, max: 1000)",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "integer",
				Minimum: &[]int{1}[0],
				Maximum: &[]int{1000}[0],
				Example: 100,
			},
		},
		{
			Name:        "cursor",
			In:          "query",
			Description: "Cursor token for cursor-based pagination",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "string",
				Example: "eyJpZCI6MTAwfQ%3D%3D",
			},
		},
		{
			Name:        "servicenow",
			In:          "query",
			Description: "Enable ServiceNow-style record format with sys_id, number, and state fields",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "boolean",
				Example: false,
			},
		},
		{
			Name:        "delay",
			In:          "query",
			Description: "Delay before response (e.g., '100ms', '1s', or just milliseconds)",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "string",
				Example: "100ms",
			},
		},
		{
			Name:        "scenario",
			In:          "query",
			Description: "ServiceNow simulation scenario. All scenarios work with pagination: 'peak_hours' (consistent delays, ideal for both), 'maintenance' (single spike per page), 'network_issues' (random delays per page), 'database_load' (single delay per page)",
			Required:    false,
			Schema: &OpenAPISchema{
				Type:    "string",
				Enum:    []any{"peak_hours", "maintenance", "network_issues", "database_load"},
				Example: "peak_hours",
			},
		},
	}
}

// buildOpenAPIResponses creates the response specifications
func (p PaginatedPayloadPlugin) buildOpenAPIResponses() map[string]OpenAPIResponse {
	return map[string]OpenAPIResponse{
		"200": {
			Description: "Successful paginated response",
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: &OpenAPISchema{
						Type: "object",
						Properties: map[string]*OpenAPISchema{
							"result": {
								Type: "array",
								Items: &OpenAPISchema{
									Type: "object",
									Properties: map[string]*OpenAPISchema{
										"id": {
											Type:        "integer",
											Description: "Unique identifier for the item",
											Example:     1,
										},
										"value": {
											Type:        "string",
											Description: "Value or description of the item",
											Example:     "Item 1",
										},
										"timestamp": {
											Type:        "string",
											Format:      "date-time",
											Description: "Timestamp when the item was generated",
										},
										"sys_id": {
											Type:        "string",
											Description: "ServiceNow system ID (when ServiceNow mode is enabled)",
											Example:     "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
										},
										"number": {
											Type:        "string",
											Description: "ServiceNow ticket number (when ServiceNow mode is enabled)",
											Example:     "INC0000001",
										},
										"state": {
											Type:        "string",
											Description: "ServiceNow state (when ServiceNow mode is enabled)",
											Example:     "New",
										},
									},
									Required: []string{"id", "value", "timestamp"},
								},
							},
							"metadata": {
								Type: "object",
								Properties: map[string]*OpenAPISchema{
									"total_count": {
										Type:        "integer",
										Description: "Total number of items across all pages",
										Example:     10000,
									},
									"page": {
										Type:        "integer",
										Description: "Current page number (page/size pagination)",
										Example:     1,
									},
									"size": {
										Type:        "integer",
										Description: "Items per page (page/size pagination)",
										Example:     100,
									},
									"limit": {
										Type:        "integer",
										Description: "Items per page (limit/offset pagination)",
										Example:     100,
									},
									"offset": {
										Type:        "integer",
										Description: "Starting position (limit/offset pagination)",
										Example:     0,
									},
									"has_more": {
										Type:        "boolean",
										Description: "Whether more pages are available",
										Example:     true,
									},
									"next_offset": {
										Type:        "integer",
										Description: "Next offset for limit/offset pagination",
										Example:     100,
									},
									"next_page": {
										Type:        "integer",
										Description: "Next page number for page/size pagination",
										Example:     2,
									},
									"next_cursor": {
										Type:        "string",
										Description: "Next cursor token for cursor-based pagination",
										Example:     "eyJpZCI6MjAwfQ%3D%3D",
									},
								},
								Required: []string{"total_count", "has_more"},
							},
						},
						Required: []string{"result", "metadata"},
					},
					Example: PaginatedResponse{
						Result: []PaginatedItem{
							{
								ID:        1,
								Value:     "Item 1",
								Timestamp: time.Now(),
							},
							{
								ID:        2,
								Value:     "ServiceNow Record 2",
								Timestamp: time.Now(),
								SysID:     "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
								Number:    "INC0000002",
								State:     "In Progress",
							},
						},
						Metadata: PaginationMetadata{
							TotalCount: 10000,
							Limit:      100,
							Offset:     0,
							HasMore:    true,
							NextOffset: &[]int{100}[0],
						},
					},
				},
			},
		},
		"400": {
			Description: "Bad request - invalid parameters",
			Content: map[string]OpenAPIMediaType{
				"text/plain": {
					Schema: &OpenAPISchema{
						Type:    "string",
						Example: "Total count must be between 1 and 1,000,000",
					},
				},
			},
		},
		"500": {
			Description: "Internal server error",
			Content: map[string]OpenAPIMediaType{
				"text/plain": {
					Schema: &OpenAPISchema{
						Type:    "string",
						Example: "Failed to encode response",
					},
				},
			},
		},
	}
}

// buildOpenAPISchemas creates the schema specifications
func (p PaginatedPayloadPlugin) buildOpenAPISchemas() map[string]*OpenAPISchema {
	return map[string]*OpenAPISchema{
		"PaginatedItem": {
			Type: "object",
			Properties: map[string]*OpenAPISchema{
				"id": {
					Type:        "integer",
					Description: "Unique identifier for the item",
				},
				"value": {
					Type:        "string",
					Description: "Value or description of the item",
				},
				"timestamp": {
					Type:        "string",
					Format:      "date-time",
					Description: "Timestamp when the item was generated",
				},
				"sys_id": {
					Type:        "string",
					Description: "ServiceNow system ID (optional)",
				},
				"number": {
					Type:        "string",
					Description: "ServiceNow ticket number (optional)",
				},
				"state": {
					Type:        "string",
					Description: "ServiceNow state (optional)",
				},
			},
			Required: []string{"id", "value", "timestamp"},
		},
		"PaginationMetadata": {
			Type: "object",
			Properties: map[string]*OpenAPISchema{
				"total_count": {
					Type:        "integer",
					Description: "Total number of items across all pages",
				},
				"page": {
					Type:        "integer",
					Description: "Current page number (page/size pagination)",
				},
				"size": {
					Type:        "integer",
					Description: "Items per page (page/size pagination)",
				},
				"limit": {
					Type:        "integer",
					Description: "Items per page (limit/offset pagination)",
				},
				"offset": {
					Type:        "integer",
					Description: "Starting position (limit/offset pagination)",
				},
				"has_more": {
					Type:        "boolean",
					Description: "Whether more pages are available",
				},
				"next_offset": {
					Type:        "integer",
					Description: "Next offset for limit/offset pagination",
				},
				"next_page": {
					Type:        "integer",
					Description: "Next page number for page/size pagination",
				},
				"next_cursor": {
					Type:        "string",
					Description: "Next cursor token for cursor-based pagination",
				},
			},
			Required: []string{"total_count", "has_more"},
		},
		"PaginatedResponse": {
			Type: "object",
			Properties: map[string]*OpenAPISchema{
				"result": {
					Type: "array",
					Items: &OpenAPISchema{
						Type:        "object",
						Description: "PaginatedItem object",
					},
				},
				"metadata": {
					Type:        "object",
					Description: "PaginationMetadata object",
				},
			},
			Required: []string{"result", "metadata"},
		},
	}
}
