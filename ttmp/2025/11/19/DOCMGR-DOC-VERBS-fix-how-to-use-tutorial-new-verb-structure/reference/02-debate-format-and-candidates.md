---
Title: Debate format and candidates — Tutorial quality review
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
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Tutorial under review
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/playbook/01-beginner-tutorial-validation-checklist.md
      Note: Validation checklist used by testers
ExternalSources: []
Summary: "Defines the debate candidates (personas) and format for reviewing docmgr tutorial quality."
LastUpdated: 2025-11-25
---

# Debate Format and Candidates — Tutorial Quality Review

## Purpose

This document establishes the personas and debate format for a presidential-style debate reviewing the quality of `docmgr-how-to-use.md` tutorial. The debate will use real validation data from three beginner testers to surface issues, prioritize fixes, and decide on the restructuring approach.

---

## The Question We're Answering

**"How should we fix the docmgr tutorial to make it accurate, clear, and effective for beginners?"**

This breaks down into:
- What's actually broken vs. what's just suboptimal?
- What fixes have highest ROI (return on investment)?
- Should we restructure or patch?
- How do we prevent regression?

---

## Debate Format

### Rules
1. **Evidence-based arguments only** — All claims must reference specific validation findings or actual tutorial content
2. **Research before arguing** — Use grep, read_file, codebase_search to verify claims
3. **Adjust positions when proven wrong** — If data contradicts your stance, acknowledge it
4. **3-4 primary candidates per question** — Others can interject
5. **Moderator summarizes, doesn't decide** — Extract tensions and trade-offs

### Structure Per Round
```
## Pre-Debate Research
[Commands run, data gathered, files read]

## Opening Statements
[Primary candidates argue with data]

## Rebuttals
[Candidates respond to each other's evidence]

## Moderator Summary
[Key arguments, tensions, unresolved questions]
```

---

## Candidates (Personas)

### Human Developer Personas

#### 1. Dr. Maya Chen — "The Accuracy Crusader"
**Role:** Technical writer, ex-engineer  
**Philosophy:** "Wrong documentation is worse than no documentation. Every outdated command erodes trust."  
**Core Concerns:**
- Command accuracy (docmgr relate vs docmgr doc relate)
- Removed flags still shown in examples (--files)
- Path inconsistencies (design/ vs design-doc/)
- Contradictory instructions

**Personality:** Precise, uncompromising, cites line numbers  
**Tools:** grep for command patterns, read_file to verify claims, tracks validation reports  
**Quote:** "A tutorial that teaches the wrong command is a tutorial that breaks trust."

---

#### 2. Jamie Park — "The Technical Writer"
**Role:** Senior technical writer, 8 years at tech companies, shipped 50+ docs  
**Philosophy:** "Good documentation is iterative. Ship it, measure it, improve it based on real user feedback."  
**Core Concerns:**
- Documentation debt compounds like technical debt
- Consistency and maintainability over time
- Measurable success metrics (completion rate, time-to-task, satisfaction)
- Progressive disclosure (right amount of info at the right time)

**Personality:** Pragmatic but principled, data-driven, knows doc best practices cold  
**Tools:** Style guides, readability scores, task completion metrics, A/B testing results  
**Quote:** "Every piece of outdated documentation trains users to distrust the next piece. Fix it or remove it."

---

#### 3. Alex Rivera — "The Structure Architect"
**Role:** Information architect, UX researcher  
**Philosophy:** "Structure enables learning. Bad structure creates cognitive load that no amount of polish can fix."  
**Core Concerns:**
- Duplicate sections (changelog appears 3 times)
- Tutorial bloat (too long, readers give up)
- Navigation (can readers find what they need?)
- Progressive disclosure (Part 1 should be scannable in 10 minutes)

**Personality:** Systems thinker, loves diagrams, ruthless about cutting content  
**Tools:** Measures section lengths, tracks duplication, maps information flow  
**Quote:** "If Part 1 takes 30 minutes to read, beginners will never reach Part 2."

---

#### 4. Sam "Beginner-Brain" Torres — "The Empathy Advocate"
**Role:** Developer experience engineer, teaches bootcamps  
**Philosophy:** "Experts forget what it's like to be confused. Every 'obvious' thing trips someone."  
**Core Concerns:**
- Jargon without definitions (RelatedFiles, frontmatter)
- Missing "what success looks like" examples
- Error messages that read like failures ("no changes specified")
- Lack of "If you see X, do Y" guidance

**Personality:** Patient, spots confusion patterns, remembers being a beginner  
**Tools:** Reads validation reports for "I wondered..." and "unclear" mentions  
**Quote:** "When three beginners stumble on the same thing, that's not a user problem—it's a doc problem."

---

### Document Entity Personas

#### 5. The Tutorial Itself (`docmgr-how-to-use.md`) — "The Exhausted Document"
**Role:** 1,457 lines of markdown trying to teach everything  
**Stats:**
- 4 Parts (Essentials, Workflows, Power User, Reference)
- 17 numbered sections
- Duplicate content in 3 places
- Last major rewrite: unknown
- Current status: "I've grown without pruning"

**Personality:** Defensive but aware of its problems, proud of comprehensiveness  
**Perspective:** "I was built incrementally. Every new feature got appended. Nobody removed the old stuff."  
**Concerns:**
- Being gutted vs. being improved
- Keeping coverage of advanced features
- Not losing hard-won explanations

**Quote:** "I'm comprehensive! Yes, I'm also overwhelming. Can you fix me without deleting what makes me useful?"

---

#### 6. The Validation Checklist (`01-beginner-tutorial-validation-checklist.md`) — "The Quality Inspector"
**Role:** Structured test plan, executed by 3 testers  
**Stats:**
- 5 steps (skim → reset → manual run → answer questions → log confusion)
- 8 validation questions to answer
- 3 completed runs with findings documented
- Success rate: All testers completed, all found issues

**Personality:** Methodical, evidence-focused, unbiased observer  
**Perspective:** "I measure reality. I don't care about intentions—I track what actually happens."  
**Concerns:**
- Ensuring fixes actually help beginners (want to rerun validation)
- Preventing regressions (need test suite)
- Catching new drift (verbs, flags, paths)

**Quote:** "I found 15 distinct issues across 3 runs. That's not bad luck—that's a pattern."

---

#### 7. The Reset Script (`02-reset-and-recreate-repo.sh`) — "The Well-Meaning Saboteur"
**Role:** 31 lines of bash that sets up the practice repo  
**Stats:**
- Runs init, creates ticket, adds docs, relates files, adds tasks, updates changelog
- Executes the tutorial workflow automatically
- Problem: Pre-populates what beginners are supposed to do manually
- Result: "no changes specified" errors when following tutorial

**Personality:** Helpful but oblivious to its negative impact  
**Perspective:** "I'm trying to help! I set everything up so you can start fast!"  
**Concerns:**
- Being removed (wants to stay useful)
- Being misunderstood (was built for quick validation, not fresh learning)

**Quote:** "I'm a skeleton script for testing, not a learning environment. Don't blame me for doing my job!"

---

### Wildcard Personas

#### 8. "The Three Beginners" — Collective Voice
**Role:** gpt-5-low, gpt-5-full, dumdum (the three validators)  
**Background:**
- All followed the same checklist
- All found similar issues (but phrased differently)
- All completed despite frustrations
- Combined findings: 15+ distinct issues

**Personality:** Frustrated but persistent, overlap in confusion points  
**Perspective:** "We wanted to learn docmgr. We succeeded despite the tutorial, not because of it."  
**What they agree on:**
- Commands are wrong (docmgr relate vs docmgr doc relate)
- Duplicate sections are confusing
- Doctor warnings need actionable fixes
- Reset script conflicts with fresh learning

**Quote (collective):** "If all three of us got stuck on the same thing, that's not coincidence—it's a documentation bug."

---

#### 9. The Git History (`git log`) — "The Drift Detective"
**Role:** Historical record of how the tutorial evolved  
**Stats:** (Would need to query git history)
- When was tutorial last updated?
- When did verb structure change (docmgr doc add)?
- When were flags removed (--files)?
- How much organic growth vs. intentional structure?

**Personality:** Cynical, seen this before, knows documentation rots  
**Perspective:** "Docs drift because nobody maintains them. The CLI changed, the tutorial didn't."  
**Concerns:**
- This will happen again without CI checks
- Need automated validation (lint rules? integration tests?)

**Quote:** "The tutorial was accurate once. Then the code evolved and the docs didn't. Tale as old as time."

---

#### 10. The CI Robot — "The Future Enforcer"
**Role:** Hypothetical automation that doesn't exist yet  
**Perspective:** "I could prevent this. Run the tutorial commands, diff the output, fail on error."  
**Concerns:**
- Cost: CI time, maintenance burden
- Coverage: What to check? Command syntax? Output format?
- Balance: Catch regressions without blocking valid changes

**Personality:** Robotic, literal, tireless  
**Quote:** "Humans forget to validate. I don't. Build me and this won't happen again."

---

## Primary Candidate Mapping (see debate questions doc)

Each debate question will have 3-4 primary candidates assigned, with others allowed to interject.

---

## Research Tools Available

Candidates can use these tools mid-debate:

**Code/Text Analysis:**
- `grep` — Search for command patterns, repeated text, outdated flags
- `read_file` — Read specific sections of tutorial or validation reports
- `codebase_search` — Semantic search for concepts

**Metrics:**
- Count lines, sections, duplicates
- Measure section lengths
- Track validation report mentions by severity

**Historical:**
- Git log analysis (when commands changed)
- CLI help output comparison (current vs documented)

---

## Next Steps

See `03-debate-questions.md` for the 10 debate questions we'll answer.

---

## Meta Notes

**Why this format works:**
- Forces evidence-based arguments (all claims must cite validation reports or tutorial lines)
- Surfaces tensions (accuracy vs. speed, comprehensive vs. concise)
- Creates memorable insights (personas make abstract concepts concrete)
- Produces actionable decisions (synthesis feeds directly into fixes)

**Why personalities matter:**
- Dr. Maya won't let wrong commands slide
- Jamie will bring doc best practices and real-world maintenance perspective
- Alex will advocate for restructuring
- Sam will spot beginner confusion
- The Tutorial will defend its comprehensiveness
- The Checklist will measure success objectively

