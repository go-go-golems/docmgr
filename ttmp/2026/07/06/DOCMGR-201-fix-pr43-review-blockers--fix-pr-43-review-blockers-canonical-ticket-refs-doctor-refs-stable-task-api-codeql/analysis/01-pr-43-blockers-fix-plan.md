---
Title: PR 43 blockers fix plan
Ticket: DOCMGR-201-fix-pr43-review-blockers
Status: active
Topics:
    - docmgr
    - cli
    - ux
    - tooling
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://pkg/commands/add.go
      Note: Fix doc add to persist canonical ticket IDs after forgiving resolution.
    - Path: repo://pkg/commands/doctor.go
      Note: Fix doctor --ticket to resolve forgiving refs before exact SQL filtering.
    - Path: repo://internal/httpapi/tickets.go
      Note: Fix task check API to accept stable refs.
    - Path: repo://internal/tasksmd/tasksmd.go
      Note: Add shared stable-ref task toggle helper for HTTP API.
    - Path: repo://ui/src/features/ticket/tabs/TicketTasksTab.tsx
      Note: Send and display stable task refs in the UI.
    - Path: repo://internal/paths/resolver.go
      Note: Triage CodeQL alert around path comparison keys.
ExternalSources:
    - https://github.com/go-go-golems/docmgr/pull/43
Summary: Implementation plan for DOCMGR-201, converting the PR 43 review blockers into concrete code changes, regression tests, and validation commands.
LastUpdated: 2026-07-06T11:40:00-04:00
WhatFor: "Guide implementation and review of the PR 43 blocker fixes."
WhenToUse: "Use before editing code for DOCMGR-201 or when checking whether the blocker fixes are complete."
---

# PR 43 blockers fix plan

## Executive summary

DOCMGR-200's PR 43 review found that the branch is directionally strong but should not merge while three correctness blockers and one merge-status blocker remain. DOCMGR-201 fixes those blockers without reopening the whole DOCMGR-200 design.

The fixes should be small and test-backed:

1. `doc add` must persist the canonical resolved ticket ID, not the user's short forgiving ref.
2. `doctor --ticket` must resolve forgiving refs before building an exact `ScopeTicket` query and must fail loudly if a requested ticket checks zero docs.
3. HTTP/UI task toggles must use stable task refs end-to-end, with positional integer IDs kept only as a compatibility fallback.
4. The Advanced Security CodeQL alert at `internal/paths/resolver.go:350` must be fixed, suppressed with justification, or dismissed in GitHub after trace review.

## Current evidence

The DOCMGR-200 review report is the evidence source:

```text
ttmp/2026/07/05/DOCMGR-200-.../analysis/02-pr-43-code-review-and-project-review.md
```

Reproduction harness:

```text
ttmp/2026/07/05/DOCMGR-200-.../scripts/04-pr43-review-experiments.sh
ttmp/2026/07/05/DOCMGR-200-.../sources/pr43-review-experiments.txt
```

## Fix 1: canonical ticket IDs in `doc add`

### Problem

`pkg/commands/add.go` resolves the ticket directory via the forgiving resolver, but `models.Document{Ticket: settings.Ticket}` persists the raw user input. A short ref such as `CANON-1` can create a doc under `CANON-1-long-canonical` with `Ticket: CANON-1`, which ticket-scoped queries then exclude.

### Proposed change

Return the full `tickets.Resolution` (or at least canonical ID + dir + root) from `findTicketDirectoryViaWorkspace`, then use `res.TicketID` in the new document and result rows.

### Regression test

Create ticket `CANON-1-long-canonical`, run `doc add --ticket CANON-1`, assert:

- new document frontmatter has `Ticket: CANON-1-long-canonical`;
- `ticket show CANON-1` lists the new document.

## Fix 2: forgiving refs in `doctor --ticket`

### Problem

`doctor --ticket` currently builds `workspace.Scope{Kind: ScopeTicket, TicketID: settings.Ticket}`. SQL compiles that to exact `d.ticket_id = ?`, so `doctor --ticket DOCMGR-200` can print `No tickets checked.` even though `ticket show DOCMGR-200` resolves correctly.

### Proposed change

After workspace discovery/index initialization and before query construction, resolve `settings.Ticket` through `tickets.Resolve`. Use `res.TicketID` in `ScopeTicket`.

If a requested ticket produces zero grouped buckets after query/filtering, return an error instead of rendering `No tickets checked.`.

### Regression test

Create `CANON-2-long-canonical`; run doctor with `--ticket CANON-2`; assert it checks the canonical ticket and emits either `ok` or findings, not zero tickets.

## Fix 3: stable task refs through HTTP and UI

### Problem

The API returns `stableId`, but `POST /api/v1/tickets/tasks/check` accepts only `ids: []int` and calls `tasksmd.ToggleChecked`. The UI sends `it.id` and displays numeric IDs.

### Proposed API shape

```json
{
  "ticket": "DOCMGR-201-fix-pr43-review-blockers",
  "refs": ["4y99"],
  "ids": [1],
  "checked": true
}
```

Rules:

- prefer `refs` when supplied;
- keep `ids` as a legacy fallback by converting to decimal strings;
- resolve stable IDs first, then positional refs;
- error with a task table when refs are unknown.

### Code changes

- Add `tasksmd.ToggleCheckedByRefs(lines []string, refs []string, checked bool)`.
- Change HTTP request struct to include `Refs []string` and legacy `IDs []int`.
- Change UI service type to `refs: string[]` (or support both during transition).
- Send `it.stableId ?? String(it.id)` from UI and display stable IDs when present.

### Regression test

HTTP API test should:

1. add/migrate a task with a stable marker;
2. `GET /tickets/tasks` and capture `stableId`;
3. `POST /tickets/tasks/check` with `refs: [stableId]`;
4. assert the task toggles;
5. assert legacy `ids: [1]` still works.

## Fix 4: CodeQL alert

### Current state

GitHub Advanced Security reports:

```text
internal/paths/resolver.go:350
Uncontrolled data used in path expression
```

The line is `NormalizedPath.matchKeys()`, which appears to construct comparison strings, not open files. Treat this as unresolved until the GitHub CodeQL trace is inspected.

### Proposed resolution path

1. Inspect the CodeQL trace.
2. If it reaches filesystem I/O, add validation before the sink.
3. If it is comparison-only, add a narrow suppression/comment according to the repo's CodeQL policy or dismiss the alert in GitHub with a false-positive explanation.
4. Document the decision in the diary.

## Validation checklist

```bash
go test ./... -count=1
go test -tags sqlite_fts5 ./... -count=1
(cd ui && pnpm build && pnpm lint)
gh pr checks 43
```

Also rerun the relevant parts of:

```bash
ttmp/2026/07/05/DOCMGR-200-.../scripts/04-pr43-review-experiments.sh
```
