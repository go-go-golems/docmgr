---
Title: 'Verbs Needing Template Support'
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
LastUpdated: 2025-11-19T21:45:00-05:00
---

# Verbs Needing Template Support

## Goal

This reference lists all docmgr verbs that have human-friendly output (`Run` methods) and could benefit from postfix template support, organized by priority.

## Context

Postfix templates are rendered after a command's human-friendly output completes. They provide LLM-oriented summaries and structured data for automation. Only verbs with substantial human output benefit from templates.

## Already Implemented ✅

1. **doc list** / **list docs**
   - Path: `templates/doc/list.templ` or `templates/list/docs.templ`
   - Status: ✅ Complete

2. **list tickets**
   - Path: `templates/list/tickets.templ`
   - Status: ✅ Complete

3. **doctor**
   - Path: `templates/doctor.templ`
   - Status: ✅ Complete

---

## High Priority (Most Useful)

### 1. **status** (`docmgr status`)

**Template Path:** `templates/status.templ`

**Current Output:**
- Summary line: `root=... config=... vocabulary=... tickets=N stale=M docs=N (design X / reference Y / playbooks Z) stale-after=N`
- Per-ticket lines (if not `--summary-only`): `TICKET 'Title' status=STATUS stale=BOOL docs=N path=PATH`

**Proposed Template Data:**
```yaml
Root: string
ConfigPath: string
VocabularyPath: string
TicketsTotal: int
TicketsStale: int
DocsTotal: int
DesignDocs: int
ReferenceDocs: int
Playbooks: int
StaleAfterDays: int
Tickets: []TicketStatus
  - Ticket: string
    Title: string
    Status: string
    Stale: bool
    DocsCount: int
    Path: string
```

**Use Cases:**
- Workspace health summaries
- Stale ticket alerts
- Doc type distribution analysis
- Progress tracking over time

**Implementation Notes:**
- Data is already collected in `Run` method
- Need to build struct before final printf
- `--summary-only` flag affects what data to include

---

### 2. **search** (`docmgr doc search`)

**Template Path:** `templates/doc/search.templ`

**Current Output:**
- Per-result lines: `path — title [ticket] :: snippet`
- With `--files`: `file — reason (source=...)`

**Proposed Template Data:**
```yaml
Query: string
File: string              # If --file filter used
Dir: string              # If --dir filter used
Status: string            # If --status filter used
DocType: string           # If --doc-type filter used
Topics: []string          # If --topics filter used
Results: []SearchResult
  - Path: string
    Title: string
    Ticket: string
    Snippet: string
    MatchedFiles: []string  # If --file used
    MatchedNotes: []string   # If --file used
TotalResults: int
```

**Use Cases:**
- Search result summaries
- Relevance scoring
- Query refinement suggestions
- Result grouping by ticket/doc type

**Implementation Notes:**
- Currently uses `filepath.Walk` with inline printing
- Need to collect results first, then print, then render template
- `--files` mode has different output format (file suggestions)

---

### 3. **tasks list** (`docmgr tasks list`)

**Template Path:** `templates/tasks/list.templ`

**Current Output:**
- Per-task lines: `[N] [x| ] text (file=path)`

**Proposed Template Data:**
```yaml
Ticket: string
File: string
TotalTasks: int
OpenTasks: int
DoneTasks: int
Tasks: []TaskInfo
  - Index: int
    Checked: bool
    Text: string
    File: string
```

**Use Cases:**
- Task completion summaries
- Progress tracking
- Task distribution analysis
- Completion rate calculations

**Implementation Notes:**
- Data already structured (tasks slice)
- Easy to add template rendering
- Simple data structure

---

## Medium Priority

### 4. **vocab list** (`docmgr vocab list`)

**Template Path:** `templates/vocab/list.templ`

**Current Output:**
- Per-item lines: `category: slug — description`

**Proposed Template Data:**
```yaml
Category: string          # Filter category (empty = all)
Topics: []VocabItem
  - Slug: string
    Description: string
DocTypes: []VocabItem
  - Slug: string
    Description: string
Intent: []VocabItem
  - Slug: string
    Description: string
Status: []VocabItem
  - Slug: string
    Description: string
```

**Use Cases:**
- Vocabulary summaries
- Category breakdowns
- Vocabulary completeness checks

**Implementation Notes:**
- Data already loaded from vocabulary file
- Simple to structure for templates
- Less commonly used command

---

## Low Priority (Less Useful)

### 5. **guidelines** (`docmgr doc guidelines`)

**Template Path:** `templates/doc/guidelines.templ`

**Current Output:**
- Raw guideline text (markdown)

**Proposed Template Data:**
```yaml
DocType: string
GuidelineText: string
Source: string            # "filesystem" or "embedded"
Path: string             # If from filesystem
```

**Use Cases:**
- Less useful since output is mostly static content
- Could add metadata about guideline source

**Implementation Notes:**
- Output is mostly static markdown
- Limited value for LLM consumption
- Low priority

---

## Not Suitable for Templates

These verbs have minimal output or are mutation commands:

- **add** - Creates documents, minimal confirmation output
- **create-ticket** - Creates tickets, minimal confirmation output
- **meta update** - Updates metadata, minimal confirmation output
- **relate** - Relates files, minimal confirmation output
- **tasks add/check/uncheck/edit/remove** - Task mutations, minimal confirmation
- **ticket close** - Closes tickets, minimal confirmation output
- **init/configure** - Setup commands, minimal output
- **layout-fix/renumber** - Utility commands, minimal output
- **import-file** - Import command, minimal output
- **vocab add** - Adds vocabulary, minimal confirmation output
- **config show** - Shows config, minimal output
- **rename-ticket** - Renames tickets, minimal confirmation output

---

## Implementation Priority Recommendation

1. **status** - High value, data already collected
2. **tasks list** - High value, simple structure
3. **search** - High value, but requires refactoring output collection
4. **vocab list** - Medium value, simple structure
5. **guidelines** - Low value, mostly static content

---

## Related

- See `reference/01-template-data-contracts-reference.md` for data structure documentation
- See `analysis/01-analysis-external-postfix-templates-for-verb-outputs.md` for design rationale
- See `log/01-implementation-diary-external-postfix-templates.md` for implementation patterns

