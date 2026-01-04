---
Title: Diary
Ticket: 005-USE-SQLITE-FTS
Status: active
Topics:
    - backend
    - docmgr
    - tooling
    - testing
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/analysis/01-analysis-fts-backed-search-refactor-search-packages.md
      Note: Analysis written during diary steps
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T14:10:00.679628143-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Capture a frequent, tiny-step diary of analysis work for adding SQLite FTS-backed search to docmgr and refactoring the search implementation out of the CLI command layer.

## Step 1: Bootstrap ticket + seed docs

This step created the ticket workspace and the initial diary + analysis documents, so the upcoming technical analysis has a place to live and can be reviewed/continued.

**Commit (code):** N/A

### What I did
- Created the ticket workspace:
  - `docmgr ticket create-ticket --ticket 005-USE-SQLITE-FTS --title "Use SQLite FTS for search; refactor search out of commands" --topics backend,docmgr,tooling,testing`
- Created the seed docs:
  - `docmgr doc add --ticket 005-USE-SQLITE-FTS --doc-type reference --title "Diary"`
  - `docmgr doc add --ticket 005-USE-SQLITE-FTS --doc-type analysis --title "Analysis: FTS-backed search + refactor search packages"`

### Why
- We want a thorough analysis trail and an actionable refactor plan before writing code, because this touches UX-critical search semantics.

### What worked
- Ticket + docs scaffolded cleanly; `analysis` doc-type is available in vocabulary.

### What didn't work
- N/A

### What I learned
- `docmgr` emits a debug log `Created FTS5 tables and triggers` during most commands; this comes from the Glazed help store’s FTS support (`glazed/pkg/help/store/fts5.go`) and indicates the binary is built with the `sqlite_fts5` build tag (FTS is likely available in-process).

### What was tricky to build
- N/A (docs only)

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- N/A

### Code review instructions
- N/A

### Technical details
- Ticket docs:
  - `docmgr/ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/analysis/01-analysis-fts-backed-search-refactor-search-packages.md`
  - `docmgr/ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/reference/01-diary.md`

## Step 2: Survey current search + identify FTS integration seam

This step mapped where “search” behavior lives today and identified the smallest safe seam to introduce FTS without breaking CLI semantics. The key observation is that metadata + reverse lookup already use an SQLite index (`Workspace.QueryDocs`), but `--query` is a post-filter substring scan over bodies in `pkg/commands/search.go`.

**Commit (code):** N/A

### What I did
- Located the current search implementation:
  - `pkg/commands/search.go` (`SearchCommand`, `SearchSettings`, post-filters, `suggestFiles`)
  - `internal/workspace/*` (schema, ingestion, query engine)
- Located existing FTS5 patterns in-repo:
  - `scenariolog/internal/scenariolog/migrate.go` (`ensureFTS5`, graceful degraded mode)
  - `glazed/pkg/help/store/fts5.go` (FTS virtual table + triggers, build-tagged `sqlite_fts5`)

### Why
- Implementing FTS in docmgr should reuse existing patterns for:
  - build tag enablement (`sqlite_fts5`),
  - runtime graceful fallback if FTS is missing,
  - and avoiding semantic drift by keeping the “search engine” separate from CLI formatting.

### What worked
- The query engine is already cleanly separated:
  - `internal/workspace/index_builder.go` builds the docs+topics+related_files tables.
  - `internal/workspace/query_docs_sql.go` compiles SQL; this is the natural place to inject FTS joins / MATCH clauses.

### What didn't work
- N/A

### What I learned
- There is no FTS table in docmgr’s workspace schema today; `--query` does not participate in SQL compilation and therefore cannot benefit from indexing or ranking.

### What was tricky to build
- N/A (analysis only)

### What warrants a second pair of eyes
- When we implement, we need careful review of any changes to reverse lookup behavior (`--file`/`--dir`), because it’s UX-critical and has explicit “compatibility” fallbacks.

## Step 3: Write analysis doc (FTS plan + package refactor plan)

This step captured a parity-first FTS integration plan (so we can speed up `--query` without changing user-visible behavior), and a concrete package refactor plan so the CLI becomes a thin adapter over a reusable search engine.

**Commit (code):** N/A

### What I did
- Wrote: `analysis/01-analysis-fts-backed-search-refactor-search-packages.md`
- Verified the origin of the “Created FTS5 tables and triggers” debug log:
  - It comes from `glazed/pkg/help/store/fts5.go` which is build-tagged with `sqlite_fts5`.

### Why
- FTS affects semantics (tokenization vs substring), so the analysis explicitly calls out how to keep parity (use FTS to narrow candidates, then substring post-filter).
- A refactor is needed to avoid duplicating search semantics across CLI and future HTTP API handlers.

### What worked
- Existing patterns in-repo (scenariolog + glazed) provide:
  - build-tag enablement conventions (`sqlite_fts5`)
  - runtime degraded mode logic (“no such module: fts5” → fall back)

### What didn't work
- N/A

### What I learned
- The current workspace index schema (`internal/workspace/sqlite_schema.go`) does not define any FTS tables; adding them is isolated and doesn’t require changing reverse lookup logic.

### What was tricky to build
- The biggest subtlety is query semantics: FTS MATCH is not “substring contains”, so parity requires either careful quoting or an explicit post-filter.

### What warrants a second pair of eyes
- The eventual decision of which schema option to adopt first (contentless vs `content='docs' + rebuild`) and the UX implications of FTS query interpretation.

## Step 4: Update design: no compatibility, schema option C, add rank ordering

This step pivoted the plan: we explicitly dropped backwards compatibility for `--query` behavior and chose schema option C (index multiple fields) plus rank ordering. This made implementation decisions much simpler because we no longer need to preserve substring semantics.

**Commit (code):** N/A

### What I did
- Wrote/updated the ticket design doc:
  - `design-doc/01-design-fts-backed-search-engine-no-compatibility.md`
- Added concrete tasks to the ticket:
  - `docmgr task add ...` (see `tasks.md`)

### Why
- The “parity-first” plan was unnecessary; we want the new behavior to be explicitly FTS-based and better for UX (title/topics/doc_type/ticket_id are queryable).

### What worked
- The design doc now cleanly specifies:
  - option C schema
  - `OrderByRank` via `bm25`
  - extract snippet moved into the shared engine package
  - CLI as a thin adapter

### What didn't work
- N/A

### What I learned
- For FTS semantics, the biggest “semantic break” is tokenization and query language; with compatibility dropped, we can treat the query string as raw FTS syntax.

### What warrants a second pair of eyes
- Validate whether we want to treat `--query` as raw FTS syntax long-term, or later add a `--query-mode` to support “phrase mode” without requiring users to know FTS syntax.

## Step 5: Implement FTS index + reusable search engine; refactor CLI to thin layer

This step implemented the core of the design:

- `docs_fts` gets created and populated during workspace ingest (option C fields).
- `Workspace.QueryDocs` supports `DocFilters.TextQuery` and `OrderByRank` ordering.
- A reusable engine package (`internal/searchsvc`) owns the search query model and snippet generation.
- The CLI search command delegates to the engine (and `--files` delegates to the shared suggestion code).

**Commit (code):** N/A

### What I did
- Implemented FTS in the workspace index:
  - `internal/workspace/sqlite_schema.go` (`ensureDocsFTS5`, `ErrFTSNotAvailable`)
  - `internal/workspace/index_builder.go` (create/populate `docs_fts`, store `ftsAvailable`)
  - `internal/workspace/query_docs.go` (`DocFilters.TextQuery`, `OrderByRank`, FTS availability check)
  - `internal/workspace/query_docs_sql.go` (JOIN + MATCH + ORDER BY bm25)
- Added reusable search engine package:
  - `internal/searchsvc/search.go` (`SearchQuery`, `SearchDocs`)
  - `internal/searchsvc/snippet.go` (`ExtractSnippet`)
  - `internal/searchsvc/date.go` (`ParseDate`)
  - `internal/searchsvc/suggest_files.go` (`SuggestFiles`, and exported helper heuristics used by other commands)
- Refactored CLI search to be a thin adapter:
  - `pkg/commands/search.go` now calls `internal/searchsvc` for docs search and file suggestions
  - Added CLI flag `--order-by path|last_updated|rank`
  - Updated completion: `cmd/docmgr/cmds/doc/search.go`
- Updated other commands that relied on old search.go helpers:
  - `pkg/commands/changelog.go` and `pkg/commands/relate.go` now call `internal/searchsvc` for git/ripgrep suggestions
- Added tests:
  - `internal/workspace/query_docs_fts5_test.go` (`//go:build sqlite_fts5`)
- Ran:
  - `gofmt -w ...`
  - `go test ./...`
  - `go test -tags sqlite_fts5 ./...`

### Why
- Centralizing search semantics in `internal/searchsvc` is the simplest way to keep CLI and future HTTP handlers consistent.

### What worked
- Both default and `-tags sqlite_fts5` test suites pass.

### What was tricky to build
- Ensuring FTS is treated as best-effort in schema creation, while making `--query` fail clearly when FTS isn’t present (instead of silently producing surprising empty results).

### What warrants a second pair of eyes
- Review the SQL compilation changes in `internal/workspace/query_docs_sql.go` for correctness and performance (JOIN placement and WHERE ordering).
