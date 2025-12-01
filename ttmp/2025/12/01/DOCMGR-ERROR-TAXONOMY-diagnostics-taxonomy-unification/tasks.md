# Tasks

## TODO

- [x] Create diagnostics core scaffolding in `pkg/diagnostics/core` (ContextPayload, Taxonomy, Severity, Stage/Symptom types, AsTaxonomy/WrapTaxonomy helpers).
- [x] Add registry + rendering plumbing in `pkg/diagnostics/rules` (registry, scoring) and `pkg/diagnostics/render` (text + JSON output).
- [x] Define frontmatter stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/frontmatter.go` (parse, schema).
- [x] Define vocabulary stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/vocabulary.go`.
- [x] Define related-files stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/related_files.go`.
- [x] Define template stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/templates.go`.
- [x] Define listing/skip stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/listing.go`.
- [x] Define workspace/staleness stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/workspace.go`.
- [x] Add taxonomy constructors per domain (factory funcs) in `pkg/diagnostics/docmgrctx/constructors.go` to centralize builder helpers for each stage/symptom.
- [x] Wire frontmatter parsing in `internal/documents/frontmatter.go` to wrap parse errors into taxonomy (line/col/snippet).
- [x] Wire workspace discovery in `internal/workspace/discovery.go` to wrap `FrontmatterErr` into taxonomy so invalid index.md is surfaced structurally (not just doctor fallback).
- [x] Wire command parse sites: return taxonomy-wrapped errors in `pkg/commands/meta_update.go`, `pkg/commands/relate.go`, `pkg/commands/rename_ticket.go`.
- [x] Wire listing/search skips: surface frontmatter parse skips as taxonomy warnings in `pkg/commands/list_docs.go` and `pkg/commands/search.go`.
- [x] Wire template validation: wrap `.templ` parse errors into TemplateSyntax taxonomy in `pkg/commands/template_validate.go`.
- [x] Add CLI adapter `cmd/docmgr/internal/diagnostics/render.go` to render rule cards on taxonomy-bearing errors while preserving glaze rows/exit codes.
- [x] Implement docmgr rule set in `pkg/diagnostics/docmgrrules` (syntax pointer, schema missing/invalid, vocabulary suggestion, related file missing, template parse, listing skip, workspace structure, staleness) with scores/actions.
- [x] Extend outputs to include diagnostics JSON/CI via shared adapter/renderer and ensure `doctor --fail-on` semantics remain stable.
- [ ] Add additional tests for frontmatter schema wrapping + diagnostics JSON flag; update help/tutorial docs (how-to-use/ci-automation) to mention diagnostics outputs.
- [x] Add listing/skip taxonomy contexts and rules, wire list_docs/search to emit taxonomy instead of silent skips.
- [x] Add workspace/staleness taxonomy and rule(s), wire discovery/doctor staleness checks.
- [x] Create shared diagnostics adapter in pkg/diagnostics/docmgr and refactor doctor to use it
- [x] Wire frontmatter parse errors to taxonomy via internal/documents/frontmatter
- [x] Wire frontmatter schema validation warnings to taxonomy (emit schema violation taxonomies in validation/doctor paths).
- [x] Wire template parse errors to taxonomy
- [x] Add listing/skip taxonomy contexts and rules, wire list_docs/search to emit taxonomy instead of silent skips
- [x] Add workspace/staleness taxonomy and rule(s), wire discovery/doctor staleness checks
- [x] Wire workspace missing_index/stale findings in doctor to emit workspace taxonomies
- [x] Wire meta_update/relate/rename_ticket error paths to taxonomy
- [x] Wire missing_index findings to workspace taxonomy in doctor
- [x] Wire meta_update/relate/rename_ticket to wrap frontmatter parse errors with taxonomy
- [x] Add CLI adapter for CI/JSON diagnostics output (shared renderer flags) to expose diagnostics in machine-readable form.
