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
  - Path: test-scenarios/testing-doc-manager/run-all.sh
    Note: Scenario suite runner; patched to build nested `scenariolog` module correctly when a repo-level go.work is present
  - Path: scenariolog/go.mod
    Note: Nested scenariolog module; `go mod tidy` may update this when fixing missing dependency sums
  - Path: scenariolog/go.sum
    Note: Nested scenariolog module dependency sums; updated so `go build` works in module mode (GOWORK=off)
  - Path: ttmp/_guidelines/skill.md
    Note: Skill-writing guidelines shown after `docmgr doc add --doc-type skill` (fixes “No guidelines found” UX)
  - Path: internal/templates/embedded/_guidelines/skill.md
    Note: Embedded skill guideline scaffolded by `docmgr init` for fresh docs roots
  - Path: internal/templates/embedded/_templates/skill.md
    Note: Embedded skill template scaffolded by `docmgr init` for fresh docs roots
  - Path: pkg/doc/using-skills.md
    Note: User-facing prompt pack docs; updated to document `--ticket` and active-ticket filtering behavior
  - Path: pkg/doc/how-to-write-skills.md
    Note: Authoring docs; updated to clarify ticket-level `skill/` vs workspace-level `skills/` convention
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

## Step 5: Make the scenario suite runnable under go.work, then run it (skills smoke included)

This step validated the ticket end-to-end by running the full `testing-doc-manager` scenario suite, including the skills smoke test. Along the way, we fixed a real-world “new contributor” footgun: with a repo-level `go.work` present, the suite’s auto-build of `scenariolog` (a nested Go module) failed unless we force module mode.

**Commit (code):** 53df7f0 — "Scenarios: build scenariolog under go.work"

**Commit (docs):** 1cb530a — "Ticket 005: record scenario run; check off smoke"

Note: I initially tried to commit from the wrong directory and thought this workspace wasn’t a git repo. It’s actually a git worktree (see `docmgr/.git`), so the work is committed above.

### What I did
- Ran the suite as documented (pinned docmgr binary):
  - `DOCMGR_PATH=/home/manuel/.local/bin/docmgr bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario`
- Hit a build failure when `run-all.sh` tried to build `scenariolog`:
  - `main module (github.com/go-go-golems/docmgr) does not contain package github.com/go-go-golems/docmgr/scenariolog/cmd/scenariolog`
- Confirmed we were in workspace mode due to a repo-level `go.work`:
  - `go env GOWORK` → `/home/manuel/workspaces/2025-12-19/add-docmgr-skills/go.work`
- Patched `test-scenarios/testing-doc-manager/run-all.sh` to build scenariolog with `GOWORK=off`.
- Re-ran the suite and hit missing dependency sums inside the nested scenariolog module:
  - `missing go.sum entry for module providing package github.com/adrg/frontmatter ...`
  - (plus several similar `missing go.sum entry` errors)
- Fixed the nested module by running:
  - `cd scenariolog && GOWORK=off go mod tidy`
- Re-ran the suite again and confirmed it completed successfully:
  - `[ok] Scenario completed at /tmp/docmgr-scenario/acme-chat-app`
- Checked off docmgr task [6] (“Run smoke”) after it passed.

### Why
- “Run the suite” is an onboarding-critical validation step for ticket 005; it catches regressions in `skill list/show` behavior and verifies the default active-ticket filtering semantics in a realistic flow.
- Contributors commonly use a repo-level `go.work`; the suite should remain runnable without requiring hidden environment knowledge.

### What worked
- The scenario suite (including step 20 “skills smoke”) passes cleanly after the fix.
- Forcing `GOWORK=off` makes `scenariolog` build reliably as its own module, independent of the repo-level workspace module set.

### What didn't work
- Initial run failed building scenariolog due to workspace mode resolving packages against the wrong module:
  - `main module (github.com/go-go-golems/docmgr) does not contain package github.com/go-go-golems/docmgr/scenariolog/cmd/scenariolog`
- Second run failed due to incomplete `scenariolog/go.sum`:
  - `missing go.sum entry for module providing package ...`

### What I learned
- A repo-level `go.work` can break nested-module builds even if you `go -C` into the nested module.
- For “self-contained nested tool builds” inside scripts, `GOWORK=off` is a pragmatic way to ensure the nested module uses its own `go.mod/go.sum`.

### What was tricky to build
- The interaction between:
  - workspace mode (`go.work`),
  - nested module boundaries (`scenariolog/go.mod`), and
  - “build a tool automatically from a script”.
  `go -C` alone isn’t sufficient; the go.work still influences module selection.

### What warrants a second pair of eyes
- Confirm that forcing `GOWORK=off` in `run-all.sh` is the intended long-term policy (vs. adding `docmgr/scenariolog` to the top-level `go.work`).
- Sanity-check that `scenariolog/go.mod/go.sum` changes from `go mod tidy` are acceptable (no accidental dep churn).

### What should be done in the future
- Consider updating `test-scenarios/testing-doc-manager/README.md` to mention go.work/workspace mode and why `run-all.sh` forces `GOWORK=off`.
- If the repo standardizes on go.work, consider adding the scenariolog module to it (if that’s desirable), so `go test ./...` workflows can include it intentionally.

### Code review instructions
- Start in:
  - `test-scenarios/testing-doc-manager/run-all.sh` (scenariolog build invocation)
  - `scenariolog/go.mod` + `scenariolog/go.sum` (tidy effects)
- Validate with:
  - `cd docmgr && DOCMGR_PATH=/home/manuel/.local/bin/docmgr bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario`

### Technical details
- Key snippet: build `scenariolog` as a nested module even when `go.work` is present:
  - `GOWORK=off go -C "${REPO_ROOT}/scenariolog" build -tags sqlite_fts5 -o "${SCENARIOLOG_PATH}" ./cmd/scenariolog`

## Step 6: Add DocType=skill guidelines + embed skill template/guidelines for `docmgr init`

This step fixes the “No guidelines found for doc-type skill” UX when creating skills via `docmgr doc add --doc-type skill`. It also adds an embedded skill template + guideline so `docmgr init` can scaffold them into fresh docs roots (new workspaces shouldn’t have to hand-create these files).

**Commit (docs+templates):** 24a0813 — "Skills: add guidelines + init scaffolding template"

### What I did
- Added a filesystem guideline at:
  - `ttmp/_guidelines/skill.md`
- Added embedded scaffolding files so `docmgr init` can seed new roots:
  - `internal/templates/embedded/_guidelines/skill.md`
  - `internal/templates/embedded/_templates/skill.md`
- Verified the guideline is discoverable in the current docs root:
  - `docmgr doc guidelines --doc-type skill --root ttmp`

### Why
- Skills are first-class docs (`DocType: skill`), so `doc add` should provide immediate guidance instead of printing a warning.
- `docmgr init` should scaffold skill docs the same way it scaffolds reference/design-doc/etc.

### What worked
- `docmgr doc guidelines --doc-type skill --root ttmp` prints the new guidelines text (no “no guideline found” error).

### What warrants a second pair of eyes
- Confirm the intended policy: runtime template/guideline resolution is filesystem-only, while embedded files are for `init` scaffolding. (Today `doc add` won’t fall back to embedded templates/guidelines unless they exist on disk.)

### Code review instructions
- Start in:
  - `ttmp/_guidelines/skill.md`
  - `internal/templates/embedded/_guidelines/skill.md`
  - `internal/templates/embedded/_templates/skill.md`
- Validate quickly with:
  - `cd docmgr && docmgr doc guidelines --doc-type skill --root ttmp`

## Step 7: Update skills docs for `--ticket`, ticket context, active-ticket filtering, and `skill/` convention

This step updates the user-facing documentation to reflect the current skills UX: `skill show` supports `--ticket` narrowing, ticket-scoped skills show ticket id + title, and `skill show` defaults to hiding skills from non-active tickets unless `--ticket` is provided. It also clarifies the directory convention difference between workspace-level skills (`ttmp/skills/`) and ticket-level skills created by `doc add` (`.../skill/`).

**Commit (docs):** fa2cf94 — "Docs: clarify skills usage and conventions"

### What I did
- Updated docs:
  - `pkg/doc/using-skills.md` (documented `--ticket` and default filtering behavior; added mention of ticket context in output)
  - `pkg/doc/how-to-write-skills.md` (clarified ticket-level `skill/` vs workspace-level `skills/` convention; adjusted examples)
- Checked off ticket tasks:
  - [7] Update docs
  - [9] Decide/confirm convention (documented as `skill/` for ticket-level)

### Why
- The UX improvements from ticket 004/005 are user-facing behavior changes; docs must match what the CLI does.
- The repo currently has both `ttmp/skills/` (workspace library) and per-ticket `skill/` folders; the docs should explain that intentional split instead of implying everything is `skills/`.

### Code review instructions
- Review the diff in:
  - `pkg/doc/using-skills.md`
  - `pkg/doc/how-to-write-skills.md`


