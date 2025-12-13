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

### Code review instructions
- Start in `pkg/commands/list.go`.
- Smoke:
  - `docmgr list`
  - `docmgr list --ticket CLEANUP-LEGACY-WALKERS`
