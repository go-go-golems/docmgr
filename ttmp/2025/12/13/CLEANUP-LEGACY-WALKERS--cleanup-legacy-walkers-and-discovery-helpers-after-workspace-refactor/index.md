---
Title: Cleanup legacy walkers and discovery helpers after Workspace refactor
Ticket: CLEANUP-LEGACY-WALKERS
Status: active
Topics:
    - refactor
    - tickets
    - docmgr-internals
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/query_docs.go
      Note: Canonical QueryDocs implementation
    - Path: pkg/commands/import_file.go
      Note: Contains findTicketDirectory definition (prime target)
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/09-cleanup-inventory-report-task-18.md
      Note: Detailed inventory of all cleanup targets
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Design spec for the Workspace API
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T10:29:17.671599467-05:00
---


# Cleanup legacy walkers and discovery helpers after Workspace refactor

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- refactor
- tickets
- docmgr-internals

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
