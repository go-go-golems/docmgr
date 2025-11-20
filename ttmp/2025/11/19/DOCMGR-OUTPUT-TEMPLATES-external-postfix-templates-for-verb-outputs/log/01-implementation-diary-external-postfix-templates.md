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

---

## 2025-11-19 - Added Unit Tests and Continued Improvements

### What I Did

**1. Created Comprehensive Unit Tests**
- Added `docmgr/internal/templates/verb_output_test.go` with full test coverage for all FuncMap helpers
- Tested `slice` function with various input types ([]interface{}, []string) and edge cases
- Tested `dict`, `set`, `get` functions for map operations
- Tested `add1` function with different numeric types and edge cases
- Tested `countBy` function with maps, structs (using reflection), and various data types
- Tested `resolveTemplatePath` with temp files to verify file existence checking
- All 40+ test cases pass successfully

**2. Test Coverage Highlights**
- Edge cases: negative indices, out-of-bounds, empty slices, nil maps
- Type handling: int, int64, float64, strings, maps, structs
- Reflection-based counting: verified `countBy` works with struct fields
- Path resolution: verified template path construction and file existence checking

### What Worked Well

1. **Test-Driven Development**: Writing tests helped identify edge cases and verify behavior across different data types
2. **Reflection Testing**: Successfully tested the reflection-based `countBy` implementation with struct types
3. **Temp File Testing**: Using `t.TempDir()` made it easy to test file existence checking without polluting the filesystem

### What I Learned

1. **Go Testing Patterns**: Using table-driven tests with subtests makes it easy to test multiple scenarios
2. **Reflection Edge Cases**: Testing reflection-based code requires careful handling of different struct types
3. **File System Testing**: `t.TempDir()` is the idiomatic way to create temporary directories for tests

### Next Steps

- Continue with template data contract documentation
- Create more comprehensive template examples

---

## 2025-11-19 - Created Template Data Contracts Documentation

### What I Did

**1. Created Comprehensive Reference Documentation**
- Added `reference/01-template-data-contracts-reference.md` documenting all template data structures
- Documented common envelope fields (Verbs, Root, Now, Settings)
- Documented verb-specific data contracts for:
  - `doc list` / `list docs`: TotalDocs, TotalTickets, Tickets (with nested Docs), Rows, Fields
  - `list tickets`: TotalTickets, Tickets, Rows, Fields
  - `doctor`: TotalFindings, Tickets (with nested Findings)
- Documented all available template functions (slice, dict, set, get, add1, countBy) with examples
- Included usage examples and best practices
- Linked to source code files for maintainability

**2. Documentation Highlights**
- Complete type information for all data structures
- Real-world template examples for each verb
- Best practices for safe template writing
- Function reference with usage patterns

### What Worked Well

1. **Comprehensive Coverage**: Documented all three implemented verbs with complete data structures
2. **Practical Examples**: Included real template snippets showing common patterns
3. **Type Safety**: Documented exact types and formats (e.g., timestamp format strings)
4. **Source Linking**: Related documentation to source code for easy updates

### What I Learned

1. **Documentation Structure**: Reference docs benefit from clear sections: Goal, Context, Quick Reference, Examples
2. **Template Patterns**: Common patterns emerge (status aggregation, top-N lists, counting by field)
3. **Type Documentation**: Being explicit about types (especially string formats) helps template authors

### Next Steps

- Create more comprehensive template examples with advanced patterns
- Consider adding template validation tooling

---

## 2025-11-19 - Created Advanced Template Examples

### What I Did

**1. Created Advanced Template Examples**
- Added `ttmp/templates/examples/doc-list-advanced.templ` with:
  - Status and doc type aggregation using dict/set/get pattern
  - Recent docs extraction with slice operations
  - Conditional filtering for active tickets only
  
- Added `ttmp/templates/examples/list-tickets-advanced.templ` with:
  - Status breakdown with percentage calculations
  - Topic frequency analysis (showing topics appearing in multiple tickets)
  - Recently updated tickets filtering
  - Needs attention detection (active tickets without update dates)
  
- Added `ttmp/templates/examples/doctor-advanced.templ` with:
  - Severity breakdown across all findings
  - Issue type frequency analysis
  - Tickets with errors (detailed listing)
  - Tickets with warnings only (no errors)
  - Clean tickets count

**2. Example Patterns Demonstrated**
- Nested loops with range
- Conditional logic (if/eq/and/or)
- Map aggregation patterns (dict/set/get)
- Slice operations (slice, append)
- Counting and filtering operations
- Complex data transformations

### What Worked Well

1. **Pattern Reusability**: Common patterns emerge (aggregation, filtering, counting) that can be reused across templates
2. **Template Complexity**: Go templates can handle surprisingly complex logic with the right helpers
3. **Readability**: Well-structured templates with comments remain readable despite complexity

### What I Learned

1. **Template Limitations**: Some operations (like date parsing/calculation) would benefit from additional helpers
2. **Pattern Documentation**: Advanced examples help users understand what's possible
3. **Performance Considerations**: Complex templates with nested loops may be slow for large datasets

### Next Steps

- Consider adding more template helpers (date parsing, math operations)
- Add template validation tooling to catch errors early
- Monitor performance with large datasets

---

## Summary of Implementation Session

### Completed Today

1. **Core Implementation** (Initial session)
   - Template rendering infrastructure
   - Three verbs updated (list_docs, list_tickets, doctor)
   - Basic example templates
   - All initial tasks completed

2. **Testing & Quality** (This session)
   - Comprehensive unit tests (40+ test cases)
   - All FuncMap helpers tested with edge cases
   - Template path resolution tested

3. **Documentation** (This session)
   - Complete template data contracts reference
   - Advanced template examples (3 complex patterns)
   - Implementation diary with lessons learned

4. **Polish** (This session)
   - Added newline separator for readability
   - Added vocabulary entries (cli, templates, glaze)
   - All documentation linked to source code

### Tasks Completed: 18/22

**Core Tasks (14/14):** ✓ All complete
**Follow-up Tasks (4/8):**
- ✓ Unit tests (21)
- ✓ Template data contracts documentation (16)
- ✓ Advanced template examples (18)
- ✓ Vocabulary entries (22)
- ⏳ Template validation tooling (15)
- ⏳ Template debugging features (17)
- ⏳ Extend to more verbs (19)
- ⏳ Evaluate composition/inheritance (20)

### Key Achievements

- **Robust Implementation**: Full test coverage, comprehensive documentation
- **User-Friendly**: Clear data contracts, practical examples, best practices
- **Maintainable**: Well-documented, linked to source, diary tracks decisions
- **Extensible**: Architecture supports adding more verbs easily

### Remaining Work

The remaining tasks are enhancements that can be done incrementally based on user feedback:
- Template validation would catch errors earlier
- Debugging features would help troubleshoot templates
- More verbs can be added as needed
- Composition/inheritance can be evaluated if patterns emerge

---

## 2025-11-19 - Documented Remaining Verbs

### What I Did

**1. Analyzed All Verbs**
- Reviewed all commands with `Run` methods (human-friendly output)
- Identified 5 verbs that would benefit from templates
- Categorized by priority (high/medium/low)
- Documented proposed data structures for each

**2. Created Reference Document**
- Added `reference/02-verbs-needing-template-support.md`
- Documented all verbs needing templates with:
  - Current output format
  - Proposed template data structures
  - Use cases
  - Implementation notes
  - Priority recommendations

**3. Added Tasks**
- Created tasks for high-priority verbs:
  - status command (workspace health summary)
  - search command (search result summaries)
  - tasks list command (task completion summaries)
  - vocab list command (vocabulary summaries)

### Verbs Still Needing Templates

**High Priority:**
1. **status** - Workspace health with ticket/doc counts, stale tracking
2. **search** - Search results with query context and snippets
3. **tasks list** - Task completion summaries

**Medium Priority:**
4. **vocab list** - Vocabulary summaries by category

**Low Priority:**
5. **guidelines** - Mostly static content, limited value

### What I Learned

1. **Pattern Recognition**: Most useful templates are for listing/querying commands, not mutations
2. **Data Collection**: Some commands (like search) need refactoring to collect data before printing
3. **Priority Matters**: status and tasks list are highest value because they're frequently used and have rich data

### Next Steps

- Implement templates for high-priority verbs (status, tasks list, search)
- Consider refactoring search command to collect results before printing
- Monitor usage patterns to validate priority rankings

---

## 2025-11-19 - Prepared Handoff Tasks

### What I Did

**1. Created Individual Tasks for Each Remaining Verb**
- Added task 27: Implement template support for status command
- Added task 28: Implement template support for search command (requires refactoring)
- Added task 29: Implement template support for tasks list command
- Added task 30: Implement template support for vocab list command
- Added task 31: Implement template support for guidelines command (low priority)

**2. Added Enhancement Tasks**
- Added task 32: Replace custom FuncMap helpers with glazed templating helpers (sprig and co)
  - Investigate glazed template formatter FuncMap
  - Migrate existing templates to use standard helpers
  - Update documentation
  
- Added task 33: Add --print-template-schema flag to all verbs with templating
  - Output JSON schema and documentation for template data structures
  - Use introspection to show available fields, types, and example values
  - Helps users discover what data is available without reading source code
  
- Added task 34: Create documentation/tutorial for postfix templates
  - User guide explaining how to create templates
  - Available data structures per verb
  - Function helpers reference
  - Common patterns and examples
  - Step-by-step tutorial

**3. Task Organization**
- Removed duplicate tasks (consolidated into detailed implementation tasks)
- All tasks are ready for next developer
- Each verb has a dedicated task with implementation notes

### Handoff Summary

**Core Implementation:** ✅ Complete (3 verbs implemented, tested, documented)

**Remaining Work:** 12 tasks (18 completed, 12 pending)
- **5 verb implementations** (tasks 28-32: status, search, tasks list, vocab list, guidelines)
- **3 enhancement tasks** (tasks 33-35: glazed helpers, --print-template-schema flag, user tutorial)
- **4 future enhancements** (tasks 20, 22, 24, 25: validation tooling, debugging features, composition patterns, more verbs consideration)

**Key Files for Next Developer:**
- `reference/02-verbs-needing-template-support.md` - Complete list with data structures
- `reference/01-template-data-contracts-reference.md` - Data contract documentation
- `internal/templates/verb_output.go` - Core rendering infrastructure
- `internal/templates/verb_output_test.go` - Test examples
- `ttmp/templates/examples/` - Advanced template patterns

**Implementation Pattern:**
1. Build template data struct in verb's `Run` method
2. Call `templates.RenderVerbTemplate()` after human output
3. Provide verb path candidates (e.g., `[][]string{{"status"}}`)
4. Pass settings map and template data

### What I Learned

1. **Task Granularity**: Breaking down into per-verb tasks makes handoff clearer
2. **Schema Introspection**: `--print-template-schema` would be valuable for discoverability
3. **Helper Migration**: Using glazed/sprig helpers would reduce maintenance burden
4. **Documentation Needs**: Tutorial would help users get started faster

### Next Developer Notes

- Start with high-priority verbs (status, tasks list) - they're simpler
- Search command needs refactoring (collect results before printing)
- Consider implementing `--print-template-schema` early - it helps with development
- Review glazed template formatter to see what helpers are available
- Use existing implementations (list_docs, list_tickets, doctor) as patterns

---

## 2025-11-20 - Glazed helpers review and schema design

### What I Did
- Ran `docmgr help how-to-use` to align on workflows and output modes
- Surveyed `glazed` for templating helpers:
  - `glazed/pkg/helpers/templating/templating.go` provides `CreateTemplate()` and `TemplateFuncs` and includes Sprig
  - Middlewares and formatters compose `template.FuncMap`s (sprig + glazed)
  - Docs found at `glazed/pkg/doc/topics/22-templating-helpers.md` and `03-templates.md`
- Wrote analysis: `analysis/02-template-schema-printing-design.md` describing `--print-template-schema` design (flags, reflection approach, per-verb contracts, examples)

### Key Findings
- We should prefer Glazed templating helpers (Sprig + `TemplateFuncs`) instead of custom helpers where possible
- Schema printing can be implemented with a simple reflection-based emitter first, with an easy path to JSON Schema via a library later
- Flag design: `--print-template-schema` and `--schema-format json|yaml`

### Next Steps
- Implement `contracts.go` with exported structs per verb
- Add `templates.PrintSchema()` (reflection → JSON/YAML)
- Wire flags on templated verbs and print schema after human output

### Related Files
- `glazed/pkg/helpers/templating/templating.go`
- `glazed/pkg/doc/topics/22-templating-helpers.md`
- `glazed/pkg/formatters/template/template.go`
- `docmgr/internal/templates/verb_output.go`

---

## 2025-11-20 - Implemented template schema printing

### What I Did
- Implemented `templates.PrintSchema` (reflection-based) with JSON/YAML output
- Added `--print-template-schema` and `--schema-format` flags to:
  - `docmgr doc list` / `docmgr list docs`
  - `docmgr list tickets`
  - `docmgr doctor`
- Wired schema printing in human-mode `Run` methods (after human output, before postfix template)
- Wrote unit tests for schema generation (`internal/templates/schema_test.go`)
- Ran module tests for `docmgr` – all green

### What Worked
- Reflection-based simple schema is sufficient for discoverability
- No need for a separate `contracts.go` at this time
- Flags integrate cleanly via glazed parameters

### What Could Be Improved
- Consider upgrading to JSON Schema library for richer typing and field docs
- Emit example values in schema (optional)

### Follow-ups
- Extend flags to future templated verbs when added (status, tasks list, search, vocab list, guidelines)
- Document the new flags in the user guide and examples

---

## 2025-11-20 - Simplified schema flag handling

### What I Did
- Simplified `--print-template-schema` handling to rely only on `settings.PrintTemplateSchema`
- Removed auxiliary flag detection helpers and extra checks
- Ensured that when the flag is set, commands print only the schema and skip all other output

### Why
- Glazed `InitializeStruct` reliably populates `settings`, so additional checks were redundant
- Keeps code paths straightforward and predictable

### Verification
- Built and ran:
  - `go run ./cmd/docmgr list docs --print-template-schema --schema-format yaml`
  - `go run ./cmd/docmgr list tickets --print-template-schema --schema-format yaml`
  - `go run ./cmd/docmgr doctor --print-template-schema --schema-format yaml`
- Observed schema-only output (no human sections or postfix templates)
