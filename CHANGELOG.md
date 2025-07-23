# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Advanced Streaming Features**:
  - Configurable delay strategies (fixed, random, progressive, burst)
  - ServiceNow-specific test scenarios (peak_hours, maintenance, network_issues, database_load)
  - ServiceNow mode with realistic record structures (sys_id, incident numbers, states)
  - Context-aware streaming with client cancellation support
  - Configurable batch sizes for flushing
  - Parameter validation and error handling

- **HugePayloadHandler**: Added support for `count` query parameter (default 10,000, up to 1,000,000)

- **Enhanced Testing**:
  - Comprehensive test suite for streaming handler
  - Tests for all delay strategies and ServiceNow scenarios
  - Parameter validation tests
  - Performance and timing tests
  - Test for huge_payload count parameter

- **Improved Documentation**:
  - Complete API reference with examples
  - ServiceNow-specific use cases and best practices
  - Detailed startup messages with example URLs
  - Performance and troubleshooting guidelines

- **Developer Experience**:
  - Random seed initialization for consistent testing
  - Better error messages and validation
  - Detailed logging of registered endpoints
  - Enhanced project structure documentation

### Changed
- **StreamingPayloadHandler**: Complete rewrite with advanced features
  - Added query parameter support (count, delay, strategy, scenario, batch_size, servicenow)
  - Improved JSON streaming with proper array formatting
  - Enhanced error handling and graceful shutdown
  - Better memory efficiency and performance

- **Main Application**: 
  - Enhanced startup messages with all available endpoints
  - Added random seed initialization
  - Improved endpoint registration logging

- **Testing Framework**:
  - Expanded test coverage from basic functionality to comprehensive scenario testing
  - Added performance and parameter validation tests

### Fixed
- **JSON Format**: Fixed invalid JSON output in streaming handler
- **Memory Efficiency**: Eliminated unnecessary buffer allocations
- **Streaming Performance**: Added proper flushing for real-time streaming
- **Error Handling**: Improved error responses and client disconnect handling

## [Previous Releases]

### Added (Previous)
- Refactorings to add a unique number to the name and to correct the id's
- Rename of the function to HugePayloadHandler to support more functionality in the future
- Added plugin architecture for payload handlers (PayloadPlugin interface)
- Added /stream_payload endpoint for streaming large JSON arrays
- Refactored StreamingPayloadHandler into its own file
- Improved documentation in README.md and handler files
- Added tests for streaming payload handler
- Initial commit with basic huge payload functionality

### Technical Improvements
- **Architecture**: Clean separation of concerns with plugin system
- **Maintainability**: Each handler in separate files for better organization
- **Extensibility**: Easy to add new payload types via PayloadPlugin interface
- **Testing**: Comprehensive test coverage for all components

### ServiceNow Integration Focus
- **Use Case Alignment**: Specifically designed for ServiceNow REST testing scenarios
- **Realistic Simulations**: Authentic ServiceNow record formats and performance patterns
- **Consultant Tools**: Ready-to-use examples and scenarios for field testing
- **Production Readiness**: Robust error handling and performance optimization

---

### Notes for Contributors
- All new features should include comprehensive tests
- ServiceNow-specific scenarios should be documented with real-world use cases
- Performance changes should be validated with benchmarks
- Breaking changes require major version bump