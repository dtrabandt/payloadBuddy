# PayloadBuddy Scenario Configuration Guide

This guide covers PayloadBuddy's configurable scenario system, which allows you to create custom testing scenarios through JSON configuration files.

## Table of Contents

- [Overview](#overview)
- [Built-in Scenarios](#built-in-scenarios)
- [Custom Scenario Configuration](#custom-scenario-configuration)
- [JSON Schema Reference](#json-schema-reference)
- [Advanced Features](#advanced-features)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

PayloadBuddy uses a sophisticated scenario system that combines:

- **Embedded Scenarios**: Core scenarios built into the binary for immediate use
- **Dynamic Loading**: User-defined scenarios loaded from `$HOME/.config/payloadBuddy/scenarios/`
- **Override Support**: User scenarios override built-in scenarios with the same `scenario_type`
- **Schema Validation**: Comprehensive JSON schema validation for all scenarios
- **Version Compatibility**: Built-in compatibility checking framework
- **Real-time Configuration**: Scenario-based defaults for count, batch_size, and ServiceNow mode

### How It Works

1. **Startup**: PayloadBuddy loads embedded scenarios first
2. **User Directory**: Scans `$HOME/.config/payloadBuddy/scenarios/` for `.json` files
3. **Validation**: Each scenario is validated against the comprehensive JSON schema
4. **Override**: User scenarios with matching `scenario_type` override embedded ones
5. **Logging**: Detailed logging shows which scenarios are loaded and any validation errors

## Built-in Scenarios

PayloadBuddy includes four core scenarios embedded in the binary. **All scenarios work with both streaming (`/stream_payload`) and pagination (`/paginated_payload`) endpoints**, adapting their behavior for each context.

### Peak Hours (`scenario=peak_hours`) - **Ideal for Both**
- **Purpose**: Simulates slower response times during peak ServiceNow usage
- **Streaming Behavior**: 200ms fixed delay between each item in the stream
- **Pagination Behavior**: 200ms fixed delay before returning each page
- **Use Case**: Testing Flow Actions, Data Stream actions, and bulk data processing resilience
- **ServiceNow Mode**: Enabled by default

**Examples:**
```bash
# Streaming endpoint
curl -u user:pass "http://localhost:8080/stream_payload?scenario=peak_hours"

# Pagination endpoint  
curl -u user:pass "http://localhost:8080/paginated_payload?scenario=peak_hours&limit=100"
```

### Maintenance Window (`scenario=maintenance`) - **Works with Both**
- **Purpose**: Simulates maintenance periods with periodic performance spikes
- **Streaming Behavior**: 500ms base delay with 2-second spikes every 500 items processed
- **Pagination Behavior**: Single delay spike applied to each page request
- **Use Case**: Testing integration resilience during ServiceNow maintenance windows
- **ServiceNow Mode**: Enabled by default

**Examples:**
```bash
# Streaming endpoint (spikes during processing)
curl -u user:pass "http://localhost:8080/stream_payload?scenario=maintenance&count=2000"

# Pagination endpoint (spike per page)
curl -u user:pass "http://localhost:8080/paginated_payload?scenario=maintenance&limit=100&page=1"
```

### Network Issues (`scenario=network_issues`) - **Works with Both**
- **Purpose**: Simulates random network delays and interruptions
- **Streaming Behavior**: 10% chance of 0-3 second random delays per item
- **Pagination Behavior**: Random delays applied to each page request
- **Use Case**: Testing retry logic, timeout handling, and partial data recovery
- **ServiceNow Mode**: Disabled by default (focuses on network simulation)

**Examples:**
```bash
# Streaming endpoint (random delays per item)
curl -u user:pass "http://localhost:8080/stream_payload?scenario=network_issues&count=1000"

# Pagination endpoint (random delays per page)
curl -u user:pass "http://localhost:8080/paginated_payload?scenario=network_issues&limit=50&offset=0"
```

### Database Load (`scenario=database_load`) - **Works with Both**
- **Purpose**: Simulates progressive performance degradation under increasing load
- **Streaming Behavior**: Delay increases by 10ms for every 100 items processed (25ms base + progressive)
- **Pagination Behavior**: Single delay applied per page (calculated based on offset/page position)
- **Use Case**: Testing large dataset processing and memory management
- **ServiceNow Mode**: Enabled by default

**Examples:**
```bash
# Streaming endpoint (progressive delays per item)
curl -u user:pass "http://localhost:8080/stream_payload?scenario=database_load&count=5000"

# Pagination endpoint (delay based on page position)
curl -u user:pass "http://localhost:8080/paginated_payload?scenario=database_load&limit=100&offset=500"
```

## Custom Scenario Configuration

### Getting Started

1. **Automatic Setup**: PayloadBuddy creates `$HOME/.config/payloadBuddy/scenarios/` on first run
2. **Create Scenarios**: Add `.json` files to define your custom scenarios
3. **Validate Scenarios**: Use `./payloadBuddy -verify <file>` to validate before deployment
4. **Schema Validation**: All scenarios are automatically validated at startup
5. **Immediate Use**: Custom scenarios are available immediately after creation

### Directory Structure

```
$HOME/.config/payloadBuddy/scenarios/
├── my-custom-scenario.json
├── high-load-test.json
├── client-specific-test.json
└── override-peak-hours.json
```

### Basic Example

Create `$HOME/.config/payloadBuddy/scenarios/basic-test.json`:

```json
{
    "schema_version": "1.0.0",
    "scenario_name": "Basic Custom Test",
    "description": "A simple custom scenario for testing",
    "scenario_type": "custom",
    "base_delay": "100ms",
    "delay_strategy": "fixed",
    "servicenow_mode": true,
    "batch_size": 50,
    "response_limits": {
        "max_count": 10000,
        "default_count": 1000
    }
}
```

**Validation:**
```bash
# Validate the scenario file
./payloadBuddy -verify $HOME/.config/payloadBuddy/scenarios/basic-test.json
```

**Usage:**
```bash
# Works with both endpoints
curl -u user:pass "http://localhost:8080/stream_payload?scenario=custom"
curl -u user:pass "http://localhost:8080/paginated_payload?scenario=custom&limit=50"
```

### Override Example

Override the built-in `peak_hours` scenario by creating `$HOME/.config/payloadBuddy/scenarios/custom-peak-hours.json`:

```json
{
    "schema_version": "1.0.0",
    "scenario_name": "Custom Peak Hours",
    "description": "Modified peak hours scenario with faster response",
    "scenario_type": "peak_hours",
    "base_delay": "100ms",
    "delay_strategy": "fixed",
    "servicenow_mode": true,
    "batch_size": 25,
    "response_limits": {
        "max_count": 50000,
        "default_count": 5000
    }
}
```

Now when you use `scenario=peak_hours`, your custom configuration will be used instead of the built-in one.

## Scenario Validation

PayloadBuddy provides built-in validation to help you create correct scenario files.

### Command-Line Validation

Use the `-verify` flag to validate scenario files before deploying them:

```bash
# Basic validation
./payloadBuddy -verify my-scenario.json

# Validate a scenario in your config directory
./payloadBuddy -verify $HOME/.config/payloadBuddy/scenarios/custom-test.json

# Validate multiple files
./payloadBuddy -verify scenario1.json
./payloadBuddy -verify scenario2.json
```

### Validation Output

**Successful validation:**
```
Validating scenario file: my-scenario.json
Validation successful!

Scenario Details:
   Name: My Custom Test
   Type: custom
   Base Delay: 100ms
   Delay Strategy: progressive
   ServiceNow Mode: enabled
   Batch Size: 50
   Author: John Doe
   Version: 1.0.0
   Tags: [testing performance]

Usage: Use this scenario with ?scenario=custom
Tip: Place this file in $HOME/.config/payloadBuddy/scenarios/ to make it available
```

**Failed validation:**
```
Validating scenario file: invalid-scenario.json
Validation failed:
scenario_name is required
```

### Best Practices

1. **Validate Early**: Always validate scenario files before deploying
2. **Test Incrementally**: Start with minimal scenarios and add complexity
3. **Use Descriptive Names**: Make scenario names and descriptions clear
4. **Version Your Scenarios**: Use metadata fields to track changes
5. **Document Parameters**: Use the description field to explain scenario behavior

## JSON Schema Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `scenario_name` | string | Human-readable name (1-100 characters) |
| `scenario_type` | string | Scenario type identifier |
| `base_delay` | string | Base delay in Go duration format or milliseconds |

### Scenario Types

| Type | Description |
|------|-------------|
| `peak_hours` | Peak usage simulation |
| `maintenance` | Maintenance window simulation |
| `network_issues` | Network instability simulation |
| `database_load` | Progressive load simulation |
| `custom` | User-defined behavior |

### Delay Strategies

| Strategy | Description |
|----------|-------------|
| `fixed` | Consistent delay between items |
| `random` | Random delay up to 2x base delay |
| `progressive` | Increasing delay over time |
| `burst` | Short delays with periodic long pauses |

### Optional Configuration

#### Response Limits
```json
"response_limits": {
    "max_count": 100000,      // Maximum items allowed
    "default_count": 10000    // Default when count not specified
}
```

#### ServiceNow Configuration
```json
"servicenow_config": {
    "record_types": ["incident", "problem", "change_request"],
    "state_rotation": ["New", "In Progress", "Resolved", "Closed"],
    "number_format": "INC%07d",
    "sys_id_format": "standard"
}
```

#### Error Injection
```json
"error_injection": {
    "enabled": true,
    "error_rate": 0.05,       // 5% error rate
    "error_types": ["timeout", "server_error"],
    "recovery_delay": "1s",
    "consecutive_error_limit": 3
}
```

#### Performance Monitoring
```json
"performance_monitoring": {
    "enabled": true,
    "metrics_interval": 1000,  // Every 1000 items
    "memory_tracking": true,
    "checkpoint_logging": true
}
```

#### Metadata
```json
"metadata": {
    "author": "Your Name",
    "created_date": "2025-01-15",
    "version": "1.0.0",
    "project": "My Test Project",
    "tags": ["testing", "performance", "custom"],
    "compatibility": {
        "min_payloadbuddy_version": "1.0.0",
        "tested_versions": ["1.0.0", "1.1.0"]
    }
}
```

## Advanced Features

### Scenario Parameters

For advanced customization, use the `scenario_parameters` section:

```json
"scenario_parameters": {
    "delay_overrides": {
        "spike_delay": "5s",
        "recovery_delay": "200ms"
    },
    "timing_patterns": {
        "intervals": [100, 500, 1000],
        "probabilities": [0.1, 0.2, 0.7],
        "thresholds": {
            "high_load_threshold": 1000,
            "critical_threshold": 5000
        }
    },
    "simulation_config": {
        "load_type": "progressive",
        "degradation_rate": 10,
        "recovery_enabled": true
    }
}
```

### Complex Scenario Example

Here's a comprehensive scenario showcasing all features:

```json
{
    "schema_version": "1.0.0",
    "scenario_name": "Enterprise Load Test",
    "description": "Comprehensive enterprise-grade load testing scenario",
    "scenario_type": "custom",
    "base_delay": "50ms",
    "delay_strategy": "progressive",
    "servicenow_mode": true,
    "batch_size": 100,
    "response_limits": {
        "max_count": 500000,
        "default_count": 25000
    },
    "scenario_parameters": {
        "delay_overrides": {
            "initial_delay": "25ms",
            "max_delay": "2s",
            "spike_probability": "0.15"
        },
        "timing_patterns": {
            "intervals": [500, 1000, 2500],
            "probabilities": [0.2, 0.3, 0.5],
            "thresholds": {
                "warning_threshold": 10000,
                "critical_threshold": 40000
            }
        },
        "simulation_config": {
            "load_type": "enterprise",
            "peak_hours": true,
            "database_contention": true,
            "memory_pressure": "moderate"
        }
    },
    "servicenow_config": {
        "record_types": ["incident", "problem", "change_request"],
        "state_rotation": ["New", "In Progress", "On Hold", "Resolved", "Closed"],
        "number_format": "INC%07d",
        "sys_id_format": "standard",
        "custom_fields": {
            "priority": ["1 - Critical", "2 - High", "3 - Moderate", "4 - Low"],
            "category": ["Hardware", "Software", "Network", "Database"]
        }
    },
    "error_injection": {
        "enabled": true,
        "error_rate": 0.02,
        "error_types": ["timeout", "server_error", "rate_limit"],
        "recovery_delay": "5s",
        "consecutive_error_limit": 2
    },
    "performance_monitoring": {
        "enabled": true,
        "metrics_interval": 2500,
        "memory_tracking": true,
        "checkpoint_logging": true
    },
    "metadata": {
        "author": "Enterprise Test Team",
        "created_date": "2025-01-15",
        "modified_date": "2025-01-20",
        "version": "2.1.0",
        "project": "Enterprise ServiceNow Integration",
        "tags": ["enterprise", "load-testing", "servicenow", "production-ready"],
        "compatibility": {
            "min_payloadbuddy_version": "1.0.0",
            "tested_versions": ["1.0.0", "1.1.0", "1.2.0"]
        }
    }
}
```

## Examples

### 1. High-Load Testing Scenario

Perfect for testing client behavior under extreme load:

```json
{
    "schema_version": "1.0.0",
    "scenario_name": "High Load Stress Test",
    "scenario_type": "custom",
    "base_delay": "10ms",
    "delay_strategy": "burst",
    "servicenow_mode": true,
    "batch_size": 200,
    "response_limits": {
        "max_count": 1000000,
        "default_count": 100000
    }
}
```

### 2. Slow Network Simulation

Simulates slow network conditions:

```json
{
    "schema_version": "1.0.0",
    "scenario_name": "Slow Network Test",
    "scenario_type": "custom",
    "base_delay": "500ms",
    "delay_strategy": "random",
    "servicenow_mode": false,
    "batch_size": 10,
    "response_limits": {
        "max_count": 1000,
        "default_count": 100
    }
}
```

### 3. ServiceNow Production Mirror

Mirrors actual ServiceNow production behavior:

```json
{
    "schema_version": "1.0.0",
    "scenario_name": "Production Mirror",
    "scenario_type": "custom",
    "base_delay": "150ms",
    "delay_strategy": "progressive",
    "servicenow_mode": true,
    "batch_size": 50,
    "servicenow_config": {
        "record_types": ["incident", "problem"],
        "state_rotation": ["1", "2", "6", "7"],
        "number_format": "INC%07d"
    },
    "error_injection": {
        "enabled": true,
        "error_rate": 0.001,
        "error_types": ["timeout"]
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Scenario Not Loading
**Problem**: Custom scenario isn't being used
**Solutions**:
- Check file placement in `$HOME/.config/payloadBuddy/scenarios/`
- Verify `.json` file extension
- Check PayloadBuddy startup logs for validation errors
- Ensure `scenario_type` matches the value used in API calls

#### 2. Validation Errors
**Problem**: Scenario file fails validation
**Solutions**:
- Check required fields: `scenario_name`, `scenario_type`, `base_delay`
- Verify delay format (e.g., "100ms", "1s", or just "500")
- Ensure `scenario_type` is one of: `peak_hours`, `maintenance`, `network_issues`, `database_load`, `custom`
- Validate date formats in metadata (YYYY-MM-DD)

#### 3. Override Not Working
**Problem**: Built-in scenario still being used instead of custom one
**Solutions**:
- Ensure `scenario_type` exactly matches the built-in scenario name
- Check that the custom scenario file is valid JSON
- Restart PayloadBuddy after adding new scenarios
- Check logs for "overriding embedded scenario" message

#### 4. Directory Not Created
**Problem**: `$HOME/.config/payloadBuddy/scenarios/` doesn't exist
**Solutions**:
- Run PayloadBuddy at least once to trigger directory creation
- Manually create the directory: `mkdir -p $HOME/.config/payloadBuddy/scenarios`
- Check permissions on the `$HOME/.config` directory
- On Windows, directory is `%USERPROFILE%\.config\payloadBuddy\scenarios`

### Debugging Tips

1. **Enable Verbose Logging**: Check PayloadBuddy startup output for scenario loading messages
2. **Validate JSON**: Use online JSON validators to check file syntax
3. **Test Incrementally**: Start with minimal scenarios and add complexity gradually
4. **Check File Permissions**: Ensure PayloadBuddy can read your scenario files
5. **Use Built-in Examples**: Copy and modify built-in scenarios as starting points

### Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/dtrabandt/payloadBuddy/issues)
- **Documentation**: Check README.md for general usage information
- **Schema Reference**: Refer to `scenarios/scenario_schema_v1.0.0.json` in the repository
- **Examples**: See the `scenarios/` directory for built-in scenario examples

---

*For more information about PayloadBuddy, see the main [README.md](README.md) file.*