---
Title: Testing numeric filename prefixes in scaffolding
Ticket: DOCMGR-101-prefix-scaffolded-files
Status: active
Topics:
    - infrastructure
    - tools
DocType: playbook
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-06T12:12:19.530328053-05:00
---


# Testing numeric filename prefixes in scaffolding

## Purpose

Validate that numeric prefixes are applied automatically and that renumber repairs ordering and updates references.

## Environment Assumptions

- `.ttmp.yaml` root points to `docmgr/ttmp`.
- A test ticket exists (e.g., this ticket).

## Commands

```bash
# 1) Add three design docs; expect 01-, 02-, 03- prefixes automatically
docmgr add --ticket DOCMGR-101-prefix-scaffolded-files --doc-type design-doc --title "A"
docmgr add --ticket DOCMGR-101-prefix-scaffolded-files --doc-type design-doc --title "B"
docmgr add --ticket DOCMGR-101-prefix-scaffolded-files --doc-type design-doc --title "C"

# 2) Manually rename one to break sequence, then repair
mv docmgr/ttmp/DOCMGR-101-*/design/02-*.md docmgr/ttmp/DOCMGR-101-*/design/99-broken.md
docmgr renumber --ticket DOCMGR-101-prefix-scaffolded-files

# 3) List docs and visually confirm ordering and updated references
docmgr list docs --ticket DOCMGR-101-prefix-scaffolded-files
```

```bash
# Command sequence
```

## Exit Criteria

- Design docs are numbered sequentially with left‑padded width (2 digits).
- Renumber restores correct sequence after manual changes.

## Notes

- MVP does not rename existing files automatically during scaffolding; use `renumber` to repair ordering.
