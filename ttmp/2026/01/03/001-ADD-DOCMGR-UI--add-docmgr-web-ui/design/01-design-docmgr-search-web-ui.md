---
Title: "Design: docmgr Search Web UI (React SPA)"
Ticket: 001-ADD-DOCMGR-UI
Status: draft
Topics:
    - docmgr
    - ux
    - web
    - search
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/api/serve.go
      Note: HTTP server entrypoint (serves `/api/v1/*` today; will also serve SPA in prod)
    - Path: internal/httpapi/server.go
      Note: REST API routes used by the UI (`/api/v1/search/*`, `/api/v1/index/refresh`, etc)
    - Path: internal/httpapi/index_manager.go
      Note: Build-on-startup + explicit refresh index lifecycle owned by server
    - Path: internal/searchsvc/search.go
      Note: Shared query engine used by CLI + HTTP (UI depends on its response shape)
    - Path: pkg/doc/docmgr-http-api.md
      Note: Existing user-facing HTTP API doc (keep in sync as UI evolves)
ExternalSources:
    - ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/01-claude-session-design.md
    - ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/test-design.html
Summary: "Design for a developer-centric search web UI that exercises docmgr’s local HTTP API (search docs, reverse lookup, file suggestions, cursor pagination, and explicit index refresh)."
LastUpdated: 2026-01-04T00:00:00Z
WhatFor: "Implement a production-embeddable SPA served by `docmgr api serve`, with a Vite dev loop and RTK Query integration."
WhenToUse: "Use as the build spec for the `001-ADD-DOCMGR-UI` implementation tasks and as a handoff doc for UI implementation/design review."
---

# Design: docmgr Search Web UI (React SPA)

## Executive Summary

Build a small, fast, developer-centric Search UI for `docmgr` that:

- Calls the existing `docmgr` HTTP API (`/api/v1/*`) for:
  - docs search (including reverse lookup),
  - file suggestions,
  - workspace/index status,
  - explicit index refresh.
- Uses cursor-based pagination (`cursor` → `nextCursor`) and a “Load more” UX.
- Provides three first-class modes: **Docs**, **Reverse Lookup**, **Files**.
- Implements keyboard-first navigation (`/`, ↑/↓, Enter, Esc, Alt+1/2/3, Cmd/Ctrl+R, `?`).
- Is a React/Vite SPA in `ui/` for development, and is embedded/served by Go for production (single binary), following the `go-web-frontend-embed` playbook.

This document is the implementation design spec and includes:

- all screens + widgets (from `sources/01-claude-session-design.md` and `sources/test-design.html`),
- ASCII wireframes,
- a concrete mapping from the YAML DSL → React/Redux/RTK Query structure (no code),
- a packaging/build plan for the Go binary.

## Goals

- Provide a “trustable” UI that makes doc search feel like an IDE feature:
  - strong information scent (ticket/type/status/topics/path/snippet/related files),
  - progressive disclosure (filters and diagnostics are collapsible),
  - explicit “index refresh” with last refresh timestamp and success toast.
- Reuse the backend query engine exactly by calling the HTTP API (no reimplementation of search semantics in JS).
- Keep the CLI and UI aligned by making the UI a thin client over stable JSON endpoints.
- Production: single Go binary serves both `/api/*` and `/` (SPA + assets), with a two-process dev loop (Vite + Go).

## Non-goals (v1)

- Full markdown document rendering and browsing within the UI.
  - v1 preview is “snippet + metadata + matched files” (as in the YAML DSL).
  - v2 can add a `/api/v1/docs/content?path=...` endpoint if desired.
- Authentication / multi-user deployment.
  - `docmgr` server is localhost-oriented; UI is a local developer tool.
- Perfect parity with CLI output formatting.
  - The UI consumes stable API fields; presentation can diverge.

## API Contract (UI-facing)

The UI relies on these existing endpoints (all relative to same origin in production):

- `GET /api/v1/healthz`
- `GET /api/v1/workspace/status`
- `POST /api/v1/index/refresh`
- `GET /api/v1/search/docs`
- `GET /api/v1/search/files`

### Response shapes (current server)

The UI should be implemented against explicit JSON shapes. For v1 we will extend the API to include the fields required by the UI (see below).

#### `GET /api/v1/search/docs` (target v1 shape)

```json
{
  "query": {
    "query": "websocket",
    "ticket": "",
    "topics": ["chat", "backend"],
    "docType": "",
    "status": "",
    "file": "",
    "dir": "",
    "externalSource": "",
    "since": "",
    "until": "",
    "createdSince": "",
    "updatedSince": "",
    "orderBy": "rank",
    "reverse": false,
    "pageSize": 200,
    "cursor": ""
  },
  "total": 123,
  "results": [
    {
      "ticket": "MEN-4242",
      "title": "Chat WebSocket Lifecycle",
      "docType": "reference",
      "status": "active",
      "topics": ["chat", "backend", "websocket"],
      "path": "2026/01/04/MEN-4242--normalize.../reference/01-chat-websocket-lifecycle.md",
      "lastUpdated": "2026-01-04T15:04:05Z",
      "snippet": "WebSocket connection lifecycle management…",
      "relatedFiles": [
        { "path": "backend/chat/ws/manager.go", "note": "WebSocket lifecycle mgmt" }
      ],
      "matchedFiles": ["backend/chat/ws/manager.go"],
      "matchedNotes": ["WebSocket lifecycle mgmt"]
    }
  ],
  "diagnostics": [],
  "nextCursor": "eyJ2IjoxLCJvIjoyMDB9"
}
```

Notes:

- `results[*].matchedFiles/matchedNotes` are populated primarily for reverse-lookup scenarios (when `file` is set); in normal docs search they can be empty.
- `diagnostics` are emitted by the backend; UI treats them as opaque, display-only items (no custom inference).
- `lastUpdated` should come from doc frontmatter (`LastUpdated`) when available.
- `relatedFiles` is the full doc related-files list (frontmatter `RelatedFiles`) and is returned for all docs (not only reverse lookup).

#### `GET /api/v1/search/files` (shape)

```json
{
  "total": 15,
  "results": [
    { "file": "backend/chat/ws/manager.go", "source": "RelatedFiles", "reason": "Referenced by docs …" }
  ]
}
```

### Search: Docs

`GET /api/v1/search/docs` supports:

- Query text (FTS5 MATCH query string): `query=...`
- Filters: `ticket`, `topics` (CSV), `docType`, `status`, `file`, `dir`, `externalSource`, date filters (`since`, `until`, `createdSince`, `updatedSince`)
- Mode switch: `reverse=true` (reverse lookup semantics)
- Ordering: `orderBy=path|last_updated|rank|...` (UI should default to `rank` for text search, `path` for reverse)
- Include flags: `includeArchived`, `includeScripts`, `includeControlDocs`
- Diagnostics: `includeDiagnostics=true` (default; used for warning badge + panel)
- Cursor pagination:
  - Request: `pageSize`, `cursor`
  - Response: `nextCursor`

UI must treat the response as the single source of truth for result ordering, ranking, and snippet behavior.

### “Open Full Doc →” (no file serving in v1)

v1 does not implement any file serving or markdown rendering endpoint.

UI behavior:

- Keep a visible “Copy path” action (button + shortcut).
- Optionally include “Open full doc” as a *disabled* or “coming soon” affordance, but do not route anywhere.
- If we want “open in editor” later, that’s a separate integration feature (not a web concern).

### Search: Files

`GET /api/v1/search/files` supports:

- `query=...` (text-ish hint)
- `ticket=...`, `topics=...` (CSV)
- `limit=...`

Results are presented as “Suggested files related to your context” with source/why fields (see wireframes).

### Index Refresh + Status

- `POST /api/v1/index/refresh`: triggers a rebuild and returns `indexedAt`, `docsIndexed`, `ftsAvailable`
- `GET /api/v1/workspace/status`: provides the same plus config/vocabulary paths for diagnostics and “trust signals”

UI should show:

- last refresh relative time (“2m ago”, “Just now”),
- a spinner / disabled refresh button while refreshing,
- a success toast on refresh completion.

## UX / Interaction Model

### Modes

1) **Docs**: general search over docs content + metadata.
2) **Reverse Lookup**: file-centric lookup “which docs reference this file/dir”.
3) **Files**: file suggestions based on query + context (ticket/topics).

Mode changes:

- set mode immediately, update URL, clear pagination cursor,
- keep query text and relevant filters where sensible (see “State & URL Sync”).

### Keyboard-first

From sources, v1 shortcuts:

- `/`: focus search input
- `↑/↓`: move focus across results
- `Enter`: select focused result (preview behavior optional in MVP)
- `Esc`: close preview (if open), otherwise clear search
- `Alt+1/2/3`: switch Docs / Reverse / Files
- `Cmd/Ctrl+R`: refresh index
- `Cmd/Ctrl+K`: copy selected result path
- `?`: open keyboard help overlay

### “Information Scent” on every result

Every result card must show:

- Title
- Ticket
- DocType (e.g. `reference`, `design-doc`, `playbook`, `index`)
- Status (`active`, `review`, `complete`, `draft`, …)
- Topics badges
- Snippet (search context)
- Path (monospace)
- Copy-to-clipboard affordance (hover-visible button)

Reverse lookup results additionally emphasize:

- matched file path(s),
- notes/explanations,
- diagnostics badge/panel when fallbacks are used (basename match, multiple candidates).

## Screens (with ASCII wireframes)

The following screens are required for v1 and must include all widgets shown in the sources.

### 1) Main Search View (empty / default)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr Search                                    [Refresh Index] 🔄 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [🔍 Search docs...                                              ] [Search] │
│  Hint:  / focus • ↑↓ navigate • Enter select • Esc clear • ? help          │
│                                                                             │
│  [●] Docs    [ ] Reverse Lookup    [ ] Files                                │
│                                                                             │
│  [Filters ▾]                                                                │
│   Ticket: [____________]  Topics: [________________]  Type: [All ▾]         │
│   Status: [All ▾]         Sort:   [Relevance ▾]                             │
│                                                                             │
│  Quick: [ ] Include archived  [✓] Include scripts  [✓] Control docs         │
│                                                                             │
│                          No search performed yet                            │
│                     Try searching for a topic or keyword                    │
│                   Or use filters to browse documentation                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

Widgets to implement (from `test-design.html` + wireframes):

- Header title
- Refresh button with relative time + “Refreshing…” disabled state
- Search input (supports Enter) + Search button
- Keyboard hint line
- Mode toggle (Docs/Reverse/Files)
- Collapsible filter row (ticket/topic/type/status/sort + clear)
- Quick toggles (includeArchived/includeScripts/includeControlDocs)
- Empty state (pre-search)

### 2) Search Results View (docs mode)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr Search                                    [Refresh Index] 🔄 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│  [🔍 websocket                                                ] [Search]    │
│  [●] Docs    [ ] Reverse Lookup    [ ] Files                               │
│                                                                             │
│  Active: [× websocket] [× chat] [× backend]          12 results   ⚠️ 2      │
│  ─────────────────────────────────────────────────────────────────────────  │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │📄 Chat WebSocket Lifecycle                        MEN-4242 • active    │ │
│  │   reference • chat, backend, websocket                                │ │
│  │   “WebSocket connection lifecycle management…”                         │ │
│  │   📂 2026/01/04/MEN-4242--.../reference/01-chat-websocket-...          │ │
│  │                                              [📋 Copy]                │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│  … more result cards …                                                     │
│                                                                             │
│                       [Load More Results] (cursor)                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

Widgets:

- Active filter chips (removable; include query itself as chip)
- Results count
- Diagnostics badge count (⚠️) when diagnostics exist
- Result cards with hover copy button (as in HTML mock)
- Loading spinner state (centered)
- “No results found” empty state (post-search)
- Cursor “Load More Results” button (only when `nextCursor != ""`)

### 3) Reverse Lookup Mode (file-centric)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr Search                                    [Refresh Index] 🔄 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│  [🔍 backend/chat/ws/manager.go                              ] [Search]     │
│  [ ] Docs    [●] Reverse Lookup    [ ] Files                               │
│                                                                             │
│  ┌─ Search by File Path ─────────────────────────────────────────────────┐ │
│  │ Enter full path, partial path, or just filename                        │ │
│  │ Examples: backend/api/register.go • register.go • ws/                   │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  Found 3 docs referencing: backend/chat/ws/manager.go        ⚠️  2          │
│  ─────────────────────────────────────────────────────────────────────────  │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │📄 Chat WebSocket Lifecycle                        MEN-4242 • active    │ │
│  │   Matched: backend/chat/ws/manager.go                                  │ │
│  │   Note: “WebSocket connection lifecycle management”                    │ │
│  │   📂 2026/01/04/MEN-4242--.../reference/01-chat-websocket-...          │ │
│  │                                              [📋 Copy]                │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│  … more results …                                                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

Reverse lookup UI specifics:

- The “Search input” is still present, but its placeholder and hint change.
- UI sets `reverse=true` for this mode.
- `file` and `dir` inputs are enabled; other filters remain available.
- Diagnostics badge is important; reverse lookup is where fallbacks occur.

### 4) Results + Preview Panel (split view)

From source wireframes (adapted; v1 preview is snippet-only):

```
┌──────────────────────────────────────┬──────────────────────────────────────┐
│ [🔍 websocket            ] [Search]  │ 📄 Chat WebSocket Lifecycle          │
│ [× websocket] [× chat]    12 results │ MEN-4242 • reference • active        │
│ ──────────────────────────────────── │ Topics: chat, backend, websocket     │
│  ┌──────────────────────────────────┐│ Updated: 2 days ago                  │
│  │✓ Chat WebSocket Lifecycle        ││ Path: 2026/01/04/MEN-4242--...        │
│  │  reference • MEN-4242            ││                                      │
│  └──────────────────────────────────┘│ Related Files:                        │
│  … results list …                    │  • backend/chat/ws/manager.go “…”     │
│                                     │  • backend/chat/ws/handler.go “…”     │
│        [Load More Results]           │                                      │
│                                     │ Preview (snippet):                    │
│                                     │ “WebSocket connection lifecycle…”     │
│                                     │                                      │
│                                     │ [📋 Copy Path] [Open Full Doc →]      │
└──────────────────────────────────────┴──────────────────────────────────────┘
```

Preview behaviors:

- Click a result card to open preview (keyboard selection behavior is optional in MVP).
- Esc closes preview.
- Preview shows metadata + snippet + related files list.
- “Open Full Doc →” is not implemented in v1 (no file serving); provide “Copy path” instead.

### 5) Diagnostics Panel (expanded)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ … header/search/mode …                                                      │
│ ⚠️  2 diagnostics  [Show Details ▾]                          5 results      │
│                                                                             │
│  ┌─ Diagnostics ──────────────────────────────────────────────────────────┐ │
│  │ ⚠️ Basename fallback used                                                │
│  │    Query: "manager.go"                                                   │
│  │    Matched: backend/chat/ws/manager.go                                   │
│  │    Suggestion: Use full path for more precise results                    │
│  │                                                                           │
│  │ ⚠️ Multiple files with same basename                                     │
│  │    Found: backend/chat/ws/manager.go, backend/api/manager.go             │
│  │    Suggestion: Add directory path to narrow results                      │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ … results …                                                                  │
└─────────────────────────────────────────────────────────────────────────────┘
```

Diagnostics requirements:

- Always display a badge count when diagnostics exist.
- Allow expanding/collapsing a diagnostic list panel.
- Each diagnostic item shows severity + message + suggestion (when available).

### 6) Files Suggestions Mode

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr Search                                    [Refresh Index] 🔄 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│  [🔍 Search for related files...                             ] [Search]    │
│  [ ] Docs    [ ] Reverse Lookup    [●] Files                               │
│                                                                             │
│  Context: Ticket [MEN-4242____]  Topics [chat, backend____]  Query [ws____] │
│                                                                             │
│  Suggested files related to your context                     15 files      │
│  ─────────────────────────────────────────────────────────────────────────  │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │📁 backend/chat/ws/manager.go                                            │ │
│  │   Source: RelatedFiles                                                  │ │
│  │   Referenced in: 3 docs (MEN-4242, MEN-5100, MEN-4300)                  │ │
│  │   “WebSocket connection lifecycle management”                           │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│  … more file cards …                                                        │
│                                                                             │
│                          [Load More Files] (optional v2)                    │
└─────────────────────────────────────────────────────────────────────────────┘
```

Files mode requirements:

- Uses `/api/v1/search/files`.
- Shows “file card” results with:
  - file path,
  - source/reason (“RelatedFiles”),
  - referenced-in docs count and a short explanation (when available).
- v1 can be `limit`-based only (no cursor); keep “Load more” as a future extension if backend adds cursor pagination here.

### 7) Mobile / Compact View (stacked)

This is a responsive layout variant, not a separate route:

```
┌───────────────────────────────┐
│ docmgr Search         🔄 2m   │
├───────────────────────────────┤
│ [🔍 Search...  ] [≡]          │
│ [●] Docs  [ ] Reverse  [ ] 📁 │
│ [× websocket] [× chat] ⚠️ 2   │
│                   12 results  │
│ ───────────────────────────── │
│ ┌───────────────────────────┐ │
│ │📄 Chat WebSocket...       │ │
│ │ reference • MEN-4242      │ │
│ │ chat, backend, websocket  │ │
│ │ Updated 2d        [View→] │ │
│ └───────────────────────────┘ │
│        [Load More Results]    │
└───────────────────────────────┘
```

Mobile requirements:

- Filters collapse behind a “hamburger”/toggle.
- Preview becomes a full-screen modal/page (“View →”).
- Copy path is still available (button or menu).

### 8) Keyboard Shortcuts Overlay

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Keyboard Shortcuts                          [×]    │
├─────────────────────────────────────────────────────────────────────────────┤
│ Navigation: / focus • ↑↓ navigate • Enter preview • Cmd/Ctrl+↵ new tab       │
│ Modes:      Alt+1 docs • Alt+2 reverse • Alt+3 files                         │
│ Actions:    Cmd/Ctrl+R refresh • Cmd/Ctrl+K copy path • ? help               │
└─────────────────────────────────────────────────────────────────────────────┘
```

Requirements:

- Open/close with `?` and Esc.
- Static content table (no API dependency).

## State, Data Flow, and URL Sync (Mapping from YAML DSL)

The source YAML DSL is a “minimal architecture view”. This section describes how it maps to real React/Redux code and file layout (still no code).

### Proposed frontend structure

`ui/` (new):

- `ui/src/app/`
  - `store` (configureStore)
  - `hooks` (typed `useAppDispatch/useAppSelector`)
- `ui/src/services/`
  - `docmgrApi` (RTK Query slice + generated hooks)
- `ui/src/features/`
  - `search/` (slice, selectors, helpers, components)
  - `workspace/` (workspace/status slice + view)
  - `preview/` (selected result + open/close state)
  - `ui/` (global UI toggles: filters open, diagnostics open, keyboard help open)
- `ui/src/components/` (reusable UI primitives: chips, badges, cards, split-pane, toast)
- `ui/src/routes/` (route-level pages)

### Store slices (YAML → concrete intent)

From YAML: `search`, `workspace`, `preview`, `ui`.

Guiding rule:

- Slices own UI state and durable “intent” (query, filters, mode, cursor).
- RTK Query owns server-state (results payloads, status/refresh responses) and caches by request args.

Concrete approach:

- `searchSlice` owns:
  - `mode`, `query`, `filters`, `cursor`, `activeChips`, and `hasSearched`.
  - It does *not* own actual result rows (those come from RTK Query), except for “accumulated list when loading more”. See Pagination below.
- `workspaceSlice` owns:
  - `health` summary and last-known workspace status payload (for displaying root/config paths and fts availability).
- `previewSlice` owns:
  - `selectedResultId` (or the full selected object, depending on implementation tradeoffs),
  - `open` flag (split pane vs closed).
- `uiSlice` owns:
  - toggles: `showFilters`, `showDiagnostics`, `showKeyboardHelp`,
  - `theme` (optional; v1 can be light-only).

### RTK Query endpoints (YAML → API calls)

Create one API slice with base URL = same origin:

- `healthCheck` → `GET /api/v1/healthz`
- `getWorkspaceStatus` → `GET /api/v1/workspace/status` (optionally poll)
- `refreshIndex` → `POST /api/v1/index/refresh`
- `searchDocs` → `GET /api/v1/search/docs`
- `searchFiles` → `GET /api/v1/search/files`

Key requirements:

- Use tag invalidation so refresh triggers status + current searches to refetch.
- Prefer request args object types:
  - `SearchDocsArgs` maps 1:1 to query params + `pageSize/cursor`.
  - `SearchFilesArgs` maps to `query/ticket/topics/limit`.

### Cursor pagination (docs)

Cursor rules:

- First page: omit `cursor` or set `cursor=""`.
- Next pages: use `nextCursor` returned by server.
- UI stores the “current cursor” as the next page cursor, not the current offset.

Two acceptable UI patterns (pick one; both are compatible with the YAML DSL):

1) **Append in slice (recommended for v1)**
   - Store accumulated `results[]` in `searchSlice`.
   - First search replaces results; load-more appends.
   - RTK Query is used as a transport layer, but slice owns the “render list”.

2) **Cache pages in RTK Query (more advanced)**
   - Keep an array of cursors and request each page as a separate cache key.
   - Selector flattens pages for rendering.
   - More moving parts; defer unless needed.

### URL sync

UI should support shareable URLs:

- Route: `/` (and optionally `/search` as an alias)
- Query params:
  - `mode=docs|reverse|files`
  - `q=...` (search query)
  - filters as separate params: `ticket`, `topics`, `docType`, `status`, `file`, `dir`, `orderBy`, toggles (`archived`, `scripts`, `control`)

Rules:

- On page load: parse URL → dispatch `setMode/setQuery/setFilters`.
- On user change: update URL (debounced) to match state.
- Cursor must NOT be encoded in URL by default (URLs should represent intent, not pagination position).

### Component mapping (YAML widgets → React components)

Below is a direct mapping of the YAML component tree to concrete React component responsibilities.

- `Layout`
  - Owns responsive structure and high-level panels.
- `Header`
  - Renders title + refresh widget.
  - Pulls from `workspace/status` + `index/refresh` for “2m ago”.
- `RefreshButton`
  - Calls `refreshIndex` mutation.
  - Shows spinner/disabled state.
  - Triggers a toast (“Index refreshed successfully!”) on success.
- `SearchBar`
  - Controlled input bound to `search.query`.
  - Enter triggers “execute search” in current mode.
  - `/` focuses (via keyboard provider).
  - Optional syntax tooltip (“FTS5 syntax” help).
- `ModeToggle`
  - Updates `search.mode`.
  - Resets cursor and accumulated results.
  - Updates placeholder text/hints according to mode.
- `FilterBar` + `QuickToggles`
  - Binds to `search.filters` state fields.
  - Shows “Clear” and controls collapsible visibility (`ui.showFilters`).
- `ActiveFilterChips`
  - Derived from `search.query` + `search.filters`.
  - Clicking chip “x” removes the associated filter and re-runs (if already searched).
- `ResultsArea`
  - Switch by mode:
    - Docs: renders `DocsResultsList`
    - Reverse: renders `ReverseResultsList` (same card UI but with “Matched:” and note emphasis)
    - Files: renders `FileSuggestionsList`
- `DocResultCard`
  - Renders “information scent” fields + copy button.
  - Click selects (and optionally opens preview).
- `DiagnosticsBadge` + `DiagnosticsPanel`
  - Badge shows count; panel toggles open/closed (progressive disclosure).
  - Must render server-provided diagnostics exactly (no heuristics).
- `PreviewPanel`
  - Right-side split pane on desktop; modal on mobile.
  - Shows metadata + snippet + related files list.
  - “Copy Path” action.
- `KeyboardHelpModal`
  - Static modal with shortcut table.

## Visual / Component Guidelines (from sources)

From `test-design.html`:

- Result cards: white background, hover highlight (border + subtle shadow)
- Status badge variants by status:
  - `active` → primary
  - `review` → warning
  - `complete` → success
  - `draft` → secondary
- Topic badges: small, low-emphasis (secondary)
- Copy button:
  - hidden by default; appears on hover
  - uses clipboard API; show toast confirmation
- Empty states:
  - pre-search: “Search docmgr documentation”
  - post-search: “No results found”
- Loading state: centered spinner

From wireframes:

- Diagnostics are a trust signal (show prominently when present).
- Filters can be collapsed by default (“Progressive Disclosure”).
- Mobile view prioritizes search + list; preview becomes “View →”.

## Build & Packaging Plan (Go + SPA) — `go-web-frontend-embed`

This section follows the `go-web-frontend-embed` skill:

### Dev topology (two-process loop)

- Vite dev server: `http://localhost:3000`
- Go API server: `http://localhost:3001`
- Vite proxies:
  - `/api/*` → `http://localhost:3001`
  - (optional) `/ws` → `ws://localhost:3001` (not needed for v1 UI)

Outcome:

- No CORS configuration required in dev.
- UI calls `/api/v1/...` as a relative path in both dev and prod.

### Production topology (single binary)

`docmgr api serve` serves:

- `/api/v1/*` = JSON API (existing)
- `/assets/*` and `/` = embedded SPA assets and SPA fallback

SPA handler invariants:

- Must never shadow `/api` (and `/ws` if later added).
- Must serve real files if present (`/assets/...`).
- Otherwise must serve `index.html` (SPA fallback).

### Directory layout + build bridge (`go generate`)

Planned layout:

- Frontend: `ui/`
- Vite build output: `ui/dist/public/`
- Canonical Go static dir: `internal/web/embed/public/`
- Go embed build tag: `-tags embed`

Build process:

1) `go generate ./internal/web`:
   - runs `pnpm -C ui run build`
   - copies `ui/dist/public/*` → `internal/web/embed/public/`
2) `go build -tags embed ./cmd/docmgr` (or normal repo build)

Makefile entry points (planned):

- `make dev-backend` (Go on `:3001`)
- `make dev-frontend` (Vite on `:3000`)
- `make ui-build` / `make ui-check` (optional)
- `make build` (generate + embed build)

CI requirements:

- install node deps before `go generate`,
- run `go generate` before `go test` if embed assets are referenced by tests.

## Open Questions (to settle during implementation)

- Enter/copy-path semantics are explicitly deferred in v1 (click + copy button are sufficient; keyboard support can be expanded later).
- Default ordering:
  - Proposal: docs mode defaults to `orderBy=rank` when query is non-empty; otherwise `path`.
- How to represent multi-topic selection:
  - v1 uses a multi-select *without suggestions*: selected topic tokens + an “Add topic” input (no autocomplete list).
