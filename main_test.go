package main

import (
	"testing"
)

func TestPayloadPlugins_Interface(t *testing.T) {
	tests := []struct {
		name         string
		plugin       PayloadPlugin
		expectedPath string
	}{
		{
			name:         "RestPayloadPlugin",
			plugin:       RestPayloadPlugin{},
			expectedPath: "/rest_payload",
		},
		{
			name:         "StreamingPayloadPlugin",
			plugin:       StreamingPayloadPlugin{},
			expectedPath: "/stream_payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Path method
			path := tt.plugin.Path()
			if path != tt.expectedPath {
				t.Errorf("Expected path %q, got %q", tt.expectedPath, path)
			}

			// Test Handler method
			handler := tt.plugin.Handler()
			if handler == nil {
				t.Error("Handler should not be nil")
			}

			// Test OpenAPISpec method
			spec := tt.plugin.OpenAPISpec()
			if spec.Path != path {
				t.Errorf("OpenAPISpec path %q doesn't match Path() %q", spec.Path, path)
			}
			if spec.Operation.Get == nil {
				t.Error("OpenAPISpec missing GET operation")
			}
		})
	}
}

func TestRegisterPlugin(t *testing.T) {
	// Save original plugins list
	originalPlugins := plugins
	defer func() {
		plugins = originalPlugins
	}()

	// Reset plugins list for testing
	plugins = []PayloadPlugin{}

	// Create a test plugin
	testPlugin := RestPayloadPlugin{}

	// Register the plugin
	registerPlugin(testPlugin)

	// Check that plugin was added
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	// Check that the right plugin was added
	if plugins[0].Path() != testPlugin.Path() {
		t.Errorf("Expected plugin path %q, got %q", testPlugin.Path(), plugins[0].Path())
	}

	// Register another plugin
	testPlugin2 := StreamingPayloadPlugin{}
	registerPlugin(testPlugin2)

	// Check that both plugins are registered
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}

	// Check that plugins are in the correct order
	expectedPaths := []string{"/rest_payload", "/stream_payload"}
	for i, plugin := range plugins {
		if plugin.Path() != expectedPaths[i] {
			t.Errorf("Expected plugin %d path %q, got %q", i, expectedPaths[i], plugin.Path())
		}
	}
}

func TestPluginRegistration_InitFunctions(t *testing.T) {
	// This test verifies that plugins are properly registered via init() functions
	// Since init() functions run automatically, we just need to check that
	// the expected plugins are registered

	expectedPlugins := map[string]bool{
		"/rest_payload":   false,
		"/stream_payload": false,
		"/openapi.json":   false,
		"/swagger":        false,
	}

	// Check that all expected plugins are registered
	for _, plugin := range plugins {
		path := plugin.Path()
		if _, exists := expectedPlugins[path]; exists {
			expectedPlugins[path] = true
		} else {
			t.Errorf("Unexpected plugin registered: %s", path)
		}
	}

	// Check that all expected plugins were found
	for path, found := range expectedPlugins {
		if !found {
			t.Errorf("Expected plugin not found: %s", path)
		}
	}

	// Check that we have the expected number of plugins
	expectedCount := len(expectedPlugins)
	actualCount := len(plugins)
	if actualCount != expectedCount {
		t.Errorf("Expected %d plugins, got %d", expectedCount, actualCount)
	}
}

func TestPayloadPluginInterface_Compliance(t *testing.T) {
	// Test that all registered plugins properly implement the PayloadPlugin interface
	for _, plugin := range plugins {
		// Test Path method
		path := plugin.Path()
		if path == "" {
			t.Errorf("Plugin %T returned empty path", plugin)
		}
		if !isValidHTTPPath(path) {
			t.Errorf("Plugin %T returned invalid HTTP path: %s", plugin, path)
		}

		// Test Handler method
		handler := plugin.Handler()
		if handler == nil {
			t.Errorf("Plugin %T returned nil handler", plugin)
		}

		// Test OpenAPISpec method
		spec := plugin.OpenAPISpec()
		if spec.Path != path {
			t.Errorf("Plugin %T: OpenAPISpec path %q doesn't match Path() %q", plugin, spec.Path, path)
		}
		if spec.Operation.Get == nil {
			t.Errorf("Plugin %T: OpenAPISpec missing GET operation", plugin)
		}
	}
}

// Helper function to validate HTTP paths
func isValidHTTPPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	if path[0] != '/' {
		return false
	}
	// Additional path validation could be added here
	return true
}

// This test verifies the setting of the default port. It's 8080,
// however, the user can override it with the -port attribute.
func TestSetupPort(t *testing.T) {
	expectedPort := "8080"

	// No port specified, should return default
	if port := setupPort(""); port != expectedPort {
		t.Errorf("Expected port %s, got %s", expectedPort, port)
	}

	// Port specified, should return that port
	if port := setupPort("9090"); port != "9090" {
		t.Errorf("Expected port 9090, got %s", port)
	}

	// Invalid port specified, should return default
	if port := setupPort("invalid_port"); port != expectedPort {
		t.Errorf("Expected default port %s for invalid input, got %s", expectedPort, port)
	}

	// Port out of range, should return default
	if port := setupPort("70000"); port != expectedPort {
		t.Errorf("Expected default port %s for out of range input, got %s", expectedPort, port)
	}

	// Port is zero, should return default
	if port := setupPort("0"); port != expectedPort {
		t.Errorf("Expected default port %s for zero input, got %s", expectedPort, port)
	}

	// Port is negative, should return default
	if port := setupPort("-1"); port != expectedPort {
		t.Errorf("Expected default port %s for negative input, got %s", expectedPort, port)
	}

	// Test edge case: port exactly at maximum valid range
	if port := setupPort("65535"); port != "65535" {
		t.Errorf("Expected port 65535, got %s", port)
	}

	// Test edge case: port just above maximum valid range
	if port := setupPort("65536"); port != expectedPort {
		t.Errorf("Expected default port %s for out of range input, got %s", expectedPort, port)
	}

	// Test edge case: port at minimum valid range
	if port := setupPort("1"); port != "1" {
		t.Errorf("Expected port 1, got %s", port)
	}
}

func TestSetupPort_Comprehensive(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{"empty_string", "", "8080", "Empty string should return default port"},
		{"valid_port_mid_range", "9090", "9090", "Valid port in middle range"},
		{"valid_port_low", "80", "80", "Valid low port number"},
		{"valid_port_high", "65000", "65000", "Valid high port number"},
		{"port_minimum", "1", "1", "Minimum valid port"},
		{"port_maximum", "65535", "65535", "Maximum valid port"},
		{"port_zero", "0", "8080", "Port zero should return default"},
		{"port_negative", "-1", "8080", "Negative port should return default"},
		{"port_negative_large", "-9999", "8080", "Large negative port should return default"},
		{"port_too_high", "65536", "8080", "Port above range should return default"},
		{"port_way_too_high", "70000", "8080", "Port way above range should return default"},
		{"invalid_string", "invalid", "8080", "Invalid string should return default"},
		{"mixed_string", "80abc", "8080", "Mixed string should return default"},
		{"float_string", "80.5", "8080", "Float string should return default"},
		{"empty_space", " ", "8080", "Space should return default"},
		{"multiple_numbers", "80 90", "8080", "Multiple numbers should return default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setupPort(tt.input)
			if result != tt.expected {
				t.Errorf("setupPort(%q) = %q, expected %q. %s", tt.input, result, tt.expected, tt.description)
			}
		})
	}
}

func TestPrintServiceNowScenarios(t *testing.T) {
	// Save original scenario manager
	originalManager := scenarioManager
	defer func() {
		scenarioManager = originalManager
	}()

	tests := []struct {
		name        string
		setupFunc   func()
		expectPanic bool
	}{
		{
			name: "with_loaded_scenarios",
			setupFunc: func() {
				scenarioManager = NewScenarioManager()
			},
			expectPanic: false,
		},
		{
			name: "with_nil_scenario_manager",
			setupFunc: func() {
				scenarioManager = nil
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected function to panic, but it didn't")
					}
				}()
			}

			printServiceNowScenarios()

			if !tt.expectPanic && scenarioManager != nil {
				// Verify scenarios are available
				scenarios := scenarioManager.ListScenarios()
				if len(scenarios) == 0 {
					t.Error("Expected at least some scenarios to be available")
				}

				// Test that we can get individual scenarios
				for _, scenarioType := range scenarios {
					scenario := scenarioManager.GetScenario(scenarioType)
					if scenario == nil {
						t.Errorf("Expected to get scenario for type %s, but got nil", scenarioType)
					}
				}
			}
		})
	}
}

func TestRegisterPluginsAndStart_PortLogic(t *testing.T) {
	// Test only the port setup logic, not the actual HTTP registration
	// since that causes conflicts when called multiple times in tests

	// Save original state
	originalParamPort := paramPort
	defer func() {
		paramPort = originalParamPort
	}()

	tests := []struct {
		name         string
		portParam    string
		expectedPort string
		description  string
	}{
		{
			name:         "default_port",
			portParam:    "8080",
			expectedPort: "8080",
			description:  "Default port should be used",
		},
		{
			name:         "custom_valid_port",
			portParam:    "9999",
			expectedPort: "9999",
			description:  "Custom valid port should be used",
		},
		{
			name:         "invalid_port_fallback",
			portParam:    "invalid",
			expectedPort: "8080",
			description:  "Invalid port should fallback to default",
		},
		{
			name:         "out_of_range_port_fallback",
			portParam:    "70000",
			expectedPort: "8080",
			description:  "Out of range port should fallback to default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*paramPort = tt.portParam

			// Test only the port setup logic
			port := setupPort(*paramPort)

			if port != tt.expectedPort {
				t.Errorf("%s: Expected port %s, got %s", tt.description, tt.expectedPort, port)
			}
		})
	}
}

func TestPrintUsageExamples(t *testing.T) {
	// Save original scenario manager
	originalManager := scenarioManager
	defer func() {
		scenarioManager = originalManager
	}()

	tests := []struct {
		name        string
		port        string
		setupFunc   func()
		expectPanic bool
	}{
		{
			name: "valid_port_8080",
			port: "8080",
			setupFunc: func() {
				scenarioManager = NewScenarioManager()
			},
			expectPanic: false,
		},
		{
			name: "valid_port_9999",
			port: "9999",
			setupFunc: func() {
				scenarioManager = NewScenarioManager()
			},
			expectPanic: false,
		},
		{
			name: "nil_scenario_manager",
			port: "8080",
			setupFunc: func() {
				scenarioManager = nil
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected function to panic, but it didn't")
					}
				}()
			}

			printUsageExamples(tt.port)
		})
	}
}

func TestStartHTTPServer_Configuration(t *testing.T) {
	// This function calls ListenAndServe which would block,
	// so we test the server configuration by checking the setup
	// We can't easily test the actual server start without complex mocking

	// Test that the function exists and doesn't panic during setup
	// The actual server start is tested in integration tests
	t.Skip("startHTTPServer calls ListenAndServe which blocks - tested in integration tests")
}

func TestMain_Refactored_Structure(t *testing.T) {
	// Test that the main function components work together
	// Save original state
	originalManager := scenarioManager
	originalPlugins := plugins
	defer func() {
		scenarioManager = originalManager
		plugins = originalPlugins
	}()

	// Set up test environment
	scenarioManager = NewScenarioManager()
	plugins = []PayloadPlugin{
		RestPayloadPlugin{},
		StreamingPayloadPlugin{},
	}

	// Test individual components that main() calls
	t.Run("scenario_manager_initialization", func(t *testing.T) {
		if scenarioManager == nil {
			t.Error("Scenario manager should be initialized")
		}

		scenarios := scenarioManager.ListScenarios()
		if len(scenarios) == 0 {
			t.Error("Expected some scenarios to be loaded")
		}
	})

	t.Run("plugin_registration", func(t *testing.T) {
		if len(plugins) == 0 {
			t.Error("Expected plugins to be registered")
		}

		for _, plugin := range plugins {
			if plugin.Path() == "" {
				t.Error("Plugin should have non-empty path")
			}
			if plugin.Handler() == nil {
				t.Error("Plugin should have non-nil handler")
			}
		}
	})

	t.Run("port_setup", func(t *testing.T) {
		port := setupPort("8080")
		if port != "8080" {
			t.Errorf("Expected port 8080, got %s", port)
		}
	})
}

// Test printServiceNowScenarios fallback logic
func TestPrintServiceNowScenarios_FallbackLogic(t *testing.T) {
	// Save original scenario manager
	originalManager := scenarioManager
	defer func() {
		scenarioManager = originalManager
	}()

	// Create a scenario manager with scenarios without descriptions to trigger fallback
	scenarioManager = NewScenarioManager()

	// Add scenarios that will trigger specific fallback cases in the switch statement
	scenarioManager.scenarios["peak_hours"] = &Scenario{
		SchemaVersion: "1.0.0",
		ScenarioName:  "Peak Hours Test",
		ScenarioType:  "peak_hours",
		BaseDelay:     "100ms",
		Description:   "", // Empty to trigger fallback
	}

	scenarioManager.scenarios["custom_test"] = &Scenario{
		SchemaVersion: "1.0.0",
		ScenarioName:  "Custom Test",
		ScenarioType:  "custom_test",
		BaseDelay:     "100ms",
		Description:   "", // Empty to trigger default case
	}

	// This should trigger the fallback logic in printServiceNowScenarios
	printServiceNowScenarios()
}
