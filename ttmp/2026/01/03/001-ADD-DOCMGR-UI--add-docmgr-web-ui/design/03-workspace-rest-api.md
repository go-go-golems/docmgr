---
Title: "Design: docmgr Workspace REST API (for full-site navigation)"
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
  - docmgr
  - ui
  - http
  - api
  - workspace
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
  - Path: cmd/docmgr/cmds/api/serve.go
    Note: Single-process server entrypoint (serves /api + SPA)
  - Path: internal/httpapi/server.go
    Note: Current API route registration; mount new /workspace endpoints here
  - Path: internal/httpapi/index_manager.go
    Note: Index lifecycle; all API handlers depend on IndexManager snapshot
  - Path: internal/httpapi/tickets.go
    Note: Existing per-ticket endpoints (/tickets/*)
  - Path: pkg/commands/list_tickets.go
    Note: Existing CLI ticket listing logic to mirror in /workspace/tickets
  - Path: internal/workspace/query_docs.go
    Note: Core query engine to power ticket/doc listing + filtering
  - Path: pkg/doc/docmgr-http-api.md
    Note: User-facing HTTP API documentation (should eventually link to this)
ExternalSources: []
Summary: "Proposed REST API that exposes the entire ttmp workspace (ticket list + facets + navigation primitives) so a designer can build a cohesive docmgr site."
LastUpdated: 2026-01-05T00:00:00Z
WhatFor: "Hand this to a UI/UX designer as the stable API contract for workspace-wide navigation (tickets list, facets, recents) and drilling into tickets/docs."
WhenToUse: "Use to implement new /api/v1/workspace/* endpoints and drive a unified docmgr web UI beyond search."
---

# Design: docmgr Workspace REST API (for full-site navigation)

This document proposes a REST API surface that exposes the *entire* `ttmp` workspace so we can build one coherent website that supports:

- Workspace-wide navigation (tickets list, active/complete filters, recents, facets).
- Workspace-wide search (already exists).
- Ticket-specific pages (already implemented).
- Document/file viewers (already implemented).

The goal is to give a designer a clear contract for “what data is available and how the UI can navigate it”.

## 0) Mental model (for designers)

`docmgr` manages a workspace rooted at a “docs root” (usually `ttmp/`).

### Core entities

- **Workspace**: one `ttmp` root + a repo root, plus an in-memory search/index snapshot.
- **Ticket**: a directory under the workspace that contains an `index.md` doc with frontmatter that defines the canonical ticket metadata:
  - `Ticket` (ID), `Title`, `Status`, `Topics`, `Owners`, `Intent`, etc.
- **Document**: a markdown file under the ticket directory (or subdirectories) with YAML frontmatter. Documents can reference repo files via `RelatedFiles`.
- **Control docs**: `tasks.md`, `changelog.md`, `README.md` (often *not* indexed because they may not have frontmatter; treat as “files under the ticket”).

### Main UI flows to support

1. Workspace home → list tickets (filter: status/topic/owner/intent).
2. Workspace home → pick a ticket → ticket page (overview/docs/tasks/graph/changelog).
3. From anywhere → search docs (text + filters) → open ticket/doc/file.
4. From a ticket/doc → open related files (safe text-only file serving).

## 1) Index lifecycle and consistency

All “browse/search/list” endpoints depend on the server’s current **Index Snapshot** (in-memory SQLite DB).

- The server builds the index on startup and supports explicit refresh:
  - `POST /api/v1/index/refresh`
- The UI should treat the index snapshot as an eventually-consistent view of disk.
  - If the user edits docs in the filesystem, they should click “Refresh Index”.

The workspace API must expose “what snapshot am I looking at?”:
- `GET /api/v1/workspace/status` (already exists).

## 2) API conventions (contract)

### 2.1 Base URL

All endpoints are rooted at:
- `/api/v1`

### 2.2 Cursor pagination (standard)

List endpoints use cursor-based pagination:
- Request: `pageSize=<int>&cursor=<opaque>`
- Response: `nextCursor` (empty string means “no more”).

Cursor should be treated as opaque by clients.

### 2.3 Filtering conventions

- String filters accept empty string as “no filter”.
- Multi-value filters are CSV lists: `topics=a,b,c`.
- Booleans are `true|false`.

### 2.4 Error envelope (standard)

All errors returned as:

```json
{
  "error": {
    "code": "invalid_argument|not_found|forbidden|index_not_ready|...",
    "message": "human readable message",
    "details": { "field": "ticket", "value": "..." }
  }
}
```

## 3) Existing endpoints (already available)

These endpoints exist today and are part of the “workspace API” story:

**Health / index**
- `GET /api/v1/healthz`
- `GET /api/v1/workspace/status`
- `POST /api/v1/index/refresh`

**Search**
- `GET /api/v1/search/docs` (cursor pagination; `orderBy=rank|path|last_updated`)
- `GET /api/v1/search/files` (suggestions)

**Ticket page (already implemented)**
- `GET /api/v1/tickets/get?ticket=...`
- `GET /api/v1/tickets/docs?ticket=...`
- `GET /api/v1/tickets/tasks?ticket=...`
- `POST /api/v1/tickets/tasks/check`
- `POST /api/v1/tickets/tasks/add`
- `GET /api/v1/tickets/graph?ticket=...`

**Doc / file viewer**
- `GET /api/v1/docs/get?path=...` (frontmatter + body best-effort)
- `GET /api/v1/files/get?root=repo|docs&path=...` (safe text-only)

For a “single cohesive site”, we need a *workspace-level browse surface* to power a ticket list page and global navigation facets.

## 4) Proposed workspace endpoints (new)

### 4.1 Workspace summary

`GET /api/v1/workspace/summary`

Purpose:
- Render the “home” dashboard without making multiple calls.

Response shape:

```json
{
  "root": "ttmp",
  "repoRoot": "/abs/repo",
  "indexedAt": "2026-01-05T00:00:00Z",
  "docsIndexed": 413,
  "stats": {
    "ticketsTotal": 128,
    "ticketsActive": 12,
    "ticketsComplete": 84,
    "ticketsReview": 9,
    "ticketsDraft": 23
  },
  "recent": {
    "tickets": [ /* TicketListItem[] */ ],
    "docs": [ /* minimal doc list items */ ]
  }
}
```

Notes:
- This is a convenience endpoint; it can be implemented using the same primitives as `/workspace/tickets` + a couple of small aggregate queries.

### 4.2 List tickets (workspace-wide)

`GET /api/v1/workspace/tickets`

Purpose:
- Power the primary “Tickets” view (table/board/list).
- This endpoint returns one row per ticket, derived from the ticket’s `index.md` frontmatter.

Query params:
- `status=`: `active|review|complete|draft|` (empty = all)
- `ticket=`: exact ticket ID match (optional)
- `topics=`: CSV, match *any* topic (optional)
- `owners=`: CSV, match *any* owner (optional)
- `intent=`: exact match (optional)
- `q=`: FTS query applied to index docs only (optional)
- `orderBy=`: `last_updated|ticket|title` (default `last_updated`)
- `reverse=`: `true|false` (default `false`)
- `includeArchived=`: `true|false` (default `true`)
- `pageSize=`: `1..1000` (default `200`)
- `cursor=`: opaque cursor
- `includeStats=`: `true|false` (default `false`)
  - When false, avoid expensive per-ticket stats like tasks count.

Response:

```json
{
  "query": {
    "q": "search string",
    "status": "active",
    "topics": ["docmgr", "ui"],
    "owners": ["manuel"],
    "intent": "long-term",
    "orderBy": "last_updated",
    "reverse": false,
    "includeArchived": true,
    "includeStats": false,
    "pageSize": 200,
    "cursor": ""
  },
  "total": 128,
  "results": [
    {
      "ticket": "001-ADD-DOCMGR-UI",
      "title": "Add docmgr Web UI",
      "status": "active",
      "topics": ["docmgr", "ui"],
      "owners": [],
      "intent": "long-term",
      "createdAt": "2026-01-03",
      "updatedAt": "2026-01-05T00:00:00Z",
      "ticketDir": "2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui",
      "indexPath": "2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/index.md",
      "snippet": "optional snippet when q is used",
      "stats": null
    }
  ],
  "nextCursor": ""
}
```

Optional `stats` (only when `includeStats=true`):

```json
{
  "docsTotal": 6,
  "tasksTotal": 27,
  "tasksDone": 25,
  "relatedFilesTotal": 17
}
```

Implementation notes:
- Mirrors the CLI `docmgr list tickets` logic from `pkg/commands/list_tickets.go` but served as JSON.
- The authoritative ticket list is “all docs with `DocType: index`”.
- For `q=`:
  - Restrict to `DocType: index` and apply `TextQuery` (FTS5); return `snippet` and possibly `rank`.

### 4.3 Ticket facets (for filters)

`GET /api/v1/workspace/facets`

Purpose:
- Drive UI filter controls without scanning client-side.

Response:

```json
{
  "statuses": ["active", "review", "complete", "draft"],
  "docTypes": ["index", "design-doc", "reference", "analysis", "sources"],
  "intents": ["short-term", "long-term", "evergreen"],
  "topics": ["docmgr", "ui", "tooling", "infra", "..."],
  "owners": ["manuel", "alex", "..."]
}
```

Notes:
- Sources of truth:
  - Prefer `vocabulary.yaml` if present (controlled vocab).
  - Otherwise derive from the indexed docs table (distinct values).

### 4.4 Workspace “recent activity”

`GET /api/v1/workspace/recent`

Purpose:
- A lightweight endpoint to show “recently updated tickets” and “recently updated docs” in the UI homepage/left nav.

Query params:
- `ticketsLimit=` (default 20)
- `docsLimit=` (default 20)
- `includeArchived=` (default true)

Response:

```json
{
  "tickets": [ /* TicketListItem[] */ ],
  "docs": [
    {
      "path": "2026/01/03/.../design/01-design.md",
      "ticket": "001-ADD-DOCMGR-UI",
      "title": "Design: docmgr Search Web UI",
      "docType": "design-doc",
      "status": "active",
      "topics": ["docmgr", "ui"],
      "updatedAt": "2026-01-05T00:00:00Z"
    }
  ]
}
```

Notes:
- This is mostly `OrderBy=last_updated` queries with small page sizes.

### 4.5 Workspace doc listing (optional)

`GET /api/v1/workspace/docs`

Purpose:
- A browse page for all docs, not just search-driven discovery.

Query params:
- `docType=`, `ticket=`, `status=`, `topics=`, `orderBy=path|last_updated`, `pageSize`, `cursor`

Response:
- Similar to `/api/v1/search/docs` results but without rank/snippet semantics unless `q` is provided.

This endpoint is optional if we’re happy with:
- workspace docs browsing = `/search/docs?query=&...` with `allowEmpty=true`.

## 5) UI primitives the designer can assume exist

If the UI follows the API surface above, the designer can assume:

- A **Tickets** page can be built around `/workspace/tickets` with a filter sidebar powered by `/workspace/facets`.
- A **Home/Dashboard** can use `/workspace/summary` and `/workspace/recent`.
- A **Ticket page** is already supported (and should be the main drill-down route).
- A **Doc viewer** is supported via `/docs/get` and a markdown renderer in the UI.
- A **File viewer** is supported via `/files/get` (safe text-only) + syntax highlighting.
- A **Global search** is supported via `/search/docs` and can be embedded anywhere.

## 6) Suggested site map (for design)

Routes (proposal):

- `/` — Workspace home (summary + recent)
- `/tickets` — Ticket list (filters + pagination)
- `/tickets/:ticket` (or `/ticket/:ticket`) — Ticket page (tabs; already)
- `/search` — Search page (already “home” today, but can become `/search`)
- `/doc?path=...` — Doc viewer (already)
- `/file?root=repo|docs&path=...` — File viewer (already)

Navigation:
- Left nav: Home, Tickets, Search, (optional) Topics, Recent
- Ticket badge anywhere links to `/ticket/:ticket`

## 7) Non-goals / boundaries (to keep scope sane)

- No auth/multi-user semantics in v1 (localhost tool).
- No “edit markdown” API in v1 (read-only viewer, plus task toggles).
- No general filesystem browsing beyond safe roots (`docs` root and `repo` root) and size-limited text serving.

