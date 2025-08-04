<div align="center">
  <img src="docs/assets/images/logo.svg" alt="PayloadBuddy Logo" width="200"/>
</div>

# payloadBuddy

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
- **/rest_payload**: Returns a REST response with a large JSON array (100,000 objects) in a single response for stress-testing REST clients
- **/stream_payload**: Advanced streaming endpoint with configurable delays, patterns, and ServiceNow simulation modes
- **/openapi.json**: Complete OpenAPI 3.1.1 specification for all endpoints
- **/swagger**: Interactive Swagger UI for API documentation and testing

### 🔐 **Security Features**
- **Basic Authentication**: Optional HTTP Basic Authentication with CLI control
- **Auto-generated Credentials**: Automatic username/password generation when not specified
- **Secure Implementation**: Constant-time comparison to prevent timing attacks
- **Documentation Access**: API documentation endpoints (`/swagger`, `/openapi.json`) remain publicly accessible even when authentication is enabled

### ⚙️ **Advanced Streaming Features**
- **Configurable Item Count**: 1 to 1,000,000 items
- **Delay Strategies**: Fixed, Random, Progressive, Burst patterns
- **ServiceNow Scenarios**: Peak hours, maintenance windows, network issues, database load
- **ServiceNow Mode**: Generates realistic ServiceNow record structures with sys_id, incident numbers, states
- **Context-Aware**: Handles client cancellation gracefully
- **Real-time Streaming**: Chunked transfer encoding with configurable flush intervals

### 📋 **Configurable Scenario System**
- **Dynamic Scenario Loading**: JSON-based scenario configuration with comprehensive schema validation
- **Embedded Scenarios**: Core scenarios built into the binary for single-executable deployment
- **User-Defined Scenarios**: Custom scenarios in `$HOME/.config/payloadBuddy/scenarios/` directory
- **Scenario Override**: User scenarios override embedded scenarios with same `scenario_type`
- **Automatic Directory Creation**: User scenario directory created automatically on first run
- **Version Compatibility**: Built-in version compatibility checking framework
- **Real-time Configuration**: Scenario-based defaults for count, batch_size, and ServiceNow mode

### 🏗️ **Architecture**
- **Plugin System**: Easily extend with new payload handlers via `PayloadPlugin` interface
- **OpenAPI 3.1.1 Integration**: Automatic documentation generation from plugin specifications
- **Separation of Concerns**: Each handler in its own file with self-documenting capabilities
- **Comprehensive Testing**: Unit tests for all scenarios, edge cases, and API documentation

### 🚀 **CI/CD & Releases**
- **Automated Testing**: GitHub Actions run tests on every PR and push to develop
- **Cross-Platform Builds**: Automatic releases for Linux, macOS, and Windows (amd64 + arm64)
- **Quality Gates**: Code coverage (80%+), linting, security scanning, and formatting checks
- **Git-flow Integration**: Seamless workflow with feature branches, releases, and automated deployments
- **Semantic Versioning**: Professional releases with changelogs and checksums

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
- Go 1.21 or newer (for building from source)

### Installation Options

#### Option 1: Download Pre-built Binaries (Recommended)
Download the latest release for your platform from [GitHub Releases](https://github.com/dtrabandt/payloadBuddy/releases):

- **Linux**: `payloadBuddy-vX.X.X-linux-amd64.tar.gz` or `payloadBuddy-vX.X.X-linux-arm64.tar.gz`
- **macOS**: `payloadBuddy-vX.X.X-darwin-amd64.tar.gz` or `payloadBuddy-vX.X.X-darwin-arm64.tar.gz`
- **Windows**: `payloadBuddy-vX.X.X-windows-amd64.zip` or `payloadBuddy-vX.X.X-windows-arm64.zip`

Extract and run:
```sh
# Linux/macOS
tar -xzf payloadBuddy-vX.X.X-linux-amd64.tar.gz
./payloadBuddy

# Windows
# Extract the .zip file and run payloadBuddy.exe
```

#### Option 2: Build from Source
1. Clone the repository:
   ```sh
   git clone https://github.com/dtrabandt/payloadBuddy.git
   cd payloadBuddy
   ```
2. Build the server:
   ```sh
   go build -o payloadBuddy
   ```

#### Option 3: Install with Go
```sh
go install github.com/dtrabandt/payloadBuddy@latest
```

### Usage

#### Basic Usage (No Authentication)
Run the server:
```sh
./payloadBuddy
```

#### With Authentication
Enable basic authentication:
```sh
# Auto-generate username and password
./payloadBuddy -auth

# Use specific credentials
./payloadBuddy -auth -user=myuser -pass=mypass
```

**Note**: When authentication is enabled, API endpoints (`/rest_payload`, `/stream_payload`) require credentials, but documentation endpoints (`/swagger`, `/openapi.json`) remain publicly accessible for better developer experience.

#### Command Line Options
For a complete list of available options, use the built-in help:
```sh
./payloadBuddy -h
# or
./payloadBuddy --help
```

**Available options:**
- `-port=<port>`: Set the HTTP server port (default: 8080)
- `-auth`: Enable basic authentication (default: false)
- `-user=<username>`: Set username (auto-generated if not specified)
- `-pass=<password>`: Set password (auto-generated if not specified)
- `-verify=<file>`: Validate a scenario file against the JSON schema and exit

The server listens on the specified port (default: 8080) and provides detailed startup information with example URLs and authentication details.

## Deployment Options

For production use or external access, see the **[DEPLOYMENT.md](DEPLOYMENT.md)** guide which covers:

### 🚀 **Quick External Access**
- **ngrok Integration**: Expose your local server to ServiceNow instances
- Perfect for demos, training, and rapid prototyping
- Easy setup with authentication and custom domains

### 🐳 **Production-Like Environment** 
- **Docker + MID-Server**: Complete containerized environment
- ServiceNow MID-Server integration for secure connections
- Load balancing with nginx and horizontal scaling
- Enterprise-ready with security best practices

### 📋 **What's Included**
- Step-by-step setup instructions
- ServiceNow configuration examples (Flow Actions, REST Messages)
- Security considerations and best practices
- Troubleshooting guides and performance optimization
- Advanced configurations including Kubernetes deployment

## API Reference

> 📖 **Interactive Documentation**: Visit `/swagger` in your browser for a complete interactive API explorer with request/response examples and the ability to test endpoints directly.

> 🔧 **OpenAPI Specification**: The complete OpenAPI 3.1.1 specification is available at `/openapi.json` for programmatic access and integration with tools like Postman, Insomnia, or code generators.

### /rest_payload
Returns 100,000 JSON objects in a single response (default, configurable via `count` parameter).

**Without Authentication:**
```sh
curl http://localhost:8080/rest_payload
```

**With Authentication:**
```sh
curl -u username:password http://localhost:8080/rest_payload
```

### /stream_payload
Advanced streaming endpoint with multiple configuration options.

#### Query Parameters

| Parameter | Description | Default | Examples |
|-----------|-------------|---------|----------|
| `count` | Number of items to stream | 10000 | `count=1000` |
| `delay` | Base delay between items | 10 | `delay=100ms`, `delay=1s`, `delay=500` |
| `strategy` | Delay pattern | fixed | `fixed`, `random`, `progressive`, `burst` |
| `scenario` | ServiceNow scenario | none | `peak_hours`, `maintenance`, `network_issues`, `database_load` |
| `batch_size` | Items per flush | 100 | `batch_size=50` |
| `servicenow` | ServiceNow mode | false | `servicenow=true` |

#### Examples

**Basic streaming (no auth):**
```sh
curl "http://localhost:8080/stream_payload?count=1000"
```

**Basic streaming (with auth):**
```sh
curl -u username:password "http://localhost:8080/stream_payload?count=1000"
```

**ServiceNow peak hours simulation:**
```sh
curl -u username:password "http://localhost:8080/stream_payload?scenario=peak_hours&servicenow=true&count=500"
```

**Random delays for network testing:**
```sh
curl -u username:password "http://localhost:8080/stream_payload?delay=200ms&strategy=random&count=200"
```

**Progressive performance degradation:**
```sh
curl -u username:password "http://localhost:8080/stream_payload?delay=50ms&strategy=progressive&count=1000"
```

**Maintenance window with spikes:**
```sh
curl -u username:password "http://localhost:8080/stream_payload?scenario=maintenance&count=2000"
```

**Burst pattern testing:**
```sh
curl -u username:password "http://localhost:8080/stream_payload?delay=10ms&strategy=burst&batch_size=25"
```

### /openapi.json
Returns the complete OpenAPI 3.1.1 specification for all endpoints.

**Example:**
```sh
# Get the OpenAPI specification (no authentication required)
curl http://localhost:8080/openapi.json

# Save it to a file for tools like Postman
curl http://localhost:8080/openapi.json > payloadBuddy-api.json
```

**Note**: This endpoint is always publicly accessible, even when authentication is enabled.

### /swagger
Interactive Swagger UI for exploring and testing the API.

**Usage:**
1. Start the server: `./payloadBuddy` or `./payloadBuddy -auth`
2. Open your browser: `http://localhost:8080/swagger` (no authentication required)
3. Explore endpoints, view schemas, and test requests directly in your browser
4. When authentication is enabled, use the "Authorize" button in Swagger UI to enter credentials for testing protected endpoints

**Note**: The Swagger UI is always publicly accessible, even when authentication is enabled. This allows you to explore the API documentation and then authenticate within Swagger UI to test protected endpoints.

> **Note:** Replace `username:password` with your actual credentials when authentication is enabled.

## ServiceNow Testing Scenarios

PayloadBuddy includes four built-in scenarios designed to simulate real ServiceNow conditions:

- **Peak Hours** (`scenario=peak_hours`): 200ms delays simulating peak usage
- **Maintenance Window** (`scenario=maintenance`): 500ms delays with 2s spikes
- **Network Issues** (`scenario=network_issues`): Random delays up to 3s
- **Database Load** (`scenario=database_load`): Progressive performance degradation

### Custom Scenario Configuration

PayloadBuddy supports user-defined scenarios through JSON configuration files with comprehensive schema validation, automatic loading, and override capabilities.

> 📖 **Complete Scenario Guide**: For detailed information about creating custom scenarios, JSON schema reference, advanced features, and troubleshooting, see **[SCENARIOS.md](SCENARIOS.md)**.

#### Quick Example

Create `$HOME/.config/payloadBuddy/scenarios/my-test.json`:

```json
{
    "schema_version": "1.0.0",
    "scenario_name": "Quick Test",
    "scenario_type": "custom",
    "base_delay": "100ms",
    "delay_strategy": "progressive",
    "servicenow_mode": true,
    "batch_size": 50
}
```

**Validation:**
```sh
# Validate your scenario file before deploying
./payloadBuddy -verify $HOME/.config/payloadBuddy/scenarios/my-test.json
```

**Usage:**
```sh
curl -u username:password "http://localhost:8080/stream_payload?scenario=custom"
```

### Key Features

- **📁 Dynamic Loading**: Scenarios loaded from `$HOME/.config/payloadBuddy/scenarios/`
- **🔄 Override Support**: User scenarios override built-in scenarios
- **✅ Schema Validation**: Comprehensive JSON schema validation
- **📊 Embedded Scenarios**: Core scenarios built into binary
- **🔧 Advanced Configuration**: Error injection, performance monitoring, custom timing patterns

## Testing

```sh
go test -v ./...                  # Run all tests
go test -v -cover ./...           # Run with coverage report
```

The test suite covers all endpoints, authentication scenarios, delay strategies, ServiceNow simulations, and edge cases. See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed testing guidelines.

## Development

PayloadBuddy uses an extensible plugin architecture. New endpoints can be added by implementing the `PayloadPlugin` interface and automatically appear in the OpenAPI specification and Swagger UI.

For detailed development guidelines, plugin creation, and contribution workflow, see [CONTRIBUTING.md](CONTRIBUTING.md).

### Project Structure
```
├── main.go                          # Server setup and plugin registration
├── *_payload_handler.go             # Endpoint implementations
├── auth.go                          # Authentication middleware
├── documentation_handler.go         # OpenAPI spec and Swagger UI
├── scenario_manager.go              # Dynamic scenario loading and management
├── scenario_validator.go            # JSON schema validation for scenarios
├── scenarios/                       # Embedded scenario JSON files and schema
│   ├── *.json                       # Built-in scenario configurations
│   └── scenario_schema_v1.0.0.json  # JSON schema for validation
├── *_test.go                        # Comprehensive test suite
├── .github/workflows/               # CI/CD automation
├── README.md                        # User documentation
├── CONTRIBUTING.md                  # Development guidelines
└── DEPLOYMENT.md                    # Deployment strategies
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

## AI Usage Disclosure

This project was developed with assistance from AI coding tools:

- **GitHub Copilot** - Used for code completion and boilerplate generation
- **Claude Code** - Used for documentation, refactoring, and code review

All AI-generated code has been reviewed, tested, and validated by human developers. The use of these tools helped accelerate development and improve code quality, particularly for Go language learning and comprehensive documentation.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:

- Development workflow and git-flow practices
- Code quality standards and formatting requirements
- Testing guidelines and TDD approach
- CI/CD pipeline and quality gates
- Project structure and plugin architecture

For bugs and feature requests, please open an issue on [GitHub](https://github.com/dtrabandt/payloadBuddy/issues).

## Releases

This project uses **Semantic Versioning** and automated releases:

- **Development**: Work happens on the `develop` branch
- **Releases**: Merges to `main` trigger automatic builds for all platforms
- **Versioning**: Git tags (e.g., `v1.0.0`) create GitHub releases with binaries
- **Platforms**: Linux, macOS, Windows (both amd64 and arm64 architectures)
- **Artifacts**: Compressed binaries with SHA256 checksums for integrity verification

See the [Releases page](https://github.com/dtrabandt/payloadBuddy/releases) for download links and changelogs.

## License

MIT License

## Logo Attribution

The PayloadBuddy logo incorporates the Go Gopher, which is licensed under [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/). The original Go Gopher was created by [Renee French](https://reneefrench.blogspot.com/) and is used with attribution as required by the license. The logo design was created by Dennis Trabandt with assistance from OpenAI ChatGPT 4.1.

**Original Go Gopher**: Created by Renee French, licensed under Creative Commons Attribution 4.0  
**Logo Design**: Dennis Trabandt (2025), incorporating the Go Gopher with proper attribution

## Authors

- Dennis Trabandt

---

*This project is designed for educational and troubleshooting purposes. Use appropriate caution in production environments and always test with realistic data volumes.*