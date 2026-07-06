---
description: 'Documentation management with the `docmgr` CLI: create and work in ticket workspaces (`ttmp/...`), add documents, relate code/files to docs, manage tasks/changelogs/metadata/vocabulary, and search/validate docs. Use when a user mentions `docmgr`, ticket docs, `docmgr doc relate`, `docmgr doc search`, YAML frontmatter validation, or asks to turn ad-hoc markdown into a structured, searchable knowledge base.'
metadata:
    title: Docmgr
    topics:
        - imported-skill
    what_for: 'Documentation management with the `docmgr` CLI: create and work in ticket workspaces (`ttmp/...`), add documents, relate code/files to docs, manage tasks/changelogs/metadata/vocabulary, and search/validate docs. Use when a user mentions `docmgr`, ticket docs, `docmgr doc relate`, `docmgr doc search`, YAML frontmatter validation, or asks to turn ad-hoc markdown into a structured, searchable knowledge base.'
    when_to_use: Use when working with Docmgr.
name: docmgr
---

# Docmgr

## Overview

Use `docmgr` to keep documentation organized into ticket workspaces, with consistent metadata and bidirectional links between code/files and docs.

Output contract (agent-friendly): successful mutations print a single line; failures exit non-zero with an actionable error (malformed `--file-note`, empty `--entry`, failed `meta update`, unknown task ID). Trust exit codes instead of parsing output. The workspace banner and coaching output only appear with the global `--verbose` flag.

## Quick Start

### Shell safety: never paste unquoted backticks

Bash/zsh treat backticks as command substitution. If you see a path rendered as `` `ttmp/.../doc.md` ``:
- Prefer removing the backticks: `ttmp/.../doc.md`
- Or quote the whole argument: ``'`ttmp/.../doc.md`'``

### First-time setup (one-time)

```bash
docmgr status --summary-only
docmgr init          # seeds vocabulary by default (--seed-vocabulary=false to skip)
docmgr vocab list
```

### Create a ticket

```bash
docmgr ticket create \
  --ticket TICKET-ID \
  --title "Descriptive Title" \
  --topics topic1,topic2
```

(`ticket create` is the canonical spelling; `create-ticket` remains as an alias. Likewise `ticket rename` replaces `rename-ticket` and `ticket list` replaces `ticket tickets`.)

### Add a document + relate files immediately

```bash
docmgr doc add --ticket TICKET-ID --doc-type analysis --title "Document Title"
docmgr doc relate --ticket TICKET-ID \
  --file-note "/abs/path/to/file.go:Why this file matters"
```

`relate` persists paths as explicit anchors (`repo://pkg/foo.go`, `ws://member/...`, `docs://...`, `abs:///...`) — see `docmgr help path-anchors`. You still ALWAYS pass absolute paths in `--file-note`; docmgr anchors them for you. Notes may contain commas and extra colons; a value without a `:` separator errors and exits 1.

### Forgiving references

- `--ticket` accepts the exact ID, a unique prefix, or a pasted workspace directory path (`2026/07/06/TICKET-ID--slug`).
- `--doc` accepts an absolute path, a cwd/repo/docs-root-relative path, a path with a duplicated `ttmp/` prefix, or a workspace-unique suffix (`design-doc/01-foo.md`). Ambiguous references list the candidates.

## Core Conventions (strict)

- Use `docmgr doc relate` (not `docmgr relate`).
- Do not pass `--doc-type` to `doc relate`; target either `--ticket TICKET-ID` (ticket index) or `--doc PATH` (specific doc).
- Always format related files as `--file-note "path:reason"` (colon separator, not dash).
- ALWAYS use absolute paths in `--file-note`; they are stored as anchored paths and stay unambiguous.
- Prefer "subdocument-first" linking: relate most files to the focused subdoc, keep `index.md` as the overview.
- Keep "RelatedFiles" tight (roughly 3-7 per ticket, not 20+).
- Store any ad-hoc scripts you create for a ticket in that ticket's `scripts/` directory under `ttmp/.../scripts` so they are tracked. Name scripts with a numerical prefix (`01-...`, `02-...`) to preserve execution order and trace investigation steps.
- Every active ticket should have a **diary** (typically `reference/02-diary.md` or similar). The diary records chronological investigation steps, what was tried, what failed, and what to do next. Read the diary before resuming work on a ticket.

## Common Workflows

### Get oriented on an existing ticket

```bash
docmgr ticket show TICKET-ID     # compact overview: status, tasks, docs, changelog
docmgr doc list --ticket TICKET-ID
docmgr task list --ticket TICKET-ID
docmgr vocab list
```

### Task bookkeeping (stable IDs)

Tasks carry stable IDs, persisted as invisible `<!-- t:xxxx -->` markers in `tasks.md`. `task add` echoes the new ID; `task list` shows them in brackets. Prefer stable IDs over positions (positions also still work).

```bash
docmgr task add --ticket TICKET-ID --text "New task"     # -> Task v0sv added ...
docmgr task list --ticket TICKET-ID                       # -> [v0sv] [ ] New task
docmgr task check --ticket TICKET-ID --id v0sv
docmgr task uncheck --ticket TICKET-ID --id v0sv
docmgr task edit --ticket TICKET-ID --id v0sv --text "Refined task"
docmgr task migrate --ticket TICKET-ID   # stamp IDs onto hand-written task lists
```

An unknown `--id` exits 1 and prints the current task table so you can pick the right ID.

### Changelog update (what + why + related files)

```bash
docmgr changelog update --ticket TICKET-ID \
  --entry "What changed and why" \
  --file-note "/abs/path/to/file.go:Reason"
```

`--entry` must be non-empty (empty entry = error, exit 1).

### Search

```bash
docmgr doc search --query "search term"
docmgr doc search --query "term" --topics backend --doc-type design-doc
docmgr doc search --file path/to/file.go
docmgr doc search --dir path/to/dir/
```

### Validate / hygiene

```bash
docmgr doctor --ticket TICKET-ID --stale-after 30
docmgr doctor --ticket TICKET-ID --fix        # safe frontmatter fixes (.bak) + anchor migration
docmgr doctor --all                           # multi-ticket: per-ticket rollup; add --details for everything
docmgr validate frontmatter --doc path/to/doc.md --suggest-fixes   # single-file check
```

Doctor checks all docs in a ticket (not just index.md); `sources/` is skipped unless `--include-sources`. `doctor --fix-anchors` migrates only legacy RelatedFiles paths to anchors.

### Structured output

Every verb (including mutations) supports `--with-glaze-output --output json|yaml|csv|table` when you need machine-readable results.

## Vocabulary

If a `doctor` warning indicates an unknown topic/doc-type/category, add it to the vocabulary (or ignore if it's intentionally out-of-vocab). Built-in doc types, intents, and statuses are always recognized.

```bash
docmgr vocab add --category topics --slug my-topic --description "Description"
```

## Reference

Load `references/docmgr.md` for the full (long-form) docmgr workflow, command reference, and best practices.
