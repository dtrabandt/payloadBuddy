# Design principles

Background reading for structural decisions in payloadBuddy. These are the conventions the
existing code already follows; match them rather than introducing a competing style.

## Clean Code

Robert C. Martin's practices, as they show up in this codebase.

**Meaningful names.** Functions have intention-revealing names: `setupPort()`,
`validateScenarioFile()`, `NewPaginatedHandler()`. Variables express purpose:
`sm` (scenario manager), `defaultServiceNowMode`, `maxCount`.

**Single responsibility.** Each file has a focused purpose — `internal/auth/auth.go` does
authentication, `internal/scenarios/manager.go` does scenario management. Functions do one
thing: `setupPort()` only handles port validation and defaults.

**Open/closed.** The `PayloadPlugin` interface allows extension without modification. New
endpoints are added without changing existing ones; the scenario system extends through JSON
configuration rather than code changes.

**Small functions.** Most functions are under 20 lines. Complex operations decompose into
smaller composable ones. This is what makes them testable in isolation.

**Comprehensive testing.** High coverage (CI floor: 80%), descriptive test names that
explain behavior, table-driven tests for multiple cases.

## Unix philosophy

**Do one thing well.** payloadBuddy tests REST client implementations against large payloads
and degraded-performance scenarios. It is not a general web server and not a monitoring
tool. Proposals that broaden the mission usually belong somewhere else.

**Work together.** Composes with other tools; standard input/output behavior works with
pipes and redirects; the command-line interface is scriptable.

**Text streams.** JSON output, HTTP text protocols, text-based configuration files,
human-readable log output.

**Small is beautiful.** Single-binary deployment, fast startup, no required external
dependencies, minimal resource usage.

**Leverage software tools.** Built on standard HTTP/REST, JSON, and OpenAPI rather than
bespoke formats.

In practice:

```bash
# Pipes and composition work naturally
curl "http://localhost:8080/stream_payload" | jq '.[0]'
payloadBuddy -verify scenario.json > validation.log

# Scriptable and automatable
payloadBuddy -port=9999 &
SERVER_PID=$!
curl "http://localhost:9999/rest_payload?count=100" > test_data.json
kill $SERVER_PID

# Standard exit codes and error handling
payloadBuddy -verify invalid.json && echo "Valid" || echo "Invalid"
```

## How these fit together

Clean Code guides implementation structure. The Unix philosophy keeps the tool composable
with existing workflows. TDD validates behavior and makes refactoring safe. The 80% coverage
floor is the mechanical backstop under all three.
