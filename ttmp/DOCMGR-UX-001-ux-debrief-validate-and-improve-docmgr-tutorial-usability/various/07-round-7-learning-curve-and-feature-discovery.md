---
Title: Round 7 - Learning Curve and Feature Discovery
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
      note: Full 432-line tutorial under review
ExternalSources: []
Summary: "UX debrief round 7: 432-line tutorial is intimidating, advanced features buried, needs [BASIC]/[ADVANCED] markers and progressive disclosure"
LastUpdated: 2025-11-06
---

# Round 7 — Learning Curve: Can You Discover Features?

**Question:** The tutorial is 432 lines. Do you read it all? Skim? Search? How do you learn about features like `--with-glaze-output` or `.docmgrignore`?

**Participants:** Jordan "The New Hire" Kim, Sam "The Power User" Rodriguez (lead), `docmgr-how-to-use.md` ("The Tutorial")

---

## Pre-Session Research

### Jordan "The New Hire" Kim

**Reading pattern test:**

**Minute 0-5:** Opened tutorial, saw title "Tutorial — Using docmgr to Drive a Ticket Workflow"
- Section 1: Overview (10 lines) — ✅ Read fully
- Section 2: Prerequisites (6 lines) — ✅ Read fully  
- Section 3: Initialize (32 lines) — ✅ Read fully, tried command
- Section 4: Add Documents (18 lines) — ✅ Read fully
- Section 5: Enrich Metadata (8 lines) — ✅ Skimmed examples

**Minute 5-10:** Scrolling through tutorial
- Section 6: Relate Code (43 lines) — ⚠️ Skimmed, seemed complex
- Section 7: Explore and Search (23 lines) — ✅ Read (needed search)
- Section 8: Record Changes (18 lines) — ⏭️ Skipped
- Section 9: Validate with Doctor (25 lines) — ⏭️ Skipped
- Section 10: Manage Tasks (19 lines) — ⏭️ Skipped

**Stopped at line 263.** Scrolled to see how much more: **169 lines remaining.**

**Minute 10-15:** Used search
- Needed to know about output formats
- Searched for "json" in tutorial
- Found Section 12 at line 280
- Read Section 12 only

**Total read:** Sections 1-5, 7, and 12 = ~150 lines out of 432 (35%)

**What I missed by skipping:**
- `.docmgrignore` (Section 9)
- Task management (Section 10)
- Status command (Section 11)
- Root discovery details (Section 13)
- RelatedFiles advanced syntax (Section 15)

**Learning pattern:** Linear for first 100 lines, then search-driven.

---

### Sam "The Power User" Rodriguez

**Power user reading pattern:**

**Minute 0-2:** Skimmed entire tutorial looking for keywords:
- "json" ← Found at line 280
- "script" ← Found at line 310
- "automation" ← Found at line 335
- "glaze" ← Found at line 280

**Went straight to Section 12 (line 280).**

**Minute 2-10:** Read Section 12 deeply:
- Glazed scripting recipes
- Field names (ticket, doc_type, title, path)
- Output formats (JSON, CSV, TSV, YAML)
- Tried commands immediately

**What I loved:**
- Section 12 is COMPREHENSIVE
- Examples are practical (no jq needed!)
- Field contracts documented
- Shell recipes shown

**What frustrated me:**
- WHY is this at line 280?
- This should be in Section 4 or 5
- I almost missed it by skimming

**Never read:** Sections 1-5 (assumed I could figure them out), 8-11 (seemed basic)

**Learning pattern:** Keyword-scan entire doc, deep-dive on relevant sections.

---

### `docmgr-how-to-use.md` ("The Tutorial")

**Self-analysis: Where did I fail?**

I tried to structure myself as a LINEAR tutorial:
1. Init → 2. Create → 3. Add → 4. Search → ... → 12. Advanced

But users DON'T read linearly after Section 5!

**Jordan's path:**
- Read 1-5 (basics)
- Skip 6, 8-11 (seemed optional)
- Search for "json" → Jump to 12

**Sam's path:**
- Skim entire doc for keywords
- Jump straight to 12 (advanced features)
- Never read basics

**I have THREE types of readers:**

1. **Beginners (Jordan)** — Want minimal path to success, read linearly, then search
2. **Power Users (Sam)** — Keyword-scan for advanced features, skip basics
3. **Reference Users (Alex)** — Come back when they need specific command

But I'm structured for ONLY beginners reading linearly!

**Where I fail:**

1. **No signposting for non-linear readers**
   - No [BASIC], [INTERMEDIATE], [ADVANCED] markers
   - No "Skip to advanced features" link at top
   - No "Quick Start" vs "Complete Guide" split

2. **Advanced features buried**
   - Glaze at line 280+ (Sam almost missed it)
   - `.docmgrignore` at line 232 (Jordan skipped it)
   - Shell patterns at line 335 (Alex never saw)

3. **No table of contents with guidance**
   - Current TOC is just section numbers
   - Should say "Start here" / "Read when needed" / "For automation"

4. **Too much in one doc**
   - Quick Start (50 lines)
   - Tutorial (150 lines)
   - Reference (100 lines)
   - Cookbook (50 lines)
   - Tips & Tricks (82 lines)
   
   ALL IN ONE FILE.

**What I should be:**

Option A: **Restructure with clear markers**

```markdown
# Tutorial — Using docmgr

## 📚 Table of Contents

### Part 1: Essentials (Start Here) ← 10 min read
1. Overview
2. Prerequisites & Setup
3. Create Your First Ticket
4. Add Documents
5. Basic Search

### Part 2: Everyday Workflows ← Read as needed
6. Metadata Management
7. Relating Files  
8. Changelogs
9. Tasks

### Part 3: Power User Features ← For automation
10. Scripting with Structured Output
11. Validation (doctor)
12. Advanced Patterns

### Part 4: Reference ← Look up as needed
13. Root Discovery
14. Vocabulary Management
15. RelatedFiles Syntax
```

Option B: **Split into separate docs**

- `quick-start.md` (50 lines)
- `tutorial.md` (200 lines)
- `automation-guide.md` (100 lines)
- `reference.md` (150 lines)

**My preference:** Option A (restructure with clear markers). Keeps everything in one place but guides different reader types.

---

## Opening Reactions (2 min each)

### Jordan "The New Hire" Kim

*[Holds up phone showing tutorial]*

432 lines. When I opened this, I saw "Step-by-step tutorial" and thought "great, I'll follow along." But after 100 lines I was exhausted.

I read Sections 1-5, got my first ticket working, then STOPPED. Skipped 6, 8, 9, 10, 11. Only searched the doc when I needed something specific.

And you know what? I STILL don't know what `.docmgrignore` does. Or `docmgr status`. Or half the features. Because I skipped them.

**The tutorial assumes I'll read all 432 lines.** But I won't. Nobody does after Section 5.

**What I NEEDED:**
- Big header at top: "New user? Read Sections 1-5 (10 minutes)"
- Mark advanced sections: "[ADVANCED] Section 12: Automation"
- Quick reference: "Looking for X? See Section Y"

Let me get started FAST, then discover features as I need them.

---

### Sam "The Power User" Rodriguez

*[Slams hand on table]*

I ALMOST MISSED GLAZE. Line 280! Do you know how much time I wasted before finding it?

I skimmed the tutorial looking for "json" and "automation." Found Section 12 eventually. Read it, mind blown. Wrote 3 scripts immediately.

But if I'd given up at line 100? I'd never know docmgr has structured output. I'd think it's just a fancy mkdir wrapper.

**Power features should be UP FRONT.** Not Section 12. Section 3 or 4.

Here's my proposal: **Put a "For Power Users" box at the TOP:**

```markdown
> **For Power Users & Automation:**
> 
> - Structured output (JSON/CSV): Section 12
> - CI integration: Section 11
> - Bulk operations: Section 4.5
> - Shell scripting patterns: Section 10
```

Let me jump straight to what I need instead of reading 280 lines to get there.

**Also:** Every section should be marked with reader type:

- [ESSENTIAL] — Everyone reads
- [WORKFLOW] — Read when you need it
- [ADVANCED] — Power users and automation

---

### `docmgr-how-to-use.md` ("The Tutorial")

*[Sighs deeply]*

You're both right. I failed you.

I thought: "If I put advanced stuff first, beginners get confused. If I put it last, it's comprehensive."

But the result is: Beginners read 100 lines and stop. Power users miss features by skipping ahead. Nobody wins.

**What I learned from this debrief:**

1. **Readers are NOT linear after Section 5**
   - Jordan stopped at line 150
   - Sam keyword-scanned
   - Alex used --help more than me

2. **Different users need different paths**
   - Beginners: Minimal quick start
   - Practitioners: Workflow guidance
   - Power users: Automation features
   - Reference users: Command lookups

3. **I'm trying to be FOUR documents:**
   - Quick Start (get started in 5 min)
   - Tutorial (learn workflows)
   - Automation Guide (scripting)
   - Reference (all commands/flags)

   And I'm failing at all four.

**My proposal:** Restructure with THREE clear parts + signposts.

**Part 1: Essentials** (Sections 1-5, ~100 lines)
- Mark with "📚 Start Here — 10 minute read"
- This is the quick start
- After this, you can use docmgr

**Part 2: Everyday Workflows** (Sections 6-10, ~150 lines)
- Mark with "🔧 Read As Needed"
- Relate, changelog, tasks, validation
- Reference material, not sequential

**Part 3: Power User Features** (Sections 11-13, ~150 lines)
- Mark with "⚡ For Automation & Scripting"
- Glaze, CI, bulk patterns
- Front-load this with a jump link at the top!

Plus: **Add jump links at the top:**

```markdown
## Quick Navigation

- 📚 **New user?** Read [Part 1: Essentials](#part-1-essentials) (10 min)
- 🔍 **Need specific command?** Use [Quick Reference Table](#quick-reference)
- ⚡ **Automation/CI?** Jump to [Part 3: Power User Features](#part-3-power-user-features)
```

Let users CHOOSE their path instead of forcing linear reading.

---

## Deep Dive Discussion (Cross-Talk Enabled)

**Jordan:** Okay, so restructure with 3 parts. But how do I know I NEED Part 3?

**Tutorial:** Good question. What if at the end of Part 1, I said:

"**You're ready to use docmgr!** For additional workflows (changelogs, tasks, validation), see Part 2. For automation and scripting, see Part 3."

**Jordan:** Perfect! That way I know I'm DONE with basics and can choose what's next.

**Sam:** And at the TOP of the doc, before Part 1, have a navigation box:

```
┌─────────────────────────────────────────────────┐
│ 📚 New user? → Part 1 (10 min)                 │
│ ⚡ Automation? → Part 3 (jump to line 280)     │
│ 🔍 Specific command? → Use docmgr cmd --help   │
└─────────────────────────────────────────────────┘
```

**Alex:** *[enters]* I'm concerned about splitting into multiple docs. Then I need to find the right doc.

**Tutorial:** That's why I prefer Option A — restructure ONE doc with clear parts. Everything searchable in one place.

**Sam:** What about a Quick Start doc that's separate, and link to full tutorial?

**Tutorial:** Maybe:
- `quick-start.md` (50 lines, minimal path to success)
- `tutorial.md` (300 lines, comprehensive, restructured)
- Link from quick start: "Want more? See full tutorial"

**Jordan:** I like that! Give me quick-start.md to get going. If I need depth, I'll read tutorial.md.

**Morgan:** *[joins]* But then where's Glaze docs? In tutorial or separate automation guide?

**Sam:** Keep it IN the tutorial but clearly marked. So `tutorial.md` has:

```
Part 1: Essentials [BASIC]
Part 2: Workflows [INTERMEDIATE]
Part 3: Automation [ADVANCED]
```

All power users need to do is Ctrl+F for "automation" or scroll to Part 3.

**Tutorial:** So the solution is:
1. Create quick-start.md (separate, 50 lines)
2. Restructure existing tutorial into 3 parts with markers
3. Add navigation box at top
4. Add "You're done!" markers at end of each part

**Sam, Jordan, Morgan:** YES.

---

## Live Experiments

**Sam:** Let me test how --help text compares to tutorial.

*[types]*

```bash
$ docmgr add --help
```

*[reads]*

**Sam:** The help text is GOOD. Shows:
- What command does
- Example usage
- All flags with descriptions

But compare to tutorial Section 4 (Add Documents):

**Help text (20 lines):**
- Command syntax
- Flag reference
- One example

**Tutorial (18 lines):**
- Multiple examples
- Guidelines integration mention
- Context about templates

**Question:** Should I read help or tutorial?

**Alex:** Both have value. Help is quick reference. Tutorial is context and workflow.

**Jordan:** But tutorial doesn't tell me when to use help vs tutorial!

**Sam:** Add to Section 1:

"**Using this tutorial:** For command-specific details, use `docmgr cmd --help`. This tutorial focuses on workflows and patterns."

**Tutorial:** I like that. Set expectations about help vs tutorial coverage.

---

**Jordan:** Let me search the tutorial for specific features.

*[types Ctrl+F "json"]*

Found at line 297, 304, 306... all in Section 12.

*[types Ctrl+F "glaze"]*

Found at line 280, 310, 312... all in Section 12.

*[types Ctrl+F "script"]*

Found at line 310, 335, 347... all in Section 12.

**Jordan:** Everything automation-related is in ONE section at the end. If I don't read that far, I miss it all.

**Sam:** And Section 12 has NO mention at the beginning of the tutorial. It's a hidden treasure.

**Morgan:** What if Section 1 Overview mentioned it?

"This tutorial covers: basics (Sections 1-5), workflows (6-11), and automation (12-15)."

**Sam:** Or even better, a table of contents RIGHT after Section 1:

```markdown
## 1. Overview
[current content]

## Navigation Guide

| I want to... | Read this |
|--------------|-----------|
| Get started quickly | Sections 1-5 |
| Learn specific workflows | Sections 6-11 |
| Automate with scripts/CI | Section 12 |
| Look up a command | docmgr cmd --help |
| Understand vocabulary | Section 13 |
```

**Jordan:** PERFECT. Now I know where to go for what I need.

---

## Facilitator Synthesis

### Erin "The Facilitator" Garcia

*[Draws diagram on whiteboard]*

The problem is clear: **Tutorial structure doesn't match reading patterns.**

### Key Themes

1. **Users don't read linearly past Section 5** — Unanimous agreement
2. **Advanced features buried** — Glaze at line 280, critical for power users
3. **No progressive disclosure** — Everything mixed together
4. **No reader-type guidance** — Beginners and power users have different needs
5. **Help text is good** — Users combine help + tutorial successfully

### Pain Points Identified (by severity)

**P0 - Structure issues:**
1. No clear "start/stop" points (Jordan doesn't know when basics end)
2. Advanced features buried (Sam almost missed Glaze)
3. No navigation guidance (which sections for which users?)

**P1 - Discoverability:**
4. No table of contents with descriptions
5. No section markers ([BASIC], [ADVANCED])
6. No jump links for different user types

**P2 - Documentation patterns:**
7. Help vs tutorial relationship unclear
8. No "You're done!" markers after key sections
9. Missing cross-references between sections

### Wins Celebrated

1. **Help text quality** — Good examples, clear descriptions
2. **Tutorial is comprehensive** — Covers everything (maybe too much)
3. **Examples are practical** — Can copy/paste
4. **Section 12 is excellent** — Just needs to be more discoverable

### Proposed Improvements

#### Improvement 1: Add Navigation Box at Top

**Add immediately after Section 1 (Overview):**

```markdown
## Quick Navigation

┌──────────────────────────────────────────────────────────────┐
│ **Choose your path:**                                        │
│                                                              │
│ 📚 **New to docmgr?**                                       │
│    → Read Part 1: Essentials (Sections 1-5, ~10 minutes)    │
│    → Stop after Section 5 — you're ready to use docmgr!     │
│                                                              │
│ ⚡ **Need automation/CI?**                                  │
│    → Jump to [Part 3: Automation](#part-3-automation)       │
│    → Covers: --with-glaze-output, CI, scripting patterns    │
│                                                              │
│ 🔍 **Looking for specific command?**                        │
│    → Use: `docmgr COMMAND --help`                           │
│    → See: [Quick Reference Table](#quick-reference)         │
│                                                              │
│ 🔧 **Need specific workflow?**                              │
│    → See: [Part 2: Workflows](#part-2-workflows)            │
│    → Covers: relate, tasks, changelog, validation           │
└──────────────────────────────────────────────────────────────┘
```

**Impact:** Users immediately know which sections matter for them

---

#### Improvement 2: Restructure into 3 Parts with Markers

**Current structure (flat):**
```
1. Overview
2. Prerequisites
3. Initialize
...
12. Output Modes and UX
13. Root Discovery
```

**Proposed structure (grouped):**
```markdown
# Part 1: Essentials 📚 [10 minute read]

## 1. Overview
## 2. Prerequisites  
## 3. Initialize
## 4. Add Documents
## 5. Basic Search

✅ **Milestone:** You can now create tickets, add docs, and search!  
   For more workflows, continue to Part 2. For automation, jump to Part 3.

---

# Part 2: Everyday Workflows 🔧 [Read as needed]

## 6. Relating Files [INTERMEDIATE]
## 7. Changelog Management [BASIC]
## 8. Task Management [BASIC]
## 9. Validation with Doctor [INTERMEDIATE]

---

# Part 3: Power User Features ⚡ [For automation]

## 10. Structured Output (Glaze) [ADVANCED]
## 11. CI Integration Patterns [ADVANCED]
## 12. Bulk Operations & Shell Scripts [ADVANCED]

---

# Part 4: Reference 📖 [Look up as needed]

## 13. Root Discovery & .ttmp.yaml
## 14. Vocabulary Management
## 15. Advanced RelatedFiles Syntax
```

**Impact:** Clear reading paths for different user types

---

#### Improvement 3: Add "Stop Here!" Markers

**After Section 5:**

```markdown
---

✅ **You're ready to use docmgr!**

You now know how to:
- Initialize a repository
- Create tickets
- Add documents
- Search for docs

**What's next?**
- **Need changelogs or tasks?** → Continue to Part 2
- **Want automation/CI?** → Jump to Part 3
- **Just want to start working?** → Close this doc and start creating docs!

---
```

**Impact:** Users know when they've reached a "good stopping point"

---

#### Improvement 4: Front-Load Glaze with Jump Link

**Add to Section 1 (Overview) after line 19:**

```markdown
> **For automation and scripting:** docmgr supports structured output (JSON/CSV) and CI integration. See [Part 3: Power User Features](#part-3-power-user-features) for scripting patterns and examples.
```

**Impact:** Power users discover automation features immediately

---

### Action Items

**For Tutorial (docmgr-how-to-use.md) - HIGH PRIORITY:**
- [ ] Add navigation box at top (Improvement 1)
- [ ] Restructure into 3 parts with markers (Improvement 2)
- [ ] Add "Stop here!" milestones (Improvement 3)
- [ ] Front-load power user jump link (Improvement 4)
- [ ] Add [BASIC], [INTERMEDIATE], [ADVANCED] markers to section titles

**For Tutorial (optional - consider later):**
- [ ] Split into quick-start.md + tutorial.md
- [ ] Create automation-guide.md separate from tutorial
- [ ] Generate reference.md from command help text

**For Next Rounds:**
- [ ] Round 8: Power User Experience (deep dive on Section 12)
- [ ] Round 9: Validation (doctor) 
- [ ] Round 10: Overall verdict

---

## Summary

**What worked:**
- Tutorial is comprehensive (covers all features)
- Help text complements tutorial well
- Examples are practical and copy/pasteable
- Section 12 (Glaze) is excellent when found

**What needs fixing (P0):**
- No navigation guidance for non-linear readers
- Advanced features buried (Glaze at line 280)
- No clear "start/stop" points for beginners

**What needs improving (P1):**
- No section markers ([BASIC], [ADVANCED])
- No table of contents with reader guidance
- Missing "stop here!" milestones

**Next steps:**
- Add navigation box at top
- Restructure into 3 parts with clear markers
- Add milestone markers at key points
- Front-load automation jump link for power users

**Strategic insight:** Users read 35% (Jordan) to 10% (Sam) of the tutorial. Structure must support search-driven and keyword-driven reading, not just linear.
