# Source

Copied from /home/manuel/.cursor/commands/docmgr.md

# docmgr — Documentation Management Workflow

## Overview

Use docmgr to create, manage, and search structured documentation organized into ticket workspaces. docmgr transforms ad-hoc markdown into a searchable knowledge base with bidirectional code-to-doc linking.

**Core value:** Structure without rigidity — consistent metadata and organization, but flexible enough for your workflow.

## Quick Start

### Shell safety: quote backticks

When copying commands from markdown/docs:

- **Never paste unquoted backticks (`...`) into your shell** — bash/zsh treat backticks as **command substitution**.
- If a doc shows a path like `` `ttmp/.../doc.md` ``:
  - **Preferred**: remove the backticks → `ttmp/.../doc.md`
  - **Or**: wrap the whole argument in quotes → `'`ttmp/.../doc.md`'` (or escape them as `\``)

### First-Time Setup (One-Time)

```bash
# Check if already initialized
docmgr status --summary-only

# If not initialized, run:
docmgr init --seed-vocabulary

# Verify
docmgr vocab list
```

### Create a Ticket

```bash
docmgr ticket create-ticket \
  --ticket TICKET-ID \
  --title "Descriptive Title" \
  --topics topic1,topic2,topic3
```

This creates `ttmp/YYYY/MM/DD/TICKET-ID--slug/` with `index.md`, `tasks.md`, and `changelog.md`.

### Add Documents

```bash
# Add a document
docmgr doc add \
  --ticket TICKET-ID \
  --doc-type analysis \
  --title "Document Title"

# Relate files immediately after creating (to the ticket index)
docmgr doc relate \
  --ticket TICKET-ID \
  --file-note "path/to/file.go:Why this file matters" \
  --file-note "path/to/other.go:Another reason"
```

**Important:** Always use `--file-note "path:reason"` format (colon separator, not dash).  
Use `docmgr doc relate` (not `docmgr relate`), and do **not** pass `--doc-type` to `doc relate` — target either:
- `--ticket TICKET-ID` to update the ticket index, or
- `--doc PATH` to update a specific document.

### Search Documentation

```bash
# Full-text search
docmgr doc search --query "search term"

# Filter by metadata
docmgr doc search --query "term" --topics backend --doc-type design-doc

# Find docs referencing a code file (reverse lookup)
docmgr doc search --file path/to/file.go

# Find docs referencing any file in a directory
docmgr doc search --dir path/to/directory/
```

## Common Workflows

### Working on an Existing Ticket

1. **Get oriented:**
   ```bash
   docmgr ticket list --ticket TICKET-ID
   docmgr doc list --ticket TICKET-ID
   docmgr task list --ticket TICKET-ID
   docmgr vocab list  # Review repository vocabulary
   ```

2. **Read existing docs:**
   - Start with `index.md` for overview
   - Review `changelog.md` for history
   - Check `tasks.md` for outstanding work
   - Read subdocuments in `analysis/`, `design-doc/`, `reference/`, etc.

3. **Work on tasks:**
   ```bash
   docmgr task check --ticket TICKET-ID --id 1,2
   docmgr task add --ticket TICKET-ID --text "New task"
   ```

4. **Relate files as you work:**
   ```bash
   docmgr doc relate --ticket TICKET-ID \
     --file-note "/abs/path/to/file.go:Why this file matters"
   ```

5. **Update changelog:**
   ```bash
   docmgr changelog update --ticket TICKET-ID \
     --entry "What changed and why" \
     --file-note "/abs/path/to/file.go:Reason"
   ```

6. **Validate:**
   ```bash
   docmgr doctor --ticket TICKET-ID --stale-after 30
   ```

### Creating Analysis Documents

```bash
# Create analysis document
docmgr doc add \
  --ticket TICKET-ID \
  --doc-type analysis \
  --title "Analysis Title"

# Relate source files immediately (typically to the ticket index)
docmgr doc relate \
  --ticket TICKET-ID \
  --file-note "path/to/source.go:Source file being analyzed" \
  --file-note "path/to/related.go:Related implementation"

# Update changelog
docmgr changelog update --ticket TICKET-ID \
  --entry "Created analysis document, related source files"
```

### Managing Metadata

```bash
# Update specific document
docmgr meta update \
  --ticket TICKET-ID \
  --doc-type analysis \
  --field Status \
  --value review

# Update all docs in ticket
docmgr meta update \
  --ticket TICKET-ID \
  --field Owners \
  --value "user1,user2"

# Update specific document by path
docmgr meta update \
  --doc ttmp/.../analysis/01-doc.md \
  --field Summary \
  --value "Updated summary"
```

### Closing a Ticket

```bash
# Check tasks first
docmgr task list --ticket TICKET-ID

# Close ticket
docmgr ticket close --ticket TICKET-ID

# Or with custom status/changelog
docmgr ticket close \
  --ticket TICKET-ID \
  --status review \
  --changelog-entry "Implementation complete, ready for review"
```

## Best Practices

### File Relating

**Always relate files when:**
- Creating a document (do it immediately after `doc add`)
- Modifying a document (relate new files you reference)
- After implementation (link key implementation files)

**Use absolute paths** for clarity:
```bash
docmgr doc relate --ticket TICKET-ID \
  --file-note "/home/user/project/src/file.go:Reason"
```

**Subdocument-first linking:**
- Relate most files to focused subdocuments (design-doc/reference/playbook)
- Keep index.md as overview linking to subdocuments
- Aim for 3-7 RelatedFiles per ticket (not 20+)

### Changelog Entries

**Always include:**
- What changed
- Why it changed
- Related files (use `--file-note`)

**Keep entries short** — mention what changed and link relevant files.

### Task Management

- Keep tasks focused on actionable steps
- Check off tasks as you complete them
- Use changelog to describe what actually changed
- Reference tasks in changelog when status changes

### Validation

**Run doctor regularly:**
```bash
# Quick check
docmgr doctor --ticket TICKET-ID

# Full workspace scan
docmgr doctor --all --stale-after 30 --fail-on error

# With structured output for CI
docmgr doctor --all --diagnostics-json diagnostics.json --fail-on warning
```

**What doctor checks:**
- Missing or invalid frontmatter
- Unknown topics/doc-types/status (warns, doesn't fail)
- Missing Note on RelatedFiles entries
- Missing files in RelatedFiles
- Stale docs (older than threshold)

## Document Types

Common doc types (custom types allowed):
- **analysis** - Code analysis, research, deep dives
- **design-doc** - Architecture and design decisions
- **reference** - API contracts, data schemas, how things work
- **playbook** - Test procedures, operational runbooks
- **index** - Ticket overview (auto-created)

## Vocabulary Management

```bash
# View vocabulary
docmgr vocab list
docmgr vocab list --category topics
docmgr vocab list --category docTypes
docmgr vocab list --category status

# Add custom entries
docmgr vocab add --category topics --slug my-topic --description "Description"
```

## Structured Output (Automation)

```bash
# JSON output for scripts
docmgr doc search --query "term" --with-glaze-output --output json

# List tickets as JSON
docmgr list tickets --with-glaze-output --output json

# Extract paths for bulk operations
docmgr list docs --ticket TICKET-ID --with-glaze-output --select path
```

## Common Patterns

### Pattern: Create Document + Relate Files

```bash
TICKET="TICKET-ID"
docmgr doc add --ticket $TICKET --doc-type analysis --title "Title"
docmgr doc relate --ticket $TICKET \
  --file-note "file1.go:Reason 1" \
  --file-note "file2.go:Reason 2"
docmgr changelog update --ticket $TICKET --entry "Created doc, related files"
```

### Pattern: Update Multiple Documents

```bash
# Update status on all design-docs
docmgr meta update \
  --ticket TICKET-ID \
  --doc-type design-doc \
  --field Status \
  --value review
```

### Pattern: Find Documentation for Code Review

```bash
# Find docs referencing a file you're reviewing
docmgr doc search --file path/to/file.go

# Find docs referencing directory
docmgr doc search --dir path/to/directory/
```

## Troubleshooting

### Path Issues

**Problem:** `docmgr doc relate` fails with path errors

**Solution:** Use absolute paths from workspace root (where `.ttmp.yaml` lives), not relative to `ttmp/`

```bash
# Wrong (relative to ttmp/)
docmgr doc relate --doc ttmp/.../doc.md --file-note "src/file.go:..."

# Right (absolute from workspace root)
docmgr doc relate --doc ttmp/.../doc.md --file-note "/abs/path/to/src/file.go:..."
```

### Backticks in copied commands

**Problem:** You copied a path that was displayed with backticks (markdown inline code), and the shell errors or runs something unexpected.

**Solution:** **Remove the backticks** before running the command, or quote/escape them.

```bash
# WRONG (backticks trigger command substitution)
docmgr meta update --doc `ttmp/.../analysis/01-doc.md` --field Summary --value "..."

# RIGHT (remove backticks)
docmgr meta update --doc ttmp/.../analysis/01-doc.md --field Summary --value "..."

# ALSO OK (quote the backticks so the shell doesn't execute them)
docmgr meta update --doc '`ttmp/.../analysis/01-doc.md`' --field Summary --value "..."
```

### File-Note Format

**Problem:** File notes not working

**Solution:** Use colon (`:`) separator, not dash or other characters

```bash
# Wrong
--file-note "file.go - Reason"

# Right
--file-note "file.go:Reason"
```

### Unknown Topics/Status

**Problem:** Doctor warns about unknown topics

**Solution:** Add to vocabulary or ignore (warnings don't fail)

```bash
docmgr vocab add --category topics --slug my-topic --description "Description"
```

## Key Commands Reference

```bash
# Tickets
docmgr ticket create-ticket --ticket ID --title "Title" --topics t1,t2
docmgr ticket list --ticket ID
docmgr ticket close --ticket ID

# Documents
docmgr doc add --ticket ID --doc-type TYPE --title "Title"
docmgr doc relate --ticket ID --file-note "path:reason"
docmgr doc search --query "term" --topics t1 --doc-type TYPE
docmgr doc list --ticket ID

# Metadata
docmgr meta update --ticket ID --field Field --value Value

# Tasks
docmgr task add --ticket ID --text "Task"
docmgr task check --ticket ID --id 1,2
docmgr task list --ticket ID

# Changelog
docmgr changelog update --ticket ID --entry "Entry" --file-note "path:reason"

# Validation
docmgr doctor --ticket ID --stale-after 30
docmgr validate frontmatter --doc path/to/doc.md --suggest-fixes

# Vocabulary
docmgr vocab list
docmgr vocab add --category CATEGORY --slug SLUG --description "Desc"

# Status
docmgr status --summary-only
```

## Integration with Debate Framework

When using the debate framework:

1. **Create debate documents:**
   ```bash
   docmgr doc add --ticket TICKET-ID --doc-type reference --title "Debate Format and Candidates"
   docmgr doc add --ticket TICKET-ID --doc-type reference --title "Debate Questions"
   ```

2. **Store debate rounds:**
   ```bash
   docmgr doc add --ticket TICKET-ID --doc-type reference --title "Debate Round N"
   ```

3. **Relate source files:**
   ```bash
   docmgr doc relate --ticket TICKET-ID \
     --file-note "path/to/source.go:Analyzed in debate"
   ```

4. **Update changelog after each round:**
   ```bash
   docmgr changelog update --ticket TICKET-ID \
     --entry "Completed debate round N on topic X"
   ```

## References

- Full documentation: `docmgr help how-to-use`
- Setup guide: `docmgr help how-to-setup`
- YAML validation: `docmgr help yaml-frontmatter-validation`
- CI automation: `docmgr help ci-automation`

---

**Use docmgr to create structured, searchable documentation that stays connected to your code.**
