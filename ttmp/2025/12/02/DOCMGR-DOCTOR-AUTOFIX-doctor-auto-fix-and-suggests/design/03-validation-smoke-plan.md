---
Title: Validation smoke plan
Ticket: DOCMGR-DOCTOR-AUTOFIX
Status: active
Topics:
    - yaml
    - validation
    - diagnostics
DocType: design
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Inventory of current validation-related smoke tests and plan for additions.
LastUpdated: 2025-12-02T00:00:00Z
---


# Validation smoke plan

## Current validation-related smokes
- `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh`
  - Exercises: vocab warnings, related file missing, listing skip (frontmatter parse), missing index/stale, template parse, diagnostics JSON.
  - Frontmatter coverage: invalid frontmatter (broken doc) → taxonomy with line/snippet; suggests validate command via rules.
- `test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh`
  - Exercises: single-file validation, failure → suggest-fixes → auto-fix (with .bak) → success.

### How to run them (landmarks for an intern)
- Build the binary: `go build -o /tmp/docmgr-bin ./cmd/docmgr`
- From `docmgr/test-scenarios/testing-doc-manager`:
  - `DOCMGR_PATH=/tmp/docmgr-bin ./00-reset.sh /tmp/docmgr-scenario`
  - `DOCMGR_PATH=/tmp/docmgr-bin ./01-create-mock-codebase.sh /tmp/docmgr-scenario`
  - `DOCMGR_PATH=/tmp/docmgr-bin ./02-init-ticket.sh /tmp/docmgr-scenario`
  - `DOCMGR_PATH=/tmp/docmgr-bin ./03-create-docs-and-meta.sh /tmp/docmgr-scenario`
  - Run smokes: `DOCMGR_PATH=/tmp/docmgr-bin ./15-diagnostics-smoke.sh /tmp/docmgr-scenario`, `DOCMGR_PATH=/tmp/docmgr-bin ./18-validate-frontmatter-smoke.sh /tmp/docmgr-scenario`
- Keep `DOCMGR_PATH` pointed to your built binary; scripts assume `/tmp/docmgr-scenario` as the sandbox repo.

## Gaps / desired additional smokes
1) Doctor + validation combo:
   - Run `doctor --doc <file>` on a broken file and ensure frontmatter parse/schema/vocab taxonomies show, pointing to help.
   - Verify diagnostics JSON when using `--diagnostics-json -` in combination with single-file validation.
2) Auto-fix via doctor (future if enabled):
   - If doctor gains `--suggest-fixes/--auto-fix`, add a scenario to exercise it.
3) Related-files in single-file mode (if added):
   - Validate that `doctor --doc` reports missing related files when added to a doc.
4) Schema warnings (when validators are added):
   - Create doc missing Summary/Owners/etc. and confirm `FrontmatterSchema` taxonomies render with hints.

## Next steps
- Extend `15-diagnostics-smoke.sh` to call `docmgr validate frontmatter` on the broken doc to show suggest/auto-fix flow in the same run (non-fatal), and optionally `doctor --doc` to show single-file diagnostics.
- Keep `18-validate-frontmatter-smoke.sh` focused on pure validation/auto-fix.
- Add new smokes later for doctor auto-fix/schema once those features land.

## Pointers to read before editing smokes
- Workflow doc: `pkg/doc/docmgr-doctor-validation-workflow.md`
- Frontmatter healing guide: `ttmp/2025/11/29/DOCMGR-YAML-001.../reference/02-frontmatter-healing-and-validation-guide.md`
- Smokes: `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh`, `18-validate-frontmatter-smoke.sh`
- Validation verb: `pkg/commands/validate_frontmatter.go`
- Doctor entry: `pkg/commands/doctor.go`
