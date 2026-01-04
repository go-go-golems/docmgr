---
Title: 'Analysis: FTS-backed search + refactor search packages'
Ticket: 005-USE-SQLITE-FTS
Status: active
Topics:
    - backend
    - docmgr
    - tooling
    - testing
DocType: analysis
Intent: ticket-specific
Owners: []
RelatedFiles:
    - Path: internal/workspace/query_docs_sql.go
      Note: FTS integration point in SQL compiler
    - Path: internal/workspace/sqlite_schema.go
      Note: Schema changes for docs_fts
    - Path: pkg/commands/search.go
      Note: Search semantics currently implemented here; refactor target
    - Path: ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/design-doc/01-design-fts-backed-search-engine-no-compatibility.md
      Note: Design doc superseding parity-first notes
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T14:10:05.593839121-05:00
WhatFor: ""
WhenToUse: ""
---



# Analysis: FTS-backed search + refactor search packages

## What this ticket is asking for

**Note:** This analysis was initially written with a “parity-first” mindset. The ticket’s current decision is **no backwards compatibility** for `--query` behavior; the authoritative plan is now in `design-doc/01-design-fts-backed-search-engine-no-compatibility.md`.

1. Update docmgr’s search implementation so `--query` can be executed using **SQLite FTS5** (instead of a Go substring scan over every candidate body).
2. Move “search logic” out of the CLI-oriented `pkg/commands` layer and into shared `internal/` / `pkg/` packages so:
   - the CLI becomes a thin adapter,
   - future HTTP APIs can reuse the same search engine,
   - search semantics don’t drift between surfaces.

This analysis is intentionally “where to look + what to refactor + what to change”, with concrete file/symbol pointers.

## Ground truth: where search behavior lives today

### CLI entry points

- Cobra wiring:
  - `cmd/docmgr/cmds/doc/search.go` (creates cobra command, dual-mode glue, completion)
  - `cmd/docmgr/cmds/root.go` (adds alias `docmgr search`)

### CLI implementation (the big one)

- `pkg/commands/search.go`
  - `type SearchCommand`
  - `type SearchSettings`
  - `(*SearchCommand).RunIntoGlazeProcessor(...)`
  - `(*SearchCommand).suggestFiles(...)`
  - helpers: `extractSnippet`, `parseDate`, git/ripgrep heuristics

This file currently mixes:

- request parsing and validation (flags → settings)
- workspace discovery and indexing
- index-backed filtering (`Workspace.QueryDocs`)
- post-filters (content substring, external source, date filters)
- output formatting (Glaze rows / bare mode)
- optional heuristics (`--files`)

### The “query engine” already exists (and is the right place to integrate FTS)

Doc filtering and reverse-lookup matching is already powered by an in-memory SQLite index:

- Workspace + root resolution:
  - `internal/workspace/workspace.go`: `DiscoverWorkspace`, `WorkspaceContext`, `Workspace`
  - `internal/workspace/config.go`: `ResolveRoot`, `.ttmp.yaml` discovery, repo root detection
- Index schema:
  - `internal/workspace/sqlite_schema.go`: `createWorkspaceSchema`, `openInMemorySQLite`
- Index ingestion:
  - `internal/workspace/index_builder.go`: `(*Workspace).InitIndex`, `ingestWorkspaceDocs`
- Query compilation + execution:
  - `internal/workspace/query_docs.go`: `(*Workspace).QueryDocs`, `DocQuery` types
  - `internal/workspace/query_docs_sql.go`: `compileDocQueryWithParseFilter(...)`

This is the “right layer” for FTS because:

- `--query` is semantically a filter over document text.
- SQLite is already the filtering backend for other criteria.

## The important current limitation: `--query` is not index-backed

Even though `SearchCommand` calls `ws.InitIndex(... IncludeBody: true)`, it does not use SQL for text matching.

Instead, it post-filters in Go:

- `pkg/commands/search.go`: it lowercases `content` (body) and does `strings.Contains(contentLower, queryLower)`.

Implications:

- performance depends on how many bodies are scanned post-query
- no ranking
- text query cannot participate in SQL query planning

## FTS prerequisites and realities in this repo

### Build tag: `sqlite_fts5` already exists as a pattern

Two relevant, proven patterns live in this repo:

1. **Scenariolog**: best-effort FTS table creation with graceful degraded mode:
   - `scenariolog/internal/scenariolog/migrate.go` (`ensureFTS5`)
2. **Glazed help store**: FTS virtual table + triggers behind a build tag:
   - `glazed/pkg/help/store/fts5.go` is `//go:build sqlite_fts5`
   - It emits the debug log line we see in docmgr output: `Created FTS5 tables and triggers`

This strongly suggests that in the current build of `docmgr`, the `sqlite_fts5` tag is already used (or at least it is supported in toolchains used for releases).

### Runtime detection is still necessary

Even if we build with `sqlite_fts5`, the safest approach is:

- attempt to create an FTS table and treat “no such module: fts5” as a supported degraded mode

Reasons:

- developers may build without the tag
- some environments may have CGO/toolchain differences
- we want predictable behavior rather than a hard crash

So: **use the scenariolog approach** (best-effort, degrade gracefully), even if we also offer build-tagged optimizations.

## Proposed FTS integration approach (parity-first)

### Goal: keep current CLI semantics unless explicitly changing them

The current user-observable semantics are:

- `--query` performs a case-insensitive substring match over markdown bodies
- ordering is primarily by path (unless caller changes options elsewhere)
- snippet is derived by `extractSnippet(content, query, 100)`

Introducing FTS can improve performance while keeping semantics stable by:

1. using FTS as the primary candidate filter (fast narrowing),
2. optionally *keeping* the substring post-filter to preserve “exactly what it did before”,
3. keeping ordering unchanged (don’t introduce ranking until a deliberate change),
4. keeping snippet extraction unchanged (for now).

This yields:

- speed improvement (often huge)
- minimal UX surprises

### Where to integrate: `internal/workspace` query compilation

The cleanest architecture is:

1. Extend `internal/workspace` query types to include a text query:
   - add `TextQuery string` to `internal/workspace/query_docs.go` → `type DocFilters`
2. Extend SQL compilation to use FTS when `TextQuery != ""`.

Concretely:

- Modify `internal/workspace/query_docs_sql.go`:
  - if `TextQuery` is non-empty:
    - join or filter against an FTS virtual table (see below)
    - add `MATCH ?` expression
    - add a fallback mode if FTS is not available

If this exists, `SearchCommand` can become:

- parse params → build `DocQuery` including `Filters.TextQuery`
- call `ws.QueryDocs(...)`
- only apply the remaining post-filters that are not indexed yet (external sources, date filters, etc.)

### Schema: add a docs FTS virtual table

Docmgr’s workspace index is in-memory and rebuilt from disk, so we have flexibility.

FTS schema options (trade-offs):

#### Option A (simple): contentless `docs_fts` and manual insertion

Create:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS docs_fts USING fts5(
  doc_id UNINDEXED,
  body,
  tokenize='unicode61'
);
```

During ingestion:

- when `opts.IncludeBody == true` and doc parses ok, insert into `docs_fts(doc_id, body)`.

Pros:
- straightforward
- no triggers
- no dependence on `content='docs'`

Cons:
- duplicate storage of body (`docs.body` + `docs_fts.body`) unless we stop storing `docs.body`

#### Option B (cleaner storage): `content='docs'` + `rebuild`

Create:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS docs_fts USING fts5(
  body,
  content='docs',
  content_rowid='doc_id',
  tokenize='unicode61'
);
```

Then after ingesting all docs (and storing `docs.body`), run:

```sql
INSERT INTO docs_fts(docs_fts) VALUES('rebuild');
```

Pros:
- avoids storing a second copy of body
- good fit for “rebuild per invocation”
- no triggers necessary for correctness (since DB is write-once)

Cons:
- requires storing `docs.body` for all indexed docs whenever you want query search
- rebuild step must run only when body exists

#### Option C (best UX, but semantic change): add more fields to FTS

`docmgr doc search --query` currently scans only body, but in practice users often expect title/topic hits too.

FTS can index multiple columns:

```sql
CREATE VIRTUAL TABLE docs_fts USING fts5(
  title,
  body,
  topics,
  doc_type,
  ticket_id,
  tokenize='unicode61'
);
```

But: this changes semantics (query will match more documents than body-only substring).

Recommendation:

- Start with body-only for strict parity, then consider a deliberate semantic expansion.

### Best-effort table creation (degraded mode)

Follow `scenariolog/internal/scenariolog/migrate.go` pattern:

- implement `ensureFTS5(ctx, db)` in `internal/workspace/sqlite_schema.go` (or a new file `fts.go`)
- it should:
  - run `CREATE VIRTUAL TABLE ... USING fts5(...)`
  - if it fails with “no such module: fts5”, treat as “FTS not available”

Then, in `Workspace.InitIndex`:

- create schema (core tables) as today
- call `ensureFTS5` outside the main DDL transaction (recommended; virtual table behavior differs across builds)
- store an “fts available” bit on the Workspace instance

### Query compilation: how to write FTS match SQL safely

If `docs_fts` exists, a common pattern is:

- join:
  - `JOIN docs_fts fts ON fts.rowid = d.doc_id`
- filter:
  - `fts MATCH ?`

But note:

- `MATCH` wants a query string in FTS syntax, not arbitrary text.
- We need to decide how to escape / interpret the user query.

Parity-first recommendation:

- treat the user string as a **phrase**:
  - for example, wrap in quotes, and escape quotes inside
  - or keep “simple token query” behavior and document it

This is a subtle but important design point:

- current substring search treats query as raw text
- FTS tokenizes; queries behave differently

Therefore, one reasonable parity strategy is:

1. Use FTS to prefilter broadly (token match).
2. Keep the substring check post-filter to preserve “contains” semantics (exact string).

This lets FTS be “fast index narrowing” rather than “semantic definition of matching”.

### Ranking and ordering: don’t change by default

FTS supports ranking via `bm25(docs_fts)`; it’s tempting to order by it.

But that changes visible output ordering.

Recommendation for first implementation:

- keep existing ordering (`OrderByPath` or `OrderByLastUpdated`)
- optionally add a new `OrderByRank` later (explicit flag and tests)

### Snippets: keep current snippet generation initially

FTS supports `snippet(docs_fts, ...)` highlighting, but:

- it introduces new formatting
- it may not be stable across sqlite versions

Recommendation:

- keep existing `extractSnippet` logic for now
- when `docs.body` is not stored, compute snippet by reading the markdown file for the matched docs only

## Refactor plan: move search logic out of `pkg/commands`

### What “out of commands” should mean here

`pkg/commands` is the CLI command implementation layer in this repo.

We want search semantics to be usable by:

- CLI (`docmgr doc search`)
- future HTTP API (ticket 004)
- tests (without invoking cobra/glaze)

So: `pkg/commands/search.go` should stop being the “engine”.

### Proposed new package boundaries

Keep the query engine in `internal/workspace` (already correct).

Move “search semantics” (glue + post-filters) into a package that is not tied to Glaze:

Option 1 (internal-only):

- `internal/searchsvc` (or `internal/search`)
  - `type Params` similar to `SearchSettings` but without CLI tags
  - `func SearchDocs(ctx context.Context, ws *workspace.Workspace, p Params) (Result, error)`
  - `func SuggestFiles(ctx, ws, p) ([]Suggestion, error)` (optional)

Option 2 (exported pkg):

- `pkg/search` with exported types that can be imported by other modules.

Given “search is an internal behavior” and we don’t have external consumers yet, Option 1 is safer initially.

### What stays in `pkg/commands/search.go`

Only:

- flag definitions (`SearchSettings`)
- parsing/validation messages for CLI
- wiring to Glaze output (row creation)
- calling into `internal/searchsvc`

### What moves out of `pkg/commands/search.go`

- Query building (DocQuery composition)
- Post-filters:
  - date parsing (`parseDate`)
  - external source matching (frontmatter read)
  - snippet extraction (`extractSnippet`)
- File suggestion heuristics (`--files` mode):
  - git log / git status / rg fallback logic

This dramatically reduces the size of `pkg/commands/search.go` and makes it less likely the HTTP API duplicates behavior.

### Bonus refactor: remove duplicated logic between glazed mode and bare mode

`SearchCommand` currently has both:

- `RunIntoGlazeProcessor` (glazed)
- `Run` (bare)

They are largely duplicated in the file.

Once the search engine is extracted, both can call the same engine and simply render results differently.

## Files and symbols likely to change (implementation checklist)

### Workspace / index

- `internal/workspace/sqlite_schema.go`
  - add FTS table creation (`ensureFTS5` style)
  - optionally keep it behind build tags or runtime best-effort
- `internal/workspace/index_builder.go`
  - if using content='docs': ensure body is stored when needed + run FTS rebuild step
  - if using contentless table: insert into FTS during ingest
- `internal/workspace/query_docs.go`
  - extend `DocFilters` with `TextQuery`
  - maybe extend `DocQueryOptions` with `Limit/Offset` later (helpful for APIs)
- `internal/workspace/query_docs_sql.go`
  - add FTS `JOIN ... MATCH ?` when TextQuery present and FTS available
  - decide how to escape/interpret query text

### Search semantics refactor

- new package: `internal/searchsvc/*`
  - main search function (calls QueryDocs)
  - snippet logic
  - date parsing
  - external source filter strategy
  - file suggestion heuristics (optional)
- `pkg/commands/search.go`
  - becomes thin wrapper calling `internal/searchsvc`
  - emit glaze rows from returned result structs

### Documentation updates (once implemented)

- Update:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`
    - “content search is substring scan” becomes “FTS-backed (with substring parity filter)” or similar

## Testing strategy (what to add/where)

Testing should focus on:

1. **Correctness parity**: the result set for `--query` should match the old substring semantics (at least for simple cases).
2. **FTS availability behavior**: if FTS is missing, search should fall back (not crash).
3. **Reverse lookup unchanged**: `--file` and `--dir` behavior must not drift.

Concrete patterns already exist:

- Build-tagged FTS tests:
  - `scenariolog/internal/scenariolog/search_fts5_test.go` (`//go:build sqlite_fts5`)

Docmgr could add:

- `internal/workspace/search_fts5_test.go` with `//go:build sqlite_fts5`
  - build an in-memory workspace index from a small fixture directory and assert MATCH results

Additionally:

- scenario tests under `test-scenarios/` (if existing patterns support it)
  - compare outputs before/after for a known fixture set

## Related docs to read (do not reinvent)

- Search internals deep dive (includes an “FTS future evolution” section):
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`
- Workspace/query architecture:
  - `pkg/doc/docmgr-codebase-architecture.md`

## Recommended next step (after this analysis)

Write a design doc (or implementation plan) that makes two explicit choices:

1. Which FTS schema option (A/B/C) to use initially (strong default: **B: content='docs' + rebuild**, parity-first).
2. Whether to keep substring post-filter for parity (strong default: **yes**, at least for v1).
