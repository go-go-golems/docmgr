---
Title: Path handling analysis
Ticket: DOCMGR-PATH-NORMALIZE
Status: draft
Topics:
    - docmgr
    - search
    - developer-experience
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: docmgr/pkg/commands/relate.go
      Note: File-note inputs are stored verbatim without normalization
    - Path: docmgr/pkg/commands/search.go
      Note: --file and --dir filters do substring/prefix checks only
    - Path: docmgr/internal/workspace/config.go
      Note: Root resolution stops at locating doc root / repo root
ExternalSources: []
Summary: >
    Documents the current behavior and shortcomings of docmgr’s file-path
    handling so we have a baseline before introducing normalization and fuzzy
    matching.
LastUpdated: 2025-11-26
---

# Path handling analysis

## Goal

- Capture today’s behavior for file-path inputs that come from relate/search verbs.
- Identify every place where normalization or canonical comparison is missing.
- Provide concrete findings that feed the normalization design doc.

## Context

docmgr stores RelatedFiles entries directly from `--file-note` arguments and
relies on literal string comparisons later when reverse-searching documents.
This means we only get a match when users remember the exact relative path they
typed the first time. Repositories that mix absolute paths, `..` segments, or
different working directories make reverse lookup practically unusable.

`workspace.ResolveRoot()` solves “where is the docs root?” but no code converts
paths to repo-relative form, de-dupes separators, or resolves against the
document folder / .ttmp.yaml location. The end result is a fragile experience
for both linking and search.

## Quick Reference

| Surface | Current behavior | Risk |
| --- | --- | --- |
| `docmgr doc relate --file-note` | Stores the literal `path` key as provided. Notes merge, but no cleaning or resolution occurs. | Duplicate entries for the same file under different relative forms; hard to search. |
| `docmgr doc relate --remove-files` | Requires exact string match with stored `RelatedFiles.Path`. | Cannot delete entries created with a different working directory. |
| `docmgr doc search --file foo` | Iterates every document and uses `strings.Contains(rf.Path, query)` plus the reverse check. | Substring-only match ignores path segments and produces both false negatives (e.g., `../../foo` vs `src/foo`) and false positives. |
| `docmgr doc search --dir dir/` | Uses `strings.HasPrefix` on the raw path string and on document paths relative to the docs root. | Relative vs absolute confusion; directories outside the docs root can never match. |
| Other consumers (tasks, changelog, etc.) | None of them interpret RelatedFiles beyond copying literal values. | Any later feature that wants to reason about paths must re-implement normalization. |

## Usage Examples

- Linking as you edit a design doc from inside `docmgr/pkg/commands` often
  results in `--file-note "../../pkg/commands/relate.go:..."`. Later, running
  `docmgr doc search --file pkg/commands/relate.go` yields no results.
- Teams that keep `.ttmp.yaml` inside a mono-repo root see both `pkg/foo.go`
  (repo-relative) and `../pkg/foo.go` (doc-relative) entries for the same file.
- Absolute paths such as `/home/manuel/workspaces/.../relate.go` get committed
  to frontmatter, leaking host-specific paths into history.

## Related

- [Path normalization design](../design/01-path-normalization-canonicalization.md)

