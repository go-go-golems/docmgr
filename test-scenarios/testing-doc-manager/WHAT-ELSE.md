# What Else to Fix / Implement / Clarify / Improve

This checklist captures findings from testing and areas to improve. We will check them off as we implement.

## CLI Consistency and Help
- [x] Use `--ticket` for init (positional not supported)
- [x] Use `--query` for search content (positional not supported)
- [ ] Audit built-in help and examples for flag consistency

## Doctor Improvements
- [x] Ignore scaffolding directories at root (prefix `_`)
- [x] Add `--ignore-dir` (repeatable) to exclude directory names
- [x] Add `--ignore-glob` patterns to exclude paths
- [x] Add `--stale-after <days>` to configure threshold
- [x] Add `--fail-on {none,warning,error}` to control exit code
- [x] Validate `Topics`, `DocType`, `Intent` against vocabulary; warn on unknowns
- [x] Warn when `RelatedFiles` paths do not exist on disk
- [ ] Consider `--ticket` filter for doctor to narrow scope (exists, ensure docs mention)
- [ ] Add JSON/yaml output examples for CI consumption

## Search Enhancements
- [ ] Improve snippet extraction (collapse blank lines; include more context)
- [ ] Add `--references-doc` (docs referencing another doc)
- [ ] Add boost for title/frontmatter matches vs body text
- [ ] Add `--sort updated` / `--sort created` options

## Templates & Guidelines
- [ ] Ensure `_templates/index.md` has valid frontmatter or is ignored by doctor
- [ ] Add examples/templates for common doc types (ADR, RFC, incident)

## DX & CI
- [ ] Add `make` targets (build, test-e2e)
- [ ] Add CI job that runs scenario and asserts `doctor --fail-on error` passes
- [ ] Add environment var to default `--root` (e.g., `DOCMGR_ROOT`)

## Documentation
- [x] Create scenario README and SCENARIO walkthrough
- [x] Create TEST-NOTES.md for observations
- [ ] Generate a troubleshooting guide (common errors & fixes)

## Nice-to-Haves
- [ ] Ticket inference from branch name
- [ ] Template variable substitution in `add`
- [ ] Migration tool `migrate --dry-run` + report
- [ ] Unit/integration/golden tests for commands
