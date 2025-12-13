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
    - Path: pkg/commands/add.go
      Note: Phase 2.1 migrated add ticket discovery to Workspace.QueryDocs
    - Path: pkg/commands/changelog.go
      Note: Phase 1.4 migrated suggestion mode to Workspace.QueryDocs
    - Path: pkg/commands/doc_move.go
      Note: Phase 3.2 migrated doc move ticket discovery to Workspace.QueryDocs (commit 770e33f)
    - Path: pkg/commands/import_file.go
      Note: Contains findTicketDirectory definition (prime target)
    - Path: pkg/commands/list.go
      Note: Phase 1.3 migrated to Workspace.QueryDocs
    - Path: pkg/commands/list_tickets.go
      Note: Phase 1.2 migrated to Workspace.QueryDocs
    - Path: pkg/commands/meta_update.go
      Note: Phase 2.2 migrated to Workspace.QueryDocs
    - Path: pkg/commands/search.go
      Note: Phase 3.1 migrated search --files suggestion path to Workspace.QueryDocs (commit eadda8d)
    - Path: pkg/commands/status.go
      Note: Phase 1.1 migrated to Workspace.QueryDocs
    - Path: pkg/commands/tasks.go
      Note: Phase 2.3 migrated to Workspace.QueryDocs
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/09-cleanup-inventory-report-task-18.md
      Note: Detailed inventory of all cleanup targets
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Design spec for the Workspace API
    - Path: ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/analysis/02-phase-2-integration-suite-failure-analysis-docmgr-path-nonexistent.md
      Note: Phase 2 validation runner failure analysis (DOCMGR_PATH pointed to missing binary)
    - Path: ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/design/01-cleanup-overview-and-migration-guide.md
      Note: Defines no-backwards-compat policy
    - Path: ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/playbook/01-intern-code-review-verification-questionnaire.md
      Note: Intern/reviewer verification questionnaire with experiments
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
