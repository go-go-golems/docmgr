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

Security note: the server is **local-first**. Bind to `127.0.0.1` by default and don’t expose it publicly unless you add authentication and threat-model it.

## 2. Quick Start

Build and run the server:

```bash
go build -tags sqlite_fts5 -o /tmp/docmgr ./cmd/docmgr
/tmp/docmgr api serve --addr 127.0.0.1:8787 --root ttmp
```

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
      "snippet": "...",
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

