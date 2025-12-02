---
Title: YAPP error-feedback port to docmgr
Ticket: DOCMGR-ERROR-TAXONOMY
Status: active
Topics:
    - yaml
    - ux
    - errors
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Port YAPP-style taxonomy, rule registry, and CLI rendering into docmgr to give structured, multi-domain validation guidance (YAML, schema, vocab, links, tasks) and generalize it for any program."
LastUpdated: 2025-12-01T11:56:30-05:00
---

# YAPP error-feedback port to docmgr

## What exists in YAPP

YAPP ships a layered error-feedback stack (see `YAPP_Box/examples/yapp/error-feedback/DEVELOPER_GUIDE.md`):
- Taxonomy (`pkg/resolver/errorx/`): typed `Taxonomy{Stage,Symptom,Path,Severity,Context}` with constructors like `NewYAMLIngestTaxonomy` and `NewSchemaConstraintTaxonomy`. Context structs (e.g., `YAMLIngestContext`, `SchemaStructureContext`) carry typed payloads (file, line, column, expected vs. actual).
- Rule registry (`pkg/resolver/rules/`): rules implement `Renderer` with `Match(*Taxonomy)` → `(bool, score)` and `Render(ctx,*Taxonomy)` → `RuleResult{Headline,Body,Severity,Actions}`. Default registry registers multiple rules and sorts by score.
- CLI integration (`cmd/yappctl/resolve_command.go`): unwraps taxonomy via `errorx.AsTaxonomy(err)`, then `reg.RenderAll` to produce multi-card guidance (text/JSON/HTML renderers are pluggable).

Key strengths worth porting:
- Strong typing for error contexts, not just string matching, enabling stage-specific messaging.
- Rule scoring to surface the most helpful guidance first.
- Separation between error production (taxonomy) and presentation (rules/renderers).

## Porting approach for docmgr (beyond YAML)

Goal: reuse the YAPP pattern so docmgr can emit structured, user-friendly guidance for YAML parsing plus other validation domains (frontmatter schema, vocabulary, file references, tasks/changelog consistency).

### 1) Introduce a docmgr taxonomy layer

Proposed package: `pkg/diagnostics/taxonomy` (mirrors `errorx`).
- Core type:
  ```go
  type Taxonomy struct {
      Stage    StageCode    // e.g., StageFrontmatterParse, StageFrontmatterSchema, StageDocLink, StageTaskValidation
      Symptom  SymptomCode  // e.g., SymptomSyntax, SymptomMissingRequired, SymptomUnknownTopic, SymptomMissingFile
      Path     string       // path in doc or field path (Summary, Topics[0], RelatedFiles[1].Path)
      Severity Severity     // error|warning|info
      Context  any          // typed contexts below
  }
  ```
- Context structs:
  - `FrontmatterParseContext{File string, Line, Column int, Snippet string, Underlying error}`
  - `FrontmatterSchemaContext{File string, Field string, Expected string, Actual string, Hint string}`
  - `VocabularyContext{File string, Field string, Value string, Known []string}`
  - `RelatedFileContext{DocPath string, FilePath string, Exists bool, Note string}`
  - `TasksContext{DocPath string, MissingIDs []int, ParseErr error}`
- Constructors file `pkg/diagnostics/taxonomy/constructors.go` akin to YAPP’s `constructors.go` to keep call sites terse.
- Helper `AsTaxonomy(err)` that unwraps wrapped errors so CLI can detect taxonomy-bearing failures.

### 2) Rule registry for docmgr

Package: `pkg/diagnostics/rules` (mirrors `pkg/resolver/rules`).
- Interface:
  ```go
  type Renderer interface {
      Match(*taxonomy.Taxonomy) (bool, int)
      Render(context.Context, *taxonomy.Taxonomy) (*RuleResult, error)
  }

  type RuleResult struct {
      Headline string
      Body     string // markdown/text
      Severity taxonomy.Severity
      Actions  []Action // label + command/args suggestions
  }
  ```
- Default registry in `pkg/diagnostics/rules/default.go` registering rules by stage:
  - YAML syntax pointer rule (reuses snippet highlighting from `pkg/frontmatter` design: `extractSnippet`, line/column).
  - Schema missing/invalid rule (maps to `FrontmatterSchemaContext`, suggests `docmgr meta update --field Summary ...`).
  - Vocabulary suggestion rule (unknown topic/status → suggest `docmgr vocab add` or closest match).
  - RelatedFiles existence rule (suggest running `ls` or fixing path).
  - Tasks/changelog lint rule (e.g., missing changelog entry for status change).
- Score bands reused (100 critical syntax, 80 schema, 70 vocabulary, 60 related-files, etc.).

### 3) CLI integration

Touchpoints:
- `cmd/docmgr/validate_frontmatter.go` (new command from design doc): wrap parse/validate errors into taxonomy before returning.
- `cmd/docmgr/meta_update.go`, `doc_add.go`, `changelog_update.go`: when validation fails, emit taxonomy-bearing errors instead of plain `fmt.Errorf`.
- Shared renderer helper in `cmd/docmgr/internal/diagnostics/render.go`:
  ```go
  func renderDiagnostics(ctx context.Context, err error) error {
      if t, ok := taxonomy.AsTaxonomy(err); ok {
          reg := rules.DefaultRegistry()
          results, _ := reg.RenderAll(ctx, t)
          fmt.Fprintln(os.Stderr, rules.RenderToText(results))
      }
      return err
  }
  ```
- Commands call `return renderDiagnostics(cmd.Context(), err)` so existing error flows stay intact but gain rich output.
- Output formats: start with text; optionally add `--with-glaze-output --output json` to emit `[]RuleResult` for CI (mirrors YAPP JSON renderer idea).

### 4) Wiring YAML improvements into taxonomy

Bridge the design doc’s `frontmatter.ParseError` to taxonomy:
- In `pkg/frontmatter/errors.go`, after constructing `ParseError{File,Field,Line,Column,Value,...}`, wrap with `taxonomy.NewFrontmatterParseTaxonomy(...)`.
- Rule `YamlSyntaxPointerRule` renders:
  - Headline: `YAML syntax at Summary (line X col Y)`
  - Body: snippet from `extractSnippet`, problem classification (unquoted colon, mapping value, etc.), fix options already computed in `ParseError.Fixes`.
  - Actions: `docmgr validate-frontmatter <file> --suggest-fixes`, `docmgr validate-frontmatter <file> --auto-fix`.

### 5) Extending beyond YAML

Examples of docmgr-specific rules and contexts:
- **Schema missing required** (FrontmatterSchemaContext): show expected type/requirements for `Title`, `Summary`, `Topics`. Action: `docmgr meta update --doc <path> --field Summary --value '...'`.
- **Vocabulary unknown** (VocabularyContext): suggest `docmgr vocab add --category topics --slug <value>` or suggest closest known slug.
- **Related file missing** (RelatedFileContext): check `os.Stat`; rule suggests updating `RelatedFiles` note or running `docmgr doc relate --remove-files`.
- **Tasks incomplete** (TasksContext): report invalid IDs or mismatch between tasks and changelog; suggest `docmgr task list`/`docmgr changelog update`.
- **Invalid status transitions**: add a StageStatusValidation for future workflows.

### 6) Pseudocode end-to-end flow (generalized)

```go
func validateFrontmatter(ctx context.Context, path string) error {
    raw, err := os.ReadFile(path)
    if err != nil { return err }

    fm, err := frontmatter.Parse(raw)
    if err != nil {
        if pe, ok := err.(*frontmatter.ParseError); ok {
            return taxonomy.NewFrontmatterParseTaxonomy(
                path, pe.Field, pe.Line, pe.Column, pe.Value, pe.Original)
        }
        return err
    }

    if warns := frontmatter.ValidateSchema(fm); len(warns) > 0 {
        return taxonomy.NewFrontmatterSchemaTaxonomy(path, warns)
    }

    return nil
}

// CLI layer
if err := validateFrontmatter(ctx, file); err != nil {
    return renderDiagnostics(ctx, err)
}
```

### 7) File/package placement (docmgr)

- `pkg/diagnostics/taxonomy/` — types, contexts, constructors, helpers.
- `pkg/diagnostics/rules/` — registry, rule implementations, text renderer (future JSON renderer).
- `pkg/diagnostics/snippet/snippet.go` — shared snippet extraction/highlighting (reused by YAML rules).
- `cmd/docmgr/internal/diagnostics/render.go` — command-side adapter.
- Integrate in:
  - `pkg/frontmatter` parsing (StageFrontmatterParse)
  - `pkg/frontmatter/validate.go` (StageFrontmatterSchema, StageVocabulary)
  - `pkg/doc/related_files.go` (StageDocLink)
  - `pkg/tasks`/`pkg/changelog` validators (StageTaskValidation/StageChangelogValidation)

### 8) Migration notes

- Start with YAML parsing (highest impact), then add schema/vocabulary and RelatedFiles rules.
- Keep original error values wrapped so logging/telemetry can still access raw errors.
- Make rules tolerant: if context casting fails, return error so CLI falls back to plain message (mirrors YAPP best practice).
- Add unit tests mirroring YAPP style: `Match` and `Render` per rule plus a renderer snapshot test for CLI text.

## How this helps DOCMGR-YAML-001

- Directly addresses Goal #1 (better errors) by layering rule-driven messaging over `ParseError`.
- Supports Goal #3/4 (validation and auto-fix) by providing actionable suggestions and commands in the rendered cards.
- Extends to future validations (vocabulary, RelatedFiles, tasks) without redesign: new contexts + rules only.
- Encourages consistent CLI UX (text + JSON outputs) similar to YAPP so automation/IDE hooks can highlight precise locations with hints.

## Unifying diagnostics systems (docmgr + YAPP + others)

We can lift YAPP’s pattern into a generic diagnostics SDK so any Go CLI (docmgr, yappctl, future tools) shares the same interface and renderers.

- Core package: `pkg/diagnostics/core/` (shared library).
  - `types.go`:
    ```go
    type StageCode string   // tool-defined
    type SymptomCode string // tool-defined
    type Severity string    // error|warning|info

    type Taxonomy struct {
        Tool     string      // "docmgr", "yappctl", etc.
        Stage    StageCode
        Symptom  SymptomCode
        Path     string      // logical path (field/module)
        Severity Severity
        Context  any         // typed per tool
        Cause    error       // original error (optional)
    }

    type ContextualError interface {
        AsTaxonomy() (*Taxonomy, bool)
    }
    ```
  - `wrap.go`: helpers `WrapTaxonomy(err error, t *Taxonomy) error` and `AsTaxonomy(err error) (*Taxonomy, bool)` to let each tool wrap its own errors with `errors.Join`/custom types.
  - `context.go`: marker interface `ContextPayload` to enforce typed contexts per tool.
  - The tool-specific packages (docmgr: `pkg/diagnostics/taxonomy`, YAPP: `pkg/resolver/errorx`) embed or alias `core.Taxonomy` and their own Stage/Symptom enums; this preserves separation while enabling shared renderers and registries.

- Renderer package: `pkg/diagnostics/render/`
  - Text renderer shared across tools: `RenderToText([]*rules.RuleResult) string`.
  - JSON renderer: `RenderToJSON([]*rules.RuleResult) ([]byte, error)` for CI/IDE.
  - HTML (optional, only if needed by a tool).

- Rule registry package: `pkg/diagnostics/rules/` (tool-agnostic shell)
  - Interfaces (below) live here; each tool can add modules that register rules via `Registry.Register` even from its own package paths.
  - Scoring and sorting implemented once; no code duplication.

Adapter strategy:
- Docmgr defines contexts/types in `pkg/diagnostics/docmgrctx` and registers its rules via the shared registry.
- YAPP retains its contexts but switches to the shared registry + renderer; its existing `Renderer` impls just import `core` and `rules`.
- Any new program can define Stage/Symptom/Context types, wrap errors via `WrapTaxonomy`, and immediately get text/JSON rendering with shared look-and-feel.

## Rule-system deep dive for helpful validation output

Shared rule contracts (lives in `pkg/diagnostics/rules/rule.go`):
```go
type Rule interface {
    Match(*core.Taxonomy) (bool, int) // score: 0-100
    Render(context.Context, *core.Taxonomy) (*RuleResult, error)
}

type RuleResult struct {
    Headline string
    Body     string // markdown/text
    Severity core.Severity
    Actions  []Action
}

type Action struct {
    Label   string
    Command string
    Args    []string
}
```

Registry (`registry.go`):
```go
type Registry struct { rules []Rule }

func (r *Registry) Register(rule Rule)               { r.rules = append(r.rules, rule) }
func (r *Registry) RenderAll(ctx context.Context, t *core.Taxonomy) ([]*RuleResult, error) {
    type scored struct {
        res  *RuleResult
        score int
    }
    var out []scored
    for _, rule := range r.rules {
        if ok, score := rule.Match(t); ok {
            res, err := rule.Render(ctx, t)
            if err != nil { return nil, err }
            out = append(out, scored{res: res, score: score})
        }
    }
    sort.Slice(out, func(i, j int) bool { return out[i].score > out[j].score })
    results := make([]*RuleResult, len(out))
    for i := range out { results[i] = out[i].res }
    return results, nil
}
```

Rendering flow for validation errors (tool code, e.g., `cmd/docmgr/internal/diagnostics/render.go`):
```go
func HandleError(ctx context.Context, err error) error {
    if t, ok := core.AsTaxonomy(err); ok {
        reg := rules.DefaultRegistry() // tool registers defaults at init
        results, rerr := reg.RenderAll(ctx, t)
        if rerr == nil {
            fmt.Fprintln(os.Stderr, render.RenderToText(results))
        }
    }
    return err // preserve original exit code/behavior
}
```

Rule authoring guidance (applies to docmgr validations):
- **YAML syntax pointer rule**: Match `StageFrontmatterParse` + `SymptomSyntax`; render headline “YAML syntax at <Field> (line X col Y)” with snippet and fix options from context; Actions: `docmgr validate-frontmatter <file> --auto-fix`.
- **Schema missing/invalid rule**: Match `StageFrontmatterSchema`; use context listing missing/invalid fields; Body lists each field with expected vs. actual; Actions: `docmgr meta update --doc <file> --field Field --value "..."`.
- **Vocabulary unknown rule**: Match `StageVocabulary`; Body shows unknown slug and nearest matches; Actions: `docmgr vocab add --category topics --slug <slug>`.
- **Related file missing rule**: Match `StageDocLink`; Body shows missing path and note; Actions: `docmgr doc relate --remove-files <path>` or fix path.
- **Tasks/changelog rule**: Match `StageTaskValidation`; Body shows stale IDs or missing changelog; Actions: `docmgr task list`, `docmgr changelog update --entry "..."`

Printing helpful information checklist (applies to all validation errors):
- Always include: file/doc path, stage, symptom, field/path, line/column (when available).
- Provide a short headline plus detailed body with Markdown bullets for fixes.
- Suggest at least one actionable command (copy-pastable) in `Actions`.
- Use severity to control emphasis; registry sorting by score ensures the most relevant card is first.
- Ensure `Render` gracefully handles unexpected context shapes (return error to fall back to plain error).

Open questions for unification:
- Where to host the shared SDK (`pkg/diagnostics/...`)—inside docmgr repo as a module importable by others, or a separate module consumed by both docmgr and YAPP?
- Versioning strategy for shared interfaces to avoid breaking existing YAPP rule implementations.

### Context interface (avoid `any`)

Add a minimal contract so contexts stay typed but generic:
```go
// pkg/diagnostics/core/context.go
type ContextPayload interface {
    Stage() StageCode    // optional: helps renderers detect mismatches
    Summary() string     // fallback if renderer cannot cast concrete type
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
    if t.Context == nil {
        return ""
    }
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

Renderer guidance:
- Try concrete cast first for rich details; if it fails, fall back to `ContextSummary()` so the user still sees location/problem.
- This replaces raw `any`, keeps contexts discoverable, and avoids renderer panics when a new context type arrives.
