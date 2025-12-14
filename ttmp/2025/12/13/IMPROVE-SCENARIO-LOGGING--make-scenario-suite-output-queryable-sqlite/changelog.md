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


## 2025-12-13

Step 2: Add scenariolog run start/end lifecycle (commit 1ecfe225f95076b8b8df77fcd7821b62ca65566f).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/cmd/scenariolog/main.go — Cobra run start/end commands
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/run.go — Run start/end DB helpers
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 2


## 2025-12-13

Step 3: Implement scenariolog exec step capture (steps row + artifacts + best-effort FTS indexing) (commit 9ac50c1f4f7314d6afbc24814f0e2144b4c056c8).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/cmd/scenariolog/main.go — Cobra exec command + exit code propagation
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/artifacts.go — Artifacts insertion
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/exec_step.go — ExecStep wrapper (capture + step rows)
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/fts.go — Best-effort FTS line indexing
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 3


## 2025-12-13

Step 4: Add FTS-backed search (library + CLI) (commit 791ffd30c5083a3e7ca3d8e0595e73de241fffd6).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/cmd/scenariolog/main.go — Cobra search command
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/search.go — FTS query API + degraded-mode error
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/search_fts5_test.go — FTS5-tagged integration test
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 4


## 2025-12-13

Step 5: Start emitting KV tags for provenance and step metadata (commit 6f32b75a1c18854aeade72b299b8d5ff1c834596).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/exec_step.go — Step-level KV tags
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/kv.go — KV upsert helper
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/run.go — Run-level KV tags
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 5


## 2025-12-13

Step 6: Integrate scenariolog into scenario harness (run-all.sh) (commit b2a11b6495d38c0bdd53662af363135760d59fbc).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/README.md — Document SCENARIOLOG_PATH + auto-build behavior
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/run-all.sh — Wrap steps with scenariolog run/exec/end + EXIT trap
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 6


## 2025-12-13

Step 7-8: Hardening exec cancellation (kill process group) + handle SIGINT as cancellation (commits 194b2d9428f508ff7365baf16c2d94d7b1b032f4, 99b1439266e7900c9dabd6ca2bf69dd27f97898e).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/cmd/scenariolog/main.go — SIGINT cancels exec
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/exec_step.go — Cancel => terminate process group
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/internal/scenariolog/procgroup_unix.go — Unix process group management
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary steps 7-8


## 2025-12-13

Step 9: Switch scenariolog query/report commands to Glazed structured output (commit ba9fef989995b2a7340664e9ba18c9fa64906f0d).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/cmd/scenariolog/glazed_cmds.go — Glazed commands (search/summary/failures/timings)
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/go.mod — Added glazed dependency
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 9


## 2025-12-13

Added reference doc: how to use scenariolog-local (with Glazed output) and captured Glaze doc-writing guidance.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/03-how-to-use-scenariolog-local.md — Usage guide
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/sources/local/glaze-help-how-to-write-good-documentation-pages-2025-12-14.txt — Glaze doc style guide
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/sources/local/glaze-help-writing-help-entries-2025-12-14.txt — Glaze help-entry authoring guide


## 2025-12-13

Wired Glazed help system into scenariolog (embedded docs + `scenariolog help how-to-use-scenariolog-local`) (commit 592f511b625b066ccc00dae84d51ff915136e732).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/cmd/scenariolog/main.go — Help system wiring
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/pkg/doc/doc.go — Embed and load help sections
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/scenariolog/pkg/doc/docs/how-to-use-scenariolog-local.md — Help section content
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/03-how-to-use-scenariolog-local.md — Pointer to embedded help
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/sources/local/glaze-help-help-system-2025-12-14.txt — Glazed help-system guide


## 2025-12-13

Step 10: Glazed runtime commands + --kv (KeyValue) and updated docs/examples; captured glaze commands-reference.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/02-diary.md — Diary step 10
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/reference/03-how-to-use-scenariolog-local.md — Document --kv usage
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/sources/local/glaze-help-commands-reference-2025-12-14.txt — Glaze commands reference
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/tasks.md — Add query recipes

