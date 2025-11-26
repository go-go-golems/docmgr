---
Title: Command Verification Findings
Ticket: DOCMGR-DOC-VERIFY
Status: active
Topics:
    - docmgr
    - documentation
    - validation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Systematic verification of all docmgr command references in documentation files"
LastUpdated: 2025-11-26T14:16:31.118877545-05:00
---

# Command Verification Findings

## Goal

Verify that all `docmgr` command references in `docmgr-how-to-use.md` and `how-to-work-on-any-ticket.md` match the actual command implementations.

## Context

The user reported that some commands might be incorrect (e.g., `docmgr relate` should be `docmgr doc relate`, and `doc relate` doesn't take `--doc-type` argument). This document tracks systematic verification of all command references.

## Verification Process

1. Checked actual command structure via `docmgr --help` and subcommand help
2. Verified command implementations in `docmgr/cmd/docmgr/cmds/`
3. Cross-referenced all command references in both documentation files

## Verified Commands

### Command Structure
- ✅ `docmgr doc relate` - Correct (not `docmgr relate`)
- ✅ `docmgr doc list` - Correct
- ✅ `docmgr list docs` - Correct (alternative form)
- ✅ `docmgr ticket list` - Correct (alias for `docmgr ticket tickets`)
- ✅ `docmgr list tickets` - Correct
- ✅ `docmgr task list` - Correct
- ✅ `docmgr meta update` - Correct
- ✅ `docmgr changelog update` - Correct
- ✅ `docmgr ticket create-ticket` - Correct
- ✅ `docmgr ticket close` - Correct
- ✅ `docmgr doc add` - Correct
- ✅ `docmgr doc search` - Correct
- ✅ `docmgr doc guidelines` - Correct
- ✅ `docmgr vocab list` - Correct
- ✅ `docmgr vocab add` - Correct
- ✅ `docmgr status` - Correct
- ✅ `docmgr doctor` - Correct
- ✅ `docmgr init` - Correct
- ✅ `docmgr completion` - Correct
- ✅ `docmgr help` - Correct

### Flag Verification

#### `docmgr doc relate`
- ✅ `--ticket` - Correct
- ✅ `--doc` - Correct
- ✅ `--file-note` - Correct (required, repeatable)
- ✅ `--remove-files` - Correct
- ❌ `--doc-type` - NOT SUPPORTED (correctly not used in docs)

#### `docmgr meta update`
- ✅ `--ticket` - Correct
- ✅ `--doc` - Correct
- ✅ `--doc-type` - Correct (used with `--ticket` to filter)
- ✅ `--field` - Correct
- ✅ `--value` - Correct

#### `docmgr ticket list` / `docmgr list tickets`
- ✅ `--ticket` - Correct (filters by ticket identifier)
- ✅ `--root` - Correct
- ✅ `--with-glaze-output` - Correct
- ✅ `--output` - Correct

#### `docmgr doc list` / `docmgr list docs`
- ✅ `--ticket` - Correct
- ✅ `--doc-type` - Correct
- ✅ `--status` - Correct
- ✅ `--topics` - Correct
- ✅ `--root` - Correct

## Issues Found

### None Found

After systematic verification:
- ✅ All command paths are correct
- ✅ All flag usage is correct
- ✅ No instances of `docmgr relate` (should be `docmgr doc relate`)
- ✅ No instances of `--doc-type` used incorrectly with `doc relate`
- ✅ All command aliases are correctly documented

## Notes

1. Both `docmgr doc list` and `docmgr list docs` work (both are valid)
2. Both `docmgr ticket list` and `docmgr list tickets` work (both are valid)
3. `docmgr ticket list` is an alias for `docmgr ticket tickets`
4. All commands support `--with-glaze-output` for structured output (correctly documented)

## Files Verified

1. ✅ `docmgr/pkg/doc/docmgr-how-to-use.md` - All commands verified
2. ✅ `docmgr/pkg/doc/how-to-work-on-any-ticket.md` - All commands verified

## Conclusion

All command references in both documentation files are correct. No fixes needed.
