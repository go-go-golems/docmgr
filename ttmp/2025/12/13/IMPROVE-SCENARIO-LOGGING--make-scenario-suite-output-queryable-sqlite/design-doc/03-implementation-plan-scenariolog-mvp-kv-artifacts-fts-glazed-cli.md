---
Title: 'Implementation plan: scenariolog MVP (KV + artifacts + FTS + Glazed CLI)'
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
    - Path: cmd/docmgr/main.go
      Note: Reference for how this repo structures Cobra commands and wiring
    - Path: internal/workspace/sqlite_schema.go
      Note: Existing SQLite pragmas / patterns / error wrapping conventions
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Primary integration target for step-level execution logging
    - Path: test-scenarios/testing-doc-manager/README.md
      Note: Scenario suite docs; will need an update once the logger exists
    - Path: ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/02-generic-sqlite-scenario-logger-go-tool.md
      Note: Primary design doc for scenariolog (schema + Glazed patterns + KV/artifacts/FTS)
ExternalSources:
    - local:glaze-help-build-first-command-2025-12-13.txt
Summary: "Concrete implementation plan for building scenariolog: schema/migrations (KV+artifacts+FTS), Glazed/Cobra CLI, and integration into the scenario harness."
LastUpdated: 2025-12-13T17:46:41.352631722-05:00
---

# Implementation plan: scenariolog MVP (KV + artifacts + FTS + Glazed CLI)

## Executive Summary

Implement a reusable “scenario flight recorder” tool (**`scenariolog`**) that:

- logs runs/steps/commands into **sqlite**
- records **KV tags** (run/step/command-scoped metadata)
- records **artifacts** (stdout/stderr + any files) with hashes/sizes
- indexes text artifacts into **FTS5** for fast “search warnings/errors” queries
- exposes a **Cobra CLI** that uses **Glazed patterns** for structured output (and optional dual-mode for human-first commands)

This doc turns the design in design-doc #2 into an actionable build plan and breaks work into small, verifiable steps.

## Problem Statement

The scenario suite produces long, interleaved output. We need a single, portable artifact that supports:

- fast “what failed?” queries
- structured timing breakdowns
- search over logs without shelling out to regex tools
- provenance and reproducibility metadata

We also want this to be **reusable** beyond docmgr scenarios, and easy to integrate into bash harnesses.

## Proposed Solution

### 1) Deliverables

- **Tool directory (self-contained)**: `scenariolog/`
  - **Go module**: `scenariolog/go.mod`
  - **Go library**: `scenariolog/internal/scenariolog`
  - open DB + pragmas
  - apply migrations (`PRAGMA user_version`)
  - create run/step/command rows
  - capture stdout/stderr to files and create `artifacts` rows
  - ingest text artifacts into `log_lines_fts`

- **CLI**: `scenariolog/cmd/scenariolog`
  - `init`, `run start|end`, `exec`
  - `index fts` (optional explicit indexing)
  - query/report commands with structured output:
    - `search`, `summary`, `failures`, `timings`

- **Scenario harness integration**
  - update `test-scenarios/testing-doc-manager/run-all.sh` to wrap steps using `scenariolog exec`
  - update scenario README with how to run + query the DB

### 2) CLI output modes (Glazed + dual-mode)

Glazed patterns from `glaze help build-first-command` enable “write once, output anywhere”:

- For report/query commands: implement `cmds.GlazeCommand` and emit `types.Row` so we automatically support `--output json|yaml|csv|table` and `--fields` / `--sort-columns`.
- For `exec` (human-first): consider dual-mode (implement both `cmds.BareCommand` + `cmds.GlazeCommand`) with a toggle like `--with-glaze-output`.

We will avoid reading Cobra flags directly; Glazed’s canonical pattern is:

- `parsedLayers.InitializeStruct(layers.DefaultSlug, &Settings{})`

### 3) Schema and migrations (v1)

Schema baseline is in design-doc #2. MVP requires:

- `scenario_runs`, `steps` (and optionally `commands`)
- `kv` (tags)
- `artifacts`
- `log_lines_fts` (FTS5 virtual table)

Migration strategy:

- `PRAGMA user_version` holds schema version.
- `scenariolog init` creates/updates schema to latest version.
- Migrations are idempotent and run inside a transaction when possible.

FTS5 strategy:

- The tool should **attempt** to create `log_lines_fts` as part of migrations.
- If sqlite lacks FTS5 (common error: `no such module: fts5`), the tool must **degrade gracefully**:
  - migrations still succeed
  - all non-FTS features work (runs/steps/kv/artifacts)
  - FTS-backed commands (`search`, `index fts`) should either:
    - return a clear error (“FTS5 not available”), or
    - behave as a no-op with an explicit warning (TBD; error is usually better)
- Recommended build for full features: `go -C scenariolog build -tags sqlite_fts5 ...` (CGO + go-sqlite3 feature build tag)

## Design Decisions

### 1) Keep artifacts as files + index text into sqlite

- Canonical record is the filesystem artifact.
- sqlite stores the metadata and provides FTS search via indexed lines.

### 2) Use prepared statements for writes

- Avoids quoting pitfalls.
- Makes it safe to store step names, paths, and arbitrary tags.

### 3) “Tags are KV, not JSON blobs”

- Simple to query and join.
- Still flexible enough for most metadata.

### 4) Use `github.com/pkg/errors` for wrapping

Matches repo conventions and keeps error chains readable.

### 5) Use `errgroup` for concurrent stream copying

One goroutine for stdout, one for stderr, plus optional tee; the pattern is clear and robust.

### 6) Interface assertions (compile-time)

For Glazed command structs:

- `var _ cmds.GlazeCommand = &MyCmd{}`
- (optional) `var _ cmds.BareCommand = &MyCmd{}`

## Alternatives Considered

Covered in design-doc #2. This doc focuses on execution steps.

## Implementation Plan

### 1.0 Phase 0 — Repo scaffolding

- [ ] Create `scenariolog/` self-contained tool directory:
  - [ ] `scenariolog/go.mod` + `scenariolog/go.sum`
  - [ ] `scenariolog/README.md` (how to build/run; how to integrate into bash harnesses)
- [ ] Create `scenariolog/internal/scenariolog/` package skeleton:
  - [ ] `db.go` (open DB, pragmas, close)
  - [ ] `migrate.go` (user_version, migrations)
  - [ ] `schema.sql` embedded or Go-DDL constants
  - [ ] `logger.go` (Run/Step/Command APIs)
  - [ ] `artifacts.go` (hash/size + insert)
  - [ ] `fts.go` (FTS ingestion helpers)
- [ ] Create CLI entrypoint at `scenariolog/cmd/scenariolog` (standalone first; can later be wrapped by `docmgr scenario ...`).

### 2.0 Phase 1 — SQLite schema + migrations (KV + artifacts + FTS)

- [ ] Implement `Open(ctx, path)`:
  - [ ] apply pragmas for file-backed DB (`foreign_keys=ON`, `journal_mode=WAL`, `synchronous=NORMAL`, busy timeout)
- [ ] Implement migrations:
  - [ ] `migrateToV1` creates: `scenario_runs`, `steps`, `kv`, `artifacts`, `log_lines_fts`
  - [ ] Ensure all tables have the necessary indexes

Verification:

- [ ] Create a tiny unit test that opens a temp DB, migrates, and asserts tables exist (`sqlite_master` queries).

### 3.0 Phase 2 — Execution wrapper + artifact capture

- [ ] Implement `Exec(ctx, ExecSpec) (ExecResult, error)`:
  - [ ] create stdout/stderr files under `log_dir`
  - [ ] start process with `exec.CommandContext`
  - [ ] copy stdout/stderr concurrently (use `errgroup`)
  - [ ] compute duration, exit code
- [ ] Insert artifact rows (`kind=stdout|stderr`, `path`, `sha256`, `size_bytes`, `is_text`)
- [ ] Insert KV tags for provenance:
  - [ ] suite name/version
  - [ ] hostname/user
  - [ ] (optional) git SHA + dirty flag (guarded / best-effort)

Verification:

- [ ] Unit test that runs `bash -lc 'echo out; echo err 1>&2; exit 3'` and verifies:
  - [ ] artifacts exist and contain expected text
  - [ ] exit code stored correctly
  - [ ] sha256/size are non-empty

### 4.0 Phase 3 — FTS ingestion + search

- [ ] Implement `IndexArtifactFTS(ctx, runID, artifactID, path)`:
  - [ ] scan file line-by-line
  - [ ] bulk insert into `log_lines_fts` (transaction + prepared statement)
  - [ ] bounds: max line length, max bytes (configurable)
- [ ] Decide when indexing happens:
  - [ ] default: index stdout/stderr automatically after `exec` completes
  - [ ] alternative: `scenariolog index fts --run-id ...` for explicit control (still useful)
- [ ] Implement `search` query:
  - [ ] `MATCH` query against FTS table
  - [ ] return artifact + line number + snippet

Verification:

- [ ] Unit test that indexes a known artifact and queries for a keyword; asserts 1+ hits with correct line numbers.

### 5.0 Phase 4 — CLI (Cobra + Glazed patterns)

- [ ] Create `scenariolog/cmd/scenariolog` with Cobra root.
- [ ] Add Glazed help system wiring (consistent with Glazed tutorial).
- [ ] Implement commands:
  - [ ] `init --db`
  - [ ] `run start --db --root-dir --suite` (returns run_id)
  - [ ] `run end --db --run-id --exit-code`
  - [ ] `exec --db --run-id --kind step --step-num --name --log-dir -- <cmd...>`
  - [ ] `search --db --query ...` (Glazed structured output)
  - [ ] `summary --db [--run-id]` (Glazed structured output)
  - [ ] `failures --db [--run-id]` (Glazed structured output)
  - [ ] `timings --db [--run-id]` (Glazed structured output)
- [ ] Adopt Glazed patterns:
  - [ ] settings structs with `glazed.parameter:"..."` tags
  - [ ] parse via `parsedLayers.InitializeStruct(...)`
  - [ ] include `settings.NewGlazedParameterLayers()` on report commands
  - [ ] include `cli.NewCommandSettingsLayer()` for `--print-schema` / `--print-parsed-parameters`

### 6.0 Phase 5 — Integrate into scenario harness

- [ ] Build `scenariolog` binary in `test-scenarios/testing-doc-manager` flow:
  - [ ] `go -C scenariolog build -o /tmp/scenariolog-local ./cmd/scenariolog`
  - [ ] pass the resulting path to the harness (e.g., `SCENARIOLOG_PATH=/tmp/scenariolog-local`)
- [ ] Modify `run-all.sh` to:
  - [ ] create a run (`scenariolog run start`)
  - [ ] wrap each script via `scenariolog exec --kind step ... -- bash ./NN-step.sh "$ROOT_DIR"`
  - [ ] finalize run (`scenariolog run end`)
- [ ] Update scenario README with:
  - [ ] how to run
  - [ ] how to query `search/failures/timings`
  - [ ] where artifacts live

### 7.0 Phase 6 — Hardening

- [ ] Signal handling: ensure CTRL-C results in a finalized step row + run row
- [ ] Busy/locking: set busy timeout; ensure we write via a single connection where appropriate
- [ ] Cleanup tooling (optional): prune old runs and remove orphan artifacts

### 8.0 Phase 7 — Docmgr integration (optional)

- [ ] Add `docmgr scenario` subcommands that shell out to scenariolog or embed library.
- [ ] Export scenario results as `docmgr` diagnostics JSON (if useful).

## Open Questions

1. **FTS size / retention**: should we index everything by default, or only stderr + “small enough” stdout?
2. **Driver choice**: stick with `mattn/go-sqlite3` (CGO) vs pure-Go later.
3. **Command granularity**: do we need `commands` table in MVP, or can we defer?
4. **CLI packaging**: standalone `scenariolog` binary vs `docmgr scenario` command group.
5. **Output mode**: do we want `exec` to be dual-mode by default?

## References

- Primary design: `ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/02-generic-sqlite-scenario-logger-go-tool.md`
- Scenario-specific baseline: `ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/01-scenario-suite-structured-logging-sqlite.md`
- Glazed tutorial output: `ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/sources/local/glaze-help-build-first-command-2025-12-13.txt`
