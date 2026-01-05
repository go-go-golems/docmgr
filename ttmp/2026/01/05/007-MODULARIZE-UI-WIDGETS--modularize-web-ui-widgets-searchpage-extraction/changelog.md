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

