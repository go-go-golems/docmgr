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
ExternalSources:
    - local:glaze-help-build-first-command-2025-12-13.txt
Summary: "Design for a reusable Go package + CLI that logs scenario runs/steps/commands into sqlite with KV tags, artifact metadata, and FTS-backed search."
LastUpdated: 2025-12-13T16:43:13.821046789-05:00
---

# Generic sqlite scenario logger (Go tool)

## Executive Summary

We want **structured, queryable execution logs** for scenario-style test harnesses (like `test-scenarios/testing-doc-manager`) without committing every harness to bespoke bash + `sqlite3` glue.

This doc proposes a reusable Go tool (working name: **`scenariolog`**) that:

- **Wraps** “units of work” (steps/scripts) and “commands” (child processes) and writes metadata to a **sqlite DB**
- **Captures** stdout/stderr to files (optionally teeing to the console) and stores **artifact records** (paths, hashes, sizes, kinds) in sqlite
- Adds **KV tags** (run/step/command-scoped) so runs explain themselves (git SHA, suite version, host, parameters)
- Adds **FTS search** over textual artifacts so “find warnings/errors” becomes a fast SQL query instead of ad-hoc `rg`
- Provides a small **query/report surface** (summary, failures, timings, search, export) so you don’t need to memorize SQL for common workflows

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

Add a small, self-contained Go tool under `scenariolog/` (and extractable later) consisting of:

- **Go module root**: `scenariolog/go.mod` (keeps dependencies/tooling isolated)
- **Library**: `scenariolog/internal/scenariolog` (schema, DB I/O, capture helpers)
- **CLI**: `scenariolog/cmd/scenariolog` (cobra-based, using Glazed patterns for structured output) that exposes the library for bash/CI usage

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
- `commands` (optional; can be added later if/when we need per-command granularity)
- `diagnostics` (optional; can be added later, ideally fed from JSON sources)

In addition, we are explicitly committing to:

- **`kv` tags** (arbitrary metadata at run/step/command scope)
- **`artifacts`** (captured files and other run outputs with metadata)
- **FTS5** search over textual artifacts (fast “find warnings/errors” queries)

**Schema versioning**:

- Use `PRAGMA user_version` to track schema version.
- Provide `scenariolog migrate --db ...` (or auto-migrate on `init`/open).

### KV tags table (run/step/command scoped)

This is the “flight recorder metadata” layer: it keeps the core schema stable while letting every run explain itself.

DDL sketch:

```sql
CREATE TABLE IF NOT EXISTS kv (
    kv_id INTEGER PRIMARY KEY AUTOINCREMENT,

    run_id TEXT NOT NULL,
    step_id TEXT,        -- nullable (scope = run if NULL)
    command_id TEXT,     -- nullable

    k TEXT NOT NULL,
    v TEXT NOT NULL,

    created_at TEXT NOT NULL DEFAULT (datetime('now')),

    -- Ensure we don't accidentally attach a command without its step
    CHECK (command_id IS NULL OR step_id IS NOT NULL),

    FOREIGN KEY (run_id) REFERENCES scenario_runs(run_id) ON DELETE CASCADE,
    FOREIGN KEY (step_id) REFERENCES steps(step_id) ON DELETE CASCADE,
    FOREIGN KEY (command_id) REFERENCES commands(command_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_kv_scope_key
    ON kv(run_id, step_id, command_id, k);

CREATE INDEX IF NOT EXISTS idx_kv_key ON kv(k);
CREATE INDEX IF NOT EXISTS idx_kv_run ON kv(run_id);
```

Notes:

- Values are stored as TEXT. If we need typed values later, add `v_json` or `v_int`/`v_real` columns.
- We can use this immediately for: `suite`, `suite_version`, `git_sha`, `git_dirty`, `hostname`, `user`, `docmgr_path`, `docmgr_version`, `ci=true`, etc.

### Artifacts table (captured files + metadata)

We store stdout/stderr as files, but we also want a normalized table that can hold “any artifact” (logs, generated files, archives, screenshots, traces).

DDL sketch:

```sql
CREATE TABLE IF NOT EXISTS artifacts (
    artifact_id INTEGER PRIMARY KEY AUTOINCREMENT,

    run_id TEXT NOT NULL,
    step_id TEXT,               -- nullable
    command_id TEXT,            -- nullable

    kind TEXT NOT NULL,         -- stdout, stderr, report, bundle, trace, etc.
    path TEXT NOT NULL,         -- preferably root_dir-relative for portability
    is_text INTEGER NOT NULL DEFAULT 1,

    size_bytes INTEGER,
    sha256 TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),

    CHECK (command_id IS NULL OR step_id IS NOT NULL),

    FOREIGN KEY (run_id) REFERENCES scenario_runs(run_id) ON DELETE CASCADE,
    FOREIGN KEY (step_id) REFERENCES steps(step_id) ON DELETE CASCADE,
    FOREIGN KEY (command_id) REFERENCES commands(command_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_artifacts_unique_path
    ON artifacts(run_id, step_id, command_id, kind, path);

CREATE INDEX IF NOT EXISTS idx_artifacts_run ON artifacts(run_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_kind ON artifacts(kind);
```

Notes:

- The `steps.stdout_path` / `steps.stderr_path` columns from design-doc #1 can either remain as convenience columns or be replaced by inserting stdout/stderr into `artifacts` and joining.
- `sha256`/`size_bytes` enables integrity checks and “bundle runs” workflows.

### FTS5: log search inside sqlite (line-oriented)

We want queries like “find `warning` in the last run” without shelling out to `rg`.

Approach:

- Keep raw logs as files in `.logs/` (artifacts are the source of truth).
- Ingest textual artifacts into an FTS5 table as **lines** so results carry a `line_num`.

DDL sketch (simple, self-contained):

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS log_lines_fts USING fts5(
    run_id UNINDEXED,
    artifact_id UNINDEXED,
    line_num UNINDEXED,
    text,
    tokenize = 'unicode61'
);

CREATE INDEX IF NOT EXISTS idx_artifacts_text_only ON artifacts(is_text);
```

Ingestion sketch:

- On `scenariolog exec` completion (or on `scenariolog index fts --run-id ...`):
  - read each text artifact (typically stdout/stderr)
  - split into lines
  - bulk insert rows into `log_lines_fts(run_id, artifact_id, line_num, text)`

Trade-offs / guardrails:

- Keep ingestion best-effort and bounded:
  - cap line length (e.g., 8–16KB) to avoid pathological lines
  - optional cap on total indexed bytes per artifact
  - skip binary-ish content (`is_text=0`)

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
- `scenariolog search --db ... --query 'warning OR error'` (FTS-backed; returns matches with artifact + line number)
- `scenariolog export junit --db ... --out report.xml` (future)

This is especially useful in CI where we want one artifact (DB) + a deterministic report.

## Design Decisions

### 1) CLI + library (not “just a library”)

- **Why**: bash harnesses and other repos can adopt it immediately; Go harnesses can embed it.
- **Trade-off**: slightly more packaging work; worth it for reuse.

### 2) sqlite via `database/sql`

- Use `database/sql` so swapping drivers is possible.
- Use `github.com/mattn/go-sqlite3` (CGO) so we can compile sqlite with optional features like FTS5.
- **FTS5 enablement**: build with the standard go-sqlite3 build tag `sqlite_fts5`.
- **Degraded mode**: if FTS5 isn't available at runtime (sqlite built without it), migrations and core logging still work; only FTS-backed search/indexing features are disabled/no-op.

### 3) Pragmas for file-backed DB

For durability + concurrency balance:

- `PRAGMA foreign_keys = ON;`
- `PRAGMA journal_mode = WAL;`
- `PRAGMA synchronous = NORMAL;`
- `PRAGMA busy_timeout = 5000;` (or DSN `_busy_timeout`)

### 4) Log artifacts stored as files, not DB blobs

- Keeps DB fast and small.
- Makes it easy to `less` logs and treat them as normal artifacts.
- We still add FTS by ingesting text into sqlite for search; raw artifacts remain the canonical record.

### 5) “Generic metadata” via tags/kv table

This keeps the core schema stable while enabling reuse:

- suite name/version
- git SHA
- host/user/OS/arch
- scenario parameters
- “expected failure” flags

### 6) Prepared statements over manual SQL quoting

Even though we have a `sqliteQuoteStringLiteral()` helper elsewhere, the scenario logger should primarily use **parameterized queries** to avoid injection/quoting bugs (especially if step names can contain weird characters).

### 7) Glazed patterns for structured CLI output (Cobra-compatible)

The `scenariolog` CLI should be a standard Cobra app, but for “report/query” commands it should adopt Glazed’s structured output patterns so we get multi-format output essentially for free.

Key patterns from `glaze help build-first-command`:

- Implement query/report commands as `cmds.GlazeCommand` and emit `types.Row` (instead of printing tables by hand).
- Parse flags via `parsedLayers.InitializeStruct(layers.DefaultSlug, &SettingsStruct{})` (avoid reading Cobra flags directly).
- Add the standard output flags by including `settings.NewGlazedParameterLayers()` (provides `--output`, `--fields`, `--sort-columns`, etc).
- Add debugging/config flags via `cli.NewCommandSettingsLayer()` (`--print-schema`, `--print-parsed-parameters`, etc).
- Bridge into Cobra via `cli.BuildCobraCommand(...)`.
- For commands that should be human-first but automation-friendly (like `exec`), use the **dual-mode** pattern (implement both `cmds.BareCommand` and `cmds.GlazeCommand`) and add a toggle flag such as `--with-glaze-output`.

Also: use compile-time interface assertions (fits our repo style):

- `var _ cmds.GlazeCommand = &MyCommand{}`
- (optionally) `var _ cmds.BareCommand = &MyCommand{}`

### 8) FTS5 for log search

- We prefer DB-backed search over shelling out to grep/rg because it’s fast, portable, and easy to integrate into reports.
- FTS indexing is generated from artifacts; it’s allowed to be best-effort and can be disabled if it causes unexpected bloat.
- The tool should **degrade gracefully** if sqlite lacks FTS5: core DB schema and artifacts remain valid; only the `log_lines_fts` table and search features are missing.

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
- [ ] Include `kv` + `artifacts` + `log_lines_fts` in the initial schema (or as schema v2+ migrations)
- [ ] Implement `Logger` with `StartRun` / `EndRun` and `StartStep` / `EndStep`
- [ ] Implement capture/execution helper using `exec.CommandContext` + `errgroup` for stdout/stderr copying
- [ ] Insert stdout/stderr captures into `artifacts` (and optionally keep `steps.stdout_path`/`stderr_path` columns in sync)
- [ ] Add FTS ingestion step that indexes stdout/stderr into `log_lines_fts` (either automatic or via `scenariolog index fts`)
- [ ] Add `cmd/scenariolog` (cobra + Glazed patterns) with:
  - [ ] `scenariolog init`
  - [ ] `scenariolog run start|end`
  - [ ] `scenariolog exec` (writes step rows + captures artifacts; dual-mode `--with-glaze-output` optional)
  - [ ] `scenariolog search` (FTS-backed, Glazed structured output)

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
- `ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/sources/local/glaze-help-build-first-command-2025-12-13.txt` (`glaze help build-first-command`)
- `test-scenarios/testing-doc-manager/README.md`
- `test-scenarios/testing-doc-manager/run-all.sh`
- `https://www.sqlite.org/pragma.html`
