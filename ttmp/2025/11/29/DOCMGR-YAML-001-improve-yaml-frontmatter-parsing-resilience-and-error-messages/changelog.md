# Changelog

## 2025-11-29

- Initial workspace created


## 2025-11-29

Created ticket for YAML frontmatter parsing improvements including enhanced errors, auto-quoting, validation command, and auto-fix mode

### Related Files

- design/01-yaml-frontmatter-parsing-improvements.md — Comprehensive design document covering all proposed improvements


## 2025-12-01

Auto-closed: ticket was active but not created today


## 2025-12-01

Added position-aware frontmatter parsing with line/snippet diagnostics; introduced 'docmgr validate frontmatter' subcommand under 'validate'; added parser unit test and a frontmatter validation smoke script.


## 2025-12-01

Added YAML preprocessing/quoting helpers and wired them into read/write paths; frontmatter writer now enforces quoting and has unit coverage; added frontmatter validation smoke script.


## 2025-12-01

Validation now emits taxonomies with fix suggestions; added --suggest-fixes/--auto-fix, delimiter/peel scrubbing, and backups; extended frontmatter smoke to cover suggest/auto-fix and reran all scenarios (00-03, 15, 18).


## 2025-12-01

Auto-fix flow polished: suppress error taxonomy after successful fix and re-parse; if re-parse fails, render the new taxonomy. Reran validation smoke (18) with success.

