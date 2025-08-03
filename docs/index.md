---
layout: default
title: Home
---

<div align="center">
  <img src="assets/images/logo.svg" alt="PayloadBuddy Logo" width="200"/>
</div>

# PayloadBuddy

A sophisticated Go server designed to test REST client implementations with large payloads and advanced streaming scenarios. This project is specifically tailored for ServiceNow consultants and developers who need to test and troubleshoot REST consumer behavior under various network and server conditions.

## Quick Start

### Download and Run
Download the latest release for your platform from [GitHub Releases](https://github.com/dtrabandt/payloadBuddy/releases).

```bash
# Linux/macOS
tar -xzf payloadBuddy-vX.X.X-linux-amd64.tar.gz
./payloadBuddy

# With authentication
./payloadBuddy -auth
```

### Key Features

- **🚀 Large Payload Testing**: REST endpoint returning up to 1M JSON objects
- **📡 Advanced Streaming**: Configurable delays, patterns, and ServiceNow scenarios  
- **🔐 Security Features**: Optional HTTP Basic Authentication
- **📖 Interactive Documentation**: Built-in Swagger UI and OpenAPI specs
- **🏗️ Plugin Architecture**: Easily extensible with new endpoints

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/rest_payload` | Large JSON response (up to 1M objects) |
| `/stream_payload` | Advanced streaming with delays and scenarios |
| `/swagger` | Interactive API documentation |
| `/openapi.json` | OpenAPI 3.1.1 specification |

## ServiceNow Integration

PayloadBuddy is specifically designed for ServiceNow consultants and developers who need to test REST integrations under realistic conditions. Many integrations fail in production due to unexpected server behavior, network issues, or performance degradation that wasn't tested during development.

### Why These Scenarios Matter

ServiceNow environments experience various performance patterns that can break poorly designed integrations:

- **Flow Actions** may timeout during peak hours when response times increase
- **REST Messages** can fail during maintenance windows with sporadic delays  
- **Scheduled Jobs** might encounter progressive slowdowns as database load increases
- **Real-time integrations** need to handle network instability gracefully

### Testing Scenarios

#### 🚀 **Peak Hours Testing**
```bash
curl -u user:pass "http://localhost:8080/stream_payload?scenario=peak_hours&servicenow=true"
```
- **Simulates**: Slower response times during peak ServiceNow usage (200ms delays)
- **Tests**: Flow Action timeout handling, bulk data processing resilience
- **Real-world impact**: Prevents integration failures during business hours

#### 🔧 **Maintenance Window Testing**  
```bash
curl -u user:pass "http://localhost:8080/stream_payload?scenario=maintenance&count=2000"
```
- **Simulates**: Maintenance periods with periodic performance spikes (500ms + 2s spikes)
- **Tests**: Integration resilience during planned ServiceNow maintenance
- **Real-world impact**: Ensures integrations survive weekly maintenance windows

#### 🌐 **Network Issues Testing**
```bash  
curl -u user:pass "http://localhost:8080/stream_payload?scenario=network_issues&count=1000"
```
- **Simulates**: Random network delays and interruptions (10% chance of 0-3s delays)
- **Tests**: Retry logic, timeout handling, partial data recovery
- **Real-world impact**: Prevents data loss during network instability

#### 📊 **Database Load Testing**
```bash
curl -u user:pass "http://localhost:8080/stream_payload?scenario=database_load&count=5000"
```
- **Simulates**: Progressive performance degradation under increasing load
- **Tests**: Large dataset processing, memory management, timeout scaling
- **Real-world impact**: Ensures integrations work with growing data volumes

### ServiceNow-Specific Features

When using `servicenow=true`, PayloadBuddy generates realistic ServiceNow record structures:

```json
{
  "sys_id": "a1b2c3d4e5f6789012345678901234567890",
  "number": "INC0000123",
  "state": "2",
  "short_description": "Sample incident description",
  "created_on": "2024-01-15 10:30:00"
}
```

This helps test:
- **Field parsing** with actual ServiceNow field names and formats
- **Record relationships** using proper sys_id references  
- **State management** with realistic ServiceNow state values
- **Date handling** with ServiceNow's datetime format

## Documentation

- **[Deployment Guide](deployment)** - ngrok, Docker, and production deployments
- **[Contributing](contributing)** - Development workflow and TDD practices  
- **[Changelog](changelog)** - Version history and release notes

## Getting Help

- [GitHub Issues](https://github.com/dtrabandt/payloadBuddy/issues) - Bug reports and feature requests
- [GitHub Releases](https://github.com/dtrabandt/payloadBuddy/releases) - Download binaries

## License

PayloadBuddy is licensed under the [MIT License](https://github.com/dtrabandt/payloadBuddy/blob/main/LICENSE.md).

## Logo Attribution

The PayloadBuddy logo incorporates the Go Gopher, which is licensed under [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/). The original Go Gopher was created by [Renee French](https://reneefrench.blogspot.com/) and is used with attribution as required by the license.

**Original Go Gopher**: Created by Renee French, licensed under Creative Commons Attribution 4.0  
**Logo Design**: Dennis Trabandt (2025), incorporating the Go Gopher with proper attribution