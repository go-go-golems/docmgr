---
Title: Diary
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/workspace/export_sqlite.go
      Note: Cobra wiring for workspace export-sqlite.
    - Path: internal/documents/walk.go
      Note: Current document walking contract (path, doc, body, readErr) → future DocHandle contract.
    - Path: internal/paths/resolver.go
      Note: |-
        Path normalization engine to be used by Workspace index + reverse lookup.
        Resolver anchors (repo/doc/config/docs-root/docs-parent) used for RelatedFiles existence checks
    - Path: internal/workspace/config.go
      Note: Existing root/config/vocab discovery helpers (basis for WorkspaceContext).
    - Path: internal/workspace/discovery.go
      Note: Existing ticket workspace discovery helpers (to be centralized in Workspace).
    - Path: internal/workspace/index_builder.go
      Note: |-
        Workspace.InitIndex + ingestion walker (Task 5).
        populate related_files norm_* columns (Task 6).
        Parse-error ingestion now infers ticket_id from ticket dir to make IncludeErrors usable under ScopeTicket.
    - Path: internal/workspace/index_builder_test.go
      Note: |-
        Index ingestion unit test (Task 5).
        assert related_files normalization keys (Task 6).
    - Path: internal/workspace/normalization.go
      Note: RelatedFiles normalization pipeline + persisted key strategy (Task 6).
    - Path: internal/workspace/query_docs.go
      Note: |-
        Nested hydration (topics/related_files) during rows iteration explains why MaxOpenConns(1) caused a hang.
        No-N+1 refactor: QueryDocs now does 1 base query + 2 batched hydration queries (topics/related_files) instead of nested per-row queries.
        Parse-error docs now keep ticket ID so doctor can group findings by ticket
    - Path: internal/workspace/query_docs_sql.go
      Note: |-
        Supports compiling parse_ok=0 query variant to surface skipped docs as diagnostics.
        Implements basename suffix fallback matching for QueryDocs RelatedFile filter (needed for scenario wonky path queries).
    - Path: internal/workspace/query_docs_test.go
      Note: |-
        Repro/guardrail tests for QueryDocs defaults + IncludeErrors behavior.
        Updated expectations for parse-error docs now carrying ticket id
    - Path: internal/workspace/skip_policy.go
      Note: Canonical ingest-time skip policy + path tagging helpers (Task 4, Spec §6).
    - Path: internal/workspace/skip_policy_test.go
      Note: Unit tests for skip policy + path tagging.
    - Path: internal/workspace/sqlite_export.go
      Note: Exports index to file; populates README table; uses VACUUM INTO.
    - Path: internal/workspace/sqlite_export_test.go
      Note: Smoke test for exported sqlite README table.
    - Path: internal/workspace/sqlite_schema.go
      Note: |-
        In-memory SQLite open + schema DDL (Task 3, Spec §9.1–§9.2).
        related_files columns expanded for canonical+fallback keys (Task 6).
        Fix deadlock+flakiness: allow multiple connections for nested hydration; unique shared in-memory DB name per Workspace to prevent cross-test leakage.
    - Path: internal/workspace/sqlite_schema_test.go
      Note: Unit smoke test for schema creation.
    - Path: internal/workspace/workspace.go
      Note: New Workspace/WorkspaceContext/DiscoverWorkspace skeleton (Task 2, Spec §5.1).
    - Path: pkg/commands/doctor.go
      Note: Ported doctor to Workspace.QueryDocs and switched RelatedFiles existence checks to doc-anchored paths.Resolver
    - Path: pkg/commands/list_docs.go
      Note: Ported list docs to Workspace.QueryDocs (Task 9).
    - Path: pkg/commands/relate.go
      Note: Ported doc relate to Workspace.QueryDocs-based doc lookup and Workspace-derived paths.Resolver normalization
    - Path: pkg/commands/search.go
      Note: Ported doc search to Workspace.QueryDocs (Task 10).
    - Path: pkg/commands/workspace_export_sqlite.go
      Note: Implements  as a classic Run() command.
    - Path: pkg/diagnostics/docmgrctx/query_docs.go
      Note: Defines QueryDocs-specific taxonomy types for diagnostics payload (parse skip + normalization fallback).
    - Path: pkg/doc/embedded_docs.go
      Note: Exports embedded pkg/doc/*.md (go:embed) for README table population.
    - Path: test-scenarios/testing-doc-manager/19-export-sqlite.sh
      Note: Scenario smoke test for export-sqlite + README table.
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: |-
        Used to validate both system and local refactor binaries.
        Runs the scenario suite; now includes export-sqlite smoke.
        Scenario suite run to validate list docs + refactor wiring.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/02-testing-strategy-integration-first.md
      Note: Decision record for when/how we add integration tests during the refactor.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/07-code-review-walkthrough-diary-driven.md
      Note: |-
        Diary-driven code review walkthrough that uses the diary as its narrative structure.
        Diary-driven code review walkthrough document (Step 15 reviewer instructions)
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/08-cleanup-inspectors-brief-task-18.md
      Note: Brief for cleanup inspectors to inventory duplicated walkers/helpers (Task 18)
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Spec driving this refactor.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/tasks.md
      Note: Task breakdown for implementation.
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T17:35:05.756386407-05:00
---





















# Diary

## Goal

Capture the step-by-step implementation of the **Workspace + in-memory SQLite index refactor** (what changed, why, and what we learned), with enough detail that a new contributor can pick up mid-stream.

## Context

- Spec: `design/01-workspace-sqlite-repository-api-design-spec.md`
- Tasks: `tasks.md`
- Key baseline code:
  - `internal/workspace/config.go` (root/config/vocab discovery)
  - `internal/workspace/discovery.go` (ticket discovery)
  - `internal/paths/resolver.go` (path normalization)
  - `internal/documents/walk.go` (walk contract we’ll reuse as `DocHandle`)

## Quick Reference

### Commands used frequently

```bash
# Ticket docs (DIARY/CHANGELOG/RELATE/TASKS): always use the system docmgr
docmgr task list --ticket REFACTOR-TICKET-REPOSITORY-HANDLING
docmgr task check --ticket REFACTOR-TICKET-REPOSITORY-HANDLING --id 1
docmgr changelog update --ticket REFACTOR-TICKET-REPOSITORY-HANDLING --entry "..."
docmgr doc relate --doc ttmp/.../reference/15-diary.md --file-note "/abs/path:note"

# Refactor testing: build a local docmgr binary and point scenarios at it
go build -o /tmp/docmgr-refactor-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-refactor-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-local
```

## Usage Examples

N/A — this document is the usage example; it’s written as we implement.

## Related

- `design/01-workspace-sqlite-repository-api-design-spec.md`
- `reference/13-design-log-repository-api.md`

## Step 1: Kickoff — diary + baseline scan

We started by putting “meta tooling” in place so the refactor stays navigable. The goal of this step wasn’t to write production code yet, but to make sure every subsequent change is tied back to documentation and is discoverable from either direction (code → docs, docs → code).

### What I did
- Created the ticket diary doc (`reference/15-diary.md`).
- Related baseline files (config/discovery/resolver/walk + spec + tasks) to the diary.
- Added an initial changelog entry to mark implementation start.

### Why
- We want the implementation to be “explainable” and searchable from code ↔ docs while we refactor.

### What worked
- `docmgr doc relate` updated relationships cleanly (no overwrite issues).

### What didn’t work
- Nothing yet.

### What I learned
- `internal/workspace` already owns config/root resolution and ticket discovery helpers, so it’s the right home for the new `Workspace` entry point.

### Technical details
- Diary doc created via:
  - `go run ./cmd/docmgr doc add --ticket REFACTOR-TICKET-REPOSITORY-HANDLING --doc-type reference --title "Diary"`
- Initial relate + changelog were recorded immediately after.

## Step 2: Implement `internal/workspace.Workspace` skeleton (Task 2, Spec §5.1)

This step introduced the new “front door” object (`Workspace`) without changing any user-facing behavior yet. The purpose is to establish a single canonical place where discovery, normalization, and later indexing will live, so that commands stop reimplementing slightly different semantics.

### What I did
- Added `internal/workspace/workspace.go`:
  - `WorkspaceContext` (Root/ConfigDir/RepoRoot + best-effort config)
  - `DiscoverWorkspace(ctx, opts)` and `NewWorkspaceFromContext(ctx)`
  - `paths.Resolver` wiring (anchors: docs root, config dir, repo root)
- Ran `go test ./...` to ensure everything still compiles.
- Commit: `5f1681b1f3ac5ed7ffa36fb4b2357b54ebaf6695` (":sparkles: First commit for the refactor towards a unified workspace and search functionality")

### Why
- This gives every CLI command a single canonical entry point to obtain:
  - consistent root/config/repo discovery, and
  - a consistent path normalizer (required for reverse lookup correctness later).

### What worked
- The repo compiles cleanly; existing tests passed.

### What didn’t work
- `LoadWorkspaceConfig()` currently returns an error for malformed config even though it prints a warning and claims to continue. In `DiscoverWorkspace` we treat that as non-fatal and proceed with `Config=nil`.

### What I learned
- Keeping discovery in `internal/workspace` avoids new packages and reuses the existing resolution chain in `config.go`.

### Technical details
- New file: `internal/workspace/workspace.go`
- Discovery heuristics implemented:
  - `Root`: `ResolveRoot(opts.RootOverride|\"ttmp\")`
  - `ConfigDir`: `.ttmp.yaml` dir if present; else `filepath.Dir(root)` as heuristic
  - `RepoRoot`: `FindRepositoryRoot()`

## Step 3: Decide testing approach (integration-first) and document it

Before we started wiring any commands to the new backend, we made an explicit call on the testing strategy. The key idea: we’re refactoring *behavioral plumbing* (discovery, filtering, reverse lookup), so scenario tests that drive the real CLI are the most cost-effective guardrail.

### What I did
- Audited `test-scenarios/testing-doc-manager/` and identified the scripts that already cover the riskiest parts of this refactor:
  - reverse lookup + wonky paths
  - path normalization
  - diagnostics taxonomy smoke
- Wrote an analysis doc with an integration-first testing plan:
  - `analysis/02-testing-strategy-integration-first.md`

### Why
- This refactor changes *how* we compute results, but the critical requirement is that the **CLI behavior stays consistent** (or changes only where explicitly intended).
- Scenario-based tests catch wiring/defaults/flag interactions that unit tests tend to miss.

### What worked
- The existing scenario already exercises:
  - `doc search --file/--dir` reverse lookup,
  - normalization across doc-relative/ttmp-relative/absolute paths,
  - and taxonomy/diagnostics emission.

### What didn’t work
- Nothing yet.

### What I learned
- We can treat `test-scenarios/testing-doc-manager` as the baseline regression suite and extend it as soon as we port the first command to `Workspace.QueryDocs`.

### Technical details
- Key scenario scripts:
  - `test-scenarios/testing-doc-manager/05-search-scenarios.sh`
  - `test-scenarios/testing-doc-manager/14-path-normalization.sh`
  - `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh`

## Step 4: Run baseline integration tests (system docmgr)

With the testing plan written down, we immediately captured a “known good” run using the installed `docmgr`. This gives us a stable baseline to compare against when the refactor starts changing internals (SQLite ingestion/query compilation) and when we port commands.

### What I did
- Ran the full integration scenario suite with the **system** `docmgr` to confirm the baseline currently passes.

### Why
- Before changing command wiring to the new `Workspace` backend, we need a “known good” reference run so we can detect regressions and unintended behavior changes.

### What worked
- The scenario completed successfully (`[ok] Scenario completed ...`).

### What didn’t work
- Nothing (this was a clean pass).

### What I learned
- The scenario already exercises the core behaviors we’re about to refactor:
  - `doc search --file/--dir` reverse lookup
  - wonky path normalization cases
  - diagnostics smoke / taxonomy wiring

### Technical details
- Command:
  - `DOCMGR_PATH=$(which docmgr) bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-baseline-2025-12-12`
- Result root:
  - `/tmp/docmgr-scenario-baseline-2025-12-12/acme-chat-app`

## Step 5: Run integration tests against the locally built (refactor) docmgr binary

Next, we repeated the same scenario run, but swapped the tested executable to a locally built binary from this repo. This is the simplest apples-to-apples integration validation loop: same scripts, same mock repo, only the binary changes.

### What I did
- Built a local `docmgr` binary from this repo and ran the same scenario suite with `DOCMGR_PATH` pointing at the local binary.

### Why
- This is the cleanest “integration test the refactor” loop:
  - same scenario inputs,
  - same scripts,
  - only the tested executable changes.

### What worked
- The scenario completed successfully against the local binary.

### What didn’t work
- Nothing (this was a clean pass).

### What I learned
- At this stage (only adding the `Workspace` skeleton), behavior is unchanged, and the end-to-end harness is already a good safety net for upcoming QueryDocs/SQLite changes.

### Technical details
- Build + run:
  - `go build -o /tmp/docmgr-refactor-local-2025-12-12 ./cmd/docmgr`
  - `DOCMGR_PATH=/tmp/docmgr-refactor-local-2025-12-12 bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-local-2025-12-12`
- Result root:
  - `/tmp/docmgr-scenario-local-2025-12-12/acme-chat-app`

## Step 6: Add in-memory SQLite schema for Workspace index (Task 3)

This step laid down the contract for what the in-memory index will store. The schema is intentionally minimal and aligned with the spec, so ingestion/query compilation can be built incrementally without “schema churn” each time we add a filter.

### What I did
- Implemented in-memory SQLite bootstrap + schema creation under `internal/workspace`:
  - `docs`
  - `doc_topics`
  - `related_files`
- Added a small unit test to ensure schema creation works.
- Commit: `5b72f4ec08faad89c1219030daf4c96ca5456db0` (":sparkles: Add initial sqlite_schema")

### Why
- Querying and reverse lookup become SQL joins, so we need a stable minimal schema before implementing ingestion and `QueryDocs`.

### What worked
- `go test ./...` passes with the new schema + test.

### What didn’t work
- Nothing yet.

### What I learned
- There was no existing SQLite wrapper in the repo, so creating a focused schema module in `internal/workspace` keeps the implementation localized and ready for ingestion.

### Technical details
- Files:
  - `internal/workspace/sqlite_schema.go`
  - `internal/workspace/sqlite_schema_test.go`

## Step 7: Canonical ingest-time skip rules + tagging (Task 4, Spec §6)

This step is about making the definition of “what is a doc” consistent. Historically, different commands implemented different skip logic (string contains, underscore-only, ignore globs), which meant “the set of docs” varied depending on which command you ran. For the SQLite index to be trustworthy, ingestion must apply a single canonical skip policy and tag docs so queries can hide/show categories consistently.

### What I did
- Implemented canonical ingest-time directory skip policy:
  - always skip `.meta/`
  - always skip underscore dirs (`_*/`) like `_templates/` and `_guidelines/`
- Implemented path-derived tagging for docs (to persist into SQLite `docs`):
  - `is_index`
  - `is_archived_path` / `is_scripts_path` / `is_sources_path`
  - `is_control_doc` (README/tasks/changelog when co-located with a ticket `index.md`)
- Added unit tests covering the tricky cases (control-doc detection + segment-based path tagging).

### Why
- These tags are the “bridge” between ingest-time policies and query-time options (`IncludeArchivedPath`, `IncludeScriptsPath`, etc.).
- A single canonical policy prevents subtle inconsistencies across commands and makes reverse lookup predictable.

### What worked
- `go test ./...` still passes with the new skip/tagging helpers + tests.

### What didn’t work
- Nothing yet.

### What I learned
- Requiring a sibling `index.md` for `is_control_doc` prevents accidental tagging of `sources/README.md` (or any nested README) as a ticket control doc.

### Technical details
- Files:
  - `internal/workspace/skip_policy.go`
  - `internal/workspace/skip_policy_test.go`
- Commit: (pending — waiting for the task checkpoint commit hash)

## Step 8: Implement ingestion walker + index build (Task 5, Spec §6 / §7.1 / §9.2)

With the schema and ingest policy in place, this step makes the index “real”: we walk the docs root once, parse each markdown file’s frontmatter, and store a normalized, queryable representation in the in-memory SQLite database. The main outcome is that later features (filters, reverse lookup, diagnostics) can be expressed as SQL queries instead of nested loops and ad-hoc maps.

### What I did
- Added an index builder to `internal/workspace`:
  - `Workspace.InitIndex(ctx, opts)` opens a new in-memory SQLite DB, applies pragmas, creates schema, and ingests docs.
  - Ingestion uses `documents.WalkDocuments` with the canonical ingest skip predicate.
- During ingestion, we store:
  - `docs` row for every `.md` encountered (absolute path, frontmatter fields, `parse_ok/parse_err`, path tags, optional body)
  - `doc_topics` rows (lowercased topic key + original case)
  - `related_files` rows for each frontmatter RelatedFiles entry, with normalized keys derived via `paths.Resolver`
- Added a unit test that builds a small temporary docs root and asserts:
  - `.meta/` and `_guidelines/` docs are skipped
  - broken frontmatter produces `parse_ok=0` + non-empty `parse_err`
  - ticket-root `tasks.md` is tagged as `is_control_doc=1`
  - topics and related_files are ingested

### Why
- This is the core prerequisite for `QueryDocs`: once data is in the DB in a stable shape, query logic becomes “compile to SQL + scan rows”.
- Indexing broken docs (but marking `parse_ok=0`) lets us surface diagnostics without silently dropping data.

### What worked
- `go test ./...` passes with the new ingestion builder + test.

### What didn’t work
- Initial version of the new test referenced helpers that didn’t exist; fixed by using `os.MkdirAll` + `os.WriteFile`.

### What I learned
- Using `paths.Resolver` with `DocPath=absPath` during ingestion is crucial: it makes doc-relative `RelatedFiles` normalize correctly, which is the entire reason reverse lookup becomes reliable with SQLite.

### Technical details
- Files:
  - `internal/workspace/index_builder.go`
  - `internal/workspace/index_builder_test.go`
  - `internal/workspace/workspace.go` (now holds `db *sql.DB` and `DB()` accessor)
- Commit: (pending — waiting for the task checkpoint commit hash)

## Step 9: Path normalization pipeline + persisted keys (Task 6, Spec §7.3 / §12.1 / §9.2)

This step tightens the contract for reverse lookup and “related files” matching. In real repos, people will write `RelatedFiles` entries in a bunch of different ways (repo-relative, absolute, doc-relative, config-relative, sometimes with `../` segments). If we only store one representation, reverse lookup becomes flaky and we regress into command-specific heuristics.

So the goal here is: **persist enough normalized keys at ingest-time** so that later query logic can do reliable SQL matching without needing to reconstruct “how the path was meant” at query time.

### What I did
- Expanded the `related_files` schema to store **multiple normalization keys**:
  - `norm_canonical` (best-effort “best” key)
  - `norm_repo_rel`, `norm_docs_rel`, `norm_doc_rel`, `norm_abs`
  - `norm_clean` (cleaned path derived from the raw string; preserves leading `..` but normalizes separators)
  - plus `anchor` and `raw_path` for debugging and UX
- Centralized ingestion-time normalization in `internal/workspace/normalization.go` so the index builder doesn’t hand-roll the mapping.
- Strengthened the ingestion test to assert that for a repo-relative related file:
  - canonical and repo-relative keys are both `backend/main.go`
  - docs-relative is empty (as expected, because that file is outside docs root)
  - doc-relative exists and ends with `backend/main.go` (we avoid asserting the exact number of `../`)
  - anchor is `repo`

### Fallback matching strategy (documented)
The planned `QueryDocs` behavior for `RelatedFile` / `RelatedDir` matching is:
1. Normalize the user-provided query path using `paths.Resolver`.
2. Match by **equality against any persisted representation**, typically in this order:
   - `norm_canonical`, `norm_repo_rel`, `norm_docs_rel`, `norm_doc_rel`, `norm_abs`, `norm_clean`
3. If we later decide we want fuzzier semantics (suffix / substring), we can add explicit SQL `LIKE` fallbacks or precomputed suffix columns — but the **MVP is strict equality across multiple representations**, which already handles most of the “same file written differently” cases.

### Technical details
- Files:
  - `internal/workspace/sqlite_schema.go`
  - `internal/workspace/index_builder.go`
  - `internal/workspace/index_builder_test.go`
  - `internal/workspace/normalization.go`
- Commit: (pending — waiting for the task checkpoint commit hash)

## Step 9: Add `workspace export-sqlite` + self-describing README table + scenario smoke test

This step turns the in-memory index into a shareable artifact. The goal is to make debugging easier (“send me the DB you’re seeing”) and to make the exported file self-explanatory: someone can open it with any SQLite browser, inspect the schema, and immediately find embedded documentation inside a dedicated table.

### What I did
- Implemented an export path that writes the in-memory workspace index to a persistent SQLite file using SQLite’s `VACUUM INTO`.
- Added a `README` table to the exported DB and populated it with:
  - `__about__.md` (a short explainer for the exported DB), and
  - all `pkg/doc/*.md` embedded docs (via `go:embed`) so the DB is self-describing.
- Added a new CLI verb: `docmgr workspace export-sqlite` (bare `Run()` only; no structured output).
- Added a scenario smoke test that:
  - runs `workspace export-sqlite`,
  - then checks `README` and a couple known rows (`__about__.md`, `docmgr-how-to-use.md`) via Python `sqlite3`.

### Why
- Exporting as a file lets us:
  - diff DBs between versions,
  - share a reproducible snapshot for debugging,
  - and inspect the schema/data with common tooling.
- The `README` table makes the DB useful even out of context: a new person can open it and learn how docmgr works without having the repo checked out.

### What worked
- The command is registered under `docmgr workspace export-sqlite` and works against the scenario repo.
- The scenario suite passes with the new smoke test included.

### What didn’t work
- A raw-string markdown snippet inside `sqlite_export.go` initially caused a Go syntax error (due to backticks). Fixed by switching to a normal concatenated string.

### What I learned
- Keeping the export command “classic” (BareCommand only) is simpler for debugging: it’s a side-effecting operation where a single success message is enough, and JSON output would add little value.

### Technical details
- Code:
  - `pkg/doc/embedded_docs.go` (exports embedded docs from `go:embed`)
  - `internal/workspace/sqlite_export.go` + `_test.go` (README table + VACUUM INTO)
  - `pkg/commands/workspace_export_sqlite.go` (command implementation; `Run()` only)
  - `cmd/docmgr/cmds/workspace/export_sqlite.go` (cobra wiring)
- Scenario:
  - `test-scenarios/testing-doc-manager/19-export-sqlite.sh`
  - `test-scenarios/testing-doc-manager/run-all.sh` (wired in)
- Commit: (pending — waiting for the task checkpoint commit hash)

## Step 10: Debug QueryDocs test hang + stabilize in-memory SQLite behavior

This step was about turning a “something fishy” test hang into a concrete, explainable failure mode, and then adjusting the SQLite setup so `QueryDocs` can safely do what it currently does (best-effort hydration) without deadlocking. The main impact is that unit tests are now deterministic again, and the query/indexing behavior around broken docs is more consistent with the spec’s intent (broken docs should be indexable and discoverable for repair workflows when explicitly requested).

The debugging also highlighted a useful implementation constraint: if we keep “nested hydration queries” inside the `rows.Next()` loop, we must not force SQLite to a single connection. In the longer term, we may still want to move to batched hydration (or a single joined query) to reduce round-trips, but correctness and debuggability come first.

### What I did
- Reproduced the hang quickly by running the specific `QueryDocs` unit test with a short timeout.
- Identified the deadlock mechanism: a main `SELECT ... FROM docs` result set was being iterated while `QueryDocs` attempted to prepare/execute additional statements on the same `*sql.DB`.
- Updated SQLite connection policy to allow multiple connections (so nested queries can progress).
- Fixed a separate “flaky semantics” failure caused by using a constant shared in-memory DB name across tests in the same process.
- Fixed a `ScopeTicket + IncludeErrors=true` mismatch by inferring `ticket_id` for parse-error docs from the ticket directory layout at ingest time.
- Re-ran `go test ./...` to confirm the suite is stable.

### Why
- A hanging test is worse than a failing test: it blocks progress and obscures root cause.
- The SQLite deadlock was a structural interaction between:
  - `database/sql` connection pooling,
  - SQLite’s connection-level serialization of work, and
  - our current `QueryDocs` design (nested hydration while iterating the main cursor).
- The “shared in-memory DB name” issue was subtle: using `cache=shared` with a fixed DSN can leak schema/data between independent `Workspace` instances in the same test process.
- For broken docs, we need ticket scoping to still work when the doc lives under a ticket directory but frontmatter parsing fails, otherwise `IncludeErrors=true` is hard to use in practice.

### What worked
- The timeout repro immediately surfaced the blocking point in a stack trace (stuck in `database/sql.(*DB).PrepareContext` from inside `QueryDocs`).
- Allowing multiple connections removed the hang and made the original test complete quickly.
- Switching to a unique in-memory DB name per `Workspace` instance eliminated cross-test contamination.
- Inferring `ticket_id` for parse-error docs brought the behavior back in line with the unit test expectations and the spec’s “repair-friendly” intent.

### What didn’t work
- After fixing the deadlock, a full `go test ./...` run started failing with “expected 1 doc with defaults, got 2”. This turned out to be shared in-memory DB state leaking between tests, not a `QueryDocs` semantic bug.

### What I learned
- `database/sql` + SQLite can deadlock surprisingly easily if you:
  - iterate a `rows` cursor, and
  - run another query on the same `*sql.DB`, and
  - cap `MaxOpenConns` at 1.
- Named shared in-memory SQLite (`file:<name>?mode=memory&cache=shared`) is convenient, but the `<name>` must be unique per workspace/test instance unless you *want* cross-connection sharing.
- “Broken doc” handling is not just about `parse_ok=0`; you often still need enough metadata (at least inferred ticket) to make the broken doc discoverable via scoped queries.

### Technical details
- **Nested hydration point**: `internal/workspace/query_docs.go` prepares `topicStmt`/`rfStmt` and calls `fetchTopics`/`fetchRelatedFiles` inside the `for rows.Next()` loop.
- **Deadlock trigger**: `internal/workspace/sqlite_schema.go` previously used `db.SetMaxOpenConns(1)` while using nested queries.
- **Fix 1 (deadlock)**: allow multiple connections (`SetMaxOpenConns(4)` / `SetMaxIdleConns(4)`).
- **Fix 2 (test flake)**: use a unique named in-memory DB per `Workspace` instance (still with `cache=shared` so multiple connections in that workspace see the same DB).
- **Fix 3 (IncludeErrors under ticket scope)**: in `internal/workspace/index_builder.go`, if parsing fails we now infer ticket ID from the `ttmp/YYYY/MM/DD/TICKET--slug/...` layout (best-effort) so `ScopeTicket` filtering can include broken docs when explicitly requested.
- **Commands run**:

```bash
go test ./internal/workspace -run TestWorkspaceQueryDocs_BasicFiltersAndReverseLookup -count=1 -timeout 5s
go test ./... -count=1
```

### What I’d do differently next time
- Add a small comment near the nested hydration logic in `QueryDocs` explaining the connection-pool requirement (or switch hydration to a single joined query early).
- In unit tests that exercise in-memory DB behavior, prefer verifying isolation assumptions explicitly (e.g., “workspace A doesn’t see workspace B docs”) so this kind of leak is caught earlier.

## Step 11: Remove nested queries / N+1 from `Workspace.QueryDocs`

This step refactored `QueryDocs` to eliminate the “query while iterating rows” pattern and the resulting N+1 behavior. The main outcome is that `QueryDocs` now executes a fixed number of queries regardless of result size (one base docs query, then at most one topics query and one related-files query), and hydration happens in-memory. This makes the implementation easier to reason about, avoids subtle connection-pool interactions, and sets us up for later diagnostics work without worrying about deadlocks or per-row overhead.

We kept the external behavior stable: ordering is still driven by the compiled SQL, parse-error docs still return as `DocHandle{Doc:nil, ReadErr:...}` when included, and topics/related files are still hydrated for parse-ok docs—just via batched lookups.

### What I did
- Reworked `internal/workspace/query_docs.go` so the main docs query is fully scanned first (capturing `doc_id` for parse-ok docs).
- Added two batch hydration helpers:
  - `fetchTopicsByDocIDs(ctx, db, docIDs)` using `WHERE doc_id IN (...)`
  - `fetchRelatedFilesByDocIDs(ctx, db, docIDs)` using `WHERE doc_id IN (...)`
- Hydrated `doc.Topics` and `doc.RelatedFiles` from `map[doc_id]...` after scanning.
- Removed the old per-doc hydration helpers (`fetchTopics` / `fetchRelatedFiles`) that enabled the nested-query pattern.
- Ran `go test ./...` to confirm behavior stayed green.

### Why
- Avoid N+1 (performance + simplicity).
- Avoid nested cursor/query interactions that can deadlock if connection limits are changed (or if future code introduces transactions/locks).
- Make it easier to add diagnostics later (Task 8) without sprinkling extra queries inside the hot loop.

### What worked
- Unit tests for `QueryDocs` still pass without changes, which suggests the refactor preserved semantics.
- The implementation became more predictable: 1 + 2 queries instead of 1 + (2 * number_of_docs).

### What didn’t work
- Nothing notable in this refactor (it was a mechanical change with good test coverage).

### What I learned
- Even on an in-memory DB, N+1 is easy to introduce accidentally when doing “hydration”. Batch hydration is often the cleanest compromise before moving to a single big join query.

### Technical details
- Files:
  - `internal/workspace/query_docs.go`
- Test command:

```bash
go test ./... -count=1
```

### What I’d do differently next time
- Consider making hydration strategy explicit (e.g., “no hydration”, “batch hydration”, “single join”) if we find commands with different needs.

## Step 12: Implement QueryDocs diagnostics (parse skips + normalization fallback)

This step made `QueryDocs` explain itself. The spec’s core UX requirement is that “broken” docs (invalid frontmatter) should not silently disappear: they should be excluded from default results, but still show up as structured diagnostics so users can repair their workspace. In practice, this means `QueryDocs` now has a dual output channel: **Docs** (the normal results) and **Diagnostics** (a list of `core.Taxonomy` entries explaining what was skipped or why matching had to fall back).

The second part of this step improves reverse lookup debuggability: when path normalization can’t derive strong keys (canonical/repo-relative/absolute) and must rely on weaker cleaned/raw matching, `QueryDocs` emits a warning diagnostic. Matching still proceeds (fallback strategy), but results become explainable instead of “mysterious”.

### What I did
- Added `pkg/diagnostics/docmgrctx/query_docs.go` with QueryDocs-specific taxonomy constructors:
  - parse-skip (`query_skipped_due_to_parse`)
  - normalization fallback (`query_normalization_fallback`)
- Refactored SQL compilation so we can explicitly compile “parse_ok=1”, “parse_ok=0”, or “no parse filter” variants:
  - `compileDocQueryWithParseFilter(..., parseOKFilter *int)`
- Updated `Workspace.QueryDocs` to:
  - emit normalization fallback diagnostics when reverse lookup inputs only yield weak keys
  - when `IncludeDiagnostics=true` and `IncludeErrors=false`, run a second compiled query with `parse_ok=0` to collect skipped docs and return them as taxonomy entries (without polluting normal results)
- Checked off Tasks 7 and 8 in the ticket task list once the behavior was implemented and tests were green.
- Ran `go test ./...` to confirm stability.

### Why
- Default behavior must hide invalid-frontmatter docs (clean output), but we still need a repair/debug path that explains what got skipped (Spec §10.6 / Decision D1=B).
- Reverse lookup failures are often “normalization ambiguity” problems; diagnostics make those cases debuggable instead of guesswork (Decision D3=B).

### What worked
- The implementation is low-risk: the main query behavior stays unchanged; diagnostics are additive.
- Tests stayed green.

### What didn’t work
- Nothing significant yet; this is the first “minimum viable” diagnostics pass and will likely evolve as we port commands.

### What I learned
- It’s much easier to build diagnostics when the query compiler can be parameterized (here: parse state) instead of doing brittle SQL-string rewrites.

### Technical details
- Files:
  - `internal/workspace/query_docs.go`
  - `internal/workspace/query_docs_sql.go`
  - `pkg/diagnostics/docmgrctx/query_docs.go`
- Test command:

```bash
go test ./... -count=1
```

### What I’d do differently next time
- Add a focused unit test that asserts the Diagnostics payload shape (stage/symptom/severity/path) for a parse-error doc when `IncludeDiagnostics=true`.

## Step 13: Port `list docs` to `Workspace.QueryDocs`

This step moved the `list docs` command off of ad-hoc filesystem walking and onto the new Workspace+SQLite index. The key goal is consistency: “what counts as a doc” and “how filters behave” should be defined in one place (`Workspace.InitIndex` + `Workspace.QueryDocs`), not re-implemented in every command. This is also the first real proof that the Workspace API is usable from command code without leaking indexing or normalization details everywhere.

We kept the user-facing output shape intact for both human mode and glaze mode, and we preserved the key behavioral quirk that `list docs` skips any `index.md` file (tickets are listed via `list tickets`).

### What I did
- Rewrote `pkg/commands/list_docs.go` to:
  - call `workspace.DiscoverWorkspace(...)`
  - build the in-memory index via `ws.InitIndex(...)`
  - query via `ws.QueryDocs(...)` using `Ticket/Status/DocType/TopicsAny` filters
- Implemented output mapping from `DocHandle` back into:
  - glaze rows (`ticket,doc_type,title,status,topics,path,last_updated`)
  - human grouped Markdown output (per-ticket grouping + sorting)
- Preserved `index.md` skipping as a post-filter on the query results.
- In glaze mode, rendered `QueryDocs` diagnostics (to keep “skipped because parse” visibility).
- Checked off Task 9 in the ticket task list.

### Why
- Remove duplicated walkers and filter semantics, and make command behavior consistent across the tool.
- Leverage the new diagnostics channel so glaze users still see what got skipped and why (instead of silent drops).

### What worked
- The command no longer does its own recursive walk or per-file parsing; it is now a thin translation layer from flags → `DocQuery` → output formatting.
- The repo continues to build and tests stay green.

### What didn’t work
- Nothing notable in this port; it was largely mechanical once QueryDocs stabilized.

### What I learned
- The “skip index.md” requirement is easiest to preserve as a post-filter until we add an explicit `IsIndex` filter to `DocFilters` (or a dedicated option).

### Technical details
- Files:
  - `pkg/commands/list_docs.go`
- Test command:

```bash
go test ./... -count=1
```

## Step 14: Port `doc search` to `Workspace.QueryDocs` + preserve wonky path UX

This step ports the `doc search` command to the Workspace+SQLite backend, which is the big “real world” consumer of the new QueryDocs API. The goal is to centralize *all* metadata filtering and reverse lookup behind QueryDocs, while keeping content search as a post-filter (FTS remains deferred). This makes `doc search` both faster (no full filesystem walk for reverse lookup) and more consistent with other commands (same skip policy, same normalization strategy, same diagnostics model when enabled).

The most important compatibility requirement here was the “wonky path” behavior: users often search by odd path forms (deep `../../..`), absolute paths, or just a basename like `register.go`. The index-backed reverse lookup already handles many normalized forms, but basename-only queries required an explicit suffix fallback to preserve the existing scenario expectations.

### What I did
- Rewrote `pkg/commands/search.go` (both glaze and human modes) to:
  - discover workspace via `workspace.DiscoverWorkspace`
  - build the in-memory index with `IncludeBody=true` (so we can generate snippets)
  - call `ws.QueryDocs` for metadata filters + reverse lookup (`--file`, `--dir`)
  - keep content search, external-source filtering, and date filtering as post-filters
- Added basename/suffix fallback matching for `RelatedFile` queries in the SQL compiler so `--file register.go` still matches paths like `backend/chat/api/register.go`.
- Ran `go test ./...` and the scenario suite to confirm behavior:
  - `test-scenarios/testing-doc-manager/05-search-scenarios.sh` (wonky path regression included)

### Why
- `doc search` was the highest-risk port because it combined: content search, metadata filters, reverse lookup, and normalization edge cases.
- Moving reverse lookup into QueryDocs is the main architectural promise of the refactor (joins, not nested loops).

### What worked
- Scenario suite passes against the local binary, including the “wonky path regression” cases.
- `doc search --file ...` and `--dir ...` now rely on the same normalization/indexing rules as the rest of the tool.

### What didn’t work
- Nothing notable during the port; the main complexity was preserving basename-only matching without making reverse lookup too fuzzy.

### What I learned
- Basename-only reverse lookup is a pragmatic UX feature. Implementing it explicitly as a suffix `LIKE '%/name.go'` fallback keeps the behavior explainable (and can be paired with diagnostics) while still leveraging the index.

### Technical details
- Files:
  - `pkg/commands/search.go`
  - `internal/workspace/query_docs_sql.go`
- Commands:

```bash
go test ./... -count=1
go build -o /tmp/docmgr-refactor-local-2025-12-13 ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-refactor-local-2025-12-13 bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-local-2025-12-13
```

## Step 14: Verify list-docs port via scenario suite + direct command runs

This step was a “trust but verify” pass after wiring `list docs` to `Workspace.QueryDocs`. The goal was to validate the refactor in the same way users experience it: run the full scenario suite against a locally built binary, then exercise `docmgr list docs` directly on the generated workspace in both human output and glaze output modes. The important outcome is confidence that the port behaves correctly in a real(istic) repo layout, not just in unit tests.

One small operational gotcha came up: piping long command output into `head` can produce `SIGPIPE` (exit code 141) even when the underlying command succeeds. That’s not a test failure; it’s just the shell pipeline closing early. The scenario run itself still reported success.

### What I did
- Built a local refactor binary.
- Ran the full integration scenario suite with `DOCMGR_PATH` pointing at that binary.
- Ran `docmgr list docs` against the scenario workspace’s `ttmp/` root:
  - human output (grouped markdown)
  - glaze JSON output (paths only) for a quick “scriptability” check

### Why
- Scenario tests catch wiring/defaults/flag interactions that unit tests miss.
- `list docs` is a high-touch command; validating both output modes is important for regressions.

### What worked
- Scenario suite completed successfully (`[ok] Scenario completed ...`).
- `list docs` produced sensible grouped output and glaze JSON rows on the scenario workspace.

### What didn’t work
- The combined command chain ended with exit code 141 due to `head` closing the pipe early (SIGPIPE). This was not a functional failure, but it did make the shell exit non-zero.

### What I learned
- When capturing “just a preview” of output in automation, prefer patterns that avoid SIGPIPE (e.g., redirect to a file and show a slice, or explicitly tolerate SIGPIPE) so success doesn’t look like failure.

### Technical details
- Commands run:

```bash
go build -o /tmp/docmgr-refactor-local-2025-12-13 ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-refactor-local-2025-12-13 bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-local-2025-12-13
/tmp/docmgr-refactor-local-2025-12-13 list docs --root /tmp/docmgr-scenario-local-2025-12-13/acme-chat-app/ttmp
/tmp/docmgr-refactor-local-2025-12-13 list docs --root /tmp/docmgr-scenario-local-2025-12-13/acme-chat-app/ttmp --with-glaze-output --output json --select path
```

## Step 15: Port `doctor` to `Workspace.QueryDocs` + align RelatedFiles checks with Workspace normalization

This step refactors `docmgr doctor` to use the new Workspace+SQLite index as its primary source of truth for “what docs exist”, instead of running its own ticket discovery and ad-hoc filesystem walkers. The intent is consistency: once `list docs`, `doc search`, and `doctor` all sit on the same `QueryDocs` foundation, we stop having three subtly different interpretations of “workspace” and “document”, and diagnostics become much easier to reason about.

The second focus was the `RelatedFiles` existence checks. Historically, `doctor` tried a handful of heuristics (repo root, config dir, cwd, etc.) to decide whether a referenced file exists. That logic was drifting away from the actual normalization logic used by the index. In this port, file existence validation now uses a **doc-anchored** `paths.Resolver`, meaning it checks the same base anchors (repo, doc dir, config, docs root, docs parent) that we rely on for reverse lookup.

### What I did
- Rewrote `pkg/commands/doctor.go` to:
  - discover the workspace via `workspace.DiscoverWorkspace`
  - build the in-memory index once via `ws.InitIndex`
  - query the doc set via `ws.QueryDocs` (with `IncludeErrors=true`, `IncludeDiagnostics=true`, and “include special paths” toggles enabled)
  - apply `--ignore-dir` / `--ignore-glob` as a post-filter on the query results (to preserve the legacy CLI behavior)
- Switched `RelatedFiles` validation to use `paths.NewResolver(ResolverOptions{DocPath: <current-doc>})` and `Normalize(...).Exists` instead of a manual candidate list.
- Tightened `--ticket` behavior so “missing index” scaffolds are only reported for the requested ticket (previously this could leak unrelated tickets into the output).
- Updated `QueryDocs` to keep a best-effort `Doc.Ticket` even for parse-error docs so `doctor` can group findings correctly without extra lookups.
- Updated `QueryDocs` unit tests to match the new parse-error behavior.

### Why
- `doctor` is one of the key “consistency” commands: if it doesn’t use the central index, then the refactor’s “single source of truth” promise is incomplete.
- Aligning `RelatedFiles` existence checks with the same resolver used by ingestion/query reduces “doctor says missing, search says found” contradictions.

### What worked
- A local `go run ./cmd/docmgr doctor --ticket ...` smoke-test succeeded and produced expected findings for the ticket being scanned.
- `go test ./...` remained green after the changes.

### What didn’t work
- In the first cut, `--ticket <ID>` could still emit “missing index” findings for other tickets, because the missing-index scaffold detection was repo-wide. This is now filtered by ticket dir prefix when `--ticket` is provided.

### What I learned
- If a command has “two discovery mechanisms” (filesystem scan + index query), it’s very easy for them to disagree in edge cases. Using QueryDocs as the baseline and layering only truly filesystem-specific checks on top is the safest route.

### What was tricky to build
- **Doc-relative file checks**: the Workspace-level resolver is constructed without a `DocPath`, so it can’t do doc-relative resolution correctly. The fix was creating a resolver per document using `ResolverOptions{DocPath: <doc>}` for `RelatedFiles` validation.
- **Ticket scoping vs repo-wide scaffolds**: “missing index” detection doesn’t naturally belong to QueryDocs (because the index is built from docs), so the ticket filter has to be applied explicitly when users scope the command.

### What warrants a second pair of eyes
- **Ticket grouping logic**: `doctor` groups docs by inferring ticket root from the `<docsRoot>/YYYY/MM/DD/<ticketDir>/...` layout. This is pragmatic, but it’s worth a careful review for off-by-one assumptions and weird path edge cases.
- **Ignore semantics**: we preserve `--ignore-glob` by post-filtering QueryDocs results. Review that we apply the globs consistently (basename + absolute path + docs-root-relative path) and don’t accidentally hide important diagnostics.

### Code review instructions (for a reviewer)
- Start with `pkg/commands/doctor.go` and look for:
  - `workspace.DiscoverWorkspace` / `ws.InitIndex` / `ws.QueryDocs` (the new “single discovery” spine)
  - `paths.NewResolver(...DocPath: indexPath...)` and `Normalize(...).Exists` (the new existence check)
  - `matchesTicketDir(...)` and the missing-index filtering under `--ticket`
- Then check `internal/workspace/query_docs.go` and confirm the parse-error behavior is still aligned with the design contract: parse-error docs still carry `ReadErr`, but now also carry a minimal `Ticket` for grouping.

### Technical details
- Files:
  - `pkg/commands/doctor.go`
  - `internal/workspace/query_docs.go`
  - `internal/workspace/query_docs_test.go`
- Commands:

```bash
gofmt -w internal/workspace/query_docs.go internal/workspace/query_docs_test.go pkg/commands/doctor.go
go test ./... -count=1
go run ./cmd/docmgr doctor --ticket REFACTOR-TICKET-REPOSITORY-HANDLING --fail-on none
```

**Commit**: `a5af7454d006e6d80989f5cb5a82b190562f42ea`

## Step 16: Port `doc relate` to Workspace-based doc lookup + normalization

This step finishes the “plumbing consistency” story for file relationships: `doc relate` no longer tries to rediscover ticket directories with ad-hoc filesystem helpers, and instead resolves its target document via the same `Workspace.QueryDocs` index that powers list/search/doctor. That means “which doc am I editing?” is now backed by the central index, not bespoke heuristics.

The second part of the change is about normalization correctness. The command now constructs its `paths.Resolver` directly from the Workspace context (docs root, config dir, repo root) and anchors it on the **target doc path**. That ensures that doc-relative paths entered by the user (and doc-relative paths already in frontmatter) normalize the same way the index does at ingest time, which reduces reverse lookup surprises.

### What I did
- Refactored `pkg/commands/relate.go` to:
  - discover a `Workspace` via `workspace.DiscoverWorkspace`
  - build the ephemeral SQLite index once via `ws.InitIndex`
  - resolve `--doc` targets via `QueryDocs(ScopeDoc)`
  - resolve `--ticket` targets via `QueryDocs(ScopeTicket + DocType=index)` and selecting `index.md`
  - construct a doc-anchored `paths.Resolver` from `ws.Context()` (including `RepoRoot`)
- Switched the “existing docs” portion of `--suggest` to query via `QueryDocs` instead of walking the filesystem, so skip rules and parse behavior match the rest of the tool.
- Smoke-tested both modes without mutating any files by issuing a no-op removal:
  - `--doc <path> --remove-files does/not/exist.go`
  - `--ticket <ID> --remove-files does/not/exist.go`

### Why
- `doc relate` is the write-path for `RelatedFiles`. If it resolves docs differently than QueryDocs, we end up with confusing “write here, read there” mismatches.
- Using a Workspace-derived resolver with `DocPath=<target doc>` makes doc-relative resolution deterministic and consistent with the ingestion/query normalization pipeline.

### What worked
- Both `--doc` and `--ticket` modes resolve their target docs through QueryDocs and correctly report no-op when no changes were requested.
- `go test ./...` stays green after the refactor.

### What didn’t work
- Nothing notable in this port; the main “gotcha” was making sure we still support the suggestion heuristics that need a filesystem root (git/ripgrep), while moving doc discovery itself to QueryDocs.

### What I learned
- `ScopeDoc` is a great “escape hatch” for commands that accept an explicit doc path: normalize it once through the Workspace resolver, then treat the absolute path as the stable key (`docs.path`).

### What was tricky to build
- **Picking the ticket index doc reliably**: for `--ticket`, QueryDocs returns “docs in the ticket”, but we still have to decide which one is the index. Filtering by `DocType=index` and then requiring `basename=index.md` kept that selection explainable.
- **Suggestions need two roots**: doc scanning can be index-backed, but git/ripgrep still needs an OS directory root; inferring the ticket dir from the doc path preserves the old behavior without reintroducing ticket-directory walkers.

### What warrants a second pair of eyes
- **Ticket-dir inference** (`inferTicketDirFromDocPath`) assumes the default docs layout `<docsRoot>/YYYY/MM/DD/<ticketDir>/...`. If we ever support alternate layouts, this is one of the first helpers that should become configurable.
- **Suggestion semantics**: moving “existing docs” from filesystem walk to QueryDocs changes which docs are considered (e.g. it now honors canonical skip rules). That’s good, but worth confirming against any workflows that relied on indexing skipped paths.

### Code review instructions (for a reviewer)
- Start with `pkg/commands/relate.go`:
  - look for `DiscoverWorkspace` → `InitIndex` → `QueryDocs` (doc resolution spine)
  - confirm `paths.NewResolver(... DocPath: targetDocPath, RepoRoot: ws.Context().RepoRoot ...)` is used for canonicalization
  - verify `--ticket` selection logic (DocType filter + `index.md` basename check)
- Run the two smoke tests that should not mutate files:

```bash
go run ./cmd/docmgr doc relate --doc ttmp/.../reference/15-diary.md --remove-files does/not/exist.go
go run ./cmd/docmgr doc relate --ticket REFACTOR-TICKET-REPOSITORY-HANDLING --remove-files does/not/exist.go
```

### Technical details
- Files:
  - `pkg/commands/relate.go`
- Commands:

```bash
gofmt -w pkg/commands/relate.go
go test ./... -count=1
```

**Commit**: `859f86d128f0acf6b8fda37976c86589e7f15861`

## Step 17: Create a “cleaning inspectors” brief for Task 18

This step sets up the next phase of the refactor (Task 18) in a way that stays tedious-but-conscientious without getting chaotic. Instead of jumping straight into deleting helpers and rewiring commands, we created a structured brief for a “crew of cleaning inspectors” to inventory all duplicated walkers/helpers and to define crisp cleanup guidelines. The goal is to turn cleanup into a checklist-driven sequence of small PRs rather than a messy, high-risk refactor blob.

The deliverable is intentionally procedural: it describes what to search for, how to classify each finding, what the canonical replacements are (Workspace discovery + QueryDocs + resolver), and how to validate each cleanup step. This should make it easy to delegate the reporting work and keep implementation decisions consistent.

### What I did
- Added a new ticket doc: `analysis/08-cleanup-inspectors-brief-task-18.md`
- Filled it with:
  - mission + scope + constraints
  - a hard rule that inspectors must not modify code (report-only work)
  - inventory table template (required fields)
  - search patterns + starting file map (known hotspots)
  - cleanup guidelines (“when you see X, replace with Y”)
  - validation guidance (tests + scenario suite)
- Related key starting-point files to the brief doc (so inspectors can jump directly).

### Why
- Task 18 touches many commands and helper paths; without a structured inventory it’s easy to miss edge cases or regress behavior.
- A written guideline set keeps “cleanup” consistent: fewer one-off choices, fewer accidental semantic changes.

### What worked
- We now have a single place that enumerates what to look for and how to report it, so implementation can proceed in small, low-risk chunks.

### What didn’t work
- Nothing yet; this is intentionally a planning/reporting step.

### What I learned
- A cleanup phase is much safer when the *reporting burden* is separated from the *implementation burden*.

### What was tricky to build
- **Balancing strictness vs flexibility**: we want strong “replace with Workspace” guidance without forbidding filesystem-level operations where they’re actually required (e.g., true file moves).

### What warrants a second pair of eyes
- **Starting map completeness**: verify we didn’t miss obvious command hotspots (especially anything still calling `findTicketDirectory` or doing `filepath.Walk*` over docs).

### Code review instructions (for a reviewer)
- Read `analysis/08-cleanup-inspectors-brief-task-18.md` and confirm:
  - the inventory template is actionable
  - the replacement guidance matches current Workspace/QueryDocs semantics
  - the “where to look first” list corresponds to actual code usage

### Technical details
- New doc:
  - `ttmp/.../analysis/08-cleanup-inspectors-brief-task-18.md`

---

## Step 18: Validate inventory report and create CLEANUP-LEGACY-WALKERS ticket (Task 18 complete)

This step completes Task 18 by validating the cleanup inventory report generated by the "inspectors", then spinning off a dedicated ticket for the cleanup work. The refactor ticket focuses on building the Workspace + QueryDocs infrastructure; the cleanup ticket focuses on systematically migrating all remaining commands to use it.

Validation confirmed the inventory report is accurate: all `findTicketDirectory` call sites (20 across 15 files), all `CollectTicketWorkspaces` usages, and all `filepath.Walk*` locations match the actual codebase. A few minor omissions were identified (e.g., `list.go:79` for `CollectTicketWorkspaces`, extra `readDocumentFrontmatter` calls in `search.go`), but the report is actionable as-is.

The new ticket (CLEANUP-LEGACY-WALKERS) includes a comprehensive design document explaining the cleanup scope, migration patterns, risk notes, and a phased PR plan (4 phases, 23 tasks). This makes the cleanup work tractable for any developer picking it up.

### What I did
- Validated the inventory report (`09-cleanup-inventory-report-task-18.md`) against actual grep results
- Created ticket CLEANUP-LEGACY-WALKERS
- Wrote `design/01-cleanup-overview-and-migration-guide.md` with:
  - Background and context (what was built, what needs cleanup)
  - Inventory summary (23 targets across 12 files)
  - Migration patterns (A: ticket discovery, B: doc enumeration, C: skip rules, D: parsing)
  - Risk notes (behavior changes, write-path commands)
  - Phased PR plan (Phase 1–4)
  - Validation strategy
- Populated `tasks.md` with 23 tasks organized by phase
- Related key files to the new ticket
- Updated changelogs for both tickets

### Why
- Task 18 is too broad for a single PR; splitting it into a dedicated ticket with granular tasks ensures traceable, low-risk migration.
- The refactor ticket can now be closed (or marked as "infrastructure complete"), while cleanup continues in the new ticket.

### What worked
- Validation using grep confirmed the report's line numbers are accurate
- The phased PR plan makes cleanup tractable (Phase 1 is low-risk, high-impact)

### What didn't work
- Nothing; this was primarily a documentation and ticket management step

### What I learned
- Validating automated/agent-generated reports before creating tickets catches minor gaps early
- Spinning off cleanup into a separate ticket keeps the infrastructure ticket focused

### What was tricky to build
- **Nothing tricky** — this was a documentation step

### What warrants a second pair of eyes
- **Phase 1 PR ordering**: verify `status.go` → `list_tickets.go` → `list.go` is the right sequence
- **Behavior change for ticket filter**: `list_tickets.go` changes from substring to exact match; may need user communication

### Code review instructions (for a reviewer)
- Review the cleanup ticket's design doc and tasks for completeness
- Verify the phased plan aligns with the inventory report
- Confirm no critical commands were missed in the task list

### Technical details
- New ticket: CLEANUP-LEGACY-WALKERS
- New docs:
  - `ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/design/01-cleanup-overview-and-migration-guide.md`
  - `ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/tasks.md` (23 tasks)

### Git commit
- (pending)
