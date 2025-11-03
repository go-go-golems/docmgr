---
Title: docmgr CLI Guide
Slug: cli-guide
Short: End-to-end guide for initializing, documenting, searching, and validating with docmgr.
Topics:
- docmgr
- documentation
- cli
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## 1. Overview

`docmgr` helps teams keep documentation close to code and current with development work. It does this by creating a small, standardized workspace per ticket, enforcing a minimal metadata contract (via YAML frontmatter), and providing practical tools to find, relate, and validate documents as code evolves.

Why this matters for you as a developer:

- When you start a ticket, you get a ready-to-use space for design notes, references, and playbooks.
- When you revisit work, you can quickly find the right docs by ticket, topic, code path, or external reference.
- When reviewing changes, you can check the health of docs (completeness, staleness, broken links) with one command.

This guide explains the core ideas and shows the main commands you’ll use day to day.

## 2. Quick Start

The commands below seed a controlled vocabulary, create a new ticket workspace, add a few documents, enrich metadata, and validate the workspace. Run them from your repository root.

```bash
# 1) Seed vocabulary (optional but recommended)
# These entries act as a shared language across docs for filtering and validation.
docmgr vocab add --category topics   --slug backend --description "Backend services"
docmgr vocab add --category topics   --slug chat    --description "Chat features"
docmgr vocab add --category topics   --slug websocket
docmgr vocab add --category docTypes --slug index
docmgr vocab add --category docTypes --slug design-doc
docmgr vocab add --category docTypes --slug reference
docmgr vocab add --category docTypes --slug playbook
docmgr vocab add --category intent   --slug long-term

# 2) Initialize a ticket workspace under ttmp/
# Creates a dedicated directory with an index and standard subfolders.
docmgr init --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket --root ttmp

# 3) Add documents
# Add a design doc, a reference doc, and a playbook to start capturing context.
docmgr add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy" --root ttmp
docmgr add --ticket MEN-4242 --doc-type reference  --title "Chat WebSocket Lifecycle"    --root ttmp
docmgr add --ticket MEN-4242 --doc-type playbook   --title "Smoke Tests for Chat"        --root ttmp

# 4) Update metadata on the ticket index
# Owners and Summary improve discoverability; RelatedFiles enable reverse lookup.
INDEX_MD="ttmp/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/index.md"
docmgr meta update --doc "$INDEX_MD" --field Owners          --value "manuel,alex"
docmgr meta update --doc "$INDEX_MD" --field Summary         --value "Unify chat HTTP paths and stabilize WebSocket flows."
docmgr meta update --doc "$INDEX_MD" --field ExternalSources --value "https://example.com/rfc/chat-api,https://example.com/ws-lifecycle"
docmgr meta update --doc "$INDEX_MD" --field RelatedFiles    --value "backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts"

# 5) Validate the workspace
# Check for missing fields, staleness, and broken file references.
docmgr doctor --root ttmp --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error
```

## 3. Core Concepts

### 4.1 Workspace Structure

Each ticket gets its own workspace under `ttmp/` (configurable with `--root`). This keeps short-lived artifacts connected to code while avoiding sprawling wiki pages. Workspaces are easy to archive or pivot as tickets evolve.

- `MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/`
  - `index.md` (frontmatter and summary)
  - `design/` (design documents)
  - `reference/` (contracts, API references)
  - `playbooks/` (operational steps, QA smoke tests)
  - `scripts/`, `sources/`, `various/`, `archive/`
  - `.meta/` (internal data)
- At root: `_templates/` and `_guidelines/` are scaffolded for consistency

### 4.2 Frontmatter Metadata

Each document starts with YAML frontmatter. This lightweight contract makes docs searchable and checkable. Think of it as a schema for documentation:

- Title, Ticket, Status, Topics, DocType, Intent
- Owners, RelatedFiles, ExternalSources, Summary, LastUpdated

`meta update` edits frontmatter safely and updates `LastUpdated` for you.

### 4.3 Vocabulary

The workspace vocabulary lives at `ttmp/vocabulary.yaml` by default (overridable via `.ttmp.yaml:vocabulary`). It defines the allowed `Topics`, `DocType`, and `Intent`. This prevents one-off spellings (“Web sockets” vs “websocket”) and keeps lists predictable for filters and automation. `doctor` warns on unknown values.

## 4. Commands

### 5.1 Vocabulary

Use vocabulary commands to establish the shared language of your project. Start small and grow with consensus. Unknown values will show up as warnings in `doctor`.
List entries:

```bash
docmgr vocab list --category topics
docmgr vocab list --category docTypes
docmgr vocab list --category intent
```

Add entries:

```bash
docmgr vocab add --category topics --slug observability --description "Logging and metrics"
docmgr vocab add --category docTypes --slug adr --description "Architecture Decision Record"
```

### 5.2 Initialize a Workspace

Run this when you start a ticket. It creates a consistent place to capture thinking and decisions. The structure is intentionally simple for quick adoption.
```bash
docmgr init --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket \
  --root ttmp [--force]
```

Creates the ticket directory, `index.md`, and `tasks.md`/`changelog.md`, and scaffolds `_templates/` and `_guidelines/` at the root.

### 5.3 Add Documents

Create additional documents as needed. Use short, descriptive titles; you can refine content later.
```bash
docmgr add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy" --root ttmp
```

### 5.4 Guidelines

Guidelines provide structure and “what good looks like” for each doc type. They help new contributors produce consistent, reviewable docs.
```bash
docmgr guidelines --doc-type design-doc --output markdown
```

Prints the guideline text for the given type. Files in `ttmp/_guidelines/` override embedded defaults.
See also: `docmgr help templates-and-guidelines` for how templates and guidelines fit together and how to customize them.

### 5.5 Update Metadata

Keep `Owners`, `Summary`, and `RelatedFiles` current. This makes search, review, and onboarding faster.
```bash
# Update a specific document
docmgr meta update --doc ttmp/MEN-4242-.../index.md --field Owners --value "manuel,alex"

# Update all docs for a ticket (optionally filter by type)
docmgr meta update --ticket MEN-4242 --doc-type design-doc --field Topics --value "chat,backend"
```

Supported fields: Title, Ticket, Status, Topics, DocType, Intent, Owners, RelatedFiles, ExternalSources, Summary.

### 5.6 List Tickets and Docs

Use listing commands to navigate by ticket. This is useful in reviews and when returning to paused work.
```bash
docmgr list tickets --root ttmp [--ticket MEN-4242]
docmgr list docs    --root ttmp --ticket MEN-4242
```

### 5.7 Search (Content + Metadata)

Search supports both content queries and metadata filters. Reverse lookups (`--file`, `--dir`) help you find docs from code paths; `--external-source` helps find docs tied to external references. Date filters surface recent activity.
```bash
# Content search
docmgr search --query "WebSocket" --ticket MEN-4242

# Metadata filters
docmgr search --ticket MEN-4242 --topics websocket,backend --doc-type design-doc

# Reverse lookup by file or directory
docmgr search --file backend/chat/api/register.go
docmgr search --dir  web/src/store/api/

# External source reference
docmgr search --external-source "https://example.com/ws-lifecycle"

# Date filters (relative and absolute)
docmgr search --updated-since "2 weeks ago" --ticket MEN-4242
docmgr search --since "last month" --until "today"

# File suggestions (heuristics: related files, git, ripgrep/grep)
docmgr search --ticket MEN-4242 --topics chat --files
```

Relative date formats supported include: `today`, `yesterday`, `last week`, `this month`, `last month`, `2 weeks ago`, as well as ISO-like absolute dates (for example, `2025-01-01`).

### 5.8 Doctor (Validation)

Run `doctor` during development and reviews. It’s a safety net to catch drift (stale docs), broken relationships (missing files), and inconsistent metadata (unknown vocabulary).
```bash
# Typical validation
docmgr doctor --root ttmp --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error

# Ignore specific paths using glob patterns
docmgr doctor --root ttmp --ignore-glob "ttmp/*/design/index.md" --fail-on warning
```

Doctor checks:

- Presence and validity of `index.md`
- Multiple `index.md` files under a single ticket
- Staleness via `LastUpdated` (configurable threshold)
- Required fields (Title, Ticket, Status, Topics)
- Unknown `Topics`, `DocType`, and `Intent` (validated against vocabulary)
- `RelatedFiles` existence on disk

`--fail-on` controls exit behavior for CI or pre-commit checks.

Ignore configuration:
- The command respects a `.docmgrignore` file at the repository root. Each non-empty line is a glob or name to ignore (comments start with `#`). Examples:
  - `.git/`, `node_modules/`, `dist/`
  - `ttmp/*/design/index.md`

### 5.9 Relate Code and Documents

Use `relate` to add or remove entries in `RelatedFiles` and to discover files to link. Each related file can carry a short note explaining why it matters.

```bash
# Add files to the ticket index
docmgr relate --ticket MEN-4242 --files \
  backend/chat/api/register.go,web/src/store/api/chatApi.ts

# Add to a specific document
docmgr relate --doc ttmp/MEN-4242-.../design/path-normalization-strategy.md \
  --files backend/chat/ws/manager.go

# See suggestions (no changes applied)
docmgr relate --ticket MEN-4242 --suggest --query WebSocket --topics chat

# Apply suggestions automatically to the ticket index
docmgr relate --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket

# Add notes for specific files (repeatable; format path:note or path=note)
docmgr relate --ticket MEN-4242 \
  --file-note "backend/chat/api/register.go:Registers chat routes (path normalization source)" \
  --file-note "web/src/store/api/chatApi.ts=Frontend API integration; must align with backend paths"
```

Suggestion output includes both a `source` and a human-readable `reason` column, such as:

- related_files → "referenced by documents"
- git_history → "recent commit activity"
- git_modified/staged/untracked → "working tree modified" / "staged for commit" / "untracked new file"
- ripgrep → "content match: <term>"

When `--apply-suggestions` is used, the combined suggestion reasons are stored as the file's note (unless overridden with `--file-note`).

### 5.10 Status (Workspace Summary)

Get a quick overview of the docs root: counts, staleness, and per-ticket breakdown.

```bash
# Full status with per-ticket rows and a summary row
docmgr status

# Only print the summary row
docmgr status --summary-only

# Adjust staleness threshold
docmgr status --stale-after 30

# Focus on one ticket
docmgr status --ticket MEN-4242
```

Status honors `.ttmp.yaml` configuration (root discovery) and uses `LastUpdated` to determine staleness.

## 5. End-to-End Workflow Example

1) Initialize a ticket workspace and add documents to capture intent early.
2) Document decisions and APIs; link to relevant code paths via `RelatedFiles`.
3) Use `search` to reconnect context as code changes (by query, topic, or file path).
4) Validate with `doctor` before merging; fix warnings or exclude known noise with ignore flags.

See `glazed/ttmp/2025-10-03/testing-doc-manager/` for runnable scripts that exercise all features end to end (including advanced `doctor` scenarios).

## 6. Troubleshooting

When the CLI returns errors, it’s usually a hint about parameter style or metadata contracts:

- "Too many arguments" on `init` or `search`: Both prefer explicit flags. Use `--ticket` for `init` and `--query` for content search.
- Doctor reports issues in `_templates` or `_guidelines`: These are scaffolding folders. Exclude them with `--ignore-dir _templates --ignore-dir _guidelines`.
- Unknown `Topics` / `DocType` / `Intent`: Either add the value with `vocab add` or correct the spelling to match the vocabulary.
- Missing `RelatedFiles`: Update the frontmatter to point to real paths (relative to the repo root) or prune stale paths.

## 7. CI Integration

Treat `doctor` as a quality gate for documentation health. This keeps docs aligned with code and prevents regressions:

```yaml
- name: Validate docs
  run: |
    docmgr doctor --root ttmp \
      --ignore-dir _templates --ignore-dir _guidelines \
      --stale-after 30 --fail-on error
```

## 8. Appendix: Tips

**Design-first**: Start with a brief executive summary and key decisions before details.

**Cite sources**: Use `ExternalSources` to link standards, RFCs, or tracking issues so reviewers can follow rationale.

**Keep ownership visible**: Maintain `Owners` and a concise `Summary` to speed up triage and handoffs.

For more, run: `docmgr help how-to-use`, `docmgr help how-to-setup`, and `docmgr help templates-and-guidelines`.


