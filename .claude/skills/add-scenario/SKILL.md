---
name: add-scenario
description: How payloadBuddy scenarios are loaded, overridden, validated, and tested. Use when creating or editing a scenario JSON file, adding a built-in scenario, debugging a scenario that is not loading or not overriding, or changing delay behavior. Covers embedded vs user scenarios, the override rule, -verify validation, and the per-endpoint delay semantics.
---

# Working with scenarios

Scenarios simulate real ServiceNow performance characteristics. `SCENARIOS.md` is the
authority on the full JSON schema — this skill covers the mechanics that are easy to get
wrong.

## Where scenarios come from

Two sources, loaded by `internal/scenarios/manager.go` at startup:

- **Embedded** — `internal/scenarios/embedded/*.json`, compiled into the binary. The built-ins
  are `peak_hours`, `maintenance`, `network_issues`, and `database_load`.
- **User** — `$HOME/.config/payloadBuddy/scenarios/*.json`, loaded dynamically at startup.

**The override rule:** a user scenario replaces an embedded one when its `scenario_type`
matches. This is the single most common source of "my changes did nothing" — the field that
decides an override is `scenario_type`, not the filename. A user file named `peak_hours.json`
whose `scenario_type` says something else is registered as an additional scenario instead of
replacing the built-in.

The manager is the source of scenario-based defaults — `count`, `batch_size`, ServiceNow mode,
and max limits — and is thread-safe for concurrent lookup. Get one via `NewManager()`.

## Delay semantics differ per endpoint

All scenarios work with both endpoints, but they mean different things, and a scenario that
behaves correctly on one can look broken on the other:

- **`/stream_payload`** applies delays **per item**, with progressive and periodic effects
  accumulating across the stream.
- **`/paginated_payload`** applies delays **per page request**, calculated from the page's
  position in the dataset.

So `database_load` ramps up item-by-item when streamed, and ramps up page-by-page when
paginated. When adding a scenario, decide what it should do in *both* contexts before
writing it — "works with streaming" is half an answer.

## Validation

`internal/scenarios/validator.go` validates every scenario against the JSON schema
(version 1.0.0) at load time, and on demand:

```bash
./payloadBuddy -verify scenario.json          # single file
./payloadBuddy -verify invalid.json && echo Valid || echo Invalid   # exit code is meaningful
```

Validation covers delay formats, version compatibility, and configuration parameters, and
reports errors per-field. Run `-verify` on any scenario file you write or edit before
assuming it loads. Get a `Validator` via `NewValidator()`.

Validation runs in two layers: the embedded `scenario_schema_v1.0.0.json` (checked via
`gojsonschema`) and then struct-level rules in `validator.go`. **Keep the two in sync** —
they drifted apart once already. If you add a field or constraint, update both.

## Tests to extend

A new or changed scenario needs coverage on both endpoints:

- `TestStreamingPayloadHandler_Scenarios` — scenario behavior while streaming.
- `TestPaginatedPayloadHandlerScenarios` — scenario compatibility with pagination.

Also assert the scenario-based defaults it contributes (count, batch size, ServiceNow mode),
since those are what callers actually observe when they pass only `?scenario=`.

For an embedded scenario, add the JSON under `internal/scenarios/embedded/` and confirm it is
picked up by a manager test — being on disk is not the same as being registered.

## Checklist

1. Write or edit the JSON; set `scenario_type` deliberately.
2. `./payloadBuddy -verify <file>` — clean before going further.
3. Decide and verify behavior on streaming *and* pagination.
4. Extend both scenario test suites.
5. Update `SCENARIOS.md` if the schema or the built-in set changed.
6. Run the `preflight` skill before committing.
