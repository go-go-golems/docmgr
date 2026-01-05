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
