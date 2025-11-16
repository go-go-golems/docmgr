---
Title: 'docmgr: scriptable output and path resolution improvements'
Ticket: DOCMGR-UX
Status: active
Topics:
    - tooling
    - ux
    - cli
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/cmd/docmgr/main.go
    - Path: docmgr/pkg/commands/list_docs.go
      Note: Docs columns and filtering
    - Path: docmgr/pkg/commands/list_tickets.go
      Note: Tickets columns and row assembly
    - Path: docmgr/pkg/commands/tasks.go
      Note: Tasks listing and mutations
    - Path: docmgr/pkg/commands/vocab_list.go
      Note: Vocabulary listing and columns
    - Path: glazed/pkg/doc/tutorials/build-first-command.md
ExternalSources:
    - glazed/pkg/doc/tutorials/build-first-command.md
    - go-go-mento/docs/how-to-use-docmgr.md
Summary: Make docmgr scripting friendlier (default-to-index verified)
LastUpdated: 2025-11-05T14:44:55.653782119-05:00
---










# docmgr: scriptable output and path resolution improvements

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- tooling
- ux
- cli

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
