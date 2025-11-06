---
Title: Filename numbering policy for newly scaffolded docs
Ticket: DOCMGR-101-prefix-scaffolded-files
Status: active
Topics:
    - infrastructure
    - tools
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-06T12:12:19.424228977-05:00
---


# Filename numbering policy for newly scaffolded docs

## Executive Summary

Apply 2‑digit numeric filename prefixes (01-, 02-, …) to all newly scaffolded documents across all ticket subdirectories to enforce a readable, deterministic order. If a folder exceeds 99 items, use 3 digits for new files. No configuration or overrides.

## Problem Statement

Ticket directories accumulate many files and natural sort by slug makes navigation noisy. We lack a first-class ordering mechanism for newly created docs without manual renames.

## Proposed Solution

- Prefix at scaffold time (create-ticket/add) for every subdirectory.
- Compute next integer by scanning sibling files; left‑pad to 2 digits (or 3 if count > 99).
- Do not rename existing files automatically; prefixes apply only to newly scaffolded docs.
- Provide a separate verb `renumber` to re‑sequence prefixes and update intra‑ticket references.

## Design Decisions

- Prefix width: 2 digits; use 3 digits only once >99 files exist in a folder.
- Scope: all subdirectories under a ticket (design, reference, playbooks, scripts, various, archive, etc.).
- Collision handling: if target exists, increment until free; fail after N attempts.
- Stability: never rewrite prefixes on existing files automatically; use `renumber` when needed.
- Determinism: order defined by creation sequence, not title.

## Alternatives Considered

- Alpha-only ordering: insufficient for dense folders.
- Date prefixes: leak chronology over intent; less predictable when backfilling docs.
- Central index-based ordering: requires manual maintenance.

## Implementation Plan

1) Scaffolding: compute next prefix per folder and apply to filename slug (2‑digit; 3‑digit if >99).
2) Doctor: warn on missing prefix only.
3) Add `renumber` verb: resequence prefixes in a ticket and update references (paths/links) within that ticket.
4) Docs: update help and guidelines; add playbook tests.

## Open Questions

- Should we support retroactive renumbering? (default no)
- How to handle >99 files in a folder with width=2? (auto-upgrade width?)
- Apply to `various/` by default?

## References

- `docmgr/cmd/docmgr/main.go` — CLI entrypoint
- Ticket index (RelatedFiles) for this ticket
