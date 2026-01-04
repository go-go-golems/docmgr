---
Title: Use SQLite FTS for search; refactor search out of commands
Ticket: 005-USE-SQLITE-FTS
Status: complete
Topics:
    - backend
    - docmgr
    - tooling
    - testing
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../glazed/pkg/help/store/fts5.go
      Note: FTS5 table + triggers behind sqlite_fts5 tag; source of debug log
    - Path: cmd/docmgr/cmds/doc/search.go
      Note: Completion for new --order-by flag
    - Path: internal/searchsvc/search.go
      Note: Reusable search engine used by CLI (and later HTTP)
    - Path: internal/searchsvc/snippet.go
      Note: Moved ExtractSnippet to core
    - Path: internal/searchsvc/suggest_files.go
      Note: Shared file suggestion heuristics
    - Path: internal/workspace/index_builder.go
      Note: |-
        Index ingestion; needs FTS population/rebuild step
        Populate docs_fts during ingest; track FTS availability
    - Path: internal/workspace/query_docs.go
      Note: |-
        DocQuery/DocFilters; add TextQuery + maybe pagination later
        Added DocFilters.TextQuery and OrderByRank
    - Path: internal/workspace/query_docs_fts5_test.go
      Note: FTS-tagged test covering MATCH + OrderByRank
    - Path: internal/workspace/query_docs_sql.go
      Note: |-
        SQL compilation; inject FTS JOIN/MATCH when TextQuery set
        JOIN docs_fts + MATCH + bm25 ordering
    - Path: internal/workspace/sqlite_schema.go
      Note: |-
        Workspace in-memory schema; add docs_fts creation + runtime degraded mode
        Added docs_fts creation helpers and ErrFTSNotAvailable
    - Path: pkg/commands/changelog.go
      Note: Switched git/ripgrep suggestions to internal/searchsvc
    - Path: pkg/commands/relate.go
      Note: Switched git/ripgrep suggestions to internal/searchsvc
    - Path: pkg/commands/search.go
      Note: |-
        Current search engine; will be split so commands become thin adapter
        CLI now thin adapter over internal/searchsvc
    - Path: scenariolog/internal/scenariolog/migrate.go
      Note: Reference implementation for best-effort FTS5 creation and degraded mode
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T17:00:21.028135653-05:00
WhatFor: ""
WhenToUse: ""
---




# Use SQLite FTS for search; refactor search out of commands

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- backend
- docmgr
- tooling
- testing

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
