---
Title: Add docmgr Web UI
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - tooling
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/api/serve.go
      Note: Serve API + UI from one process
    - Path: internal/httpapi/server.go
      Note: HTTP API handlers consumed by the UI
    - Path: internal/paths/resolver.go
      Note: Path normalization and fuzzy matching
    - Path: internal/web/spa.go
      Note: SPA fallback handler (serve UI from Go)
    - Path: internal/workspace/index_builder.go
      Note: Index ingestion (docs/topics/related_files)
    - Path: internal/workspace/query_docs.go
      Note: Workspace.QueryDocs API used by search
    - Path: internal/workspace/query_docs_sql.go
      Note: Reverse lookup SQL for --file/--dir
    - Path: pkg/commands/search.go
      Note: Doc search command implementation (flags
    - Path: ui/src/features/search/SearchPage.tsx
      Note: Main Search UI page (modes/filters/pagination/preview)
    - Path: ui/src/features/ticket/TicketPage.tsx
      Note: Crash fix for sec.items null in tasks response
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T13:38:20.372344818-05:00
WhatFor: ""
WhenToUse: ""
---




# Add docmgr Web UI

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Design**: `design/01-design-docmgr-search-web-ui.md`

## Status

Current status: **active**

## Topics

- docmgr
- ux
- cli
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
