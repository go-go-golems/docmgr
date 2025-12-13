---
Title: Code review walkthrough (diary-driven)
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/query_docs.go
      Note: QueryDocs implementation reviewed in Steps 10-12.
    - Path: internal/workspace/workspace.go
      Note: Workspace entry point reviewed in Step 2.
    - Path: pkg/commands/list_docs.go
      Note: list docs port reviewed in Step 13.
    - Path: pkg/commands/search.go
      Note: doc search port reviewed in Step 14.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Design spec referenced for spec section numbers.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/15-diary.md
      Note: Implementation diary used as the narrative anchor for this review walkthrough.
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T00:00:00Z
---


# Code Review Walkthrough (Diary-Driven)

## Goal

Guide code reviewers through the Workspace+SQLite refactor implementation step-by-step, using the **implementation diary** (`reference/15-diary.md`) as the narrative anchor. For each diary step, this document explains what was built, where it lives (files + key symbols), how to test it, what needs careful review, and what was tricky.

## Context

This walkthrough complements:
- **Design spec** (`design/01-workspace-sqlite-repository-api-design-spec.md`): the "what and why" of the architecture
- **Senior code review guide** (`analysis/03-code-review-guide-senior.md`): spec-to-code mapping + risk assessment
- **Implementation diary** (`reference/15-diary.md`): the full implementation narrative

This document is for reviewers who want to follow the **implementation timeline** and understand the **evolution of thinking** (failed attempts, learned lessons, iterative refinements).

## How to Use This Document

1. **Read linearly** if you want the full story (setup → core → queries → command ports)
2. **Jump to specific steps** if you only care about certain subsystems:
   - **Steps 1-5**: scaffolding + testing strategy
   - **Steps 6-9**: SQLite schema + ingestion + normalization + export
   - **Steps 10-12**: query engine (debug hang, optimize, diagnostics)
   - **Steps 13-14**: command ports (`list docs`, `search`)

For each step, the structure is:
- **What was built**: high-level summary
- **Where it is**: files + key symbols (functions, types)
- **How to exercise**: commands/tests to run
- **Needs scrutiny**: what could go wrong / what needs extra eyes
- **What was tricky**: implementation gotchas / debugging notes

---

## Step 1: Kickoff — diary + baseline scan

**What was built**: Project scaffolding for tracking implementation progress.

**Where it is**:
- `ttmp/.../reference/15-diary.md` (the diary itself)
- Changelog entries in `ttmp/.../changelog.md`

**How to exercise**:
- N/A (meta step)

**Needs scrutiny**:
- N/A

**What was tricky**:
- Nothing; this was purely organizational setup.

---

## Step 2: Implement `internal/workspace.Workspace` skeleton (Task 2, Spec §5.1)

**What was built**: The new entry-point object for workspace discovery and normalization, without changing any user-facing behavior yet.

**Where it is**:
- **File**: `internal/workspace/workspace.go`
- **Key symbols**:
  - `type Workspace`
  - `type WorkspaceContext`
  - `func DiscoverWorkspace(ctx, opts) (*Workspace, error)`
  - `func NewWorkspaceFromContext(ctx) (*Workspace, error)`
  - `func (*Workspace) Resolver() *paths.Resolver`

**How to exercise**:
```go
// Unit test pattern:
ws, err := workspace.NewWorkspaceFromContext(workspace.WorkspaceContext{
    Root:      "/path/to/ttmp",
    ConfigDir: "/path/to/repo",
    RepoRoot:  "/path/to/repo",
})
assert.NoError(t, err)
resolver := ws.Resolver()
assert.NotNil(t, resolver)
```

**Needs scrutiny**:
- **Discovery heuristics** in `DiscoverWorkspace`:
  - ConfigDir fallback (`filepath.Dir(root)`) may be wrong in monorepos
  - RepoRoot is required (hard error if `FindRepositoryRoot` fails)
- **Config load error handling**: malformed config currently produces a warning but we proceed with `Config=nil`; verify this is acceptable for all commands

**What was tricky**:
- `LoadWorkspaceConfig` returns an `error` even when it prints a warning and "continues". We treat that as non-fatal in `DiscoverWorkspace`, but the interface is confusing.

**Commit**: `5f1681b1f3ac5ed7ffa36fb4b2357b54ebaf6695`

---

## Step 3-5: Testing strategy + baseline runs

**What was built**: Testing approach doc + baseline + local-binary scenario runs.

**Where it is**:
- `ttmp/.../analysis/02-testing-strategy-integration-first.md`
- Scenario scripts: `test-scenarios/testing-doc-manager/run-all.sh` (and friends)

**How to exercise**:
```bash
# System docmgr
DOCMGR_PATH=$(which docmgr) bash test-scenarios/testing-doc-manager/run-all.sh /tmp/baseline

# Local build
go build -o /tmp/docmgr-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/local
```

**Needs scrutiny**:
- Confirm baseline and local runs produce identical output for commands not yet refactored

**What was tricky**:
- Nothing; this is pure validation infrastructure.

---

## Step 6: Add in-memory SQLite schema (Task 3, Spec §9.1–§9.2)

**What was built**: The contract for what the in-memory index will store.

**Where it is**:
- **File**: `internal/workspace/sqlite_schema.go`
- **Key symbols**:
  - `func openInMemorySQLite(ctx) (*sql.DB, error)`
  - `func createWorkspaceSchema(ctx, db) error`
  - Tables: `docs`, `doc_topics`, `related_files`

**How to exercise**:
```go
db, err := openInMemorySQLite(ctx)
assert.NoError(t, err)
err = createWorkspaceSchema(ctx, db)
assert.NoError(t, err)

// Verify tables exist
var name string
db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='docs'").Scan(&name)
assert.Equal(t, "docs", name)
```

**Needs scrutiny**:
- **Pragmas** (`journal_mode=OFF`, `synchronous=OFF`): verify these are safe for in-memory use
- **Schema indexes**: confirm indexes on `ticket_id`, `parse_ok`, `path_tags`, `norm_repo_rel`, `norm_abs` support expected query patterns
- **Foreign keys**: verify `ON DELETE CASCADE` is desired (deleting a doc deletes its topics/related_files)

**What was tricky**:
- Nothing; schema is intentionally minimal at this stage.

**Commit**: `5b72f4ec08faad89c1219030daf4c96ca5456db0`

---

## Step 7: Canonical ingest-time skip rules + tagging (Task 4, Spec §6)

**What was built**: Single canonical definition of "what gets indexed" and "what tags get applied".

**Where it is**:
- **File**: `internal/workspace/skip_policy.go`
- **Key symbols**:
  - `type PathTags`
  - `func DefaultIngestSkipDir(relPath, d) bool`
  - `func ComputePathTags(docPath) PathTags`
  - `func containsPathSegment(slashPath, seg) bool` (segment-boundary checks)

**How to exercise**:
```go
tags := workspace.ComputePathTags("/path/to/ttmp/2025/12/12/TICKET/tasks.md")
assert.True(t, tags.IsControlDoc)  // sibling index.md exists

tags = workspace.ComputePathTags("/path/to/ttmp/2025/12/12/TICKET/archive/old.md")
assert.True(t, tags.IsArchivedPath)

tags = workspace.ComputePathTags("/path/to/ttmp/2025/12/12/myarchive-project/doc.md")
assert.False(t, tags.IsArchivedPath)  // false positive avoidance
```

**Needs scrutiny**:
- **Segment-boundary logic**: confirm `containsPathSegment` correctly differentiates `/archive/` from `/myarchive/`
- **Control doc detection**: confirm `hasSiblingIndex` file IO is acceptable during ingestion (happens once per file)
- **Skip policy consistency**: verify `.meta/` and `_*/` skipping matches all other parts of the refactor (no gaps)

**What was tricky**:
- The segment-boundary check using `"/"+seg+"/"` is subtle but necessary to avoid false positives.

---

## Step 8: Implement ingestion walker + index build (Task 5, Spec §6 / §7.1 / §9.2)

**What was built**: The core ingestion logic that populates the in-memory SQLite index.

**Where it is**:
- **Files**: `internal/workspace/index_builder.go`, `internal/workspace/workspace.go`
- **Key symbols**:
  - `func (*Workspace) InitIndex(ctx, opts) error`
  - `func ingestWorkspaceDocs(ctx, db, wctx, opts) error`
  - `type BuildIndexOptions`

**How to exercise**:
```go
ws, _ := workspace.NewWorkspaceFromContext(workspace.WorkspaceContext{
    Root: testRoot, ConfigDir: testRepo, RepoRoot: testRepo,
})
err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false})
assert.NoError(t, err)

db := ws.DB()
var count int
db.QueryRow("SELECT COUNT(*) FROM docs").Scan(&count)
assert.Greater(t, count, 0)
```

**Needs scrutiny**:
- **Transaction handling**: confirm ingestion is all-or-nothing (single tx wraps the entire walk)
- **Parse error handling**: verify broken docs get `parse_ok=0` + `parse_err` populated correctly
- **Ticket ID inference**: when parse fails, confirm the `ttmp/YYYY/MM/DD/TICKET--slug/...` layout is used to infer `ticket_id`
- **Path normalization during ingest**: confirm `paths.Resolver` is constructed with `DocPath=absPath` so doc-relative `RelatedFiles` normalize correctly

**What was tricky**:
- Per-doc `paths.Resolver` construction during ingestion is key; using the wrong anchor makes doc-relative paths fail.

---

## Step 9: Path normalization pipeline + persisted keys (Task 6, Spec §7.3 / §12.1 / §9.2)

**What was built**: Expanded `related_files` schema to persist multiple normalized keys for reliable reverse lookup.

**Where it is**:
- **Files**: `internal/workspace/normalization.go`, `internal/workspace/sqlite_schema.go`, `internal/workspace/index_builder.go`
- **Key symbols**:
  - `func normalizeRelatedFilePath(resolver, raw) NormalizedKeys`
  - Columns: `norm_canonical`, `norm_repo_rel`, `norm_docs_rel`, `norm_doc_rel`, `norm_abs`, `norm_clean`, `anchor`, `raw_path`

**How to exercise**:
```sql
-- Query exported DB to see normalized keys
SELECT raw_path, norm_canonical, norm_repo_rel, norm_doc_rel, anchor
FROM related_files
WHERE raw_path = 'pkg/commands/search.go';
```

**Needs scrutiny**:
- **Fallback coverage**: verify at least one `norm_*` column is always populated (even if repo root detection fails)
- **Anchor assignment**: confirm `anchor` field correctly identifies which resolution path succeeded
- **Canonical key selection**: verify priority order (repo > docs > doc > abs > clean) is reasonable

**What was tricky**:
- Storing 7 different representations (including original raw) means `related_files` rows are "wide", but this is the entire reason reverse lookup becomes reliable.

---

## Step 10: Debug QueryDocs test hang + stabilize in-memory SQLite

**What was built**: Fixes for deadlock + test flakiness in `QueryDocs`.

**Where it is**:
- **Files**: `internal/workspace/sqlite_schema.go`, `internal/workspace/index_builder.go`
- **Key changes**:
  - `db.SetMaxOpenConns(4)` / `SetMaxIdleConns(4)` (from 1)
  - Unique in-memory DB name per `Workspace`
  - Inferred `ticket_id` for parse-error docs

**How to exercise**:
```bash
# Repro the original hang (before fix):
go test ./internal/workspace -run TestWorkspaceQueryDocs -count=1 -timeout 5s
# Should complete quickly (not hang)
```

**Needs scrutiny**:
- **Connection pool size**: confirm 4 connections is enough for nested hydration patterns in `QueryDocs`
- **Unique DB naming**: verify the counter/uniqueness mechanism doesn't leak memory or descriptors
- **Ticket ID inference**: confirm the `ttmp/YYYY/MM/DD/TICKET--slug/...` pattern matching is correct

**What was tricky**:
- The deadlock only manifested with `MaxOpenConns=1` + nested queries; it was easy to miss until a test hung.
- Shared in-memory DB name caused test cross-contamination (flaky failures).

---

## Step 11: Remove nested queries / N+1 from `Workspace.QueryDocs`

**What was built**: Refactor of `QueryDocs` to use batch hydration instead of per-row queries.

**Where it is**:
- **File**: `internal/workspace/query_docs.go`
- **Key symbols**:
  - `func fetchTopicsByDocIDs(ctx, db, docIDs) (map[int64][]string, error)`
  - `func fetchRelatedFilesByDocIDs(ctx, db, docIDs) (map[int64]models.RelatedFiles, error)`
  - Hydration now happens after base `docs` query completes

**How to exercise**:
- Same QueryDocs unit tests should pass (no behavioral change)
- Query count: 1 base + 2 batch (not 1 base + 2N nested)

**Needs scrutiny**:
- **Batch query correctness**: verify `WHERE doc_id IN (...)` with placeholders is safe and correct
- **Hydration completeness**: confirm all parse-ok docs get topics/related_files populated
- **Parse-error docs**: confirm broken docs are not hydrated (Doc stays nil)

**What was tricky**:
- Building the `IN (...)` clause with correct placeholder count and arg ordering.

---

## Step 12: Implement QueryDocs diagnostics (parse skips + normalization fallback)

**What was built**: Diagnostics payload in `DocQueryResult`.

**Where it is**:
- **Files**: `pkg/diagnostics/docmgrctx/query_docs.go`, `internal/workspace/query_docs.go`, `internal/workspace/query_docs_sql.go`
- **Key symbols**:
  - `type QuerySkippedDueToParse` (context payload)
  - `type QueryNormalizationFallback` (context payload)
  - `func newQuerySkippedTaxonomy(...)` / `func newQueryNormalizationFallbackTaxonomy(...)`
  - `func compileDocQueryWithParseFilter(..., parseOKFilter *int)`

**How to exercise**:
```go
result, _ := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Options: workspace.DocQueryOptions{
        IncludeErrors: false,  // hide broken docs from Docs
        IncludeDiagnostics: true,  // but emit them as Diagnostics
    },
})
assert.Greater(t, len(result.Diagnostics), 0)  // broken docs appear here
```

**Needs scrutiny**:
- **Diagnostics completeness**: confirm `parse_ok=0` docs are surfaced as diagnostics when `IncludeErrors=false` + `IncludeDiagnostics=true`
- **Normalization fallback warnings**: verify they're emitted when reverse lookup inputs have weak/empty keys
- **Taxonomy shape**: confirm `Tool`, `Stage`, `Symptom`, `Severity`, `Context`, `Path` are populated correctly

**What was tricky**:
- Parameterized SQL compiler (`compileDocQueryWithParseFilter`) to support "parse_ok=0" diagnostic queries without duplicating logic.

---

## Step 13: Port `list docs` to `Workspace.QueryDocs`

**What was built**: First command integration using the new Workspace API.

**Where it is**:
- **File**: `pkg/commands/list_docs.go`
- **Key changes**:
  - Removed `filepath.Walk` + manual parsing
  - Added `workspace.DiscoverWorkspace` + `ws.InitIndex` + `ws.QueryDocs`
  - Preserved `index.md` skipping as post-filter

**How to exercise**:
```bash
# Build local binary
go build -o /tmp/docmgr-local ./cmd/docmgr

# Run against a test workspace
/tmp/docmgr-local list docs --root /path/to/ttmp

# Glaze output
/tmp/docmgr-local list docs --root /path/to/ttmp --with-glaze-output --output json

# Compare with system docmgr
diff <(docmgr list docs --root /path/to/ttmp) <(/tmp/docmgr-local list docs --root /path/to/ttmp)
```

**Needs scrutiny**:
- **Output shape preservation**: verify human and glaze output match old behavior (same fields, same formatting)
- **Filter semantics**: confirm ticket/status/doc-type/topics filters still work as expected
- **Diagnostics rendering**: in glaze mode, confirm broken docs emit diagnostics (old behavior)
- **Performance**: index build happens on every invocation; verify it's acceptably fast for real workspaces

**What was tricky**:
- Preserving the old `index.md` skip while `IncludeIndexMD` is not yet a QueryDocs option required a post-filter check.

---

## Step 14: Port `doc search` to `Workspace.QueryDocs` + preserve wonky path UX

**What was built**: `doc search` command now uses `Workspace.QueryDocs` for metadata + reverse lookup, keeping content search as a post-filter.

**Where it is**:
- **Files**: `pkg/commands/search.go`, `internal/workspace/query_docs_sql.go`
- **Key changes**:
  - Removed `filepath.Walk` + manual parsing + manual reverse lookup
  - Added `ws.QueryDocs` with `RelatedFile` / `RelatedDir` filters
  - Added basename suffix fallback matching (`LIKE '%/basename'`) for UX preservation
  - Content search, external-source, and date filters remain as post-filters

**How to exercise**:
```bash
# Metadata search
docmgr-local doc search --ticket MEN-3475 --topics backend --doc-type design

# Reverse lookup
docmgr-local doc search --file pkg/commands/search.go
docmgr-local doc search --file search.go  # basename-only (should still work)

# Directory reverse lookup
docmgr-local doc search --dir pkg/commands/

# Wonky paths (scenario suite regression cases)
docmgr-local doc search --file "../backend/chat/api/register.go"
docmgr-local doc search --file "/absolute/path/to/file.go"

# Content search (post-filter)
docmgr-local doc search "WebSocket"

# Combined
docmgr-local doc search "API" --topics backend --file internal/workspace/query_docs.go
```

**Needs scrutiny**:
- **Basename fallback matching**: verify `LIKE '%/name.go'` doesn't match too broadly (e.g., doesn't match `/myname.go`)
- **Content search correctness**: confirm `IncludeBody=true` populates `handle.Body` and content filter works
- **Date filters**: confirm `--since`, `--until`, `--created-since`, `--updated-since` still work as post-filters
- **External source**: confirm substring matching on `ExternalSources` still works

**What was tricky**:
- Basename-only reverse lookup required explicit suffix matching logic in the SQL compiler without making it too fuzzy.
- Preserving the old `--file register.go` UX (scenario suite expected it) while keeping the normalization pipeline principled.

---

## Key Implementation Files (Quick Reference)

### Core Infrastructure
- `internal/workspace/workspace.go` — Workspace entry point + context
- `internal/workspace/sqlite_schema.go` — DB open + schema DDL
- `internal/workspace/skip_policy.go` — Canonical skip/tagging rules

### Ingestion
- `internal/workspace/index_builder.go` — InitIndex + ingest walker
- `internal/workspace/normalization.go` — RelatedFiles normalization pipeline

### Query Engine
- `internal/workspace/query_docs.go` — QueryDocs API + hydration (no N+1)
- `internal/workspace/query_docs_sql.go` — DocQuery → SQL compiler

### Diagnostics
- `pkg/diagnostics/docmgrctx/query_docs.go` — QueryDocs taxonomy types

### Command Ports
- `pkg/commands/list_docs.go` — Ported to Workspace.QueryDocs
- `pkg/commands/search.go` — Ported to Workspace.QueryDocs

### Utilities
- `internal/workspace/sqlite_export.go` — Export to persistent file + README table

---

## Overall Review Checklist

### Schema & Indexes
- [ ] Tables match spec (docs, doc_topics, related_files)
- [ ] Indexes cover expected query patterns
- [ ] Foreign key constraints are correct (CASCADE behavior)
- [ ] Pragmas are appropriate for in-memory use

### Ingestion
- [ ] Skip rules match spec (skip `.meta/`, `_*/`; tag archive/scripts/control docs)
- [ ] Parse errors are handled gracefully (parse_ok=0 + parse_err)
- [ ] Path normalization is correct (per-doc resolver, multiple keys persisted)
- [ ] Transaction safety (all-or-nothing)

### Query Compilation
- [ ] Scope handling is correct (Repo/Ticket/Doc)
- [ ] Filter semantics match spec (exact match, OR for TopicsAny/RelatedFile/RelatedDir)
- [ ] Visibility defaults are correct (hide archived/scripts/control docs by default)
- [ ] SQL is safe (parameterized, no injection)
- [ ] Basename fallback matching is not too fuzzy

### Query Execution
- [ ] No N+1 (fixed number of queries regardless of result size)
- [ ] Hydration is correct (topics/related_files populated for parse-ok docs)
- [ ] Parse-error docs return with Doc=nil + ReadErr when IncludeErrors=true
- [ ] Ordering is correct (OrderByPath/OrderByLastUpdated + Reverse)

### Diagnostics
- [ ] Parse skips are emitted when IncludeDiagnostics=true
- [ ] Normalization fallback warnings are emitted
- [ ] Taxonomy shape is correct and renderable

### Command Ports
- [ ] list docs output matches old behavior (human + glaze)
- [ ] doc search output matches old behavior (snippets, matched files, etc.)
- [ ] Filters work as expected (ticket, topics, status, doc-type, file, dir)
- [ ] Content search still works (post-filter)
- [ ] Date filters still work (post-filter)
- [ ] External source filter still works (post-filter)

### Integration
- [ ] Scenario suite passes (`test-scenarios/testing-doc-manager/run-all.sh`)
- [ ] Unit tests pass (`go test ./...`)
- [ ] No regressions in existing commands not yet ported

---

## Related Documents

- **Implementation diary**: `reference/15-diary.md` (full narrative)
- **Design spec**: `design/01-workspace-sqlite-repository-api-design-spec.md` (architecture + decisions)
- **Senior code review guide**: `analysis/03-code-review-guide-senior.md` (spec-to-code mapping)
- **Testing strategy**: `analysis/02-testing-strategy-integration-first.md` (when/how we test)
- **Task list**: `tasks.md` (remaining work)
