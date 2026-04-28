# Refactoring Plan — Go Modernization

Target: Go 1.26.2 standards (structure + code)

## Tier 1 — Bug fixes (zero logic risk)

- [x] 1.1 Fix broken cursor stubs in `paginated_payload_handler.go`
- [x] 1.2 Annotate silent end-of-stream writes in `streaming_payload_handler.go`

## Tier 2 — Modern Go language features

- [x] 2.1 Update `go.mod` to Go 1.26.2 + add toolchain directive
- [x] 2.2 Update CI action versions (setup-go, cache, golangci-lint) + Go version
- [x] 2.3 Replace `log.Printf` with `log/slog` in `scenario_manager.go`
- [x] 2.4 Use Go 1.22 method-routing (`"GET /path"`) in `registerPlugins()`
- [x] 2.5 Replace `for i := 0; i < n; i++` with `for i := range n` (Go 1.22)
- [x] 2.6 Use `slices` package where applicable

## Tier 3 — Package restructure

- [x] 3.1 Introduce `internal/` layout:
  - `internal/auth/`
  - `internal/scenarios/`
  - `internal/handlers/`
  - `internal/openapi/`
- [x] 3.2 Replace `init()`-based plugin auto-registration with explicit wiring in `main.go`
- [x] 3.3 Eliminate mutable globals (`plugins`, auth vars) behind constructors

## Tier 4 — Polish

- [x] 4.1 Raise CI coverage floor to 80% — total coverage is 82.2%
- [x] 4.2 Add `.golangci.yml` with curated linter set
- [x] 4.3 Cursor round-trip tests added in `internal/handlers/paginated_test.go`
