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


## 2026-01-04

UI: persist selected doc in URL, allow ctrl-click Open links, and render snippet markdown with match highlighting

### Related Files

- pkg/doc/docmgr-web-ui.md — Document sel/preview params
- ui/src/features/search/SearchPage.tsx — sel/preview URL sync


## 2026-01-04

Design ticket page HTTP API + Web UI (overview/docs/tasks/graph/changelog)

### Related Files

- ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design/02-ticket-page-api-and-web-ui.md — Ticket page design spec


## 2026-01-05

Docs: audit current React SPA (Search/Doc/File/Ticket) and propose a widget + design-system architecture for upcoming Workspace navigation pages.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md — New analysis doc capturing architecture + widget plan


## 2026-01-05

Fix: Ticket UI tolerates malformed tasks response where sections[].items is null (prevents runtime crash in openTasks and Tasks tab).

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/ticket/TicketPage.tsx — Defensive asArray() wrapper for section items


## 2026-01-05

Diary: backfilled SearchPage modularization work (ticket 007) and the TicketPage tasks null crash fix steps.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/01-diary.md — Added Steps 24–33


## 2026-01-05

Design + scaffold Workspace navigation UI (AppShell + /workspace pages) (commit b1900d1)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/App.tsx — Add /workspace nested routes
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/search/widgets/SearchHeader.tsx — Add Workspace entry link
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ui/src/features/workspace/WorkspaceLayout.tsx — Workspace shell with TopBar + SideNav


## 2026-01-05

Add playbook for implementing new UI pages/widgets (post-refactor conventions)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/playbook/01-playbook-implementing-new-ui-pages-and-widgets-post-refactor.md — Central workflow and conventions


## 2026-01-05

Design doc: embed Workspace ASCII screenshots verbatim (from sources/workspace-page.md)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design-doc/01-design-workspace-navigation-ui-post-refactor.md — Appendix includes ASCII designs for quick reference

