---
Title: docmgr HTTP API
Slug: http-api
Short: Run docmgr as a local HTTP server with a JSON search API (v1), cursor pagination, and explicit index refresh.
Topics:
- docmgr
- api
- search
- http
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# docmgr HTTP API

## 1. Overview

`docmgr` can run as a local HTTP server exposing a versioned JSON API for searching documentation using the same query engine as the CLI.

This is intended for:
- A web UI (or IDE plugin) that needs stable JSON responses
- Avoiding per-request CLI spawns
- Reusing docmgr’s reverse lookup + path normalization + diagnostics

See also: `docmgr-web-ui.md` (Slug: `web-ui`) for running the bundled Search Web UI.

Security note: the server is **local-first**. Bind to `127.0.0.1` by default and don’t expose it publicly unless you add authentication and threat-model it.

## 2. Quick Start

Build and run the server:

```bash
go build -tags "sqlite_fts5,embed" -o /tmp/docmgr ./cmd/docmgr
/tmp/docmgr api serve --addr 127.0.0.1:8787 --root ttmp
```

Notes:
- `sqlite_fts5` enables full-text search.
- `embed` bundles the web UI assets into the binary (optional for API-only usage).

Check health:

```bash
curl -s http://127.0.0.1:8787/api/v1/healthz
```

Refresh the index (explicit):

```bash
curl -s -X POST http://127.0.0.1:8787/api/v1/index/refresh
```

Search:

```bash
curl -s "http://127.0.0.1:8787/api/v1/search/docs?query=websocket&orderBy=rank&pageSize=50"
```

## 3. Concepts

### 3.1. Index Lifecycle (IndexManager)

The server builds an in-memory SQLite index on startup and reuses it for requests.

- Startup: discover workspace + build index
- Runtime: queries read from the current index
- Refresh: `POST /api/v1/index/refresh` rebuilds the index from disk and swaps it in atomically

This is intentionally “refresh-on-demand” for simplicity (no file watching yet).

### 3.2. Query Semantics (FTS5)

The `query` parameter uses SQLite FTS5 `MATCH` syntax and is **not** a substring/contains search.

- Build/install with `-tags sqlite_fts5` to enable full-text search.
- If FTS is unavailable and a request includes `query`, the API returns an error.

Ranking:

- `orderBy=rank` orders by `bm25` score (best matches first).

### 3.3. Cursor Pagination

`GET /api/v1/search/docs` supports cursor-based pagination:

- Request: `pageSize` + `cursor`
- Response: `nextCursor` (opaque; pass it back as `cursor` to fetch the next page)

For v1, cursors may be implemented internally using offsets but are treated as opaque by clients.

## 4. Running the Server

### 4.1. Command

```bash
docmgr api serve --addr 127.0.0.1:8787 --root ttmp
```

Flags:
- `--addr`: bind address (default `127.0.0.1:8787`)
- `--root`: docs root directory (default `ttmp`)
- `--cors-origin`: if set, adds CORS headers for browser-based UIs

## 5. API Reference (v1)

Base path: `/api/v1`

### 5.1. Health

`GET /api/v1/healthz`

Response:

```json
{ "ok": true }
```

### 5.2. Workspace Status

`GET /api/v1/workspace/status`

Purpose: show which workspace is currently indexed and basic index metadata.

Response (shape):

```json
{
  "root": "/abs/path/to/ttmp",
  "configDir": "/abs/path",
  "repoRoot": "/abs/path/to/repo",
  "configPath": "/abs/path/to/.ttmp.yaml",
  "vocabularyPath": "/abs/path/to/ttmp/vocabulary.yaml",
  "indexedAt": "2026-01-04T21:05:04.583Z",
  "docsIndexed": 200,
  "ftsAvailable": true
}
```

### 5.2.1. Workspace Summary

`GET /api/v1/workspace/summary`

Purpose: render the Workspace home/dashboard with one call (basic stats + recent tickets + recent docs).

Response (shape):

```json
{
  "root": "/abs/path/to/ttmp",
  "repoRoot": "/abs/path/to/repo",
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
    "tickets": [
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
        "snippet": "",
        "stats": null
      }
    ],
    "docs": [
      {
        "path": "2026/01/03/.../design/03-workspace-rest-api.md",
        "ticket": "001-ADD-DOCMGR-UI",
        "title": "Design: docmgr Workspace REST API (for full-site navigation)",
        "docType": "design-doc",
        "status": "active",
        "topics": ["docmgr", "ui", "http", "api", "workspace"],
        "updatedAt": "2026-01-05T00:00:00Z"
      }
    ]
  }
}
```

### 5.3. Refresh Index

`POST /api/v1/index/refresh`

Response (shape):

```json
{
  "refreshed": true,
  "indexedAt": "2026-01-04T21:05:04.583Z",
  "docsIndexed": 200,
  "ftsAvailable": true
}
```

### 5.3.1. Workspace Tickets

`GET /api/v1/workspace/tickets`

Purpose: list tickets (workspace-wide) derived from the ticket `index.md` docs (`DocType: index`).

Query parameters:
- `status` (string): `active|review|complete|draft|` (empty = all)
- `ticket` (string): exact ticket ID match (optional)
- `topics` (string): comma-separated, match any topic (optional)
- `owners` (string): comma-separated, match any owner (optional)
- `intent` (string): exact match (optional)
- `q` (string): full-text query (FTS5) applied to index docs only (optional)
- `orderBy` (string): `last_updated|ticket|title` (default `last_updated`)
- `reverse` (bool, default `false`)
- `includeArchived` (bool, default `true`)
- `includeStats` (bool, default `false`): when true, computes per-ticket stats (tasks/docs/related files)
- `pageSize` (int, default `200`, max `1000`)
- `cursor` (string, optional)

Response (shape):

```json
{
  "query": {
    "q": "",
    "status": "active",
    "ticket": "",
    "topics": ["docmgr", "ui"],
    "owners": [],
    "intent": "",
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
      "snippet": "",
      "stats": null
    }
  ],
  "nextCursor": ""
}
```

### 5.3.2. Workspace Facets

`GET /api/v1/workspace/facets`

Purpose: drive Workspace filters (statuses/docTypes/intents/topics/owners).

Query parameters:
- `includeArchived` (bool, default `true`)

Response (shape):

```json
{
  "statuses": ["active", "review", "complete", "draft"],
  "docTypes": ["index", "design-doc", "reference", "analysis", "sources"],
  "intents": ["short-term", "long-term", "evergreen"],
  "topics": ["docmgr", "ui", "tooling"],
  "owners": ["manuel", "alex"]
}
```

Notes:
- The server prefers `vocabulary.yaml` for `statuses/docTypes/intents/topics` when present, but falls back to deriving from indexed docs.

### 5.3.3. Workspace Recent

`GET /api/v1/workspace/recent`

Purpose: show “recently updated tickets” and “recently updated docs”.

Query parameters:
- `ticketsLimit` (int, default `20`, max `1000`)
- `docsLimit` (int, default `20`, max `1000`)
- `includeArchived` (bool, default `true`)

Response (shape):

```json
{
  "tickets": [ /* TicketListItem[] (same as /workspace/tickets results) */ ],
  "docs": [
    {
      "path": "2026/01/03/.../design/03-workspace-rest-api.md",
      "ticket": "001-ADD-DOCMGR-UI",
      "title": "Design: docmgr Workspace REST API (for full-site navigation)",
      "docType": "design-doc",
      "status": "active",
      "topics": ["docmgr", "ui"],
      "updatedAt": "2026-01-05T00:00:00Z"
    }
  ]
}
```

### 5.3.4. Workspace Topics

`GET /api/v1/workspace/topics`

Purpose: list topics (workspace-wide) with basic counts.

Query parameters:
- `includeArchived` (bool, default `true`)

Response (shape):

```json
{
  "total": 42,
  "results": [
    { "topic": "docmgr", "docsTotal": 120, "ticketsTotal": 14, "updatedAt": "2026-01-05T00:00:00Z" }
  ]
}
```

### 5.3.5. Workspace Topic Detail

`GET /api/v1/workspace/topics/get`

Query parameters:
- `topic` (string, required)
- `includeArchived` (bool, default `true`)
- `docsLimit` (int, default `20`, max `1000`)

Response (shape):

```json
{
  "topic": "docmgr",
  "stats": {
    "ticketsTotal": 14,
    "ticketsActive": 6,
    "ticketsComplete": 4,
    "ticketsReview": 2,
    "ticketsDraft": 2
  },
  "tickets": [ /* TicketListItem[] */ ],
  "docs": [ /* RecentDocItem[] */ ]
}
```

### 5.4. Search Docs

`GET /api/v1/search/docs`

Query parameters:

- `query` (string): FTS5 `MATCH` query string
- `ticket` (string)
- `topics` (string): comma-separated
- `docType` (string)
- `status` (string)
- `file` (string): reverse lookup
- `dir` (string): reverse lookup
- `externalSource` (string)
- `since` (string)
- `until` (string)
- `createdSince` (string)
- `updatedSince` (string)

Visibility toggles (defaults mirror CLI behavior):
- `includeArchived` (bool, default `true`)
- `includeScripts` (bool, default `true`)
- `includeControlDocs` (bool, default `true`)
- `includeDiagnostics` (bool, default `true`)
- `includeErrors` (bool, default `false`)

Sorting:
- `orderBy`: `path|last_updated|rank` (default `path`)
- `reverse` (bool, default `false`)

Reverse lookup notes:

- `reverse=true` searches docs by `RelatedFiles` references.
- It normally requires `file` or `dir`.
- As a convenience, if `reverse=true` and `file`/`dir` are empty but `query` is set, the server treats `query` as `file`.

Pagination:
- `pageSize` (int, default `200`, max `1000`)
- `cursor` (string, optional)

Response (shape):

```json
{
  "query": { "query": "websocket", "pageSize": 50, "cursor": "" },
  "total": 12,
  "results": [
    {
      "ticket": "MEN-4242",
      "title": "Chat WebSocket Lifecycle",
      "docType": "reference",
      "status": "active",
      "topics": ["chat", "backend", "websocket"],
      "path": "2026/01/04/MEN-4242--.../reference/01-chat-websocket-lifecycle.md",
      "lastUpdated": "2026-01-04T15:04:05Z",
      "snippet": "...",
      "relatedFiles": [
        { "path": "backend/chat/ws/manager.go", "note": "WebSocket lifecycle (scenario)" }
      ],
      "matchedFiles": ["backend/chat/ws/manager.go"],
      "matchedNotes": ["WebSocket lifecycle (scenario)"]
    }
  ],
  "diagnostics": [],
  "nextCursor": "..."
}
```

### 5.5. Suggest Files

`GET /api/v1/search/files`

Query parameters:
- `query` (string)
- `ticket` (string)
- `topics` (string): comma-separated
- `limit` (int, default `200`, max `1000`)

Response (shape):

```json
{
  "total": 3,
  "results": [
    { "file": "backend/chat/ws/manager.go", "source": "related_files", "reason": "..." }
  ]
}
```

### 5.6. Get Document (markdown + frontmatter)

`GET /api/v1/docs/get`

Query parameters:
- `path` (string, required): doc-relative path under the docs root (same value as `SearchDocResult.path`)

Response (shape):

```json
{
  "path": "2026/01/03/TICKET--slug/design/01-doc.md",
  "doc": {
    "title": "...",
    "ticket": "...",
    "status": "...",
    "topics": ["..."],
    "docType": "...",
    "intent": "...",
    "owners": ["..."],
    "relatedFiles": [{ "path": "internal/foo.go", "note": "..." }],
    "externalSources": [],
    "summary": "",
    "lastUpdated": "2026-01-04T19:22:44-05:00",
    "whatFor": "",
    "whenToUse": ""
  },
  "relatedFiles": [{ "path": "internal/foo.go", "note": "..." }],
  "body": "# Markdown…",
  "stats": { "sizeBytes": 12345, "modTime": "2026-01-04T19:22:44-05:00" },
  "diagnostic": null
}
```

Notes:
- If the document frontmatter fails to parse, `doc` will be omitted and `diagnostic` may be present; `body` still returns the markdown body (best-effort).

### 5.7. Get File (text-only)

`GET /api/v1/files/get`

Query parameters:
- `path` (string, required): a file path (repo-relative is recommended; absolute paths are only accepted if they resolve inside the allowed root)
- `root` (string, optional): `repo|docs` (default `repo`)

Response (shape):

```json
{
  "path": "internal/httpapi/server.go",
  "root": "repo",
  "language": "go",
  "contentType": "text/x-go; charset=utf-8",
  "truncated": false,
  "content": "package httpapi\n...",
  "stats": { "sizeBytes": 12345, "modTime": "2026-01-04T19:22:44-05:00" }
}
```

Safety behavior:
- Requests are constrained to repo root or docs root.
- Path traversal and symlink-escape reads are rejected.
- Binary files and non-UTF8 are rejected (`unsupported_media_type`).
- Large files may be truncated (see `truncated`).

## 6. Error Handling

All error responses use a stable JSON envelope:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "must provide at least a query or filter",
    "details": {}
  }
}
```

Common error codes:
- `index_not_ready` (503): the index is not initialized
- `invalid_cursor` (400): cursor is malformed
- `fts_not_available` (400): request uses `query` but FTS is unavailable
- `internal` (500): unexpected server error

## 7. Troubleshooting

### Query returns `fts_not_available`

Build with `-tags sqlite_fts5` and restart the server:

```bash
go build -tags sqlite_fts5 -o /tmp/docmgr ./cmd/docmgr
/tmp/docmgr api serve --addr 127.0.0.1:8787 --root ttmp
```

### Results don’t reflect recent file changes

Call refresh:

```bash
curl -s -X POST http://127.0.0.1:8787/api/v1/index/refresh
```
