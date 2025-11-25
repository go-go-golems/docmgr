---
Title: Jamie's proposed debate question changes
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - tutorial
DocType: working-note
Intent: short-term
Owners:
    - jamie-park
RelatedFiles:
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/03-debate-questions.md
      Note: Original debate questions
ExternalSources: []
Summary: "Technical writer's perspective on which debate questions need changing and why."
LastUpdated: 2025-11-25
---

# Jamie's Proposed Question Changes

## Context

As a technical writer with 8 years shipping documentation, I've reviewed the 10 debate questions and want to propose changes that better reflect documentation best practices and measurable quality standards.

---

## Problems with Current Question Set

### Issue 1: Missing Readability and Cognitive Load
The current questions focus on accuracy and structure but don't address **readability**—the single biggest predictor of tutorial success. We need to ask: "Can users actually process this information?"

### Issue 2: No Question About Terminology Consistency
Three validation reports mention confusion about jargon (frontmatter, RelatedFiles, vocabulary). We need a dedicated question about terminology and definitions.

### Issue 3: "Minimum Viable Fix" (Q4) Is Too Vague
Q4 asks "what 3-5 fixes give biggest improvement?" but doesn't define "improvement." Are we measuring:
- Time to complete tutorial?
- Error rate?
- User satisfaction?
- Return rate (do they come back to re-read)?

We need measurable success criteria BEFORE deciding what to fix.

### Issue 4: Missing Information Architecture
No question asks: "Can users find what they need?" Navigation, search, cross-references, and progressive disclosure are absent from the debate.

---

## Proposed Changes

### REPLACE Question 4 (Minimum Viable Fix)

**Old Q4:** "If we only fix 3-5 issues, which ones give us the biggest improvement in beginner success?"

**New Q4:** "What does 'success' mean for this tutorial? How do we measure if our fixes actually help beginners?"

**Why:**
- Forces us to define measurable criteria FIRST (time-to-task, error rate, satisfaction)
- Establishes baseline for before/after comparison
- Prevents arguing about "high ROI" without agreeing on ROI definition
- Standard technical writing practice: define success metrics before optimization

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "We need: time to complete Part 1, commands that error out, and post-tutorial confidence survey."
- **The Validation Checklist** (Inspector) — "I can measure: completion time, issues encountered, confusion points logged."
- **The Three Beginners** (Collective) — "Success is: we finish without getting stuck, understand what we did, feel confident trying more."
- **Sam Torres** (Empathy) — "Success is when beginners don't feel stupid. Measure 'I felt lost' moments."

**Decision Point:** Agree on 3-5 measurable success criteria that we'll track before/after fixes.

---

### REVISE Question 5 (Priority After Metrics)

**Revised Q5:** "Based on our success metrics and validation data, which fixes should we implement? (Priority order, not effort analysis)"

**Why:**
- With AI tools, implementation effort is less relevant than impact
- Focus on: what will measurably improve the metrics we defined in Q4?
- Use validation data to prioritize by user pain, not developer hours

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "Fix anything that blocks task completion or causes errors. Metrics tell us what matters."
- **Dr. Maya Chen** (Accuracy) — "Command accuracy fixes eliminate a whole class of errors. Start there."
- **The Three Beginners** (Collective) — "We struggled most with: wrong commands, duplicate sections, unclear errors. Fix those."
- **The Validation Checklist** (Inspector) — "The data shows: 100% of testers hit wrong commands, 66% confused by duplicates."

**Decision Point:** Ordered list of fixes based on measured impact on success metrics.

---

### ADD New Question 6b (After Tutorial Length)

**New Q6b:** "How do we handle jargon, definitions, and terminology consistency?"

**Why:**
- All three validation reports mention confusion about terms (frontmatter, RelatedFiles, vocabulary)
- Glossary exists but appears too late (after terms are used)
- No consistency in capitalization (RelatedFiles vs. related files, Ticket vs. ticket)
- Technical writing standard: define terms at first use or link to glossary

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "We need: glossary moved to top, inline definitions at first use, consistent capitalization in style guide."
- **Sam Torres** (Empathy) — "Beginners stumbled on 'frontmatter' and 'docs root.' Those need immediate definitions or we lose them."
- **The Three Beginners** (Collective) — "We kept wondering: what's a ticket workspace? Is it different from a ticket? Confusion."
- **The Tutorial** (Document) — "I have a glossary! But it's in Section 2. Maybe move it earlier?"

**Decision Point:** 
1. Where does glossary go? (Before Part 1? As sidebar?)
2. Which terms get inline definitions?
3. Create style guide for term capitalization?

---

### REPLACE Question 8 (Error Messages)

**Old Q8:** "Beginners see 'Error: no changes specified' and think they failed. Should the tutorial explain every error message or just the common ones?"

**New Q8:** "How do we write error messages and troubleshooting guidance that actually helps beginners recover?"

**Why:**
- Old question focuses on tutorial scope ("explain all errors or just top 3?")
- Real issue: error messages THEMSELVES are confusing ("no changes specified" sounds like user error)
- Technical writing best practice: improve the error messages at the source, then document edge cases
- Need to distinguish: CLI error message improvements vs. tutorial troubleshooting section

**Primary Candidates:**
- **Jamie Park** (Technical Writer) — "Fix the CLI error messages first. 'No changes to apply' is clearer than 'no changes specified.'"
- **Dr. Maya Chen** (Accuracy) — "Document what each error MEANS and what to do. CLI changes are out of scope for this ticket."
- **Sam Torres** (Empathy) — "Errors need: what happened, why it happened, what to do next. Every single one."
- **The Reset Script** (Saboteur) — "Or just fix me so beginners never hit this error!"

**Decision Point:**
1. Which errors need CLI message improvements? (separate ticket?)
2. Which errors need tutorial troubleshooting section?
3. Format for troubleshooting: inline after each section or dedicated appendix?

---

### MERGE Q10 + Q11 into Single Maintenance Question

**New Q10:** "How do we prevent this from happening again? (Both automation AND human ownership)"

**Why:**
- Automation alone can't prevent conceptual drift
- Human ownership alone can't scale to catch every command change
- Need both: CI/automation for regressions + ownership for maintenance
- Merge prevention (automation) and maintenance (humans) into one holistic strategy

**Primary Candidates:**
- **CI Robot** (Future Enforcer) — "Run tutorial commands in CI. Fail on syntax errors. But I can't catch outdated concepts."
- **Git History** (Drift Detective) — "Automation catches syntax. Humans catch drift. You need both."
- **Jamie Park** (Technical Writer) — "Three-layer defense: CI for commands, quarterly human review, assigned owner for updates."
- **The Tutorial** (Document) — "Give me both: automation to catch obvious breaks, and an owner who cares about quality."

**Decision Point:**
1. **Automation:** What CI checks? (command syntax, output validation, link checking)
2. **Ownership:** Who owns tutorial? (role: tech writer, doc team, maintainer)
3. **Review:** When do humans review? (quarterly? triggered by CLI changes? post-validation runs?)
4. **Process:** How do contributors update? (style guide, contribution guidelines)
5. **Metrics:** Rerun validation quarterly? Track metrics over time?

---

## Summary of Changes

### Questions to REPLACE:
- **Q4** (Minimum Viable Fix) → New Q4 (Define Success Metrics)
- **Q5** (Duplicate Content) → New Q5 (Priority by Metrics) — renumber old Q5-Q9 to Q6-Q10
- **Q8** (Error Message Scope) → New Q9 (Error Message & Troubleshooting Quality)
- **Q10+Q11** (Regression + Maintenance) → Merged into single Q11 (Prevention & Maintenance)

### Questions to ADD:
- **Q7** (Jargon & Terminology) — After structure decisions

### New Question Count: 12 questions (was 10)

---

## Revised Question Flow

```
Foundation (Q1-Q2): Go/no-go, Patch vs Restructure
  ↓
Triage (Q3): What's broken? (severity)
  ↓
Metrics (Q4): How do we measure success?
  ↓
Priority (Q5): Which fixes to implement? (based on metrics, not effort)
  ↓
Structure (Q6): Duplicates — Delete or Consolidate?
  ↓
Terminology (Q7): Jargon & Definitions (NEW)
  ↓
Length (Q8): Tutorial Length — Split or Trim?
  ↓
Mechanics (Q9): Error Messages & Troubleshooting (REVISED)
  ↓
Commands (Q10): Command Accuracy — Fix Scope?
  ↓
Workflow (Q11): Reset Script Problem
  ↓
Prevention & Maintenance (Q12): Automation + Human Ownership (MERGED)
```

---

## Rationale (Why Technical Writing Needs These)

### Define Success Metrics First (New Q4)
**Industry Standard:** Google's tech writing team measures:
- Time to complete task
- Error rate (commands that fail)
- Helpfulness rating (post-tutorial survey)
- Return rate (do users come back to re-read?)

Without metrics, "fix what matters" is just opinions. With metrics, we can prove ROI.

### Jargon Kills Tutorials (New Q6b)
**Research:** Nielsen Norman Group found that unexplained jargon is the #2 reason users abandon technical docs (after "can't find what I need").

Every undefined term adds cognitive load. Three beginners stumbled on the same terms—that's not coincidence, that's a pattern.

### Error Messages Are UX (New Q8)
**Industry Standard:** Microsoft's error message guidelines:
1. What happened
2. Why it happened
3. What to do next

"Error: no changes specified" fails all three. Compare to:
"No changes to apply. The files you specified are already related with these notes. Use --remove-files to unlink them."

That's user-centered writing.

### Maintenance Prevents Rot (New Q11)
**Industry Standard:** Every doc needs:
- Owner (role: "tech writer," "doc team," "tutorial maintainer")
- Review cadence (quarterly, or triggered by CLI changes)
- Contribution guidelines (how others can help)
- Style guide (tone, terminology, example format)

Without these, docs become orphaned and drift. This tutorial drifted because no one owned it.

---

## Recommendation

Accept these proposed changes. The revised 12-question debate will produce:
1. **Measurable success criteria** (Q4) — so we know if fixes work
2. **Metric-driven priorities** (Q5) — what to fix based on validation data, not effort
3. **Terminology strategy** (Q7) — reduce cognitive load from jargon
4. **Better error UX** (Q9) — help users recover from errors
5. **Holistic maintenance** (Q12) — automation + human ownership to prevent drift

This is how professional technical writing teams approach doc quality, adapted for AI-assisted development (effort becomes less relevant than impact).

---

## Next Step

Manuel: Do you accept these changes? If yes, I'll update `03-debate-questions.md` with the new/revised questions.

