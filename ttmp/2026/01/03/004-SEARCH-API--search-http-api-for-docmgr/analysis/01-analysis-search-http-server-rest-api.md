---
Title: 'Analysis: Search HTTP server + REST API'
Ticket: 004-SEARCH-API
Status: active
Topics:
    - backend
    - docmgr
    - tooling
DocType: analysis
Intent: ticket-specific
Owners: []
RelatedFiles:
    - Path: internal/workspace/config.go
      Note: Root resolution and vocabulary path logic
    - Path: internal/workspace/workspace.go
      Note: Workspace discovery used by any server
    - Path: pkg/commands/search.go
      Note: Source of current search semantics
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T21:33:28.202978527-05:00
WhatFor: ""
WhenToUse: ""
---


# Analysis: Search HTTP server + REST API

## Why this document exists

This ticket is about adding an HTTP server surface to `docmgr` so another tool (a web UI, scripts, IDE integration, etc.) can query the same doc index/search capabilities that the CLI exposes.

This analysis focuses on:

1. **Running an HTTP server against a docmgr ticket workspace** (workspace discovery, root resolution, index lifecycle, caching, concurrency, safety defaults).
2. **Building a search REST API for docmgr search** (how to map the existing CLI search semantics to JSON endpoints without forking behavior).

This is written as a “where to look + what to refactor” map, with file + symbol pointers and explicit design constraints.

## Context: what docmgr search is today

### Key architectural fact: there is already a reusable query engine

The `docmgr` CLI already has a “backend” architecture that is very close to what an HTTP server would want:

- `workspace.DiscoverWorkspace(...)` resolves root + config + repo root (`internal/workspace/workspace.go`, `internal/workspace/config.go`).
- `ws.InitIndex(...)` builds an **in-memory SQLite index** each invocation (`internal/workspace/index_builder.go`).
- `ws.QueryDocs(...)` queries that index with **scope + filters + options** (`internal/workspace/query_docs.go`, `internal/workspace/query_docs_sql.go`).

This is explicitly documented in:

- `docmgr/pkg/doc/docmgr-codebase-architecture.md`
- `docmgr/pkg/doc/docmgr-cli-guide.md`
- `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md` (excellent deep dive; treat it as the primary reference for search behavior)

### Key pragmatic fact: “content search” is not in SQLite (yet)

The search CLI (`docmgr doc search`) uses the SQLite index primarily for **metadata filtering** and **reverse lookup**:

- ticket (`--ticket`)
- topics (`--topics`)
- doc-type (`--doc-type`)
- status (`--status`)
- reverse lookup (`--file`, `--dir`)

But the “full text” aspect of `--query` is currently a **Go substring scan** over bodies:

- `pkg/commands/search.go`: lowercases body and checks `strings.Contains(...)`.

So, when designing an HTTP API, the first decision is whether to:

- mirror current semantics (substring scan), or
- implement a real FTS path and accept a semantic change.

This analysis assumes “**match current CLI semantics first**” unless explicitly choosing a behavior change (see “Potential refactors”).

## The two big concepts requested

## 1) Running an HTTP server to work against a docmgr ticket workspace

### What “work against a ticket workspace” means in code

The workspace is the **documentation root**, typically `ttmp/`, that contains ticket directories and their docs.

The authoritative “how do I find the root?” behavior lives here:

- `internal/workspace/config.go`
  - `ResolveRoot(root string) string`
  - `FindTTMPConfigPath() (string, error)` (discovers `.ttmp.yaml`)
  - `FindRepositoryRoot() (string, error)`
  - `ResolveVocabularyPath() (string, error)` (finds `vocabulary.yaml`)
- `internal/workspace/workspace.go`
  - `DiscoverWorkspace(ctx, opts)` (builds `WorkspaceContext`)

Important semantics to preserve:

- If the caller passes a non-default `--root`, `ResolveRoot` treats it as explicit and anchors relative paths on CWD.
- If the caller passes default `ttmp`, `ResolveRoot` is allowed to apply `.ttmp.yaml` and git-root fallbacks.

For an HTTP server, this naturally maps to:

- a startup-time `--root` flag (or env var),
- plus *optional* per-request override (very risky unless strictly constrained).

### Index lifecycle: CLI rebuilds per invocation; a server must not

Current `Workspace.InitIndex` policy is explicitly: rebuild from scratch per invocation.

- `internal/workspace/index_builder.go`: `InitIndex(...)` closes any existing DB and rebuilds the entire in-memory DB.

For an HTTP server, doing that per request would be:

- too slow (startup cost per request),
- noisy (re-reading hundreds of docs),
- and concurrency-hostile (requests racing to rebuild/close DB).

So, a server needs an index lifecycle strategy, for example:

- **Build once at startup**, then:
  - refresh on explicit request (`POST /api/v1/index/rebuild`),
  - and/or refresh on a timer,
  - and/or refresh on file system events (watcher).

Practical recommendation (incremental complexity order):

1. **Build once + manual rebuild endpoint** (lowest complexity; good enough for early UI).
2. Add **debounced file watcher** later (optional).

### Concurrency & thread-safety notes

`Workspace` has mutable state:

- `w.db` can be replaced/closed during `InitIndex`.

But `QueryDocs` uses `w.db` for queries and then hydrates topics/related_files. This is safe if:

- `w.db` is not concurrently closed/replaced while queries run.

Therefore, a server should wrap access with a lock, for example:

- `RWMutex`: readers hold RLock during `QueryDocs`, writers hold Lock during `InitIndex`.

Alternative (simpler but less efficient):

- never rebuild in-process (only rebuild by restarting the server).

### Default safety: bind localhost, no auth by default

Docmgr is typically local tooling. A server should default to:

- bind to `127.0.0.1` (not `0.0.0.0`)
- no authentication by default (but allow adding later)
- explicit CORS policy if a browser UI is involved

### “Docmgr ticket workspace” also implies directory conventions

The discovery/doctor code uses “workspace scaffold markers” to recognize ticket directories:

- `internal/workspace/discovery.go`: `workspaceStructureMarkers = []string{"design", "reference", "playbooks", ...}`

Notes:

- This list is **directory-name based** and includes `design` but not `design-doc`.
- The repo currently contains both `design/` and `design-doc/` directories in older/newer tickets.

Implication for server work:

- If the server includes “workspace status” endpoints that emulate doctor/status behavior, it should either:
  - reuse these utilities as-is (and accept the marker mismatch), or
  - update marker lists to include both `design` and `design-doc` (potential refactor).

This is a concrete “where to refactor” candidate that likely matters for UI expectations.

## 2) Building a search REST API for docmgr search

### Mapping the CLI surface to a JSON endpoint

The most direct “keep semantics consistent” strategy is:

1. Parse query parameters (HTTP) into a struct that mirrors CLI `SearchSettings`.
2. Reuse the same internal pipeline the CLI uses:
   - `workspace.DiscoverWorkspace`
   - `ws.InitIndex` (but now cached, see above)
   - `ws.QueryDocs` (metadata/reverse lookup)
   - then apply the same post-filters for:
     - substring content search (`--query`)
     - external sources (`--external-source`)
     - date filters (`--since`, `--until`, `--created-since`, `--updated-since`)
   - format results into JSON response.

Relevant CLI implementation files:

- `pkg/commands/search.go`
  - `SearchCommand.RunIntoGlazeProcessor(...)`: the “real” engine path used by the cobra command
  - `SearchCommand.suggestFiles(...)`: a separate mode (`--files`)
- `cmd/docmgr/cmds/doc/search.go`: cobra wiring, dual-mode + glazed toggle

### What to reuse vs what to re-implement

#### Reuse: workspace index + query types

Core reusable types:

- `internal/workspace/query_docs.go`
  - `DocQuery`, `Scope`, `DocFilters`, `DocQueryOptions`
  - `DocHandle`, `DocQueryResult`
- `internal/workspace/query_docs_sql.go`
  - `compileDocQuery...` builds safe SQL with bound parameters

These are already “API shaped”: they express stable query semantics and return structured handles + diagnostics.

#### Likely re-implement: output formatting

The CLI uses Glaze to emit rows + multiple formats. The HTTP API should just emit JSON.

So the server should build a purpose-built response type that is either:

- explicitly versioned (recommended), or
- mirrors the Glaze row schema (convenient for parity, but less idiomatic).

#### Avoid: invoking cobra/Glaze from HTTP handlers

It’s tempting to call `SearchCommand` and capture JSON output, but that couples HTTP behavior to CLI plumbing and makes error handling, pagination, and performance harder to control.

Instead, factor the reusable search engine into a shared package (see “Potential refactors”).

### Search semantics you must decide explicitly

These decisions affect API contracts and UI expectations.

#### A) What does `query` mean?

Current CLI: `strings.Contains(strings.ToLower(body), strings.ToLower(query))`.

Pros:
- predictable
- matches existing behavior

Cons:
- no ranking
- no tokenization
- potentially slow if you have to scan many doc bodies

Recommendation for first version:

- Keep substring behavior as-is for parity.
- Add explicit “future: FTS” notes; do not silently change semantics.

#### B) Pagination + limits

`QueryDocs` currently returns all matching docs (in-memory). For an HTTP API:

- you likely want `limit` + `offset` (or cursor).

Where to implement:

- Option 1: add LIMIT/OFFSET support inside `compileDocQueryWithParseFilter(...)` (more invasive, but efficient).
- Option 2: fetch all then slice (simpler, but wastes work).

Since this ticket is “analysis + design”, the key thing is to call this out so the design doc can pick one.

#### C) Include/exclude categories by default

SQL compiler defaults to hiding:

- archived paths
- scripts paths
- control docs

unless explicitly included via `DocQueryOptions`.

The CLI search command **explicitly includes** them (it sets include flags to true). A REST API should decide:

- follow CLI search defaults (include these by default), or
- follow workspace query defaults (exclude unless requested).

Recommendation:

- Match `docmgr doc search` behavior for the `/search` endpoint; otherwise UI results will surprise users.

#### D) What counts as “created since”?

CLI uses `os.Stat(path).ModTime()` as a proxy for “created time”.

This is not stable “creation time” semantics, but it’s consistent with the current behavior.

If the API surfaces this, document it clearly.

#### E) Reverse lookup matching and path normalization

Reverse lookup is the UX-critical part:

- Users pass arbitrary paths (repo-relative, abs, doc-relative, sometimes just basenames).
- Docmgr normalizes and stores multiple representations for each related file, then matches with a fallback strategy.

Core code to understand:

- `internal/paths/resolver.go` (normalization anchors: repo root, docs root, doc path)
- `internal/workspace/index_builder.go` (normalizes and stores related_files representations)
- `internal/workspace/query_docs_sql.go` (`relatedFileExistsClause`, `relatedDirExistsClause`, path-key generation)
- `pkg/commands/search.go` (extra matching used for presentation/explanations in `--file` mode)

If the HTTP API includes reverse lookup endpoints, it should reuse `QueryDocs` rather than re-implementing matching logic in HTTP handlers.

### Where to look: files + symbols map

#### Existing search implementation (CLI)

- `docmgr/pkg/commands/search.go`
  - `type SearchCommand`
  - `type SearchSettings`
  - `(*SearchCommand).RunIntoGlazeProcessor(...)`
  - `(*SearchCommand).suggestFiles(...)`
- `docmgr/cmd/docmgr/cmds/doc/search.go` (`NewSearchCommand` wiring)
- `docmgr/cmd/docmgr/cmds/root.go` (adds the `docmgr search` alias)

#### Workspace discovery & root resolution

- `docmgr/internal/workspace/config.go`
  - `ResolveRoot`
  - `.ttmp.yaml` parsing via `LoadWorkspaceConfig`
  - `FindRepositoryRoot`
  - `ResolveVocabularyPath`
- `docmgr/internal/workspace/workspace.go`
  - `DiscoverWorkspace`
  - `WorkspaceContext`

#### Index build + schema

- `docmgr/internal/workspace/index_builder.go`
  - `(*Workspace).InitIndex`
  - `ingestWorkspaceDocs`
  - `inferTicketIDFromPath` (best-effort ticket inference for parse failures)
- `docmgr/internal/workspace/sqlite_schema.go`
  - `openInMemorySQLite`
  - `createWorkspaceSchema`

#### Query engine

- `docmgr/internal/workspace/query_docs.go`
  - `(*Workspace).QueryDocs`
  - `DocQuery`, `DocFilters`, `DocQueryOptions`
- `docmgr/internal/workspace/query_docs_sql.go`
  - `compileDocQueryWithParseFilter`

#### Ticket discovery scaffolding (likely relevant for server “workspace status” endpoints)

- `docmgr/internal/workspace/discovery.go`
  - `FindTicketScaffoldsMissingIndex`
  - `workspaceStructureMarkers`

### What to potentially refactor (high-value seams)

This is the “what will hurt if we don’t refactor” list for implementing an HTTP API cleanly.

#### 1) Extract a reusable search engine from `pkg/commands/search.go`

Problem:
- The CLI search command currently mixes:
  - request parsing (flags → settings),
  - query execution (`QueryDocs`),
  - post-filters,
  - output formatting (Glaze rows / human output),
  - and optional file-suggestion heuristics (`--files`).

For an HTTP API, we want:
- a clean function like `Search(ctx, ws, params) (results, diagnostics, error)`
- plus a separate “formatting” layer for:
  - CLI (Glaze/human)
  - HTTP (JSON)

Refactor target:
- new internal package (e.g. `internal/search` or `pkg/searchsvc`) that both CLI and server can call.

#### 2) Add explicit index caching / lifecycle helpers for server use

Problem:
- `InitIndex` always rebuilds and swaps `w.db`.
- There is no concurrency guard in `Workspace` itself.

Refactor targets:
- an `IndexManager` (holds a `*workspace.Workspace`, an `RWMutex`, and build timestamps)
- explicit methods:
  - `EnsureIndex(ctx)` (build once)
  - `RebuildIndex(ctx)` (swap safely)

This can live in a server package rather than in `workspace`, but it must exist somewhere.

#### 3) Consider adding SQL-level pagination (if the API needs it)

Problem:
- `QueryDocs` returns all matches; handlers may need `limit/offset`.

Refactor target:
- extend `DocQueryOptions` with `Limit`/`Offset`
- implement LIMIT/OFFSET in `compileDocQueryWithParseFilter(...)`.

If this is postponed, document in the design:
- “server slices in-memory for now; may be slow on huge workspaces”.

#### 4) Decide whether to index external sources / owners / summary

Problem:
- External sources filtering is a post-filter that re-reads frontmatter per doc.
- Owners/summary are not indexed either (at least not in current schema).

Refactor target:
- extend schema + ingestion to store these fields so API queries are fast and fully index-backed.

This is optional for first version, but relevant if the UI needs it.

#### 5) Directory naming consistency for doc types (design vs design-doc)

Problem:
- The repo contains both `design/` and `design-doc/` ticket subdirectories.
- Some detection logic uses directory names (`workspaceStructureMarkers`).
- New tickets currently scaffold `design/` even though vocabulary doc-type is `design-doc`.

Refactor targets (choose one, explicitly):
- unify on `design-doc/` everywhere, or
- keep `design/` as the canonical directory and treat `design-doc` as a vocabulary label only, or
- support both and normalize in tooling.

This matters for:
- server endpoints that list ticket workspace contents,
- UI navigation expectations,
- doctor/reporting accuracy.

## Related documentation (start here)

### Search internals & API-like contracts

- `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`
- `docmgr/pkg/doc/docmgr-cli-guide.md` (Search section)
- `docmgr/pkg/doc/docmgr-codebase-architecture.md` (Workspace + Query system)

### Workspace config + root resolution

- `docmgr/pkg/doc/docmgr-how-to-use.md` (how `.ttmp.yaml` and root resolution affects CLI usage)
- `docmgr/internal/workspace/config.go` (source of truth)

### Testing guidance relevant to changes in query semantics

- `docmgr/test-scenarios/` (scenario-based expectations)
- historical design/analysis around workspace refactors:
  - search via `docmgr doc search --query "Workspace.QueryDocs"` and follow the most relevant “REFACTOR-TICKET-REPOSITORY-HANDLING” docs.
