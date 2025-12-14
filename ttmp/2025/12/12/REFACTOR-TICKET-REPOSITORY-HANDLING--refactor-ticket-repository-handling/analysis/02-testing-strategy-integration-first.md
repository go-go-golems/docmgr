---
Title: Testing strategy (integration-first)
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: test-scenarios/testing-doc-manager/05-search-scenarios.sh
      Note: Covers search + reverse lookup semantics; must keep consistent through refactor.
    - Path: test-scenarios/testing-doc-manager/14-path-normalization.sh
      Note: Existing integration coverage for wonky path forms (doc/ttmp/abs/basename).
    - Path: test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh
      Note: Existing integration coverage for taxonomy/diagnostics behaviors.
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Primary integration harness we will run against baseline vs refactor.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Spec sections for QueryDocs + skip rules + diagnostics the tests should lock down.
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T17:41:45.961181169-05:00
---


# Testing strategy (integration-first)

## Goal

Make the refactor safe by locking down **user-visible behavior** while we replace the lookup implementation with `internal/workspace.Workspace` backed by an **in-memory SQLite index**.

## Why integration tests first (for this refactor)

This refactor is mostly about **behavior + consistency**, not “clever algorithms”:
- discovery + root resolution
- skip rules (what’s included/excluded and what is “hidden by default”)
- reverse lookup + path normalization
- diagnostics behavior for broken/partial states

Those are best tested by driving the real CLI against a realistic workspace, because unit tests tend to:
- miss command wiring / flags / defaults,
- miss interactions between config/root + walk + normalization,
- and drift as we move logic between packages.

## What we already have (use it as the baseline)

There is already an end-to-end harness at:
- `test-scenarios/testing-doc-manager/`

It includes scripts that are directly relevant to this refactor:
- `05-search-scenarios.sh`: search + reverse lookup expectations (file/dir) and “wonky paths”
- `14-path-normalization.sh`: relates the same file using doc-relative, ttmp-relative, and absolute; then verifies `doc search --file` matches regardless of path form
- `15-diagnostics-smoke.sh`: exercises taxonomy output across doctor/list/template/frontmatter parsing

**Takeaway:** we should extend these scenarios incrementally as we port commands to `Workspace.QueryDocs`.

## When to start testing (timing)

Start **now**, but stage it so we don’t waste effort:

1) **Immediately (today): keep the existing scenarios running as a regression baseline**
   - Run `test-scenarios/testing-doc-manager/run-all.sh` against:
     - the “current” released/system docmgr (baseline), and later
     - the refactor binary (to compare behavior).

2) **As soon as Task [3–7] land (schema + ingest + QueryDocs compiler), add/extend scenarios**
   - The first time we port *one* command (likely `doc search` or `list docs`) to `Workspace.QueryDocs`,
     we should add scenario assertions that cover:
     - default skip rules (§6) and default visibility toggles (§5.2 options)
     - invalid frontmatter handling (`parse_ok=0` excluded by default, but diagnostics emitted) (§7)
     - reverse lookup results consistent with current behavior (§10 examples)

3) **Later (after we stabilize semantics), add small unit tests**
   - Only for the pieces that are easiest to pin down deterministically:
     - SQL compilation (DocQuery → SQL fragments + args)
     - skip-rule tagging helpers (path → tags)
     - edge-case path normalization behavior (but a lot is already covered by `14-path-normalization.sh`)

## Proposed integration test strategy for this refactor

### A) Baseline-first comparison

Run the existing scenario suite with:
- **system docmgr** (baseline)
- **refactor docmgr** (built from this repo)

The goal is not byte-for-byte identical output, but:
- identical exit codes for the same mode (`--fail-on`, etc.)
- same sets of returned docs for key queries
- same presence/absence of key diagnostics categories

### B) Add “QueryDocs smoke” coverage once QueryDocs exists

Add one new scenario script (or extend `05-search-scenarios.sh`) to cover:
- “ticket scope + status/docType filters” (basic filtering)
- “repo scope + RelatedFile[] / RelatedDir[]” (reverse lookup joins)
- “include/exclude control docs/scripts/archive” (visibility defaults)

### C) Add “broken states” scenarios

Use (or extend) `15-diagnostics-smoke.sh` patterns to validate:
- invalid YAML frontmatter doc is indexed but excluded unless `IncludeErrors=true`
- diagnostics emitted for:
  - parse failures
  - missing related files
  - normalization fallbacks

## Practical decision

For this refactor, integration tests should be the **primary safety net**. Unit tests are optional and should be added only where they reduce iteration time or prevent accidental query/SQL regressions.

