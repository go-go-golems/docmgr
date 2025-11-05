---
Title: Reference — Current outputs and flags
Ticket: DOCMGR-UX
Status: active
Topics:
    - tooling
    - ux
    - cli
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-05T14:07:05.531023232-05:00
---


# Reference — Current outputs and flags

## Goal
Document how to get machine-parseable output from docmgr commands today and the flags/toggles involved.

## Context
Most list-style commands are wired with Glazed dual-mode. Use the toggle flag `--with-glaze-output` and select an output format via `--output json|yaml|csv|table`.

## Quick Reference

- Tickets (JSON):
  - `docmgr list tickets --with-glaze-output --output json`
- Tickets (CSV):
  - `docmgr list tickets --with-glaze-output --output csv`
- Docs for a ticket (JSON):
  - `docmgr list docs --ticket TICKET --with-glaze-output --output json`
- Docs for a ticket (CSV):
  - `docmgr list docs --ticket TICKET --with-glaze-output --output csv`
- Tasks (JSON):
  - `docmgr tasks list --ticket TICKET --with-glaze-output --output json`
- Vocab (JSON):
  - `docmgr vocab list --with-glaze-output --output json`
- Search (JSON):
  - `docmgr search --query "..." --with-glaze-output --output json`

## Usage Examples

- Extract ticket directory path (CSV parsing):
  - `docmgr list tickets --with-glaze-output --output csv | awk -F',' '$1=="DOCMGR-UX"{print $2; exit}'`
- Compute index.md path from ticket dir:
  - `TICKET_DIR=$(docmgr list tickets --with-glaze-output --output csv | awk -F',' 'NR>1 && $1=="REORG-FEATURE-STRUCTURE"{print $2; exit}') && echo "$TICKET_DIR/index.md"`
- List docs for a ticket and print only paths (CSV parsing):
  - `docmgr list docs --ticket REORG-FEATURE-STRUCTURE --with-glaze-output --output csv | awk -F',' 'NR>1{print $4}'`
    - Columns order for CSV (current wiring): `ticket,doc_type,title,path,...`

## Related

- `docmgr/cmd/docmgr/main.go` (Dual-mode wiring with Glazed)
- `glazed/pkg/doc/tutorials/build-first-command.md` (structured output patterns)
