# Guidelines: Skill Plans

## Purpose
Skills are packaged workflows defined by a `skill.yaml` plan and an exported `SKILL.md` body. The plan tells docmgr what to package (files, help output), while the body teaches LLMs (and humans) *how to work*, not just what exists.

Skills are meant to be discoverable via:
- `docmgr skill list` (filter by topics, file/dir, ticket)
- `docmgr skill show <query>` (load and apply a skill plan)

## Required Elements
- **skill.yaml metadata**
  - `skill.name`: lowercase hyphenated name
  - `skill.description`: short trigger description
  - `skill.what_for`: outcome this workflow ensures
  - `skill.when_to_use`: trigger conditions ("Use when ...")
  - `skill.topics`: topics for discovery
- **skill.yaml sources**
  - `sources`: explicit `file` and/or `binary-help` entries
  - `sources[].append_to_body`: append source content into the main SKILL.md body (output file is not written; auto sections are suppressed when present; title is skipped if the content already starts with one)
  - `output`: optional export options (`skill_dir_name`, `skill_md`)
- **SKILL.md body** (exported)
  - **Overview**: 2–5 sentences (why it matters; what it enforces)
  - **When to Use**: Concrete triggers + examples
  - **Process**: Step-by-step actions (copy/paste commands when possible)
  - **Verification**: Checklist to prevent “I think I’m done” drift

## Recommended SKILL.md Structure
Use a structure that’s easy for both humans and LLMs to follow:

- `## Overview`
- `## When to Use`
- `## The Iron Law` (the non-negotiable rule)
- `## The Process` (steps)
- `## Red Flags` (common rationalizations + counters)
- `## Verification Checklist`
- `## Integration` (other skills that pair well)
- `## Examples` (good vs bad)

## Best Practices
- **Be explicit**: Use strong modal language (“MUST”, “NEVER”, “STOP”) where appropriate.
- **Optimize for discovery**: Ensure `skill.topics`, `skill.what_for`, and `skill.when_to_use` contain the words people will actually search for.
- **Prefer commands over prose**: If there’s a canonical command sequence, include it as a `bash` block.
- **Make validation unskippable**: Add a checklist and expected success criteria (tests/linters/scenario suites).
- **Keep it small and reusable**: If a skill becomes too broad, split it into multiple skills and cross-link.

## Notes
- DocType skill documents still exist as workflow docs, but `docmgr skill list/show` use `skill.yaml` plans.

## References
- `pkg/doc/how-to-write-skills.md` — Full guidance on writing and enforcing skills (recommended reading)
- `pkg/doc/using-skills.md` — How to discover and load skills via CLI
