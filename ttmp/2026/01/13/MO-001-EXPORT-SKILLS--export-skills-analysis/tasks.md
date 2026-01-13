# Tasks

## TODO

- [ ] Add tasks here

- [x] Define skill.yaml schema v1: required fields (skill.name, skill.description, what_for, when_to_use, topics), optional fields (title, license, compatibility), sources (file, binary-help), output stanza, and validation constraints (Agent Skills name/description rules, path normalization).
- [x] Implement plan discovery: scan ttmp/skills/ for workspace plans; when --ticket is provided, resolve ticket path and scan <ticket>/skills/; decide whether workspace plans are included alongside ticket plans and document the behavior.
- [x] Add YAML plan parser + validator with clear error messages (missing fields, invalid name, unsupported source type, bad output path). Include unit tests for schema validation edge cases.
- [x] Refactor docmgr skill list to operate on parsed skill.yaml plans instead of DocType skill docs; preserve existing flags (--ticket, --topics, --file, --dir) by mapping to plan metadata and explicit file sources.
- [x] Refactor docmgr skill show to load plans and render a plan summary; add optional --resolve flag to print resolved content (file bodies and captured help output) without packaging.
- [x] Implement ticket status filtering for plans: hide non-active ticket plans by default unless --ticket is provided, mirroring current skill list/show behavior.
- [x] Implement binary-help source resolver: run $binary help <topic>, capture stdout, normalize line endings, and write to references/ output path. Add timeout + error handling, and record failures in export logs.
- [x] Implement skill export command: resolve a plan into a skill directory (SKILL.md + references/), enforce Agent Skills constraints, then invoke package_skill.py to emit .skill to an output dir.
- [x] Implement skill import command: accept .skill file or directory, extract SKILL.md + references, generate a skill.yaml plan (file sources) under ttmp/skills/ or ticket skills/ when --ticket provided.
- [x] Add plan discovery tests: workspace-only, ticket-only, mixed; ensure path matching and title/slug matching behaviors are deterministic.
- [ ] Add export/import integration tests: plan -> .skill -> plan roundtrip, and binary-help capture with a stub test binary that returns deterministic help output.
- [x] Update user docs: revise using-skills.md, how-to-write-skills.md, and guidelines to explain plan-based skills and new export/import commands.
- [x] Add migration guidance: note that DocType: skill docs remain workflow docs but are no longer used by skill verbs; document how to convert a DocType skill into a plan if desired.
- [x] Add security and UX safeguards: ensure skill list/show never executes binaries unless --resolve or export is explicit; add warnings when binaries are not found on PATH.
