
# Docmgr

## Overview

Use `docmgr` to keep documentation organized into ticket workspaces, with consistent metadata and bidirectional links between code/files and docs.

## Quick Start

### Shell safety: never paste unquoted backticks

Bash/zsh treat backticks as command substitution. If you see a path rendered as `` `ttmp/.../doc.md` ``:
- Prefer removing the backticks: `ttmp/.../doc.md`
- Or quote the whole argument: ``'`ttmp/.../doc.md`'``

### First-time setup (one-time)

```bash
docmgr status --summary-only
docmgr init --seed-vocabulary
docmgr vocab list
```

### Create a ticket

```bash
docmgr ticket create-ticket \
  --ticket TICKET-ID \
  --title "Descriptive Title" \
  --topics topic1,topic2
```

### Add a document + relate files immediately

```bash
docmgr doc add --ticket TICKET-ID --doc-type analysis --title "Document Title"
docmgr doc relate --ticket TICKET-ID \
  --file-note "/abs/path/to/file.go:Why this file matters"
```

## Core Conventions (strict)

- Use `docmgr doc relate` (not `docmgr relate`).
- Do not pass `--doc-type` to `doc relate`; target either `--ticket TICKET-ID` (ticket index) or `--doc PATH` (specific doc).
- Always format related files as `--file-note "path:reason"` (colon separator, not dash).
- Prefer absolute paths in `--file-note` for clarity and copy/paste safety.
- Prefer “subdocument-first” linking: relate most files to the focused subdoc, keep `index.md` as the overview.
- Keep “RelatedFiles” tight (roughly 3–7 per ticket, not 20+).

## Common Workflows

### Get oriented on an existing ticket

```bash
docmgr ticket list --ticket TICKET-ID
docmgr doc list --ticket TICKET-ID
docmgr task list --ticket TICKET-ID
docmgr vocab list
```

### Task bookkeeping

```bash
docmgr task add --ticket TICKET-ID --text "New task"
docmgr task check --ticket TICKET-ID --id 1,2
```

### Changelog update (what + why + related files)

```bash
docmgr changelog update --ticket TICKET-ID \
  --entry "What changed and why" \
  --file-note "/abs/path/to/file.go:Reason"
```

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
docmgr validate frontmatter --doc path/to/doc.md --suggest-fixes
```

## Vocabulary

If a `doctor` warning indicates an unknown topic/doc-type/category, add it to the vocabulary (or ignore if it’s intentionally out-of-vocab):

```bash
docmgr vocab add --category topics --slug my-topic --description "Description"
```

## Reference

Load `references/docmgr.md` for the full (long-form) docmgr workflow, command reference, and best practices.
