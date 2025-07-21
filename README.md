# gohugePayloadServer

A simple Go server designed to test REST client implementations with large payloads and streaming data. This project is intended for training and troubleshooting scenarios where REST consumers (such as ServiceNow) struggle to process or return large datasets in memory.

## Purpose

This server helps consultants and developers:
- Simulate REST endpoints that return very large JSON payloads.
- Test streaming endpoints for clients that need to process data incrementally.
- Identify and troubleshoot issues with clients (e.g., ServiceNow Flow Actions) that cannot handle large responses.
- Train on best practices for handling large data transfers in REST APIs.

## Features

- **/huge_payload**: Returns a large JSON array (100,000 objects) in a single response for stress-testing REST clients.
- **/stream_payload**: Streams a large JSON array (10,000 objects) in chunked encoding, allowing clients to process data as it arrives.
- **Plugin Architecture**: Easily extend the server with new payload or streaming plugins by implementing the `PayloadPlugin` interface.

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

By default, the server listens on port 8080. You can test the endpoints:

#### Example Requests
- Huge payload:
  ```sh
  curl http://localhost:8080/huge_payload
  ```
- Streaming payload:
  ```sh
  curl http://localhost:8080/stream_payload
  ```

## Customization
- Adjust the payload size or structure in `huge_payload_handler.go` and `streaming_payload_handler.go` as needed for your testing scenario.
- Add new plugins by implementing the `PayloadPlugin` interface and registering them in `main.go`.

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
