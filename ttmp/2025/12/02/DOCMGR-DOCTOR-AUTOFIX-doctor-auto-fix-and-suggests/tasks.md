# Tasks

## TODO

- [ ] Design how doctor will expose suggest/auto-fix:
  - [ ] Define flags (e.g., `--suggest-fixes`, `--auto-fix`) and scope (per-file? per-ticket?).
  - [ ] Decide safety constraints (backups, dry-run, confirm prompts).
- [ ] Reuse fix generator from `pkg/commands/validate_frontmatter.go`:
  - [ ] Extract or share `generateFixes`/`normalizeDelimiters`/`scrubStrayDelimiters`/`peelTrailingNonKeyLines`/`PreprocessYAML` for doctor use.
  - [ ] Ensure taxonomies carry `Fixes` so rules can render suggestions in doctor output.
- [ ] Integrate into doctor flows:
  - [ ] Ticket scan paths (index + subdocs) should offer suggestions; decide when to auto-fix vs suggest-only.
  - [ ] Single-file mode (`doctor --doc`) should mirror `validate frontmatter` behaviors if flags are set.
- [ ] Update rules/messages:
  - [ ] Ensure frontmatter rules mention the new doctor flags and reference help (`docmgr help yaml-frontmatter-validation`).
- [ ] Tests and smokes:
  - [ ] Add unit coverage for doctor auto-fix invocation.
  - [ ] Extend diagnostics smoke to exercise doctor suggest/auto-fix (with backups).
- [ ] Docs:
  - [ ] Update `pkg/doc/docmgr-diagnostics-and-rules.md`, CLI guides, and the YAML validation help page to mention doctor’s suggest/auto-fix behavior.
- [ ] Design doctor suggest/auto-fix flags and scope (per-file/per-ticket) with safety (backups/dry-run).
