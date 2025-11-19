---
Title: 'Code Refactoring: Hierarchical Commands and Utilities'
Ticket: DOCMGR-REFACTOR
Status: active
Topics:
    - docmgr
    - refactoring
    - architecture
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/01-debate-round-01-architecture-and-code-organization.md
      Note: Hierarchical structure decision
    - Path: ../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/02-debate-round-02-command-implementation-patterns.md
      Note: Duplication analysis and utility extraction
    - Path: ../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/04-debate-round-04-error-handling-and-user-experience.md
      Note: Error handling improvements
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/cmd/docmgr/cmds/root.go
      Note: |-
        Hierarchical Cobra command tree
        Hierarchical commands wiring
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/documents/frontmatter.go
      Note: Frontmatter helpers moved to internal
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/documents/walk.go
      Note: Document walker utility
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/templates/templates.go
      Note: Template scaffolding now internal
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/workspace/config.go
      Note: Config helpers moved to internal/workspace
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/doctor.go
      Note: Doctor command rows now emit contextual errors
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/meta_update.go
      Note: Now uses internal/documents
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/search.go
      Note: Search command suggestion/errors now include context
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/tasks.go
      Note: Wrapped tasks commands errors with context
ExternalSources: []
Summary: Round 2 document utilities shipped; CLI verified with new docs while kicking off Round 4 error-handling improvements.
LastUpdated: 2025-11-18T20:50:06.490255749-05:00
---








# Code Refactoring: Hierarchical Commands and Utilities

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- refactoring
- architecture

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
