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
  - path: docmgr/ttmp/templates/doc/list.templ
    note: Example template for doc list
  - path: docmgr/ttmp/templates/list/tickets.templ
    note: Example template for ticket list
  - path: docmgr/ttmp/templates/doctor.templ
    note: Example template for doctor report
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

- Add parameters to the verb’s settings struct:
  - `PrintTemplateSchema bool  \`glazed.parameter:"print-template-schema"\``
  - `SchemaFormat string       \`glazed.parameter:"schema-format"\`` (default `json`)
- In `Run`, early-return when `PrintTemplateSchema` is true:
  1. Build a minimal representative `templateData` shape (map/structs).
  2. Call `templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)`.
  3. `return nil`.

Use `list_docs.go`, `list_tickets.go`, and `doctor.go` as examples.

## 3) Build template data for the verb

- After printing the human output, assemble a typed struct or a `map[string]any` carrying:
  - Totals and summaries (e.g., `TotalDocs`, `TotalTickets`, `TotalFindings`).
  - Collections with stable item fields (e.g., `Tickets`, `Rows`, `Findings`).
  - Keep names consistent with other verbs when it makes sense.

## 4) Render postfix template (if present)

- Call `templates.RenderVerbTemplate(verbCandidates, root, settingsMap, templateData)`.
  - `verbCandidates` are ordered lists like `[][]string{{"status"}}` or `{{"doc","search"}}`.
  - `settingsMap` is a simple `map[string]any` (only settings you want exposed).
  - `root` is the docs root for resolving `ttmp/templates/...`.
- Canonical resolution (already implemented):
  - `ttmp/templates/<group>/<verb>.templ` (e.g., `list/tickets.templ`)
  - or `ttmp/templates/<verb>.templ` (e.g., `doctor.templ`)

## 5) Create an example template

- Place files under `ttmp/templates/...` using the canonical path.
- Use Glazed/Sprig helpers where possible (string/list/dict/math/date utils).
- Keep templates small, readable, and focused on LLM-friendly summaries.

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

- Templates are human-mode only; Glaze mode never renders postfix templates.
- Prefer stable, predictable data field names; keep types simple.
- Use `glamour`-rendered output for readability (already in the code).
- Root resolution matters; `workspace.ResolveRoot` is the source of truth.
- If schema grows complex, consider migrating to JSON Schema later.

# Reference Implementations

- Schema printing: `docmgr/internal/templates/schema.go`.
- Example verbs: `list_docs.go`, `list_tickets.go`, `doctor.go`.
- Example templates: `ttmp/templates/doc/list.templ`, `ttmp/templates/list/tickets.templ`, `ttmp/templates/doctor.templ`.

# Suggested Next Verbs (in order)

1. `status` — high value; add template data and example template.
2. `tasks list` — summarize open/done totals; simple data.
3. `search` — refactor to collect results, then render template.
4. `vocab list` — categorize topics/docTypes/intent/status.
5. `guidelines` — optional; consider metadata around source.

# Command Snippets

```bash
# Human output
go run ./cmd/docmgr list docs
go run ./cmd/docmgr list tickets
go run ./cmd/docmgr doctor

# Schema only
go run ./cmd/docmgr list docs --print-template-schema --schema-format yaml

# After edits
go build ./...
go test ./internal/templates -run TestPrintSchema
docmgr doctor --ticket DOCMGR-OUTPUT-TEMPLATES --fail-on error
```

# Handoff Notes

- Follow the patterns in the implemented verbs.
- Keep templates and data contracts small but expressive.
- Always update diary and changelog; relate changed files with short notes.


