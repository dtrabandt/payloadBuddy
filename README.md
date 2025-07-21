# gohugePayloadServer

A sophisticated Go server designed to test REST client implementations with large payloads and advanced streaming scenarios. This project is specifically tailored for ServiceNow consultants and developers who need to test and troubleshoot REST consumer behavior under various network and server conditions.

## Purpose

This server helps consultants and developers:
- Simulate REST endpoints that return very large JSON payloads
- Test streaming endpoints with configurable delays and patterns
- Reproduce ServiceNow-specific scenarios (peak hours, maintenance windows, etc.)
- Identify and troubleshoot issues with clients that cannot handle large responses
- Train on best practices for handling large data transfers in REST APIs

## Features

### 🚀 **Core Endpoints**
- **/huge_payload**: Returns a large JSON array (100,000 objects) in a single response for stress-testing REST clients
- **/stream_payload**: Advanced streaming endpoint with configurable delays, patterns, and ServiceNow simulation modes

### ⚙️ **Advanced Streaming Features**
- **Configurable Item Count**: 1 to 1,000,000 items
- **Delay Strategies**: Fixed, Random, Progressive, Burst patterns
- **ServiceNow Scenarios**: Peak hours, maintenance windows, network issues, database load
- **ServiceNow Mode**: Generates realistic ServiceNow record structures with sys_id, incident numbers, states
- **Context-Aware**: Handles client cancellation gracefully
- **Real-time Streaming**: Chunked transfer encoding with configurable flush intervals

### 🏗️ **Architecture**
- **Plugin System**: Easily extend with new payload handlers via `PayloadPlugin` interface
- **Separation of Concerns**: Each handler in its own file
- **Comprehensive Testing**: Unit tests for all scenarios and edge cases

## Use Cases

### ServiceNow Integration Testing
Perfect for testing ServiceNow REST integrations that might fail with large datasets:
- Flow Actions that timeout on large responses
- REST Message processing limits
- Memory constraints in ServiceNow instances
- Peak hour performance degradation

### Performance Testing
- Client timeout behavior under various delays
- Memory usage with large streaming responses
- Network interruption handling
- Progressive performance degradation simulation

## Getting Started

### Prerequisites
- Go 1.24.5 or newer

### Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/dtrabandt/gohugePayloadServer.git
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

The server listens on port 8080 and provides detailed startup information with example URLs.

## API Reference

### /huge_payload
Returns 100,000 JSON objects in a single response.

```sh
curl http://localhost:8080/huge_payload
```

### /stream_payload
Advanced streaming endpoint with multiple configuration options.

#### Query Parameters

| Parameter | Description | Default | Examples |
|-----------|-------------|---------|----------|
| `count` | Number of items to stream | 10000 | `count=1000` |
| `delay` | Base delay between items | 0 | `delay=100ms`, `delay=1s`, `delay=500` |
| `strategy` | Delay pattern | fixed | `fixed`, `random`, `progressive`, `burst` |
| `scenario` | ServiceNow scenario | none | `peak_hours`, `maintenance`, `network_issues`, `database_load` |
| `batch_size` | Items per flush | 100 | `batch_size=50` |
| `servicenow` | ServiceNow mode | false | `servicenow=true` |

#### Examples

**Basic streaming:**
```sh
curl "http://localhost:8080/stream_payload?count=1000"
```

**ServiceNow peak hours simulation:**
```sh
curl "http://localhost:8080/stream_payload?scenario=peak_hours&servicenow=true&count=500"
```

**Random delays for network testing:**
```sh
curl "http://localhost:8080/stream_payload?delay=200ms&strategy=random&count=200"
```

**Progressive performance degradation:**
```sh
curl "http://localhost:8080/stream_payload?delay=50ms&strategy=progressive&count=1000"
```

**Maintenance window with spikes:**
```sh
curl "http://localhost:8080/stream_payload?scenario=maintenance&count=2000"
```

**Burst pattern testing:**
```sh
curl "http://localhost:8080/stream_payload?delay=10ms&strategy=burst&batch_size=25"
```

## ServiceNow Testing Scenarios

### Peak Hours (`scenario=peak_hours`)
- Simulates slower response times during peak ServiceNow usage
- 200ms base delay between items
- Perfect for testing Flow Action timeouts

### Maintenance Window (`scenario=maintenance`)
- Simulates maintenance periods with periodic spikes
- 500ms base delay with 2s spikes every 500 items
- Tests resilience during ServiceNow maintenance windows

### Network Issues (`scenario=network_issues`)
- Random network delays and interruptions
- 10% chance of 0-3 second delays
- Simulates unstable network conditions

### Database Load (`scenario=database_load`)
- Progressive performance degradation
- Delay increases as more items are processed
- Simulates database performance issues under load

## Testing

Run the test suite:
```sh
go test -v
```

Tests cover:
- Basic functionality
- Parameter validation
- ServiceNow mode
- All delay strategies
- All scenarios
- Error conditions
- Performance expectations

## Development

### Adding New Plugins
1. Implement the `PayloadPlugin` interface:
   ```go
   type MyPlugin struct{}
   
   func (m MyPlugin) Path() string { return "/my_endpoint" }
   func (m MyPlugin) Handler() http.HandlerFunc { return MyHandler }
   ```

2. Register in `init()`:
   ```go
   func init() {
       registerPlugin(MyPlugin{})
   }
   ```

### Project Structure
```
├── main.go                           # Server setup and plugin registration
├── huge_payload_handler.go           # Large single-response endpoint
├── streaming_payload_handler.go      # Advanced streaming endpoint
├── *_test.go                        # Comprehensive test suite
├── README.md                        # This file
├── CHANGELOG.md                     # Version history
└── go.mod                           # Go module definition
```

## Best Practices for Large Payloads

### For Clients (ServiceNow, etc.)
- **Implement streaming parsers** instead of loading entire responses
- **Set appropriate timeouts** based on expected data sizes
- **Use pagination** when possible to reduce response sizes
- **Monitor memory usage** during large data processing
- **Implement retry logic** with exponential backoff

### For Servers
- **Use chunked transfer encoding** for real-time streaming
- **Implement client cancellation** handling
- **Add compression** (gzip) for large responses
- **Set reasonable limits** on response sizes
- **Monitor server resources** during load testing

### ServiceNow Specific
- **Flow Actions**: Set timeout values based on expected response times
- **REST Messages**: Consider using pagination for large datasets
- **Scheduled Jobs**: Use smaller batch sizes for processing
- **Performance**: Test during peak hours with realistic data volumes

## Contributing

Contributions are welcome! Areas for improvement:
- Additional ServiceNow scenarios
- More delay patterns
- Performance optimizations
- Additional output formats
- Monitoring and metrics

Please open issues or submit pull requests.

## License

MIT License

## Authors

- Dennis Trabandt

---

*This project is designed for educational and troubleshooting purposes. Use appropriate caution in production environments and always test with realistic data volumes.*