# Changelog

## 2025-12-13

- Initial workspace created


## 2025-12-13

Created cleanup ticket derived from REFACTOR-TICKET-REPOSITORY-HANDLING Task 18. Imported inventory of 23 cleanup targets across 12 command files. Defined phased PR plan (4 phases, 23 tasks) and migration patterns.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/design/01-cleanup-overview-and-migration-guide.md — Cleanup overview and migration guide


## 2025-12-13

Step 1: Migrated status.go to QueryDocs. Replaced CollectTicketWorkspaces + filepath.Walk with Workspace.QueryDocs for ticket discovery and doc enumeration. Preserved output format and behavior. Tests pass.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/status.go — Migrated to QueryDocs


## 2025-12-13

Step 1.1: Migrated status.go to Workspace.QueryDocs (commit f61606c)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/status.go — Migrated to QueryDocs; removed CollectTicketWorkspaces + filepath.Walk


## 2025-12-13

Step 1.2: Migrated list_tickets.go to Workspace.QueryDocs (commits 024993a, f23a876); spec: no backwards compatibility

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/list_tickets.go — Replaced CollectTicketWorkspaces; exact-match ticket filter
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/design/01-cleanup-overview-and-migration-guide.md — No backwards compatibility policy


## 2025-12-13

Step 1.3: Migrated list.go to Workspace.QueryDocs (commit 0ec09da)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/list.go — Replaced CollectTicketWorkspaces


## 2025-12-13

Diary: add 'what should be done in the future' sections for Steps 1–3

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/reference/01-diary.md — Added future-work guidance sections for reviewer triage


## 2025-12-13

Step 1.4: Migrated changelog.go suggestion doc-scan to Workspace.QueryDocs (commit 09e1e6f)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/changelog.go — Replaced filepath.Walk + readDocumentFrontmatter with QueryDocs


## 2025-12-13

Step 2.1: Migrated add.go ticket discovery to Workspace.QueryDocs (commit a512739)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/add.go — Replaced findTicketDirectory with QueryDocs


## 2025-12-13

Step 2.2: Migrated meta_update.go to Workspace.QueryDocs (commit 3458a46)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/meta_update.go — Replaced findTicketDirectory + findMarkdownFiles with QueryDocs


## 2025-12-13

Created intern code review verification questionnaire playbook with YAML DSL (22 questions + experiments)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/playbook/01-intern-code-review-verification-questionnaire.md — Verification playbook for onboarding/review


## 2025-12-13

Step 2.3: Migrated tasks.go ticket discovery to Workspace.QueryDocs (commit 234f42c)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/tasks.go — Replaced substring dir scan + findTicketDirectory fallback with QueryDocs


## 2025-12-13

Validation: integration scenario suite passed after Phase 2 (DOCMGR_PATH=/tmp/docmgr-scenario-local)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/README.md — How to build/pin DOCMGR_PATH
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/run-all.sh — Scenario suite run (Phase 2)


## 2025-12-13

Step 3.1: Migrated search.go suggestion mode to Workspace.QueryDocs (commit eadda8d)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/search.go — Replaced filepath.Walk/readDocumentFrontmatter in suggestion mode with QueryDocs


## 2025-12-13

Step 3.2: Migrated doc_move.go ticket discovery to Workspace.QueryDocs (commit 770e33f)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/doc_move.go — Replaced findTicketDirectory with QueryDocs


## 2025-12-13

Step 3.2: Migrated ticket_move.go ticket discovery to Workspace.QueryDocs (commit 5ce1a88)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/ticket_move.go — Replaced findTicketDirectory with QueryDocs


## 2025-12-13

Step 3.2: Migrated ticket_close.go ticket discovery to Workspace.QueryDocs (commit 35de822)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/ticket_close.go — Replaced findTicketDirectory with QueryDocs


## 2025-12-13

Step 3.3: Migrated rename_ticket.go discovery to Workspace.QueryDocs (commit 5ddd75c)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/rename_ticket.go — Replaced findTicketDirectory with QueryDocs


## 2025-12-13

Step 3.3: Migrated renumber.go discovery to Workspace.QueryDocs (commit 9fd2c8a)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/renumber.go — Replaced findTicketDirectory with QueryDocs


## 2025-12-13

Step 3.3: Migrated layout_fix.go discovery to Workspace.QueryDocs (commit c72e0db)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/layout_fix.go — Replaced legacy root scan/findTicketDirectory with QueryDocs


## 2025-12-13

Step 3: Migrated import_file.go to Workspace.QueryDocs (commit d2b357a)

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/import_file.go — Use QueryDocs to resolve ticket directory

