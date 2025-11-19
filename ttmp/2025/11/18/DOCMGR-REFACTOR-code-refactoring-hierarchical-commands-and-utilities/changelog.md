# Changelog

## 2025-11-18

- Initial workspace created


## 2025-11-18

Added cmd/docmgr/cmds hierarchy for Cobra commands and moved config/workspace/template helpers under internal/.

### Related Files

- cmd/docmgr/cmds/root.go — Cobra hierarchy
- cmd/docmgr/main.go — Root now delegates to cmds
- internal/templates/templates.go — template helpers moved
- internal/workspace/config.go — config resolution moved


## 2025-11-18

Added internal/documents utilities (frontmatter + walk) and migrated key commands to them; updated Cobra hierarchy verification and CLI tests.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/documents/frontmatter.go — New helper
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/documents/walk.go — New walker
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/import_file.go — Uses helpers
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/meta_update.go — Refactored to use helpers
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/relate.go — Uses helpers


## 2025-11-18

Wrapped high-traffic CLI listings with contextual errors, fixed config warnings, and exercised the new frontmatter helpers end-to-end by creating + updating docs via the CLI.

### Highlights

- Fixed `ResolveRoot` warnings so malformed `.ttmp.yaml` files surface real parse errors.
- Added contextual errors for ticket/doc/workspace listings, meta updates, and layout fix dry-runs.
- Imported two new documents (`design-doc/01-error-context-rollout.md`, `playbooks/01-cli-regression-checklist.md`) via `doc add`, then updated metadata with `meta update` to confirm the helper pipeline works.
- Verified doc relationships by running `doc relate`, `tasks check`, and `changelog update` from the CLI.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/workspace/config.go — config parsing warnings
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/list_tickets.go — ticket list error context
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/list_docs.go — doc list error context
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/list.go — workspace list error context
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/status.go — summary row error context
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/meta_update.go — glaze row reporting fixes
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/import_file.go — safer external source metadata writes
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/layout_fix.go — dry-run/move logging with context

## 2025-11-18

Added contextual errors to all tasks CLI commands and re-ran go test.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/tasks.go — tasks command error handling


## 2025-11-18

Extended Round 4 error context work to doc search and doctor commands; verified via go test + live CLI runs.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/doctor.go — doctor findings addRow context
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/search.go — search result/suggestion rows wrapped


## 2025-11-18

Captured a Round 4 implementation diary describing the search/doctor error-context rollout and task closure.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/ttmp/2025/11/18/DOCMGR-REFACTOR-code-refactoring-hierarchical-commands-and-utilities/various/2025-11-18-implementation-diary.md — New diary

