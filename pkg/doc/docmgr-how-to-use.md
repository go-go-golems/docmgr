---
Title: Tutorial — Using docmgr to Drive a Ticket Workflow
Slug: how-to-use
Short: Complete guide to creating tickets, adding docs, searching, and managing documentation workflows with docmgr.
Topics:
- docmgr
- tutorial
- workflow
- documentation
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Tutorial — Using docmgr to Drive a Ticket Workflow

**Based on UX findings from 10-round heated debrief with 7 participants.**

## Quick Navigation

┌──────────────────────────────────────────────────────────────────────────────┐
│ **Choose your path:**                                                        │
│                                                                              │
│ 📚 **New to docmgr?**                                                       │
│    → Read [Part 1: Essentials](#part-1-essentials-📚) (10 minutes)          │
│    → You'll be ready to create tickets and docs                             │
│                                                                              │
│ ⚡ **Need automation/CI?**                                                  │
│    → Jump to [Part 3: Power Features](#part-3-power-user-features-⚡)       │
│    → Covers: JSON output, scripting, CI integration                         │
│                                                                              │
│ 🔍 **Looking for specific command?**                                        │
│    → Use: `docmgr COMMAND --help`                                           │
│    → Or search this doc with Ctrl+F                                         │
│                                                                              │
│ 🔧 **Need specific workflow?**                                              │
│    → See: [Part 2: Everyday Workflows](#part-2-everyday-workflows-🔧)       │
│    → Covers: relating files, tasks, changelogs, validation                  │
└──────────────────────────────────────────────────────────────────────────────┘

> **For power users:** docmgr supports structured output (JSON/CSV), CI integration, and bulk operations. See [Part 3](#part-3-power-user-features-⚡) for automation examples.

---

## Overview

docmgr transforms ad-hoc markdown documentation into a structured, searchable knowledge base by organizing docs into ticket workspaces with consistent metadata. This structure enables powerful features like bidirectional code-to-doc linking, full-text search with metadata filtering, automated validation, and scriptable output for CI/CD integration. The tool enforces just enough structure to make documentation discoverable and maintainable without imposing rigid constraints on your workflow.

**Core value proposition:**
- **Structure without rigidity** — Consistent metadata and organization, but unknown topics/doc-types are allowed
- **Bidirectional linking** — Find docs from code files (code review) and code from docs (implementation)
- **Automated quality** — Validation catches broken links, missing files, and stale docs before they accumulate
- **Automation-ready** — Stable JSON/CSV output for scripts, CI/CD, and reporting

If you're used to unstructured markdown files, docmgr adds metadata and command overhead but pays back through discoverability, team collaboration, and quality enforcement. Break-even is typically 10-20 tickets or when multiple people need to navigate the documentation.

**Working discipline:** 
- Use `docmgr` commands to update frontmatter (metadata)
- Write document body content (markdown) in your editor
- Keep `tasks.md` and `changelog.md` current via CLI commands for consistency

---

## Key Concepts (Glossary)

Quick definitions for terms used in this tutorial:

- **Ticket** — An identifier for a unit of work (like JIRA-123 or FEAT-042)
- **Ticket workspace** — A directory containing all docs related to a ticket
- **Docs root** — The `ttmp/` directory that contains all ticket workspaces
- **Frontmatter** — YAML metadata at the top of markdown files (Title, Topics, Status, etc.)
- **RelatedFiles** — Code files referenced in a doc's frontmatter with notes explaining why
- **Vocabulary** — Defines valid topics/docTypes/intent for validation (not enforcement)

---

# Part 1: Essentials 📚

**[10 minute read — START HERE]**

This part covers everything you need to start using docmgr.

---

## 1. Prerequisites [BASIC]

**Required:**
- `docmgr` available on PATH (install it first)
- A directory to work in

**Recommended:**
- Git repository (makes RelatedFiles paths more meaningful, but not required)

> **Note:** docmgr works without Git. It just uses the file system. Git is only useful for making file paths in RelatedFiles more meaningful to your team.

---

## 2. First-Time Setup [BASIC]

**⚠️ IMPORTANT:** Run this ONCE per repository before creating your first ticket.

### Check if Already Initialized

Before running init, check if your workspace is already set up:

```bash
docmgr status --summary-only
```

**If initialized, you'll see:**
```
root=/path/to/ttmp config=/path/.ttmp.yaml vocabulary=/path/vocabulary.yaml tickets=0 stale=0 docs=0
```

**If NOT initialized, you'll see:**
```
Error: root directory does not exist: /path/ttmp
```

If you see the first output (with root, vocabulary paths), **you're already set up!** Skip to [Section 3](#3-create-your-first-ticket-basic).

### Initialize the Documentation Workspace

If not initialized, run:

```bash
docmgr init --seed-vocabulary
```

This creates:

```
ttmp/
├── vocabulary.yaml     # Defines topics/docTypes (used for validation warnings)
├── _templates/         # Document templates (used by 'docmgr add')
├── _guidelines/        # Writing guidelines (see 'docmgr guidelines')
└── .docmgrignore       # Files to exclude from validation
```

**Verify it worked:**

```bash
# Check initialization status
docmgr status --summary-only

# View seeded vocabulary 
docmgr vocab list
```

Expected output from `docmgr status`:
```
root=/your/path/ttmp vocabulary=/your/path/ttmp/vocabulary.yaml tickets=0 docs=0
```

Expected output from `docmgr vocab list`:
```
topics: chat — Chat backend and frontend surfaces
topics: backend — Backend services
topics: websocket — WebSocket lifecycle & events
etc...
...
```

If you see these, initialization succeeded!

**What's vocabulary.yaml?**
- Defines valid topics (backend, frontend, etc.) and doc types (design-doc, reference, etc.)
- Used by `docmgr doctor` to warn about unknown values (NOT enforced — you can use any topics)
- `--seed-vocabulary` pre-fills it with common defaults
- Add custom entries with: `docmgr vocab add --category topics --slug your-topic`

**Note:** Running `docmgr init` multiple times is safe (idempotent) — it won't overwrite existing files unless you use `--force`.

> **Advanced setup:** For multi-repo setups or custom paths, see `docmgr help how-to-setup`.

---

## 3. Create Your First Ticket [BASIC]

```bash
docmgr create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket
```

**What this creates:**

```
ttmp/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/
├── index.md        # Ticket overview (you're here)
├── tasks.md        # Todo list
├── changelog.md    # History of changes
├── design-doc/     # Created when you add a design-doc
├── reference/      # Created when you add a reference doc
├── playbook/       # Created when you add a playbook
└── <doc-type>/     # Any other doc-type creates its own subdir
```

**Understanding index.md:**

The `index.md` file is your ticket's single entry point. It:
- Summarizes what the ticket does (one-line Summary in frontmatter + Overview section in body)
- Points to key docs and code via `RelatedFiles` in frontmatter
- Serves as anchor for validation checks (`docmgr doctor`)

**Best practice:** 
- Keep index.md body content concise (~50 lines of markdown)
- Update frontmatter via `docmgr meta update` commands
- Write Overview, Status, Next Steps in the body content (below frontmatter)
- Prefer a subdocument-first linking pattern: relate most implementation files to focused subdocuments (design-doc/reference/playbook), and have `index.md` link to those subdocuments instead of listing every file directly.
- When relating files (anywhere), always include notes (`--file-note "path:why-this-file-matters"`); file notes are required in our workflow.

> **Smart Default:** Documents you add will automatically inherit topics (`chat,backend,websocket`), owners, and status from the ticket. No need to repeat them!

---

## 4. Add Documents [BASIC]

Add documents to organize your thinking:

```bash
docmgr add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
docmgr add --ticket MEN-4242 --doc-type reference --title "Chat WebSocket Lifecycle"
docmgr add --ticket MEN-4242 --doc-type playbook --title "Smoke Tests for Chat"
```

**What happens:**
- Each doc is created from a template in `_templates/`
- Frontmatter fields (Title, Ticket, Topics) are auto-filled
- Files get numeric prefixes (01-, 02-, 03-) to keep them ordered
- Topics/owners/status inherited from ticket (no repetition!)
- The file is stored under a subdirectory named exactly after its doc-type (e.g., `design-doc/`, `reference/`, `playbook/`, or your custom type)

**Common doc types:**
- `design-doc` — Architecture and design decisions
- `reference` — API contracts, data schemas, how things work
- `playbook` — Test procedures, operational runbooks
- Custom types are allowed and create their own subdirectory (e.g., `til/`, `analysis/`)

> **Tip:** Want structure guidance? Run: `docmgr guidelines --doc-type design-doc`

---

## 5. Search for Documents [BASIC]

Find docs by content or metadata:

```bash
# Full-text search
docmgr search --query "WebSocket"

# Filter by metadata
docmgr search --query "API" --topics backend --doc-type design-doc

# Find docs that reference a code file (reverse lookup)
docmgr search --file backend/api/register.go

# Find docs referencing any file in a directory
docmgr search --dir backend/api/
```

**Common usecases:**
- **Discovery:** "What have we documented about authentication?"
- **Code review:** "What's the design for this file I'm reviewing?"
- **Refactoring:** "Which docs mention this directory I'm changing?"

Search is fast (< 100ms even with 200+ docs) and case-insensitive.

---

✅ **Milestone: You're Ready to Use docmgr!**

You now know how to:
- Initialize a repository (once)
- Create tickets
- Add documents
- Search for docs

**What's next?**
- **Need changelogs, tasks, or validation?** → Continue to [Part 2](#part-2-everyday-workflows-🔧)
- **Want automation and scripting?** → Jump to [Part 3](#part-3-power-user-features-⚡)
- **Ready to work?** → Start creating docs! Come back here when you need more features.

---

# Part 2: Everyday Workflows 🔧

**[Read as needed — workflow reference]**

This part covers common workflows beyond the basics.

---

## 6. Managing Metadata [INTERMEDIATE]

Metadata (frontmatter) defines how docs are discovered, filtered, and validated. docmgr provides the `meta update` command to modify frontmatter fields programmatically, ensuring valid YAML syntax, consistent formatting, and automated timestamp updates. This approach is particularly powerful for bulk operations (updating status across all design docs) and automation (syncing metadata from external systems), while keeping single-doc updates simple through command shortcuts.

### Update Specific Document

```bash
# Update index.md (default target when using --ticket; no --doc needed)
docmgr meta update --ticket MEN-4242 --field Summary --value "Unify API paths"
docmgr meta update --ticket MEN-4242 --field Status --value review
docmgr meta update --ticket MEN-4242 --field Owners --value "manuel,alex"

# Update a specific subdocument (use --doc with explicit path)
DOC="ttmp/MEN-4242-normalize-chat-api/reference/03-foobar.md"
docmgr meta update --doc "$DOC" --field Summary --value "Unify API paths"
```

Note: When you omit --doc and provide --ticket, commands like `meta update` target the ticket’s `index.md` by default.

### Bulk Updates Across Documents

For updating multiple docs at once, use `--ticket` and `--doc-type`:

```bash
# Update status on all design-docs in a ticket
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value review

# Update all docs in a ticket (any type)
docmgr meta update --ticket MEN-4242 \
    --field Owners --value "manuel,alex"
```

### When to Use Each Approach

| Your Task | Command Pattern | Example |
|-----------|----------------|---------|
| Update 1 field on 1 doc | `--doc PATH --field X --value Y` | Update summary on design doc |
| Update 1 field on all design-docs | `--ticket T --doc-type design-doc --field X` | Mark all designs as reviewed |
| Update 1 field on all docs | `--ticket T --field X --value Y` | Update all owners |
| Automation/scripts | Any of above in shell loops | CI validation |

**Philosophy:** Use `docmgr meta update` for frontmatter because it ensures:
- Valid YAML syntax
- Consistent formatting  
- Proper validation (warns about unknown topics)
- Automated LastUpdated timestamps
- Trackable changes through the tool

**Write your document body content** (markdown below the frontmatter) in your editor as usual.

### Shell Patterns for Multiple Docs

```bash
# Use shell variable to avoid repeating ticket
TICKET=MEN-4242
docmgr add --ticket $TICKET --doc-type design-doc --title "Architecture"
docmgr add --ticket $TICKET --doc-type reference --title "API Contracts"
docmgr add --ticket $TICKET --doc-type playbook --title "Smoke Tests"
```

---

## 7. Relating Files to Docs [INTERMEDIATE]

Bidirectional linking between documentation and code is one of docmgr's most powerful features. By relating code files to docs with explanatory notes, you create a navigation map that answers two critical questions: "What's the design for this code file?" (code review context) and "Which code implements this design?" (implementation reference). The `docmgr relate` command manages these relationships in frontmatter, while `docmgr search --file` provides instant reverse lookup from any code file to its related documentation.

### The Workflow

**When to relate files:**
1. **During design** — Identify which code files will implement your design
2. **After implementation** — Link the key implementation files
3. **Before code review** — So reviewers can find context

### Basic Usage

```bash
# Relate files to ticket index with notes (repeat --file-note)
docmgr relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "backend/ws/manager.go:WebSocket lifecycle management"
```

### Relating with Notes (ALWAYS)

**Notes are required.** Always provide a note for each file when running `docmgr relate` or `docmgr changelog`. Notes turn file lists into navigation maps that explain why a file is linked.

```bash
docmgr relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "backend/ws/manager.go:WebSocket lifecycle management"
```

**Result in frontmatter:**
```yaml
RelatedFiles:
    - path: backend/api/register.go
      note: Registers API routes (normalization logic)
    - path: backend/ws/manager.go
      note: WebSocket lifecycle management
```

Now readers know WHY each file matters and WHERE to start reading.

### Reverse Lookup (Code Review Superpower)

Find design context from code files:

```bash
# During code review: "What's the design for this file?"
$ docmgr search --file backend/api/register.go

MEN-4242/design-doc/01-path-normalization.md — Path Normalization [MEN-4242] ::
  file=backend/api/register.go note=Registers API routes
```

**Saves 2-3 minutes per file review** by surfacing design context instantly.

### Choosing Files to Relate

**DO relate:**
- ✅ Key implementation files (handlers, services, core logic)
- ✅ Files reviewers need to understand the feature
- ✅ Files that would impact docs if refactored

**DON'T relate:**
- ❌ Every file (creates noise)
- ❌ Generated files or build artifacts
- ❌ Test files (unless documenting test strategy)

**Rule of thumb:** 3-7 files per ticket. More than 10? You're probably over-relating.

### Writing Good Notes

**Good notes explain WHY a file matters:**

```yaml
# ❌ Bad (states the obvious)
- path: auth/handler.go
  note: Auth handler

# ✅ Good (explains role and responsibilities)
- path: auth/handler.go
  note: Implements login, logout, refresh endpoints; validates credentials
```

**Template:** `[What it does]; [Key responsibilities or functions]`

### Subdocument-first linking

Prefer relating most files to the specific subdocument that explains them (design-doc/reference/playbook) rather than directly to `index.md`. Keep `index.md` minimal and use it to reference those subdocuments.

- Why: Subdocuments keep context close to code and prevent `index.md` bloat.
- How: Add `RelatedFiles` on the subdocument; in `index.md`, add a short link to that subdocument.
- Exception: Truly ticket-wide anchors (e.g., top-level diagrams or scripts) can live on `index.md`.

Example:
```bash
# Relate implementation files to the design doc (preferred)
docmgr relate --doc ttmp/MEN-4242/design-doc/01-path-normalization-strategy.md \
  --file-note "backend/api/register.go:Normalization entrypoint and router setup"
```

### Index Playbook (Quick Checklist)

**Keep your ticket index clean and current with this 3-step workflow:**

1) **Relate files with notes** to the ticket index

```bash
docmgr relate --ticket MEN-4242 \
  --file-note \"backend/api/register.go:Registers API routes (normalization logic)\" \
  --file-note \"web/src/store/api/chatApi.ts:Frontend API integration\"
```

2) **Update the one-line Summary** in frontmatter

```bash
docmgr meta update --ticket MEN-4242 \
  --field Summary \
  --value "MEN-4242: normalize API paths; update WS lifecycle; docs + tests."
```

3) **Validate** to catch issues

```bash
docmgr doctor --ticket MEN-4242 --stale-after 30 --fail-on error
```

Run this checklist whenever you've made significant progress on a ticket.

---

## 8. Recording Changes [BASIC]

Track progress in `changelog.md`:

```bash
# With related files and notes
docmgr changelog update --ticket MEN-4242 \
  --file-note "backend/api/register.go:Path normalization source"
```

Changelogs are dated automatically. Keep entries short — mention what changed and link relevant files.

### Changelog Hygiene (Always Link Files and Provide Notes)

**Best practice:** When you add a changelog entry, always include file notes and also relate the exact files you changed to the relevant subdocument(s) (design-doc/reference/playbook). Keep `index.md` as a concise map that links to those subdocuments. Then validate.

**The workflow:**

1) Relate files with notes (to ticket index):

```bash
docmgr relate --ticket MEN-4242 \
  --file-note \"backend/api/register.go:Path normalization source\" \
  --file-note \"web/src/store/api/chatApi.ts:Frontend integration\"
```

2) Add changelog entry (mention linked files):

```bash
docmgr changelog update --ticket MEN-4242 \
  --entry "Normalized API paths; linked backend/api/register.go and chatApi.ts with notes."
```

3) Validate:

```bash
docmgr doctor --ticket MEN-4242 --stale-after 30 --fail-on error
```

**Why this matters:** Changelogs are a record of WHAT changed. RelatedFiles document WHICH code implements it. Together they give complete context.

**Note:** `RelatedFiles` in YAML supports both `Path`/`Note` and `path`/`note` formats. Use `docmgr relate` commands to maintain consistency.

---

## 9. Managing Tasks [BASIC]

Track concrete steps in `tasks.md`:

```bash
# Add tasks
docmgr tasks add --ticket MEN-4242 --text "Update API docs for /v2"
docmgr tasks add --ticket MEN-4242 --text "Add WebSocket lifecycle diagram"

# Check off completed tasks
docmgr tasks check --ticket MEN-4242 --id 1,2

# List tasks
docmgr tasks list --ticket MEN-4242
```

Output shows checkboxes: `[x]` for done, `[ ]` for pending.

---

## 10. Validation with Doctor [INTERMEDIATE]

Check for problems before they bite you:

```bash
# Validate all docs
docmgr doctor --all --stale-after 30 --fail-on error

# Validate specific ticket
docmgr doctor --ticket MEN-4242
```

**What doctor checks:**
- ✅ Missing or invalid frontmatter
- ✅ Unknown topics/doc-types (warns, doesn't fail)
- ✅ Missing Note on RelatedFiles entries (warns)
- ✅ Missing files in RelatedFiles
- ✅ Stale docs (older than --stale-after days)

**Common warnings:**
- Unknown topic (not in vocabulary.yaml) — Add it with `docmgr vocab add`
- Missing file in RelatedFiles — Fix path or remove entry
- Stale doc — Update content or adjust --stale-after threshold

### Suppressing Noise with .docmgrignore

`docmgr init` creates `ttmp/.docmgrignore`. Add patterns to ignore:

```
.git/
_templates/
_guidelines/
archive/
2023-*
2024-*
```

Doctor automatically respects these patterns.

---

✅ **Milestone: You Can Now Use All Core Features!**

You know: init, create, add, search, metadata, relate, changelog, tasks, validation.

**What's next?**
- **Need automation?** → Continue to [Part 3](#part-3-power-user-features-⚡)
- **Done for now?** → Start working! Refer back when you need advanced features.

---

# Part 3: Power User Features ⚡

**[For automation, scripting, and CI]**

This part covers advanced features for power users and automation.

---

## 11. Structured Output (Glaze) [ADVANCED]

Every docmgr command that produces output can render it in multiple structured formats (JSON, CSV, YAML, TSV) through the Glaze framework. This design decouples the command's business logic from its output format, enabling the same command to serve both human users (with readable tables and text) and automation scripts (with parseable JSON or CSV). The stable field contracts ensure your scripts won't break when docmgr is updated, making it safe to build CI/CD integrations, reporting dashboards, and bulk operation scripts on top of docmgr.

### Quick Examples

```bash
# JSON output (for scripts)
docmgr list tickets --with-glaze-output --output json

# CSV output (for spreadsheets)
docmgr list docs --with-glaze-output --output csv > docs.csv

# Extract just paths (one per line)
docmgr list docs --ticket MEN-4242 --with-glaze-output --select path

# Validate in CI with proper exit code
docmgr doctor --all --fail-on error || exit 1
```

### Available Output Formats

- `json` — Valid JSON, parseable
- `csv` — Comma-separated (for spreadsheets)
- `tsv` — Tab-separated
- `yaml` — YAML format
- `table` — ASCII table (human-readable)

### Stable Field Names (API Contract)

Use these with `--fields`, `--filter`, `--select`:

**Tickets:**
- `ticket`, `title`, `status`, `topics`, `path`, `last_updated`

**Docs:**
- `ticket`, `doc_type`, `title`, `status`, `topics`, `path`, `last_updated`

**Tasks:**
- `index`, `checked`, `text`, `file`

**Vocabulary:**
- `category`, `slug`, `description`

### Field Selection Examples

```bash
# Paths only (newline-separated)
docmgr list docs --ticket MEN-4242 --with-glaze-output --select path

# Custom columns (CSV)
docmgr list docs --with-glaze-output --output csv \
  --fields doc_type,title,path

# Templated output
docmgr list docs --ticket MEN-4242 --with-glaze-output \
  --select-template '{{.doc_type}}: {{.title}}' --select _0
```

### Automation Patterns

**Pattern 1: Find and update stale docs**

```bash
# Find docs older than 60 days, mark for review
docmgr search --updated-since "60 days ago" --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value "needs-review"
  done
```

**Pattern 2: CI validation**

```bash
#!/bin/bash
# .github/workflows/validate-docs.yml

if ! docmgr doctor --all --stale-after 14 --fail-on error; then
  echo "ERROR: Documentation validation failed"
  # Get list of issues
  docmgr doctor --all --with-glaze-output --output json | \
    jq -r '.[] | select(.issue != "none") | "\(.path): \(.message)"'
  exit 1
fi
```

**Pattern 3: Weekly doc report**

```bash
# Generate report of doc activity
docmgr status --stale-after 7 --with-glaze-output --output json | \
  jq -r '.docs[] | select(.stale) | "\(.ticket): \(.title) (stale \(.days_since_update) days)"'
```

**Pattern 4: Bulk operations**

```bash
# Create similar tickets
for i in {1..5}; do
    TICKET=PROJ-00$i
    docmgr create-ticket --ticket $TICKET --title "Feature $i" --topics backend
    docmgr add --ticket $TICKET --doc-type design-doc --title "Design $i"
done

# Update all docs of a type
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value complete
```

---

## 12. CI Integration Examples [ADVANCED]

### GitHub Actions

```yaml
name: Validate Docs

on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install docmgr
        run: go install github.com/go-go-golems/docmgr@latest
      - name: Validate docs
        run: docmgr doctor --all --fail-on error
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check docs aren't broken
if ! docmgr doctor --all --fail-on error; then
  echo "ERROR: Doc validation failed. Fix issues or use 'git commit --no-verify'"
  exit 1
fi
```

### Makefile Integration

```makefile
.PHONY: docs-validate docs-report

docs-validate:
	docmgr doctor --all --stale-after 30 --fail-on error

docs-report:
	@docmgr status --with-glaze-output --output table
	@echo ""
	@docmgr search --updated-since "7 days ago"
```

---

✅ **Milestone: You Can Now Automate Everything!**

You know: structured output, CI integration, bulk operations, scripting patterns.

**What's next?**
- Read [Part 4: Reference](#part-4-reference-📖) for advanced topics
- Or close this doc and build your automation!

---

# Part 4: Reference 📖

**[Look up as needed — advanced topics]**

---

## 13. List and Status Commands [BASIC]

Explore what's in your workspace:

```bash
# List all tickets
docmgr list tickets

# List all docs in a ticket
docmgr list docs --ticket MEN-4242

# List just one ticket (useful for checking)
docmgr list tickets --ticket MEN-4242

# Check overall workspace status
docmgr status
docmgr status --summary-only
docmgr status --stale-after 30

# Structured output for scripting
docmgr list tickets --with-glaze-output --output json
docmgr list docs --ticket MEN-4242 --with-glaze-output --output csv
```

**Common usecases:**
- `list tickets` — See all your tickets at a glance
- `list docs --ticket T` — What docs exist for this ticket?
- `status` — Health check: how many tickets, docs, any stale?
- `status --summary-only` — Just the totals, no per-ticket detail

---

## 14. Iterate and Maintain [INTERMEDIATE]

**Keep your documentation workspace healthy over time:**

### Regular Maintenance Tasks

**Update metadata as work progresses:**
```bash
# Keep Summary current
docmgr meta update --doc ttmp/TICKET-.../index.md --field Summary --value "Current state..."

# Update Status when transitioning
docmgr meta update --ticket MEN-4242 --field Status --value complete

# Keep Owners accurate
docmgr meta update --ticket MEN-4242 --field Owners --value "current,team,members"
```

**Maintain RelatedFiles:**
```bash
# Add files as you implement
docmgr relate --ticket MEN-4242 --files new/file.go \
  --file-note "new/file.go:What this file does"

# Remove files if refactored away
docmgr relate --ticket MEN-4242 --remove-files old/file.go
```

**Update index.md body regularly:**
- Overview (goal, scope, constraints)
- Status (one-line current state)
- Next steps (short checklist)
- Key Links (important docs and code)

**Keep tasks.md and changelog.md current:**
```bash
# Check off tasks as you complete them
docmgr tasks check --ticket MEN-4242 --id 1,2

# Add changelog entries after significant changes
docmgr changelog update --ticket MEN-4242 --entry "Completed authentication flow"
```

**Run validation periodically:**
```bash
# Check for issues
docmgr doctor --ticket MEN-4242

# Or check entire workspace
docmgr doctor --all --stale-after 30
```

**Consult guidelines when writing:**
```bash
docmgr guidelines --doc-type design-doc
docmgr guidelines --doc-type reference
```

---

## 15. Root Discovery and Configuration [ADVANCED]

### How docmgr Finds the Docs Root

1. Looks for `.ttmp.yaml` walking up from CWD
2. If found, uses `root` field from that file
3. If not found, defaults to `<cwd>/ttmp` or `<git-root>/ttmp` if in Git repo

### Custom Configuration (.ttmp.yaml)

Create at repository root:

```yaml
root: ttmp
vocabulary: ttmp/vocabulary.yaml
```

Useful for:
- Multi-repo setups
- Custom root directory names
- Centralizing vocabulary across repos

**Most users don't need this** — defaults work for typical setups.

---

## 15. Vocabulary Management [INTERMEDIATE]

Vocabulary defines valid topics, doc types, and intents (used for warnings, not enforcement).

```bash
# List vocabulary
docmgr vocab list

# Add custom topic
docmgr vocab add --category topics --slug frontend \
  --description "Frontend code and components"

# Add custom doc type
docmgr vocab add --category docTypes --slug til \
  --description "Today I Learned entries"
```

**Remember:** Unknown topics/doc-types are allowed. They just trigger warnings in `docmgr doctor`. The vocabulary is for documentation and team consistency, not enforcement.

---

## 16. Numeric Prefixes [INTERMEDIATE]

**What happens automatically:**
- New docs get numeric prefixes: `01-`, `02-`, `03-`
- Keeps files ordered in directory listings
- Switches to 3 digits after 99 files
- Ticket-root files (`index.md`, `tasks.md`, `changelog.md`) are exempt

**Resequencing:**

```bash
# If you delete files and want to renumber
docmgr renumber --ticket MEN-4242
```

This updates prefixes and fixes internal links between docs.

**Doctor warns if files are missing prefixes** (you can suppress with `.docmgrignore`).

---

## 17. Tips and Best Practices

### Workflow Recommendations

**1. Relate files with notes (always)**
- Not just which files, but WHY they matter
- Helps code reviewers and future developers

**2. Keep index.md concise (~50 lines)**
- One-line Summary in frontmatter
- Brief Overview section
- Links to key docs and files
- Current status and next steps

**3. Update changelog regularly**
- After significant changes
- Link files you modified
- Keep entries short

**4. Use doctor in CI**
- Catch broken links and missing files
- Adjust --stale-after to your team's pace
- Use .docmgrignore for false positives

### Shell Gotchas

**Parentheses in directory names:**
```bash
# If ticket name has parens, quote/escape:
cd "ttmp/MEN-XXXX-name-\(with-parens\)"
```

**Tab completion:**
- Most shells support tab completion for paths
- Helps with long ticket directory names

---

## Appendix: Quick Reference

### Common Commands

```bash
# Setup (once per repo)
docmgr init --seed-vocabulary

# Create ticket
docmgr create-ticket --ticket YOUR-123 --title "..." --topics ...

# Add docs
docmgr add --ticket YOUR-123 --doc-type TYPE --title "..."

# Search
docmgr search --query "..."
docmgr search --file path/to/file.go

# Relate files
docmgr relate --ticket YOUR-123 --file-note "path:note" --file-note "path2:note2"

# Validate
docmgr doctor --ticket YOUR-123

# Tasks
docmgr tasks add --ticket YOUR-123 --text "..."
docmgr tasks check --ticket YOUR-123 --id 1,2

# Changelog
docmgr changelog update --ticket YOUR-123 --entry "..."

# Automation
docmgr list docs --with-glaze-output --output json
```

### Field Names for Metadata

**Common fields in frontmatter:**
- `Title` — Document title
- `Ticket` — Ticket identifier
- `Status` — active, draft, review, complete
- `Topics` — Array of topics (inherited from ticket)
- `DocType` — design-doc, reference, playbook, etc.
- `Intent` — long-term, temporary, etc.
- `Owners` — Array of owner names
- `RelatedFiles` — Array of paths (with optional notes)
- `ExternalSources` — Array of URLs
- `Summary` — One-line description