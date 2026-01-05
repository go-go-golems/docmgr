---
Title: Doc Serving API and Document Viewer UI
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - tooling
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/documents/read_document.go
      Note: Doc parsing helper (frontmatter + body)
    - Path: internal/httpapi/server.go
      Note: Where to add /api/v1/docs/* and /api/v1/files/* endpoints
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/02-single-doc.md
      Note: UX spec snapshot
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/03-single-doc.html
      Note: UI target for doc viewer
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/single-doc.html
      Note: UI target for doc viewer
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/single-doc.md
      Note: UX spec snapshot
    - Path: ui/src/features/search/SearchPage.tsx
      Note: Search UI entrypoint that should link into doc viewer
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T19:49:02.266829388-05:00
WhatFor: ""
WhenToUse: ""
---



# Analysis: Doc Serving API and Document Viewer UI

## Goal

Add a “view document” workflow to the new web UI:

- Search results → open a doc viewer page for a specific doc path.
- Viewer renders markdown nicely (Bootstrap-friendly) and supports syntax highlighting.
- RelatedFiles entries become actionable: open related source files and display them with highlighting.

Reference UI/UX targets:

- `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/03-single-doc.html`
- `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/02-single-doc.md`

## Current State (what exists today)

- We have a search UI (`ui/`) that consumes `/api/v1/search/docs` and shows:
  - snippet + metadata,
  - relatedFiles list (from frontmatter),
  - preview panel (snippet-only).
- The server provides:
  - `GET /api/v1/search/docs` returning `path` (doc-relative path under docs root)
  - `GET /api/v1/workspace/status` returning `root` and `repoRoot`
- We do not have:
  - a “fetch this doc body” endpoint,
  - a safe “fetch this source file content” endpoint,
  - any markdown rendering beyond simple snippet display.

## UX Requirements (derived from single-doc.html)

### Document viewer layout

1) Header block
- Title
- Ticket • docType • status badge

2) Metadata section (table)
- Ticket, Status, Doc Type, Intent, Topics, Owners, Last Updated
- Copy-to-clipboard for ticket and doc path (and maybe title)

3) Related files section
- Each file shows path + note
- Copy path button
- “Open related code” becomes a real action:
  - either open a file viewer route (recommended), or open inline modal/pane.

4) Doc content section
- Render markdown into HTML
- Styled headings, paragraphs, lists
- Inline code styling
- Fenced code blocks with syntax highlighting

5) Sticky action bar
- keyboard hints (Esc close/back, Ctrl/Cmd+C copy all)
- Copy path
- Back to search

### What “document view” must support

- Markdown features commonly used by docmgr docs:
  - headings, lists, emphasis, inline code, fenced code blocks
  - tables (GFM) are likely useful
  - links (relative + absolute)
  - images (optional in v1; see below)

### What “source file view” must support

- Render plain text with line-preserving formatting.
- Syntax highlight by language:
  - use extension inference (`.go`, `.ts`, `.md`, `.yaml`, etc)
  - allow a `lang` override for edge cases.
- Show metadata:
  - resolved absolute path? (maybe hidden; prefer safe relative)
  - last modified
  - size (and whether truncated)
- Guardrails:
  - refuse binary files,
  - refuse large files above a limit (or truncate).

## API Work Needed

### Overview

We need an API that can safely serve:

1) A doc (frontmatter metadata + markdown body) given a doc path returned by search.
2) A “source file” (text) given a related file path (typically relative to repo root).
3) (Optional) Assets referenced by markdown (images) so rendered markdown can display them.

In all cases:

- Must prevent path traversal and accidental disclosure of arbitrary files on disk.
- Must clearly distinguish “doc paths under docs root” vs “repo files under repo root”.

### Proposed endpoints (v1)

#### 1) Get a doc (markdown + metadata)

`GET /api/v1/docs/get?path=<docRelPath>`

Input:
- `path`: doc-relative path under workspace docs root (same value as `SearchResult.Path`)

Output:
```json
{
  "path": "2026/01/03/.../design-doc/01-foo.md",
  "doc": {
    "title": "...",
    "ticket": "...",
    "status": "...",
    "docType": "...",
    "intent": "...",
    "topics": ["..."],
    "owners": ["..."],
    "lastUpdated": "2026-01-04T..."
  },
  "relatedFiles": [
    { "path": "internal/searchsvc/search.go", "note": "..." }
  ],
  "body": "## Markdown…",
  "stats": { "sizeBytes": 12345, "modTime": "..." }
}
```

Implementation notes:
- Read via `documents.ReadDocumentWithFrontmatter(absPath)` and return:
  - frontmatter fields
  - body (without YAML frontmatter delimiters)
- Ensure `path` resolves to a file under `workspace.Context().Root`.

#### 2) Get a source file (text) for “related code”

`GET /api/v1/files/get?path=<repoRelOrAbs>&root=repo|docs`

Input:
- `path`:
  - preferred: repo-relative path (e.g. `internal/httpapi/server.go`)
  - allow absolute path only if it’s inside the allowed roots
- `root`:
  - default `repo`
  - allow `docs` for “open markdown file in raw mode”

Output:
```json
{
  "path": "internal/httpapi/server.go",
  "root": "repo",
  "language": "go",
  "contentType": "text/plain; charset=utf-8",
  "truncated": false,
  "content": "package httpapi\n\n..."
}
```

Implementation notes:
- Allowed roots:
  - docs root (`workspace.Context().Root`)
  - repo root (`workspace.Context().RepoRoot`)
- Security / traversal:
  - reject `..` escapes by resolving to abs and checking it’s within the root
  - optionally resolve symlinks (`EvalSymlinks`) to avoid “symlink out of root” bypasses
- File classification:
  - detect binary by scanning for NUL bytes
  - enforce max size (example: 1–2 MB) or return truncated payload with `truncated=true`
- Language inference:
  - by extension (server-side) to keep UI simple; UI can still override.

#### 3) (Optional) Serve doc-relative assets (images)

This becomes necessary as soon as markdown docs include:

`![img](./img.png)` or `![img](assets/foo.png)`

Options:

A) JSON API (base64) — simple but inefficient.

`GET /api/v1/docs/asset?docPath=<docRelPath>&rel=<relPath>`

B) Raw bytes endpoint — preferred for browsers.

`GET /api/v1/docs/asset/<docDir>/<rel>` (mounted under `/api` so Vite proxy works)

Constraints:
- Must resolve relative to the doc’s directory under docs root.
- Must ensure the resolved path stays inside docs root.
- Must set correct `Content-Type` (`image/png`, etc).

Recommendation:
- Defer assets until the viewer is working for text/code/docs; add as a follow-up once a real doc demonstrates need.

## UI Work Needed

### Routes

Add two new routes to the SPA:

- `/doc?path=<docRelPath>` — document viewer page
- `/file?path=<filePath>&root=repo|docs` — source file viewer page

Both should:
- show a “Back” action via browser history (and optionally “Back to Search” link)
- keep search state intact via existing URL sync

### Doc viewer page (React)

Responsibilities:

- Fetch doc via `GET /api/v1/docs/get?path=...`.
- Render metadata panel and related files list per `single-doc.html`.
- Render markdown body:
  - headings, lists, inline code, code fences.
- Provide actions:
  - copy doc path
  - copy all markdown content
  - open related file → navigate to `/file?path=...`

### File viewer page (React)

Responsibilities:

- Fetch file via `GET /api/v1/files/get`.
- Render header (file path + copy + open-in-new-tab).
- Render content with syntax highlighting.
- Consider “line numbers” later (nice-to-have).

## Markdown rendering + syntax highlighting (library choices)

Bootstrap does not render markdown; we need a renderer.

### Option 1 (recommended): Client-side markdown rendering (React)

Libraries:
- `react-markdown` (markdown → React tree)
- `remark-gfm` (tables, strikethrough, task lists)
- `rehype-highlight` + `highlight.js` CSS theme for fenced code highlighting

Notes:
- Security: keep raw HTML disabled (default in `react-markdown`) to avoid XSS via markdown.
- Code fences:
  - read language from triple-backtick info string, pass as `language-<lang>` class
  - `rehype-highlight` applies highlight.js server-side in the React render pipeline.

Pros:
- Fast to implement in the existing UI stack.
- No server-side HTML rendering complexity.

Cons:
- Adds frontend bundle weight.
- Highlight.js is “best-effort”; code-block language coverage depends on included languages (default auto is OK for MVP).

### Option 2: Server-side rendering (Go) using Goldmark + Chroma

We already have `goldmark` and `chroma` in `go.mod` indirectly (not currently used).

Approach:
- Server endpoint returns rendered HTML + metadata.
- UI simply injects HTML into the page (requires strict sanitization/trust boundary).

Pros:
- Smaller JS; “canonical rendering” across clients.
- Chroma can be great for code.

Cons:
- HTML injection increases XSS risk unless sanitized.
- Harder to theme/iterate quickly in UI.

Recommendation:
- Start with Option 1 (client-side) and keep raw HTML disabled.

### Source file highlighting

For code viewer:
- Reuse highlight.js via:
  - `highlight.js` language inference by extension (server) or by client map
  - Render as `<pre><code class="language-go">…</code></pre>`
- For markdown “code fences”: same pipeline.

## Security & Safety Considerations (must-do)

Even for a local-first tool, file serving must guard against:

- path traversal (`../../etc/passwd`)
- symlink escape (a symlink inside the root pointing outside)
- huge files causing memory pressure
- binary blobs being rendered as “text”

Minimum guardrails:

- Only serve files that resolve within:
  - docs root (for docs + doc assets)
  - repo root (for related code)
- Enforce size limits and return `truncated` metadata.
- For binary files: return `415 unsupported_media_type` with a clear JSON error.

## Implementation Steps (MVP)

1) Backend
- Add `GET /api/v1/docs/get`:
  - resolve docRelPath → abs (docs root) safely
  - read frontmatter/body via `documents.ReadDocumentWithFrontmatter`
  - return JSON payload
- Add `GET /api/v1/files/get`:
  - safe path resolution into repo root
  - read with size limit + binary detection
  - return JSON payload including inferred language
- Add tests:
  - traversal rejects
  - binary rejects
  - size limit truncation behavior

2) Frontend
- Add RTK Query endpoints:
  - `getDoc`
  - `getFile`
- Add new routes `/doc` and `/file`.
- Add “Open doc” link from search results:
  - in result card or preview panel, navigate to `/doc?path=...`.
- Implement markdown rendering:
  - pick Option 1 (`react-markdown` + `remark-gfm` + `rehype-highlight`).
- Implement file viewer with highlight.

3) Follow-ups (after MVP)
- Doc assets (images) endpoint.
- “Open in editor” integration (out of scope for browser-only UI; would need protocol handler or local helper).
- Better navigation inside doc (TOC, anchor links, scroll spy).
