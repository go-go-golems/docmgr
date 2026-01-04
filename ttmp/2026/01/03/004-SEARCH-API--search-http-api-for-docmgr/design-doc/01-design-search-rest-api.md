---
Title: 'Design: Search REST API'
Ticket: 004-SEARCH-API
Status: draft
Topics:
    - backend
    - docmgr
    - tooling
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/query_docs.go
      Note: Query types to expose through REST
    - Path: internal/workspace/query_docs_sql.go
      Note: Filter compilation and ordering
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T21:33:33.020911519-05:00
WhatFor: ""
WhenToUse: ""
---


# Design: Search REST API

## Executive Summary

Add a local HTTP server mode to `docmgr` that exposes a versioned JSON REST API for searching docs using the same workspace/index/query engine as the CLI (`workspace.DiscoverWorkspace` → `InitIndex` → `QueryDocs` + the existing search post-filters).

This enables a web UI (and other tools) to call `docmgr` search without shelling out to the CLI and without re-implementing reverse lookup/path normalization.

## Problem Statement

`docmgr` has a powerful search command (`docmgr doc search`), but it’s CLI-only. A web UI or IDE integration needs:

- an always-on process (no per-request `go run` / binary spawn),
- stable JSON contracts (not Glaze rows and not human text),
- an index that is built once and reused across requests,
- explicit endpoints for:
  - searching docs (content + metadata + reverse lookup),
  - (optionally) suggesting related files (`--files` mode),
  - rebuilding the index when docs change.

We must preserve existing search semantics by default to avoid UI/CLI drift.

## Proposed Solution

### Overview

Implement:

1. A new Cobra command to run the server (proposed name: `docmgr api serve`, final bikeshed TBD).
2. A small internal HTTP package that:
   - owns workspace discovery and index lifecycle (build once; rebuild on demand),
   - exposes REST endpoints for search,
   - serializes results to JSON.
3. A shared “search engine” function/service so both CLI and HTTP can call the same logic (to prevent divergence).

### Server command (CLI)

#### Proposed CLI surface

Example usage:

```bash
docmgr api serve --addr 127.0.0.1:8787 --root ttmp
```

Flags:

- `--addr` (default `127.0.0.1:8787`)
- `--root` (default `ttmp`, resolved via `workspace.ResolveRoot`)
- `--log-level` (optional; follow project conventions if/when added)
- `--cors-origin` (optional; if web UI is a browser app served from a different origin)

Default binding should be localhost-only.

### Index lifecycle

At startup:

1. Discover workspace with `workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: root})`.
2. Build index with `ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: true})` (body needed for `query` substring scan).

During runtime:

- Search requests query the current index.
- A rebuild endpoint triggers `InitIndex` again (swap DB safely).

Concurrency:

- Protect `QueryDocs` vs `InitIndex` with an `RWMutex`.

### REST API (v1)

#### Conventions

- Base path: `/api/v1`
- JSON only.
- All endpoints return:
  - success: `2xx` with a JSON object
  - failure: non-`2xx` with a JSON error payload

Error payload shape:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "must provide at least a query or filter",
    "details": {
      "field": "query"
    }
  }
}
```

#### Endpoint: Health

`GET /api/v1/healthz`

Response:

```json
{ "ok": true }
```

#### Endpoint: Workspace status (optional but useful)

`GET /api/v1/workspace/status`

Purpose:
- Let a UI confirm which root/config/vocabulary are in use and whether an index is built.

Response (proposed):

```json
{
  "root": "/abs/path/to/ttmp",
  "configPath": "/abs/path/to/.ttmp.yaml",
  "vocabularyPath": "/abs/path/to/ttmp/vocabulary.yaml",
  "indexedAt": "2026-01-03T21:40:00Z",
  "docsIndexed": 200
}
```

#### Endpoint: Rebuild index

`POST /api/v1/index/rebuild`

Semantics:
- Rebuild the in-memory index from disk.
- Returns metadata about the rebuild.

Response:

```json
{
  "rebuilt": true,
  "indexedAt": "2026-01-03T21:40:00Z",
  "docsIndexed": 200
}
```

#### Endpoint: Search docs

`GET /api/v1/search/docs`

Query parameters (mirrors `docmgr doc search` flags as closely as possible):

- `query` (string): content substring query
- `ticket` (string): ticket filter; if set, server also scopes to that ticket
- `topics` (string): comma-separated list; OR semantics (“any topic”)
- `docType` (string)
- `status` (string)
- `file` (string): reverse lookup by related file
- `dir` (string): reverse lookup by related dir
- `externalSource` (string): substring match against `ExternalSources` (post-filter)
- `since` (string): compare against `LastUpdated`
- `until` (string): compare against `LastUpdated`
- `createdSince` (string): compare against `os.Stat(...).ModTime()` (CLI parity)
- `updatedSince` (string): compare against `LastUpdated`

Visibility toggles (match CLI default behavior):

- `includeArchived` (bool, default `true`)
- `includeScripts` (bool, default `true`)
- `includeControlDocs` (bool, default `true`)
- `includeDiagnostics` (bool, default `true`)
- `includeErrors` (bool, default `false`)

Sorting:

- `orderBy` = `path` | `last_updated` (default `path`)
- `reverse` (bool, default `false`)

Pagination (explicit design choice; recommended to implement in v1):

- `limit` (int, default `200`, max `1000`)
- `offset` (int, default `0`)

Request validation (match CLI behavior):

- If all filters are empty (`query`, `ticket`, `topics`, `docType`, `status`, `file`, `dir`, `externalSource`, `since/until/created/updated`), return `400 invalid_argument`.

Response shape:

```json
{
  "query": {
    "query": "Workspace.QueryDocs",
    "ticket": "",
    "topics": ["backend"],
    "docType": "",
    "status": "",
    "file": "",
    "dir": "",
    "externalSource": "",
    "since": "",
    "until": "",
    "createdSince": "",
    "updatedSince": "",
    "orderBy": "path",
    "reverse": false,
    "limit": 200,
    "offset": 0
  },
  "total": 12,
  "results": [
    {
      "ticket": "MEN-1234",
      "title": "Doc Search: Implementation and API Guide",
      "docType": "reference",
      "status": "active",
      "topics": ["docmgr", "ux"],
      "path": "2026/01/03/001-ADD-DOCMGR-UI--.../reference/02-doc-search-implementation-and-api-guide.md",
      "snippet": "..."
    }
  ],
  "diagnostics": []
}
```

Notes:

- `path` should be docs-root relative (to match CLI output and allow deterministic linking).
- `snippet` should reuse the CLI snippet rules for parity.
- For `file` searches, include additional fields so the UI can show which related file entries matched:
  - `matchedFiles` (string list)
  - `matchedNotes` (string list)

#### Endpoint: Suggest related files (optional; maps to `doc search --files`)

`GET /api/v1/search/files`

Query parameters (subset):

- `query` (string): seed term for heuristics (optional)
- `ticket` (string): scope suggestions to a ticket (optional)
- `topics` (string): comma-separated
- `limit` (int, default `200`, max `1000`)

Response:

```json
{
  "total": 42,
  "results": [
    { "file": "pkg/commands/search.go", "source": "related_files", "reason": "Referenced by docs in ticket ..." }
  ]
}
```

### Implementation structure (proposed packages)

The goal is to keep HTTP concerns separate from doc indexing/search semantics.

Proposed files/packages (subject to repo conventions):

- `cmd/docmgr/cmds/api/serve.go`
  - Cobra command wiring
- `pkg/commands/api_serve.go` (optional)
  - If following the “commands live in pkg/commands” pattern, but HTTP may be simpler to keep out of Glaze/dual-mode.
- `internal/httpapi/server.go`
  - `type Server struct { ... }`
  - routing, JSON helpers, error formatting
- `internal/httpapi/index_manager.go`
  - `type IndexManager struct { ws *workspace.Workspace; mu sync.RWMutex; indexedAt time.Time; ... }`
- `internal/searchsvc/search.go`
  - `SearchDocs(ctx, ws, params) (results, diagnostics, error)`
  - `SuggestFiles(...)` (if included)

### Compatibility goals

- **Parity-first**: v1 REST search should match `docmgr doc search` semantics, especially for reverse lookup.
- **Versioned**: breaking changes require `/api/v2` (or explicit feature flags).
- **Local-first**: bind localhost by default.

## Design Decisions

1. **Stdlib `net/http` over a router dependency**
   - Rationale: there is currently no HTTP server stack in this module; keep dependencies minimal unless a UI requires routing features.
2. **Index built once, rebuilt explicitly**
   - Rationale: avoids per-request rebuild cost and avoids complex file watcher correctness initially.
3. **Reuse workspace/query engine**
   - Rationale: prevents semantic drift; reverse lookup/path normalization is already battle-tested in the query engine.
4. **Keep substring search semantics for `query` in v1**
   - Rationale: matches CLI behavior; switching to FTS is a deliberate future change.
5. **Add pagination to the API even if CLI doesn’t have it**
   - Rationale: UIs need incremental loading; we can implement server-side slicing now and later optimize with SQL LIMIT/OFFSET.

## Alternatives Considered

1. **Shell out to `docmgr doc search --with-glaze-output --output json` from the server**
   - Rejected: slow, hard to manage concurrency, hard to paginate, couples to CLI output quirks.
2. **Expose Glaze rows directly as the API**
   - Rejected: Glaze row schema is optimized for CLI table/csv use; it’s not an explicit long-lived API contract.
3. **Implement a new search engine specifically for HTTP**
   - Rejected: would duplicate reverse lookup + filtering semantics and drift from CLI.
4. **Add FTS5 immediately and change `query` semantics**
   - Rejected for v1: semantics change + extra complexity. Keep as a follow-up after parity version ships.

## Implementation Plan

1. **Create server command**
   - Add `cmd/docmgr/cmds/api/serve.go` and attach it in `cmd/docmgr/cmds/root.go`.
2. **Build minimal HTTP server**
   - Implement `/api/v1/healthz`.
   - Implement JSON error helper.
3. **Implement index manager**
   - Discover workspace once at startup.
   - Build index once.
   - Guard rebuild/query with `RWMutex`.
4. **Implement `/api/v1/search/docs`**
   - Parse query params into a struct mirroring CLI `SearchSettings`.
   - Execute `QueryDocs` and apply post-filters matching `pkg/commands/search.go`.
   - Return JSON response with `results`, `total`, `diagnostics`.
5. **(Optional) Implement `/api/v1/index/rebuild`**
6. **(Optional) Implement `/api/v1/search/files`**
7. **Refactor for reuse**
   - Extract snippet + post-filter logic into a shared internal package so CLI and HTTP call the same functions.
8. **Add tests**
   - At minimum: request parsing + smoke search against a small fixture workspace.
   - If adding pagination in SQL: unit tests for `compileDocQueryWithParseFilter` output.

## Open Questions

1. Should the server support per-request `root` overrides, or is root strictly a startup-time choice?
2. Do we want an endpoint to fetch a document’s full body/frontmatter (`GET /api/v1/docs/...`) for UI navigation, or is that out of scope for “search API”?
3. Should `query` remain substring-only forever (parity), or do we plan a `/api/v2` that adds FTS and ranking?
4. Should search results include `lastUpdated` explicitly (not currently in CLI row output) to support UI sorting?

## References

- Analysis: `docmgr/ttmp/2026/01/03/004-SEARCH-API--search-http-api-for-docmgr/analysis/01-analysis-search-http-server-rest-api.md`
- Search implementation deep-dive: `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`
- Workspace/query architecture: `docmgr/pkg/doc/docmgr-codebase-architecture.md`
