---
Title: Diagnostics taxonomy handover
Ticket: DOCMGR-ERROR-TAXONOMY
Status: active
Topics:
    - errors
    - ux
    - yaml
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: analysis/01-yapp-error-feedback-port-to-docmgr.md
      Note: Pattern analysis
    - Path: design/02-generic-diagnostics-interface-rollout.md
      Note: Design
    - Path: pkg/diagnostics/docmgrctx/constructors.go
      Note: Constructors helper
    - Path: playbook/01-how-to-add-a-new-diagnostics-domain.md
      Note: Step-by-step
    - Path: tasks.md
      Note: Open tasks
    - Path: test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh
      Note: Smoke coverage
    - Path: working-note/01-diagnostics-integration-diary.md
      Note: Work log
ExternalSources: []
Summary: Handover guide for continuing the diagnostics taxonomy/rules rollout in docmgr.
LastUpdated: 2025-12-01T17:15:00-05:00
---


# Diagnostics taxonomy handover

## Goal
A clear, self-contained handoff so a new contributor can continue the diagnostics taxonomy/rules rollout without hunting for context. It highlights what’s built, what’s left, where to read, and how to resume work with confidence.

## Onboarding reading list
Read these in order to understand the intent and pattern before editing code:
- `design/02-generic-diagnostics-interface-rollout.md` — overall design.
- `analysis/01-yapp-error-feedback-port-to-docmgr.md` — source pattern from YAPP.
- `playbook/01-how-to-add-a-new-diagnostics-domain.md` — step-by-step domain addition (now references constructors).
- `working-note/01-diagnostics-integration-diary.md` — chronological log of what changed.
- `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh` — end-to-end smoke.
- `tasks.md` — remaining TODOs for this ticket.
- `pkg/doc/docmgr-diagnostics-and-rules.md` — help-system page describing architecture, rules, taxonomies, and CLI usage.

## What’s implemented (code map)
Docmgr now uses a shared diagnostics stack (taxonomy + rules + adapter) across several domains.
- Core/adapter: `pkg/diagnostics/core`, `pkg/diagnostics/rules`, `pkg/diagnostics/render`, `pkg/diagnostics/docmgr/adapter.go`.
- Taxonomies: `pkg/diagnostics/docmgrctx/` (frontmatter, vocabulary, related_files, templates, listing, workspace) plus `constructors.go` for consistent builders.
- Rules: `pkg/diagnostics/docmgrrules/` (syntax/schema, vocab, related, template, listing, workspace) registered in `default.go`.
- Renderer: `pkg/diagnostics/docmgr/adapter.go` now supports context-attached collectors; `docmgr doctor --diagnostics-json <path|->` writes JSON rule output for CI.
- Wiring in commands:
  - Frontmatter parse → taxonomy: `internal/documents/frontmatter.go`.
  - Template parse → taxonomy: `pkg/commands/template_validate.go`.
  - Listing skips → taxonomy: `pkg/commands/list_docs.go`.
  - Workspace missing_index/stale → taxonomy rendering: `pkg/commands/doctor.go`.
  - Meta/relate parse errors propagate taxonomy (frontmatter wrapper).
- Smoke coverage: vocab/related/listing/workspace/frontmatter/template via `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh`.

## What’s left (priority order)
Focus on tests/docs to finish the rollout:
1) **Tests/docs** — Add additional coverage for schema wiring + diagnostics JSON flag; update CLI docs (how-to-use/ci-automation) so users see the new diagnostics outputs.

## How to proceed (actionable steps)
1) Testing/docs:
   - Add rule/wiring tests (frontmatter schema, diagnostics JSON plumbing in doctor/listing).
   - Update help/tutorial to mention diagnostics output and new `--diagnostics-json` flag.

## Pseudocode snippets
Use constructors (`constructors.go`) and the shared adapter:
```go
// Schema warnings → taxonomy
for _, w := range validateFrontmatterContent(doc) {
    tax := docmgrctx.NewFrontmatterSchema(file, w.Field, w.Message, core.SeverityWarning)
    docmgr.RenderTaxonomy(ctx, tax)
}
```

```bash
# Emit diagnostics JSON for CI
docmgr doctor --all --diagnostics-json diagnostics.json --fail-on warning
```

## Runbook (quick commands)
- Smoke: `./test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh /tmp/docmgr-scenario`
- Doctor JSON: `docmgr doctor --all --diagnostics-json /tmp/diag.json --fail-on none`
- Tests: `go test ./pkg/diagnostics/docmgrrules ./pkg/commands`

## File pointers (jump here when coding)
- Constructors helper: `pkg/diagnostics/docmgrctx/constructors.go`
- Adapter: `pkg/diagnostics/docmgr/adapter.go`
- Doctor wiring: `pkg/commands/doctor.go`
- Listing wiring: `pkg/commands/list_docs.go`
- Frontmatter wrapper: `internal/documents/frontmatter.go`
- Template wrapper: `pkg/commands/template_validate.go`
- Rules registry: `pkg/diagnostics/docmgrrules/default.go`

## Status snapshot
Changelog and diary capture daily changes; tasks.md lists open items. The smoke script currently demonstrates vocab/related/listing/workspace/frontmatter/template warnings/errors; expect a template parse error in its output as part of coverage.
