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
      Note: Phase 1.4 migrated suggestion doc-scan to Workspace.QueryDocs (commit 09e1e6f)
    - Path: pkg/commands/meta_update.go
      Note: Phase 2.2 migrated ticket discovery + enumeration to Workspace.QueryDocs (commit 3458a46)
    - Path: pkg/commands/status.go
      Note: Phase 1.1 migration to Workspace.QueryDocs (commit f61606c)
    - Path: pkg/commands/tasks.go
      Note: Phase 2.3 migrated tasks ticket discovery to Workspace.QueryDocs (commit 234f42c)
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
