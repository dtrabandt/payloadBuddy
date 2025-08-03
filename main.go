package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Release 0.1.0
var Version = "0.2.1"

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
	paramPort = flag.String("port", "8080", "Port to run the HTTP server on")
)

// Setup the port for the HTTP server.
// If the provided port is empty or not possible to parse,
// it defaults to 8080. It also defaults to 8080 if the port is out of range.
func setupPort(desiredPort string) string {
	defaultPort := "8080"

	i, err := strconv.Atoi(desiredPort)
	if err != nil || i <= 0 || i > 65535 {
		return defaultPort // Return default port if parsing fails or invalid port
	} else {
		// Ensure the port is within valid range
		if i < 1 || i > 65535 {
			return defaultPort // Return default port if out of range
		} else {
			return desiredPort // Return the valid port specified by the user
		}
	}
}

// main is the entry point for the payloadBuddy application.
// It starts an HTTP server on port 8080 and registers all plugin endpoints.
// The server returns large JSON payloads for testing REST client implementations.
func main() {
	// Parse command line flags
	flag.Parse()

	// Setup authentication if enabled
	setupAuthentication()

	// Register plugins with conditional authentication middleware
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

	port := setupPort(*paramPort)
	addr := ":" + port

	fmt.Printf("\nStarting payloadBuddy %s on http://localhost:%s\n", Version, port)

	// Print authentication info if enabled
	printAuthenticationInfo()

	fmt.Println("\nAvailable endpoints:")
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/rest_payload", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/openapi.json", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/swagger", port)))

	fmt.Println("\nRest Payload examples:")
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/rest_payload", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/rest_payload?count=5000", port)))

	fmt.Println("\nStreaming examples:")
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload?count=1000&delay=100ms", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload?scenario=peak_hours&servicenow=true", port)))
	fmt.Printf("  %s\n", getExampleURL(fmt.Sprintf("http://localhost:%s/stream_payload?delay=50ms&strategy=random&batch_size=50", port)))

	fmt.Println("\nServiceNow test scenarios:")
	fmt.Printf("  - peak_hours: Simulates ServiceNow during peak hours\n")
	fmt.Printf("  - maintenance: Simulates maintenance windows with spikes\n")
	fmt.Printf("  - network_issues: Random network delays\n")
	fmt.Printf("  - database_load: Progressive database load simulation\n")

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
