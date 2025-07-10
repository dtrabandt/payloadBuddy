package main

import (
	"fmt"
	"net/http"
	"os"
)

// main is the entry point for the gohugePayloadServer application.
// It starts an HTTP server on port 8080 and registers the /huge_payload endpoint.
// The server returns large JSON payloads for testing REST client implementations.
func main() {
	// Register the /huge_payload endpoint with its handler function.
	http.HandleFunc("/huge_payload", HugePayloadHandler)

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
