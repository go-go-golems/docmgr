---
Title: Add incrementing numeric prefixes to scaffolded files
Ticket: DOCMGR-101-prefix-scaffolded-files
Status: active
Topics:
    - infrastructure
    - tools
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/cmd/docmgr/main.go
      Note: Changelog update now dual-mode
    - Path: docmgr/pkg/commands/add.go
      Note: Apply 2-digit prefixes on scaffold
    - Path: docmgr/pkg/commands/changelog.go
      Note: Added BareCommand; prints reminder after update
    - Path: docmgr/pkg/commands/doctor.go
      Note: Warn missing numeric prefix
    - Path: docmgr/pkg/commands/filename_prefix.go
      Note: Prefix helpers
    - Path: docmgr/pkg/commands/renumber.go
      Note: Resequence prefixes; update references
    - Path: docmgr/pkg/commands/vocab_add.go
      Note: Add --root flag and context print
    - Path: docmgr/pkg/commands/vocab_list.go
      Note: Add --root flag and context print
    - Path: docmgr/pkg/models/document.go
      Note: Parse RelatedFiles with Path/Note and path/note (case-insensitive)
ExternalSources: []
Summary: Apply 2‑digit numeric prefixes (01-, 02-) to all newly scaffolded docs across all subdirectories; doctor warns on missing prefix; add a renumber verb that also updates intra‑ticket references.
LastUpdated: 2025-11-06T12:12:40.671526961-05:00
---









# Add incrementing numeric prefixes to scaffolded files

## Overview

Scope is simplified and opinionated:

- Apply 2‑digit numeric prefixes to all subdirectories under a ticket. If a folder exceeds 99 items, switch to 3 digits for new files.
- No configuration or override flags.
- Doctor: warn only on missing prefix.
- New verb: renumber — renames files within the same ticket to restore sequential prefixes and updates references within that ticket.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- infrastructure
- tools

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
