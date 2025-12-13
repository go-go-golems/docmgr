---
Title: Generic sqlite scenario logger (Go tool)
Ticket: IMPROVE-SCENARIO-LOGGING
Status: active
Topics:
    - testing
    - tooling
    - diagnostics
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Scenario harness entrypoint (primary integration target)
    - Path: test-scenarios/testing-doc-manager/README.md
      Note: Scenario suite docs + execution conventions
    - Path: ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/01-scenario-suite-structured-logging-sqlite.md
      Note: Scenario-specific schema + bash wrapper baseline (this doc generalizes it)
    - Path: internal/workspace/sqlite_schema.go
      Note: Existing Go sqlite patterns (pragmas, connection management, quoting helper)
ExternalSources: []
Summary: "Design for a reusable Go package + CLI that logs scenario runs/steps/commands into sqlite with captured stdout/stderr artifacts, reusable beyond docmgr."
LastUpdated: 2025-12-13T16:43:13.821046789-05:00
---

# Generic sqlite scenario logger (Go tool)

## Executive Summary

We want **structured, queryable execution logs** for scenario-style test harnesses (like `test-scenarios/testing-doc-manager`) without committing every harness to bespoke bash + `sqlite3` glue.

This doc proposes a reusable Go tool (working name: **`scenariolog`**) that:

- **Wraps** “units of work” (steps/scripts) and “commands” (child processes) and writes metadata to a **sqlite DB**
- **Captures** stdout/stderr to files (optionally teeing to the console) and stores **artifact paths** in sqlite
- Provides a small **query/report surface** (summary, failures, timings, warning search, export) so you don’t need to memorize SQL for common workflows

It intentionally generalizes the scenario-specific approach in design-doc #1 into a tool that can be reused for:

- other integration suites
- CI pipelines (artifact DB + logs)
- local automation/migration scripts (anything that is “a sequence of commands and checks”)

## Problem Statement

### What we have

Design-doc #1 already sketches a working MVP: a bash harness that uses `sqlite3` + redirections to store **runs/steps/commands** and per-step logs.

### What’s missing (and why a Go tool helps)

Bash + `sqlite3` is fine for a single suite, but it gets painful when we want to reuse it:

- **Reusability**: every suite reimplements schema init, quoting, timestamps, file layout, and reporting.
- **Correctness + ergonomics**: quoting SQL in bash is brittle; handling signals + exit codes across process trees is easy to get subtly wrong.
- **Portability**: the `sqlite3` CLI isn’t always present; cross-platform behavior (mac/linux) differs in small ways.
- **Extensibility**: as soon as we want more metadata (argv, cwd, env, tags, artifacts, expected failures), bash glue grows fast.

We want a single, boring, predictable tool: “run steps, capture output, write sqlite”.

## Proposed Solution

### Deliverable

Add a small Go module in this repo (or extracted later) consisting of:

- **Library**: `internal/scenariolog` (schema, DB I/O, capture helpers)
- **CLI**: `cmd/scenariolog` (cobra-based) that exposes the library for bash/CI usage

The tool should be **generic**: it must not depend on `docmgr` output formats. Docmgr-specific parsing (doctor JSON, validation warnings, etc.) can be an optional plugin or a separate extractor command later.

### Core workflow (CLI)

The key primitive is a wrapper that executes a process and logs it:

- `scenariolog init --db /tmp/scenario/.scenario-run.db`
- `scenariolog run start --db ... --root-dir /tmp/scenario --suite testing-doc-manager`
- `scenariolog exec --db ... --run-id ... --kind step --step-num 1 --name 00-reset --log-dir .logs -- bash ./00-reset.sh /tmp/scenario`
- `scenariolog exec --db ... --run-id ... --kind step --step-num 2 --name 01-create-mock-codebase --log-dir .logs -- bash ./01-create-mock-codebase.sh /tmp/scenario`
- `scenariolog run end --db ... --run-id ... --exit-code $?`

Optionally, for fine-grain tracking inside a step:

- Export env vars: `SCENARIOLOG_DB`, `SCENARIOLOG_RUN_ID`, `SCENARIOLOG_STEP_ID`
- Provide a tiny bash helper that replaces direct calls:
  - `docmgr ...` → `scenariolog exec --kind cmd --parent-step-id "$SCENARIOLOG_STEP_ID" -- docmgr ...`

### Minimal integration with the existing scenario suite

Phase 1 should keep changes small:

- `run-all.sh` uses `scenariolog` to wrap **scripts** (step-level logging)
- no need to edit individual step scripts yet (command-level can come later)

This keeps the suite readable and preserves today’s control flow.

### Schema strategy

**Compatibility-first**: use the same mental model as design-doc #1 (runs → steps → commands), but implement it with prepared statements in Go and allow “generic metadata” growth.

Suggested schema (v1) is effectively design-doc #1:

- `scenario_runs`
- `steps`
- `commands` (optional)
- `diagnostics` (optional)

Small extensions that improve reuse without exploding scope:

- `tags` (or `kv`) table for arbitrary key/value metadata at run/step/command scope
- `artifacts` table for captured file paths with content hash/size (optional)

**Schema versioning**:

- Use `PRAGMA user_version` to track schema version.
- Provide `scenariolog migrate --db ...` (or auto-migrate on `init`/open).

### Output capture

The tool should capture stdout/stderr robustly:

- Write to `log_dir/step-02-stdout.txt` + `log_dir/step-02-stderr.txt`
- Optionally tee to console (to keep current UX) via `--tee` / `--tee-stdout` / `--tee-stderr`
- Store paths **relative** to `root_dir` in sqlite for portability (but allow absolute paths on disk)

Implementation sketch (library):

- record start timestamp (RFC3339Nano) and start monotonic `time.Now()`
- create log files
- `exec.CommandContext(ctx, argv[0], argv[1:]...)`
- use `errgroup.Group` to copy stdout/stderr concurrently (and optionally tee)
- wait, compute duration, record exit code, update row
- on signal/cancel, forward signals to the child process group and still finalize the DB row as best-effort

### Query/report surface

We still want raw SQL freedom, but the tool should cover the “80% queries”:

- `scenariolog summary --db ... --run-id ...`
  - exit code, duration, failed steps, slowest steps
- `scenariolog failures --db ...`
- `scenariolog timings --db ... --top 10`
- `scenariolog grep --db ... --pattern '\\bwarning\\b' --scope step-stdout`
- `scenariolog export junit --db ... --out report.xml` (future)

This is especially useful in CI where we want one artifact (DB) + a deterministic report.

## Design Decisions

### 1) CLI + library (not “just a library”)

- **Why**: bash harnesses and other repos can adopt it immediately; Go harnesses can embed it.
- **Trade-off**: slightly more packaging work; worth it for reuse.

### 2) sqlite via `database/sql`

- Use `database/sql` so swapping drivers is possible.
- This repo already imports `github.com/mattn/go-sqlite3` (CGO). Start there and consider a pure-Go driver later if needed.

### 3) Pragmas for file-backed DB

For durability + concurrency balance:

- `PRAGMA foreign_keys = ON;`
- `PRAGMA journal_mode = WAL;`
- `PRAGMA synchronous = NORMAL;`
- `PRAGMA busy_timeout = 5000;` (or DSN `_busy_timeout`)

### 4) Log artifacts stored as files, not DB blobs

- Keeps DB fast and small.
- Makes it easy to `less` logs and treat them as normal artifacts.
- Optional: store small previews (first N lines) in DB later for faster “grep-like” queries.

### 5) “Generic metadata” via tags/kv table

This keeps the core schema stable while enabling reuse:

- suite name/version
- git SHA
- host/user/OS/arch
- scenario parameters
- “expected failure” flags

### 6) Prepared statements over manual SQL quoting

Even though we have a `sqliteQuoteStringLiteral()` helper elsewhere, the scenario logger should primarily use **parameterized queries** to avoid injection/quoting bugs (especially if step names can contain weird characters).

## Alternatives Considered

### A) Keep everything in bash + `sqlite3` (design-doc #1 only)

- **Pros**: fastest MVP; no new binary.
- **Cons**: reinvents quoting, schema mgmt, concurrency, reporting; not easily reusable outside this one suite.

### B) JSONL logs + `jq`

- **Pros**: simple, ubiquitous tooling.
- **Cons**: ad-hoc analytics; harder to join across runs; harder to maintain “canonical queries”.

### C) OpenTelemetry tracing + exporter

- **Pros**: rich ecosystem, spans, timelines.
- **Cons**: heavy for this use case; the sqlite DB *is* the local analytics store we want.

## Implementation Plan

### 1) MVP: `init` + `exec` wrapper (step-level)

- [ ] Create package `internal/scenariolog` (schema DDL, open DB, pragmas, migrations via `PRAGMA user_version`)
- [ ] Implement `Logger` with `StartRun` / `EndRun` and `StartStep` / `EndStep`
- [ ] Implement capture/execution helper using `exec.CommandContext` + `errgroup` for stdout/stderr copying
- [ ] Add `cmd/scenariolog` (cobra) with:
  - [ ] `scenariolog init`
  - [ ] `scenariolog run start|end`
  - [ ] `scenariolog exec` (writes step rows + captures artifacts)

### 2) Integrate with `test-scenarios/testing-doc-manager`

- [ ] Update `run-all.sh` to wrap each script invocation with `scenariolog exec --kind step ...`
- [ ] Ensure output experience remains decent (stderr markers + optional tee)
- [ ] Document how to query DB + logs in the scenario README

### 3) Command-level granularity (optional)

- [ ] Add a bash helper `sl_cmd()` that wraps `docmgr` invocations with `scenariolog exec --kind cmd --parent-step-id ...`
- [ ] Populate `commands` table with argv, cwd, env (as needed)

### 4) Reporting + exports

- [ ] `scenariolog summary` (failures + slowest steps)
- [ ] `scenariolog export junit` (CI integration)
- [ ] Optional: `scenariolog diff --run a --run b`

## Open Questions

- **Tool location**: keep inside `docmgr` repo vs extract to a separate module once stabilized?
- **Binary name**: `scenariolog` vs `scenario-log` vs `scenario-db` (avoid collision with existing tools).
- **Driver**: stick with `mattn/go-sqlite3` (CGO) or add a pure-Go option?
- **Retention**: do we add a built-in cleanup/prune command for `.logs/` and old runs?
- **Streaming UX**: default to tee output to console or keep quiet and rely on logs + summary?
- **Schema**: do we want a single generic “spans” table eventually, or is runs/steps/commands sufficient forever?

## References

- `ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/01-scenario-suite-structured-logging-sqlite.md`
- `test-scenarios/testing-doc-manager/README.md`
- `test-scenarios/testing-doc-manager/run-all.sh`
- `https://www.sqlite.org/pragma.html`
