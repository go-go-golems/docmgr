# Changelog

## 2026-07-06

- Initial workspace created


## 2026-07-06

Step 1: Created DOCMGR-201 ticket, diary, fix plan, and task list for PR 43 blocker fixes.

### Related Files

- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/06/DOCMGR-201-fix-pr43-review-blockers--fix-pr-43-review-blockers-canonical-ticket-refs-doctor-refs-stable-task-api-codeql/analysis/01-pr-43-blockers-fix-plan.md — Blocker fix plan
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/06/DOCMGR-201-fix-pr43-review-blockers--fix-pr-43-review-blockers-canonical-ticket-refs-doctor-refs-stable-task-api-codeql/reference/01-implementation-diary.md — Chronological implementation diary


## 2026-07-06

Step 2: Implemented PR 43 blocker fixes: doc add persists canonical ticket IDs, doctor resolves forgiving ticket refs, HTTP/UI task checks use stable refs with legacy id fallback, and resolver comparison keys avoid filepath normalization at CodeQL alert site.

### Related Files

- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/internal/httpapi/tickets.go — refs API for task check with ids fallback
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/internal/paths/resolver.go — Comparison-only normalization for CodeQL alert site
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/internal/tasksmd/tasksmd.go — Stable-ref task toggle helper
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/pkg/commands/add.go — Canonical ticket ID persistence for doc add
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/pkg/commands/doctor.go — Forgiving doctor --ticket resolution
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ui/src/features/ticket/tabs/TicketTasksTab.tsx — UI task tab uses stable refs
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ui/src/services/docmgrApi.ts — UI API mutation sends refs


## 2026-07-06

Step 4: Fixed local embed/public generation for release hooks: Dagger UI build excludes node_modules/dist, uses frozen pnpm install, and make goreleaser now depends on ui-build.

### Related Files

- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/Makefile — goreleaser depends on ui-build
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/internal/web/generate_build.go — Non-interactive Dagger UI build fix

