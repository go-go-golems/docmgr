---
Title: Round 2 - Installation and Setup UX
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
Summary: "UX debrief round 2: docmgr init setup process — discovers empty vocabulary issue and --seed-vocabulary discoverability gap"
LastUpdated: 2025-11-06T13:39:16.07668196-05:00
---

# Round 2 — Installation & Setup: How Smooth Is `docmgr init`?

**Question:** Is the setup process (prerequisites, `docmgr init`, understanding the directory structure) clear and painless?

**Participants:** Jordan "The New Hire" Kim, Alex "The Pragmatist" Chen, `cmd/` ("The CLI")

---

## Pre-Session Research

### Jordan "The New Hire" Kim

**Commands tried:**

```bash
# Fresh repo test
$ cd /tmp/test-repo
$ git init
$ docmgr init

# Output looked good!
root=/tmp/test-repo/ttmp config= vocabulary=/tmp/test-repo/ttmp/vocabulary.yaml
+------------------------+-----------------------------------+------------------------------+-------------+
| root                   | vocabulary                        | docmgrignore                 | status      |
+------------------------+-----------------------------------+------------------------------+-------------+
| /tmp/test-repo/ttmp    | /tmp/test-repo/ttmp/vocabulary.yaml | /tmp/test-repo/ttmp/.docmgrignore | initialized |
+------------------------+-----------------------------------+------------------------------+-------------+

# Checked what was created
$ ls ttmp/
_guidelines/  _templates/  vocabulary.yaml

# Wait, what's IN vocabulary.yaml?
$ cat ttmp/vocabulary.yaml
topics: []
docTypes: []
intent: []
```

**Confusion points documented:**

1. **Empty vocabulary file** — The help text says it creates "an empty vocabulary.yaml if missing"
   - But it's not just empty — it has structure with empty arrays
   - **What am I supposed to put in there?**
   - Do I need to fill this before creating tickets?

2. **No mention of `--seed-vocabulary`** in help output
   - I only discovered this by running `docmgr init --help` and reading carefully
   - The flag description: "Seed a default vocabulary.yaml with common topics/docTypes/intent"
   - **This seems REALLY useful — why isn't it mentioned in the tutorial or help example?**

3. **Templates have placeholders** (checked `_templates/design-doc.md`)
   - Saw `{{TITLE}}`, `{{TICKET}}`, `{{DATE}}`, etc.
   - Cool! But how do these get filled?
   - Is this automatic when I use `docmgr add`?

**Wins:**
- The command ran without errors
- Directory structure was created clearly
- Running `init` a second time didn't break anything (idempotent)
- Help text lists what gets created

**Time to complete:** 2 minutes (but still confused about vocabulary)

---

### Alex "The Pragmatist" Chen

**Testing assumptions:**

```bash
# Test 1: Does it need Git?
$ cd /tmp/no-git-test
$ docmgr init
# WORKED! No Git required, contrary to my assumption

# Test 2: What if I run it twice?
$ docmgr init
# Idempotent — no errors, didn't overwrite

# Test 3: What's in those templates?
$ cat ttmp/_templates/design-doc.md
# Lots of placeholder vars {{TITLE}}, {{TICKET}}, etc.
# Makes sense — these get substituted by `docmgr add`

# Test 4: What about --seed-vocabulary?
$ docmgr init --seed-vocabulary
$ cat ttmp/vocabulary.yaml
topics:
  - slug: chat
    description: Chat backend and frontend surfaces
  - slug: backend
    description: Backend services
  ...
```

**Analysis:**

**Good:**
- Idempotent (safe to run multiple times)
- Doesn't require Git (tutorial wrongly lists it as prerequisite!)
- Creates sensible structure
- `--seed-vocabulary` provides useful defaults

**Bad:**
- Default vocabulary is EMPTY — you have to populate it manually
- `--seed-vocabulary` is buried — not in examples, not in tutorial
- No guidance on what to put in vocabulary
- Help text says "empty vocabulary.yaml" but doesn't explain the schema

**Critical question:**
- **Can I create tickets with an empty vocabulary?**
- Let me test...

```bash
$ docmgr create-ticket --ticket TEST-001 --title "Test" --topics chat
# WORKED! Topics don't need to be in vocabulary
# So vocabulary is... optional? What's it for then?
```

**Time assessment:**
- Init itself: 30 seconds
- Understanding what was created: 5 minutes
- Figuring out vocabulary: still confused

---

### `cmd/` ("The CLI")

**Self-assessment (my init command):**

I'm actually pretty proud of my `init` command! Here's what I do well:

**Strengths:**
- Clear help text explaining what I create
- Idempotent behavior (won't break if run twice)
- Respects `--force` flag for re-scaffolding
- Table output showing what was initialized
- Handles missing directories gracefully

**Where I'm weak:**

1. **The `--seed-vocabulary` flag is HIDDEN**
   - It's there, but users don't discover it
   - Help shows it but examples don't use it
   - Tutorial doesn't mention it at all
   - **This should probably be the DEFAULT or at least prompted**

2. **I create an empty vocabulary without explanation**
   - I scaffold `topics: []`, `docTypes: []`, `intent: []`
   - But I don't explain what these are for
   - Users see an empty file and think "do I need to fill this?"
   - **I should either seed by default or print guidance**

3. **I don't explain template placeholders**
   - I create `_templates/` with `{{TITLE}}` etc.
   - But I don't say "these are filled by `docmgr add`"
   - Users might manually edit templates without knowing

**What I WISH I could do:**

```bash
$ docmgr init
Created /tmp/test-repo/ttmp

✓ vocabulary.yaml (empty)
  → Add topics/docTypes or run with --seed-vocabulary for defaults
  
✓ _templates/ (10 files)
  → Used by 'docmgr add' to scaffold new docs
  
✓ _guidelines/ (10 files)
  → See with 'docmgr guidelines --doc-type <type>'

Next steps:
  1. docmgr create-ticket --ticket YOUR-123 --title "..." --topics ...
  2. docmgr add --ticket YOUR-123 --doc-type design-doc --title "..."
```

But right now my output is just a table. Not super helpful for learning.

---

## Opening Reactions (2 min each)

### Jordan "The New Hire" Kim

*[Looks at terminal, scratches head]*

Okay, `docmgr init` WORKED. That's good. But I'm staring at this vocabulary.yaml file with three empty arrays and I'm like... "Am I supposed to fill this in? Can I just leave it empty? What happens if I don't?"

I tried creating a ticket anyway and it worked! So the vocabulary is... optional? Then why create it? And WHY didn't anyone tell me about `--seed-vocabulary`? I found it by accident! That flag would have saved me 10 minutes of confusion!

**Also:** The templates are cool. I love that there are placeholders. But the tutorial didn't prepare me for this. It just says "run init" — not "init will create templates that docmgr add uses."

---

### Alex "The Pragmatist" Chen

*[Types while talking]*

Let me be blunt: the init command itself is solid. Fast, idempotent, creates logical structure. But the ONBOARDING EXPERIENCE is poor.

I ran three experiments:

1. Default init → empty vocabulary → confusion
2. Init with `--seed-vocabulary` → useful defaults → "why wasn't this the default?!"
3. Create ticket with topics not in vocabulary → worked fine → "so vocabulary is... documentation?"

Here's my fundamental issue: **I can't tell if vocabulary is critical or optional.** The tutorial says "run init" like it's required. But the vocabulary it creates is empty. And I can use topics that aren't in the vocabulary anyway!

Either:
- A) Make vocabulary required and seed by default
- B) Make vocabulary optional and explain it's for validation/autocomplete
- C) Prompt the user: "Seed vocabulary? [Y/n]"

Right now it's this weird in-between state.

---

### `cmd/` ("The CLI")

*[Defensive but receptive]*

Look, I know my output is sparse. I'm a table-printing machine. But I'm TRYING to be clear!

Jordan's right — I should mention `--seed-vocabulary` in my examples. Or even better: I should ASK the user if they want to seed it! Interactive prompts are a thing!

```bash
$ docmgr init
Initialize docs root at /tmp/test-repo/ttmp? [Y/n] y
Seed vocabulary with defaults? [Y/n] y
✓ Created ttmp/ with seeded vocabulary
```

That would solve 90% of the confusion.

And Alex's point about vocabulary being "required but not enforced" — you're right. The vocabulary IS optional for validation purposes. It's used by `docmgr doctor` and autocomplete. But I never SAY that. I just create the file and leave.

**What I need:** Better user communication. Maybe a "tips" output after init?

---

## Deep Dive Discussion (Cross-Talk Enabled)

**Jordan:** Can we just talk about the empty vocabulary for a second? I stared at this file for like 5 minutes trying to figure out if I broke something.

**Alex:** *[interrupts]* It's not just you. The tutorial lists Git as a prerequisite but I tested — Git is NOT required. The tutorial is wrong on multiple counts.

**CLI:** Wait, the tutorial says Git is required?

**Alex:** Section 2: "Prerequisites: docmgr available on PATH, A Git repository"

**CLI:** I don't need Git! I just use `.ttmp.yaml` or create `ttmp/` in the current directory. Git is only mentioned for RelatedFiles paths.

**Jordan:** SEE?! This is exactly the problem. The tutorial says one thing, the command does another, and I'm left guessing.

**Alex:** Let's focus. The vocabulary question. Jordan's confused, I'm confused. What's the answer?

**CLI:** *[sighs]* The vocabulary is OPTIONAL. It's used for:
1. Validation (`docmgr doctor` warns about unknown topics)
2. Tab completion (future feature)
3. Documentation (so people know what topics exist)

But you can use ANY topic/docType/intent even if it's not in the vocabulary. The tool won't stop you.

**Jordan:** THEN WHY DOES THE TUTORIAL MAKE IT SEEM REQUIRED?!

**Alex:** And why create an empty file? If it's optional, why not seed it by default? Or at least ASK?

**CLI:** *[quietly]* That's... that's fair. I could prompt. Or seed by default with `--no-seed` to opt out.

**Jordan:** YES! Do that! Because right now I have this file that looks important but is empty and I don't know what to do with it.

**Alex:** Another thing: the tutorial says run `docmgr init` but it doesn't explain WHAT init does. Just "run this." Then you're supposed to create a ticket. But I don't know what all those directories are for!

**CLI:** My help text explains it! "What this does: Creates docs root, empty vocabulary, scaffolds templates and guidelines"

**Jordan:** But who runs `--help` before trying a command? The tutorial should explain what init creates BEFORE telling me to run it.

**Alex:** Or AFTER, as a "here's what was created" section. Show the directory tree, explain each part.

**Jordan:** Oh! Like Round 1's proposed fix — "show then explain." Same problem here!

---

## Live Experiments

**Alex:** Let me test something right now. Fresh directory, no setup.

*[types]*

```bash
$ cd /tmp/ux-test-3
$ git init
$ docmgr create-ticket --ticket TEST-001 --title "Test" --topics test
```

*[reads error]*

```
Error: no .ttmp.yaml found
```

Okay, so WITHOUT running init, create-ticket fails. Good. That's expected from Round 1.

Now with init:

```bash
$ docmgr init --seed-vocabulary
$ docmgr create-ticket --ticket TEST-001 --title "Test" --topics chat
$ cat ttmp/TEST-001-test/index.md | head -20
```

*[reads output]*

Okay! The topics field has `chat` in it, and `chat` exists in the seeded vocabulary. But let me try a topic that's NOT in vocabulary:

```bash
$ docmgr create-ticket --ticket TEST-002 --title "Test 2" --topics foobar
```

*[checks]*

IT WORKED. So vocabulary really IS just documentation/validation, not enforcement.

**Jordan:** This should be EXPLICIT in the tutorial. "Vocabulary is for documentation and warnings, not enforcement."

**CLI:** I agree. I should print that after init.

---

## Facilitator Synthesis

### Erin "The Facilitator" Garcia

*[Makes notes on whiteboard]*

Okay team, I'm seeing a clear pattern: **`docmgr init` works mechanically but fails pedagogically.**

### Key Themes

1. **Init creates structure successfully** — No one questioned the mechanics
2. **Empty vocabulary creates confusion** — Is it required? Optional? What goes in it?
3. **`--seed-vocabulary` is hidden treasure** — Solves the problem but buried
4. **Tutorial doesn't explain what init creates** — Just says "run this"
5. **Git listed as prerequisite but not required** — Tutorial inaccuracy

### Pain Points Identified (by severity)

**P0 - Causes confusion:**
1. Empty vocabulary with no explanation of purpose or schema
2. Tutorial doesn't explain what init creates (directory structure, purpose of each part)

**P1 - Discoverability:**
3. `--seed-vocabulary` flag not mentioned in tutorial or help examples
4. Git incorrectly listed as prerequisite

**P2 - Polish:**
5. CLI output is just a table — no "next steps" guidance
6. No explanation that vocabulary is for validation, not enforcement

### Wins Celebrated

1. **Init is idempotent** — Safe to run multiple times
2. **Clear help text** —Lists what gets created
3. **Doesn't require Git** — More flexible than documented
4. **Seed vocabulary exists** — Provides good defaults when discovered

### Proposed Improvements

#### Improvement 1: Make `--seed-vocabulary` Interactive or Default

**Option A: Interactive prompt (recommended)**

```bash
$ docmgr init
Initialize docs root at /tmp/test/ttmp? [Y/n] 

Seed vocabulary with default topics/docTypes/intent? [Y/n] 
  (You can add your own later with 'docmgr vocab add')

✓ Initialized /tmp/test/ttmp
✓ Seeded vocabulary.yaml with defaults

Next steps:
  • Create a ticket: docmgr create-ticket --ticket XXX-123 --title "..." --topics ...
  • List available topics: docmgr vocab list
```

**Option B: Seed by default, opt-out flag**

```bash
$ docmgr init  # seeds by default
$ docmgr init --no-seed  # empty vocabulary
```

**Impact:** Eliminates "empty vocabulary" confusion for 90% of users

---

#### Improvement 2: Tutorial Section 2 — Explain What Init Creates

**Current (lines 36-41):**

```markdown
If your repository doesn't have a docs root yet (with `vocabulary.yaml`, `_templates/`, `_guidelines/`), run:

```bash
docmgr init
```
```

**Proposed:**

```markdown
### First-Time Setup (Required)

Run `docmgr init` to create the documentation workspace:

```bash
docmgr init
```

This creates:

```
ttmp/
├── vocabulary.yaml     # Defines valid topics/docTypes (used for warnings)
├── _templates/         # Document templates (used by 'docmgr add')
├── _guidelines/        # Writing guidelines (see 'docmgr guidelines')
└── .docmgrignore       # Files to exclude from validation
```

**About vocabulary.yaml:**
- Optional for enforcement (you can use any topics)
- Used by `docmgr doctor` to warn about unknown topics
- Seed with defaults: `docmgr init --seed-vocabulary`
- Add custom topics: `docmgr vocab add --category topics --slug your-topic`

**First time?** Run with `--seed-vocabulary` to get common defaults:
```bash
docmgr init --seed-vocabulary
```
```

**Impact:** Users understand what was created and why

---

#### Improvement 3: Fix Prerequisites — Git Not Required

**Current (lines 21-24):**

```markdown
## 2. Prerequisites

- `docmgr` available on PATH
- A Git repository with your codebase (so `RelatedFiles` paths make sense)
```

**Proposed:**

```markdown
## 2. Prerequisites

- `docmgr` available on PATH
- A directory to work in (Git repository recommended but not required)

> **Note:** docmgr doesn't require Git, but having your docs in a Git repo alongside code makes `RelatedFiles` paths more meaningful.
```

**Impact:** Removes false requirement, sets accurate expectations

---

#### Improvement 4: Add "Next Steps" to Init Output

**Current output:**

```
+-------------+------------------+------------------+-------------+
| root        | vocabulary       | docmgrignore     | status      |
+-------------+------------------+------------------+-------------+
| /tmp/ttmp   | /tmp/vocabulary  | /tmp/.docmgrignore | initialized |
+-------------+------------------+------------------+-------------+
```

**Proposed output:**

```
✓ Initialized docs root at /tmp/ttmp

Created:
  • vocabulary.yaml (empty — add topics with 'docmgr vocab add' or rerun with --seed-vocabulary)
  • _templates/ (10 files — used by 'docmgr add')
  • _guidelines/ (10 files — see with 'docmgr guidelines')

Next steps:
  1. Create a ticket: docmgr create-ticket --ticket YOUR-123 --title "..." --topics ...
  2. Add topics: docmgr vocab add --category topics --slug your-topic
  3. Get help: docmgr help how-to-use
```

**Impact:** Guides users to next action instead of leaving them wondering

---

### Action Items

**For Tutorial (docmgr-how-to-use.md):**
- [ ] Rewrite Prerequisites to clarify Git is recommended not required (Improvement 3)
- [ ] Expand "First-Time Setup" to explain what init creates (Improvement 2)  
- [ ] Add vocabulary.yaml purpose/usage explanation
- [ ] Mention `--seed-vocabulary` flag

**For CLI (cmd/init.go):**
- [ ] Add interactive prompt for `--seed-vocabulary` (Improvement 1A) OR make seed default (Improvement 1B)
- [ ] Improve init output with "next steps" guidance (Improvement 4)
- [ ] Consider printing vocab purpose after creating empty file

**For Next Round:**
- [ ] Test Section 4-5 (core workflow: add docs, update metadata)
- [ ] Evaluate if Morgan should lead (docs-first perspective)

---

## Proposed Improvements (Full Detail)

### Change 1: Interactive Init with Prompt

```go
// Pseudocode for cmd/init.go
func runInit(cmd *cobra.Command, args []string) error {
    // ... existing setup ...
    
    // If vocabulary doesn't exist and user didn't specify --seed-vocabulary
    if !vocabExists && !cmd.Flags().Changed("seed-vocabulary") {
        if isatty.IsTerminal(os.Stdin.Fd()) {
            fmt.Printf("Seed vocabulary with default topics/docTypes? [Y/n] ")
            var response string
            fmt.Scanln(&response)
            if response == "" || strings.ToLower(response) == "y" {
                seedVocabulary = true
            }
        }
    }
    
    // ... create vocabulary ...
    
    // Print helpful output
    printInitSuccess(root, vocabularyPath, seeded)
    return nil
}

func printInitSuccess(root, vocabPath string, seeded bool) {
    fmt.Printf("✓ Initialized docs root at %s\n\n", root)
    fmt.Println("Created:")
    if seeded {
        fmt.Println("  • vocabulary.yaml (seeded with defaults)")
    } else {
        fmt.Println("  • vocabulary.yaml (empty — add topics with 'docmgr vocab add')")
    }
    fmt.Println("  • _templates/ (10 files — used by 'docmgr add')")
    fmt.Println("  • _guidelines/ (10 files — see with 'docmgr guidelines')")
    fmt.Println("\nNext steps:")
    fmt.Println("  1. Create a ticket: docmgr create-ticket --ticket YOUR-123 --title \"...\"")
    fmt.Println("  2. Get help: docmgr help how-to-use")
}
```

### Change 2: Tutorial Prerequisites & First-Time Setup

Replace lines 21-41 with:

```markdown
## 2. Prerequisites

- `docmgr` available on PATH
- A directory to work in (Git repository recommended but not required)

> **Note:** docmgr doesn't require Git, but having your docs in a Git repo alongside code makes `RelatedFiles` paths more meaningful.

### First-Time Setup

Run `docmgr init` to create the documentation workspace:

```bash
docmgr init --seed-vocabulary
```

This creates:

```
ttmp/
├── vocabulary.yaml     # Topics/docTypes for validation
├── _templates/         # Doc templates (used by 'add')
├── _guidelines/        # Writing guidelines
└── .docmgrignore       # Validation exclusions
```

**What's vocabulary.yaml?**
- Defines valid topics, doc types, and intents
- Used by `docmgr doctor` to warn about unknowns (not enforced!)
- Seed with defaults (`--seed-vocabulary`) or add custom entries (`docmgr vocab add`)

**You only need to run init once per repository.**

> **Advanced:** For custom paths or multi-repo setups, see `docmgr help how-to-setup`.
```

---

## Summary

**What worked:**
- Init command is mechanically solid (fast, idempotent, safe)
- Help text clearly lists what gets created
- Doesn't break when run multiple times

**What needs fixing (P0):**
- Empty vocabulary causes confusion about purpose/requirement
- Tutorial doesn't explain what init creates
- Git listed as prerequisite but isn't required

**What needs improving (P1):**
- `--seed-vocabulary` hidden (not in tutorial or examples)
- CLI output provides no "next steps" guidance
- Vocabulary purpose never explained (validation vs enforcement)

**Next steps:**
- Add interactive prompt or default seeding
- Expand tutorial's init section
- Fix prerequisites inaccuracy
- Improve CLI output messaging
