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

### Understanding Superpowers: A Developer's Perspective

If you're new to this project, think of Superpowers as a "playbook system" for AI coding assistants. Just like a sports team has playbooks that define exactly how to execute different plays, Superpowers provides coding agents with structured workflows (called "skills") that ensure consistent, high-quality development practices.

The genius of Superpowers lies in its enforcement mechanism. Rather than suggesting "maybe you should write tests first," it mandates "you MUST write a failing test before any production code." This isn't just a preference—it's baked into the system through carefully crafted instructions that prevent agents from skipping steps or taking shortcuts.

What makes this particularly interesting is how it handles the challenge of getting AI agents to follow processes reliably. Anyone who's worked with AI assistants knows they can be creative—sometimes too creative—and might skip important steps in favor of what seems like a faster path. Superpowers solves this by using very explicit instructions, "red flags" that catch common rationalizations, and a mandatory checking step that happens before the agent even responds to the user.

## Repository Structure

The Superpowers repository is organized with a clear separation of concerns. The core skills library lives in the `skills/` directory and is platform-agnostic—meaning the same skill content works whether you're using Claude Code, Codex, or OpenCode. Platform-specific code lives in separate directories (`.codex/`, `.opencode/`, `.claude-plugin/`) and handles the mechanics of how skills are discovered, loaded, and injected into each platform's agent system.

**Repository:** https://github.com/obra/superpowers

```
superpowers/
├── skills/              # Core skill library (platform-agnostic)
│   ├── brainstorming/
│   ├── writing-plans/
│   ├── subagent-driven-development/
│   └── ...
├── lib/
│   └── skills-core.js  # Shared skill discovery/parsing (Codex + OpenCode)
│                       # See: https://github.com/obra/superpowers/blob/main/lib/skills-core.js
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

Superpowers supports three different AI coding agent platforms, each with its own architecture and capabilities. Understanding these differences is crucial because they affect how skills are discovered and loaded. The core insight here is that while the skill content itself is platform-agnostic, the mechanism for getting skills into the agent's context varies significantly between platforms.

### 1. Claude Code (Plugin Marketplace)

**Installation:**
- Via plugin marketplace: `/plugin install superpowers@superpowers-marketplace`
- Plugin metadata in [`superpowers/.claude-plugin/plugin.json`](https://github.com/obra/superpowers/blob/main/.claude-plugin/plugin.json)
- Actual injection mechanism not visible in repository (likely similar to OpenCode)

**Characteristics:**
- Plugin marketplace distribution
- Version: 4.0.0 (from [`plugin.json`](https://github.com/obra/superpowers/blob/main/.claude-plugin/plugin.json))
- Skills update automatically when plugin updates

### 2. Codex (Node.js CLI)

**Installation:**
- Manual: Clone repo to `~/.codex/superpowers`
- Add bootstrap section to `~/.codex/AGENTS.md`
- Agent runs CLI commands to discover/load skills

**Implementation:**
- CLI script: [`superpowers/.codex/superpowers-codex`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-codex) (Node.js executable)
- Commands:
  - `bootstrap` - Load complete bootstrap with all skills
  - `use-skill <name>` - Load specific skill
  - `find-skills` - List all available skills
- Uses shared [`superpowers/lib/skills-core.js`](https://github.com/obra/superpowers/blob/main/lib/skills-core.js) module
- Bootstrap file: [`superpowers/.codex/superpowers-bootstrap.md`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-bootstrap.md)

**How it works:**
1. Agent runs `superpowers-codex bootstrap` at session start
2. Script outputs markdown with:
   - Bootstrap instructions
   - List of all skills with descriptions
   - Instructions to check skills before ANY task
3. Agent manually invokes `superpowers-codex use-skill <name>` to load skills
4. Skills output as markdown that agent reads

**Tool Mapping** (from [`superpowers/.codex/superpowers-bootstrap.md`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-bootstrap.md)):
```markdown
**Tool Mapping for Codex:**
When skills reference tools you don't have, substitute your equivalent tools:
- `TodoWrite` → `update_plan` (your planning/task tracking tool)
- `Task` tool with subagents → Tell the user that subagents aren't available in Codex yet and you'll do the work the subagent would do
- `Skill` tool → `~/.codex/superpowers/.codex/superpowers-codex use-skill` command (already available)
- `Read`, `Write`, `Edit`, `Bash` → Use your native tools with similar functions
```

The Codex approach is interesting because it relies on the agent's ability to execute shell commands. The agent must actively participate in loading skills by running CLI commands, which means the bootstrap process is more of a "conversation starter" than an automatic injection. This gives the agent more control but also requires more discipline to follow the process.

### 3. OpenCode (Plugin System)

**Installation:**
- Clone to `~/.config/opencode/superpowers`
- Symlink plugin: `ln -sf ~/.config/opencode/superpowers/.opencode/plugin/superpowers.js ~/.config/opencode/plugin/superpowers.js`
- Restart OpenCode
- Installation docs: [`superpowers/.opencode/INSTALL.md`](https://github.com/obra/superpowers/blob/main/.opencode/INSTALL.md)

**Implementation:**
- Plugin file: [`superpowers/.opencode/plugin/superpowers.js`](https://github.com/obra/superpowers/blob/main/.opencode/plugin/superpowers.js)
- Plugin hooks into OpenCode events:
  - `session.created` - Inject full bootstrap
  - `session.compacted` - Re-inject compact bootstrap
- Provides two custom tools:
  - `use_skill` - Load and inject skill content
  - `find_skills` - List all available skills
- Uses shared [`superpowers/lib/skills-core.js`](https://github.com/obra/superpowers/blob/main/lib/skills-core.js) module

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

**Tool Mapping** (from [`superpowers/.opencode/plugin/superpowers.js`](https://github.com/obra/superpowers/blob/main/.opencode/plugin/superpowers.js)):
```javascript
**Tool Mapping for OpenCode:**
When skills reference tools you don't have, substitute OpenCode equivalents:
- `TodoWrite` → `update_plan`
- `Task` tool with subagents → Use OpenCode's subagent system (@mention)
- `Skill` tool → `use_skill` custom tool
- `Read`, `Write`, `Edit`, `Bash` → Your native tools
```

OpenCode's plugin system is the most sophisticated of the three implementations. It hooks into the platform's event system to automatically inject skills at the right moments, and it handles context compaction gracefully by re-injecting a compact version of the bootstrap. This means the agent doesn't need to remember to load skills—the system ensures they're always available.

## Skill Discovery Mechanism

One of the most important insights from analyzing Superpowers is how skill discovery works. At first glance, you might assume the system uses semantic search—asking "what skills are relevant to this task?" and letting the AI figure it out. But Superpowers takes a different, more explicit approach that's worth understanding.

### How Skills Are Found

**Not semantic search** - Skills are discovered through explicit listing:

The system doesn't rely on the AI's ability to semantically match tasks to skills. Instead, it provides a complete list of all available skills with their descriptions upfront, and the agent is responsible for matching the current task to the appropriate skill description. This explicit approach reduces ambiguity and ensures consistent behavior.

1. **Bootstrap Process:**
   - Bootstrap includes complete list of all skills with descriptions
   - Each skill has YAML frontmatter (parsed by [`superpowers/lib/skills-core.js`](https://github.com/obra/superpowers/blob/main/lib/skills-core.js) `extractFrontmatter()` function):
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
   - Skill discovery handled by [`superpowers/lib/skills-core.js`](https://github.com/obra/superpowers/blob/main/lib/skills-core.js) `findSkillsInDir()` function

### How Agents Are Told to Search

The "using-superpowers" skill (auto-loaded at bootstrap) contains the core instructions. This skill is located at [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md).

This is where Superpowers gets really interesting from a prompt engineering perspective. The system doesn't just suggest checking for skills—it mandates it with language strong enough to prevent the common AI tendency to skip steps. The instructions are designed to catch the agent before it can rationalize its way out of following the process.

**Critical Rules** (exact quotes from [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md)):

```markdown
<EXTREMELY-IMPORTANT>
If you think there is even a 1% chance a skill might apply to what you are doing, you ABSOLUTELY MUST read the skill.

IF A SKILL APPLIES TO YOUR TASK, YOU DO NOT HAVE A CHOICE. YOU MUST USE IT.

This is not negotiable. This is not optional. You cannot rationalize your way out of this.
</EXTREMELY-IMPORTANT>

## The Rule

**Check for skills BEFORE ANY RESPONSE.** This includes clarifying questions. Even 1% chance means invoke the Skill tool first.
```

**Red Flags Table** (exact quote from [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md)):

```markdown
## Red Flags

These thoughts mean STOP—you're rationalizing:

| Thought | Reality |
|---------|---------|
| "This is just a simple question" | Questions are tasks. Check for skills. |
| "I need more context first" | Skill check comes BEFORE clarifying questions. |
| "Let me explore the codebase first" | Skills tell you HOW to explore. Check first. |
| "I can check git/files quickly" | Files lack conversation context. Check for skills. |
| "Let me gather information first" | Skills tell you HOW to gather information. |
| "This doesn't need a formal skill" | If a skill exists, use it. |
| "I remember this skill" | Skills evolve. Read current version. |
| "This doesn't count as a task" | Action = task. Check for skills. |
| "The skill is overkill" | Simple things become complex. Use it. |
| "I'll just do this one thing first" | Check BEFORE doing anything. |
| "This feels productive" | Undisciplined action wastes time. Skills prevent this. |
```

The red flags table is particularly clever—it anticipates the exact rationalizations that AI agents (and humans) use to skip process steps. By explicitly calling these out and providing counter-arguments, the system prevents the agent from convincing itself that "this time is different" or "I can skip this step."

**Skill Priority** (from [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md)):
```markdown
## Skill Priority

When multiple skills could apply, use this order:

1. **Process skills first** (brainstorming, debugging) - these determine HOW to approach the task
2. **Implementation skills second** (frontend-design, mcp-builder) - these guide execution

"Let's build X" → brainstorming first, then implementation skills.
"Fix this bug" → debugging first, then domain-specific skills.
```

**Flow Diagram** (from [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md)):
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

Once an agent has discovered a relevant skill, it needs to know how to use it. Superpowers skills are more than just documentation—they're executable workflows with clear entry points, step-by-step processes, and explicit integration points with other skills.

### Skill Structure

Each skill follows this structure. Example from [`superpowers/skills/brainstorming/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/brainstorming/SKILL.md):

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

Skills follow several usage patterns that ensure they're applied correctly:

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

To make this concrete, let's look at how the brainstorming skill works. This skill demonstrates many of the patterns we've discussed. Full skill file: [`superpowers/skills/brainstorming/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/brainstorming/SKILL.md)

**Trigger** (from frontmatter):
```yaml
---
name: brainstorming
description: "You MUST use this before any creative work - creating features, building components, adding functionality, or modifying behavior. Explores user intent, requirements and design before implementation."
---
```

**Process** (exact from [`superpowers/skills/brainstorming/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/brainstorming/SKILL.md)):
1. Understand current project context
2. Ask questions one at a time to refine idea
3. Propose 2-3 approaches with trade-offs
4. Present design in sections (200-300 words)
5. Validate after each section
6. Write design document to `docs/plans/YYYY-MM-DD-<topic>-design.md`
7. Commit design document

**Integration** (from [`superpowers/skills/brainstorming/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/brainstorming/SKILL.md)):
```markdown
**Implementation (if continuing):**
- Ask: "Ready to set up for implementation?"
- Use superpowers:using-git-worktrees to create isolated workspace
- Use superpowers:writing-plans to create detailed implementation plan
```

## Agent-Specific Prompts

Superpowers doesn't just provide skills for the main agent—it also defines specialized subagents that handle specific tasks like code review and implementation. These subagents have carefully crafted prompts that ensure they perform their roles correctly. Understanding these prompts is crucial because they reveal how Superpowers maintains quality through structured review processes.

### Code Reviewer Agent

**Location:** 
- [`superpowers/agents/code-reviewer.md`](https://github.com/obra/superpowers/blob/main/agents/code-reviewer.md) - Agent definition
- [`superpowers/skills/requesting-code-review/code-reviewer.md`](https://github.com/obra/superpowers/blob/main/skills/requesting-code-review/code-reviewer.md) - Prompt template

**Purpose:** Review completed work against plan and coding standards

**Template Structure** (from [`superpowers/skills/requesting-code-review/code-reviewer.md`](https://github.com/obra/superpowers/blob/main/skills/requesting-code-review/code-reviewer.md)):
```markdown
Task tool (superpowers:code-reviewer):
  Use template at requesting-code-review/code-reviewer.md

  WHAT_WAS_IMPLEMENTED: [from implementer's report]
  PLAN_OR_REQUIREMENTS: Task N from [plan-file]
  BASE_SHA: [commit before task]
  HEAD_SHA: [current commit]
  DESCRIPTION: [task summary]
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

**Location:** [`superpowers/skills/subagent-driven-development/`](https://github.com/obra/superpowers/tree/main/skills/subagent-driven-development)

Subagent-driven development is one of Superpowers' most sophisticated features. Instead of having one agent work through an entire plan, it dispatches fresh subagents for each task. This prevents context pollution and ensures each task gets focused attention. The prompts for these subagents are templates that get filled with task-specific context at dispatch time.

**Three Prompt Templates:**

#### 1. Implementer Prompt

**File:** [`superpowers/skills/subagent-driven-development/implementer-prompt.md`](https://github.com/obra/superpowers/blob/main/skills/subagent-driven-development/implementer-prompt.md)

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

Notice how the implementer prompt emphasizes asking questions upfront. This prevents the common problem of subagents making assumptions and implementing the wrong thing. The self-review checklist is also important—it catches issues before they're passed to reviewers.

#### 2. Spec Compliance Reviewer

**File:** [`superpowers/skills/subagent-driven-development/spec-reviewer-prompt.md`](https://github.com/obra/superpowers/blob/main/skills/subagent-driven-development/spec-reviewer-prompt.md)

**Purpose:** Verify implementation matches spec (nothing more, nothing less)

**Critical Instruction** (exact quote):
```markdown
## CRITICAL: Do Not Trust the Report

The implementer finished suspiciously quickly. Their report may be incomplete,
inaccurate, or optimistic. You MUST verify everything independently.

**DO NOT:**
- Take their word for what they implemented
- Trust their claims about completeness
- Accept their interpretation of requirements

**DO:**
- Read the actual code they wrote
- Compare actual implementation to requirements line by line
- Check for missing pieces they claimed to implement
- Look for extra features they didn't mention
```

**Checks:**
- Missing requirements: Did they implement everything requested?
- Extra/unneeded work: Did they build things not requested?
- Misunderstandings: Did they interpret requirements differently?

**Output:**
- ✅ Spec compliant (if everything matches after code inspection)
- ❌ Issues found: [list specifically what's missing or extra, with file:line references]

**Key Principle:** Verify by reading code, not by trusting report

The spec reviewer's instruction to "DO NOT Trust the Report" is particularly important. It recognizes that implementers might claim they've done everything correctly, but the actual code might tell a different story. This skepticism is built into the review process to catch issues early.

#### 3. Code Quality Reviewer

**File:** [`superpowers/skills/subagent-driven-development/code-quality-reviewer-prompt.md`](https://github.com/obra/superpowers/blob/main/skills/subagent-driven-development/code-quality-reviewer-prompt.md)

**Purpose:** Verify implementation is well-built (clean, tested, maintainable)

**Critical Rule:** Only dispatch AFTER spec compliance review passes

**Uses:** `superpowers:code-reviewer` template (same as code reviewer agent at [`superpowers/skills/requesting-code-review/code-reviewer.md`](https://github.com/obra/superpowers/blob/main/skills/requesting-code-review/code-reviewer.md))

**Output:** Strengths, Issues (Critical/Important/Minor), Assessment

### Subagent-Driven Development Workflow

The subagent-driven development workflow is a carefully orchestrated process that ensures quality through multiple review stages. Each task goes through the same rigorous process, which might seem slow but actually prevents costly mistakes and rework.

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

To understand how Superpowers enforces quality, it's helpful to examine a few key skills in detail. These skills demonstrate the patterns and enforcement mechanisms used throughout the system. They're not just suggestions—they're mandatory workflows with built-in safeguards against skipping steps.

### Test-Driven Development (TDD)

**Location:** [`superpowers/skills/test-driven-development/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/test-driven-development/SKILL.md)

**Iron Law** (exact quote):
```markdown
## The Iron Law

NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

**Process:** Red-Green-Refactor
1. RED: Write failing test
2. Verify RED: Watch it fail (MANDATORY)
3. GREEN: Minimal code to pass
4. Verify GREEN: Watch it pass (MANDATORY)
5. REFACTOR: Clean up (keep tests green)

**Enforcement** (exact quotes from [`superpowers/skills/test-driven-development/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/test-driven-development/SKILL.md)):
```markdown
Write code before the test? Delete it. Start over.

**No exceptions:**
- Don't keep it as "reference"
- Don't "adapt" it while writing tests
- Don't look at it
- Delete means delete

Implement fresh from tests. Period.
```

The skill includes an extensive rationalization table covering common excuses like "Too simple to test", "I'll test after", "Already manually tested", etc., each with a counter-argument explaining why the excuse doesn't hold.

The TDD skill's enforcement is particularly strict. It recognizes that the temptation to write code first is strong, and it provides no escape hatches. The "delete it, start over" instruction might seem harsh, but it's necessary because keeping the code creates a psychological barrier to writing proper tests—you'll be tempted to adapt the tests to match the code rather than writing tests that define the desired behavior.

**Verification Checklist:**
- Every new function/method has a test
- Watched each test fail before implementing
- Each test failed for expected reason
- Wrote minimal code to pass each test
- All tests pass
- Output pristine (no errors, warnings)

### Systematic Debugging

**Location:** [`superpowers/skills/systematic-debugging/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/systematic-debugging/SKILL.md)

**Iron Law** (exact quote):
```markdown
## The Iron Law

NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

The systematic debugging skill addresses one of the most common problems in software development: the tendency to apply quick fixes without understanding the root cause. Anyone who's spent hours debugging knows the frustration of fixing symptoms only to have the problem reappear in a different form. This skill enforces a methodical approach that prevents that cycle.

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

The "if 3+ fixes failed" rule is particularly insightful. It recognizes that sometimes the problem isn't with the implementation—it's with the architecture itself. After multiple failed fix attempts, the skill instructs the agent to step back and question whether the fundamental approach is sound, rather than continuing to patch symptoms.

**Red Flags:**
- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "One more fix attempt" (when already tried 2+)
- Each fix reveals new problem in different place → Architectural problem

### Writing Plans

**Location:** [`superpowers/skills/writing-plans/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/writing-plans/SKILL.md)

**Purpose:** Create detailed implementation plans for zero-context engineers

The writing-plans skill is fascinating because it assumes the worst-case scenario: an engineer with zero context about your codebase, questionable taste, and limited knowledge of your toolset. This might seem pessimistic, but it ensures that plans are detailed enough for anyone to follow, which is crucial when plans might be executed by subagents or in separate sessions where context is lost.

**Assumptions** (exact quote from [`superpowers/skills/writing-plans/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/writing-plans/SKILL.md)):
```markdown
Write comprehensive implementation plans assuming the engineer has zero context for our codebase and questionable taste. Document everything they need to know: which files to touch for each task, code, testing, docs they might need to check, how to test it. Give them the whole plan as bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.

Assume they are a skilled developer, but know almost nothing about our toolset or problem domain. Assume they don't know good test design very well.
```

**Task Granularity:** Each step is 2-5 minutes
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement minimal code" - step
- "Run tests and make sure they pass" - step
- "Commit" - step

The granularity here is intentional. Breaking tasks into 2-5 minute steps might seem excessive, but it ensures that each step is small enough to verify independently. This prevents the common problem of "I thought I did everything" when actually several steps were skipped or done incorrectly.

**Plan Document Structure** (exact from [`superpowers/skills/writing-plans/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/writing-plans/SKILL.md)):
```markdown
# [Feature Name] Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

---

### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

**Step 1: Write the failing test**

```python
def test_specific_behavior():
    result = function(input)
    assert result == expected
```

**Step 2: Run test to verify it fails**

Run: `pytest tests/path/test.py::test_name -v`
Expected: FAIL with "function not defined"

**Step 3: Write minimal implementation**

```python
def function(input):
    return expected
```

**Step 4: Run test to verify it passes**

Run: `pytest tests/path/test.py::test_name -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tests/path/test.py src/path/file.py
git commit -m "feat: add specific feature"
```
```

**Remember** (from [`superpowers/skills/writing-plans/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/writing-plans/SKILL.md)):
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

The bootstrap process is how Superpowers gets its instructions into the agent's context. This happens differently on each platform, but the goal is the same: ensure the agent knows about skills and understands how to use them before it starts working.

### Bootstrap Content

**Core Components:**
1. "using-superpowers" skill content (full or compact)
2. Tool mapping instructions
3. Skills naming/priority rules
4. Critical rules about skill checking

**Codex Bootstrap** (from [`superpowers/.codex/superpowers-bootstrap.md`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-bootstrap.md)):
- Agent runs `superpowers-codex bootstrap`
- Script outputs markdown with:
  - Bootstrap instructions
  - Available skills list (with descriptions)
  - Usage instructions
  - Auto-loads "using-superpowers" skill
- Bootstrap file content includes:
```markdown
<EXTREMELY_IMPORTANT>
You have superpowers.

**Tool for running skills:**
- `~/.codex/superpowers/.codex/superpowers-codex use-skill <skill-name>`

**Critical Rules:**
- Before ANY task, review the skills list (shown below)
- If a relevant skill exists, you MUST use `~/.codex/superpowers/.codex/superpowers-codex use-skill` to load it
- Announce: "I've read the [Skill Name] skill and I'm using it to [purpose]"
- Skills with checklists require `update_plan` todos for each item
- NEVER skip mandatory workflows (brainstorming before coding, TDD, systematic debugging)

IF A SKILL APPLIES TO YOUR TASK, YOU DO NOT HAVE A CHOICE. YOU MUST USE IT.
</EXTREMELY_IMPORTANT>
```

**OpenCode Bootstrap** (from [`superpowers/.opencode/plugin/superpowers.js`](https://github.com/obra/superpowers/blob/main/.opencode/plugin/superpowers.js)):
- Plugin injects at `session.created` event
- Full bootstrap includes complete "using-superpowers" skill
- Compact bootstrap re-injected after `session.compacted`
- Uses `client.session.prompt()` with `noReply: true` for persistence
- Bootstrap content generated by `getBootstrapContent()` function in the plugin

### Skill Resolution

When multiple skills with the same name exist (for example, a project-specific skill and a superpowers skill), the system needs rules for which one to use. This is handled through priority ordering:

**Priority Order (OpenCode):**
1. Project skills (`.opencode/skills/`)
2. Personal skills (`~/.config/opencode/skills/`)
3. Superpowers skills (`~/.config/opencode/superpowers/skills/`)

**Naming** (from [`superpowers/.opencode/INSTALL.md`](https://github.com/obra/superpowers/blob/main/.opencode/INSTALL.md)):
- `project:skill-name` - Force project skill lookup
- `skill-name` - Searches project → personal → superpowers
- `superpowers:skill-name` - Force superpowers skill lookup

**Codex** (from [`superpowers/.codex/superpowers-bootstrap.md`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-bootstrap.md)):
- Personal skills override superpowers when names match
- `superpowers:skill-name` forces superpowers lookup
- `skill-name` searches personal first, then superpowers

## Key Design Patterns

Throughout the Superpowers system, several design patterns emerge that are worth understanding. These patterns aren't accidental—they're carefully chosen to solve specific problems in getting AI agents to follow structured processes reliably.

### 1. Mandatory Workflows

Skills enforce mandatory workflows through:
- Strong language (`EXTREMELY-IMPORTANT`, "YOU MUST", "NO CHOICE") - see [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md)
- Red flags tables (prevent rationalization) - see [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md) lines 42-58
- Iron Laws ("NO PRODUCTION CODE WITHOUT FAILING TEST FIRST") - see [`superpowers/skills/test-driven-development/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/test-driven-development/SKILL.md) line 34
- Verification checklists (can't mark complete without checking boxes) - see [`superpowers/skills/test-driven-development/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/test-driven-development/SKILL.md) lines 327-340

### 2. Two-Stage Review

Subagent-driven development uses two-stage review:
1. Spec compliance review (did they build what was requested?)
2. Code quality review (is it well-built?)

**Critical:** Code quality review only after spec compliance passes

The two-stage review process is important because it separates "did we build the right thing?" from "did we build it well?" This prevents situations where beautiful, well-written code doesn't actually solve the problem, or where correct code is so poorly written that it's unmaintainable. By checking spec compliance first, you ensure correctness before optimizing for quality.

### 3. Template-Based Subagents

Subagent prompts are templates with placeholders:
- Task-specific context filled at dispatch time
- Full task text provided (don't make subagent read file)
- Scene-setting context included
- Self-review checklists built-in

The template-based approach ensures consistency while allowing customization. Rather than writing a new prompt for each task, the system fills in placeholders with task-specific information. The instruction to "don't make subagent read file" is important—it ensures the subagent has all the context it needs upfront, preventing the common problem of subagents misunderstanding requirements because they didn't read the plan carefully.

**Example template** (from [`superpowers/skills/subagent-driven-development/implementer-prompt.md`](https://github.com/obra/superpowers/blob/main/skills/subagent-driven-development/implementer-prompt.md)):
```markdown
Task tool (general-purpose):
  description: "Implement Task N: [task name]"
  prompt: |
    ## Task Description
    
    [FULL TEXT of task from plan - paste it here, don't make subagent read file]
    
    ## Context
    
    [Scene-setting: where this fits, dependencies, architectural context]
```

### 4. Skill Chaining

Skills reference other skills explicitly:
- Format: `superpowers:skill-name` or `@skill-name`
- Required sub-skills listed
- Workflow dependencies documented

Skill chaining creates a workflow system where skills build on each other. For example, the brainstorming skill leads to writing-plans, which leads to executing-plans. By explicitly documenting these dependencies, the system ensures agents follow complete workflows rather than skipping steps or using skills in isolation.

**Example** (from [`superpowers/skills/brainstorming/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/brainstorming/SKILL.md)):
```markdown
**Implementation (if continuing):**
- Ask: "Ready to set up for implementation?"
- Use superpowers:using-git-worktrees to create isolated workspace
- Use superpowers:writing-plans to create detailed implementation plan
```

**Another example** (from [`superpowers/skills/writing-plans/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/writing-plans/SKILL.md)):
```markdown
# [Feature Name] Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.
```

### 5. Platform Abstraction

Skills are platform-agnostic:
- Skills reference Claude Code tools
- Platform-specific code handles tool mapping
- Bootstrap adapts instructions per platform
- Same skill content works across platforms

Platform abstraction is crucial for maintainability. By keeping skills platform-agnostic and handling tool mapping at the platform level, Superpowers can support multiple platforms without duplicating skill content. When a skill needs updating, it only needs to be changed once, and all platforms benefit. The tool mapping happens transparently, so skills can reference "TodoWrite" and each platform adapts it to its native tool.

**Tool mapping examples:**
- Codex: [`superpowers/.codex/superpowers-bootstrap.md`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-bootstrap.md) lines 9-14
- OpenCode: [`superpowers/.opencode/plugin/superpowers.js`](https://github.com/obra/superpowers/blob/main/.opencode/plugin/superpowers.js) lines 32-47

## Comparison: How Each Platform Works

To help you understand the differences between platforms, here's a side-by-side comparison. The key insight is that each platform has different capabilities and constraints, which affects how Superpowers integrates with them:

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

If you're implementing a similar skills system for docmgr, here are the key insights from analyzing Superpowers. These aren't just technical details—they're design decisions that solve real problems in getting AI agents to follow structured processes.

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

Based on this analysis, here are concrete recommendations for implementing a skills system in docmgr. These recommendations are derived from what works well in Superpowers and what patterns solve the core challenges of getting AI agents to follow structured processes:

1. **Use explicit skill listing** (not semantic search) - Bootstrap should list all skills with descriptions
2. **Strong enforcement language** - Use `<EXTREMELY-IMPORTANT>` tags and "YOU MUST" language
3. **Red flags tables** - Prevent common rationalizations
4. **YAML frontmatter** - `name` and `description` fields for matching
5. **Platform abstraction** - Skills reference docmgr tools, platform adapts them
6. **Skill chaining** - Skills explicitly reference other required skills
7. **Verification checklists** - Can't mark complete without checking boxes
8. **Template-based subagents** - If using subagents, use templates with placeholders
