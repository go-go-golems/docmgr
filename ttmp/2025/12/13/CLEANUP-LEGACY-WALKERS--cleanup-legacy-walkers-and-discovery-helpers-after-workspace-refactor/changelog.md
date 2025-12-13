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

