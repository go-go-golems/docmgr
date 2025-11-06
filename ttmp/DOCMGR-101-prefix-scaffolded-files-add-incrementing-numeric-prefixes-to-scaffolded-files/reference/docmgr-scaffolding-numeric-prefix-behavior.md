---
Title: 'docmgr scaffolding: numeric prefix behavior'
Ticket: DOCMGR-101-prefix-scaffolded-files
Status: active
Topics:
    - infrastructure
    - tools
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-06T12:12:19.467417463-05:00
---


# docmgr scaffolding: numeric prefix behavior

## Goal

Define the user-facing contract for numeric filename prefixes during scaffolding: automatic behavior, doctor warnings, and the renumber verb.

## Context

Applies automatically when creating docs under a ticket with `docmgr add` (and during `create-ticket` scaffolding). No configuration or per-invocation overrides.

## Quick Reference

Behavior:
- New files get a 2‑digit prefix (01-, 02-, …) in all subdirectories; when a folder exceeds 99 items, new files use 3 digits (100-, 101-, …).
- Existing files are not renamed automatically.
- Doctor warns when a file is missing a numeric prefix.

Renumber verb (new):
- `docmgr renumber --ticket <TICKET>` resequences prefixes in the ticket and updates references within that ticket.
- Safe rename behavior: skip files with uncommitted changes (TBD), or require clean working state.

## Usage Examples

```bash
# Create docs; prefixes applied automatically
docmgr add --ticket DOCMGR-101-prefix-scaffolded-files --doc-type design-doc --title "Ordering and collision handling"
docmgr add --ticket DOCMGR-101-prefix-scaffolded-files --doc-type reference  --title "Prefix behavior"
docmgr add --ticket DOCMGR-101-prefix-scaffolded-files --doc-type playbook   --title "Manual checks"

# Repair ordering and update links within the ticket
docmgr renumber --ticket DOCMGR-101-prefix-scaffolded-files
```

## Related

- Design: Filename numbering policy for newly scaffolded docs
- Playbook: Testing numeric filename prefixes in scaffolding
