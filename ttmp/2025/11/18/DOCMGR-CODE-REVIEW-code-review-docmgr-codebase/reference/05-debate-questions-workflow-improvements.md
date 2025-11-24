---
Title: Debate Questions - Workflow Improvements
Ticket: DOCMGR-CODE-REVIEW
Status: active
Topics:
    - docmgr
    - code-review
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Five debate questions exploring workflow improvements for closing tasks/tickets and updating status/intent fields, with focus on LLM coding agent usage
LastUpdated: 2025-11-19T14:20:12.084692546-05:00
---

# Debate Questions - Workflow Improvements

## Goal

This document lists all debate questions for exploring workflow improvements around closing tasks/tickets and updating document status/intent fields. Each question maps to specific candidates who will argue positions backed by codebase evidence.

## Question Flow

The questions build on each other:
1. **Foundation:** What's the problem? (Round 1)
2. **Approach:** What verbs/commands do we need? (Round 2)
3. **Lifecycle:** How should status/intent transitions work? (Round 3)
4. **Automation:** What can be automated vs. manual? (Round 4)
5. **LLM Usage:** How should LLM coding agents use these workflows? (Round 5)

## The Questions

### Round 1: What's the actual workflow friction?

**Question:** When closing a ticket or completing work, what are the current pain points? How many commands does a user need to run, and what's the cognitive load?

**Primary Candidates:**
- **Casey** (The New User) — First-time user perspective, confusion points
- **Jordan** (The LLM Expert) — LLM agent workflow analysis, command sequence optimization
- **The Task Manager** (`tasks.go`) — Current task command patterns

**Secondary Candidates (can interject):**
- Sarah (The Pragmatist) — Cost/benefit analysis
- Taylor (DX Expert) — Human developer patterns

**Research Needed:**
- Count command invocations in typical "close ticket" workflow
- Analyze reminder messages in task commands
- Survey common status/intent value combinations
- Review git history for status update patterns

**Key Tensions:**
- Manual control vs. automation
- Flexibility vs. consistency
- Number of commands vs. command complexity

---

### Round 2: What new verbs or command patterns should we add?

**Question:** Should we add high-level verbs like `ticket close` or `ticket complete` that combine multiple operations? Or should we enhance existing commands with flags? What's the right abstraction level?

**Primary Candidates:**
- **Sarah** (The Pragmatist) — Cost of new commands vs. benefit
- **Jordan** (The LLM Expert) — LLM command composability, token efficiency
- **The Metadata Updater** (`meta_update.go`) — Current update patterns

**Secondary Candidates (can interject):**
- Alex (The Architect) — Command structure and consistency
- Taylor (DX Expert) — Human developer mental models
- The Task Manager — Integration opportunities

**Research Needed:**
- Analyze existing command patterns (verb + subcommand vs. flags)
- Count how many commands touch status/intent
- Review similar tools (git, kubectl) for verb patterns
- Map common operation combinations

**Key Tensions:**
- New top-level verbs vs. enhancing existing commands
- Explicit commands vs. implicit automation
- Backwards compatibility vs. breaking changes

---

### Round 3: How should status and intent lifecycle transitions work?

**Question:** Should status/intent transitions be explicit (user chooses) or implicit (derived from task completion)? Should we enforce valid transitions? Should status be vocabulary-controlled like intent?

**Primary Candidates:**
- **Alex** (The Architect) — State machine design, lifecycle management
- **The Data Guardian** (`document.go`) — Status/Intent field semantics
- **The Task Manager** — Task completion → status relationship

**Secondary Candidates (can interject):**
- Jordan (LLM Expert) — LLM workflow patterns
- Taylor (DX Expert) — Human developer ergonomics
- Casey — User expectations

**Research Needed:**
- Analyze current Status value patterns (grep for Status: in docs)
- Review Intent vocabulary structure
- Map common status transitions
- Identify invalid state combinations

**Key Tensions:**
- Free-form Status vs. vocabulary-controlled Status
- Explicit transitions vs. implicit derivation
- Flexibility vs. validation

---

### Round 4: What should be automated vs. manual?

**Question:** When all tasks are checked, should status automatically change? Should intent updates be prompted or automatic? What operations should require explicit user confirmation vs. happening silently?

**Primary Candidates:**
- **Jordan** (The LLM Expert) — Automation opportunities, LLM workflow patterns
- **The Task Manager** — "All tasks done" detection
- **The Metadata Updater** — Bulk operation patterns

**Secondary Candidates (can interject):**
- Sarah — Cost of automation complexity
- Taylor (DX Expert) — Human developer control expectations
- Casey — User control expectations

**Research Needed:**
- Analyze task completion → status update correlation
- Count manual status updates after task completion
- Review reminder messages for automation hints
- Survey user preferences (if available)

**Key Tensions:**
- Automation vs. user control
- Silent operations vs. explicit confirmation
- Smart defaults vs. predictable behavior

---

### Round 5: How should LLM coding agents use these workflows?

**Question:** Since docmgr is primarily used by LLM coding agents, how should the workflow design optimize for LLM usage? What makes commands LLM-friendly vs. human-friendly? Should we have separate high-level commands for LLMs vs. composable low-level commands?

**Primary Candidates:**
- **Jordan** (The LLM Expert) — LLM command patterns, token efficiency, error handling
- **Taylor** (DX Expert) — Human developer needs, dual-mode design
- **The Task Manager** — Current command structure and LLM usability

**Secondary Candidates (can interject):**
- Sarah — Implementation complexity of dual-mode design
- Alex — Architecture for supporting both LLMs and humans
- The Metadata Updater — Command composability patterns

**Research Needed:**
- Analyze current command output formats (structured vs. human-readable)
- Review LLM prompt patterns that use docmgr
- Count token usage for common workflows
- Identify LLM error patterns (parsing failures, ambiguous outputs)
- Survey LLM vs. human command usage patterns

**Key Tensions:**
- High-level LLM commands vs. composable low-level commands
- Structured output (JSON/YAML) vs. human-readable output
- Explicit vs. implicit operations (LLMs prefer explicit)
- Token efficiency vs. command clarity
- Single-mode vs. dual-mode (LLM + human) design

---

## Question Dependencies

- **Round 1 → Round 2:** Understanding friction informs verb design
- **Round 2 → Round 3:** Command structure affects lifecycle design
- **Round 3 → Round 4:** Lifecycle rules determine automation opportunities
- **Round 4 → Round 2:** Automation needs inform command design
- **Round 5 → All:** LLM usage patterns inform all design decisions (primary use case)
- **Round 2 → Round 5:** Command design must work for both LLMs and humans

## Expected Outcomes

After all rounds, we should have:
- Clear understanding of current workflow friction
- Candidate solutions for new verbs/commands
- Lifecycle transition model
- Automation strategy
- LLM-optimized workflow design

These will feed into a design document and RFC.

## Related

- [Debate Format and Candidates](./04-debate-format-and-candidates-workflow-improvements.md) - Candidate personas
- [Intent and Status Fields Analysis](../analysis/11-intent-and-status-fields-analysis.md) - Current state analysis
- Original playbook: `/home/manuel/workspaces/2025-11-03/.../playbook-using-debate-framework-for-technical-rfcs.md`
