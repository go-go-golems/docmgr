---
Title: 'Playbook — Handoff: Docmgr improvements — context and starting points'
Ticket: DOC
Status: active
Topics:
    - infrastructure
    - backend
DocType: playbook
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/pkg/commands/add.go
    - Path: docmgr/pkg/commands/config.go
      Note: Root/config resolution and DOCMGR_CONFIG
    - Path: docmgr/pkg/commands/init.go
    - Path: docmgr/pkg/commands/meta_update.go
    - Path: docmgr/pkg/commands/status.go
    - Path: docmgr/pkg/commands/vocab_add.go
    - Path: go-go-mento/ttmp/how-to-use.md
      Note: Workspace playbook for efficient usage
ExternalSources: []
Summary: Unified repo-root detection; default stale-after=30; relate --from-git documented; warnings in status; updated docs.
LastUpdated: 2025-11-04T17:56:55.025825377-05:00
---







# Playbook — Handoff: Docmgr improvements — context and starting points

## Who this is for
New developer joining the DOC ticket in a fresh session. Use `docmgr` to navigate, update tasks, and continue implementation.

## TL;DR
- Root is resolved via `.ttmp.yaml` at workspace root → `go-go-mento/ttmp`.
- Mutating commands echo: `root/config/vocabulary` pre‑write.
- `status` shows `config_path` and `vocabulary_path` in summary.
- Unknown `docType` is accepted and stored under `various/`.
- `.docmgrignore` with globs removes the need for `--ignore-dir`.

## Environment Assumptions
- You are at the workspace root (where `.ttmp.yaml` lives).
- `docmgr` is on PATH; `docmgr status` works.

## Start here (commands)
```bash
# Confirm resolution
docmgr status --summary-only

# Explore docs and tasks
docmgr list docs   --ticket DOC
docmgr tasks list  --ticket DOC

# Open key docs (from the list output): index, plan, design, playbook
```

## What’s implemented
- CLI: context echo on `add`, `init`, `meta update`, `vocab add`.
- CLI: `status` prints config/vocabulary paths (human + structured).
- Repo detection supports `.git` file with `gitdir:`.
- Docs: multi‑repo + `.docmgrignore` guidance; workspace playbook.

## What’s next
- `docmgr configure` to write `.ttmp.yaml`
- `init --seed-vocabulary` flag
- `relate --from-git` to suggest/apply changed files
- Warnings for multiple plausible roots and fallback to `<cwd>/ttmp`
- Tasks verbs improvements (attach owners/links/related/notes; listing UX)

See `./../tasks.md` for the authoritative, granular checklist (each task lists related files).

## Key files to skim
- CLI commands: `docmgr/pkg/commands/*.go` (especially `config.go`, `status.go`, `add.go`, `init.go`, `meta_update.go`, `vocab_add.go`, `tasks.go`)
- Docs: `docmgr/pkg/doc/*.md`, `go-go-mento/ttmp/how-to-use.md`

## Useful links (docs in this ticket)
- Index: `./../index.md`
- Plan: `./../various/plan-docmgr-ux-and-multi-repo-improvements.md`
- Design — Tasks UX: `./../design/design-tasks-verbs-task-metadata-and-listing-ux.md`
- Playbook (workspace): `./../../how-to-use.md`
- Tasks: `./../tasks.md`
- Changelog: `./../changelog.md`

## Daily flow
1) `docmgr status` → confirm root/config/vocabulary
2) `docmgr tasks list --ticket DOC` → pick next task
3) Implement; keep docs accurate via `docmgr meta update`/`docmgr relate`
4) `docmgr changelog update` with small, frequent entries
5) `docmgr doctor --stale-after 30` before pushing

## Notes
- If status shows `root=.../ttmp (fallback)`, create `.ttmp.yaml` or export `DOCMGR_ROOT`.
- Prefer `docmgr relate` to add RelatedFiles with rationale.
