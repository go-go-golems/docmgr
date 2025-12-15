---
Title: Cleanup Inventory Report - Task 18
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/discovery.go
      Note: Legacy discovery helpers (CollectTicketWorkspaces
    - Path: pkg/commands/import_file.go
      Note: Defines findTicketDirectory helper (prime cleanup target)
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/08-cleanup-inspectors-brief-task-18.md
      Note: Original brief for cleanup inspectors
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T10:23:25.535621879-05:00
---




# Cleanup Inventory Report — Task 18

**Mission**: Inventory all duplicated walkers, helpers, and discovery logic that should be migrated to the canonical `Workspace` + `QueryDocs` API.

**Status**: Complete inventory — ready for implementation PRs.

**Reference**: See `analysis/08-cleanup-inspectors-brief-task-18.md` for the original brief and methodology.

---

## Executive Summary

This report inventories **23 cleanup targets** across **12 command files** and **1 core helper module**. The cleanup is organized into **5 categories**:

- **Discovery** (ticket/doc discovery): 8 targets
- **Traversal** (filesystem walking): 7 targets  
- **Filtering/Skip Rules** (manual skip logic): 3 targets
- **Parsing** (manual frontmatter parsing): 4 targets
- **Normalization** (path resolution): 1 target

**Priority**: High-impact targets are marked with ⚠️. These are commands that re-implement core discovery/traversal logic and should be migrated first to establish the Workspace API as the single source of truth.

---

## Inventory Table

### Category: Discovery

#### 1. `findTicketDirectory` helper (PRIME TARGET)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/import_file.go:112` |
| **Category** | discovery |
| **Current behavior** | Calls `workspace.CollectTicketWorkspaces(root, nil)`, then iterates to find first match where `ws.Doc.Ticket == ticket`. Returns directory path or error. Used by 15+ command files. |
| **Proposed canonical replacement** | `ws.QueryDocs(ScopeTicket, Ticket=ticketID, DocType=index)` → select `index.md` → infer ticket dir from doc path (`filepath.Dir(indexPath)`) |
| **Migration note** | This is the **most duplicated helper**. All 15+ call sites should migrate to Workspace-based discovery. Behavior change: currently returns first match silently; Workspace API can surface duplicates via diagnostics. |
| **Action** | **delete** (after migrating all callers) |
| **Risk** | Medium — many call sites, but migration is straightforward |

**Call sites** (15 files):
- `pkg/commands/add.go:180`
- `pkg/commands/changelog.go:99, 115, 386, 399`
- `pkg/commands/doc_move.go:144, 148`
- `pkg/commands/import_file.go:166` (self-reference)
- `pkg/commands/meta_update.go:184`
- `pkg/commands/renumber.go:126`
- `pkg/commands/rename_ticket.go:93, 184`
- `pkg/commands/search.go:449, 1038`
- `pkg/commands/tasks.go:53`
- `pkg/commands/ticket_close.go:110, 211`
- `pkg/commands/ticket_move.go:114`

---

#### 2. ⚠️ `workspace.CollectTicketWorkspaces` usage in `status.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/status.go:156, 357` |
| **Category** | discovery |
| **Current behavior** | Calls `workspace.CollectTicketWorkspaces(settings.Root, nil)` to enumerate tickets, then walks each ticket directory with `filepath.Walk` to count docs. Filters by `--ticket` manually. |
| **Proposed canonical replacement** | `ws.InitIndex(...)` → `ws.QueryDocs(ScopeTicket, Ticket=filter, DocType=...)` → aggregate counts from query results |
| **Migration note** | Currently does **double work**: discovers tickets via `CollectTicketWorkspaces`, then walks filesystem again. Should use QueryDocs for both ticket discovery and doc enumeration. |
| **Action** | **wrap** → migrate to QueryDocs |
| **Risk** | Low — output shape can be preserved |

---

#### 3. ⚠️ `workspace.CollectTicketWorkspaces` usage in `list_tickets.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/list_tickets.go:174, 282` |
| **Category** | discovery |
| **Current behavior** | Calls `workspace.CollectTicketWorkspaces(settings.Root, nil)`, filters by `--ticket` (substring match) and `--status` (exact match), sorts by `LastUpdated`. |
| **Proposed canonical replacement** | `ws.InitIndex(...)` → `ws.QueryDocs(ScopeTicket, Ticket=filter, Status=filter)` → sort results |
| **Migration note** | Currently uses substring match for `--ticket` (`strings.Contains`). Workspace API uses exact match. **Behavior change**: ticket filter semantics will change from substring to exact match. |
| **Action** | **wrap** → migrate to QueryDocs |
| **Risk** | Medium — ticket filter semantics change (substring → exact) |

---

#### 4. `workspace.CollectTicketScaffoldsWithoutIndex` usage in `doctor.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/doctor.go:298` |
| **Category** | discovery |
| **Current behavior** | Calls `workspace.CollectTicketScaffoldsWithoutIndex(settings.Root, skipFn)` to detect ticket-like directories missing `index.md`. Filters by `--ticket` manually. |
| **Proposed canonical replacement** | Keep as-is OR extend Workspace to detect missing-index scaffolds via filesystem scan (this is inherently filesystem-level, not index-backed). |
| **Migration note** | This is a **filesystem-level check** (detects dirs with scaffold markers but no index.md). The index can't help here because broken/missing docs aren't indexed. Consider keeping this helper but moving it under Workspace for consistency. |
| **Action** | **keep** (with justification) OR **move** to `workspace.DiscoverWorkspace` as an optional check |
| **Risk** | Low — this is inherently filesystem-level |

---

#### 5. `findTicketDirectory` usage in `changelog.go` (suggestion mode)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/changelog.go:99, 115, 386, 399` |
| **Category** | discovery |
| **Current behavior** | Uses `findTicketDirectory` to resolve ticket dir for suggestion search root. Then walks filesystem with `filepath.Walk` to collect RelatedFiles from existing docs. |
| **Proposed canonical replacement** | `ws.InitIndex(...)` → `ws.QueryDocs(ScopeTicket, Ticket=ticketID)` → extract RelatedFiles from query results |
| **Migration note** | Suggestion mode currently walks filesystem to find docs. Should use QueryDocs instead. |
| **Action** | **wrap** → migrate to QueryDocs |
| **Risk** | Low |

---

#### 6. `findTicketDirectory` usage in `add.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/add.go:180` |
| **Category** | discovery |
| **Current behavior** | Uses `findTicketDirectory` to locate ticket dir, then creates new doc under `<ticketDir>/<docType>/`. |
| **Proposed canonical replacement** | `ws.QueryDocs(ScopeTicket, Ticket=ticketID, DocType=index)` → select index.md → infer ticket dir |
| **Migration note** | Write-path operation. Discovery should be canonical, but file creation is inherently filesystem-level. |
| **Action** | **wrap** → migrate discovery to QueryDocs |
| **Risk** | Low |

---

#### 7. `findTicketDirectory` usage in `meta_update.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/meta_update.go:184` |
| **Category** | discovery |
| **Current behavior** | Uses `findTicketDirectory` to locate ticket dir, then calls `findMarkdownFiles(ticketDir, docType)` to enumerate docs. |
| **Proposed canonical replacement** | `ws.InitIndex(...)` → `ws.QueryDocs(ScopeTicket, Ticket=ticketID, DocType=docType)` |
| **Migration note** | Currently walks filesystem to find markdown files. Should use QueryDocs. |
| **Action** | **wrap** → migrate to QueryDocs |
| **Risk** | Low |

---

#### 8. `findTicketDirectory` usage in `tasks.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/tasks.go:53` |
| **Category** | discovery |
| **Current behavior** | Uses `findTicketDirectory` as fallback after trying directory-name-based heuristics. |
| **Proposed canonical replacement** | `ws.QueryDocs(ScopeTicket, Ticket=ticketID, DocType=index)` → infer ticket dir |
| **Migration note** | Write-path operation (reads/writes `tasks.md`). Discovery should be canonical. |
| **Action** | **wrap** → migrate discovery to QueryDocs |
| **Risk** | Low |

---

### Category: Traversal

#### 9. ⚠️ `filepath.Walk` in `status.go` (doc counting)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/status.go:188, 387, 465` (3 occurrences) |
| **Category** | traversal |
| **Current behavior** | Walks ticket directory with `filepath.Walk`, skips `index.md`, parses frontmatter for each `.md` file, counts by `DocType`. |
| **Proposed canonical replacement** | `ws.QueryDocs(ScopeTicket, Ticket=ticketID)` → aggregate counts from query results |
| **Migration note** | Currently walks filesystem and parses frontmatter manually. Should use QueryDocs. **Behavior change**: currently skips parse errors silently; QueryDocs can surface them via diagnostics. |
| **Action** | **delete** → replace with QueryDocs |
| **Risk** | Low — output shape can be preserved |

---

#### 10. ⚠️ `filepath.Walk` in `changelog.go` (suggestion mode)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/changelog.go:122` |
| **Category** | traversal |
| **Current behavior** | Walks search root (ticket dir or docs root) with `filepath.Walk`, parses frontmatter for each `.md`, extracts RelatedFiles, filters by topics if provided. |
| **Proposed canonical replacement** | `ws.InitIndex(...)` → `ws.QueryDocs(ScopeTicket|ScopeRepo, TopicsAny=topics)` → extract RelatedFiles from query results |
| **Migration note** | Currently walks filesystem and parses frontmatter manually. Should use QueryDocs. |
| **Action** | **delete** → replace with QueryDocs |
| **Risk** | Low |

---

#### 11. ⚠️ `filepath.Walk` in `search.go` (suggestion mode)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/search.go:473` |
| **Category** | traversal |
| **Current behavior** | Walks ticket directory (or root) with `filepath.Walk`, parses frontmatter for each `.md`, extracts RelatedFiles. Used in `--suggest` mode. |
| **Proposed canonical replacement** | `ws.InitIndex(...)` → `ws.QueryDocs(ScopeTicket|ScopeRepo, TopicsAny=topics)` → extract RelatedFiles |
| **Migration note** | **Note**: `search.go` was already ported to QueryDocs for main search path (Step 14 in diary). This is the **suggestion mode** path that still uses filesystem walking. |
| **Action** | **delete** → replace with QueryDocs |
| **Risk** | Low — main search path already uses QueryDocs |

---

#### 12. `filepath.WalkDir` in `renumber.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/renumber.go:91, 149` |
| **Category** | traversal |
| **Current behavior** | Walks ticket directory with `filepath.WalkDir` to update ticket references in markdown files. Also walks subdirectories to collect files for renumbering. |
| **Proposed canonical replacement** | **Keep as-is** — this is a **write-path operation** that modifies file contents. Filesystem traversal is appropriate here. Discovery of ticket dir should use QueryDocs. |
| **Migration note** | Write-path operations (file content modification) are inherently filesystem-level. Only the ticket discovery should migrate to QueryDocs. |
| **Action** | **keep** (with justification) — only migrate ticket discovery |
| **Risk** | Low — write-path operations are filesystem-level |

---

#### 13. `filepath.WalkDir` in `layout_fix.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/layout_fix.go:129` |
| **Category** | traversal |
| **Current behavior** | Walks ticket directory with `filepath.WalkDir` to fix layout issues (renumbering, moving files). |
| **Proposed canonical replacement** | **Keep as-is** — write-path operation. Only migrate ticket discovery. |
| **Migration note** | Write-path operations are filesystem-level. Only ticket discovery should migrate. |
| **Action** | **keep** (with justification) — only migrate ticket discovery |
| **Risk** | Low |

---

#### 14. `filepath.WalkDir` in `rename_ticket.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/rename_ticket.go:151` |
| **Category** | traversal |
| **Current behavior** | Uses `documents.WalkDocuments` (canonical walker) to update ticket references in markdown files. |
| **Proposed canonical replacement** | **Keep as-is** — `documents.WalkDocuments` is the canonical walker. This is appropriate for write-path operations. |
| **Migration note** | `documents.WalkDocuments` is the canonical walker and is appropriate for write-path operations. No change needed. |
| **Action** | **keep** (with justification) |
| **Risk** | None |

---

#### 15. `filepath.Walk` in `doctor.go` (findIndexFiles)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/doctor.go:843` |
| **Category** | traversal |
| **Current behavior** | Recursively searches for `index.md` files. Used for detecting missing index.md in ticket directories. |
| **Proposed canonical replacement** | **Keep as-is** OR use `ws.QueryDocs(ScopeRepo, DocType=index)` to find all index.md files. |
| **Migration note** | This is a filesystem-level check for missing index.md. Could use QueryDocs, but the check is inherently about "what's missing", so filesystem scan may be more appropriate. |
| **Action** | **keep** (with justification) OR migrate to QueryDocs if it simplifies the logic |
| **Risk** | Low |

---

### Category: Filtering/Skip Rules

#### 16. Manual skip logic in `status.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/status.go:198` |
| **Category** | filtering/skip |
| **Current behavior** | Manually skips `index.md` files during `filepath.Walk`. |
| **Proposed canonical replacement** | QueryDocs automatically excludes `index.md` when `IsIndex` filter is not set, OR use post-filter on query results. |
| **Migration note** | Skip logic should be handled by Workspace ingest skip policy + QueryDocs filters. |
| **Action** | **delete** → handled by QueryDocs |
| **Risk** | Low |

---

#### 17. Manual skip logic in `renumber.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/renumber.go:143` |
| **Category** | filtering/skip |
| **Current behavior** | Manually skips directories starting with `_` and `.meta` during traversal. |
| **Proposed canonical replacement** | **Keep as-is** — write-path operation. Skip logic is appropriate for filesystem traversal. |
| **Migration note** | Write-path operations need filesystem traversal. Skip logic is appropriate here. |
| **Action** | **keep** (with justification) |
| **Risk** | None |

---

#### 18. Manual skip logic in `changelog.go` (suggestion mode)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/changelog.go:123` |
| **Category** | filtering/skip |
| **Current behavior** | Skips directories and non-markdown files during `filepath.Walk`. |
| **Proposed canonical replacement** | QueryDocs handles skip logic automatically. |
| **Migration note** | Should use QueryDocs instead of manual walking. |
| **Action** | **delete** → handled by QueryDocs |
| **Risk** | Low |

---

### Category: Parsing

#### 19. ⚠️ Manual frontmatter parsing in `status.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/status.go:201, 394, 472` |
| **Category** | parsing |
| **Current behavior** | Calls `readDocumentFrontmatter(path)` for each `.md` file during `filepath.Walk`. |
| **Proposed canonical replacement** | QueryDocs returns parsed frontmatter from index. |
| **Migration note** | Currently parses frontmatter manually. QueryDocs provides parsed frontmatter from index. |
| **Action** | **delete** → use QueryDocs results |
| **Risk** | Low |

---

#### 20. Manual frontmatter parsing in `changelog.go` (suggestion mode)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/changelog.go:126` |
| **Category** | parsing |
| **Current behavior** | Calls `readDocumentFrontmatter(path)` during `filepath.Walk`. |
| **Proposed canonical replacement** | QueryDocs provides parsed frontmatter. |
| **Migration note** | Should use QueryDocs instead of manual parsing. |
| **Action** | **delete** → use QueryDocs results |
| **Risk** | Low |

---

#### 21. Manual frontmatter parsing in `search.go` (suggestion mode)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/search.go:486, 508` |
| **Category** | parsing |
| **Current behavior** | Calls `readDocumentFrontmatter(path)` during `filepath.Walk` in suggestion mode. |
| **Proposed canonical replacement** | QueryDocs provides parsed frontmatter. |
| **Migration note** | Main search path already uses QueryDocs. Suggestion mode should migrate too. |
| **Action** | **delete** → use QueryDocs results |
| **Risk** | Low |

---

#### 22. Manual frontmatter parsing in `doctor.go`

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/doctor.go:409, 635, 1023` |
| **Category** | parsing |
| **Current behavior** | Calls `readDocumentFrontmatter(path)` for missing-index checks and RelatedFiles validation. |
| **Proposed canonical replacement** | **Keep as-is** — `doctor` already uses QueryDocs for main doc enumeration (Step 15 in diary). These are edge-case checks that need manual parsing for diagnostics. |
| **Migration note** | `doctor` already migrated to QueryDocs. These are edge-case checks (missing index, RelatedFiles validation) that may need manual parsing for line/col diagnostics. |
| **Action** | **keep** (with justification) — edge-case diagnostics |
| **Risk** | None |

---

### Category: Normalization

#### 23. Path resolution in `changelog.go` (suggestion mode)

| Field | Value |
|-------|-------|
| **Location** | `pkg/commands/changelog.go:115` |
| **Category** | normalization |
| **Current behavior** | Uses `findTicketDirectory` to resolve ticket dir, then uses that as search root. |
| **Proposed canonical replacement** | `ws.QueryDocs(ScopeTicket, Ticket=ticketID)` → use Workspace context for path resolution |
| **Migration note** | Should use Workspace resolver for consistent path normalization. |
| **Action** | **wrap** → use Workspace resolver |
| **Risk** | Low |

---

## Cleanup Guidelines

### When you see X, replace with Y

#### Ticket Discovery

**Pattern**: `findTicketDirectory(root, ticket)`

**Replacement**:
```go
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: root})
if err != nil {
    return err
}
if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{}); err != nil {
    return err
}
docs, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.ScopeTicket,
    Ticket: ticket,
    DocType: "index",
})
if err != nil || len(docs) == 0 {
    return fmt.Errorf("ticket not found: %s", ticket)
}
indexDoc := docs[0]
ticketDir := filepath.Dir(indexDoc.Path)
```

**Rationale**: Single canonical discovery path via Workspace index.

---

#### Doc Enumeration

**Pattern**: `filepath.Walk(root, func(...) { readDocumentFrontmatter(...) })`

**Replacement**:
```go
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: root})
if err != nil {
    return err
}
if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{}); err != nil {
    return err
}
docs, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.ScopeTicket, // or ScopeRepo
    Ticket: ticket, // optional
    DocType: docType, // optional
    TopicsAny: topics, // optional
})
```

**Rationale**: Single canonical traversal via Workspace index. Skip rules, parsing, and filtering are handled consistently.

---

#### Manual Skip Rules

**Pattern**: Manual checks like `strings.HasPrefix(name, "_")` or `name == ".meta"`

**Replacement**: QueryDocs automatically applies canonical skip rules. Use `IncludeArchivedPath`, `IncludeScriptsPath`, etc. to control what's included.

**Rationale**: Skip rules are centralized in `internal/workspace/skip_policy.go`. Commands should not re-implement them.

---

#### Manual Frontmatter Parsing

**Pattern**: `readDocumentFrontmatter(path)` or `documents.ReadDocumentWithFrontmatter(path)`

**Replacement**: QueryDocs returns parsed frontmatter in `DocHandle.Doc`. Only re-parse if you need line/col diagnostics.

**Rationale**: Frontmatter is parsed once during ingestion. QueryDocs provides parsed results.

---

#### Path Normalization

**Pattern**: Manual path resolution or `filepath.Join`/`filepath.Rel` without Workspace context

**Replacement**: Use `ws.Resolver()` or `paths.NewResolver(ResolverOptions{DocPath: docPath, RepoRoot: ws.Context().RepoRoot})`

**Rationale**: Consistent normalization anchors (repo root, docs root, config dir, doc dir) ensure reverse lookup correctness.

---

## PR Plan (Recommended Sequencing)

### Phase 1: Low-Risk, High-Impact (Foundation)

**PR 1.1**: Migrate `status.go` to QueryDocs
- **Target**: `pkg/commands/status.go`
- **Changes**: Replace `CollectTicketWorkspaces` + `filepath.Walk` with `QueryDocs`
- **Risk**: Low — output shape preserved
- **Validation**: Run `docmgr status` in both human and glaze modes, compare output

**PR 1.2**: Migrate `list_tickets.go` to QueryDocs
- **Target**: `pkg/commands/list_tickets.go`
- **Changes**: Replace `CollectTicketWorkspaces` with `QueryDocs`
- **Risk**: Medium — ticket filter semantics change (substring → exact)
- **Validation**: Run `docmgr list tickets --ticket PARTIAL` and verify behavior change is acceptable

**PR 1.3**: Migrate `changelog.go` suggestion mode to QueryDocs
- **Target**: `pkg/commands/changelog.go`
- **Changes**: Replace `filepath.Walk` + manual parsing with `QueryDocs`
- **Risk**: Low
- **Validation**: Run `docmgr changelog update --ticket X --suggest` and verify suggestions match

---

### Phase 2: Medium-Risk (Write-Path Discovery)

**PR 2.1**: Migrate `add.go` ticket discovery to QueryDocs
- **Target**: `pkg/commands/add.go`
- **Changes**: Replace `findTicketDirectory` with `QueryDocs`-based discovery
- **Risk**: Low — write-path operation, discovery only changes
- **Validation**: Run `docmgr doc add --ticket X --doc-type analysis --title "Test"`

**PR 2.2**: Migrate `meta_update.go` to QueryDocs
- **Target**: `pkg/commands/meta_update.go`
- **Changes**: Replace `findTicketDirectory` + `findMarkdownFiles` with `QueryDocs`
- **Risk**: Low
- **Validation**: Run `docmgr meta update --ticket X --field Status --value review`

**PR 2.3**: Migrate `tasks.go` ticket discovery to QueryDocs
- **Target**: `pkg/commands/tasks.go`
- **Changes**: Replace `findTicketDirectory` fallback with `QueryDocs`
- **Risk**: Low
- **Validation**: Run `docmgr task list --ticket X`

---

### Phase 3: High-Risk (Behavior-Sensitive)

**PR 3.1**: Migrate `search.go` suggestion mode to QueryDocs
- **Target**: `pkg/commands/search.go` (suggestion path only)
- **Changes**: Replace `filepath.Walk` + manual parsing with `QueryDocs`
- **Risk**: Low — main search path already uses QueryDocs
- **Validation**: Run `docmgr doc search --suggest --ticket X`

**PR 3.2**: Migrate remaining `findTicketDirectory` callers
- **Targets**: `doc_move.go`, `ticket_move.go`, `ticket_close.go`, `rename_ticket.go`, `renumber.go`, `layout_fix.go`
- **Changes**: Replace `findTicketDirectory` with `QueryDocs`-based discovery
- **Risk**: Medium — write-path operations, but discovery is straightforward
- **Validation**: Run each command with `--ticket X` and verify behavior

---

### Phase 4: Cleanup (Delete Legacy Helpers)

**PR 4.1**: Delete `findTicketDirectory` helper
- **Target**: `pkg/commands/import_file.go:112`
- **Changes**: Remove function definition
- **Risk**: Low — all callers migrated
- **Validation**: `go test ./...` and scenario suite

**PR 4.2**: Consider deprecating `CollectTicketWorkspaces` (if unused)
- **Target**: `internal/workspace/discovery.go:26`
- **Changes**: Mark as deprecated OR delete if truly unused
- **Risk**: Medium — verify no external callers
- **Validation**: Search codebase for remaining callers

---

## Risk Notes

### Behavior Changes

1. **Ticket filter semantics** (`list_tickets.go`):
   - **Current**: Substring match (`strings.Contains(doc.Ticket, filter)`)
   - **After**: Exact match (`Ticket == filter`)
   - **Impact**: Users using partial ticket IDs may see different results
   - **Mitigation**: Document change, consider adding `--ticket-contains` flag if needed

2. **Parse error handling** (`status.go`):
   - **Current**: Silently skips docs with parse errors
   - **After**: QueryDocs can surface parse errors via diagnostics
   - **Impact**: More visibility into broken docs (positive change)
   - **Mitigation**: Use `IncludeErrors=false` by default to preserve current behavior

3. **Missing index.md detection** (`doctor.go`):
   - **Current**: Uses `CollectTicketScaffoldsWithoutIndex` (filesystem scan)
   - **After**: Could use QueryDocs, but missing docs aren't indexed
   - **Impact**: May need to keep filesystem scan for this check
   - **Mitigation**: Keep `CollectTicketScaffoldsWithoutIndex` OR extend Workspace to handle this

### Performance Considerations

1. **Index build overhead**: Commands that currently do lightweight filesystem walks will now build a full SQLite index. This is a one-time cost per command invocation.
   - **Mitigation**: Consider caching index across command invocations (future work)

2. **Memory usage**: SQLite index holds parsed frontmatter in memory. For very large repos, this could be significant.
   - **Mitigation**: Current implementation is in-memory only. Consider persistence if needed (future work)

### Compatibility

1. **External callers**: `CollectTicketWorkspaces` may be used by external tools or scripts.
   - **Mitigation**: Keep function but mark as deprecated, or verify no external usage

2. **Write-path operations**: Commands that modify files (`renumber`, `layout_fix`, `rename_ticket`) will still need filesystem traversal.
   - **Mitigation**: Only migrate discovery, keep filesystem traversal for write operations

---

## Validation Strategy

### Unit Tests

For each PR:
1. Run `go test ./...` to ensure no regressions
2. Add unit tests for QueryDocs-based paths if missing
3. Test edge cases (missing tickets, parse errors, empty results)

### Integration Tests

For each PR:
1. Run scenario suite: `bash test-scenarios/testing-doc-manager/run-all.sh`
2. Test both human and glaze output modes
3. Compare output before/after migration (where applicable)

### Manual Smoke Tests

For each PR:
1. Run the command in both modes (human/glaze)
2. Test with `--ticket` filter
3. Test with edge cases (missing ticket, parse errors, empty results)

---

## Deliverable Checklist

- [x] Inventory table filled for all matches of search patterns
- [x] Clear mapping from each duplicate to its Workspace-based replacement
- [x] Prioritized PR plan (what to do first, what to defer)
- [x] Risks called out explicitly (behavior changes, performance, compatibility)

---

## Next Steps

1. **Review this report** with implementers
2. **Start with Phase 1** (low-risk, high-impact migrations)
3. **Validate each PR** with scenario suite before merging
4. **Update diary** after each PR with lessons learned

---

## Related Files

- `analysis/08-cleanup-inspectors-brief-task-18.md` — Original brief
- `reference/15-diary.md` — Implementation diary (Steps 1-17)
- `internal/workspace/workspace.go` — Workspace API
- `internal/workspace/query_docs.go` — QueryDocs implementation
- `internal/workspace/discovery.go` — Legacy discovery helpers
- `pkg/commands/import_file.go` — `findTicketDirectory` definition
