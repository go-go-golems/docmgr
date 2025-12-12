---
Title: 'Design Spec: workspace.Workspace SQLite-backed repository lookup API'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T16:54:14-05:00
---

# Design Spec: `workspace.Workspace` SQLite-backed repository lookup API

## 1) Summary

This document explains a refactor that centralizes how docmgr finds tickets and documents, and how it supports searching and reverse lookup. It is written for a new engineer joining the team: it provides background, a mental model, concrete examples, and a precise API direction.

We will centralize ticket/document discovery and lookup behind a first-class object in `internal/workspace`:

- `workspace.Workspace` is the API entry point for repo/ticket/doc lookup.
- Each CLI invocation constructs a `Workspace`, **eagerly builds an in-memory SQLite index**, runs queries, and exits.
- Query entry point: `QueryDocs(ctx, DocQuery) (DocQueryResult, error)` (structured request/response).
- Reverse lookup is implemented as **constraints/filters** compiled into SQL joins (not a separate “reverse scope” mode).

This spec consolidates decisions from debate rounds and the interactive design log.

## 2) Goals / non-goals

### Goals
- **Single source of truth** for:
  - ticket discovery and ticket-id → directory resolution
  - document enumeration + filtering semantics
  - reverse lookup (`RelatedFiles`) using consistent normalization rules
- **Deterministic behavior**: stable ordering, explicit inclusion policies, explicit diagnostics.
- **Explainability**: skip/parse/normalization issues are surfaced as diagnostics rather than silent skipping.
- **SQLite-backed implementation**: reduce hand-written correlation and nested scans; enable richer, composable queries.

### Non-goals (for initial implementation)
- Long-running index refresh / daemon mode
- Persistent on-disk DB lifecycle
- Performance tuning beyond “reasonable for CLI invocation”
- Vocabulary validation inside lookup (kept separate per Q8)

## 3) Background: why this refactor exists (problem statement)

Today, the codebase contains multiple implementations of “find tickets” and “find docs”, spread across commands. They differ in:
- what directories they skip,
- whether they include `index.md`,
- whether invalid frontmatter is silently skipped or surfaced,
- how reverse lookup against `RelatedFiles` is matched (normalized vs raw string ops),
- and how ticket filtering behaves (exact match vs substring).

For a new engineer, the key takeaway is: **the concept of “the set of docs in the workspace” is not currently canonical**. That makes features harder to build and harder to debug.

This refactor introduces a single centralized API (`workspace.Workspace`) with a single ingestion step (SQLite index build) and a single query API (`QueryDocs`). Commands then become thin: they translate flags into `DocQuery`, run `QueryDocs`, and format output.

## 3) References (debates + log)

- Debate Q6 (ticket identity): `reference/06-debate-round-6-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md`
- Debate Q7 (scope): `reference/07-debate-round-7-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md`
- Debate Q8 (vocab/config boundaries): `reference/08-debate-round-8-q8-how-do-we-keep-vocabulary-config-concerns-from-leaking-everywhere.md`
- Debate Q11 (QueryDocs): `reference/09-debate-round-9-q11-design-querydocs-ctx-scope-filters.md`
- Debate Q1 (repository/workspace object): `reference/10-debate-round-10-q1-what-is-the-repository-object-and-what-does-it-own.md`
- Debate Q2 (broken states): `reference/11-debate-round-11-q2-how-should-we-represent-broken-and-partial-states.md`
- Debate Q3 (semantics/skip rules): `reference/12-debate-round-12-q3-what-are-the-semantics-of-filters-and-enumeration.md`
- Debate Round 13 (SQLite backend): `reference/14-debate-round-13-sqlite-index-backend-influences-lookup-and-reverse-lookup.md`
- Interactive design log (decisions): `reference/13-design-log-repository-api.md`

## 4) Architecture overview

### 4.1 High-level flow (per CLI invocation)

1. Construct `workspace.Workspace`:
   - supports both best-effort discovery (CLI) and injected context (tests)
   - requires `Root`, `ConfigDir`, `RepoRoot` invariants for the core object
2. Eagerly ingest workspace into an in-memory SQLite DB:
   - enumerate docs according to canonical skip policy
   - parse frontmatter; store parse failures (`parse_ok=0`, `parse_err`)
   - normalize `RelatedFiles` paths into canonical keys
3. Execute `QueryDocs` / ticket queries as SQL.
4. Return results + diagnostics (if requested/needed).

### 4.1.1 The mental model (for interns)

Think of docmgr as a “tiny documentation database” that is rebuilt on demand:

- The **filesystem** is the source of truth.
- On each CLI run, we scan a subset of that filesystem (according to skip rules) and load it into an **in-memory SQLite database**.
- Once loaded, nearly all features are expressed as **queries** instead of “walk + nested loops”.

This reduces duplication, ensures consistency, and makes reverse lookup (“show me docs referencing this file”) a simple join.

### 4.2 Layering / ownership

- `internal/workspace`:
  - new `Workspace` object
  - index construction and query compilation
  - canonical enumeration/skip policy for ingestion
- `internal/documents`:
  - YAML/frontmatter parsing (`ReadDocumentWithFrontmatter`)
- `internal/paths`:
  - path normalization (`paths.Resolver`) used to compute stable join keys
- `pkg/diagnostics/core` + `pkg/diagnostics/docmgrctx`:
  - taxonomy objects used in diagnostics/repair messaging

## 5) Workspace layout basics (what we are indexing)

At a high level, docmgr workspaces are organized under a docs root (often `ttmp/`) with ticket directories, for example:

- `ttmp/YYYY/MM/DD/TICKET--slug/index.md` (ticket index)
- `ttmp/YYYY/MM/DD/TICKET--slug/design/...` (ticket docs)
- `ttmp/YYYY/MM/DD/TICKET--slug/archive/...` (archived artifacts)
- `ttmp/YYYY/MM/DD/TICKET--slug/.meta/...` (implementation metadata; we skip this entirely)

Each markdown doc generally begins with YAML frontmatter. Two relevant frontmatter fields are:
- `Ticket`: ticket identifier (used for ticket scoping)
- `RelatedFiles`: list of code paths related to the doc

## 5) Public API (initial shape)

### 5.1 Construction

```go
type WorkspaceContext struct {
    Root      string
    ConfigDir string
    RepoRoot  string
    Config    *WorkspaceConfig // best-effort loaded config (may be nil initially)
}

func NewWorkspaceFromContext(ctx WorkspaceContext) (*Workspace, error)

type DiscoverOptions struct {
    RootOverride string
}
func DiscoverWorkspace(ctx context.Context, opts DiscoverOptions) (*Workspace, error)
```

Intern tip: even though the API supports injected construction for tests, **CLI commands should almost always call** `DiscoverWorkspace` and not re-run discovery logic themselves.

### 5.2 QueryDocs

```go
type DocQuery struct {
    Scope   Scope
    Filters DocFilters
    Options DocQueryOptions
}

type ScopeKind int
const (
    ScopeRepo ScopeKind = iota
    ScopeTicket
    ScopeDoc
)

type Scope struct {
    Kind     ScopeKind
    TicketID string // if ScopeTicket
    DocPath  string // if ScopeDoc
}

type DocFilters struct {
    Ticket string  // exact match (Decision 13)
    DocType string
    Status string

    // Reverse lookup constraints (Option B)
    RelatedFile string
    RelatedDir  string

    // Topics semantics deferred; planned as SQL-extensible.
    TopicsAny []string
}

type DocQueryOptions struct {
    IncludeBody   bool
    IncludeErrors bool
    OrderBy       OrderBy
    Reverse       bool

    // Default visibility policy (derived from ingest tags)
    IncludeArchivedPath bool
    IncludeControlDocs  bool
    IncludeScriptsPath  bool

    // Optional: emit diagnostics explaining skips/fallbacks
    IncludeDiagnostics bool
}

type DocHandle struct {
    Path    string
    Doc     *models.Document
    Body    string
    ReadErr error
}

type DocQueryResult struct {
    Docs        []DocHandle
    Diagnostics []core.Taxonomy // Decision: reuse taxonomy objects
}
```

Notes:
- Reverse lookup is **filters/constraints** compiled to SQL joins, not a separate scope kind.
- Contradictory queries are **hard errors** (Decision 16 / D2).

### 5.3 Example call sites (how commands will use it)

#### Example: “List docs for a ticket”

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: "MEN-3475"},
    Filters: workspace.DocFilters{
        Status: "active",
    },
    Options: workspace.DocQueryOptions{
        IncludeDiagnostics: true, // show what got skipped
    },
})
```

#### Example: “Reverse lookup: docs referencing a file”

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Filters: workspace.DocFilters{
        RelatedFile: "pkg/commands/search.go",
    },
    Options: workspace.DocQueryOptions{
        IncludeDiagnostics: true,
    },
})
```

Under the hood, this becomes a join against the `related_files` table (see §10 examples).

#### Example: “Search defaults (hide noisy categories)”

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Filters: workspace.DocFilters{
        DocType: "design",
    },
    Options: workspace.DocQueryOptions{
        IncludeArchivedPath: false,
        IncludeControlDocs:  false,
        IncludeScriptsPath:  false,
    },
})
```

## 6) Canonical skip rules (ingest-time)

These rules determine what enters the SQLite index.

### 6.1 Directories

- **Skip `.meta/` entirely** (Decision 6).
- **Skip all underscore dirs (`_*/`)** (Decision 7).
- **Include `archive/` but tag** `is_archived_path=1` (Decision 5).
- **Include `scripts/` but tag** `is_scripts_path=1` and default-hide (Decision 8).
- **Include `sources/` but tag** `is_sources_path=1` (Decision 9); do not hide by default.

### 6.2 Control docs at ticket root

Include `README.md`, `tasks.md`, `changelog.md`, but tag `is_control_doc=1` and default-hide (Decision 10).

### 6.3 `index.md` inclusion

Include `index.md` by default in doc queries (Decision 12). (Implementation may tag `is_index=1`.)

## 7) Why SQLite here (and what it buys us)

### 7.1 The “before” state (manual correlation)

Historically, reverse lookup and filtering required code like:
- walk every markdown file,
- parse frontmatter,
- scan `RelatedFiles` arrays,
- normalize paths (sometimes),
- perform matching.

This logic is easy to duplicate and easy to get subtly inconsistent.

### 7.2 The “after” state (joins instead of loops)

With an in-memory SQLite index:
- each doc becomes a row in `docs`
- each related file entry becomes a row in `related_files`
- reverse lookup becomes a join + predicate (no nested loops)

This is a major simplification for new contributors: to add a feature, you usually add:
- one column/table in ingestion, and
- one compilation rule in query generation.

## 7) Parse errors + diagnostics contract

### 7.1 Invalid frontmatter docs

- Ingest indexes invalid-frontmatter docs with:
  - `parse_ok=0`, `parse_err` populated (Decision 11).
- Default queries exclude these docs, but **emit diagnostics** explaining skipped docs (Decision 16 / D1=B).

### 7.2 Contradictory/invalid queries

- **Hard error** (Decision 16 / D2=A).

### 7.3 Reverse lookup normalization failure

- Emit diagnostics + apply documented fallback strategy (Decision 16 / D3=B).
  - Example fallback: try repo-relative key; if unavailable, try absolute key; if unavailable, fall back to cleaned raw.

### 7.4 Diagnostics representation

Diagnostics are represented as existing `core.Taxonomy` objects (`pkg/diagnostics/core/types.go`), not a new bespoke type.

**Open design note**: make “message” + “suggestions/hints” uniformly available across domains (either extend `core.Taxonomy` or standardize context payload interface).

## 8) Intern notes: diagnostics taxonomy in practice

Docmgr already uses taxonomy objects for frontmatter parsing and schema issues. You’ll see these in code as:
- an error being returned,
- and later extracted via `core.AsTaxonomy(err)`.

We reuse the taxonomy model inside `DocQueryResult.Diagnostics` so tooling/UIs can render consistent messages.

Open enhancement: make “message” and “suggestions/hints” uniformly accessible without domain-specific type checks, even though some contexts already carry these fields (e.g. `FrontmatterParseContext.Fixes`).

## 9) SQLite index model (minimum viable)

### 8.1 Required tables

- `docs`
- `doc_topics`
- `related_files`

Optionally (later):
- `tickets` (for ticket-level edge cases and validation queries)

### 8.2 Recommended columns

`docs`:
- `path`, `ticket_id`, `doc_type`, `status`, `intent`, `title`, `last_updated`
- `parse_ok`, `parse_err`
- tags: `is_index`, `is_archived_path`, `is_scripts_path`, `is_sources_path`, `is_control_doc`
- `body` (optional; can be skipped initially)

`related_files`:
- `doc_id`, `note`
- normalized keys (at least one canonical key; store additional fallback keys):
  - `norm_repo_rel`, `norm_abs`, `norm_clean`, `anchor`

`doc_topics`:
- `doc_id`, `topic_lower` (or store original + lower)

## 10) Concrete query examples (DocQuery → SQL mental model)

This section gives new engineers an intuition for how the API maps to the index. (Exact SQL may differ; this is conceptual.)

### 10.1 Ticket scope + status filter

DocQuery:
- scope: ticket
- filter: status=active

SQL shape:
- `WHERE docs.ticket_id = ? AND docs.status = ?`

### 10.2 Reverse lookup by file

DocQuery:
- scope: repo
- filter: RelatedFile="pkg/commands/search.go"

SQL shape:
- join `related_files` and match a normalized key:
  - `JOIN related_files rf ON rf.doc_id = docs.doc_id`
  - `WHERE rf.norm_repo_rel = ?`

### 10.3 Include/exclude tagged path categories

If `IncludeArchivedPath=false`, SQL adds:
- `AND docs.is_archived_path = 0`

If `IncludeControlDocs=false`, SQL adds:
- `AND docs.is_control_doc = 0`

If `IncludeScriptsPath=false`, SQL adds:
- `AND docs.is_scripts_path = 0`

### 10.4 Invalid frontmatter diagnostics

If `IncludeErrors=false`, invalid docs (`parse_ok=0`) are excluded from `Docs` but an entry is added to `Diagnostics` that points at the `docs.path` and carries a taxonomy describing the parse failure.

## 11) Migration plan (API-level)

1. Introduce `workspace.Workspace` and implement ingestion + a minimal `QueryDocs`.
2. Port command-by-command:
   - `search` → `QueryDocs` + (optional) content search layer
   - `list docs` → `QueryDocs` with default policies
   - `doctor` → `QueryDocs(IncludeErrors=true, IncludeDiagnostics=true)` + ticket-level checks
   - `relate` → use the same normalization keys as the index
3. Remove/retire duplicated walkers and inconsistent skip semantics.

## 12) Open questions (explicitly deferred)

- Topic match semantics (any vs all vs configurable) — deferred (Decision 13 section).
- Directory reverse lookup semantics for `RelatedDir` (match related_files vs doc path vs either) — deferred.
- DiscoverWorkspace failure/fallback behavior for missing config/repo-root — parked.
- Whether to expose a richer boolean `Where Expr` in `DocQuery` — postponed until needed.


