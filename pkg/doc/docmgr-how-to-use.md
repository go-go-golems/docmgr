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

Working discipline: As you work, keep `tasks.md` and `changelog.md` current via the CLI. Prefer `docmgr tasks ...` and `docmgr changelog update` over manual edits so indexes and dates stay consistent.

## 2. Prerequisites

- `docmgr` available on PATH
- A Git repository with your codebase (so `RelatedFiles` paths make sense)

> For repository setup (including vocabulary), see: `docmgr help how-to-setup`.

## 3. Initialize the Ticket Workspace

```bash
docmgr create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket
```

This creates `ttmp/MEN-4242-.../` with `index.md`, `tasks.md`, and `changelog.md` under a standard structure.
If your repository doesn’t have a docs root yet (with `vocabulary.yaml`, `_templates/`, `_guidelines/`), run:

```bash
docmgr init
```

### What this index is for

This file is the ticket’s single entry point. It:

- Summarizes what the ticket does (one‑line Summary in frontmatter + this Overview)
- Points to the most important docs and code via `RelatedFiles` (frontmatter) and sections below
- Helps newcomers navigate quickly to design, reference, playbooks, and key source files
- Serves as the anchor for docmgr checks (health/staleness/links); keep it short and up‑to‑date

Keep this index concise. Put details in design/reference docs; use notes on `RelatedFiles` to explain why a file matters.

Update the body of `index.md` throughout the ticket — not just frontmatter via `meta update`. Maintain:

- Overview (goal, scope, constraints)
- Key Links (docs, code, data assets)
- Status (one‑line)
- Next steps (short checklist)

## 4. Add Documents

```bash
docmgr add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
docmgr add --ticket MEN-4242 --doc-type reference  --title "Chat WebSocket Lifecycle"
docmgr add --ticket MEN-4242 --doc-type playbook   --title "Smoke Tests for Chat"
```

Optionally consult guidelines for structure:

```bash
# Human-readable guideline text (default)
docmgr guidelines --doc-type design-doc

# Structured output (for tooling)
docmgr guidelines --doc-type design-doc --with-glaze-output --output yaml
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
docmgr list tickets --ticket MEN-4242
docmgr list docs    --ticket MEN-4242

# Structured
docmgr list tickets --with-glaze-output --output json

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

Use the changelog continuously during development (small, frequent entries). Treat it as your working log of decisions and progress, not just an end-of-ticket summary.

### What `changelog.md` is for

- Running log of notable changes, decisions, and learnings during the ticket
- Timestamped entries to reconstruct context later (e.g., date‑grouped notes)
- Lightweight status anchor for reviewers; keep lines short and clear
- Link to PRs, commits, references as relevant; add related files with short notes
- Update frequently as work progresses; prefer many small entries over one big dump

## 9. Validate with Doctor

```bash
# Preferred (when .docmgrignore is present): flags not needed
docmgr doctor --root ttmp --stale-after 30 --fail-on error

# Ad-hoc suppression example (optional)
docmgr doctor --root ttmp --ignore-glob "ttmp/*/design/index.md" --fail-on warning
```

Warnings to expect in real projects:

- Unknown topic/docType/intent (if not in vocabulary)
- Missing files listed in `RelatedFiles`
- Multiple `index.md` under a ticket (use `--ignore-glob` to suppress known duplicates)

## 10. Manage Tasks

Use the `tasks` commands to track the concrete steps for your ticket directly in `tasks.md`.

```bash
# List tasks with indexes
docmgr tasks list --ticket MEN-4242

# Add a new task
docmgr tasks add --ticket MEN-4242 --text "Update API docs for /chat/v2"

# Check / uncheck by id or ids (comma-separated)
docmgr tasks check   --ticket MEN-4242 --id 1,2,4
docmgr tasks uncheck --ticket MEN-4242 --id 1,2,4

# Edit and remove
docmgr tasks edit   --ticket MEN-4242 --id 2 --text "Align frontend routes with backend"
docmgr tasks remove --ticket MEN-4242 --id 3,5

Note: `--id` accepts a comma-separated list for operations that target IDs (check, uncheck, remove).
```

Tasks are standard Markdown checkboxes (`- [ ]` / `- [x]`). The commands only edit the specific task line, preserving the rest of the file.

### What `tasks.md` is for

- Canonical, machine‑readable checklist for the ticket (Markdown checkboxes)
- Tracks day‑to‑day execution; keep it current as tasks start/finish
- Break work into small, actionable items; optionally tag owners inline
- Use the `docmgr tasks` commands to add/check/edit/remove without manual formatting; update tasks as you start/finish work items so reviewers can track progress in real time

## 11. Check Workspace Status

Use `status` to see a concise overview of the docs under the root, including staleness based on `LastUpdated`:

```bash
docmgr status
docmgr status --summary-only
docmgr status --stale-after 30
```

## 12. Output Modes and UX

docmgr supports human-friendly defaults and structured output via Glaze.

- Human-friendly (default):
  - list tickets/docs: concise one-liners (ticket/title/status/topics/path/updated)
  - status: summary line (+ per-ticket lines unless `--summary-only`)
  - search: `path — title [ticket] :: snippet`; `--files` shows `file — reason (source=...)`
  - guidelines: raw guideline text (or list types with `--list`)
  - tasks list: `[#idx] [x| ] text (file=...)`
  - vocab list: `category: slug — description`

- Structured output:
  - Add `--with-glaze-output --output json|yaml|csv|table`
  - Available on: `list tickets`, `list docs`, `status`, `search`, `guidelines`, `vocab list`, `tasks list`

Examples:
```bash
# Human
docmgr list tickets
docmgr status --summary-only
docmgr search --query websocket
docmgr guidelines --doc-type design-doc

# Structured
docmgr list tickets --with-glaze-output --output json
docmgr status --with-glaze-output --output table
docmgr search --query websocket --with-glaze-output --output yaml
docmgr guidelines --doc-type design-doc --with-glaze-output --output json
```

### Glazed scripting recipes (no jq)

Use Glazed flags with `--with-glaze-output` to get machine-friendly output directly from docmgr commands.

- Paths only (newline-separated)
  - Tickets (all):
    ```bash
    docmgr list tickets --with-glaze-output --select path
    ```
  - Docs in a ticket:
    ```bash
    docmgr list docs --ticket MEN-4242 --with-glaze-output --select path
    ```

- CSV/TSV with specific columns
  - Single column (no header):
    ```bash
    docmgr list docs --ticket MEN-4242 --with-glaze-output --output csv --with-headers=false --fields path
    ```
  - Multiple columns (tab-separated):
    ```bash
    docmgr list docs --ticket MEN-4242 --with-glaze-output --output tsv --with-headers=true --fields doc_type,title,path
    ```

- Hide columns you don't need
  ```bash
  docmgr list docs --ticket MEN-4242 --with-glaze-output --output csv --fields ticket,doc_type,title,path --filter ticket
  ```

- One-line templated output per row
  ```bash
  docmgr list docs --ticket MEN-4242 --with-glaze-output \
    --select-template '{{.doc_type}} {{.title}}: {{.path}}' --select _0
  ```

### Stable column contracts (for scripting)

Use these field names with `--fields`, `--filter`, and `--select`.

- Tickets (`docmgr list tickets`):
  - `ticket,title,status,topics,path,last_updated`
- Docs (`docmgr list docs --ticket TICKET`):
  - `ticket,doc_type,title,status,topics,path,last_updated`
- Tasks (`docmgr tasks list --ticket TICKET`):
  - `index,checked,text,file`
- Vocabulary (`docmgr vocab list`):
  - `category,slug,description`

Discover quickly:

```bash
docmgr list docs --with-glaze-output --output csv --with-headers=true | sed -n '1p'
```

### Root discovery and shell gotchas

- `.ttmp.yaml` discovery walks up from CWD. If you need consistent behavior from nested subdirs, set an absolute `root` in `.ttmp.yaml` or run from repo root.
- When neither flag nor `.ttmp.yaml` is present, docmgr anchors the default root to the Git repository root if found (`<git-root>/ttmp`), else uses `<cwd>/ttmp`.
- `.ttmp.yaml` does not need to live in the repository root. In multi-repo or monorepo setups, place it at a parent directory to centralize a shared docs root and point different repos at distinct `ttmp/` directories as needed via `root` or `vocabulary`.
- Avoid parentheses in ticket dir names; quote/escape if you must use them:
  ```bash
  cd "ttmp/MEN-XXXX-name-\(with-parens\)"
  ```

## 13. Iterate and Maintain

- Keep `Owners`, `Summary`, and `RelatedFiles` current
- Regularly update `index.md`, `changelog.md`, and `tasks.md` as work progresses
- Use `guidelines`

## 14. Advanced: RelatedFiles with notes and ignores

### Structured RelatedFiles (with notes)

Prefer structured entries with short notes to explain why a file matters.

```yaml
RelatedFiles:
  - path: pinocchio/pkg/webchat/forwarder.go
    note: SEM mapping; projector side-channel source
  - path: pkg/snapshots/sqlite_store.go
    note: SQLite SnapshotStore (MVP persistence)
```

CLI examples:
```bash
# Add files + notes to the ticket index
docmgr relate --ticket MEN-3083 \
  --files pkg/webchat/forwarder.go,pkg/snapshots/sqlite_store.go \
  --file-note "pkg/webchat/forwarder.go:SEM mapping; projector side-channel source" \
  --file-note "pkg/snapshots/sqlite_store.go:SQLite SnapshotStore (MVP persistence)"

# Add to a specific document
docmgr relate --doc ttmp/MEN-3083-.../design/sem-event-flow.md \
  --files pkg/timeline/controller.go \
  --file-note "pkg/timeline/controller.go:TUI timeline lifecycle (apply/render)"

# Let docmgr suggest related files and store reasons as notes
docmgr relate --ticket MEN-3083 --suggest --apply-suggestions --query timeline --topics conversation,events
```

### Ignore noise with .docmgrignore

`docmgr init` creates `ttmp/.docmgrignore` with sensible defaults (`.git/`, `_templates/`, `_guidelines/`). Place `.docmgrignore` at your docs root (e.g., `ttmp/.docmgrignore`) to add more patterns. One pattern per line. `doctor` automatically respects these patterns, so you can omit `--ignore-dir`/`--ignore-glob` in most cases.

Common entries:
```
.git/
_templates/
_guidelines/
node_modules/
dist/
coverage/
2024-*
2025-*
```