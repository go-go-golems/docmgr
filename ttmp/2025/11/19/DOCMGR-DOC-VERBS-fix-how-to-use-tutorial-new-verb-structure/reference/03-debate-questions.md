---
Title: Debate questions — Tutorial quality review
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
ExternalSources: []
Summary: "The 10 debate questions for reviewing and fixing docmgr tutorial quality, with primary candidates mapped."
LastUpdated: 2025-11-25
---

# Debate Questions — Tutorial Quality Review

## Purpose

This document lists the 10 debate questions that will drive the tutorial quality review. Each question builds on the previous, forcing candidates to use validation data to make evidence-based arguments.

---

## Question Progression Strategy

```
Foundation (Q1-Q2): Should we fix this? What's the scope?
  ↓
Priority (Q3-Q4): What's broken vs. suboptimal? What fixes first?
  ↓
Structure (Q5-Q6): Do we restructure or patch? How do we handle bloat?
  ↓
Mechanics (Q7-Q8): How do we fix specific issues? (commands, errors, workflow)
  ↓
Prevention (Q9-Q10): How do we prevent regression? How do we validate success?
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
- Effort estimate for patch vs. restructure
- Risk of making things worse

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

### Round 4: What's the Minimum Viable Fix? (ROI)

**Question:** "If we only fix 3-5 issues, which ones give us the biggest improvement in beginner success?"

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "Quick wins: Fix verbs, add verification steps after each command, one concise changelog section."
- **Dr. Maya Chen** (Accuracy) — "All wrong commands must be fixed, not just the common one."
- **Sam Torres** (Empathy) — "Add 'what success looks like' examples. Beginners don't know if they're on track."
- **Alex Rivera** (Structure) — "Remove duplicate sections. Reduces confusion for zero cost."

**What needs researching:**
- Which fixes unblock the most beginners?
- Which fixes prevent the most confusion?
- Which fixes take <1 hour vs. >1 day?
- What's the 80/20 rule here?

**Decision point:** Agree on the 3-5 high-ROI fixes to do first.

---

### Round 5: Duplicate Content — Delete or Consolidate? (Structure)

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

### Round 6: Tutorial Length — Split or Trim? (Structure)

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
- What do other successful tutorials do?

**Decision point:** What stays in Part 1, what moves to Part 2?

---

### Round 7: Command Accuracy — Fix Scope? (Mechanics)

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

### Round 8: Error Messages — Add Guidance? (UX)

**Question:** "Beginners see 'Error: no changes specified' and think they failed, but it just means nothing needs updating. Should the tutorial explain every error message or just the common ones?"

**Primary Candidates:**
- **Sam Torres** (Empathy) — "Every error message needs an 'If you see X, it means Y, do Z' callout."
- **The Three Beginners** (Collective) — "The 'no changes' error confused all of us. That one needs explanation."
- **Jamie Park** (Technical Writer) — "Add a troubleshooting section. Top 3 errors inline, everything else in one place. Standard doc pattern."
- **The Reset Script** (Saboteur) — "Or fix me so beginners don't hit this error during the tutorial!"

**What needs researching:**
- What errors did validation testers encounter?
- Which errors caused confusion vs. clear failures?
- Can we improve the CLI error messages themselves?
- What's the tutorial's role vs. CLI's role in error explanation?

**Decision point:** Which errors need tutorial guidance? Should we fix the reset script or the tutorial?

---

### Round 9: The Reset Script Problem — Fix or Remove? (Workflow)

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

### Round 10: Regression Prevention — How? (Process)

**Question:** "We're fixing these issues now, but how do we prevent the tutorial from drifting again when the CLI changes?"

**Primary Candidates:**
- **CI Robot** (Future Enforcer) — "Run the tutorial commands in CI. Diff the output. Fail on syntax errors."
- **Git History** (Drift Detective) — "Link CLI changes to doc updates. Require doc PRs when verbs change."
- **The Validation Checklist** (Inspector) — "Rerun validation quarterly. I'm cheap and I catch real problems."
- **Dr. Maya Chen** (Accuracy) — "Document the verb mapping (old→new) and automate checks against it."

**What needs researching:**
- What CI checks are feasible? (cost, maintenance, flakiness)
- Can we lint for command patterns? (grep for "docmgr relate" without "doc")
- How often does the CLI change in ways that affect tutorials?
- What's the long-term maintenance burden?

**Decision point:** What automation do we build? What process changes do we make?

---

## Question Dependencies

```
Q1 (Go/No-Go) → Determines if we continue to Q2
Q2 (Patch vs Restructure) → Determines fix approach for Q3-Q8
Q3 (Severity Triage) → Creates prioritized list for Q4
Q4 (Minimum Viable Fix) → Identifies high-ROI fixes
Q5-Q6 (Structure) → Decides on restructuring scope
Q7-Q8 (Mechanics) → Decides on specific fix implementations
Q9 (Workflow) → Fixes testing/learning conflict
Q10 (Prevention) → Ensures fixes stick
```

---

## Expected Outputs

After all 10 rounds:
- **Severity-ranked issue list** (from Q3)
- **High-ROI fix list** (from Q4)
- **Structure decision** (from Q5-Q6)
- **Command fix scope** (from Q7)
- **Error guidance strategy** (from Q8)
- **Reset script decision** (from Q9)
- **CI/process changes** (from Q10)

These feed directly into:
- **Action plan** (what to fix, in what order)
- **RFC or design doc** (if restructuring)
- **Task list** (concrete next steps)

---

## Research Commands Candidates Will Use

```bash
# Count issues by severity
grep -E "(MINOR|MEDIUM|HIGH|CRITICAL)" validation-reports/*.md | wc -l

# Find command pattern usage
grep -rn "docmgr relate" docmgr/pkg/doc/

# Count duplicate sections
grep -n "Record Changes in Changelog" docmgr/pkg/doc/docmgr-how-to-use.md

# Measure section lengths
awk '/^# Part 1/,/^# Part 2/' docmgr/pkg/doc/docmgr-how-to-use.md | wc -l

# Check CLI actual syntax
docmgr doc relate --help
docmgr relate --help 2>&1 | head -5

# List what reset script does
cat docmgr/ttmp/.../script/02-reset-and-recreate-repo.sh | grep "docmgr"
```

---

## Next Step

Ready to start Round 1! Manuel, should I proceed with the debate or do you want to adjust the questions/candidates first?

