---
Title: UX Debrief - Validate and Improve docmgr Tutorial Usability
Ticket: DOCMGR-UX-001
Status: complete
Topics:
    - ux
    - documentation
    - usability
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/doc/docmgr-how-to-setup.md
      Note: Original setup tutorial (276 lines) - basis for v2
    - Path: pkg/doc/docmgr-how-to-use.md
      Note: Tutorial under review (432 lines)
    - Path: playbooks/01-tutorial-v2-restructured-based-on-ux-findings.md
      Note: Daily usage tutorial (1064 lines)
    - Path: playbooks/02-setup-tutorial-v2-restructured.md
      Note: Repository setup tutorial (858 lines)
    - Path: playbooks/03-ci-and-automation-guide.md
      Note: CI/CD and automation guide (789 lines)
ExternalSources: []
Summary: UX debrief for docmgr tutorial using heated brainstorm format; identifies P0 init ordering issues and P1 jargon/flow problems
LastUpdated: 2025-11-06T14:59:16.634500808-05:00
---







# UX Debrief - Validate and Improve docmgr Tutorial Usability

## Overview

This ticket uses a **heated UX user survey debrief/brainstorm format** (inspired by the debate framework) to validate and improve the `docmgr-how-to-use.md` tutorial. Instead of a presidential debate, this is a collaborative but passionate exploration where 7 participants (4 developer personas, 2 personified code entities, 1 facilitator) try out the tool, research the code/docs, and discuss findings with data.

**Goal:** Identify specific pain points and wins in the tutorial, propose concrete improvements with before/after examples.

**Completed:**
- ✅ **Round 1: First Impressions** (552 lines) — Init ordering issue, jargon problems
- ✅ **Round 2: Installation & Setup** (654 lines) — Empty vocabulary confusion, Git prerequisite error
- ✅ **Round 3: Core Workflow** (785 lines) — --ticket flag repetition (4-15× per session)
- ✅ **Round 4: Metadata Management** (616 lines) — CLI for bulk, manual for single-doc edits
- ✅ **Round 5: Relating Files** (681 lines) — Notes are GOLD, --suggest needs docs
- ✅ **Round 6: Search & Discovery** (712 lines) — Fast/accurate but output format needs work
- ✅ **Round 7: Learning Curve** (738 lines) — Tutorial too dense, needs [BASIC]/[ADVANCED] markers
- ✅ **Rounds 8-10 Summary** (358 lines) — Power user verdict, validation effectiveness, final assessment
- ✅ **Framework documents** (2 reference docs with 7 personas, 10 questions, methodology)

**Key Findings Summary:**

**Round 1 (First Impressions):**
- **P0:** Tutorial shows `create-ticket` BEFORE explaining `docmgr init` is required
- **P0:** Error message when init not run doesn't guide user to solution  
- **P1:** Section 3 explains concepts before showing output (should be "show then explain")
- **P1:** Jargon ("frontmatter", "docs root") used without definition
- **Win:** `docmgr --help` → `docmgr help how-to-use` breadcrumb is excellent

**Round 2 (Installation & Setup):**
- **P0:** Empty vocabulary causes confusion (users don't know if it's required or optional)
- **P0:** Tutorial doesn't explain what `docmgr init` creates (just says "run this")
- **P1:** `--seed-vocabulary` flag hidden (not in tutorial or examples)
- **P1:** Git incorrectly listed as prerequisite (not actually required)
- **P2:** CLI output is just a table (no "next steps" guidance)
- **Win:** Init command is mechanically solid (fast, idempotent, safe)

**Round 3 (Core Workflow):**
- **P0:** `--ticket` flag required on every `add` command (typed 4-15× per session at scale)
- **P0:** `meta update` requires full paths (59+ characters common)
- **P1:** No CWD-based ticket inference (can't infer from directory context)
- **P1:** Bulk operation patterns not documented in tutorial
- **P2:** Smart defaults (numeric prefixes, topic inheritance) not highlighted
- **Win:** Core abstraction (tickets → docs → metadata) is sound
- **Win:** Template auto-filling, unknown doc-type handling work perfectly

**Rounds 4-7 (Condensed Findings):**
- **Round 4 (Meta Update):** CLI verbose for single-field edits; better for bulk operations
- **Round 5 (Relate Files):** Notes are valuable; --suggest needs documentation; feature works well
- **Round 6 (Search):** Works fast, reverse lookup powerful; output format needs improvement
- **Round 7 (Learning Curve):** Tutorial too dense (432 lines); advanced features buried; needs structure

**Top 5 P0/P1 Issues Identified:**
1. **CWD-based ticket inference** — 40% typing reduction for common workflows
2. **Tutorial restructuring** — Add [BASIC]/[ADVANCED] markers, consider splitting
3. **Init vocabulary seeding** — Interactive prompt or default seeding
4. **Search output formatting** — Better spacing, indentation, clarity
5. **Jargon + "show then explain"** — Fix accessibility for junior developers

**Cross-Cutting Wins:**
- Core abstraction (tickets → docs) is sound
- Template auto-filling, smart defaults work perfectly
- Glaze scripting beloved by power users
- Help text quality generally good

**Rounds 8-10 (Final Assessment):**
- **Round 8 (Power Users):** Glaze scripting is phenomenal; stable API, CI-friendly, fast at scale
- **Round 9 (Validation):** Doctor catches real issues; staleness warnings need tuning (too aggressive)
- **Round 10 (Overall Verdict):** **STRONG YES from all participants** with P0 fixes

**Final Consensus:** Tool is SOLID (core abstraction right, features work, performance good). Main issues are documentation structure and CLI verbosity. All fixable.

**Status:** ✅ ALL 7 DETAILED ROUNDS COMPLETE + 2 SUMMARY DOCS — UX Debrief Finished

**Deliverables:** 
- 14 documents total
- 2 reference docs (framework, participants, questions)
- 7 full debate rounds (~600-800 lines each with passionate discussion)
- 2 summary docs (rounds 4-7 and 8-10 condensed findings)
- Total documentation: ~5,100 lines of UX analysis

**Final Deliverable:**
✅ **UX Findings Report** (design doc) — Comprehensive synthesis with:
- 15 ranked issues (5 P0, 5 P1, 5 P2)
- 4-phase implementation plan
- Decision matrices and alternatives considered
- Metrics for success

**Next Steps:**
1. ✅ Design Doc complete — All findings synthesized
2. Optionally: Create RFC with sprint-level implementation details
3. Optionally: Draft Tutorial v2 based on feedback
4. Implement P0 fixes (Phase 1: Tutorial restructure — 2-4 hours)

**Top 5 P0 Fixes (Must Do):**
1. Restructure tutorial — Split Quick Start (50 lines) + Guide (200) + Reference
2. Add CWD-based --ticket inference (40% typing reduction)
3. Init vocabulary seeding — Interactive prompt or default seed
4. Define jargon on first use (frontmatter, docs root, etc.)
5. Add value proposition section ("Why use this vs mkdir?")

**Strategic Insight:** docmgr is a FORCE MULTIPLIER for teams. Sam wrote 4 automation scripts in 1 hour. Morgan sees it as chaos prevention at scale. Fix the onboarding, and adoption will be strong.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- ux
- documentation
- usability

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
