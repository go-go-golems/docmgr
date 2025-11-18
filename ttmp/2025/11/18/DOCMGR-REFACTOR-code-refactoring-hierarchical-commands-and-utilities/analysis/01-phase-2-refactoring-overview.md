---
Title: Phase 2: Refactoring Overview
Ticket: DOCMGR-REFACTOR
Status: active
Topics:
    - docmgr
    - refactoring
    - architecture
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Hierarchical command structure, extract utilities to internal/, improve error handling
LastUpdated: 2025-11-18T12:20:00.000000000-05:00
---

# Phase 2: Refactoring Overview

## Goal

Restructure the codebase for better organization, reduce duplication, and improve maintainability. Implements the architectural decisions from Round 1.

## Context

From **Debate Round 10** synthesis:
- **Current structure**: Flat `pkg/commands/*.go` (29 files)
- **Target**: Hierarchical `cmd/docmgr/cmds/` + extracted `internal/`
- **Effort**: 2-4 weeks

## Tasks (7 items)

### Hierarchical Command Structure (Round 1)

1. ✅ **Create cmd/docmgr/cmds/ hierarchy**:
   ```
   cmd/docmgr/cmds/
   ├── doc/add/, doc/list/, doc/search/
   ├── ticket/create/, ticket/list/
   ├── vocab/add/, vocab/list/
   ├── changelog/update/
   └── ...
   ```

2. ✅ **Move helpers to internal/**:
   - `pkg/commands/config.go` → `internal/workspace/config.go`
   - `pkg/commands/workspaces.go` → `internal/workspace/discovery.go`
   - `pkg/commands/templates.go` → `internal/templates/templates.go`
   - `pkg/commands/guidelines.go` → `internal/guidelines/`

**Impact**: Clear separation of CLI interface from business logic, better discoverability.

### Extract Utilities (Round 2 - Duplication Fixes)

3. ✅ **Extract internal/documents/frontmatter.go**:
   - `ReadDocumentWithFrontmatter(path) (*Document, string, error)`
   - `WriteDocumentWithFrontmatter(path, doc, body) error`
   - Consolidates 4 different frontmatter implementations

4. ✅ **Extract internal/documents/walk.go**:
   - `WalkDocuments(root, fn) error`
   - Consolidates 18 directory walk operations

5. ✅ **Migrate commands to use utilities**:
   - Update all 20+ commands to import `internal/documents`
   - Remove duplicated frontmatter/walk code
   - Test each migration

**Impact**: Eliminates duplication, single place to fix bugs, consistent behavior.

### Error Handling Improvements (Round 4)

6. ✅ **Add error context to bare returns**:
   - 72 `return err` statements need context
   - Pattern: `return fmt.Errorf("parsing settings: %w", err)`
   - Focus on high-traffic commands first

7. ✅ **Fix config.go silent error swallowing**:
   - Distinguish "not found" (expected) from "malformed" (unexpected)
   - Log warnings for malformed config
   - Use `LoadConfigOrWarn()` pattern

**Impact**: Better error messages help users debug issues quickly.

## Success Criteria

- [ ] Commands organized by domain (doc, ticket, vocab)
- [ ] Zero duplication in frontmatter parsing or directory walking
- [ ] Helpers extracted to `internal/`, commands are thin wrappers
- [ ] Error messages include context (which setting, which file, etc.)
- [ ] All tests pass after migration

## Migration Strategy

**Incremental approach** (from Sarah's recommendations):

1. **Extract utilities first** (week 1)
   - Create `internal/documents/`
   - Extract `frontmatter.go`, `walk.go`
   - Add tests

2. **Migrate 1-2 commands as pilot** (week 2)
   - Choose `add` and `search` (high usage)
   - Update to use `internal/documents`
   - Validate behavior unchanged

3. **Create hierarchical structure** (week 3)
   - Create `cmd/docmgr/cmds/` directories
   - Move commands incrementally
   - Update `main.go` registration

4. **Finish migration** (week 4)
   - Move remaining commands
   - Update error handling
   - Clean up old `pkg/commands/`

## References

- [Debate Round 1: Architecture Decision](../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/01-debate-round-01-architecture-and-code-organization.md)
- [Debate Round 2: Duplication Analysis](../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/02-debate-round-02-command-implementation-patterns.md)
- [Debate Round 4: Error Handling](../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/04-debate-round-04-error-handling-and-user-experience.md)
- [Debate Round 5: Package Boundaries](../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/05-debate-round-05-package-boundaries-and-internal-dependencies.md)

## Timeline

**Estimated effort**: 2-4 weeks
**Priority**: MEDIUM (builds on Phase 1, enables future work)
**Dependencies**: Phase 1 documentation helps during migration
