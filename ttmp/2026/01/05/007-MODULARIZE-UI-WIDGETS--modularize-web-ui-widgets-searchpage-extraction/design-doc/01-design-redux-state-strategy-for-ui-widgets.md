---
Title: 'Design: Redux state strategy for UI widgets'
Ticket: 007-MODULARIZE-UI-WIDGETS
Status: active
Topics:
    - ui
    - web
    - ux
    - docmgr
    - refactor
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/features/search/SearchPage.tsx
      Note: Current mixed local+Redux state; primary migration target.
    - Path: ui/src/features/search/hooks/useSearchUrlSync.ts
      Note: URL restore/write remains as a hook even with Redux.
    - Path: ui/src/features/search/searchSlice.ts
      Note: Existing Redux intent state for Search (mode/query/filters).
    - Path: ui/src/services/docmgrApi.ts
      Note: RTK Query server-state layer; target owner for search results.
ExternalSources: []
Summary: Guidelines and a migration plan for moving appropriate widget/page state into Redux Toolkit slices and RTK Query, while keeping ephemeral UI state local.
LastUpdated: 2026-01-05T09:33:19.168765976-05:00
WhatFor: Reduce page complexity and duplication by standardizing which UI state lives in Redux/RTK Query vs local React state, starting with SearchPage.
WhenToUse: When refactoring large pages into widgets/hooks or adding new pages so state ownership stays consistent across the app.
---


# Design: Redux state strategy for UI widgets

## Executive Summary

We should not “move everything into Redux”. We should move *shared and/or persistent state* into Redux Toolkit slices, move *server state* into RTK Query (already Redux-backed), and keep *ephemeral UI state* local (`useState`) and side effects local (`useEffect`) where it wires to browser APIs.

For the docmgr web UI, the biggest immediate win is on the Search page: stop keeping “docsResults/docsNextCursor/diagnostics” as page-local state and instead treat them as RTK Query state (cached in Redux). Redux slices should own the query intent (mode/query/filters) and any cross-widget UI state we want to persist/share (optionally selection), but not transient toggles like “shortcuts modal open”.

## Problem Statement

The current UI has a mixed state story:
- Redux (`searchSlice`) holds query intent: mode, query string, filters.
- RTK Query (`docmgrApi`) holds server state, but Search currently uses *lazy queries* and then copies results into local `useState`.
- Pages repeatedly implement toast/error/copy patterns with local state.

This leads to:
- duplicated state (RTK Query cache + local state for the same response),
- harder-to-reason-about behavior (URL restore + selection + “auto-run search”),
- increased coupling within large page components.

We want a coherent policy: “what belongs in Redux/RTK Query” vs “what stays local”.

## Proposed Solution

### 1) Classify UI state into 3 buckets

**A) Server state (move into RTK Query; do not duplicate in useState)**
- Anything that is a response to an API request and can be re-derived from a query key:
  - Search docs results + diagnostics + pagination cursors
  - Ticket docs/tasks/graph
  - Workspace status/summary

**B) Shared/persistent client state (Redux slices)**
- Inputs and settings that:
  - are shared across multiple widgets on a page, or
  - are persisted in the URL / expected to survive navigation, or
  - represent “user intent” rather than “one-off UI toggle”.

Examples (Search):
- already in slice: `mode`, `query`, `filters.*`
- optional slice additions:
  - `selectedPath` (if we decide it should persist across widgets/routes beyond URL)
  - `lastSubmittedAt` / `hasSearched` (if we want consistent “empty state” semantics without local booleans)

**C) Ephemeral UI state (keep local `useState`)**
- Modal open/closed, drawers, temporary toasts, input focus refs, transient form text:
  - `showShortcuts`, `showFilterDrawer`, `showPreviewModal`
  - “toast visible for 2 seconds”
  - `newTaskText` in Ticket page

These should remain local unless we explicitly want global behavior (e.g. a global toast host).

### 2) Adopt a “single source of truth” rule

For any API call, exactly one of the following owns the result:
- RTK Query cache (preferred), or
- a Redux slice (only if we need a custom cache/merge model that RTK Query cannot express cleanly).

Avoid: “RTK Query fetch → copy response into local state” unless you are implementing an explicit *client-side transform* that needs a second representation.

### 3) Search page migration target

Current SearchPage state is still a mix:
- **Redux slice**: query intent (`searchSlice`)
- **Local state**: UI toggles + “server response copies” (docsResults, totals, cursor, diagnostics)

Target:
- `searchSlice` continues to own mode/query/filters (and optionally selection path if desired).
- `docmgrApi.searchDocs` query becomes the owner of docs results + diagnostics + cursor.
- For pagination/“Load more”, use one of:
  1) **RTK Query merge** (`serializeQueryArgs` + `merge` + `forceRefetch`) to accumulate results per query key, or
  2) a small “searchResultsSlice” keyed by a stable query hash (store `results[]`, `nextCursor`) while the network fetch remains RTK Query.

Option (1) is preferred because it keeps server state in RTK Query, but it requires careful cache key design to avoid cross-query contamination.

### 4) URL sync and side effects stay as hooks

Even with Redux, we still need `useEffect` for:
- reading `window.location.search` on mount and dispatching `setMode/setQuery/setFilter`
- writing debounced URL updates (or using router navigation helpers)
- global keyboard listeners (unless you want a dedicated hotkeys component mounted once at app root)

The key is: the hook should dispatch actions and read selectors; it should not also own server state.

## Design Decisions

### Decision: Keep RTK Query as the owner of server state
Rationale:
- It is already Redux-backed.
- It provides request de-duplication, caching, and invalidation.
- It keeps “API state” consistent across pages and future widgets.

### Decision: Keep ephemeral UI toggles local by default
Rationale:
- Putting every modal boolean into Redux tends to increase global coupling and re-render scope.
- Local state is simpler, more testable, and matches React’s model for ephemeral view state.

### Decision: Prefer URL as persistence for shareable view state
Rationale:
- The product already supports “restore state on reload/share link” in Search.
- The URL is the most user-visible persistence mechanism; Redux alone does not give shareable links.

### Decision: Introduce slice state only when it’s “intent” or “shared”
Rationale:
- Redux state should be meaningful at the feature level, not just a mirror of component internals.

## Alternatives Considered

### Alternative: Put everything in Redux (no useState)
Rejected because:
- it over-centralizes ephemeral view state,
- increases boilerplate and coupling,
- makes local UI interactions harder to reason about.

### Alternative: Avoid Redux entirely, use only component state + hooks
Rejected because:
- query intent is shared across multiple components/widgets,
- RTK Query is already established and beneficial,
- consistent state ownership matters as we add Workspace pages.

## Implementation Plan

### Phase 0: Pre-work (documentation + guardrails)
- Keep validating each extraction with `pnpm -C ui lint` and `pnpm -C ui build`.
- Keep refactors behavior-preserving until the state move is explicitly validated.

### Phase 1: Search results: move “server copies” into RTK Query
- Keep `useLazySearchDocsQuery`, but remove local “server copies” (`docsResults/docsTotal/docsNextCursor/docsDiagnostics`) from `SearchPage.tsx` and render from the RTK Query cache (`searchDocsState.data`).
- Implement “Load more” via RTK Query merge:
  - `serializeQueryArgs` excludes `cursor` so all pages for the same query intent share one cache entry.
  - `merge` appends new pages (deduped by `ticket:path`) and updates `nextCursor/total/diagnostics`.
  - `forceRefetch` returns true when `cursor` changes so pagination fetches don’t get skipped.
- Ensure “Clear” resets the lazy query state (`searchDocsState.reset()` / `searchFilesState.reset()`) so the UI returns to the uninitialized/empty state.

Status: implemented in commit `0cd16e6`.

### Phase 2: Selection persistence decision
- Keep selection local (current), or:
- Add `selectedPath` to `searchSlice` if other widgets/pages need to observe it.

### Phase 3: Optional global toast host
- If we want consistent toast UX across pages, add a `uiSlice` with a toast queue and a `ToastHost` mounted in `App.tsx`.
- Otherwise keep page-local toasts.

### Phase 4: Apply the same policy to upcoming Workspace pages
- For each new page:
  - server state → RTK Query
  - user intent / shareable state → slice + URL
  - ephemeral toggles → local state

## Open Questions

1) For Search pagination, do we want RTK Query “merge cache” behavior or a separate slice for aggregated results?
   - Chosen: RTK Query “merge cache” (implemented in `0cd16e6`).
2) Do we want selection path in Redux, or is URL persistence sufficient?
3) Should we standardize a global toast host now, or wait until Workspace pages exist?

## References

- Architecture analysis (ticket 001): `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`
- Modularization ticket tasks: `ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/tasks.md`
- Search intent state: `ui/src/features/search/searchSlice.ts`
- API layer (RTK Query): `ui/src/services/docmgrApi.ts`
