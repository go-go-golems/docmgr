---
Title: Post-refactor UI componentization roadmap
Ticket: 007-MODULARIZE-UI-WIDGETS
Status: active
Topics:
    - ui
    - web
    - ux
    - docmgr
    - refactor
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md
      Note: |-
        Baseline architecture analysis and the original widget/system proposal this doc builds on.
        Baseline architecture doc referenced for comparison
    - Path: ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/design-doc/01-design-redux-state-strategy-for-ui-widgets.md
      Note: |-
        “3-bucket” state strategy; defines what belongs in RTK Query vs slices vs local state.
        State ownership strategy referenced by roadmap
    - Path: ui/src/components/DocCard.tsx
      Note: |-
        Domain card currently coupled to Search-specific CSS; high ROI target for design-system decoupling.
        High ROI styling decoupling target
    - Path: ui/src/components/MarkdownBlock.tsx
      Note: Shared markdown renderer used by DocViewer and Ticket overview.
    - Path: ui/src/components/PageHeader.tsx
      Note: Shared page header primitive; sets the direction for consistent page chrome.
    - Path: ui/src/components/ToastHost.tsx
      Note: |-
        Global toast UX primitive backed by Redux; replaced page-local timers.
        Global toast primitive for shared UX
    - Path: ui/src/features/search/SearchPage.tsx
      Note: Orchestrator example and remaining state to audit
    - Path: ui/src/features/ticket/TicketPage.tsx
      Note: Orchestrator example and tab-widget composition
    - Path: ui/src/styles/design-system.css
      Note: |-
        Shared design-system stylesheet; future componentization should prefer dm-* utilities here.
        Home of dm-* utilities and shared styling
ExternalSources: []
Summary: Roadmap for the next phase of React componentization and design-system consolidation, based on the refactors completed in ticket 007 (Search/Ticket modularization, shared primitives, and Redux/RTK Query state cleanup).
LastUpdated: 2026-01-05T12:21:55.511876646-05:00
WhatFor: Keep continuing UI refactors “small batch” while converging on a reusable widget/primitives architecture suitable for new Workspace navigation pages.
WhenToUse: When deciding whether to extract a widget vs. a primitive, where to put new components, and which remaining refactors are highest ROI (DocCard styling, dm-* utilities, shared list/preview patterns, and state audits).
---


# Post-refactor UI componentization roadmap

## Executive summary

Ticket `007-MODULARIZE-UI-WIDGETS` moved the UI from “a few large pages with embedded widgets” to a feature-first structure with:
- Route components acting as orchestrators (Search and Ticket are now much smaller).
- Widgets extracted into feature directories (Search widgets and Ticket tabs).
- A growing set of shared UI primitives (`PageHeader`, `ApiErrorAlert`, `MarkdownBlock`, etc).
- A clarified state model (RTK Query for server state, slices for shared/persistent intent, local state for ephemeral UI).
- A global toast system (`ToastHost` + `useToast`) that removed repeated `setTimeout` logic across routes.

The next phase is about **coherence**, not “more extraction for its own sake”:
1) Tighten a minimal “docmgr design system” layer (dm-* utilities + neutral primitives) without overbuilding.
2) Extract shared patterns only when they are already proven in multiple contexts (list+preview, chips, cards).
3) Establish a stable “shell + pages + widgets” structure so new Workspace pages don’t re-invent page chrome.

## What changed since the original analysis (ticket 001)

The original architecture doc (`001`) predicted that keeping Search as a kitchen-sink page would block Workspace navigation pages. That pressure is now relieved because we have working exemplars:

### Widget extraction patterns are proven
- Search widgets live under `ui/src/features/search/widgets/` and are composed by a thinner `SearchPage.tsx`.
- Ticket tabs live under `ui/src/features/ticket/tabs/` and are composed by a thinner `TicketPage.tsx`.

This is now the default pattern for future pages: “page orchestrator + widgets”.

### The shared primitives layer exists (but needs coherence)
Primitives already extracted:
- Feedback: `ApiErrorAlert`, `LoadingSpinner`, `EmptyState`, `ToastHost`
- Display: `MarkdownBlock`, `CodeBlock`, `DiagnosticCard`
- Page chrome: `PageHeader`, `PathHeader`, `RelatedFilesList`

Follow-up is to standardize naming/styling so shared primitives don’t inherit Search-only CSS or page-local spacing quirks.

### State ownership is clearer
We now have a consistent rule that’s been applied successfully:
- Server state → RTK Query cache (no “copy the response into local state”).
- Shared/persistent intent → Redux slice.
- Ephemeral UI mechanics → local `useState`/`useEffect`.

The next phase is an explicit audit (task 85) so we apply that rule consistently across the remaining pages/widgets.

## Component taxonomy (what we build next, and where)

To prevent “mega components” from creeping back in, use a small taxonomy with expectations:

### A) Primitives (UI kit layer)
Small reusable pieces that wrap Bootstrap patterns and encode docmgr UX conventions.
- Props in, markup out, minimal domain knowledge.
- Examples: `PageHeader`, `EmptyState`, `ToastHost`, `MarkdownBlock`, `CodeBlock`.

### B) Patterns (opinionated assemblies)
Reused assemblies that combine primitives and encode a layout pattern.
- Extract only once they exist in 2–3 places.
- Example candidates (once proven again in Workspace pages):
  - list + preview (desktop pane + mobile modal)
  - filters bar + active chips + clear
  - “meta table” rows for docs/tickets

### C) Domain components (docmgr concepts)
Reusable components that know docmgr types and semantics.
- Examples: `DocCard`, potential future `TicketCard`, `TopicCard`.
- Must not depend on Search-only CSS.

### D) Widgets (page sections)
Composable page blocks that own light UI state and wire interactions.
- Examples: Search widgets, Ticket tab bodies.

### E) Pages (route orchestrators)
Route params + URL sync + data fetching + widget composition.
- Size guideline: ~150–350 LOC; helpers should move out.

## Directory structure: evolve without a big-bang move

Current structure is workable (feature-first + shared `components/`). The risk is creating shared patterns inside feature widgets.

Incremental direction:
1) Keep feature-local UI in `ui/src/features/<feature>/widgets` and `ui/src/features/<feature>/components`.
2) Keep primitives in `ui/src/components/` for now.
3) Once shared components grow enough, introduce subfolders (no churn until it’s worth it):
   - `ui/src/components/ui/` (primitives)
   - `ui/src/components/domain/` (DocCard, TicketCard, etc)
   - `ui/src/components/patterns/` (list+preview, filters+chips, etc)

## High-ROI next componentization work

### 1) Decouple `DocCard` styling from Search-specific CSS (tasks 79, 83, 84)
Problem: `DocCard` is domain logic plus Search-derived class names (`result-*`), which blocks reuse in Workspace pages.

Preferred direction:
- Add a minimal dm-* namespace in `ui/src/styles/design-system.css`:
  - `.dm-mono` (paths/ticket ids)
  - `.dm-card` (neutral card density)
  - `.dm-meta` (small muted metadata rows)
  - `.dm-chip` (filter/topic chips)
  - `.dm-section` (consistent vertical spacing)
- Update `DocCard` to use dm-* (or a neutral `Card` primitive wrapper).
- Keep Search-only layout (split pane grid) in `search.css` only.

Success criteria:
- `DocCard` can be used in Search, Ticket, and a future Workspace “Recent docs” widget without inheriting Search-only styling.

### 2) Extract “list + preview” pattern (only after one more reuse)
We already have this pattern twice:
- Search docs results + preview pane/modal.
- Ticket documents tab list + preview.

Avoid extracting a generic pattern too early. Instead:
- Build one more consumer (Workspace “Recent docs” + preview) and then extract a neutral pattern component.
- Keep it layout-focused and pass rendering functions (`renderRow`, `renderPreview`) so it stays reusable.

### 3) CSS design-system consolidation (tasks 66–68)
Goal: make it easy to build new pages without creating page-local CSS for basic typography/spacing.
- Keep `search.css` as “layout only”.
- Prefer dm-* utilities for spacing/typography/card density.

### 4) State audit (task 85)
Deliverable should be a table:
`state field` → `owner (local/slice/RTK)` → `why` → `migration plan`.

This audit should drive any future “move widget-local state into slices” work, rather than doing it by intuition.

## Refactor playbook (keep it safe)

Rules that have worked and should continue:
- Extract in-context: don’t create “shared” abstractions until a second consumer exists.
- Keep commits small (1–2 extractions/retrofitting areas).
- Prefer deletion over addition when refactoring (prove ROI by removing duplicates).
- Validate frequently (`pnpm -C ui lint`, `pnpm -C ui build`).

## Suggested short-term sequence

1) Implement dm-* utilities and decouple `DocCard` styling (unblocks reuse everywhere).
2) Add a thin shell widget set (`TopBar`, `SideNav`, `AppShell`) and use it for new Workspace pages first.
3) Create one Workspace page that reuses existing patterns (e.g., “Recent docs”) to justify extracting list+preview.
4) Run the state audit and decide which remaining state deserves a slice.
