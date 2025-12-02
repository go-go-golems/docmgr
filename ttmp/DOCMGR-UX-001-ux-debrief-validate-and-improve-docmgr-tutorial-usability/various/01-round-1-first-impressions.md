---
Title: Round 1 - First Impressions
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
RelatedFiles:
    - path: pkg/doc/docmgr-how-to-use.md
      note: Tutorial under review (432 lines)
    - path: cmd/root.go
      note: CLI help text structure
ExternalSources: []
Summary: "UX debrief round 1: First contact with docmgr — Can users get started?"
LastUpdated: 2025-11-06T13:33:27.329495941-05:00
---

# Round 1 — First Impressions: Can You Get Started?

**Question:** When you first open `docmgr-how-to-use.md`, can you figure out what to do? What are your first 5 minutes like?

**Participants:** Jordan "The New Hire" Kim (lead), Alex "The Pragmatist" Chen, `docmgr-how-to-use.md` ("The Tutorial")

---

## Pre-Session Research

### Jordan "The New Hire" Kim

**Commands tried:**

```bash
# First, I ran help
$ docmgr --help
# Output was clean! I saw "Quick usage: docmgr help how-to-use"

# Followed the breadcrumb
$ docmgr help how-to-use
# Got the tutorial in my terminal — nice!

# Tried to follow Section 3 literally
$ docmgr create-ticket --ticket TEST-001 --title "My first ticket" --topics test
# ERROR: No .ttmp.yaml found and no vocabulary.yaml
# Wait, I thought the tutorial said "run docmgr init if needed"?

# Went back, ran init first
$ docmgr init
# Created ttmp/, vocabulary.yaml, templates, etc.

# Tried create-ticket again
$ docmgr create-ticket --ticket TEST-001 --title "My first ticket" --topics test
# SUCCESS! Created ttmp/TEST-001-my-first-ticket/
```

**Confusion points documented:**

1. **Line 36-37 of tutorial:** "If your repository doesn't have a docs root yet... run `docmgr init`"
   - This is AFTER the create-ticket command
   - I already tried create-ticket and it failed
   - **Should this be in Prerequisites or Section 2?**

2. **What is a "ticket workspace"?** (Line 17)
   - Is "ticket" like a JIRA ticket? GitHub issue?
   - Or is it docmgr-specific terminology?
   - Took me 2 minutes to understand it's just an identifier

3. **Frontmatter** (mentioned line 47, 54)
   - I've never heard this word
   - Only figured out it means YAML header because of context
   - **Could use a tooltip or parenthetical**

4. **Section 3.1 "What this index is for"** (Lines 43-59)
   - This is 17 lines of explanation BEFORE I've created anything
   - I skipped it initially, came back later
   - **Feels like it breaks flow**

**Wins:**
- `docmgr --help` pointing to `docmgr help how-to-use` was PERFECT
- Tutorial in terminal = no context switching
- First command example (line 30-33) was copy/pasteable

**Time to first success:** 8 minutes (would be 3 if init was mentioned earlier)

---

### Alex "The Pragmatist" Chen

**Comparison to alternatives:**

| Task | docmgr | mkdir + vim | Notion |
|------|--------|-------------|--------|
| Create ticket workspace | `docmgr create-ticket` (1 command) | `mkdir -p ttmp/TICKET/{design,reference}` + create files (3-5 commands) | Click "New Page", fill template (2-3 min) |
| Add document | `docmgr add` (1 command) | `vim ttmp/.../design/my-doc.md` + write frontmatter manually | Click, type | 
| Time to first useful artifact | 8 min (incl. init) | 2 min | 5 min |

**Skeptical questions:**

1. **Does `docmgr` respect `.gitignore`?**
   - Tutorial doesn't mention this
   - Checked: it creates `.docmgrignore` (line 420) but ignores .git by default
   - **Not clear if search/doctor respect .gitignore**

2. **What's the overhead over time?**
   - Every doc has frontmatter (12+ lines)
   - Every action needs a `--ticket` flag
   - **Is this just YAML busywork?**

3. **Can I use this incrementally?**
   - What if I have existing markdown docs?
   - Tutorial doesn't cover migration
   - **Adoption friction unclear**

**Time-to-value assessment:**

- **Immediate value:** Structure enforcement, searchable metadata
- **Delayed value:** Validation (doctor), relationships (relate)
- **Break-even:** Probably after 5-10 tickets
- **Verdict:** Would try on one project before rolling out

---

### `docmgr-how-to-use.md` ("The Tutorial")

**Self-assessment (structure issues):**

I know I'm 432 lines long. I KNOW. But I'm trying to be comprehensive!

**What I'm proud of:**
- Section 1 is short (10 lines)
- Code examples everywhere (50+)
- "Heads-up" boxes for advanced topics

**Where I feel dense:**

1. **Section 3 mixes Quick Start with Deep Dive**
   - Lines 28-35: Quick "do this"
   - Lines 36-41: Conditional init (breaks flow)
   - Lines 43-59: Philosophy dump about index.md
   - **Users want to CREATE SOMETHING first, understand later**

2. **I use jargon without defining it:**
   - "Frontmatter" (line 47, 54, 84)
   - "Ticket workspace" (line 28, 36)
   - "Docs root" (line 23, 36)
   - "RelatedFiles" (line 48)
   - **Assumes prior knowledge**

3. **My flow is documentation-centric, not task-centric:**
   - Sections = "Initialize", "Add Docs", "Enrich Metadata"
   - NOT "How to document a feature", "How to track progress"
   - **Missing the "why" narrative**

4. **Advanced features buried:**
   - Glaze scripting at Section 12 (line 280+)
   - `.docmgrignore` at line 232 (inside Doctor section)
   - Suggestions (`--suggest`) at line 99
   - **Power users won't discover these**

**What I fear:**
- Users skim me, miss critical details (like needing `docmgr init`)
- Users get to Section 5 and think "this is too much overhead"
- Users never discover advanced features that would win them over

---

## Opening Reactions (2 min each)

### Jordan "The New Hire" Kim

*[Slumps in chair, looks tired]*

Okay, real talk? I got it working eventually, but I hit a wall in the first 3 minutes. The tutorial says "create a ticket" BEFORE it says "run init if needed." I tried the command, it failed with a cryptic error about `.ttmp.yaml`, and I had to scroll around looking for what I missed.

When I finally ran `docmgr init` and tried again, everything clicked! The help output was great, the commands worked, the structure made sense. But that first failure? Almost made me rage quit.

**Also:** What the hell is "frontmatter"? I Googled it. I'm a 6-month developer. Just say "YAML metadata" or something!

---

### Alex "The Pragmatist" Chen

*[Pulls up terminal, types while talking]*

I timed it. 8 minutes to create my first ticket. `mkdir` + `vim` takes me 2 minutes. So docmgr needs to save me 6+ minutes elsewhere or this is a net loss.

The structure it creates is nice — I'll give it that. The frontmatter is consistent, the directories are logical. But I keep asking: **when does this pay off?** Search? Validation? Relationships? The tutorial doesn't sell me on the ROI.

And don't get me started on the `--ticket` flag. I get why it's there, but if I'm in a ticket directory, can't it infer the ticket? Every. Single. Command. Needs. This. Flag.

**Verdict so far:** Jury's still out. Show me the magic.

---

### `docmgr-how-to-use.md` ("The Tutorial")

*[Defensive but earnest]*

Look, I know I screwed up the init ordering! Lines 36-41 should be in Section 2 or Prerequisites. I was trying to be "just-in-time" with information but I buried a CRITICAL prerequisite.

And yes, I use jargon. I was written by developers for developers and I assumed people know what frontmatter is. I'M SORRY, OKAY?

But here's what I want to say: I'm trying to teach EVERYTHING. Commands, concepts, workflows, edge cases, advanced features. Maybe I need to split into "Quick Start" and "Complete Reference"? Maybe I need a glossary?

**What I need from this debrief:** Tell me what to cut. Tell me what to move. Tell me where you got lost. I WANT to be better.

---

## Deep Dive Discussion (Cross-Talk Enabled)

**Jordan:** Can we talk about the init thing? Like, seriously, I wasted 3 minutes on an error that could have been prevented if Prerequisites said "Run `docmgr init` first if this is a new repo."

**Tutorial:** You're right. I KNOW you're right. But look at line 26 — I mention `docmgr help how-to-setup` for repository setup. Isn't that enough?

**Jordan:** NO! Because I'm following YOUR instructions linearly. Section 3 says "do this command." I do it. It fails. THEN you tell me about init. That's backwards!

**Alex:** *[interrupts]* Actually, let me pull up the help text real quick...

```bash
$ docmgr create-ticket --help
```

*[reads output]*

The help says `--root` defaults to 'ttmp' or `.ttmp.yaml` root. But it doesn't say "Hey, run `docmgr init` first!" The error message when it fails is also not great. It just says "no .ttmp.yaml found."

**Jordan:** EXACTLY! The error message doesn't guide me to the solution!

**Alex:** Okay, so we have TWO problems:
1. Tutorial ordering (init should be earlier or more prominent)
2. Error message UX (should suggest running init)

**Tutorial:** *[taking notes furiously]* Okay, okay. What if Section 3 started with "First time? Run `docmgr init`" in a big box, THEN the create-ticket command?

**Jordan:** YES. Or even better: what if Prerequisites said:

```
## 2. Prerequisites

- docmgr available on PATH
- A Git repository
- **Run `docmgr init` once per repository**
```

**Alex:** I like that. Elevate init to prerequisite status, not a conditional footnote.

**Tutorial:** *[sighs]* Fine. I'll do it. What else?

**Jordan:** The jargon! "Frontmatter", "ticket workspace", "docs root". Can we at least parenthesize the first use? Like "frontmatter (YAML metadata)" ?

**Tutorial:** I can do that. But won't it make me wordier?

**Alex:** You're already 432 lines. What's 5 more words of clarity?

**Tutorial:** *[grumbles]* Fair.

**Jordan:** Oh! And Section 3.1 — "What this index is for" — it's 17 lines of explanation BEFORE I've seen the thing. Can that move AFTER I create the ticket? Like "Now look at what was created..."?

**Alex:** Agreed. Show, THEN explain. Not explain, then show.

**Tutorial:** *[nods slowly]* That... that makes sense. I was trying to frontload understanding, but you're saying it's cognitive overload?

**Jordan:** Exactly! I want to succeed first, understand later.

---

## Live Experiments

**Alex:** Let me try something right now. I'm going to pretend I've never used docmgr and just run help.

*[types]*

```bash
$ docmgr --help
```

*[reads]*

Okay, the help is REALLY good. It says "Quick usage: `docmgr help how-to-use`" right at the top. That's the breadcrumb Jordan followed. So the discoverability path is solid.

*[types]*

```bash
$ docmgr help how-to-setup
```

*[reads]*

Okay, this OTHER tutorial talks about init, vocabulary, templates. So if someone reads how-to-setup first, they'd know. But the tutorial we're reviewing assumes you either:
1. Already ran init, OR
2. Will read how-to-setup separately

**Jordan:** But who reads TWO tutorials before trying ONE command?

**Alex:** Exactly. The "how-to-use" tutorial should be self-contained for the common path.

**Tutorial:** So... merge init into how-to-use, or make Prerequisites more explicit?

**Jordan:** Explicit Prerequisites. Don't make me read two docs.

---

## Facilitator Synthesis

### Erin "The Facilitator" Garcia

*[flips through notes, marks whiteboard]*

Alright, team. Here's what I'm hearing:

### Key Themes

1. **Init timing is a critical blocker** — Users hit it in first 3 minutes
2. **Jargon assumes prior knowledge** — Accessibility issue for juniors
3. **Show-then-explain > Explain-then-show** — Flow issue
4. **Discoverability path (help → tutorial) is excellent** — Big win
5. **Value proposition unclear early** — Alex is skeptical about ROI

### Pain Points Identified (by severity)

**P0 - Blocks getting started:**
1. Init ordering: Tutorial says "create-ticket" before ensuring init is run
2. Error message: When init not run, error doesn't guide user to solution

**P1 - Degrades experience:**
3. Jargon without definitions: "frontmatter", "docs root", "ticket workspace"
4. Section 3.1 breaks flow: 17 lines of explanation before seeing output

**P2 - Reduces long-term adoption:**
5. Value proposition unclear: Why use this vs mkdir/vim?
6. Advanced features buried: Glaze, suggestions, etc.

### Wins Celebrated

1. **`docmgr --help` → `docmgr help how-to-use` breadcrumb** — Perfect discoverability
2. **Tutorial accessible in terminal** — No context switching
3. **Copy/pasteable examples** — Lowers friction
4. **First command success feels good** — Once you get past init

### Proposed Improvements (Quick Wins)

#### Improvement 1: Elevate init to Prerequisites

**Before (lines 21-27):**

```markdown
## 2. Prerequisites

- `docmgr` available on PATH
- A Git repository with your codebase (so `RelatedFiles` paths make sense)

> For repository setup (including vocabulary), see: `docmgr help how-to-setup`.
```

**After:**

```markdown
## 2. Prerequisites

- `docmgr` available on PATH
- A Git repository

### First-Time Setup

If this is your first time using docmgr in this repository, run:

```bash
docmgr init
```

This creates:
- `ttmp/` (docs root directory)
- `vocabulary.yaml` (doc types, topics)
- `_templates/` and `_guidelines/`

> For advanced setup options, see: `docmgr help how-to-setup`.
```

**Impact:** Prevents init failure in Section 3

---

#### Improvement 2: Add Jargon Glossary / First-Use Parentheticals

**Quick fix:** First use of jargon gets parenthetical

- Line 47: "Summarizes what the ticket does (one‑line Summary in **frontmatter** (YAML metadata)..."
- Line 28: "This creates a **ticket workspace** (a directory for all docs related to this work item)..."

**Better fix:** Add Section 1.5 "Key Concepts" with 5-line glossary before Section 2

---

#### Improvement 3: Reorder Section 3 (Show → Explain)

**Current flow:**
1. Command (line 30-34)
2. Conditional init (line 36-41)
3. Explanation of index (line 43-59)

**Proposed flow:**
1. First-time notice: "Already ran `docmgr init`? Great! If not, see Prerequisites."
2. Command (create-ticket)
3. "Here's what was created:" (show directory structure)
4. Explanation of index (now they have context)

---

#### Improvement 4: Better Error Message When Init Not Run

**Current error:**
```
Error: no .ttmp.yaml found
```

**Proposed error:**
```
Error: No docs root found (missing .ttmp.yaml or ttmp/ directory)

To initialize docmgr for this repository, run:
  docmgr init

For more help: docmgr help how-to-setup
```

**Severity:** P0 (this is a CLI change, not tutorial change)

---

### Action Items

**For Tutorial (docmgr-how-to-use.md):**
- [ ] Move init to Prerequisites (Improvement 1)
- [ ] Add glossary or first-use parentheticals (Improvement 2)
- [ ] Reorder Section 3 to "show then explain" (Improvement 3)
- [ ] Add value proposition paragraph in Section 1 (addresses Alex's ROI question)

**For CLI (cmd/ error handling):**
- [ ] Improve error message when `.ttmp.yaml` not found (Improvement 4)
- [ ] Consider: auto-detect ticket from CWD to reduce `--ticket` flag usage

**For Next Round:**
- [ ] Test Section 2 (init UX) in detail
- [ ] Explore Section 4-5 (core workflow) with Morgan leading

---

## Proposed Improvements (Full Detail)

### Change 1: Prerequisites Section Rewrite

```markdown
## 2. Prerequisites

- `docmgr` available on PATH
- A Git repository

### First-Time Setup (Required)

If this is your first time using docmgr in this repository:

```bash
docmgr init
```

This creates:
- `ttmp/` — Your documentation root directory  
- `vocabulary.yaml` — Defines doc types, topics, and intents  
- `_templates/` — Document templates  
- `_guidelines/` — Writing guidelines  

You only need to run this once per repository.

> **Advanced:** For multi-repo setups or custom paths, see `docmgr help how-to-setup`.
```

### Change 2: Section 1 — Add Value Proposition

After line 19 ("Working discipline..."), add:

```markdown
**Why use docmgr?** It gives you:
- **Structured docs** that LLMs can navigate (metadata, relationships)
- **Searchable knowledge base** across all your tickets
- **Validation** that catches broken links, missing files, stale docs
- **Scriptable output** for automation and CI/CD

If you're used to `mkdir` and `vim`, docmgr adds structure overhead but pays back with discoverability and quality checks.
```

### Change 3: Section 3 — Reorder for Flow

Lines 28-59 become:

```markdown
## 3. Create Your First Ticket

```bash
docmgr create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket
```

This creates `ttmp/MEN-4242--normalize-chat-api-paths-and-websocket-lifecycle/` with:

```
ttmp/MEN-4242--.../
├── index.md        # Ticket overview (you're here)
├── tasks.md        # Todo list
├── changelog.md    # History of changes
├── design/         # Design docs
├── reference/      # Reference docs
└── various/        # Other docs
```

### Understanding index.md

The `index.md` file is your ticket's single entry point. It:
- Summarizes what the ticket does (one‑line Summary in **frontmatter** (YAML metadata) + Overview section)
- Points to key docs and code via `RelatedFiles`
- Serves as the anchor for validation checks

**Best practice:** Keep index.md concise (~50 lines). Put details in design/reference docs.
```

---

## Summary

**What worked:**
- Discoverability (`docmgr --help` → tutorial) is excellent
- Tutorial in terminal removes friction
- Examples are copy/pasteable

**What needs fixing (P0):**
- Init timing/ordering blocks first use
- Error messages don't guide to solution

**What needs improving (P1):**
- Jargon accessibility for junior developers
- Show-then-explain flow in Section 3
- Value proposition clarity

**Next steps:**
- Revise Prerequisites section
- Improve error messages in CLI
- Test revised tutorial with fresh user
