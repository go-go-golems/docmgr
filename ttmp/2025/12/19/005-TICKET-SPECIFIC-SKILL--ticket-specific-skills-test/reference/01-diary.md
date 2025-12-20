---
Title: Diary
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
  - Path: pkg/commands/skill_list.go
    Note: Prints ticket id+title for ticket-scoped skills (and emits ticket_title in structured output)
  - Path: pkg/commands/skill_show.go
    Note: Prints ticket id+title, supports --ticket scoping, and filters non-active ticket skills by default
  - Path: test-scenarios/testing-doc-manager/20-skills-smoke.sh
    Note: Scenario coverage for --ticket, ambiguity, and active-ticket filtering
  - Path: ttmp/2025/12/19/005-TICKET-SPECIFIC-SKILL--ticket-specific-skills-test/skill/01-systematic-debugging.md
    Note: Ticket-scoped skill fixture (generated via doc add)
  - Path: ttmp/2025/12/19/005-TICKET-SPECIFIC-SKILL--ticket-specific-skills-test/skill/02-test-driven-development.md
    Note: Ticket-scoped skill fixture (generated via doc add)
  - Path: ttmp/2025/12/19/005-TICKET-SPECIFIC-SKILL--ticket-specific-skills-test/skill/03-documenting-as-you-code.md
    Note: Ticket-scoped skill fixture (generated via doc add)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-20T00:00:00Z
---

# Diary

## Goal

Create ticket-scoped skill fixtures and validate the UX around ticket scoping: skills should clearly show which ticket they belong to (including ticket title), `skill show` should support `--ticket` to narrow the search, and by default `skill show` should not surface skills from completed tickets (unless `--ticket` is provided).

## Step 1: Create ticket 005 + generate ticket-scoped skill fixtures

This step created a dedicated ticket to hold test “ticket-scoped” skills and generated a few skill documents inside it, so we can validate both discovery (`skill list`) and loading (`skill show`) behavior.

### What I did
- Created ticket workspace:
  - `docmgr ticket create-ticket --ticket 005-TICKET-SPECIFIC-SKILL --title "Ticket-specific skills test" --topics skills,cli,ux`
- Created three ticket-scoped skill docs via doc verb:
  - `docmgr doc add --ticket 005-TICKET-SPECIFIC-SKILL --doc-type skill --title "Systematic Debugging"`
  - `docmgr doc add --ticket 005-TICKET-SPECIFIC-SKILL --doc-type skill --title "Test-Driven Development"`
  - `docmgr doc add --ticket 005-TICKET-SPECIFIC-SKILL --doc-type skill --title "Documenting as You Code"`

### Why
- We need concrete fixtures to validate that ticket-scoped skills are discoverable and that the UX around ticket identity is clear.

### What worked
- `skill list --ticket 005-TICKET-SPECIFIC-SKILL` returns exactly the three skills.
- The generated docs live under the ticket directory (not `ttmp/skills/`), i.e. they’re ticket-scoped.

### What warrants a second pair of eyes
- The repo currently uses `doc add --doc-type skill` to create ticket skills under `.../skill/` (singular), not `.../skills/`. Confirm this is intended naming.

### Code review instructions
- Start in `ttmp/.../005-.../skill/` to see the generated docs and their frontmatter.

## Step 2: Print ticket id + ticket title for ticket-scoped skills

This step improved the UX: when a skill is ticket-scoped, we print the ticket id and the ticket title in both list and show outputs.

**Commit (code):** d889527 — "Skills: show ticket title for ticket-scoped skills"

### What I did
- Updated `skill list` human output to include:
  - `Ticket: <ID> — <Ticket Title>`
- Updated structured output to emit:
  - `ticket`, `ticket_title`
- Updated `skill show` to print ticket id + ticket title.

### Why
- Ticket-scoped skills can otherwise be confusing (same skill names can exist across tickets).
- Ticket title is an important human cue for which context the skill belongs to.

### What was tricky to build
- Resolving ticket title requires a ticket index query (the title is stored in the ticket index doc).

## Step 3: Finish the `--ticket` usability “in practice” (PATH vs local build)

This step resolved the user-facing confusion where `docmgr` on PATH did not support `--ticket` because it was an older binary.

### What I did
- Verified the binary mismatch:
  - PATH `docmgr` was an older install (missing `--ticket`)
  - `/tmp/docmgr-local` and the repo build supported `--ticket`
- Installed the updated build to `/home/manuel/.local/bin/docmgr` so normal `docmgr ...` invocations have the new flags.

### Why
- If PATH points at an older binary, UX improvements look “broken” even if code is correct.

### What worked
- After install, `docmgr skill show --ticket "005-TICKET-SPECIFIC-SKILL" documenting-as-you-code` works as expected.

## Step 4: Filter out skills from non-active tickets by default (unless `--ticket` is provided)

This step changed default behavior in `skill show`: if you don’t specify `--ticket`, we hide skills that belong to non-active tickets (e.g. `complete`). This does NOT apply when `--ticket` is set (because then you’re explicitly scoping).

**Commit (code):** e6bd5a7 — "Skill show: hide non-active ticket skills by default"

### What I did
- Updated matching candidate selection in `skill show`:
  - If `--ticket` is empty, filter out skills whose `Ticket` points to a ticket whose status is not `active`.
  - Workspace-level skills (no ticket) remain visible.
- Extended `20-skills-smoke.sh`:
  - Create a skill under a ticket, close that ticket, then assert:
    - without `--ticket` → not found
    - with `--ticket` → found

### Why
- Old/completed tickets shouldn’t pollute the default “find me a skill” experience.
- Explicit `--ticket` should still work for archaeology/debugging.

### What warrants a second pair of eyes
- The exact definition of “active” (we currently treat anything not equal to `active` as excluded). Confirm whether `review` should also count as “active enough”.


