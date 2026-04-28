package handlers

import (
	"net/http"

	"github.com/dtrabandt/payloadBuddy/internal/openapi"
)

// PayloadPlugin is implemented by every HTTP endpoint registered with the server.
type PayloadPlugin interface {
	Path() string
	Handler() http.HandlerFunc
	OpenAPISpec() openapi.PathSpec
}
