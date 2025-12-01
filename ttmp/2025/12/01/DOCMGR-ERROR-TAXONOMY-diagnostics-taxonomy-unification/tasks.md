# Tasks

## TODO

- [ ] Create diagnostics core scaffolding in `pkg/diagnostics/core` (ContextPayload, Taxonomy, Severity, Stage/Symptom types, AsTaxonomy/WrapTaxonomy helpers).
- [ ] Add registry + rendering plumbing in `pkg/diagnostics/rules` (registry, scoring) and `pkg/diagnostics/render` (text + JSON output).
- [ ] Define frontmatter stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/frontmatter.go` (parse, schema).
- [ ] Define vocabulary stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/vocabulary.go`.
- [ ] Define related-files stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/related_files.go`.
- [ ] Define template stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/templates.go`.
- [ ] Define listing/skip stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/listing.go`.
- [ ] Define workspace/staleness stage/symptom enums + contexts in `pkg/diagnostics/docmgrctx/workspace.go`.
- [ ] Add taxonomy constructors per domain (factory funcs) in `pkg/diagnostics/docmgrctx/constructors.go`.
- [ ] Wire frontmatter parsing in `internal/documents/frontmatter.go` to wrap parse errors into taxonomy (line/col/snippet).
- [ ] Wire workspace discovery in `internal/workspace/discovery.go` to wrap `FrontmatterErr` into taxonomy before doctor consumes it.
- [ ] Wire command parse sites: return taxonomy-wrapped errors in `pkg/commands/meta_update.go`, `pkg/commands/relate.go`, `pkg/commands/rename_ticket.go`.
- [ ] Wire listing/search skips: surface frontmatter parse skips as taxonomy warnings in `pkg/commands/list_docs.go` and `pkg/commands/search.go`.
- [ ] Wire template validation: wrap `.templ` parse errors into TemplateSyntax taxonomy in `pkg/commands/template_validate.go`.
- [ ] Add CLI adapter `cmd/docmgr/internal/diagnostics/render.go` to render rule cards on taxonomy-bearing errors while preserving glaze rows/exit codes.
- [ ] Implement docmgr rule set in `pkg/diagnostics/docmgrrules` (syntax pointer, schema missing/invalid, vocabulary suggestion, related file missing, template parse, listing skip, workspace structure, staleness) with scores/actions.
- [ ] Extend outputs to include diagnostics JSON/CI and ensure `doctor --fail-on` semantics remain stable.
- [ ] Add tests: unit/snapshot coverage for taxonomy wrapping, rules matching/rendering, renderers; update help/tutorial docs to describe diagnostics behavior.
