package main

import (
	"flag"
	"fmt"
	mathRand "math/rand"
	"net/http"
	"os"
	"time"
)

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

// main is the entry point for the payloadBuddy application.
// It starts an HTTP server on port 8080 and registers all plugin endpoints.
// The server returns large JSON payloads for testing REST client implementations.
func main() {
	// Parse command line flags
	flag.Parse()

	// Initialize random seed for delay variations in streaming
	mathRand.Seed(time.Now().UnixNano())

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

	port := "8080"
	addr := ":" + port

	fmt.Printf("\nStarting payloadBuddy on http://localhost:%s\n", port)

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

	// Start the HTTP server and log any fatal errors.
	if err := http.ListenAndServe(addr, nil); err != nil {
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
