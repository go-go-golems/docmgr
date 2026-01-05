# Changelog

## 2026-01-03

- Initial workspace created


## 2026-01-03

Created diary and exhaustive guide to doc search implementation, CLI surface, and extension points

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/pkg/commands/search.go — Implementation reference for guide
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/01-diary.md — Research and validation diary for search guide
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md — Full documentation of search implementation + usage


## 2026-01-04

Step 6: Implement MVP Search Web UI + embed serving (commit 04a4d52)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/internal/web/spa.go — Serve SPA fallback
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/SearchPage.tsx — UI MVP


## 2026-01-04

Implement URL sync + structured diagnostics + inline API errors in Web UI

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/SearchPage.tsx — URL sync


## 2026-01-04

Add analysis for doc serving API + document viewer UI (markdown + code highlighting)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/01-doc-serving-api-and-document-viewer-ui.md — Proposed endpoints and UI plan


## 2026-01-04

Add doc/file serving endpoints and doc/file viewer routes with frontend markdown rendering (commit bacf9f9)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/internal/httpapi/docs_files.go — Doc/file serving endpoints
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/doc/DocViewerPage.tsx — Doc viewer route


## 2026-01-04

Finish UI MVP (shortcuts + mobile preview/filters), update API/UI docs, and set release builds to sqlite_fts5+embed

### Related Files

- .goreleaser.yaml — Release tags + UI generate
- Makefile — Build/install tags
- pkg/doc/docmgr-http-api.md — Document/file endpoints
- ui/src/features/search/SearchPage.tsx — Shortcuts + mobile preview/filter drawer

