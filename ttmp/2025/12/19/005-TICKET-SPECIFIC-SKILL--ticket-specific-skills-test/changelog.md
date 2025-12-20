# Changelog

## 2025-12-19

- Initial workspace created

## 2025-12-20

- Created ticket-scoped skill fixtures and improved skill UX to show ticket id + title for ticket-scoped skills. Also ensured `skill show --ticket ...` works with the installed PATH binary, and updated `skill show` to hide skills from non-active tickets by default unless `--ticket` is provided.
- Added intern onboarding guide reference document for next-day handoff.

### Commits

- d889527 — Skills: show ticket title for ticket-scoped skills
- e6bd5a7 — Skill show: hide non-active ticket skills by default

### Docs

- `reference/02-intern-onboarding-guide.md` — Intern onboarding guide for ticket 005


## 2025-12-19

Scenarios: make run-all build nested scenariolog under go.work (GOWORK=off), then run full suite (skills smoke passes).

### Related Files

- /home/manuel/workspaces/2025-12-19/add-docmgr-skills/docmgr/scenariolog/go.sum — Update sums (via go mod tidy) so scenariolog builds cleanly
- /home/manuel/workspaces/2025-12-19/add-docmgr-skills/docmgr/test-scenarios/testing-doc-manager/run-all.sh — Force GOWORK=off for scenariolog build so suite runs under repo-level go.work


## 2025-12-19

Docs/templates: add DocType=skill guideline + embed skill template/guideline so init can scaffold skills; remove 'No guidelines found' UX for skills.

### Related Files

- /home/manuel/workspaces/2025-12-19/add-docmgr-skills/docmgr/internal/templates/embedded/_guidelines/skill.md — Embedded skill guideline seeded by docmgr init
- /home/manuel/workspaces/2025-12-19/add-docmgr-skills/docmgr/internal/templates/embedded/_templates/skill.md — Embedded skill template seeded by docmgr init
- /home/manuel/workspaces/2025-12-19/add-docmgr-skills/docmgr/ttmp/_guidelines/skill.md — New guidelines shown after doc add for skill docs

