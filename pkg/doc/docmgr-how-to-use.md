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
│    → Covers: relating files, tasks, changelogs, validation, working on existing tickets │
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
- **Vocabulary** — Optional list of topics/docTypes/intent used for validation warnings (not hard enforcement). View with `docmgr vocab list` to familiarize yourself with repository conventions.

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
ttmp/YYYY/MM/DD/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/
├── index.md        # Ticket overview (you're here)
├── tasks.md        # Todo list
├── changelog.md    # History of changes
├── design-doc/     # Created when you add a design-doc
├── reference/      # Created when you add a reference doc
├── playbook/       # Created when you add a playbook
└── <doc-type>/     # Any other doc-type creates its own subdir
```

> **Note:** Tickets are stored under `ttmp/YYYY/MM/DD/` using the date the ticket was created. This keeps workspaces organized chronologically. You can override the layout with `--path-template` if needed.

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

### Best Practice: Relate Files When Creating Documents

**⚠️ Important:** Every time you create or modify a document, relate the source files you used or referenced. This creates bidirectional links that make your documentation discoverable from code and vice versa.

**Immediately after creating a document:**

```bash
# After creating a design-doc, relate the files it describes
docmgr doc add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
docmgr doc relate --ticket MEN-4242 --doc-type design-doc \
  --file-note "backend/api/register.go:Implements path normalization logic" \
  --file-note "backend/api/router.go:Route configuration"

# After creating a reference doc, relate the files it documents
docmgr doc add --ticket MEN-4242 --doc-type reference --title "Chat WebSocket Lifecycle"
docmgr doc relate --ticket MEN-4242 --doc-type reference \
  --file-note "backend/ws/manager.go:WebSocket connection lifecycle"
```

**Why this matters:**
- Code reviewers can find design docs instantly: `docmgr doc search --file backend/api/register.go`
- Future developers can answer "Where's the spec for this file?" without searching
- Documentation stays connected to implementation, preventing drift

**Workflow tip:** Make relating files part of your document creation routine. If you're analyzing code to write a doc, relate those files immediately. See [Relating Files to Docs](#8-relating-files-to-docs-intermediate) for detailed guidance.

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

### Interpreting Results

```bash
# No results
docmgr doc search --query "nonexistent-term"
# Output: (no results)

# Multiple results with snippets (default human output)
docmgr doc search --query "API"
# Example:
# 2025/11/19/MEN-4242-chat-persistence/reference/02-api-contracts.md — Chat API Contracts [MEN-4242]
# ... "Normalized API paths for chat endpoints" ...
#
# 2025/11/20/MEN-4300-auth/reference/01-auth-api.md — Auth API [MEN-4300]
# ... "All API requests require JWT tokens" ...

# Narrow with metadata filters
docmgr doc search --query "API" --topics backend --doc-type design-doc

# Script/CI-friendly JSON
docmgr doc search --query "API" --with-glaze-output --output json
```

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

> **Important:** Always use `docmgr doc relate` (not `docmgr relate`). The `doc relate` command does not support a `--doc-type` flag; use `--ticket` to target the ticket index or `--doc PATH` to target a specific document.

### The Workflow

**When to relate files:**
1. **When creating a document** — Relate the source files you're analyzing or documenting (best practice: do this immediately after `docmgr doc add`)
2. **When modifying a document** — Relate any new files you reference or analyze
3. **During design** — Identify which code files will implement your design
4. **After implementation** — Link the key implementation files
5. **Before code review** — So reviewers can find context

**Key principle:** Every document creation or modification should include relating the relevant source files. This creates bidirectional links that make documentation discoverable from code and prevents documentation drift.

### Basic Usage

```bash
# Relate files to ticket index with notes (repeat --file-note)
docmgr doc relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "backend/ws/manager.go:WebSocket lifecycle management"
```

### Relating with Notes (ALWAYS)

**Notes are required.** Always provide a note for each file when running `docmgr doc relate` or `docmgr changelog`. Notes turn file lists into navigation maps that explain why a file is linked. The legacy `\-\-files` flag was removed to enforce this behavior; use repeated `--file-note "path:reason"` entries instead.

For examples, see Basic Usage above.

#### File-note format

> **Format:** `--file-note "FILE_PATH:DESCRIPTIVE_NOTE"`
>
> The colon (`:`) separates the file path from the note.
>
> Examples:
>
> - ✅ `--file-note "backend/api/register.go:Registers API routes"`
> - ✅ `--file-note "web/src/store/api/chatApi.ts:Frontend integration"`
> - ❌ `--file-note "backend/api/register.go - Registers API routes"` (wrong delimiter)

**Re-running with the same notes:** If you call `docmgr doc relate` again with identical `--file-note` entries (and nothing else), docmgr now emits a warning row like `status=noop` with a reason such as `file-note entries matched existing notes` instead of failing. Add a new note, change the note text, or use `--remove-files` to make a real change.

### When to Relate to Ticket vs. Subdocument

> Use this quick decision guide to keep `index.md` concise and implementation details in the right place.

**Relate to ticket index (`--ticket`) when:**
- You’re establishing high-level context or overview links
- The file is core to understanding the ticket’s scope
- You want a minimal map from the ticket entry point

Example:
```bash
docmgr doc relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Entry point and router setup"
```

**Relate to a subdocument (`--doc PATH`) when:**
- The file implements a specific design/reference/playbook
- You want tightly scoped relationships per document type
- Reviewers should land on a focused doc, not the ticket index

Example:
```bash
DOC="ttmp/2025/11/19/MEN-4242-chat-persistence/design-doc/01-path-normalization-strategy.md"
docmgr doc relate --doc "$DOC" \
  --file-note "backend/api/register.go:Normalization entrypoint and router setup"
```

**Path handling tips:**
- **Use `--ticket` when possible** — Simplest approach, automatically targets the ticket's `index.md` or works with `--doc-type`
- **For subdocuments, use paths relative to workspace root** — Paths should start from your workspace root (where `.ttmp.yaml` lives), not from `ttmp/`
- **Use shell variables for long paths** — Store the document path in a variable to avoid typos:
  ```bash
  DOC="docmgr/ttmp/2025/11/19/MEN-4242-chat-persistence/reference/01-carapace-analysis.md"
  docmgr doc relate --doc "$DOC" --file-note "carapace/storage.go:Storage system"
  ```
- **Common mistake:** Using paths relative to `ttmp/` instead of workspace root — docmgr expects paths from the workspace root where the config file lives

### Advanced patterns and best practices

**Reverse lookup:** See [Search for Documents](#6-search-for-documents-basic) for reverse lookup examples.

**Subdocument-first linking**

```bash
docmgr doc relate --doc ttmp/MEN-4242/design-doc/01-path-normalization-strategy.md \
  --file-note "backend/api/register.go:Normalization entrypoint and router setup"
```

- Keep `index.md` as an overview; relate most files to the specific design/reference/playbook document that explains them.
- Aim for 3-7 `RelatedFiles` entries per ticket; more than 10 usually means you're listing everything.
- Run the quick checklist: relate files → update the Summary → `docmgr doctor --ticket YOUR-123 --stale-after 30`.

**Best practice: Relate files immediately after document creation/modification**

Make relating files part of your document workflow:

1. **Create the document** → `docmgr doc add --ticket T --doc-type TYPE --title "Title"`
2. **Immediately relate source files** → `docmgr doc relate --ticket T --doc-type TYPE --file-note "path:reason"`
3. **Update changelog** → `docmgr changelog update --ticket T --entry "Created doc, related files"`
4. **Validate** → `docmgr doctor --ticket T`

This ensures documentation stays connected to code from the start, preventing "orphaned" docs that reference files but aren't discoverable from those files.

## 9. Recording Changes [BASIC]

Changelog entries are your running field notes. They document *when* something changed, *why* it changed, and *where* to look in the code. Keeping them current prevents "why was this file touched?" detective work weeks later.

Track progress in `changelog.md`:

```bash
# With related files and notes
docmgr changelog update --ticket MEN-4242 \
  --file-note "backend/api/register.go:Path normalization source"
```

Changelogs are dated automatically. Keep entries short — mention what changed and link relevant files.

**Best practice:** When you add a changelog entry, always include file notes and also relate the exact files you changed to the relevant subdocument(s) (design-doc/reference/playbook). Keep `index.md` as a concise map that links to those subdocuments. Then validate.

**The workflow:**

1) Relate files with notes (see [Relating Files to Docs](#8-relating-files-to-docs-intermediate))

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

### Quick Hands-on

```bash
# Edit a task inline (safe updates to just the target line)
docmgr task edit --ticket MEN-4242 --id 1 --text "Write API docs for /v2"

# Remove a task while keeping the rest intact
docmgr task remove --ticket MEN-4242 --id 1
```

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

See [Vocabulary Management](#16-vocabulary-management-intermediate) for listing and customizing status values.

Suggested transitions (not enforced):
- `draft` → `active` → `review` → `complete` → `archived`
- `review` → `active` (send back for fixes)
- `complete` → `active` (reopen; unusual, call it out in the changelog)

`docmgr doctor` warns (does not fail) if a ticket uses a status value that's not part of the vocabulary and lists the valid values plus the `docmgr vocab list --category status` command to help you correct or extend the list.

To customize status values, see [Vocabulary Management](#16-vocabulary-management-intermediate).

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

# Capture diagnostics for CI (JSON)
docmgr doctor --all --diagnostics-json diagnostics.json --fail-on warning
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

## 12.5. Working on Existing Tickets [INTERMEDIATE]

> **When to use this:** You're taking over an existing ticket workspace and need to get oriented quickly while keeping metadata aligned.

Whenever you inherit ticket `<TICKET-ID>` inside repository `<REPO-PATH>`, follow this playbook to get oriented, understand the current context, and keep the workspace in a compliant state. Every section builds on the previous one so you can ramp up quickly without dropping important metadata.

### Step 0: Confirm the Workspace and Refresh docmgr Basics

Before touching ticket files, confirm that you can run docmgr at the repository root. This ensures the help system is available and that the ticket metadata already exists.

```bash
cd <REPO-PATH>

docmgr help how-to-use
docmgr ticket list --ticket <TICKET-ID>
docmgr doc list --ticket <TICKET-ID>
docmgr task list --ticket <TICKET-ID>
```

> **Note:** Both `docmgr ticket list` and `docmgr list tickets` work (they're aliases). Similarly, `docmgr doc list` and `docmgr list docs` are both valid forms.

**Review the repository vocabulary:** Familiarize yourself with the team's shared vocabulary by viewing the available topics, doc types, intent values, and status values. This helps you understand the repository's conventions and use consistent metadata when creating or updating documents.

```bash
# View all vocabulary entries
docmgr vocab list

# View specific categories
docmgr vocab list --category topics
docmgr vocab list --category docTypes
docmgr vocab list --category intent
docmgr vocab list --category status
```

If any command fails, fix the repository setup (see `docmgr help how-to-setup`) before proceeding.

### Step 1: Review the Ticket Source Material

Read all existing documentation in order so you understand why the ticket exists and what has already been attempted.

1. Open the ticket index (typically `ttmp/YYYY/MM/DD/<TICKET-ID>-.../index.md`) for the canonical summary.
2. Inspect implementation diaries under `log/`, reading entries chronologically to catch historical context.
3. Review the current `tasks.md` and `changelog.md` to see outstanding work and completed changes.
4. Skim any background docs referenced by `docmgr doc list --ticket <TICKET-ID>` and note prerequisites or dependencies.

### Step 2: Start with the Highest-Priority Tasks

Always begin with the next unchecked task so progress stays orderly. Use the CLI to see and update task status as you work.

```bash
docmgr task list --ticket <TICKET-ID>
docmgr task check --ticket <TICKET-ID> --id <TASK-ID>
```

Update the list as soon as you complete a meaningful unit of work and capture any new subtasks that emerge.

### Step 3: Keep Files and the Changelog in Sync

Every modification must be traceable. Relate files immediately after edits and log the change so future maintainers know what happened and why.

```bash
docmgr doc relate --ticket <TICKET-ID> \
  --file-note "/ABS/PATH/TO/FILE:Why this file matters right now"

docmgr changelog update --ticket <TICKET-ID> \
  --entry "What changed and why" \
  --file-note "/ABS/PATH/TO/FILE:Reason"
```

> **Important:** Always use `docmgr doc relate` (not `docmgr relate`). The `doc relate` command does not support `--doc-type` flag; use `--ticket` or `--doc PATH` instead.

Use absolute paths for clarity, and group related changes into a single changelog entry with multiple `--file-note` values if needed.

### Step 4: Maintain an Implementation Diary

After each significant step, jot down what you tried, what succeeded or failed, and what to do next. Append to the active diary in `log/` or create a new note under `log/various/` if no diary exists yet. These entries become the institutional memory for the ticket.

### Step 5: Capture Repo-Specific Intelligence

Document any local setup, build commands, unusual comparison steps, or environment switches that apply to this repository. Add these notes to the ticket workspace (often `various/` or a dedicated reference doc) so the next person can reproduce your environment without guesswork.

### Step 6: Track Known Issues and Immediate Focus

Keep a running list of blockers, gaps, and temporary workarounds. Update it whenever you discover a new risk so planning conversations have an up-to-date source of truth. Mention follow-up tasks if they cannot be addressed immediately.

### Step 7: Close the Ticket When Done

When all tasks are complete and work is ready for review or deployment, use `ticket close` to atomically update status, changelog, and timestamps. Status values follow the shared status vocabulary (`draft → active → review → complete → archived`, with `review → active` and occasional `complete → active` re-openings), so confirm the exact slug before closing.

```bash
# Check if all tasks are done
docmgr task list --ticket <TICKET-ID>

# Close with defaults (status=complete)
docmgr ticket close --ticket <TICKET-ID>

# Or close with custom status
docmgr ticket close --ticket <TICKET-ID> --status review --changelog-entry "Implementation complete, ready for review"
```

**What `ticket close` does:**
- Updates Status (default: `complete`, override with `--status`)
- Optionally updates Intent (via `--intent`)
- Appends a changelog entry
- Updates LastUpdated timestamp
- Warns if tasks aren't all done (doesn't fail)

**Pro tip:** When you check off the last task with `docmgr task check`, it automatically suggests running `ticket close`.

**Status cheat sheet:**

```bash
# Inspect the current status vocabulary (including custom entries)
docmgr vocab list --category status --with-glaze-output --output table
```

`docmgr doctor` warns—but will not fail—if a ticket uses a status value outside the vocabulary. Update the ticket or extend the vocabulary with `docmgr vocab add --category status --slug <slug> --description "..."`

### Quick Reference: docmgr Helpers at a Glance

Re-run the status and metadata commands whenever you context-switch to ensure nothing drifted:

```bash
docmgr status --summary-only
docmgr ticket list
docmgr meta update --ticket <TICKET-ID> --field Status --value active
docmgr ticket close --ticket <TICKET-ID>  # When done
```

**Where to go next:** After finishing this checklist, return to the task list, confirm priorities with the ticket owner, and continue iterating through tasks, file relations, changelog entries, and diary updates. When all tasks are complete, use `docmgr ticket close` to finalize the work. This loop keeps every ticket workspace healthy and auditable.

---

✅ **Milestone: You Can Now Use All Core Features!**

You know: init, create, add, search, metadata, relate, changelog, tasks, close, validation, working on existing tickets.

**What's next?**
- **Need automation?** → Continue to [Part 3](#part-3-power-user-features-⚡)
- **Done for now?** → Start working! Refer back when you need advanced features.

---

# Part 3: Power User Features ⚡

**[For automation, scripting, and CI]**

This part covers advanced features for power users and automation.

---

## 13. Automation and CI [ADVANCED]

Most docmgr commands support structured output for automation. Add `--with-glaze-output --output json` to any command to get machine-readable results. This enables CI validation, bulk operations, and reporting dashboards without custom parsers.

**Quick examples:**

```bash
# JSON output for scripts
docmgr list tickets --with-glaze-output --output json

# Validate in CI with proper exit code
docmgr doctor --all --fail-on error || exit 1

# Extract paths for bulk operations
docmgr list docs --ticket MEN-4242 --with-glaze-output --select path
```

**For complete automation patterns, CI integration examples, and field contracts, see:**

```bash
docmgr help ci-automation
```

That guide covers GitHub Actions, GitLab CI, pre-commit hooks, Makefile integration, bulk operation patterns, and stable field names for scripting.

---

✅ **Milestone: You Can Now Automate Everything!**

You know: structured output basics and where to find complete automation patterns.

**What's next?**
- Read [Part 4: Reference](#part-4-reference-📖) for list/status commands and vocabulary
- Or close this doc and start working!

---

# Part 4: Reference 📖

**[Look up as needed — advanced topics]**

---

## 14. List and Status Commands [BASIC]

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
- `list tickets` or `ticket list` — See all your tickets at a glance (both forms work)
- `list docs --ticket T` or `doc list --ticket T` — What docs exist for this ticket? (both forms work)
- `status` — Health check: how many tickets, docs, any stale?
- `status --summary-only` — Just the totals, no per-ticket detail

> **Note:** Command aliases: `docmgr ticket list` is an alias for `docmgr ticket tickets`, and both `docmgr list docs` and `docmgr doc list` work identically. Use whichever form feels more natural.

---

## 15. Iterate and Maintain [INTERMEDIATE]

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
See [Validation with Doctor](#12-validation-with-doctor-intermediate) for commands and options.

**Consult guidelines when writing:**
```bash
docmgr doc guidelines --doc-type design-doc
docmgr doc guidelines --doc-type reference
```

**Advanced maintenance:** For layout fixes, config debugging, and multi-repo setups, see `docmgr help advanced-workflows`.

---

## 16. Vocabulary Management [INTERMEDIATE]

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

**Verify changes and categories:**
```bash
# Supported categories include: topics, status, docTypes

# Before adding a doc type (expect no 'til' entry)
docmgr vocab list --category docTypes | grep -E "^docTypes: til" || true

# Add and verify
docmgr vocab add --category docTypes --slug til --description "Today I Learned entries"
docmgr vocab list --category docTypes | grep -E "^docTypes: til"
```

**Remember:** Unknown topics/doc-types are allowed. They just trigger warnings in `docmgr doctor`. The vocabulary is for documentation and team consistency, not enforcement.

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

**4. Use tasks to track progress**
- Add tasks using `docmgr task add`
- Check off tasks as you complete them
- Add changelog entries after significant changes

### Shell Gotchas

**Parentheses in directory names:**
```bash
# If ticket name has parens, quote/escape:
cd "ttmp/MEN-XXXX-name-\(with-parens\)"
```

Don't use ! in strings because it confuses shells like zsh. 
Avoid : in file notes themselves since it confuses the argument parser (which uses it to separate the path from the note).
