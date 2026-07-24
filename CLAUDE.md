# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

payloadBuddy is a single-binary HTTP server for testing REST clients — primarily ServiceNow
integrations — against large payloads, streamed responses, and simulated degraded
performance.

Authoritative documentation lives elsewhere; prefer it over restating:
[CONTRIBUTING.md](CONTRIBUTING.md) (human contributor guide),
[README.md](README.md) (user-facing API reference and ServiceNow integration guide),
[SCENARIOS.md](SCENARIOS.md) (scenario schema), [DEPLOYMENT.md](DEPLOYMENT.md),
and `.github/workflows/` (the real definition of CI).

## Development Commands

```bash
go build -o payloadBuddy                       # build
./payloadBuddy                                 # run without authentication
./payloadBuddy -auth                           # run with auto-generated credentials
./payloadBuddy -auth -user=admin -pass=secret  # run with custom credentials
./payloadBuddy -verify scenario.json           # validate a scenario file

go test -v ./...                               # all tests
go test -v -race ./...                         # as CI runs them
go test -v -run TestRestPayloadHandler ./...   # one test pattern
go mod tidy                                    # clean up dependencies
```

`./...` is mandatory — this is a multi-package module, and a root-only `go test` silently
skips everything under `internal/`.

Formatting is automatic: a PostToolUse hook (`.claude/settings.json`) runs `gofmt -s -w` on
every `.go` file as it is edited. Use `gofmt -s -l .` to verify the whole tree; CI fails on
any output from it.

## Architecture

```
main.go                      Server bootstrap, explicit plugin wiring, flag parsing
internal/auth/auth.go        HTTP Basic Auth: Config, Setup(), Middleware(), PrintInfo()
internal/handlers/
  plugin.go                  PayloadPlugin interface: Path(), Handler(), OpenAPISpec()
  rest.go                    /rest_payload — single large response, up to 1,000,000 objects
  streaming.go               /stream_payload — fixed/random/progressive/burst delays
  paginated.go               /paginated_payload — limit/offset, page/size, cursor
  docs.go                    /openapi.json — OpenAPI 3.1.1, built via Init(allPlugins, authCfg)
  swagger.go                 /swagger — interactive Swagger UI (CDN assets carry SRI hashes)
  helpers.go                 Shared handler utilities, incl. queryParser — check here first
internal/openapi/types.go    OpenAPI 3.1.1 structs (the only place they are defined)
internal/scenarios/
  manager.go                 Manager, NewManager() — loads embedded + user scenarios
  validator.go               Validator, NewValidator() — schema + struct-level rules, -verify
  types.go                   Scenario data structures
  embedded/                  Built-in scenarios compiled into the binary
```

Four invariants. Breaking one of these is a bug even when the tests pass:

1. **Plugins are wired explicitly** in the slice `main()` builds. No `init()` registration,
   no mutable plugin globals. Routes use Go 1.22 method-routing (`"GET /path"`).
2. **Auth is per-route middleware.** `auth.Setup(enabled, user, pass)` returns a `Config`;
   `authCfg.Middleware(handler)` wraps data endpoints only. `/swagger` and `/openapi.json`
   are registered without it and stay public by design, so API docs remain reachable while
   `/rest_payload`, `/stream_payload`, and `/paginated_payload` are protected under `-auth`.
   Credential comparison is constant-time — do not hand-roll a replacement.
3. **User scenarios override embedded ones by `scenario_type`**, not by filename. User
   scenarios load from `$HOME/.config/payloadBuddy/scenarios/*.json`.
4. **OpenAPI structs live only in `internal/openapi/types.go`.** Handlers import them.

Scenarios apply to both endpoints but mean different things: streaming applies delays **per
item**, pagination applies them **per page request**. ServiceNow mode generates realistic
record structures (sys_id, incident numbers, states) and the pagination endpoint is
compatible with ServiceNow Data Stream actions.

## Conventions

- **Test first.** Red → green → refactor. CI enforces an 80% total coverage floor.
- **Validate query params, don't silently default.** Parse through `newQueryParser(r)`
  (`internal/handlers/helpers.go`), check `q.Err()` once, and return **HTTP 400** before
  writing any output. Silently substituting a default for bad input hides client bugs behind
  200s — that is how the panics found in the 2026-07-24 review reached the handlers. New
  endpoints follow this shape and document a `400` in their `OpenAPISpec()`.
- Table-driven tests; `httptest.NewRecorder()` for handlers.
- `log/slog`, not `log.Printf`.
- Go 1.22+ idioms already adopted throughout: method routing, `for i := range n`, `slices`.
- Small single-purpose functions with intention-revealing names (`setupPort()`,
  `NewPaginatedHandler()`).
- Dependencies are injected via constructors and struct fields, never package-level globals.
- Run the `preflight` skill before committing. Feature PRs target `develop`, not `main`.

## Skills

| Skill | Use it when |
|---|---|
| `go-tdd` | Writing tests, adding a feature, refactoring — TDD cycle and test patterns |
| `add-endpoint` | Adding an endpoint or `PayloadPlugin` |
| `add-scenario` | Creating, editing, or debugging a scenario JSON file |
| `preflight` | Before committing or opening a PR — runs the full CI gate locally |
