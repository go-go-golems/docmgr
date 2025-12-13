---
Title: Diary
Ticket: IMPROVE-SCENARIO-LOGGING
Status: active
Topics:
    - testing
    - tooling
    - diagnostics
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: scenariolog/go.mod
      Note: Self-contained tool module (dependencies + toolchain)
    - Path: scenariolog/README.md
      Note: Build instructions + FTS5 enablement/fallback notes
    - Path: scenariolog/cmd/scenariolog/main.go
      Note: Cobra entrypoint (currently `init`; will grow)
    - Path: scenariolog/internal/scenariolog/db.go
      Note: SQLite open + pragmas (file-backed DB)
    - Path: scenariolog/internal/scenariolog/migrate.go
      Note: Schema migrations + FTS5 graceful fallback behavior
    - Path: scenariolog/internal/scenariolog/migrate_test.go
      Note: Migration tests (including degraded mode expectations)
    - Path: ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/tasks.md
      Note: Ticket checklist; keep in sync with implementation progress
    - Path: ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md
      Note: Step-by-step implementation plan
ExternalSources: []
Summary: "Implementation diary for building scenariolog (sqlite runs/steps + KV + artifacts + FTS fallback) and integrating it into the scenario suite."
LastUpdated: 2025-12-13T17:57:59.478120178-05:00
---

# Diary

## Goal

Capture the step-by-step implementation story for `scenariolog`: what changed, why, what worked, what failed, and what we learned, so future continuation and code review are fast.

## Context

This ticket is building a sqlite-backed “flight recorder” for scenario-style integration suites. The desired end state is:

- scenario runs/steps recorded in sqlite
- stdout/stderr captured as artifacts
- KV tags for provenance
- FTS-backed search over text logs, **with graceful degradation** if FTS5 is unavailable

Primary specs:

- `design-doc/02-generic-sqlite-scenario-logger-go-tool.md`
- `design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md`

## Quick Reference

Build commands (repo root):

```bash
go -C scenariolog test ./...
go -C scenariolog build -tags sqlite_fts5 -o /tmp/scenariolog-local ./cmd/scenariolog
```

Initialize DB:

```bash
/tmp/scenariolog-local init --db /tmp/scenario/.scenario-run.db
```

## Usage Examples

## Step 1: Create self-contained module + migrations + FTS fallback

This step turned the design into a real, buildable codebase by creating a standalone `scenariolog/` Go module with an initial Cobra CLI (`init`) and schema migrations. The first non-trivial edge we hit was **FTS5 availability**: some sqlite builds omit it, so migrations must degrade gracefully instead of failing hard.

**Commit (code):** 41d66c1dd66d8f8839b81d3612afd5b0e63745cb — "scenariolog: scaffold module, sqlite migrations, FTS fallback"

### What I did
- Created the self-contained tool directory and module:
  - `scenariolog/go.mod`
  - `scenariolog/README.md`
  - `scenariolog/cmd/scenariolog/main.go` (Cobra `init --db`)
- Implemented sqlite open + pragmas (`scenariolog/internal/scenariolog/db.go`)
- Implemented migrations with `PRAGMA user_version` (`scenariolog/internal/scenariolog/migrate.go`)
- Added migration unit tests (`scenariolog/internal/scenariolog/migrate_test.go`)
- Implemented **best-effort FTS5** creation:
  - if sqlite reports `no such module: fts5`, migrations still succeed and FTS features are disabled
  - if built with `-tags sqlite_fts5`, migrations create `log_lines_fts` successfully

### Why
- We need a concrete artifact we can integrate into the bash harness soon.
- FTS5 is a nice-to-have for “search warnings/errors”, but should not block the core “runs/steps/kv/artifacts” DB from being usable everywhere.

### What worked
- `go -C scenariolog test ./...` passes (degraded mode).
- `go -C scenariolog test -tags sqlite_fts5 ./...` passes (full FTS mode).
- `scenariolog init --db ...` runs migrations successfully.

### What didn't work
- Initially, migrations failed hard with:
  - `no such module: fts5`
  when running without the `sqlite_fts5` build tag.

### What I learned
- With `github.com/mattn/go-sqlite3`, enabling FTS5 is commonly done via the build tag `sqlite_fts5`.
- “FTS everywhere” is not a safe assumption; the migration layer needs an explicit degraded path.

### What was tricky to build
- Avoiding a “one-way” migration: a DB created on a non-FTS system should still be able to gain `log_lines_fts` later when run under an FTS-enabled build. We handle this by attempting `ensureFTS5` even when `user_version` is already at v1.

### What warrants a second pair of eyes
- The error-matching logic for FTS5 unavailability (`no such module: fts5`) is string-based; ensure it’s robust enough for our environments.
- Confirm `PRAGMA journal_mode=WAL` and busy timeout are appropriate defaults for our write patterns.

### What should be done in the future
- N/A (for this step).

### Code review instructions
- Start at `scenariolog/internal/scenariolog/migrate.go` (`Migrate`, `migrateToV1`, `ensureFTS5`).
- Run:
  - `go -C scenariolog test ./...`
  - `go -C scenariolog test -tags sqlite_fts5 ./...`

### Technical details
- Schema created in v1:
  - `scenario_runs`, `steps`, `commands`, `kv`, `artifacts`
  - `log_lines_fts` only when FTS5 is available

### What I'd do differently next time
- Add a small explicit “capabilities” query (e.g., `scenariolog capabilities`) early, so it’s obvious when FTS is enabled.

## Related

- `design-doc/02-generic-sqlite-scenario-logger-go-tool.md`
- `design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md`
