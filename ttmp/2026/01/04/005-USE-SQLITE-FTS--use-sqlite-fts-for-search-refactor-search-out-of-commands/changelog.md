# Changelog

## 2026-01-04

- Initial workspace created


## 2026-01-04

Create ticket + diary + analysis doc; analyze FTS integration and refactor plan for search

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/pkg/commands/search.go — Current implementation to be refactored
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/analysis/01-analysis-fts-backed-search-refactor-search-packages.md — Verbose plan for FTS-backed query + moving search engine out of commands
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/reference/01-diary.md — Diary of analysis steps


## 2026-01-04

Implement FTS-backed search (schema option C), add OrderByRank, extract reusable search engine, refactor CLI search

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/internal/searchsvc/search.go — Reusable search engine
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/internal/workspace/index_builder.go — Populate docs_fts and track FTS availability
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/internal/workspace/query_docs_fts5_test.go — FTS-tagged test
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/internal/workspace/query_docs_sql.go — SQL MATCH + bm25 ordering
- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/pkg/commands/search.go — CLI delegates to internal/searchsvc


## 2026-01-04

Update analysis doc note: parity-first superseded; no-backcompat plan lives in design doc

### Related Files

- /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/04/005-USE-SQLITE-FTS--use-sqlite-fts-for-search-refactor-search-out-of-commands/analysis/01-analysis-fts-backed-search-refactor-search-packages.md — Clarify superseded parity-first assumptions

