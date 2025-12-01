---
Title: Streamline templates and guidelines - move to embedded scaffolding
Ticket: DOCMGR-STREAMLINE-TEMPLATES
Status: active
Topics:
    - docmgr
    - templates
    - guidelines
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/verb-templates
      Note: Moved verb templates from ttmp/templates/ to codebase examples
    - Path: internal/templates/embedded.go
      Note: New embedded FS loading for templates/guidelines (scaffolding only)
    - Path: internal/templates/templates.go
      Note: Removed legacy string map fallback from runtime resolution
    - Path: pkg/commands/add.go
      Note: Updated to use filesystem-only template loading
    - Path: pkg/commands/guidelines_cmd.go
      Note: Updated to use filesystem-only guideline loading
    - Path: pkg/commands/scaffold.go
      Note: Updated to use embedded templates/guidelines for init scaffolding
    - Path: pkg/doc/templates-and-guidelines.md
      Note: Comprehensive rewrite following glazed style guide
    - Path: pkg/doc/verb-templates-and-schema.md
      Note: New comprehensive guide for verb templates and schema introspection
    - Path: ttmp/2025/12/01/DOCMGR-STREAMLINE-TEMPLATES-streamline-templates-and-guidelines-move-to-embedded-scaffolding/analysis/01-template-analysis-useful-vs-slop.md
      Note: Comprehensive analysis identifying useful vs slop templates
ExternalSources: []
Summary: 'Streamlined templates and guidelines system: moved to embedded FS for scaffolding only, removed legacy fallbacks, moved verb templates to examples/, and created comprehensive documentation following glazed style guide'
LastUpdated: 2025-12-01T15:01:32.400651902-05:00
---




# Streamline templates and guidelines - move to embedded scaffolding

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- templates
- guidelines

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
