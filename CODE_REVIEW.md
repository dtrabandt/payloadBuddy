# Code Review Findings — 2026-07-24

Full-codebase review of `develop` @ `4f9a63f`. All findings were verified by running the
code (live server + linters), not inferred from reading alone.

> **Status:** #1–#6, #11, #12, #13 and #17 are **fixed** on `feature/code_review`, each with a
> regression test; every original repro was re-run against a live server. The remaining open
> findings are #7–#10, #14–#16 and #18–#21, plus the test-quality items at the end of this
> document. Resolved entries are marked ✅ inline.

**Baseline at time of review:** `go vet ./...` clean · `gofmt -s -l .` clean ·
`go test ./... -cover` passes — total **82.7%** (CI floor 80%; main 62.0%, handlers 81.9%,
scenarios 85.2%, auth 97.8%, `internal/openapi` untested).

**Suggested order of work:** fix the lint config (#7) first so CI actually enforces
anything, then the two panics (#1, #2) and the cursor clamp (#3) with tests, then
`WriteTimeout` (#4). Coverage headroom is thin — add boundary tests before any refactor
that could shed coverage.

---

## HIGH

### ✅ 1. Unauthenticated panic: `batch_size=0` → integer divide by zero

`internal/handlers/streaming.go:124` — `if i%batchSize == 0`. `batchSize` comes straight
from `getIntParam(r, "batch_size", ...)` with no validation.

Verified: `GET /stream_payload?count=5&batch_size=0` produced
`http: panic serving 127.0.0.1:47758: runtime error: integer divide by zero` plus a full
stack trace on stdout; connection aborted mid-response (curl exit 18). The process
survives only because `net/http` recovers per-connection.

**Fix:** reject `batchSize < 1` with 400, or clamp to the default.

### ✅ 2. Unauthenticated panic: `strategy=random` with `delay=0` or a negative delay

`internal/handlers/streaming.go:188` → `internal/handlers/helpers.go:84`.
`secureRandInt63n(int64(baseDelay*2))` calls `crypto/rand.Int`, which panics when
`max <= 0`.

Verified: `?count=3&delay=0&strategy=random` and `?count=3&delay=-5ms&strategy=random`
both panic with `crypto/rand: argument to Int is <= 0`. `TestApplyDelay_ZeroBaseDelay`
only exercises `FixedDelay`, so this is untested.

**Fix:** return early when `baseDelay <= 0` in the `RandomDelay` case, and reject negative
durations in `getDurationParam`.

### ✅ 3. Cursor pagination bypasses the 1000-item page cap — 182 MB / 32 s from one GET

`internal/handlers/paginated.go:100-124`. The `page` branch clamps `size` to 1..1000
(:109) and the `offset` branch clamps `limit` (:119), but the **`cursor` branch does no
clamping at all**. `parseCursor` (:198) falls back to the raw `limit` query param whenever
the cursor is undecodable or its embedded limit is out of range.

Verified: `GET /paginated_payload?total=1000000&limit=1000000&cursor=notbase64!!&servicenow=true`
→ HTTP 200, **1,000,000 items, 182,666,989 bytes, 32.25 s** (`servicenow=true` triggers 32
`crypto/rand` calls per item, so 32M syscalls). A few concurrent requests will OOM the host.

**Fix:** clamp `pageSize` to 1..1000 once, after the switch.

### ✅ 4. `WriteTimeout: 30s` silently truncates every non-trivial stream into invalid JSON

`main.go:147` vs `internal/handlers/streaming.go`. The server-wide write deadline kills
long streams mid-flight.

Verified: `GET /stream_payload?count=4000&delay=10ms` returned **HTTP 200 with 2,901 of
4,000 items and no closing `]`** at exactly 30.06 s. The documented default
(`count=10000`, `delay=10ms` ≈ 100 s) can *never* complete. This defeats the tool's primary
purpose, and clients see a 200 with malformed JSON rather than an error.

**Fix:** `WriteTimeout: 0` — the handler's `ctx.Done()` handling (streaming.go:83-88,
209-214) is already correct and is the right cancellation mechanism. Keep
`ReadHeaderTimeout`. This went unnoticed because streaming tests all use
`httptest.NewRecorder()`, which has no write deadline.

---

## MEDIUM

### ✅ 5. `/rest_payload` silently ignores invalid `count`

`internal/handlers/rest.go:27-31`. Verified: `count=-5`, `count=abc`, and `count=99999999`
all return HTTP 200 with the 10,000-item default. The OpenAPI spec (:57-58) advertises
min 1 / max 1000000, and both sibling endpoints return 400. Silent fallback hides client
bugs. **Fix:** 400 on unparseable/out-of-range.

### ✅ 6. Released binaries never get their version stamped

`.github/workflows/release.yml:101` uses `-X main.version=...`, but `main.go:16` declares
`var Version` (capital V). `-X` against a nonexistent symbol is silently ignored, so
**every published binary reports `0.3.0`** regardless of tag. **Fix:** `-X main.Version=`.

### 7. `.golangci.yml` is invalid for golangci-lint v2 — the CI lint job cannot be passing

Verified with golangci-lint v2.1.6:

- `config verify` → `additional properties 'exclude-rules' not allowed` (issues) and
  `additional properties 'linters-settings' not allowed`
- `run` → `can't load config: gofmt is a formatter` (exit 3, no analysis performed)

`.github/workflows/test.yml` pins `version: latest`, which resolves to v2.x. Also
`gosimple` (`.golangci.yml:7`) no longer exists in v2 (folded into staticcheck).

**Fix:** migrate to `linters.settings`, `linters.exclusions.rules`, a `formatters:` block
for gofmt/goimports, and drop `gosimple`.

Running with a corrected config surfaced 44 real issues, including #8 below,
`main_test.go:142` `isValidHTTPPath` unused, and the broken test noted in coverage gaps.

### 8. Scenario `delay_strategy` is a dead assignment — custom scenarios can't use random/progressive/burst

`internal/handlers/streaming.go:155` — `strategy = calculatedStrategy` is flagged by both
ineffassign and staticcheck SA4006. After that assignment `strategy` is never read; the
`switch strategy` block only runs in the `sm == nil || scenario == ""` branch. So a user
scenario with `scenario_type: "custom"` and `delay_strategy: "progressive"` always gets a
flat `base_delay`.

**Fix:** have `GetScenarioDelay` return a base delay and apply the strategy switch on all
paths.

### 9. Pagination scenarios hardcode itemIndex 0, so all documented position-based behavior is absent

`internal/handlers/paginated.go:89` — `sm.GetScenarioDelay(scenario, 0)`. Consequences
from `manager.go:68-83`: `database_load` degradation is `itemIndex/100*10ms` → always
**0**, so page 1 and page 500 get identical delays; `maintenance` hits
`itemIndex%500==0` → **every** page gets the 2 s spike. CLAUDE.md claims pagination applies
"position-based calculations".

**Fix:** compute `startIndex` before the delay and pass it as `itemIndex`.

### 10. Cursor pagination drops the client's page size and accepts negative offsets

`internal/handlers/paginated.go:198-224`. `createCursor` hardcodes `Limit: 100` and
`parseCursor` prefers the cursor's embedded limit — verified that
`?total=1000&limit=500&cursor=<createCursor(0)>` returns **100** items. Separately, a
cursor with a negative id is accepted: `{"id":-500,"limit":100}` →
`{"id":-499,"value":"Item -499"}`.

**Fix:** clamp `cd.ID` to >= 0 and carry the effective limit into generated cursors.

> Partially fixed alongside #3: `parseCursor` now clamps `cd.ID` to >= 0, so negative item
> IDs are gone. The dropped page size is still open — `createCursor` continues to hardcode
> `Limit: 100`.

### ✅ 11. Validator rejects valid scenario files (diverges from the shipped JSON Schema)

`internal/scenarios/validator.go:195` and `:202`. `consecutive_error_limit` and
`metrics_interval` are optional-with-defaults in
`embedded/scenario_schema_v1.0.0.json`, but the Go validator errors on their zero value. A
file containing `"performance_monitoring": {"enabled": true}` fails `-verify` and is
silently dropped at startup by `loadUser`.

**Fix:** validate only when non-zero, or apply the schema defaults first.

### ✅ 12. The embedded JSON Schema is never used

`internal/scenarios/embedded/scenario_schema_v1.0.0.json` is embedded but explicitly
skipped at `manager.go:161`, and `github.com/xeipuuv/gojsonschema` is a direct go.mod
dependency used **only** in `internal/handlers/rest_test.go:38`. All validation is
hand-rolled and has already drifted (see #11; the schema also enforces `batch_size` max
10000 and `additionalProperties: false`, neither of which the code checks).

**Fix:** validate against the embedded schema, or delete it so it stops advertising
unenforced rules.

---

## LOW

13. ✅ **`validateDelayFormat` regex is mis-anchored** — `validator.go:139`: `^(...)|\d+$`
    applies `^` only to the first alternative and `$` only to the second, so `"100msJUNK"`
    and `"garbage100"` both match (caught downstream by `ParseDelay`, so no live impact),
    and `"-5"` matches `\d+$` → `ParseDelay` returns `-5ms`. The same broken pattern is
    copied into the JSON Schema. **Fix:** `^\d+(\.\d+)?(ns|us|μs|ms|s|m|h)$|^\d+$`.
14. **Modulo bias in credentials** — `internal/auth/auth.go:93`: `charset[b[i]%62]`; 256
    mod 62 = 8, so the first 8 charset chars are ~25% more likely. Negligible for a dev
    tool; `rand.Text()` (Go 1.24+) is a one-line fix.
15. **`-user`/`-pass` without `-auth` silently leaves the server open** — `main.go:36` →
    `auth.go:22-24` returns early when disabled, with no warning.
16. **GH Actions script injection via tag name** — `release.yml:61` interpolates
    `${{ steps.version.outputs.version }}` (from `GITHUB_REF`) directly into an
    `awk`/shell command in a `contents: write` job. Pass via `env:` instead.
17. ✅ **Swagger UI pulls three assets from unpkg.com with no SRI** — `docs.go:127,136,137`,
    served unauthenticated. The handler itself is a `const` with zero templating, so there
    is no XSS/injection surface in the handler. Add `integrity`/`crossorigin`, or vendor
    via `go:embed`. Also `Content-Type` should be `text/html; charset=utf-8`
    (`docs.go:121`).
18. **Manual `Transfer-Encoding: chunked`** — `streaming.go:68`; net/http manages this
    itself.
19. **`http.Error` after the body has started** — `streaming.go:106`, after `[\n` and items
    are already flushed. Unreachable today, but would inject plain text into the JSON
    stream and log "superfluous WriteHeader".
20. **Unknown `?scenario=` values silently accepted** — `paginated.go:62` /
    `streaming.go:44` → `manager.go:89` returns generic defaults with HTTP 200 for
    `?scenario=typo`.
21. **`ValidateScenarioFile` calls `os.Exit(1)`** — `validator.go:104-116`, and duplicates
    the `os.Stat` that `ValidateScenarioFileContent` already performs. Library code
    shouldn't exit; this is also why it needs subprocess tests to cover.

---

## Areas that are genuinely fine

- **Context cancellation in streaming is correct.** `select` on `ctx.Done()` at the top of
  the loop (`streaming.go:83`) and inside `applyDelay` (`:209-214`). No goroutines are
  spawned anywhere, so there are no leak paths; the `time.After` timer lives at most
  `delay`.
- **Auth middleware is sound.** Both comparisons are computed before branching (no
  short-circuit timing leak), `crypto/subtle` is used correctly, and both failure paths
  emit byte-identical responses. Route wiring at `main.go:51-57` correctly leaves only
  `/swagger` and `/openapi.json` public; all three data endpoints are wrapped. Go 1.22
  method routing 405s non-GET before any handler runs.
- **`Manager` concurrency.** The map is written only inside `NewManager()` and is read-only
  afterwards, so the missing mutex is safe despite the "thread-safe" wording in CLAUDE.md.
  CI runs `-race`.
- **User-scenario file handling.** `filepath.WalkDir` doesn't descend symlinked
  directories; the `HasPrefix` containment check at `manager.go:197` is redundant but
  harmless. `$HOME/.config` is a boundary the user already controls.
- **No integer overflow is reachable.** `strconv.Atoi` rejects out-of-range values, and
  every `min(startIndex+pageSize, totalCount)` path was traced — `endIndex` only equals
  `totalCount` when `startIndex` is already within `pageSize` of it, so `actualSize` can't
  blow up from arithmetic (only from the missing cursor clamp, #3).

---

## Test coverage gaps

- ✅ **Untested panics:** no test passes `batch_size=0`/negative (#1) or `strategy=random`
  with `delay=0` (#2). — covered by `TestStreamingPayloadHandler_InvalidBatchSize`,
  `TestStreamingPayloadHandler_RandomStrategyZeroDelay` and
  `TestApplyDelay_RandomStrategyNonPositiveBase`.
- ✅ **No oversized-`limit` cursor test** (#3). — covered by
  `TestPaginatedPayloadHandlerCursorPageSizeClamp` and
  `TestPaginatedPayloadHandlerNegativeCursorID`.
- ✅ **No test asserts a stream terminates with `]` under a real `http.Server`** — every
  streaming test uses `httptest.NewRecorder()`, which has no write deadline. This is
  precisely why #4 is invisible to CI. — covered by `TestStreamTerminatesUnderRealServer`
  in `main_test.go`, which runs a real `http.Server` with the timeouts from
  `newHTTPServer` and was confirmed to fail (`unexpected EOF` at 30 s) when the old
  `WriteTimeout` is restored.
- **`TestApplyDelay_NetworkIssuesScenario` (`streaming_test.go:343-357`) tests nothing.**
  staticcheck SA4004: the loop body ends in an unconditional `break`, so it runs one
  iteration, and `hitShortDelay = true` is set unconditionally right before it — the final
  assertion can never fail.
- ✅ `RestPayloadHandler` invalid-count branches (#5) and `parseCursor`'s negative-ID path are
  uncovered. — covered by `TestRestPayloadHandler_InvalidCount`,
  `TestRestPayloadHandler_ValidCount` and `TestPaginatedPayloadHandlerNegativeCursorID`.
- Dead/sloppy test code: ~~`main_test.go:142` `isValidHTTPPath` is unused~~ (removed);
  `main_test.go:101,108` accept `t` and never use it; `docs_test.go:112` shadows `exists`
  from :86.

Total coverage after this pass: **84.3%** (was 82.7%) — main 62.5%, handlers 85.8%,
scenarios 85.7%, auth 97.8%, `internal/openapi` still untested.
