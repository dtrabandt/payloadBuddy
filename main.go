package main

import (
	"fmt"
	"net/http"
	"os"
)

// PayloadPlugin is an interface that must be implemented by
// any plugin that wants to register a payload handler.
// It provides the Path and Handler methods for the HTTP endpoint.
type PayloadPlugin interface {
	Path() string
	Handler() http.HandlerFunc
}

// plugins holds the list of registered payload plugins.
var plugins []PayloadPlugin

// registerPlugin adds a new PayloadPlugin to the list of
// registered plugins. It is called by the init function
// of each plugin implementation.
func registerPlugin(p PayloadPlugin) {
	plugins = append(plugins, p)
}

// main is the entry point for the gohugePayloadServer application.
// It starts an HTTP server on port 8080 and registers the /huge_payload endpoint.
// The server returns large JSON payloads for testing REST client implementations.
func main() {
	// Register plugins
	for _, p := range plugins {
		http.HandleFunc(p.Path(), p.Handler())
	}

	port := "8080"
	addr := ":" + port
	fmt.Printf("Starting server on http://localhost:%s/huge_payload\n", port)

	// Start the HTTP server and log any fatal errors.
	if err := http.ListenAndServe(addr, nil); err != nil {
		// Print error to stderr and exit with non-zero code.
		fmt.Fprintf(os.Stderr, "Server failed to start: %v\n", err)
		os.Exit(1)
	}
}

// HugePayloadPlugin implements PayloadPlugin for large JSON payloads
type HugePayloadPlugin struct{}

// Path returns the HTTP path for the huge payload endpoint.
func (h HugePayloadPlugin) Path() string { return "/huge_payload" }

// Handler returns the handler function for the huge payload endpoint.
func (h HugePayloadPlugin) Handler() http.HandlerFunc { return HugePayloadHandler }

// StreamingPayloadPlugin implements PayloadPlugin for streaming data
type StreamingPayloadPlugin struct{}

// Path returns the HTTP path for the streaming payload endpoint.
func (s StreamingPayloadPlugin) Path() string { return "/stream_payload" }

// Handler returns the handler function for the streaming payload endpoint.
func (s StreamingPayloadPlugin) Handler() http.HandlerFunc { return StreamingPayloadHandler }

func init() {
	registerPlugin(HugePayloadPlugin{})
	registerPlugin(StreamingPayloadPlugin{})
}
