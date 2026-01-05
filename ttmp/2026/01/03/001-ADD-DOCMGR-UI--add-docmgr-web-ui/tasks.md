# Tasks

## TODO

- [x] Write UI implementation plan from `design/01-design-docmgr-search-web-ui.md`
- [x] Extend `/api/v1/search/docs` results to include `lastUpdated` + `relatedFiles[]` (no file serving)
- [x] Create `ui/` React+Vite app skeleton (routing, layout, CSS baseline)
- [x] Add RTK store + RTK Query `docmgrApi` slice (health/status/refresh/search endpoints)
- [x] Implement header: title + refresh button + last refresh “time ago” + toast on success
- [x] Implement search bar: input + Search button + keyboard hint line (`/`, Enter, Esc)
- [x] Implement mode toggle (Docs / Reverse / Files) + per-mode placeholders/hints
- [x] Implement filter bar (ticket/topics/type/status/orderBy/file/dir + clear) + collapsible UI
- [x] Implement Topics multi-select with no suggestions (token list + “Add topic” input)
- [x] Implement quick toggles (includeArchived/includeScripts/includeControlDocs)
- [x] Implement active filter chips (remove chip → update filters → rerun search if active)
- [x] Implement docs results list:
- [x] Result card UI (status badge, topic badges, snippet, monospace path, hover copy button)
- [x] Loading spinner + pre-search empty state + post-search empty state
- [x] Cursor pagination (“Load more results” → append + nextCursor)
- [x] Implement reverse lookup mode (sets `reverse=true` and emphasizes matched file + note)
- [x] Implement diagnostics badge + expandable diagnostics panel (server-provided items)
- [x] Implement preview panel:
- [x] Desktop split pane (select result → show metadata + snippet + related files)
- [x] Mobile modal/page variant (“View →”)
- [x] Copy path (no file serving / no in-app full doc rendering)
- [x] Implement keyboard shortcuts overlay modal (`?`) + core shortcuts:
- [x] `/` focus input, ↑/↓ navigation, Enter select/open, Esc close/clear
- [x] Alt+1/2/3 mode switching
- [x] Cmd/Ctrl+R refresh index
- [x] Cmd/Ctrl+K copy selected path
- [x] Implement files suggestions mode (`/api/v1/search/files`) + file cards
- [x] Implement URL sync (`mode`, `q`, filters) and restore state on reload
- [x] Add responsive styling (compact layout, filter drawer, preview modal)
- [x] Add doc viewer + file viewer (doc serving MVP):
- [x] Backend: implement `GET /api/v1/docs/get?path=...` (doc metadata + body)
- [x] Backend: implement `GET /api/v1/files/get?path=...&root=repo|docs` (safe text-only)
- [x] Backend: add tests (traversal rejection, symlink escape, binary rejection, size limit/truncation)
- [x] Frontend: add RTK Query endpoints `getDoc` + `getFile`
- [x] Frontend: add routes `/doc?path=...` and `/file?path=...&root=...`
- [x] Frontend: render markdown in doc viewer (client-side) using `react-markdown` + `remark-gfm`
- [x] Frontend: add code highlighting for fenced blocks using `rehype-highlight` + a highlight.js theme
- [x] Frontend: wire “Open doc” from search results and “Open related file” from RelatedFiles
- [x] Embed packaging (per `go-web-frontend-embed`):
- [x] Add `internal/web` embed + SPA fallback handler
- [x] Add `go generate` bridge to build/copy `ui/dist/public` into `internal/web/embed/public`
- [x] Wire SPA handler into `docmgr api serve` without shadowing `/api`
- [x] Add Makefile targets for `dev-frontend`, `dev-backend`, and embed build
- [x] Add minimal regression test: `GET /` serves `index.html` when embed assets exist
- [x] Update `pkg/doc` docs for “Web UI” (how to run dev, how to run embedded)

## Done

- [x] Trace doc search implementation from CLI → Workspace.QueryDocs
- [x] Write exhaustive search implementation + API/CLI guide
- [x] Validate search behavior with go run examples
- [x] Run docmgr doctor for the ticket
- [x] Upload diary + guide PDFs to reMarkable
- [x] Render diagnostics as a structured list (severity/message/suggestion) instead of raw JSON
- [x] Show API error details inline (code/message/details) for search failures
