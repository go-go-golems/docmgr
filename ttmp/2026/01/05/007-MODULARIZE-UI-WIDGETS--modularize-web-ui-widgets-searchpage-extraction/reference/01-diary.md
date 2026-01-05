---
Title: Diary
Ticket: 007-MODULARIZE-UI-WIDGETS
Status: active
Topics:
    - ui
    - web
    - ux
    - docmgr
    - refactor
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/components/CodeBlock.tsx
      Note: Shared code rendering used by FileViewer
    - Path: ui/src/components/DiagnosticCard.tsx
      Note: Shared diagnostic rendering extracted from DocViewer
    - Path: ui/src/components/DocCard.tsx
      Note: Decoupled from Search-specific styling; now reusable across pages
    - Path: ui/src/components/MarkdownBlock.tsx
      Note: Shared markdown rendering primitive
    - Path: ui/src/components/PageHeader.tsx
      Note: Shared page header primitive used across routes
    - Path: ui/src/components/PathHeader.tsx
      Note: Switched to dm-path-pill
    - Path: ui/src/components/ToastHost.tsx
      Note: Global toast rendering and timeout management
    - Path: ui/src/features/doc/DocViewerPage.tsx
      Note: |-
        Uses DiagnosticCard primitive
        No more page-local toast timers
    - Path: ui/src/features/file/FileViewerPage.tsx
      Note: |-
        Uses CodeBlock primitive
        No more page-local toast timers
    - Path: ui/src/features/search/SearchPage.tsx
      Note: No more page-local toast timers
    - Path: ui/src/features/search/widgets/SearchDocsResults.tsx
      Note: Uses DocCard doc-object API
    - Path: ui/src/features/search/widgets/SearchFilesResults.tsx
      Note: Switched to dm-* card/path styles
    - Path: ui/src/features/search/widgets/SearchHeader.tsx
      Note: Search page header widget retrofit
    - Path: ui/src/features/ticket/TicketPage.tsx
      Note: No more page-local toast timers
    - Path: ui/src/features/ticket/components/TicketHeader.tsx
      Note: Ticket header widget retrofit
    - Path: ui/src/features/ticket/tabs/TicketDocumentsTab.tsx
      Note: Uses DocCard doc-object API
    - Path: ui/src/features/toast/toastSlice.ts
      Note: Redux slice for toast queue
    - Path: ui/src/features/toast/useToast.ts
      Note: Hook used by pages
    - Path: ui/src/lib/time.ts
      Note: Shared time/date formatting helpers
    - Path: ui/src/styles/design-system.css
      Note: dm-* utility aliases for reusable cards/path pills
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-05T10:53:27.605970965-05:00
WhatFor: ""
WhenToUse: ""
---





# Diary

## Goal

Record the incremental widget/component refactor work for ticket `007-MODULARIZE-UI-WIDGETS`, including the reasoning behind extractions, validation commands, and follow-up tasks.

## Step 1: Set up the modularization ticket + task plan

I created a dedicated ticket to track the Search/Ticket page modularization work as a sequence of small, low-risk extractions. The intent was to keep behavior stable while moving toward a coherent widget architecture that we can reuse for upcoming pages (Workspace/Tickets/Topics/Recent).

This step established the “refactor rules of engagement” (small commits; run lint/build; extract shared primitives only once they’ve been exercised in-context) so later steps could move quickly without ambiguity.

### What I did
- Created ticket `007-MODULARIZE-UI-WIDGETS` and wrote a detailed task list in `tasks.md`
- Started a changelog trail in `changelog.md` for each batch of extractions

### Why
- Keep refactor scope explicit and reviewable
- Make progress measurable (checklist) and easy to continue between sessions

### What worked
- Task list sequencing made it easy to keep commits small and focused

### What didn't work
- N/A

### What I learned
- “Extract shared primitives later” is easier to enforce when tasks explicitly separate widgetization vs. shared primitives

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Keep RelatedFiles lists tight (avoid ballooning ticket index relations)

### Code review instructions
- Start with `ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/tasks.md`
- Validate later steps by running `pnpm -C ui lint` and `pnpm -C ui build`

### Technical details
- Ticket path: `ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/`

## Step 2: Split SearchPage into widgets (in-context primitives)

SearchPage was the main “kitchen sink” component, so we split it into widget files and smaller hooks/components while preserving behavior. Shared primitives were only extracted when we hit duplication in multiple widgets/pages (errors/spinners/empty states/path headers).

The outcome is that `SearchPage.tsx` is now mostly an orchestrator: it wires URL-sync, selection, RTK Query, and delegates UI to small widget components under `features/search/widgets/`.

**Commit (code):** 159be3b — "Search: split page into widgets"

### What I did
- Extracted Search widgets into `ui/src/features/search/widgets/`
- Introduced shared primitives in `ui/src/components/` (error/spinner/empty state/path headers/page header)
- Split Search CSS to keep layout rules scoped (`ui/src/styles/search.css`)
- Validated with:
  - `pnpm -C ui lint`
  - `pnpm -C ui build`

### Why
- Prepare for more pages without duplicating “page-local widget” code patterns
- Make it easier to reason about state ownership boundaries per widget

### What worked
- Widget boundary: “pure UI props in, callbacks out” kept extraction easy

### What didn't work
- N/A

### What I learned
- Extracting shared primitives only after a second use keeps the “design system” from becoming speculative

### What was tricky to build
- Keeping URL selection/preview behavior stable while breaking up render trees

### What warrants a second pair of eyes
- Search URL sync + selection ordering (ensure no double-apply of selection on mount)

### What should be done in the future
- Add a manual UX pass for Search hotkeys + URL restore (tracked in `tasks.md`)

### Code review instructions
- Start with `ui/src/features/search/SearchPage.tsx`
- Then browse `ui/src/features/search/widgets/` for extracted components

### Technical details
- Shared primitives introduced/used: `ui/src/components/ApiErrorAlert.tsx`, `ui/src/components/LoadingSpinner.tsx`, `ui/src/components/EmptyState.tsx`, `ui/src/components/PathHeader.tsx`, `ui/src/components/PageHeader.tsx`, `ui/src/components/RelatedFilesList.tsx`

## Step 3: Move Search docs “server state” into RTK Query (pagination merge)

Search docs results were previously copied into local component state, which made pagination and “Clear” semantics harder to reason about and more bug-prone. This step moved docs result ownership into RTK Query cache and implemented cursor-based pagination via `merge` while keeping UI behavior the same.

This clarified our state model: RTK Query owns server results, the Search slice owns persistent UI intent, and local state is reserved for ephemeral UI concerns (e.g., transient selection UI mechanics).

**Commit (code):** 0cd16e6 — "Search: keep docs results in RTK Query (pagination merge + clear/reset)"

### What I did
- Updated `ui/src/services/docmgrApi.ts` `searchDocs` endpoint to merge pages and ignore cursor in the cache key
- Removed local docs results copies from `SearchPage.tsx` and rendered directly from RTK Query data
- Kept “Clear” behavior by resetting the lazy query state

### Why
- Avoid duplicated sources-of-truth for results
- Make pagination accumulation explicit and testable at the API layer

### What worked
- RTK Query `serializeQueryArgs` + `merge` + `forceRefetch` model matched the cursor pagination needs cleanly

### What didn't work
- N/A

### What I learned
- Cursor parameters often should *not* be part of the cache key when you want incremental accumulation into a single result set

### What was tricky to build
- Ensuring “Clear” resets the visible results without leaving stale merged cache in view

### What warrants a second pair of eyes
- Confirm `forceRefetch` conditions don’t accidentally refetch when unrelated query parts are stable

### What should be done in the future
- Add a manual UX pass for “docs load more + clear + URL restore” (tracked in `tasks.md`)

### Code review instructions
- Start with `ui/src/services/docmgrApi.ts` searchDocs endpoint
- Then check `ui/src/features/search/SearchPage.tsx` docs query usage and clear/reset path

### Technical details
- Cursor excluded from cache key: `serializeQueryArgs`
- Accumulate pages: `merge`
- Refetch when cursor changes: `forceRefetch`

## Step 4: Split TicketPage into tab widgets

TicketPage had grown into the next largest “everything page”, so we applied the same pattern as Search: keep the route component small and extract tab bodies into separate files. This sets the precedent for future multi-section pages (Workspace dashboard, Tickets listing) to be composed of small widgets rather than a single mega component.

**Commit (code):** f2461be — "Ticket: split page into tab widgets"

### What I did
- Extracted `TicketHeader` and `TicketTabs`
- Extracted tab bodies into `ui/src/features/ticket/tabs/` (overview/documents/tasks/graph/changelog)
- Preserved preview and selection behavior in the Documents tab

### Why
- Match SearchPage refactor style for consistency
- Make it easier to work on one tab’s UI/state at a time

### What worked
- Tab bodies as “page-widgets” stayed clean because the orchestrator passes down narrow props

### What didn't work
- N/A

### What I learned
- A “tab widget” boundary is a natural place to keep local `useState` for ephemeral form inputs (e.g., `newTaskText`)

### What was tricky to build
- Keeping query `skip` logic correct per-tab to avoid unnecessary fetching while navigating tabs

### What warrants a second pair of eyes
- Ensure tab switching doesn’t leave behind stale `doc` URL params in non-Documents tabs (or vice versa)

### What should be done in the future
- Follow-up: unify headers and toast patterns across Search/Ticket/Doc/File

### Code review instructions
- Start with `ui/src/features/ticket/TicketPage.tsx`
- Then inspect `ui/src/features/ticket/tabs/` and `ui/src/features/ticket/components/`

### Technical details
- Ticket Documents preview now reuses `ui/src/components/PathHeader.tsx` + `ui/src/components/RelatedFilesList.tsx`

## Step 5: Retrofit page headers + unify date formatting helper

With Search and Ticket now decomposed into widgets, header duplication became more obvious. This step standardizes how pages render their “title + subtitle + actions” block and introduces a shared date formatting helper to stop copy/pasting `new Date(...).toLocaleString()` in multiple places.

This doesn’t change data flow, but it makes it easier to keep typography/spacing consistent as we add new Workspace-style pages.

**Commit (code):** d80c73f — "UI: retrofit headers and date formatting"

### What I did
- Added `formatDate()` to `ui/src/lib/time.ts`
- Extended `ui/src/components/PageHeader.tsx` to support `titleClassName` and configurable bottom margin
- Refactored `ui/src/features/search/widgets/SearchHeader.tsx` and `ui/src/features/ticket/components/TicketHeader.tsx` to use `PageHeader`
- Removed TicketPage’s local `formatDate` helper and imported the shared one

### Why
- Reduce header layout duplication across pages
- Keep date formatting consistent and reusable

### What worked
- `PageHeader` fit both “simple page title” (Doc/File) and “Search header with action button” without losing behavior

### What didn't work
- N/A

### What I learned
- Keeping `PageHeader` flexible via class props is a low-cost way to avoid adding multiple header variants too early

### What was tricky to build
- Avoiding accidental empty subtitle rows in `PageHeader` when a widget had “optional” subtitle parts

### What warrants a second pair of eyes
- Visual spacing/regression: compare Search header spacing before/after (`mb-4` vs `mb-3`)

### What should be done in the future
- Replace page-local toast timers with a shared toast host (`tasks.md` task 76)

### Code review instructions
- Start with `ui/src/components/PageHeader.tsx`
- Then check `ui/src/features/search/widgets/SearchHeader.tsx` and `ui/src/features/ticket/components/TicketHeader.tsx`

### Technical details
- Shared date helper: `ui/src/lib/time.ts` `formatDate(iso?: string)`

## Step 6: Introduce MarkdownBlock primitive for consistent markdown rendering

We already had multiple pages rendering markdown (DocViewer “Content” and TicketOverview’s `index.md`) with similar glue (`ReactMarkdown`, `remark-gfm`, and sometimes code highlighting). This step centralizes that into a single component so future pages/widgets can render markdown consistently and we can evolve styling/behavior in one place.

This is also a “design system” anchor: markdown rendering is a foundational widget primitive (cards, docs previews, workspace summaries).

**Commit (code):** 866399f — "UI: add MarkdownBlock primitive"

### What I did
- Added `ui/src/components/MarkdownBlock.tsx` (ReactMarkdown + remarkGfm + rehypeHighlight)
- Refactored `ui/src/features/doc/DocViewerPage.tsx` and `ui/src/features/ticket/tabs/TicketOverviewTab.tsx` to use it

### Why
- Remove duplicated markdown glue across pages
- Make code highlighting consistent wherever markdown is rendered

### What worked
- Global highlight stylesheet was already imported (`ui/src/index.css`), so switching to `rehype-highlight` remained consistent

### What didn't work
- N/A

### What I learned
- Markdown rendering is “shared primitive” territory quickly; extracting it early tends to pay off because many pages need it

### What was tricky to build
- Keeping the component interface minimal (`markdown` string + optional highlight toggle) without leaking ReactMarkdown internals everywhere

### What warrants a second pair of eyes
- Verify markdown rendering parity on TicketOverview’s `index.md` (especially code fences + tables)

### What should be done in the future
- Consider whether Search snippets should reuse `MarkdownBlock` (or stay specialized because of query term highlighting)

### Code review instructions
- Start with `ui/src/components/MarkdownBlock.tsx`
- Then check replacement sites in `ui/src/features/doc/DocViewerPage.tsx` and `ui/src/features/ticket/tabs/TicketOverviewTab.tsx`

### Technical details
- Plugins: `remark-gfm` (tables/task lists), `rehype-highlight` (code fences)

## Step 7: Extract CodeBlock + DiagnosticCard primitives

After introducing `MarkdownBlock`, the next obvious UI duplication was “render highlighted code HTML in a styled `<pre>`” (FileViewer) and “render parse diagnostic taxonomy” (DocViewer). This step extracted both into `ui/src/components/` so other pages/widgets can reuse them without re-copying markup and classnames.

The goal is to keep primitives small and narrow (1 job, 1 file), so future extractions (toasts, card styling) have a stable foundation.

**Commit (code):** a832d20 — "UI: extract CodeBlock and DiagnosticCard"

### What I did
- Added `ui/src/components/CodeBlock.tsx` and used it in `ui/src/features/file/FileViewerPage.tsx`
- Added `ui/src/components/DiagnosticCard.tsx` and used it in `ui/src/features/doc/DocViewerPage.tsx`

### Why
- Reduce duplication and keep code/diagnostic rendering consistent
- Make it easier to improve styling in one place later (design system cleanup)

### What worked
- FileViewer and DocViewer both swapped cleanly because the primitives are “render-only”

### What didn't work
- N/A

### What I learned
- Extracting “leaf primitives” tends to be the safest kind of refactor (no behavior changes, just markup consolidation)

### What was tricky to build
- Ensuring `CodeBlock` preserved the exact `hljs` class naming so existing highlight CSS still applies

### What warrants a second pair of eyes
- Visual parity check for FileViewer code block scroll height and code highlighting

### What should be done in the future
- Consider whether Search file preview should reuse `CodeBlock` if it ever renders highlighted code HTML

### Code review instructions
- Start with `ui/src/components/CodeBlock.tsx` and `ui/src/components/DiagnosticCard.tsx`
- Verify replacement sites:
  - `ui/src/features/file/FileViewerPage.tsx`
  - `ui/src/features/doc/DocViewerPage.tsx`

### Technical details
- `CodeBlock` expects already-highlighted HTML (`hljs.highlight(...).value`) and renders it via `dangerouslySetInnerHTML`

## Step 8: Expand follow-up tasks for remaining cleanup

With the core widget split done, the remaining work is mostly “cohesion”: toasts, neutral card styling, and a small set of `.dm-*` utilities so future pages can compose widgets without inventing new ad-hoc CSS. I expanded the task list to make these phases explicit and to separate “build the primitive” from “retrofit pages”.

This also adds an explicit audit task for `useState/useEffect` usage so we can decide (with evidence) what belongs in local component state vs. a Redux slice vs. RTK Query cache.

### What I did
- Added follow-up tasks (80–86) to break down toast/design-system/widget-skeleton work into actionable steps

### Why
- Reduce ambiguity about “what’s next” and keep the remaining refactor work small-batch and reviewable

### What worked
- `docmgr task add` made it easy to extend the checklist without manual renumbering

### What didn't work
- N/A

### What I learned
- The “hard part” is no longer splitting pages—it’s establishing shared primitives without prematurely over-generalizing them

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Start with toast system (tasks 80–82) since it removes duplicated timers across 4 pages

### Code review instructions
- Review new tasks at `ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/tasks.md`

### Technical details
- N/A

## Step 9: Validate UI build + lint after primitives work

After adding shared primitives and updating docs, I ran the UI lint and build to ensure TypeScript and bundling still succeed. This is the main safety net for these extraction-heavy refactors since we’re mostly reorganizing component boundaries and imports.

No failures surfaced in this pass, so the current state is safe to continue from (next likely step: ToastHost).

### What I did
- Ran `pnpm -C ui lint`
- Ran `pnpm -C ui build`

### Why
- Catch missing imports/exports, type errors, and accidental circular deps early

### What worked
- Both commands passed

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Keep lint/build in the loop when tackling the toast refactor (it will touch many files)

### Code review instructions
- Run `pnpm -C ui lint` and `pnpm -C ui build`

### Technical details
- N/A

## Step 10: Add a global toast system and remove per-page timers

Multiple pages were each managing their own toast UI state and `setTimeout` cleanup logic. That creates duplicated code, inconsistent timeouts/messages, and subtle leak risks when navigating quickly. This step introduces a single global toast queue in Redux and a `ToastHost` that owns timeout cleanup, then retrofits Search/Ticket/Doc/File to use `useToast()`.

This is one of the highest-ROI extractions so far because it removes copy/paste logic across four routes and establishes a shared “UX primitive” that Workspace pages will also need.

**Commit (code):** 8be000a — "UI: add global toast system"  
**Commit (code):** fc60503 — "UI: use ToastHost in doc/file viewers"  
**Commit (code):** 2e44767 — "UI: use ToastHost in TicketPage"  
**Commit (code):** 5ef282d — "UI: use ToastHost in SearchPage"

### What I did
- Added `ui/src/features/toast/toastSlice.ts` (Redux toast queue) and `ui/src/features/toast/useToast.ts` (imperative API)
- Added `ui/src/components/ToastHost.tsx` and mounted it in `ui/src/App.tsx`
- Updated `ui/src/app/store.ts` to include `toast` reducer
- Removed page-local toast state + `setTimeout` cleanup from:
  - `ui/src/features/search/SearchPage.tsx`
  - `ui/src/features/ticket/TicketPage.tsx`
  - `ui/src/features/doc/DocViewerPage.tsx`
  - `ui/src/features/file/FileViewerPage.tsx`

### Why
- Avoid duplicated timer logic and inconsistent behavior
- Make it easy for any widget/page to “toast” without owning UI layout

### What worked
- Redux slice + host pattern kept page changes small (mostly deleting code)

### What didn't work
- N/A

### What I learned
- Centralizing time-based UI cleanup (timeouts) is safer than scattering `setTimeout` across route components

### What was tricky to build
- Ensuring timers are cleaned up both when a toast is removed and when the host unmounts (navigation or HMR)

### What warrants a second pair of eyes
- UX parity: toast positioning/stacking vs the old Search-only toast container

### What should be done in the future
- Consider a consistent message vocabulary (“Copied” vs “Copied path: …”) once DocCard styling work starts

### Code review instructions
- Start with `ui/src/features/toast/toastSlice.ts` and `ui/src/components/ToastHost.tsx`
- Verify the four retrofit sites listed above compile and behave the same

### Technical details
- `ToastHost` uses a `Map<toastId, timeoutHandle>` to ensure one timer per toast and to avoid leaking timers

## Step 11: Decouple DocCard from Search-only styling and reuse it across pages

`DocCard` started life as a “Search result card”, which meant it carried Search-centric class names (`result-*`, `topic-badge`, `copy-btn`) and an API that required the caller to pass a bunch of parallel props. Now that we’re actively building a coherent widget/primitives system, that coupling becomes a blocker: we want the same doc list item to render in Search, Ticket, and future Workspace widgets without inheriting Search page concerns.

This step refactors `DocCard` into a domain component with a single `doc` object prop, introduces a `dm-*` utility namespace in the shared design-system stylesheet, and updates Search/Ticket callsites to use the new API. The immediate goal is to keep behavior stable but make DocCard re-usable in more contexts with less prop churn.

### What I did
- Updated `ui/src/styles/design-system.css` to add `dm-*` utility aliases for card/path/snippet styling and removed the now-unused `.toast-container` rule
- Refactored `ui/src/components/DocCard.tsx`:
  - Accept `doc: DocCardDoc` instead of parallel props
  - Use `dm-*` classnames (neutral styling)
  - Add Enter/Space keyboard activation for accessibility
- Updated callsites:
  - `ui/src/features/search/widgets/SearchDocsResults.tsx` now passes `doc={r}`
  - `ui/src/features/ticket/tabs/TicketDocumentsTab.tsx` now passes `doc={{...d, ticket}}`
- Updated remaining `result-*` consumers to use dm-*:
  - `ui/src/features/search/widgets/SearchFilesResults.tsx`
  - `ui/src/components/PathHeader.tsx`

### Why
- Make `DocCard` a true domain component (usable outside Search)
- Establish a minimal `dm-*` design-system namespace to prevent “page CSS leaks” as Workspace pages land

### What worked
- The “doc object” prop shape reduced callsite verbosity and makes it easier to thread through doc types consistently
- Keeping old CSS selectors as aliases avoided surprises while we migrate remaining usages

### What didn't work
- N/A

### What I learned
- When a component is used in multiple pages, switching from “parallel props” → “single domain object + small overrides” lowers maintenance cost quickly

### What was tricky to build
- Keeping styling stable while renaming classnames; the alias approach lets us migrate gradually without breaking consumers

### What warrants a second pair of eyes
- Visual parity: confirm Search docs results and Ticket docs tab still look correct (hover/selected/copy button behavior)

### What should be done in the future
- Consider a small “pattern” component for list+preview once Workspace pages add a third consumer (avoid premature abstraction)

### Code review instructions
- Start with `ui/src/components/DocCard.tsx` and `ui/src/styles/design-system.css`
- Then verify the refactored usages:
  - `ui/src/features/search/widgets/SearchDocsResults.tsx`
  - `ui/src/features/ticket/tabs/TicketDocumentsTab.tsx`
  - `ui/src/features/search/widgets/SearchFilesResults.tsx`
  - `ui/src/components/PathHeader.tsx`

### Technical details
- New classnames are dm-* but the old selectors remain as aliases temporarily for safer migration
