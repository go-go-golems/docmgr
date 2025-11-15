---
Title: Docmgr — root resolution, status API, unknown doc types, docs
Ticket: DOC
Status: active
Topics:
    - infrastructure
    - backend
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/README.md
      Note: 'Docs: quick start includes configure + seed'
    - Path: docmgr/cmd/docmgr/main.go
      Note: 'Cobra: register configure command'
    - Path: docmgr/pkg/commands/add.go
    - Path: docmgr/pkg/commands/changelog.go
    - Path: docmgr/pkg/commands/config.go
    - Path: docmgr/pkg/commands/configure.go
      Note: 'CLI: write .ttmp.yaml (configure)'
    - Path: docmgr/pkg/commands/doctor.go
    - Path: docmgr/pkg/commands/import_file.go
    - Path: docmgr/pkg/commands/init.go
      Note: 'CLI: add --seed-vocabulary flag'
    - Path: docmgr/pkg/commands/meta_update.go
    - Path: docmgr/pkg/commands/relate.go
    - Path: docmgr/pkg/commands/status.go
    - Path: docmgr/pkg/commands/tasks.go
    - Path: docmgr/pkg/commands/vocab_add.go
    - Path: docmgr/pkg/doc/docmgr-cli-guide.md
    - Path: docmgr/pkg/doc/docmgr-how-to-setup.md
      Note: 'Docs: configure/seed guidance'
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
    - Path: docmgr/test-scenarios/testing-doc-manager/02-init-ticket.sh
      Note: 'Scenario: use init --seed-vocabulary'
    - Path: docmgr/test-scenarios/testing-doc-manager/08-configure.sh
      Note: 'Scenario: test configure command'
    - Path: docmgr/test-scenarios/testing-doc-manager/09-relate-from-git.sh
    - Path: docmgr/test-scenarios/testing-doc-manager/10-status-warnings.sh
    - Path: docmgr/test-scenarios/testing-doc-manager/11-changelog-file-notes.sh
    - Path: docmgr/test-scenarios/testing-doc-manager/12-vocab-add-output.sh
    - Path: docmgr/test-scenarios/testing-doc-manager/run-all.sh
    - Path: glazed/pkg/doc/topics/how-to-write-good-documentation-pages.md
      Note: Style guide referenced for writing quality
    - Path: glazed/pkg/doc/tutorials/05-build-first-command.md
      Note: Glazed tutorial used for command patterns
    - Path: go-go-mento/ttmp/DOC-docmgr-root-resolution-status-api-unknown-doc-types-docs/design/design-tasks-verbs-task-metadata-and-listing-ux.md
    - Path: go-go-mento/ttmp/DOC-docmgr-root-resolution-status-api-unknown-doc-types-docs/playbooks/playbook-handoff-docmgr-improvements-context-and-starting-points.md
    - Path: go-go-mento/ttmp/how-to-use.md
ExternalSources:
    - local:glazed-build-first-command.md
    - local:glazed-doc-style-guide.md
Summary: Unified repo-root detection; default stale-after=30; relate --from-git documented; warnings in status; updated docs.
LastUpdated: 2025-11-04T17:56:55.02560684-05:00
---





















# Docmgr — root resolution, status API, unknown doc types, docs

## Overview

This ticket improves developer UX and predictability for docmgr in multi‑repo codebases and server/CLI workflows. It makes root resolution explicit, surfaces the active config and vocabulary in outputs, accepts unknown doc types under `various/`, and clarifies documentation (including how `.docmgrignore` removes the need for ignore flags).

Implemented so far:
- DOCMGR_CONFIG support; `.ttmp.yaml` discovery remains the primary source
- Repo detection understands `.git` as a file with `gitdir:`
- `status` shows `config_path` and `vocabulary_path` (also in human summary)
- Mutating commands (`add`, `init`, `meta update`, `vocab add`) echo `root/config/vocabulary` before writes
- Server: unknown `docType` saved under `various/`; `/api/status` endpoint
- Docs updated (how‑to, CLI guide, tutorial); new planning doc
 - Docs updated (how‑to, CLI guide, tutorial); local playbook added; new planning doc

Planned next:
- `docmgr configure` to write `.ttmp.yaml` in multi‑repo setups
- `init --seed-vocabulary` to pre‑populate common DocTypes/Topics
- `relate --from-git` to suggest/apply changed files from git ranges
- Warnings: multiple plausible `ttmp/` roots; fallback to `<cwd>/ttmp` without config
- Reconcile `stale-after` default vs documented guidance

## Key Links

- Planning document: [various/plan-docmgr-ux-and-multi-repo-improvements.md](./various/plan-docmgr-ux-and-multi-repo-improvements.md)
- Tutorial: [docmgr-multi-repo-and-server](../../docmgr/pkg/doc/tutorials/docmgr-multi-repo-and-server.md)
- How‑to: [how-to-use](../../docmgr/pkg/doc/docmgr-how-to-use.md), [how-to-setup](../../docmgr/pkg/doc/docmgr-how-to-setup.md), [cli-guide](../../docmgr/pkg/doc/docmgr-cli-guide.md)
- Playbook: [Using docmgr Efficiently (go-go-mento)](../../how-to-use.md)
- Source touchpoints: see RelatedFiles frontmatter
- External Sources: see ExternalSources frontmatter (none yet)

## Status

Current status: **active**

Recent highlights:
- Status output extended; CLI/server rebuilt
- Docs clarified: `.docmgrignore` with globs obviates `--ignore-dir/--ignore-glob`

Next steps:
- Implement `configure` command and `init --seed-vocabulary`
- Add `relate --from-git` and warnings for fallback/multiple roots
- Defer server work (startup logging, endpoints docs) for this phase

## Topics

- infrastructure
- backend

## Tasks

See [tasks.md](./tasks.md) for the current, granular checklist (implementation and docs).

## Changelog

See [changelog.md](./changelog.md) for dated notes and decisions.

## Structure

- design/ — Architecture and design documents
- reference/ — API contracts and reference docs
- playbooks/ — Operational steps and QA procedures
- scripts/ — Utility scripts and automation
- various/ — Working notes and research
- archive/ — Deprecated or reference‑only artifacts
