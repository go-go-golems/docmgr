---
Title: Cleanup Overview and Migration Guide
Ticket: CLEANUP-LEGACY-WALKERS
Status: active
Topics:
    - refactor
    - tickets
    - docmgr-internals
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/09-cleanup-inventory-report-task-18.md
      Note: Detailed inventory of all cleanup targets with line numbers
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Design spec for the Workspace API being adopted
    - Path: internal/workspace/workspace.go
      Note: Canonical Workspace API implementation
    - Path: internal/workspace/query_docs.go
      Note: QueryDocs implementation - the replacement for legacy walkers
    - Path: pkg/commands/import_file.go
      Note: Contains findTicketDirectory helper definition (prime cleanup target)
    - Path: internal/workspace/discovery.go
      Note: Contains CollectTicketWorkspaces (legacy discovery helper)
ExternalSources: []
Summary: Migrate all docmgr commands from legacy discovery/walking patterns to the canonical Workspace+QueryDocs API
LastUpdated: 2025-12-13T10:29:25.71572102-05:00
---

# Cleanup Overview and Migration Guide

## 1. Background and Context

This ticket continues the work started in **REFACTOR-TICKET-REPOSITORY-HANDLING**, which introduced a centralized `workspace.Workspace` object backed by an in-memory SQLite index for document discovery and metadata handling.

## Compatibility policy (non-negotiable)

This cleanup explicitly does **not** preserve backwards compatibility.

- We will not add compatibility flags, shims, or fallback behaviors to emulate legacy semantics.
- Behavior changes are acceptable and expected as part of removing duplicated walkers and consolidating semantics in `Workspace.QueryDocs`.
- When legacy behavior differs from the canonical QueryDocs semantics, **QueryDocs wins**.

### What was built

The refactor ticket delivered:

1. **`internal/workspace/workspace.go`** — A `Workspace` object that discovers the docs root, loads config, and provides a `paths.Resolver` for consistent path normalization.

2. **`internal/workspace/sqlite_schema.go`** — An in-memory SQLite schema with tables for `docs`, `doc_topics`, and `related_files`, populated at command startup.

3. **`internal/workspace/query_docs.go`** — A `QueryDocs(ctx, DocQuery)` API that translates structured queries into SQL and returns `DocHandle` results with hydrated metadata.

4. **Canonical skip rules** — `internal/workspace/skip_policy.go` defines which directories are skipped during ingestion (`.meta/`, `_*/`, `scripts/`, etc.) and which path tags are applied (`is_archived_path`, `is_control_doc`, etc.).

5. **Path normalization** — `internal/workspace/normalization.go` uses `paths.Resolver` to store multiple normalized representations of each `RelatedFile` path, enabling robust reverse lookup.

### What still needs cleanup

Several commands were ported to use the new `Workspace` API during the refactor:
- `list docs` (Task 9)
- `search` (Task 10)
- `doctor` (Task 16)
- `relate` (Task 17)

However, **many other commands still use legacy patterns**:
- **`findTicketDirectory`** — A helper in `pkg/commands/import_file.go` that calls `CollectTicketWorkspaces` and iterates to find a ticket by ID. Used by 15+ commands.
- **`workspace.CollectTicketWorkspaces`** — Walks the entire docs root to enumerate tickets. Used directly by `status.go`, `list_tickets.go`, `list.go`.
- **`filepath.Walk` / `filepath.WalkDir`** — Manual filesystem traversal with ad-hoc skip rules and frontmatter parsing.
- **`readDocumentFrontmatter`** — Manual parsing of frontmatter instead of using index results.

This cleanup ticket systematically migrates all remaining commands to use `Workspace.QueryDocs`, then deletes the legacy helpers.

---

## 2. Inventory Summary

The detailed inventory is in:

> `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/09-cleanup-inventory-report-task-18.md`

### By the numbers

| Category | Count | Description |
|----------|-------|-------------|
| **Discovery** | 8 | `findTicketDirectory`, `CollectTicketWorkspaces` usage |
| **Traversal** | 7 | `filepath.Walk`, `filepath.WalkDir` for doc enumeration |
| **Filtering** | 3 | Manual skip logic (should be handled by QueryDocs) |
| **Parsing** | 4 | `readDocumentFrontmatter` instead of index results |
| **Normalization** | 1 | Ad-hoc path resolution |
| **Total** | **23** | Cleanup targets across 12 command files |

### Prime targets

| Helper | Location | Call sites |
|--------|----------|------------|
| `findTicketDirectory` | `pkg/commands/import_file.go:112` | 20 call sites across 15 files |
| `CollectTicketWorkspaces` | `internal/workspace/discovery.go` | 6 call sites (status, list_tickets, list, import_file) |
| `CollectTicketScaffoldsWithoutIndex` | `internal/workspace/discovery.go` | 1 call site (doctor) |

---

## 3. Migration Patterns

### Pattern A: Ticket Discovery

**Legacy code:**
```go
ticketDir, err := findTicketDirectory(settings.Root, settings.Ticket)
if err != nil {
    return err
}
```

**Canonical replacement:**
```go
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
if err != nil {
    return err
}
if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{}); err != nil {
    return err
}
result, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope:   workspace.ScopeTicket,
    Filters: workspace.DocFilters{Ticket: settings.Ticket, DocType: "index"},
})
if err != nil || len(result.Docs) == 0 {
    return fmt.Errorf("ticket not found: %s", settings.Ticket)
}
ticketDir := filepath.Dir(result.Docs[0].Path)
```

### Pattern B: Document Enumeration

**Legacy code:**
```go
err := filepath.Walk(ticketDir, func(path string, info os.FileInfo, err error) error {
    if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
        return nil
    }
    doc, err := readDocumentFrontmatter(path)
    // ... process doc
})
```

**Canonical replacement:**
```go
result, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope:   workspace.ScopeTicket,
    Filters: workspace.DocFilters{Ticket: ticketID},
})
if err != nil {
    return err
}
for _, h := range result.Docs {
    // h.Doc contains parsed frontmatter
    // h.Path is the absolute path
    // ... process h
}
```

### Pattern C: Manual Skip Rules

**Legacy code:**
```go
if strings.HasPrefix(info.Name(), "_") || info.Name() == ".meta" {
    return filepath.SkipDir
}
```

**Canonical replacement:**

QueryDocs automatically applies skip rules from `skip_policy.go`. Use options to control visibility:

```go
result, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.ScopeTicket,
    Options: workspace.DocQueryOptions{
        IncludeArchivedPath: false, // default: exclude archive/
        IncludeScriptsPath:  false, // default: exclude scripts/
        IncludeControlDocs:  false, // default: exclude control docs
    },
})
```

### Pattern D: Manual Frontmatter Parsing

**Legacy code:**
```go
doc, err := readDocumentFrontmatter(path)
if err != nil {
    // skip or handle
}
docType := doc.DocType
```

**Canonical replacement:**

QueryDocs returns parsed frontmatter in `DocHandle.Doc`:

```go
for _, h := range result.Docs {
    docType := h.Doc.DocType  // already parsed
    topics := h.Topics        // already hydrated
    related := h.RelatedFiles // already hydrated
}
```

---

## 4. Risk Notes

### Behavior Changes

1. **Ticket filter semantics** (`list_tickets.go`):
   - **Current**: Substring match (`strings.Contains`)
   - **After**: Exact match
   - This is intentional: we do not support compatibility modes.

2. **Parse error visibility**:
   - **Current**: Many commands silently skip docs with parse errors
   - **After**: QueryDocs can surface them via `IncludeErrors=true` and `Diagnostics`
   - This is intentional: commands should adopt the QueryDocs defaults and options, not legacy heuristics.

3. **Index build overhead**:
   - Commands now build an in-memory SQLite index at startup
   - This is a one-time cost per invocation (typically < 100ms for moderate repos)

### Write-Path Commands

Commands that **modify files** still need filesystem traversal for the write portion:
- `renumber.go` — updates file contents
- `layout_fix.go` — moves files
- `rename_ticket.go` — updates references

For these commands, **only ticket discovery** should migrate to QueryDocs. The write-path traversal remains filesystem-based.

---

## 5. Phased PR Plan

### Phase 1: Foundation (Low Risk, High Impact)

| PR | Target | Changes | Risk |
|----|--------|---------|------|
| 1.1 | `status.go` | Replace `CollectTicketWorkspaces` + `filepath.Walk` with QueryDocs | Low |
| 1.2 | `list_tickets.go` | Replace `CollectTicketWorkspaces` with QueryDocs | Medium (filter semantics) |
| 1.3 | `list.go` | Replace `CollectTicketWorkspaces` with QueryDocs | Low |
| 1.4 | `changelog.go` | Replace suggestion-mode walking with QueryDocs | Low |

### Phase 2: Write-Path Discovery

| PR | Target | Changes | Risk |
|----|--------|---------|------|
| 2.1 | `add.go` | Replace `findTicketDirectory` with QueryDocs discovery | Low |
| 2.2 | `meta_update.go` | Replace `findTicketDirectory` + `findMarkdownFiles` with QueryDocs | Low |
| 2.3 | `tasks.go` | Replace `findTicketDirectory` fallback with QueryDocs | Low |

### Phase 3: Remaining Commands

| PR | Target | Changes | Risk |
|----|--------|---------|------|
| 3.1 | `search.go` | Migrate suggestion-mode walking (main path already done) | Low |
| 3.2 | `doc_move.go`, `ticket_move.go`, `ticket_close.go` | Replace `findTicketDirectory` | Low |
| 3.3 | `rename_ticket.go`, `renumber.go`, `layout_fix.go` | Replace discovery only (keep write-path walking) | Low |

### Phase 4: Delete Legacy Helpers

| PR | Target | Changes | Risk |
|----|--------|---------|------|
| 4.1 | `import_file.go` | Delete `findTicketDirectory` function | Low (all callers migrated) |
| 4.2 | `discovery.go` | Deprecate/delete `CollectTicketWorkspaces` if unused | Medium (verify no external callers) |

---

## 6. Validation Strategy

### For each PR

1. **Unit tests**: `go test ./...`
2. **Integration tests**: `bash test-scenarios/testing-doc-manager/run-all.sh`
3. **Manual smoke tests**:
   - Run command with `--ticket` filter
   - Run command without filter (repo-wide)
   - Test edge cases (missing ticket, parse errors)
4. **Compare output**: Before/after for human and glaze modes

### Baseline

The refactor ticket established integration test baselines:
- System `docmgr`: All scenarios pass
- Local build: All scenarios pass

Maintain this baseline throughout the cleanup.

---

## 7. Related Documents

- **Inventory Report**: `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/09-cleanup-inventory-report-task-18.md`
- **Design Spec**: `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md`
- **Refactor Diary**: `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/15-diary.md`
- **Inspectors Brief**: `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/08-cleanup-inspectors-brief-task-18.md`

