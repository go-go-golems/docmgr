---
Title: Verb Output Templates and Template Schema
Slug: verb-templates-and-schema
Short: Guide to creating and using postfix templates for command output, and introspecting template data schemas.
Topics:
- docmgr
- templates
- cli
- automation
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Verb Output Templates and Template Schema

Verb output templates enable you to append custom, structured output to docmgr commands after they complete their normal human-friendly output. This system decouples command logic from output formatting, allowing you to create LLM-friendly summaries, automation-friendly data structures, or custom reports without modifying the command code. Templates are optional and only render when present, making them perfect for team-specific workflows or integration with external tools.

## Core Concepts

### Verb Templates

Verb templates are Go text templates that render **after** a command finishes its normal output. They receive structured data from the command and can format it however you need—as YAML summaries, JSON snippets, markdown reports, or any other text format.

**Key characteristics:**
- **Postfix rendering**: Templates append to command output, they don't replace it
- **Optional**: Commands work normally without templates; templates only render if present
- **Human mode only**: Templates render in classic/human output mode, never in Glaze structured mode
- **Non-fatal**: Template errors print warnings to stderr but don't fail the command

### Template Schema

Template schema introspection lets you discover what data is available to templates without reading source code. Each command that supports templates can print its data structure using the `--print-template-schema` flag, showing field names, types, and nested structures in JSON or YAML format.

**Use cases:**
- Understanding available fields before writing a template
- Documenting template data contracts
- Validating template correctness
- Generating template scaffolding

## How Verb Templates Work

### Template Resolution

Templates are resolved from the docs root (`ttmp/` by default) using a canonical path based on the command's verb structure:

- **Grouped verbs**: `templates/<group>/<verb>.templ`
  - `docmgr doc list` → `templates/doc/list.templ`
  - `docmgr list tickets` → `templates/list/tickets.templ`
  
- **Single-level verbs**: `templates/<verb>.templ`
  - `docmgr doctor` → `templates/doctor.templ`
  - `docmgr status` → `templates/status.templ`

**Resolution order:**
1. Check for template at canonical path
2. If found, render it with command data
3. If not found, command completes normally (no error)

### Common Template Data Envelope

All verb templates receive a common envelope of data available to every template:

```go
type CommonTemplateData struct {
    Verbs    []string               // Full verb path, e.g., ["docmgr", "doc", "list"]
    Root     string                 // Absolute docs root used
    Now      time.Time              // Rendering timestamp
    Settings map[string]interface{} // Parsed layer values (flags, config)
}
```

In templates, access these as:
- `{{ .Verbs }}` — Command path array
- `{{ .Root }}` — Docs root directory
- `{{ .Now }}` — Current timestamp
- `{{ .Settings }}` — Command settings map

### Verb-Specific Data

Each command provides its own data structure in addition to the common envelope. For example, `doc list` provides:

```go
type DocListTemplateData struct {
    TotalDocs    int
    TotalTickets int
    Tickets      []struct {
        Ticket string
        Docs   []struct {
            DocType string
            Title   string
            Status  string
            Topics  []string
            Updated string
            Path    string
        }
    }
}
```

Access verb-specific data directly in templates: `{{ .TotalDocs }}`, `{{ range .Tickets }}`, etc.

## Discovering Template Data with Schema Introspection

Before writing a template, discover what data is available using the `--print-template-schema` flag:

```bash
# Print schema for doc list command
docmgr doc list --print-template-schema

# Print schema in YAML format
docmgr doc list --print-template-schema --schema-format yaml

# Print schema for doctor command
docmgr doctor --print-template-schema
```

**Example schema output:**

```json
{
  "type": "object",
  "properties": {
    "TotalDocs": {
      "type": "integer"
    },
    "TotalTickets": {
      "type": "integer"
    },
    "Tickets": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "Ticket": {
            "type": "string"
          },
          "Docs": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "DocType": {
                  "type": "string"
                },
                "Title": {
                  "type": "string"
                },
                "Status": {
                  "type": "string"
                }
              }
            }
          }
        }
      }
    }
  }
}
```

**When to use schema introspection:**
- Before writing a new template
- When debugging template access errors
- When documenting template data contracts
- When validating template correctness

## Creating Verb Templates

### Step 1: Discover Available Data

First, print the schema to understand what fields are available:

```bash
docmgr doc list --print-template-schema --schema-format yaml
```

### Step 2: Create Template File

Create the template file at the canonical path. For `docmgr doc list`, create:

```
ttmp/templates/doc/list.templ
```

### Step 3: Write Template Content

Use Go text template syntax with the data structure from the schema:

```go
{{- /* LLM-oriented summary of docs */ -}}
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

### Step 4: Test Template

Run the command to see your template rendered:

```bash
docmgr doc list
# Normal output appears first...
# Then your template output appears
```

## Template Functions Available

Verb templates have access to Go's standard template functions plus Sprig template functions for common operations:

**Common functions:**
- `slice` — Extract sub-slices: `{{ slice .Items 0 5 }}`
- `first`, `last` — Get first/last element: `{{ first .Items }}`
- `has`, `hasKey` — Check existence: `{{ if has "key" .Map }}`
- `default` — Provide defaults: `{{ .Value | default "unknown" }}`
- `upper`, `lower`, `title` — String transformations
- `join` — Join slices: `{{ join .Topics ", " }}`
- `len` — Get length: `{{ len .Items }}`

**Custom functions:**
- `countBy` — Count items matching a condition (used in doctor templates)

See the [Sprig documentation](http://masterminds.github.io/sprig/) for the complete function list.

## Example Templates

### Simple Summary Template

**File:** `templates/doc/list.templ`

```go
{{- /* LLM-oriented summary of docs */ -}}
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

### Doctor Command Summary

**File:** `templates/doctor.templ`

```go
---
doctor_summary:
  findings_total: {{ .TotalFindings }}
  by_ticket:
  {{- range .Tickets }}
    - ticket: {{ .Ticket }}
      errors: {{ countBy .Findings "ERROR" }}
      warnings: {{ countBy .Findings "WARNING" }}
      oks: {{ countBy .Findings "OK" }}
  {{- end }}
guidance: |
  Address ERRORs first. For WARNINGs, add owner and due date.
```

### Status Command Report

**File:** `templates/status.templ`

```go
---
workspace_status:
  total_tickets: {{ .TicketsTotal }}
  stale_tickets: {{ .TicketsStale }}
  total_docs: {{ .DocsTotal }}
  doc_breakdown:
    design: {{ .DesignDocs }}
    reference: {{ .ReferenceDocs }}
    playbooks: {{ .Playbooks }}
  stale_after_days: {{ .StaleAfterDays }}
```

## Best Practices

### 1. Keep Templates Focused

Templates should provide **summary or structured data**, not duplicate the command's normal output. Use them for:
- LLM-friendly summaries
- Automation data structures
- Team-specific reporting formats
- Integration with external tools

### 2. Use YAML for Structured Output

YAML is more readable than JSON for human consumption and works well with LLMs:

```go
---
summary:
  total: {{ .Total }}
  items:
  {{- range .Items }}
  - name: {{ .Name }}
    status: {{ .Status }}
  {{- end }}
```

### 3. Handle Missing Data Gracefully

Use `default` to handle missing or zero values:

```go
{{ .Value | default "unknown" }}
{{ if .OptionalField }}{{ .OptionalField }}{{ end }}
```

### 4. Document Template Purpose

Add comments explaining what the template produces:

```go
{{- /* LLM-oriented summary for automation */ -}}
{{- /* Team-specific status report format */ -}}
```

### 5. Test with Schema First

Always check the schema before writing templates:

```bash
docmgr <verb> --print-template-schema
```

### 6. Keep Templates Versioned

Store templates in your repository so they're versioned with your documentation:

```
ttmp/
├── templates/
│   ├── doc/
│   │   └── list.templ
│   └── doctor.templ
```

## Commands Supporting Templates

The following commands support verb output templates:

- `doc list` / `list docs` — Lists all documents
- `list tickets` — Lists all tickets
- `doctor` — Validation and health checks
- `status` — Workspace status summary
- `search` — Document search results
- `tasks list` — Task listings
- `vocab list` — Vocabulary listings
- `doc guidelines` — Guideline display

Each command provides its own data structure. Use `--print-template-schema` to discover the exact fields available.

## Reference Examples

Example templates are available in the codebase at `examples/verb-templates/`. Copy these to `ttmp/templates/` in your workspace to use them, or use them as reference when creating your own:

```bash
# Copy example templates to your workspace
cp -r examples/verb-templates/* ttmp/templates/
```

**Available examples:**
- `examples/verb-templates/doc/list.templ` — Document listing summary
- `examples/verb-templates/list/tickets.templ` — Ticket listing summary
- `examples/verb-templates/doctor.templ` — Doctor findings summary
- `examples/verb-templates/status.templ` — Status report
- `examples/verb-templates/examples/*.templ` — Advanced examples

## Troubleshooting

### Template Not Rendering

**Check:**
1. Template file exists at canonical path: `ttmp/templates/<group>/<verb>.templ`
2. Running in human mode (not `--with-glaze-output`)
3. Template syntax is valid Go template syntax

**Debug:**
```bash
# Check if template file exists
ls -la ttmp/templates/doc/list.templ

# Test template syntax (if using a Go template validator)
```

### Template Errors

Template errors print warnings to stderr but don't fail the command. Check stderr for details:

```bash
docmgr doc list 2>&1 | grep -i "template\|warning"
```

### Unknown Fields

Use schema introspection to see available fields:

```bash
docmgr doc list --print-template-schema
```

### Template Functions Not Working

Ensure you're using valid Sprig functions. Check the [Sprig documentation](http://masterminds.github.io/sprig/) for available functions.

## Integration with Automation

Verb templates are ideal for automation workflows:

**CI/CD Integration:**
```bash
# Generate LLM-friendly summary for ticket review
docmgr doc list --ticket MEN-1234 > summary.yaml

# Extract structured data for reporting
docmgr doctor --all | grep -A 100 "doctor_summary" > health-report.yaml
```

**LLM Integration:**
Templates can format output specifically for LLM consumption, providing structured summaries that LLMs can easily parse and reason about.

## Advanced: Custom Template Functions

Commands can provide custom template functions beyond Sprig. Check command documentation or source code for command-specific functions like `countBy` used in doctor templates.

