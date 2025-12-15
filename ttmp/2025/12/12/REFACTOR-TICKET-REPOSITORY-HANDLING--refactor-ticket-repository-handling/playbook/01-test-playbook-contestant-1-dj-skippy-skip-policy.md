---
Title: 'Test Playbook: Contestant #1 DJ Skippy (Skip Policy)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/skip_policy.go
      Note: Skip predicate + path tag logic under test
    - Path: internal/workspace/skip_policy_test.go
      Note: Unit tests for skip predicate + tagging
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T20:08:04.436369958-05:00
---


# Test Playbook: Contestant #1 DJ Skippy (Skip Policy)

## Purpose

This playbook validates the **canonical ingest-time skip rules** and **path tagging** behavior used by the Workspace SQLite index.

In the refactor, “DJ Skippy” is the personality for:

- directory skip predicate: `DefaultIngestSkipDir`
- path tagging logic for docs rows: `ComputePathTags`

These two routines are foundational because:

- they define what gets indexed (and what is excluded entirely),
- they define tags that later drive query defaults (hide `archive/`, `scripts/`, ticket control docs),
- and they must stay consistent across all future command ports (`list docs`, `search`, `doctor`, etc.).

## Environment Assumptions

- You are in the `docmgr` repo working tree.
- You have Go installed (the tests use `go test`).
- You can run SQLite-backed tests (uses the sqlite3 driver via `database/sql`).

This playbook does not require a real workspace. It uses unit tests and temporary directories.

## Commands

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr && \
go test ./internal/workspace -run 'TestDefaultIngestSkipDir|TestComputePathTags_' -count=1 -v
```

## Exit Criteria

- The command exits 0.
- Output shows all `skip_policy_test.go` tests passing.

Key behavioral guarantees validated by these tests:

- `.meta/` is always skipped.
- `_*/` directories are always skipped.
- `archive/`, `scripts/`, `sources/` are **not** skipped (they are included and later tagged).
- Ticket-root control docs are tagged only when a sibling `index.md` exists:
  - `tasks.md`, `README.md`, `changelog.md` at ticket root => `is_control_doc=1`
  - `design/README.md` without sibling `index.md` => `is_control_doc=0`
- Path segment tagging is boundary-safe:
  - tags trigger on `/archive/`, `/scripts/`, `/sources/` segments (not substrings like `myarchive`).

## Notes

### What this is *actually* testing

The tests live in:

- `internal/workspace/skip_policy_test.go`

and directly call the functions under test:

- `DefaultIngestSkipDir` (directory filter)
- `ComputePathTags` (tag assignment)

### Why this is observable today (even though ingestion is “internal”)

Even if you can’t “watch” the skip policy in a running CLI command yet, these unit tests cover:

- the exact rules as code,
- the tricky cases (control docs require sibling `index.md`),
- and segment boundary correctness.

### Optional integration cross-check (proves the walker actually applies the skip policy)

`skip_policy_test.go` validates the functions in isolation. To prove ingestion applies them,
run Ingrid’s ingestion test too:

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr && \
go test ./internal/workspace -run TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles -count=1 -v
```

That test asserts that docs under `.meta/` and `_guidelines/` do not appear in the DB at all.
