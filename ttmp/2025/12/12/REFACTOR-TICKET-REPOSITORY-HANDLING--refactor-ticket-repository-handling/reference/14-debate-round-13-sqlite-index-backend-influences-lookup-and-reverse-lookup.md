---
Title: 'Debate Round 13 — SQLite-backed index: how backend influences lookup + reverse lookup'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/paths/resolver.go
      Note: Path normalization keys that the SQLite index should store
    - Path: pkg/commands/search.go
      Note: Current reverse lookup implementation (file/dir filters + path matching)
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/07-debate-round-7-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md
      Note: Scope vs reverse lookup design
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/09-debate-round-9-q11-design-querydocs-ctx-scope-filters.md
      Note: QueryDocs contract + options
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/13-design-log-repository-api.md
      Note: Interactive API decisions
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T15:56:48-05:00
---


# Debate Round 13 — SQLite-backed index: how backend influences lookup + reverse lookup

## Goal

Explore how knowing we’ll implement lookup/search via an **in-memory SQLite index** influences the design of the new `workspace.Workspace` lookup API—especially **reverse lookup**—since we won’t need to hand-correlate data structures.

**Prompt**: “If the data is indexed into SQLite, what should the public API look like so it stays clean and stable, while enabling powerful queries (including reverse lookup)?”

**Non-goals** (sticking to earlier debate framing):
- not performance benchmarking
- not security
- not backwards compatibility

## Context / ties to earlier rounds

- **Q11 / Round 9 (`QueryDocs`)**: we picked a single entry point, debated request/response shape, deterministic ordering, parse/error handling.
- **Q7 / Round 7 (scope)**: reverse lookup is currently special-case logic; question is whether it’s scope vs filter vs helper.
- **Q2 / Round 11 (broken states)**: how broken docs/tickets are represented in results.
- **Q3 / Round 12 (semantics + enumeration)**: consistent skip rules and match semantics.
- **Design log (`reference/13-design-log-repository-api.md`)**: we’re converging on `workspace.Workspace` + structured request/response.

## Pre-Debate Research (current behavior to replace)

### Evidence A — Reverse lookup exists today but is hand-rolled

Current reverse lookup behavior lives in `pkg/commands/search.go` and is implemented by:
- walking docs
- parsing frontmatter
- normalizing query path and related-file paths (sometimes)
- matching via a mix of normalization helpers and string ops

This is exactly what a SQLite join eliminates: instead of “scan N docs and check each RF list”, we can:
- store `doc_path` rows
- store `related_file_path` rows
- query joins with a normalized key

### Evidence B — Normalization must still exist, but can be centralized

Even with SQLite, you still need consistent normalization:
- “how do we canonicalize a user query path?”
- “how do we canonicalize `RelatedFiles.Path` values from frontmatter?”

SQLite doesn’t remove normalization; it makes matching **a query** instead of manual loops.

## Opening Statements

### Mara (Staff Engineer) — “Keep the API semantic; SQLite is an internal accelerator”

**Position**: Public API should remain **semantic** and not “leak SQL-ness”.

- `QueryDocs(ctx, DocQuery) (DocQueryResult, error)` stays.
- `DocQuery.Scope` can include reverse lookup variants (ReverseFile/ReverseDir).
- Implementation may use SQLite under the hood, but callers shouldn’t know.

**Rationale**:
- Callers shouldn’t have to care whether results come from walking the FS or a DB.
- Keeping a stable semantic API allows swapping SQLite in/out (or combining with fallback).

**API consequence**:
- Structured request/response becomes even more valuable: we can add new query features without adding more methods.

### Jon (Senior Engineer) — “Expose a small query language (but not raw SQL)”

**Position**: If we’re using SQLite, we can safely expose a **more expressive query surface** than today’s flags, but still not raw SQL.

Example: allow compound logic beyond “AND all filters”:
- topic matches ANY of {a,b}
- (doc_type IN …) AND (status != archived)
- reverse lookup constraints combined with metadata filters

**Rationale**:
- SQLite makes complex filters cheap to express and maintain internally.
- A richer query AST avoids proliferation of ad-hoc options.

**API consequence**:
- Add an expression form in `DocQuery`, e.g. `Where Expr` alongside `Filters`.

### `pkg/commands/*` (bloc) — “We want a single call that maps from flags”

**Position**: Keep the top-level API easy:
- “flags in → one `DocQuery` → results out”

Commands shouldn’t need to build complex boolean ASTs; they should pass simple fields and let `workspace.Workspace` compile it.

**Rationale**:
- CLI is still the main consumer.
- The implementation can become smarter without increasing command complexity.

### `paths.Resolver` — “SQLite does not remove path anchoring; define canonical keys”

**Position**: DB-backed reverse lookup only works if we define **canonical keys** for stored paths and queries.

Proposed approach:
- store multiple forms per path: `raw`, `clean`, `canonical`, `abs`, `repo_relative`
- pick one canonical comparable key (e.g. `repo_relative`) when available
- fall back to others when repo root is unknown

**Rationale**:
- Normalization is the linchpin; without it, SQLite just stores inconsistent strings faster.

## Debate: How SQLite changes the design

### 1) Reverse lookup becomes a join instead of a loop

**Today**:
- walk docs
- read doc
- for each related file entry, match against query

**With SQLite**:
- `docs` table: one row per doc
- `related_files` table: one row per (doc_id, normalized_path, note)
- Query:
  - reverse file: `SELECT docs.* FROM docs JOIN related_files ... WHERE related_files.norm = ?`
  - reverse dir: `... WHERE related_files.norm LIKE 'dir/%'` (or better: use a normalized directory key)

**Design impact**:
Reverse lookup fits naturally as a **Scope** variant because it changes “which docs are in the set”.

### 2) Filters become WHERE clauses; semantics must be specified (still)

SQLite makes it easy to implement consistent semantics, but we must decide:
- ticket filter: exact vs substring vs configurable (Q3)
- topics: any-match case-insensitive (current)
- include/exclude `index.md`
- include invalid frontmatter docs or not (Q2/Q11)

**Design impact**:
`DocQueryOptions` needs explicit policy fields so semantics are deterministic and shared.

### 3) Broken states require row modeling

In SQLite, “broken” is not an exception; it’s a row state:
- `docs.parse_ok BOOLEAN`
- `docs.parse_err TEXT` (or a classified taxonomy code)
- `tickets.index_missing BOOLEAN`
- `tickets.index_parse_ok BOOLEAN`
- etc.

**Design impact**:
This strongly suggests “handles + diagnostics” rather than skipping.

### 4) Incremental indexing (optional) changes lifecycle, not API

SQLite introduces the question: when do we build/refresh the index?
- build once per command invocation (simple)
- lazily build on first query (still simple)
- maintain incremental updates (more complex)

**Design impact**:
This should remain internal; the API should not require callers to orchestrate indexing.
At most, expose an explicit `w.RefreshIndex(...)` for advanced workflows.

### 5) Determinism becomes easy

SQLite queries can specify ordering deterministically:
- `ORDER BY path ASC` default
- `ORDER BY last_updated DESC` optionally

**Design impact**:
We can guarantee deterministic ordering in `DocQueryOptions`.

## Concrete proposal (SQLite-aware but API-stable)

### Keep the same semantic API

```go
type DocQuery struct {
    Scope   Scope
    Filters DocFilters
    Options DocQueryOptions
}

type Scope struct {
    Kind ScopeKind
    TicketID string
    DocPath  string
    ReverseFile string
    ReverseDir  string
}
```

### Add an optional “advanced where” expression (only if needed)

Only if we want richer logic than “AND of filters”, add:

```go
type DocQuery struct {
    Scope   Scope
    Filters DocFilters
    Where   *Expr // optional
    Options DocQueryOptions
}
```

This keeps CLI simple (commands mostly use `Filters`) while enabling power users later.

## Moderator Summary

### What SQLite changes (agreement)
- Reverse lookup becomes natural and consistent via joins.
- Deterministic ordering becomes trivial to guarantee.
- Broken states are easier to represent as row states (not just errors).

### What SQLite should NOT change (agreement)
- The API should remain semantic (`workspace.Workspace.QueryDocs(...)`), not DB-shaped.
- Normalization remains essential; SQLite does not remove the need for `paths.Resolver`-style canonicalization.

### Main tension (disagreement)
- Do we expand the public API to support richer boolean query logic (Jon), or keep it strictly in `Filters` + `Scope` (Mara/commands)?

### Open questions
- Should reverse lookup live in `Scope` (recommended) or in `Filters`?
- Do we want an expression language (`Where Expr`) now or later?
- How do we model normalization keys for stable joins (repo-relative preferred, but fallback rules)?
- How do we represent parse errors in results: `DocHandle.ReadErr` vs `Diagnostics` vs both?

## Next Steps

1. Decide whether reverse lookup is a `Scope` variant (likely yes).
2. Decide whether we need a `Where` expression now or postpone.
3. Define normalized path keys stored in the index (what canonical form?).
4. Define indexing lifecycle: build once per invocation vs lazy build.


