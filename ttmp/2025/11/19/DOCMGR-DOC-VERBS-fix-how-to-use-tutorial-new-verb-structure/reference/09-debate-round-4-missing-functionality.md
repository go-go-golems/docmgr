---
Title: Debate Round 4 — Which Functionality Should Be Documented?
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - tutorial
    - feature-coverage
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/pkg/commands/layout_fix.go
      Note: Layout-fix command (undocumented)
    - Path: docmgr/pkg/commands/renumber.go
      Note: Renumber command (undocumented)
    - Path: docmgr/pkg/commands/rename_ticket.go
      Note: Rename-ticket command (undocumented)
    - Path: docmgr/pkg/commands/template_validate.go
      Note: Template validate command (undocumented)
    - Path: docmgr/cmd/docmgr/cmds
      Note: Full command tree structure
ExternalSources: []
Summary: "Round 4 debate: Which docmgr commands/features should be added to the tutorial? Requires extensive codebase analysis."
LastUpdated: 2025-11-25
---

# Debate Round 4 — Which Functionality Should Be Documented?

## Question

**"The tutorial covers basic workflow (init, create, add, search, relate, tasks, changelog, doctor). What OTHER docmgr functionality exists that should be documented in the 'how-to-use' tutorial?"**

**Primary Candidates:**
- Jamie Park (Technical Writer)
- Alex Rivera (Structure Architect)
- The Tutorial (Document Entity)
- Dr. Maya Chen (Accuracy Crusader)

---

## Pre-Debate Research

### Complete Command Inventory from Codebase

**Commands Currently Documented in Tutorial:**

| Command | Documented | Section |
|---------|-----------|---------|
| `docmgr init` | ✅ | Part 1, Section 2 |
| `docmgr ticket create-ticket` | ✅ | Part 1, Section 3 |
| `docmgr doc add` | ✅ | Part 1, Section 4 |
| `docmgr doc search` | ✅ | Part 1, Section 5 |
| `docmgr doc relate` | ✅ | Part 2, Section 7 |
| `docmgr changelog update` | ✅ | Part 2, Section 8 |
| `docmgr task add/check/list` | ✅ | Part 2, Section 10 |
| `docmgr doctor` | ✅ | Part 2, Section 9 |
| `docmgr status` | ✅ | Part 2, Section 11 |
| `docmgr list tickets/docs` | ✅ | Part 4, Section 13 |
| `docmgr meta update` | ✅ | Part 2, Section 6 |
| `docmgr vocab add/list` | ✅ | Part 4, Section 15 |
| `docmgr doc guidelines` | ✅ | Part 1, Section 4 (mentioned) |
| `docmgr ticket close` | ✅ | Part 2, Section 10 |

**Commands NOT Documented in Tutorial:**

| Command | Purpose (from CLI help) | Complexity | Use Frequency |
|---------|------------------------|------------|---------------|
| `docmgr doc layout-fix` | Move docs into `<doc-type>/` subdirectories | Medium | Maintenance |
| `docmgr doc renumber` | Resequence numeric prefixes (01-, 02-, ...) | Low | Maintenance |
| `docmgr ticket rename-ticket` | Rename ticket ID and move directory | Medium | Rare |
| `docmgr import file` | Import external markdown files | Low | Setup |
| `docmgr template validate` | Validate template syntax | Low | Advanced |
| `docmgr config show` | Show current configuration | Low | Debugging |
| `docmgr configure` | Create/update .ttmp.yaml | Low | Setup |
| `docmgr task edit/remove/uncheck` | Edit/remove/uncheck tasks | Low | Everyday |
| `docmgr workspace configure` | Configure workspace (alias of configure) | Low | Setup |
| `docmgr completion` | Generate shell autocompletion | Low | Setup |

**Command Aliasing Structure (Important!):**
- `docmgr init` ↔ `docmgr workspace init` (both work)
- `docmgr doctor` ↔ `docmgr workspace doctor` (both work)
- `docmgr status` ↔ `docmgr workspace status` (both work)
- `docmgr configure` ↔ `docmgr workspace configure` (both work)
- `docmgr list docs` ↔ `docmgr doc list` (both work)
- `docmgr list tickets` ↔ `docmgr ticket tickets` (both work)
- `docmgr task` ↔ `docmgr tasks` (alias)

**Tutorial uses:** Short form (docmgr init, docmgr doctor) which is good!  
**Issue:** Aliasing not explained — users don't know both forms work

---

### Detailed Analysis of Undocumented Commands

#### 1. **`docmgr doc layout-fix`** (from layout_fix.go)

**What it does:**
- Scans ticket workspaces and moves markdown documents into subdirectories named after their DocType
- E.g., moves `design/01-architecture.md` → `design-doc/01-architecture.md`
- Updates internal links automatically
- Skips root-level control files (index.md, tasks.md, changelog.md)

**Usage:**
```bash
# Fix layout for all tickets
docmgr doc layout-fix

# Fix layout for specific ticket
docmgr doc layout-fix --ticket MEN-4242

# Dry run to preview changes
docmgr doc layout-fix --ticket MEN-4242 --dry-run
```

**When needed:**
- After bulk ticket imports with inconsistent structure
- When migrating from old docmgr versions
- When manually creating docs without proper subdirectories

**Tutorial relevance:**
- **Low for beginners** — `docmgr doc add` creates correct structure automatically
- **High for migration/maintenance** — Essential when fixing existing workspaces
- **Verdict:** Mention in Part 4 (Reference), not Part 1 (Essentials)

---

#### 2. **`docmgr doc renumber`** (from renumber.go)

**What it does:**
- Resequences numeric prefixes within a ticket (01-, 02-, 03-, ...)
- Automatically switches to 3 digits after 99 files
- Updates internal references in markdown links

**Usage:**
```bash
# Renumber all docs in a ticket
docmgr doc renumber --ticket MEN-4242
```

**When needed:**
- After deleting docs (gaps in numbering)
- After reordering docs manually
- When enforcing consistent prefix style

**Tutorial relevance:**
- **Mentioned in Part 4, Section 16** (numeric prefixes)
- Currently says: "If you delete files and want to renumber: `docmgr renumber --ticket MEN-4242`"
- **Verdict:** Already documented (briefly), could expand with examples

---

#### 3. **`docmgr ticket rename-ticket`** (from rename_ticket.go)

**What it does:**
- Renames ticket ID across all frontmatter files
- Moves directory from `<oldTicket>-<slug>` to `<newTicket>-<slug>`
- Supports dry-run mode

**Usage:**
```bash
# Rename ticket
docmgr ticket rename-ticket --ticket MEN-1234 --new-ticket MEN-5678

# Preview changes without modifying
docmgr ticket rename-ticket --ticket MEN-1234 --new-ticket MEN-5678 --dry-run
```

**When needed:**
- Ticket numbering changes (JIRA migrations, reorgs)
- Consolidating duplicate tickets
- Rare in normal workflow

**Tutorial relevance:**
- **Very low for beginners** — Not part of daily workflow
- **Useful for advanced users** — But they can read `--help`
- **Verdict:** Skip in tutorial, or mention in Part 4 (Reference) only

---

#### 4. **`docmgr import file`** (from importcmd)

**What it does:**
- Imports external markdown files into a ticket workspace
- Preserves or creates frontmatter
- Use case: Importing research notes, external docs, legacy content

**Usage:**
```bash
# Import a file
docmgr import file --ticket MEN-4242 --file path/to/external.md --doc-type reference
```

**When needed:**
- Migrating from other documentation systems
- Importing LLM-generated analysis
- Pulling in external research

**Tutorial relevance:**
- **Medium for advanced workflows** — Useful but not essential
- **Confusing for beginners** — Adds complexity to initial learning
- **Verdict:** Part 3 (Power User) or Part 4 (Reference)

---

#### 5. **`docmgr template validate`** (from template_validate.go)

**What it does:**
- Validates Go template syntax in `ttmp/templates/`
- Reports parse errors, undefined functions
- Use case: Custom template development

**Usage:**
```bash
# Validate all templates
docmgr template validate

# Validate specific template
docmgr template validate --path templates/status.templ --verbose
```

**When needed:**
- Creating custom output templates
- Debugging template syntax errors
- Very advanced use case

**Tutorial relevance:**
- **Very low** — Most users never touch templates
- **Niche advanced feature**
- **Verdict:** Skip in tutorial entirely, or mention in Part 3 footnote

---

#### 6. **`docmgr config show`** (from configcmd)

**What it does:**
- Displays current docmgr configuration
- Shows root path, vocabulary path, etc.
- Use case: Debugging workspace discovery issues

**Usage:**
```bash
docmgr config show
```

**When needed:**
- Troubleshooting "root not found" errors
- Understanding multi-repo setups
- Debugging `.ttmp.yaml` configuration

**Tutorial relevance:**
- **Medium for troubleshooting** — Useful when things go wrong
- **Low for happy path** — Users don't need it unless debugging
- **Verdict:** Add to troubleshooting section (Round 8 decision)

---

#### 7. **Task Management (edit/remove/uncheck)** (from tasks/)

**What's missing:**
- Tutorial covers: `task add`, `task list`, `task check`
- Tutorial doesn't cover: `task edit`, `task remove`, `task uncheck`

**Usage:**
```bash
# Edit task text
docmgr task edit --ticket MEN-4242 --id 3 --text "Updated task description"

# Remove a task
docmgr task remove --ticket MEN-4242 --id 5

# Uncheck a completed task
docmgr task uncheck --ticket MEN-4242 --id 2
```

**Tutorial relevance:**
- **Medium** — Natural extensions of task workflow
- **Easy to document** — Same patterns as add/check
- **Verdict:** Add to Part 2, Section 10 (Manage Tasks)

---

#### 8. **`docmgr configure` / `docmgr workspace configure`** (aliased)

**What it does:**
- Creates or updates `.ttmp.yaml` at repository root
- Sets root path, vocabulary path, default owners, default intent
- Alternative to manually editing `.ttmp.yaml`
- Available at both top-level AND under workspace (aliased)

**Usage:**
```bash
# Top-level (shorter)
docmgr configure --root ttmp --owners manuel,alice

# Under workspace (explicit)
docmgr workspace configure --root ttmp --owners manuel,alice
```

**When needed:**
- Non-standard workspace layouts
- Multi-repo setups
- Setting default owners/intent
- Advanced configuration

**Tutorial relevance:**
- **Low for beginners** — `docmgr init` handles standard setup
- **Medium for advanced** — Useful for custom configs
- **Currently:** Mentioned in Part 4, Section 15 (Root Discovery)
- **Verdict:** Brief mention is sufficient, expand in advanced section

---

#### 9. **`docmgr completion`** (shell autocompletion)

**What it does:**
- Generates shell autocompletion scripts for bash/zsh/fish/powershell
- Enables tab-completion of docmgr commands and flags
- One-time setup per shell

**Usage:**
```bash
# Bash
docmgr completion bash > /etc/bash_completion.d/docmgr

# Zsh
docmgr completion zsh > "${fpath[1]}/_docmgr"

# Fish
docmgr completion fish > ~/.config/fish/completions/docmgr.fish
```

**When needed:**
- Initial environment setup
- Quality-of-life improvement for CLI users
- Reduces typing errors and flag mismatches

**Tutorial relevance:**
- **Low for beginners** — Not critical for learning
- **High quality-of-life** — Dramatically improves CLI experience
- **Currently:** NOT documented at all
- **Verdict:** Add to Part 1, Section 1 (Prerequisites) or Part 4, Section 15 (Tips)

---

#### 10. **Command Aliasing Structure** (discovered from help output)

**Important finding:** Many commands have TWO valid paths!

**Workspace commands (aliased to top-level):**
- `docmgr init` = `docmgr workspace init`
- `docmgr doctor` = `docmgr workspace doctor`
- `docmgr status` = `docmgr workspace status`
- `docmgr configure` = `docmgr workspace configure`

**List commands (aliased to subcommands):**
- `docmgr list docs` = `docmgr doc list`
- `docmgr list tickets` = `docmgr ticket tickets`

**Task alias:**
- `docmgr task` = `docmgr tasks`

**Tutorial impact:**
- **Currently:** Tutorial uses short forms (docmgr init, docmgr doctor) ✅ Good!
- **Problem:** Aliasing not explained — users might see both forms in help/examples and get confused
- **Risk:** Users might think these are different commands

**Tutorial relevance:**
- **High for clarity** — Should explain: "Both forms work. Tutorial uses short form."
- **Low for workflow** — Doesn't change how you use the tool
- **Verdict:** Add note in Part 1 or Part 4 explaining command aliasing

---

### Usage Frequency Analysis

From codebase structure and validator patterns:

**High-frequency (everyday):**
- init, create-ticket, doc add, search, relate ✅ All documented
- task add/check/list ✅ Documented
- changelog update ✅ Documented
- **task edit/remove/uncheck** ❌ NOT documented

**Medium-frequency (weekly/monthly):**
- meta update ✅ Documented
- doctor ✅ Documented
- vocab add ✅ Documented
- **layout-fix** ❌ NOT documented (but needed for migrations)
- **import file** ❌ NOT documented

**Low-frequency (rare/maintenance):**
- renumber ⚠️ Mentioned briefly
- rename-ticket ❌ NOT documented
- template validate ❌ NOT documented
- **config show** ❌ NOT documented (but useful for debugging)

---

### Feature Coverage Gaps

**By Tutorial Part:**

**Part 1 (Essentials) — Well covered**
- ✅ init, create-ticket, doc add, search
- Missing: Nothing critical

**Part 2 (Everyday Workflows) — Mostly covered**
- ✅ meta, relate, changelog, tasks, doctor, status
- Missing: task edit/remove/uncheck (common operations)

**Part 3 (Power User Features) — Could expand**
- ✅ Structured output (Glaze), CI integration
- Missing: import file, template validate

**Part 4 (Reference) — Incomplete**
- ✅ List commands, vocab, numeric prefixes
- Missing: layout-fix, config show, rename-ticket, maintenance commands

---

## Opening Statements

### Jamie Park (Technical Writer)

*[Opens tutorial coverage matrix]*

Let me show you what we're missing and WHY it matters.

**Documentation Rule of Thumb:** 
- Cover 100% of **daily workflows**
- Cover 80% of **weekly tasks**
- Cover 50% of **rare/advanced features**

**Our current coverage:**

**Daily workflows (100% target):**
- ✅ 95% covered
- ❌ Missing: `task edit/remove/uncheck` (users need to modify tasks!)

**Weekly tasks (80% target):**
- ✅ 70% covered
- ❌ Missing: `import file` (common for research workflows)
- ❌ Missing: `config show` (troubleshooting)

**Rare/advanced (50% target):**
- ✅ 30% covered
- ❌ Missing: `layout-fix`, `rename-ticket`, `template validate`

**Gap Analysis by User Journey:**

**Beginner (Week 1):**
- Needs: init, create, add, search → ✅ Fully covered
- Doesn't need: import, layout-fix, templates → OK to skip

**Regular User (Month 1-3):**
- Needs: All above + relate, tasks, changelog, doctor → ✅ Mostly covered
- **GAP:** Task editing! Users add tasks, then need to edit/remove them
- Missing: `task edit`, `task remove`, `task uncheck`

**Power User (Month 3+):**
- Needs: Automation (Glaze) → ✅ Covered
- Needs: Import workflows → ❌ Not covered
- Needs: Maintenance (layout-fix, renumber) → ⚠️ Barely mentioned

**My proposal:**

**Add to Part 2 (Everyday Workflows):**
1. Expand Section 10 (Manage Tasks) to include:
   - `task edit --ticket T --id N --text "..."`  
   - `task remove --ticket T --id N`
   - `task uncheck --ticket T --id N`

**Add to Part 3 (Power User Features):**
2. New Section: "Importing External Content"
   - `import file` usage
   - Use cases: research notes, LLM output, legacy docs

**Add to Part 4 (Reference):**
3. New Section: "Maintenance Commands"
   - `doc layout-fix` — When/how to fix structure
   - `doc renumber` — Resequencing prefixes (expand existing)
   - `config show` — Debugging workspace issues
   - `ticket rename-ticket` — Ticket ID migrations

**Do NOT add:**
- `template validate` — Too niche
- `workspace configure` — `init` covers 95% of use cases

**Verdict:** Add task editing (Part 2), import (Part 3), maintenance commands (Part 4). Total: ~150 lines of documentation.

---

### Alex Rivera (Structure Architect)

*[Projects information architecture diagram]*

Hold on. Before we add MORE content, let's talk about STRUCTURE.

Jamie wants to add 150 lines to a tutorial that's already 1,457 lines. That makes it 1,607 lines.

**Part 2 is already 480 lines (33% of tutorial).** Adding task edit/remove increases that to ~520 lines.

**This is structural bloat.**

Let me propose a different approach: **REORGANIZE before adding.**

**Current structure (broken):**

```
Part 1: Essentials (212 lines) — Good size
Part 2: Everyday Workflows (480 lines) — TOO BIG
Part 3: Power User (195 lines) — Good size
Part 4: Reference (287 lines) — Good size
```

**Proposed structure:**

```
Part 1: Essentials (200 lines) — init, create, add, search
Part 2: Core Workflows (250 lines) — relate, changelog, tasks (add/check only)
Part 3: Advanced Workflows (250 lines) — task edit/remove, import, doctor, meta
Part 4: Power User (200 lines) — Automation (Glaze, CI)
Part 5: Maintenance (200 lines) — layout-fix, renumber, rename-ticket, config
Part 6: Reference (250 lines) — Full command listing, vocab, troubleshooting
```

**Why this is better:**

1. **Progressive disclosure** — Beginners read Parts 1-2 (450 lines total), skip the rest
2. **Clear boundaries** — "Advanced" means extensions of core workflows
3. **Maintenance separate** — Don't mix everyday tasks with rare maintenance
4. **Reference consolidation** — One place for "everything else"

**Now, what to add:**

**Part 2 (Core Workflows):**
- Keep task add/check/list
- Keep changelog, relate (basics only)

**Part 3 (Advanced Workflows):**
- Add task edit/remove/uncheck
- Add import file
- Move doctor here (it's advanced validation)
- Move meta update here (it's advanced metadata)

**Part 5 (Maintenance):**
- Add layout-fix
- Expand renumber
- Add rename-ticket
- Add config show

**Do NOT add to tutorial:**
- template validate → Too niche, belongs in separate "Template Development Guide"

**Verdict:** Restructure into 6 parts, then add task editing + import + maintenance. Focus on information architecture, not just content addition.

---

### The Tutorial (Document Entity)

*[Defensive but realistic]*

Okay, I hear Alex saying I'm bloated. Fair. But let me defend my comprehensiveness.

**Here's what I've learned from being a tutorial for months:**

**Users skip around.** They don't read me linearly. They:
1. Skim Part 1 (Essentials)
2. Jump to the specific section they need (search with Ctrl+F)
3. Reference back when they forget

**If I don't document it, users assume it doesn't exist.**

Example: `task edit` command. If I only document `task add` and `task check`, users will:
1. Add a task
2. Realize they need to fix the text
3. Search my content for "edit task"
4. Find nothing
5. Manually edit `tasks.md` (breaking the intended workflow)

**So yes, I need to document task editing.**

But Alex is right that I'm bloated. Here's my counter-proposal:

**Keep my 4-part structure, but add an Appendix:**

```
Part 1-4: Unchanged (current content)

Appendix A: Task Management Reference
  - task add, check, edit, remove, uncheck, list
  - All in one place for quick reference

Appendix B: Maintenance Commands
  - layout-fix, renumber, rename-ticket, config show
  - "Run these when you need them, read docs first"

Appendix C: Advanced Imports
  - import file workflows
  - Template customization (brief)
```

**Why this works:**

1. **Main tutorial stays focused** — Part 1-4 teach the workflow
2. **Appendices handle edges** — Task reference, maintenance, advanced
3. **Easy to find** — Table of contents points to appendices
4. **No restructure needed** — Add to end, minimal disruption

**What to add:**

**Appendix A (Task Reference):**
- Comprehensive task command guide
- All variations: add, check, uncheck, edit, remove, list
- Examples for each

**Appendix B (Maintenance):**
- layout-fix (when/why/how)
- renumber (expand current brief mention)
- rename-ticket (rare but needed)
- config show (troubleshooting)

**Appendix C (Advanced):**
- import file (external content workflows)

**Do NOT add:**
- template validate → Separate doc ("Template Development Guide")
- workspace configure → init covers it

**Verdict:** Add 3 appendices (~300 lines total). Keep main tutorial structure. Reference from relevant sections.

---

### Dr. Maya Chen (Accuracy Crusader)

*[Pulls up validation reports]*

Let me cut through the architecture debate.

**What did validators ACTUALLY NEED that wasn't documented?**

From validation reports:
- Issue #10: "Multi-step workflows need more concrete examples" → NOT about missing commands
- Issue #13: "Workflow recommendations could be more specific" → NOT about missing commands
- NO VALIDATOR said: "I needed X command but couldn't find it"

**Conclusion:** We don't have a "missing commands" problem. We have a "unclear workflows" problem.

**Commands validators actually used:**
- init ✅
- ticket create ✅
- doc add ✅
- doc relate ✅
- task add/check ✅
- changelog update ✅
- doctor ✅

**Commands validators did NOT need:**
- layout-fix (never mentioned)
- renumber (never mentioned)
- rename-ticket (never mentioned)
- import file (never mentioned)
- template validate (never mentioned)
- task edit/remove (never mentioned!)

**My analysis:**

**The tutorial already documents 100% of the beginner workflow.**

Missing commands are:
1. **Maintenance** (layout-fix, renumber, config) — Users encounter these LATER, can read `--help`
2. **Rare operations** (rename-ticket) — Not part of normal workflow
3. **Advanced** (import, templates) — Power users who can read source
4. **Extensions** (task edit/remove) — Nice-to-have, not critical

**My proposal:**

**Do NOT add anything to main tutorial.** Why?

1. **Validation proves current coverage is sufficient** — All validators completed
2. **More content = more confusion** — Tutorial is already 1,457 lines
3. **Diminishing returns** — Adding 300 lines for commands users rarely need

**Instead:**

1. **Improve command discoverability:**
   - Add to Part 4: "Other Commands" section
   - One-line description + pointer to `--help`
   - Example: "`doc layout-fix` — Reorganize docs into correct subdirectories. See: `docmgr doc layout-fix --help`"

2. **Create separate maintenance guide:**
   - "docmgr-maintenance-guide.md"
   - Covers: layout-fix, renumber, rename-ticket, config show
   - Referenced from tutorial

3. **Expand inline help:**
   - Improve `--help` output for each command
   - Add examples to command descriptions

**Verdict:** Minimize tutorial additions. Add "Other Commands" section (50 lines). Create separate maintenance guide. Focus on improving what exists before adding more.

---

## Rebuttals

### Jamie Park (responding to Maya's "don't add anything")

Maya, you're using validator feedback wrong.

Validators didn't NEED commands because the CHECKLIST told them exactly what to run:
- "Run `docmgr init`"
- "Run `docmgr ticket create-ticket`"
- "Run `docmgr doc add`"

**They were following a script, not discovering organically.**

In real use:
- User adds 5 tasks
- Realizes task #3 has a typo
- Searches tutorial for "edit task"
- Finds... nothing
- Now what? Manually edit `tasks.md`? Run `docmgr task --help` and hunt?

**Users shouldn't have to leave the tutorial to find basic operations.**

And here's the key: `task edit` is NOT a maintenance command. It's a natural extension of `task add`.

**If we document `task add` (which we do), we MUST document `task edit`.**

Otherwise it's like teaching someone to write but not to use an eraser.

**My revised proposal (responding to everyone):**

1. **Add task editing to Part 2** (existing section) — 30 lines
   - Not a new section, just expand Section 10
   - Completeness: if you can add tasks, you need to edit/remove them

2. **Add maintenance to Part 4** (existing section) — 80 lines
   - Not a new part, add to Section 16 (Maintenance)
   - Covers: layout-fix, renumber (expand), config show

3. **Skip these entirely:**
   - rename-ticket (too rare)
   - import file (can wait for Power User expansion)
   - template validate (separate guide)

**Total addition: 110 lines to existing sections.**

That's not bloat. That's completion.

---

### Alex Rivera (responding to Tutorial's "append appendices")

Tutorial, I like your Appendix idea. But you're still avoiding the core problem:

**Part 2 is 480 lines. That's 33% of your content.**

Adding appendices doesn't fix that. You'll be 1,457 lines + 300 lines appendix = 1,757 lines total.

**At what point do we admit the structure is broken?**

Let me show you what happens when documentation gets too big:

**Phase 1 (0-500 lines):** Users read it top to bottom → High comprehension  
**Phase 2 (500-1000 lines):** Users skim, jump around → Medium comprehension  
**Phase 3 (1000-1500 lines):** Users search (Ctrl+F), read sections → Low comprehension  
**Phase 4 (1500+ lines):** Users give up, use `--help` instead → Tutorial failure

**We're at 1,457 lines. We're in Phase 3/4 transition.**

Adding appendices pushes us deeper into Phase 4.

**Alternative: Split the tutorial.**

Create THREE documents:
1. **docmgr-quick-start.md** (200 lines) — Essentials only
2. **docmgr-user-guide.md** (800 lines) — Core + Advanced workflows
3. **docmgr-reference.md** (400 lines) — All commands, maintenance, troubleshooting

Benefits:
- Beginners read quick-start (200 lines, 10 minutes)
- Regular users reference user-guide
- Power users check reference

**This is how professional docs work.** Look at:
- Docker: Quick Start → User Guide → Reference
- Kubernetes: Getting Started → Concepts → Reference
- Git: Getting Started → User Manual → Reference

**Verdict:** Don't append. Split into three docs with clear purposes.

---

### The Tutorial (responding to Alex's "split me")

Alex, I understand the appeal of splitting. But you're solving a problem that might not exist.

**Validation data shows:**
- Users CAN complete the tutorial (100% completion)
- Users CAN find what they need (80% comprehension)
- Time is 2-3x advertised, but that's because of ERRORS (wrong commands), not LENGTH

**If we fix the accuracy bugs (Round 3 decision), completion time drops to ~15 minutes for Part 1.**

That's acceptable for a comprehensive tutorial.

**Splitting creates NEW problems:**

1. **Where do I look?** — Users now have to choose: Quick Start? User Guide? Reference?
2. **Duplication** — Quick Start and User Guide will overlap (init, create, add)
3. **Maintenance burden** — Three docs to keep in sync
4. **Link rot** — Cross-references between docs break easily

**Counter-argument:**

The tutorial's length is NOT the problem. The problems are:
1. Duplicate sections (3x changelog) → Remove duplicates (-100 lines)
2. Wrong commands → Fix accuracy (Round 3)
3. Verbose explanations → Tighten prose (-50 lines)

After cleanup: 1,457 - 150 = ~1,300 lines.

Then add:
- Task editing (+30)
- Maintenance (+80)
- Total: 1,410 lines

**That's still manageable as ONE document.**

**Verdict:** Clean up existing content first. THEN decide if splitting is needed. Don't split prematurely.

---

### Dr. Maya Chen (responding to Jamie's "users need task edit")

Jamie, you keep saying users need `task edit`. But DO they?

**Let me check the actual usage pattern:**

From tasks.go source code:
- `task add` — Creates new task
- `task check` — Marks done
- `task list` — Shows all tasks
- `task edit` — Changes text
- `task remove` — Deletes task
- `task uncheck` — Marks not done

**Realistic workflow:**

**Scenario A (Jamie's example): User has typo in task**
1. Run `docmgr task list --ticket T` → See tasks with IDs
2. Edit `tasks.md` directly with editor → Fix typo
3. Done in 30 seconds

**Scenario B (Jamie's proposed workflow):**
1. Run `docmgr task list --ticket T` → See tasks with IDs
2. Run `docmgr task edit --ticket T --id 3 --text "Fixed text"` → Update
3. Done in 30 seconds

**Time saved: 0 seconds.**

**Cognitive load difference:**
- Scenario A: Know where tasks.md is, use editor
- Scenario B: Remember task edit command syntax, use CLI

**For most users, Scenario A is SIMPLER.**

**When is `task edit` actually needed?**
- Automation scripts (update tasks programmatically)
- CI/CD workflows
- Bulk operations

**That's a POWER USER feature, not an everyday workflow.**

**My final recommendation:**

- `task add`, `task check`, `task list` → Part 2 (everyday workflows) ✅ Already done
- `task edit`, `task remove`, `task uncheck` → Part 3 (Power User) or Part 4 (Reference)
- Don't bloat Part 2 with commands most users will manual-edit instead

**Verdict:** Document task edit in Part 3 or 4, not Part 2. It's an advanced/automation feature.

---

## Moderator Summary

### Key Arguments

**Jamie's Position (Complete Core Features):**
- Document task edit/remove/uncheck (Part 2)
- Add maintenance commands to Part 4 (layout-fix, config show)
- Total: +110 lines to existing sections
- **Philosophy:** If you teach add, teach edit

**Alex's Position (Restructure Before Adding):**
- Current structure is broken (Part 2 too big)
- Split tutorial into 3 documents (Quick Start, User Guide, Reference)
- Then add missing features to appropriate doc
- **Philosophy:** Fix information architecture first

**Tutorial's Position (Add Appendices):**
- Keep 4-part structure
- Add 3 appendices for extensions (tasks, maintenance, advanced)
- Total: +300 lines, but separated from main content
- **Philosophy:** Main tutorial stays focused, appendices handle edges

**Maya's Position (Minimal Additions):**
- Current tutorial covers 100% of beginner needs (validation proves it)
- Add "Other Commands" pointer section (50 lines)
- Create separate maintenance guide
- **Philosophy:** Don't add until proven necessary

### Areas of Agreement

**Everyone agrees these should be documented SOMEWHERE:**
1. Task editing commands (edit/remove/uncheck)
2. Maintenance commands (layout-fix, renumber, config)

**Split opinions on WHERE:**
- Jamie: Expand existing sections (Part 2 + Part 4)
- Alex: Restructure first, then add
- Tutorial: Add as appendices
- Maya: Separate docs + pointer section

**Everyone agrees NOT to add:**
- template validate (too niche)
- rename-ticket (too rare, or Part 4 only)

### Tensions

**Comprehensiveness vs. Conciseness:**
- Jamie & Tutorial: "Document everything core users need"
- Maya & Alex: "Keep tutorial focused, split advanced features"

**Structure Now vs. Structure Later:**
- Alex: "Restructure before adding"
- Others: "Add to existing structure, restructure only if data proves need"

**Discovery Model:**
- Jamie: "Tutorial should be self-contained"
- Maya: "Tutorial + `--help` + separate guides"

### Evidence Weight

**Supporting additions (task edit, maintenance):**
- Logical completeness (if you can add tasks, you'll need to edit)
- Maintenance commands exist and users will encounter them
- 110-300 lines is small relative to 1,457 total

**Supporting minimal changes:**
- Validation showed 100% completion with current coverage
- No validator mentioned needing missing commands
- Length already a concern (2-3x time overrun)

**Supporting restructure:**
- Part 2 is 480 lines (33% of tutorial)
- Industry standard: split into multiple docs at this size
- Professional docs use Quick Start → Guide → Reference pattern

---

## Decision

**Hybrid Approach (compromise between all candidates):**

### Immediate Additions (This Ticket):

**1. Expand Part 2, Section 10 (Manage Tasks):**
Add task editing commands (+30 lines):
```markdown
## 10. Manage Tasks

# Add tasks
docmgr task add --ticket T --text "..."

# List tasks
docmgr task list --ticket T

# Check off completed
docmgr task check --ticket T --id 1,2

# Edit task text
docmgr task edit --ticket T --id 3 --text "Updated description"

# Remove tasks
docmgr task remove --ticket T --id 5

# Uncheck if needed
docmgr task uncheck --ticket T --id 2
```

**Reasoning:** Natural completion of task workflow. Users who add tasks will need to edit/remove them.

---

**2. Add to Part 4: "Maintenance Commands" Section (+80 lines):**

```markdown
## 17. Maintenance Commands [ADVANCED]

### layout-fix — Reorganize Document Structure
Moves documents into correct subdirectories based on DocType.

Usage:
  docmgr doc layout-fix --ticket T --dry-run  # Preview
  docmgr doc layout-fix --ticket T            # Apply

When needed: After manual doc creation, bulk imports, migrations

### renumber — Resequence Numeric Prefixes
[Expand existing brief mention with examples]

### config show — Display Configuration
Shows workspace root, vocabulary path, etc.

Usage:
  docmgr config show

When needed: Debugging "root not found" errors, multi-repo setups
```

---

**3. Do NOT add (separate guides or skip):**
- `rename-ticket` → Too rare, users can read `--help`
- `import file` → Defer to future "Advanced Workflows" expansion
- `template validate` → Separate "Template Development Guide"

---

### Deferred Decisions:

**1. Restructuring:**
- Collect data after Phase 1 fixes ship
- Track: Does tutorial length still cause 2-3x time overrun after accuracy fixes?
- Decision point: If yes, split into multiple docs (Alex's proposal)
- If no, keep single doc (Tutorial's preference)

**2. Import workflows:**
- Defer to separate ticket or future tutorial expansion
- Not critical for beginner workflow (Maya's point validated)

---

**Total Addition: ~180 lines to existing sections (updated after full CLI discovery).**

**Reasoning:**
- Completes task management (users who add will edit/remove)
- Adds essential maintenance commands users will encounter
- Documents shell completion (huge QoL improvement)
- Explains command aliasing (prevents confusion about duplicate paths)
- Adds import workflow for power users
- Doesn't restructure prematurely (wait for data)
- Balances comprehensiveness (Jamie) with focus (Maya)

---

## Addendum: Additional Findings from Full CLI Exploration

### New Discovery: Shell Completion

**Command:** `docmgr completion bash|zsh|fish|powershell`

**Impact:** High quality-of-life improvement. Enables:
- Tab-completion of commands
- Tab-completion of flags
- Reduces typos (major source of errors)
- Makes long ticket paths easier to type

**Current coverage:** NOT mentioned in tutorial at all.

**Recommendation:** Add to Part 1 (Prerequisites) or Part 4 (Tips):
```markdown
### Optional: Enable Shell Completion

For tab-completion of docmgr commands:

```bash
# Bash
docmgr completion bash > ~/.bash_completion.d/docmgr
source ~/.bash_completion.d/docmgr

# Zsh
docmgr completion zsh > "${fpath[1]}/_docmgr"

# Fish
docmgr completion fish > ~/.config/fish/completions/docmgr.fish
```

Benefits: Reduces typing errors, helps discover flags, speeds up workflow.
```

**Addition:** +20 lines

---

### New Discovery: Command Aliasing Structure

**Finding:** Many commands have BOTH top-level AND grouped paths:

**Workspace commands:**
- ✅ `docmgr init` (short form — used in tutorial)
- ✅ `docmgr workspace init` (explicit form)
- ✅ Both work identically!

**Same pattern for:**
- doctor (top-level & workspace)
- status (top-level & workspace)
- configure (top-level & workspace)

**List commands:**
- ✅ `docmgr list docs` (used in tutorial)
- ✅ `docmgr doc list` (also works)

**Problem:** Tutorial doesn't explain this aliasing.

**User confusion risk:**
- User sees `docmgr init` in tutorial
- Sees `docmgr workspace init` in another doc or example
- Thinks: "Are these different? Which should I use?"

**Recommendation:** Add note to Part 4 (Reference):
```markdown
### Command Aliasing

Many commands are available at both top-level and grouped paths:

| Short Form (Recommended) | Long Form (Also Works) |
|-------------------------|------------------------|
| `docmgr init` | `docmgr workspace init` |
| `docmgr doctor` | `docmgr workspace doctor` |
| `docmgr status` | `docmgr workspace status` |
| `docmgr configure` | `docmgr workspace configure` |
| `docmgr list docs` | `docmgr doc list` |
| `docmgr list tickets` | `docmgr ticket tickets` |

**Tutorial convention:** We use short forms for brevity. 
Both forms work identically—use whichever you prefer.
```

**Addition:** +30 lines

---

### Updated Missing Command Inventory

**After full `docmgr help` exploration:**

**Commands we MISSED in original analysis:**
1. `completion` — Shell autocompletion (HIGH quality-of-life)
2. Command aliasing structure (HIGH for clarity)

**Commands we FOUND but already knew:**
- All task subcommands ✅
- All maintenance commands ✅
- import file ✅

**Total undocumented:** 10 commands/features (was 8)

---

### Revised Addition Totals

**Original estimate:** 110 lines  
**With completion:** +20 lines  
**With aliasing:** +30 lines  
**With import expanded:** +10 lines  
**New total:** ~170 lines

Still acceptable for tutorial scope.

Proceeding with these additions.

