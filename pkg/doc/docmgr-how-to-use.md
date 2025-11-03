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

## 8. Validate with Doctor

```bash
docmgr doctor --root ttmp --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error
```

Warnings to expect in real projects:

- Unknown topic/docType/intent (if not in vocabulary)
- Missing files listed in `RelatedFiles`
- Multiple `index.md` under a ticket (use `--ignore-glob` to suppress known duplicates)

## 9. Check Workspace Status

Use `status` to see a concise overview of the docs under the root, including staleness based on `LastUpdated`:

```bash
docmgr status
docmgr status --summary-only
docmgr status --stale-after 30
```

## 10. Iterate and Maintain

- Keep `Owners`, `Summary`, and `RelatedFiles` current
- Use `guidelines` to keep structure consistent
- Re-run `doctor` before merging changes


