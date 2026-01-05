---
Title: 'Playbook: Implementing new UI pages and widgets (post-refactor)'
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - tooling
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design-doc/01-design-workspace-navigation-ui-post-refactor.md
      Note: Primary UI design spec
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md
      Note: Design target for new pages
    - Path: ui/src/components/DocCard.tsx
      Note: Reusable domain card
    - Path: ui/src/components/ToastHost.tsx
      Note: Global toast primitive
    - Path: ui/src/services/docmgrApi.ts
      Note: RTK Query endpoint patterns
    - Path: ui/src/styles/design-system.css
      Note: dm-* design system utilities
ExternalSources: []
Summary: Step-by-step playbook for adding new UI pages/widgets to the docmgr React SPA using the post-refactor architecture (feature widgets, shared primitives, RTK Query ownership, dm-* design system, and docmgr+git hygiene).
LastUpdated: 2026-01-05T13:15:40.751635287-05:00
WhatFor: Keep new Workspace pages consistent and prevent regressions into “mega page components” by following a repeatable implementation workflow and directory/state conventions.
WhenToUse: Whenever creating a new route/page, extracting a widget/pattern, adding a new RTK Query endpoint, or doing a cross-page refactor of shared UI primitives.
---


# Playbook: Implementing new UI pages and widgets (post-refactor)

## Purpose

Provide a repeatable, low-risk workflow for implementing new pages and widgets for the docmgr Web UI (Workspace navigation pages in particular), covering:
- Where new files should live (features vs shared components vs lib vs styles)
- Which primitives/patterns to reuse
- State ownership rules (RTK Query vs slices vs local)
- How to validate (lint/build) and how to keep docs (docmgr) and commits clean

## Environment Assumptions

- Repo root: `docmgr/` (this playbook assumes commands run from this directory)
- Tooling: `pnpm`, Node, and Go installed
- UI is a Vite SPA under `ui/` and the backend serves it via embed in production
- Server runs locally at `127.0.0.1:3001` (API base is `/api/v1`)
- Git working tree is clean before starting

Optional but recommended:
- Two-terminal dev loop:
  - terminal A: `make dev-backend`
  - terminal B: `make dev-frontend`

## Commands

```bash
# 0) Baseline sanity (before you touch anything)
git status --porcelain
pnpm -C ui lint
pnpm -C ui build

# 1) (Optional) run dev loop while iterating
# make dev-backend
# make dev-frontend
```

## Implementation Workflow (repeatable)

### 1) Pick the right “home” for new code

Use these locations by default:
- **Route/page orchestrators**: `ui/src/features/<feature>/<X>Page.tsx`
  - Workspace pages live under `ui/src/features/workspace/`
  - Keep orchestrators small; push UI sections into widgets.
- **Feature-only widgets**: `ui/src/features/<feature>/widgets/`
- **Feature-only helper components** (not full widgets): `ui/src/features/<feature>/components/`
- **Feature hooks**: `ui/src/features/<feature>/hooks/`
- **Shared primitives / domain components**: `ui/src/components/`
- **Shared utilities**: `ui/src/lib/` (e.g., `time.ts`, `clipboard.ts`, `apiError.ts`)
- **API layer**: `ui/src/services/docmgrApi.ts` (RTK Query)
- **Styles**:
  - Shared: `ui/src/styles/design-system.css` (prefer dm-* utilities)
  - Page layout only: `ui/src/styles/<page>.css` (e.g., `search.css`)

Rule of thumb:
- If it is used in 2+ routes, it’s a candidate for `ui/src/components/`.
- If it is used only within one feature, keep it in that feature directory.
- Avoid “speculative” shared patterns: extract only after a second real consumer exists.

### 2) Use the existing primitives first

Before adding new markup helpers, look for (and prefer) existing primitives:
- `PageHeader`, `PathHeader`
- `ApiErrorAlert`, `LoadingSpinner`, `EmptyState`
- `ToastHost` + `useToast`
- `MarkdownBlock`, `CodeBlock`, `DiagnosticCard`
- `RelatedFilesList`, `DocCard`

### 3) State ownership rules (post-refactor)

Use the “3 buckets” strategy:
- **Server state** (responses, caches, pagination merges): RTK Query.
  - Don’t copy RTK Query `data` into local `useState` unless you are explicitly building a draft/edit buffer.
- **Shared/persistent UI intent** (filters, sort, view modes, time range): Redux slice.
  - Create a slice only when 2+ widgets need it or when it must persist across navigation.
- **Ephemeral UI mechanics** (drawer open, modal open, hover, local input text): local `useState`/`useEffect`.

### 4) RTK Query patterns to reuse

- Prefer adding endpoints in `ui/src/services/docmgrApi.ts` with clear request/response types.
- Use tags for invalidation (keep them coarse: `Workspace`, `Ticket`, `Search`).
- For cursor pagination that accumulates into one list: use `serializeQueryArgs` + `merge` + `forceRefetch` (pattern already used for Search docs).

### 5) Design-system / CSS rules

- Prefer `dm-*` utilities (in `ui/src/styles/design-system.css`) over inventing new ad-hoc CSS.
- Keep page-specific layout rules in page stylesheets (e.g., split panes, sticky sidebars).
- Avoid coupling shared components to a specific page’s CSS (e.g., don’t require Search-only classes).

### 6) Add the route + stub page first (then fill widgets)

For new Workspace routes:
- Add the route in `ui/src/App.tsx` (nested under `/workspace` where possible).
- Create the page as a thin orchestrator.
- Add widgets one at a time, and keep each widget focused.

### 7) Validate and commit in small batches

```bash
# Validate after each small batch
pnpm -C ui lint
pnpm -C ui build

# Commit code changes (small, scoped)
git status --porcelain
git diff
git add path/to/files...
git diff --cached --stat
git commit -m "UI: <scoped change>"
git rev-parse HEAD
```

### 8) docmgr hygiene (after each step/commit)

```bash
# Check tasks when complete
docmgr task check --ticket 001-ADD-DOCMGR-UI --id <N>[,<M>...]

# Update changelog with why + key files
docmgr changelog update --ticket 001-ADD-DOCMGR-UI \
  --entry "..." \
  --file-note "/abs/path/to/file.tsx:Reason"

# Relate key files to the relevant design doc or diary
docmgr doc relate --doc ttmp/.../reference/01-diary.md \
  --file-note "/abs/path/to/file.tsx:Why it matters"
```

## Exit Criteria

- The new page/route renders without runtime errors.
- `pnpm -C ui lint` passes.
- `pnpm -C ui build` passes.
- Any new shared component lives in the right place and does not depend on page-local CSS.
- Ticket bookkeeping is updated:
  - tasks checked
  - changelog entry added
  - key files related
- Changes are committed in small, reviewable commits.

## Notes

- Avoid building “generic shared patterns” (e.g., list+preview) until at least 3 real pages use the pattern.
- If you must move files, do it in a separate “move-only” commit and keep behavioral changes separate.
