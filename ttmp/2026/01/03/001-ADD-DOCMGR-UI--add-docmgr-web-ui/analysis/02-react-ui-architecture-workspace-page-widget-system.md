---
Title: React UI architecture + Workspace page widget system
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ui
    - web
    - workspace
    - ux
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design/03-workspace-rest-api.md
      Note: API contract referenced by the widget data-dependency mapping.
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md
      Note: Source UX spec; widget inventory derived from these ASCII designs.
    - Path: ui/src/App.tsx
      Note: Top-level routing and Provider setup; baseline for adding Workspace routes.
    - Path: ui/src/components/DocCard.tsx
      Note: Shared domain card but styled via Search-centric classes; candidate for design-system separation.
    - Path: ui/src/features/search/SearchPage.tsx
      Note: Largest page; primary source of extraction candidates (filters/results/preview/shortcuts).
    - Path: ui/src/features/ticket/TicketPage.tsx
      Note: Second-largest page; tab widgets model maps to Workspace dashboard composition.
    - Path: ui/src/services/docmgrApi.ts
      Note: RTK Query API layer and hook exports; impacts all widget data dependencies.
ExternalSources: []
Summary: Audit of the current React SPA architecture and a proposed widget/design-system structure for the upcoming Workspace navigation pages.
LastUpdated: 2026-01-05T08:18:56.392823046-05:00
WhatFor: Guide incremental refactors (extract widgets/primitives) and inform implementation of the Workspace/Home/Tickets/Topics/Recent UI pages.
WhenToUse: Before adding new Workspace navigation pages or when reorganizing existing Search/Ticket pages into shared widgets and design-system primitives.
---



# React UI architecture + Workspace page widget system

## Summary

The current React UI is a Vite SPA using React Router for page routing, Redux Toolkit + RTK Query for API/data, and Bootstrap (plus a small amount of custom CSS in `ui/src/App.css`) for styling. The app is functional and consistent, but some route components have grown into “everything pages” (notably `SearchPage.tsx`), which makes it hard to evolve the UI into a multi-page navigation experience (Workspace/Tickets/Topics/Recent/etc).

This document maps the current architecture, identifies the biggest “implicit widgets” inside existing pages, and proposes a widget/page/design-system architecture for the Workspace navigation pages described in `sources/workspace-page.md`.

## Current UI architecture (as implemented)

### Build/runtime
- Vite SPA under `ui/` (with `node_modules` checked in for this workspace).
- Bootstrap is imported globally in `ui/src/main.tsx` (`bootstrap/dist/css/bootstrap.min.css`).
- Global styles live in `ui/src/index.css` and feature-ish styles (for search/results, markdown/code, etc) currently live in `ui/src/App.css`.

### Entry points + routing
- `ui/src/main.tsx` renders `<App />` into `#root` inside `<StrictMode />`.
- `ui/src/App.tsx` provides the Redux store and configures routes:
  - `/` → `SearchPage`
  - `/doc` → `DocViewerPage` (query param `path=...`)
  - `/file` → `FileViewerPage` (query params `root=repo|docs`, `path=...`)
  - `/ticket/:ticket` → `TicketPage`

### Data access (RTK Query)
- Central API definition: `ui/src/services/docmgrApi.ts`.
  - Uses `fetchBaseQuery({ baseUrl: '/api/v1' })`.
  - Encapsulates API request/response types and exports hooks.
  - Tagging strategy: `tagTypes: ['Workspace', 'Search', 'Ticket']` with fairly clean invalidation for refresh and ticket mutations.
- Store wiring: `ui/src/app/store.ts` with:
  - `docmgrApi.reducer` and `docmgrApi.middleware`
  - `searchReducer` (local UI state for Search)

### “Layering” as it exists today

The current structure is close to “feature-first”, but route pages and reusable pieces are somewhat mixed:
- **Route-level pages** live under `ui/src/features/*/*Page.tsx`.
- **Reusable components** live under `ui/src/components/` (`DocCard`, `MermaidDiagram`).
- **API/service layer** lives under `ui/src/services/` (`docmgrApi.ts`).
- **Redux app plumbing** lives under `ui/src/app/` (`store.ts`, `hooks.ts`).

This is a good start, but the biggest pressure point is that “page-local widgets” are defined inline inside pages (especially Search), so there isn’t a natural home for shared widgets once Workspace nav pages arrive.

## Page-by-page audit (what exists + what’s getting big)

### Search (`ui/src/features/search/SearchPage.tsx`, ~1649 LOC)

This file currently contains:
- Page shell/layout (header, search bar, mode toggles)
- Workspace refresh button + `useGetWorkspaceStatusQuery` integration
- Form and filter UI (including a mobile filter drawer)
- URL sync (read/write search state and selected doc via URL params)
- Keyboard shortcuts (global `keydown` listener, focus management, selection navigation)
- Search execution logic (docs search, file search, pagination)
- Results rendering (cards using `DocCard`)
- Preview panel + mobile preview modal
- Diagnostics rendering
- A mini “design system” inlined as helper components/functions:
  - `useIsMobile`
  - `toErrorBanner`
  - toast state management
  - markdown snippet rendering + term highlighting (`MarkdownSnippet`, `highlightReactNode`, etc)
  - `TopicMultiSelect`
  - `DiagnosticList`

**Biggest maintainability risks**
- The file is currently a “kitchen sink”; adding a second large, interactive page (Workspace dashboard) will likely duplicate patterns (header layout, toasts, errors, drawers/modals, cards, list virtualization).
- Several utilities and UI patterns are duplicated across pages already (`toErrorBanner` and “copy-to-clipboard toast” exist in multiple places with minor differences).

**Immediate extraction candidates (high ROI)**
- `useIsMobile()` → shared hook (or a simple `useMediaQuery`).
- Toast logic + copy-to-clipboard helper:
  - `useToast()` (imperative `toast.success(...)`, `toast.error(...)`)
  - `useClipboard()` (`copy(text)` with consistent fallback/error text)
- Error mapping and rendering:
  - `ApiErrorAlert` or `ErrorBanner` component shared by Search/Doc/File/Ticket.
- Search-specific widgets:
  - `SearchHeader` (title + refresh)
  - `SearchModeToggle` (Docs / Reverse / Files)
  - `SearchFiltersPanel` (desktop)
  - `SearchFiltersDrawer` (mobile)
  - `SearchResultsList`
  - `SearchPreviewPanel` / `SearchPreviewModal`
  - `SearchDiagnosticsPanel`
- Snippet rendering:
  - `MarkdownSnippet` can become a reusable widget for any “excerpt with highlighted terms”, useful again in Topics/Recent views.

### Doc viewer (`ui/src/features/doc/DocViewerPage.tsx`, ~259 LOC)

This page is already fairly modular, but repeats patterns that should become shared:
- Toast state for copy actions (`Copy path`, `Copy markdown`)
- Error banner mapping (`toErrorBanner`)
- Generic “page header” layout with Back + Search buttons
- A “Related files” section which is similar in spirit to Ticket’s preview list (a pattern that will recur).

Extraction candidates:
- `PageHeader` (title/subtitle + right-side actions)
- `RelatedFilesList` (render list with copy/open actions)
- Shared `ApiErrorAlert` + `useClipboard` + `useToast`

### File viewer (`ui/src/features/file/FileViewerPage.tsx`, ~154 LOC)

This page is small but also repeats:
- Toast + copy patterns
- Error mapping (`toErrorBanner`)
- Page header layout
- Code presentation styling (`docmgr-code` in `ui/src/App.css`)

This page is a good example of where “design system primitives” will pay off: a shared `CodeBlock` component could unify the code styling and avoid each page doing the same `pre` boilerplate.

### Ticket page (`ui/src/features/ticket/TicketPage.tsx`, ~653 LOC)

The Ticket page is already structured around a tab model, but it contains multiple “tab bodies” inline:
- Overview tab = multiple cards (metadata, stats, key docs, open tasks, index.md)
- Documents tab = grouped lists + preview pane
- Tasks tab = progress, sections, checkboxes, add task
- Graph tab = Mermaid diagram + DSL `<details>`
- Changelog tab = link to `changelog.md`

Extraction candidates (each can be a widget):
- `TicketHeader`
- `TicketTabs`
- `TicketOverviewTab`
  - `TicketStatsCard`, `KeyDocsCard`, `OpenTasksCard`, `IndexDocCard`
- `TicketDocumentsTab`
  - `TicketDocsByTypeList`, `TicketDocPreviewCard`
- `TicketTasksTab`
  - `TaskSections`, `AddTaskCard`
- `TicketGraphTab` (already depends on `MermaidDiagram`)
- Shared “doc card” concept:
  - `DocCard` is already a shared component, but it still carries Search-specific naming/styling (`result-card`, etc) which suggests we should separate *design system card* vs *domain card*.

## Existing widgets/components to reorganize (specific callouts)

### `DocCard` (`ui/src/components/DocCard.tsx`)
- **What it is today**: a domain card (doc result) + some design-system concerns (layout, spacing) + Search-derived CSS classes (`result-card`, `result-title`, etc).
- **What to change over time**:
  - Keep `DocCard` as a domain component, but move generic styling primitives into a shared `Card` primitive (or rename the CSS to be domain-neutral, e.g. `.dm-card`).
  - Move `timeAgo()` into a shared `lib/time.ts` so the same formatting is used everywhere (Search and other future pages will want this too).
- **Why it matters for Workspace pages**: Recent documents and topic pages will likely want a “document list item” that is a smaller variant of the same component.

### Inline “widgets” inside `SearchPage.tsx`
These are effectively widgets already; they just don’t have a file boundary yet:
- URL synchronization (read/write query params)
- keyboard shortcuts + selection model
- filter controls (desktop) + filter drawer (mobile)
- results list + pagination (“load more”)
- preview panel + preview modal
- markdown snippet rendering + highlighting terms
- diagnostics list rendering

Any of these moving into their own file makes the page easier to reason about and makes reuse feasible when we add a Tickets page (filters + list + preview patterns are very similar).

### Inline “tab widgets” inside `TicketPage.tsx`
Treat each tab body as a widget:
- Overview = a dashboard (cards + lists + markdown)
- Documents = grouped list + preview panel
- Tasks = checklist + add form
- Graph = diagram + debug details

This is nearly the same structural problem as the Workspace dashboard: multiple “cards” assembled into one page. Extracting these tab widgets is a direct rehearsal for building `WorkspaceHomePage`.

## Coherent design system: what we should standardize now

The UI already uses Bootstrap heavily; a “docmgr design system” can be implemented as a thin wrapper layer on top of Bootstrap, plus a small set of CSS variables/utilities.

### Design primitives (shared, not domain-specific)
Recommended to define as reusable React components (wrapping Bootstrap) + a small stylesheet:
- Layout: `AppShell`, `PageHeader`, `Section`, `Stack`, `Grid`
- Feedback: `Toast`, `EmptyState`, `LoadingSpinner`, `InlineError`, `ApiErrorAlert`
- Navigation: `TopBar`, `SideNav`, `Breadcrumbs`, `Tabs`
- Interaction: `IconButton`, `ButtonRow`, `SearchInput`, `FilterChip`, `Drawer`, `Modal`
- Display: `StatPill`, `MetaRow`, `TagList`, `ProgressBar`, `CodeBlock`, `MarkdownBlock`

### Design tokens (thin layer over Bootstrap)
Even with Bootstrap, it’s worth standardizing a few tokens to keep pages cohesive:
- Spacing scale for “section gaps” (`--dm-space-1..4`)
- Card density variants (`compact`, `default`) for desktop vs mobile
- Consistent “mono” text style for paths/ticket IDs (`.dm-mono`)
- Consistent page backgrounds and content max widths (avoid mixing ad-hoc `.search-container` with plain `.container`)

### CSS strategy (pragmatic and incremental)

Current state:
- `ui/src/index.css` holds global-ish base styles (body background, highlight.js import).
- `ui/src/App.css` mixes:
  - Search page layout/styles (`.search-container`, `.results-grid.split`, `.preview-panel`, etc)
  - shared-ish utilities (`.docmgr-markdown`, `.docmgr-code`)

Proposed direction:
- Keep Bootstrap as the “baseline UI kit”.
- Introduce a small `dm-*` namespace for docmgr-specific utilities and patterns:
  - `.dm-mono`, `.dm-card`, `.dm-section`, `.dm-empty`, `.dm-kbd`, etc.
- Split “design system utilities” from “page-specific layout”:
  - `ui/src/styles/design-system.css` (tokens + utilities + shared patterns like markdown/code blocks)
  - `ui/src/styles/search.css` (Search-only layout, split preview grid)
  - (later) `ui/src/styles/ticket.css`, `ui/src/styles/workspace.css` if needed

This avoids a future where adding Workspace pages requires editing a single mega CSS file with coupled concerns.

### File organization for design system
Today, `ui/src/App.css` contains both global-ish styles and Search-specific styles. A clearer split:
- `ui/src/styles/`:
  - `globals.css` (body/background, highlight.js theme import)
  - `design-system.css` (card/list/pill patterns, `.dm-*` utilities)
  - `search.css` (only Search-specific layout like split preview grid)
- Or keep CSS collocated by widgets, and only keep tokens/utilities globally.

## Workspace navigation pages: widget architecture proposal

`sources/workspace-page.md` describes a multi-page navigation shell with a shared top bar and a left nav. Treat this as a product surface: we want a stable layout shell with swappable content widgets.

### Widget sizing and ownership (rules of thumb)

To keep the Workspace pages maintainable, prefer these boundaries:
- **Page (route component)**: orchestrates data + composes widgets; minimal local state; avoid embedding complex render helpers. Target size: ~150–300 LOC.
- **Widget**: a reusable page section (often “one card” or “one panel”), owns presentation state (expanded/collapsed, local sorting, etc) and composes smaller components. Target size: ~100–250 LOC.
- **Feature component**: business interaction unit (e.g. “ticket task checkbox list with mutation”), can be used by multiple widgets/pages. Target size: ~100–250 LOC.
- **Design-system primitive**: tiny, highly reusable (`PageHeader`, `StatPill`, `ApiErrorAlert`). Target size: ~30–150 LOC.

These are not hard limits, but they help prevent another `SearchPage.tsx` situation as the UI grows.

### Page inventory implied by the designs
- `WorkspaceHomePage` (Dashboard)
- `TicketsPage` (table view + card view, and later a board view)
- `TopicsPage` (topic browser)
- `TopicDetailPage` (topic drill-down)
- `RecentActivityPage` (timeline)
- A shared `AppShell` that adapts for mobile (“compact mobile navigation”)

### Shell widgets (shared across pages)
These are the widgets we should build once and re-use everywhere:
- `TopBar`
  - left: product name + optional breadcrumb
  - right: global actions (`Search`, `Refresh`, last indexed time)
- `SideNav`
  - Home / Tickets / Search / Topics / Recent
  - “secondary nav” area (quick stats, quick links) when on desktop
- `ContentHeader`
  - per-page title, optional subtitle, optional right-side actions (sort/view toggles)

**How this maps to current code**
- The current pages are “full-bleed containers” without a shared shell; each page reimplements its own header (Search title + refresh, Doc/File back/search buttons, Ticket title + tabs).
- The first concrete design-system win is a `PageHeader` primitive that can be reused by all existing pages *and* become the top section of the Workspace shell.

### Dashboard (Design 1) widget breakdown
Proposed widgets:
- `WorkspaceOverviewCard`
  - workspace root(s), indexed timestamp, document counts, FTS availability
- `QuickStatsCard` (can live in the sidebar or as a main widget on mobile)
  - tickets total + status counts (active/review/complete/draft)
  - documents total
- `TicketsByStatusWidget`
  - could start as a simple “4 columns with counts” (as shown) and later become a chart
- `RecentActivityWidget`
  - list of recent ticket/doc updates, with per-item actions
  - re-usable on `RecentActivityPage` as a “sectioned timeline” variant
- `QuickLinksWidget`
  - nav shortcuts (All Tickets, All Topics, Stale Docs)

### Tickets list (Design 2 + 3) widget breakdown
Shared widgets:
- `TicketsFiltersBar`
  - status dropdown, owner dropdown, intent dropdown, topic token input
  - active filter chips + Clear
- `TicketsSidebarFacets`
  - topics facet list + owner facet list + status facet list
- `TicketsViewToggle`
  - Table / Cards / Board (Board can be “disabled” initially)
- `TicketsSortSelect`
  - order by last updated, created, status, etc
- `TicketsPaginator`
  - Load more / infinite scroll

View-specific widgets:
- `TicketsTable`
  - row widget: `TicketRow`
- `TicketsCardGrid`
  - card widget: `TicketCard` (progress, docs/files count, updated time)
- (Later) `TicketsBoard`
  - column widget: `TicketsBoardColumn` (status)
  - drag/drop interactions (postpone until rest is stable)

### Topics browser + detail (Design 5 + 6)
Widgets:
- `TopicsGrid`
  - card widget: `TopicCard`
- `TopicSummaryCard`
  - counts, description, related topics chips/links
- `TopicTicketsByStatus`
  - expandable sections: `TopicTicketSection` (Active/Review/etc)
- `RecentDocsList`
  - list item widget: `DocListItem` (title, ticket, doc-type, updated)

### Recent activity (Design 7)
Widgets:
- `ActivityTimeline`
  - group widget: `ActivityDaySection` (Today/Yesterday/This Week)
  - item widget: `ActivityItem` (ticket/doc/task events)

### Mobile navigation (Design 8)
Mobile-first pattern:
- Replace always-visible `SideNav` with:
  - `MobileTopBar` (hamburger + search + refresh)
  - `MobileQuickActions` (links)
  - “dashboard cards” stacked (`QuickStatsCard`, `RecentActivityWidget`, etc)

This implies we want the shell widgets to support both “sidebar” and “drawer” presentations without rewriting each page.

### Data dependencies (so widgets don’t invent ad-hoc fetch patterns)

The designs are explicitly “REST API contract-driven” (see `design/03-workspace-rest-api.md`). The widget boundaries above map cleanly to API calls:
- `TopBar`:
  - `GET /api/v1/workspace/status` (already in `docmgrApi.ts` as `getWorkspaceStatus`)
  - `POST /api/v1/index/refresh` (already in `docmgrApi.ts` as `refreshIndex`)
- `WorkspaceHomePage`:
  - Ideally a single `GET /api/v1/workspace/summary` (proposed in the design doc) to avoid N calls for stats + recents.
- `TicketsPage`:
  - `GET /api/v1/workspace/tickets` (proposed) for the main list
  - optionally `GET /api/v1/workspace/facets` (if added later) to drive sidebar counts without extra client-side aggregation
- `TopicsPage` / `TopicDetailPage`:
  - `GET /api/v1/workspace/topics` and `GET /api/v1/workspace/topics/:topic` (names TBD; proposed pattern)
- `RecentActivityPage`:
  - `GET /api/v1/workspace/activity` (names TBD; proposed pattern)

**RTK Query organization recommendation**
- Keep a single `docmgrApi` for now (small codebase, shared error envelope).
- Group endpoints in the file by “domain section” (workspace/search/tickets/docs/files) and export hooks in the same order, so discoverability stays high as the API grows.

## How to reorganize existing widgets/files without a big-bang rewrite

### 1) Introduce a stable “shell + pages + widgets” directory layout (incrementally)
Keep the current pages working, but establish a place for shared widgets as Workspace pages land:
- `ui/src/pages/` — route components (thin orchestrators)
- `ui/src/widgets/` — reusable page sections (TopBar, SideNav, Filters, Lists)
- `ui/src/ui/` (or `ui/src/shared/`) — design system primitives (PageHeader, Toast, etc)
- Keep `ui/src/services/` and `ui/src/app/` as-is.

Concrete proposed tree (one possible shape):

```text
ui/src/
  app/
    store.ts
    hooks.ts
  services/
    docmgrApi.ts
  lib/
    apiError.ts        # parse + format API errors
    clipboard.ts       # copy helper
    time.ts            # timeAgo/formatDate
  ui/                  # design system primitives (thin Bootstrap wrappers)
    PageHeader.tsx
    ApiErrorAlert.tsx
    EmptyState.tsx
    ToastHost.tsx
    MarkdownBlock.tsx
    CodeBlock.tsx
  widgets/
    shell/
      TopBar.tsx
      SideNav.tsx
      AppShell.tsx
    search/
      SearchFiltersPanel.tsx
      SearchResultsList.tsx
      SearchPreviewPanel.tsx
    tickets/
      TicketsFiltersBar.tsx
      TicketCard.tsx
      TicketsTable.tsx
    topics/
      TopicCard.tsx
      TopicsGrid.tsx
    activity/
      ActivityTimeline.tsx
  pages/
    SearchPage/
      SearchPage.tsx
    TicketPage/
      TicketPage.tsx
    DocViewerPage/
      DocViewerPage.tsx
    FileViewerPage/
      FileViewerPage.tsx
    WorkspaceHomePage/
      WorkspaceHomePage.tsx
```

This gives us “homes” for:
- extraction candidates from Search/Ticket (widgets and shared UI)
- new Workspace pages (pages + widgets) without forcing a rewrite of existing code on day 1.

### 2) Split the biggest pages along widget boundaries
Use file size as a forcing function:
- `SearchPage.tsx` (1649 LOC) should become:
  - `SearchPage.tsx` (layout + orchestration)
  - `SearchForm.tsx`, `SearchFilters*`, `SearchResultsList.tsx`, `SearchPreview*`, `SearchDiagnostics.tsx`
  - `searchHooks.ts` (URL sync, keyboard shortcuts, selection model)
- `TicketPage.tsx` (653 LOC) should become:
  - `TicketPage.tsx` (tab selection + top-level queries)
  - `TicketOverviewTab.tsx`, `TicketDocumentsTab.tsx`, `TicketTasksTab.tsx`, `TicketGraphTab.tsx`

Additional “too-big” signals (even if LOC is smaller):
- Multiple copies of the same helper in multiple pages (toast/copy/error) → move to `ui/` or `lib/`.
- CSS file that mixes global tokens + one-page layout rules (`App.css`) → split tokens/utilities from page-specific layout.

### 3) Normalize duplicated patterns into shared primitives
Concrete duplicates to remove over time:
- `toErrorBanner` / `apiErrorMessage` → shared `apiErrorMessage(err)` + shared `ApiErrorAlert`
- Clipboard copying/toasts → shared `useClipboard` + `useToast`
- “time ago” formatting appears both in `DocCard` and Search page utilities → shared `timeAgo()`

## Suggested extraction sequence (so refactors stay reviewable)

If we treat the Workspace navigation UI as the “final set of pages”, the best order is:

1. **Shared primitives first (no routing changes yet)**
   - `ApiErrorAlert` (unify error envelope parsing and rendering)
   - `useClipboard` + `useToast` (remove repeated copy/toast code)
   - `timeAgo` / `formatDate` helpers
   - `MarkdownBlock` / `CodeBlock` (standardize doc and code rendering styles)

2. **Split Search into widgets (largest risk reducer)**
   - Extract filter panel/drawer, results list, preview panel/modal, diagnostics panel.
   - Extract “behavior hooks”: URL sync, selection model, keyboard shortcuts.
   - Keep the route component thin and focused on composition.

3. **Split Ticket tab bodies into widgets**
   - Each tab becomes a file; keep the data fetching in `TicketPage` or in per-tab widgets (but be consistent).
   - This directly exercises the “dashboard composed of cards” pattern needed for `WorkspaceHomePage`.

4. **Introduce the Workspace shell**
   - Build `AppShell` with `TopBar` + `SideNav` and move existing pages inside it incrementally.
   - Add new routes for Workspace pages and implement them as widget composition.

This sequence minimizes the chance that Workspace UI work gets blocked by ongoing refactors, because steps 1–2 are immediately useful regardless of future pages.

## Git staging/commit hygiene for this work (docs + refactors)

This change set (analysis + eventual refactors) will be easiest to review if we keep commits small and scoped:
- Prefer one commit per extraction (e.g. “extract useToast/useClipboard”, “extract TicketOverviewTab”, etc).
- Avoid bundling unrelated CSS restyling with behavior changes.
- Before each commit:
  - `git status --porcelain`
  - `git diff --stat` and `git diff --cached --stat`
  - `git diff --cached --name-only`

## References
- Workspace page designs: `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md`
- Current routes: `ui/src/App.tsx`
- Current API client: `ui/src/services/docmgrApi.ts`
