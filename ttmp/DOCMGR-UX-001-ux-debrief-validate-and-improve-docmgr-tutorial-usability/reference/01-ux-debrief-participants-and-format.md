---
Title: UX Debrief Participants and Format
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
Summary: "Defines participants and format for heated UX debrief/brainstorm session on docmgr tutorial usability"
LastUpdated: 2025-11-06T13:31:05.76414679-05:00
---

# UX Debrief Participants and Format

## Purpose

This document defines the participants, their perspectives, and the format for a heated UX user survey debrief / brainstorm discussion about docmgr's tutorial (`how-to-use-docmgr.md`). Unlike the presidential debate format, this is a **collaborative but passionate** exploration where participants can:

- Try out the tool themselves
- Research code and documentation
- Call out pain points loudly
- Propose improvements on the fly
- Interrupt with "Wait, I tried that and..."

**Goal:** Validate and improve how easy docmgr is to use based on the tutorial.

## Debrief Format Rules

### Session Structure

1. **Pre-Session Research** (30 min simulation)
   - Each participant gets hands-on time with docmgr
   - Can read tutorial, try commands, examine code
   - Document findings, confusion points, and wins

2. **Heated Discussion Rounds** (per question)
   - **Opening Reactions** (2 min each) — Gut reactions, first impressions
   - **Dive Deep** (5 min) — Share specific findings, pain points, delights
   - **Cross-Talk** (open) — Participants interrupt, agree, disagree vehemently
   - **Synthesis** (facilitator) — Extract themes, action items

3. **Live Experiments** (throughout)
   - "Wait, let me try that right now..."
   - Participants can demo problems in real-time
   - Code/docs examined on the spot

### Discussion Norms

**Encouraged:**
- Passionate disagreement ("That's absurd! Look at this...")
- Real-time tool testing ("Let me pull up a terminal...")
- Concrete examples ("On line 47 it says X but I expected Y")
- Developer empathy ("A new hire would be lost here")
- Quick prototypes ("What if we changed this to...")

**Discouraged:**
- Vague complaints ("This feels weird")
- Untested assumptions ("I think users would...")
- Bikeshedding minor details without data
- Personal attacks (attack ideas, not people)

## Participants (7 Personas)

### 1. Developer Personas (4 participants)

#### Jordan "The New Hire" Kim
**Role:** Junior developer, 6 months out of bootcamp  
**Background:** First time using a CLI documentation tool  
**Perspective:** "What's a frontmatter? Why so many flags?"  
**Main Concerns:**
- Cognitive load of commands
- Discoverability of features
- Error message clarity
- Examples clarity

**Tools/Methods:**
- Follow tutorial step-by-step literally
- Get confused by jargon
- Ask "why" questions constantly
- Test edge cases accidentally

**Personality:** Earnest, easily overwhelmed, asks basic questions without shame

---

#### Alex "The Pragmatist" Chen
**Role:** Senior backend engineer, 5 years experience  
**Background:** Used to Confluence, Notion, markdown in repos  
**Perspective:** "Does this save me time vs `mkdir` and `vim`?"  
**Main Concerns:**
- Time-to-value
- Overhead vs manual workflows
- Integration with existing tools
- Migration cost

**Tools/Methods:**
- Compare commands to bash equivalents
- Measure keystrokes/time
- Test automation potential
- Check if it respects `.gitignore`

**Personality:** Skeptical, efficiency-obsessed, needs convincing

---

#### Sam "The Power User" Rodriguez
**Role:** Tech lead, loves automation  
**Background:** Heavy tmux/vim/CLI user, writes shell scripts daily  
**Perspective:** "Show me the JSON output and Glaze flags"  
**Main Concerns:**
- Scriptability
- Performance at scale
- Consistency of output formats
- Advanced features visibility

**Tools/Methods:**
- Try `--with-glaze-output` immediately
- Test with 100+ docs
- Write test scripts
- Read source code when docs unclear

**Personality:** Demanding, detail-oriented, appreciates power but critiques rough edges

---

#### Morgan "The Docs-First" Taylor
**Role:** Staff engineer, documentation advocate  
**Background:** Maintains team wiki, writes RFCs  
**Perspective:** "Good docs tools shape good thinking"  
**Main Concerns:**
- Does structure enforce good practices?
- Discoverability of relationships
- Search quality
- Long-term maintainability

**Tools/Methods:**
- Test search across 20+ tickets
- Try to break relationships
- Check staleness detection
- Validate metadata consistency

**Personality:** Thoughtful, systems-thinker, cares about team workflows

---

### 2. Personified Code/System Entities (2 participants)

#### `docmgr-how-to-use.md` ("The Tutorial")
**Stats:**
- 432 lines
- 15 sections
- 50+ code examples
- 3 "heads-up" boxes

**Perspective:** "I try to be helpful but I'm DENSE"  
**What it wants:** To be read sequentially and understood  
**What it fears:** Being skimmed, misinterpreted, or too intimidating  

**Tools/Methods:**
- Can quote own sections
- Knows what's covered vs missing
- Aware of structure/flow issues
- Defensive about criticism but willing to improve

**Personality:** Earnest, slightly insecure, wants to be loved

---

#### `cmd/` ("The CLI")
**Stats:**
- 28 commands
- 100+ flags total
- Cobra-based
- Glaze output support

**Perspective:** "I'm powerful but am I approachable?"  
**What it wants:** To feel intuitive and consistent  
**What it fears:** Users falling back to manual workflows  

**Tools/Methods:**
- Can show actual command signatures
- Knows flag inconsistencies
- Aware of help text quality
- Can demo actual behavior

**Personality:** Functional but anxious about UX, wants validation

---

### 3. Wildcard Participant (1)

#### Erin "The Facilitator" Garcia
**Role:** UX researcher / product person  
**Background:** Runs user studies, reads analytics  
**Perspective:** "What does the data say? What do users *actually* do?"  

**Tools/Methods:**
- Analyzes task completion rates (simulated)
- Identifies drop-off points
- Proposes A/B test ideas
- Synthesizes conflicting feedback

**Personality:** Neutral but firm, data-driven, keeps discussion on track

**Special Powers:**
- Can call "Time Out!" to refocus
- Asks "What would success look like?"
- Forces prioritization ("Fix top 3 only")

## Research Methods Available

Each participant can use these tools during pre-session research or live during discussion:

### Hands-On Testing
- Install and run `docmgr` commands
- Create test tickets
- Try tutorial examples verbatim
- Intentionally make mistakes

### Code Examination
- Read CLI source (`cmd/`, `pkg/`)
- Check help text generation
- Trace command execution
- Review error handling

### Documentation Analysis
- Read tutorial line-by-line
- Map sections to commands
- Check for gaps/redundancy
- Validate examples

### User Flow Simulation
- "Think aloud" while following tutorial
- Time common tasks
- Count steps to accomplish goals
- Identify confusion points

### Pattern Analysis
- Compare to manual workflows (mkdir, vim)
- Check CLI best practices
- Analyze output formats
- Review error handling patterns

## Output Artifacts

### Per Discussion Round
Each round produces a markdown document with:

```markdown
## Pre-Session Research
[What each participant discovered/tested]

## Opening Reactions
[Gut reactions from each participant]

## Deep Dive Discussion
[Passionate exchange of findings, with data]

## Live Experiments
[Real-time tests during discussion]

## Facilitator Synthesis
[Key themes, pain points, wins, action items]

## Proposed Improvements
[Specific changes with before/after]
```

### Final Deliverables

1. **UX Findings Report** (design doc)
   - Top pain points ranked by severity
   - Quick wins identified
   - Major improvements proposed
   - Before/after examples

2. **Tutorial Improvement RFC**
   - Specific restructuring proposals
   - New examples/sections
   - Cuts/consolidations
   - Success metrics

3. **CLI Enhancement Proposals**
   - Flag/command improvements
   - Help text updates
   - Error message rewrites
   - New convenience features

## Related

- [UX Debrief Questions and Research Areas](./02-ux-debrief-questions-and-research-areas.md)
- Tutorial under review: `pkg/doc/docmgr-how-to-use.md`
- Debate framework inspiration: `go-go-mento/ttmp/REORG-FEATURE-STRUCTURE-.../playbooks/playbook-using-debate-framework-for-technical-rfcs.md`
