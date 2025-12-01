---
Title: doc list --ticket fails in multi-repo setup and silently skips broken documents
Ticket: DOCMGR-BUG-001
Status: active
Topics:
    - bug
    - multi-repo
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: 'doc list command has two issues: 1) may fail in multi-repo setups due to root resolution, 2) silently skips documents with broken frontmatter without warnings'
LastUpdated: 2025-12-01T15:45:32.21490823-05:00
---


# doc list --ticket fails in multi-repo setup and silently skips broken documents

## Overview

Two related issues with `docmgr doc list`:

1. **Multi-repo root resolution**: May fail to find tickets when run from workspace root in multi-repo setups
2. **Silent skipping**: Documents with broken frontmatter are silently skipped without warnings, making it appear as if no documents exist

See the bug report in `reference/01-bug-report-doc-list-ticket-fails-in-multi-repo-setup.md` for detailed analysis and proposed fixes.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- bug
- multi-repo

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
