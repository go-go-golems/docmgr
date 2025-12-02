---
Title: Ticket move duplicates ticket name in directory when title contains ticket identifier
Ticket: DOCMGR-TICKET-MOVE-001
Status: complete
Topics:
    - bug
    - ticket-move
    - path-template
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/utils/slug.go
      Note: Added SlugifyTitleForTicket helper function to strip ticket identifiers from titles
    - Path: pkg/utils/slug_test.go
      Note: Added comprehensive unit tests for SlugifyTitleForTicket
ExternalSources: []
Summary: When ticket title contains the ticket identifier, ticket move command creates directory names with duplicate ticket identifiers (e.g., TEST-9999-test-9999-...)
LastUpdated: 2025-12-02T10:53:26.560548891-05:00
---




# Ticket move duplicates ticket name in directory when title contains ticket identifier

## Overview

When a ticket's `index.md` title field contains the ticket identifier (e.g., "TEST-9999: Description"), the `docmgr ticket move` command creates directory names with duplicate ticket identifiers. This occurs because the slug is computed from the title (which includes the ticket name), and then the path template combines `{{TICKET}}` with `{{SLUG}}`, resulting in patterns like `TEST-9999-test-9999-description`.

**Example:**
- Ticket: `TEST-9999`
- Title in index.md: `TEST-9999: Test ticket with ticket in title`
- Resulting directory: `TEST-9999-test-9999-test-ticket-with-ticket-in-title`

The bug affects both `ticket create-ticket` (when title includes ticket) and `ticket move` (when moving tickets with titles that include the ticket identifier).

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- bug
- ticket-move
- path-template

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
