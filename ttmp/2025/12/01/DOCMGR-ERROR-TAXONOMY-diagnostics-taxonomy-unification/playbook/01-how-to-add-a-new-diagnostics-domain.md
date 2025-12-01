---
Title: How to add a new diagnostics domain
Ticket: DOCMGR-ERROR-TAXONOMY
Status: active
Topics:
    - errors
    - ux
    - yaml
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: analysis/01-yapp-error-feedback-port-to-docmgr.md
      Note: Pattern source
    - Path: design/02-generic-diagnostics-interface-rollout.md
      Note: Design context
    - Path: pkg/commands/list_docs.go
      Note: Listing taxonomy wiring
    - Path: pkg/diagnostics/docmgrctx/listing.go
      Note: Listing taxonomy
    - Path: pkg/diagnostics/docmgrctx/workspace.go
      Note: Workspace/staleness taxonomy
    - Path: pkg/diagnostics/docmgrrules/listing_rule.go
      Note: Listing rule
    - Path: pkg/diagnostics/docmgrrules/workspace_rule.go
      Note: Workspace rule
ExternalSources: []
Summary: Playbook for adding a new diagnostics domain (taxonomy, rules, wiring, tests) in docmgr.
LastUpdated: 2025-12-01T13:32:04-05:00
---



# How to add a new diagnostics domain

This playbook guides a new contributor through extending docmgr’s diagnostics system (taxonomy + rules + adapter). For context, read:
- Design: `design/02-generic-diagnostics-interface-rollout.md`
- Analysis: `analysis/01-yapp-error-feedback-port-to-docmgr.md`
- Examples: existing domains under `pkg/diagnostics/docmgrctx` and `pkg/diagnostics/docmgrrules` (vocabulary, related-files, frontmatter, template, listing, workspace).

## Purpose
Step-by-step recipe to add a new diagnostics domain (stage/symptom/context, rules, wiring, tests) in docmgr.

## Environment Assumptions
- Go toolchain available; repo checked out.
- Diagnostics core scaffold present (`pkg/diagnostics/core`, `.../docmgrctx`, `.../docmgrrules`, adapter).
- gofmt and go test available.

## Commands / Steps

1) **Define stage/symptom + context**
   - Add `pkg/diagnostics/docmgrctx/<domain>.go`.
   - Export `Stage<Domain>`, `Symptom<Something>`, context struct implementing `Stage()`/`Summary()`.
   - Add constructor returning `*core.Taxonomy` (set Severity, Path, Context, Cause).

2) **Add a rule**
   - Create `pkg/diagnostics/docmgrrules/<domain>_rule.go`.
   - Implement `Match` (stage + symptom) and `Render` (headline/body/actions).
   - Keep actions copy-paste friendly (docmgr commands).

3) **Register the rule**
   - Update `pkg/diagnostics/docmgrrules/default.go` to `Register(&<Domain>Rule{})`.

4) **Wire taxonomy emission**
   - Locate the producing code (e.g., command parse, validator, walker).
   - Wrap the error/condition with the constructor and call adapter:
     ```go
     docmgr.RenderTaxonomy(ctx, docmgrctx.New<Domain>Taxonomy(...))
     ```
   - Replace silent skips with warnings where appropriate (list/search).

5) **Adapter reuse**
   - Commands should not duplicate rendering; use `pkg/diagnostics/docmgr/adapter.go`.

6) **Tests**
   - Add/extend rule tests under `pkg/diagnostics/docmgrrules`.
   - Run: `go test ./pkg/diagnostics/docmgrrules ./pkg/commands`.

7) **Scenario (optional)**
   - If helpful, add a script under `docmgr/test-scenarios/testing-doc-manager/` that triggers the new domain and shows doctor output.

## Exit Criteria
- Stage/symptom/context and constructor in `docmgrctx`.
- Rule registered and rendering helpful guidance.
- Producing code wraps errors into taxonomy and renders via adapter.
- gofmt + go test pass; scenario (if added) demonstrates the warning/error.
