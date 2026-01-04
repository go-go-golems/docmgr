---
Title: Search HTTP API for docmgr
Ticket: 004-SEARCH-API
Status: active
Topics:
    - backend
    - docmgr
    - tooling
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/config.go
      Note: Root resolution (.ttmp.yaml
    - Path: internal/workspace/discovery.go
      Note: Ticket scaffold detection; may need marker updates (design vs design-doc)
    - Path: internal/workspace/index_builder.go
      Note: InitIndex + ingestion; needs caching/locking for server
    - Path: internal/workspace/query_docs.go
      Note: Workspace.QueryDocs public query types; core engine for REST
    - Path: internal/workspace/query_docs_sql.go
      Note: SQL compilation for filters and reverse lookup
    - Path: internal/workspace/workspace.go
      Note: Workspace discovery + context used by both CLI and future server
    - Path: pkg/commands/search.go
      Note: CLI search implementation; primary semantics to mirror in REST
    - Path: ttmp/vocabulary.yaml
      Note: Restored analysis doc-type to allow analysis docs in tickets
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T21:31:55.197633946-05:00
WhatFor: ""
WhenToUse: ""
---


# Search HTTP API for docmgr

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
