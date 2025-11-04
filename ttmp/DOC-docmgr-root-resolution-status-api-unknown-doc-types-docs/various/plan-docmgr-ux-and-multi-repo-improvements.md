---
Title: Plan — Docmgr UX and Multi-repo Improvements
Ticket: DOC
Status: active
Topics:
    - infrastructure
    - backend
DocType: planning
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/README.md
      Note: Server configuration and endpoint documentation
    - Path: docmgr/cmd/docmgr-server/main.go
      Note: Server root resolution and /api/status
    - Path: docmgr/pkg/commands/add.go
      Note: Echo resolved context before writes
    - Path: docmgr/pkg/commands/config.go
      Note: Root/config discovery; DOCMGR_CONFIG support
    - Path: docmgr/pkg/commands/init.go
      Note: Echo resolved context; scaffolding behavior
    - Path: docmgr/pkg/commands/meta_update.go
      Note: Echo resolved context; field updates
    - Path: docmgr/pkg/commands/status.go
      Note: Add config/vocabulary fields to status
    - Path: docmgr/pkg/commands/vocab_add.go
      Note: Echo resolved context; repo root detection
    - Path: docmgr/pkg/doc/docmgr-cli-guide.md
      Note: Quick-start doctor example without ignore flags
    - Path: docmgr/pkg/doc/docmgr-how-to-setup.md
      Note: Multi-repo and CI guidance
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Clarified .docmgrignore and doctor usage
    - Path: docmgr/pkg/doc/tutorials/docmgr-multi-repo-and-server.md
      Note: New tutorial for multi-repo + server
ExternalSources: []
Summary: Unified repo-root detection; default stale-after=30; relate --from-git documented; warnings in status; updated docs.
LastUpdated: 2025-11-04T17:56:55.026456434-05:00
---









# Plan — Docmgr UX and Multi-repo Improvements

## 1. Problem Statement (What was confusing)
- Root resolution in multi-repo led to writes in the wrong `ttmp/` until `.ttmp.yaml` lived at the workspace root; `.git` as a file (gitdir) made repo discovery non-obvious.
- Mutating commands didn’t show where they wrote (easy to seed the wrong `vocabulary.yaml`).
- Multi-repo onboarding not emphasized; users miss the “place `.ttmp.yaml` one dir up” guidance.
- `init` defaults feel bare; empty `vocabulary.yaml` weakens early validation.
- `status` omits `config_path` and `vocabulary_path`; verification requires extra steps.
- Root discovery spec references `.git/` dir, but many repos use a `.git` file with `gitdir:`.
- Docs suggest `--stale-after 30` while defaults are 14.

## 2. Goals and Non‑Goals
### Goals
- Make root resolution predictable in multi-repo workspaces and visible on every write.
- Reduce setup friction: better onboarding docs and optional seeded vocabulary.
- Improve status/doctor telemetry for quick verification and CI use.
- Provide QoL commands for config and git-driven relating.
### Non‑Goals
- Changing the on‑disk workspace layout.
- Implementing a GUI.

## 3. Proposed Features and Behavior
1) Echo resolved context on every mutating command
   - Before persisting, print: `root=… config=… vocabulary=…`.
   - If falling back to `<cwd>/ttmp` (no config), print a prominent hint to create `.ttmp.yaml` or pass `--root`.
   - If multiple plausible `ttmp/` roots exist in the workspace, print a warning.

2) Smarter root detection
   - `.ttmp.yaml` wins (nearest up‑tree). Resolve relative paths from the config file’s directory.
   - Support `.git` as a file (read `gitdir:`) in repo detection helpers.
   - Honor `DOCMGR_ROOT`; add `DOCMGR_CONFIG` to point at a config file explicitly.

3) First‑class multi‑repo onboarding
   - New `docmgr configure --root <path> [--write-config <dir>]` to write `.ttmp.yaml` with normalized paths and defaults.
   - Update quick‑starts to begin with `docmgr status` and a “multi‑repo” callout.

4) Optional seeded defaults on `init`
   - Flag `--seed-vocabulary` pre-populates common DocTypes (`index`, `design-doc`, `reference`, `playbook`, `code-review`) and Topics (`backend`, `frontend`).
   - Idempotent (skip if already present).

5) Enhance `status`
   - Add columns/fields: `config_path`, `vocabulary_path`, `stale_after_days` (and indicate default vs configured in human output).

6) Clearer warnings
   - When using fallback `<cwd>/ttmp`, log a one‑liner with suggested remediation.

7) QoL: git‑assisted relating
   - `docmgr relate --from-git <refspec> --base <base> --ticket <TCK>` suggests files changed in a PR/range and applies them to `RelatedFiles` (with optional `--apply`).

## 4. Design Details and Touchpoints
- Root/config resolution (CLI): `docmgr/pkg/commands/config.go` — extend `ResolveRoot`, `ResolveVocabularyPath`, add `LoadTTMPConfig`, respect `DOCMGR_CONFIG`.
- Repo detection: update `findRepoRoot()` helpers (e.g., `vocab_add.go`) to parse `.git` file and `gitdir:`.
- Mutating commands: `add.go`, `init.go`, `meta_update.go`, `vocab_add.go`, `changelog.go`, `tasks.go`, `import_file.go`, `relate.go` — print resolved context prior to writes.
- `status`: `pkg/commands/status.go` — include config/vocabulary paths and stale source.
- `doctor`: continue to auto‑apply `.docmgrignore` globs; document this (done).
- New command: `pkg/commands/configure.go` — writes `.ttmp.yaml` with normalized, repo‑relative paths.
- Git relating: extend `relate.go` with a `--from-git` mode that shells out to `git diff --name-only` given a base and ref.

## 5. UX and Output Examples
### Mutating command preface
```
root=/workspace/go-go-mento/ttmp config=/workspace/.ttmp.yaml vocabulary=/workspace/go-go-mento/ttmp/vocabulary.yaml
```
Warn fallback:
```
root=/current/dir/ttmp (fallback) — create .ttmp.yaml or pass --root to avoid writing to the wrong place
```

### Status additions
Human:
```
root=… config=… vocabulary=… stale-after=30 (configured)
```
Structured (existing glaze output gains fields `config_path`, `vocabulary_path`).

## 6. Implementation Plan (Phased)
1) Plumbing and visibility
   - Add `DOCMGR_CONFIG` support; extend config loaders and `ResolveRoot`/`ResolveVocabularyPath`.
   - Implement pre‑write context printing in mutating commands.
   - Update `status` output with paths and stale source.

2) Repo detection improvements
   - Enhance `findRepoRoot` to handle `.git` files (parse `gitdir:`).
   - Add unit tests for repo detection edge‑cases.

3) Onboarding and seeding
   - Implement `docmgr configure` writing `.ttmp.yaml`.
   - `init --seed-vocabulary` flag with idempotent writes.
   - Update embedded docs (done for `.docmgrignore`; extend for multi‑repo quick‑start with `configure`).

4) Git‑assisted relate
   - `relate --from-git` plumbing (parse git output, optional `--apply`).

## 7. Acceptance Criteria
- Mutating commands always print resolved root/config/vocabulary before changes.
- `status` shows `config_path`, `vocabulary_path`, and the active `stale_after_days` source.
- Repo discovery supports `.git` file with `gitdir:`.
- `configure` generates a correct `.ttmp.yaml` at the selected directory.
- `init --seed-vocabulary` seeds common DocTypes/Topics when empty.
- `relate --from-git` suggests and can apply related files from git ranges.
- Docs clearly show `.docmgrignore` globs obviate `--ignore-dir`.

## 8. Risks and Mitigations
- Over‑verbose output: gate context printing behind `--quiet` (default noisy on writes).
- Mis‑resolving config: include absolute paths in logs; prefer `.ttmp.yaml` nearest up‑tree; `DOCMGR_CONFIG` explicit override.
- Git plumbing portability: avoid porcelain‑dependent flags; stick to `git diff --name-only`.

## 9. Testing Plan
- Unit tests: config resolution, `.git` file parsing, repo detection fallbacks, seeded vocabulary idempotency.
- CLI snapshot tests: mutating prefaces, `status` structured output fields present.
- E2E: temporary workspace that mimics multi‑repo layout; verify writes land under the intended `root`.

## 10. Milestones
- M1: Context printing + status fields.
- M2: `.git` file support + tests.
- M3: `configure` command + docs.
- M4: `init --seed-vocabulary` + docs.
- M5: `relate --from-git` + demos.

## 11. Open Questions
- Should `stale-after` default increase from 14 to 30, or stay 14 but highlight doc guidance?
- Should unknown DocTypes always land in `various/` regardless of vocabulary presence, or warn when vocabulary forbids?
