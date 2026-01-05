# Tasks

## TODO

### Guardrails (keep refactor safe)
- [x] Confirm current UI builds before changes (`pnpm -C ui build`)
- [x] Confirm current UI lint passes before changes (`pnpm -C ui lint`)
- [x] Keep diffs behavior-preserving (no UX changes; extraction-only)
- [x] Keep each commit scoped to 1–2 extractions max

### High ROI extraction: Search page “leaf widgets” (low coupling)
Goal: shrink `ui/src/features/search/SearchPage.tsx` by moving pure subcomponents/helpers into dedicated files.

- [x] Extract `useIsMobile` into `ui/src/features/search/hooks/useIsMobile.ts`
- [x] Extract `MarkdownSnippet` + highlighting helpers into `ui/src/features/search/components/MarkdownSnippet.tsx`
- [x] Extract `DiagnosticList` into `ui/src/features/search/components/DiagnosticList.tsx`
- [x] Extract `TopicMultiSelect` into `ui/src/features/search/components/TopicMultiSelect.tsx`
- [x] Update `SearchPage.tsx` to consume extracted modules (no behavior changes)

### High ROI extraction: shared utilities (duplication reducer)
Goal: eliminate repeated patterns across Search/Doc/File/Ticket.

- [x] Introduce `ui/src/lib/time.ts` (`timeAgo`, `formatDate` as needed)
- [x] Introduce `ui/src/lib/clipboard.ts` (`copyToClipboard(text)` wrapper + consistent errors)
- [x] Introduce `ui/src/lib/apiError.ts` (parse error envelope; `apiErrorMessage(err)` helper)
- [ ] (Optional) Replace page-local duplicates in Search only first; expand to other pages in follow-up ticket

### High ROI extraction: Search page behavior hooks (highest impact, more risk)
Goal: make the route component a thin orchestrator by extracting behavior into hooks.

- [x] Extract URL sync into `ui/src/features/search/hooks/useSearchUrlSync.ts`
- [x] Reads initial mode/query/filters from URL
- [x] Writes mode/query/filters to URL with debounce
- [x] Preserves current behavior for `sel` + `preview` params
- [ ] Extract keyboard shortcuts into `ui/src/features/search/hooks/useSearchHotkeys.ts`
  - [ ] `/` focus search input
  - [ ] `?` open shortcuts modal
  - [ ] Arrow navigation + Enter open + Esc clear/close
  - [ ] Alt+1/2/3 mode switching
  - [ ] Cmd/Ctrl+R refresh, Cmd/Ctrl+K copy selected path
- [x] Extract selection model into `ui/src/features/search/hooks/useSearchSelection.ts`
- [x] Selected index/path; desktop vs mobile preview behavior preserved

### CSS cleanup (de-couple “design system” from Search-only layout)
- [x] Split `ui/src/App.css` into shared utilities vs Search-only layout
- [x] Keep classnames stable for now (minimize churn)

### Validation
- [x] `pnpm -C ui lint`
- [x] `pnpm -C ui build`
- [ ] Quick manual check: Search page still supports keyboard shortcuts + preview panel + URL restore

### Docmgr bookkeeping
- [x] Relate touched files to `index.md` (`docmgr doc relate --ticket 007-MODULARIZE-UI-WIDGETS ...`)
- [x] Update `changelog.md` with each extraction batch (`docmgr changelog update --ticket 007-MODULARIZE-UI-WIDGETS ...`)

## Done
- [x] Redux cleanup: make RTK Query own Search docs results (no local copies)
- [x] Implement searchDocs pagination merge (serializeQueryArgs ignores cursor; merge appends; forceRefetch on cursor change)
- [x] Refactor SearchPage: remove docsResults/docsTotal/docsNextCursor/docsDiagnostics local state; render from RTK Query data
- [x] Refactor SearchPage: derive hasSearched from lazy query state; ensure Clear resets queries + selection + auto-search latch
- [ ] Manual UX check: docs search, load more, clear, URL-restore auto-search, selection/preview, hotkeys

## TODO (Widget componentization: primary)

Notes:
- Keyboard shortcuts (tasks 18–23) are intentionally secondary; prioritize extracting UI widgets and shared primitives first.
- Keep behavior stable and commit in small batches (1–2 extractions per commit) to keep reviews manageable.

### Shared primitives (to support Workspace pages + shrink page files)
- [ ] Add `PageHeader` primitive (title/subtitle + right-side actions)
- [ ] Add `LoadingSpinner` primitive (Bootstrap wrapper used across pages)
- [ ] Add `EmptyState` primitive (title + body + optional actions)
- [ ] Add `ApiErrorAlert` primitive (render `apiErrorFromUnknown` output + details disclosure)
- [ ] Add `RelatedFilesList` widget (copy/open actions; used by DocViewer + Search preview + Ticket)
- [ ] Replace page-local `toErrorBanner` duplication with `ApiErrorAlert` where applicable (start with Search/Ticket)

### Search page: split into widgets (hotkeys remain in-page for now)
Goal: make `ui/src/features/search/SearchPage.tsx` a thin orchestrator (~200–350 LOC) by extracting UI composition into dedicated widgets.

- [ ] Introduce `ui/src/features/search/widgets/` directory (home for Search-only widgets)
- [ ] Extract `SearchHeader` (title + refresh + workspace status)
- [ ] Extract `SearchBar` (input + placeholder + keyboard hint + submit button)
- [ ] Extract `SearchModeToggle` (Docs/Reverse/Files buttons)
- [ ] Extract `SearchActiveChips` (computed filter chips row)
- [ ] Extract `SearchFiltersDesktop` (non-mobile filter panel)
- [ ] Extract `SearchFiltersDrawer` (mobile modal; share inner form fields with desktop)
- [ ] Extract `SearchDiagnosticsPanel` + diagnostics toggle button
- [ ] Extract `SearchFilesResults` (render files results list + empty state)
- [ ] Extract `SearchDocsResults` (render docs list, “Load more”, and empty state)
- [ ] Extract `SearchPreviewPanel` (desktop right-side preview)
- [ ] Extract `SearchPreviewModal` (mobile preview modal)
- [ ] Follow-up: move duplicated “path header + copy/open actions” into shared primitives (`PageHeader`, `RelatedFilesList`) and simplify Search widgets

### Ticket page: split into widgets (tab bodies become files)
Goal: split `ui/src/features/ticket/TicketPage.tsx` into tab widgets so it becomes a router+data orchestrator and establishes the “dashboard of cards” pattern used by Workspace pages.

- [ ] Introduce `ui/src/features/ticket/tabs/` directory (one file per tab body)
- [ ] Extract `TicketHeader` (title/subtitle + stats + actions)
- [ ] Extract `TicketTabs` (tab selection UI)
- [ ] Extract `TicketOverviewTab` (cards + key docs + open tasks + index doc)
- [ ] Extract `TicketDocumentsTab` (docs list/grouping + preview panel)
- [ ] Extract `TicketTasksTab` (sections list + add task form; keep `newTaskText` local)
- [ ] Extract `TicketGraphTab` (Mermaid graph + debug details)
- [ ] Extract `TicketChangelogTab` (changelog link + rendering as applicable)
- [ ] Identify shared “Doc list + preview” patterns between Search and Ticket and factor shared pieces into `ui/src/components/`

### Design system coherence (incremental, avoid churn)
- [ ] Ensure new primitives/widgets use `ui/src/styles/design-system.css` patterns first (avoid adding new ad-hoc CSS)
- [ ] Add/standardize a minimal `.dm-*` utility set as needed (mono paths, compact cards, section spacing)
- [ ] Keep Search-only layout rules in `ui/src/styles/search.css` (avoid leaking page layout into design system)

### Validation + docmgr bookkeeping (per extraction batch)
- [ ] After each extraction batch: `pnpm -C ui lint` + `pnpm -C ui build`
- [ ] After each batch: `docmgr changelog update --ticket 007-MODULARIZE-UI-WIDGETS --entry \"...\" --file-note ...`
- [ ] After each batch: `docmgr doc relate --ticket 007-MODULARIZE-UI-WIDGETS --file-note \"/abs/path:reason\"`
