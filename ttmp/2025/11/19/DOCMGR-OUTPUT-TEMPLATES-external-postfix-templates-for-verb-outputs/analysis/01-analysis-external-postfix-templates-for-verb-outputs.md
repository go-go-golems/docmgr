---
Title: 'Analysis: External postfix templates for verb outputs'
Ticket: DOCMGR-OUTPUT-TEMPLATES
Status: active
Topics:
    - cli
    - templates
    - glaze
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/cmd/docmgr/cmds/common/common.go
      Note: Dual-mode builder and default glaze settings
    - Path: docmgr/pkg/commands/doctor.go
      Note: Human-mode grouped report suitable for postfix templating
    - Path: docmgr/pkg/commands/list_docs.go
      Note: Dual-mode list docs implementation (rows + human)
    - Path: docmgr/pkg/commands/list_tickets.go
      Note: Dual-mode list tickets implementation (rows + human)
    - Path: glazed/pkg/cli/cobra.go
      Note: Command run flow and verbs extraction
    - Path: glazed/pkg/cmds/runner/run.go
      Note: Command type switching and glaze processor wiring
    - Path: glazed/pkg/formatters/template/template.go
      Note: Formatter executing Go templates over table rows (reusable)
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T20:23:21.873380559-05:00
---


# Analysis: External postfix templates for verb outputs

### Purpose and scope

- Enable per‑verb external templates to append a configurable “postfix” to human output.
- Location: `ttmp/templates/$group/$verb.templ` (rooted at docs root).
- Each verb provides a typed struct as template data to steer LLMs with custom summaries.
- Applies by default in classic/human mode; opt‑in behavior for Glaze mode is discussed below.

### High‑level behavior

- After a command completes its normal human output, docmgr checks for a matching template file:
  - Grouped verbs: `templates/doc/list.templ`, `templates/list/tickets.templ`, etc.
  - Single‑level verbs: `templates/doctor.templ`
  - Fallbacks:
    - If `templates/$group/$verb.templ` exists, render it.
    - Else if `templates/$verb.templ` exists (for single‑level verbs), render it.
    - Else do nothing.
- The template is rendered with a typed data object provided by the verb.
- Rendering errors print a concise warning to stderr and are non‑fatal.

### Where to hook this in the codebase

- Runner decision for Glaze vs classic output and place to inject postfix:
  - `glazed/pkg/cli/cobra.go` (`runCobraCommand`): decides classic vs Glaze and runs the verb.
  - Use `GetVerbsFromCobraCommand` to compute `["docmgr","doc","list"] → group=doc, verb=list`.
- Dual‑mode commands and defaults:
  - `docmgr/cmd/docmgr/cmds/common/common.go`: wires dual mode and Glaze defaults.
- Examples of classic/human output we will append to:
  - `docmgr/pkg/commands/list_docs.go` (human run aggregates by ticket, prints bullets)
  - `docmgr/pkg/commands/list_tickets.go` (similar pattern for tickets)
  - `docmgr/pkg/commands/doctor.go` (grouped markdown report)
- Reusable templating infrastructure:
  - `glazed/pkg/formatters/template/template.go` shows current use of `text/template` over rows; we can reuse its philosophy and function maps.

### Template file resolution

- Docs root resolved as usual (CWD → `.ttmp.yaml` → default `ttmp`).
- Relative to docs root:
  - For nested verbs: `templates/$group/$verb.templ`
    - Examples: `templates/doc/list.templ`, `templates/list/tickets.templ`, `templates/doc/search.templ`
  - For single verbs: `templates/$verb.templ`
    - Examples: `templates/doctor.templ`, `templates/status.templ`
- Optional future extension:
  - Support subpaths along the entire verb chain: `templates/doc/list/docs.templ` (not needed now).
  - Configure `templatesRoot` in `.ttmp.yaml` to override `ttmp/templates`.

### When to render

- Default: render in classic/human mode only.
- Glaze mode:
  - Off by default (scripts expect machine output).
  - Add `--with-postfix-template` to explicitly append postfix even in Glaze mode (printed after structured output on stdout; never interleave within JSON/CSV payload).
- Global flag(s) proposal:
  - `--no-postfix-template` to disable.
  - `--with-postfix-template` to force-enable (including Glaze mode).
  - Env overrides: `DOCMGR_POSTFIX_TEMPLATE=0|1`.

### Template data contract

Principle: Each verb exposes a typed Go struct that captures its computed, human‑friendly state in a stable format for templates. For Glaze verbs, also provide rows and simple stats.

- Common envelope (available to all templates):
  - `Verbs []string` — full path, e.g., `["docmgr","doc","list"]`
  - `Root string` — absolute docs root used
  - `Now time.Time` — rendering timestamp
  - `Settings map[string]any` — parsed layer values relevant to the verb

- List Docs (`docmgr doc list` or `docmgr list docs`)
  - Proposed data:
    ```yaml
    TotalDocs: int
    TotalTickets: int
    Tickets:
      - Ticket: string
        Docs:
          - DocType: string
            Title: string
            Status: string
            Topics: []string
            Updated: string
            Path: string
    Rows: []map[string]any         # same fields as Glaze rows
    Fields: []string                # stable field names
    ```

- List Tickets (`docmgr list tickets`)
  - Proposed data:
    ```yaml
    TotalTickets: int
    Tickets:
      - Ticket: string
        Title: string
        Status: string
        Topics: []string
        Path: string
        LastUpdated: string
    Rows: []map[string]any
    Fields: []string
    ```

- Doctor (`docmgr doctor`)
  - Proposed data:
    ```yaml
    TotalFindings: int
    Tickets:
      - Ticket: string
        Findings:
          - Issue: string
            Severity: string
            Message: string
            Path: string
    ```` 

Notes:
- For Glaze verbs, `Rows` and `Fields` mirror what `--with-glaze-output` would produce.
- Classic verbs provide richer, domain‑specific summaries in addition to any rows.

### Example templates

- `ttmp/templates/doc/list.templ`
  ```gotemplate
  {{- /* LLM‑oriented summary of docs */ -}}
  ---
  summary:
    docs: {{ .TotalDocs }}
    tickets: {{ .TotalTickets }}
  guidance: |
    Prefer the most recently updated docs per ticket when summarizing.
  top_docs:
  {{- range $t := .Tickets }}
    {{- range $d := $t.Docs | slice 0 1 }}
    - ticket: {{ $t.Ticket }}
      title: {{ $d.Title }}
      type: {{ $d.DocType }}
      status: {{ $d.Status }}
      path: {{ $d.Path }}
    {{- end }}
  {{- end }}
  ```

- `ttmp/templates/list/tickets.templ`
  ```gotemplate
  ---
  ticket_overview:
    total: {{ .TotalTickets }}
    statuses:
    {{- $m := dict }}
    {{- range .Tickets }}{{- $s := .Status }}{{- $m = set $m $s (add1 (or (get $m $s) 0)) }}{{- end }}
    {{- range $k, $v := $m }}
      - {{ $k }}: {{ $v }}
    {{- end }}
  guidance: |
    Focus on active tickets first. Summarize changes since the last update.
  ```

- `ttmp/templates/doctor.templ`
  ```gotemplate
  ---
  doctor_summary:
    findings_total: {{ .TotalFindings }}
    by_ticket:
    {{- range .Tickets }}
      - ticket: {{ .Ticket }}
        errors: {{ countBy .Findings \"ERROR\" }}
        warnings: {{ countBy .Findings \"WARNING\" }}
        oks: {{ countBy .Findings \"OK\" }}
    {{- end }}
  guidance: |
    Address ERRORs first. For WARNINGs, add owner and due date.
  ```

Function helpers
- Reuse a minimal, safe `template.FuncMap` (e.g., `slice`, `dict`, `set`, `get`, `add1`, `countBy`) similar to what `glazed` uses for its template formatter. Keep the surface small and deterministic.

### Sample data shapes (abbreviated)

- List docs (classic/human run):
  ```yaml
  TotalDocs: 3
  TotalTickets: 2
  Tickets:
    - Ticket: MEN-4242
      Docs:
        - DocType: design-doc
          Title: Path Normalization
          Status: active
          Topics: [backend, routing]
          Updated: 2025-11-19 14:20
          Path: 2025/11/19/MEN-4242/path-normalization.md
  ```

- Doctor:
  ```yaml
  TotalFindings: 4
  Tickets:
    - Ticket: MEN-4242
      Findings:
        - Issue: missing-frontmatter
          Severity: WARNING
          Message: "Document missing Owners"
          Path: 2025/11/19/MEN-4242/design/01-x.md
  ```

### Implementation outline (no code changes yet)

- Add a small “postfix renderer” utility:
  - Resolve docs root (respect `.ttmp.yaml`) and compute template path using `GetVerbsFromCobraCommand`.
  - If template exists, render with the struct provided by the verb, plus the common envelope.
  - Append to stdout after the classic output block.
- Introduce a lightweight interface implemented by verbs:
  - `type TemplateDataProvider interface { TemplateData(ctx context.Context, parsed *layers.ParsedLayers) (any, error) }`
  - Classic verbs return a domain struct; Glaze verbs can return `Rows/Fields` plus a concise summary struct.
  - If a verb does not implement the interface, skip rendering.
- Flags:
  - Root level (default off in Glaze): `--with-postfix-template`, `--no-postfix-template`.
  - Env: `DOCMGR_POSTFIX_TEMPLATE=0|1`.
- Testing surface:
  - Add template files under `ttmp/templates/` and verify rendering for:
    - `doc list` (aka `list docs`)
    - `list tickets`
    - `doctor`

### Risks and considerations

- Avoid polluting Glaze output: never interleave templated text inside JSON/CSV; only append after.
- Keep the data contract stable; prefer additive changes.
- Enforce deterministic templates (no network/env).
- Long outputs: consider truncation helpers in templates (optional).

### Next steps

- Finalize the `TemplateDataProvider` shape and the common envelope fields.
- Implement postfix renderer and wire into classic run path.
- Add initial helpers and example templates under `ttmp/templates/`.
- Roll out to 3 verbs; iterate from feedback.
