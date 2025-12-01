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
  - [ ] Update `renderFrontmatterParseTaxonomy`/callers so doctor/list/search/meta/relate/import surface the richer context
- [ ] Quoting/preprocessing helpers
  - [ ] Add `pkg/frontmatter` with `needsQuoting`, `quoteValue`, and preprocessing (auto-quote risky scalars) + unit tests
  - [ ] Optionally invoke preprocessing before parse in read path (feature-flag or default) to reduce user-facing YAML errors
- [x] Validate-frontmatter CLI verb
- [x] Implement `pkg/commands/validate_frontmatter.go` (file or `--ticket`, `--suggest-fixes`, `--auto-fix` with .bak backups)
- [x] Wire into command tree under `cmd/docmgr/cmds/` and ensure actions in diagnostics rules point to the real verb
- [ ] Field/schema validation with hints
  - [ ] Define frontmatter schema validators (required + optional with hints) and integrate into doctor + validate-frontmatter
  - [ ] Emit schema diagnostics via `docmgrctx.NewFrontmatterSchema` with severity warning/error as appropriate
- [ ] Write-path hardening
  - [ ] Reuse quoting helpers in `internal/documents/WriteDocumentWithFrontmatter` and all command-level writers (add, meta update, create_ticket, doc_move, rename_ticket, ticket_close, import)
  - [ ] Add focused tests to confirm writer output quotes colons/hashes/@ and preserves unquoted safe scalars
- [ ] Testing and smoke coverage
  - [ ] Unit tests: quoting helpers, preprocessing, parser problem classification, validate-frontmatter behavior
  - [ ] Smoke: extend `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh` (and add a small validation scenario if needed) to cover bad frontmatter parsing, line/col/snippet output, validation verb, and auto-fix path
- [ ] Docs and changelog
  - [ ] Update `pkg/doc/docmgr-diagnostics-and-rules.md`, CLI guide/how-to-use with validate-frontmatter usage and YAML UX changes
- [x] Update ticket `changelog.md` with implemented milestones and note any new flags/verbs
