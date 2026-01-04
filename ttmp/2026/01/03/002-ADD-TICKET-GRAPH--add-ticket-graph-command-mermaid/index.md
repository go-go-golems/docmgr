---
Title: Add ticket graph command (Mermaid)
Ticket: 002-ADD-TICKET-GRAPH
Status: complete
Topics:
    - docmgr
    - cli
    - tooling
    - diagnostics
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: Makefile
      Note: Ensures hooks and local make targets don't use parent go.work
    - Path: cmd/docmgr/cmds/ticket/graph.go
      Note: Cobra wrapper for ticket graph
    - Path: cmd/docmgr/cmds/ticket/ticket.go
      Note: Attach point for new
    - Path: internal/paths/resolver.go
      Note: Required for canonical file node keys + safe display labels
    - Path: internal/workspace/query_docs.go
      Note: Graph command should enumerate docs and expand transitive closure using this API
    - Path: internal/workspace/query_docs_sql.go
      Note: Defines OR semantics for RelatedFile/RelatedDir filters used for expansion
    - Path: pkg/commands/ticket_graph.go
      Note: New ticket graph command implementation
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T21:15:05.559840731-05:00
WhatFor: ""
WhenToUse: ""
---




# Add ticket graph command (Mermaid)

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- cli
- tooling
- diagnostics

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
