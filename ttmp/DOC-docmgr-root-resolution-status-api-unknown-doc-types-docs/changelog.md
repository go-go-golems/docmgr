---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Unified repo-root detection; default stale-after=30; relate --from-git documented; warnings in status; updated docs.
LastUpdated: 2025-11-04T17:56:55.025154732-05:00
---



# Changelog

## 2025-11-04

- Initial workspace created


## 2025-11-04

Server: add /api/status; root resolution via env + .ttmp.yaml; unknown doc types -> various/; scans include various/; improved logging. Docs updated.


## 2025-11-04

Docs: clarify that .docmgrignore (with globs) removes the need for --ignore-dir/--ignore-glob; updated how-to-use, how-to-setup, and CLI guide examples accordingly.


## 2025-11-04

Vocabulary: add docType=planning; Planning doc reclassified to planning; Added implementation tasks and related files.


## 2025-11-04

Impl: status shows config/vocabulary; DOCMGR_CONFIG supported; .git gitdir supported; echo context on mutating commands.


## 2025-11-04

Plan: added surfaced issues as tasks; related updated docs to ticket; updated plan summary.


## 2025-11-04

Planning doc: related code and documentation files with notes.


## 2025-11-04

Planning doc: related code and documentation files with notes.


## 2025-11-04

Tasks: appended Related file notes per task in tasks.md.


## 2025-11-04

Design: tasks verbs + task metadata model + listing UX; added crosslinking strategy with changelog.


## 2025-11-04

Playbook: added go-go-mento/ttmp/how-to-use.md documenting efficient docmgr usage.


## 2025-11-04

Handoff playbook added and linked from index; related key files added with notes.


## 2025-11-04

CLI: added configure; init: --seed-vocabulary; updated README + how-to; updated scenarios; ran tests (passing)

### Related Files

- docmgr/cmd/docmgr/main.go
- docmgr/pkg/commands/configure.go
- docmgr/pkg/commands/init.go


## 2025-11-04

Imported Glazed tutorial and style guide; added relate --from-git; added status warnings; updated CLI guide.

### Related Files

- glazed/pkg/doc/topics/how-to-write-good-documentation-pages.md
- glazed/pkg/doc/tutorials/05-build-first-command.md


## 2025-11-04

Unify repo-root detection; default stale-after=30; status warnings; updated CLI guide

### Related Files

- docmgr/pkg/commands/config.go
- docmgr/pkg/commands/doctor.go
- docmgr/pkg/commands/status.go


## 2025-11-04

Extended tests: relate --from-git; status warnings; changelog file-notes; vocab add output (vocabulary_path). All passing.

### Related Files

- docmgr/test-scenarios/testing-doc-manager/09-relate-from-git.sh
- docmgr/test-scenarios/testing-doc-manager/10-status-warnings.sh
- docmgr/test-scenarios/testing-doc-manager/11-changelog-file-notes.sh
- docmgr/test-scenarios/testing-doc-manager/12-vocab-add-output.sh


## 2025-11-04

Verify file notes rendering

### Related Files

- docmgr/pkg/commands/configure.go — CLI configure command
- docmgr/pkg/commands/init.go — --seed-vocabulary flag


## 2025-12-01

Auto-closed: ticket was active but not created today

