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


## 2026-01-05

SearchPage: split into widgets + introduce shared primitives (EmptyState/LoadingSpinner/ApiErrorAlert/RelatedFilesList); lint+build (commit 159be3b)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/ApiErrorAlert.tsx — Shared API error banner used by SearchPage
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/SearchPage.tsx — Now a thin orchestrator delegating UI to widgets (filters/results/preview/etc)
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/widgets/SearchDocsResults.tsx — Docs results list widget (DocCard rendering + load more)
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/widgets/SearchPreviewPanel.tsx — Desktop preview widget; uses shared RelatedFilesList


## 2026-01-05

Search preview: extract PathHeader (path + action buttons) and reuse in Search preview content (task 56)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/PathHeader.tsx — Shared primitive for path label + monospace path + action row
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/widgets/SearchPreviewContent.tsx — Uses PathHeader for Copy/Open actions


## 2026-01-05

TicketPage: split into tabs (overview/documents/tasks/graph/changelog) + extract TicketHeader/TicketTabs; reuse shared primitives (ApiErrorAlert/LoadingSpinner/PathHeader/RelatedFilesList)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/ticket/TicketPage.tsx — Now orchestrator (~200 LOC) delegating to tab widgets
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/ticket/tabs/TicketDocumentsTab.tsx — Documents tab list + preview; uses PathHeader and RelatedFilesList
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/ticket/tabs/TicketTasksTab.tsx — Tasks tab; keeps newTaskText local


## 2026-01-05

Retrofit DocViewer/FileViewer to use shared primitives (PageHeader, ApiErrorAlert, LoadingSpinner, copyToClipboard, RelatedFilesList); DocCard uses shared timeAgo

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/DocCard.tsx — Use shared timeAgo helper
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/PageHeader.tsx — New shared page header (title/subtitle/actions)
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/doc/DocViewerPage.tsx — Uses shared header/error/spinner/clipboard + RelatedFilesList
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/file/FileViewerPage.tsx — Uses shared header/error/spinner/clipboard


## 2026-01-05

UI: retrofit headers + date formatting (d80c73f); add MarkdownBlock primitive (866399f)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/MarkdownBlock.tsx — Shared markdown rendering
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/PageHeader.tsx — Standardize page headers
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/lib/time.ts — Shared formatDate helper


## 2026-01-05

UI: extract CodeBlock + DiagnosticCard primitives (a832d20)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/CodeBlock.tsx — Share FileViewer code rendering
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/components/DiagnosticCard.tsx — Share doc parse diagnostic rendering


## 2026-01-05

Add follow-up tasks for toast system, dm-* utilities, DocCard restyle, and state audit

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/05/007-MODULARIZE-UI-WIDGETS--modularize-web-ui-widgets-searchpage-extraction/tasks.md — Track next refactor phases

