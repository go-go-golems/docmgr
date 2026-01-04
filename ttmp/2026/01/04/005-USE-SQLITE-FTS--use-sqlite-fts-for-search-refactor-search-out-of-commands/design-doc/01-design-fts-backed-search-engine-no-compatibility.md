---
Title: 'Design: FTS-backed search engine (no compatibility)'
Ticket: 005-USE-SQLITE-FTS
Status: draft
Topics:
    - backend
    - docmgr
    - tooling
    - testing
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/searchsvc/search.go
      Note: Engine boundary the design targets
    - Path: internal/workspace/sqlite_schema.go
      Note: docs_fts schema and FTS availability
    - Path: pkg/commands/search.go
      Note: Thin CLI wrapper
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T15:34:47.928865577-05:00
WhatFor: ""
WhenToUse: ""
---


# Design: FTS-backed search engine (no compatibility)

## Executive Summary

Replace docmgr’s `--query` search implementation with a **SQLite FTS5-backed** query path (no backwards compatibility guarantees) and refactor the CLI search command so it becomes a thin adapter over a reusable search engine package.

This design explicitly chooses:

- **FTS schema option C**: index `title`, `body`, `topics`, `doc_type`, `ticket_id`.
- **Ranking support**: add `OrderByRank` using `bm25(...)`.
- **Snippet**: keep the current `extractSnippet` behavior for now, but move it into the reusable core package.

## Problem Statement

Today `docmgr doc search --query` is implemented as a Go substring scan over bodies after `Workspace.QueryDocs(...)` returns candidates. This has three major downsides:

- slow on large workspaces (body scanning is O(total bytes scanned))
- no ranking (results are typically ordered by path)
- search logic is embedded in `pkg/commands/search.go`, which makes it difficult to reuse for HTTP APIs and increases risk of semantic drift

We do **not** need behavior exact-match compatibility with the current substring semantics. This ticket is allowed to change behavior.

## Proposed Solution

### 1) Add FTS to the in-memory workspace index

Extend the workspace SQLite schema to create a virtual FTS5 table:

- Table name: `docs_fts`
- Model: **contentless** FTS table, populated during ingest
- Tokenizer: `unicode61`

Schema (Option C):

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS docs_fts USING fts5(
  title,
  body,
  topics,
  doc_type,
  ticket_id,
  tokenize='unicode61'
);
```

Populate it by inserting rows with `rowid = docs.doc_id` during ingestion:

```sql
INSERT INTO docs_fts(rowid, title, body, topics, doc_type, ticket_id)
VALUES (?, ?, ?, ?, ?, ?);
```

Where:

- `title/body/doc_type/ticket_id` come from the parsed frontmatter and body.
- `topics` is a deterministic string representation derived from `doc.Topics` (for example, comma-separated in original order).

FTS availability:

- Best-effort create `docs_fts`.
- If sqlite lacks fts5 (`no such module: fts5`), treat it as “FTS unavailable”:
  - metadata-only search stays available
  - text query (`--query`) returns a clear error

### 2) Teach the query engine about text queries and rank ordering

Extend `internal/workspace` query types:

- Add `TextQuery string` to `DocFilters`.
- Add `OrderByRank` to `OrderBy`.

Update SQL compilation:

- If `TextQuery` is non-empty, join FTS and add `WHERE docs_fts MATCH ?`.
- If `OrderByRank`, order by `bm25(docs_fts)` (ascending = better).

Notes:

- Because we don’t need compatibility, we treat the raw user query string as an FTS query string. If parsing fails, return a useful error.

### 3) Extract a reusable search engine package

Create a package that is not tied to Glaze/Cobra:

- Suggested: `internal/searchsvc`

Core types:

- `type SearchQuery struct { ... }` (covers all search inputs, including visibility toggles, ordering, pagination later)
- `type SearchResult struct { ... }`
- `type SearchResponse struct { Total int; Results []SearchResult; Diagnostics []... }`

Core entrypoint:

- `SearchDocs(ctx context.Context, ws *workspace.Workspace, q SearchQuery) (SearchResponse, error)`

Responsibilities:

- Convert `SearchQuery` → `workspace.DocQuery`
- Call `ws.QueryDocs(...)`
- Apply remaining non-indexed filters (external sources, date filters) if still desired
- Generate snippets using `extractSnippet` moved to `internal/searchsvc/snippet.go`

### 4) Make the CLI command thin

Refactor `pkg/commands/search.go` to:

- parse flags into the existing CLI-only struct
- map into `searchsvc.SearchQuery`
- call `searchsvc.SearchDocs`
- render results (glaze rows / bare text) without re-implementing search semantics

## Design Decisions

1. **No backwards compatibility**
   - We do not preserve substring semantics. `--query` becomes FTS semantics.
2. **Option C FTS schema**
   - Indexing multiple fields is more useful than body-only and enables better UX immediately.
3. **`OrderByRank` with `bm25`**
   - Enables meaningful result ordering when using `--query`.
4. **Keep snippet logic (for now)**
   - Moving it to core prevents CLI/HTTP duplication; improving snippets can be a follow-up.
5. **Search engine outside `pkg/commands`**
   - Prevents semantic drift and makes HTTP API work straightforward.

## Alternatives Considered

1. Body-only FTS (schema option B/A)
   - Rejected: less useful than indexing title/topics/doc_type/ticket_id.
2. Preserve substring semantics by post-filtering after FTS
   - Rejected: explicit requirement is no exact-match compatibility.
3. Implement FTS in `pkg/commands` only
   - Rejected: makes it harder to reuse for HTTP and will drift.

## Implementation Plan

1. Add tasks to the ticket (see `tasks.md`).
2. Implement `docs_fts` creation and ingest population.
3. Extend `workspace.DocFilters` with `TextQuery`; extend ordering with `OrderByRank`.
4. Add SQL compilation logic for `MATCH` and `bm25`.
5. Create `internal/searchsvc` and move:
   - `extractSnippet`
   - date parsing (if still needed)
   - external source filtering (if still needed)
6. Refactor `pkg/commands/search.go` to call `internal/searchsvc`.
7. Add tests:
   - `//go:build sqlite_fts5` test proving `MATCH` returns expected hits and `OrderByRank` ordering works.
8. Update related docs (search implementation guide).

## Open Questions

1. Should we add a `--query-mode` (raw FTS vs quoted phrase) or just accept raw FTS syntax?
2. Should the FTS query always scope by ticket/doc-type before matching (for speed), or rely on SQLite planner?

## References

- Analysis: `ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/analysis/01-analysis-fts-backed-search-refactor-search-packages.md`
- Current CLI search command: `pkg/commands/search.go`
- Workspace schema and query engine:
  - `internal/workspace/sqlite_schema.go`
  - `internal/workspace/index_builder.go`
  - `internal/workspace/query_docs.go`
  - `internal/workspace/query_docs_sql.go`
