---
Title: Cleanup inspectors brief (Task 18)
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/discovery.go
      Note: Legacy ticket discovery helpers still used by some commands
    - Path: internal/workspace/query_docs.go
      Note: Canonical doc enumeration API for replacements
    - Path: pkg/commands/import_file.go
      Note: Defines legacy findTicketDirectory helper (prime cleanup target)
    - Path: pkg/commands/status.go
      Note: Uses CollectTicketWorkspaces + Walk; candidate for Workspace migration
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T10:12:05.519875905-05:00
---


# Cleaning Inspectors Brief — Task 18 (“Cleanup duplicated walkers/helpers”)

## Mission

You are a “cleaning inspectors” crew. Your job is **not** to implement the cleanup yet. Your job is to produce a **high-signal report** that tells implementers exactly what needs attention to complete Task 18:

> Cleanup: remove/retire duplicated walkers and helpers; enforce single canonical traversal via Workspace (Spec §11.3).

The report should make it hard to miss anything, and easy to execute in small, safe PRs.

## Scope (what you’re responsible for)

You are looking for:
- **Duplicate discovery logic** (ticket/doc discovery, root/config detection, doc walking)
- **Duplicate filtering/skip rules** (what files/dirs are skipped; control-doc handling; archive/scripts/sources)
- **Duplicate frontmatter parsing** paths (manual read/parse vs using the centralized ingestion/index)
- **Residual “findTicketDirectory” / “Walk” style implementations** that should be routed through `workspace.Workspace`

You are *not* required to:
- Change user-facing behavior (that comes later, in implementation PRs)
- Redesign the Workspace API (assume the current Workspace+SQLite shape is the direction)

## Constraints (repo rules you must respect)
- **Don’t remove debug logging code.**
- **Never remove lines marked `XXX`.**
- Prefer incremental changes: one command at a time, with tests.
- Keep the “system docmgr” workflow (diary/changelog/relate updates during actual implementation).

## Desired output (what you must deliver)

Produce a document (this one) containing:

1. **Inventory table** of cleanup targets (see template below)
2. **Cleanup guidelines**: “when you see X, replace it with Y”
3. **PR plan**: recommended sequencing (safe/independent chunks first)
4. **Risk notes**: which refactors are likely to be behavior-sensitive

### Inventory table template (required)

For each item you find, record:
- **Location**: file + function/symbol
- **Category**: discovery | traversal | filtering/skip | parsing | normalization | output grouping
- **Current behavior**: what it does today (1–3 sentences)
- **Proposed canonical replacement**:
  - Workspace discovery: `workspace.DiscoverWorkspace(...)`
  - Doc listing/search: `ws.InitIndex(...)` + `ws.QueryDocs(...)`
  - Path normalization: `ws.Resolver()` or doc-anchored `paths.NewResolver(ResolverOptions{DocPath: ...})`
- **Migration note**: expected behavior change risks + how to validate
- **Action**: delete | wrap | move | keep (with justification)

## What to look for (search patterns)

Start with these patterns (they’re known hotspots):
- `findTicketDirectory(` (legacy ticket dir discovery helper; often re-implements semantics)
- `workspace.CollectTicketWorkspaces(` / `CollectTicketScaffoldsWithoutIndex(` (legacy ticket scanning)
- `filepath.Walk(` / `filepath.WalkDir(` (ad-hoc traversal)
- `documents.WalkDocuments(` (canonical doc walker; should usually be used via Workspace, not re-implemented)
- `documents.ReadDocumentWithFrontmatter` / `readDocumentFrontmatter` (manual parse paths)

Also watch for:
- Any command re-implementing skip rules (e.g., `_*/`, `.meta/`, control docs)
- Any command doing path canonicalization without using `paths.Resolver`

## Where to look first (starting map)

### High-priority commands (likely cleanup targets)

These currently contain `findTicketDirectory` and/or ad-hoc `filepath.Walk*` logic and should be reviewed for migration to Workspace-based discovery:
- `pkg/commands/search.go` (still uses `findTicketDirectory` + `filepath.Walk` in suggestion mode)
- `pkg/commands/status.go` (uses `CollectTicketWorkspaces` and walks ticket dirs)
- `pkg/commands/list_tickets.go` / `pkg/commands/list.go` (ticket discovery via `CollectTicketWorkspaces`)
- `pkg/commands/changelog.go` (ticket-dir discovery and filesystem search root selection)
- `pkg/commands/tasks.go` (ticket-dir helper usage)
- `pkg/commands/add.go` (ticket-dir discovery on write path)
- `pkg/commands/meta_update.go` (walks and updates frontmatter; likely migratable to QueryDocs-based doc sets)
- `pkg/commands/renumber.go` / `pkg/commands/layout_fix.go` (WalkDir over ticket trees)
- `pkg/commands/doc_move.go`, `pkg/commands/ticket_move.go`, `pkg/commands/ticket_close.go`, `pkg/commands/rename_ticket.go` (ticket-dir discovery; some operations are inherently filesystem-level, but discovery should be canonical)

### Core legacy “discovery” helpers
- `pkg/commands/import_file.go` defines `findTicketDirectory(...)` (this is the main legacy helper to inventory and potentially retire)
- `internal/workspace/discovery.go` (`CollectTicketWorkspaces`, `CollectTicketScaffoldsWithoutIndex`) (still used by some commands; likely should be replaced by Workspace index-backed enumeration and/or a Workspace-level helper)

### Canonical sources (what “good” looks like)
- `internal/workspace/workspace.go` (Workspace context + resolver)
- `internal/workspace/index_builder.go` (ingest contract + skip policy integration)
- `internal/workspace/query_docs.go` / `query_docs_sql.go` (single “doc set” truth)
- `internal/workspace/skip_policy.go` (canonical skip/tag semantics)
- `internal/paths/resolver.go` (normalization contract)

## Cleanup guidelines (draft rules for implementers)

When you see…
- **Ticket discovery via `findTicketDirectory`**
  - Prefer: `ws.QueryDocs(ScopeTicket, DocType=index)` + selecting `index.md` where you need the ticket root doc.
  - If you truly need the filesystem ticket directory, infer from the index doc’s path instead of re-scanning.

- **Doc enumeration via filesystem walking**
  - Prefer: `ws.InitIndex(...)` + `ws.QueryDocs(...)` for doc sets.
  - Only keep filesystem walks when the command fundamentally operates on non-doc files, or needs raw directory structure operations.

- **Manual skip rules**
  - Prefer: Workspace ingest skip + path tags (archive/scripts/sources/control docs).
  - If a command has extra ignore semantics (e.g., `doctor --ignore-glob`), treat them as *post-filters* on QueryDocs results.

- **Manual parsing of frontmatter**
  - Prefer: the index’s `parse_ok` / `parse_err` fields surfaced through QueryDocs.
  - Only re-parse when you need line/col/snippet diagnostics for output.

## How to validate proposed cleanups (test guidance)

For each proposed refactor, suggest a concrete validation:
- Unit tests: `go test ./...`
- Scenario suite (when behavior-sensitive): `test-scenarios/testing-doc-manager/run-all.sh`
- CLI smoke tests: run the specific command in both “human” and “glaze” modes if applicable.

## Deliverable checklist (inspectors sign-off)
- [ ] Inventory table filled for all matches of the search patterns above (or justified exclusions)
- [ ] Clear mapping from each duplicate to its Workspace-based replacement
- [ ] Prioritized PR plan (what to do first, what to defer)
- [ ] Risks called out explicitly (behavior changes, performance, compatibility)

