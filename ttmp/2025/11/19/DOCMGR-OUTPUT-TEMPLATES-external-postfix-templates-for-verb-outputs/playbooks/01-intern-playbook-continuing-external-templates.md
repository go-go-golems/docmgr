---
Title: Intern Playbook — Continuing External Postfix Templates
Ticket: DOCMGR-OUTPUT-TEMPLATES
Status: active
Topics:
  - cli
  - templates
  - glaze
DocType: playbook
Intent: long-term
Owners:
  - manuel
RelatedFiles:
  - path: docmgr/internal/templates/verb_output.go
    note: Postfix template rendering entrypoint and FuncMap
  - path: docmgr/internal/templates/schema.go
    note: PrintSchema to output template schema (JSON/YAML)
  - path: docmgr/pkg/commands/list_docs.go
    note: Example wiring for postfix templates and schema flags
  - path: docmgr/pkg/commands/list_tickets.go
    note: Example wiring for postfix templates and schema flags
  - path: docmgr/pkg/commands/doctor.go
    note: Example wiring for postfix templates and schema flags
  - path: docmgr/pkg/commands/status.go
    note: Example wiring for postfix templates and schema flags
  - path: docmgr/pkg/commands/tasks.go
    note: Example wiring for postfix templates and schema flags
  - path: docmgr/pkg/commands/search.go
    note: Example wiring for postfix templates and schema flags (refactored to collect results first)
  - path: docmgr/pkg/commands/vocab_list.go
    note: Example wiring for postfix templates and schema flags
  - path: docmgr/pkg/commands/guidelines_cmd.go
    note: Example wiring for postfix templates and schema flags
  - path: docmgr/ttmp/templates/doc/list.templ
    note: Example template for doc list
  - path: docmgr/ttmp/templates/list/tickets.templ
    note: Example template for ticket list
  - path: docmgr/ttmp/templates/doctor.templ
    note: Example template for doctor report
  - path: docmgr/ttmp/templates/status.templ
    note: Example template for status command
  - path: docmgr/ttmp/templates/tasks/list.templ
    note: Example template for tasks list command
  - path: docmgr/ttmp/templates/doc/search.templ
    note: Example template for search command
  - path: docmgr/ttmp/templates/vocab/list.templ
    note: Example template for vocab list command
  - path: docmgr/ttmp/templates/doc/guidelines.templ
    note: Example template for guidelines command
ExternalSources: []
Summary: Step-by-step guide for implementing postfix templates and template schema printing for additional verbs, with validation, examples, and documentation best practices.
LastUpdated: 2025-11-20T00:00:00Z
---

# Purpose

Continue the rollout of external postfix templates across docmgr verbs and maintain a consistent developer workflow. This playbook shows how to add templates, print schemas, test outputs, and document your changes.

# Prerequisites

- Go toolchain installed and working.
- This repo checked out; run commands from repository root unless specified.
- Run `docmgr help how-to-use` to familiarize yourself with the workflow and output modes.
- Read and execute the tasks for this ticket DOCMGR-OUTPUT-TEMPLATES

# Quick Checklist (Definition of Done)

- [ ] Human output remains unchanged and readable.
- [ ] Postfix template renders (if a file exists at the canonical path).
- [ ] `--print-template-schema` prints schema only (no extra output).
- [ ] Example template added under `ttmp/templates/...`.
- [ ] Diary and changelog updated; related files linked.
- [ ] `docmgr doctor` shows no new critical issues.

# Step-by-Step

## 1) Pick a verb and locate its Run method

- Files are under `docmgr/pkg/commands/`.
- You need a command with a human-friendly `Run` method (not only Glaze).
- Good next targets: `status`, `tasks list`, `search`, `vocab list`, `guidelines`.

## 2) Add schema flags (if the verb supports templating)

- **Add import**: `"github.com/go-go-golems/docmgr/internal/templates"`
- Add parameters to the verb's settings struct:
  - `PrintTemplateSchema bool  \`glazed.parameter:"print-template-schema"\``
  - `SchemaFormat string       \`glazed.parameter:"schema-format"\`` (default `json`)
- Add flag definitions in `New*Command()`:
  - `parameters.NewParameterDefinition("print-template-schema", ...)`
  - `parameters.NewParameterDefinition("schema-format", ...)`
- **Important**: Add early-return in BOTH `RunIntoGlazeProcessor` AND `Run` methods when `PrintTemplateSchema` is true:
  1. Call `workspace.ResolveRoot(settings.Root)` first (if using root).
  2. Build a minimal representative `templateData` shape (map/structs).
  3. Call `templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)`.
  4. `return nil`.

Use `list_docs.go`, `list_tickets.go`, `doctor.go`, `status.go`, and `tasks.go` as examples.

## 3) Build template data for the verb

- After printing the human output, assemble a typed struct or a `map[string]any` carrying:
  - Totals and summaries (e.g., `TotalDocs`, `TotalTickets`, `TotalFindings`, `TotalTasks`, `OpenTasks`, `DoneTasks`).
  - Collections with stable item fields (e.g., `Tickets`, `Rows`, `Findings`, `Tasks`).
  - Keep names consistent with other verbs when it makes sense.
- **Tip**: Match the structure of your human output - if you're already collecting data for display, reuse that structure for template data.
- **Calculate derived values in Go**: Don't rely on template math functions (they're limited). Calculate percentages, ratios, etc. in Go code before passing to template.

## 4) Render postfix template (if present)

- **Important**: Ensure `workspace.ResolveRoot(settings.Root)` is called before this step.
- Call `templates.RenderVerbTemplate(verbCandidates, root, settingsMap, templateData)`.
  - `verbCandidates` are ordered lists like `[][]string{{"status"}}` or `{{"tasks","list"}}` or `{{"doc","search"}}`.
  - `settingsMap` is a simple `map[string]any` (only settings you want exposed to templates).
  - `root` should be the absolute path to the docs root (use `filepath.Abs()` if needed).
- Canonical resolution (already implemented):
  - Single verb: `ttmp/templates/<verb>.templ` (e.g., `status.templ`, `doctor.templ`)
  - Grouped verb: `ttmp/templates/<group>/<verb>.templ` (e.g., `list/tickets.templ`, `tasks/list.templ`)
- The function returns `bool` indicating if a template was found/rendered (you can ignore it).

## 5) Create an example template

- Place files under `ttmp/templates/...` using the canonical path (see step 4).
- **Available template functions**: Check `internal/templates/verb_output.go` `getTemplateFuncMap()` for available functions:
  - `slice`, `dict`, `set`, `get`, `add1`, `countBy`
  - **Note**: Math functions like `div`, `mul`, `sub` are NOT available. Calculate values in Go code instead.
- Keep templates small, readable, and focused on LLM-friendly summaries (YAML format works well).
- Use conditionals (`{{- if ... }}`) to filter collections (e.g., only show open tasks).

## 6) Test locally

- Human output (TTY pretty):
  - `go run ./cmd/docmgr <verb> [flags...]`
- Schema only:
  - `go run ./cmd/docmgr <verb> --print-template-schema --schema-format yaml`
- Template present:
  - Human output appears as before, then a blank line, then the template output.

## 7) Update docs and hygiene

- Diary: Add what you did, what worked, and what needs follow-up.
- Changelog: Short entry; include related files with notes.
- Relate important code/docs to the ticket or sub-docs.
- Run: `docmgr doctor --ticket DOCMGR-OUTPUT-TEMPLATES --fail-on error`.

# Tips and Gotchas

- **Templates are human-mode only**: Glaze mode never renders postfix templates.
- **Always call `workspace.ResolveRoot`**: Before using `settings.Root`, call `workspace.ResolveRoot(settings.Root)` to get the correct resolved path.
- **Both methods need schema early return**: Add the schema printing early return in BOTH `RunIntoGlazeProcessor` AND `Run` methods.
- **Template function limitations**: The FuncMap doesn't include math functions (`div`, `mul`, `sub`). Calculate derived values (percentages, ratios) in Go code before passing to template.
- **Absolute paths for root**: When calling `RenderVerbTemplate`, use an absolute path for the root parameter (use `filepath.Abs()` if needed).
- Prefer stable, predictable data field names; keep types simple.
- Use `glamour`-rendered output for readability (already in the code).
- Root resolution matters; `workspace.ResolveRoot` is the source of truth.
- If schema grows complex, consider migrating to JSON Schema later.
- **Test both outputs**: Always verify human output is unchanged AND schema printing works correctly.

# Reference Implementations

- Schema printing: `docmgr/internal/templates/schema.go`.
- Template rendering: `docmgr/internal/templates/verb_output.go`.
- Example verbs: `list_docs.go`, `list_tickets.go`, `doctor.go`, `status.go`, `tasks.go`, `search.go`, `vocab_list.go`, `guidelines_cmd.go`.
- Example templates: 
  - `ttmp/templates/doc/list.templ`
  - `ttmp/templates/list/tickets.templ`
  - `ttmp/templates/doctor.templ`
  - `ttmp/templates/status.templ`
  - `ttmp/templates/tasks/list.templ`
  - `ttmp/templates/doc/search.templ`
  - `ttmp/templates/vocab/list.templ`
  - `ttmp/templates/doc/guidelines.templ`

# Suggested Next Verbs (in order)

1. ~~`status`~~ ✅ — Completed; template data with ticket info and totals.
2. ~~`tasks list`~~ ✅ — Completed; simple data with open/done totals.
3. ~~`search`~~ ✅ — Completed; refactored to collect results, then render template.
4. ~~`vocab list`~~ ✅ — Completed; categorize topics/docTypes/intent/status.
5. ~~`guidelines`~~ ✅ — Completed; consider metadata around source.

# Command Snippets

```bash
# Human output (verify unchanged)
go run ./cmd/docmgr list docs
go run ./cmd/docmgr list tickets
go run ./cmd/docmgr doctor
go run ./cmd/docmgr status --summary-only
go run ./cmd/docmgr tasks list --ticket DOCMGR-OUTPUT-TEMPLATES

# Schema only (verify schema-only output)
go run ./cmd/docmgr list docs --print-template-schema --schema-format yaml
go run ./cmd/docmgr status --print-template-schema --schema-format yaml
go run ./cmd/docmgr tasks list --ticket DOCMGR-OUTPUT-TEMPLATES --print-template-schema --schema-format yaml

# After edits
go build ./cmd/docmgr
go test ./internal/templates -run TestPrintSchema
docmgr doctor --ticket DOCMGR-OUTPUT-TEMPLATES --fail-on error
```

# Handoff Notes

- Follow the patterns in the implemented verbs (`status.go` and `tasks.go` are good recent examples).
- Keep templates and data contracts small but expressive.
- Always update diary and changelog; relate changed files with short notes.
- Remember: calculate derived values in Go, not in templates (math functions are limited).
- Always test both human output (should be unchanged) and schema printing (should be schema-only).

# Lessons Learned (from status and tasks list implementations)

1. **Template functions are limited**: The FuncMap doesn't include `div`, `mul`, `sub`. Calculate percentages/ratios in Go code before passing to template.

2. **Root resolution is critical**: Always call `workspace.ResolveRoot(settings.Root)` before using the root path, and use absolute paths when calling `RenderVerbTemplate`.

3. **Both methods need schema early return**: Don't forget to add the schema printing early return in BOTH `RunIntoGlazeProcessor` AND `Run` methods.

4. **Match human output structure**: When building template data, reuse the same data structures you're already collecting for human output - it makes the code cleaner and easier to maintain.

5. **Simple templates work best**: Keep templates focused on LLM-friendly summaries. Complex calculations should happen in Go code.

6. **Test thoroughly**: Always verify:
   - Human output remains unchanged
   - Schema printing outputs only schema (no extra text)
   - Template renders correctly when present
   - Build succeeds without errors


