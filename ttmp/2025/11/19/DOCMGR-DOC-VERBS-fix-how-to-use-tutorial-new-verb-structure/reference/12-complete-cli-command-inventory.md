---
Title: Complete CLI command inventory — Tutorial coverage analysis
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - cli
    - documentation
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/cmd/docmgr/cmds/root.go
      Note: CLI command tree registration
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Tutorial being analyzed for coverage
ExternalSources: []
Summary: "Complete inventory of all docmgr commands from CLI help output, showing what's documented vs missing."
LastUpdated: 2025-11-25
---

# Complete CLI Command Inventory

## Purpose

This document catalogs ALL docmgr commands discovered through `docmgr help` exploration and maps them to tutorial coverage. Created during Round 4 debate to ensure no functionality is overlooked.

---

## Command Tree Structure

### Top-Level Commands (from `docmgr help`)

```
docmgr
├── changelog
│   └── update
├── completion (bash|zsh|fish|powershell)
├── config
│   └── show
├── configure
├── doc
│   ├── add
│   ├── guidelines
│   ├── layout-fix
│   ├── list
│   ├── relate
│   ├── renumber
│   └── search
├── doctor
├── help
├── import
│   └── file
├── init
├── list
│   ├── docs
│   └── tickets
├── meta
│   └── update
├── status
├── task (alias: tasks)
│   ├── add
│   ├── check
│   ├── edit
│   ├── list
│   ├── remove
│   └── uncheck
├── template
│   └── validate
├── ticket
│   ├── close
│   ├── create-ticket
│   ├── rename-ticket
│   └── tickets
├── vocab
│   ├── add
│   └── list
└── workspace
    ├── configure
    ├── doctor
    ├── init
    └── status
```

**Total command groups:** 16  
**Total subcommands:** 35+  

---

## Command Aliasing Map

### Workspace Commands (Available at Two Levels)

| Short Form (Top-Level) | Long Form (Under workspace) | Tutorial Uses |
|------------------------|----------------------------|---------------|
| `docmgr init` | `docmgr workspace init` | Short form ✅ |
| `docmgr doctor` | `docmgr workspace doctor` | Short form ✅ |
| `docmgr status` | `docmgr workspace status` | Short form ✅ |
| `docmgr configure` | `docmgr workspace configure` | Neither (rare) |

**Both forms work identically.** The short form is syntactic sugar for convenience.

---

### List Commands (Available at Two Levels)

| Short Form | Long Form | Tutorial Uses |
|-----------|-----------|---------------|
| `docmgr list docs` | `docmgr doc list` | Short form ✅ |
| `docmgr list tickets` | `docmgr ticket tickets` | Short form ✅ |

---

### Task Command (Alias)

| Primary | Alias | Tutorial Uses |
|---------|-------|---------------|
| `docmgr task` | `docmgr tasks` | Primary ✅ |

---

### Tutorial Impact

**Good:** Tutorial consistently uses short forms (docmgr init, docmgr doctor, docmgr list docs).

**Problem:** Aliasing is NOT explained anywhere in tutorial.

**Risk:** Users who see both forms in different contexts might think:
- "Are these different commands?"
- "Which is correct?"
- "Is one deprecated?"

**Solution:** Add command aliasing section to Part 4 (Reference).

---

## Coverage Analysis: Documented vs. Undocumented

### ✅ DOCUMENTED in Tutorial

**Part 1 (Essentials):**
- `docmgr init` (Section 2)
- `docmgr ticket create-ticket` (Section 3)
- `docmgr doc add` (Section 4)
- `docmgr doc search` (Section 5)
- `docmgr doc guidelines` (Section 4, mentioned)

**Part 2 (Everyday Workflows):**
- `docmgr meta update` (Section 6)
- `docmgr doc relate` (Section 7)
- `docmgr changelog update` (Section 8)
- `docmgr doctor` (Section 9)
- `docmgr task add` (Section 10)
- `docmgr task check` (Section 10)
- `docmgr task list` (Section 10)
- `docmgr status` (Section 11)

**Part 3 (Power User):**
- `docmgr list tickets --with-glaze-output` (Section 11)
- `docmgr list docs --with-glaze-output` (Section 11)

**Part 4 (Reference):**
- `docmgr list tickets` (Section 13)
- `docmgr list docs` (Section 13)
- `docmgr status` (Section 13)
- `docmgr vocab list` (Section 15)
- `docmgr vocab add` (Section 15)
- `docmgr doc renumber` (Section 16, brief mention)
- `docmgr ticket close` (Section 10)

**Total documented:** ~20 commands (core workflow covered)

---

### ❌ NOT DOCUMENTED in Tutorial

**Everyday/Useful:**
1. `docmgr task edit` — Edit task text
2. `docmgr task remove` — Remove tasks
3. `docmgr task uncheck` — Uncheck completed tasks
4. `docmgr completion` — Shell autocompletion setup
5. `docmgr config show` — Debug configuration

**Maintenance/Advanced:**
6. `docmgr doc layout-fix` — Reorganize doc structure
7. `docmgr doc renumber` — Resequence prefixes (mentioned briefly)
8. `docmgr import file` — Import external files
9. `docmgr configure` — Create/update .ttmp.yaml

**Rare/Very Advanced:**
10. `docmgr ticket rename-ticket` — Rename ticket IDs
11. `docmgr template validate` — Validate template syntax

**Meta Information:**
12. Command aliasing structure — Not explained

**Total undocumented:** 12 commands/features

---

## Priority Ranking for Documentation

### MUST ADD (Everyday Workflow):

**1. Task editing commands** (Priority: HIGH)
- `task edit/remove/uncheck`
- **Why:** Natural extension of task workflow
- **Impact:** Users who add tasks will need to edit/remove them
- **Location:** Part 2, Section 10 (expand existing section)
- **Lines:** +30

**2. Shell completion** (Priority: HIGH)
- `completion bash|zsh|fish`
- **Why:** Huge quality-of-life improvement, reduces typos
- **Impact:** Makes CLI much easier to use
- **Location:** Part 1 (Prerequisites) OR Part 4 (Tips)
- **Lines:** +20

**3. Command aliasing** (Priority: HIGH)
- Explain: docmgr init ↔ workspace init, etc.
- **Why:** Prevents confusion about "duplicate" commands
- **Impact:** Users see both forms in wild, need explanation
- **Location:** Part 4 (Reference)
- **Lines:** +30

---

### SHOULD ADD (Maintenance/Power User):

**4. Maintenance commands** (Priority: MEDIUM)
- `doc layout-fix`, `config show`, expand `renumber`
- **Why:** Users will encounter these during maintenance
- **Impact:** Currently users have no guidance
- **Location:** Part 4, Section 17 (new or expanded)
- **Lines:** +60

**5. Import workflow** (Priority: MEDIUM)
- `import file` usage and examples
- **Why:** Power users import external content
- **Impact:** Enables research/LLM content workflows
- **Location:** Part 3 (Power User Features)
- **Lines:** +40

---

### CAN SKIP (Rare/Too Advanced):

**6. Configure command** (Priority: LOW)
- Already mentioned in Part 4, Section 15
- Most users use `init` instead
- **Action:** Keep brief mention, don't expand

**7. Rename ticket** (Priority: LOW)
- Very rare operation
- **Action:** Users can read `--help` when needed

**8. Template validate** (Priority: LOW)
- Only for custom template development
- **Action:** Create separate guide if needed

---

## Recommended Additions to Tutorial

### Summary

**Total additions: ~180 lines across 5 sections**

| Section | Addition | Lines | Priority |
|---------|----------|-------|----------|
| Part 1 or 4 | Shell completion | +20 | HIGH |
| Part 2, §10 | Task editing | +30 | HIGH |
| Part 3 | Import workflow | +40 | MEDIUM |
| Part 4, §17 | Maintenance commands | +60 | MEDIUM |
| Part 4, §18 | Command aliasing | +30 | HIGH |

**HIGH priority: 80 lines** (completion + task editing + aliasing)  
**MEDIUM priority: 100 lines** (maintenance + import)

---

## Command Documentation Status Table

### Complete Inventory

| Command | Documented? | Location | Notes |
|---------|------------|----------|-------|
| **workspace** |
| init | ✅ Yes | Part 1, §2 | Short form used |
| doctor | ✅ Yes | Part 2, §9 | Short form used |
| status | ✅ Yes | Part 2, §11 | Short form used |
| configure | ⚠️ Brief | Part 4, §15 | Could expand |
| **ticket** |
| create-ticket | ✅ Yes | Part 1, §3 | Core workflow |
| close | ✅ Yes | Part 2, §10 | Well documented |
| tickets | ✅ Yes | Part 4, §13 | Via list tickets |
| rename-ticket | ❌ No | — | Too rare |
| **doc** |
| add | ✅ Yes | Part 1, §4 | Core workflow |
| search | ✅ Yes | Part 1, §5 | Core workflow |
| relate | ✅ Yes | Part 2, §7 | Core workflow |
| guidelines | ✅ Yes | Part 1, §4 | Mentioned |
| list | ✅ Yes | Part 4, §13 | Via list docs |
| layout-fix | ❌ No | — | **SHOULD ADD** |
| renumber | ⚠️ Brief | Part 4, §16 | **SHOULD EXPAND** |
| **task** |
| add | ✅ Yes | Part 2, §10 | Core workflow |
| check | ✅ Yes | Part 2, §10 | Core workflow |
| list | ✅ Yes | Part 2, §10 | Core workflow |
| edit | ❌ No | — | **MUST ADD** |
| remove | ❌ No | — | **MUST ADD** |
| uncheck | ❌ No | — | **MUST ADD** |
| **changelog** |
| update | ✅ Yes | Part 2, §8 | Core workflow |
| **meta** |
| update | ✅ Yes | Part 2, §6 | Core workflow |
| **vocab** |
| add | ✅ Yes | Part 4, §15 | Well documented |
| list | ✅ Yes | Part 4, §15 | Well documented |
| **import** |
| file | ❌ No | — | **SHOULD ADD** |
| **template** |
| validate | ❌ No | — | Too advanced |
| **config** |
| show | ❌ No | — | **SHOULD ADD** |
| **list** |
| docs | ✅ Yes | Part 4, §13 | Core workflow |
| tickets | ✅ Yes | Part 4, §13 | Core workflow |
| **completion** |
| (bash/zsh/fish) | ❌ No | — | **SHOULD ADD** |
| **configure** |
| (top-level) | ⚠️ Brief | Part 4, §15 | Alias of workspace configure |

---

## Key Discoveries from Full CLI Exploration

### 1. Shell Completion is Missing!

**Command:** `docmgr completion bash|zsh|fish|powershell`

**Why it matters:**
- Enables tab-completion of commands (docmgr d<tab> → doc, doctor)
- Enables tab-completion of flags (--<tab> → list all flags)
- Reduces typing errors (major pain point from validation)
- Speeds up workflow significantly

**Impact:** HIGH quality-of-life improvement for all users.

**Current tutorial:** NO mention of completion at all!

**Recommendation:** Add to Part 1 (Optional Setup) or Part 4 (Tips & Tricks).

---

### 2. Command Aliasing is Undocumented!

**Finding:** Many commands work at TWO levels.

**Examples:**
- `docmgr init` = `docmgr workspace init`
- `docmgr doctor` = `docmgr workspace doctor`
- `docmgr list docs` = `docmgr doc list`

**Tutorial behavior:** Uses short forms consistently ✅

**Problem:** Never EXPLAINS that both forms work.

**User confusion risk:**
- User sees `docmgr init` in tutorial
- Sees `docmgr workspace init` in error message or other doc
- Wonders: "Did I use the wrong one? Are these different?"

**Recommendation:** Add "Command Aliasing" section to Part 4.

---

### 3. Task Editing Commands are Critical

**Missing:**
- `task edit` — Change task text
- `task remove` — Delete tasks
- `task uncheck` — Mark incomplete

**Why critical:**
- Tutorial teaches `task add` and `task check`
- Logical next need: "How do I fix a typo in a task?"
- Without docs, users will manually edit `tasks.md` (breaking intended workflow)

**Recommendation:** Add to Part 2, Section 10 (expand existing task section).

---

### 4. Config Show is Debugging Essential

**Command:** `docmgr config show`

**What it shows:**
- Current root path
- Vocabulary path
- .ttmp.yaml resolution
- Configuration source (flag/config/default)

**Why it matters:**
- Debugging "root not found" errors
- Understanding multi-repo setups
- Verifying configuration

**Current tutorial:** Mentions configuration (Part 4, §15) but not `config show` command.

**Recommendation:** Add to Part 4 or troubleshooting section.

---

## Updated Documentation Plan

### HIGH Priority Additions (Must Add):

**1. Part 1, Section 1a (Optional) OR Part 4, Section 18 (Tips):**
```markdown
### Shell Completion (Optional Setup)

Enable tab-completion for docmgr commands:

```bash
# Bash
docmgr completion bash > ~/.bash_completion.d/docmgr
source ~/.bash_completion.d/docmgr

# Zsh (oh-my-zsh)
docmgr completion zsh > "${fpath[1]}/_docmgr"
# Then restart shell

# Fish
docmgr completion fish > ~/.config/fish/completions/docmgr.fish
```

Benefits:
- Tab-complete commands (docmgr d<TAB> → doc, doctor)
- Tab-complete flags (--<TAB> → see all flags)
- Reduces typing errors
- Speeds up workflow

Once set up, type `docmgr <TAB><TAB>` to see all commands.
```

**Lines:** +25  
**Impact:** Huge QoL improvement, reduces typos (major error source)

---

**2. Part 2, Section 10 (Expand Manage Tasks):**
```markdown
## 10. Manage Tasks

### Add and Check Tasks
[Existing content]

### Edit and Remove Tasks

```bash
# Edit task text
docmgr task edit --ticket MEN-4242 --id 3 --text "Updated task description"

# Remove a task
docmgr task remove --ticket MEN-4242 --id 5

# Uncheck a task if marked done by mistake
docmgr task uncheck --ticket MEN-4242 --id 2
```

**When to use:**
- **edit**: Fix typos, update task description as work evolves
- **remove**: Delete obsolete or duplicate tasks
- **uncheck**: Reopen a task if work needs redoing

**Tip:** For quick fixes, you can also edit `tasks.md` directly. The task commands are useful for scripts and automation.
```

**Lines:** +35  
**Impact:** Completes task workflow, natural extension of add/check

---

**3. Part 4, Section 18 (NEW: Command Aliasing):**
```markdown
## 18. Command Aliasing [REFERENCE]

Many docmgr commands are available at both top-level and grouped paths. Both forms work identically.

### Workspace Commands

| Short Form (Recommended) | Long Form (Also Works) |
|-------------------------|------------------------|
| `docmgr init` | `docmgr workspace init` |
| `docmgr doctor` | `docmgr workspace doctor` |
| `docmgr status` | `docmgr workspace status` |
| `docmgr configure` | `docmgr workspace configure` |

### List Commands

| Short Form (Recommended) | Long Form (Also Works) |
|-------------------------|------------------------|
| `docmgr list docs` | `docmgr doc list` |
| `docmgr list tickets` | `docmgr ticket tickets` |

### Task Command Alias

- `docmgr task` = `docmgr tasks` (both accepted)

### Tutorial Convention

This tutorial uses **short forms** for brevity:
- `docmgr init` (not `docmgr workspace init`)
- `docmgr doctor` (not `docmgr workspace doctor`)
- `docmgr list docs` (not `docmgr doc list`)

**Both forms are correct.** Use whichever you prefer. The short forms are slightly more concise; the long forms are more explicit about the command group.
```

**Lines:** +35  
**Impact:** Prevents confusion about "duplicate" commands

---

### MEDIUM Priority Additions:

**4. Part 3, Section 12a (NEW: Importing External Content):**
```markdown
## 12a. Importing External Content [ADVANCED]

Import markdown files from external sources:

```bash
# Import a file with explicit doc-type
docmgr import file --ticket MEN-4242 \
  --file path/to/research.md \
  --doc-type reference

# Import preserves existing frontmatter if present
# If frontmatter is missing, docmgr adds it
```

**Use cases:**
- Importing LLM-generated analysis
- Migrating from other documentation systems
- Pulling in external research notes
- Consolidating scattered markdown files

**Tip:** Imported files are moved into the ticket workspace with proper structure. Original file is preserved unless you specify `--move`.
```

**Lines:** +40  
**Impact:** Enables power user workflows

---

**5. Part 4, Section 17 (Expand: Maintenance Commands):**
```markdown
## 17. Maintenance Commands [ADVANCED]

### layout-fix — Reorganize Document Structure

Moves documents into subdirectories matching their DocType:

```bash
# Preview changes (dry-run)
docmgr doc layout-fix --ticket MEN-4242 --dry-run

# Apply changes
docmgr doc layout-fix --ticket MEN-4242
```

**When needed:**
- After bulk imports with inconsistent structure
- When migrating from old docmgr versions
- When manually created docs aren't in correct subdirectories

**What it does:**
- Moves `design/01-doc.md` → `design-doc/01-doc.md` (if DocType is "design-doc")
- Updates internal markdown links automatically
- Skips root files (index.md, tasks.md, changelog.md)

---

### config show — Display Configuration

Shows how docmgr resolved your configuration:

```bash
docmgr config show
```

**Output shows:**
- Root directory path
- Vocabulary file path
- Configuration source (.ttmp.yaml, defaults, or flags)
- Default owners and intent

**When needed:**
- Debugging "root not found" errors
- Understanding multi-repo setups
- Verifying .ttmp.yaml is being used

---

### renumber — Resequence Numeric Prefixes

[Expand existing brief mention with examples]

```bash
# Renumber all docs in a ticket
docmgr doc renumber --ticket MEN-4242
```

**When needed:**
- After deleting docs (gaps in numbering: 01, 02, 05, 07)
- After reordering docs manually
- To enforce clean sequential prefixes

**What it does:**
- Renumbers to: 01-, 02-, 03-, ... (or 001- after 99 files)
- Updates internal markdown links
- Preserves file content and frontmatter
```

**Lines:** +65  
**Impact:** Covers maintenance scenarios users will encounter

---

## Total Additional Documentation

**HIGH priority (must add):**
- Shell completion: +25 lines
- Task editing: +35 lines
- Command aliasing: +35 lines
**Subtotal: ~95 lines**

**MEDIUM priority (should add):**
- Import workflow: +40 lines
- Maintenance commands: +65 lines
**Subtotal: ~105 lines**

**TOTAL: ~200 lines of new documentation**

(Previous estimate was 110 lines, now 200 after discovering completion and aliasing)

---

## Acceptance Criteria

**After additions, tutorial should:**
- ✅ Document all everyday workflow commands (task editing)
- ✅ Document high-impact QoL features (shell completion)
- ✅ Explain command structure (aliasing)
- ✅ Cover maintenance operations (layout-fix, config show, renumber)
- ✅ Enable power user workflows (import)
- ✅ Maintain focus (skip rare commands like rename-ticket)

**Validation:**
- Run `docmgr help` and verify each command either documented OR explicitly skipped
- Check that all documented commands actually exist (no wrong syntax)
- Verify examples work in fresh environment

---

## Next Step

This inventory feeds directly into Round 4 (Missing Functionality) debate decisions and Phase 1 implementation checklist.

