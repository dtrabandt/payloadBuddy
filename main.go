package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

var Version = "0.2.5"

// Global scenario manager
var scenarioManager *ScenarioManager

// PayloadPlugin is an interface that must be implemented by
// any plugin that wants to register a payload handler.
// It provides the Path, Handler, and OpenAPISpec methods for the HTTP endpoint.
type PayloadPlugin interface {
	Path() string
	Handler() http.HandlerFunc
	OpenAPISpec() OpenAPIPathSpec
}

// plugins holds the list of registered payload plugins.
var plugins []PayloadPlugin

// registerPlugin adds a new PayloadPlugin to the list of
// registered plugins. It is called by the init function
// of each plugin implementation.
func registerPlugin(p PayloadPlugin) {
	plugins = append(plugins, p)
}

// Setup the variables from the command line flags.
var (
	paramPort   = flag.String("port", "8080", "Port to run the HTTP server on")
	paramVerify = flag.String("verify", "", "Validate a scenario file against the JSON schema and exit")
)

// Setup the port for the HTTP server.
// If the provided port is empty or not possible to parse,
// it defaults to 8080. It also defaults to 8080 if the port is out of range.
func setupPort(desiredPort string) string {
	defaultPort := "8080"

	i, err := strconv.Atoi(desiredPort)
	if err != nil || i < 1 || i > 65535 {
		return defaultPort // Return default port if parsing fails or out of range
	}
	return desiredPort // Return the valid port specified by the user
}

// verifyScenarioFile validates a scenario file using the scenario validator
func verifyScenarioFile(filePath string) {
	validator := NewScenarioValidator()
	validator.ValidateScenarioFile(filePath)
}

// registerPlugins registers all plugins with conditional authentication middleware
func registerPlugins() {
	for _, p := range plugins {
		path := p.Path()
		// Exclude documentation endpoints from authentication for better UX
		if path == "/swagger" || path == "/openapi.json" {
			http.HandleFunc(path, p.Handler())
			fmt.Printf("Registered endpoint: %s (no auth)\n", path)
		} else {
			http.HandleFunc(path, basicAuthMiddleware(p.Handler()))
			fmt.Printf("Registered endpoint: %s\n", path)
		}
	}
}

// printStartupInfo prints application startup information and usage examples
func printStartupInfo(port string) {
	fmt.Printf("\nStarting payloadBuddy %s on http://localhost:%s\n", Version, port)

	// Print authentication info if enabled
	printAuthenticationInfo()

	// Print usage examples
	printUsageExamples(port)
}

// initializeServer registers plugins and prepares server startup
func initializeServer() string {
	registerPlugins()
	port := setupPort(*paramPort)
	printStartupInfo(port)
	return port
}

// printUsageExamples prints all the usage examples and scenarios
func printUsageExamples(port string) {
	fmt.Println("\nAvailable endpoints:")
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/rest_payload", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/paginated_payload", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/openapi.json", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/swagger", port)))

	fmt.Println("\nRest Payload examples:")
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/rest_payload", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/rest_payload?count=5000", port)))

	fmt.Println("\nPagination examples (ServiceNow Data Stream compatible):")
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/paginated_payload?limit=100&offset=0&servicenow=true", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/paginated_payload?page=2&size=50&servicenow=true", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/paginated_payload?scenario=peak_hours&servicenow=true", port)))

	fmt.Println("\nStreaming examples:")
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload?count=1000&delay=100ms", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload?scenario=peak_hours&servicenow=true", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload?delay=50ms&strategy=random&batch_size=50", port)))

	printServiceNowScenarios()
}

// getScenarioUsageContext returns usage context information for scenarios
func getScenarioUsageContext(scenarioType string) string {
	switch scenarioType {
	case "peak_hours":
		return " • Best for: both streaming and pagination testing"
	case "maintenance":
		return " • Best for: streaming (periodic spikes), pagination (single spike per page)"
	case "network_issues":
		return " • Best for: both (random delays simulate real network conditions)"
	case "database_load":
		return " • Best for: streaming (progressive degradation), pagination (single delay per page)"
	default:
		return ""
	}
}

// printServiceNowScenarios prints all available ServiceNow scenarios including custom ones
func printServiceNowScenarios() {
	fmt.Println("\nServiceNow test scenarios (compatible with both streaming and pagination):")

	// Get all scenario types from the scenario manager
	scenarioTypes := scenarioManager.ListScenarios()

	for _, scenarioType := range scenarioTypes {
		scenario := scenarioManager.GetScenario(scenarioType)
		usageContext := getScenarioUsageContext(scenarioType)

		if scenario != nil && scenario.Description != "" {
			// Use full scenario description with usage context
			description := scenario.Description
			// Make description more concise for startup output
			if len(description) > 80 {
				description = description[:77] + "..."
			}
			fmt.Printf("  - %s: %s\n", scenarioType, description)
			if usageContext != "" {
				fmt.Printf("    %s\n", usageContext)
			}
		} else {
			// Fallback descriptions for scenarios without description
			switch scenarioType {
			case "peak_hours":
				fmt.Printf("  - %s: Simulates ServiceNow during peak hours\n", scenarioType)
			case "maintenance":
				fmt.Printf("  - %s: Simulates maintenance windows with spikes\n", scenarioType)
			case "network_issues":
				fmt.Printf("  - %s: Random network delays\n", scenarioType)
			case "database_load":
				fmt.Printf("  - %s: Progressive database load simulation\n", scenarioType)
			default:
				fmt.Printf("  - %s: Custom scenario\n", scenarioType)
			}
			if usageContext != "" {
				fmt.Printf("    %s\n", usageContext)
			}
		}
	}
}

// startHTTPServer starts the HTTP server with proper configuration
func startHTTPServer(port string) {
	addr := ":" + port

	fmt.Println("\nPress Ctrl+C to stop the server")

	// Start the HTTP server with proper timeouts to prevent resource exhaustion
	server := &http.Server{
		Addr:         addr,
		Handler:      nil, // Use DefaultServeMux
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		// Print error to stderr and exit with non-zero code.
		fmt.Fprintf(os.Stderr, "Server failed to start: %v\n", err)
		os.Exit(1)
	}
}

// main is the entry point for the payloadBuddy application.
// It starts an HTTP server on port 8080 and registers all plugin endpoints.
// The server returns large JSON payloads for testing REST client implementations.
func main() {
	// Parse command line flags
	flag.Parse()

	// Handle scenario file verification
	if *paramVerify != "" {
		verifyScenarioFile(*paramVerify)
		return
	}

	// Initialize scenario manager
	scenarioManager = NewScenarioManager()

	// Setup authentication if enabled
	setupAuthentication()

	// Initialize server components
	port := initializeServer()
	startHTTPServer(port)
}

// RestPayloadPlugin implements PayloadPlugin for large JSON payloads
type RestPayloadPlugin struct{}

// Path returns the HTTP path for the rest payload endpoint.
func (h RestPayloadPlugin) Path() string { return "/rest_payload" }

// Handler returns the handler function for the rest payload endpoint.
func (h RestPayloadPlugin) Handler() http.HandlerFunc { return RestPayloadHandler }

// StreamingPayloadPlugin implements PayloadPlugin for streaming data
type StreamingPayloadPlugin struct{}

// Path returns the HTTP path for the streaming payload endpoint.
func (s StreamingPayloadPlugin) Path() string { return "/stream_payload" }

// Handler returns the handler function for the streaming payload endpoint.
func (s StreamingPayloadPlugin) Handler() http.HandlerFunc { return StreamingPayloadHandler }

func init() {
	registerPlugin(RestPayloadPlugin{})
	registerPlugin(StreamingPayloadPlugin{})
}
