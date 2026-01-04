---
Title: Add better usage examples to docmgr command help
Ticket: 003-BETTER-EXAMPLES
Status: archived
Topics:
    - docmgr
    - cli
    - docs
    - ux
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/doc/doc.go
      Note: Add examples to doc group help (commit 8ec1c61)
    - Path: cmd/docmgr/cmds/ticket/ticket.go
      Note: Add examples to ticket group help (commit 8ec1c61)
    - Path: cmd/docmgr/cmds/workspace/workspace.go
      Note: Add examples to workspace group help (commit 8ec1c61)
    - Path: pkg/commands/add.go
      Note: Update doc add help examples
    - Path: pkg/commands/create_ticket.go
      Note: Update create-ticket help examples + README template
    - Path: pkg/commands/relate.go
      Note: Update doc relate help examples (remove suggest
    - Path: pkg/commands/search.go
      Note: Add real example usage for search (commit 8ec1c61)
    - Path: pkg/commands/tasks.go
      Note: Add examples to task subcommands (commit 8ec1c61)
    - Path: pkg/commands/ticket_move.go
      Note: Add examples and clarify path-template usage (commit 8ec1c61)
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T21:15:10.36053916-05:00
WhatFor: ""
WhenToUse: ""
---




# Add better usage examples to docmgr command help

## Overview

Add 2 copy/paste-ready, realistic CLI examples to the long help text of every `docmgr` command, and validate each example by actually running it.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- cli
- docs
- ux

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
