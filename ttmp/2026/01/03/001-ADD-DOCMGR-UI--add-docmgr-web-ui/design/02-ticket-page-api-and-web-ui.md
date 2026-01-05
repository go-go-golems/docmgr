---
Title: Ticket Page API + Web UI Design
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
  - docmgr
  - ui
  - http
  - api
  - tickets
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
  - Path: internal/httpapi/server.go
    Note: Where to mount new `/api/v1/tickets/*` endpoints
  - Path: internal/httpapi/index_manager.go
    Note: Index snapshot + workspace discovery for API handlers
  - Path: internal/searchsvc/search.go
    Note: Existing `ticket` filtering + doc results shaping
  - Path: pkg/commands/tasks.go
    Note: Current tasks.md parsing + mutation logic (to extract into a reusable package)
  - Path: pkg/commands/ticket_graph.go
    Note: Mermaid ticket graph generation (to reuse via API)
  - Path: ui/src/App.tsx
    Note: Add new route for ticket page
  - Path: ui/src/services/docmgrApi.ts
    Note: Add ticket endpoints to RTK Query
  - Path: ui/src/features/search/SearchPage.tsx
    Note: Link ticket badge / path to open ticket page
  - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/topic-page.md
    Note: UI design reference (ASCII specs for tabs + widgets)
ExternalSources: []
Summary: "Design the HTTP API and React UI for a ticket-specific page (overview/docs/tasks/graph/changelog) based on sources/topic-page.md."
LastUpdated: 2026-01-05T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# Ticket Page API + Web UI Design

This document designs:

1) A **ticket-specific page** in the docmgr web UI (tabs: Overview / Documents / Tasks / Graph / Changelog), and
2) A **ticket page API surface** that the UI can call without shelling out to the CLI.

Primary UI reference:
- `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/topic-page.md`

## 0. Goals / Non-goals

### Goals
- Provide a “single place” to review a ticket:
  - ticket metadata and quick stats
  - key documents
  - tasks progress + ability to check/uncheck tasks
  - Mermaid graph of docs ↔ related files
  - changelog browsing
- Keep navigation fluid:
  - easy back/forth between Search → Ticket → Doc → File
  - deep-linkable URLs (so links can be shared)
- Reuse existing primitives where possible:
  - `search/docs` for listing docs in a ticket
  - `docs/get` and `files/get` for viewer surfaces
  - `ticket graph` + `tasks` parsing logic (but extracted into reusable packages, not invoked as CLI)

### Non-goals (v1)
- Full in-browser editing of doc bodies (markdown editor).
- Full task editing UI parity (move/reorder sections, free-form edits) — we’ll design endpoints, but UI can ship read-only first.
- Auth/multi-user (assume local).

## 1. UX / Information Architecture

### 1.1. Routes

Proposed routes:
- `/ticket/:ticket` — ticket page (tabs)
- `/doc?path=...` — doc viewer (already exists)
- `/file?root=repo|docs&path=...` — file viewer (already exists)

Ticket page query params:
- `tab=overview|documents|tasks|graph|changelog` (default `overview`)
- `doc=<docRelPath>` optional selection within Documents tab (for a right-side preview drawer on desktop)
- `file=<repoRelPath>` optional selection within Graph tab side panel (optional)

### 1.2. Navigation expectations

From Search results:
- Clicking the **ticket badge** opens `/ticket/<ticket>` (new).
- Viewing a document (Open doc) stays as-is (`/doc?path=...`), but should include a “Ticket” link back to `/ticket/<ticket>`.

Within Ticket page:
- Documents tab uses the same doc viewer route `/doc?path=...` for “View →”.
- Related files open `/file?...` (already).

### 1.3. Tabs and widgets (from `topic-page.md`)

We should implement at least these widgets per tab:

**Overview**
- Header: ticket id + title + Actions menu
- Metadata panel: Status / Created / Updated / Topics / Owners / Intent
- Quick stats: docs count, tasks done/total, related files count, progress bar
- Summary and Current status sections (rendered markdown from index.md)
- “Key documents” (top few docs by importance / doc type)
- “Active tasks” (top N unchecked tasks with quick-check)
- “Related files” list with copy/open

**Documents**
- Group by doc type (design-doc / reference / analysis / sources / index / etc)
- Each card shows: title, short summary, updated time, related files count, and “View →”
- Optional: filter/search within the ticket’s docs (client-side filter + server pagination as needed)

**Tasks**
- Group tasks by section headings in tasks.md
- Show progress
- Allow check/uncheck (fast)
- Allow add task (append to a section)

**Graph**
- Render Mermaid graph of docs ↔ related files
- Side panel: selected node details, “Open doc”, “Open file”, “Copy path”
- Graph insights (hub docs, most referenced files)

**Changelog**
- List entries (date grouping)
- Clicking entry opens markdown viewer for changelog.md (or a structured view if we parse it)

## 1.4. ASCII Screenshots (Full Pages)

These are “full page” ASCII screenshots meant to be used as implementation targets and for designer handoff. They are intentionally verbose and include expected widgets, layout, and navigation affordances.

Legend:
- `[...]` = a button or control
- `(...)` = secondary text
- `▸` = navigates to another route
- “Preview” panels are optional on mobile (use a modal instead).

### 1.4.1. Desktop: Ticket Page Shell (All Tabs)

```
┌───────────────────────────────────────────────────────────────────────────────────────────────┐
│ docmgr Web                                                                                     │
│ [Search]  [Ticket]                                                                             │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ Ticket: 001-ADD-DOCMGR-UI — Add docmgr Web UI                         [Refresh Index] [⋯]     │
│ Path: 2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui                  Updated: 2026-01-05     │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ [Overview] [Documents] [Tasks] [Graph] [Changelog]                                           │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ (Tab content below; some tabs optionally show a right-side Preview/Details panel on desktop)  │
└───────────────────────────────────────────────────────────────────────────────────────────────┘
```

### 1.4.2. Desktop: Overview Tab

Route: `/ticket/001-ADD-DOCMGR-UI?tab=overview`

```
┌───────────────────────────────────────────────────────────────────────────────────────────────┐
│ Ticket: 001-ADD-DOCMGR-UI — Add docmgr Web UI                         [Refresh Index] [⋯]     │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ [Overview] [Documents] [Tasks] [Graph] [Changelog]                                           │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ ┌─ Metadata ──────────────────────────────────────┐  ┌─ Quick Stats ─────────────────────────┐ │
│ │ Status: active         Intent: long-term        │  │ Docs:  6   Tasks: 25/27   Files: 17  │ │
│ │ Topics: [docmgr] [ui] [api] [http]              │  │ Progress: [███████████████░░] 93%     │ │
│ │ Owners: (none)                                  │  │ Indexed: 10m ago  Workspace: ttmp      │ │
│ │ Created: 2026-01-03    Updated: 2026-01-05      │  └───────────────────────────────────────┘ │
│ └─────────────────────────────────────────────────┘                                            │
│                                                                                               │
│ ┌─ Summary (from index.md) ─────────────────────────────────────────────────────────────────┐ │
│ │ Build a React-based web UI for docmgr that provides search and document viewing...         │ │
│ │                                                                                           │ │
│ │ (render markdown)                                                                          │ │
│ └───────────────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                               │
│ ┌─ Current Status (from index.md) ─────────────────────────────────────────────────────────┐ │
│ │ ✅ Search UI works                                                                          │ │
│ │ ✅ HTTP API runs                                                                            │ │
│ │ 🚧 Ticket page pending                                                                      │ │
│ └───────────────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                               │
│ ┌─ Key Documents ────────────────────────────────────────┐  ┌─ Active Tasks ─────────────────┐ │
│ │ design/01-design-docmgr-search-web-ui.md        [Open]  │  │ [ ] API: /tickets/get           │ │
│ │ design/02-ticket-page-api-and-web-ui.md         [Open]  │  │ [ ] UI: /ticket/:ticket route   │ │
│ │ analysis/01-doc-serving-api-and-document-viewer.md[Open]│  │ [ ] Graph tab rendering          │ │
│ │ …                                                       │  │ [View all ▸]                     │ │
│ └────────────────────────────────────────────────────────┘  └────────────────────────────────┘ │
│                                                                                               │
│ ┌─ Related Files (top) ─────────────────────────────────────────────────────────────────────┐ │
│ │ internal/httpapi/server.go                         [Open ▸] [Copy]                         │ │
│ │ ui/src/services/docmgrApi.ts                       [Open ▸] [Copy]                         │ │
│ │ …                                                                                          │ │
│ └───────────────────────────────────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────────────────────────────────┘
```

### 1.4.3. Desktop: Documents Tab (Grouped + Preview)

Route: `/ticket/001-ADD-DOCMGR-UI?tab=documents`

Optional deep link with selection preview:
- `/ticket/001-ADD-DOCMGR-UI?tab=documents&doc=2026/01/03/.../design/01-design-docmgr-search-web-ui.md`

```
┌───────────────────────────────────────────────────────────────────────────────────────────────┐
│ Ticket: 001-ADD-DOCMGR-UI — Add docmgr Web UI                         [Refresh Index] [⋯]     │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ [Overview] [Documents] [Tasks] [Graph] [Changelog]                                           │
├───────────────────────────────────────────────────────────────┬───────────────────────────────┤
│ Documents (6)                                     [Filter ▾]  │ Preview                        │
│ ────────────────────────────────────────────────────────────── │ ───────────────────────────── │
│ DESIGN-DOC (2)                                                  │ Title: docmgr Search Web UI   │
│  ┌───────────────────────────────────────────────────────────┐ │ DocType: design-doc           │
│  │ docmgr Search Web UI (React SPA)                           │ │ Status: active  Updated: 1d   │
│  │ (short summary…)                                           │ │ Path: …/design/01-…md         │
│  │ Updated: 2026-01-04 • Related files: 17                    │ │ [Open doc ▸] [Open folder ▸]  │
│  │ …/design/01-design-docmgr-search-web-ui.md           [Open] │ │                               │
│  └───────────────────────────────────────────────────────────┘ │ ┌─ Related Files ───────────┐ │
│  ┌───────────────────────────────────────────────────────────┐ │ │ internal/httpapi/server.go │ │
│  │ Ticket Page API + Web UI Design                             │ │ │ ui/src/App.tsx             │ │
│  │ …/design/02-ticket-page-api-and-web-ui.md             [Open] │ │ │ …                          │ │
│  └───────────────────────────────────────────────────────────┘ │ └────────────────────────────┘ │
│                                                                │                               │
│ ANALYSIS (1)                                                    │ ┌─ Excerpt ─────────────────┐ │
│  ┌───────────────────────────────────────────────────────────┐ │ │ (render snippet/markdown) │ │
│  │ Doc Serving API and Document Viewer UI                     │ │ │ …                          │ │
│  │ …/analysis/01-doc-serving-api-and-document-viewer-ui.md[Open]│ └────────────────────────────┘ │
│  └───────────────────────────────────────────────────────────┘ │                               │
│                                                                │ [Close preview]                │
├───────────────────────────────────────────────────────────────┴───────────────────────────────┤
│ Tips: Ctrl/Cmd+Click Open to new tab • Click row to preview • Enter opens selected            │
└───────────────────────────────────────────────────────────────────────────────────────────────┘
```

### 1.4.4. Desktop: Tasks Tab (Checklist + Mutations)

Route: `/ticket/001-ADD-DOCMGR-UI?tab=tasks`

```
┌───────────────────────────────────────────────────────────────────────────────────────────────┐
│ Ticket: 001-ADD-DOCMGR-UI — Add docmgr Web UI                         [Refresh Index] [⋯]     │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ [Overview] [Documents] [Tasks] [Graph] [Changelog]                                           │
├───────────────────────────────────────────────────────────────┬───────────────────────────────┤
│ Tasks: 25/27 done                                              │ Task Details (optional)        │
│ [███████████████████████░░] 93%                                 │ ───────────────────────────── │
│                                                                │ Selected: (none)               │
│ ┌─ TODO ─────────────────────────────────────────────────────┐ │                               │
│ │ [ ] API: resolve ticket dir helper                          │ │ (Optional right panel for      │
│ │ [ ] API: GET /tickets/get                                   │ │  multi-line tasks or references)│
│ │ [ ] UI: route /ticket/:ticket                               │ │                               │
│ │ [ ] UI: Documents tab grouping                               │ │                               │
│ │ [ ] UI: Graph tab rendering                                  │ │                               │
│ │                                                             │ │                               │
│ │ Add: [____________________________________] [Add to TODO]   │ │                               │
│ └─────────────────────────────────────────────────────────────┘ │                               │
│                                                                │                               │
│ ┌─ Done ─────────────────────────────────────────────────────┐ │                               │
│ │ [x] Persist selected doc in URL                              │ │                               │
│ │ [x] Render snippets as markdown                              │ │                               │
│ └─────────────────────────────────────────────────────────────┘ │                               │
├───────────────────────────────────────────────────────────────┴───────────────────────────────┤
│ Behavior: checking toggles immediately; failures show inline error and revert                 │
└───────────────────────────────────────────────────────────────────────────────────────────────┘
```

### 1.4.5. Desktop: Graph Tab (Mermaid + Side Panel)

Route: `/ticket/001-ADD-DOCMGR-UI?tab=graph`

Optional selection deep link:
- `/ticket/001-ADD-DOCMGR-UI?tab=graph&doc=...` or `/ticket/...&file=internal/httpapi/server.go`

```
┌───────────────────────────────────────────────────────────────────────────────────────────────┐
│ Ticket: 001-ADD-DOCMGR-UI — Add docmgr Web UI                         [Refresh Index] [⋯]     │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ [Overview] [Documents] [Tasks] [Graph] [Changelog]                                           │
├───────────────────────────────────────────────────────────────┬───────────────────────────────┤
│ Graph                                                        │ Selected Node                   │
│ [Direction: TD ▾] [Rebuild Graph] [Fit] [Zoom -] [Zoom +]     │ ───────────────────────────── │
│                                                                │ Kind: doc                       │
│ ┌───────────────────────────────────────────────────────────┐ │ Title: Ticket Page API + UI     │
│ │                                                           │ │ Path: …/design/02-ticket-…md    │
│ │  (Mermaid canvas - interactive)                            │ │ Updated: 2026-01-05             │
│ │                                                           │ │                                 │
│ │    ┌──────────────┐         ┌───────────────┐             │ │ Outgoing:                        │
│ │    │ index.md      │ ─────▶  │ design/01…    │             │ │  → internal/httpapi/server.go     │
│ │    └──────┬───────┘         └───────┬───────┘             │ │  → ui/src/App.tsx                 │
│ │           │                          │                      │ │                                 │
│ │           ▼                          ▼                      │ │ Related Files:                    │
│ │    ┌──────────────┐         ┌───────────────┐              │ │  • internal/httpapi/server.go     │
│ │    │ analysis/01… │         │ ui/src/App.tsx│              │ │  • ui/src/services/docmgrApi.ts   │
│ │    └──────────────┘         └───────────────┘              │ │  • …                              │
│ │                                                           │ │                                 │
│ └───────────────────────────────────────────────────────────┘ │ [Open doc ▸] [Open file ▸] [Copy]│
│                                                                │                                 │
├───────────────────────────────────────────────────────────────┴───────────────────────────────┤
│ Tips: click node to select • ctrl/cmd+click Open opens new tab • copy path copies relative    │
└───────────────────────────────────────────────────────────────────────────────────────────────┘
```

### 1.4.6. Desktop: Changelog Tab (Entries + Preview)

Route: `/ticket/001-ADD-DOCMGR-UI?tab=changelog`

Option A (v1): open the doc viewer for `changelog.md` (simple).
Option B (polish): parse entries + preview in this tab.

```
┌───────────────────────────────────────────────────────────────────────────────────────────────┐
│ Ticket: 001-ADD-DOCMGR-UI — Add docmgr Web UI                         [Refresh Index] [⋯]     │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│ [Overview] [Documents] [Tasks] [Graph] [Changelog]                                           │
├───────────────────────────────────────────────────────────────┬───────────────────────────────┤
│ Changelog                                                     │ Preview                         │
│ [Open changelog.md ▸]                                         │ ─────────────────────────────  │
│                                                                │ 2026-01-05                      │
│ ┌─ 2026-01-05 ───────────────────────────────────────────────┐ │ - Added docs/get endpoint       │
│ │ • Add doc serving endpoints                                 │ │ - Added file viewer route       │
│ │ • Fix snippet highlight recursion                            │ │                                 │
│ └─────────────────────────────────────────────────────────────┘ │ (render markdown excerpt)      │
│                                                                │                                 │
│ ┌─ 2026-01-04 ───────────────────────────────────────────────┐ │                                 │
│ │ • Add embed packaging                                        │ │                                 │
│ │ • Add Vite dev proxy                                         │ │                                 │
│ └─────────────────────────────────────────────────────────────┘ │                                 │
│                                                                │ [View full ▸]                    │
└────────────────────────────────────────────────────────────────┴───────────────────────────────┘
```

### 1.4.7. Mobile: Ticket Page (Stacked + Modals)

Routes match desktop; layout changes:
- Tabs become a horizontally scrollable row.
- Right-side preview panels become modal sheets.

```
┌───────────────────────────────────────┐
│ 001-ADD-DOCMGR-UI        [⋯]          │
│ Add docmgr Web UI                     │
├───────────────────────────────────────┤
│ [Overview][Docs][Tasks][Graph][More▾] │
├───────────────────────────────────────┤
│ (Overview tab)                         │
│ Status: active   Progress: 93%         │
│ Docs: 6  Tasks: 25/27  Files: 17       │
│                                       │
│ Summary (markdown)                    │
│ …                                     │
│                                       │
│ Key docs                              │
│ - design/01…                 [Open▸]  │
│ - design/02…                 [Open▸]  │
│                                       │
│ Active tasks                           │
│ - [ ] API: /tickets/get                │
│ - [ ] UI: /ticket route                │
└───────────────────────────────────────┘

(Documents tab: tapping a doc opens a modal preview with [Open doc ▸] button)
```

## 2. Backend Model: What counts as a “ticket”

Ticket root is a directory under docs root with:
- `index.md` (has frontmatter; canonical ticket metadata)
- `tasks.md` (markdown checkboxes; usually *no* frontmatter)
- `changelog.md` (markdown; usually *no* frontmatter)
- optional folders: `design/`, `analysis/`, `reference/`, `sources/`, etc.

Canonical “ticket metadata” source:
- Prefer index.md frontmatter (`DocType: index`, Ticket: <id>).

Created date source:
- Derive from the ticket directory path: `ttmp/YYYY/MM/DD/TICKET--slug/` (best-effort).

## 3. API Design (v1)

Base path: `/api/v1`

### 3.1. Ticket summary

`GET /api/v1/tickets/get?ticket=<TICKET-ID>`

Returns enough to render the Overview header + metadata + quick stats.

Response shape:

```json
{
  "ticket": "001-ADD-DOCMGR-UI",
  "title": "Add docmgr Web UI",
  "status": "active",
  "intent": "long-term",
  "owners": [],
  "topics": ["docmgr", "ux", "tooling", "web"],
  "createdAt": "2026-01-03",
  "updatedAt": "2026-01-05T00:00:00Z",
  "ticketDir": "2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui",
  "indexPath": "2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/index.md",
  "stats": {
    "docsTotal": 6,
    "tasksTotal": 27,
    "tasksDone": 25,
    "relatedFilesTotal": 17
  }
}
```

Implementation notes:
- Resolve `ticketDir` via workspace + index doc query (same approach as tasks.go).
- `docsTotal`: query index (scope: ticket, includeControlDocs=true).
- `tasksTotal/tasksDone`: parse tasks.md (see §3.3).
- `relatedFilesTotal`: aggregate unique repo file keys referenced by docs in the ticket (likely union of `RelatedFiles` across docs).

### 3.2. Ticket docs list

`GET /api/v1/tickets/docs?ticket=<TICKET-ID>&pageSize=200&cursor=...&groupBy=docType`

The server can implement this by delegating to the existing search engine:
- `workspace.QueryDocs` with scope ticket + `OrderBy=last_updated|path`
- include a minimal body? (not needed; viewer uses `docs/get`)

Response shape:

```json
{
  "ticket": "001-ADD-DOCMGR-UI",
  "total": 6,
  "results": [
    {
      "path": "2026/01/03/.../design/01-design-docmgr-search-web-ui.md",
      "title": "docmgr Search Web UI (React SPA)",
      "docType": "design-doc",
      "status": "active",
      "topics": ["docmgr", "ui"],
      "lastUpdated": "2026-01-04T19:22:44-05:00",
      "relatedFilesCount": 17
    }
  ],
  "nextCursor": ""
}
```

Notes:
- If `groupBy=docType`, the UI can still group client-side; server grouping is optional.
- The design intentionally avoids duplicating `/api/v1/search/docs` unless we need a different shape.

### 3.3. Ticket tasks

`GET /api/v1/tickets/tasks?ticket=<TICKET-ID>`

Return a structured representation suitable for a Tasks tab:

```json
{
  "ticket": "001-ADD-DOCMGR-UI",
  "tasksPath": "2026/01/03/.../tasks.md",
  "stats": { "total": 27, "done": 25 },
  "sections": [
    {
      "title": "TODO",
      "items": [{ "id": 1, "checked": true, "text": "Write UI implementation plan..." }]
    },
    {
      "title": "Done",
      "items": [{ "id": 49, "checked": true, "text": "Trace doc search implementation..." }]
    }
  ]
}
```

To support interaction:

`POST /api/v1/tickets/tasks/check`

```json
{ "ticket": "001-ADD-DOCMGR-UI", "ids": [20, 21], "checked": true }
```

`POST /api/v1/tickets/tasks/add`

```json
{ "ticket": "001-ADD-DOCMGR-UI", "section": "TODO", "text": "New task text" }
```

Implementation notes:
- Extract parsing/mutation logic from `pkg/commands/tasks.go` into a reusable internal package (e.g. `internal/tasks`).
- Preserve task IDs as the “appearance order” indices (as today).
- Preserve file formatting as best-effort (don’t rewrite unrelated parts of tasks.md).

### 3.4. Ticket graph

`GET /api/v1/tickets/graph?ticket=<TICKET-ID>&format=markdown|mermaid&direction=TD|LR`

Response shape:

```json
{
  "ticket": "001-ADD-DOCMGR-UI",
  "format": "mermaid",
  "mermaid": "graph TD\n...",
  "stats": { "nodes": 11, "edges": 23 },
  "insights": {
    "mostReferencedFiles": [{ "path": "internal/httpapi/server.go", "count": 4 }]
  }
}
```

Implementation notes:
- Reuse the existing graph builder in `pkg/commands/ticket_graph.go`, but extract core logic into a package callable by HTTP without glazed.
- Keep Mermaid IDs stable and safe (already tested in `pkg/commands/ticket_graph_test.go`).

### 3.5. Ticket changelog

Option A (simplest v1): serve as a document via the existing doc endpoint:
- UI uses `GET /api/v1/docs/get?path=<ticketDir>/changelog.md` and renders markdown in the doc viewer.

Option B (nicer): parse changelog into structured entries:
- `GET /api/v1/tickets/changelog?ticket=...`

This is optional; start with Option A unless the UX suffers.

## 4. UI Design: Ticket Page (React)

### 4.1. Top-level layout

- Left column (main): tab content
- Right column (optional): selection side panel (desktop) for docs/tasks/graph

Desktop:
- show split view for Documents tab (“selected doc preview”)

Mobile:
- use modal drawers for selection details (similar to Search page behavior)

### 4.2. Actions menu (header)

Actions are API-driven when possible:
- Refresh index (already: `POST /api/v1/index/refresh`)
- Run doctor (future endpoint)
- Copy ticket id, copy ticket dir, copy index path

### 4.3. Data fetching strategy

Use RTK Query endpoints:
- `getTicketSummary`
- `getTicketDocs`
- `getTicketTasks`
- `updateTicketTasks` (mutation)
- `getTicketGraph`

Caching:
- Tag tickets by `Ticket:<id>`
- Invalidate when tasks mutate or index refresh happens.

### 4.4. Tabs mapping to API

**Overview**
- `getTicketSummary`
- `getTicketTasks` (only counts + first N open tasks)
- `getTicketDocs` (for “Key documents”)

**Documents**
- `getTicketDocs` + client-side grouping
- “View →” uses existing `/doc?path=...`

**Tasks**
- `getTicketTasks`
- mutations to check/uncheck/add

**Graph**
- `getTicketGraph` (mermaid string)
- render with Mermaid (client-side)

**Changelog**
- open doc viewer for changelog.md (Option A)

## 5. Phased Implementation Plan (suggested)

Phase 1 (read-only MVP):
- Add `/ticket/:ticket` route with Overview + Documents using only existing endpoints:
  - `search/docs?ticket=...`
  - `docs/get` for index.md
- For Tasks and Changelog: render tasks.md/changelog.md via doc viewer (markdown), no structured controls yet.
- For Graph: show Mermaid output as text first (or preformatted).

Phase 2 (interactive):
- Add tasks endpoints (parse + check/uncheck/add).
- Add graph endpoint wrapping existing Mermaid generator.

Phase 3 (polish):
- Add insights/hub stats
- Add doc selection side panel and deep-link via `doc=...`

## 6. Open Decisions

- Should ticket navigation be by:
  - `ticket id` only (`/ticket/001-ADD-DOCMGR-UI`), or
  - ticket dir path (more precise, but uglier)?
- Should changelog be parsed into structured entries, or remain “markdown document view”?
- For tasks: do we preserve “task indices are implicit by scan order” (current behavior), or introduce stable IDs?
