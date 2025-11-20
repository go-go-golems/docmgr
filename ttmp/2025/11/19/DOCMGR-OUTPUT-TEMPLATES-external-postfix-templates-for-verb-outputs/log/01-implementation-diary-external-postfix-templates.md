---
Title: 'Implementation Diary: External Postfix Templates'
Ticket: DOCMGR-OUTPUT-TEMPLATES
Status: active
Topics:
    - cli
    - templates
    - glaze
DocType: log
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/internal/templates/verb_output.go
      Note: Core template rendering infrastructure with path resolution and FuncMap helpers
    - Path: docmgr/pkg/commands/doctor.go
      Note: Added template data struct building and postfix template rendering
    - Path: docmgr/pkg/commands/list_docs.go
      Note: Added template data struct building and postfix template rendering after human output
    - Path: docmgr/pkg/commands/list_tickets.go
      Note: Added template data struct building and postfix template rendering after human output
    - Path: docmgr/ttmp/templates/doc/list.templ
      Note: Example template demonstrating LLM-oriented summary with doc counts
    - Path: docmgr/ttmp/templates/doctor.templ
      Note: Example template demonstrating doctor summary with findings grouped by severity
    - Path: docmgr/ttmp/templates/list/tickets.templ
      Note: Example template demonstrating ticket overview with status aggregation using dict/set helpers
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T21:00:00-05:00
---


# Implementation Diary: External Postfix Templates

Implementation diary for DOCMGR-OUTPUT-TEMPLATES tracking the journey of implementing external postfix templates for verb outputs.

---

## 2025-11-19 - Implementation Complete

### What I Did

**1. Created Template Rendering Infrastructure**
- Implemented `docmgr/internal/templates/verb_output.go` with core rendering logic
- Created `RenderVerbTemplate` function that handles template path resolution, reading, parsing, and rendering
- Implemented `resolveTemplatePath` to compute canonical template paths:
  - Grouped verbs: `templates/$group/$verb.templ` (e.g., `templates/doc/list.templ`)
  - Single verbs: `templates/$verb.templ` (e.g., `templates/doctor.templ`)
- Built common template data envelope with `Verbs`, `Root`, `Now`, and `Settings` fields
- Created safe template FuncMap with helpers: `slice`, `dict`, `set`, `get`, `add1`, `countBy`

**2. Updated Three Verbs**
- **`list_docs`** (`pkg/commands/list_docs.go`):
  - Built template data struct with `TotalDocs`, `TotalTickets`, `Tickets` (nested `Docs`), `Rows`, `Fields`
  - Added template rendering after human output completes
  - Supports both `templates/doc/list.templ` and `templates/list/docs.templ` paths
  
- **`list_tickets`** (`pkg/commands/list_tickets.go`):
  - Built template data struct with `TotalTickets`, `Tickets`, `Rows`, `Fields`
  - Added template rendering after human output completes
  - Uses `templates/list/tickets.templ` path

- **`doctor`** (`pkg/commands/doctor.go`):
  - Built template data struct with `TotalFindings` and `Tickets` (nested `Findings`)
  - Added template rendering after human output completes
  - Uses `templates/doctor.templ` path
  - Fixed settings parsing issue (settings were not being parsed in `Run` method)

**3. Created Example Templates**
- `ttmp/templates/doc/list.templ` - LLM-oriented summary with doc counts and top docs per ticket
- `ttmp/templates/list/tickets.templ` - Ticket overview with status counts using dict/set helpers
- `ttmp/templates/doctor.templ` - Doctor summary with findings grouped by severity

### What Worked Well

1. **Template Path Resolution**: The canonical path approach worked cleanly - each verb has a single, predictable template location. The implementation handles both grouped and single-level verbs correctly.

2. **FuncMap Helpers**: The template helpers (`dict`, `set`, `get`, `slice`, `countBy`) provide enough power for common operations while staying safe and deterministic. The `countBy` function with reflection support handles both maps and structs nicely.

3. **Non-Fatal Error Handling**: Printing warnings to stderr and continuing execution means templates are truly optional - if a template is missing or malformed, the command still works normally.

4. **Data Structure Design**: Building typed structs in each verb's `Run` method gives us full control over what data is available to templates, including derived/computed fields that aren't in the Glaze rows.

5. **Integration with Existing Code**: The implementation fits cleanly into the existing dual-mode architecture - templates only render in classic/human mode, never interfering with Glaze output.

### What Didn't Work / Challenges

1. **Initial Template Path Resolution Bug**: First implementation didn't properly handle the "docmgr" root command in the verb path. Fixed by stripping "docmgr" before computing paths.

2. **Settings Parsing in Doctor**: Initially forgot to parse settings in the `Run` method (they were only parsed in `RunIntoGlazeProcessor`). This caused compilation errors when trying to access `settings.Root` and other fields.

3. **countBy Function Complexity**: The `countBy` helper needed to handle multiple data types (maps, structs, slices). Used reflection for struct field access, which works but adds complexity. Considered making it more specific to the `Severity` field but kept it generic for flexibility.

4. **Template Data Merging**: Initially tried to merge verb-specific data directly into the common envelope, but Go templates work better when data is structured as a map. Ended up merging maps and also supporting struct data via a `Data` field.

5. **Verb Path Candidates**: For commands like `list_docs` that can be invoked as both `docmgr doc list` and `docmgr list docs`, needed to support multiple candidate paths. The current implementation tries candidates in order until one exists.

### What I Learned

1. **Go Template Reflection**: Learned how to use `reflect` package to access struct fields dynamically in template helpers. This enables generic helpers that work with both maps and structs.

2. **Template FuncMap Design**: Keeping FuncMaps minimal and deterministic is important for security and predictability. Each helper should have a clear, single purpose.

3. **Path Resolution Patterns**: The canonical path approach (single path per verb, no fallbacks) simplifies the mental model and makes templates easier to discover. Users know exactly where to put templates.

4. **Error Handling Philosophy**: Non-fatal template errors fit the use case perfectly - templates are enhancements, not requirements. Commands should work fine without them.

5. **Data Structure Flexibility**: Building template data structures separately from Glaze rows gives flexibility to include computed/derived fields that make sense for LLM consumption but aren't needed for structured output.

### What I Should Do Better in the Future

1. **Test Earlier**: Should have tested template rendering earlier in the process, especially the path resolution logic. Caught the "docmgr" root stripping issue during testing rather than during implementation.

2. **More Comprehensive FuncMap Testing**: The `countBy` function with reflection is complex - should test it with various data types to ensure it works correctly in all cases. Consider adding unit tests for template helpers.

3. **Template Validation**: Consider adding a `docmgr template validate` command to check template syntax before runtime. This would catch template errors earlier.

4. **Documentation**: Should document the template data contracts more explicitly - what fields are available in each verb's template context. This helps users write effective templates.

5. **Template Examples**: The example templates are minimal. Should create more comprehensive examples showing advanced patterns (nested loops, conditionals, complex data transformations).

6. **Path Resolution Edge Cases**: Consider edge cases like:
   - What if a template exists for both `templates/doc/list.templ` and `templates/list/docs.templ`? (Currently uses first found)
   - Should we support template inheritance or composition?
   - What about template directories vs single files?

7. **Performance Considerations**: Template rendering happens synchronously after output. For large datasets, consider if async rendering or caching would be beneficial.

8. **Template Debugging**: Add a `--debug-template` flag that shows:
   - Which template path was resolved
   - What data is being passed to the template
   - Template parsing/rendering errors with more context

### Files Changed

- `docmgr/internal/templates/verb_output.go` - Core template rendering infrastructure
- `docmgr/pkg/commands/list_docs.go` - Added template data building and rendering
- `docmgr/pkg/commands/list_tickets.go` - Added template data building and rendering  
- `docmgr/pkg/commands/doctor.go` - Added template data building and rendering, fixed settings parsing
- `docmgr/ttmp/templates/doc/list.templ` - Example template for doc list
- `docmgr/ttmp/templates/list/tickets.templ` - Example template for list tickets
- `docmgr/ttmp/templates/doctor.templ` - Example template for doctor

### Next Steps

- Monitor usage and gather feedback on template patterns
- Consider adding more verbs (status, search, etc.)
- Evaluate need for template composition/inheritance
- Add template validation tooling
- Document template data contracts more thoroughly
