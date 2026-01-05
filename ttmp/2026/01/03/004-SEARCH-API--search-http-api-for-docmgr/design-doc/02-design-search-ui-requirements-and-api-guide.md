---
Title: "Design: Search UI Requirements + HTTP API Guide"
Ticket: 004-SEARCH-API
Status: draft
Topics:
  - backend
  - docmgr
  - tooling
  - ux
  - web
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
  - Path: internal/httpapi/server.go
    Note: REST API endpoints and JSON shapes
  - Path: internal/searchsvc/search.go
    Note: Shared query engine backing CLI + HTTP
  - Path: internal/workspace/query_docs.go
    Note: Core query model (filters, ordering)
ExternalSources: []
Summary: "A designer-facing guide to docmgr’s search concepts and the v1 HTTP API, plus UX goals and feature wishlist for a high-end search UI."
LastUpdated: 2026-01-04T22:05:00-05:00
WhatFor: "Enable a designer to create a best-in-class search UI that exercises docmgr’s search capabilities and supports developer workflows."
WhenToUse: "When designing the Web UI for docmgr search and planning interaction patterns that map to API capabilities."
---

# Design: Search UI Requirements + HTTP API Guide

## 1. Audience + Goal

This document is for a **product/designer** designing a “killer” search experience for docmgr. It explains:

- what docmgr is and how developers use it day-to-day,
- what “search” means in docmgr (it’s more than text: it’s metadata + reverse lookup),
- what the HTTP API can do today (v1),
- what UI features we want to **exercise** as developers,
- what we’d like to add next (UI-driven roadmap signals).

## 2. What is docmgr (mental model)

docmgr is a documentation system that keeps “ticket-scoped” documentation close to code. It expects a docs root (by default `ttmp/`) containing ticket directories and markdown files with structured YAML frontmatter.

### 2.1. Key entities (as users experience them)

- **Docs root** (`ttmp/`): the canonical place where docmgr looks for documentation.
- **Ticket workspace**: a directory for a single ticket (e.g. `ttmp/YYYY/MM/DD/MEN-4242--slug/`).
- **Document**: a markdown file inside a ticket workspace (e.g. `design-doc/01-foo.md`), with frontmatter fields:
  - `Ticket`, `Title`, `DocType`, `Status`, `Topics`, `Intent`, `Owners`, `Summary`, `LastUpdated`
  - `RelatedFiles` (code-to-doc links, with optional notes)
  - `ExternalSources` (URLs / refs linked to the ticket/doc)
- **Control docs**: `index.md`, `tasks.md`, `changelog.md` are “special” operational docs used for ticket tracking.

### 2.2. Why developers care (primary workflows)

Developers use docmgr search to answer questions like:

1. **Discovery / recall**: “Where did we decide X?” “Do we have a design doc for this area?”
2. **Code review context**: “I’m looking at `backend/chat/ws/manager.go` — what docs are related?”
3. **Implementation navigation**: “Show me all docs for ticket MEN-4242, filtered to design docs.”
4. **Maintenance**: “What docs were updated recently?” “What’s stale?”
5. **Linking and traceability**: “Which docs reference this external RFC or URL?”

So search UI must support both:
- “Google-style” text searching (with a good results UX), and
- “structured navigation” over tickets/doc types/status/topics/files.

## 3. What “search” means in docmgr

Search is a combined query over:

- **Full-text content** (`query`): FTS-backed (SQLite FTS5 `MATCH`) when enabled.
- **Metadata filters**: ticket, topics, doc type, status.
- **Reverse lookup**:
  - `file`: find docs that relate to a file (or include it in `RelatedFiles`)
  - `dir`: find docs relating to a directory
- **External sources**: filter docs referencing a URL / external reference string.
- **Time filters**: “since/until/updatedSince/createdSince”.

Search results include:
- doc path (docs-root relative),
- title/ticket/docType/status/topics,
- snippet (best-effort excerpt around query),
- file match details (when searching by `file`): `matchedFiles`, `matchedNotes`,
- diagnostics (optional): warnings about “fallback matching” when path normalization had to guess.

## 4. The HTTP API (v1) – what exists today

Base path: `/api/v1`

The server is started via:

```bash
docmgr api serve --addr 127.0.0.1:8787 --root ttmp
```

Index lifecycle:
- built on startup,
- refreshed explicitly via `POST /api/v1/index/refresh`.

### 4.1. Health

`GET /api/v1/healthz`

Used by:
- UI startup “server reachable” check
- uptime indicator

### 4.2. Workspace status (environment + index state)

`GET /api/v1/workspace/status`

Used by:
- showing which repo/root the server is serving,
- showing “indexedAt / docsIndexed / ftsAvailable” state,
- quick “something is wrong” debugging.

### 4.3. Index refresh

`POST /api/v1/index/refresh`

Used by:
- a “Refresh index” button in UI,
- developer workflow after editing docs.

This is a big UX affordance: it’s the simplest way to make the UI “feel live” without a file watcher.

### 4.4. Search docs (core endpoint)

`GET /api/v1/search/docs`

Parameters (selected):

- `query`: **FTS5 MATCH query string** (not substring search)
- `orderBy`: `path|last_updated|rank`
- `topics`: comma-separated
- `ticket`, `docType`, `status`
- `file` / `dir` (reverse lookup)
- `externalSource`
- `since`, `until`, `createdSince`, `updatedSince` (date expressions supported)
- toggles: `includeArchived`, `includeScripts`, `includeControlDocs`, `includeDiagnostics`, `includeErrors`

Pagination:
- `pageSize` (default 200; max 1000)
- `cursor` (opaque)
- response includes `nextCursor`

### 4.5. Suggest files

`GET /api/v1/search/files`

Purpose:
- “Given this ticket/topic/query, which code files are likely relevant?”

This powers a UI panel like “Related code files” or “Suggested files to relate” (even if the actual relate action remains in CLI for now).

## 5. Important semantics designers must understand

### 5.1. `query` is not “contains”

`query` uses **SQLite FTS5** `MATCH` syntax. This has UX consequences:

- Quoting matters.
- Operators exist (AND/OR/NOT, prefix matching with `*`, etc.).
- A single string may error if syntactically invalid for MATCH.

UI implication:
- Provide **inline help** (“FTS query syntax”) and **error recovery**.
- Consider “basic mode” vs “advanced mode”, where basic mode safely escapes user input (if we add server support later).

### 5.2. Ranking exists

`orderBy=rank` returns “best matches first”.

UI implication:
- default sort may be `rank` when `query` is present.
- show sort dropdown (Path / Last Updated / Relevance).

### 5.3. Cursor pagination

Cursor pagination supports:
- infinite scroll,
- “Load more” button,
- consistent UX without requiring the UI to manage offsets.

UI implication:
- keep the **current cursor** as part of the query state.
- show “N results shown of total” and a progress indicator.

### 5.4. Index refresh is explicit

Docs change on disk won’t show up until refresh is called.

UI implication:
- show an “Index stale?” hint (we can later add file watching).
- make refresh accessible and fast-feeling (spinner + last refreshed timestamp).

### 5.5. Reverse lookup is a first-class feature

Reverse lookup (`file`/`dir`) is key for code review workflows.

UI implication:
- a dedicated “Search by file path” input
- accept partial/basename queries (e.g. `register.go`) and show any diagnostics explaining fallback behavior

## 6. “Killer” Search UI – feature wishlist (developer-centric)

This is what we’d like to be able to do as developers.

### 6.1. Fast “global” search with strong information scent

Results list items should show:
- Title (primary)
- Ticket ID + status + doc type (secondary)
- Topics chips
- Snippet
- Path (copyable)
- “Last updated” (if/when API returns it; currently not included in v1 results)

### 6.2. Faceted filters (chips + counts)

Even if the API doesn’t return facet counts yet, the UI can:
- show selected filters as removable chips,
- provide “filter pickers”:
  - ticket (autocomplete)
  - topics (multi-select)
  - doc type (select)
  - status (select)
- optionally show counts once we add a facets endpoint.

### 6.3. Search modes (tabs)

We want a UI that makes search intent explicit:

1. **Docs** (default): `GET /search/docs`
2. **Reverse lookup**: file/dir oriented UI, still backed by `/search/docs`
3. **Suggested files**: `/search/files`

### 6.4. Preview / reader panel

We want to click a result and see:
- rendered markdown preview, or at minimum the snippet + metadata.

Note: v1 API does not provide full document body. Options:
- show snippet-only for now,
- add a follow-up endpoint later: `GET /api/v1/docs/content?path=...`.

### 6.5. Deep links + shareable URLs

The UI should:
- reflect current query state in the URL (including filters and cursor state),
- allow copying a “link to this search”.

### 6.6. Keyboard-first UX

Developers live on the keyboard. Ideal:
- `/` focuses search
- arrow keys navigate results
- `Enter` opens preview
- `Cmd/Ctrl+Enter` opens in new tab (if preview is separate)
- `Esc` closes panels / clears.

### 6.7. “Refresh Index” affordance

Must be obvious and low-friction:
- button with last refreshed timestamp from `/workspace/status`
- indicates running refresh when pressed
- disables while refresh in-flight

### 6.8. Diagnostics display (trust-building)

Search can include `diagnostics` that explain why a reverse lookup match relied on weaker fallbacks.

We want:
- a small “warnings” indicator in results header,
- expandable panel showing diagnostics (developer trust feature).

### 6.9. Error handling UX

We want:
- clear inline errors when query is invalid (e.g. invalid cursor, invalid FTS query),
- a quick “reset to basic search” action,
- graceful empty state (“no results”) with suggestions (remove filters, switch sort, etc.).

### 6.10. History + saved searches

Nice-to-have:
- local history of recent searches (persisted)
- saved searches (“My ticket dashboard”, “Stale docs > 30 days”, etc.)

## 7. UI requirements mapped to API calls

### 7.1. App startup

1) poll `GET /api/v1/healthz`
2) fetch `GET /api/v1/workspace/status`
3) if index not ready, offer “Refresh index” (POST) or show error.

### 7.2. Running a search

- always call `GET /api/v1/search/docs` with current query state
- on “load more”, call again with `cursor=nextCursor`
- if filters change, reset cursor.

### 7.3. Suggested files sidebar

- call `GET /api/v1/search/files?ticket=...&topics=...&query=...`
- show list of files with “reason” and “source” to build trust.

## 8. Open questions (for designer + devs)

1. Should we design “basic query” and “advanced query” modes, anticipating we add an API knob to safely escape a user’s plain text into FTS?
2. Do we need “ticket list” / “topics list” endpoints for autocompletion and facet building, or is it acceptable to require users to paste values initially?
3. Should the UI embed a markdown reader (requires a “get doc content” endpoint), or is link-out sufficient for v1?
4. How much of the “ticket workflow” (tasks, changelog, close) should the UI eventually surface vs being search-only?

## 9. Success criteria (how we’ll judge the UI)

For developer workflows, a “killer” UI means:

- Finding the right doc in <10 seconds for common queries.
- Reverse lookup “paste file path → get docs” works reliably and builds trust.
- Filter UX is fast and doesn’t feel like a form.
- Index refresh is obvious and eliminates “why didn’t my change show up?” confusion.
- The UI is shareable (URL state) and keyboard-friendly.

