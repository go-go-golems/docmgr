---
Title: UX Findings Report - Executive Summary
Ticket: DOCMGR-UX-001
Status: active
Topics:
    - ux
    - documentation
    - usability
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - path: pkg/doc/docmgr-how-to-use.md
      note: Tutorial under review (432 lines)
    - path: various/01-round-1-first-impressions.md
      note: Round 1 findings (552 lines)
    - path: various/02-round-2-installation-and-setup-ux.md
      note: Round 2 findings (654 lines)
    - path: various/03-round-3-core-workflow-creating-and-adding-docs.md
      note: Round 3 findings (785 lines)
    - path: various/04-round-4-metadata-management-meta-update-vs-manual.md
      note: Round 4 findings (616 lines)
    - path: various/05-round-5-relating-files-feature-value.md
      note: Round 5 findings (681 lines)
    - path: various/06-round-6-search-and-discovery-effectiveness.md
      note: Round 6 findings (712 lines)
    - path: various/07-round-7-learning-curve-and-feature-discovery.md
      note: Round 7 findings (738 lines)
ExternalSources: []
Summary: "Comprehensive UX findings from 10-round heated debrief: docmgr is solid but needs tutorial restructuring and CLI ergonomics fixes"
LastUpdated: 2025-11-06T14:37:15.873575651-05:00
---

# UX Findings Report — Executive Summary

## Executive Summary

After 10 rounds of heated UX debrief with 7 participants (4 developer personas, 2 personified code entities, 1 facilitator), we conducted comprehensive hands-on testing of docmgr's tutorial and CLI. **The verdict: docmgr is fundamentally SOLID with excellent core abstractions, but suffers from tutorial structure issues and CLI verbosity that block adoption.**

**Key Findings:**
- ✅ **Core product is sound** — Tickets → Docs → Metadata hierarchy makes sense
- ✅ **Power features are excellent** — Glaze scripting, validation, relationships all work well
- ⚠️ **Tutorial structure blocks discovery** — 432 lines, no progressive disclosure, power features buried
- ⚠️ **CLI verbosity causes friction** — --ticket typed 4-15× per session, 59+ char paths common

**Recommendation:** Fix the Top 5 P0 issues and docmgr goes from "works well" to "delightful to use."

---

## Problem Statement

The `docmgr-how-to-use.md` tutorial (432 lines) is the primary onboarding document, but participants identified significant usability issues that block adoption:

1. **New users hit blockers in first 5 minutes** (init ordering, jargon)
2. **Power users almost miss critical features** (Glaze at line 280)
3. **Repetitive CLI usage creates friction** (--ticket flag typed 15× per session)
4. **No clear guidance** on when to use CLI vs manual editing, when to use features

These issues are **fixable** — they're about documentation structure and CLI ergonomics, not fundamental design flaws.

---

## What We Tested

### Methodology

**7 Participants:**
- Jordan "The New Hire" Kim (junior developer, first-time CLI doc tool user)
- Alex "The Pragmatist" Chen (senior engineer, efficiency-focused)
- Sam "The Power User" Rodriguez (tech lead, automation-focused)
- Morgan "The Docs-First" Taylor (staff engineer, documentation advocate)
- `docmgr-how-to-use.md` (the Tutorial itself, personified)
- `cmd/` (the CLI itself, personified)
- Erin "The Facilitator" Garcia (UX researcher, synthesizer)

**10 Research Questions:**
1. First Impressions — Can you get started?
2. Installation & Setup — Is init clear?
3. Core Workflow — Creating & adding docs intuitive?
4. Metadata Management — CLI vs manual editing?
5. Relating Files — Is this feature worth it?
6. Search & Discovery — Can you find things?
7. Learning Curve — Can you discover features?
8. Power User Experience — Does it scale?
9. Validation — Does doctor help or nag?
10. Overall Assessment — Would you use this?

**Testing Approach:**
- Hands-on commands with real docmgr installation
- Timing measurements (keystroke counts, execution time)
- Scale testing (5-20 tickets, 60-80 docs)
- Reading pattern analysis (what users actually read vs skip)

---

## Key Findings by Round

### Round 1: First Impressions ⚠️ P0 Blockers Found

**Pain Points:**
- Tutorial shows `create-ticket` command BEFORE explaining `docmgr init` is required
- Error message "no .ttmp.yaml found" doesn't suggest running init
- Jargon ("frontmatter", "docs root") used without definition
- Section 3 explains concepts before showing output (breaks flow)

**Wins:**
- `docmgr --help` → `docmgr help how-to-use` breadcrumb is excellent
- Tutorial accessible in terminal (no context switching)
- Copy/pasteable examples

**Impact:** New users hit failure in first 3 minutes (Jordan's experience)

---

### Round 2: Installation & Setup ⚠️ Vocabulary Confusion

**Pain Points:**
- Init creates empty vocabulary (users confused if it's required/optional)
- `--seed-vocabulary` flag hidden (not in tutorial or examples)
- Tutorial doesn't explain WHAT init creates (just says "run this")
- Git incorrectly listed as prerequisite (works fine without Git)

**Wins:**
- Init is idempotent (safe to run multiple times)
- Help text clear about what gets created
- Seeded vocabulary provides good defaults (when discovered)

**Impact:** Users stare at empty vocabulary.yaml wondering what to do

---

### Round 3: Core Workflow ⚠️ Repetitive Typing

**Pain Points:**
- `--ticket` flag required on every command (typed 4-15× per session)
- `meta update` requires full paths (59+ characters common)
- No CWD-based ticket inference (can't infer from directory)
- Bulk operation patterns not documented

**Wins:**
- Core abstraction (tickets → docs) is sound
- Topic/owner inheritance (set once on ticket, flows to docs)
- Automatic numeric prefixes (01-, 02-) keep files ordered
- Unknown doc-types handled gracefully (go to various/)

**Impact:** Morgan typed --ticket 15× for 5 tickets, calls it "brutal at scale"

---

### Round 4: Metadata Management ✅ Clarity Needed

**Pain Points:**
- Tutorial doesn't clearly emphasize using CLI for frontmatter updates
- Path verbosity for single-doc updates (97 chars)
- Users uncertain when to use commands vs editing files

**Wins:**
- CLI excels at bulk operations (one command updates 10 docs)
- Validation ensures proper YAML syntax
- Automation-friendly for scripts and CI

**Impact:** Users need clearer guidance on command patterns

**Updated Philosophy:**
- **Frontmatter updates** → Use `docmgr meta update` commands
- **Body content** → Write/edit in your editor
- **Bulk operations** → CLI shines (1 command, many docs)
- **Automation/scripts** → CLI only approach

---

### Round 5: Relating Files ✅ Notes Are GOLD

**Pain Points:**
- `--suggest` heuristics unexplained (how does it work?)
- No workflow guidance (when to relate in dev process?)
- No best practices for notes (what makes a good note?)

**Wins:**
- Notes transform file lists into navigation maps (Morgan: "GOLD")
- Reverse lookup powerful (`--file` search finds docs from code)
- ROI positive (saves 2 min per code review)
- Automation-friendly (can script relate operations)

**Impact:** Feature is valuable but underdocumented

**Key Quote (Morgan):** "Notes turn file lists into navigation maps. Instead of just listing 10 files, I can explain why each matters. This is 10× better than plain links."

---

### Round 6: Search & Discovery ✅ Works But Dense

**Pain Points:**
- Output format dense (everything on one line, hard to scan)
- No visual hierarchy in results
- Case sensitivity behavior unclear

**Wins:**
- Fast (sub-second even with 200+ docs)
- Accurate results
- Filters combine well (query + topics + doc-type)
- Reverse lookup (file/dir search) is powerful
- Structured output perfect for scripting

**Impact:** Search works but output needs formatting improvement

---

### Round 7: Learning Curve ⚠️ P0 Structure Issue

**Pain Points:**
- 432 lines intimidating (no clear entry points)
- Advanced features buried (Glaze at line 280)
- No progressive disclosure or reader-type guidance
- No [BASIC]/[ADVANCED] markers

**Wins:**
- Tutorial is comprehensive (covers all features)
- Help text quality good
- Examples practical

**Impact:** Users read 35% (Jordan) to 10% (Sam) of tutorial, miss features

**Reading Patterns Discovered:**
- Jordan: Read sections 1-5 linearly, then search for specifics
- Sam: Keyword-scan for "json", "automation", jump to Section 12
- Alex: Use `--help` more than tutorial

---

### Rounds 8-10: Power User Verdict ✅ Strong YES

**Round 8 (Power Users):**
- Glaze scripting is phenomenal (Sam wrote 4 scripts in 1 hour)
- Stable API, CI-friendly, fast at scale
- Should be front-loaded (line 50, not line 280)

**Round 9 (Validation):**
- Doctor catches real issues (broken paths, typos)
- Staleness warnings too aggressive (78% false positive rate)
- `.docmgrignore` works well

**Round 10 (Overall Verdict):**
- **All participants say YES to adoption** (with P0 fixes)
- Strategic value: chaos prevention at team scale
- Tactical value: automation force multiplier

---

## Top Issues Ranked by Severity

### P0 — Must Fix (Blocks Adoption)

| # | Issue | Impact | Found In |
|---|-------|--------|----------|
| 1 | Tutorial structure (no clear entry points) | Power users miss features, beginners overwhelmed | Round 7 |
| 2 | CWD-based --ticket inference missing | 40% typing reduction lost | Round 3 |
| 3 | Init vocabulary seeding (empty by default) | Users confused about requirement | Round 2 |
| 4 | Init ordering in tutorial | Users fail in first 3 minutes | Round 1 |
| 5 | Jargon undefined | Blocks junior developers | Round 1 |

### P1 — Should Fix (Degrades Experience)

| # | Issue | Impact | Found In |
|---|-------|--------|----------|
| 6 | Search output format dense | Hard to scan visually | Round 6 |
| 7 | --suggest undocumented | Can't use feature | Round 5 |
| 8 | Bulk operation patterns missing | Users reinvent patterns | Round 3, 4 |
| 9 | Meta update context unclear | Users unsure when to use CLI | Round 4 |
| 10 | Staleness defaults aggressive | 78% false positive rate | Round 9 |

### P2 — Nice to Have (Polish)

| # | Issue | Impact | Found In |
|---|-------|--------|----------|
| 11 | --file flag for relative paths | Less verbose meta update | Round 3 |
| 12 | --json/--csv shortcuts | Shorter than --with-glaze-output | Round 8 |
| 13 | Fuzzy search | Typo tolerance | Round 6 |
| 14 | Link validation in doctor | Catch broken internal links | Round 9 |
| 15 | Evergreen doc marking | Skip stale checks | Round 9 |

---

## What Works Excellently

### Core Abstractions ✅

**Unanimous consensus:** Tickets → Docs → Metadata hierarchy is RIGHT.

- Makes sense conceptually
- Scales from 1 ticket to 100+
- Supports team workflows
- No one questioned the fundamental design

### Smart Defaults ✅

Features that "just work":
- **Topic inheritance** — Set on ticket, flows to all docs
- **Numeric prefixes** — Automatic (01-, 02-), keeps files ordered
- **Unknown doc-types** — Go to various/ but preserve DocType field
- **Template auto-filling** — Placeholders substituted automatically

### Power Features ✅

**Glaze scripting** (Round 8):
- Sam wrote 4 automation scripts in 1 hour
- Stable API, CI-friendly, fast at scale
- JSON/CSV/TSV all work perfectly

**Validation (doctor)** (Round 9):
- Catches broken RelatedFiles paths
- Detects typos in topics/doc-types
- Exit codes work for CI
- `.docmgrignore` suppresses noise

**Relationships (relate)** (Round 5):
- Notes are "GOLD" (Morgan's quote)
- Reverse lookup powerful for code reviews
- Saves 2 min per file review

---

## Proposed Solutions

### Solution 1: Tutorial Restructuring (P0)

**Problem:** 432 lines, no clear entry points, advanced features buried.

**Solution:** Restructure into 3 parts with navigation:

```markdown
# Tutorial — Using docmgr

## Quick Navigation

┌──────────────────────────────────────────────────────────────┐
│ 📚 New user? → Part 1: Essentials (10 min)                  │
│ ⚡ Automation? → Part 3: Power Features (jump to line 200)  │
│ 🔍 Specific command? → docmgr COMMAND --help                │
└──────────────────────────────────────────────────────────────┘

# Part 1: Essentials 📚 [10 minute read - START HERE]
1. Overview & Prerequisites
2. Initialize Repository  
3. Create Tickets & Add Docs
4. Basic Search
5. Metadata Basics

✅ **Milestone:** You can now use docmgr! Continue to Part 2 or jump to Part 3.

# Part 2: Everyday Workflows 🔧 [Read as needed]
6. Relating Files
7. Changelogs & Tasks
8. Validation (doctor)

# Part 3: Power User Features ⚡ [For automation]
9. Structured Output (Glaze)
10. CI Integration
11. Bulk Operations & Scripting
```

**Impact:** Clear paths for beginners, practitioners, and power users

---

### Solution 2: CWD-Based Ticket Inference (P0)

**Problem:** --ticket typed 4-15× per session (brutal at scale).

**Solution:** Infer ticket from current working directory:

```bash
# Explicit (always works)
$ docmgr add --ticket PROJ-001 --doc-type design-doc --title "X"

# Inferred (when in ticket directory)
$ cd ttmp/PROJ-001-feature/
$ docmgr add --doc-type design-doc --title "X"
Using ticket: PROJ-001 (inferred from CWD)
```

**Implementation:** Parse CWD for `ttmp/TICKET-slug/` pattern, extract TICKET ID.

**Impact:** 40% typing reduction (Morgan's measurement)

---

### Solution 3: Init Vocabulary Seeding (P0)

**Problem:** Empty vocabulary causes confusion ("Is this required?").

**Solution:** Interactive prompt or seed by default:

```bash
$ docmgr init

Initialize docs root at /tmp/test/ttmp? [Y/n] 

Seed vocabulary with defaults (chat, backend, websocket topics)? [Y/n] 

✓ Initialized /tmp/test/ttmp
✓ Seeded vocabulary.yaml

Next steps:
  • Create ticket: docmgr create-ticket --ticket XXX-123 --title "..."
  • List topics: docmgr vocab list
```

**Impact:** Eliminates empty vocabulary confusion for 90% of users

---

### Solution 4: Prerequisites Clarity (P0)

**Problem:** Tutorial says run create-ticket, then mentions init as footnote. Users fail first try.

**Solution:** Elevate init to Prerequisites:

```markdown
## 2. Prerequisites

- docmgr available on PATH
- A directory to work in (Git recommended but not required)

### First-Time Setup (REQUIRED)

Run `docmgr init` to create the documentation workspace:

```bash
docmgr init --seed-vocabulary
```

This creates:
- ttmp/ — Docs root
- vocabulary.yaml — Topics/docTypes (seeded with defaults)
- _templates/ — Used by 'docmgr add'
- _guidelines/ — See with 'docmgr guidelines'

**You only run this once per repository.**
```

**Impact:** Prevents first-use failure

---

### Solution 5: Define Jargon (P0)

**Problem:** "Frontmatter", "docs root", "ticket workspace" undefined.

**Solution:** Add glossary or first-use definitions:

```markdown
## 1.5 Key Concepts

- **Ticket** — An identifier (like JIRA/GitHub issue) for a unit of work
- **Ticket workspace** — A directory containing all docs for a ticket
- **Docs root** — The ttmp/ directory containing all tickets
- **Frontmatter** — YAML metadata at the top of markdown files
- **RelatedFiles** — Code files referenced in a doc's frontmatter
```

**Impact:** Accessibility for junior developers

---

## Implementation Plan

### Phase 1: Tutorial Fixes (Immediate) — 2-4 hours

**P0 Fixes (Tutorial Only):**
1. ✅ Add navigation box at top with reader paths
2. ✅ Restructure into 3 parts with [BASIC]/[ADVANCED] markers
3. ✅ Move init to Prerequisites with clear explanation
4. ✅ Add glossary or first-use jargon definitions
5. ✅ Add "when to use CLI vs manual" section

**Deliverable:** Updated tutorial (same length, better structure)

---

### Phase 2: CLI Ergonomics (High Value) — 1-2 days

**P0-P1 Fixes (CLI Code):**
1. ✅ Implement CWD-based ticket inference
2. ✅ Add --file flag for relative paths in meta update
3. ✅ Interactive prompt for init --seed-vocabulary
4. ✅ Improve error message when init not run
5. ✅ Add --json/--csv shortcuts for --with-glaze-output

**Deliverable:** ergonomics release (v0.x+1)

---

### Phase 3: Documentation Polish (Medium Priority) — 4-6 hours

**P1-P2 Fixes (Tutorial + Help):**
1. ✅ Document --suggest heuristics clearly
2. ✅ Add "Relate Workflow" section
3. ✅ Add automation patterns throughout
4. ✅ Improve search output format
5. ✅ Add bulk operation examples

**Deliverable:** Comprehensive tutorial with all usecases

---

### Phase 4: Advanced Features (Low Priority) — Future

**P2 Fixes (CLI Code):**
1. ⏳ Fuzzy search
2. ⏳ Link validation in doctor
3. ⏳ Evergreen doc marking
4. ⏳ Ranking indicators in search

**Deliverable:** Polish release

---

## Metrics for Success

**Measure tutorial improvements:**
- [ ] New user time-to-first-ticket < 5 minutes (currently 8 min)
- [ ] Power user discovers Glaze within 5 minutes (currently 20+ min)
- [ ] Users read 50%+ of relevant sections (currently 35%)

**Measure CLI improvements:**
- [ ] --ticket typing count reduced by 40% with CWD inference
- [ ] meta update path length reduced from 59 to <25 chars
- [ ] Init vocabulary confusion drops to <10% of users

**Measure adoption:**
- [ ] 3 new teams try docmgr in next quarter
- [ ] Positive feedback from 2+ teams
- [ ] CI integration examples in production use

---

## Alternatives Considered

### Alternative 1: Split Tutorial into Multiple Docs

**Option:** Create quick-start.md (50 lines) + tutorial.md (300 lines) + automation-guide.md

**Rejected because:**
- Adds discoverability burden (which doc to read?)
- Users prefer one searchable document
- Can achieve same goal with clear sections

**Chosen instead:** Restructure one doc with clear parts and navigation

---

### Alternative 2: Make Vocabulary Required

**Option:** Enforce topics must be in vocabulary (fail if unknown)

**Rejected because:**
- Too restrictive for exploratory work
- Unknown doc-types to various/ is good flexibility
- Warnings (current approach) better than errors

**Chosen instead:** Keep flexible, document that it's for validation not enforcement

---

### Alternative 3: Ticket Context File (.docmgr-context)

**Option:** Store current ticket in .docmgr-context file to avoid --ticket flag

**Rejected because:**
- Another hidden file to track
- CWD inference simpler and more intuitive
- Context file requires manual management

**Chosen instead:** CWD inference with explicit --ticket override

---

## Risks and Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Tutorial restructuring breaks existing user workflows | Low | Users adapt easily, structure improves |
| CWD inference guesses wrong ticket | Medium | Print "Using ticket: X (inferred)", allow explicit override |
| Interactive prompts annoy CI/scripts | Medium | Detect TTY, skip prompts in non-interactive mode |
| Empty vocabulary still confuses users | Low | Seed by default in prompt, document clearly |

---

## Open Questions

**For immediate implementation:**
- ✅ Should --seed-vocabulary be default or prompted? → **Prompted (consensus)**
- ✅ Should tutorial be split or restructured? → **Restructured (consensus)**
- ✅ Should init be in Prerequisites or Section 3? → **Prerequisites (consensus)**

**For future consideration:**
- How to handle CWD inference in subdirectories? (design/, reference/)
- Should we add --json shortcut or keep --with-glaze-output? (Both?)
- Should doctor stale check be opt-in or just higher default? (Needs testing)

---

## References

### Debate Rounds (Full Detail)

- [Round 1: First Impressions](../various/01-round-1-first-impressions.md) (552 lines)
- [Round 2: Installation & Setup](../various/02-round-2-installation-and-setup-ux.md) (654 lines)
- [Round 3: Core Workflow](../various/03-round-3-core-workflow-creating-and-adding-docs.md) (785 lines)
- [Round 4: Metadata Management](../various/04-round-4-metadata-management-meta-update-vs-manual.md) (616 lines)
- [Round 5: Relating Files](../various/05-round-5-relating-files-feature-value.md) (681 lines)
- [Round 6: Search & Discovery](../various/06-round-6-search-and-discovery-effectiveness.md) (712 lines)
- [Round 7: Learning Curve](../various/07-round-7-learning-curve-and-feature-discovery.md) (738 lines)

### Summary Documents

- [Rounds 8-10 Final Assessment](../various/12-rounds-8-10-final-assessment.md) (358 lines)
- [Framework: Participants and Format](../reference/01-ux-debrief-participants-and-format.md)
- [Framework: Questions and Research Areas](../reference/02-ux-debrief-questions-and-research-areas.md)

### Source Material

- Tutorial under review: `pkg/doc/docmgr-how-to-use.md` (432 lines)
- Methodology inspiration: `go-go-mento/ttmp/REORG-FEATURE-STRUCTURE-.../playbooks/playbook-using-debate-framework-for-technical-rfcs.md`

---

## Approval

**Decision needed:** Approve Top 5 P0 fixes for implementation.

**Stakeholders:** docmgr maintainers, tutorial authors, potential adopters.

**Next step:** Create RFC with detailed implementation plan for P0 fixes.
