---
Title: Diagnostics Taxonomy, Rules, and CLI Integration
Slug: diagnostics-taxonomy-and-rules
Short: End-to-end guide to docmgr’s diagnostics taxonomy, rule renderers, CLI verbs, and JSON outputs.
Topics:
- docmgr
- diagnostics
- validation
- cli
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Diagnostics Taxonomy, Rules, and CLI Integration

Docmgr’s diagnostics system turns raw errors into actionable guidance for humans and machines. It wraps parsing/validation failures into structured taxonomies, renders guidance via rules, and exposes the output through CLI verbs (human text and JSON for CI). This page explains the architecture, the existing domains, how CLI verbs emit diagnostics, and how to extend the system.

## 1. Architecture Overview

The diagnostics stack is split into small packages so domain logic stays reusable and testable:
- Core types (`pkg/diagnostics/core`): `Taxonomy`, `ContextPayload`, severities, stage/symptom codes, and error wrapping (`WrapWithCause`, `AsTaxonomy`).
- Taxonomy contexts (`pkg/diagnostics/docmgrctx`): Domain-specific stages, symptom codes, and context structs (frontmatter, vocabulary, related-files, templates, listing, workspace) plus constructors in `constructors.go`.
- Rules (`pkg/diagnostics/docmgrrules`): Renderers that match a taxonomy and produce `RuleResult` with headline, body, severity, and suggested actions.
- Renderer/adapter (`pkg/diagnostics/docmgr/adapter.go`): Default registry + text rendering; supports collectors for JSON output and context attachment (`ContextWithRenderer`).
- Output formatting (`pkg/diagnostics/render`): Text and JSON helpers.
- Rule registry (`pkg/diagnostics/rules`): Rule registration and scoring.

The flow: a verb detects an issue → builds a taxonomy via constructors → `docmgr.RenderTaxonomy` runs registered rules → text goes to stderr (default) and optionally into a collector for JSON.

## 2. Data Flow and Usage Patterns

Diagnostics are emitted in a uniform pattern so verbs stay simple:
```go
// Wrap an error with taxonomy for Go error chains
return core.WrapWithCause(err, docmgrctx.NewFrontmatterParse(path, line, col, snippet, problem, err))

// Render ad-hoc taxonomy (e.g., missing related file)
docmgr.RenderTaxonomy(ctx, docmgrctx.NewRelatedFileMissing(docPath, rf.Path, rf.Note))

// Attach a collector for JSON output (doctor uses this for --diagnostics-json)
renderer := docmgr.NewRenderer(docmgr.WithCollector())
ctx = docmgr.ContextWithRenderer(ctx, renderer)
// ... emit taxonomies ...
data, _ := renderer.JSON() // pretty JSON of all rule results
```

Rules stay decoupled from producers: producers only construct taxonomies; renderers decide how to display or serialize them.

## 3. Taxonomy Domains (Stages, Symptoms, Contexts)

Each domain defines stage/symptom codes and context structs under `pkg/diagnostics/docmgrctx`:
- **Frontmatter (`frontmatter.go`)**: `StageFrontmatterParse`, `SymptomYAMLSyntax`, `SymptomSchemaViolation`; contexts carry file, line/col, snippet, field, and detail. Constructors: `NewFrontmatterParse`, `NewFrontmatterSchema`.
- **Vocabulary (`vocabulary.go`)**: `StageVocabulary`, `SymptomUnknownValue`; context holds file, field, offending value, known values. Constructor: `NewVocabularyUnknown`.
- **Related files (`related_files.go`)**: `StageRelatedFiles`, `SymptomMissingFile`; context includes doc path, related path, and note. Constructor: `NewRelatedFileMissing`.
- **Templates (`templates.go`)**: `StageTemplateParse`, `SymptomTemplateParseError`; context has template path and problem. Constructor: `NewTemplateParse`.
- **Listing (`listing.go`)**: `StageListing`, `SymptomSkippedDueToParse`; context records command (`list_docs`/`search`), file, and reason. Constructor: `NewListingSkip`.
- **Workspace (`workspace.go`)**: `StageWorkspace`, `SymptomMissingIndex`, `SymptomStaleDoc`; contexts note missing ticket path or staleness metadata. Constructors: `NewWorkspaceMissingIndex`, `NewWorkspaceStale`.

All constructors are re-exported via `pkg/diagnostics/docmgrctx/constructors.go` for consistent usage in verbs.

## 4. Rules and Rendering

Rules live in `pkg/diagnostics/docmgrrules` and register in `default.go`:
- Frontmatter syntax (`FrontmatterSyntaxRule`): Points to YAML line/col and suggests `docmgr validate-frontmatter`.
- Frontmatter schema (`FrontmatterSchemaRule`): Highlights missing/invalid fields and suggests `docmgr meta update --field ...`.
- Vocabulary suggestion (`VocabularySuggestionRule`): Lists known values and offers `vocab add` / `vocab list`.
- Related file missing (`RelatedFileMissingRule`): Suggests removing/fixing the path.
- Template parse (`TemplateParseRule`): Surfaces `.templ` parsing errors with file and parser message.
- Listing skip (`ListingSkipRule`): Explains why list/search skipped a doc (usually bad frontmatter).
- Workspace (`WorkspaceRule`): Covers missing index and stale docs.

Renderer: `docmgr.RenderTaxonomy` looks up matches in the default registry, renders text to stderr, and, if a collector is attached, accumulates `RuleResult` objects for JSON (`render.RenderToJSON`).

## 5. CLI Verb Integration

Diagnostics are emitted from verbs and helpers so users see consistent guidance:
- **doctor** (`pkg/commands/doctor.go`): Emits workspace missing index/stale, frontmatter schema (required fields + missing Status/Topics), vocabulary warnings, related file missing, and invalid frontmatter anywhere under the ticket. Supports `--diagnostics-json <path|->` to write rule results for CI while preserving `--fail-on` semantics.
- **list docs / search** (`pkg/commands/list_docs.go`, `search.go`): Emit listing-skip taxonomies when a doc is skipped due to bad frontmatter instead of silently ignoring it.
- **template validate** (`pkg/commands/template_validate.go`): Wraps `.templ` parse errors into template taxonomies so users see parser details.
- **meta update / relate / rename-ticket** (`pkg/commands/meta_update.go`, `relate.go`, `rename_ticket.go`): Wrap frontmatter parse errors into taxonomies for actionable output.
- **workspace discovery** (`internal/workspace/discovery.go`) and **frontmatter parsing** (`internal/documents/frontmatter.go`): Wrap parse errors so callers (doctor, listing) receive taxonomies in error chains.

### QueryDocs-driven diagnostics (unified index-backed commands)

Several core commands are backed by the workspace index + `QueryDocs` (search/list/doctor). In those flows, diagnostics can be produced directly by `QueryDocs` in addition to the “classic” command-level checks:

- **Parse-error visibility (diagnostics)**: invalid-frontmatter docs are excluded from normal results by default, but can be surfaced as structured diagnostics (so users can repair documents without silently losing them).
- **Normalization fallback warnings**: when reverse lookup must use weaker matching (for example, basename/suffix fallback), `QueryDocs` can emit a warning diagnostic to explain why the match happened.

The taxonomy constructors for these live in:

- `pkg/diagnostics/docmgrctx/query_docs.go`

### Example: doctor with JSON output

```bash
# Human + JSON (file)
docmgr doctor --all --stale-after 30 --diagnostics-json diagnostics.json --fail-on warning

# Human + JSON to stdout (useful in CI pipelines)
docmgr doctor --ticket MEN-4242 --diagnostics-json - --fail-on error
```

JSON payload is an array of `RuleResult`:
```json
[
  {
    "Headline": "Unknown vocabulary value for Topics",
    "Body": "File: ttmp/.../index.md\nField: Topics\nValue: \"custom\"\nKnown values: chat, backend\n",
    "Severity": "warning",
    "Actions": [
      { "Label": "Add to vocabulary", "Command": "docmgr", "Args": ["vocab", "add", "--category", "topics", "--slug", "custom"] }
    ]
  }
]
```

## 6. Extending to New Domains

The fastest path to add a domain is captured in `ttmp/2025/12/01/DOCMGR-ERROR-TAXONOMY-diagnostics-taxonomy-unification/playbook/01-how-to-add-a-new-diagnostics-domain.md`, but the essentials are:
1. Add stage/symptom + context and constructor under `pkg/diagnostics/docmgrctx`.
2. Implement a rule under `pkg/diagnostics/docmgrrules` and register it in `default.go`.
3. Emit taxonomies from the producing code path and call `docmgr.RenderTaxonomy`.
4. Add tests (rule snapshot/behavior) and, if needed, a smoke scenario.
5. Run `go test ./pkg/diagnostics/... ./pkg/commands` and relevant scenario scripts.

## 7. Testing and Smoke Coverage

Automated coverage keeps diagnostics stable across verbs:
- Unit tests: `pkg/diagnostics/docmgrrules` (rules), `pkg/diagnostics/docmgr` (renderer JSON collector), plus command tests where applicable.
- Smoke: `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh` exercises vocabulary, related files, listing skips, workspace staleness/missing index, frontmatter parse, template parse, and writes diagnostics JSON.
- Quick checks: `go test ./pkg/diagnostics/... ./pkg/commands` validates rule wiring and command helpers.

## 8. Key File Map

Use this list to jump directly to the implementation:
- Core types: `pkg/diagnostics/core/types.go`
- Taxonomies: `pkg/diagnostics/docmgrctx/*.go`, constructors in `constructors.go`
- Rules and registry: `pkg/diagnostics/docmgrrules/*.go`, `default.go`
- Renderer/adapter: `pkg/diagnostics/docmgr/adapter.go` (+ tests)
- Command wiring: `pkg/commands/doctor.go`, `list_docs.go`, `search.go`, `template_validate.go`, `meta_update.go`, `relate.go`, `rename_ticket.go`
- Helpers: `internal/documents/frontmatter.go`, `internal/workspace/discovery.go`
- Smoke: `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh`
