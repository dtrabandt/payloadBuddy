# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [v0.2.0] - 2025-07-31

### Added

- **README**:
  - Added some lines about `flags` functionality to display a usage menu upon utilizing `-h` or `--help`

- **Port Selection Feature**:
  - Added `-port` command-line parameter to allow users to specify the HTTP server port (default: 8080)

### Changed

- Removed old leftover comments from Go 1.20.0 behaviour. Removed it as the code was not existing anymore.

### Fixed

## [v0.1.0] - 2025-07-31

### Added
- **OpenAPI 3.1.1 Integration**:
  - Complete OpenAPI specification generation at `/openapi.json`
  - Interactive Swagger UI at `/swagger` endpoint
  - Automatic documentation collection from all registered plugins
  - CORS-enabled OpenAPI endpoint for cross-origin access
  - Self-documenting plugin architecture with `OpenAPISpec()` method
  - Comprehensive parameter documentation with examples and constraints
  - Authentication information included in specifications when enabled

- **Documentation Enhancements**:
  - Interactive API exploration through Swagger UI
  - Programmatic API specification access for tools like Postman/Insomnia
  - Detailed parameter descriptions with ServiceNow-specific examples
  - Complete request/response schema definitions
  - OpenAPI-compliant error response documentation
- **Basic Authentication System**:
  - HTTP Basic Authentication middleware with CLI control
  - `-auth` flag to enable/disable authentication
  - `-user` and `-pass` flags for custom credentials
  - Auto-generation of secure username/password when not specified
  - Constant-time comparison to prevent timing attacks
  - Clear credential display on server startup
  - Authentication status in endpoint examples

- **Advanced Streaming Features**:
  - Configurable delay strategies (fixed, random, progressive, burst)
  - ServiceNow-specific test scenarios (peak_hours, maintenance, network_issues, database_load)
  - ServiceNow mode with realistic record structures (sys_id, incident numbers, states)
  - Context-aware streaming with client cancellation support
  - Configurable batch sizes for flushing
  - Parameter validation and error handling

- **RestPayloadHandler**:
  - Renamed the endpoint from HugePayload to RestPayload as well the url from /huge_payload to /rest_payload
  - Added support for `count` query parameter (default 10,000, up to 1,000,000)

- **Enhanced Testing**:
  - Comprehensive test suite for streaming handler
  - Tests for all delay strategies and ServiceNow scenarios
  - Parameter validation tests
  - Performance and timing tests
  - Test for huge_payload count parameter
  - Authentication middleware tests (with/without credentials)
  - Security edge case testing
  - OpenAPI specification validation tests
  - Swagger UI functionality tests
  - Plugin interface compliance testing

- **Improved Documentation**:
  - Complete API reference with examples
  - ServiceNow-specific use cases and best practices
  - Detailed startup messages with example URLs
  - Performance and troubleshooting guidelines
  - Authentication setup and usage documentation
  - Command-line options reference

- **Developer Experience**:
  - Random seed initialization for consistent testing
  - Better error messages and validation
  - Detailed logging of registered endpoints
  - Enhanced project structure documentation
  - Refactored authentication code into separate `auth.go` file
  - Improved code organization and maintainability

- **CI/CD Pipeline**:
  - GitHub Actions workflow for automated testing on PRs and pushes to develop
  - Automated release builds triggered by git tags (semantic versioning)
  - Cross-platform binary builds (Linux, macOS, Windows) for both amd64 and arm64
  - Comprehensive test suite including unit tests, linting, security scanning
  - Test coverage enforcement (minimum 80% coverage required)
  - Code quality checks with golangci-lint and gosec security scanner
  - Automated GitHub releases with compressed binaries and checksums
  - Git-flow integration with branch protection and automated workflows
  - Version embedding in binaries for better traceability

### Changed
- **PayloadPlugin Interface**: Extended to include OpenAPI specification generation
  - All plugins now implement `OpenAPISpec()` method for automatic documentation
  - Backward compatibility maintained with existing handler functionality
  - Enhanced plugin registration to include documentation endpoints

- **StreamingPayloadHandler**: Complete rewrite with advanced features
  - Added query parameter support (count, delay, strategy, scenario, batch_size, servicenow)
  - Improved JSON streaming with proper array formatting
  - Enhanced error handling and graceful shutdown
  - Better memory efficiency and performance
  - Comprehensive OpenAPI specification with all parameters documented

- **Main Application**: 
  - Enhanced startup messages with all available endpoints (including new documentation endpoints)
  - Added random seed initialization
  - Improved endpoint registration logging  
  - Command-line flag parsing and validation
  - Authentication status display and credential output
  - Refactored authentication logic into separate module for better organization
  - Automatic registration of documentation plugins alongside payload handlers

- **Testing Framework**:
  - Expanded test coverage from basic functionality to comprehensive scenario testing
  - Added performance and parameter validation tests

### Fixed
- **JSON Format**: Fixed invalid JSON output in streaming handler
- **Memory Efficiency**: Eliminated unnecessary buffer allocations
- **Streaming Performance**: Added proper flushing for real-time streaming
- **Error Handling**: Improved error responses and client disconnect handling

### Notes for Contributors
- All new features should include comprehensive tests
- ServiceNow-specific scenarios should be documented with real-world use cases
- Performance changes should be validated with benchmarks
- Breaking changes require major version bump

[Unreleased]: https://github.com/dennistrabandt/payloadBuddy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dennistrabandt/payloadBuddy/releases/tag/v0.1.0