---
Title: Design — Tasks verbs, task metadata, and listing UX
Ticket: DOC
Status: active
Topics:
    - infrastructure
    - backend
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ""
    - Path: ""
    - Path: ""
    - Path: ""
    - Path: ""
ExternalSources: []
Summary: Unified repo-root detection; default stale-after=30; relate --from-git documented; warnings in status; updated docs.
LastUpdated: 2025-11-04T17:56:55.025393803-05:00
---




# Design — Tasks verbs, task metadata, and listing UX

## Executive Summary

We will extend the tasks feature to carry actionable context (owners, related files, external links, notes) and improve both editing verbs and list output. The goals are: faster execution, better reviewability, and durable crosslinking to changelog entries and related docs.

## Problem Statement

- Tasks often lack sufficient context to execute without hunting for files/links.
- Editing verbs handle only the task text; attaching context requires manual edits.
- Listing output is hard to scan and doesn’t surface what’s needed (owners, links, next actions).
- Changelog updates aren’t reflected in tasks, and vice‑versa.

## Proposed Solution

### A. Task metadata model (in tasks.md)
Represent optional metadata inline, staying markdown‑friendly:

- Owners: `@name1,@name2`
- Related: file paths with optional notes: `path (note)`; multiple separated by `;`
- Links: URLs (e.g., PRs/issues/refs)
- Notes: short free‑form hint

Format appended to the task line after an em dash:

- `- [ ] Title — Owners: @alice — Related: pkg/x.go (handler); web/y.ts (slice) — Links: https://… — Notes: quick-win`

Parsing remains tolerant: missing fields are ignored.

### B. New/extended verbs
- `tasks attach --id N --files a,b [--file-note "a:note" --file-note "b=note"]` → merges into Related
- `tasks owners  --id N --owners alice,bob` → sets Owners
- `tasks links   --id N --urls https://a,https://b` → sets Links
- `tasks note    --id N --text "…"` → sets/updates Notes
- `tasks check|uncheck` keep behavior; `tasks edit` continues to edit title text

All verbs accept `--match` as alternative to `--id` (first match semantics), and echo `root/config/vocabulary` before writes.

### C. Listing UX
- Human default: one task per line, showing
  - `[idx] [x| ] Title — Owners: … — Related: … — Links: … — Notes: …`
- Flags:
  - `--with-related` (print short related summary)
  - `--with-owners`, `--with-links`, `--with-notes`
  - `--group-by owners|checked` to group sections
  - `--next` (print only unchecked, owner‑filtered top set)
- Structured: `--with-glaze-output --output table|json|yaml` includes parsed fields with columns: `id,checked,title,owners[],related[],links[],notes`.

### D. Changelog crosslinking
- `tasks link-changelog --id N --entry-id <id>` to append `[#id]` to tasks line and embed backlink in changelog entry body.
- `changelog update --ticket ... --from-task N` to create entry and back‑populate the task with the entry id.

## Design Decisions
- Stay markdown‑native; avoid sidecar YAML to keep diffs/simple edits easy.
- Tolerant parsing to avoid breaking existing task lines.
- Verbs remain idempotent; repeated attaches merge rather than duplicate.
- Human output emphasizes scannability; structured output is first‑class for tooling.

## Alternatives Considered
- Separate YAML frontmatter per task block: more structure, but heavy editing experience.
- JSON blocks per task: machine‑friendly, less ergonomic in reviews.

## Implementation Plan
1) Parser/formatter
   - Extend `parseTasksFromLines` to extract `Owners`, `Related`, `Links`, `Notes` sections.
   - Helpers to merge/serialize metadata back to a single line.
2) Verbs
   - Add `attach`, `owners`, `links`, `note`, `link-changelog` subcommands.
   - Accept `--id` or `--match` consistently; echo context pre‑write.
3) Listing
   - Add flags `--with-related/owners/links/notes`, `--group-by`, `--next`.
   - Structured output adds columns with arrays for related and links.
4) Crosslinking
   - Implement `--from-task` on changelog; teach tasks to embed `[#chg-YYYYMMDD-N]`.
5) Docs
   - Update how‑to and CLI guide; add examples for each verb and listing mode.

## Open Questions
- Should we allow multi‑line tasks (wrapped metadata on next line)?
- Should owners map to people directory for validation?

## References
- Ticket index and plan documents
- Current tasks implementation: `docmgr/pkg/commands/tasks.go`
- Changelog integration: `docmgr/pkg/commands/changelog.go`
