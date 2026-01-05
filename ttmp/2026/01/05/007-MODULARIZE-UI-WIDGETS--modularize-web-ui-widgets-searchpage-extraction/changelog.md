# Changelog

## 2026-01-05

- Initial workspace created


## 2026-01-05

Extract leaf SearchPage widgets/hooks (MarkdownSnippet, DiagnosticList, TopicMultiSelect, useIsMobile) to reduce page size and enable further modularization; verified lint+build.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/SearchPage.tsx — Now imports extracted widgets
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/components/MarkdownSnippet.tsx — New


## 2026-01-05

Extract shared helpers into ui/src/lib (timeAgo + copyToClipboard) and switch SearchPage to import them; lint+build verified.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/lib/clipboard.ts — New
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/lib/time.ts — New


## 2026-01-05

Extract api error envelope parsing into ui/src/lib/apiError.ts and use it in SearchPage to build error banners.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/lib/apiError.ts — New


## 2026-01-05

Extract SearchPage URL state sync into useSearchUrlSync (restores from URL and debounced writes; preserves sel+preview behavior).

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/hooks/useSearchUrlSync.ts — New


## 2026-01-05

Extract Search selection model into a hook and split App.css into design-system vs Search-specific styles; lint+build verified.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/hooks/useSearchSelection.ts — New
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/styles/design-system.css — New
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/styles/search.css — New


## 2026-01-05

Design: define Redux/RTK Query state ownership policy for widgets/pages (what moves to slices vs stays local), with a SearchPage migration plan.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/design-doc/01-design-redux-state-strategy-for-ui-widgets.md — New design doc


## 2026-01-05

Search: move docs results out of local state into RTK Query (pagination merge + clear/reset) (commit 0cd16e6)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/SearchPage.tsx — stop copying search responses into useState; render from RTK Query state and reset on Clear
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/services/docmgrApi.ts — searchDocs now merges paginated pages into one cache entry (cursor excluded from cache key)

