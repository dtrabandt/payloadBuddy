package main

import (
	"testing"
)

func TestRestPayloadPlugin_Interface(t *testing.T) {
	plugin := RestPayloadPlugin{}

	// Test Path method
	path := plugin.Path()
	expectedPath := "/rest_payload"
	if path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}

	// Test Handler method
	handler := plugin.Handler()
	if handler == nil {
		t.Error("Handler should not be nil")
	}

	// Test that handler function matches expected function
	// We can't directly compare function pointers, but we can call it
	// This is implicitly tested in other handler tests
}

func TestStreamingPayloadPlugin_Interface(t *testing.T) {
	plugin := StreamingPayloadPlugin{}

	// Test Path method
	path := plugin.Path()
	expectedPath := "/stream_payload"
	if path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}

	// Test Handler method
	handler := plugin.Handler()
	if handler == nil {
		t.Error("Handler should not be nil")
	}

	// Test that handler function matches expected function
	// We can't directly compare function pointers, but we can call it
	// This is implicitly tested in other handler tests
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
		"/rest_payload":       false,
		"/stream_payload":     false,
		"/paginated_payload":  false,
		"/openapi.json":       false,
		"/swagger":            false,
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
}
