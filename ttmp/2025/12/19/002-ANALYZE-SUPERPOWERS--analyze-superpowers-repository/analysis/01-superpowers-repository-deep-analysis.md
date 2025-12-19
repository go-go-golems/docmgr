---
Title: Superpowers Repository Deep Analysis
Ticket: 002-ANALYZE-SUPERPOWERS
Status: active
Topics:
    - analysis
    - skills
    - agents
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T14:06:10.31978731-05:00
---

# Superpowers Repository Deep Analysis

## Executive Summary

Superpowers is a skills-based workflow system for coding agents that enforces structured development processes through composable "skills" - markdown documents containing workflows, checklists, and mandatory processes. The system supports three agent platforms (Claude Code, Codex, OpenCode) with platform-specific implementations for skill discovery and injection.

**Key Finding:** The system uses explicit skill listing and description matching rather than semantic search. Agents are instructed to check for applicable skills BEFORE ANY response (including clarifying questions), enforced through strong language and "red flags" tables that prevent rationalization.

## Repository Structure

```
superpowers/
├── skills/              # Core skill library (platform-agnostic)
│   ├── brainstorming/
│   ├── writing-plans/
│   ├── subagent-driven-development/
│   └── ...
├── lib/
│   └── skills-core.js  # Shared skill discovery/parsing (Codex + OpenCode)
├── .codex/             # Codex platform implementation
│   ├── INSTALL.md
│   ├── superpowers-bootstrap.md
│   └── superpowers-codex (CLI script)
├── .opencode/          # OpenCode platform implementation
│   ├── INSTALL.md
│   └── plugin/superpowers.js
├── .claude-plugin/     # Claude Code plugin metadata
│   └── plugin.json
├── commands/           # Shortcut commands that trigger skills
├── agents/             # Agent definitions (code-reviewer)
└── tests/              # Test suites for each platform
```

## Supported Agents and Platforms

### 1. Claude Code (Plugin Marketplace)

**Installation:**
- Via plugin marketplace: `/plugin install superpowers@superpowers-marketplace`
- Plugin metadata in `.claude-plugin/plugin.json`
- Actual injection mechanism not visible in repository (likely similar to OpenCode)

**Characteristics:**
- Plugin marketplace distribution
- Version: 4.0.0
- Skills update automatically when plugin updates

### 2. Codex (Node.js CLI)

**Installation:**
- Manual: Clone repo to `~/.codex/superpowers`
- Add bootstrap section to `~/.codex/AGENTS.md`
- Agent runs CLI commands to discover/load skills

**Implementation:**
- CLI script: `~/.codex/superpowers/.codex/superpowers-codex`
- Commands:
  - `bootstrap` - Load complete bootstrap with all skills
  - `use-skill <name>` - Load specific skill
  - `find-skills` - List all available skills
- Uses shared `lib/skills-core.js` module

**How it works:**
1. Agent runs `superpowers-codex bootstrap` at session start
2. Script outputs markdown with:
   - Bootstrap instructions
   - List of all skills with descriptions
   - Instructions to check skills before ANY task
3. Agent manually invokes `superpowers-codex use-skill <name>` to load skills
4. Skills output as markdown that agent reads

**Tool Mapping:**
- `TodoWrite` → `update_plan`
- `Task` with subagents → Tell user subagents unavailable, do work directly
- `Skill` tool → `superpowers-codex use-skill` command

### 3. OpenCode (Plugin System)

**Installation:**
- Clone to `~/.config/opencode/superpowers`
- Symlink plugin: `ln -sf ~/.config/opencode/superpowers/.opencode/plugin/superpowers.js ~/.config/opencode/plugin/superpowers.js`
- Restart OpenCode

**Implementation:**
- Plugin hooks into OpenCode events:
  - `session.created` - Inject full bootstrap
  - `session.compacted` - Re-inject compact bootstrap
- Provides two custom tools:
  - `use_skill` - Load and inject skill content
  - `find_skills` - List all available skills

**How it works:**
1. Plugin automatically injects bootstrap at session creation
2. Bootstrap includes full "using-superpowers" skill content
3. Skills inserted as synthetic user messages with `noReply: true` (persists across compaction)
4. After context compaction, plugin re-injects compact bootstrap
5. Agent uses `use_skill` tool to load additional skills

**Skill Resolution Priority:**
1. Project skills (`.opencode/skills/`) - Highest
2. Personal skills (`~/.config/opencode/skills/`)
3. Superpowers skills (`~/.config/opencode/superpowers/skills/`)

**Tool Mapping:**
- `TodoWrite` → `update_plan`
- `Task` with subagents → OpenCode's `@mention` system
- `Skill` tool → `use_skill` custom tool

## Skill Discovery Mechanism

### How Skills Are Found

**Not semantic search** - Skills are discovered through explicit listing:

1. **Bootstrap Process:**
   - Bootstrap includes complete list of all skills with descriptions
   - Each skill has YAML frontmatter:
     ```yaml
     ---
     name: skill-name
     description: Use when [condition] - [what it does]
     ---
     ```

2. **Matching Process:**
   - Agent reads skill descriptions
   - Matches current task to skill description
   - No semantic search - explicit description matching

3. **Skill Locations:**
   - Superpowers skills: `skills/` directory (platform-specific path)
   - Personal skills: User's personal skills directory
   - Project skills: `.opencode/skills/` (OpenCode only)

### How Agents Are Told to Search

The "using-superpowers" skill (auto-loaded at bootstrap) contains the core instructions:

**Critical Rules:**
- `<EXTREMELY-IMPORTANT>` tags emphasize mandatory checking
- **Rule:** "Check for skills BEFORE ANY RESPONSE" - even clarifying questions
- **Rule:** "If you think there is even a 1% chance a skill might apply, you ABSOLUTELY MUST read the skill"
- **Rule:** "IF A SKILL APPLIES TO YOUR TASK, YOU DO NOT HAVE A CHOICE. YOU MUST USE IT."

**Red Flags Table:**
Prevents common rationalizations:
- "This is just a simple question" → Questions are tasks, check for skills
- "I need more context first" → Skill check comes BEFORE clarifying questions
- "Let me explore the codebase first" → Skills tell you HOW to explore
- "I remember this skill" → Skills evolve, read current version
- "This doesn't need a formal skill" → If a skill exists, use it

**Skill Priority:**
1. Process skills first (brainstorming, debugging) - determine HOW to approach
2. Implementation skills second (frontend-design, mcp-builder) - guide execution

**Flow Diagram:**
```
User message received
    ↓
Might any skill apply? (even 1% chance)
    ↓ YES
Invoke Skill tool
    ↓
Announce: "Using [skill] to [purpose]"
    ↓
Has checklist?
    ↓ YES
Create TodoWrite todo per item
    ↓
Follow skill exactly
    ↓
Respond (including clarifications)
```

## Skill Usage Instructions

### Skill Structure

Each skill follows this structure:

```markdown
---
name: skill-name
description: Use when [condition] - [what it does]
---

# Skill Title

## Overview
[What the skill does]

## When to Use
[Trigger conditions]

## The Process
[Step-by-step workflow]

## Red Flags
[Common mistakes to avoid]

## Integration
[Other skills this requires/references]
```

### Usage Patterns

**1. Direct Invocation:**
- Skills say "Use this skill exactly as written" or "Adapt principles to context"
- Some skills are "rigid" (TDD, debugging) - must follow exactly
- Some skills are "flexible" (patterns) - adapt principles

**2. Skill Announcement:**
- Skills instruct agents to announce usage: "I'm using the [skill-name] skill to [purpose]"
- Example: "I'm using the writing-plans skill to create the implementation plan."

**3. Skill Chaining:**
- Skills explicitly reference other skills they require
- Format: `superpowers:skill-name` or `@skill-name`
- Example: `writing-plans` says "REQUIRED SUB-SKILL: Use superpowers:executing-plans"

**4. Checklists:**
- Skills with checklists require `TodoWrite` todos for each item
- Example: TDD skill has verification checklist before marking complete

### Example: Brainstorming Skill

**Trigger:** "You MUST use this before any creative work"

**Process:**
1. Understand current project context
2. Ask questions one at a time to refine idea
3. Propose 2-3 approaches with trade-offs
4. Present design in sections (200-300 words)
5. Validate after each section
6. Write design document to `docs/plans/YYYY-MM-DD-<topic>-design.md`
7. Commit design document

**Integration:**
- After design: Use `superpowers:using-git-worktrees` to create workspace
- Then: Use `superpowers:writing-plans` to create implementation plan

## Agent-Specific Prompts

### Code Reviewer Agent

**Location:** `agents/code-reviewer.md` and `skills/requesting-code-review/code-reviewer.md`

**Purpose:** Review completed work against plan and coding standards

**Template Structure:**
```
Task tool (superpowers:code-reviewer):
  WHAT_WAS_IMPLEMENTED: [description]
  PLAN_OR_REQUIREMENTS: [plan reference]
  BASE_SHA: [commit before]
  HEAD_SHA: [current commit]
  DESCRIPTION: [summary]
```

**Output Format:**
- Strengths: What's well done
- Issues: Critical / Important / Minor (with file:line references)
- Recommendations: Improvements
- Assessment: Ready to merge? (Yes/No/With fixes)

**Review Checklist:**
- Code Quality: Separation of concerns, error handling, type safety, DRY, edge cases
- Architecture: Design decisions, scalability, performance, security
- Testing: Tests verify logic (not mocks), edge cases, integration tests
- Requirements: All plan requirements met, matches spec, no scope creep
- Production Readiness: Migration strategy, backward compatibility, documentation

### Subagent-Driven Development Prompts

**Location:** `skills/subagent-driven-development/`

**Three Prompt Templates:**

#### 1. Implementer Prompt (`implementer-prompt.md`)

**Structure:**
```
Task tool (general-purpose):
  description: "Implement Task N: [task name]"
  prompt: |
    ## Task Description
    [FULL TEXT of task from plan]
    
    ## Context
    [Scene-setting: where this fits, dependencies]
    
    ## Before You Begin
    If you have questions, ask them now.
    
    ## Your Job
    1. Implement exactly what task specifies
    2. Write tests (following TDD if required)
    3. Verify implementation works
    4. Commit your work
    5. Self-review (see below)
    6. Report back
    
    ## Before Reporting Back: Self-Review
    - Completeness: Did I fully implement everything?
    - Quality: Is this my best work?
    - Discipline: Did I avoid overbuilding (YAGNI)?
    - Testing: Do tests verify behavior?
    
    ## Report Format
    - What you implemented
    - What you tested and test results
    - Files changed
    - Self-review findings
    - Any issues or concerns
```

**Key Features:**
- Emphasizes asking questions BEFORE starting
- Includes self-review checklist
- Requires full task text (don't make subagent read file)
- Scene-setting context provided

#### 2. Spec Compliance Reviewer (`spec-reviewer-prompt.md`)

**Purpose:** Verify implementation matches spec (nothing more, nothing less)

**Critical Instruction:**
- "DO NOT Trust the Report" - must read actual code
- "The implementer finished suspiciously quickly. Their report may be incomplete, inaccurate, or optimistic."

**Checks:**
- Missing requirements: Did they implement everything requested?
- Extra/unneeded work: Did they build things not requested?
- Misunderstandings: Did they interpret requirements differently?

**Output:**
- ✅ Spec compliant (if everything matches after code inspection)
- ❌ Issues found: [list specifically what's missing or extra, with file:line references]

**Key Principle:** Verify by reading code, not by trusting report

#### 3. Code Quality Reviewer (`code-quality-reviewer-prompt.md`)

**Purpose:** Verify implementation is well-built (clean, tested, maintainable)

**Critical Rule:** Only dispatch AFTER spec compliance review passes

**Uses:** `superpowers:code-reviewer` template (same as code reviewer agent)

**Output:** Strengths, Issues (Critical/Important/Minor), Assessment

### Subagent-Driven Development Workflow

**Process:**
1. Read plan, extract all tasks with full text, create TodoWrite
2. For each task:
   a. Dispatch implementer subagent with full task text + context
   b. Implementer asks questions? → Answer, provide context
   c. Implementer implements, tests, commits, self-reviews
   d. Dispatch spec reviewer subagent
   e. Spec reviewer confirms code matches spec?
      - No → Implementer fixes spec gaps → Re-review
      - Yes → Continue
   f. Dispatch code quality reviewer subagent
   g. Code quality reviewer approves?
      - No → Implementer fixes quality issues → Re-review
      - Yes → Mark task complete
3. After all tasks: Dispatch final code reviewer for entire implementation
4. Use `superpowers:finishing-a-development-branch`

**Red Flags:**
- Never skip reviews (spec compliance OR code quality)
- Never start code quality review before spec compliance passes
- Never proceed with unfixed issues
- Never dispatch multiple implementation subagents in parallel
- Never make subagent read plan file (provide full text instead)

## Key Skills Analysis

### Test-Driven Development (TDD)

**Location:** `skills/test-driven-development/SKILL.md`

**Iron Law:** "NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST"

**Process:** Red-Green-Refactor
1. RED: Write failing test
2. Verify RED: Watch it fail (MANDATORY)
3. GREEN: Minimal code to pass
4. Verify GREEN: Watch it pass (MANDATORY)
5. REFACTOR: Clean up (keep tests green)

**Enforcement:**
- Write code before test? Delete it. Start over.
- No exceptions: Don't keep as "reference", don't "adapt" it
- Extensive rationalization table (common excuses and why they're wrong)
- Red flags: Code before test, test passes immediately, "tests after achieve same purpose"

**Verification Checklist:**
- Every new function/method has a test
- Watched each test fail before implementing
- Each test failed for expected reason
- Wrote minimal code to pass each test
- All tests pass
- Output pristine (no errors, warnings)

### Systematic Debugging

**Location:** `skills/systematic-debugging/SKILL.md`

**Iron Law:** "NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST"

**Four Phases (must complete each before proceeding):**

1. **Root Cause Investigation:**
   - Read error messages carefully
   - Reproduce consistently
   - Check recent changes
   - Gather evidence (especially multi-component systems)
   - Trace data flow (see `root-cause-tracing.md`)

2. **Pattern Analysis:**
   - Find working examples
   - Compare against references (read COMPLETELY, don't skim)
   - Identify differences
   - Understand dependencies

3. **Hypothesis and Testing:**
   - Form single hypothesis
   - Test minimally (one variable at a time)
   - Verify before continuing
   - When you don't know: Say so, ask for help

4. **Implementation:**
   - Create failing test case
   - Implement single fix
   - Verify fix
   - If 3+ fixes failed: Question architecture (not just try again)

**Red Flags:**
- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "One more fix attempt" (when already tried 2+)
- Each fix reveals new problem in different place → Architectural problem

### Writing Plans

**Location:** `skills/writing-plans/SKILL.md`

**Purpose:** Create detailed implementation plans for zero-context engineers

**Assumptions:**
- Engineer has zero context for codebase
- Engineer has questionable taste
- Engineer doesn't know good test design well
- Engineer is skilled developer but unfamiliar with toolset/problem domain

**Task Granularity:** Each step is 2-5 minutes
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement minimal code" - step
- "Run tests and make sure they pass" - step
- "Commit" - step

**Plan Document Structure:**
```markdown
# [Feature Name] Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans

**Goal:** [One sentence]
**Architecture:** [2-3 sentences]
**Tech Stack:** [Key technologies]

---

### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

**Step 1: Write the failing test**
[Complete code]

**Step 2: Run test to verify it fails**
[Exact command with expected output]

**Step 3: Write minimal implementation**
[Complete code]

**Step 4: Run test to verify it passes**
[Exact command with expected output]

**Step 5: Commit**
[Exact git commands]
```

**Remember:**
- Exact file paths always
- Complete code in plan (not "add validation")
- Exact commands with expected output
- Reference relevant skills with @ syntax
- DRY, YAGNI, TDD, frequent commits

**Execution Handoff:**
After saving plan, offer two options:
1. Subagent-Driven (this session) - Use `superpowers:subagent-driven-development`
2. Parallel Session (separate) - New session uses `superpowers:executing-plans`

## Configuration and Bootstrap

### Bootstrap Content

**Core Components:**
1. "using-superpowers" skill content (full or compact)
2. Tool mapping instructions
3. Skills naming/priority rules
4. Critical rules about skill checking

**Codex Bootstrap:**
- Agent runs `superpowers-codex bootstrap`
- Script outputs markdown with:
  - Bootstrap instructions
  - Available skills list (with descriptions)
  - Usage instructions
  - Auto-loads "using-superpowers" skill

**OpenCode Bootstrap:**
- Plugin injects at `session.created` event
- Full bootstrap includes complete "using-superpowers" skill
- Compact bootstrap re-injected after `session.compacted`
- Uses `client.session.prompt()` with `noReply: true` for persistence

### Skill Resolution

**Priority Order (OpenCode):**
1. Project skills (`.opencode/skills/`)
2. Personal skills (`~/.config/opencode/skills/`)
3. Superpowers skills (`~/.config/opencode/superpowers/skills/`)

**Naming:**
- `project:skill-name` - Force project skill lookup
- `skill-name` - Searches project → personal → superpowers
- `superpowers:skill-name` - Force superpowers skill lookup

**Codex:**
- Personal skills override superpowers when names match
- `superpowers:skill-name` forces superpowers lookup
- `skill-name` searches personal first, then superpowers

## Key Design Patterns

### 1. Mandatory Workflows

Skills enforce mandatory workflows through:
- Strong language (`EXTREMELY-IMPORTANT`, "YOU MUST", "NO CHOICE")
- Red flags tables (prevent rationalization)
- Iron Laws ("NO PRODUCTION CODE WITHOUT FAILING TEST FIRST")
- Verification checklists (can't mark complete without checking boxes)

### 2. Two-Stage Review

Subagent-driven development uses two-stage review:
1. Spec compliance review (did they build what was requested?)
2. Code quality review (is it well-built?)

**Critical:** Code quality review only after spec compliance passes

### 3. Template-Based Subagents

Subagent prompts are templates with placeholders:
- Task-specific context filled at dispatch time
- Full task text provided (don't make subagent read file)
- Scene-setting context included
- Self-review checklists built-in

### 4. Skill Chaining

Skills reference other skills explicitly:
- Format: `superpowers:skill-name` or `@skill-name`
- Required sub-skills listed
- Workflow dependencies documented

### 5. Platform Abstraction

Skills are platform-agnostic:
- Skills reference Claude Code tools
- Platform-specific code handles tool mapping
- Bootstrap adapts instructions per platform
- Same skill content works across platforms

## Comparison: How Each Platform Works

| Aspect | Codex | OpenCode | Claude Code |
|--------|-------|----------|-------------|
| **Discovery** | Manual CLI invocation | Automatic plugin injection | Plugin marketplace |
| **Bootstrap** | Agent runs `bootstrap` command | Plugin injects at session start | (Not visible in repo) |
| **Skill Loading** | `use-skill <name>` command | `use_skill` tool | (Not visible) |
| **Persistence** | Agent must re-invoke | Synthetic messages with `noReply: true` | (Not visible) |
| **Context Compaction** | Manual re-invocation | Auto re-injection | (Not visible) |
| **Tool Mapping** | Bootstrap instructions | Plugin provides mapping | (Not visible) |
| **Custom Tools** | None (uses CLI) | `use_skill`, `find_skills` | (Not visible) |

## Key Insights for docmgr Implementation

### 1. Skill Discovery
- **Not semantic search** - Explicit listing with descriptions
- Bootstrap includes all skills upfront
- Agent matches task to skill description
- YAML frontmatter: `name` and `description` (trigger condition)

### 2. Mandatory Checking
- Strong enforcement language prevents skipping
- Red flags tables prevent rationalization
- Check BEFORE any response (even clarifying questions)
- 1% chance rule: If skill might apply, must read it

### 3. Skill Structure
- YAML frontmatter for metadata
- Description serves as trigger condition
- Skills contain workflows, checklists, examples
- Skills reference other skills explicitly

### 4. Platform Adaptation
- Skills are platform-agnostic
- Platform code handles tool mapping
- Bootstrap adapts instructions per platform
- Same skill content works everywhere

### 5. Subagent Patterns
- Template-based prompts with placeholders
- Full context provided (don't make subagent read files)
- Two-stage review: spec compliance then code quality
- Self-review checklists before reporting

### 6. Workflow Enforcement
- Iron Laws and Core Principles
- Verification checklists
- Red flags and rationalization handling
- Integration with other skills

## Recommendations for docmgr Skills Implementation

1. **Use explicit skill listing** (not semantic search) - Bootstrap should list all skills with descriptions
2. **Strong enforcement language** - Use `<EXTREMELY-IMPORTANT>` tags and "YOU MUST" language
3. **Red flags tables** - Prevent common rationalizations
4. **YAML frontmatter** - `name` and `description` fields for matching
5. **Platform abstraction** - Skills reference docmgr tools, platform adapts them
6. **Skill chaining** - Skills explicitly reference other required skills
7. **Verification checklists** - Can't mark complete without checking boxes
8. **Template-based subagents** - If using subagents, use templates with placeholders
