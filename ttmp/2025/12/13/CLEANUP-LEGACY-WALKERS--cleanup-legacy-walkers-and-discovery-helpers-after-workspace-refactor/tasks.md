# Tasks

## Phase 1: Foundation (Low Risk, High Impact)

- [x] [1] Migrate `status.go` to QueryDocs — Replace `CollectTicketWorkspaces` + `filepath.Walk` with `ws.QueryDocs`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 1.1
- [x] [2] Migrate `list_tickets.go` to QueryDocs — Replace `CollectTicketWorkspaces`. Note: ticket filter changes from substring to exact match. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 1.2
- [x] [3] Migrate `list.go` to QueryDocs — Replace `CollectTicketWorkspaces`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 1.3
- [x] [4] Migrate `changelog.go` suggestion mode to QueryDocs — Replace `filepath.Walk` + manual parsing. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 1.4

## Phase 2: Write-Path Discovery

- [x] [5] Migrate `add.go` ticket discovery to QueryDocs — Replace `findTicketDirectory`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 2.1
- [x] [6] Migrate `meta_update.go` to QueryDocs — Replace `findTicketDirectory` + `findMarkdownFiles`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 2.2
- [ ] [7] Migrate `tasks.go` ticket discovery to QueryDocs — Replace `findTicketDirectory` fallback. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 2.3

## Phase 3: Remaining Commands

- [ ] [8] Migrate `search.go` suggestion mode to QueryDocs — Main search path already done; suggestion mode still uses `filepath.Walk`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3.1
- [ ] [9] Migrate `doc_move.go` ticket discovery to QueryDocs — Replace both `findTicketDirectory` calls. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3.2
- [ ] [10] Migrate `ticket_move.go` ticket discovery to QueryDocs — Replace `findTicketDirectory`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3.2
- [ ] [11] Migrate `ticket_close.go` ticket discovery to QueryDocs — Replace both `findTicketDirectory` calls. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3.2
- [ ] [12] Migrate `rename_ticket.go` discovery to QueryDocs — Replace `findTicketDirectory` (keep `WalkDocuments` for write-path). See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3.3
- [ ] [13] Migrate `renumber.go` discovery to QueryDocs — Replace `findTicketDirectory` (keep `filepath.WalkDir` for write-path). See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3.3
- [ ] [14] Migrate `layout_fix.go` discovery to QueryDocs — Replace `findTicketDirectory` (keep `filepath.WalkDir` for write-path). See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3.3
- [ ] [15] Migrate `import_file.go` caller to QueryDocs — This file defines `findTicketDirectory` and also uses it. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 3

## Phase 4: Delete Legacy Helpers

- [ ] [16] Delete `findTicketDirectory` function — After all callers migrated, remove from `pkg/commands/import_file.go:112`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 4.1
- [ ] [17] Deprecate/delete `CollectTicketWorkspaces` — Verify no external callers, then deprecate or remove from `internal/workspace/discovery.go`. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 4.2
- [ ] [18] Deprecate/delete `CollectTicketScaffoldsWithoutIndex` — After `doctor.go` migration, assess if still needed. See: design/01-cleanup-overview-and-migration-guide.md §5 Phase 4.2

## Validation

- [ ] [19] Run integration test suite after Phase 1 — `bash test-scenarios/testing-doc-manager/run-all.sh`
- [ ] [20] Run integration test suite after Phase 2
- [ ] [21] Run integration test suite after Phase 3
- [ ] [22] Final integration test suite after Phase 4
- [ ] [23] Manual smoke tests for all migrated commands

