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

The commands below initialize a documentation workspace with seeded vocabulary, create a ticket workspace, add documents, enrich metadata, and validate. Run them from your repository root.

```bash
# 1) Check if already initialized
docmgr status --summary-only
# If error "root directory does not exist", proceed with init

# 2) Initialize the docs root with seeded vocabulary
# Creates ttmp/, vocabulary.yaml (with defaults), templates, and guidelines
docmgr init --seed-vocabulary

# 3) Verify initialization
docmgr vocab list  # Should show seeded topics (chat, backend, websocket)

# 4) Create a ticket workspace under ttmp/
# Creates a dedicated directory with index, tasks, changelog, and standard subfolders.
docmgr ticket create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket

By default the workspace lives under `ttmp/YYYY/MM/DD/<ticket>-<slug>/`. Override this with `--path-template` when you need a different hierarchy (placeholders: `{{YYYY}}`, `{{MM}}`, `{{DD}}`, `{{DATE}}`, `{{TICKET}}`, `{{SLUG}}`, `{{TITLE}}`).

# 5) Add documents
# Add a design doc, a reference doc, and a playbook to start capturing context.
docmgr doc add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
docmgr doc add --ticket MEN-4242 --doc-type reference  --title "Chat WebSocket Lifecycle"
docmgr doc add --ticket MEN-4242 --doc-type playbook   --title "Smoke Tests for Chat"

# 6) Update metadata on the ticket index
# Owners and Summary improve discoverability; RelatedFiles enable reverse lookup.
INDEX_MD=$(find ttmp -type f -path "*/MEN-4242-*/index.md" -print -quit)
test -n "$INDEX_MD"
docmgr meta update --doc "$INDEX_MD" --field Owners          --value "manuel,alex"
docmgr meta update --doc "$INDEX_MD" --field Summary         --value "Unify chat HTTP paths and stabilize WebSocket flows."
docmgr meta update --doc "$INDEX_MD" --field ExternalSources --value "https://example.com/rfc/chat-api,https://example.com/ws-lifecycle"
docmgr meta update --doc "$INDEX_MD" --field RelatedFiles    --value "backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts"

# 7) Validate the workspace
# Check for missing fields, staleness, and broken file references.
# When `.docmgrignore` is present, you can omit ignore flags entirely.
docmgr doctor --root ttmp --stale-after 30 --fail-on error
```

### Output modes (human vs structured)

Many verbs now support dual output modes:

- Default: human-friendly text (ideal for terminals and LLM prompts)
- Structured: enable with `--with-glaze-output`, then select format via `--output json|yaml|csv|table`

Examples:

```bash
# Human-readable (default)
docmgr ticket list

# Structured
docmgr ticket list --with-glaze-output --output json
docmgr doc search --query websocket --with-glaze-output --output yaml
```

## 3. Core Concepts

### 3.1 Root Configuration and Discovery

You rarely need `--root`. docmgr resolves the docs root in this order:

- Flag: `--root /abs/or/relative/path` (relative paths are anchored to CWD)
- `.ttmp.yaml` nearest to CWD: `root: ttmp` (interpreted relative to the config file location)
- Git repository root: `<git-root>/ttmp` if a `.git/` directory is found while walking up
- Fallback: `<cwd>/ttmp`

Vocabulary path is resolved similarly via `.ttmp.yaml:vocabulary` (absolute or relative to the config); otherwise defaults to `<root>/vocabulary.yaml`.


### 3.2 Workspace Structure

Each ticket gets its own workspace under `ttmp/` (configurable with `--root`). This keeps short-lived artifacts connected to code while avoiding sprawling wiki pages. Workspaces are easy to archive or pivot as tickets evolve.

- `MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/`
  - `index.md` (frontmatter and summary)
  - `design-doc/` (design documents)
  - `reference/` (contracts, API references)
  - `playbook/` (operational steps, QA smoke tests)
  - `<doc-type>/` (custom types create their own subdir)
  - `scripts/`, `sources/`, `archive/`
  - `.meta/` (internal data)
- At root: `_templates/` and `_guidelines/` are scaffolded for consistency

Slugification of the directory and filenames:

- Lowercase; any non‑alphanumeric is replaced with `-`; multiple `-` are collapsed; trim leading/trailing `-`.
- Example: `go-go-mento: Webchat/Web hydration and integration reference` → `go-go-mento-webchat-web-hydration-and-integration-reference`.

### 3.3 Frontmatter Metadata

Each document starts with YAML frontmatter. This lightweight contract makes docs searchable and checkable. Think of it as a schema for documentation:

- Title, Ticket, Status, Topics, DocType, Intent
- Owners, RelatedFiles, ExternalSources, Summary, LastUpdated

`meta update` edits frontmatter safely and updates `LastUpdated` for you.

### 3.4 Vocabulary

The workspace vocabulary lives at `ttmp/vocabulary.yaml` by default (overridable via `.ttmp.yaml:vocabulary`). It defines the allowed `Topics`, `DocType`, and `Intent`. This prevents one-off spellings (“Web sockets” vs “websocket”) and keeps lists predictable for filters and automation. `doctor` warns on unknown values.

## 4. Commands

### 4.1 Vocabulary

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

### 4.2 Initialize a Docs Root

Run this once per repository (or shared parent) to create the docs root with vocabulary, templates, guidelines, and a default `.docmgrignore`.

```bash
# Recommended: seed with common defaults
docmgr init --seed-vocabulary

# Or initialize with empty vocabulary
docmgr init

# Force re-scaffold templates/guidelines
docmgr init --force
```

Creates the `ttmp/` directory if missing, and scaffolds `_templates/` and `_guidelines/`. With `--seed-vocabulary`, populates `vocabulary.yaml` with common topics (chat, backend, websocket), doc types (design-doc, reference, playbook), and intents (long-term).

### 4.3 Create a Ticket Workspace

Run this when you start a ticket. It creates a consistent place to capture thinking and decisions.
```bash
docmgr ticket create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket \
  [--force]
```

Creates the ticket directory with `index.md`, and `tasks.md`/`changelog.md` under a standard structure.

### 4.4 Add Documents

Create additional documents as needed. Use short, descriptive titles; you can refine content later.
```bash
docmgr doc add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
docmgr doc add --ticket MEN-4242 --doc-type til        --title "TIL — Hydration end-to-end"

# Optional overrides (taken from ticket by default)
docmgr doc add --ticket MEN-4242 \
  --doc-type til \
  --title "TIL conv-id vs run-id hydration curl debugging 2025-11-03" \
  --topics hydration,persistence,conversation,bug \
  --owners manuel,alex \
  --status active \
  --intent short-term \
  --external-sources https://example.com/a,https://example.com/b \
  --summary "debugging notes" \
  --related-files backend/chat/api/register.go,web/src/store/api/chatApi.ts
```

Notes:
- `doc-type` values come from your workspace vocabulary (`ttmp/vocabulary.yaml`).
- If a doc type has a template at `ttmp/_templates/<docType>.md`, its body is rendered automatically.
- Unknown/other doc types are accepted and stored under a subdirectory named after the doc-type (frontmatter `DocType` is still set for filtering).

### 4.4.1 Move Documents Between Tickets

If a document was created under the wrong ticket, move it and rewrite its Ticket field:
```bash
docmgr doc move --doc ttmp/2025/12/01/MEN-4242-.../reference/01-chat-websocket-lifecycle.md \
  --dest-ticket MEN-5678 \
  --overwrite

# Optional: change the subdirectory under the destination ticket
docmgr doc move --doc path/to/doc.md --dest-ticket MEN-5678 --dest-dir reference/migrations
```

The command writes the destination copy with an updated Ticket frontmatter value and deletes the source after a successful move. Use `--overwrite` if a file with the same name already exists at the destination.

### 4.5 Guidelines

Guidelines provide structure and “what good looks like” for each doc type. They help new contributors produce consistent, reviewable docs.
```bash
# Human-readable guideline text
docmgr doc guidelines --doc-type design-doc

# Structured output (for tooling)
docmgr doc guidelines --doc-type design-doc --with-glaze-output --output json
```

Prints the guideline text for the given type. Files in `ttmp/_guidelines/` override embedded defaults.
See also: `docmgr help templates-and-guidelines` for how templates and guidelines fit together and how to customize them.

### 4.6 Update Metadata

Keep `Owners`, `Summary`, and `RelatedFiles` current. This makes search, review, and onboarding faster.
```bash
# Update a specific document
docmgr meta update --doc ttmp/YYYY/MM/DD/MEN-4242-.../index.md --field Owners --value "manuel,alex"

# Update all docs for a ticket (optionally filter by type)
docmgr meta update --ticket MEN-4242 --doc-type design-doc --field Topics --value "chat,backend"
```

Supported fields: Title, Ticket, Status, Topics, DocType, Intent, Owners, RelatedFiles, ExternalSources, Summary.

**Note on Status:** Status is vocabulary-guided (see `docmgr vocab list --category status`). Unknown values trigger warnings in `doctor` but don't fail operations.

### 4.7 List Tickets and Docs

Use listing commands to navigate by ticket. This is useful in reviews and when returning to paused work.
```bash
docmgr ticket list [--ticket MEN-4242]
docmgr doc list    --ticket MEN-4242
```

### 4.8 Search (Content + Metadata)

Search supports both content queries and metadata filters. Reverse lookups (`--file`, `--dir`) help you find docs from code paths; `--external-source` helps find docs tied to external references. Date filters surface recent activity.
```bash
# Content search
docmgr doc search --query "WebSocket" --ticket MEN-4242

# Metadata filters
docmgr doc search --ticket MEN-4242 --topics websocket,backend --doc-type design-doc

# Reverse lookup by file or directory
docmgr doc search --file backend/chat/api/register.go
docmgr doc search --dir  web/src/store/api/

# External source reference
docmgr doc search --external-source "https://example.com/ws-lifecycle"

# Date filters (relative and absolute)
docmgr doc search --updated-since "2 weeks ago" --ticket MEN-4242
docmgr doc search --since "last month" --until "today"

# File suggestions (heuristics: related files, git, ripgrep/grep)
docmgr doc search --ticket MEN-4242 --topics chat --files

# Relate changed files from git status (modified, staged, untracked)
docmgr doc relate --ticket MEN-4242 --suggest --from-git

# Apply changed files directly to the ticket index with notes
docmgr doc relate --ticket MEN-4242 --suggest --from-git --apply-suggestions
```

Relative date formats supported include: `today`, `yesterday`, `last week`, `this month`, `last month`, `2 weeks ago`, as well as ISO-like absolute dates (for example, `2025-01-01`).

### 4.9 Relate Files

Link code files to documentation for bidirectional navigation. Relating files enables powerful reverse lookup: find design docs from code files during review.

```bash
# Relate files to ticket index with explanatory notes
docmgr doc relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "backend/ws/manager.go:WebSocket lifecycle management"

# Suggest files from git changes
docmgr doc relate --ticket MEN-4242 --suggest --from-git

# Apply suggestions automatically
docmgr doc relate --ticket MEN-4242 --suggest --from-git --apply-suggestions

# Remove files
docmgr doc relate --ticket MEN-4242 --remove-files old/file.go
```

Notes explain WHY each file matters, turning file lists into navigation maps.

### 4.10 Changelog

Track progress and decisions in `changelog.md`:

```bash
# Simple entry
docmgr changelog update --ticket MEN-4242 --entry "Normalized API paths"

# With related files
docmgr changelog update --ticket MEN-4242 \
  --file-note "backend/api/register.go:Path normalization source"
```

### 4.11 Tasks

Manage concrete steps in `tasks.md`:

```bash
# Add tasks
docmgr task add --ticket MEN-4242 --text "Update API docs"

# Check off tasks
docmgr task check --ticket MEN-4242 --id 1,2

# List tasks
docmgr task list --ticket MEN-4242
```

### 4.12 Doctor (Validation)

Run `doctor` during development and reviews. It's a safety net to catch drift (stale docs), broken relationships (missing files), and inconsistent metadata (unknown vocabulary).

```bash
# Typical validation
docmgr doctor --all --stale-after 30 --fail-on error

# Ignore specific paths using glob patterns
docmgr doctor --ignore-glob "ttmp/*/design-doc/index.md" --fail-on warning

# Validate specific ticket
docmgr doctor --ticket MEN-4242
```

Doctor checks:

- Presence and validity of `index.md`
- Multiple `index.md` files under a single ticket
- Staleness via `LastUpdated` (configurable threshold)
- Required fields (Title, Ticket, Status, Topics)
- Unknown `Topics`, `DocType`, and `Intent` (validated against vocabulary)
- `RelatedFiles` existence on disk

`--fail-on` controls exit behavior for CI or pre-commit checks.

**Ignore configuration:**
- The command respects a `.docmgrignore` file at the repository root or at the docs root (`ttmp/`)
- Common patterns: `.git/`, `_templates/`, `_guidelines/`, `archive/`, date-based tickets like `2023-*/`

## 5. Testing the CLI (Dual Mode)

For docmgr contributors or power users: use a temporary root to avoid touching your repo during tests. The following matrix exercises both human-friendly output (default) and structured outputs (with `--with-glaze-output`).

```bash
# Build
go build -o /tmp/docmgr ./cmd/docmgr

# Create temp root and seed a workspace
ROOT=$(mktemp -d /tmp/docmgr-tests-XXXXXXXX)
/tmp/docmgr init --root "$ROOT"
/tmp/docmgr ticket create-ticket --ticket TST-1000 --title "Dual Mode Test" --topics demo,test --root "$ROOT"
/tmp/docmgr doc add  --ticket TST-1000 --doc-type design-doc --title "Design One" --root "$ROOT"

# list tickets
/tmp/docmgr ticket list --root "$ROOT"
/tmp/docmgr ticket list --root "$ROOT" --with-glaze-output --output json

# list docs
/tmp/docmgr doc list --root "$ROOT" --ticket TST-1000
/tmp/docmgr doc list --root "$ROOT" --ticket TST-1000 --with-glaze-output --output table

# status
/tmp/docmgr status --root "$ROOT"
/tmp/docmgr status --root "$ROOT" --with-glaze-output --output json

# guidelines
/tmp/docmgr doc guidelines --list --root "$ROOT"
/tmp/docmgr doc guidelines --doc-type design-doc --root "$ROOT"
/tmp/docmgr doc guidelines --doc-type design-doc --root "$ROOT" --with-glaze-output --output json

# tasks list
/tmp/docmgr task list --ticket TST-1000 --root "$ROOT"
/tmp/docmgr task list --ticket TST-1000 --root "$ROOT" --with-glaze-output --output csv

# search
/tmp/docmgr doc search --root "$ROOT" --ticket TST-1000 --query workspace
/tmp/docmgr doc search --root "$ROOT" --ticket TST-1000 --query workspace --with-glaze-output --output yaml

# cleanup (optional)
rm -rf "$ROOT"
```

Expected high-level behavior:

- Human mode prints readable one-liners or markdown text.
- Structured mode honors `--output` (json/yaml/csv/table) with the same data.
- Guidelines print the raw guideline content in human mode; list mode enumerates available types.
- Tasks list shows at least one seeded task from `init`.

---

## Related Documentation

For more detailed guides:

- **Daily usage:** `docmgr help how-to-use` — Complete tutorial with workflows, search, and power features
- **Repository setup:** `docmgr help how-to-setup` — Initialize workspace, configure vocabulary, customize templates
- **CI/automation:** `docmgr help ci-and-automation` — GitHub Actions, GitLab CI, hooks, Makefile, reporting
- **Templates:** `docmgr help templates-and-guidelines` — Customization guide

This CLI guide provides a quick reference. For step-by-step workflows and detailed explanations, see the tutorials above.
