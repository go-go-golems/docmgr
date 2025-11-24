---
Title: 'Fix: Template frontmatter parsing with placeholders'
Ticket: DOCMGR-FRONTMATTER
Status: complete
Topics:
    - docmgr
    - infrastructure
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/templates/templates.go
      Note: Fixed ExtractFrontmatterAndBody to handle templates with placeholders by manually stripping frontmatter when library parsing fails
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T15:25:29.510021777-05:00
---



# Fix: Template frontmatter parsing with placeholders

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- infrastructure

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
