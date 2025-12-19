---
Title: Diary
Ticket: 002-ANALYZE-SUPERPOWERS
Status: active
Topics:
    - analysis
    - skills
    - agents
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T14:06:10.31978731-05:00
---

# Diary

## Goal

Document the step-by-step analysis of the Superpowers repository to understand:
- Which prompts are used for different coding agents
- Which agents are supported (Claude Code, Codex, OpenCode)
- How things are configured for each platform
- How the system tells models to search for skills
- How the system tells models to use skills

## Step 1: Initial Exploration and Repository Structure

Started by exploring the repository structure to understand the overall organization. Found three platform-specific directories: `.codex/`, `.opencode/`, and `.claude-plugin/`, indicating support for three different agent platforms.

**What I searched for:**
- Repository root structure
- Platform-specific directories
- README.md for overview

**What I found:**
- Three platform implementations: Codex (Node.js CLI), OpenCode (plugin system), Claude Code (plugin marketplace)
- Skills stored in `skills/` directory, each with a `SKILL.md` file
- Shared core library at `lib/skills-core.js` used by Codex and OpenCode
- Commands directory with trigger files for specific workflows
- Agents directory with code-reviewer agent definition

**What I got from it:**
- Clear separation of platform-specific implementations
- Skills are the core abstraction, platform-specific code handles discovery/loading
- Commands provide shortcuts to trigger specific skills

## Step 2: Understanding Platform-Specific Implementations

Examined how each platform loads and injects skills into agent sessions. Found three distinct approaches:

**What I searched for:**
- `.codex/INSTALL.md` - Codex installation and bootstrap process
- `.codex/superpowers-bootstrap.md` - Bootstrap instructions for Codex
- `.codex/superpowers-codex` - CLI script for Codex
- `.opencode/INSTALL.md` - OpenCode installation
- `.opencode/plugin/superpowers.js` - OpenCode plugin implementation
- `.claude-plugin/plugin.json` - Claude Code plugin metadata

**What I found:**

**Codex:**
- Uses a Node.js CLI script (`superpowers-codex`) that can be invoked by the agent
- Bootstrap process: Agent runs `superpowers-codex bootstrap` which outputs markdown instructions
- Skills loaded via `superpowers-codex use-skill <name>` command
- Bootstrap includes list of all available skills and instructions to check before ANY task
- Uses shared `lib/skills-core.js` for skill discovery

**OpenCode:**
- Plugin system that hooks into `session.created` and `session.compacted` events
- Automatically injects bootstrap content via `client.session.prompt()` with `noReply: true`
- Provides two custom tools: `use_skill` and `find_skills`
- Skills inserted as synthetic user messages that persist across context compaction
- Re-injects compact bootstrap after context compaction events

**Claude Code:**
- Plugin marketplace system (minimal files found - likely uses similar injection pattern)
- Plugin.json defines metadata but actual injection mechanism not visible in repo

**What I got from it:**
- Codex: Manual invocation model (agent must run CLI commands)
- OpenCode: Automatic injection model (plugin handles everything)
- Claude Code: Plugin marketplace (details not in repo, likely similar to OpenCode)

## Step 3: Analyzing Skill Discovery and Search Instructions

Examined how the system instructs agents to search for and discover skills.

**What I searched for:**
- `skills/using-superpowers/SKILL.md` - Core skill that establishes the system
- Bootstrap files for each platform
- Skill discovery mechanisms

**What I found:**

**The "using-superpowers" skill is the foundation:**
- Contains `<EXTREMELY-IMPORTANT>` tags emphasizing mandatory skill checking
- Rule: "Check for skills BEFORE ANY RESPONSE" - even clarifying questions
- Rule: "If you think there is even a 1% chance a skill might apply, you ABSOLUTELY MUST read the skill"
- Includes red flags table showing common rationalizations to avoid
- Defines skill priority (process skills first, then implementation skills)

**Bootstrap injection patterns:**

**Codex:**
- Bootstrap markdown includes list of all skills with descriptions
- Instructions tell agent to run `superpowers-codex bootstrap` at session start
- Agent must manually invoke `superpowers-codex use-skill` to load skills

**OpenCode:**
- Plugin automatically injects bootstrap at `session.created` event
- Bootstrap includes full "using-superpowers" skill content
- Compact version re-injected after `session.compacted` events
- Tool mapping instructions included (TodoWrite→update_plan, etc.)

**What I got from it:**
- Discovery happens via explicit skill listing in bootstrap
- Search is not semantic - agent sees full list and must match task to skill description
- Mandatory checking enforced through strong language and red flags
- Skills have YAML frontmatter with `name` and `description` fields for matching

## Step 4: Understanding Skill Usage Instructions

Examined how skills tell agents to use them and what happens when loaded.

**What I searched for:**
- Example skills: `brainstorming/SKILL.md`, `writing-plans/SKILL.md`, `subagent-driven-development/SKILL.md`
- Skill structure and format
- How skills reference other skills

**What I found:**

**Skill structure:**
- YAML frontmatter with `name` and `description` (description is the trigger condition)
- Description format: "Use when [condition] - [what it does]"
- Skills contain detailed instructions, workflows, checklists
- Skills can reference other skills using `superpowers:skill-name` syntax

**Usage patterns:**

**Direct invocation:**
- Skills say "Use this skill exactly as written"
- Some skills are "rigid" (TDD, debugging) - must follow exactly
- Some skills are "flexible" (patterns) - adapt principles to context

**Subagent prompts:**
- `subagent-driven-development` includes three prompt templates:
  - `implementer-prompt.md` - Instructions for implementer subagent
  - `spec-reviewer-prompt.md` - Instructions for spec compliance reviewer
  - `code-quality-reviewer-prompt.md` - Instructions for code quality reviewer
- These are templates filled with task-specific context when dispatching subagents

**Skill chaining:**
- Skills explicitly reference other skills they require
- Example: `writing-plans` says "REQUIRED SUB-SKILL: Use superpowers:executing-plans"
- Skills announce when they're being used: "I'm using the writing-plans skill..."

**What I got from it:**
- Skills are self-contained workflows with clear entry points
- Subagent prompts are parameterized templates, not static instructions
- Skill descriptions serve as matching criteria for agents
- Skills enforce their own usage through explicit instructions

## Step 5: Analyzing Agent-Specific Prompts

Examined the different prompt structures used for different agent types.

**What I searched for:**
- `agents/code-reviewer.md` - Code reviewer agent definition
- `skills/requesting-code-review/code-reviewer.md` - Code review prompt template
- Subagent prompt templates in `subagent-driven-development/`
- Test files showing skill triggering

**What I found:**

**Code Reviewer Agent:**
- Defined in `agents/code-reviewer.md` with detailed role description
- Used via `requesting-code-review` skill
- Template at `skills/requesting-code-review/code-reviewer.md` with placeholders:
  - `{WHAT_WAS_IMPLEMENTED}`
  - `{PLAN_OR_REQUIREMENTS}`
  - `{BASE_SHA}`, `{HEAD_SHA}`
  - `{DESCRIPTION}`
- Output format: Strengths, Issues (Critical/Important/Minor), Assessment

**Subagent Prompt Templates:**

**Implementer:**
- Template includes task description, context, "Before You Begin" section
- Emphasizes asking questions before starting
- Includes self-review checklist before reporting back
- Report format: what implemented, tests, files changed, self-review findings

**Spec Reviewer:**
- Focus: Verify implementation matches spec (nothing more, nothing less)
- Critical instruction: "DO NOT Trust the Report" - must read actual code
- Checks for: missing requirements, extra/unneeded work, misunderstandings
- Output: ✅ Spec compliant OR ❌ Issues found with file:line references

**Code Quality Reviewer:**
- Only dispatched after spec compliance passes
- Uses `superpowers:code-reviewer` template
- Focus: Clean, tested, maintainable code
- Output: Strengths, Issues by severity, Assessment

**What I got from it:**
- Agents have specific roles and output formats
- Prompts are templates with placeholders filled at dispatch time
- Two-stage review process: spec compliance first, then code quality
- Strong emphasis on verification (don't trust reports, read code)

## Step 6: Understanding Tool Mapping and Platform Adaptation

Examined how skills written for Claude Code are adapted for other platforms.

**What I searched for:**
- Tool mapping sections in bootstrap files
- Platform-specific adaptations
- How commands trigger skills

**What I found:**

**Tool mappings (consistent across platforms):**
- `TodoWrite` → `update_plan` (Codex/OpenCode)
- `Task` with subagents → Platform-specific (Codex: tell user, OpenCode: @mention)
- `Skill` tool → Platform-specific command/tool
- File operations → Native platform tools

**Commands:**
- `commands/brainstorm.md` - Triggers brainstorming skill
- `commands/write-plan.md` - Triggers writing-plans skill  
- `commands/execute-plan.md` - Triggers executing-plans skill
- Commands are shortcuts that say "Use skill X exactly as written"

**Platform differences:**
- Codex: Agent must run CLI commands, manual skill loading
- OpenCode: Automatic injection, custom tools, event-driven
- Claude Code: Plugin marketplace (details not visible in repo)

**What I got from it:**
- Skills are platform-agnostic, platform code handles tool mapping
- Commands provide convenient shortcuts
- Adaptation happens at bootstrap/tool level, not in skill content
- Skills reference Claude Code tools, platform adapts them

## Step 7: Examining Skill Examples and Patterns

Looked at specific skill implementations to understand patterns and structure.

**What I searched for:**
- `test-driven-development/SKILL.md` - TDD skill
- `systematic-debugging/SKILL.md` - Debugging skill
- Skill triggering test files

**What I found:**

**TDD Skill:**
- Very prescriptive: "NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST"
- Red-Green-Refactor cycle with verification steps
- Extensive rationalization table (common excuses and why they're wrong)
- Red flags section: "STOP and Start Over" conditions
- Verification checklist before marking complete

**Systematic Debugging:**
- Four-phase process: Root Cause → Pattern Analysis → Hypothesis → Implementation
- Iron Law: "NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST"
- Red flags table for common mistakes
- Emphasis on evidence gathering, especially for multi-component systems
- If 3+ fixes fail: Question architecture (not just try again)

**Common patterns across skills:**
- Strong opening statements (Iron Laws, Core Principles)
- When to use / When NOT to use sections
- Red flags / Rationalizations tables
- Verification checklists
- Integration with other skills
- Examples (good vs bad)

**What I got from it:**
- Skills are highly prescriptive with strong enforcement language
- Skills include extensive anti-patterns and rationalization handling
- Skills are designed to prevent common mistakes through explicit rules
- Skills reference each other creating a workflow system

## Summary of Key Findings

**How skills are discovered:**
- Bootstrap lists all skills with descriptions
- Agent matches task to skill description (not semantic search)
- Skills have YAML frontmatter: `name` and `description` (trigger condition)

**How agents are told to search:**
- "using-superpowers" skill: Check BEFORE ANY RESPONSE (even clarifying questions)
- Rule: "If 1% chance skill applies, MUST read it"
- Red flags table prevents rationalization
- Skill priority: Process skills first, then implementation skills

**How agents are told to use skills:**
- Skills say "Use exactly as written" or "Adapt principles"
- Skills announce their usage: "I'm using skill X to..."
- Skills reference other required skills
- Skills include checklists that must be followed

**Platform differences:**
- Codex: Manual CLI invocation, agent runs commands
- OpenCode: Automatic plugin injection, custom tools, event-driven
- Claude Code: Plugin marketplace (details not in repo)

**Agent types:**
- Code reviewer: Template-based with placeholders
- Implementer: Task-focused with self-review
- Spec reviewer: Verify matches spec (read code, don't trust report)
- Code quality reviewer: Only after spec compliance passes
