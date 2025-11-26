---
Title: Normalize docmgr file search paths
Ticket: DOCMGR-PATH-NORMALIZE
Status: active
Topics:
    - docmgr
    - search
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: docmgr/pkg/commands/relate.go
      Note: Canonicalizes --file-note/--remove-files entries and suggestion sources
    - Path: docmgr/pkg/commands/search.go
      Note: Reverse lookup now uses normalized comparisons + suffix-based fuzzy matching
    - Path: docmgr/internal/workspace/config.go
      Note: Provides doc root + config resolution leveraged by the path resolver
ExternalSources: []
Summary: >
    Baselines docmgr’s previous path handling, implements canonical path
    normalization + fuzzy search for relate/search verbs, and backstops the
    behavior with unit tests plus a scenario playbook.
LastUpdated: 2025-11-26T18:40:00-05:00
---

# Normalize docmgr file search paths

## Overview

- Created `reference/01-path-handling-analysis.md` to capture the original behavior.
- Drafted `design/01-path-normalization-canonicalization.md` covering the resolver,
  canonical storage rules, and search updates.
- Implemented the resolver + relate/search wiring, added Go regression tests,
  and documented a `14-path-normalization.sh` playbook to exercise the new UX.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active** — implementation + regression coverage complete; ready for review.

## Topics

- docmgr
- search

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
