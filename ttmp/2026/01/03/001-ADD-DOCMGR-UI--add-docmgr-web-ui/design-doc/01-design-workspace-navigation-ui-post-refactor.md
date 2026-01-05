---
Title: 'Design: Workspace navigation UI (post-refactor)'
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - tooling
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md
      Note: Baseline analysis + widget/primitives strategy.
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design/03-workspace-rest-api.md
      Note: Proposed REST endpoints for workspace navigation pages; UI should align with this.
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md
      Note: |-
        Source ASCII designs for Home/Tickets/Topics/Recent + mobile.
        Source ASCII designs
    - Path: ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/analysis/01-post-refactor-ui-componentization-roadmap.md
      Note: Post-refactor roadmap that informs folder/component decisions.
    - Path: ui/src/App.tsx
      Note: |-
        Routing; Workspace shell will be introduced here.
        Workspace routes
    - Path: ui/src/components/DocCard.tsx
      Note: Domain card reused across Search/Ticket and future Workspace widgets.
    - Path: ui/src/components/PageHeader.tsx
      Note: Shared page header primitive to reuse in Workspace shell/pages.
    - Path: ui/src/components/ToastHost.tsx
      Note: Global toast UX primitive (refresh/copy feedback).
    - Path: ui/src/features/search/widgets/SearchHeader.tsx
      Note: Entry link to Workspace
    - Path: ui/src/features/workspace/WorkspaceHomePage.tsx
      Note: Dashboard MVP using workspace status
    - Path: ui/src/features/workspace/WorkspaceLayout.tsx
      Note: Workspace shell implementation
    - Path: ui/src/services/docmgrApi.ts
      Note: RTK Query endpoints; Workspace endpoints will be added here.
    - Path: ui/src/styles/design-system.css
      Note: dm-* utilities and shared styling; Workspace should prefer these over page-local CSS.
ExternalSources: []
Summary: UI design for implementing the Workspace navigation pages (Home/Tickets/Topics/Recent + mobile) using the now-refactored widget/primitives architecture (post ticket 007), aligned with the workspace REST API design.
LastUpdated: 2026-01-05T13:00:14.19672193-05:00
WhatFor: Provide a concrete UI architecture, route structure, widget breakdown, and incremental implementation plan for Workspace pages without regressing into “mega page” components.
WhenToUse: When implementing new Workspace routes or adding/adjusting workspace REST endpoints; use as the source of truth for component boundaries and state ownership.
---


# Design: Workspace navigation UI (post-refactor)

## Executive Summary

We will implement a new **Workspace navigation shell** (TopBar + SideNav + content area) and a set of Workspace routes (Home/Tickets/Topics/Topic Detail/Recent) based on `sources/workspace-page.md`. The design leverages the recent refactor work (ticket `007`) that introduced a consistent widget architecture and shared UI primitives (headers, toasts, markdown/code blocks, domain cards).

The implementation is incremental and safe:
- Add `/workspace` routes and shell without breaking existing `/` (Search) and `/ticket/:ticket` flows.
- Use RTK Query for server state, Redux slices only for shared/persistent UI intent, and local state for ephemeral UI mechanics.
- Extract shared patterns only when proven in-context (e.g., list+preview once Workspace adds a third consumer).

## Problem Statement

The UI currently supports Search/Doc/File/Ticket routes, but it lacks:
- A coherent “full-site” navigation model (Home/Tickets/Topics/Recent).
- A shared shell that prevents each page from re-implementing header/nav/refresh behavior.
- A predictable widget boundary system so new pages don’t grow into kitchen-sink components.

`sources/workspace-page.md` defines a multi-page experience that requires both:
1) A stable shell and reusable widgets (filters, cards, lists, nav).
2) New workspace-level endpoints (summary, tickets list, topics, activity) to avoid client-side N+1 calls and ad-hoc aggregation.

## Proposed Solution

### 1) Routing and shell

Introduce a new route group under `/workspace/*` that renders an `AppShell` and nested pages via React Router:
- `/workspace` → Workspace Home/Dashboard
- `/workspace/tickets` → Tickets list (table view initially, cards toggle later)
- `/workspace/topics` → Topics browser
- `/workspace/topics/:topic` → Topic detail
- `/workspace/recent` → Recent activity

Existing routes remain intact:
- `/` stays Search (for now)
- `/ticket/:ticket` stays Ticket page
- `/doc`, `/file` remain viewer routes

The shell includes:
- `TopBar`: product title, breadcrumb, Search button, Refresh button, “indexed X ago”
- `SideNav`: Home/Tickets/Search/Topics/Recent

### 2) Widget breakdown (from workspace-page.md)

**Workspace Home**
- `WorkspaceOverviewCard` (workspace roots + indexed time + doc count)
- `QuickStatsCard` (ticket status counts; requires workspace summary endpoint)
- `TicketStatsWidget` (tickets-by-status bar chart-ish view; requires summary)
- `RecentTicketsWidget` (recently updated tickets; requires “recent tickets” endpoint)
- `RecentDocsWidget` (recently updated docs; uses `DocCard` or a new `DocListItem`)

**Tickets list**
- `TicketsFiltersBar` (status, topics, owner, intent + Clear)
- `TicketsSidebarFacets` (topics/owners/status counts; optional)
- `TicketsTable` (first pass) + `TicketsCardGrid` (toggle later)

**Topics**
- `TopicsGrid` (topic cards, ticket counts)
- `TopicSummaryCard` + `TopicTicketsByStatus` + `TopicRecentDocs`

**Recent activity**
- `TimeRangeToggle`
- `ActivityTimeline` (grouped sections)

### 3) State ownership rules (post-refactor)

We apply the state strategy from ticket 007:
- Server state: RTK Query (workspace summary, lists, activity)
- Shared/persistent UI intent (filters/view mode/sort): Redux slice per page domain (`ticketsSlice`, `topicsSlice`) when needed
- UI mechanics (drawer open/closed, modal open, selection): local component state

### 4) API alignment

Workspace pages should align with the proposed endpoints in `design/03-workspace-rest-api.md`. The UI will be built “endpoint-first” so we don’t reimplement aggregation in the client. When endpoints are missing, the UI should show explicit placeholders rather than inventing client-side scans.

## Design Decisions

### Keep `/` as Search initially
Rationale: avoids breaking existing workflows and keeps the refactor low-risk. We can later choose to make `/` redirect to `/workspace` once Workspace is feature-complete.

### Shell is a nested route (not copied per page)
Rationale: ensures consistent page chrome, avoids header duplication, and makes mobile nav behavior centralized.

### Shared primitives first; shared “patterns” later
Rationale: primitives (toasts, headers, cards) are safe and already proven. Patterns (list+preview, filters+chips) should be extracted only once Workspace introduces a third consumer so we don’t over-generalize prematurely.

## Alternatives Considered

### Replace Search as the homepage immediately
Rejected for now: too much UX churn while the Workspace pages are still being built.

### Make Workspace pages separate from existing routes (no shared shell)
Rejected: would recreate duplication (headers, refresh logic, nav), reintroducing the “mega page” problem.

### Put all UI state in Redux
Rejected: page-local mechanics (drawers/modals/temporary selection) are cheaper and safer as local state; Redux is reserved for shared/persistent intent and cross-widget coordination.

## Implementation Plan

### Phase 1: Shell + route scaffolding (UI-first)
1) Add `/workspace/*` route group and placeholder pages.
2) Implement `AppShell` with `TopBar` (refresh + status) and `SideNav`.
3) Add a “Workspace” link from Search header to bootstrap navigation.

### Phase 2: Dashboard MVP
1) Implement `WorkspaceOverviewCard` using existing workspace status data.
2) Add placeholders for stats/recent widgets gated on new endpoints.

### Phase 3: Workspace endpoints + page fill-in
1) Implement minimal workspace summary endpoint(s) per `design/03-workspace-rest-api.md`.
2) Implement tickets list endpoint (paged/sortable/filterable) and wire Tickets page.
3) Implement topics list + topic detail endpoints and wire Topics pages.
4) Implement recent activity endpoint and wire Recent page.

### Phase 4: Mobile nav + design-system tightening
1) Add mobile drawer for SideNav and compact top bar layout.
2) Standardize remaining dm-* utilities and replace any lingering Search-only styling.

## Open Questions

1) Should `/` remain Search long-term, or should Workspace Home become the new default entry point?
2) Should viewer routes (`/doc`, `/file`, `/ticket/:ticket`) be inside the Workspace shell (for consistent nav), or remain “standalone pages”?
3) Do we want a breadcrumb in TopBar for deep pages (Tickets → Topic → Ticket)?
4) Exact endpoint payload shapes: confirm against `design/03-workspace-rest-api.md` before implementing server-side.

## References

- `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md`
- `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design/03-workspace-rest-api.md`
- `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`
- `ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/analysis/01-post-refactor-ui-componentization-roadmap.md`

## Appendix: ASCII screenshots (verbatim)

The following ASCII designs are copied verbatim from `sources/workspace-page.md` so this design doc is self-contained.
<!-- BEGIN ASCII SCREENSHOTS: workspace-page.md -->
# ASCII Designs for Workspace Navigation Pages

## Design 1: Workspace Home / Dashboard

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr                                    [🔍 Search]  [🔄 Refresh] 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│ ┌─ Nav ────────────┐                                                        │
│ │ [●] Home         │                                                        │
│ │ [ ] Tickets      │   ┌─ WORKSPACE OVERVIEW ──────────────────────────┐  │
│ │ [ ] Search       │   │                                                │  │
│ │ [ ] Topics       │   │ ttmp/                                          │  │
│ │ [ ] Recent       │   │ /Users/dev/projects/docmgr                     │  │
│ │                  │   │                                                │  │
│ ├──────────────────┤   │ Indexed: Jan 5, 2026 12:34 PM                  │  │
│ │ 📊 Quick Stats   │   │ Documents: 413                                 │  │
│ │ ───────────      │   │                                                │  │
│ │ Tickets:    128  │   └────────────────────────────────────────────────┘  │
│ │ Active:      12  │                                                        │
│ │ Review:       9  │   ┌─ TICKET STATS ─────────────────────────────────┐  │
│ │ Complete:    84  │   │                                                 │  │
│ │ Draft:       23  │   │  ┌──────────────────────────────────────────┐  │  │
│ │                  │   │  │         Tickets by Status                │  │  │
│ │ 📌 Quick Links   │   │  │  ┌────────┬────────┬────────┬────────┐  │  │  │
│ │ ───────────      │   │  │  │ Active │ Review │Complete│ Draft  │  │  │  │
│ │ [Recent Activity]│   │  │  │   12   │   9    │   84   │   23   │  │  │  │
│ │ [All Tickets]    │   │  │  │  [██]  │  [█]   │ [████] │  [█]   │  │  │  │
│ │ [All Topics]     │   │  │  └────────┴────────┴────────┴────────┘  │  │  │
│ │ [Stale Docs]     │   │  │                                          │  │  │
│ └──────────────────┘   │  │  Active: 12 tickets   •   9 in review   │  │  │
│                        │  │  84 completed         •   23 drafts      │  │  │
│                        │  │                                          │  │  │
│                        │  │  [View All Tickets →]                    │  │  │
│                        │  └──────────────────────────────────────────┘  │  │
│                        └─────────────────────────────────────────────────┘  │
│                                                                             │
│ ┌─ RECENT ACTIVITY ──────────────────────────────────────────────────────┐ │
│ │                                                                         │ │
│ │ Recently Updated Tickets                               [View All →]    │ │
│ │ ─────────────────────────                                              │ │
│ │                                                                         │ │
│ │ 📋 001-ADD-DOCMGR-UI: Add docmgr Web UI               active  •  2h ago│ │
│ │    backend, docmgr, tooling, ux, web                                   │ │
│ │    Tasks: 25/27 (93%)                                    [Open →]      │ │
│ │                                                                         │ │
│ │ 📋 005-USE-SQLITE-FTS: FTS-backed search engine        draft  •  5h ago│ │
│ │    backend, docmgr, tooling, testing                                   │ │
│ │    Tasks: 3/8 (38%)                                      [Open →]      │ │
│ │                                                                         │ │
│ │ 📋 003-DOC-VALIDATION: Enhanced doc validation        review  •  1d ago│ │
│ │    docmgr, tooling, quality                                            │ │
│ │    Tasks: 12/12 (100%)                                   [Open →]      │ │
│ │                                                                         │ │
│ │ Recently Updated Documents                             [View All →]    │ │
│ │ ──────────────────────────                                             │ │
│ │                                                                         │ │
│ │ 📄 Design: docmgr Search Web UI             001-ADD-DOCMGR-UI  •  2h   │ │
│ │    design-doc                                             [View →]     │ │
│ │                                                                         │ │
│ │ 📄 FTS-backed search engine                005-USE-SQLITE-FTS  •  5h   │ │
│ │    design-doc                                             [View →]     │ │
│ │                                                                         │ │
│ │ 📄 Validation Rules Reference              003-DOC-VALIDATION  •  1d   │ │
│ │    reference                                              [View →]     │ │
│ │                                                                         │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Design 2: Tickets List Page (Table View)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr > Tickets                          [🔍 Search]  [🔄 Refresh] 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│ ┌─ Nav ────────────┐                                                        │
│ │ [ ] Home         │  ┌─ FILTERS ──────────────────────────────────────┐  │
│ │ [●] Tickets      │  │ Status: [All ▾]  Topics: [___________]  🔍    │  │
│ │ [ ] Search       │  │ Owner:  [All ▾]  Intent: [All ▾]      [Clear] │  │
│ │ [ ] Topics       │  └────────────────────────────────────────────────┘  │
│ │ [ ] Recent       │                                                        │
│ │                  │  Active: [× backend] [× ui]                128 tickets│
│ ├──────────────────┤  ──────────────────────────────────────────────────── │
│ │ 🏷️ Topics        │                                                        │
│ │ ───────────      │  ┌─ TICKETS ──────────────────────────────────────┐  │
│ │ backend    (45)  │  │                                                 │  │
│ │ docmgr     (38)  │  │ Ticket ID          Title            Status  📊  │  │
│ │ tooling    (32)  │  │ ─────────────────────────────────────────────  │  │
│ │ ui         (18)  │  │                                                 │  │
│ │ testing    (15)  │  │ 001-ADD-       Add docmgr Web   active   93%   │  │
│ │ infra      (12)  │  │ DOCMGR-UI      UI                       25/27  │  │
│ │ [View All]       │  │                backend, docmgr, tooling...     │  │
│ │                  │  │                Updated 2h ago      [Open →]    │  │
│ │ 👤 Owners        │  │                                                 │  │
│ │ ───────────      │  │ 005-USE-       FTS-backed search draft   38%   │  │
│ │ manuel     (23)  │  │ SQLITE-FTS     engine                    3/8   │  │
│ │ alex       (18)  │  │                backend, docmgr, tooling...     │  │
│ │ (none)     (87)  │  │                Updated 5h ago      [Open →]    │  │
│ │                  │  │                                                 │  │
│ │ 📅 Status        │  │ 003-DOC-       Enhanced doc      review  100%  │  │
│ │ ───────────      │  │ VALIDATION     validation                12/12 │  │
│ │ active     (12)  │  │                docmgr, tooling, quality        │  │
│ │ review      (9)  │  │                Updated 1d ago      [Open →]    │  │
│ │ complete   (84)  │  │                                                 │  │
│ │ draft      (23)  │  │ 002-HTTP-API   HTTP API design   complete 100% │  │
│ │                  │  │                                          18/18 │  │
│ └──────────────────┘  │                backend, api, http               │  │
│                       │                Updated 3d ago      [Open →]    │  │
│                       │                                                 │  │
│                       │ [Load More (124 remaining)]                    │  │
│                       └─────────────────────────────────────────────────┘  │
│                                                                             │
│                       Sort: [Last Updated ▾]  View: [Table] [Cards] [Board]│
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Design 3: Tickets List (Card View)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr > Tickets                          [🔍 Search]  [🔄 Refresh] 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│ Filters: [× active] [× backend] [× ui]                     128 tickets     │
│ Sort: [Last Updated ▾]  View: [Table] [●Cards] [Board]        [Clear All]  │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│ ┌───────────────────────────────┬───────────────────────────────┬─────────┐│
│ │ 📋 001-ADD-DOCMGR-UI          │ 📋 005-USE-SQLITE-FTS         │ 📋 003- ││
│ │ Add docmgr Web UI             │ FTS-backed search engine      │ Enhance ││
│ │                               │                               │ doc val ││
│ │ active                        │ draft                         │ review  ││
│ │ backend • docmgr • tooling... │ backend • docmgr • testing... │ docmgr..││
│ │                               │                               │         ││
│ │ 📊 Progress: 93%              │ 📊 Progress: 38%              │ 📊 100% ││
│ │ [████████████████░░] 25/27    │ [███░░░░░░░░░░░░░] 3/8        │ [█████] ││
│ │                               │                               │  12/12  ││
│ │ 📄 6 docs  •  📁 17 files     │ 📄 3 docs  •  📁 8 files      │ 📄 5    ││
│ │ Updated: 2h ago               │ Updated: 5h ago               │ 1d ago  ││
│ │                               │                               │         ││
│ │         [Open Ticket →]       │         [Open Ticket →]       │ [Open →]││
│ └───────────────────────────────┴───────────────────────────────┴─────────┘│
│                                                                             │
│ ┌───────────────────────────────┬───────────────────────────────┬─────────┐│
│ │ 📋 002-HTTP-API               │ 📋 004-SEARCH-API             │ 📋 006- ││
│ │ HTTP API design               │ Search UI Requirements        │ Vocabu..││
│ │                               │                               │         ││
│ │ complete                      │ active                        │ draft   ││
│ │ backend • api • http          │ backend • ui • ux • web       │ docmgr..││
│ │                               │                               │         ││
│ │ 📊 Progress: 100%             │ 📊 Progress: 67%              │ 📊 25%  ││
│ │ [██████████████████] 18/18    │ [█████████░░░░░░░] 8/12       │ [██░░░] ││
│ │                               │                               │  2/8    ││
│ │ 📄 4 docs  •  📁 11 files     │ 📄 2 docs  •  📁 5 files      │ 📄 1    ││
│ │ Updated: 3d ago               │ Updated: 1w ago               │ 2w ago  ││
│ │                               │                               │         ││
│ │         [Open Ticket →]       │         [Open Ticket →]       │ [Open →]││
│ └───────────────────────────────┴───────────────────────────────┴─────────┘│
│                                                                             │
│                       [Load More (122 remaining)]                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Design 5: Topics Browser Page

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr > Topics                           [🔍 Search]  [🔄 Refresh] 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│ ┌─ Nav ────────────┐                                                        │
│ │ [ ] Home         │  Browse by Topic                                       │
│ │ [ ] Tickets      │  ────────────────────────────────────────────────────  │
│ │ [ ] Search       │                                                        │
│ │ [●] Topics       │  ┌─────────────────┬─────────────────┬─────────────┐  │
│ │ [ ] Recent       │  │ 🏷️ backend      │ 🏷️ docmgr       │ 🏷️ tooling  │  │
│ └──────────────────┘  │ 45 tickets      │ 38 tickets      │ 32 tickets  │  │
│                       │                 │                 │             │  │
│                       │ Core backend    │ docmgr tool dev │ Dev tooling │  │
│                       │ services and    │ and maintenance │ and infra   │  │
│                       │ infrastructure  │                 │             │  │
│                       │                 │                 │             │  │
│                       │ [Browse →]      │ [Browse →]      │ [Browse →]  │  │
│                       └─────────────────┴─────────────────┴─────────────┘  │
│                                                                             │
│                       ┌─────────────────┬─────────────────┬─────────────┐  │
│                       │ 🏷️ ui           │ 🏷️ testing      │ 🏷️ infra    │  │
│                       │ 18 tickets      │ 15 tickets      │ 12 tickets  │  │
│                       │                 │                 │             │  │
│                       │ User interface  │ Testing and QA  │ Infrastructure │
│                       │ and UX work     │ automation      │ and ops     │  │
│                       │                 │                 │             │  │
│                       │ [Browse →]      │ [Browse →]      │ [Browse →]  │  │
│                       └─────────────────┴─────────────────┴─────────────┘  │
│                                                                             │
│                       ┌─────────────────┬─────────────────┬─────────────┐  │
│                       │ 🏷️ api          │ 🏷️ http         │ 🏷️ web      │  │
│                       │ 11 tickets      │ 9 tickets       │ 8 tickets   │  │
│                       │                 │                 │             │  │
│                       │ API design and  │ HTTP services   │ Web tech    │  │
│                       │ implementation  │ and protocols   │ and SPAs    │  │
│                       │                 │                 │             │  │
│                       │ [Browse →]      │ [Browse →]      │ [Browse →]  │  │
│                       └─────────────────┴─────────────────┴─────────────┘  │
│                                                                             │
│                       [View All Topics (24) →]                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Design 6: Topic Detail Page (Drilled Down)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr > Topics > backend                 [🔍 Search]  [🔄 Refresh] 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│ ← Back to Topics                                                            │
│                                                                             │
│ ┌─ TOPIC: backend ────────────────────────────────────────────────────┐    │
│ │                                                                      │    │
│ │ 45 tickets  •  127 documents  •  89 related files                   │    │
│ │                                                                      │    │
│ │ Core backend services and infrastructure                            │    │
│ │                                                                      │    │
│ │ Related Topics: api (28), http (15), tooling (12), infra (9)        │    │
│ │                                                                      │    │
│ └──────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│ ┌─ FILTERS ───────────────────────────────────────────────────────────┐    │
│ │ Status: [All ▾]  Owner: [All ▾]  Intent: [All ▾]      [Clear]      │    │
│ └──────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│ ┌─ ACTIVE TICKETS (12) ──────────────────────────────────── [Expand ▼] ┐   │
│ │                                                                        │   │
│ │ 📋 001-ADD-DOCMGR-UI: Add docmgr Web UI                   active  2h  │   │
│ │    backend, docmgr, tooling, ux, web                                  │   │
│ │    93% complete (25/27 tasks)                             [Open →]   │   │
│ │                                                                        │   │
│ │ 📋 004-SEARCH-API: Search UI Requirements                 active  1w  │   │
│ │    backend, ui, ux, web                                               │   │
│ │    67% complete (8/12 tasks)                              [Open →]   │   │
│ │                                                                        │   │
│ │ [Show 10 more... ▼]                                                   │   │
│ └────────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│ ┌─ REVIEW (3) ───────────────────────────────────────────── [Expand ▼] ┐   │
│ │                                                                        │   │
│ │ 📋 003-DOC-VALIDATION: Enhanced doc validation           review  1d   │   │
│ │    docmgr, tooling, quality, backend                                  │   │
│ │    100% complete (12/12 tasks)                            [Open →]   │   │
│ │                                                                        │   │
│ │ [Show 2 more... ▼]                                                    │   │
│ └────────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│ ┌─ RECENT DOCUMENTS ─────────────────────────────────────────────────────┐ │
│ │                                                                         │ │
│ │ 📄 Design: docmgr Search Web UI         001-ADD-DOCMGR-UI  •  2h ago  │ │
│ │    design-doc                                              [View →]   │ │
│ │                                                                         │ │
│ │ 📄 HTTP API design and implementation   002-HTTP-API  •  3d ago       │ │
│ │    design-doc                                              [View →]   │ │
│ │                                                                         │ │
│ │ [View All Documents (127) →]                                           │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Design 7: Recent Activity Page

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ docmgr > Recent Activity                  [🔍 Search]  [🔄 Refresh] 2m ago  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│ ┌─ Nav ────────────┐                                                        │
│ │ [ ] Home         │  ┌─ TIME RANGE ─────────────────────────────────┐    │
│ │ [ ] Tickets      │  │ [●] Today  [ ] This Week  [ ] This Month     │    │
│ │ [ ] Search       │  │ [ ] Last 3 Months  [ ] All Time              │    │
│ │ [ ] Topics       │  └──────────────────────────────────────────────┘    │
│ │ [●] Recent       │                                                        │
│ └──────────────────┘  ┌─ TODAY (Jan 5, 2026) ────────────────────────┐    │
│                       │                                               │    │
│                       │ 14:30  📋 001-ADD-DOCMGR-UI                   │    │
│                       │        Status changed: active                 │    │
│                       │        Task completed: #25 Cmd/Ctrl+R refresh │    │
│                       │                                               │    │
│                       │ 12:15  📄 Design: docmgr Search Web UI        │    │
│                       │        Document updated                       │    │
│                       │        001-ADD-DOCMGR-UI  •  design-doc       │    │
│                       │                                               │    │
│                       │ 09:45  📋 005-USE-SQLITE-FTS                  │    │
│                       │        Task added: Implement FTS5 triggers    │    │
│                       │                                               │    │
│                       │ 09:30  📄 FTS-backed search engine            │    │
│                       │        Document created                       │    │
│                       │        005-USE-SQLITE-FTS  •  design-doc      │    │
│                       └───────────────────────────────────────────────┘    │
│                                                                             │
│                       ┌─ YESTERDAY (Jan 4, 2026) ─────────────────────┐    │
│                       │                                               │    │
│                       │ 16:20  📋 003-DOC-VALIDATION                  │    │
│                       │        Status changed: review → complete      │    │
│                       │        All tasks completed (12/12)            │    │
│                       │                                               │    │
│                       │ 14:10  📄 Validation Rules Reference          │    │
│                       │        Document updated                       │    │
│                       │        003-DOC-VALIDATION  •  reference       │    │
│                       │                                               │    │
│                       │ 11:00  📋 001-ADD-DOCMGR-UI                   │    │
│                       │        3 tasks completed: #22, #23, #24       │    │
│                       │                                               │    │
│                       │ [Show 4 more... ▼]                            │    │
│                       └───────────────────────────────────────────────┘    │
│                                                                             │
│                       ┌─ THIS WEEK ───────────────────────────────────┐    │
│                       │                                               │    │
│                       │ Jan 3  📋 001-ADD-DOCMGR-UI created           │    │
│                       │ Jan 3  📄 6 documents created in ticket       │    │
│                       │ Jan 2  📋 004-SEARCH-API updated              │    │
│                       │                                               │    │
│                       │ [Show More ▼]                                 │    │
│                       └───────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Design 8: Compact Mobile Navigation

```
┌───────────────────────────────┐
│ ☰ docmgr            🔍  🔄    │
├───────────────────────────────┤
│                               │
│ ┌─ Home ──────────────────┐  │
│ │                         │  │
│ │ 📊 Quick Stats          │  │
│ │ ─────────────           │  │
│ │ Tickets:  128           │  │
│ │ Active:    12           │  │
│ │ Review:     9           │  │
│ │ Docs:     413           │  │
│ │                         │  │
│ │ [View All Tickets →]    │  │
│ └─────────────────────────┘  │
│                               │
│ ┌─ Recent ────────────────┐  │
│ │                         │  │
│ │ 📋 001-ADD-DOCMGR-UI    │  │
│ │    active  •  2h ago    │  │
│ │    [Open →]             │  │
│ │                         │  │
│ │ 📋 005-USE-SQLITE-FTS   │  │
│ │    draft  •  5h ago     │  │
│ │    [Open →]             │  │
│ │                         │  │
│ │ [View All →]            │  │
│ └─────────────────────────┘  │
│                               │
│ ┌─ Quick Actions ─────────┐  │
│ │ [🔍 Search]             │  │
│ │ [📋 All Tickets]        │  │
│ │ [🏷️ Topics]             │  │
│ │ [📅 Recent Activity]    │  │
│ └─────────────────────────┘  │
│                               │
└───────────────────────────────┘
```

These designs show:
1. **Home/Dashboard** - Workspace overview with stats and recent activity
2. **Tickets List (Table)** - Filterable table view with sidebar
3. **Tickets List (Cards)** - Grid of ticket cards with progress indicators
4. **Kanban Board** - Drag-and-drop board organized by status
5. **Topics Browser** - Topic cards with ticket counts
6. **Topic Detail** - Drilled-down view of tickets by topic
7. **Recent Activity** - Timeline of workspace changes
8. **Mobile Navigation** - Compact mobile-first layout

All designs follow the REST API contract from the design doc and support the key workspace navigation flows!
<!-- END ASCII SCREENSHOTS -->
