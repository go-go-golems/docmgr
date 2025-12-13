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
    - Path: scenariolog/README.md
      Note: Build instructions + FTS5 enablement/fallback notes
    - Path: scenariolog/cmd/scenariolog/main.go
      Note: Cobra entrypoint (currently `init`; will grow)
    - Path: scenariolog/go.mod
      Note: Self-contained tool module (dependencies + toolchain)
    - Path: scenariolog/internal/scenariolog/db.go
      Note: SQLite open + pragmas (file-backed DB)
    - Path: scenariolog/internal/scenariolog/migrate.go
      Note: Schema migrations + FTS5 graceful fallback behavior
    - Path: scenariolog/internal/scenariolog/migrate_test.go
      Note: Migration tests (including degraded mode expectations)
    - Path: scenariolog/internal/scenariolog/search.go
      Note: FTS search implementation
    - Path: scenariolog/internal/scenariolog/search_fts5_test.go
      Note: FTS5 test
    - Path: ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md
      Note: Step-by-step implementation plan
    - Path: ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/tasks.md
      Note: Ticket checklist; keep in sync with implementation progress
ExternalSources: []
Summary: Implementation diary for building scenariolog (sqlite runs/steps + KV + artifacts + FTS fallback) and integrating it into the scenario suite.
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

## Step 2: Add run lifecycle (run start/end)

This step added the minimal “run lifecycle” layer: the CLI can now start a run, print a generated `run_id`, and later finalize the run with an exit code and computed duration. This unlocks the next steps (steps/scripts and artifacts) because we now have a stable parent row to attach everything to.

**Commit (code):** 1ecfe225f95076b8b8df77fcd7821b62ca65566f — "scenariolog: add run start/end lifecycle"

### What I did
- Added run ID generation (`scenariolog/internal/scenariolog/ids.go`)
- Added `StartRun` / `EndRun` helpers and a unit test (`scenariolog/internal/scenariolog/run.go`, `run_test.go`)
- Extended the Cobra CLI with:
  - `scenariolog run start --db ... --root-dir ... [--suite ...] [--run-id ...]`
  - `scenariolog run end --db ... --run-id ... --exit-code ...`

### Why
- The schema is run-scoped; without a real run lifecycle, the harness can’t reliably attach step/command rows.

### What worked
- `go -C scenariolog test ./...` passes.
- `scenariolog run start` prints a run id and writes a row to `scenario_runs`.
- `scenariolog run end` finalizes the run and stores `exit_code` + `duration_ms`.

### What didn't work
- N/A (for this step).

### What I learned
- Keeping timestamps in RFC3339Nano makes duration computation straightforward, but parsing needs a small amount of tolerance for legacy/alternate formats.

### What was tricky to build
- Ensuring we can compute duration robustly even if timestamps are malformed or clocks jump (we clamp negative durations to 0 instead of failing).

### What warrants a second pair of eyes
- The `NewRunID` format: confirm it’s acceptable for downstream usage (sorting, readability, file naming).

### What should be done in the future
- N/A (for this step).

### Code review instructions
- Start at `scenariolog/internal/scenariolog/run.go` and `scenariolog/cmd/scenariolog/main.go` (run commands).
- Validate end-to-end:

```bash
go -C scenariolog build -o /tmp/scenariolog-local ./cmd/scenariolog
RUN_ID=$(/tmp/scenariolog-local run start --db /tmp/scenario-run-test.db --root-dir /tmp/scenario --suite test-suite)
/tmp/scenariolog-local run end --db /tmp/scenario-run-test.db --run-id \"$RUN_ID\" --exit-code 0
rm -f /tmp/scenario-run-test.db
```

### Technical details
- Inserts into `scenario_runs` on start.
- Updates `scenario_runs.completed_at`, `exit_code`, and `duration_ms` on end.

## Step 3: Exec step wrapper (capture stdout/stderr + artifacts + best-effort FTS indexing)

This step implemented the first “real” value of the tool: wrapping a step command so it gets recorded in sqlite and produces portable log artifacts. We write a `steps` row, capture stdout/stderr into files, insert `artifacts` rows with sha256/size, and (when available) index the artifacts into the FTS table for future search.

**Commit (code):** 9ac50c1f4f7314d6afbc24814f0e2144b4c056c8 — "scenariolog: exec step capture + artifacts"

### What I did
- Implemented `ExecStep` in `scenariolog/internal/scenariolog/exec_step.go`:
  - inserts step start row
  - runs the command with `exec.CommandContext`
  - concurrently copies stdout/stderr into files via `errgroup`
  - finalizes the step row with exit code + duration
  - hashes artifacts and inserts `artifacts` rows
  - best-effort indexes lines into `log_lines_fts` (no-op if missing)
- Added artifact insertion helper: `scenariolog/internal/scenariolog/artifacts.go`
- Added FTS indexing helper: `scenariolog/internal/scenariolog/fts.go`
- Added Cobra `exec` command:
  - `scenariolog exec --db ... --run-id ... --root-dir ... --log-dir ... --step-num ... --name ... -- <cmd...>`
  - propagates the wrapped command’s exit code
- Added unit tests for `ExecStep` that avoid shell init file noise.

### Why
- This is the core mechanism the scenario harness will use: each step script becomes one `scenariolog exec ... bash ./NN-step.sh "$ROOT_DIR"` call.

### What worked
- `scenariolog exec` produces files `step-NN-stdout.txt` / `step-NN-stderr.txt` and inserts corresponding rows into sqlite.
- Non-zero exit codes are preserved and returned by the CLI process.

### What didn't work
- N/A (for this step).

### What I learned
- `bash -lc` can leak user shell init into stderr; tests should use `bash --noprofile --norc -c` for determinism.

### What was tricky to build
- Ensuring we still finalize the step row even when file creation or command startup fails (we best-effort write an exit code of 127).

### What warrants a second pair of eyes
- Signal/process-group handling is not implemented yet (CTRL-C may leave child processes running). We need to harden this before using it heavily in CI.

### What should be done in the future
- Add explicit process-group handling and more robust cancellation semantics (tracked in remaining Phase 2 hardening tasks).

### Code review instructions
- Start at `scenariolog/internal/scenariolog/exec_step.go` and `scenariolog/cmd/scenariolog/main.go` (`exec`).
- Run:

```bash
go -C scenariolog test ./...
go -C scenariolog build -o /tmp/scenariolog-local ./cmd/scenariolog
DB=/tmp/scenario-run-test.db
RUN_ID=$(/tmp/scenariolog-local run start --db \"$DB\" --root-dir /tmp --suite test-suite)
/tmp/scenariolog-local exec --db \"$DB\" --run-id \"$RUN_ID\" --root-dir /tmp --log-dir . --step-num 1 --name demo -- bash --noprofile --norc -c 'echo out; echo err 1>&2; exit 3'
/tmp/scenariolog-local run end --db \"$DB\" --run-id \"$RUN_ID\" --exit-code 0
rm -f \"$DB\"
```

## Step 4: Add FTS-backed search command (library + CLI)

This step made the indexed log lines actually queryable from the CLI by adding an FTS search API and a `scenariolog search` command. Importantly, it respects degraded mode: if the `log_lines_fts` table doesn’t exist (FTS5 unavailable), search returns a clear error instead of silently lying.

**Commit (code):** 791ffd30c5083a3e7ca3d8e0595e73de241fffd6 — "scenariolog: add FTS search command"

### What I did
- Added `SearchFTS` in `scenariolog/internal/scenariolog/search.go`
- Added Cobra `search` command in `scenariolog/cmd/scenariolog/main.go`
- Added an FTS5-tagged test that indexes a file and queries for a keyword (`scenariolog/internal/scenariolog/search_fts5_test.go`)

### Why
- The whole point of indexing into sqlite is to make “find warnings/errors” a fast query instead of manual grep.

### What worked
- `go -C scenariolog test -tags sqlite_fts5 ./...` validates search end-to-end.

### What didn't work
- N/A (for this step).

### What warrants a second pair of eyes
- SQL query semantics: confirm `MATCH` + `run_id` filtering is correct and performant for our expected data sizes.

### Code review instructions
- Start at `scenariolog/internal/scenariolog/search.go` and the `search` command wiring in `scenariolog/cmd/scenariolog/main.go`.

## Related

- `design-doc/02-generic-sqlite-scenario-logger-go-tool.md`
- `design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md`
