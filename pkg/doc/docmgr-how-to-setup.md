---
Title: Tutorial — Setting up docmgr in a Codebase
Slug: how-to-setup
Short: Initialize docmgr in your repository, configure vocabulary, customize templates, and manage the documentation workspace.
Topics:
- docmgr
- setup
- configuration
- repository
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Tutorial — Setting up docmgr in a Codebase

**For repository maintainers and team leads setting up docmgr.**

## Quick Navigation

┌──────────────────────────────────────────────────────────────────────────────┐
│ **Choose your path:**                                                        │
│                                                                              │
│ 📚 **First-time setup?**                                                    │
│    → Read [Part 1: Repository Setup](#part-1-repository-setup-📚)           │
│    → Initialize, configure, seed vocabulary (15 minutes)                    │
│                                                                              │
│ 🔧 **Need CI/automation?**                                                  │
│    → See separate guide: `docmgr help ci-and-automation`                    │
│    → Covers: GitHub Actions, GitLab, hooks, Makefile, reporting            │
│                                                                              │
│ 📖 **Need reference?**                                                      │
│    → See [Part 2: Reference](#part-2-reference-📖)                          │
│    → Configuration, vocabulary, troubleshooting                             │
│                                                                              │
│ 🚀 **Just using docmgr?**                                                   │
│    → See `docmgr help how-to-use` instead                                   │
│    → This doc is for maintainers setting up the workspace                   │
└──────────────────────────────────────────────────────────────────────────────┘

---

## Overview

Setting up docmgr establishes a shared documentation infrastructure for your entire team. This one-time initialization creates the vocabulary (your team's documentation language), templates (structure scaffolds), and guidelines (quality standards) that ensure consistency across all tickets and contributors. The setup process balances sensible defaults with customization points, allowing teams to adopt docmgr quickly while evolving their documentation standards over time.

**This guide covers:** Repository initialization, vocabulary management, template customization, and configuration for multi-repo setups.

**Intended audience:** Repository maintainers, tech leads, and documentation champions responsible for setting up shared tooling.

**Daily users:** If you just want to create tickets and add docs, see `docmgr help how-to-use` instead. This guide is for the person setting up the workspace for the team.

---

## Key Concepts for Setup

Setup-specific terms used in this guide:

- **Docs root** — The `ttmp/` directory containing all documentation (default location)
- **Vocabulary** — Central `vocabulary.yaml` defining valid topics, doc types, and intents
- **Templates** — Markdown files in `_templates/` used by `docmgr doc add` to scaffold new docs
- **Guidelines** — Writing guidance in `_guidelines/` shown via `docmgr doc guidelines`
- **.ttmp.yaml** — Optional config file for custom root paths or multi-repo setups
- **Slugification** — How ticket titles become directory names (lowercase, dashes, normalized)

---

# Part 1: Repository Setup 📚

**[15 minute read — Start here for new repositories]**

---

## 1. Check Current Setup [BASIC]

Before initializing, check if docmgr is already set up:

```bash
docmgr status --summary-only
```

**If initialized:**
```
root=/path/ttmp vocabulary=/path/vocabulary.yaml tickets=N docs=M
```

**If NOT initialized:**
```
Error: root directory does not exist: /path/ttmp
```

If already initialized, skip to [Section 3 (Vocabulary)](#3-managing-vocabulary-intermediate) to customize or verify your vocabulary.

---

## 2. Initialize Repository [BASIC]

**Run once per repository:**

```bash
docmgr init
```

**What this creates:**

```
ttmp/
├── vocabulary.yaml     # Seeded with common topics/docTypes/intent/status
│                       # (pass --seed-vocabulary=false for an empty vocabulary)
├── _templates/         # Document scaffolds (design-doc.md, reference.md, etc.)
├── _guidelines/        # Writing guidelines for each doc type
└── .docmgrignore       # Validation exclusions (.git/, _templates/, etc.)
```

**Verify initialization:**

```bash
# Check status
docmgr status --summary-only

# View seeded vocabulary
docmgr vocab list

# See templates
docmgr ticket list  # Will be empty but proves root exists
```

**Understanding what was created:**

- **vocabulary.yaml** — Defines valid topics (backend, frontend, websocket), doc types (design-doc, reference, playbook), and intents (long-term). Used for validation warnings, not enforcement.

- **_templates/** — Contains markdown templates for each doc type with placeholders (`{{TITLE}}`, `{{TICKET}}`, etc.) that `docmgr doc add` fills automatically.

- **_guidelines/** — Writing guidance shown via `docmgr doc guidelines --doc-type TYPE`. Customize these to encode your team's standards.

- **.docmgrignore** — Like `.gitignore` but for docmgr validation. Excludes `.git/`, `_templates/`, `_guidelines/` by default.

**Note:** Running `docmgr init` is idempotent (safe to run multiple times). It won't overwrite existing files unless you use `--force`.

---

## 3. Managing Vocabulary [INTERMEDIATE]

Vocabulary defines your team's shared language for documentation.

### View Current Vocabulary

```bash
# List all vocabulary
docmgr vocab list

# List specific category
docmgr vocab list --category topics
docmgr vocab list --category docTypes
```

### Add Custom Entries

```bash
# Add topics for your architecture
docmgr vocab add --category topics --slug frontend \
  --description "Frontend app and components"

docmgr vocab add --category topics --slug infrastructure \
  --description "Infrastructure and deployment"

# Add custom doc types
docmgr vocab add --category docTypes --slug til \
  --description "Today I Learned (short notes capturing lessons)"

docmgr vocab add --category docTypes --slug analysis \
  --description "Analysis documents and research"

# Add intent values
docmgr vocab add --category intent --slug temporary \
  --description "Short-lived documentation"
```

**Guidance for vocabulary:**
- **Topics** — Reflect your architecture/domains (backend, frontend, database, api, etc.)
- **DocTypes** — How readers approach the doc (design-doc, reference, playbook, til, etc.)
- **Intent** — Longevity expectations (long-term for persistent docs, short-term for active work, throwaway for experiments)
- **Status** — Workflow state (draft, active, review, complete, archived) — vocabulary-guided, not enforced

**Keep vocabulary small initially** (5-8 topics, 5-7 doc types). Evolve with team consensus via PRs.

### Status Vocabulary (NEW)

Status values are now vocabulary-guided, allowing teams to customize workflow states:

```bash
# View current status values
docmgr vocab list --category status

# Add custom status
docmgr vocab add --category status --slug blocked \
  --description "Work blocked by external dependencies"

# Add another
docmgr vocab add --category status --slug on-hold \
  --description "Work paused temporarily"
```

**Default status values:**
- `draft` — Initial draft state
- `active` — Active work in progress
- `review` — Ready for review
- `complete` — Work completed
- `archived` — Archived/completed work

**Suggested transitions (not enforced):**
```
draft → active → review → complete → archived
review → active (rework)
complete → active (reopen, unusual)
```

Doctor warns on unknown status values but doesn't fail, encouraging consistency while allowing flexibility.

### Using Custom Doc Types

```bash
# After adding 'til' to vocabulary
docmgr doc add --ticket MEN-XXXX --doc-type til --title "TIL — WebSocket Reconnection"
```

If a template exists at `ttmp/_templates/til.md`, it will be used. Otherwise the doc is created under a subdirectory named after its doc-type (for example `til/`) with `DocType: til` so it still participates in search and validation.

---

## 4. Customizing Templates and Guidelines [INTERMEDIATE]

Templates and guidelines enforce house style and quality standards.

### Templates

Templates are in `ttmp/_templates/<docType>.md` and use placeholders:

- `{{TITLE}}` — Document title
- `{{TICKET}}` — Ticket identifier  
- `{{DATE}}` — Current timestamp
- `{{TOPICS}}` — YAML-formatted topics array
- `{{OWNERS}}` — YAML-formatted owners array
- `{{SUMMARY}}` — Summary text

**When you run `docmgr doc add`, these placeholders are automatically filled.**

**Customize templates** by editing files in `_templates/`. Use `docmgr init --force` to re-scaffold if you want to reset to defaults.

### Guidelines

Guidelines are shown via `docmgr doc guidelines --doc-type TYPE` and help writers understand:
- What sections to include
- What quality standards to meet
- What reviewers look for

**Preview guidelines:**

```bash
docmgr doc guidelines --doc-type design-doc
docmgr doc guidelines --doc-type reference
docmgr doc guidelines --list  # Show all available types
```

**Customize guidelines** by editing files in `_guidelines/`.

**Best practices:**
- Templates give structure (sections, scaffolds)
- Guidelines give intent (what to write in each section, quality expectations)
- See `docmgr help templates-and-guidelines` for detailed customization guidance

---

✅ **Milestone: Repository is Set Up!**

Your team can now:
- Create tickets with `docmgr ticket create`
- Add docs with `docmgr doc add`
- Search with `docmgr doc search`

**What's next?**
- Set up CI validation (see separate **CI and Automation Guide**)
- Customize vocabulary for your domain
- Share setup with your team

> **For CI integration:** See the separate **CI and Automation Guide** for GitHub Actions, GitLab CI, pre-commit hooks, Makefile integration, reporting, and monitoring.

---

# Part 2: Reference 📖

**[Advanced configuration and troubleshooting]**

---

## 5. Repository Configuration (.ttmp.yaml) [ADVANCED]

**Most teams don't need this** — defaults work well. Use `.ttmp.yaml` for:
- Custom root directory names
- Multi-repo/monorepo setups
- Centralized vocabulary across repos

### Create Configuration

```bash
docmgr configure --root ttmp --owners manuel --intent long-term --vocabulary ttmp/vocabulary.yaml
```

Or manually create `.ttmp.yaml` at repository root:

```yaml
root: ttmp
defaults:
  owners: [manuel, alex]
  intent: long-term
vocabulary: ttmp/vocabulary.yaml
```

**Fields:**
- `root` — Default docs root (overrides built-in `ttmp`)
- `defaults.owners` — Applied to new ticket indexes
- `defaults.intent` — Default intent for new docs
- `vocabulary` — Path to vocabulary file

### Root Resolution Order

1. `--root` flag (explicit)
2. `.ttmp.yaml:root` (relative to config file location)
3. `<git-root>/ttmp` (if `.git/` found walking up)
4. `<cwd>/ttmp` (fallback)

**Note:** `.ttmp.yaml` doesn't need to live at repository root. In monorepos, place it at a parent directory to centralize configuration.

---

## 6. Repository Conventions [BASIC]

**Directory structure:**

```
repository/
├── .ttmp.yaml              # Optional configuration
├── ttmp/                   # Docs root (default name)
│   ├── vocabulary.yaml     # Shared vocabulary
│   ├── .docmgrignore       # Validation exclusions
│   ├── _templates/         # Doc scaffolds
│   ├── _guidelines/        # Writing guidance
│   │
│   ├── TICKET-001-slug/    # Ticket workspace
│   │   ├── index.md
│   │   ├── tasks.md
│   │   ├── changelog.md
│   │   ├── design-doc/        # Created when you add a design-doc
│   │   ├── reference/         # Created when you add a reference doc
│   │   ├── playbook/          # Created when you add a playbook
│   │   └── <doc-type>/        # Any other doc-type creates its own subdir
│   │
│   └── TICKET-002-slug/    # Another ticket
│       └── ...
```

**Ticket workspace naming:**
- Format: `<TICKET>-<slug>/`
- Slug from title: lowercase, alphanumerics and dashes only
- Example: "Chat WebSocket Lifecycle" → `chat-websocket-lifecycle`

**Per-ticket directories:**
- `index.md` — Ticket overview
- Doc-type subdirectories are created on demand by `docmgr doc add` (for example `design-doc/`, `reference/`, `playbook/`, or custom types like `til/`)
- `scripts/`, `sources/`, `archive/` — Other content
- `.meta/` — Internal metadata

---

## 7. Numeric Prefixes [INTERMEDIATE]

**Automatic prefixing:**
- New docs get prefixes: `01-`, `02-`, `03-`
- Keeps files ordered in listings
- Switches to 3 digits after 99 files
- Exempt files: `index.md`, `README.md`, `tasks.md`, `changelog.md`

**Resequencing:**

```bash
# After deleting files, renumber and update internal links
docmgr renumber --ticket MEN-4242
```

**Doctor behavior:**
- Warns when subdirectory markdown files are missing numeric prefixes
- Exempt files don't trigger warnings
- Suppress with `.docmgrignore` patterns if needed

---

## 8. Adding New Doc Types [INTERMEDIATE]

Extend doc types via vocabulary, then use immediately.

**Workflow:**

1) **Add to vocabulary:**

```bash
docmgr vocab add --category docTypes --slug til \
  --description "Today I Learned entries"
```

2) **Verify it's there:**

```bash
docmgr vocab list --category docTypes
```

3) **Create template (optional):**

Edit `ttmp/_templates/til.md` with your preferred structure. If no template exists, docs are still created and stored under `til/` with `DocType: til`.

4) **Use it:**

```bash
docmgr doc add --ticket MEN-XXXX --doc-type til --title "TIL — WebSocket Reconnection"
```

**Common custom doc types teams add:**
- `til` — Today I Learned notes
- `analysis` — Deep-dive research and analysis
- `code-review` — Code review summaries
- `decision` — Architecture decision records (ADRs)
- `retro` — Sprint retrospectives

---

## 9. Migrating Existing Docs [INTERMEDIATE]

**If you have existing markdown docs, migrate them into docmgr structure:**

**Workflow:**

1) **Create ticket workspace:**

```bash
docmgr ticket create --ticket MIGRATE-001 --title "Existing Docs" --topics migration
```

2) **Move files into ticket directory:**

```bash
# Move to appropriate subdirectory
mv docs/old-design.md ttmp/MIGRATE-001-existing-docs/design-doc/01-old-design.md
mv docs/api-ref.md ttmp/MIGRATE-001-existing-docs/reference/01-api-ref.md
```

3) **Add or fix frontmatter via commands:**

```bash
# Set ticket
docmgr meta update --doc ttmp/MIGRATE-001-.../design-doc/01-old-design.md \
  --field Ticket --value MIGRATE-001

# Set topics
docmgr meta update --doc ttmp/MIGRATE-001-.../design-doc/01-old-design.md \
  --field Topics --value "backend,api"

# Add summary
docmgr meta update --doc ttmp/MIGRATE-001-.../design-doc/01-old-design.md \
  --field Summary --value "Original API design documentation"
```

4) **Validate:**

```bash
docmgr doctor --ticket MIGRATE-001 --fail-on error
```

5) **Repeat for other docs**, organizing into appropriate tickets.

---

✅ **Milestone: Repository Fully Initialized!**

Your repository now has:
- ✅ Docs root with vocabulary
- ✅ Templates and guidelines  
- ✅ Validation configured
- ✅ Ready for team use

**What's next?**
- Set up CI (see [Part 2](#part-2-ci-integration-🔧))
- Share setup with team
- Create first ticket: `docmgr ticket create`

---

# Part 2: CI Integration 🔧

**[Enforce documentation quality automatically]**

---

## 10. Suppressing Validation Noise [INTERMEDIATE]

Use `.docmgrignore` to suppress validation warnings for specific files or patterns.

**Common patterns for `ttmp/.docmgrignore`:**

```
# VCS and tooling
.git/
_templates/
_guidelines/
node_modules/
dist/
coverage/

# Archive old tickets (don't validate)
archive/
2023-*/
2024-*/

# Ignore specific problematic files
ttmp/*/design-doc/index.md
ttmp/LEGACY-*/

# Ignore drafts and experiments
**/draft-*.md
**/scratch-*.md
```

**Docmgr automatically respects `.docmgrignore`** as workspace-wide ingest policy. The workspace matcher is backed by `github.com/denormal/go-gitignore`, loads `.docmgrignore` from the repository hierarchy (including the docs root and nested ticket/script directories), and prunes ignored paths before Markdown frontmatter parsing. This affects doctor, list, search, status, and other index-backed commands. Use `docmgr ignore explain <path>` to debug which rule ignored or included a path.

---

## 11. Operational Tips [INTERMEDIATE]

**Ongoing maintenance for repository maintainers:**

### Keep Vocabulary Stable
- Socialize vocabulary changes via PRs
- Don't add topics/doc-types ad-hoc
- Review quarterly, prune unused entries

### Enforce Index Quality
- Require `Owners`, `Summary`, and `RelatedFiles` on every `index.md`
- Use `docmgr doctor` to catch missing fields
- Template the reminder in PR templates

### Revisit Templates/Guidelines
- Review quarterly
- Incorporate lessons learned
- Adjust based on team feedback

### Monitor Workspace Health

```bash
# Quick health check
docmgr status

# Detailed status
docmgr doctor --all --fail-on none  # Show all warnings
```

---

## 12. Troubleshooting [INTERMEDIATE]

**Basic validation job:**

```yaml
validate-docs:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - name: Install docmgr
      run: go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
    - name: Validate
      run: docmgr doctor --all --stale-after 30 --fail-on error
```

### Adjusting Strictness

**Start lenient, increase over time:**

```bash
# Phase 1: Only errors (broken links, missing files)
docmgr doctor --all --fail-on error

# Phase 2: Add staleness check
docmgr doctor --all --stale-after 60 --fail-on error

# Phase 3: Make staleness stricter
docmgr doctor --all --stale-after 30 --fail-on error

# Phase 4: Treat warnings as errors (strict)
docmgr doctor --all --stale-after 30 --fail-on warning
```

**Use `.docmgrignore`** to suppress false positives rather than lowering standards.

---

## 13. Reporting in CI [INTERMEDIATE]

Generate documentation health reports:

```yaml
doc-report:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - name: Install docmgr
      run: go install -tags sqlite_fts5 github.com/go-go-golems/docmgr@latest
    
    - name: Generate report
      run: |
        echo "=== Documentation Status ==="
        docmgr status
        
        echo ""
        echo "=== Stale Docs (>30 days) ==="
        docmgr status --stale-after 30 --with-glaze-output --output json | \
          jq -r '.docs[] | select(.stale) | "[\(.ticket)] \(.title) — \(.days_since_update) days"'
        
        echo ""
        echo "=== Recent Activity (Last 7 Days) ==="
        docmgr doc search --updated-since "7 days ago"
```

---

## 14. Operational Tips [INTERMEDIATE]

**Ongoing maintenance:**

### Keep Vocabulary Stable
- Socialize vocabulary changes via PRs
- Don't add topics/doc-types ad-hoc
- Review quarterly, prune unused entries

### Enforce Index Quality
- Require `Owners`, `Summary`, and `RelatedFiles` on every `index.md`
- Use doctor to catch missing fields
- Template the index.md reminder in PR templates

### Use Search in Reviews
```bash
# During code review: find design context
docmgr doc search --file path/to/changed-file.go

# During architecture review: find related docs
docmgr doc search --query "authentication" --doc-type design-doc
```

### Treat Doctor Warnings as Tech Debt
- Unknown topics → Add to vocabulary or fix typo
- Stale docs → Update or mark as evergreen (use `.docmgrignore`)
- Missing files → Fix path or remove from RelatedFiles
- Track follow-ups in issues

### Revisit Templates/Guidelines
- Review quarterly
- Incorporate lessons learned
- Adjust based on team feedback

### Use Tasks and Changelog
```bash
# Encourage consistent tracking
docmgr task add --ticket T --text "Task description"
docmgr changelog update --ticket T --entry "What changed"
```

---

# Part 3: Reference 📖

**[Advanced topics and troubleshooting]**

---

### Issue: Commands Can't Find Root

**Symptom:** `Error: root directory does not exist`

**Solutions:**
```bash
# Check status to see what root it's looking for
docmgr status --summary-only

# Run from repository root
cd /path/to/repo && docmgr status

# Or set absolute path in .ttmp.yaml
# In .ttmp.yaml:
root: /absolute/path/to/ttmp
```

### Issue: Vocabulary Not Found

**Symptom:** Using topics but doctor warns "unknown topics"

**Solutions:**
```bash
# Check vocabulary location
docmgr status --summary-only
# Look at vocabulary= path

# List what's in vocabulary
docmgr vocab list --category topics

# Add missing topics
docmgr vocab add --category topics --slug YOUR-TOPIC
```

### Issue: Too Many Doctor Warnings

**Solutions:**
1. Add patterns to `.docmgrignore`
2. Adjust `--stale-after` threshold
3. Fix legitimate issues flagged

```bash
# See what's being flagged
docmgr doctor --all --with-glaze-output --output json | \
  jq -r '.[] | select(.issue != "none") | "\(.path): \(.message)"'
```

### Issue: Parentheses in Ticket Names

**Symptom:** Shell errors when cd'ing into ticket directories

**Solutions:**
```bash
# Quote/escape directory names with parens
cd "ttmp/MEN-XXXX-name-\(with-parens\)"

# Or avoid parens in ticket titles
```

---

## 13. Multi-Repo and Monorepo Setup [ADVANCED]

**Scenario:** Multiple repos sharing vocabulary.

### Option A: Centralized Config

Place `.ttmp.yaml` at parent directory:

```
parent/
├── .ttmp.yaml              # Points to shared vocabulary
├── repo-a/
│   └── ttmp-a/
└── repo-b/
    └── ttmp-b/
```

`.ttmp.yaml`:
```yaml
root: %(repo)/ttmp-%(repo-name)
vocabulary: /absolute/path/to/shared/vocabulary.yaml
```

### Option B: Separate Roots, Shared Vocabulary

Each repo has its own `ttmp/` but references shared vocabulary:

```
repo/
├── .ttmp.yaml
└── ttmp/
```

`.ttmp.yaml`:
```yaml
root: ttmp
vocabulary: /shared/path/vocabulary.yaml
```

**Most teams use default setup.** Only use `.ttmp.yaml` if you have specific needs.

---

## 14. Vocabulary Philosophy [REFERENCE]

**What vocabulary does:**
- Defines valid topics, doc types, intents
- Used by `docmgr doctor` for validation warnings
- NOT enforced — you can use any topics/doc-types
- Helps team maintain consistency

**Why it's not enforced:**
- Allows exploratory work (try new topics without approval)
- Unknown/custom doc-types create their own `<doc-type>/` subdirectory (flexible)
- Warnings better than errors (doesn't block progress)

**Think of vocabulary as:**
- Documentation of team conventions
- Input for autocomplete (future feature)
- Basis for validation warnings (not blockers)

---

## 15. Quick Reference

### Initialization Commands

```bash
# Initialize with seeded vocabulary
docmgr init

# Initialize empty
docmgr init

# Force re-scaffold templates/guidelines
docmgr init --force

# Check if initialized
docmgr status --summary-only
```

### Vocabulary Commands

```bash
# List all
docmgr vocab list

# List category
docmgr vocab list --category topics

# Add entry
docmgr vocab add --category topics --slug NAME --description "..."

# Structured output
docmgr vocab list --with-glaze-output --output json
```

### Configuration Commands

```bash
# Create .ttmp.yaml
docmgr configure --root ttmp --owners USER --vocabulary ttmp/vocabulary.yaml

# Check configuration
docmgr status --summary-only
# Shows: root=... config=... vocabulary=...
```



## Related Documentation

- **Daily usage:** `docmgr help how-to-use` — Creating tickets, adding docs, searching
- **CI/automation:** See **CI and Automation Guide** (playbooks/03-ci-and-automation-guide.md) — GitHub Actions, hooks, reporting
- **Templates:** `docmgr help templates-and-guidelines` — Customization guide
- **CLI reference:** `docmgr help cli-guide` — Command overview
