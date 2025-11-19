---
Title: Error Context Rollout
Ticket: DOCMGR-REFACTOR
Status: review
Topics:
    - docmgr
    - error-handling
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/workspace/config.go
      Note: Config parse warnings now propagate real errors
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/list_docs.go
      Note: List docs now wraps AddRow errors with context
ExternalSources: []
Summary: Tracks the plan for wrapping bare errors
LastUpdated: 2025-11-18T21:07:35.792171526-05:00
---



---
Title: Error Context Rollout
Ticket: DOCMGR-REFACTOR
Status: draft
Topics:
  - docmgr
  - error-handling
DocType: design-doc
Intent: long-term
Owners:
  - manuel
RelatedFiles: []
ExternalSources: []
Summary: >
  Tracks the plan for wrapping bare errors
LastUpdated: 2025-11-18
---

# Error Context Rollout

## Executive Summary

Round 4 focuses on making docmgr's CLI errors actionable. We now treat frontmatter reads/writes as first-class helpers (`internal/documents`) and the most frequently used listings (`ticket tickets`, `doc docs`, `list`, `status`, `meta update`, `layout fix`) wrap all propagated errors with context about the command, target ticket, or path involved. Config parsing now surfaces actual YAML failures so users know why `ResolveRoot` fell back to defaults.

## Problem Statement

72 `return err` sites and silent config fallbacks meant users routinely saw messages such as "failed to add row" or "invalid config" with no actionable context. When `.ttmp.yaml` contained a typo docmgr would silently ignore it, making the root directory resolution unpredictable.

## Proposed Solution

Treat each high-traffic command as a user-facing API:

1. Wrap every Glaze row emission with `fmt.Errorf("failed to add … for %s: %w", subject, err)` so scripts immediately know which ticket/doc triggered the failure.
2. Ensure helper functions (`appendSourceMetadata`, layout mover, meta update loops) wrap IO/YAML failures with the absolute path, preventing generic "permission denied" guesses.
3. Update `ResolveRoot` to log parse failures with the actual YAML error while still falling back so users can fix `.ttmp.yaml` without blocking the workflow.
4. Use CLI commands (`doc add`, `meta update`, `doc relate`, `changelog update`, `tasks check`) to validate the helpers end-to-end on live documents.

## Design Decisions

- Focus on commands invoked during every workflow (listings/meta/import/layout) before touching long-tail utilities (doctor/tasks).
- Keep `ResolveRoot` resilient: log warnings to stderr so CI logs capture them, but do not exit to keep docmgr usable even if `.ttmp.yaml` is broken.
- Store newly created docs (`design-doc/01-error-context-rollout.md`, `playbooks/01-cli-regression-checklist.md`) plus test evidence inside the ticket to serve as future regression anchors.

## Alternatives Considered

- Introduce a global `UserError` struct right now. Deferred because wrapping the existing errors delivers immediate value without designing a new error taxonomy.
- Fail hard on malformed configs. Rejected for now; warnings + fallbacks keep existing automation unblocked while surfacing the root cause.

## Implementation Plan

1. Continue wrapping remaining `return err` sites (search, doctor, tasks) with descriptive `fmt.Errorf`.
2. Add thin helpers for repeated `gp.AddRow` patterns once the majority of commands carry context.
3. Expand CLI regression checklist playbook so future contributors can replay the doc creation + metadata flow quickly.

## Open Questions

- Should we eventually add structured error codes for scripting? (Left for later discussion once wrappers land.)
- How aggressively should we warn about missing `docmgr relate/changelog` steps (maybe via `tasks` reminders)?

## References

- Debate Round 04 — Error Handling and User Experience
- Working note: Frontmatter Walkthrough (this ticket)
