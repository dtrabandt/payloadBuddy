---
name: go-tdd
description: Test-driven development workflow and Go test patterns for payloadBuddy. Use when writing or modifying tests, adding a feature or delay strategy, refactoring a handler or package, or investigating test coverage. Covers the red-green-refactor cycle, table-driven test structure, httptest handler testing, testing auth-protected endpoints, and the test command reference.
---

# TDD for payloadBuddy

This project treats TDD as a core methodology, not a suggestion. Write the failing test
first; it defines the requirement and becomes the executable documentation of the behavior.

## The cycle

```bash
# 1. RED — write the failing test first
go test -v -run TestNewFeature ./...    # must fail: the feature does not exist yet

# 2. GREEN — write the minimum code that makes it pass
go test -v -run TestNewFeature ./...

# 3. REFACTOR — improve structure while the tests stay green
go test -v ./...
```

Rules that make the cycle worth doing:

- Start with the simplest failing test, not the complete one.
- Only refactor when everything is green. Re-run tests after each refactoring step.
- Never adjust a test to match code that turned out to behave differently than intended —
  decide which one is wrong first.
- Report failures with their output. A red run is never summarized as green.

## Test commands

```bash
go test -v ./...                     # everything (./... is mandatory — this is a multi-package module)
go test -v -run TestSpecific ./...   # one test pattern
go test -v -race ./...               # race detection, as CI runs it
go test -v -cover ./...              # coverage summary
go test -v -short ./...              # skip long-running tests
```

`./...` is not optional. The module is split across `internal/auth`, `internal/handlers`,
`internal/openapi`, and `internal/scenarios`; a bare `go test -v` in the repo root only
exercises the root package and silently skips every one of them.

The full suite should stay fast — under about 5 seconds. If a new test needs real time to
pass, gate it behind `testing.Short()`.

## Table-driven tests

The default shape for anything with more than one input:

```go
func TestNewDelayStrategy(t *testing.T) {
    tests := []struct {
        name     string
        strategy string
        expected DelayStrategy
    }{
        {"exponential strategy", "exponential", ExponentialDelay},
        {"invalid strategy", "invalid", FixedDelay},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation here
        })
    }
}
```

Name subtests after the behavior they pin down, so a failure reads as a sentence.

## Handler tests

Use `httptest.NewRecorder()` and drive the handler end-to-end — status code, headers, and
decoded JSON body. Validate the response structure, not just that a 200 came back.

Testing auth is two cases, and both are worth having:

```go
// unauthenticated
cfg := auth.Setup(false, "", "")

// protected — wrap the handler and assert 401 without credentials, 200 with them
handler := auth.Setup(true, "user", "pass").Middleware(h)
```

Remember the project's split: `/rest_payload`, `/stream_payload`, and `/paginated_payload`
are protected when `-auth` is on, while `/swagger` and `/openapi.json` stay public. A test
that asserts auth on a documentation endpoint is asserting a bug.

## What to cover

- Happy path, edge cases, and error conditions — not just the happy path.
- Boundary values: `count=0`, `count=1`, `count=1000000`, and past the max limit.
- Parameter validation and malformed input.
- Scenario behavior on both endpoints. The existing suites to extend are
  `TestStreamingPayloadHandler_Scenarios` (delays per item) and
  `TestPaginatedPayloadHandlerScenarios` (delays per page). See the `add-scenario` skill.
- OpenAPI specs: every endpoint documented, parameters matching the implementation,
  security scheme correct.

CI enforces a hard 80% total coverage floor. Coverage is a floor, not a goal — a covered
line that asserts nothing is worse than an uncovered one, because it reads as tested.

## Unit vs integration

Unit tests exercise one function with no network or file I/O. Integration tests drive the
HTTP handlers through the middleware stack and assert on real JSON. This project needs both;
prefer the unit test when it can pin the behavior down, and reach for the integration test
when the thing you care about is the wiring.

## Design principles

Broader background on the Clean Code and Unix-philosophy conventions this codebase follows
is in [references/design-principles.md](references/design-principles.md). Read it when
making structural decisions — naming, splitting functions, adding a package, or judging
whether a change fits the project's scope.
