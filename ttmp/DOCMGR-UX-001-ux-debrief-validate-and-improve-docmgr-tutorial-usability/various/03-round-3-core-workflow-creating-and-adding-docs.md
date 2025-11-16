---
Title: Round 3 - Core Workflow Creating and Adding Docs
Ticket: DOCMGR-UX-001
Status: active
Topics:
    - ux
    - documentation
    - usability
DocType: various
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "UX debrief round 3: core workflow (create ticket, add docs, update metadata) — finds repetitive --ticket flags and verbose meta update paths"
LastUpdated: 2025-11-06T13:45:45.108629718-05:00
---

# Round 3 — Core Workflow: Creating & Adding Documents

**Question:** The bread-and-butter workflow (create ticket, add docs, update metadata) — is it intuitive? Too many steps? Right abstractions?

**Participants:** Jordan "The New Hire" Kim, Morgan "The Docs-First" Taylor (lead), `cmd/` ("The CLI")

---

## Pre-Session Research

### Jordan "The New Hire" Kim

**Workflow test: Document a small feature**

```bash
# Setup
$ cd /tmp/my-project
$ git init
$ docmgr init --seed-vocabulary

# Step 1: Create ticket
$ docmgr create-ticket --ticket FEAT-001 --title "Add dark mode" --topics chat,frontend

# Step 2: Add design doc
$ docmgr add --ticket FEAT-001 --doc-type design-doc --title "Dark Mode Design"

# Step 3: Add reference doc  
$ docmgr add --ticket FEAT-001 --doc-type reference --title "Color Tokens"

# Step 4: Add playbook
$ docmgr add --ticket FEAT-001 --doc-type playbook --title "Testing Dark Mode"

# Step 5: Update design doc summary
$ docmgr meta update --doc ttmp/FEAT-001-add-dark-mode/design/01-dark-mode-design.md --field Summary --value "Design for dark mode feature"
```

**Step count:** 5 commands to create a basic ticket workspace with 3 docs

**Observations:**

1. **Every command needs `--ticket FEAT-001`**
   - I typed `--ticket FEAT-001` FOUR times (steps 2-5)
   - That's 17 characters × 4 = 68 extra characters
   - Gets tedious fast

2. **meta update requires full path**
   - I had to type: `ttmp/FEAT-001-add-dark-mode/design/01-dark-mode-design.md`
   - That's 59 characters!
   - Tab completion helps but still verbose

3. **Numeric prefixes are auto-added**
   - My docs got `01-`, `02-`, `03-` automatically
   - This is NICE! Keeps things ordered
   - Tutorial doesn't mention this feature

4. **Topics auto-inherited from ticket**
   - I set `--topics chat,frontend` on the ticket
   - All docs automatically got those topics
   - Saved me from repeating it

**Confusion points:**

- **What if I'm already IN the ticket directory?**
  ```bash
  $ cd ttmp/FEAT-001-add-dark-mode
  $ docmgr add --doc-type design-doc --title "Another Doc"
  Error: --ticket is required
  ```
  Can't it infer the ticket from my CWD?

- **How do I update metadata on multiple files?**
  - Tutorial shows `meta update --ticket X --doc-type Y`
  - But that updates ALL docs of that type
  - What if I want to update just one doc's owner but need the path?

**Wins:**
- Commands are predictable (same flag names)
- Output tables are clear
- Templates get filled automatically
- Numeric prefixes keep things tidy

**Time:** 3 minutes to create structure, 2 minutes typing paths/flags = 5 minutes total

---

### Morgan "The Docs-First" Taylor

**Testing across 5 tickets to evaluate workflow at scale:**

```bash
# Created 5 tickets, each with 3-5 docs
$ for i in {1..5}; do
    docmgr create-ticket --ticket PROJ-00$i --title "Feature $i" --topics backend
    docmgr add --ticket PROJ-00$i --doc-type design-doc --title "Design $i"
    docmgr add --ticket PROJ-00$i --doc-type reference --title "Reference $i"  
    docmgr add --ticket PROJ-00$i --doc-type playbook --title "Playbook $i"
done

# That's 5 + 15 = 20 commands
# Typed --ticket 15 times
```

**Workflow pain points at scale:**

1. **Flag repetition is BRUTAL**
   - At 5 tickets × 3 docs each, I typed `--ticket PROJ-00X` 15 times
   - This doesn't scale
   - Shell variable helps but feels hacky:
     ```bash
     T=PROJ-001
     docmgr add --ticket $T --doc-type design-doc --title "X"
     docmgr add --ticket $T --doc-type reference --title "Y"
     ```

2. **No bulk operations**
   - What if I want to add the same 3 doc types to every ticket?
   - No `add-batch` or template workflow
   - Each `add` is a separate command

3. **Metadata updates are path-hell**
   - After creating 15 docs, I wanted to update summaries
   - Had to construct paths like:
     ```
     ttmp/PROJ-001-feature-1/design/01-design-1.md
     ttmp/PROJ-001-feature-1/reference/01-reference-1.md
     ...
     ```
   - Used `docmgr list docs --ticket PROJ-001 --with-glaze-output --select path` to get paths
   - But that's meta-tooling on top of the tool!

4. **CWD inference would help**
   - If I'm in `ttmp/PROJ-001-feature-1/`, why can't it infer `--ticket PROJ-001`?
   - Git does this (infers repo from CWD)
   - Would save SO much typing

**What DOES work well:**

- **Topic/Owner inheritance**
  - Set once on ticket, flows to all docs
  - Changed my mind on owners? Update ticket, docs follow
  - Good abstraction

- **Unknown doc types go to `various/`**
  - I tried `--doc-type til` (not in vocabulary)
  - Went to `various/01-til-doc.md`
  - DocType field still says "til"
  - Good! Doesn't force me into predefined buckets

- **Numeric prefixes** 
  - Automatically incrementing (01-, 02-, 03-)
  - Keeps files ordered in directory listings
  - Don't have to think about it

**Proposed improvement:**

Add a context-aware mode:
```bash
$ cd ttmp/PROJ-001-feature-1
$ docmgr add --doc-type design-doc --title "X"  # infers --ticket PROJ-001
$ docmgr meta update design/01-* --field Owners --value manuel  # glob support
```

**Time assessment:**
- 20 commands for 5 tickets × 3 docs = 8 minutes
- With CWD inference: ~5 minutes (40% faster)

---

### `cmd/` ("The CLI")

**Self-assessment:**

I provide three main commands for the core workflow:

1. **`create-ticket`** — Creates workspace structure
2. **`add`** — Adds document from template
3. **`meta update`** — Updates frontmatter fields

**What I do well:**

- **Consistent flag naming**
  - `--ticket`, `--doc-type`, `--title` across commands
  - Users learn once, apply everywhere

- **Smart defaults**
  - Topics/owners/status inherit from ticket
  - Templates auto-filled with placeholders
  - Numeric prefixes auto-incremented

- **Flexible doc-type handling**
  - Known types → proper subdirectory
  - Unknown types → `various/` with DocType preserved
  - No enforcement, just organization

**Where I'm repetitive:**

1. **`--ticket` flag everywhere**
   - Required on `add` even if you're in ticket directory
   - Required on `meta update` unless using `--doc`
   - I KNOW this is tedious
   - But how do I infer ticket from CWD safely?

2. **`meta update` needs full paths or wildcards**
   - `--doc ttmp/TICKET-slug/design/01-file.md`
   - 59+ character paths are common
   - Alternative: `--ticket X --doc-type Y` updates ALL docs (maybe too broad)
   - Missing middle ground: update ONE doc by name

3. **No batch operations**
   - Each `add` is separate
   - No "add these 3 doc types to this ticket"
   - Users write shell loops

**What I'm considering:**

**Option A: CWD-based ticket inference**
```bash
$ cd ttmp/PROJ-001-feature-1
$ docmgr add --doc-type design-doc --title "X"
# Infers --ticket PROJ-001 from directory name
```

**Risk:** What if they're in a subdirectory? What if directory renamed?

**Option B: Shorter path syntax for meta update**
```bash
$ docmgr meta update --ticket PROJ-001 --file design/01-* --field Summary --value "X"
# --file is relative to ticket dir
```

**Option C: Ticket context file**
```bash
$ echo "PROJ-001" > .docmgr-context
$ docmgr add --doc-type design-doc --title "X"  # reads ticket from context file
```

**Risk:** Another hidden file, another thing to track

**My preference:** Option A (CWD inference) with explicit `--ticket` override

---

## Opening Reactions (2 min each)

### Jordan "The New Hire" Kim

*[Flexes fingers from typing]*

Okay, the workflow WORKS. I get it. Create ticket, add docs, update metadata. Logical. Clean. But OH MY GOD the typing.

I typed `--ticket FEAT-001` four times in 5 minutes. And that's a SMALL ticket! Morgan created 5 tickets with 3 docs each and probably wanted to cry.

The `meta update` command is the worst offender. I had to type a 59-character path to update one field. FIFTY-NINE CHARACTERS. For ONE field update. 

If I'm already in the ticket directory, can't you just... know which ticket I'm working on? Git knows I'm in a repo. Make knows I'm in a project. Why can't docmgr know I'm in FEAT-001?

**What I loved:**
- Numeric prefixes (01-, 02-) — automatic, helpful
- Topic inheritance — set once, forget
- Unknown doc types handled gracefully

**What I hated:**
- Repeating `--ticket` over and over
- Paths in `meta update`

---

### Morgan "The Docs-First" Taylor

*[Pulls up spreadsheet with timing data]*

I documented 5 tickets with 3 docs each. 20 commands total. 8 minutes of typing. I timed it.

**The abstraction is RIGHT:**
- Tickets contain docs ✓
- Docs have types and metadata ✓
- Templates provide structure ✓

**The execution is VERBOSE:**
- Every command needs `--ticket`
- Metadata updates need full paths
- No bulk operations

Here's the thing: Jordan found it tedious at 1 ticket. I found it painful at 5 tickets. What about a team doing 50 tickets over 6 months? This will drive people back to `mkdir` and `vim`.

**The fix is obvious:**
1. Infer ticket from CWD when possible
2. Allow relative paths in `meta update` when `--ticket` is specified
3. Add batch operations or at least document shell loop patterns

I'm willing to trade explicitness for ergonomics here. Make `--ticket` optional when CWD provides context, require explicit `--ticket` otherwise.

**What I appreciated:**
- Topic inheritance (huge time saver)
- Unknown doc types to `various/` (flexible!)
- Numeric prefixes (automatic ordering)
- Consistent flag names

---

### `cmd/` ("The CLI")

*[Nods in agreement]*

You're both right. The `--ticket` repetition is my biggest design issue.

I was being EXPLICIT — every command is self-contained, no hidden state. But Jordan and Morgan are showing me that explicitness becomes tedious.

Here's my proposal: **CWD-based inference with explicit override**

```bash
# Explicit (always works)
$ docmgr add --ticket PROJ-001 --doc-type design-doc --title "X"

# Inferred (works if CWD is in a ticket directory)
$ cd ttmp/PROJ-001-feature-1
$ docmgr add --doc-type design-doc --title "X"
# Infers --ticket PROJ-001 from directory name

# Override (if CWD is in ticket but you want different ticket)
$ cd ttmp/PROJ-001-feature-1
$ docmgr add --ticket PROJ-002 --doc-type reference --title "Y"
```

**Implementation:**
1. Parse CWD path
2. If it contains `ttmp/TICKET-slug/` pattern, extract TICKET
3. Use as default for `--ticket` flag
4. Allow explicit `--ticket` to override

**For meta update:**
```bash
# Current (verbose)
$ docmgr meta update --doc ttmp/PROJ-001-x/design/01-file.md --field Summary --value "X"

# Proposed (shorter)
$ docmgr meta update --ticket PROJ-001 --file design/01-file.md --field Summary --value "X"
# --file is relative to ticket dir

# Or with CWD inference
$ cd ttmp/PROJ-001-x
$ docmgr meta update --file design/01-file.md --field Summary --value "X"
```

Sound reasonable?

---

## Deep Dive Discussion (Cross-Talk Enabled)

**Morgan:** Okay, CWD inference is great. But what about safety? What if I'm in the wrong directory?

**CLI:** That's why I'd keep explicit `--ticket` override. And I could print which ticket I inferred:

```bash
$ cd ttmp/PROJ-001-x
$ docmgr add --doc-type design-doc --title "Test"
Using ticket: PROJ-001 (inferred from /tmp/test/ttmp/PROJ-001-x)
Created: ttmp/PROJ-001-x/design/02-test.md
```

**Jordan:** I like that! Shows me what's happening but doesn't force me to type it.

**Morgan:** What if I `mv` the directory? Say I rename `PROJ-001-feature-1` to `PROJ-001-dark-mode`?

**CLI:** The ticket ID is in the directory name (`PROJ-001`), not the slug. So renaming `PROJ-001-feature-1` → `PROJ-001-dark-mode` still infers `PROJ-001`.

**Jordan:** Wait, can I rename directories freely?

**CLI:** Technically yes, but if you rename to something without the ticket ID prefix, I can't infer it anymore. And links might break.

**Morgan:** So there's a `renumber` or `rename` command for this?

**CLI:** There's `docmgr renumber` which resequences numeric prefixes and updates links. But directory renaming is manual.

**Jordan:** That's... fine? I mean, I probably shouldn't rename ticket directories anyway.

**Alex:** *[enters the conversation]* Can we talk about batch operations? Morgan mentioned wanting to add the same 3 doc types to multiple tickets.

**Morgan:** YES! Like:

```bash
$ docmgr add-batch --tickets PROJ-001,PROJ-002,PROJ-003 \
    --doc-types design-doc,reference,playbook \
    --titles "Design,Reference,Playbook"
```

**CLI:** That's... complex. What if one ticket fails? Do I roll back? What about different titles per ticket?

**Alex:** Maybe simpler: just document the shell pattern?

```bash
$ for t in PROJ-{001..003}; do
    docmgr add --ticket $t --doc-type design-doc --title "Design"
done
```

**Morgan:** Fine, but put that in the tutorial! Section 4-5 show individual commands but not patterns for repetitive tasks.

**Jordan:** Also, can we talk about the `meta update` path situation?

**CLI:** My proposal: add `--file` flag that's relative to ticket dir. Shorter than `--doc` with full path.

**Jordan:** MUCH better. So:

```bash
# Instead of
$ docmgr meta update --doc ttmp/PROJ-001-x/design/01-file.md --field Summary --value "X"

# I can do
$ docmgr meta update --ticket PROJ-001 --file design/01-file.md --field Summary --value "X"

# Or with CWD inference
$ cd ttmp/PROJ-001-x
$ docmgr meta update --file design/01-file.md --field Summary --value "X"
```

**Morgan:** And glob support?

**CLI:** Maybe. `--file "design/01-*"` to match multiple files?

**Jordan:** Yes! That covers my "update one doc" and Morgan's "update several docs" use cases.

---

## Live Experiments

**Morgan:** Let me test the current workflow and time it.

*[types]*

```bash
$ cd /tmp/test
$ time (
    docmgr create-ticket --ticket TEST-999 --title "Test" --topics test
    docmgr add --ticket TEST-999 --doc-type design-doc --title "Design"
    docmgr add --ticket TEST-999 --doc-type reference --title "Reference"
    docmgr add --ticket TEST-999 --doc-type playbook --title "Playbook"
)
```

**Output:**
```
real    0m1.253s
user    0m0.956s
sys     0m0.297s
```

**Morgan:** Okay, 1.25 seconds for 4 commands. Fast. But I typed 112 characters total. Let me measure typing time...

*[types same commands at normal speed]*

~15 seconds of actual typing. The bottleneck isn't the tool, it's me typing `--ticket TEST-999` three times!

**Jordan:** Try with a shell variable?

**Morgan:**

```bash
$ T=TEST-999
$ time (
    docmgr add --ticket $T --doc-type design-doc --title "Design"
    docmgr add --ticket $T --doc-type reference --title "Reference"
    docmgr add --ticket $T --doc-type playbook --title "Playbook"
)
```

**Output:**
```
real    0m0.892s
user    0m0.689s
sys     0m0.203s
```

**Morgan:** 40% less typing. Still feels hacky. But proves the point — if I remove `--ticket` repetition, it's much faster.

**CLI:** This is great data. The tool is fast (< 1.5s for 4 commands), but the UX overhead is typing flags.

---

## Facilitator Synthesis

### Erin "The Facilitator" Garcia

*[Reviews notes]*

Alright team, I'm seeing clear consensus: **The workflow is sound but the UX is verbose.**

### Key Themes

1. **Core abstraction is correct** — Tickets → Docs → Metadata hierarchy makes sense
2. **Flag repetition is the main pain** — `--ticket` typed 4-15 times per session
3. **Path verbosity in meta update** — 59+ character paths are common
4. **CWD inference would help** — Strong consensus on this feature
5. **Smart defaults work well** — Topic inheritance, numeric prefixes, unknown doc-type handling

### Pain Points Identified (by severity)

**P0 - Causes significant friction:**
1. `--ticket` flag required on every `add` command (repetitive typing)
2. `meta update` requires full paths (59+ characters common)

**P1 - Workflow gaps:**
3. No documented patterns for bulk operations (shell loops not in tutorial)
4. Can't infer ticket from CWD (forces explicit flags even in context)

**P2 - Tutorial gaps:**
5. Numeric prefix feature not explained
6. Topic inheritance not highlighted as a win
7. No examples of shell variable patterns for repetitive tasks

### Wins Celebrated

1. **Automatic numeric prefixes** — Keeps files ordered without manual intervention
2. **Topic/owner inheritance** — Set once on ticket, flows to all docs
3. **Unknown doc types handled gracefully** — Goes to `various/` but preserves DocType
4. **Consistent flag naming** — Learn once, apply everywhere
5. **Template auto-filling** — Placeholders substituted automatically

### Proposed Improvements

#### Improvement 1: CWD-Based Ticket Inference

**Implementation:**

```go
// Pseudocode for ticket inference
func inferTicketFromCWD() string {
    cwd := os.Getwd()
    // Match pattern: ttmp/TICKET-slug/...
    re := regexp.MustCompile(`ttmp/([A-Z]+-\d+)`)
    if match := re.FindStringSubmatch(cwd); match != nil {
        return match[1]  // Return TICKET part
    }
    return ""  // No inference possible
}

func getTicket(cmd *cobra.Command) (string, error) {
    // Explicit --ticket flag takes precedence
    if cmd.Flags().Changed("ticket") {
        return cmd.Flags().GetString("ticket"), nil
    }
    
    // Try CWD inference
    if inferred := inferTicketFromCWD(); inferred != "" {
        fmt.Fprintf(os.Stderr, "Using ticket: %s (inferred from CWD)\n", inferred)
        return inferred, nil
    }
    
    return "", errors.New("--ticket required (could not infer from CWD)")
}
```

**Usage:**

```bash
# Explicit (always works)
$ docmgr add --ticket PROJ-001 --doc-type design-doc --title "X"

# Inferred (works from ticket directory)
$ cd ttmp/PROJ-001-feature-1
$ docmgr add --doc-type design-doc --title "X"
Using ticket: PROJ-001 (inferred from CWD)
Created: ttmp/PROJ-001-feature-1/design/02-x.md
```

**Impact:** Reduces typing by ~40% for common workflows

---

#### Improvement 2: Relative Paths in `meta update`

**Add `--file` flag that's relative to ticket directory:**

```bash
# Current (verbose)
$ docmgr meta update --doc ttmp/PROJ-001-x/design/01-file.md --field Summary --value "X"

# Proposed (shorter)
$ docmgr meta update --ticket PROJ-001 --file design/01-file.md --field Summary --value "X"

# With CWD inference
$ cd ttmp/PROJ-001-x
$ docmgr meta update --file design/01-file.md --field Summary --value "X"

# With glob support
$ docmgr meta update --file "design/*" --field Status --value review
```

**Implementation:**

```go
func resolveFile(ticket, file string, root string) (string, error) {
    if filepath.IsAbs(file) || strings.HasPrefix(file, "ttmp/") {
        // Already absolute or full path
        return file, nil
    }
    
    // Resolve relative to ticket directory
    ticketDir := findTicketDir(ticket, root)
    return filepath.Join(ticketDir, file), nil
}
```

**Impact:** Reduces path typing from 59+ to ~20 characters

---

#### Improvement 3: Tutorial Examples for Bulk Operations

**Add Section 4.5 to tutorial:**

```markdown
### Working with Multiple Documents

**Pattern: Add same doc types to multiple tickets**

```bash
# Using shell variable
TICKET=PROJ-001
docmgr add --ticket $TICKET --doc-type design-doc --title "Design"
docmgr add --ticket $TICKET --doc-type reference --title "Reference"
docmgr add --ticket $TICKET --doc-type playbook --title "Playbook"
```

**Pattern: Create similar tickets in batch**

```bash
for i in {1..5}; do
    docmgr create-ticket --ticket PROJ-00$i --title "Feature $i" --topics backend
    docmgr add --ticket PROJ-00$i --doc-type design-doc --title "Design $i"
done
```

**Pattern: Update all docs of a type**

```bash
# Update all design-docs for a ticket
docmgr meta update --ticket PROJ-001 --doc-type design-doc --field Status --value review
```
```

**Impact:** Users discover patterns without frustration

---

#### Improvement 4: Highlight Smart Defaults in Tutorial

**Add callout box in Section 4:**

```markdown
> **Smart Defaults**
>
> Documents automatically inherit topics, owners, and status from their parent ticket. 
> This means you set these fields ONCE on the ticket, and all docs get them for free.
>
> Want to override? Use flags like `--topics`, `--owners` on the `add` command.
```

**Impact:** Users appreciate the feature instead of discovering it accidentally

---

### Action Items

**For CLI (high priority):**
- [ ] Implement CWD-based ticket inference (Improvement 1)
- [ ] Add `--file` flag for relative paths in `meta update` (Improvement 2)
- [ ] Print "Using ticket: X (inferred)" when CWD inference succeeds

**For Tutorial (medium priority):**
- [ ] Add Section 4.5 with bulk operation patterns (Improvement 3)
- [ ] Add callout highlighting topic/owner inheritance (Improvement 4)
- [ ] Mention numeric prefix feature explicitly
- [ ] Show shell variable pattern (`T=PROJ-001`) for repetitive tasks

**For Next Round:**
- [ ] Test Section 6 (Relate files) with Morgan leading
- [ ] Or jump to Section 7 (Search) if relate seems straightforward

---

## Proposed Improvements (Full Detail)

### Change 1: Tutorial Section 4 — Add Bulk Patterns

**Insert after current Section 4 (Add Documents):**

```markdown
## 4.5 Working with Multiple Documents

When creating many documents, use shell patterns to reduce repetition:

### Pattern 1: Multiple docs for one ticket

```bash
# Use a shell variable to avoid retyping ticket
TICKET=MEN-4242
docmgr add --ticket $TICKET --doc-type design-doc --title "Architecture"
docmgr add --ticket $TICKET --doc-type reference --title "API Contracts"
docmgr add --ticket $TICKET --doc-type playbook --title "Smoke Tests"
```

### Pattern 2: Same structure across multiple tickets

```bash
# Create 5 similar tickets with design docs
for i in {1..5}; do
    TICKET=PROJ-00$i
    docmgr create-ticket --ticket $TICKET --title "Feature $i" --topics backend
    docmgr add --ticket $TICKET --doc-type design-doc --title "Design $i"
done
```

### Pattern 3: Bulk metadata updates

```bash
# Update all design-docs for a ticket
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value review
```

> **Tip:** Once CWD inference is available (planned feature), you can omit `--ticket` 
> when running commands from within a ticket directory.
```

### Change 2: Highlight Inheritance Feature

**Add callout in Section 4:**

```markdown
## 4. Add Documents

```bash
docmgr add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
```

> **Smart Defaults:** Documents automatically inherit `Topics`, `Owners`, and `Status` 
> from their parent ticket. Set these once on the ticket, and all docs get them. 
> Override per-doc using `--topics`, `--owners`, or `--status` flags if needed.
```

---

## Summary

**What worked:**
- Core abstraction (tickets → docs → metadata) is sound
- Smart defaults (inheritance, numeric prefixes, unknown doc-types)
- Consistent flag naming across commands
- Template auto-filling works seamlessly

**What needs fixing (P0):**
- `--ticket` flag repetition (typed 4-15× per session)
- `meta update` path verbosity (59+ characters)

**What needs improving (P1):**
- No CWD-based ticket inference (forces explicit flags)
- Bulk operation patterns not documented
- Smart defaults not highlighted

**Next steps:**
- Implement CWD inference for `--ticket`
- Add `--file` flag with relative paths for `meta update`
- Document shell patterns in tutorial
- Highlight inheritance and numeric prefixes
