---
name: add-endpoint
description: Step-by-step workflow for adding a new HTTP endpoint to payloadBuddy as a PayloadPlugin. Use when adding an endpoint, route, or handler, or when implementing the PayloadPlugin interface. Covers the plugin interface, explicit wiring in main.go, the authentication decision, OpenAPI documentation, and which existing helpers to reuse.
---

# Adding an endpoint

Endpoints are plugins. Adding one means implementing `PayloadPlugin` and wiring it in
explicitly — there is no auto-registration to lean on.

## Order of work

Test first (see the `go-tdd` skill for the cycle and test patterns).

**1. Write the failing test.** `internal/handlers/<name>_test.go`. Assert the plugin's
`Path()`, drive `Handler()` through `httptest`, and assert the shape of `OpenAPISpec()`.

**2. Implement the plugin.** `internal/handlers/<name>.go`, satisfying the interface in
`internal/handlers/plugin.go`:

```go
type PayloadPlugin interface {
    Path() string
    Handler() http.HandlerFunc
    OpenAPISpec() OpenAPIPathSpec
}
```

If the endpoint needs the scenario manager, take it as a struct field the way
`StreamingPayloadPlugin{SM}` and `PaginatedPayloadPlugin{SM}` do, and inject it — do not
reach for a package-level global.

**3. Wire it in `main.go`.** Append the plugin to the slice that `main()` builds. This is
deliberate: there is no `init()` auto-registration and no mutable plugin global, so a plugin
that is not in that slice does not exist. Routes are registered with Go 1.22 method-routing,
i.e. `"GET /path"`, not bare `"/path"`.

**4. Decide authentication — deliberately.** The rule the project follows:

- Data endpoints wrap in `authCfg.Middleware(handler)` and are protected under `-auth`.
- Documentation endpoints (`/swagger`, `/openapi.json`) are registered *without* the wrapper
  and stay public, so API docs remain reachable while data is protected.

New endpoints that return payload data belong in the first group. If you are adding
something documentation-like, match the second and say so in the OpenAPI security section.

**5. Verify it surfaces.** The plugin's `OpenAPISpec()` flows into `/openapi.json` via
`DocumentationPlugin`, which is initialized with the full plugin slice and the auth config
(`Init(allPlugins, authCfg)`). Start the server and confirm the endpoint appears in both
`/openapi.json` and the `/swagger` UI. If it is missing from either, step 3 was skipped.

**6. Update the docs.** `README.md` carries the user-facing API reference — new endpoint,
its query parameters, and a curl example. Startup output examples use curl format so they
can be pasted into ServiceNow Flow Actions.

**7. Run the gate.** See the `preflight` skill before committing.

## Reuse rather than reinvent

- `internal/handlers/helpers.go` — shared handler utilities; check here before writing a new
  parameter parser or response writer.
- `internal/openapi/types.go` — the OpenAPI 3.1.1 structs. All spec types live here; do not
  define a parallel set in the handlers package.
- `internal/handlers/plugin.go` — the interface itself.
- `internal/auth` — `auth.Setup()` and `Config.Middleware()`; never hand-roll credential
  comparison, the existing one is constant-time on purpose.
- `internal/scenarios` — `NewManager()` for scenario lookup and defaults.

## Existing plugins as reference

| Plugin | Path | Shape |
|---|---|---|
| `RestPayloadPlugin` | `/rest_payload` | single large response, up to 1,000,000 objects |
| `StreamingPayloadPlugin{SM}` | `/stream_payload` | streamed, per-item delays, scenarios |
| `PaginatedPayloadPlugin{SM}` | `/paginated_payload` | limit/offset, page/size, cursor |
| `DocumentationPlugin` | `/openapi.json` | public, built from all plugin specs |
| `SwaggerUIPlugin` | `/swagger` | public, interactive UI |

Pick the closest one and follow its structure.
