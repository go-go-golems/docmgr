---
Title: Diary
Ticket: CLEANUP-LEGACY-WALKERS
Status: active
Topics:
    - refactor
    - tickets
    - docmgr-internals
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/commands/changelog.go
      Note: |-
        Phase 1.4 migrated suggestion doc-scan to Workspace.QueryDocs (commit 09e1e6f)
        Phase 4.1 removed remaining findTicketDirectory callsites (commit 3751433)
    - Path: pkg/commands/doc_move.go
      Note: Phase 3.2 migrated doc move to Workspace.QueryDocs (commit 770e33f)
    - Path: pkg/commands/import_file.go
      Note: Phase 4.1 deleted findTicketDirectory helper (commit 3751433)
    - Path: pkg/commands/layout_fix.go
      Note: Phase 3.3 migrated layout-fix discovery to Workspace.QueryDocs (commit c72e0db)
    - Path: pkg/commands/meta_update.go
      Note: Phase 2.2 migrated ticket discovery + enumeration to Workspace.QueryDocs (commit 3458a46)
    - Path: pkg/commands/status.go
      Note: Phase 1.1 migration to Workspace.QueryDocs (commit f61606c)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T10:38:05.321452661-05:00
---







# Diary

## Goal

Capture the step-by-step implementation of the **cleanup phase** after the Workspace+SQLite refactor: migrating remaining commands off legacy walkers (`CollectTicketWorkspaces`, `findTicketDirectory`, `filepath.Walk*`, manual frontmatter parsing) onto the canonical `workspace.Workspace` + `Workspace.QueryDocs` API.

## Context

This ticket is a follow-on to **REFACTOR-TICKET-REPOSITORY-HANDLING** which introduced:

- `internal/workspace/workspace.go` (discovery + resolver)
- `internal/workspace/index_builder.go` (in-memory SQLite index ingestion)
- `internal/workspace/query_docs.go` (canonical lookup + filters + diagnostics)

The goal here is to make command behavior consistent by **removing duplicated discovery/walk logic** across `pkg/commands/*` and then deleting the legacy helpers once no longer used.

## Quick Reference

```bash
# Code validation
gofmt -w pkg/commands/status.go
go test ./... -count=1

# Ticket workflow
docmgr task check --ticket CLEANUP-LEGACY-WALKERS --id 1
docmgr doc relate --doc ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/reference/01-diary.md \
  --file-note "/home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/status.go:Phase 1.1 migration to Workspace.QueryDocs"
docmgr changelog update --ticket CLEANUP-LEGACY-WALKERS --entry "..." --file-note "/abs/path:note"
```

## Usage Examples

Use this diary while reviewing the cleanup PRs. Each step should contain:

- the intent of the change (what legacy helper got removed),
- the exact files touched,
- the validation commands run,
- and any compatibility notes (behavior preserved vs intentionally changed).

## Related

- `ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/design/01-cleanup-overview-and-migration-guide.md`
- `ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/tasks.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/15-diary.md`

## Step 1: Migrate `status` to Workspace+QueryDocs (Phase 1.1)

This step removes the last major “status-style” legacy traversal: enumerating tickets with `CollectTicketWorkspaces` and then doing a per-ticket `filepath.Walk` to count docs by DocType. The replacement builds the Workspace index once and computes the same aggregates from `QueryDocs` results, keeping output stable while aligning semantics with the rest of the tool.

**Commit (code):** `f61606c` — "Cleanup: migrate status to QueryDocs"

### What I did
- Updated `pkg/commands/status.go` to use `workspace.DiscoverWorkspace` + `ws.InitIndex` + `ws.QueryDocs` instead of:
  - `workspace.CollectTicketWorkspaces`, and
  - `filepath.Walk` + `readDocumentFrontmatter`.
- Introduced a small helper (`computeStatusTickets`) that:
  - scans `QueryDocs(ScopeRepo)` results,
  - identifies ticket “index” docs (`index.md` / `DocType=index`) for ticket metadata, and
  - counts non-index docs by DocType (`design-doc`, `reference`, `playbook`).
- Ran `gofmt` and `go test ./... -count=1`.

### Why
- Make `status` consistent with the canonical ingestion + skip rules and avoid ad-hoc traversal.
- Reduce duplicated frontmatter parsing and ticket discovery logic (prerequisite for deleting legacy helpers later).

### What worked
- Unit tests stayed green.
- No behavior changes needed in the CLI output formatting; the change is purely internal.

### What didn't work
- Nothing notable; this was a mechanical port once `QueryDocs` stabilized.

### What I learned
- `QueryDocs` is sufficient to compute status aggregates without any filesystem walking, but preserving behavior requires explicitly including “special path” categories where the legacy code would have counted them (control docs, scripts, archive).

### What was tricky to build
- **Replicating ticket discovery correctly** without `CollectTicketWorkspaces`: we now infer the ticket directory by locating the `index.md` row for each ticket. This relies on `DocType=index` and/or basename `index.md`, which is correct for docmgr tickets but worth keeping in mind for any “non-standard” docs layouts.
- **Visibility defaults**: `QueryDocs` hides certain tagged categories unless explicitly enabled via options. For status, we had to decide which categories should contribute to counts and set the include options accordingly.

### What warrants a second pair of eyes
- **Counting semantics**: confirm we’re still counting the same set of files the old `filepath.Walk(ticketDir)` would have seen (especially around tagged paths like `archive/`, `scripts/`, control docs, and anything skipped by canonical ingestion rules).
- **Index doc identification**: confirm the `index.md`/`DocType=index` detection is the right invariant for status aggregation (and doesn’t accidentally “promote” a non-ticket index doc).

### What should be done in the future
- If reviewers/users report “missing docs” in counts, treat it as a **QueryDocs visibility/tagging semantics** question (not a reason to reintroduce walking). Decide whether the status command should include/exclude additional tagged categories by adjusting QueryDocs options, and document the chosen semantics.
- If the `index.md`/`DocType=index` assumption ever becomes invalid for some workspace layouts, that’s an **architecture constraint**: either enforce it harder (doctor/scaffold) or introduce a single canonical way to identify ticket roots—do not reintroduce per-command heuristics.

### Code review instructions
- Start in `pkg/commands/status.go` and review `computeStatusTickets`.
- Smoke:
  - `docmgr status --summary-only`
  - `docmgr status --ticket CLEANUP-LEGACY-WALKERS`

## Step 2: Migrate `list tickets` to Workspace+QueryDocs (Phase 1.2)

This step migrates `list tickets` off the legacy `CollectTicketWorkspaces` walker and onto `Workspace.QueryDocs` with a `DocType=index` filter. The result is that ticket enumeration, metadata hydration, and ordering are now backed by the same canonical ingestion+skip rules used by `list docs`, `search`, `doctor`, and `relate`, while keeping the human output stable.

**Commit (code):** `024993a` — "Cleanup: migrate list tickets to QueryDocs"

### What I did
- Updated `pkg/commands/list_tickets.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - `QueryDocs(ScopeRepo, Filters{DocType=index, Ticket, Status}, OrderByLastUpdated DESC)` instead of `CollectTicketWorkspaces`.
- Ran `gofmt` and `go test ./... -count=1`.

### Why
- Remove duplicated ticket discovery logic and centralize semantics in QueryDocs.
- Keep `list tickets` consistent with the Workspace refactor’s “single source of truth”.

### What worked
- Tests and lint stayed green.
- No changes required to the human-facing Markdown rendering or template schema behavior.

### What was tricky to build
- **Path handling across output modes**: `QueryDocs` returns absolute paths; the legacy code used `TicketWorkspace.Path` (ticket dir) directly. We now compute a root-relative path for display, which needs careful handling when `--root` is passed or when root is resolved via config.
- **Ordering**: old behavior was “newest first” based on `index.md` frontmatter `LastUpdated`; this is now driven by `OrderByLastUpdated` and verified in code.

### Behavior notes
- **Ticket filter semantics** are now **exact match** (via `ticket_id = ?`) rather than substring matching. This was an intentional cleanup per the migration guide.

### What warrants a second pair of eyes
- **Filter semantics**: confirm that switching from substring match to exact match is acceptable across both human and glaze output, and that it’s documented clearly enough for users.
- **Tasks counting**: we now derive the ticket dir from the `index.md` path; verify that `countTasksInTicket(ticketDir)` still points at the correct `tasks.md` for all ticket layouts.

### What should be done in the future
- If any scripts/docs were relying on substring semantics for `--ticket`, **update those call sites** to pass the exact ticket ID (or move them to a different selection mechanism). Per the spec, we should not add compatibility flags like `--ticket-contains`.
- If users get confused by the behavior change, the right follow-up is **documentation** (help text / docs) and **tests** that codify exact-match semantics—not a compatibility layer.

### Code review instructions
- Start in `pkg/commands/list_tickets.go` and review `queryTicketIndexDocs`.
- Smoke:
  - `docmgr list tickets`
  - `docmgr list tickets --ticket CLEANUP-LEGACY-WALKERS`

## Step 3: Migrate `list` to Workspace+QueryDocs (Phase 1.3)

This step migrates the legacy `docmgr list` command (workspaces listing) off `CollectTicketWorkspaces` and onto `Workspace.QueryDocs` with a `DocType=index` filter. The output is still the same conceptual table (ticket/title/status/topics/path/last_updated), but it is now derived from the canonical in-memory index.

**Commit (code):** `0ec09da` — "Cleanup: migrate list to QueryDocs"

### What I did
- Updated `pkg/commands/list.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - `QueryDocs(ScopeRepo, Filters{DocType=index, Ticket, Status})` instead of `CollectTicketWorkspaces`.
- Ran `gofmt` and `go test ./... -count=1`.

### Why
- Remove the last remaining direct `CollectTicketWorkspaces` usage in Phase 1 commands.
- Consolidate ticket/workspace enumeration behind QueryDocs (single semantics).

### What worked
- Tests and lint stayed green.

### What was tricky to build
- **Path output**: QueryDocs gives an `index.md` file path; we derive the ticket directory and emit a root-relative path for consistency.
- **Ordering semantics**: Query order is now explicit (`OrderByLastUpdated DESC`) instead of relying on walker ordering.

### What warrants a second pair of eyes
- **Column semantics**: confirm `path` should be ticket-dir-root-relative (not the `index.md` file path, and not absolute) now that we’ve removed compatibility constraints.
- **Filter semantics**: confirm ticket filtering is exact-match (QueryDocs semantics) and that’s acceptable for this legacy alias command.

### What should be done in the future
- If `docmgr list` and `docmgr list tickets` diverge again, treat it as a **regression**: they should be thin wrappers over the same QueryDocs semantics, not independent walkers.
- If the `path` column becomes contentious (absolute vs root-relative vs “ticket dir”), resolve it by **declaring a single contract** and updating downstream scripts/tests accordingly (no shims).

### Code review instructions
- Start in `pkg/commands/list.go`.
- Smoke:
  - `docmgr list`
  - `docmgr list --ticket CLEANUP-LEGACY-WALKERS`

## Step 4: Migrate `changelog update --suggest` doc-scan to QueryDocs (Phase 1.4)

This step removes the last place in Phase 1 where we were still doing a manual `filepath.Walk` over markdown files and parsing frontmatter ad-hoc to seed suggestions. The suggestions pipeline now sources “referenced by documents” from the Workspace index (`QueryDocs`), which keeps semantics centralized and removes duplicated skip/parsing logic.

**Commit (code):** `09e1e6f` — "Cleanup: migrate changelog suggest to QueryDocs"

### What I did
- Updated `pkg/commands/changelog.go` suggestion mode to build a Workspace index and use `QueryDocs` to collect `RelatedFiles` from existing docs, instead of:
  - `filepath.Walk(searchRoot)` and
  - `readDocumentFrontmatter(path)`.
- Kept the other suggestion sources unchanged (git history, ripgrep, git status).

### Why
- Eliminate the last Phase 1 manual doc traversal/parsing and consolidate semantics in QueryDocs.
- Reduce duplicated skip rules and parsing behavior across commands.

### What worked
- Unit tests remained green.
- Suggestion output is still explainable by “sources” (documents/git/ripgrep/status), but document-derived suggestions now come from a single canonical index.

### What was tricky to build
- **Scope selection**: ticket-scoped suggestions should query `ScopeTicket` when we have `--ticket`, otherwise `ScopeRepo` (equivalent to scanning the whole root).
- **Visibility semantics**: deciding which tagged categories of docs should contribute to “referenced by documents” suggestions is now expressed via QueryDocs options (not via walker heuristics).

### What warrants a second pair of eyes
- **Suggestion semantics shift**: confirm it’s acceptable that the doc-derived suggestion source now reflects QueryDocs’ ingestion/visibility behavior rather than “whatever `filepath.Walk` happened to see”.
- **ChangelogFile mode**: if users call `--changelog-file` without `--ticket`, confirm the repo-scope query is the right contract for doc-derived suggestions.

### What should be done in the future
- If reviewers observe missing/extra “referenced by documents” suggestions, treat it as a **QueryDocs visibility/options contract** issue (adjust QueryDocs options or documentation), not a reason to reintroduce walking/parsing.
- Once Phase 2 removes `findTicketDirectory`, consider whether `--changelog-file` should infer ticket scope from the provided file path (architecture choice), but keep it centralized (no per-command heuristics).

### Code review instructions
- Start in `pkg/commands/changelog.go` and review the `s.Suggest` block inside `RunIntoGlazeProcessor`.
- Smoke:
  - `docmgr changelog update --ticket CLEANUP-LEGACY-WALKERS --suggest --query QueryDocs`

## Step 5: Migrate `add` ticket discovery to Workspace+QueryDocs (Phase 2.1)

This step removes `add`’s dependency on the legacy `findTicketDirectory` helper. Instead of rediscovering ticket directories by walking/collecting workspaces, `docmgr add` now resolves the ticket workspace by querying the Workspace index for the ticket’s `index.md` and deriving the ticket directory from that canonical result.

**Commit (code):** `a512739` — "Cleanup: migrate add ticket discovery to QueryDocs"

### What I did
- Updated `pkg/commands/add.go` to resolve the ticket directory via `workspace.DiscoverWorkspace` + `ws.InitIndex` + `ws.QueryDocs(ScopeTicket, DocType=index)` instead of `findTicketDirectory`.
- Kept the write-path behavior the same (create a new doc under `<doc-type>/` and seed metadata from the ticket index doc).

### Why
- Remove legacy ticket discovery from write-path commands so we can delete `findTicketDirectory` later.
- Make ticket resolution consistent across commands by using the same canonical index and normalization rules.

### What worked
- Tests and lint stayed green.
- `docmgr add` no longer depends on legacy discovery helpers.

### What was tricky to build
- **Root resolution**: Workspace discovery may resolve `--root` differently than a naive relative path. The command now pins its root to the Workspace-resolved root to keep template/guideline loading consistent.
- **Index uniqueness**: `QueryDocs(ScopeTicket, DocType=index)` is assumed to return exactly one result; ambiguity is treated as an error.

### What warrants a second pair of eyes
- **Ticket workspace selection**: confirm that using the ticket’s `DocType=index` doc as the canonical anchor is the correct invariant for all supported layouts.
- **Error messaging**: confirm the “ticket not found / ambiguous” errors are actionable for users (especially during partial/scaffolded tickets).

### What should be done in the future
- When Phase 4 deletes `findTicketDirectory`, ensure all remaining write-path commands use this same QueryDocs-based contract (no fallback walkers).
- If multiple index-doc edge cases show up in real repos, treat it as an **architecture constraint** (enforce uniqueness / improve doctor findings), not a reason to reintroduce ad-hoc search.

### Code review instructions
- Start in `pkg/commands/add.go` and review ticket resolution in `createDocument`.

## Step 6: Migrate `meta update` to Workspace+QueryDocs (Phase 2.2)

This step removes `meta update`’s dependency on legacy ticket discovery and manual doc enumeration (`findTicketDirectory` + `findMarkdownFiles`). Ticket resolution and “which docs to update” are now derived from the Workspace index via `QueryDocs`.

**Commit (code):** `3458a46` — "Cleanup: migrate meta update to QueryDocs"

### What I did
- Updated `pkg/commands/meta_update.go` so that when `--ticket` is used:
  - it discovers the Workspace and builds the index once, then
  - uses `QueryDocs` to select either:
    - the ticket’s `index.md` (default), or
    - all docs of `--doc-type` within the ticket.
- Deleted the legacy `findMarkdownFiles` walker/parsing helper from this command.

### Why
- Centralize “what docs exist” and “how we filter them” behind QueryDocs.
- Remove duplicated traversal/parsing and unblock deleting `findTicketDirectory` later.

### What worked
- Tests and lint stayed green.
- The command still performs the same write-path operation (read frontmatter → update field → write back).

### What was tricky to build
- **Path normalization**: `QueryDocs` returns slash-cleaned absolute paths; write-path functions expect filesystem paths. The implementation normalizes via `filepath.FromSlash`.
- **DocType scoping**: enumeration is now driven by the index’s parsed `doc_type` field, not by “walk everything and parse on the fly”.

### What warrants a second pair of eyes
- **Enumeration semantics**: confirm that using `ScopeTicket + DocType filter` is the intended contract for “update all docs of this type”.
- **Index doc selection**: confirm the “default when only --ticket is specified is index.md” rule is still correct and robust.

### What should be done in the future
- Once Phase 4 removes `findTicketDirectory`, verify no other commands reintroduce manual enumeration helpers like `findMarkdownFiles`.
- If someone reports “meta update skipped a doc”, treat it as a QueryDocs ingestion/visibility semantics issue first, not a reason to resurrect walkers.

### Code review instructions
- Start in `pkg/commands/meta_update.go` and review `applyMetaUpdate`.

## Step 7: Migrate `tasks` ticket discovery to Workspace+QueryDocs (Phase 2.3)

This step removes `tasks`’ remaining legacy ticket directory heuristics (substring directory scan + `findTicketDirectory` fallback). When `--tasks-file` is not provided, the command now resolves the ticket’s `tasks.md` by querying the Workspace index and selecting the `tasks.md` control doc within the ticket scope.

**Commit (code):** `234f42c` — "Cleanup: migrate tasks ticket discovery to QueryDocs"

### What I did
- Updated `pkg/commands/tasks.go` so `loadTasksFile` is Workspace-backed when `--tasks-file` is not set:
  - `DiscoverWorkspace` + `InitIndex`
  - `QueryDocs(ScopeTicket, IncludeControlDocs=true)` and pick the `tasks.md` doc by basename
- Threaded `context.Context` through `loadTasksFile` and `editTaskLine` callsites so we can reuse the caller’s context.

### Why
- Remove the last Phase 2 usage of `findTicketDirectory` (and the ad-hoc substring heuristics) so we can delete the helper later.
- Make “control doc resolution” consistent with the canonical index semantics.

### What worked
- Tests and lint stayed green.
- `tasks` commands now resolve `tasks.md` without any filesystem directory scanning.

### What was tricky to build
- **Doc selection**: QueryDocs doesn’t have a “basename filter”, so we query the ticket scope and then select `tasks.md` in-memory.
- **Visibility**: `tasks.md` is a control doc, so the query must opt into `IncludeControlDocs=true`.

### What warrants a second pair of eyes
- **Control-doc contract**: confirm that resolving `tasks.md` via “control doc within ScopeTicket” is the intended long-term API contract for these commands (and not, e.g., “derive from index.md’s directory”).
- **Error behavior**: confirm the “tasks.md not found for ticket” error is appropriate when the file is missing or skipped.

### What should be done in the future
- If we later add more control docs, consider adding a small QueryDocs filter option for “basename/path” to avoid querying-and-filtering in multiple commands.
- After Phase 4 deletes `findTicketDirectory`, ensure no other commands resurrect name-based directory guessing.

### Code review instructions
- Start in `pkg/commands/tasks.go` and review `loadTasksFile` and `findTasksFileViaWorkspace`.

## Step 8: Migrate `search --files` suggestion doc-scan to Workspace+QueryDocs (Phase 3.1)

This step removes the last `search`-side manual filesystem scan in suggestion mode. Both code paths that emit “referenced by documents” file suggestions (`--with-glaze-output` and human output) now source RelatedFiles via `Workspace.QueryDocs` instead of `filepath.Walk` + `readDocumentFrontmatter`.

**Commit (code):** `eadda8d` — "Cleanup: migrate search suggest to QueryDocs"

### What I did
- Updated `pkg/commands/search.go` suggestion mode to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - `QueryDocs(ScopeTicket|ScopeRepo, TopicsAny, IncludeControlDocs=true)` and collect `RelatedFiles` from the indexed docs.
- Removed use of:
  - `filepath.Walk(...)` over markdown files, and
  - ad-hoc `readDocumentFrontmatter(...)` parsing in suggestion mode.
- Preserved the external heuristics (git history, git status, ripgrep/grep).

### Why
- Consolidate doc discovery + metadata parsing behind QueryDocs for consistent semantics and skip rules.
- Reduce duplicated frontmatter parsing and ad-hoc directory traversal patterns.

### What worked
- Tests and lint stayed green.
- Suggestion output continues to include sources (`related_files`, `git_*`, `ripgrep`) with the doc-derived portion now coming from the canonical index.

### What was tricky to build
- **TicketDir for heuristics**: git/ripgrep heuristics still need a working directory; for ticket-scoped suggestions we derive it from the ticket’s `index.md` path (QueryDocs) instead of `findTicketDirectory`.
- **Topic filtering**: topic filtering is now expressed as `DocFilters.TopicsAny` (index-backed), not “parse frontmatter during walk”.

### What warrants a second pair of eyes
- **Scope semantics**: confirm that `--ticket` correctly limits doc-derived suggestions to `ScopeTicket`, while non-ticket uses `ScopeRepo`.
- **Visibility options**: confirm `IncludeControlDocs=true` is correct so suggestions include `RelatedFiles` from ticket root docs (index/tasks/changelog) as legacy walk did.

### What should be done in the future
- If more commands need “derive ticket directory”, consider centralizing the QueryDocs-based derivation into a shared helper (avoid reintroducing `findTicketDirectory`).
- If suggestion semantics drift, treat it as an index/QueryDocs contract issue (options + docs), not a reason to re-add walking/parsing.

### Code review instructions
- Start in `pkg/commands/search.go` and review both suggestion paths:
  - `suggestFiles(...)` (glaze mode)
  - `if settings.Files { ... }` (human mode)

## Step 9: Migrate `doc move` ticket discovery to Workspace+QueryDocs (Phase 3.2)

This step removes `doc move`’s remaining legacy ticket directory resolution (`findTicketDirectory`) for both the source ticket (derived from the document’s frontmatter) and the destination ticket. The command now discovers the Workspace once, builds the index once, and resolves each ticket directory by querying the ticket’s `index.md` via `QueryDocs`.

**Commit (code):** `770e33f` — "Cleanup: migrate doc move to QueryDocs"

### What I did
- Updated `pkg/commands/doc_move.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - resolve `srcTicketDir` and `destTicketDir` via `QueryDocs(ScopeTicket, DocType=index)` instead of `findTicketDirectory`.
- Threaded `context.Context` into `applyMove` so the migration uses the caller’s context (consistent with the rest of the QueryDocs ports).

### Why
- Remove legacy ticket discovery from write-path commands so we can delete `findTicketDirectory` later.
- Ensure “ticket directory resolution” follows the same canonical semantics as other migrated commands.

### What worked
- Tests and lint stayed green.
- The command still preserves relative subpaths under the ticket and enforces “stay within ticket” constraints.

### What was tricky to build
- **Single Workspace lifecycle**: this command needs to resolve two tickets; we avoid doing two separate discoveries/index builds by reusing a single Workspace/index instance.
- **Path conversions**: QueryDocs returns slash-cleaned paths; write-path uses filesystem paths. The implementation normalizes with `filepath.FromSlash` before `filepath.Dir`.

### What warrants a second pair of eyes
- **Index uniqueness assumption**: confirm it’s correct to treat “more than one index doc for a ticket” as a hard error in this write-path command.
- **Root contracts**: confirm that resolving `settings.Root` to `ws.Context().Root` doesn’t break any callers relying on relative-root behavior.

### What should be done in the future
- If more commands need “ticket directory resolution”, consider extracting the QueryDocs-based pattern into a shared helper so we don’t duplicate this logic (without reintroducing legacy heuristics).
- Once Phase 4 removes `findTicketDirectory`, add a regression guard (test or lint pattern) to prevent reintroducing it.

### Code review instructions
- Start in `pkg/commands/doc_move.go` and review `applyMove` and `resolveTicketDirViaWorkspace`.

## Step 10: Migrate `ticket move` ticket discovery to Workspace+QueryDocs (Phase 3.2)

This step migrates `ticket move` off the legacy `findTicketDirectory` helper. The command now discovers the Workspace and builds the index once, then resolves the source ticket directory by querying the ticket’s `index.md` via `QueryDocs` and deriving the directory from that canonical path.

**Commit (code):** `5ce1a88` — "Cleanup: migrate ticket move to QueryDocs"

### What I did
- Updated `pkg/commands/ticket_move.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - resolve `srcDir` via `QueryDocs(ScopeTicket, DocType=index)` instead of `findTicketDirectory`.
- Threaded `context.Context` into `applyMove` to align with the rest of the QueryDocs migrations.

### Why
- Remove remaining `findTicketDirectory` usage in ticket move so we can delete the helper later.
- Make ticket directory resolution consistent with the canonical Workspace index semantics.

### What worked
- Tests and lint stayed green.
- The rename/move semantics and “touch LastUpdated in index.md” behavior are unchanged.

### What was tricky to build
- **Helper duplication**: avoid duplicating helper symbols in the `commands` package by reusing the existing QueryDocs-based ticket-dir resolver.
- **Root resolution**: the command now pins `settings.Root` to `ws.Context().Root` (Workspace resolution) before rendering destination paths.

### What warrants a second pair of eyes
- **Path template rendering**: confirm that using Workspace-resolved root doesn’t subtly change destination paths for callers who pass relative roots.
- **Error behavior**: confirm “ticket not found / ambiguous” is acceptable for this command (write-path).

### What should be done in the future
- Consider centralizing the QueryDocs-based “ticket dir from index doc” pattern into an explicit shared helper file (not buried in another command) once more commands migrate.

### Code review instructions
- Start in `pkg/commands/ticket_move.go` and review `applyMove`.

## Step 11: Migrate `ticket close` ticket discovery to Workspace+QueryDocs (Phase 3.2)

This step removes `ticket close`’s remaining dependency on legacy ticket directory discovery (`findTicketDirectory`). Both the structured output path and human output path now discover the Workspace, build the index, and resolve the ticket directory via `QueryDocs` (ticket `index.md`) before operating on `index.md`, `tasks.md`, and `changelog.md`.

**Commit (code):** `35de822` — "Cleanup: migrate ticket close to QueryDocs"

### What I did
- Updated `pkg/commands/ticket_close.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - resolve `ticketDir` via the shared `resolveTicketDirViaWorkspace(...)` helper (QueryDocs-based).
- Removed the two `findTicketDirectory` call sites (glaze + bare) from this command.

### Why
- Continue eliminating legacy discovery helpers so we can delete them in Phase 4.
- Ensure ticket close semantics follow the same canonical discovery/path normalization rules as other migrated commands.

### What worked
- Tests and lint stayed green.
- Behavior of task-count warning, index frontmatter update, and changelog append remains the same (only discovery changed).

### What was tricky to build
- **Dual code paths**: the command has both Glaze output and human output implementations; both needed the same Workspace-backed ticketDir resolution.
- **Control docs**: `ticket close` interacts with ticket-root docs; ticketDir resolution must match the ticket’s canonical directory as defined by the indexed `index.md`.

### What warrants a second pair of eyes
- **Root resolution**: confirm pinning `settings.Root` to `ws.Context().Root` is correct for all invocation contexts (nested cwd, root overrides).
- **Failure modes**: confirm the “ticket not found / ambiguous” errors are acceptable for this write-path command.

### What should be done in the future
- Consider refactoring the duplicated “discover workspace + init index + resolve ticket dir” boilerplate into a small shared helper (still QueryDocs-backed) used by all write-path ticket commands.

### Code review instructions
- Start in `pkg/commands/ticket_close.go` and review both `RunIntoGlazeProcessor` and `Run` ticket resolution blocks.

## Step 12: Migrate `rename-ticket` discovery to Workspace+QueryDocs (Phase 3.3)

This step removes `rename-ticket`’s remaining dependency on legacy ticket directory discovery (`findTicketDirectory`). The command already uses `documents.WalkDocuments` for the write-path (updating frontmatter across the ticket), so the migration here is purely about resolving the current ticket directory via `Workspace.QueryDocs`.

**Commit (code):** `5ddd75c` — "Cleanup: migrate rename ticket discovery to QueryDocs"

### What I did
- Updated `pkg/commands/rename_ticket.go` (both glaze + human paths) to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - resolve `oldDir` via `resolveTicketDirViaWorkspace(...)` (QueryDocs-based) instead of `findTicketDirectory`.
- Kept the write-path behavior unchanged:
  - `documents.WalkDocuments(oldDir, ...)` updates frontmatter Ticket fields, then
  - `os.Rename(oldDir, newDir)` moves the directory.

### Why
- Continue eliminating `findTicketDirectory` callers to unblock Phase 4 deletion.
- Ensure rename-ticket uses canonical Workspace semantics for “what is the ticket directory?”.

### What worked
- Tests and lint stayed green.
- `--dry-run` remains a pure planning mode (no filesystem changes).

### What was tricky to build
- **Two output modes**: both Glaze and bare modes had their own `findTicketDirectory` call sites that needed migration.
- **Root resolution**: pinning `settings.Root` to `ws.Context().Root` ensures consistent behavior when invoked from nested directories.

### What warrants a second pair of eyes
- **Directory naming invariants**: confirm the “preserve remainder suffix after `<ticket>`” logic still matches all real-world ticket directory names (especially older layouts).
- **Discovery assumptions**: confirm that basing discovery on the ticket’s `index.md` is acceptable for rename operations (and that tickets without index.md are considered invalid for rename).

### What should be done in the future
- If rename flows need to handle partially scaffolded tickets, address it via an explicit contract + tests (not via reintroducing directory heuristics).

### Code review instructions
- Start in `pkg/commands/rename_ticket.go` and review `RunIntoGlazeProcessor` + `Run` ticketDir resolution and the `updateTicketFrontmatter` walk.

## Step 13: Migrate `renumber` discovery to Workspace+QueryDocs (Phase 3.3)

This step removes `renumber`’s remaining dependency on legacy ticket directory discovery (`findTicketDirectory`). The command continues to use `filepath.WalkDir` for the write-path (renaming files and updating intra-ticket references), but the initial ticket workspace resolution is now derived from the Workspace index via `QueryDocs`.

**Commit (code):** `9fd2c8a` — "Cleanup: migrate renumber discovery to QueryDocs"

### What I did
- Updated `pkg/commands/renumber.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - resolve `ticketDir` via `resolveTicketDirViaWorkspace(...)` (QueryDocs-based) instead of `findTicketDirectory`.
- Threaded `context.Context` into `applyRenumber` so it uses the caller’s context.

### Why
- Continue removing legacy discovery helpers across Phase 3 so Phase 4 can delete them.
- Keep discovery semantics consistent with the canonical Workspace index.

### What worked
- Tests and lint stayed green.
- Renumber behavior (prefix resequencing + reference rewriting) is unchanged; only discovery moved.

### What was tricky to build
- **Context threading**: `applyRenumber` is called from both glaze and bare paths; both needed to pass through `ctx`.
- **Write-path separation**: discovery migrates to QueryDocs while the file rename/reference rewrite stays as the existing filesystem walk (intentional per spec).

### What warrants a second pair of eyes
- **TicketDir correctness**: confirm that deriving ticketDir from the indexed `index.md` path is correct for all layouts this command supports.
- **Reference rewrite safety**: confirm the link rewrite still only affects intra-ticket paths and doesn’t accidentally rewrite external links (pre-existing risk, but worth reviewing while touching this code).

### What should be done in the future
- Consider adding a targeted test fixture for renumber (paths + rewritten links) so future refactors don’t regress link rewrite behavior.

### Code review instructions
- Start in `pkg/commands/renumber.go` and review `applyRenumber` (ticketDir resolution) and `updateTicketReferences`.

## Step 14: Migrate `layout-fix` discovery to Workspace+QueryDocs (Phase 3.3)

This step removes `layout-fix`’s remaining legacy discovery logic: a mix of `findTicketDirectory` for `--ticket` and an ad-hoc `os.ReadDir(root)` scan that only worked for flat ticket layouts. Ticket discovery is now driven by the Workspace index via `QueryDocs`, which correctly finds tickets under the date-based path template.

**Commit (code):** `c72e0db` — "Cleanup: migrate layout-fix discovery to QueryDocs"

### What I did
- Updated `pkg/commands/layout_fix.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - resolve ticket directories via QueryDocs:
    - `--ticket`: `resolveTicketDirViaWorkspace(...)`
    - no `--ticket`: `QueryDocs(ScopeRepo, DocType=index)` and derive ticketDir from each `index.md`.
- Threaded `context.Context` into `applyLayoutFix` and updated both glaze + bare callsites.

### Why
- Remove legacy discovery helpers to unblock Phase 4 deletion.
- Fix the “scan all tickets” behavior for modern date-based layouts (QueryDocs is the canonical discovery).

### What worked
- Tests and lint stayed green.
- The write-path behavior (WalkDir + move docs + update intra-ticket references) is unchanged.

### What was tricky to build
- **Ticket enumeration**: the old root scan was layout-sensitive; QueryDocs enumeration now needs deduplication and stable ordering of ticketDirs.
- **Scope/visibility**: we intentionally keep enumeration limited (no archive/scripts) unless explicitly targeted by `--ticket`.

### What warrants a second pair of eyes
- **Enumeration semantics change**: confirm it’s acceptable that `layout-fix` now finds tickets under date-based templates (previously it effectively didn’t).
- **Skip rules**: confirm that excluding archive/scripts in the index query matches the intended contract for this maintenance command.

### What should be done in the future
- If we want to support fixing archived tickets, add an explicit flag/option and document it (rather than broadening defaults silently).

### Code review instructions
- Start in `pkg/commands/layout_fix.go` and review the new ticket directory enumeration logic in `applyLayoutFix`.

## Step 15: Migrate `import file` ticket discovery to Workspace+QueryDocs (Phase 3)

This step migrates `import file` off the legacy `findTicketDirectory` helper (which internally depended on `CollectTicketWorkspaces`). The command now discovers the Workspace, builds the index, and resolves the ticket directory via `QueryDocs` (ticket `index.md`) before writing into `sources/local/` and updating `.meta/sources.yaml` and the ticket `index.md`.

**Commit (code):** `d2b357a` — "Cleanup: migrate import file to QueryDocs"

### What I did
- Updated `pkg/commands/import_file.go` to:
  - `DiscoverWorkspace` + `InitIndex`, then
  - resolve `ticketDir` via `resolveTicketDirViaWorkspace(...)` (QueryDocs-based) instead of `findTicketDirectory`.
- Threaded `context.Context` into `importFile` and updated both glaze + bare callsites.

### Why
- Eliminate another `findTicketDirectory` callsite and remove reliance on legacy workspace collection.
- Keep ticket directory resolution consistent with the canonical Workspace index.

### What worked
- Tests and lint stayed green.
- Behavior of the import itself is unchanged (copy file into `sources/local/`, append `.meta/sources.yaml`, and add `local:<name>` to `index.md` ExternalSources).

### What was tricky to build
- **Context threading**: both Glaze and bare modes now call `importFile(ctx, ...)` so the Workspace lifecycle is consistent with other QueryDocs migrations.
- **TicketDir contract**: importing relies on writing into the ticket directory; this must be derived from the canonical index doc path, not directory heuristics.

### What warrants a second pair of eyes
- **Index update semantics**: confirm adding `local:<destName>` to `ExternalSources` remains the intended contract and doesn’t conflict with the new source metadata stored in `.meta/sources.yaml`.
- **Legacy helper cleanup**: confirm there are no remaining `findTicketDirectory` callsites besides the ones we intend to migrate/delete in Phase 4.

### What should be done in the future
- Once Phase 4 deletes `findTicketDirectory`, consider moving the QueryDocs-based “ticket dir resolver” into an explicit shared helper file to avoid it living incidentally in another command’s file.

### Code review instructions
- Start in `pkg/commands/import_file.go` and review `importFile` ticketDir resolution and the subsequent writes.

## Step 16: Delete legacy `findTicketDirectory` helper (Phase 4.1)

This step deletes the legacy `findTicketDirectory` helper that lived in `pkg/commands/import_file.go` and depended on `CollectTicketWorkspaces`. Before deleting it, we removed the last remaining callsites (notably in `changelog.go`), so there are no compatibility shims left and ticket resolution is now consistently Workspace+QueryDocs-backed.

**Commit (code):** `3751433` — "Cleanup: delete findTicketDirectory helper"

### What I did
- Updated `pkg/commands/changelog.go` to remove the remaining `findTicketDirectory` callsites and instead resolve:
  - `changelog.md` path from the ticket directory via Workspace+QueryDocs, and
  - ticket `searchRoot` for heuristics via the QueryDocs-backed ticket-dir resolver.
- Deleted the `findTicketDirectory(...)` function from `pkg/commands/import_file.go`.

### Why
- This is the prerequisite cleanup that makes it possible to delete the legacy ticket discovery stack (Phase 4).
- Remove the last hidden dependency on `CollectTicketWorkspaces` from the commands layer.

### What worked
- Tests and lint stayed green after deletion.
- No CLI flags or outputs changed; this is a pure internal refactor.

### What was tricky to build
- **Changelog edge paths**: `changelog update` has multiple modes (`--changelog-file`, `--suggest`, human vs glaze). Removing the helper required touching all those resolution points without changing behavior.

### What warrants a second pair of eyes
- **Changelog heuristics searchRoot**: confirm ticket-scoped heuristics still run in the ticket directory (not the repo root) and that failures remain best-effort for suggestions.
- **Deletion safety**: confirm there are no remaining production callsites (only historical docs mention it, which is fine).

### What should be done in the future
- If we see repeated “discover workspace + init index + resolve ticketDir” boilerplate, consider extracting a small shared helper for commands (still QueryDocs-backed) to keep callsites consistent.

### Code review instructions
- Start in `pkg/commands/changelog.go` and verify there are no references to `findTicketDirectory`.
- Start in `pkg/commands/import_file.go` and confirm the helper is deleted and ticketDir is derived via QueryDocs.
