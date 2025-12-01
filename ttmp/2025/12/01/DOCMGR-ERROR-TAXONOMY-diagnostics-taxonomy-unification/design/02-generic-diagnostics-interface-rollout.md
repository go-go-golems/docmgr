---
Title: Generic diagnostics interface rollout
Ticket: DOCMGR-ERROR-TAXONOMY
Status: active
Topics:
    - yaml
    - ux
    - errors
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/diagnostics/core/types.go
      Note: Taxonomy core types and wrapping
    - Path: pkg/diagnostics/docmgrctx/related_files.go
      Note: Related files taxonomy context
    - Path: pkg/diagnostics/docmgrctx/vocabulary.go
      Note: Vocabulary taxonomy context
    - Path: pkg/diagnostics/docmgrrules/default.go
      Note: Registry seeding for docmgr rules
    - Path: pkg/diagnostics/docmgrrules/related_files_rule.go
      Note: Related files rule
    - Path: pkg/diagnostics/docmgrrules/related_files_rule_test.go
      Note: Rule test
    - Path: pkg/diagnostics/docmgrrules/vocabulary_rule.go
      Note: Vocabulary rule
    - Path: pkg/diagnostics/docmgrrules/vocabulary_rule_test.go
      Note: Rule test
    - Path: pkg/diagnostics/render/render.go
      Note: Text/JSON renderers for diagnostics
    - Path: pkg/diagnostics/rules/rules.go
      Note: Rule registry and scoring
    - Path: test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh
      Note: Diagnostics smoke exercising docmgr binary
ExternalSources: []
Summary: Design a generic diagnostics interface (ContextPayload-based) and map docmgr validation/error surfaces to it for consistent rule-driven output.
LastUpdated: 2025-12-01T11:56:30-05:00
---



# Generic diagnostics interface rollout

## Objective

Adopt the generic diagnostics interface (ContextPayload) across docmgr so validation and error reporting flow through a shared taxonomy + rule renderer. Goal: consolidate parsing/validation errors (frontmatter, vocabulary, related files, template syntax, tasks) into structured, actionable output for both human and machine consumers.

## Generic interface (recap)

Shared core (proposal):
```go
// pkg/diagnostics/core/context.go
type ContextPayload interface {
    Stage() StageCode    // optional: renderer can reject mismatches
    Summary() string     // fallback string if concrete cast fails
}

type Taxonomy struct {
    Tool     string
    Stage    StageCode
    Symptom  SymptomCode
    Path     string
    Severity Severity
    Context  ContextPayload
    Cause    error
}

func (t *Taxonomy) ContextSummary() string {
    if t.Context == nil { return "" }
    return t.Context.Summary()
}
```

Example docmgr context:
```go
type FrontmatterParseContext struct {
    File    string
    Line    int
    Column  int
    Snippet string
    Problem string
}
func (c *FrontmatterParseContext) Stage() core.StageCode { return StageFrontmatterParse }
func (c *FrontmatterParseContext) Summary() string {
    return fmt.Sprintf("%s:%d:%d %s", c.File, c.Line, c.Column, c.Problem)
}
```

Renderers try concrete casts first; otherwise fall back to `ContextSummary()` to avoid panics while still showing location/problem.

## Docmgr validation and error surfaces (inventory)

- Frontmatter parse/write helper: `internal/documents/frontmatter.go`
  - `ReadDocumentWithFrontmatter`: wraps `adrg/frontmatter` errors as `fmt.Errorf("parse frontmatter...")`.
  - `WriteDocumentWithFrontmatter`: YAML encode/write errors.
- Workspace discovery: `internal/workspace/discovery.go`
  - `CollectTicketWorkspaces` stores `FrontmatterErr` when index.md fails to parse; doctor consumes this.
- Doctor command: `pkg/commands/doctor.go`
  - Emits findings for `invalid_frontmatter`, `missing_required_fields`, `unknown_topics/status`, `missing_index`, stale docs, glob ignores.
  - Uses glazed rows + human renderer; currently plain strings, no taxonomy/rules.
- Listing commands:
  - `pkg/commands/list_docs.go`: silently skips docs with invalid frontmatter via `readDocumentFrontmatter`.
  - `pkg/commands/list_tickets.go` (not shown above) likely similar when enumerating tickets; needs taxonomy for parse failures instead of skip.
  - `pkg/commands/search.go` (frontmatter-based filtering) skips on parse failure.
- Mutators that parse frontmatter:
  - `pkg/commands/meta_update.go`: `ReadDocumentWithFrontmatter`, errors bubble as `failed to parse frontmatter`.
  - `pkg/commands/relate.go`: `failed to read document frontmatter` when updating RelatedFiles.
  - `pkg/commands/rename_ticket.go`: updates Ticket in all frontmatters; returns wrapped parse/write errors.
  - `pkg/commands/add.go` (doc creation) writes new frontmatter; errors from filesystem and template rendering.
- Template validation: `pkg/commands/template_validate.go`
  - Validates `.templ` files; prints `ERROR: path: parse error` or returns `validation failed` count.
- Generation/printing helpers:
  - Human printers in doctor/list commands use glamour; glaze rows (JSON/CSV) map 1:1 to fields (`issue`, `message`, `path`, etc.).
  - Error wrapping is ad hoc (`fmt.Errorf`), and skips hide parse issues from users (list/search).

## Proposed taxonomy + rules mapping (docmgr)

Stages (examples):
- `StageFrontmatterParse` — raw YAML/frontmatter parsing failures.
- `StageFrontmatterSchema` — required/malformed fields, type mismatches.
- `StageVocabulary` — unknown topic/status/docType/intent.
- `StageDocLink` — RelatedFiles missing or empty notes.
- `StageTemplateSyntax` — `.templ` parsing.
- `StageListingParse` — list/search/collect skipping due to parse issues.
- `StageWorkspaceStructure` — missing index.md, invalid scaffolds.
- `StageStaleness` — stale LastUpdated thresholds.

Symptoms:
- `SymptomSyntax`, `SymptomMissingRequired`, `SymptomTypeMismatch`, `SymptomUnknownValue`, `SymptomMissingFile`, `SymptomTemplateParse`, `SymptomSkippedDueToParse`, `SymptomMissingIndex`, `SymptomStale`.

Contexts (tool-specific, implement ContextPayload):
- `FrontmatterParseContext{File,Line,Column,Snippet,Problem}`.
- `FrontmatterSchemaContext{File,Field,Expected,Actual,Hint}`.
- `VocabularyContext{File,Field,Value,Known []string}`.
- `RelatedFileContext{DocPath,FilePath,Exists bool,Note string}`.
- `TemplateParseContext{File,Err string}`.
- `ListingSkipContext{File,Command string,Reason string}` (for list/search skips).
- `WorkspaceStructureContext{Path string, Missing string}`.
- `StalenessContext{File string, LastUpdated time.Time, ThresholdDays int}`.

Rules (shared registry):
- YAML syntax pointer rule → uses `FrontmatterParseContext` snippet + fix actions (`validate-frontmatter`/`meta update`).
- Schema missing/invalid rule → `FrontmatterSchemaContext` with per-field hints.
- Vocabulary suggestion rule → `VocabularyContext` with closest matches and `docmgr vocab add` actions.
- RelatedFiles missing file rule → `RelatedFileContext` with `docmgr doc relate --remove-files` or path fix.
- Template parse rule → `TemplateParseContext` with `docmgr template validate --path` action.
- Listing skip visibility rule → `ListingSkipContext` surfaces previously silent skips (list/search) as warnings.
- Workspace structure rule → `WorkspaceStructureContext` suggests rerunning `ticket create-ticket`.
- Staleness rule → `StalenessContext` summarizing age and suggesting update/changelog.

## Application plan to docmgr codebase

1) Core plumbing
   - Add `pkg/diagnostics/core` with `ContextPayload`, `Taxonomy`, `Severity`, `StageCode`, `SymptomCode`, `AsTaxonomy/WrapTaxonomy`.
   - Add `pkg/diagnostics/render` (text + JSON) and `pkg/diagnostics/rules` (registry + scoring).
   - Add `pkg/diagnostics/docmgrctx` for docmgr context structs and constructors.

2) Command-layer adapter
   - New `cmd/docmgr/internal/diagnostics/render.go` with `HandleError(ctx, err) error`:
     ```go
     if t, ok := core.AsTaxonomy(err); ok {
         reg := rules.DefaultRegistry() // populated in init() by docmgr rules
         results, _ := reg.RenderAll(ctx, t)
         fmt.Fprintln(os.Stderr, render.RenderToText(results))
     }
     return err
     ```
   - Commands returning errors (`doctor`, `meta update`, `relate`, `rename-ticket`, `template validate`) call `return HandleError(ctx, err)` at top-level Run.

3) Wrap parse/validation sites with taxonomy
   - `internal/documents/frontmatter.go`: when `frontmatter.Parse` fails, wrap with `docmgrctx.NewFrontmatterParseTaxonomy(path, line, col, snippet, problem, err)`.
   - `internal/workspace/discovery.go`: on `FrontmatterErr`, wrap before storing; doctor receives taxonomy instead of opaque error.
   - `pkg/commands/doctor.go`: when emitting findings (invalid_frontmatter, missing_required_fields, unknown_topics/status, missing_index, stale), create taxonomy per finding and pass to renderer instead of hand-formatted strings; still emit glaze rows for machine output.
   - `pkg/commands/list_docs.go` / `search.go`: instead of silent skip on parse error, wrap in `ListingSkipContext` with `SymptomSkippedDueToParse` and render as warning (while continuing).
   - `pkg/commands/meta_update.go`, `relate.go`, `rename_ticket.go`: wrap frontmatter parse/write failures as `StageFrontmatterParse` (parse) or `StageFrontmatterSchema` (unknown field) to display actionable guidance.
   - `pkg/commands/template_validate.go`: wrap template parse errors as `StageTemplateSyntax` with `TemplateParseContext` so renderer prints helpful card.

4) Rule registration
   - `pkg/diagnostics/docmgrrules/default.go`: register docmgr rules (syntax pointer, schema missing/invalid, vocabulary, related-file, template parse, listing skip, staleness, workspace structure).
   - Optionally reuse YAPP rules when compatible; key is ContextPayload alignment.

5) Outputs
   - Human: text renderer (cards) before existing glamour output; keep glamour for summaries but prepend diagnostics block.
   - Glaze/JSON: add `--with-glaze-output --output json` to emit `[]RuleResult` for CI/IDE.
   - Ensure exit codes remain consistent (`fail-on` in doctor).

6) Safety/compat
   - If `AsTaxonomy` fails (unexpected context), commands fall back to current behavior.
   - ContextPayload avoids `any` yet doesn’t force renderers to know every struct; `Summary()` supplies safe fallback.

## Pseudocode: doctor invalid frontmatter

```go
if ws.FrontmatterErr != nil {
    t := docmgrctx.NewFrontmatterParseTaxonomy(
        filepath.Join(ticketPath, "index.md"),
        lineFromErr(ws.FrontmatterErr),
        colFromErr(ws.FrontmatterErr),
        snippet(path, line), // reuse snippet helper
        classifyProblem(ws.FrontmatterErr),
        ws.FrontmatterErr,
    )
    _ = render.HandleTaxonomy(ctx, t) // prints rule cards

    row := types.NewRow(
        types.MRP("ticket", filepath.Base(ticketPath)),
        types.MRP("issue", "invalid_frontmatter"),
        types.MRP("severity", "error"),
        types.MRP("message", t.ContextSummary()),
        types.MRP("path", ticketPath),
    )
    _ = gp.AddRow(ctx, row)
    highestSeverity = maxInt(highestSeverity, 2)
    continue
}
```

## Rollout steps

1. Scaffold `pkg/diagnostics/core`, `render`, `rules`, `docmgrctx`, `docmgrrules` with ContextPayload interface and registry.
2. Wrap frontmatter/template parsing and validation entry points with taxonomy constructors (files above).
3. Replace silent skips in list/search with warning taxonomies (still continue execution).
4. Add renderer adapter in CLI commands; ensure human + glaze outputs include diagnostics (text + optional JSON).
5. Add unit tests for `ContextPayload` implementations, taxonomy wrapping helpers, and rule rendering snapshots (syntax pointer, vocabulary, template parse).
6. Document new behavior in help/tutorial (mention actionable diagnostics, JSON output).

## Open questions

- Where to host shared diagnostics core (inside docmgr module vs shared module for YAPP + docmgr)?
- Versioning approach so YAPP rule implementations remain compatible with ContextPayload changes.
- Whether list/search should emit warnings by default or behind a flag (to avoid noisy output in large repos).
