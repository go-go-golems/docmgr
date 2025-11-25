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

### Optional: Enable Shell Completion

Typing long command names gets old fast. Enable the built-in completion once and every follow-up exercise is less error-prone.

```bash
# Generate completion script for your shell
docmgr completion bash
docmgr completion zsh
docmgr completion fish
docmgr completion powershell
```

- **Bash:** `docmgr completion bash | sudo tee /etc/bash_completion.d/docmgr >/dev/null` then restart your shell (or `source ~/.bashrc`).
- **Zsh:** `docmgr completion zsh > ~/.zfunc/_docmgr`, add `fpath+=~/.zfunc` to `.zshrc`, then `autoload -Uz compinit && compinit`.
- **Fish:** `docmgr completion fish > ~/.config/fish/completions/docmgr.fish`.
- **PowerShell:** `docmgr completion powershell | Out-String | Invoke-Expression` for the current session; persist by adding the command to your `$PROFILE`.

Once configured, `docmgr doc rel<TAB>` expands to `docmgr doc relate`, and flags like `--with-glaze-output` autocomplete automatically.

---

## 2. Key Concepts (Glossary) [BASIC]

These definitions show up throughout the tutorial. Skim them once so the later sections flow faster.

- **Ticket** — An identifier for a unit of work (like JIRA-123 or FEAT-042)
- **Ticket workspace** — Directory that contains every doc for a ticket
- **Docs root** — The `ttmp/` directory that stores all ticket workspaces
- **Frontmatter** — YAML metadata block at the top of each markdown doc
- **RelatedFiles** — Code references stored in frontmatter with notes explaining why a file matters
- **Vocabulary** — Optional list of topics/docTypes/intent used for validation warnings (not hard enforcement)

---

## 3. First-Time Setup [BASIC]

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

If you see the first output (with root, vocabulary paths), **you're already set up!** Skip to [Section 4](#4-create-your-first-ticket-basic).

### Initialize the Documentation Workspace

If not initialized, run:

```bash
docmgr init --seed-vocabulary
```

This creates:

```
ttmp/
├── vocabulary.yaml     # Defines topics/docTypes (used for validation warnings)
├── _templates/         # Document templates (used by 'docmgr doc add')
├── _guidelines/        # Writing guidelines (see 'docmgr doc guidelines')
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

## 4. Create Your First Ticket [BASIC]

```bash
docmgr ticket create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket
```

This creates `ttmp/YYYY/MM/DD/MEN-4242-.../` with `index.md`, `tasks.md`, and `changelog.md` under a standard structure. Use `--path-template` to override the relative directory layout (placeholders: `{{YYYY}}`, `{{MM}}`, `{{DD}}`, `{{DATE}}`, `{{TICKET}}`, `{{SLUG}}`, `{{TITLE}}`). If your repository doesn’t have a docs root yet (with `vocabulary.yaml`, `_templates/`, `_guidelines/`), run `docmgr init` first.

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

## 5. Add Documents [BASIC]

Add documents to organize your thinking:

```bash
docmgr doc add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
docmgr doc add --ticket MEN-4242 --doc-type reference --title "Chat WebSocket Lifecycle"
docmgr doc add --ticket MEN-4242 --doc-type playbook --title "Smoke Tests for Chat"
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

> **Tip:** Want structure guidance? Run: `docmgr doc guidelines --doc-type design-doc`

---

## 6. Search for Documents [BASIC]

Find docs by content or metadata:

```bash
# Full-text search
docmgr doc search --query "WebSocket"

# Filter by metadata
docmgr doc search --query "API" --topics backend --doc-type design-doc

# Find docs that reference a code file (reverse lookup)
docmgr doc search --file backend/api/register.go

# Find docs referencing any file in a directory
docmgr doc search --dir backend/api/
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

## 7. Managing Metadata [INTERMEDIATE]

Metadata (frontmatter) is the fuel that powers docmgr's search, validation, and automation. When the YAML is accurate, teammates can filter by status/topics and the `doctor` command can warn you before stale docs pile up. This section shows how to update fields safely without breaking formatting or forgetting timestamps.

docmgr provides the `meta update` command to modify frontmatter fields programmatically, ensuring valid YAML syntax, consistent formatting, and automated timestamp updates. This approach is particularly powerful for bulk operations (updating status across all design docs) and automation (syncing metadata from external systems), while keeping single-doc updates simple through command shortcuts.

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
docmgr doc add --ticket $TICKET --doc-type design-doc --title "Architecture"
docmgr doc add --ticket $TICKET --doc-type reference --title "API Contracts"
docmgr doc add --ticket $TICKET --doc-type playbook --title "Smoke Tests"
```

---

## 8. Relating Files to Docs [INTERMEDIATE]

> **Why this matters:** Relating files creates a breadcrumb trail between the prose (design/reference/playbook) and the code you just touched. Reviewers can jump from a file path to the design doc in one command, and future you can answer "Where is the spec for this file?" without spelunking.

Bidirectional linking between documentation and code is one of docmgr's most powerful features. By relating code files to docs with explanatory notes, you create a navigation map that answers two critical questions: "What's the design for this code file?" (code review context) and "Which code implements this design?" (implementation reference). The `docmgr doc relate` command manages these relationships in frontmatter, while `docmgr doc search --file` provides instant reverse lookup from any code file to its related documentation.

### The Workflow

**When to relate files:**
1. **During design** — Identify which code files will implement your design
2. **After implementation** — Link the key implementation files
3. **Before code review** — So reviewers can find context

### Basic Usage

```bash
# Relate files to ticket index with notes (repeat --file-note)
docmgr doc relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "backend/ws/manager.go:WebSocket lifecycle management"
```

### Relating with Notes (ALWAYS)

**Notes are required.** Always provide a note for each file when running `docmgr doc relate` or `docmgr changelog`. Notes turn file lists into navigation maps that explain why a file is linked. The legacy `\-\-files` flag was removed to enforce this behavior; use repeated `--file-note "path:reason"` entries instead.

```bash
docmgr doc relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "backend/ws/manager.go:WebSocket lifecycle management"
```

### Advanced patterns

**Structured RelatedFiles (with notes)**

```yaml
RelatedFiles:
  - path: backend/api/register.go
    note: Registers API routes (normalization logic)
  - path: backend/ws/manager.go
    note: WebSocket lifecycle management
```

**Reverse lookup (during code review)**

```bash
docmgr doc search --file backend/api/register.go
```

**Subdocument-first linking**

```bash
docmgr doc relate --doc ttmp/MEN-4242/design-doc/01-path-normalization-strategy.md \
  --file-note "backend/api/register.go:Normalization entrypoint and router setup"
```

- Keep `index.md` as an overview; relate most files to the specific design/reference/playbook document that explains them.
- Aim for 3-7 `RelatedFiles` entries per ticket; more than 10 usually means you're listing everything.
- Run the quick checklist: relate files → update the Summary → `docmgr doctor --ticket YOUR-123 --stale-after 30`.

## 9. Recording Changes [BASIC]

Changelog entries are your running field notes. They document *when* something changed, *why* it changed, and *where* to look in the code. Keeping them current prevents "why was this file touched?" detective work weeks later.

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
docmgr doc relate --ticket MEN-4242 \
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

**Note:** `RelatedFiles` in YAML supports both `Path`/`Note` and `path`/`note` formats. Use `docmgr doc relate` commands to maintain consistency.

---

## 10. Managing Tasks [BASIC]

Tasks capture the atomic steps required to finish a ticket. They pair nicely with changelog entries: tasks show the *plan*, changelog entries show the *history*. Keep the list short, actionable, and owned so anyone opening the ticket knows what remains.

Track concrete steps in `tasks.md`:

```bash
# Add tasks
docmgr task add --ticket MEN-4242 --text "Update API docs for /v2"
docmgr task add --ticket MEN-4242 --text "Add WebSocket lifecycle diagram"

# Check off completed tasks
docmgr task check --ticket MEN-4242 --id 1,2

# List tasks
docmgr task list --ticket MEN-4242
```

### Edit, uncheck, or remove tasks

```bash
# Fix the text without touching other lines
docmgr task edit --ticket MEN-4242 --id 2 --text "Align frontend routes with backend"

# Re-open work that needs more attention
docmgr task uncheck --ticket MEN-4242 --id 2

# Remove noisy items (docmgr keeps the rest of the list intact)
docmgr task remove --ticket MEN-4242 --id 3
```

Output shows checkboxes: `[x]` for done, `[ ]` for pending.

**When to use which doc:**
- Keep *tasks* focused on actionable steps ("Add monitoring dashboard").
- Use the *changelog* to describe what actually changed ("Added grafana dashboard; linked api/metrics.go").
- Reference tasks inside the changelog when the status changes (e.g., "Task #4 complete").

---

## 11. Closing Tickets [INTERMEDIATE]

Closing a ticket is more than flipping Status to `complete`. The `ticket close` command updates status, writes a final changelog line, and warns if open tasks remain so reviewers can trust that the documentation reflects production reality.

When you've finished work on a ticket, use `ticket close` to atomically update status, changelog, and metadata. Ticket status must match the shared vocabulary; see [Status Vocabulary & Transitions](#status-vocabulary--transitions) if you're unsure which value to pick.

```bash
# Close with defaults (status=complete)
docmgr ticket close --ticket MEN-4242

# Close with custom status
docmgr ticket close --ticket MEN-4242 --status archived

# Close with custom changelog message
docmgr ticket close --ticket MEN-4242 --changelog-entry "All requirements implemented, ready for production"

# Close and update intent
docmgr ticket close --ticket MEN-4242 --intent long-term
```

**What `ticket close` does:**
- Updates Status (default: `complete`, override with `--status`)
- Optionally updates Intent (via `--intent`)
- Appends a changelog entry (default: "Ticket closed")
- Updates LastUpdated timestamp
- Warns if tasks aren't all done (doesn't fail)

**Structured output for automation:**
```bash
# Get machine-readable results
docmgr ticket close --ticket MEN-4242 --with-glaze-output --output json

# Example output:
{
  "ticket": "MEN-4242",
  "all_tasks_done": true,
  "open_tasks": 0,
  "done_tasks": 5,
  "status": "complete",
  "operations": {
    "status_updated": true,
    "intent_updated": false,
    "changelog_updated": true
  }
}
```

### Status Vocabulary & Transitions

Status values are vocabulary-guided (teams can customize). Default values keep work flowing in a predictable direction:
- `draft` — Initial draft state
- `active` — Active work in progress
- `review` — Ready for review
- `complete` — Work completed
- `archived` — Archived/completed work

Discover the current list (including custom entries) with:

```bash
docmgr vocab list --category status --with-glaze-output --output yaml
```

Suggested transitions (not enforced):
- `draft` → `active` → `review` → `complete` → `archived`
- `review` → `active` (send back for fixes)
- `complete` → `active` (reopen; unusual, call it out in the changelog)

`docmgr doctor` warns (does not fail) if a ticket uses a status value that's not part of the vocabulary and lists the valid values plus the `docmgr vocab list --category status` command to help you correct or extend the list.

Add custom status values with:

```bash
docmgr vocab add --category status --slug on-hold --description "Work paused"
```

**Pro tip:** When you check off the last task, `task check` suggests running `ticket close`:
```bash
docmgr task check --ticket MEN-4242 --id 3
# Output: 💡 All tasks complete! Consider closing the ticket: docmgr ticket close --ticket MEN-4242
```

---

## 12. Validation with Doctor [INTERMEDIATE]

`docmgr doctor` is your safety net. It catches broken frontmatter, missing files, stale docs, and unknown vocabulary before a reviewer—or worse, a user—finds them. Run it early and often, especially before you open a PR or hand off work.

Check for problems before they bite you:

```bash
# Validate all docs
docmgr doctor --all --stale-after 30 --fail-on error

# Validate specific ticket
docmgr doctor --ticket MEN-4242
```

**What doctor checks:**
- ✅ Missing or invalid frontmatter (all markdown files)
- ✅ Unknown topics/doc-types/status (warns, doesn't fail)
- ✅ Missing Note on RelatedFiles entries (warns)
- ✅ Missing files in RelatedFiles
- ✅ Stale docs (older than --stale-after days)

**Common warnings:**
- Unknown topic (not in vocabulary.yaml) — Add it with `docmgr vocab add`
- Missing file in RelatedFiles — Fix path or remove entry
- Stale doc — Update content or adjust --stale-after threshold
- Invalid frontmatter — Fix YAML syntax errors

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

You know: init, create, add, search, metadata, relate, changelog, tasks, close, validation.

**What's next?**
- **Need automation?** → Continue to [Part 3](#part-3-power-user-features-⚡)
- **Done for now?** → Start working! Refer back when you need advanced features.

---

# Part 3: Power User Features ⚡

**[For automation, scripting, and CI]**

This part covers advanced features for power users and automation.

---

## 13. Structured Output (Glaze) [ADVANCED]

Automation-minded teams often need machine-readable output. Glaze lets every docmgr command emit JSON/YAML/CSV/TSV without writing new scripts. Once you understand the field contracts, you can feed docmgr data straight into CI pipelines, dashboards, or quick one-off shell loops.

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
docmgr doc search --updated-since "60 days ago" --with-glaze-output --output json | \
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
    docmgr ticket create-ticket --ticket $TICKET --title "Feature $i" --topics backend
    docmgr doc add --ticket $TICKET --doc-type design-doc --title "Design $i"
done

# Update all docs of a type
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value complete
```

---

## 14. CI Integration Examples [ADVANCED]

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
	@docmgr status --with-glaze-output --output yaml
	@echo ""
	@docmgr doc search --updated-since "7 days ago"
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

## 15. List and Status Commands [BASIC]

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

Sample human output (default):

```
Docs root: `/home/you/projects/chat-app/ttmp`
Paths are relative to this root.

## Tickets (2)

### MEN-4242 — Chat Persistence
- Status: **active**
- Topics: backend, chat
- Tasks: 2 open / 5 done
- Updated: 2025-11-19 14:20
- Path: `2025/11/19/MEN-4242-chat-persistence`
```

`docmgr list docs` mirrors the same style, grouped by ticket with per-document bullet summaries (doc type, status, topics, updated, path).

**Common usecases:**
- `list tickets` — See all your tickets at a glance
- `list docs --ticket T` — What docs exist for this ticket?
- `status` — Health check: how many tickets, docs, any stale?
- `status --summary-only` — Just the totals, no per-ticket detail

---

## 16. Iterate and Maintain [INTERMEDIATE]

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
# Add files as you implement (notes required)
docmgr doc relate --ticket MEN-4242 \
  --file-note "new/file.go:What this file does"

# Remove files if refactored away
docmgr doc relate --ticket MEN-4242 --remove-files old/file.go
```

**Update index.md body regularly:**
- Overview (goal, scope, constraints)
- Status (one-line current state)
- Next steps (short checklist)
- Key Links (important docs and code)

**Keep tasks.md and changelog.md current:**
```bash
# Check off tasks as you complete them
docmgr task check --ticket MEN-4242 --id 1,2

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
docmgr doc guidelines --doc-type design-doc
docmgr doc guidelines --doc-type reference
```

---

## 17. Root Discovery and Configuration [ADVANCED]

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

## 18. Vocabulary Management [INTERMEDIATE]

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

## 19. Numeric Prefixes [INTERMEDIATE]

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

## 20. Command Aliasing [REFERENCE]

The CLI ships friendly aliases so you can type the style that matches your brain—or the screenshots in older docs. Each pair executes the exact same command and produces the same output.

| Primary command | Alias | When to use |
|-----------------|-------|-------------|
| `docmgr init` | `docmgr workspace init` | Prefer the `workspace` form when scripting multiple roots |
| `docmgr doctor` | `docmgr workspace doctor` | Helpful when thinking in terms of workspace health |
| `docmgr status` | `docmgr workspace status` | Same data; workspace prefix clarifies the scope |
| `docmgr doc list` | `docmgr list docs` | Use whichever ordering matches your muscle memory |

Check the aliases that are available in your build:

```bash
docmgr help workspace init
docmgr help workspace doctor
docmgr help workspace status
docmgr help doc list
```

If you document a workflow, mention both names the first time (e.g., "`docmgr doc list` (alias: `docmgr list docs`)") so new users are never blocked by mismatched wording.

---

## 21. Tips and Best Practices

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

## Appendix A: Troubleshooting Common Errors

### "Error: no changes specified"

- **What it means:** You ran `docmgr meta update` or `docmgr doc relate` without providing any fields/flags that would change the file.
- **Common causes:** Forgetting `--field Summary --value ...` or running `doc relate` without any `--file-note` entries.
- **Fix it:** Re-run the command with at least one change, e.g. `docmgr meta update --ticket MEN-4242 --field Status --value review`.

### `"Unknown topic: <slug>"`

- **What it means:** The topic in your document isn't listed in `ttmp/vocabulary.yaml`, so validation produced a warning.
- **Common causes:** New feature areas, typos, or collaborators using different casing.
- **Fix it:** Either add the topic via `docmgr vocab add --category topics --slug your-topic` **or** update the doc's `Topics` list to an existing slug.

### "Must specify --doc or --ticket"

- **What it means:** Commands like `doc relate`, `meta update`, and `changelog update` need to know which file to touch.
- **Common causes:** Running a command from deep inside the repo without flags, or copying an example that omitted `--ticket`.
- **Fix it:** Add `--ticket YOUR-123` to target the ticket's `index.md`, or pass `--doc path/to/doc.md` for a subdocument.

### "open <path>: no such file or directory"

- **What it means:** docmgr tried to read/write a file that doesn't exist yet.
- **Common causes:** Typos in `--doc` paths, moving files without updating `RelatedFiles`, or running commands from the wrong directory/root.
- **Fix it:** Double-check the path relative to the docs root, run `docmgr status --summary-only` to confirm the root, and re-run from the repository root for consistent discovery.

### "warning: document stale (45 days old)"

- **What it means:** `docmgr doctor` found a doc whose `LastUpdated` timestamp is older than the `--stale-after` threshold.
- **Common causes:** Work paused, docs written once and never revisited, or automated updates not touching the content body.
- **Fix it:** Review the content, make necessary updates (even adding a short status note counts), then rerun `docmgr doctor --ticket YOUR-123 --stale-after 30` to verify the warning cleared. Adjust `--stale-after` only if the cadence is truly different.

---

## Appendix B: Quick Reference

### Common Commands

```bash
# Setup (once per repo)
docmgr init --seed-vocabulary

# Create ticket
docmgr ticket create-ticket --ticket YOUR-123 --title "..." --topics ...

# Add docs
docmgr doc add --ticket YOUR-123 --doc-type TYPE --title "..."

# Search
docmgr doc search --query "..."
docmgr doc search --file path/to/file.go

# Relate files
docmgr doc relate --ticket YOUR-123 --file-note "path:note" --file-note "path2:note2"

# Validate
docmgr doctor --ticket YOUR-123

# Tasks
docmgr task add --ticket YOUR-123 --text "..."
docmgr task check --ticket YOUR-123 --id 1,2

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
# Human vs Glaze output

`docmgr doctor` now mirrors other dual-mode commands:

```bash
# Human-readable report (default)
docmgr doctor --ticket MEN-4242

# Structured output (default JSON)
docmgr doctor --ticket MEN-4242 --with-glaze-output --output yaml
```

The human report groups findings per ticket with Markdown bullets, while `--with-glaze-output` switches back to Glaze rows (table/json/yaml/csv) for automation.
