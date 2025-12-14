---
Title: Cleanup Inspection Report — Final Verification (Post-CLEANUP-LEGACY-WALKERS)
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/reference/01-diary.md
      Note: Diary documenting all cleanup work completed
    - Path: internal/workspace/discovery.go
      Note: Legacy helpers removed; only FindTicketScaffoldsMissingIndex remains (diagnostic-only)
    - Path: pkg/commands/doc_move.go
      Note: Contains shared resolveTicketDirViaWorkspace helper (canonical replacement)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T10:45:00.000000000-05:00
---

# Cleanup Inspection Report — Final Verification

**Date**: 2025-12-13  
**Inspector**: AI Assistant (re-running Task 18 inspection brief)  
**Context**: Post-CLEANUP-LEGACY-WALKERS completion verification

## Executive Summary

✅ **Primary cleanup objectives COMPLETE**: All legacy discovery helpers (`findTicketDirectory`, `CollectTicketWorkspaces`, `CollectTicketScaffoldsWithoutIndex`) have been successfully removed or migrated. The codebase now uses `Workspace.QueryDocs` as the single canonical source for document discovery and enumeration.

✅ **Remaining filesystem operations are LEGITIMATE**: All remaining `filepath.Walk`/`WalkDir` usage is for write-path operations (file renaming, frontmatter updates) or diagnostic scans (missing index detection), which are inherently filesystem-level and cannot be replaced by QueryDocs.

✅ **Frontmatter parsing is LEGITIMATE**: All remaining `readDocumentFrontmatter`/`ReadDocumentWithFrontmatter` usage is in write-path commands that need to read, modify, and write frontmatter back to disk. These are not discovery operations.

## Verification Methodology

1. **Grep verification** of all search patterns from the original brief
2. **Code review** of remaining matches to confirm legitimacy
3. **Cross-reference** with CLEANUP-LEGACY-WALKERS diary to confirm completion
4. **Pattern analysis** to identify any remaining duplication risks

---

## Inventory: Legacy Helpers Status

### ✅ DELETED: `findTicketDirectory`

| Field | Value |
|-------|-------|
| **Original Location** | `pkg/commands/import_file.go:112` |
| **Status** | **DELETED** (Phase 4.1, commit `3751433`) |
| **Replacement** | `resolveTicketDirViaWorkspace(ctx, ws, ticketID)` in `pkg/commands/doc_move.go:210` |
| **Remaining Callsites** | **0** (verified via grep) |

**Verification**: No production callsites found. All commands now use Workspace-backed ticket resolution.

---

### ✅ DELETED: `workspace.CollectTicketWorkspaces`

| Field | Value |
|-------|-------|
| **Original Location** | `internal/workspace/discovery.go` |
| **Status** | **DELETED** (Phase 4.2, commit `37d68a4`) |
| **Replacement** | `ws.QueryDocs(ScopeRepo, Filters{DocType=index})` |
| **Remaining Callsites** | **0** (verified via grep) |

**Verification**: Deleted from `internal/workspace/discovery.go`. All ticket enumeration now uses QueryDocs.

---

### ✅ RENAMED: `CollectTicketScaffoldsWithoutIndex` → `FindTicketScaffoldsMissingIndex`

| Field | Value |
|-------|-------|
| **Original Location** | `internal/workspace/discovery.go` |
| **Status** | **RENAMED** (Phase 4.2, commit `37d68a4`) |
| **New Location** | `internal/workspace/discovery.go:18` |
| **Category** | **diagnostic** (not discovery) |
| **Justification** | This is a filesystem-level diagnostic scan for tickets missing `index.md`. QueryDocs cannot represent this state (no doc to index), so a filesystem walk is required. The rename clarifies this is a diagnostic tool, not a general-purpose discovery helper. |
| **Action** | **KEEP** (with justification) |

**Verification**: Only used by `doctor.go` for missing-index detection. This is legitimate and cannot be replaced by QueryDocs.

---

## Inventory: Remaining Filesystem Operations

### ✅ LEGITIMATE: `filepath.Walk` / `filepath.WalkDir` Usage

All remaining filesystem walks are **legitimate** and fall into these categories:

#### Category 1: Write-Path Operations (File Renaming/Moving)

| Location | Function | Purpose | Justification |
|----------|----------|---------|---------------|
| `pkg/commands/layout_fix.go:166` | `applyLayoutFix` | Renames files and updates intra-ticket references | Write-path operation; QueryDocs cannot perform renames |
| `pkg/commands/renumber.go:91,161` | `applyRenumber` | Resequences numeric prefixes and rewrites references | Write-path operation; QueryDocs cannot perform renames |

**Action**: **KEEP** — These are write-path operations that require filesystem-level file moves/renames.

#### Category 2: Diagnostic Scans (Missing Index Detection)

| Location | Function | Purpose | Justification |
|----------|----------|---------|---------------|
| `internal/workspace/discovery.go:20` | `FindTicketScaffoldsMissingIndex` | Detects ticket-like dirs missing `index.md` | Diagnostic scan; QueryDocs cannot represent missing docs |
| `pkg/commands/doctor.go:843` | `findIndexFiles` | Finds all `index.md` files for validation | Diagnostic scan; used for `--all` mode validation |

**Action**: **KEEP** — These are diagnostic operations that require filesystem-level scanning.

#### Category 3: Non-Document Operations

| Location | Function | Purpose | Justification |
|----------|----------|---------|---------------|
| `pkg/commands/template_validate.go:118` | `Run` | Scans templates directory (`.templ` files) | Not a document operation; templates are not indexed |

**Action**: **KEEP** — This operates on template files, not documents.

---

## Inventory: Remaining Frontmatter Parsing

### ✅ LEGITIMATE: `readDocumentFrontmatter` / `ReadDocumentWithFrontmatter` Usage

All remaining frontmatter parsing is **legitimate** and occurs in write-path commands that need to:
1. Read existing frontmatter
2. Modify it
3. Write it back to disk

| Location | Function | Purpose | Justification |
|----------|----------|---------|---------------|
| `pkg/commands/add.go:210` | `createDocument` | Reads ticket `index.md` to seed new doc metadata | Write-path: creating new doc |
| `pkg/commands/doctor.go:409,635,1023` | Various | Reads frontmatter for validation/diagnostics | Diagnostic: needs line/col info for errors |
| `pkg/commands/search.go:319,1256` | `RunIntoGlazeProcessor` | Reads frontmatter for external source filtering | Post-filter on QueryDocs results; external sources not indexed |
| `pkg/commands/meta_update.go:309` | `applyMetaUpdate` | Reads/modifies/writes frontmatter | Write-path: updating frontmatter |
| `pkg/commands/import_file.go:210` | `importFile` | Reads/modifies ticket `index.md` | Write-path: updating index.md |
| `pkg/commands/layout_fix.go:194` | `applyLayoutFix` | Reads frontmatter for reference rewriting | Write-path: updating references |
| `pkg/commands/ticket_close.go:133,251` | `RunIntoGlazeProcessor` / `Run` | Reads/modifies `index.md` and `changelog.md` | Write-path: updating control docs |
| `pkg/commands/doc_move.go:146` | `applyMove` | Reads source doc frontmatter | Write-path: moving doc |
| `pkg/commands/relate.go:422` | `applyRelate` | Reads/modifies target doc frontmatter | Write-path: updating RelatedFiles |

**Action**: **KEEP** — All are write-path operations that require reading and writing frontmatter. QueryDocs provides discovery, but write operations need to parse/write files.

---

## Inventory: Canonical Document Walker

### ✅ LEGITIMATE: `documents.WalkDocuments` Usage

| Location | Function | Purpose | Justification |
|----------|----------|---------|---------------|
| `internal/workspace/index_builder.go:99` | `InitIndex` | Canonical ingestion walker | **This IS the canonical walker** used by Workspace ingestion |
| `pkg/commands/rename_ticket.go:162` | `updateTicketFrontmatter` | Updates Ticket field across all docs in ticket | Write-path: bulk frontmatter update |

**Action**: **KEEP** — `index_builder.go` usage is the canonical ingestion path. `rename_ticket.go` usage is a write-path bulk update operation.

---

## Potential Minor Improvements (Non-Critical)

### Observation: Duplicate Ticket Directory Resolution Helpers

There are two similar helpers for resolving ticket directories via Workspace:

1. **`findTicketDirectoryViaWorkspace`** in `pkg/commands/add.go:378`
   - Returns `(ticketDir, resolvedRoot, error)`
   - Used by `add.go` only

2. **`resolveTicketDirViaWorkspace`** in `pkg/commands/doc_move.go:210`
   - Returns `(ticketDir, error)`
   - Used by 10+ commands (tasks, import_file, changelog, rename_ticket, layout_fix, renumber, ticket_move, ticket_close, doc_move)

**Analysis**: The `doc_move.go` helper is the de facto standard (used by most commands). The `add.go` helper differs only in also returning `resolvedRoot`, which `add.go` uses for template/guideline lookup.

**Recommendation**: **LOW PRIORITY** — Consider consolidating these helpers in a future cleanup, but this is not a critical issue. The duplication is minimal and both are Workspace-backed (not legacy).

**Action**: **DEFER** — Not a blocker for Task 18 completion.

---

## Cleanup Guidelines Verification

### ✅ Guideline 1: Ticket Discovery

**Guideline**: "Prefer `ws.QueryDocs(ScopeTicket, DocType=index)` + selecting `index.md`"

**Status**: **COMPLETE** — All commands now use `resolveTicketDirViaWorkspace` or equivalent QueryDocs-based resolution.

**Exceptions**: None.

---

### ✅ Guideline 2: Doc Enumeration

**Guideline**: "Prefer `ws.InitIndex(...)` + `ws.QueryDocs(...)` for doc sets"

**Status**: **COMPLETE** — All read-path commands (`list docs`, `list tickets`, `search`, `doctor`, `status`) use QueryDocs.

**Exceptions**: Write-path operations (`renumber`, `layout_fix`) still use `filepath.WalkDir` for file renaming, which is legitimate.

---

### ✅ Guideline 3: Manual Skip Rules

**Guideline**: "Prefer Workspace ingest skip + path tags"

**Status**: **COMPLETE** — All discovery uses canonical skip policy via QueryDocs.

**Exceptions**: `doctor --ignore-glob` applies post-filters on QueryDocs results (as intended per guideline).

---

### ✅ Guideline 4: Manual Frontmatter Parsing

**Guideline**: "Prefer the index's `parse_ok` / `parse_err` fields surfaced through QueryDocs"

**Status**: **COMPLETE** — All discovery uses QueryDocs. Remaining parsing is for write-path operations.

**Exceptions**: Write-path commands (`add`, `meta_update`, `import_file`, `ticket_close`, `doc_move`, `relate`, `layout_fix`) parse frontmatter to modify and write it back, which is legitimate.

---

## Risk Assessment

### ✅ No High-Risk Items Remaining

All remaining filesystem operations and frontmatter parsing are:
1. **Legitimate** (write-path or diagnostic operations)
2. **Well-scoped** (not duplicating discovery logic)
3. **Documented** (clear purpose in code comments)

### Low-Risk Observations

1. **Helper consolidation opportunity**: Two similar ticket-dir resolution helpers could be consolidated, but this is low priority and not a blocker.

---

## Validation Checklist

- [x] **Inventory table filled** for all search patterns
- [x] **Clear mapping** from duplicates to Workspace-based replacements
- [x] **Legitimate exceptions** identified and justified
- [x] **Risks called out** (none critical)

---

## Final Verdict

✅ **Task 18 is COMPLETE**

All legacy discovery helpers have been removed or migrated. All remaining filesystem operations and frontmatter parsing are legitimate write-path or diagnostic operations that cannot be replaced by QueryDocs.

**Recommendation**: **CLOSE** Task 18 and the CLEANUP-LEGACY-WALKERS ticket. The codebase now has a single canonical discovery path via `Workspace.QueryDocs`, with legitimate filesystem operations clearly separated and justified.

---

## Sign-Off

**Inspector**: AI Assistant  
**Date**: 2025-12-13  
**Status**: ✅ **VERIFIED COMPLETE**

All cleanup objectives met. No critical issues found. Minor consolidation opportunities identified but not blockers.

