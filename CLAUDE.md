# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
```bash
go build -o payloadBuddy        # Build the application
./payloadBuddy                  # Run without authentication
./payloadBuddy -auth            # Run with auto-generated credentials
./payloadBuddy -auth -user=admin -pass=secret  # Run with custom credentials
```

### Testing
```bash
go test -v                             # Run all tests with verbose output
go test -v -run TestRestPayloadHandler # Run specific test pattern
go test -v ./...                       # Run tests recursively (single package project)
```

### Development Workflow
```bash
go mod tidy                            # Clean up dependencies
go build && go test -v                 # Build and test in sequence
```

## Architecture Overview

### Plugin-Based Architecture
The server uses a plugin system where endpoints are registered via the `PayloadPlugin` interface:
- Each handler (rest_payload_handler.go, streaming_payload_handler.go, documentation_handler.go) implements `PayloadPlugin`
- Plugins self-register in their `init()` functions using `registerPlugin()`
- Main server automatically discovers and registers all plugins with authentication middleware
- Each plugin provides its own OpenAPI specification via the `OpenAPISpec()` method

### Core Components

**main.go**: Server bootstrap, plugin registration, and HTTP server setup
- Manages the plugin registry and applies authentication middleware to all endpoints
- Handles command-line flag parsing and server startup messaging

**auth.go**: HTTP Basic Authentication system
- Provides `basicAuthMiddleware()` that wraps all endpoints
- Handles credential generation, validation, and display
- Uses constant-time comparison for security against timing attacks
- Authentication is optional (controlled via `-auth` flag)

**rest_payload_handler.go**: Single large response endpoint (`/rest_payload`)
- Returns configurable number of JSON objects (default 10,000, max 1,000,000)
- Uses `count` query parameter for customization

**streaming_payload_handler.go**: Advanced streaming endpoint (`/stream_payload`)
- Real-time JSON streaming with chunked transfer encoding
- Supports multiple delay strategies (fixed, random, progressive, burst)
- ServiceNow-specific simulation scenarios (peak_hours, maintenance, network_issues, database_load)
- Configurable via query parameters: count, delay, strategy, scenario, batch_size, servicenow

**documentation_handler.go**: OpenAPI 3.1.1 specification and Swagger UI endpoints
- `/openapi.json`: Complete OpenAPI specification for all endpoints
- `/swagger`: Interactive Swagger UI for API documentation and testing
- Automatic collection of specifications from all registered plugins
- CORS-enabled for cross-origin access to OpenAPI specification

### ServiceNow Integration Focus
This server is specifically designed for ServiceNow REST integration testing:
- ServiceNow mode generates realistic record structures (sys_id, incident numbers, states)
- Scenarios simulate real ServiceNow performance characteristics
- Examples in startup output use curl format for easy ServiceNow Flow Action integration

### Authentication Flow
1. Command-line flags parsed in main()
2. `setupAuthentication()` configures credentials (auto-generated or custom)
3. `basicAuthMiddleware()` wraps all plugin handlers
4. Credentials displayed on startup for development use
5. All endpoints protected when `-auth` flag is used

### Testing Strategy
Tests are structured to handle both authenticated and non-authenticated scenarios:
- Set `*enableAuth = false` to disable auth in tests
- Use `basicAuthMiddleware()` wrapper for testing auth scenarios
- Comprehensive coverage of all delay strategies, scenarios, and parameter combinations
- OpenAPI specification testing ensures all endpoints are properly documented
- Swagger UI functionality is validated through automated tests