---
Title: "Source: Single doc view (UX snapshot)"
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ux
    - web
    - search
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Terminal-style snapshot of the desired single-document viewer UX."
LastUpdated: 2026-01-04T19:45:00-05:00
---

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Design: FTS-backed search engine (no compatibility)                         │
│ 005-USE-SQLITE-FTS • design-doc • draft                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│ METADATA                                                                    │
│ ───────────────────────────────────────────────────────────────────────     │
│ Ticket:         005-USE-SQLITE-FTS                                          │
│ Status:         draft                                                       │
│ Doc Type:       design-doc                                                  │
│ Intent:         long-term                                                   │
│ Topics:         backend, docmgr, tooling, testing                           │
│ Owners:         (none)                                                      │
│ Last Updated:   2026-01-04 15:34:47                                         │
│                                                                             │
│ RELATED FILES (3)                                                           │
│ ───────────────────────────────────────────────────────────────────────     │
│ • internal/searchsvc/search.go                                              │
│   Engine boundary the design targets                                        │
│                                                                             │
│ • internal/workspace/sqlite_schema.go                                       │
│   docs_fts schema and FTS availability                                      │
│                                                                             │
│ • pkg/commands/search.go                                                    │
│   Thin CLI wrapper                                                          │
│                                                                             │
│ ═══════════════════════════════════════════════════════════════════════════ │
│                                                                             │
│ ## Executive Summary                                                        │
│                                                                             │
│ Replace docmgr's `--query` search implementation with a **SQLite FTS5-      │
│ backed** query path (no backwards compatibility guarantees) and refactor    │
│ the CLI search command so it becomes a thin adapter over a reusable search  │
│ engine package.                                                             │
│                                                                             │
│ This design explicitly chooses:                                             │
│                                                                             │
│ - **FTS schema option C**: index `title`, `body`, `topics`, `doc_type`,    │
│   `ticket_id`.                                                              │
│ - **Ranking support**: add `OrderByRank` using `bm25(...)`.                │
│ - **Snippet**: keep the current `extractSnippet` behavior for now, but     │
│   move it into the reusable core package.                                   │
│                                                                             │
│ ## Problem Statement                                                        │
│                                                                             │
│ Today `docmgr doc search --query` is implemented as a Go substring scan     │
│ over bodies after `Workspace.QueryDocs(...)` returns candidates. This has   │
│ three major downsides:                                                      │
│                                                                             │
│ - slow on large workspaces (body scanning is O(total bytes scanned))       │
│ - no ranking (results are typically ordered by path)                        │
│ - search logic is embedded in `pkg/commands/search.go`, which makes it     │
│   difficult to reuse for HTTP APIs and increases risk of semantic drift    │
│                                                                             │
│ We do **not** need behavior exact-match compatibility with the current     │
│ substring semantics. This ticket is allowed to change behavior.             │
│                                                                             │
│ ## Proposed Solution                                                        │
│                                                                             │
│ ### 1) Add FTS to the in-memory workspace index                             │
│                                                                             │
│ Extend the workspace SQLite schema to create a virtual FTS5 table:          │
│                                                                             │
│ - Table name: `docs_fts`                                                    │
│ - Model: **contentless** FTS table, populated during ingest                │
│ - Tokenizer: `unicode61`                                                    │
│                                                                             │
│ Schema (Option C):                                                          │
│                                                                             │
│ ```sql                                                                      │
│ CREATE VIRTUAL TABLE IF NOT EXISTS docs_fts USING fts5(                     │
│   title,                                                                    │
│   body,                                                                     │
│   topics,                                                                   │
│   doc_type,                                                                 │
│   ticket_id,                                                                │
│   tokenize='unicode61'                                                      │
│ );                                                                          │
│ ```                                                                         │
│                                                                             │
│ [scroll down for more...]                                                   │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│ [Tab] Switch section  [↑↓] Scroll  [Esc] Close  [Ctrl+C] Copy all          │
└─────────────────────────────────────────────────────────────────────────────┘
```
