# Tasks

## TODO

- [ ] Add tasks here

- [x] Update smoke scenario docs/scripts to build with `-tags sqlite_fts5`
- [x] Exercise `--order-by rank` in the scenario suite

- [x] Add docs_fts (FTS5) table to workspace schema (option C: title/body/topics/doc_type/ticket_id), best-effort create
- [x] Populate docs_fts during workspace ingest (rowid=doc_id); decide topics serialization
- [x] Extend workspace query model: DocFilters.TextQuery + OrderByRank (bm25)
- [x] Update query SQL compiler to JOIN docs_fts + MATCH + ORDER BY bm25 for rank
- [x] Create internal/searchsvc reusable engine types (SearchQuery/SearchResult)
- [x] Move extractSnippet into internal/searchsvc and use it from both CLI/engine
- [x] Refactor pkg/commands/search.go to be a thin adapter over internal/searchsvc
- [x] Add sqlite_fts5-tagged tests for MATCH + OrderByRank
- [ ] Update search documentation (implementation guide) to describe FTS semantics and ranking
