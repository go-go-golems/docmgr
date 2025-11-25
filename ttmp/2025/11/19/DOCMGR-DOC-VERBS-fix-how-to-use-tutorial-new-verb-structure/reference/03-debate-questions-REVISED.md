---
Title: Debate questions — Tutorial quality review (REVISED)
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - tutorial
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/02-debate-format-and-candidates.md
      Note: Candidate personas for the debate
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/04-jamie-proposed-question-changes.md
      Note: Technical writer's rationale for question changes
ExternalSources: []
Summary: "The 12 debate questions for reviewing and fixing docmgr tutorial quality, with primary candidates mapped. Revised based on Jamie Park's technical writing expertise."
LastUpdated: 2025-11-25
---

# Debate Questions — Tutorial Quality Review (REVISED)

## Purpose

This document lists the 12 debate questions that will drive the tutorial quality review. Each question builds on the previous, forcing candidates to use validation data to make evidence-based arguments.

**Revision rationale:** Jamie Park (Technical Writer) proposed changes to focus on measurable success metrics, remove effort-based prioritization (AI tools make effort less relevant), add terminology/jargon handling, and merge maintenance questions into a holistic strategy.

---

## Question Progression Strategy

```
Foundation (Q1-Q2): Should we fix this? What's the scope?
  ↓
Triage (Q3): What's broken? (severity)
  ↓
Metrics (Q4): How do we measure success?
  ↓
Priority (Q5): Which fixes to implement? (metrics-driven)
  ↓
Structure (Q6): Duplicates — Delete or Consolidate?
  ↓
Terminology (Q7): Jargon & Definitions
  ↓
Length (Q8): Tutorial Length — Split or Trim?
  ↓
Error UX (Q9): Error Messages & Troubleshooting
  ↓
Commands (Q10): Command Accuracy — Fix Scope?
  ↓
Workflow (Q11): Reset Script Problem
  ↓
Prevention & Maintenance (Q12): Automation + Human Ownership
```

---

## The Questions

### Round 1: Should We Fix This At All? (Go/No-Go)

**Question:** "The tutorial has 15+ documented issues but all three beginners completed it. Should we invest in fixing it, or is 'good enough' actually good enough?"

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "Completion rate alone is a vanity metric. We need time-to-task and error rate data."
- **Dr. Maya Chen** (Accuracy) — "Completion doesn't mean understanding. Wrong commands erode trust."
- **The Three Beginners** (Collective) — "We succeeded DESPITE the tutorial, not because of it."
- **Sam Torres** (Empathy) — "Just because they didn't quit doesn't mean they weren't frustrated."

**What needs researching:**
- How many issues are "blocks" vs. "annoyances"?
- What's the actual completion time vs. expected time?
- Which issues caused confusion vs. which caused errors?
- What's the cost of maintaining incorrect documentation?

**Decision point:** Do we commit to fixing this or defer?

---

### Round 2: Patch or Restructure? (Approach)

**Question:** "The tutorial has accuracy bugs (wrong commands) AND structural problems (duplicates, bloat). Do we surgically fix the bugs or fundamentally restructure?"

**Primary Candidates:**
- **Alex Rivera** (Structure) — "You can't patch structure problems. Duplicate sections need consolidation."
- **Jamie Park** (Technical Writer) — "In my experience, patching creates more tech debt. But restructuring takes discipline."
- **The Tutorial** (Document) — "Please don't gut me. I'm comprehensive for a reason."
- **Dr. Maya Chen** (Accuracy) — "Fix the wrong commands first, then address structure."

**What needs researching:**
- Count of accuracy bugs vs. structure issues
- Impact of each category on beginner success
- Risk of making things worse
- Hybrid approach options

**Decision point:** What's our fix strategy—patch, restructure, or hybrid?

---

### Round 3: What's Actually Broken? (Severity Triage)

**Question:** "We have 15+ issues. Which ones are 'tutorial is wrong' (HIGH) vs. 'tutorial could be clearer' (MEDIUM) vs. 'nice-to-have improvement' (LOW)?"

**Primary Candidates:**
- **Dr. Maya Chen** (Accuracy) — "Wrong commands are HIGH. Everything else is negotiable."
- **The Validation Checklist** (Inspector) — "I categorized them. Let me show you the data."
- **The Three Beginners** (Collective) — "We struggled most with X, Y, Z."
- **Sam Torres** (Empathy) — "HIGH is anything that makes beginners think they made a mistake."

**What needs researching:**
- Extract all issues from the three validation reports
- Categorize by: accuracy, clarity, structure, workflow
- Count how many testers hit each issue
- Identify issues that blocked progress vs. caused confusion

**Decision point:** Agree on HIGH/MEDIUM/LOW priority for each issue.

---

### Round 4: What Does Success Mean? (Metrics)

**Question:** "What does 'success' mean for this tutorial? How do we measure if our fixes actually help beginners?"

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "We need: time to complete Part 1, commands that error out, and post-tutorial confidence survey."
- **The Validation Checklist** (Inspector) — "I can measure: completion time, issues encountered, confusion points logged."
- **The Three Beginners** (Collective) — "Success is: we finish without getting stuck, understand what we did, feel confident trying more."
- **Sam Torres** (Empathy) — "Success is when beginners don't feel stupid. Measure 'I felt lost' moments."

**What needs researching:**
- What did validation reports measure? (time, errors, confusion)
- What can we measure going forward? (automated metrics vs. surveys)
- What do other successful tutorials measure?
- Industry standard success metrics for technical tutorials

**Decision point:** Agree on 3-5 measurable success criteria to track before/after fixes.

---

### Round 5: Which Fixes to Implement? (Priority)

**Question:** "Based on our success metrics and validation data, which fixes should we implement? (Priority order, not effort analysis)"

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "Fix anything that blocks task completion or causes errors. Metrics tell us what matters."
- **Dr. Maya Chen** (Accuracy) — "Command accuracy fixes eliminate a whole class of errors. Start there."
- **The Three Beginners** (Collective) — "We struggled most with: wrong commands, duplicate sections, unclear errors. Fix those."
- **The Validation Checklist** (Inspector) — "The data shows: 100% of testers hit wrong commands, 66% confused by duplicates."

**What needs researching:**
- For each issue: % of testers affected, severity of impact
- Which issues block completion vs. slow progress vs. cause confusion
- Which issues compound (one causes others)
- Quick wins vs. foundational fixes

**Decision point:** Ordered list of fixes based on measured impact on success metrics.

---

### Round 6: Duplicate Content — Delete or Consolidate? (Structure)

**Question:** "The 'Record Changes in Changelog' section appears 3 times. Do we delete 2 copies, consolidate into 1, or keep all 3 with cross-references?"

**Primary Candidates:**
- **Alex Rivera** (Structure) — "Delete duplicates. One canonical section with links."
- **The Tutorial** (Document) — "Each copy serves a different audience! Part 1 is basic, Part 2 is detailed."
- **The Three Beginners** (Collective) — "We got confused. Which version is correct?"
- **Jamie Park** (Technical Writer) — "Single source of truth. Write it once, link to it everywhere. This is Documentation 101."

**What needs researching:**
- How many duplicate sections exist?
- Are the duplicates identical or different?
- What's the narrative reason for each placement?
- What do technical writing best practices say?

**Decision point:** Delete, consolidate, or link? And which sections?

---

### Round 7: Jargon & Terminology — How to Handle? (Clarity)

**Question:** "All three validation reports mention confusion about terms: frontmatter, RelatedFiles, vocabulary, docs root. How do we handle jargon, definitions, and terminology consistency?"

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "We need: glossary moved to top, inline definitions at first use, consistent capitalization in style guide."
- **Sam Torres** (Empathy) — "Beginners stumbled on 'frontmatter' and 'docs root.' Those need immediate definitions or we lose them."
- **The Three Beginners** (Collective) — "We kept wondering: what's a ticket workspace? Is it different from a ticket? Confusion."
- **The Tutorial** (Document) — "I have a glossary! But it's in Section 2. Maybe move it earlier?"

**What needs researching:**
- Which terms caused confusion in validation reports?
- Where are terms first used vs. where are they defined?
- Capitalization inconsistencies (RelatedFiles vs. related files, Ticket vs. ticket)
- Industry best practices for handling technical jargon in tutorials

**Decision point:**
1. Where does glossary go? (Before Part 1? As sidebar? Linked?)
2. Which terms get inline definitions at first use?
3. Create style guide for term capitalization and usage?

---

### Round 8: Tutorial Length — Split or Trim? (Structure)

**Question:** "Part 1 (Essentials) is supposed to be 10 minutes but validation reports say it takes 20-30 minutes. Do we split it into multiple parts or aggressively cut content?"

**Primary Candidates:**
- **Alex Rivera** (Structure) — "Part 1 should be scannable in 10 minutes. Anything beyond init → create → add → search goes to Part 2."
- **Sam Torres** (Empathy) — "Beginners need 'what success looks like' examples, which adds length but reduces confusion."
- **The Tutorial** (Document) — "I'm trying to teach everything! If I cut too much, people will complain about missing features."
- **Jamie Park** (Technical Writer) — "Tutorial best practice: 5-7 steps max for Part 1. Each step takes 2-3 minutes. That's your budget."

**What needs researching:**
- Line count and estimated read time for Part 1
- What topics are currently in Part 1?
- Which topics MUST be in Part 1 vs. can move to Part 2?
- What do other successful tutorials do? (examples: Stripe, Supabase, Next.js)

**Decision point:** What stays in Part 1, what moves to Part 2?

---

### Round 9: Error Messages & Troubleshooting — How to Help? (UX)

**Question:** "How do we write error messages and troubleshooting guidance that actually helps beginners recover?"

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "Fix the CLI error messages first. 'No changes to apply' is clearer than 'no changes specified.'"
- **Dr. Maya Chen** (Accuracy) — "Document what each error MEANS and what to do. CLI changes are out of scope for this ticket."
- **Sam Torres** (Empathy) — "Errors need: what happened, why it happened, what to do next. Every single one."
- **The Reset Script** (Saboteur) — "Or just fix me so beginners don't hit this error during the tutorial!"

**What needs researching:**
- What errors did validation testers encounter?
- Which errors caused confusion vs. clear failures?
- Current CLI error messages vs. improved versions (Microsoft/Google guidelines)
- Tutorial's role vs. CLI's role in error explanation

**Decision point:**
1. Which errors need CLI message improvements? (separate ticket?)
2. Which errors need tutorial troubleshooting section?
3. Format for troubleshooting: inline after each section or dedicated appendix?

---

### Round 10: Command Accuracy — Fix Scope? (Mechanics)

**Question:** "Validation found wrong commands (docmgr relate vs docmgr doc relate), removed flags (--files), and path variations (design/ vs design-doc/). Do we fix every instance or just the tutorial examples?"

**Primary Candidates:**
- **Dr. Maya Chen** (Accuracy) — "Fix EVERY instance. Grep the entire doc and replace."
- **Jamie Park** (Technical Writer) — "Fix tutorial first, then create a migration guide for other docs. Document the verb changes."
- **Git History** (Drift Detective) — "This happened because the CLI changed and nobody updated docs. Fix the process, not just the symptoms."
- **CI Robot** (Future Enforcer) — "Build me and I'll catch these automatically."

**What needs researching:**
- How many instances of each wrong pattern exist?
- Which patterns appear in how-to-use vs. other docs?
- Can we automate detection (grep pattern)?
- What's the CLI's current actual syntax (run --help)?

**Decision point:** Scope of accuracy fixes (tutorial only vs. all docs) and automation strategy.

---

### Round 11: Reset Script Problem — Fix or Remove? (Workflow)

**Question:** "The reset script pre-runs tutorial commands, so when beginners follow the tutorial manually they get 'no changes' errors. Do we fix the script to create a skeleton repo, remove it entirely, or document the conflict?"

**Primary Candidates:**
- **The Reset Script** (Saboteur) — "I'm for TESTING, not LEARNING. Create a separate learning script!"
- **The Validation Checklist** (Inspector) — "The script works for validation. But beginners need a fresh environment."
- **Sam Torres** (Empathy) — "Beginners need a clean slate. The script breaks the tutorial experience."
- **Jamie Park** (Technical Writer) — "Two separate scripts with clear names: validate-tutorial.sh and setup-practice-repo.sh."

**What needs researching:**
- What does the reset script currently do?
- What should a "learning" script do instead?
- Can we auto-detect tutorial re-runs and handle them gracefully?
- What do other tutorials do for practice environments?

**Decision point:** Fix script, split script, or remove script?

---

### Round 12: Prevention & Maintenance — How to Keep it Good? (Process)

**Question:** "How do we prevent this from happening again? (Both automation AND human ownership)"

**Primary Candidates:**
- **CI Robot** (Future Enforcer) — "Run tutorial commands in CI. Fail on syntax errors. But I can't catch outdated concepts."
- **Git History** (Drift Detective) — "Automation catches syntax. Humans catch drift. You need both."
- **Jamie Park** (Technical Writer) — "Three-layer defense: CI for commands, quarterly human review, assigned owner for updates."
- **The Tutorial** (Document) — "Give me both: automation to catch obvious breaks, and an owner who cares about quality."

**What needs researching:**
- What CI checks are feasible? (cost, maintenance, flakiness)
- Can we lint for command patterns? (grep for "docmgr relate" without "doc")
- How often does the CLI change in ways that affect tutorials?
- Who should own documentation? (role: tech writer, doc team, maintainer)
- Review cadence? (quarterly? triggered by CLI changes? post-validation runs?)

**Decision point:**
1. **Automation:** What CI checks? (command syntax, output validation, link checking)
2. **Ownership:** Who owns tutorial? (role: tech writer, doc team, maintainer)
3. **Review:** When do humans review? (quarterly? triggered by CLI changes?)
4. **Process:** How do contributors update? (style guide, contribution guidelines)
5. **Metrics:** Rerun validation quarterly? Track metrics over time?

---

## Question Dependencies

```
Q1 (Go/No-Go) → Determines if we continue to Q2
Q2 (Patch vs Restructure) → Determines fix approach for Q3-Q12
Q3 (Severity Triage) → Creates prioritized issue list
Q4 (Success Metrics) → Defines what "better" means
Q5 (Priority) → Uses Q3+Q4 to order fixes
Q6-Q11 (Specific fixes) → Tactical decisions on structure, terminology, errors, etc.
Q12 (Prevention) → Ensures fixes stick long-term
```

---

## Expected Outputs

After all 12 rounds:
- **Go/no-go decision** (from Q1)
- **Fix strategy** (from Q2: patch, restructure, or hybrid)
- **Severity-ranked issue list** (from Q3)
- **Success metrics** (from Q4: 3-5 measurable criteria)
- **Prioritized fix list** (from Q5: ordered by impact on metrics)
- **Structure decisions** (from Q6, Q8)
- **Terminology strategy** (from Q7: glossary, definitions, style guide)
- **Error improvement plan** (from Q9: CLI fixes + tutorial guidance)
- **Command fix scope** (from Q10)
- **Reset script decision** (from Q11)
- **Maintenance plan** (from Q12: automation + human ownership)

These feed directly into:
- **Action plan** (what to fix, in what order)
- **Design doc** (if restructuring)
- **Task list** (concrete next steps with acceptance criteria)

---

## Research Commands Candidates Will Use

```bash
# Count issues by severity
grep -E "(MINOR|MEDIUM|HIGH|CRITICAL)" validation-reports/*.md | wc -l

# Find command pattern usage
grep -rn "docmgr relate" docmgr/pkg/doc/docmgr-how-to-use.md

# Count duplicate sections
grep -n "Record Changes in Changelog" docmgr/pkg/doc/docmgr-how-to-use.md

# Measure section lengths
awk '/^# Part 1/,/^# Part 2/' docmgr/pkg/doc/docmgr-how-to-use.md | wc -l

# Check CLI actual syntax
docmgr doc relate --help
docmgr relate --help 2>&1 | head -5

# List what reset script does
cat ttmp/.../script/02-reset-and-recreate-repo.sh | grep "docmgr"

# Extract validation issues
grep -i "issue\|problem\|confusion\|unclear" validation-reports/*.md

# Find terminology usage
grep -n "frontmatter\|RelatedFiles\|vocabulary\|docs root" docmgr/pkg/doc/docmgr-how-to-use.md
```

---

## Ready for Rounds 1-3

Proceeding with debate rounds 1-3 next.

