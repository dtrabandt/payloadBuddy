---
name: preflight
description: Run payloadBuddy's full CI quality gate locally before committing or opening a PR. Use before committing, before pushing, when preparing a pull request, or when asked whether changes will pass CI. Mirrors .github/workflows/test.yml exactly - formatting, vet, race tests, the 80% coverage floor, linting, security scan, and build.
---

# Preflight

Run this before committing. It reproduces `.github/workflows/test.yml` locally, in the same
order, so a green run here means a green run in CI.

## The gate

```bash
# 1. Formatting — must print nothing. CI fails on any output.
gofmt -s -l .

# 2. Vet
go vet ./...

# 3. Tests with race detection
go test -v -race -coverprofile=coverage.out ./...

# 4. Coverage floor — total must be >= 80%
go tool cover -func=coverage.out | tail -1

# 5. Lint
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./...

# 6. Security scan
gosec ./...        # go install github.com/securego/gosec/v2/cmd/gosec@latest

# 7. Build
go build -v ./...
```

Notes on each:

- **Formatting** is normally already satisfied — a PostToolUse hook runs `gofmt -s -w` on
  every `.go` file as it is edited. Step 1 is the check that the tree as a whole is clean,
  including files changed outside the editor.
- **`./...` is mandatory** everywhere. This is a multi-package module; a root-only command
  silently skips `internal/auth`, `internal/handlers`, `internal/openapi`, and
  `internal/scenarios`.
- **Coverage** is a hard CI gate at 80% total, not a target to approach. If a change drops
  below it, add the missing tests rather than lowering the bar.
- **Lint** uses the curated linter set in `.golangci.yml` (v2 schema: linters in `linters:`,
  `gofmt`/`goimports` in `formatters:`). CI runs the action with `--timeout=5m`; the `go run`
  form above needs no separate install. Verify config changes with
  `golangci-lint config verify` before relying on a run.
- **`gosec`** is not vendored — install it if missing, or note that it was skipped. Do not
  report the gate as passed with a step missing.

## Reporting

State the actual result of each step. If a step fails, show its output and say which step;
if a step was skipped (missing tool), say that explicitly. A run with a failing or skipped
step is not a passing preflight.

## After it is green

CI triggers on pull requests to `develop` or `main` and on pushes to `develop`.

- **Feature PRs target `develop`, not `main`.**
- Merges to `main` are for releases; tags matching `v*` trigger the cross-platform release
  build, which runs this same test suite as its quality gate before building.
- The release pipeline builds Linux, macOS, and Windows (amd64 + arm64), attaches archives
  and SHA256 checksums, and generates the changelog from `CHANGELOG.md`. Version comes from
  `-ldflags="-X main.version=v1.0.0"`, semantic versioning, pre-releases supported.

`.github/workflows/test.yml` and `release.yml` are the authority. If this skill and a
workflow file disagree, the workflow is right — fix this skill.
