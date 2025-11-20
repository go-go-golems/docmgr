---
Title: Template Schema Printing Design
Ticket: DOCMGR-OUTPUT-TEMPLATES
Status: draft
Topics:
    - templates
    - cli
    - glaze
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/internal/templates/verb_output.go
      Note: Postfix rendering and FuncMap
    - Path: glazed/pkg/doc/topics/22-templating-helpers.md
      Note: Docs for templating helpers and usage
    - Path: glazed/pkg/formatters/template/template.go
      Note: Template formatter uses func maps
    - Path: glazed/pkg/helpers/templating/templating.go
      Note: Glazed CreateTemplate and TemplateFuncs (sprig included)
ExternalSources: []
Summary: Design for a --print-template-schema flag to output per-verb template data schema
LastUpdated: 2025-11-20T00:00:00Z
---


## Purpose

Add a consistent `--print-template-schema` capability to verbs that support postfix templates, so users can introspect the data available to templates without reading the code.

## Scope

- Verbs: list docs (done), list tickets (done), doctor (done), status, search, tasks list, vocab list, guidelines
- Output: JSON (default) or YAML via `--schema-format json|yaml`
- Mode: Human/Classic commands only (never in Glaze structured mode)

## Requirements

- Emit a stable description of the template data contract for each verb
- Minimal runtime overhead; no changes to actual command behavior unless flag is passed
- Keep implementation centralized and re-usable across verbs

## Data Contract Strategy

We standardize per-verb exported Go structs to represent the template data. These types live in `docmgr/internal/templates/contracts.go` (new) to be re-used by both rendering and schema generation.

Examples (sketch):

```go
// Common envelope already exists via CommonTemplateData
type StatusTemplateData struct {
  TicketsTotal   int
  TicketsStale   int
  DocsTotal      int
  DesignDocs     int
  ReferenceDocs  int
  Playbooks      int
  StaleAfterDays int
  Tickets        []struct {
    Ticket    string
    Title     string
    Status    string
    Stale     bool
    DocsCount int
    Path      string
  }
}

type SearchTemplateData struct {
  Query         string
  File          string
  Dir           string
  Status        string
  DocType       string
  Topics        []string
  Results       []struct {
    Path         string
    Title        string
    Ticket       string
    Snippet      string
    MatchedFiles []string
    MatchedNotes []string
  }
  TotalResults int
}
```

## Schema Generation Approaches

Option A: Reflection-driven JSON Schema (recommended)
- Use a JSON Schema library (e.g., `github.com/invopop/jsonschema` or `github.com/alecthomas/jsonschema`) to derive schema from the Go types.
- Pros: Accurate, low maintenance, supports enums/descriptions via struct tags
- Cons: Adds a dependency

Option B: Hand-rolled reflection to a simplified schema
- Emit a simple, readable schema (type names, fields, nested objects, arrays)
- Pros: Zero new deps, easy to control output
- Cons: Less expressive, more code to maintain

Decision: Start with Option B (simple reflection) for fast iteration; optionally add JSON Schema later.

## CLI Design

- `--print-template-schema` (bool): If set, print schema for the verb’s template data and exit 0
- `--schema-format json|yaml` (string, default json): Choose schema output format

Behavior:
- Command runs normal human output; after that, if `--print-template-schema` is set, print the schema block
- If `--with-glaze-output` is set, skip schema entirely (schema is for human templating only)

## Implementation Outline

1) Centralize types
- Create `docmgr/internal/templates/contracts.go` with exported structs for each verb’s template data
- Reference these types in the verbs when constructing the data passed to `RenderVerbTemplate`

2) Schema helper
- Add `templates/ schema.go`:
  - `func PrintSchema(w io.Writer, v any, format string) error`
  - Reflect on `v` (or a zero instance of the type) to build a simple map-based schema representation
  - Marshal as JSON or YAML

3) Wire per verb
- Add flags:
  - `--print-template-schema`
  - `--schema-format`
- In each `Run`, after human output and before postfix template rendering:
  - If flag set, call `templates.PrintSchema(os.Stdout, StatusTemplateData{}, format)` (passing a zero value of the type or the filled instance)
  - Then continue to postfix templating (or exit early – either is acceptable; prefer printing after the normal output so users see both)

4) Docs & discoverability
- Reference schema printing in the ticket docs and in examples
- Update the examples under `ttmp/templates/` as needed with comments showing `.fields`

## Example Pseudocode

```go
// In status.go
var printTemplateSchema bool
var schemaFormat string

func init() {
  // cobra flag setup
  // flags.BoolVar(&printTemplateSchema, "print-template-schema", false, "Print template schema after output")
  // flags.StringVar(&schemaFormat, "schema-format", "json", "Schema format: json|yaml")
}

func (c *StatusCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
  // ... existing human output code ...

  // Build template data
  data := StatusTemplateData{ /* filled from counters */ }

  if printTemplateSchema {
    _ = templates.PrintSchema(os.Stdout, data, schemaFormat)
  }

  // Render postfix template
  _ = templates.RenderVerbTemplate([][]string{{"status"}}, settings.Root, settingsMap, data)
  return nil
}
```

## Risks & Mitigations

- Drift between constructed maps and exported types
  - Mitigation: Use the exported types in code directly when building template data
- Verb-specific differences
  - Mitigation: Keep the schemas per-verb; centralize the common envelope in `CommonTemplateData`

## Next Steps

- [ ] Add `contracts.go` with exported template data structs
- [ ] Implement `templates.PrintSchema()` (simple reflection to JSON/YAML)
- [ ] Add flags to templated verbs and wire `PrintSchema` after human output
- [ ] Document usage in ticket reference and examples
