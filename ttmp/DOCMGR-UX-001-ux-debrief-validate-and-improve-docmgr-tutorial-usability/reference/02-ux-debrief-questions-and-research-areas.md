---
Title: UX Debrief Questions and Research Areas
Ticket: DOCMGR-UX-001
Status: active
Topics:
    - ux
    - documentation
    - usability
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "10 research questions for UX debrief, mapped to participants and research methods"
LastUpdated: 2025-11-06T13:31:07.133044107-05:00
---

# UX Debrief Questions and Research Areas

## Purpose

This document lists all research questions for the UX debrief sessions, maps primary participants to each question, and outlines research methods each participant should use before the discussion.

## Question Flow Overview

The questions build from first impressions → specific workflows → advanced usage → overall assessment:

1. **First Contact** (Questions 1-2) — Can you even get started?
2. **Core Workflows** (Questions 3-5) — Do basic tasks make sense?
3. **Discovery & Learning** (Questions 6-7) — Can you figure out features?
4. **Power User Experience** (Questions 8-9) — Does it scale?
5. **Meta Assessment** (Question 10) — What's the verdict?

## Questions and Participant Assignments

### Question 1: First Impressions — Can You Get Started?

**Question:** When you first open `docmgr-how-to-use.md`, can you figure out what to do? What are your first 5 minutes like?

**Primary Participants:**
- Jordan "The New Hire" Kim (lead)
- Alex "The Pragmatist" Chen
- `docmgr-how-to-use.md` ("The Tutorial")

**Research Methods:**
- Jordan: Read sections 1-3, try to create first ticket, document confusion points
- Alex: Compare to getting started with other tools, time to first success
- Tutorial: Identify structure issues, missing signposts, jargon

**Key Focus Areas:**
- Prerequisites clarity
- Command discoverability
- Conceptual model (what is a "ticket workspace"?)
- First example quality

---

### Question 2: Installation & Setup — How Smooth Is `docmgr init`?

**Question:** Is the setup process (prerequisites, `docmgr init`, understanding the directory structure) clear and painless?

**Primary Participants:**
- Jordan "The New Hire" Kim
- Alex "The Pragmatist" Chen
- `cmd/` ("The CLI")

**Research Methods:**
- Jordan: Follow section 3 literally, document what's unclear
- Alex: Test in fresh repo, check for assumptions (Git required? Go version?)
- CLI: Review `init` command help text, error messages

**Key Focus Areas:**
- What gets created and why
- Vocabulary.yaml purpose
- Template/guidelines purpose
- Error handling when prerequisites missing

---

### Question 3: Core Workflow — Creating & Adding Documents

**Question:** The bread-and-butter workflow (create ticket, add docs, update metadata) — is it intuitive? Too many steps? Right abstractions?

**Primary Participants:**
- Jordan "The New Hire" Kim
- Morgan "The Docs-First" Taylor (lead)
- `cmd/` ("The CLI")

**Research Methods:**
- Jordan: Follow sections 4-5, count steps, note friction
- Morgan: Create 5 test tickets with different doc types, assess structure
- CLI: Compare command UX (create-ticket vs add vs meta update)

**Key Focus Areas:**
- Step count vs value
- Flag naming consistency
- Examples sufficiency
- When to use CLI vs manual editing

---

### Question 4: Metadata Management — `docmgr meta update` vs Manual Editing

**Question:** Section 5 shows `docmgr meta update` for frontmatter. Is this better than just opening the file? When would you use which?

**Primary Participants:**
- Alex "The Pragmatist" Chen (lead)
- Sam "The Power User" Rodriguez
- Morgan "The Docs-First" Taylor

**Research Methods:**
- Alex: Time both methods, assess keystroke overhead
- Sam: Test batch operations, check for scriptability
- Morgan: Try complex metadata updates, check for validation

**Key Focus Areas:**
- When CLI adds value
- Bulk update support
- Error handling (typos, invalid values)
- Discoverability of field names

---

### Question 5: Relating Files — Is This Feature Worth It?

**Question:** Section 6 is all about `docmgr relate`. Does this pull its weight? Is `--suggest` magical or confusing?

**Primary Participants:**
- Morgan "The Docs-First" Taylor (lead)
- Sam "The Power User" Rodriguez
- Alex "The Pragmatist" Chen

**Research Methods:**
- Morgan: Test relate with 10+ files, check suggestion quality
- Sam: Try `--suggest --apply-suggestions`, assess trust level
- Alex: Compare to manually editing RelatedFiles in YAML

**Key Focus Areas:**
- Suggestion accuracy
- Note-taking UX
- When to use vs manual edit
- Value proposition clarity

---

### Question 6: Search & Discovery — Can You Find Things?

**Question:** Section 7 shows search. If you're 3 weeks into a project with 20 tickets, can you actually find what you need?

**Primary Participants:**
- Morgan "The Docs-First" Taylor (lead)
- Sam "The Power User" Rodriguez
- Jordan "The New Hire" Kim

**Research Methods:**
- Morgan: Create 20-ticket test corpus, run searches, assess relevance
- Sam: Test metadata filters, reverse lookups, date filters
- Jordan: Try natural language queries, assess whether you'd use this

**Key Focus Areas:**
- Search result quality
- Filter combinations
- Output format clarity
- Performance with real corpus

---

### Question 7: Learning Curve — Can You Discover Features?

**Question:** The tutorial is 432 lines. Do you read it all? Skim? Search? How do you learn about features like `--with-glaze-output` or `.docmgrignore`?

**Primary Participants:**
- Jordan "The New Hire" Kim
- Sam "The Power User" Rodriguez (lead)
- `docmgr-how-to-use.md` ("The Tutorial")

**Research Methods:**
- Jordan: Skim tutorial, note what you skip, try `docmgr --help`
- Sam: Look for advanced features, check if "heads-up" boxes help
- Tutorial: Assess structure (progressive disclosure? Reference vs tutorial?)

**Key Focus Areas:**
- Skimmability
- Help text vs tutorial coverage
- Advanced feature discoverability
- Information architecture

---

### Question 8: Power User Experience — Does It Scale?

**Question:** Section 12 shows Glaze scripting. Is this actually usable for automation? What about performance at scale?

**Primary Participants:**
- Sam "The Power User" Rodriguez (lead)
- Alex "The Pragmatist" Chen
- `cmd/` ("The CLI")

**Research Methods:**
- Sam: Write real scripts (e.g., CI check), test Glaze flags
- Alex: Test with large corpus (100+ docs), check speed
- CLI: Review output format consistency, flag naming

**Key Focus Areas:**
- Glaze documentation quality
- Field name stability
- Performance characteristics
- Scripting ergonomics

---

### Question 9: Validation & Maintenance — Does `docmgr doctor` Help?

**Question:** Sections 9, 11 cover validation. Does this actually prevent problems or just nag?

**Primary Participants:**
- Morgan "The Docs-First" Taylor (lead)
- Alex "The Pragmatist" Chen
- `cmd/` ("The CLI")

**Research Methods:**
- Morgan: Test doctor with messy corpus, assess signal vs noise
- Alex: Check if warnings would actually run in CI (false positive rate)
- CLI: Review warning categories, suppression mechanisms

**Key Focus Areas:**
- Warning relevance
- `.docmgrignore` usability
- Actionability of errors
- CI integration story

---

### Question 10: Overall Assessment — Would You Use This?

**Question:** After trying everything, would you adopt docmgr for your team? Why or why not? What are the top 3 blockers and top 3 wins?

**Primary Participants:**
- All participants
- Erin "The Facilitator" Garcia (synthesizes)

**Research Methods:**
- Everyone: Rank pain points and delights
- Erin: Facilitate prioritization, identify quick wins
- CLI & Tutorial: Respond to top criticisms

**Key Focus Areas:**
- Adoption barriers
- Killer features
- Missing features
- Value proposition clarity

## Research Preparation Checklist

Before each discussion round, participants should:

**For Developer Personas:**
- [ ] Install docmgr and run `docmgr --help`
- [ ] Read the tutorial section(s) relevant to the question
- [ ] Try the workflow hands-on (create test ticket/docs)
- [ ] Document at least 2 specific pain points or wins
- [ ] Prepare concrete examples (command output, line numbers)

**For Code/System Personas:**
- [ ] Review own structure (Tutorial: TOC, examples; CLI: help text, flags)
- [ ] Identify gaps or inconsistencies
- [ ] Prepare to show actual code/docs when challenged
- [ ] Be ready to defend design decisions with data

**For Facilitator:**
- [ ] Review all participants' pre-session findings
- [ ] Identify common themes and conflicts
- [ ] Prepare follow-up questions
- [ ] Draft synthesis framework for the question

## Discussion Round Template

Each question gets a separate discussion round document:

```
## Question N: [Title]

## Pre-Session Research

### Jordan "The New Hire" Kim
[Commands tried, confusion points, findings]

### Alex "The Pragmatist" Chen
[Timing data, comparisons, analysis]

[... all participants ...]

## Opening Reactions (2 min each)

### Jordan
[Gut reaction]

### Alex
[Gut reaction]

[... all participants ...]

## Deep Dive Discussion (5-10 min)

[Passionate exchange with data, interruptions encouraged]

## Live Experiments

[Real-time testing during discussion]

## Facilitator Synthesis

### Key Themes
- [Theme 1]
- [Theme 2]

### Pain Points Identified
1. [Pain point with severity]
2. [Pain point with severity]

### Wins Celebrated
1. [What works well]

### Proposed Improvements
1. [Specific change with before/after]
2. [Specific change with before/after]

### Action Items
- [ ] [Concrete task]
- [ ] [Concrete task]
```

## Success Criteria

A good discussion round should produce:

1. **At least 3 concrete pain points** with specific examples (line numbers, command output)
2. **At least 2 proposed improvements** with before/after
3. **Evidence-based decisions** (not just opinions)
4. **Prioritized action items** (not everything is P0)
5. **Cross-participant consensus or documented disagreement**

## Related

- [UX Debrief Participants and Format](./01-ux-debrief-participants-and-format.md)
- Tutorial under review: `pkg/doc/docmgr-how-to-use.md`
- CLI source: `cmd/` and `pkg/commands/`
