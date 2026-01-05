---
Title: Diary
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - tooling
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/httpapi/tickets.go
      Note: Ticket API endpoints (/tickets/get/docs/tasks/graph) (commits 522e678,4a82f9d)
    - Path: internal/httpapi/tickets_test.go
      Note: Basic handler test coverage for ticket endpoints (commit 522e678)
    - Path: .goreleaser.yaml
      Note: Release tags sqlite_fts5
    - Path: Makefile
      Note: Default build/install tags sqlite_fts5
    - Path: cmd/docmgr/cmds/api/serve.go
      Note: Serve API + SPA from one process
    - Path: internal/httpapi/docs_files.go
      Note: Doc/file serving endpoints (commit bacf9f9)
    - Path: internal/httpapi/path_safety.go
      Note: Safe path resolution + symlink escape protection (commit bacf9f9)
    - Path: internal/httpapi/server.go
      Note: Allow empty browse; reverse query->file; orderBy guards
    - Path: internal/tasksmd/tasksmd.go
      Note: Parse/mutate tasks.md for ticket tasks API (commit 522e678)
    - Path: internal/ticketgraph/ticketgraph.go
      Note: Mermaid builder for ticket graph API (commit 522e678)
    - Path: internal/tickets/resolve.go
      Note: Resolve ticket dir/index.md from workspace index (commit 522e678)
    - Path: internal/searchsvc/search.go
      Note: Add lastUpdated+relatedFiles to search results for UI
    - Path: internal/web/generate_build.go
      Note: go generate bridge to build/copy Vite assets
    - Path: internal/web/spa.go
      Note: SPA fallback handler (never shadow /api)
    - Path: pkg/doc/docmgr-http-api.md
      Note: Document/file endpoints docs (Step 10)
    - Path: pkg/doc/docmgr-web-ui.md
      Note: |-
        User docs for running the UI (dev + embedded)
        Viewer routes + shortcuts docs (Step 10)
        URL params docs for selection restore (Step 11)
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/01-doc-serving-api-and-document-viewer-ui.md
      Note: Doc serving API + viewer research and plan
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/02-single-doc.md
      Note: UX snapshot (terminal-style doc view)
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/03-single-doc.html
      Note: Mock doc viewer UI spec
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/single-doc.html
      Note: Mock doc viewer UI spec
    - Path: ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/single-doc.md
      Note: UX snapshot (terminal-style doc view)
    - Path: ui/src/App.css
      Note: Responsive tweaks + selected result styling (Step 10)
    - Path: ui/src/features/doc/DocViewerPage.tsx
      Note: Doc viewer route + markdown rendering (commit bacf9f9)
    - Path: ui/src/features/file/FileViewerPage.tsx
      Note: File viewer route + syntax highlighting (commit bacf9f9)
    - Path: ui/src/features/search/SearchPage.tsx
      Note: |-
        MVP search UI (modes
        Wire Open doc/Open file navigation (commit bacf9f9)
        Mobile preview modal + filter drawer + keyboard shortcuts (Step 10)
        URL sel/preview state + Link opens + markdown snippets (Step 11)
        Ticket badge links to /ticket/:ticket (commit 522e678)
    - Path: ui/src/features/ticket/TicketPage.tsx
      Note: Ticket page tabs (overview/docs/tasks/graph/changelog) (commits 522e678,4a82f9d)
    - Path: ui/src/components/MermaidDiagram.tsx
      Note: Client-side Mermaid rendering for ticket graph (commit 4a82f9d)
    - Path: ui/src/services/docmgrApi.ts
      Note: |-
        RTK Query client for docmgr HTTP API
        RTK Query endpoints getDoc/getFile (commit bacf9f9)
        Ticket endpoints (/tickets/*) (commit 522e678)
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-05T00:20:58-05:00
WhatFor: ""
WhenToUse: ""
---







# Diary

## Goal

This diary captures the research and writing process for documenting `docmgr` search: how it’s implemented, what the CLI/API surface is, and how it can be extended. It’s meant to preserve the “how I found it” trail (commands, files, sharp edges), not just the final guide.

## Step 1: Create the ticket and make docmgr runnable locally

This step established the workspace for the work (ticket + diary) and ensured I can run the `docmgr` CLI from this repo to validate search behavior while reading the implementation. The main outcome is a reproducible invocation pattern for `docmgr` in this repo despite a `go.work`/toolchain mismatch.

I hit an immediate build/run issue when trying to run `docmgr` from the repo root: the `go.work` file’s `go` directive is lower than the per-module `go` requirements, and the workspace includes a module that requires a newer Go patch release than the Go tool available in this environment. I worked around this by running the `docmgr` module directly with `GOWORK=off`, which is sufficient for documenting and exercising `docmgr`’s search features.

### What I did
- Read workflow instructions in `~/.cursor/commands/docmgr.md` and `~/.cursor/commands/diary.md`.
- Verified `docmgr` can resolve this repo’s `.ttmp.yaml` configuration (docs root is `docmgr/ttmp`).
- Ran `docmgr` module commands with `GOWORK=off` to avoid `go.work` incompatibilities.
- Created ticket `001-ADD-DOCMGR-UI` and created this diary document in the ticket workspace.

### Why
- I need an actual ticket workspace under `docmgr/ttmp/` to store the exhaustive search guide and keep a detailed diary while researching.
- I need a reliable way to execute `docmgr doc search` so that the written guide matches actual behavior.

### What worked
- `GOWORK=off go run ./cmd/docmgr status --summary-only` runs successfully from `docmgr/` and discovers `.ttmp.yaml`.
- Ticket creation works and created the expected workspace layout under `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/`.

### What didn't work
- Running `go run` at the repo root failed with:
  - `module glazed listed in go.work file requires go >= 1.24.2, but go.work lists go 1.23`
  - `module pinocchio listed in go.work file requires go >= 1.25.4 ...`

### What I learned
- `go.work` can block running a single module even if that module would build fine on its own; `GOWORK=off` is a practical way to focus on one module for investigation and documentation.

### What was tricky to build
- The interaction between:
  - `go.work`’s `go` version directive,
  - per-module `go` requirements and `toolchain` directives,
  - and the workspace including modules with stricter requirements than the local Go toolchain.

### What warrants a second pair of eyes
- Whether the intended workflow for this mono-repo is to keep `go.work` usable for all modules, or to treat each module as independently runnable (and document `GOWORK=off` as the standard workaround).

### What should be done in the future
- N/A (for the search guide work). If this repo is meant to be a coherent `go.work` workspace, it likely needs a follow-up ticket to reconcile Go/toolchain requirements across modules.

### Code review instructions
- N/A (no code changes in this step).

### Technical details
- Environment: `go version go1.25.3 linux/amd64`
- Successful command (from `docmgr/`): `GOWORK=off go run ./cmd/docmgr status --summary-only`
- Ticket created via: `GOWORK=off go run ./cmd/docmgr ticket create-ticket --ticket 001-ADD-DOCMGR-UI --title "Add docmgr Web UI" --topics docmgr,ux,cli,tooling`

## Step 2: Trace doc search from CLI down to the index query

This step mapped the `docmgr doc search` execution path end-to-end: CLI wiring (cobra + glazed), the user-visible flags and modes, and the internal query engine (`Workspace.InitIndex` + `Workspace.QueryDocs`). The main deliverable from this step is a set of “entry points” and invariants I can now rely on while writing the exhaustive guide.

The most important structural finding is that “search” is a hybrid: *metadata + reverse lookup* (ticket/topics/doc-type/status/related_files) are handled by an in-memory SQLite index (`internal/workspace`), while *content search*, *external source filtering*, and some *date filtering* are applied as post-filters in the command layer (`pkg/commands/search.go`). That split matters for extension work: adding indexed filters requires schema/ingest changes, while post-filters can often be implemented purely in the command.

### What I did
- Read the CLI entrypoint and command tree wiring:
  - `docmgr/cmd/docmgr/main.go`
  - `docmgr/cmd/docmgr/cmds/root.go`
  - `docmgr/cmd/docmgr/cmds/doc/search.go`
- Read the command implementation:
  - `docmgr/pkg/commands/search.go`
  - `docmgr/pkg/commands/document_utils.go`
- Read the workspace/index/query implementation:
  - `docmgr/internal/workspace/workspace.go`
  - `docmgr/internal/workspace/index_builder.go`
  - `docmgr/internal/workspace/sqlite_schema.go`
  - `docmgr/internal/workspace/query_docs.go`
  - `docmgr/internal/workspace/query_docs_sql.go`
- Read path normalization and reverse-lookup matching utilities:
  - `docmgr/internal/paths/resolver.go`
  - `docmgr/internal/workspace/normalization.go`
- Read how `RelatedFiles` are canonicalized when users run `docmgr doc relate` (search depends on this data quality):
  - `docmgr/pkg/commands/relate.go`

### Why
- The guide needs to explain search behavior “from the inside out”: what is indexed, what is filtered after, and what normalization/matching rules apply for reverse lookups.

### What worked
- The architecture is relatively clean to explain: CLI → command → workspace discovery → index build → SQL query → post-filters → output.
- Reverse lookup has an explicit normalization strategy: persisted `related_files` rows store multiple normalized representations and are queried with a best-effort key set (plus basename-only suffix matching).

### What didn't work
- N/A (this step was pure reading). The only earlier blocker remains the `go.work` mismatch, handled via `GOWORK=off`.

### What I learned
- `docmgr search` is an alias for `docmgr doc search` (implemented in `docmgr/cmd/docmgr/cmds/root.go` by cloning the doc search cobra command and renaming `Use`).
- `docmgr doc search` has two execution modes:
  - “search documents” mode (default): builds the index with `IncludeBody=true`, runs `Workspace.QueryDocs`, then applies post-filters for content/external source/dates.
  - “suggest files” mode (`--files`): builds the index with `IncludeBody=false` and blends suggestions from `RelatedFiles` + git history + ripgrep + git status.
- The index is rebuilt per invocation into an in-memory SQLite database (see `Workspace.InitIndex` + `openInMemorySQLite`).

### What was tricky to build
- The reverse-lookup behavior is spread across layers:
  - SQL filtering uses persisted normalized keys (workspace index).
  - The command layer also re-normalizes per-document when printing “matched file” details for `--file`.
  - Path matching intentionally uses multiple fallbacks (canonical/repo-relative/docs-relative/doc-relative/abs/clean/raw + basename suffix patterns), which is great for UX but easy to misunderstand if not documented explicitly.

### What warrants a second pair of eyes
- Confirm the intended semantics of `--created-since`: it currently uses filesystem `ModTime()` as “created time” in `docmgr/pkg/commands/search.go`, which is not a true creation timestamp on most filesystems.
- Confirm whether `docmgr doc search` should include control docs (`README.md`, `tasks.md`, `changelog.md`) and `archive/` by default; the command currently sets `IncludeControlDocs=true` and `IncludeArchivedPath=true` when calling `QueryDocs`.

### What should be done in the future
- If/when search becomes performance-sensitive, consider moving content search to an indexed form (e.g., SQLite FTS) instead of a post-filter substring scan. This is a structural follow-up, not needed for the documentation deliverable.

### Code review instructions
- N/A (no code changes in this step).

### Technical details
- Main command implementation: `docmgr/pkg/commands/search.go`
- Index schema and ingest: `docmgr/internal/workspace/sqlite_schema.go`, `docmgr/internal/workspace/index_builder.go`
- Query compilation and reverse lookup SQL: `docmgr/internal/workspace/query_docs_sql.go`

## Step 3: Exercise `docmgr doc search` against this repo’s workspace

This step validated the mental model from Step 2 by running real `docmgr doc search` commands against the existing `docmgr/ttmp` workspace in this repo. The goal was to confirm the key behaviors I plan to document (content search, reverse lookup, directory lookup, output modes, and file suggestion heuristics) using concrete commands and outputs.

Because this repo’s `ttmp` already contains a lot of historical tickets and related file links, it’s a good real-world dataset for “messy path” behavior: some RelatedFiles entries are canonical repo-relative paths, and others are odd relative paths from past workspaces. The search implementation’s normalization+fallback strategy appears designed exactly for this kind of reality.

### What I did
- Ran reverse lookup by exact-ish file path:
  - `GOWORK=off go run ./cmd/docmgr doc search --file docmgr/pkg/commands/search.go --with-glaze-output --output table`
- Ran reverse lookup by basename (suffix matching):
  - `GOWORK=off go run ./cmd/docmgr doc search --file search.go --with-glaze-output --output table`
- Ran directory reverse lookup:
  - `GOWORK=off go run ./cmd/docmgr doc search --dir pkg/commands --with-glaze-output --output table`
- Ran content search to confirm snippet extraction and indexing includes body:
  - `GOWORK=off go run ./cmd/docmgr doc search --query "Workspace.QueryDocs" --with-glaze-output --output table`
- Ran file suggestion mode (heuristics blend):
  - `GOWORK=off go run ./cmd/docmgr doc search --ticket 001-ADD-DOCMGR-UI --files`

### Why
- The written guide needs to be grounded in observable behavior and command outputs, not just code reading.

### What worked
- `--file` finds documents that reference a file even when the stored path representation differs (repo-relative vs odd relative), consistent with the index’s multi-key matching strategy.
- Basename-only queries (e.g., `--file search.go`) return results (the SQL layer adds suffix `LIKE` patterns for basename-only queries).
- `--dir` returns results for `pkg/commands` (directory prefix matching across the same set of normalized keys).
- `--with-glaze-output --output table` produces a stable, machine-friendly table format, including extra columns (`file`, `file_note`) when `--file` is used.

### What didn't work
- N/A (no unexpected failures in this round).

### What I learned
- In this repo’s dataset, RelatedFiles entries include both clean canonical paths like `pkg/commands/search.go` and very long relative paths (e.g., `../../../../.../docmgr/pkg/commands/search.go`). The combination of:
  - storing multiple normalized forms in `related_files`,
  - generating query key sets via `queryPathKeys(...)`,
  - and basename-only suffix patterns
  is enough to match across those variants.

### What was tricky to build
- Interpreting `--files` output: it can be noisy because it always includes “recent commit activity” results from git history even when there’s no query/topic term to focus the heuristic. This is still useful, but the guide should explain that the output is an unranked multi-source suggestion stream.

### What warrants a second pair of eyes
- Confirm whether `--files` should require at least one of `--query` or `--topics` to avoid “just show me recent commits” behavior for new/empty tickets.

### What should be done in the future
- N/A for the documentation deliverable.

### Code review instructions
- N/A (no code changes in this step).

### Technical details
- Example table output command: `GOWORK=off go run ./cmd/docmgr doc search --file docmgr/pkg/commands/search.go --with-glaze-output --output table`

## Step 4: Write the exhaustive search guide and validate ticket docs

This step produced the main deliverable: a verbose, implementation-grounded guide to `docmgr` search covering CLI usage, internal architecture, indexing, reverse lookup semantics, output modes, templating, and concrete extension playbooks. I wrote it iteratively using the findings from earlier steps (code reading + live command validation) so the document reflects how the tool actually behaves in practice.

I also ran `docmgr doctor` for the new ticket to ensure the created docs are structurally valid and updated the ticket changelog to record the documentation deliverable and its key reference files.

### What I did
- Wrote `reference/02-doc-search-implementation-and-api-guide.md` for ticket `001-ADD-DOCMGR-UI`.
- Related the core implementation files to the guide document (so reverse lookup from code → guide works).
- Related the same core implementation files to the ticket `index.md` (so the ticket overview stays a good entry point).
- Ran doctor on the ticket:
  - `GOWORK=off go run ./cmd/docmgr doctor --ticket 001-ADD-DOCMGR-UI --stale-after 30`
- Updated the ticket changelog with file notes pointing at the diary, the guide, and `pkg/commands/search.go`.

### Why
- The user request explicitly asks for a detailed, prose-heavy document that goes from implementation → usage, and for a “frequent” diary trail while doing research and writing.
- Running `doctor` is a quick sanity check that the output is consistent with docmgr’s own expectations (frontmatter + layout + relationships).

### What worked
- The guide can be navigated both “top-down” (user workflows) and “bottom-up” (index schema → QueryDocs → post-filters).
- Relating the relevant code files to the guide makes `docmgr doc search --file ...` a practical navigation tool for future readers.
- `doctor` reports all checks passed for this ticket.

### What didn't work
- N/A (no failures encountered in writing/validation).

### What I learned
- The glazed long help is the most reliable source for the “structured output toolchain” flags; the short help only lists the docmgr-specific flags.
- `--select` is a glazed “single field per line” facility that works as expected when paired with template output (e.g., `--output template --select path`), and should be documented accordingly.

### What was tricky to build
- Keeping the guide accurate about the CLI’s dual-mode + glazed flag behavior required checking the live `--help --long-help` output rather than relying on assumptions or older examples.

### What warrants a second pair of eyes
- Review the guide’s claims about which parts of search are index-backed vs post-filtered to ensure they stay aligned as docmgr evolves (especially if content search moves into SQLite FTS in the future).

### What should be done in the future
- N/A for this documentation deliverable.

### Code review instructions
- Start in `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`.
- Validate via:
  - `cd docmgr`
  - `GOWORK=off go run ./cmd/docmgr doc search --file docmgr/pkg/commands/search.go --with-glaze-output --output table`
  - `GOWORK=off go run ./cmd/docmgr doctor --ticket 001-ADD-DOCMGR-UI --stale-after 30`

### Technical details
- Guide doc: `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`
- Ticket changelog: `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/changelog.md`

## Step 5: Upload diary + search guide to reMarkable

This step published the two primary docs (the diary and the search guide) to my reMarkable so they can be reviewed and annotated away from the workstation. The key outcome is that both markdown files were converted to PDFs (with YAML frontmatter stripped) and uploaded into a mirrored ticket folder under `ai/YYYY/MM/DD/` on the device.

### What I did
- Read `~/.cursor/commands/remarkable.md` and followed the 3-step workflow (dry-run, then upload).
- Uploaded:
  - `reference/01-diary.md` → `01-diary.pdf`
  - `reference/02-doc-search-implementation-and-api-guide.md` → `02-doc-search-implementation-and-api-guide.pdf`

### Why
- Makes it easier to do long-form reading and markup on the device without losing the frontmatter/metadata noise in the PDF.

### What worked
- Dry-run correctly showed the remote destination:
  - `ai/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/`
- Actual upload succeeded without needing `--force` (no collisions).

### What didn't work
- N/A.

### What I learned
- `--mirror-ticket-structure` is the safest default for docmgr tickets because it avoids name collisions under `ai/YYYY/MM/DD/` and keeps the device folder structure aligned with ticket layout.

### What was tricky to build
- N/A (tooling already exists; this was a straightforward publish step).

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- N/A.

### Code review instructions
- N/A.

### Technical details
- Commands:
  - Dry run:
    - `python3 /home/manuel/.local/bin/remarkable_upload.py --ticket-dir /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui --mirror-ticket-structure --dry-run /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/01-diary.md /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`
  - Upload:
    - `python3 /home/manuel/.local/bin/remarkable_upload.py --ticket-dir /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui --mirror-ticket-structure /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/01-diary.md /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/02-doc-search-implementation-and-api-guide.md`
- Remote destination:
  - `ai/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/`

## Step 6: Implement a first usable Search Web UI + embedded serving

This step turns the earlier “search implementation research” into something you can actually use: a React/Vite Search UI that calls the local `docmgr` HTTP API, plus the Go plumbing to serve the SPA from `docmgr api serve` for a single-binary “just run it” mode.

The key behavior goals for the MVP were: (1) browse all docs with empty query, (2) docs search ordered by rank when a text query exists, (3) reverse lookup mode that searches by related file paths, and (4) cursor “load more” pagination. I also wired a minimal preview panel and basic diagnostics visibility to make the UI practically useful during development.

**Commit (code):** 04a4d52 — "web-ui: add React SPA, HTTP search tweaks, and embedded serving"

### What I did
- Implemented API tweaks so a UI can work naturally:
  - allow empty query to “browse all docs”
  - prevent `orderBy=rank` from erroring when `query` is empty (fallback to `path`)
  - when `reverse=true` and `file/dir` are empty, treat `query` as `file` for reverse lookup convenience
- Added UI-field backfill to the search response:
  - `lastUpdated`
  - `relatedFiles[]` (full list)
  - always-present `matchedFiles[]/matchedNotes[]` arrays
- Scaffolded `ui/` (React+Vite+TS), styled with Bootstrap, and implemented the `SearchPage`:
  - Docs / Reverse / Files modes
  - filter bar + toggles + topics multi-select (no suggestions)
  - results list with “Load more” (cursor pagination)
  - copy path + simple toast
  - minimal preview panel (desktop split view)
- Implemented Go SPA serving + embed pipeline:
  - `internal/web` provides SPA fallback handler + `go generate` build/copy bridge
  - `docmgr api serve` serves `/api/*` and, if assets exist, serves `/` + SPA fallback
- Added docs:
  - `pkg/doc/docmgr-web-ui.md` (“how to run dev and embedded mode”)
  - updated `pkg/doc/docmgr-http-api.md` for new response fields + reverse convenience semantics

### Why
- A web UI should not need to re-implement docmgr search semantics; the server/API should be the single source of truth.
- The MVP UI needs a “browse” workflow (empty query) because it’s a common way to discover docs by ticket/type/status/topics without knowing exact terms.
- Serving the SPA from Go makes it easy to ship one binary and avoids needing a separate “frontend deployment” concept for a local-first tool.

### What worked
- The UI can be run in a two-process loop (Vite `:3000` with `/api` proxy → backend `:3001`).
- The same UI can be served directly from `docmgr api serve` (embedded/disk-backed assets) at `http://127.0.0.1:3001/`.
- Cursor pagination works via the API’s `nextCursor` contract.

### What didn't work
- Initial Vite scaffolding created a non-React template due to a CLI invocation mistake; regenerated using `pnpm create vite@latest ui -t react-ts`.
- pnpm initially refused to run `esbuild` postinstall scripts; resolved by allowing `esbuild` in `ui/package.json` via `pnpm.onlyBuiltDependencies`.

### What I learned
- Reverse lookup is easiest to use when the “main search box” in Reverse mode edits the file path filter (not the text query).
- `OrderByRank` is a “text-query-only” concept at the SQL layer; it should be treated as a UI convenience, not a hard requirement.

### What was tricky to build
- Getting “serve SPA + serve API” routing correct so `/api/*` is never shadowed by the SPA fallback.
- Balancing “API correctness” with “UI ergonomics”:
  - empty query should be allowed for browsing,
  - rank ordering shouldn’t explode on empty query,
  - reverse lookup should be strict but ergonomic (query→file mapping).

### What warrants a second pair of eyes
- The API semantics changes:
  - confirm allowing empty query is acceptable and doesn’t cause surprising performance regressions
  - confirm the `reverse=true` query→file mapping is the right contract (and doesn’t mask accidental client bugs)
- The Go SPA handler:
  - confirm fallback logic never intercepts `/api/*`
  - confirm embed/disk modes behave as intended (especially around missing assets)

### What should be done in the future
- Implement URL sync (`mode`, `q`, filters) so searches are shareable.
- Replace the diagnostics JSON blob with a structured diagnostic list.
- Implement the keyboard shortcuts modal and richer navigation (arrow selection, etc).
- Add mobile preview modal/drawer behavior (right now it’s a desktop-first split panel).

### Code review instructions
- Start in:
  - `ui/src/features/search/SearchPage.tsx`
  - `internal/httpapi/server.go`
  - `internal/searchsvc/search.go`
  - `internal/web/spa.go`
  - `cmd/docmgr/cmds/api/serve.go`
- Validate:
  - `make dev-backend` and `make dev-frontend` → open `http://localhost:3000/`
  - `curl http://127.0.0.1:3001/api/v1/search/docs?query=&orderBy=rank&pageSize=10`
  - `curl http://127.0.0.1:3001/api/v1/search/docs?reverse=true&file=pkg/commands/skill_list.go&pageSize=10`

### Technical details
- Dev topology:
  - Vite: `http://localhost:3000/` (proxy `/api` → `http://127.0.0.1:3001`)
  - Backend: `go run -tags sqlite_fts5 ./cmd/docmgr api serve --addr 127.0.0.1:3001 --root ttmp`
- Embedded build:
  - `go generate ./internal/web`
  - `go build -tags \"sqlite_fts5,embed\" ./cmd/docmgr`

### What I'd do differently next time
- Start by locking down “reverse lookup UX” and “empty query browse semantics” in the API contract before building the UI, because those two behaviors shape a surprising amount of UI control flow.

## Step 7: Analyze doc serving API + doc viewer UI (markdown + code highlighting)

This step scoped the next big UX leap: turning a search hit into a fully readable document view (title/metadata/related files + rendered markdown) and making `RelatedFiles` actionable by serving and syntax-highlighting source files.

I used the provided mockups (`single-doc.html` / `single-doc.md`) as the “pixel-level” spec, then worked backwards to identify the minimal backend endpoints and frontend libraries we’d need to ship an MVP safely (path traversal protection, binary detection, size limits).

### What I did
- Read the target mock UI and UX snapshot:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/03-single-doc.html`
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/02-single-doc.md`
- Reviewed the existing backend surfaces we can extend:
  - `internal/httpapi/server.go` (route registration + error format)
  - `internal/documents` parsing helpers (frontmatter + body)
  - `internal/workspace.Context` roots (`root`, `repoRoot`) from `GET /api/v1/workspace/status`
- Wrote an analysis document with:
  - endpoint proposals (`/api/v1/docs/get`, `/api/v1/files/get`, optional assets)
  - security constraints (root restrictions, traversal/symlink, binary/size checks)
  - UI route design (`/doc?path=...`, `/file?path=...`)
  - markdown rendering options + recommended libs

### Why
- The search UI is useful, but snippets aren’t enough; developers need a first-class “read the document” flow with the same metadata/context cues docmgr is built around.
- `RelatedFiles` is central to docmgr’s value prop; without a way to open/view related code, the UI leaves a lot of power on the table.

### What worked
- The mock demonstrates a clear layout and interaction model that maps cleanly to our existing UI components + Bootstrap styling.
- The backend already has reliable primitives for reading markdown bodies and parsing frontmatter.

### What didn't work
- N/A (research/design step).

### What I learned
- Bootstrap helps with layout/styling, but markdown rendering and syntax highlighting need dedicated libraries.
- The safest v1 approach is “serve markdown as text + render on client without raw HTML” to avoid introducing an XSS surface.

### What was tricky to build
- Designing file serving endpoints that are ergonomic for the UI but still safe:
  - must not allow reading arbitrary files outside docs root / repo root,
  - must reject binary blobs and avoid memory blowups from huge files.

### What warrants a second pair of eyes
- Security boundary choices:
  - whether absolute paths should ever be accepted (even if inside root),
  - whether we should `EvalSymlinks` to prevent symlink-escape reads.
- Library choices for markdown/highlighting (bundle size vs simplicity).

### What should be done in the future
- Implement the proposed endpoints + doc/file viewer routes.
- Add optional doc asset serving (images) once there’s a concrete example needing it.

### Code review instructions
- Read the analysis doc:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/01-doc-serving-api-and-document-viewer-ui.md`

### Technical details
- Candidate frontend libs (client-side rendering):
  - `react-markdown`
  - `remark-gfm`
  - `rehype-highlight` + `highlight.js` CSS theme

## Step 8: Implement doc/file serving endpoints (safe text-only)

This step implemented the minimal backend needed for a real document viewer: endpoints to fetch a doc’s parsed frontmatter + markdown body, and to fetch an arbitrary text file (repo or docs) with strict safety guardrails. This unblocks the UI work for “open document” and “open related file”.

I treated this as a security-sensitive change even though docmgr is local-first: the endpoints must not allow path traversal, symlink escapes, or accidental binary/huge-file reads.

**Commit (code):** bacf9f9 — "web-ui: add doc viewer and safe file serving"

### What I did
- Added HTTP API routes:
  - `GET /api/v1/docs/get?path=...`
  - `GET /api/v1/files/get?path=...&root=repo|docs`
- Implemented safe root resolution + symlink-escape protection:
  - `internal/httpapi/path_safety.go` (`resolveFileWithin`, `tryRelWithin`)
- Implemented text-only file reading with guardrails:
  - max size cap (`2 MiB`) + `truncated` flag
  - reject NUL bytes (binary) and non-UTF8
- Added unit tests for traversal/symlink/binary/truncation:
  - `internal/httpapi/docs_files_test.go`

### Commands
- `gofmt -w internal/httpapi/*.go`
- `go test ./internal/httpapi -count=1`

### What worked
- Endpoints are wired into `internal/httpapi/server.go` and return the existing structured error shape.
- Safety checks cover:
  - `../` traversal (forbidden)
  - symlink escaping the allowed root (forbidden)
  - binary payloads (415)
  - huge files (truncated)

### What didn't work
- N/A (no blockers in this step).

### What I learned
- Path safety needs both:
  - a cheap “clean + rel” check (pre-filesystem), and
  - an `EvalSymlinks` check (post-filesystem) to block symlink escapes.

### What was tricky to build
- Keeping “we should support weird stored RelatedFiles paths” in mind without loosening the security boundary:
  - we still only serve files that ultimately resolve inside repo root or docs root.

### What warrants a second pair of eyes
- Confirm the root-resolution + `EvalSymlinks` behavior is correct and not overly strict for common RelatedFiles formats.
- Confirm `2 MiB` is a sensible cap for MVP file rendering.

### What should be done in the future
- Add doc asset/image serving once a concrete doc needs it.

### Code review instructions
- Start in:
  - `internal/httpapi/docs_files.go`
  - `internal/httpapi/path_safety.go`
  - `internal/httpapi/server.go`
  - `internal/httpapi/docs_files_test.go`

## Step 9: Add doc viewer + file viewer pages (frontend markdown rendering)

This step made the doc-serving endpoints usable by the UI: new SPA routes for viewing a document and viewing a file, plus client-side markdown rendering with code highlighting. It also wires “Open doc” and “Open related file” into the existing search preview.

**Commit (code):** bacf9f9 — "web-ui: add doc viewer and safe file serving"

### What I did
- Added new UI routes:
  - `/doc?path=...` → `ui/src/features/doc/DocViewerPage.tsx`
  - `/file?root=repo|docs&path=...` → `ui/src/features/file/FileViewerPage.tsx`
- Added API client endpoints:
  - `getDoc` + `getFile` in `ui/src/services/docmgrApi.ts`
- Added markdown rendering in the doc viewer:
  - `react-markdown` + `remark-gfm` (tables, etc)
  - `rehype-highlight` + highlight.js theme CSS
- Added code viewer highlighting via `highlight.js` (server returns inferred `language`)
- Wired navigation from search:
  - Preview pane “Open doc” button
  - Related files list “Open” button
- Checked off ticket tasks `30–38` in `tasks.md`.

### Commands
- `pnpm -C ui add react-markdown remark-gfm rehype-highlight highlight.js`
- `pnpm -C ui build`
- `go test ./... -count=1`
- `docmgr task check --tasks-file ttmp/.../tasks.md --id 30,31,32,33,34,35,36,37,38`

### What worked
- The UI can now show a full document rendered as markdown, with highlighted fenced code blocks.
- Related file paths open in a file viewer with highlighting (and remain safe because the server enforces roots + traversal/symlink guardrails).

### What didn't work
- N/A (no blockers in this step).

### What I learned
- Even for a local-first tool, “serve arbitrary file by path” endpoints must behave like an untrusted boundary: strict root allowlist + symlink escape checks + size limits.

### What was tricky to build
- Balancing UX and safety for weird `RelatedFiles` paths: the server primarily serves repo-relative paths, but also tries a resolver-based fallback while still enforcing repo-root containment.

### What warrants a second pair of eyes
- Confirm the fallback resolver behavior in `/api/v1/files/get` is correct and can’t be used to read outside repo/docs roots.
- Confirm highlight.js language mapping is acceptable for Go/TS/MD-heavy repos.

### What should be done in the future
- Add doc asset/image serving once needed by a real doc (to support `![](...)` in markdown).
- Improve viewer UX (TOC, anchor links, line numbers) if it becomes a daily driver.

### Code review instructions
- Start in:
  - `ui/src/features/doc/DocViewerPage.tsx`
  - `ui/src/features/file/FileViewerPage.tsx`
  - `ui/src/services/docmgrApi.ts`
  - `ui/src/features/search/SearchPage.tsx`
  - `internal/httpapi/docs_files.go`

## Step 10: Finish remaining UI MVP tasks + update docs/build configs

This step closed out the remaining “MVP” tasks in ticket 001 for the search UI: keyboard shortcuts, mobile-friendly preview and filters, and basic responsive polish. It also updates the public docs and build/release configs to reflect that the default build should include both FTS and embedded UI assets.

### What I did
- UI:
  - Added a real shortcuts modal (`?`) and implemented the MVP shortcut set:
    - `↑/↓` selection, `Enter` open selected doc, `Esc` close preview/modal, `Alt+1/2/3` mode switching, `Ctrl/Cmd+K` copy selected path.
  - Added mobile preview modal (instead of the desktop split pane).
  - Added a mobile filter drawer (modal) and responsive spacing tweaks.
  - Added “selected” styling for the active result.
- Docs:
  - Updated `pkg/doc/docmgr-http-api.md` with `/api/v1/docs/get` and `/api/v1/files/get`.
  - Updated `pkg/doc/docmgr-web-ui.md` with viewer routes and shortcuts.
- Build configs:
  - Updated `Makefile` default build/install tags to `sqlite_fts5,embed`.
  - Updated `.goreleaser.yaml` to build with `sqlite_fts5,embed` and generate UI assets.
- Checked off ticket tasks `18,20,22,23,24,26,29`.

### Commands
- `pnpm -C ui build`
- `go test ./... -count=1`
- `docmgr task check --tasks-file ttmp/.../tasks.md --id 18,20,22,23,24,26,29`

### What worked
- Mobile UX is now usable: tap a result → preview modal; filters open in a drawer modal.
- Keyboard navigation is good enough for daily driving.

### What was tricky to build
- Making keyboard shortcuts not interfere with typing into inputs (`isEditableTarget` gate).

### What warrants a second pair of eyes
- Confirm the keyboard handler dependency list in `SearchPage.tsx` doesn’t cause performance issues (it currently rebinds the listener when relevant state changes).
- Confirm the GoReleaser `flags` syntax is correct for your CI setup.

### What should be done in the future
- Consider code-splitting the UI bundle if size becomes a concern (highlight/markdown libs add weight).

## Step 11: Persist selection in URL, enable ctrl-click Open, and render snippet markdown

This step improves navigation ergonomics: selecting a result is now reflected in the URL so returning from `/doc` or `/file` restores the preview/selection state. The “Open” actions are real links so you can ctrl-click / middle-click to open in a new tab. Finally, snippets render as markdown and highlight the matching query terms.

### What I did
- Search UI:
  - Persist selection in URL via `sel=<docRelPath>` and restore it after search.
  - Persist mobile preview state via `preview=true`.
  - Convert “Open doc” / “Open” buttons to links (`<Link>`) so browsers can open new tabs.
  - Render snippets with `react-markdown` + `remark-gfm` and highlight query terms using `<mark>`.
- Doc viewer:
  - Convert related file “Open” to a real link (ctrl-click / new tab).
- Docs:
  - Documented `sel`/`preview` URL params in `pkg/doc/docmgr-web-ui.md`.
- Checked off ticket tasks `46–48`.

### Commands
- `pnpm -C ui build`
- `docmgr task check --tasks-file ttmp/.../tasks.md --id 46,47,48`

### What was tricky to build
- Highlighting matches without enabling raw HTML in markdown rendering (kept XSS surface low by transforming React nodes).

### What warrants a second pair of eyes
- Confirm URL restore logic for `sel` doesn’t fight with the existing filter URL-sync loop.

## Step 12: Design the ticket page API + Web UI (tabs + widgets)

This step turned the `sources/topic-page.md` concept into a concrete design for a ticket-specific page that can be implemented without shelling out to the CLI. The main outcome is a crisp API surface (`/api/v1/tickets/*`) and UI routing contract (`/ticket/:ticket?tab=...`) that fit cleanly into the existing `IndexManager` + React Router app.

**Commit (docs):** `cd16089` — "ticket(001): design ticket page API and web UI"

### What I did
- Wrote `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design/02-ticket-page-api-and-web-ui.md`:
  - routes, tab structure, data contracts, phased implementation plan
  - decisions about treating `index.md` frontmatter as canonical ticket metadata
- Expanded the ticket task list for API/UI implementation breakdown (commit `2fd4fb3`).

### Why
- The Search UI is useful, but it lacks a “ticket cockpit” view. The ticket page is the missing navigation hub for docs + tasks + graph + changelog.

### What worked
- The design aligned well with existing backend structure:
  - reuse `IndexManager` snapshot + `Workspace.QueryDocs` under the hood
  - keep doc/file viewers as separate routes (`/doc`, `/file`)

### What didn't work
- N/A (design-only).

### What I learned
- `tasks.md` / `changelog.md` are control docs that aren’t necessarily indexed, so ticket APIs must read them directly from disk (same safety constraints as `/files/get`).

### What was tricky to build
- Designing a ticket resolution mechanism that is stable and unambiguous:
  - ticket ID → find `DocType: index` → ticket dir
  - avoid relying on heuristic directory scans in the HTTP server.

### What warrants a second pair of eyes
- Confirm the contract for “ticket discovery” is correct: a ticket is identified by the unique `index.md` document (`DocType: index`) and its parent directory.

### What should be done in the future
- If we want to support tickets without frontmatter index docs, add a fallback resolver (but not needed for v1).

### Code review instructions
- Read `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design/02-ticket-page-api-and-web-ui.md` and sanity-check the endpoint shapes and URL params.

## Step 13: Add full-page ASCII screenshots to the ticket page design + upload

This step made the design doc more actionable by embedding “full page” ASCII screenshots for each tab plus a mobile layout. The immediate benefit is implementation alignment: it’s easy to notice missing widgets or navigation affordances while coding.

**Commit (docs):** `e6a5f29` — "docs(ticket): add full ASCII screenshots"

### What I did
- Added a dedicated `ASCII Screenshots (Full Pages)` section to the ticket design doc.
- Uploaded the updated design doc to reMarkable (overwrite).

### Commands
- `python3 /home/manuel/.local/bin/remarkable_upload.py --ticket-dir ... --mirror-ticket-structure --force ttmp/.../design/02-ticket-page-api-and-web-ui.md`

### What was tricky to build
- Keeping the screenshots “complete” without turning them into lorem ipsum walls; the goal is to specify layout + widgets, not the final copy.

### What warrants a second pair of eyes
- N/A (docs-only), but reviewers should confirm the screenshots reflect the intended feature set.

## Step 14: Implement ticket API endpoints on top of IndexManager (no CLI shell-outs)

This step implemented the new ticket API surface in the Go HTTP server and introduced reusable internal packages for ticket resolution, tasks parsing/mutation, and graph generation. The key outcome: the web UI can fetch ticket-specific views via `/api/v1/tickets/*` with safe filesystem access patterns.

**Commit (code):** `522e678` — "feat(ticket-page): add ticket API and UI"

### What I did
- Backend:
  - Added `internal/tickets/resolve.go` to resolve a ticket to its `index.md` and ticket dir via `Workspace.QueryDocs`.
  - Added `internal/tasksmd/tasksmd.go` to parse `tasks.md` into sections + stable numeric IDs (scan order) and to mutate tasks (check/uncheck + append).
  - Added `internal/ticketgraph/ticketgraph.go` to produce Mermaid DSL from docs + related files.
  - Added `internal/httpapi/tickets.go` implementing:
    - `GET /api/v1/tickets/get`
    - `GET /api/v1/tickets/docs` (cursor pagination)
    - `GET /api/v1/tickets/tasks`
    - `POST /api/v1/tickets/tasks/check`
    - `POST /api/v1/tickets/tasks/add`
    - `GET /api/v1/tickets/graph`
  - Mounted the handlers in `internal/httpapi/server.go`.
  - Added `internal/httpapi/tickets_test.go` to cover a basic end-to-end flow.
- UI (initial pass):
  - Added `ui/src/features/ticket/TicketPage.tsx` and route `/ticket/:ticket`.
  - Added RTK Query endpoints in `ui/src/services/docmgrApi.ts`.
  - Linked ticket badges from Search + Doc viewer to `/ticket/:ticket`.

### Commands
- `gofmt -w ...`
- `go test ./...`
- `pnpm -C ui lint`
- `pnpm -C ui build`

### What worked
- The API layer fits cleanly into the current server: it uses the same `IndexManager.WithWorkspace` pattern and reuses the existing cursor encoding.

### What didn't work
- Initially saw `404` on `/api/v1/tickets/*` while testing through Vite proxy: the running backend process was stale (see Step 16).

### What was tricky to build
- `tasks.md` isn’t guaranteed to be indexed; reading/mutating it safely required:
  - root-bound resolution (`resolveFileWithin`)
  - best-effort behavior when tasks.md doesn’t exist (return `exists=false`).

### What warrants a second pair of eyes
- Confirm task ID stability semantics are acceptable (IDs are “scan order at read time”; mutations preserve line positions but edits may renumber later).
- Confirm graph canonicalization keys match `QueryDocs` reverse-lookup behavior closely enough for UX.

### What should be done in the future
- Add stronger per-handler tests around path safety (e.g., tasks path traversal attempts), similar to `docs_files_test.go`.

### Code review instructions
- Start with:
  - `internal/httpapi/tickets.go`
  - `internal/tickets/resolve.go`
  - `internal/tasksmd/tasksmd.go`
  - `internal/ticketgraph/ticketgraph.go`
  - `ui/src/features/ticket/TicketPage.tsx`

## Step 15: Improve ticket Overview and render the Mermaid graph in the UI

This step made the ticket page feel “real”: the Overview tab now shows key docs, open tasks, and renders `index.md` content; the Graph tab renders Mermaid client-side instead of showing raw DSL.

**Commit (code):** `4a82f9d` — "feat(ticket-page): overview and mermaid graph"

### What I did
- UI:
  - Overview tab fetches and renders `index.md` body markdown via the existing `/api/v1/docs/get`.
  - Overview tab shows “Key documents” and “Open tasks” (quick checkboxes).
  - Graph tab renders Mermaid using a small React wrapper component:
    - `ui/src/components/MermaidDiagram.tsx`
    - dependency: `pnpm -C ui add mermaid`
  - Kept a `<details>` block with raw Mermaid DSL for debugging.
- Ticket tasks bookkeeping:
  - Checked off the “Overview tab” task in `ttmp/.../tasks.md`.

### Commands
- `pnpm -C ui add mermaid`
- `pnpm -C ui lint`
- `pnpm -C ui build`

### What was tricky to build
- Mermaid is heavy and significantly increases the production bundle. Acceptable for now, but a clear candidate for code-splitting later.

### What warrants a second pair of eyes
- Confirm `mermaid.initialize({ securityLevel: 'strict' })` is sufficient and that we’re not accidentally enabling HTML injection via Mermaid.

## Step 16: Debug 404s for ticket endpoints (stale backend process)

I hit a confusing failure where the UI and `curl` against `http://localhost:3000/api/v1/tickets/*` returned `404`, despite the code being implemented. The underlying issue was simply that the running `go run ... api serve` process was started before the new commit and hadn’t been restarted.

### What I did
- Verified:
  - `curl http://127.0.0.1:3001/api/v1/healthz` → `200 OK`
  - `curl http://127.0.0.1:3001/api/v1/tickets/get?...` → `404 Not Found`
  - which strongly implies “old server binary still running”.
- Restarted the backend process in tmux (`docmgr-dev` pane 0).
- Re-verified:
  - `curl http://127.0.0.1:3001/api/v1/tickets/get?...` → `200 OK`
  - `curl http://127.0.0.1:3000/api/v1/tickets/get?...` → `200 OK` (via Vite proxy).

### Commands
- Backend restart:
  - `go run -tags sqlite_fts5 ./cmd/docmgr api serve --addr 127.0.0.1:3001 --root ttmp`
- Checks:
  - `curl -i http://127.0.0.1:3001/api/v1/tickets/get?ticket=DOCMGR-002`
  - `curl -i http://127.0.0.1:3000/api/v1/tickets/get?ticket=DOCMGR-002`

### What I learned
- When testing new endpoints with `go run`, a green health check doesn’t guarantee you’re running the latest code; always validate a newly added route explicitly.

## Step 17: Start React UI architecture audit + Workspace page widget planning

I started a focused audit of the current React SPA to understand the architectural “shape” of the existing pages (Search/Doc/File/Ticket) before designing the new Workspace page. The goal is to propose a widget/component hierarchy that stays aligned with the current code, but also nudges us toward a coherent, reusable design system.

In parallel, I created a dedicated analysis document for this audit so that the conclusions (folder layout, reusable widgets, component sizing boundaries, and design-system primitives) are captured in one place and can guide the final set of pages.

### What I did
- Located the React app entry points and routes:
  - `ui/src/main.tsx` (bootstrap import + root render)
  - `ui/src/App.tsx` (react-router routes)
- Skimmed the shared API client and state setup:
  - `ui/src/services/docmgrApi.ts` (RTK Query API surface)
  - `ui/src/app/store.ts` (Redux store wiring)
- Read the Workspace page source spec:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md`
- Created a new ticket analysis doc to write into:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`

### Why
- The current pages already contain “proto-widgets” embedded inside large page files; extracting the implicit patterns now should reduce future thrash when adding Workspace/Home/Topics/etc.
- The Workspace page spec is bigger than the current pages; designing its widget boundaries up front is the easiest way to keep it maintainable.

### What worked
- The codebase already follows a reasonable high-level split (`app/`, `features/`, `components/`, `services/`), and it’s using RTK Query (good foundation for consistent data access).

### What didn't work
- Minor workflow footgun: I typed `ls -ლა` (unicode flags) by accident and got `ls: invalid option -- 'á'`. No functional impact, just noise worth avoiding.

### What I learned
- The SPA is Vite + React Router + Redux Toolkit/RTK Query + Bootstrap, with the route-level “pages” under `ui/src/features/*/*Page.tsx`.
- Several “page-local components” (helpers, rendering widgets, parsing/highlighting) currently live inside the page files; that’s likely where we’ll find the best extraction candidates for a shared widget library.

### What was tricky to build
- N/A (this step was research + doc scaffolding).

### What warrants a second pair of eyes
- The proposed widget/design-system boundaries should be reviewed against the intended future pages (Workspace/Tickets/Topics/Recent) so we don’t overfit to the current Search/Ticket UI.

### What should be done in the future
- Continue the audit by mapping each page into: layout shell → widgets → smaller components, then design a target folder architecture that supports re-use.

### Code review instructions
- Start with the analysis doc draft:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`
- Then skim the current “page” entry points:
  - `ui/src/App.tsx`
  - `ui/src/features/search/SearchPage.tsx`
  - `ui/src/features/ticket/TicketPage.tsx`

### Technical details
- Commands I used:
  - `rg -n "createRoot|BrowserRouter|Routes|rtk-query" ui/src -S`
  - `docmgr doc add --ticket 001-ADD-DOCMGR-UI --doc-type analysis --title "React UI architecture + Workspace page widget system"`

## Step 18: Map current pages into widgets + draft the Workspace nav widget inventory

I went deeper on the “big page” files (Search and Ticket) to identify concrete widget boundaries and repeatable UI primitives. The emphasis here is to treat the new Workspace navigation experience as a *product surface* with a stable shell and swappable content, rather than accreting more page-specific logic into monolithic route components.

I also walked the `sources/workspace-page.md` ASCII designs and translated them into an initial widget inventory (shell widgets + page-specific widgets), so we can design and implement the Workspace pages by composing reusable parts instead of rewriting the same patterns repeatedly.

### What I did
- Quantified current “page sizes” to identify the biggest extraction candidates:
  - `SearchPage.tsx` (~1649 LOC) and `TicketPage.tsx` (~653 LOC)
- Read through the major sections of:
  - `ui/src/features/search/SearchPage.tsx` (URL sync, keyboard shortcuts, filters/drawer, results/preview)
  - `ui/src/features/ticket/TicketPage.tsx` (tab bodies, repeated list/preview patterns)
  - Shared components and styling:
    - `ui/src/components/DocCard.tsx`
    - `ui/src/App.css` and `ui/src/index.css`
- Read the Workspace navigation page designs:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md`
- Drafted the first pass of:
  - “What’s getting big”
  - “Extraction candidates”
  - “Shell widgets”
  - “Workspace page widget breakdown”
  into the new analysis doc.

### Why
- The Workspace pages (Home/Tickets/Topics/Recent) will introduce multiple new large UI surfaces; without a widget/design-system strategy, we’ll end up duplicating toasts/errors/modals/layout logic across pages.
- The current Search page already contains many implicit widgets; making them explicit now gives us reusable building blocks for Workspace pages.

### What worked
- Bootstrap + RTK Query already provide strong “defaults”; we can create a coherent design system mostly as thin wrappers + tokenized CSS, rather than inventing a new component library from scratch.

### What didn't work
- N/A (no implementation yet; this was architecture mapping + doc writing).

### What I learned
- `DocCard` is already shared between Search and Ticket, but its styling and naming are Search-centric (`result-card`), which is a signal that we should separate:
  - design-system “Card” primitives vs
  - domain-specific “DocCard/TicketCard/etc”.

### What was tricky to build
- N/A (research + documentation).

### What warrants a second pair of eyes
- The proposed directory/layout reorg (`pages/`, `widgets/`, `ui/shared/`) should be validated against how we want to evolve the codebase (e.g. whether to adopt a formal feature-sliced structure or keep the current `features/*Page.tsx` pattern and only add `widgets/`).

### What should be done in the future
- Extract shared primitives first (`useToast`, `useClipboard`, `ApiErrorAlert`, `PageHeader`) before moving files around, so refactors stay incremental and low risk.

### Code review instructions
- Read the analysis doc (new content starts near the top and continues through the widget inventory):
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`
- Skim the “big page” files to see the implicit widgets in place today:
  - `ui/src/features/search/SearchPage.tsx`
  - `ui/src/features/ticket/TicketPage.tsx`

### Technical details
- Commands I used:
  - `wc -l ui/src/features/search/SearchPage.tsx ui/src/features/ticket/TicketPage.tsx`
  - `rg -n "^## " ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/sources/workspace-page.md`

## Step 19: Specify widget sizing rules + a concrete file tree for a coherent design system

I tightened the analysis doc into something implementable: not just “extract widgets”, but explicit rules of thumb for ownership/sizing, a proposed directory tree, and clear guidance on how to incrementally migrate without a big-bang rewrite. The intent is to make the upcoming Workspace pages (Home/Tickets/Topics/Recent) feel cohesive while still being achievable in small refactor commits.

I also added a “data dependency” section to keep widget boundaries aligned with the Workspace REST API design, so we don’t end up with ad-hoc fetching logic scattered across components.

### What I did
- Expanded `analysis/02-react-ui-architecture-workspace-page-widget-system.md` with:
  - widget/page sizing guidance
  - explicit shell widget mapping to current page headers
  - data dependency mapping to the Workspace REST API design doc
  - a concrete proposed `ui/src/` directory tree (pages/widgets/ui/lib)
  - additional “too-big” refactor triggers (dup helpers, mixed CSS)

### Why
- The organization and naming conventions are the “design system” as much as the CSS is; if we don’t define these now, each new page will invent its own patterns.

### What worked
- The existing code already suggests natural boundaries (e.g. Ticket tab bodies, Search’s filter/result/preview areas); the file tree proposal just gives those boundaries a stable home.

### What didn't work
- N/A.

### What I learned
- Even in a Bootstrap-heavy UI, we benefit from explicit design-system primitives (PageHeader, ApiErrorAlert, EmptyState), because they eliminate duplication and keep interactions consistent across pages.

### What was tricky to build
- N/A (documentation only).

### What warrants a second pair of eyes
- Sanity check that the proposed tree doesn’t conflict with team preferences (e.g. keeping everything under `features/` vs introducing `pages/` + `widgets/`).

### What should be done in the future
- Before moving files, implement the shared primitives and switch existing pages over; once behavior is stable, migrate file locations.

### Code review instructions
- Review the new “implementable” sections in:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`

### Technical details
- No new commands beyond editing the analysis doc.

## Step 20: Link the analysis doc to key code + update the ticket changelog

I finalized the documentation bookkeeping: validated frontmatter, linked the new analysis doc to the most relevant UI/code/spec files, and recorded the work in the ticket changelog. This keeps the “architecture plan” discoverable from both doc search and from the document’s frontmatter relationships.

### What I did
- Validated doc frontmatter:
  - `docmgr validate frontmatter --doc /abs/.../analysis/02-react-ui-architecture-workspace-page-widget-system.md`
  - `docmgr validate frontmatter --doc /abs/.../reference/01-diary.md`
- Related key files to the analysis doc (kept to 7 total related files):
  - `ui/src/App.tsx`, `ui/src/services/docmgrApi.ts`
  - `ui/src/features/search/SearchPage.tsx`, `ui/src/features/ticket/TicketPage.tsx`
  - `ui/src/components/DocCard.tsx`
  - `sources/workspace-page.md`, `design/03-workspace-rest-api.md`
- Updated the ticket changelog with an entry for the new analysis doc.

### Why
- Relationships and changelog entries make it much easier to find this decision record later (and avoid re-deriving the same conclusions when we start moving code around).

### What worked
- Using absolute paths for `--file-note` avoids ambiguity and copy/paste errors.

### What didn't work
- I initially ran `docmgr validate frontmatter --doc ttmp/...` and it resolved as `ttmp/ttmp/...` under the docs root. Using an absolute path fixed it.

### What I learned
- `docmgr doc relate` supports removals via `--remove-files`, which is useful to keep `RelatedFiles` within the recommended size.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- N/A (bookkeeping only).

### What should be done in the future
- Once code refactors begin, keep the same discipline: relate the extracted widget/primitives to the focused doc(s) and record the rationale in the changelog per step.

### Code review instructions
- Review the analysis doc and its `RelatedFiles` section:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`
- Review the changelog entry:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/changelog.md`

### Technical details
- Commands I used:
  - `docmgr doc relate --doc /abs/.../analysis/02-... --file-note "..."`
  - `docmgr doc relate --doc /abs/.../analysis/02-... --remove-files "/abs/path/to/file"`
  - `docmgr changelog update --ticket 001-ADD-DOCMGR-UI --entry "..." --file-note "..."`

## Step 21: Update the new analysis doc (topics + CSS strategy + extraction sequence)

I incorporated follow-up edits into the new analysis doc: updated the frontmatter topics to better reflect the document’s purpose (UI/web/workspace) and expanded the content with a pragmatic CSS strategy and a suggested extraction sequence. This makes the doc more directly actionable as a “playbook” for refactoring Search/Ticket and for building the upcoming Workspace navigation pages.

### What I did
- Updated analysis doc frontmatter Topics to: `docmgr, ui, web, workspace, ux`.
- Added:
  - a CSS strategy section (split design-system utilities from page-specific layout)
  - a suggested extraction sequence (shared primitives → Search widgets → Ticket tabs → AppShell + Workspace pages)
- Re-validated frontmatter:
  - `docmgr validate frontmatter --doc /abs/.../analysis/02-react-ui-architecture-workspace-page-widget-system.md`

### Why
- The doc is meant to guide the multi-page UI build; it should be tagged and structured so it’s easy to discover and apply later.

### What worked
- The frontmatter schema accepted the topic updates and the doc remains valid.

### What didn't work
- N/A.

### What I learned
- The most important “design system” decision is separating shared utilities/tokens from one-page layout CSS early; it prevents style coupling as more pages are added.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- Confirm the chosen extraction order aligns with implementation priorities (e.g. if Workspace pages must land sooner than Search refactors).

### What should be done in the future
- When we start implementing the refactors, keep commits scoped to one extraction and update this doc with the “final tree” that actually emerges.

### Code review instructions
- Review:
  - `ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/analysis/02-react-ui-architecture-workspace-page-widget-system.md`

### Technical details
- Commands:
  - `docmgr validate frontmatter --doc /home/manuel/.../analysis/02-react-ui-architecture-workspace-page-widget-system.md --suggest-fixes`

## Step 22: Upload the diary and Workspace REST API design doc to reMarkable (overwrite)

I uploaded the updated diary and the Workspace REST API design doc to the reMarkable device. The first upload attempt failed because the PDFs already existed on-device; I reran the upload with `--force` to overwrite them with the latest versions.

### What I did
- Ran a ticket-aware upload with mirrored structure and overwrite enabled:
  - `python3 /home/manuel/.local/bin/remarkable_upload.py --force --ticket-dir ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui --mirror-ticket-structure ...`
- Confirmed successful replacements:
  - `01-diary.pdf` → `ai/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/`
  - `03-workspace-rest-api.pdf` → `ai/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/design/`

### Why
- Keeping the reMarkable copies current makes review/annotation easier without needing a browser session.

### What worked
- `--mirror-ticket-structure` keeps files organized and `--force` resolves conflicts cleanly when PDFs already exist.

### What didn't work
- Without `--force`, `rmapi put` failed with: `entry already exists (use --force to recreate, --content-only to replace content)`.

### What I learned
- For iterative docs, it’s better to either (a) overwrite explicitly with `--force`, or (b) upload into a versioned folder via `--remote-ticket-root` to preserve older PDFs.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- Decide on a consistent convention: always overwrite PDFs for “living docs”, or always upload into a new versioned subfolder to preserve history.

### Code review instructions
- N/A (upload-only).

### Technical details
- Command:
  - `python3 /home/manuel/.local/bin/remarkable_upload.py --force --ticket-dir ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui --mirror-ticket-structure ttmp/.../reference/01-diary.md ttmp/.../design/03-workspace-rest-api.md`
