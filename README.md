# gohugePayloadServer

A simple Go server designed to test REST client implementations with large payloads. This project is intended for training and troubleshooting scenarios where REST consumers (such as ServiceNow) struggle to process or return large datasets in memory.

## Purpose

This server helps consultants and developers:
- Simulate REST endpoints that return very large JSON payloads.
- Identify and troubleshoot issues with clients (e.g., ServiceNow Flow Actions) that cannot handle large responses.
- Train on best practices for handling large data transfers in REST APIs.

## Use Case

In some environments, such as ServiceNow, REST calls from Flow Actions may fail or hang if the response payload is too large to fit in memory. This server allows you to:
- Generate and serve large JSON responses.
- Test how your client applications behave under these conditions.
- Develop strategies for paginating, streaming, or otherwise managing large data transfers.

## Getting Started

### Prerequisites
- Go 1.18 or newer

### Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/your-org/gohugePayloadServer.git
   cd gohugePayloadServer
   ```
2. Build the server:
   ```sh
   go build -o gohugePayloadServer
   ```

### Usage
Run the server:
```sh
./gohugePayloadServer
```

By default, the server listens on port 8080. You can test the endpoint (e.g., `/payload`) to receive a large JSON response.

#### Example Request
```sh
curl http://localhost:8080/huge_payload
```

## Running Tests

This project includes schema-based unit tests to ensure the structure of JSON responses matches the documented schema.  
To run all tests (including schema validation):

```sh
go test ./...
```
See `huge_payload_handler_test.go` for example tests using gojsonschema.

### Example: JSON Schema Validation Test
```go
import (
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/xeipuuv/gojsonschema"
)

func TestHugePayloadHandler_JSONSchema(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/huge_payload", nil)
    w := httptest.NewRecorder()

    HugePayloadHandler(w, req)

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
        t.Fatalf("Schema validation failed: %v", err)
    }
    if !result.Valid() {
        for _, err := range result.Errors() {
            t.Errorf("Schema error: %s", err)
        }
    }
}
```

## Customization
- Adjust the payload size or structure in `hugePayloadHandler.go` as needed for your testing scenario.
- Modify endpoints or add new ones to simulate different REST behaviors.

## Dependencies

This project uses the following Go package for schema-based testing:

- [github.com/xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema): Used in test cases to validate JSON API responses against a JSON Schema.


## Best Practices for Handling Large Payloads
- **Pagination:** Break large datasets into smaller pages.
- **Streaming:** Use chunked transfer encoding or streaming APIs.
- **Compression:** Enable gzip or similar compression for responses.
- **Timeouts & Limits:** Set reasonable client/server timeouts and payload size limits.

## Contributing
Contributions are welcome! Please open issues or submit pull requests for improvements.

## License
MIT License

## Authors
- Dennis Trabandt

---

*This project is for educational and troubleshooting purposes. Use with caution in production environments.*
