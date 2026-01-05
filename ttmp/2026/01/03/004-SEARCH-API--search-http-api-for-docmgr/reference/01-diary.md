---
Title: Diary
Ticket: 004-SEARCH-API
Status: active
Topics:
    - backend
    - docmgr
    - tooling
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T21:32:02.555262832-05:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Capture a frequent, tiny-step diary of analysis + design work for the `004-SEARCH-API` ticket, with enough concrete commands and file pointers that someone else can continue without re-deriving context.

## Step 1: Bootstrap ticket + seed docs

This step created the ticket workspace and the initial documents needed to capture analysis and design work. It also re-enabled the `analysis` doc-type in the workspace vocabulary so the ticket can contain a proper analysis document rather than overloading `reference` or `design-doc`.

**Commit (code):** N/A

### What I did
- Read the workflow refs from `~/.cursor/commands/`: `docmgr.md`, `diary.md`, `git-commit-instructions.md`.
- Created the ticket workspace:
  - `docmgr ticket create-ticket --ticket 004-SEARCH-API --title "Search HTTP API for docmgr" --topics backend,docmgr,tooling`
- Created seed docs:
  - `docmgr doc add --ticket 004-SEARCH-API --doc-type reference --title "Diary"`
  - Added `analysis` back into `docmgr/ttmp/vocabulary.yaml`
  - `docmgr doc add --ticket 004-SEARCH-API --doc-type analysis --title "Analysis: Search HTTP server + REST API"`
  - `docmgr doc add --ticket 004-SEARCH-API --doc-type design-doc --title "Design: Search REST API"`

### Why
- The ticket needs two artifacts:
  - a verbose “what/where/why” analysis document, and
  - a concrete REST API design document to implement next.
- `analysis` docs already exist historically in this repo (`ttmp/.../analysis/...`), so restoring the doc-type in `vocabulary.yaml` makes the workflow consistent again.

### What worked
- `docmgr` accepted the updated vocabulary and created an `analysis/` directory inside the ticket workspace.

### What didn't work
- N/A

### What I learned
- This repo’s current `ttmp/vocabulary.yaml` was missing the `analysis` doc-type, even though older tickets already use `analysis/` directories; `docmgr doc add` enforces vocabulary membership.

### What was tricky to build
- N/A (no code changes)

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- N/A

### Code review instructions
- N/A

### Technical details
- Vocabulary change: `docmgr/ttmp/vocabulary.yaml`
- Ticket docs:
  - `docmgr/ttmp/2026/01/03/004-SEARCH-API--search-http-api-for-docmgr/analysis/01-analysis-search-http-server-rest-api.md`
  - `docmgr/ttmp/2026/01/03/004-SEARCH-API--search-http-api-for-docmgr/design-doc/01-design-search-rest-api.md`

## Step 2: Survey current search + workspace index/query engine

This step mapped the existing “search” implementation onto the internal workspace/query architecture so the forthcoming HTTP API can reuse the same primitives (or refactor them cleanly). The key result is that `docmgr doc search` is a hybrid: metadata + reverse-lookup filtering is SQLite-backed, but content search is currently a Go substring scan over bodies (not FTS).

**Commit (code):** N/A

### What I did
- Grepped for existing search + HTTP server code:
  - `rg -n "SearchCommand|QueryDocs\\(" cmd pkg internal`
  - `rg -n "net/http|ListenAndServe" cmd pkg internal`
- Read core implementation files:
  - `pkg/commands/search.go` (`SearchCommand`, `SearchSettings`, post-filters)
  - `internal/workspace/*` (`Workspace`, `InitIndex`, `QueryDocs`, SQL compilation)
- Read the already-written deep guide on search internals:
  - `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`

### Why
- A REST API should not re-implement doc search semantics; it should reuse the workspace/query engine so CLI and API stay consistent.

### What worked
- Clear seams exist:
  - `workspace.DiscoverWorkspace` + `ws.InitIndex(...)` + `ws.QueryDocs(...)` give a reusable “query engine”.
  - `SearchCommand` adds post-filters (content substring, external sources, date filters) and output formatting.

### What didn't work
- No existing HTTP server implementation exists in this repo’s `docmgr` module; this needs to be added.

### What I learned
- The current in-memory SQLite schema (`internal/workspace/sqlite_schema.go`) does not include an FTS table; “full-text” search is best-effort post-filtering in `pkg/commands/search.go`.

### What was tricky to build
- N/A (analysis only)

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- As part of implementing the HTTP API, consider extracting a shared “search service” that both the CLI (`SearchCommand`) and server handlers call, to avoid semantic drift.

### Code review instructions
- Start with `pkg/commands/search.go`, then follow the calls into:
  - `internal/workspace/workspace.go`
  - `internal/workspace/index_builder.go`
  - `internal/workspace/query_docs.go`
  - `internal/workspace/query_docs_sql.go`

## Step 3: Write the verbose analysis doc (what to build, where to look, refactors)

This step consolidated the code survey into a structured “map”: what an HTTP server needs (root resolution, index lifecycle, concurrency), what the existing search engine already provides, and which refactors would prevent HTTP/CLI semantic drift. It also explicitly calls out decisions the design must make (pagination, category visibility defaults, query semantics).

**Commit (code):** N/A

### What I did
- Read root-resolution and repo-root logic to understand how a server should interpret `--root`:
  - `internal/workspace/config.go` (`ResolveRoot`, `.ttmp.yaml` discovery, `FindRepositoryRoot`)
- Wrote: `analysis/01-analysis-search-http-server-rest-api.md`

### Why
- The server design needs to be grounded in the existing query engine (Workspace + QueryDocs) instead of reinventing search in HTTP handlers.

### What worked
- The analysis doc cleanly separates:
  - “workspace + index lifecycle” concerns, and
  - “search REST API” concerns (semantics + parity).

### What didn't work
- N/A

### What I learned
- Ticket-scaffold detection uses a hardcoded directory marker list (`internal/workspace/discovery.go`) that does not currently include `design-doc/`; this may impact any server endpoint that tries to emulate doctor/status workspace detection.

### What was tricky to build
- N/A (docs only)

### What warrants a second pair of eyes
- Review the “potential refactors” list in the analysis doc and confirm which are in-scope for the first implementation iteration vs deliberate follow-ups.

### What should be done in the future
- If the server needs pagination, consider pushing LIMIT/OFFSET into `Workspace.QueryDocs` rather than slicing in handlers.

### Code review instructions
- Start with `docmgr/ttmp/2026/01/03/004-SEARCH-API--search-http-api-for-docmgr/analysis/01-analysis-search-http-server-rest-api.md`

## Step 4: Draft the REST API design document (v1 endpoints + contracts)

This step translated the analysis into an implementable REST API design: endpoints, query parameters mapped from CLI flags, response/error schemas, and an implementation plan with proposed packages and concurrency strategy.

**Commit (code):** N/A

### What I did
- Wrote: `design-doc/01-design-search-rest-api.md` (versioned `/api/v1` plan)

### Why
- A design doc gives a stable target for the next phase (implementation) and makes API decisions explicit before code is written.

### What worked
- The design keeps parity-first behavior by default (substring `query`, reuse `QueryDocs` for reverse lookup).

### What didn't work
- N/A

### What I learned
- The most important “don’t drift” lever is to extract shared search logic (post-filters + snippet) so both CLI and HTTP can call the same functions.

### What was tricky to build
- Choosing defaults that match CLI search (include archived/scripts/control docs) even though `QueryDocs` defaults hide them unless explicitly enabled.

### What warrants a second pair of eyes
- Endpoint naming + whether the server should expose doc bodies (`GET /api/v1/docs/...`) as part of “search API” scope.

### What should be done in the future
- If/when adding real full-text (FTS), treat it as a versioned semantic change (`/api/v2`) or an explicit query mode parameter.

## Step 5: Update design for FTS + cursor pagination + IndexManager

The codebase now has a shared query engine (`internal/searchsvc`) and FTS-backed `--query`, so the REST API design was updated to match the new reality:

- `query` is a SQLite FTS5 `MATCH` query string (no substring compatibility guarantees)
- add `orderBy=rank`
- use cursor-based pagination
- build index on startup and refresh explicitly via an `IndexManager`

**Commit (code):** N/A

### What I did
- Updated: `design-doc/01-design-search-rest-api.md`
- Updated: `tasks.md` to reflect implementation work

### Why
- The API must track the authoritative search behavior (shared engine + FTS) to avoid UI/CLI drift.

### What worked
- Mapping REST parameters directly to `internal/searchsvc.SearchQuery` keeps semantics aligned.

### What didn't work
- N/A

## Step 6: Start implementing the HTTP server and REST API

Implemented a first cut of the HTTP server and endpoints, centered around an `IndexManager` that builds once and refreshes explicitly.

**Commit (code):** pending

### What I did
- Added: `cmd/docmgr/cmds/api` (`docmgr api serve`)
- Added: `internal/httpapi` (`IndexManager`, server, handlers)
- Implemented endpoints: `/api/v1/healthz`, `/api/v1/workspace/status`, `/api/v1/index/refresh`, `/api/v1/search/docs`, `/api/v1/search/files`
- Added minimal tests for cursor + index-not-ready

### What worked
- Reusing `internal/searchsvc.SearchDocs` means the server is thin and inherits the same filters/snippets/diagnostics behavior as the CLI.

### What should be done next
- Add an end-to-end httptest that builds an index from a small fixture workspace and exercises `/search/docs`.
