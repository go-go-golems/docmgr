---
Title: Intern onboarding guide
Ticket: 005-TICKET-SPECIFIC-SKILL
Status: active
Topics:
    - skills
    - cli
    - ux
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
  - Path: ttmp/2025/12/19/005-TICKET-SPECIFIC-SKILL--ticket-specific-skills-test/reference/01-diary.md
    Note: Step-by-step diary with commands + commits for ticket 005
  - Path: ttmp/2025/12/19/004-BETTER-SKILL-SHOWING--improve-skill-show-ux-for-intuitive-skill-name-matching/design/01-skill-show-and-list-ux-analysis.md
    Note: Architecture + reasoning behind skill show/list implementation choices
  - Path: ttmp/2025/12/19/004-BETTER-SKILL-SHOWING--improve-skill-show-ux-for-intuitive-skill-name-matching/reference/01-diary.md
    Note: Ticket 004 diary (includes post-close follow-ups references)
  - Path: pkg/commands/skill_show.go
    Note: `docmgr skill show` matching + ticket filter + active-ticket filtering
  - Path: pkg/commands/skill_list.go
    Note: `docmgr skill list` output (ticket display + load command generation)
  - Path: pkg/commands/skills_helpers.go
    Note: Load-command selection logic (slug → title → path)
  - Path: test-scenarios/testing-doc-manager/20-skills-smoke.sh
    Note: End-to-end test coverage for skill UX and ticket filtering
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T20:07:17.346792155-05:00
WhatFor: ""
WhenToUse: ""
---

# Intern onboarding guide

## Goal

Give a new intern enough context and an actionable checklist to safely continue improving the **skills UX** in docmgr (ticket-scoped skills, ticket context in outputs, `--ticket` narrowing, and default filtering of skills from non-active tickets).

## Context

We just shipped a set of UX improvements around `docmgr skill list` and `docmgr skill show`:

- `skill show` supports positional queries and multiple matching strategies (slug/title/path).
- `skill list` prints copy/pasteable “Load:” commands.
- Ticket-scoped skills display **ticket id + ticket title**.
- `skill show` (when no `--ticket`) hides skills belonging to non-active tickets, while `--ticket` overrides that.

There are two relevant tickets:

- **004-BETTER-SKILL-SHOWING**: the original UX improvement work (now complete).
- **005-TICKET-SPECIFIC-SKILL**: fixture ticket + follow-on improvements and “intern landing pad”.

## Quick Reference

### The “what to read first” list (in order)

- **Ticket 005 diary (this is the shortest path to context):**
  - `ttmp/2025/12/19/005-TICKET-SPECIFIC-SKILL--ticket-specific-skills-test/reference/01-diary.md`
- **Ticket 004 analysis + diary (architecture + why):**
  - `ttmp/2025/12/19/004-BETTER-SKILL-SHOWING--improve-skill-show-ux-for-intuitive-skill-name-matching/design/01-skill-show-and-list-ux-analysis.md`
  - `ttmp/2025/12/19/004-BETTER-SKILL-SHOWING--improve-skill-show-ux-for-intuitive-skill-name-matching/reference/01-diary.md`
- **Key code:**
  - `pkg/commands/skill_show.go`
  - `pkg/commands/skill_list.go`
  - `pkg/commands/skills_helpers.go`
- **Key scenario test:**
  - `test-scenarios/testing-doc-manager/20-skills-smoke.sh`

### The “how to validate” commands

Run these from `docmgr/`:

```bash
# Build once (optional; `docmgr` is installed on PATH too)
go build -o /tmp/docmgr-local ./cmd/docmgr

# Full scenario suite (includes skills smoke)
DOCMGR_PATH="$(command -v docmgr)" bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario

# Quick checks (local repo)
docmgr skill list
docmgr skill show systematic-debugging
docmgr skill show --ticket "005-TICKET-SPECIFIC-SKILL" documenting-as-you-code
```

### Intern TODO checklist (copy/paste)

These tasks are also tracked in `docmgr task list --ticket 005-TICKET-SPECIFIC-SKILL`.

- [ ] Read ticket 005 diary + ticket 004 analysis
- [ ] Run scenario suite and ensure skills smoke passes
- [ ] Update docs to reflect:
  - positional `skill show`
  - `skill show --ticket` narrowing
  - ticket id+title printing
  - “active tickets only by default” behavior
- [ ] Add guideline/template for `DocType=skill` (so `doc add --doc-type skill` has good scaffolding)
- [ ] Decide/confirm directory convention for ticket-scoped skills: `skill/` vs `skills/`
- [ ] Review “active” semantics (should `review` count as active?)
- [ ] Optional perf cleanup: avoid repeated ticket index queries inside list/show

### Where to store future research/notes

- Store new research notes under `ttmp/YYYY-MM-DD/...` and keep using a numbered diary format for multi-step work.

## Usage Examples

### Example: Narrow a show command to a ticket (disambiguation)

```bash
docmgr skill show --ticket "005-TICKET-SPECIFIC-SKILL" documenting-as-you-code
```

### Example: Find the right skill when names clash

1) List skills (copy/paste the `Load:` line):

```bash
docmgr skill list
```

2) Use the generated `Load:` command (it will prefer slug → title → path).

## Related

- Ticket 005 diary: `ttmp/2025/12/19/005-TICKET-SPECIFIC-SKILL--ticket-specific-skills-test/reference/01-diary.md`
- Ticket 004 analysis: `ttmp/2025/12/19/004-BETTER-SKILL-SHOWING--improve-skill-show-ux-for-intuitive-skill-name-matching/design/01-skill-show-and-list-ux-analysis.md`
