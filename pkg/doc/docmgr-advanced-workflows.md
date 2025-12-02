---
Title: Advanced Workflows — Import Files, Root Configuration, and Maintenance
Slug: advanced-workflows
Short: Advanced docmgr workflows for importing external files, configuring custom roots, and maintaining documentation workspaces.
Topics:
- docmgr
- advanced
- workflows
- configuration
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Advanced Workflows — Import Files, Root Configuration, and Maintenance

## Overview

This guide covers advanced docmgr workflows that most users encounter after they've mastered the basics. These features solve specific problems: importing external research artifacts, debugging configuration issues, reorganizing legacy ticket structures, and managing multi-repo documentation setups.

**When to read this:**
- You need to import PDFs, LLM outputs, or external specs into tickets
- docmgr seems to be looking at the wrong root directory
- You're migrating legacy tickets or setting up multi-repo documentation
- You want to understand how `.ttmp.yaml` configuration works

**Prerequisites:**
- Comfortable with basic docmgr workflows (see `docmgr help how-to-use`)
- Understanding of ticket workspaces and frontmatter

---

## 1. Import External Files

Third-party specs, customer notes, or LLM brainstorms rarely live inside docmgr at first. `docmgr import file` captures those artifacts in your ticket so reviewers can see exactly which research you referenced without digging through drives or chat logs.

### Typical workflows

```bash
# Import a markdown brainstorm from /tmp
docmgr import file --ticket MEN-4242 --file /tmp/chat-llm-notes.md

# Import a PDF spec and rename it on the way in
docmgr import file --ticket MEN-4242 \
  --file ~/Downloads/websocket-spec.pdf \
  --name websocket-spec-v2
```

- Files land under `ttmp/.../MEN-4242--.../sources/local/`. The `--name` flag changes the basename while keeping the original extension.
- `.meta/sources.yaml` records every import (type, original path, timestamp) so you can refresh or replace them later.
- `index.md` automatically gains an `ExternalSources` entry such as `local:websocket-spec-v2.pdf`, making the attachment searchable via `docmgr doc search --external-source`.

### Keep imports discoverable

1. **Link within docs:** Mention `sources/local/...` inside the design/reference doc that consumes the material so reviewers know where supporting evidence lives.
2. **Relate downstream files:** After implementing the feature, run `docmgr doc relate` on the design doc or ticket index, noting that it depends on the imported file.
3. **Refresh stale artifacts:** If the upstream doc changes, re-run `docmgr import file ... --name existing-name` (or delete + import) and mention the update in `changelog.md`.

When importing large bundles (LLM dumps, screenshots, transcripts), add a short `sources/README.md` summarizing each file and why it matters. Treat it like an appendix so newcomers can skim before opening the raw assets.

---

## 2. Root Discovery and Configuration

### How docmgr Finds the Docs Root

docmgr searches for the docs root in this order:

1. `--root` flag (if provided)
2. `DOCMGR_CONFIG` environment variable pointing to `.ttmp.yaml`
3. `.ttmp.yaml` file (walking up from current directory)
4. Git repository root + `/ttmp` (if in a Git repo)
5. Current directory + `/ttmp` (fallback)

### Custom Configuration (.ttmp.yaml)

Create at repository root:

```yaml
root: ttmp
vocabulary: ttmp/vocabulary.yaml
```

**Use cases:**
- Multi-repo setups where different repos share a centralized docs root
- Custom root directory names (e.g., `docs/` instead of `ttmp/`)
- Centralizing vocabulary across multiple repositories
- Monorepo setups with multiple documentation workspaces

**Most users don't need this** — defaults work for typical single-repo setups.

### Debugging Configuration Issues

When docmgr behaves as if it's reading the wrong root or vocabulary, use `docmgr config show` to see the entire resolution process:

```bash
# Show config resolution from the current directory
docmgr config show

# Point at another docs root (useful in scripts/CI)
docmgr config show --root /tmp/doc-workspace/ttmp

# Combine with a custom config file
DOCMGR_CONFIG=../configs/docs/.ttmp.yaml docmgr config show
```

The output lists every source checked (CLI flag, `DOCMGR_CONFIG`, `.ttmp.yaml` while walking up, Git-root fallback) and clearly marks the source that was actually used, along with the effective `root`/`vocabulary` paths.

**Common issues:**
- Running commands from nested subdirectories without absolute paths in `.ttmp.yaml`
- Multiple `.ttmp.yaml` files in parent directories (first one found wins)
- Expecting Git root discovery when not in a Git repository
- Environment variables overriding local config files

---

## 3. Layout Fix — Reorganize Doc Directories

Legacy tickets or manual file moves can leave documents sitting at the ticket root, which breaks the "one folder per doc-type" structure and makes templates hard to reuse. `docmgr doc layout-fix` rescans the ticket, moves each markdown file into the directory that matches its `DocType`, and updates internal links in one shot.

```bash
# Preview what would move (recommended)
docmgr doc layout-fix --ticket MEN-4242 --dry-run

# Apply the fix
docmgr doc layout-fix --ticket MEN-4242
```

**What it does:**
- Reads the `DocType` frontmatter from each markdown file
- Moves files into `<doc-type>/` subdirectories (e.g., `design-doc/`, `reference/`)
- Skips root control files (`index.md`, `README.md`, `tasks.md`, `changelog.md`)
- Creates missing doc-type directories automatically
- Updates all markdown links within the ticket to reflect new paths

**When to use:**
- Migrating old tickets that predate the doc-type directory structure
- After manually moving files and realizing links are broken
- Cleaning up tickets where docs ended up in the wrong directories

**Caution:** When **not** using `--ticket`, it walks the entire docs root—handy after large migrations, but start with a dry run so the move list is readable.

---

## 4. Multi-Repo and Monorepo Setups

### Scenario 1: Shared Documentation Root

Multiple repositories sharing a centralized documentation workspace:

```
/home/team/docs/
├── .ttmp.yaml          # Shared config
├── ttmp/
│   ├── vocabulary.yaml
│   ├── PROJ-001-.../   # From repo A
│   └── PROJ-002-.../   # From repo B
```

**In each repository's `.ttmp.yaml`:**
```yaml
root: /home/team/docs/ttmp
vocabulary: /home/team/docs/ttmp/vocabulary.yaml
```

### Scenario 2: Monorepo with Multiple Workspaces

```
monorepo/
├── .ttmp.yaml          # Points to shared vocabulary
├── backend/
│   └── ttmp/           # Backend docs
├── frontend/
│   └── ttmp/           # Frontend docs
└── shared-vocabulary.yaml
```

**Root `.ttmp.yaml`:**
```yaml
vocabulary: shared-vocabulary.yaml
```

**When working in `backend/`:**
```bash
cd backend
docmgr status  # Uses backend/ttmp automatically
```

**When working in `frontend/`:**
```bash
cd frontend
docmgr status  # Uses frontend/ttmp automatically
```

### Scenario 3: Custom Root Name

If you prefer `docs/` instead of `ttmp/`:

```yaml
root: docs
vocabulary: docs/vocabulary.yaml
```

All docmgr commands will now look for `docs/` instead of `ttmp/`.

---

## 5. Best Practices for Advanced Setups

**Configuration:**
- Use absolute paths in `.ttmp.yaml` when working across multiple repos
- Keep vocabulary centralized if teams share terminology
- Document your setup in the repository README
- Test `docmgr config show` from different directories to verify discovery

**Imports:**
- Add a `sources/README.md` when importing multiple files
- Link imported files in the docs that reference them
- Use `--name` to give imports meaningful names
- Update `.meta/sources.yaml` if you manually refresh files

**Maintenance:**
- Run `layout-fix --dry-run` before applying to preview changes
- Use `config show` when debugging unexpected behavior
- Keep `.ttmp.yaml` in version control
- Document custom configurations in your team wiki

---

## 6. Numeric Prefixes and Renumber

Numeric prefixes keep long directories readable (`01-overview.md`, `02-api.md`, …). docmgr adds them automatically when you scaffold new docs, but deletes, renames, and bulk moves can knock the ordering out of sync.

**What happens automatically:**
- New docs get numeric prefixes: `01-`, `02-`, `03-`
- Keeps files ordered in directory listings
- Switches to 3 digits after 99 files
- Ticket-root files (`index.md`, `tasks.md`, `changelog.md`) are exempt

### When to run `docmgr doc renumber`

Run the renumber command whenever you:
- Delete or insert docs mid-sequence and want clean numbering again.
- Import older tickets whose files never had prefixes.
- Rearrange files manually and need docmgr to update intra-ticket links.

```bash
# Resequence every doc under a ticket and fix references
docmgr doc renumber --ticket MEN-4242
```

`docmgr doc renumber` walks every doc-type directory, renames files to the next sequential prefix (switching to 3 digits once you exceed 99 files), and updates all markdown links inside the ticket so nothing breaks. Commit or stash unrelated changes first—the command edits every file that still references old paths.

> No `--dry-run` flag yet, so lean on Git to preview the diff if you need to approve the rename list.

**Doctor warns if files are missing prefixes** (you can suppress with `.docmgrignore`).

---

## See Also

- `docmgr help how-to-use` — Core tutorial
- `docmgr help ci-automation` — CI/CD integration patterns
- `docmgr help how-to-setup` — Initial setup guide

