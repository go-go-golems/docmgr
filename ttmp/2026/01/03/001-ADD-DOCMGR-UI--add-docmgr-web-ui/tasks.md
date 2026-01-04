# Tasks

## TODO

- [ ] Write UI implementation plan from `design/01-design-docmgr-search-web-ui.md`
- [ ] Decide v1 scope for “Updated” + “Related Files” + “Open Full Doc” (UI-only vs API extensions)
- [ ] Create `ui/` React+Vite app skeleton (routing, layout, CSS baseline)
- [ ] Add RTK store + RTK Query `docmgrApi` slice (health/status/refresh/search endpoints)
- [ ] Implement header: title + refresh button + last refresh “time ago” + toast on success
- [ ] Implement search bar: input + Search button + keyboard hint line (`/`, Enter, Esc)
- [ ] Implement mode toggle (Docs / Reverse / Files) + per-mode placeholders/hints
- [ ] Implement filter bar (ticket/topics/type/status/orderBy/file/dir + clear) + collapsible UI
- [ ] Implement quick toggles (includeArchived/includeScripts/includeControlDocs)
- [ ] Implement active filter chips (remove chip → update filters → rerun search if active)
- [ ] Implement docs results list:
  - [ ] Result card UI (status badge, topic badges, snippet, monospace path, hover copy button)
  - [ ] Loading spinner + pre-search empty state + post-search empty state
  - [ ] Cursor pagination (“Load more results” → append + nextCursor)
- [ ] Implement reverse lookup mode (sets `reverse=true` and emphasizes matched file + note)
- [ ] Implement diagnostics badge + expandable diagnostics panel (server-provided items)
- [ ] Implement preview panel:
  - [ ] Desktop split pane (select result → show metadata + snippet + related files)
  - [ ] Mobile modal/page variant (“View →”)
  - [ ] Copy path + “Open full doc” placeholder action
- [ ] Implement keyboard shortcuts overlay modal (`?`) + core shortcuts:
  - [ ] `/` focus input, ↑/↓ navigation, Enter select/open, Esc close/clear
  - [ ] Alt+1/2/3 mode switching
  - [ ] Cmd/Ctrl+R refresh index
  - [ ] Cmd/Ctrl+K copy selected path
- [ ] Implement files suggestions mode (`/api/v1/search/files`) + file cards
- [ ] Implement URL sync (`mode`, `q`, filters) and restore state on reload
- [ ] Add responsive styling (compact layout, filter drawer, preview modal)
- [ ] Embed packaging (per `go-web-frontend-embed`):
  - [ ] Add `internal/web` embed + SPA fallback handler
  - [ ] Add `go generate` bridge to build/copy `ui/dist/public` into `internal/web/embed/public`
  - [ ] Wire SPA handler into `docmgr api serve` without shadowing `/api`
  - [ ] Add Makefile targets for `dev-frontend`, `dev-backend`, and embed build
- [ ] Add minimal regression test: `GET /` serves `index.html` when embed assets exist
- [ ] Update `pkg/doc` docs for “Web UI” (how to run dev, how to run embedded)

## Done

- [x] Trace doc search implementation from CLI → Workspace.QueryDocs
- [x] Write exhaustive search implementation + API/CLI guide
- [x] Validate search behavior with go run examples
- [x] Run docmgr doctor for the ticket
- [x] Upload diary + guide PDFs to reMarkable
