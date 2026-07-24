package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dtrabandt/payloadBuddy/internal/auth"
	"github.com/dtrabandt/payloadBuddy/internal/handlers"
	"github.com/dtrabandt/payloadBuddy/internal/scenarios"
)

var Version = "0.3.0"

var (
	paramPort     = flag.String("port", "8080", "Port to run the HTTP server on")
	paramVerify   = flag.String("verify", "", "Validate a scenario file against the JSON schema and exit")
	paramAuth     = flag.Bool("auth", false, "Enable basic authentication")
	paramUsername = flag.String("user", "", "Username for basic auth (auto-generated if empty)")
	paramPassword = flag.String("pass", "", "Password for basic auth (auto-generated if empty)")
)

func main() {
	flag.Parse()

	if *paramVerify != "" {
		v := scenarios.NewValidator()
		v.ValidateScenarioFile(*paramVerify)
		return
	}

	sm := scenarios.NewManager()
	authCfg := auth.Setup(*paramAuth, *paramUsername, *paramPassword)

	docPlugin := &handlers.DocumentationPlugin{}
	allPlugins := []handlers.PayloadPlugin{
		handlers.RestPayloadPlugin{},
		handlers.StreamingPayloadPlugin{SM: sm},
		handlers.PaginatedPayloadPlugin{SM: sm},
		docPlugin,
		handlers.SwaggerUIPlugin{},
	}
	docPlugin.Init(allPlugins, authCfg)

	mux := http.NewServeMux()
	for _, p := range allPlugins {
		path := p.Path()
		if path == "/swagger" || path == "/openapi.json" {
			mux.HandleFunc("GET "+path, p.Handler())
			fmt.Printf("Registered endpoint: %s (no auth)\n", path)
		} else {
			mux.HandleFunc("GET "+path, authCfg.Middleware(p.Handler()))
			fmt.Printf("Registered endpoint: %s\n", path)
		}
	}

	port := setupPort(*paramPort)
	printStartupInfo(port, sm, authCfg)
	startHTTPServer(port, mux)
}

func setupPort(desiredPort string) string {
	i, err := strconv.Atoi(desiredPort)
	if err != nil || i < 1 || i > 65535 {
		return "8080"
	}
	return desiredPort
}

func printStartupInfo(port string, sm *scenarios.Manager, authCfg *auth.Config) {
	fmt.Printf("\nStarting payloadBuddy %s on http://localhost:%s\n", Version, port)
	authCfg.PrintInfo()
	printUsageExamples(port, sm, authCfg)
}

func printUsageExamples(port string, sm *scenarios.Manager, authCfg *auth.Config) {
	url := func(path string) string { return authCfg.ExampleURL("http://localhost:" + port + path) }

	fmt.Println("\nAvailable endpoints:")
	fmt.Printf("  %s\n", url("/rest_payload"))
	fmt.Printf("  %s\n", url("/stream_payload"))
	fmt.Printf("  %s\n", url("/paginated_payload"))
	fmt.Printf("  %s\n", url("/openapi.json"))
	fmt.Printf("  %s\n", url("/swagger"))

	fmt.Println("\nRest Payload examples:")
	fmt.Printf("  %s\n", url("/rest_payload"))
	fmt.Printf("  %s\n", url("/rest_payload?count=5000"))

	fmt.Println("\nPagination examples (ServiceNow Data Stream compatible):")
	fmt.Printf("  %s\n", url("/paginated_payload?limit=100&offset=0&servicenow=true"))
	fmt.Printf("  %s\n", url("/paginated_payload?page=2&size=50&servicenow=true"))
	fmt.Printf("  %s\n", url("/paginated_payload?scenario=peak_hours&servicenow=true"))

	fmt.Println("\nStreaming examples:")
	fmt.Printf("  %s\n", url("/stream_payload?count=1000&delay=100ms"))
	fmt.Printf("  %s\n", url("/stream_payload?scenario=peak_hours&servicenow=true"))
	fmt.Printf("  %s\n", url("/stream_payload?delay=50ms&strategy=random&batch_size=50"))

	printServiceNowScenarios(sm)
}

func printServiceNowScenarios(sm *scenarios.Manager) {
	fmt.Println("\nServiceNow test scenarios (compatible with both streaming and pagination):")
	for _, scenarioType := range sm.ListScenarios() {
		scenario := sm.GetScenario(scenarioType)
		usageContext := scenarioUsageContext(scenarioType)
		if scenario != nil && scenario.Description != "" {
			desc := scenario.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("  - %s: %s\n", scenarioType, desc)
		} else {
			fmt.Printf("  - %s\n", scenarioType)
		}
		if usageContext != "" {
			fmt.Printf("    %s\n", usageContext)
		}
	}
}

func scenarioUsageContext(scenarioType string) string {
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

// newHTTPServer builds the server with the timeouts appropriate for a payload
// testing tool. WriteTimeout is deliberately disabled: /stream_payload responses
// are unbounded by design, and a write deadline truncates them mid-array, leaving
// the client with an HTTP 200 and malformed JSON. Cancellation is handled by the
// streaming handler through the request context instead. ReadHeaderTimeout still
// bounds how long a client may take to send its request headers.
func newHTTPServer(port string, mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       120 * time.Second,
	}
}

func startHTTPServer(port string, mux *http.ServeMux) {
	server := newHTTPServer(port, mux)
	fmt.Println("\nPress Ctrl+C to stop the server")
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed to start: %v\n", err)
		os.Exit(1)
	}
}
