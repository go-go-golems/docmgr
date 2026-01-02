---
Title: 'Codex skills: test playbook'
Ticket: 006-SKILL-CREATE
Status: active
Topics:
    - skills
    - cli
    - ux
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: .codex/skills/cursor-debate/SKILL.md
      Note: Codex skill implementing the debate workflow
    - Path: .codex/skills/cursor-diary/SKILL.md
      Note: Codex skill implementing the diary workflow
    - Path: .codex/skills/cursor-docmgr/SKILL.md
      Note: Codex skill for docmgr command workflows
    - Path: .codex/skills/cursor-git-commit/SKILL.md
      Note: Codex skill for git commit hygiene workflow
    - Path: .codex/skills/cursor-remarkable-upload/SKILL.md
      Note: Codex skill for uploading markdown to reMarkable
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-02T09:55:26.275425211-05:00
WhatFor: ""
WhenToUse: ""
---


# Codex skills: test playbook

## Purpose

Verify that the repo-scoped Codex skills in `.codex/skills/` load correctly and trigger (explicitly and implicitly) with the expected behavior.

## Environment Assumptions

- You have Codex CLI installed and authenticated (`codex` command works).
- You run Codex from within this repository (so `.codex/skills/` is in scope).
- You are testing the skills created from `~/.cursor/commands/*`:
  - `cursor-debate`
  - `cursor-diary`
  - `cursor-docmgr`
  - `cursor-git-commit`
  - `cursor-remarkable-upload`

## Commands

### 0) Verify skill files exist in this repo

```bash
ls -la .codex/skills
find .codex/skills -maxdepth 3 -type f -name 'SKILL.md' -print
```

### 1) Validate skill format (offline)

Use the built-in validator shipped with the system `skill-creator` skill:

```bash
python ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/cursor-debate
python ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/cursor-diary
python ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/cursor-docmgr
python ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/cursor-git-commit
python ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/cursor-remarkable-upload
```

### 2) Start Codex and confirm skills load

```bash
codex
```

In the interactive session:
- Run `/skills` (or start typing `$`) and confirm the five `cursor-*` skills appear.
- If they do not appear: quit and restart Codex, and ensure the file is exactly named `SKILL.md` (no symlinks).

### 3) Explicit invocation tests (deterministic)

Run each prompt and confirm Codex follows the referenced workflow/template:

1) `cursor-diary`
   - Prompt: `Use $cursor-diary and keep a diary for this session. Start with Step 1.`
   - Expected: diary step structure; commands + failures captured; docmgr loop when applicable.

2) `cursor-debate`
   - Prompt: `Use $cursor-debate to explore two approaches to implementing a new CLI command in this repo.`
   - Expected: research-first section, opening statements, rebuttals, moderator summary.

3) `cursor-docmgr`
   - Prompt: `Use $cursor-docmgr. I need commands to create a ticket, add a design doc, relate files, and update changelog.`
   - Expected: `docmgr ticket create-ticket`, `docmgr doc add`, `docmgr doc relate --file-note ...`, `docmgr changelog update ...`.

4) `cursor-git-commit`
   - Prompt: `Use $cursor-git-commit. I accidentally staged dist/ and .env; how do I fix it before committing?`
   - Expected: `git reset HEAD ...`, `git rm --cached ...` guidance, and `.gitignore`/amend instructions.

5) `cursor-remarkable-upload`
   - Prompt: `Use $cursor-remarkable-upload. Dry-run upload of /abs/path/to/doc.md to reMarkable, mirroring ticket structure.`
   - Expected: dry-run first, avoid overwrite, mention prerequisites (`rmapi`, `pandoc`, `xelatex`).

### 4) Implicit invocation tests (trigger quality)

Ask without `$...` and confirm Codex chooses the correct skill:

- `Please keep a diary of what you're doing, step by step, and include commands.`
- `Let's debate the trade-offs between two designs before we implement anything.`
- `How do I relate code files to a doc in docmgr?`
- `I'm about to commit; what should I check first?`
- `Can you upload this markdown doc to my reMarkable as a PDF?`

If the wrong skill triggers:
- Narrow the skill’s `description` to be more specific about when it should/shouldn’t trigger.
- Add discriminators (tool names, file types, or explicit phrases) to the `description` line (must stay single-line and <=500 chars).

```bash
# (Intentionally left blank; see command blocks above.)
```

## Exit Criteria

- `/skills` shows all five `cursor-*` skills.
- Each skill can be invoked explicitly and produces the expected structure/output.
- Implicit prompts select the intended skill most of the time (and misfires are documented with a proposed description tweak).

## Notes

- Skill locations + precedence (high → low) include: `$CWD/.codex/skills`, parent `.codex/skills`, repo root `.codex/skills`, then `~/.codex/skills`.
- Codex ignores symlinked skill directories and rejects malformed YAML or multi-line / over-length `name`/`description`.
