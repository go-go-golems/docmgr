# Tasks

## Implementation plan (high level)
1) Position-aware parsing + diagnostics: extract YAML block, parse with yaml.Node, surface line/col/snippet/problem via `docmgrctx.NewFrontmatterParse`, and plumb through `readDocumentFrontmatter` callers (doctor/list/search/meta/relate/import). Add problem classification helpers.
2) Quoting helpers and preprocessing: `pkg/frontmatter` helpers for `needsQuoting/quoteValue`, preprocessing to auto-quote risky scalars, and reuse them in both read (optional) and write paths.
3) CLI validation verb: `docmgr validate-frontmatter` (file or `--ticket`, with `--suggest-fixes`/`--auto-fix`, backups) using the new parser/validator and emitting diagnostics.
4) Schema/field validation: define frontmatter schema validators with hints (required + optional fields) and emit schema diagnostics in doctor and validate-frontmatter.
5) Write-path hardening: ensure all writers use quoting helpers (`internal/documents/frontmatter.go` and calling commands) and keep LastUpdated behavior intact.
6) Testing and smoke: unit tests for helpers/parser/quoting, plus scenario updates to `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh` (and/or a new scenario) to cover bad frontmatter, validation verb, and auto-fix.
7) Docs/changelog: update `pkg/doc/docmgr-diagnostics-and-rules.md`, CLI guide/how-to-use, and ticket changelog with new commands/behavior.

## TODO

- [x] Position-aware frontmatter parsing and diagnostics
- [x] Add YAML block extraction + yaml.Node parsing to `internal/documents/frontmatter.go`, returning line/col/snippet/problem
- [x] Add problem classification helper (mapping values not allowed, did not find expected key, etc.) for clearer messages
- [x] Update `renderFrontmatterParseTaxonomy`/callers so doctor/list/search/meta/relate/import surface the richer context
- [x] Quoting/preprocessing helpers
- [x] Add `pkg/frontmatter` with `needsQuoting`, `quoteValue`, and preprocessing (auto-quote risky scalars) + unit tests
- [x] Optionally invoke preprocessing before parse in read path (feature-flag or default) to reduce user-facing YAML errors
- [x] Validate-frontmatter CLI verb
- [x] Implement `pkg/commands/validate_frontmatter.go` (file or `--ticket`, `--suggest-fixes`, `--auto-fix` with .bak backups)
- [x] Wire into command tree under `cmd/docmgr/cmds/` and ensure actions in diagnostics rules point to the real verb
- [x] Auto-fix and suggestions
- [x] Add `--suggest-fixes` and `--auto-fix` flags to `docmgr validate frontmatter` with .bak backups
- [x] Implement fix generator (quote unsafe scalars, optionally block scalars for long values) and reuse in CLI/doctor surfaces
- [x] Ensure validate frontmatter emits taxonomies and uses the rules renderer (collector + stdout) for suggestions/fixes, not ad-hoc prints
- [x] Add/adjust rules so frontmatter parse/schema taxonomies can surface suggested fixes (actions or embedded suggestions)
- [x] Ensure diagnostics rules can surface suggested fixes (either in context or alongside rendered output)
- [ ] Field/schema validation with hints
  - [ ] Define frontmatter schema validators (required + optional with hints) and integrate into doctor + validate-frontmatter
  - [ ] Emit schema diagnostics via `docmgrctx.NewFrontmatterSchema` with severity warning/error as appropriate
- [x] Write-path hardening
- [x] Reuse quoting helpers in `internal/documents/WriteDocumentWithFrontmatter` and all command-level writers (add, meta update, create_ticket, doc_move, rename_ticket, ticket_close, import)
- [x] Add focused tests to confirm writer output quotes colons/hashes/@ and preserves unquoted safe scalars
- [ ] Testing and smoke coverage
- [x] Unit tests: extend coverage to suggest/auto-fix behaviors (normalize delimiters, peel trailing lines, scrub stray delimiters, auto-fix success/failure)
  - [ ] Smoke: extend `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh` (and add a small validation scenario if needed) to cover bad frontmatter parsing, line/col/snippet output, validation verb, and auto-fix path
  - [x] Smoke: extend `18-validate-frontmatter-smoke.sh` to exercise `--suggest-fixes` and `--auto-fix` (verify .bak creation and re-parse success)
- [ ] Docs and changelog
  - [ ] Update `pkg/doc/docmgr-diagnostics-and-rules.md`, CLI guide/how-to-use with validate-frontmatter usage and YAML UX changes
- [x] Update ticket `changelog.md` with implemented milestones and note any new flags/verbs
- [ ] Additional validation polish
  - [ ] Suppress/adjust taxonomy emission after successful auto-fix so a passing run doesn’t show an error taxonomy
  - [ ] Ensure doctor/list/search consume richer frontmatter context without double-reporting after auto-fix (if integration needed)
