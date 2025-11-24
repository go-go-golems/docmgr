---
Title: 'Template Data Contracts Reference'
Ticket: DOCMGR-OUTPUT-TEMPLATES
Status: active
Topics:
    - cli
    - templates
    - glaze
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T21:30:00-05:00
---

# Template Data Contracts Reference

## Goal

This reference documents the exact data structures available in postfix templates for each verb. Use this when writing templates to know what fields are available and their types.

## Context

Postfix templates are rendered after a command's human-friendly output completes. Each verb builds a typed data structure containing its output state, which is then merged with a common envelope and passed to the template.

Templates are located at:
- Grouped verbs: `templates/$group/$verb.templ` (e.g., `templates/doc/list.templ`)
- Single verbs: `templates/$verb.templ` (e.g., `templates/doctor.templ`)

## Common Envelope

All templates receive these common fields:

```yaml
Verbs: []string          # Full verb path, e.g., ["docmgr", "doc", "list"]
Root: string             # Absolute docs root path
Now: time.Time           # Rendering timestamp
Settings: map[string]any  # Parsed layer values (root, ticket, status, etc.)
```

Access in templates:
```gotemplate
{{ .Verbs }}      # ["docmgr", "doc", "list"]
{{ .Root }}       # "/path/to/ttmp"
{{ .Now }}        # 2025-11-19 21:30:00 -0500 EST
{{ .Settings.root }}  # "ttmp"
```

## Verb-Specific Data Contracts

### `docmgr doc list` / `docmgr list docs`

**Template Paths:**
- `templates/doc/list.templ`
- `templates/list/docs.templ`

**Data Structure:**

```yaml
TotalDocs: int           # Total number of documents
TotalTickets: int        # Total number of tickets
Tickets: []TicketInfo    # Documents grouped by ticket
Rows: []map[string]any   # Same fields as Glaze rows
Fields: []string         # Stable field names: ["ticket", "doc_type", "title", "status", "topics", "path", "last_updated"]
```

**TicketInfo:**
```yaml
Ticket: string           # Ticket identifier
Docs: []DocInfo          # Documents in this ticket
```

**DocInfo:**
```yaml
DocType: string          # Document type (e.g., "design-doc", "reference")
Title: string            # Document title
Status: string           # Document status
Topics: []string         # List of topics
Updated: string          # Last updated timestamp (format: "2006-01-02 15:04")
Path: string             # Relative path from docs root
```

**Example Template Usage:**
```gotemplate
{{- range .Tickets }}
Ticket: {{ .Ticket }}
  Docs: {{ len .Docs }}
  {{- range .Docs | slice 0 1 }}
  Latest: {{ .Title }} ({{ .DocType }})
  {{- end }}
{{- end }}
```

---

### `docmgr list tickets`

**Template Path:**
- `templates/list/tickets.templ`

**Data Structure:**

```yaml
TotalTickets: int        # Total number of tickets
Tickets: []TicketInfo    # List of tickets
Rows: []map[string]any   # Same fields as Glaze rows
Fields: []string         # Stable field names: ["ticket", "title", "status", "topics", "tasks_open", "tasks_done", "path", "last_updated"]
```

**TicketInfo:**
```yaml
Ticket: string           # Ticket identifier
Title: string            # Ticket title
Status: string           # Ticket status (e.g., "active", "complete")
Topics: []string         # List of topics
Path: string             # Relative path from docs root
LastUpdated: string      # Last updated timestamp (format: "2006-01-02 15:04")
```

**Example Template Usage:**
```gotemplate
{{- $m := dict }}
{{- range .Tickets }}{{- $s := .Status }}{{- $m = set $m $s (add1 (or (get $m $s) 0)) }}{{- end }}
Status breakdown:
{{- range $k, $v := $m }}
  {{ $k }}: {{ $v }}
{{- end }}
```

---

### `docmgr doctor`

**Template Path:**
- `templates/doctor.templ`

**Data Structure:**

```yaml
TotalFindings: int       # Total number of findings (excluding "none" issues)
Tickets: []TicketFindings  # Findings grouped by ticket
```

**TicketFindings:**
```yaml
Ticket: string           # Ticket identifier
Findings: []Finding      # List of findings for this ticket
```

**Finding:**
```yaml
Issue: string            # Issue type (e.g., "missing_frontmatter", "unknown_topics")
Severity: string         # Severity level: "ERROR", "WARNING", "OK"
Message: string          # Human-readable message
Path: string             # Path to the file with the issue
```

**Note:** Findings with `Issue == "none"` are excluded from the template data (they represent "all checks passed").

**Example Template Usage:**
```gotemplate
{{- range .Tickets }}
{{ .Ticket }}:
  Errors: {{ countBy .Findings "ERROR" }}
  Warnings: {{ countBy .Findings "WARNING" }}
{{- end }}
```

---

## Available Template Functions

All templates have access to these helper functions:

### `slice(start, end, slice)`
Extract a sub-slice from an array.

```gotemplate
{{- range .Tickets | slice 0 3 }}
  {{ .Ticket }}
{{- end }}
```

### `dict(key1, val1, key2, val2, ...)`
Create a map from key-value pairs.

```gotemplate
{{- $m := dict "active" 0 "complete" 0 }}
```

### `set(map, key, value)`
Set a value in a map (creates map if nil).

```gotemplate
{{- $m = set $m "active" (add1 (get $m "active")) }}
```

### `get(map, key)`
Get a value from a map (returns nil if missing).

```gotemplate
{{- $count := get $m "active" }}
```

### `add1(number)`
Increment a number by 1. Handles int, int64, float64.

```gotemplate
{{- add1 5 }}  # 6
{{- add1 (get $m "count") }}
```

### `countBy(slice, value)`
Count items in a slice matching a value. Works with:
- Maps: checks `Severity` or `Status` fields
- Structs: checks `Severity` field via reflection

```gotemplate
{{- countBy .Findings "ERROR" }}    # Count findings with Severity="ERROR"
{{- countBy .Tickets "active" }}    # Count tickets with Status="active"
```

---

## Usage Examples

### Example 1: Status Aggregation (list tickets)

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
```

### Example 2: Top Documents Per Ticket (doc list)

```gotemplate
{{- range .Tickets }}
{{ .Ticket }}:
  {{- range .Docs | slice 0 1 }}
  Latest: {{ .Title }} ({{ .DocType }}, {{ .Status }})
  {{- end }}
{{- end }}
```

### Example 3: Findings Summary (doctor)

```gotemplate
---
doctor_summary:
  findings_total: {{ .TotalFindings }}
  by_ticket:
  {{- range .Tickets }}
    - ticket: {{ .Ticket }}
      errors: {{ countBy .Findings "ERROR" }}
      warnings: {{ countBy .Findings "WARNING" }}
  {{- end }}
```

---

## Best Practices

1. **Always check for empty slices** before iterating:
   ```gotemplate
   {{- if gt (len .Tickets) 0 }}
   {{- range .Tickets }}
   ...
   {{- end }}
   {{- end }}
   ```

2. **Use `slice` for limiting results**:
   ```gotemplate
   {{- range .Docs | slice 0 5 }}  # First 5 docs
   ```

3. **Handle nil maps** when using `get`:
   ```gotemplate
   {{- $val := or (get $m "key") 0 }}  # Default to 0 if nil
   ```

4. **Format timestamps** if needed (they're already formatted as strings):
   ```gotemplate
   Updated: {{ .Updated }}  # Already formatted as "2006-01-02 15:04"
   ```

5. **Use `countBy` for aggregations**:
   ```gotemplate
   {{- countBy .Findings "ERROR" }}  # More reliable than manual counting
   ```

---

## Related

- See `analysis/01-analysis-external-postfix-templates-for-verb-outputs.md` for design rationale
- See `log/01-implementation-diary-external-postfix-templates.md` for implementation details
- Example templates in `ttmp/templates/`
