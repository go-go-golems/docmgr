# Changelog

## 2025-12-13

- Initial workspace created


## 2025-12-13

Created design doc for sqlite-based scenario logging

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/01-scenario-suite-structured-logging-sqlite.md — Full spec with schema


## 2025-12-13

Added reusable Go/sqlite scenario logger design-doc + brainstorm reference (idea bank for future phases).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/02-generic-sqlite-scenario-logger-go-tool.md — Reusable Go tool design-doc (CLI + library)
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/01-brainstorm-scenario-logging-ideas-wild-useful.md — Brainstorm/idea bank for schema/capture/reporting/diffing


## 2025-12-13

Added detailed implementation-plan design-doc for phased execution.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md — Detailed plan + checklists


## 2025-12-13

Step 1: Scaffold scenariolog module + migrations; add graceful FTS5 fallback (commit 41d66c1dd66d8f8839b81d3612afd5b0e63745cb; build with -tags sqlite_fts5 for full search).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/cmd/scenariolog/main.go — CLI init command
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/go.mod — Self-contained module
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/migrate.go — Migrations + FTS5 best-effort
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 1

