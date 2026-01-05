# Tasks

## TODO

### Guardrails (keep refactor safe)
- [ ] Confirm current UI builds before changes (`pnpm -C ui build`)
- [ ] Confirm current UI lint passes before changes (`pnpm -C ui lint`)
- [ ] Keep diffs behavior-preserving (no UX changes; extraction-only)
- [ ] Keep each commit scoped to 1–2 extractions max

### High ROI extraction: Search page “leaf widgets” (low coupling)
Goal: shrink `ui/src/features/search/SearchPage.tsx` by moving pure subcomponents/helpers into dedicated files.

- [ ] Extract `useIsMobile` into `ui/src/features/search/hooks/useIsMobile.ts`
- [ ] Extract `MarkdownSnippet` + highlighting helpers into `ui/src/features/search/components/MarkdownSnippet.tsx`
- [ ] Extract `DiagnosticList` into `ui/src/features/search/components/DiagnosticList.tsx`
- [ ] Extract `TopicMultiSelect` into `ui/src/features/search/components/TopicMultiSelect.tsx`
- [ ] Update `SearchPage.tsx` to consume extracted modules (no behavior changes)

### High ROI extraction: shared utilities (duplication reducer)
Goal: eliminate repeated patterns across Search/Doc/File/Ticket.

- [ ] Introduce `ui/src/lib/time.ts` (`timeAgo`, `formatDate` as needed)
- [ ] Introduce `ui/src/lib/clipboard.ts` (`copyToClipboard(text)` wrapper + consistent errors)
- [ ] Introduce `ui/src/lib/apiError.ts` (parse error envelope; `apiErrorMessage(err)` helper)
- [ ] (Optional) Replace page-local duplicates in Search only first; expand to other pages in follow-up ticket

### High ROI extraction: Search page behavior hooks (highest impact, more risk)
Goal: make the route component a thin orchestrator by extracting behavior into hooks.

- [ ] Extract URL sync into `ui/src/features/search/hooks/useSearchUrlSync.ts`
  - [ ] Reads initial mode/query/filters from URL
  - [ ] Writes mode/query/filters to URL with debounce
  - [ ] Preserves current behavior for `sel` + `preview` params
- [ ] Extract keyboard shortcuts into `ui/src/features/search/hooks/useSearchHotkeys.ts`
  - [ ] `/` focus search input
  - [ ] `?` open shortcuts modal
  - [ ] Arrow navigation + Enter open + Esc clear/close
  - [ ] Alt+1/2/3 mode switching
  - [ ] Cmd/Ctrl+R refresh, Cmd/Ctrl+K copy selected path
- [ ] Extract selection model into `ui/src/features/search/hooks/useSearchSelection.ts`
  - [ ] Selected index/path; desktop vs mobile preview behavior preserved

### CSS cleanup (de-couple “design system” from Search-only layout)
- [ ] Split `ui/src/App.css` into shared utilities vs Search-only layout
- [ ] Keep classnames stable for now (minimize churn)

### Validation
- [ ] `pnpm -C ui lint`
- [ ] `pnpm -C ui build`
- [ ] Quick manual check: Search page still supports keyboard shortcuts + preview panel + URL restore

### Docmgr bookkeeping
- [ ] Relate touched files to `index.md` (`docmgr doc relate --ticket 007-MODULARIZE-UI-WIDGETS ...`)
- [ ] Update `changelog.md` with each extraction batch (`docmgr changelog update --ticket 007-MODULARIZE-UI-WIDGETS ...`)

## Done
