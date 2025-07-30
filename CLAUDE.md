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
3. `basicAuthMiddleware()` wraps API endpoints (excludes documentation endpoints)
4. Credentials displayed on startup for development use
5. API endpoints protected when `-auth` flag is used (documentation endpoints remain public)

### Authentication Exclusions
- **Documentation endpoints are public**: `/swagger` and `/openapi.json` are excluded from authentication
- **API endpoints require auth**: `/rest_payload` and `/stream_payload` require authentication when `-auth` is enabled
- **Rationale**: Standard practice to keep API documentation publicly accessible while protecting data endpoints
- **Implementation**: Conditional middleware application in main.go based on endpoint path

### Testing Strategy
Tests are structured to handle both authenticated and non-authenticated scenarios:
- Set `*enableAuth = false` to disable auth in tests
- Use `basicAuthMiddleware()` wrapper for testing auth scenarios
- Comprehensive coverage of all delay strategies, scenarios, and parameter combinations
- OpenAPI specification testing ensures all endpoints are properly documented
- Swagger UI functionality is validated through automated tests

## Test-Driven Development (TDD) Workflow

This project follows TDD practices and Claude Code is well-suited for TDD workflows:

### TDD Cycle with Claude Code
```bash
# 1. RED: Write failing test first
go test -v -run TestNewFeature     # Should fail - feature doesn't exist yet

# 2. GREEN: Write minimal code to make test pass
# Implement just enough to make the test pass

# 3. REFACTOR: Improve code while keeping tests green
go test -v                         # Ensure all tests still pass
```

### TDD Best Practices for This Project

#### 1. **Test-First Development**
- Always write tests before implementing new features
- Start with the simplest failing test
- Use table-driven tests for multiple scenarios
- Example pattern:
```go
func TestNewDelayStrategy(t *testing.T) {
    tests := []struct {
        name     string
        strategy string
        expected DelayStrategy
    }{
        {"exponential strategy", "exponential", ExponentialDelay},
        {"invalid strategy", "invalid", FixedDelay},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation here
        })
    }
}
```

#### 2. **Fast Feedback Loop**
- Run tests frequently: `go test -v`
- Run specific tests: `go test -v -run TestSpecificFunction`
- Use `go test -v ./...` for full project coverage
- Tests should complete quickly (< 5 seconds for full suite)

#### 3. **Comprehensive Test Coverage**
- Test happy paths, edge cases, and error conditions
- Include boundary value testing (e.g., count=0, count=1000000)
- Test authentication scenarios (with/without auth)
- Validate HTTP status codes, headers, and response formats
- Test parameter validation and error handling

#### 4. **Refactoring Safety**
- Only refactor when all tests are green
- Run tests after each refactoring step
- Use tests as documentation of expected behavior
- Maintain test coverage during refactoring

### TDD Workflow Examples

#### Adding a New Delay Strategy
```bash
# 1. RED: Write failing test
# Add test for ExponentialDelay in streaming_payload_handler_test.go

# 2. GREEN: Minimal implementation
# Add ExponentialDelay constant and case in getDelayStrategy()

# 3. REFACTOR: Improve implementation
# Add proper exponential delay calculation in applyDelay()
```

#### Adding New Endpoint
```bash
# 1. RED: Write plugin test
# Create test for new plugin implementing PayloadPlugin interface

# 2. GREEN: Basic plugin implementation
# Implement minimal Path(), Handler(), OpenAPISpec() methods

# 3. REFACTOR: Complete implementation
# Add full handler logic, parameter validation, OpenAPI documentation
```

### Testing Guidelines

#### **Unit Tests**
- Test individual functions in isolation
- Mock external dependencies if needed
- Focus on business logic and edge cases
- Fast execution (no network calls, file I/O)

#### **Integration Tests**
- Test HTTP handlers end-to-end
- Use `httptest.NewRecorder()` for HTTP testing
- Test middleware integration (authentication)
- Validate JSON response structures

#### **OpenAPI Testing**
- Validate OpenAPI specification structure
- Test that all endpoints are documented
- Verify parameter definitions match implementation
- Check security scheme configuration

### TDD Commands Reference
```bash
# Development cycle
go test -v                           # Run all tests
go test -v -run TestSpecific         # Run specific test pattern
go test -v -short                    # Skip long-running tests
go test -v -cover                    # Show test coverage

# Continuous testing (with external tools)
# Install: go install github.com/cosmtrek/air@latest
air                                  # Auto-restart on file changes

# Test with race detection
go test -v -race                     # Detect race conditions
```

### TDD Benefits in This Project
- **Rapid iteration**: Fast feedback on new features
- **Regression prevention**: Catch breaking changes immediately  
- **Documentation**: Tests serve as executable documentation
- **Refactoring confidence**: Safe to improve code structure
- **Quality assurance**: Edge cases and error conditions covered