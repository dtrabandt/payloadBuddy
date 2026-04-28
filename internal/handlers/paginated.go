package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/openapi"
	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

// PaginatedItem represents a single object in a paginated response.
type PaginatedItem struct {
	ID        int       `json:"id"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	SysID     string    `json:"sys_id,omitempty"`
	Number    string    `json:"number,omitempty"`
	State     string    `json:"state,omitempty"`
}

// PaginationMetadata contains pagination state for the current page.
type PaginationMetadata struct {
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page,omitempty"`
	Size       int     `json:"size,omitempty"`
	Limit      int     `json:"limit,omitempty"`
	Offset     int     `json:"offset,omitempty"`
	HasMore    bool    `json:"has_more"`
	NextOffset *int    `json:"next_offset,omitempty"`
	NextPage   *int    `json:"next_page,omitempty"`
	NextCursor *string `json:"next_cursor,omitempty"`
}

// PaginatedResponse is the envelope returned by the /paginated_payload endpoint.
type PaginatedResponse struct {
	Result   []PaginatedItem    `json:"result"`
	Metadata PaginationMetadata `json:"metadata"`
}

// PaginatedPayloadPlugin implements PayloadPlugin for the /paginated_payload endpoint.
type PaginatedPayloadPlugin struct {
	SM *scenarios.Manager
}

func (PaginatedPayloadPlugin) Path() string { return "/paginated_payload" }

func (p PaginatedPayloadPlugin) Handler() http.HandlerFunc {
	return NewPaginatedHandler(p.SM)
}

// NewPaginatedHandler returns an http.HandlerFunc backed by the given scenario manager.
func NewPaginatedHandler(sm *scenarios.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scenario := strings.ToLower(r.URL.Query().Get("scenario"))

		var defaultCount, maxCount, defaultBatchSize int
		var defaultSNMode bool
		if sm != nil && scenario != "" {
			defaultBatchSize, defaultSNMode, maxCount, defaultCount = sm.GetScenarioConfig(scenario)
		} else {
			defaultCount = 10000
			maxCount = 1000000
			defaultBatchSize = 100
		}

		totalCount := getIntParam(r, "total", defaultCount)
		limit := getIntParam(r, "limit", defaultBatchSize)
		offset := getIntParam(r, "offset", 0)
		page := getIntParam(r, "page", 1)
		size := getIntParam(r, "size", defaultBatchSize)
		cursor := r.URL.Query().Get("cursor")

		snMode := defaultSNMode
		if v := r.URL.Query().Get("servicenow"); v != "" {
			snMode = v == "true"
		}
		delay := getDurationParam(r, "delay", 0)

		if totalCount <= 0 || totalCount > maxCount {
			http.Error(w, fmt.Sprintf("Total count must be between 1 and %d", maxCount), http.StatusBadRequest)
			return
		}

		if scenario != "" && sm != nil {
			scenarioDelay, _ := sm.GetScenarioDelay(scenario, 0)
			if scenarioDelay > 0 {
				time.Sleep(scenarioDelay)
			}
		} else if delay > 0 {
			time.Sleep(delay)
		}

		var startIndex, pageSize int
		var paginationType string

		switch {
		case cursor != "":
			paginationType = "cursor"
			startIndex, pageSize = parseCursor(cursor, limit)
		case r.URL.Query().Has("page") || r.URL.Query().Has("size"):
			paginationType = "page"
			if page < 1 {
				page = 1
			}
			if size <= 0 || size > 1000 {
				size = 100
			}
			startIndex = (page - 1) * size
			pageSize = size
		default:
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

		if startIndex >= totalCount {
			response := PaginatedResponse{
				Result:   []PaginatedItem{},
				Metadata: buildMetadata(paginationType, totalCount, startIndex, pageSize, page, size, limit, offset, false),
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
			return
		}

		endIndex := min(startIndex+pageSize, totalCount)
		actualSize := endIndex - startIndex

		items := make([]PaginatedItem, actualSize)
		for i := range actualSize {
			itemID := startIndex + i + 1
			if snMode {
				items[i] = PaginatedItem{
					ID:        itemID,
					Value:     fmt.Sprintf("ServiceNow Record %d", itemID),
					Timestamp: time.Now(),
					SysID:     generateSysID(),
					Number:    fmt.Sprintf("INC%07d", itemID),
					State:     []string{"New", "In Progress", "Resolved", "Closed"}[itemID%4],
				}
			} else {
				items[i] = PaginatedItem{ID: itemID, Value: fmt.Sprintf("Item %d", itemID), Timestamp: time.Now()}
			}
		}

		hasMore := endIndex < totalCount
		response := PaginatedResponse{
			Result:   items,
			Metadata: buildMetadata(paginationType, totalCount, startIndex, pageSize, page, size, limit, offset, hasMore),
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func buildMetadata(paginationType string, totalCount, startIndex, pageSize, page, size, limit, offset int, hasMore bool) PaginationMetadata {
	m := PaginationMetadata{TotalCount: totalCount, HasMore: hasMore}
	switch paginationType {
	case "page":
		m.Page = page
		m.Size = size
		if hasMore {
			next := page + 1
			m.NextPage = &next
		}
	case "cursor":
		m.Limit = pageSize
		if hasMore {
			next := createCursor(startIndex + pageSize)
			m.NextCursor = &next
		}
	default:
		m.Limit = limit
		m.Offset = offset
		if hasMore {
			next := offset + limit
			m.NextOffset = &next
		}
	}
	return m
}

func parseCursor(cursor string, defaultLimit int) (int, int) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil || len(decoded) == 0 {
		return 0, defaultLimit
	}
	var cd struct {
		ID    int `json:"id"`
		Limit int `json:"limit"`
	}
	if err := json.Unmarshal(decoded, &cd); err != nil {
		return 0, defaultLimit
	}
	lim := cd.Limit
	if lim <= 0 || lim > 1000 {
		lim = defaultLimit
	}
	return cd.ID, lim
}

func createCursor(startID int) string {
	cd := struct {
		ID    int `json:"id"`
		Limit int `json:"limit"`
	}{ID: startID, Limit: 100}
	data, _ := json.Marshal(cd)
	return base64.URLEncoding.EncodeToString(data)
}

func (PaginatedPayloadPlugin) OpenAPISpec() openapi.PathSpec {
	return openapi.PathSpec{
		Path: "/paginated_payload",
		Operation: openapi.OpenAPIPath{
			Get: &openapi.OpenAPIOperation{
				Summary:     "Get paginated JSON payload",
				Description: "Returns paginated JSON data supporting limit/offset, page/size, and cursor-based pagination for ServiceNow Data Stream actions",
				Tags:        []string{"pagination", "servicenow"},
				Parameters: []openapi.OpenAPIParameter{
					{Name: "total", In: "query", Description: "Total items across all pages (default: 10000, max: 1000000)", Schema: &openapi.OpenAPISchema{Type: "integer", Minimum: &[]int{1}[0], Maximum: &[]int{1000000}[0], Example: 10000}},
					{Name: "limit", In: "query", Description: "Items per page for limit/offset pagination (default: 100, max: 1000)", Schema: &openapi.OpenAPISchema{Type: "integer", Minimum: &[]int{1}[0], Maximum: &[]int{1000}[0], Example: 100}},
					{Name: "offset", In: "query", Description: "Starting position for limit/offset pagination (default: 0)", Schema: &openapi.OpenAPISchema{Type: "integer", Minimum: &[]int{0}[0], Example: 0}},
					{Name: "page", In: "query", Description: "Page number for page/size pagination (default: 1)", Schema: &openapi.OpenAPISchema{Type: "integer", Minimum: &[]int{1}[0], Example: 1}},
					{Name: "size", In: "query", Description: "Items per page for page/size pagination (default: 100, max: 1000)", Schema: &openapi.OpenAPISchema{Type: "integer", Minimum: &[]int{1}[0], Maximum: &[]int{1000}[0], Example: 100}},
					{Name: "cursor", In: "query", Description: "Cursor token for cursor-based pagination", Schema: &openapi.OpenAPISchema{Type: "string"}},
					{Name: "servicenow", In: "query", Description: "Enable ServiceNow-style record format", Schema: &openapi.OpenAPISchema{Type: "boolean", Example: false}},
					{Name: "delay", In: "query", Description: "Delay before response", Schema: &openapi.OpenAPISchema{Type: "string", Example: "100ms"}},
					{Name: "scenario", In: "query", Description: "ServiceNow simulation scenario", Schema: &openapi.OpenAPISchema{Type: "string", Enum: []any{"peak_hours", "maintenance", "network_issues", "database_load"}, Example: "peak_hours"}},
				},
				Responses: map[string]openapi.OpenAPIResponse{
					"200": {Description: "Successful paginated response"},
					"400": {Description: "Bad request — invalid parameters"},
					"500": {Description: "Internal server error"},
				},
			},
		},
		Schemas: map[string]*openapi.OpenAPISchema{
			"PaginatedItem": {Type: "object", Properties: map[string]*openapi.OpenAPISchema{
				"id": {Type: "integer"}, "value": {Type: "string"}, "timestamp": {Type: "string", Format: "date-time"},
				"sys_id": {Type: "string"}, "number": {Type: "string"}, "state": {Type: "string"},
			}, Required: []string{"id", "value", "timestamp"}},
			"PaginationMetadata": {Type: "object", Properties: map[string]*openapi.OpenAPISchema{
				"total_count": {Type: "integer"}, "has_more": {Type: "boolean"},
				"next_offset": {Type: "integer"}, "next_page": {Type: "integer"}, "next_cursor": {Type: "string"},
			}, Required: []string{"total_count", "has_more"}},
		},
	}
}
