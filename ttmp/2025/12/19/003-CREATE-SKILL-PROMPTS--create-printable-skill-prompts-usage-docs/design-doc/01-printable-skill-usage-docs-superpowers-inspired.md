---
Title: Printable skill usage docs (Superpowers-inspired)
Ticket: 003-CREATE-SKILL-PROMPTS
Status: active
Topics:
    - skills
    - prompts
    - docs
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T17:17:59.313515848-05:00
---

# Printable skill usage docs (Superpowers-inspired)

## Executive Summary

We want a **printable set of “prompt cards” for LLMs** (similar in style to the printable docs in `docmgr/pkg/doc/`) that provide **copy/paste-ready prompts** for enforcing skill discipline in an LLM-driven coding session.

These docs will explicitly borrow the most effective Superpowers prompt techniques: **explicit skill listing (not semantic guessing), mandatory “skill check before any response,” anti-rationalization red-flags, clear tool-mapping, and proven prompt templates** (including subagent prompts and two-stage review).

Outcome: a developer new to the project can print/read a short set of documents and quickly get **LLM session prompts** for:
- Bootstrapping “skills mode” (rules, tool mapping, and discipline)
- Finding/selecting/loading the right skill
- Enforcing checklist → task tracking
- Running task execution with subagent-style roles (implementer + reviewers)

## Problem Statement

We’re adding “skills” support to docmgr (see ticket `001-ADD-CLAUDE-SKILLS`), but we currently lack **printable, copy/paste prompt packs** that teach an LLM *how to behave* in a skill-driven workflow.

Without these docs:
- Each user re-invents their own “system prompt,” so behavior becomes inconsistent.
- LLM sessions drift into “just start coding” unless the prompt actively enforces “check skills first.”
- We can’t easily audit or evolve our prompts over time (no canonical prompt library).
- We miss the strongest benefit of skills systems: enforcing disciplined process (brainstorming before building, TDD, systematic debugging, review gates).

## Proposed Solution

Add a small set of embedded help docs under `docmgr/pkg/doc/` that are intentionally **print-friendly**, **copy/paste-ready**, and structured like the best printable docs we already ship (clear navigation, short sections, and immediately usable blocks).

These documents will:
- Provide “prompt cards” developers can paste into LLM tools.
- Teach the LLM the “skills workflow” (discover → load → announce → checklist → execute), as instructions to the model (not to the human).
- Explain **tool mapping** (Superpowers-style) into docmgr verbs (e.g., checklist → `docmgr task ...`, skill lookup → `docmgr skill list/show`).
- Provide subagent-role prompts (implementer + spec reviewer + code reviewer) to reduce scope creep and enforce verification.

### Converged scope: one canonical doc (bootstrap + mapping + red flags)

We will start with **one canonical “LLM bootstrap” doc** that is basically the Superpowers `using-superpowers` skill, but rewritten to:
- instruct the model to use **docmgr commands** to list/load skills
- map checklist discipline to **docmgr tasks**
- keep the “non-negotiable” tone and phrasing as close as possible

### Where these docs live (and why)

We will place the printable docs in `docmgr/pkg/doc/` so they’re embedded into docmgr’s help system via Go embed.

- Embedding is driven by `//go:embed *` in `docmgr/pkg/doc/doc.go`.
- Help-system loading is done by `docmgr/pkg/doc/doc.go:AddDocToHelpSystem(...)`.
- This means adding markdown files to `docmgr/pkg/doc/` is enough for them to ship with docmgr.

### Relationship to Superpowers and prior analysis

These docs will be “Superpowers-inspired” but *docmgr-native*:
- We will reuse the proven phrasing patterns and prompt structure (including “red flags”).
- We will include **full file paths + exact quotes + GitHub links** when referencing Superpowers source material, so readers can verify claims and track provenance.

## Design Decisions

### 1) Multiple short prompt-pack docs, not one mega-doc

We’ll create **a small set of focused printable docs**:
- One “bootstrap / rules / mapping” prompt pack (paste into any session).
- One “task execution + review roles” prompt pack (subagent templates).
- Optionally one “authoring prompts” doc (prompts that help an LLM create/maintain skill docs).

This reduces cognitive load and makes the docs easier to print as separate “handouts.”

**Update:** based on current direction, we’ll implement **one doc first** (the canonical bootstrap prompt pack) and only split into multiple docs later if it becomes too large to be usable.

### 2) Explicit listing and matching (not semantic “search for skills”)

Superpowers’ key insight is to avoid relying on implicit semantic matching. Instead:
- provide an explicit list of skills + trigger descriptions
- force the agent to choose from that list

In docmgr terms, this lines up naturally with `docmgr skill list` + `docmgr skill show` (see `001-ADD-CLAUDE-SKILLS` design).

### 3) “Skill check before any response” as a first-class norm

We will adopt the Superpowers technique of mandating a skills check **even before clarifying questions**. This is mostly documentation + prompt discipline, not just code.

### 4) Copy/paste-ready “prompt cards”

These docs are meant to be used during real work. Each will include blocks that are ready to paste into an LLM:
- bootstrap (“you have skills; here is the list; you must check before responding”)
- mapping (“TodoWrite → docmgr task”; “Skill tool → docmgr skill show”)
- subagent prompts (implementer, spec reviewer, code reviewer)

### 5) Use existing docmgr doc conventions

We’ll mirror the “printable tutorial” patterns already used in:
- `docmgr/pkg/doc/docmgr-how-to-use.md`
- `docmgr/pkg/doc/docmgr-cli-guide.md`
- `docmgr/pkg/doc/docmgr-ci-automation.md`

## Alternatives Considered

### A) Put these docs under `ttmp/` instead of `pkg/doc/`

Rejected for this goal: ticket docs are great for project-specific context, but these “how to use skills” docs should ship with docmgr and be accessible via `docmgr help ...` even in fresh repos.

### B) Only add one doc (single “skills tutorial”)

Rejected: prompt packs become noisy and hard to find, and authoring guidance gets buried. The split docs approach keeps usage, prompts, and authoring separate.

### C) Rely on semantic search across skill bodies

Rejected: Superpowers demonstrates that explicit listing + mandatory selection is more reliable for LLMs and easier to audit.

## Implementation Plan

### 1) Create embedded printable docs (new files)

- [ ] **Create** `docmgr/pkg/doc/using-skills.md`
  - **Goal**: one doc that teaches the model how to use both docmgr + skills in one go
  - **Model**: as close as possible to Superpowers `skills/using-superpowers/SKILL.md`, but:
    - replaces “Skill tool” with `docmgr skill list` / `docmgr skill show`
    - includes docmgr tool mapping (TodoWrite → docmgr tasks)
    - includes the same red-flags / rationalization table
    - includes upstream provenance links to Superpowers source files

### 2) Ensure docs appear in help output

- [ ] Verify `docmgr help` lists the new docs (embedding is automatic via `//go:embed *` in `docmgr/pkg/doc/doc.go`)
- [ ] Optionally update the comment list in `docmgr/pkg/doc/doc.go` to mention the new docs (non-functional, but helps maintainers)

### 3) Add cross-links between docs

- [ ] Add “See also” sections to link:
  - `docmgr-how-to-use` ↔ `docmgr-llm-skills-bootstrap-prompt-pack`
  - `templates-and-guidelines` ↔ `docmgr-llm-skill-authoring-prompts` (if we create it)
  - `docmgr-diagnostics-and-rules` ↔ prompts troubleshooting (what to do if docmgr outputs warnings/errors)

### 4) Capture Superpowers provenance in our docs

- [ ] For each key technique we adopt (e.g., “check for skills before any response”), include:
  - Superpowers file path
  - exact quote blocks
  - upstream link (GitHub)
  - brief “what we changed” note (docmgr adaptation)

## Open Questions

1. Should `docmgr-skills-prompt-pack.md` be `ShowPerDefault: true`, or should it be “opt-in” to avoid overwhelming new users?
2. Should we include one “platform integration” appendix (Claude Code vs Codex vs OpenCode) now, or keep it docmgr-centric until docmgr’s skills UX is fully implemented?
3. Do we want a dedicated `SectionType` for prompt packs / printable cards, or keep using `GeneralTopic`?

## References

- `docmgr/pkg/doc/docmgr-how-to-use.md` (style reference)
- `docmgr/pkg/doc/docmgr-cli-guide.md` (style reference)
- Ticket `001-ADD-CLAUDE-SKILLS`: `docmgr/ttmp/2025/12/19/001-ADD-CLAUDE-SKILLS--add-claude-skills-support/design-doc/01-skills-implementation-plan.md`
- Ticket `002-ANALYZE-SUPERPOWERS`: `docmgr/ttmp/2025/12/19/002-ANALYZE-SUPERPOWERS--analyze-superpowers-repository/analysis/01-superpowers-repository-deep-analysis.md`

### Superpowers source references (upstream)

These are the **exact upstream files** we should treat as “source of truth” when drafting our prompt packs:

- **Repo root**: [`obra/superpowers`](https://github.com/obra/superpowers)

- **Core “skills discipline” prompt** (mandatory check-before-any-response, red flags table):
  - [`skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md)

- **Codex bootstrap instructions** (explicit “tool mapping” + mandatory workflows):
  - [`./.codex/superpowers-bootstrap.md`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-bootstrap.md)
  - (CLI that prints bootstrap + lists skills): [`./.codex/superpowers-codex`](https://github.com/obra/superpowers/blob/main/.codex/superpowers-codex)

- **OpenCode plugin injection + tool mapping** (how bootstrap is inserted; compaction re-injection):
  - [`./.opencode/plugin/superpowers.js`](https://github.com/obra/superpowers/blob/main/.opencode/plugin/superpowers.js)

- **Subagent-driven development prompt templates** (implementer + spec reviewer + quality reviewer):
  - [`skills/subagent-driven-development/implementer-prompt.md`](https://github.com/obra/superpowers/blob/main/skills/subagent-driven-development/implementer-prompt.md)
  - [`skills/subagent-driven-development/spec-reviewer-prompt.md`](https://github.com/obra/superpowers/blob/main/skills/subagent-driven-development/spec-reviewer-prompt.md)
  - [`skills/subagent-driven-development/code-quality-reviewer-prompt.md`](https://github.com/obra/superpowers/blob/main/skills/subagent-driven-development/code-quality-reviewer-prompt.md)

- **Code review agent + prompt template** (severity buckets, diff-range workflow):
  - Agent definition: [`agents/code-reviewer.md`](https://github.com/obra/superpowers/blob/main/agents/code-reviewer.md)
  - Template referenced by workflows: [`skills/requesting-code-review/code-reviewer.md`](https://github.com/obra/superpowers/blob/main/skills/requesting-code-review/code-reviewer.md)

- **Shared skill discovery + parsing logic** (frontmatter parsing, “find skills” recursion, shadowing rules):
  - [`lib/skills-core.js`](https://github.com/obra/superpowers/blob/main/lib/skills-core.js)
