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
LastUpdated: 2025-11-04T17:56:55.026326019-05:00
---



# Tasks

## TODO

### Active (CLI/Docs only)

- [x] Add 'docmgr configure' to write .ttmp.yaml — Related: docmgr/pkg/commands/config.go (base); new: docmgr/pkg/commands/configure.go
- [x] Init flag --seed-vocabulary for default types/topics — Related: docmgr/pkg/commands/init.go; docmgr/pkg/commands/vocabulary.go (seed vocabulary)
- [x] relate --from-git to suggest/apply changed files — Related: docmgr/pkg/commands/relate.go (--from-git mode)
- [x] Warn if multiple plausible ttmp roots exist in workspace — Related: docmgr/pkg/commands/config.go (detect multiple roots)
- [x] Warn prominently when falling back to <cwd>/ttmp (no .ttmp.yaml) — Related: docmgr/pkg/commands/config.go (fallback warning)
- [x] Unify repo-root detection across commands; factor .git file (gitdir) support — Related: docmgr/pkg/commands/vocab_add.go; config.go; doctor.go (unify repo detection)
- [x] Reconcile default stale-after (14) vs docs (30); choose and document — Related: docmgr/pkg/commands/status.go; docmgr/pkg/doc/docmgr-how-to-setup.md (stale-after guidance)
- [x] Add docmgr configure onboarding docs and quick-start callout — Related: docmgr/pkg/doc/docmgr-how-to-setup.md; docmgr/pkg/doc/docmgr-cli-guide.md (onboarding)
- [x] Mutating commands include target path in structured/human rows (path fields) — Related: docmgr/pkg/commands/add.go; init.go; meta_update.go; vocab_add.go (include paths in outputs)
- [x] Update README with configuration (root/config/vocabulary) and .docmgrignore guidance — Related: docmgr/README.md (configuration)
- [x] Docs: adopt `.docmgrignore` guidance and de-emphasize ignore flags; show day-to-day flow — Related: go-go-mento/ttmp/how-to-use.md
- [x] Build and install updated docmgr CLI — Related: docmgr/cmd/docmgr/main.go (CLI entrypoint)

### Completed

- [x] Echo context on mutating commands (root/config/vocabulary) — Related: docmgr/pkg/commands/add.go; init.go; meta_update.go; vocab_add.go (echo context)
- [x] Enhance status to show config_path, vocabulary_path, stale source — Related: docmgr/pkg/commands/status.go (config/vocabulary in status)
- [x] Support .git file (gitdir:) in repo detection — Related: docmgr/pkg/commands/vocab_add.go (gitdir support); docmgr/pkg/commands/config.go

## Navigation

- [Index](./index.md)
- [Plan](./various/plan-docmgr-ux-and-multi-repo-improvements.md)
- [Changelog](./changelog.md)
