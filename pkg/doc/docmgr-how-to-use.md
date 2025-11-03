---
Title: Tutorial — Using docmgr to Drive a Ticket Workflow
Slug: how-to-use
Short: Step-by-step tutorial to create, relate, search, and validate docs for a ticket.
Topics:
- docmgr
- tutorial
- workflow
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

## 1. Overview

This tutorial walks a developer through using `docmgr` to manage the documentation for a single ticket from start to finish: initialize a workspace, add documents, link code, search, and validate.

## 2. Prerequisites

- `docmgr` available on PATH
- A Git repository with your codebase (so `RelatedFiles` paths make sense)

> For repository setup (including vocabulary), see: `docmgr help how-to-setup`.

## 3. Initialize the Ticket Workspace

```bash
docmgr init --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket --root ttmp
```

This creates `ttmp/MEN-4242-.../` with `index.md` and scaffolds `_templates/` and `_guidelines/` at the root (if missing).

### What this index is for

This file is the ticket’s single entry point. It:

- Summarizes what the ticket does (one‑line Summary in frontmatter + this Overview)
- Points to the most important docs and code via `RelatedFiles` (frontmatter) and sections below
- Helps newcomers navigate quickly to design, reference, playbooks, and key source files
- Serves as the anchor for docmgr checks (health/staleness/links); keep it short and up‑to‑date

Keep this index concise. Put details in design/reference docs; use notes on `RelatedFiles` to explain why a file matters.

## 4. Add Documents

```bash
docmgr add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy" --root ttmp
docmgr add --ticket MEN-4242 --doc-type reference  --title "Chat WebSocket Lifecycle"    --root ttmp
docmgr add --ticket MEN-4242 --doc-type playbook   --title "Smoke Tests for Chat"        --root ttmp
```

Optionally consult guidelines for structure:

```bash
docmgr guidelines --doc-type design-doc --output markdown
```
See also: `docmgr help templates-and-guidelines` for customization and best practices.

## 5. Enrich Metadata

```bash
INDEX_MD="ttmp/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/index.md"
docmgr meta update --doc "$INDEX_MD" --field Owners          --value "manuel,alex"
docmgr meta update --doc "$INDEX_MD" --field Summary         --value "Unify chat HTTP paths and stabilize WebSocket flows."
docmgr meta update --doc "$INDEX_MD" --field ExternalSources --value "https://example.com/rfc/chat-api,https://example.com/ws-lifecycle"
docmgr meta update --doc "$INDEX_MD" --field RelatedFiles    --value "backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts"
```

## 6. Relate Code and Docs

Link code paths to your ticket so reviewers can jump from code to context:

```bash
# Add files to the ticket index
docmgr relate --ticket MEN-4242 --files \
  backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts

# Get suggestions with explanations (no changes applied)
docmgr relate --ticket MEN-4242 --suggest --query WebSocket --topics chat

# Apply suggestions to the ticket index (reasons are saved as notes)
docmgr relate --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket

# Add or update notes for specific files
docmgr relate --ticket MEN-4242 \
  --file-note "backend/chat/api/register.go:Registers chat routes (path normalization source)" \
  --file-note "web/src/store/api/chatApi.ts=Frontend API integration; must align with backend paths"
```

Suggestion output includes both `source` and `reason` (for example, "recent commit activity", "working tree modified", "referenced by documents"). When applying suggestions, combined reasons are saved as the file's note unless overridden with `--file-note`.

## 7. Explore and Search

```bash
docmgr list tickets --root ttmp --ticket MEN-4242
docmgr list docs    --root ttmp --ticket MEN-4242

# Content search
docmgr search --query "WebSocket" --ticket MEN-4242

# Metadata filters
docmgr search --ticket MEN-4242 --topics websocket,backend --doc-type design-doc

# Reverse lookups
docmgr search --file backend/chat/api/register.go
docmgr search --dir  web/src/store/api/

# External source
docmgr search --external-source "https://example.com/ws-lifecycle"

# Date filters
docmgr search --updated-since "1 day ago" --ticket MEN-4242
```

## 8. Record Changes in Changelog

Append dated entries to `changelog.md` and include related files when useful:

```bash
# Minimal entry
docmgr changelog update --ticket MEN-4242 --entry "Normalized chat API paths"

# With related files and notes
docmgr changelog update --ticket MEN-4242 \
  --files backend/chat/api/register.go,web/src/store/api/chatApi.ts \
  --file-note "backend/chat/api/register.go:Source of path normalization" \
  --file-note "web/src/store/api/chatApi.ts=Frontend integration"

# Use suggestions (print only) or apply them
docmgr changelog update --ticket MEN-4242 --suggest --query WebSocket
docmgr changelog update --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket
```

### What `changelog.md` is for

- Running log of notable changes, decisions, and learnings during the ticket
- Timestamped entries to reconstruct context later (e.g., date‑grouped notes)
- Lightweight status anchor for reviewers; keep lines short and clear
- Link to PRs, commits, references as relevant; add related files with short notes
- Update frequently as work progresses; prefer many small entries over one big dump

## 9. Validate with Doctor

```bash
docmgr doctor --root ttmp --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error
```

Warnings to expect in real projects:

- Unknown topic/docType/intent (if not in vocabulary)
- Missing files listed in `RelatedFiles`
- Multiple `index.md` under a ticket (use `--ignore-glob` to suppress known duplicates)

## 10. Manage Tasks

Use the `tasks` commands to track the concrete steps for your ticket directly in `tasks.md`.

```bash
# List tasks with indexes
docmgr tasks list --ticket MEN-4242 --root ttmp

# Add a new task
docmgr tasks add --ticket MEN-4242 --text "Update API docs for /chat/v2" --root ttmp

# Check / uncheck by id
docmgr tasks check   --ticket MEN-4242 --id 1 --root ttmp
docmgr tasks uncheck --ticket MEN-4242 --id 1 --root ttmp

# Edit and remove
docmgr tasks edit   --ticket MEN-4242 --id 2 --text "Align frontend routes with backend" --root ttmp
docmgr tasks remove --ticket MEN-4242 --id 3 --root ttmp
```

Tasks are standard Markdown checkboxes (`- [ ]` / `- [x]`). The commands only edit the specific task line, preserving the rest of the file.

### What `tasks.md` is for

- Canonical, machine‑readable checklist for the ticket (Markdown checkboxes)
- Tracks day‑to‑day execution; keep it current as tasks start/finish
- Break work into small, actionable items; optionally tag owners inline
- Use the `docmgr tasks` commands to add/check/edit/remove without manual formatting

## 11. Check Workspace Status

Use `status` to see a concise overview of the docs under the root, including staleness based on `LastUpdated`:

```bash
docmgr status
docmgr status --summary-only
docmgr status --stale-after 30
```

## 12. Iterate and Maintain

- Keep `Owners`, `Summary`, and `RelatedFiles` current
- Regularly update `index.md`, `changelog.md`, and `tasks.md` as work progresses
- Use `guidelines` to keep structure consistent
- Re-run `doctor` before merging changes


