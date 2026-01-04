---
Title: Add Claude Skills Support
Ticket: 001-ADD-CLAUDE-SKILLS
Status: archived
Topics:
    - features
    - skills
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/doc/doc.go
      Note: Document command structure
    - Path: cmd/docmgr/cmds/root.go
      Note: Command registration pattern
    - Path: cmd/docmgr/cmds/vocab/vocab.go
      Note: Vocabulary command structure
    - Path: internal/documents/frontmatter.go
      Note: Frontmatter parsing
    - Path: internal/documents/walk.go
      Note: Document walking utilities
    - Path: internal/workspace/discovery.go
      Note: Workspace discovery
    - Path: internal/workspace/index_builder.go
      Note: Index ingest to persist WhatFor/WhenToUse
    - Path: internal/workspace/query_docs.go
      Note: Key constraint for skills fields (QueryDocs hydrates from SQLite)
    - Path: internal/workspace/query_docs_sql.go
      Note: Query compilation updates for new columns
    - Path: internal/workspace/sqlite_schema.go
      Note: Schema changes required for skill fields
    - Path: pkg/commands/search.go
      Note: Path filtering semantics to reuse for skill list
    - Path: pkg/commands/vocab_list.go
      Note: List command implementation pattern
    - Path: pkg/doc/docmgr-codebase-architecture.md
      Note: Enhanced architecture documentation with detailed explanations
    - Path: pkg/doc/docmgr-how-to-add-cli-verbs.md
      Note: Created CLI verb implementation guide with step-by-step instructions
    - Path: pkg/models/document.go
      Note: Document model definition
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T21:12:49.73755553-05:00
WhatFor: ""
WhenToUse: ""
---





# Add Claude Skills Support

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- features
- skills

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
