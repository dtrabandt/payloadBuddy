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
curl http://localhost:8080/payload
```

## Customization
- Adjust the payload size or structure in `handler.go` as needed for your testing scenario.
- Modify endpoints or add new ones to simulate different REST behaviors.

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
