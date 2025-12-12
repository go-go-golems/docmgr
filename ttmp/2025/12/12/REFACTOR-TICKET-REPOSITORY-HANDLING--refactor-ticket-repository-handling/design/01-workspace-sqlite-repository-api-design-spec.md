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
LastUpdated: 2025-12-12T16:58:40-05:00
---

# Design Spec: `workspace.Workspace` SQLite-backed repository lookup API

## 0) Previous Session Summary

This document continues work from a previous design session. Here's what was accomplished:

**Analysis Phase**: A comprehensive codebase analysis (`analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`) identified that ticket and document discovery logic is duplicated across multiple commands (`list_tickets.go`, `list_docs.go`, `search.go`, `doctor.go`, `import_file.go`, etc.), with inconsistent behaviors:
- Different skip rules (some skip `_*/` dirs, some don't)
- Different ticket filtering semantics (exact match vs substring)
- Different error handling (silent skip vs diagnostics)
- Different reverse lookup implementations (normalized paths vs raw string matching)

**Debate Phase**: Thirteen structured debate rounds explored key design questions:
- **Q6**: What is a ticket? (ID vs directory vs `index.md` frontmatter)
- **Q7**: How should we model scope in lookups? (repo-wide vs ticket-scoped vs doc-scoped)
- **Q8**: How do we keep vocabulary/config concerns from leaking everywhere?
- **Q9-Q11**: Detailed `QueryDocs` API design (request/response shape, filters, options)
- **Q12-Q13**: Skip rules, broken state representation, filter semantics
- **Round 13**: How SQLite backend influences lookup/reverse lookup design

**Design Log**: An interactive design session (`reference/13-design-log-repository-api.md`) captured 16 concrete decisions, including:
- Package placement: extend `internal/workspace` (not a new `internal/repository`)
- Construction: support both discovery (CLI-friendly) and injection (test-friendly)
- Skip rules: canonical policies for `.meta/`, `archive/`, `scripts/`, `sources/`, control docs
- Error handling: index invalid frontmatter docs but exclude from default queries; emit diagnostics
- Reverse lookup: model as filters/constraints compiled to SQL joins (not a separate scope mode)

**Current State**: This spec consolidates all decisions into a single authoritative design document. The next phase is implementation.

## 1) Summary

This document explains a refactor that centralizes how docmgr finds tickets and documents, and how it supports searching and reverse lookup. It is written for a new engineer joining the team: it provides background, a mental model, concrete examples, and a precise API direction.

We will centralize ticket/document discovery and lookup behind a first-class object in `internal/workspace`:

- `workspace.Workspace` is the API entry point for repo/ticket/doc lookup.
- Each CLI invocation constructs a `Workspace`, **eagerly builds an in-memory SQLite index**, runs queries, and exits.
- Query entry point: `QueryDocs(ctx, DocQuery) (DocQueryResult, error)` (structured request/response).
- Reverse lookup is implemented as **constraints/filters** compiled into SQL joins (not a separate "reverse scope" mode).

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

Today, the codebase contains multiple implementations of "find tickets" and "find docs", spread across commands. They differ in:
- what directories they skip,
- whether they include `index.md`,
- whether invalid frontmatter is silently skipped or surfaced,
- how reverse lookup against `RelatedFiles` is matched (normalized vs raw string ops),
- and how ticket filtering behaves (exact match vs substring).

For a new engineer, the key takeaway is: **the concept of "the set of docs in the workspace" is not currently canonical**. That makes features harder to build and harder to debug.

### 3.1 Concrete examples of current inconsistencies

**Example 1: Ticket filtering semantics differ**

In `pkg/commands/list_tickets.go` (line 116), ticket filtering uses substring match:
```go
if settings.Ticket != "" && !strings.Contains(doc.Ticket, settings.Ticket) {
    return nil
}
```

But in `pkg/commands/list_docs.go` (line 195), ticket filtering uses exact match:
```go
if settings.Ticket != "" && doc.Ticket != settings.Ticket {
    return nil
}
```

**Impact**: A user running `docmgr list tickets --ticket MEN` might see `MEN-3475`, but `docmgr list docs --ticket MEN` returns nothing. This is confusing and breaks scripting.

**Example 2: Skip rules differ**

`internal/documents/walk.go` (line 42) skips directories starting with `_`:
```go
if d.Name() != "." && strings.HasPrefix(d.Name(), "_") {
    return fs.SkipDir
}
```

But `pkg/commands/list_docs.go` (line 172) uses raw `filepath.Walk` and does **not** skip `_*/` directories. It only skips `index.md` files (line 184).

**Impact**: `docmgr list docs` might include docs from `ttmp/_guidelines/`, while `docmgr search` (which uses `documents.WalkDocuments`) excludes them. The "set of docs" is ambiguous.

**Example 3: Reverse lookup normalization is inconsistent**

`pkg/commands/search.go` (lines 321-362) uses `paths.NewResolver` and `paths.MatchPaths` for reverse lookup:
```go
docResolver := paths.NewResolver(paths.ResolverOptions{
    DocsRoot:  settings.Root,
    DocPath:   path,
    ConfigDir: configDir,
})
fileQueryNorm := docResolver.Normalize(fileQueryRaw)
for _, entry := range related {
    if paths.MatchPaths(fileQueryNorm, entry.norm) {
        fileMatch = true
        // ...
    }
}
```

But `pkg/commands/doctor.go` uses a bespoke "candidate list" strategy (lines 186-188) that doesn't reuse `paths.Resolver`:
```go
// doctor.go uses separate logic: repo root, config dir, parent of config dir, cwd, etc.
```

**Impact**: A path normalized by `relate` might not match the same path checked by `doctor`, leading to false positives/negatives in validation.

**Example 4: Invalid frontmatter handling differs**

`pkg/commands/list_docs.go` (line 189) emits diagnostics in glaze mode but silently skips in human mode:
```go
doc, err := readDocumentFrontmatter(path)
if err != nil {
    docmgr.RenderTaxonomy(ctx, docmgrctx.NewListingSkip("list_docs", path, err.Error(), err))
    return nil
}
```

But `pkg/commands/search.go` (line 290) silently skips invalid frontmatter in all modes:
```go
doc, content, err := readDocumentWithContent(path)
if err != nil {
    return nil // Skip files with invalid frontmatter
}
```

**Impact**: Users can't reliably discover broken docs; some commands surface them, others don't.

### 3.2 Why this matters

These inconsistencies create several problems:

1. **User confusion**: Same flag, different behavior across commands.
2. **Debugging difficulty**: When a doc "should" appear but doesn't, you have to check which command's walker was used.
3. **Feature development friction**: Adding a new filter requires re-implementing traversal logic or choosing which existing walker to reuse (and inheriting its quirks).
4. **Testing complexity**: Each command's behavior must be tested independently, even though they're conceptually doing the same thing.

### 3.3 The solution: centralized API

This refactor introduces a single centralized API (`workspace.Workspace`) with a single ingestion step (SQLite index build) and a single query API (`QueryDocs`). Commands then become thin: they translate flags into `DocQuery`, run `QueryDocs`, and format output.

**Benefits**:
- **Single source of truth**: One canonical "set of docs" definition.
- **Consistent semantics**: Ticket filtering, skip rules, error handling are uniform.
- **Easier testing**: Test the `Workspace` API once; commands become trivial wrappers.
- **Easier feature development**: Add a filter? Extend `DocFilters` and the SQL compiler. No need to touch multiple walkers.

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

At a high level, docmgr workspaces are organized under a docs root (often `ttmp/`) with ticket directories. The ticket layout can be templated, but by default follows this structure:

- `ttmp/YYYY/MM/DD/TICKET--slug/index.md` (ticket index)
- `ttmp/YYYY/MM/DD/TICKET--slug/design/...` (ticket docs)
- `ttmp/YYYY/MM/DD/TICKET--slug/archive/...` (archived artifacts)
- `ttmp/YYYY/MM/DD/TICKET--slug/.meta/...` (implementation metadata; we skip this entirely)

Each markdown doc generally begins with YAML frontmatter. Two relevant frontmatter fields are:
- `Ticket`: ticket identifier (used for ticket scoping)
- `RelatedFiles`: list of code paths related to the doc

### 5.1 Path normalization: why it matters for reverse lookup

One of the trickiest parts of reverse lookup is **path normalization**. Consider this scenario:

A doc at `ttmp/2025/12/12/MEN-3475--refactor/design/api.md` has frontmatter:
```yaml
RelatedFiles:
  - Path: "pkg/commands/search.go"
  - Path: "../internal/workspace/discovery.go"
  - Path: "~/workspace/docmgr/cmd/docmgr/main.go"
```

When a user queries `--file pkg/commands/search.go`, how do we match it?

**The problem**: The same logical file can be represented in many ways:
- Absolute: `/home/user/workspace/docmgr/pkg/commands/search.go`
- Repo-relative: `pkg/commands/search.go`
- Doc-relative: `../../../../pkg/commands/search.go`
- With `~`: `~/workspace/docmgr/pkg/commands/search.go`
- With different separators: `pkg\commands\search.go` (Windows)

Additionally, it's often unclear which directory relative paths start from, and things can be inconsistent even within a single file. A user might mix repo-relative paths (`pkg/commands/search.go`) with doc-relative paths (`../internal/workspace/discovery.go`) in the same `RelatedFiles` list, making reliable matching difficult.

**The solution**: `internal/paths.Resolver` normalizes paths into multiple representations and picks a canonical key. The SQLite index stores these normalized keys, enabling reliable joins.

**How it works** (simplified):
1. `paths.Resolver` is constructed with anchors: `DocsRoot`, `DocPath`, `ConfigDir`, `RepoRoot`.
2. For each raw path, it tries resolving against each anchor in priority order (repo > doc > config > docs-root > docs-parent).
3. It produces a `NormalizedPath` with multiple representations: `Canonical`, `RepoRelative`, `DocsRelative`, `DocRelative`, `Abs`.
4. The index stores the canonical key (preferring `RepoRelative` if available) plus fallback keys.

**Example normalization**:
- Raw: `"pkg/commands/search.go"` (from frontmatter)
- Resolved against repo root: `RepoRelative = "pkg/commands/search.go"` → **canonical key**
- Also stored: `Abs = "/home/user/workspace/docmgr/pkg/commands/search.go"` (fallback)

When querying `--file pkg/commands/search.go`:
1. Normalize the query path using the same resolver logic.
2. Match against stored canonical keys (and fallbacks if needed).
3. Join `related_files` → `docs` to return matching documents.

This ensures that `"pkg/commands/search.go"` matches `"pkg/commands/search.go"` even if one was stored as repo-relative and the other was normalized differently.

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
    // Lists are treated as OR: match docs that reference any of the specified files/dirs
    RelatedFile []string  // match docs referencing any of these files
    RelatedDir  []string  // match docs referencing files in any of these directories

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

#### Example: "Reverse lookup: docs referencing a file"

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Filters: workspace.DocFilters{
        RelatedFile: []string{"pkg/commands/search.go"},
    },
    Options: workspace.DocQueryOptions{
        IncludeDiagnostics: true,
    },
})
```

#### Example: "Reverse lookup: docs referencing any of multiple files"

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Filters: workspace.DocFilters{
        RelatedFile: []string{
            "pkg/commands/search.go",
            "pkg/commands/list_docs.go",
            "internal/workspace/discovery.go",
        },
    },
    Options: workspace.DocQueryOptions{
        IncludeDiagnostics: true,
    },
})
```

Under the hood, these become joins against the `related_files` table with OR conditions (see §10 examples).

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

### 9.1 Required tables

- `docs`: one row per markdown document
- `doc_topics`: many-to-many relationship (doc → topics)
- `related_files`: one row per `RelatedFiles` entry

Optionally (later):
- `tickets`: for ticket-level edge cases and validation queries (can be derived from `docs` where `is_index=1`)

### 9.2 Recommended schema (detailed)

**`docs` table**:
```sql
CREATE TABLE docs (
    doc_id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,              -- absolute path to .md file
    ticket_id TEXT,                          -- from frontmatter Ticket field
    doc_type TEXT,                           -- from frontmatter DocType
    status TEXT,                             -- from frontmatter Status
    intent TEXT,                             -- from frontmatter Intent
    title TEXT,                              -- from frontmatter Title
    last_updated TEXT,                      -- ISO8601 timestamp from frontmatter
    
    -- Parse state
    parse_ok INTEGER NOT NULL DEFAULT 1,    -- 1 if frontmatter parsed successfully, 0 otherwise
    parse_err TEXT,                          -- error message if parse_ok=0
    
    -- Path category tags (for filtering)
    is_index INTEGER NOT NULL DEFAULT 0,    -- 1 if path ends with /index.md
    is_archived_path INTEGER NOT NULL DEFAULT 0,  -- 1 if path contains /archive/
    is_scripts_path INTEGER NOT NULL DEFAULT 0,   -- 1 if path contains /scripts/
    is_sources_path INTEGER NOT NULL DEFAULT 0,   -- 1 if path contains /sources/
    is_control_doc INTEGER NOT NULL DEFAULT 0,     -- 1 if basename is README.md, tasks.md, or changelog.md
    
    -- Optional: full document body (can be skipped initially for memory)
    body TEXT                                -- markdown content (only if IncludeBody=true)
);
CREATE INDEX idx_docs_ticket_id ON docs(ticket_id);
CREATE INDEX idx_docs_parse_ok ON docs(parse_ok);
CREATE INDEX idx_docs_path_tags ON docs(is_archived_path, is_scripts_path, is_control_doc);
```

**`doc_topics` table**:
```sql
CREATE TABLE doc_topics (
    doc_id INTEGER NOT NULL,
    topic_lower TEXT NOT NULL,              -- lowercase topic for case-insensitive matching
    topic_original TEXT,                      -- original case (for display)
    PRIMARY KEY (doc_id, topic_lower),
    FOREIGN KEY (doc_id) REFERENCES docs(doc_id) ON DELETE CASCADE
);
CREATE INDEX idx_doc_topics_topic ON doc_topics(topic_lower);
```

**`related_files` table**:
```sql
CREATE TABLE related_files (
    rf_id INTEGER PRIMARY KEY,
    doc_id INTEGER NOT NULL,
    note TEXT,                                -- optional note from RelatedFiles entry
    
    -- Normalized path keys (multiple representations for reliable matching)
    norm_repo_rel TEXT,                      -- repo-relative path (preferred canonical key)
    norm_abs TEXT,                           -- absolute path (fallback)
    norm_clean TEXT,                         -- cleaned relative path (fallback)
    anchor TEXT,                             -- which anchor was used (repo/doc/config/etc)
    
    -- Original raw path from frontmatter (for display/debugging)
    raw_path TEXT,
    
    FOREIGN KEY (doc_id) REFERENCES docs(doc_id) ON DELETE CASCADE
);
CREATE INDEX idx_related_files_doc_id ON related_files(doc_id);
CREATE INDEX idx_related_files_norm_repo_rel ON related_files(norm_repo_rel);
CREATE INDEX idx_related_files_norm_abs ON related_files(norm_abs);
```

**Why multiple normalized keys?** Path normalization can fail or produce different results depending on context. Storing multiple representations allows the query compiler to try fallback matching strategies:
1. First try `norm_repo_rel` (most reliable if repo root is known)
2. Fall back to `norm_abs` (works even if repo root detection failed)
3. Fall back to `norm_clean` (last resort)

This matches the fallback strategy in `paths.Resolver.Normalize()`.

## 10) Concrete query examples (DocQuery → SQL mental model)

This section gives new engineers an intuition for how the API maps to the index. (Exact SQL may differ; this is conceptual.)

### 10.1 Simple ticket scope + status filter

**DocQuery**:
```go
DocQuery{
    Scope: Scope{Kind: ScopeTicket, TicketID: "MEN-3475"},
    Filters: DocFilters{Status: "active"},
    Options: DocQueryOptions{IncludeErrors: false},
}
```

**Compiled SQL** (conceptual):
```sql
SELECT docs.* FROM docs
WHERE docs.ticket_id = 'MEN-3475'
  AND docs.status = 'active'
  AND docs.parse_ok = 1  -- exclude invalid frontmatter
ORDER BY docs.last_updated DESC;
```

**Notes**:
- `ScopeTicket` adds `WHERE docs.ticket_id = ?`
- `Status` filter adds `AND docs.status = ?`
- `IncludeErrors=false` adds `AND docs.parse_ok = 1`
- Default ordering is `last_updated DESC` (newest first)

### 10.2 Reverse lookup by file (with fallback matching)

**DocQuery**:
```go
DocQuery{
    Scope: Scope{Kind: ScopeRepo},
    Filters: DocFilters{
        RelatedFile: []string{"pkg/commands/search.go"},
    },
    Options: DocQueryOptions{IncludeDiagnostics: true},
}
```

**Compiled SQL** (conceptual):
```sql
SELECT DISTINCT docs.* FROM docs
INNER JOIN related_files rf ON rf.doc_id = docs.doc_id
WHERE (
    rf.norm_repo_rel = 'pkg/commands/search.go'  -- try canonical key first
    OR rf.norm_abs = '/absolute/path/to/pkg/commands/search.go'  -- fallback
    OR rf.norm_clean = 'pkg/commands/search.go'  -- last resort
)
  AND docs.parse_ok = 1
ORDER BY docs.last_updated DESC;
```

**DocQuery with multiple files** (OR semantics):
```go
DocQuery{
    Scope: Scope{Kind: ScopeRepo},
    Filters: DocFilters{
        RelatedFile: []string{
            "pkg/commands/search.go",
            "pkg/commands/list_docs.go",
        },
    },
}
```

**Compiled SQL** (conceptual):
```sql
SELECT DISTINCT docs.* FROM docs
INNER JOIN related_files rf ON rf.doc_id = docs.doc_id
WHERE (
    -- First file
    (rf.norm_repo_rel = 'pkg/commands/search.go' OR rf.norm_abs = '...' OR rf.norm_clean = '...')
    OR
    -- Second file (OR semantics)
    (rf.norm_repo_rel = 'pkg/commands/list_docs.go' OR rf.norm_abs = '...' OR rf.norm_clean = '...')
)
  AND docs.parse_ok = 1
ORDER BY docs.last_updated DESC;
```

**Notes**:
- `RelatedFile` filter triggers a `JOIN` with `related_files`
- Query compiler normalizes the input path and tries multiple matching strategies
- If normalization fails, a diagnostic is emitted (if `IncludeDiagnostics=true`)
- `DISTINCT` prevents duplicate docs if a doc has multiple `RelatedFiles` entries matching the query

### 10.3 Complex query: ticket + topics + exclude noisy categories

**DocQuery**:
```go
DocQuery{
    Scope: Scope{Kind: ScopeTicket, TicketID: "MEN-3475"},
    Filters: DocFilters{
        TopicsAny: []string{"refactor", "tickets"},
    },
    Options: DocQueryOptions{
        IncludeArchivedPath: false,
        IncludeControlDocs:  false,
        IncludeScriptsPath:  false,
    },
}
```

**Compiled SQL** (conceptual):
```sql
SELECT DISTINCT docs.* FROM docs
INNER JOIN doc_topics dt ON dt.doc_id = docs.doc_id
WHERE docs.ticket_id = 'MEN-3475'
  AND docs.parse_ok = 1
  AND docs.is_archived_path = 0  -- exclude archive/
  AND docs.is_control_doc = 0     -- exclude README.md, tasks.md, changelog.md
  AND docs.is_scripts_path = 0     -- exclude scripts/
  AND dt.topic_lower IN ('refactor', 'tickets')  -- match any topic
GROUP BY docs.doc_id
ORDER BY docs.last_updated DESC;
```

**Notes**:
- `TopicsAny` requires a `JOIN` with `doc_topics` and `IN` clause
- Multiple path category exclusions are combined with `AND`
- `GROUP BY` ensures each doc appears once even if it matches multiple topics

### 10.4 Reverse lookup by directory

**DocQuery**:
```go
DocQuery{
    Scope: Scope{Kind: ScopeRepo},
    Filters: DocFilters{
        RelatedDir: []string{"pkg/commands/"},
    },
    Options: DocQueryOptions{IncludeDiagnostics: true},
}
```

**Compiled SQL** (conceptual):
```sql
SELECT DISTINCT docs.* FROM docs
INNER JOIN related_files rf ON rf.doc_id = docs.doc_id
WHERE (
    rf.norm_repo_rel LIKE 'pkg/commands/%'  -- directory prefix match
    OR rf.norm_abs LIKE '%/pkg/commands/%'  -- absolute fallback
)
  AND docs.parse_ok = 1
ORDER BY docs.last_updated DESC;
```

**DocQuery with multiple directories** (OR semantics):
```go
DocQuery{
    Scope: Scope{Kind: ScopeRepo},
    Filters: DocFilters{
        RelatedDir: []string{
            "pkg/commands/",
            "internal/workspace/",
        },
    },
}
```

**Compiled SQL** (conceptual):
```sql
SELECT DISTINCT docs.* FROM docs
INNER JOIN related_files rf ON rf.doc_id = docs.doc_id
WHERE (
    rf.norm_repo_rel LIKE 'pkg/commands/%' OR rf.norm_repo_rel LIKE 'internal/workspace/%'
    OR rf.norm_abs LIKE '%/pkg/commands/%' OR rf.norm_abs LIKE '%/internal/workspace/%'
)
  AND docs.parse_ok = 1
ORDER BY docs.last_updated DESC;
```

**Notes**:
- `RelatedDir` uses `LIKE` with `%` suffix for prefix matching
- Directory normalization follows the same fallback strategy as file matching
- **Open question**: Should `RelatedDir` also match docs whose `path` is inside the directory? (Currently deferred; see §12)

### 10.5 Include/exclude tagged path categories (detailed)

The `DocQueryOptions` visibility flags control which path categories appear in results:

**If `IncludeArchivedPath=false`**, SQL adds:
```sql
AND docs.is_archived_path = 0
```

**If `IncludeControlDocs=false`**, SQL adds:
```sql
AND docs.is_control_doc = 0
```

**If `IncludeScriptsPath=false`**, SQL adds:
```sql
AND docs.is_scripts_path = 0
```

**Rationale**: These categories are indexed (for discoverability and reverse lookup) but hidden by default to keep results clean. Users can opt-in via flags.

### 10.6 Invalid frontmatter diagnostics

**DocQuery**:
```go
DocQuery{
    Scope: Scope{Kind: ScopeRepo},
    Options: DocQueryOptions{
        IncludeErrors: false,        -- exclude from Docs
        IncludeDiagnostics: true,    -- but emit diagnostics
    },
}
```

**Compiled SQL** (conceptual):
```sql
-- Main query: only valid docs
SELECT docs.* FROM docs
WHERE docs.parse_ok = 1
ORDER BY docs.last_updated DESC;

-- Separate query for diagnostics: invalid docs
SELECT docs.path, docs.parse_err FROM docs
WHERE docs.parse_ok = 0;
```

**Post-processing**:
- Valid docs go into `DocQueryResult.Docs`
- Invalid docs are converted to `core.Taxonomy` diagnostics and added to `DocQueryResult.Diagnostics`
- Each diagnostic includes:
  - `Path`: the doc path
  - `Context`: a `FrontmatterParseContext` with parse error details
  - `Severity`: `SeverityError`
  - `Cause`: the original parse error

This allows commands like `doctor` to surface broken docs without polluting normal search results.

## 11) Migration plan (API-level)

### 11.1 Phase 1: Core infrastructure

**Goal**: Build the `workspace.Workspace` API and SQLite index.

**Tasks**:
1. Create `internal/workspace/workspace.go` with:
   - `Workspace` struct
   - `WorkspaceContext` struct
   - `NewWorkspaceFromContext` constructor
   - `DiscoverWorkspace` helper (best-effort discovery)
2. Create `internal/workspace/index.go` with:
   - SQLite schema creation
   - Ingestion logic (walk docs, parse frontmatter, normalize paths, insert rows)
   - Skip rule implementation (canonical policy)
3. Create `internal/workspace/query.go` with:
   - `DocQuery`, `DocFilters`, `DocQueryOptions` types
   - `DocHandle`, `DocQueryResult` types
   - `QueryDocs` method (SQL compilation + execution)
   - Query compiler (DocQuery → SQL with proper joins/WHERE clauses)

**Success criteria**:
- `DiscoverWorkspace` successfully builds an index from a real workspace
- `QueryDocs` with simple filters (ticket, status, doc-type) returns correct results
- Invalid frontmatter docs are indexed with `parse_ok=0`
- Path normalization produces consistent keys

### 11.2 Phase 2: Port commands (one at a time)

**Strategy**: Port commands incrementally, keeping old code alongside new code initially. Use feature flags or separate code paths to compare behavior.

#### 11.2.1 Port `list docs`

**Current behavior** (`pkg/commands/list_docs.go`):
- Walks entire docs root
- Skips `index.md` files
- Filters by ticket (exact match), status, doc-type, topics
- Silently skips invalid frontmatter (in human mode) or emits diagnostics (in glaze mode)

**New implementation**:
```go
func (c *ListDocsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    settings := &ListDocsSettings{}
    // ... parse settings ...
    
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
        RootOverride: settings.Root,
    })
    if err != nil {
        return fmt.Errorf("failed to discover workspace: %w", err)
    }
    
    query := workspace.DocQuery{
        Scope: workspace.Scope{Kind: workspace.ScopeRepo},
        Filters: workspace.DocFilters{
            Ticket:  settings.Ticket,
            Status:  settings.Status,
            DocType: settings.DocType,
            TopicsAny: settings.Topics,
        },
        Options: workspace.DocQueryOptions{
            IncludeErrors: false,  // match current behavior
            IncludeDiagnostics: true,  // emit skip diagnostics in glaze mode
            IncludeArchivedPath: false,  // hide by default
            IncludeControlDocs: false,   // hide by default
            IncludeScriptsPath: false,   // hide by default
        },
    }
    
    res, err := ws.QueryDocs(ctx, query)
    if err != nil {
        return fmt.Errorf("query failed: %w", err)
    }
    
    // Emit diagnostics if any
    for _, diag := range res.Diagnostics {
        docmgr.RenderTaxonomy(ctx, diag)
    }
    
    // Emit rows
    for _, handle := range res.Docs {
        // ... format row from handle.Doc ...
        gp.AddRow(ctx, row)
    }
    
    return nil
}
```

**Changes**:
- Replace `filepath.Walk` with `ws.QueryDocs`
- Remove custom skip logic (now handled by canonical skip rules)
- Remove custom frontmatter parsing (now handled by index)
- Use `DocQuery` filters instead of manual filtering

**Validation**:
- Compare output of old vs new implementation on a real workspace
- Ensure ticket filtering is exact match (not substring)
- Ensure invalid frontmatter handling matches current behavior

#### 11.2.2 Port `search`

**Current behavior** (`pkg/commands/search.go`):
- Supports content search (full-text)
- Supports reverse lookup (`--file`, `--dir`)
- Uses `paths.Resolver` for normalization
- Complex filtering logic (ticket, topics, doc-type, status, dates, external sources)

**New implementation**:
```go
func (c *SearchCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    // ... parse settings ...
    
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
        RootOverride: settings.Root,
    })
    if err != nil {
        return fmt.Errorf("failed to discover workspace: %w", err)
    }
    
    query := workspace.DocQuery{
        Scope: workspace.Scope{Kind: workspace.ScopeRepo},
        Filters: workspace.DocFilters{
            Ticket:    settings.Ticket,
            DocType:   settings.DocType,
            Status:    settings.Status,
            TopicsAny: settings.Topics,
            // Reverse lookup (OR semantics): empty strings result in nil slices
            RelatedFile: func() []string {
                if settings.File != "" {
                    return []string{settings.File}
                }
                return nil
            }(),
            RelatedDir: func() []string {
                if settings.Dir != "" {
                    return []string{settings.Dir}
                }
                return nil
            }(),
        },
        Options: workspace.DocQueryOptions{
            IncludeBody: settings.Query != "",  // only load body if content search needed
            IncludeDiagnostics: true,
        },
    }
    
    res, err := ws.QueryDocs(ctx, query)
    if err != nil {
        return fmt.Errorf("query failed: %w", err)
    }
    
    // Apply content search filter (post-query, since SQLite FTS is deferred)
    for _, handle := range res.Docs {
        if settings.Query != "" {
            if !strings.Contains(strings.ToLower(handle.Body), strings.ToLower(settings.Query)) {
                continue
            }
        }
        // ... format and emit row ...
    }
    
    return nil
}
```

**Changes**:
- Replace `filepath.Walk` + manual filtering with `ws.QueryDocs`
- Reverse lookup (`--file`, `--dir`) now uses `RelatedFile`/`RelatedDir` filters
- Content search remains post-query (can be optimized later with SQLite FTS)
- Date filtering and external source filtering can be added to `DocFilters` later

**Validation**:
- Test reverse lookup with various path formats (absolute, relative, `~`)
- Ensure normalization matches current `paths.Resolver` behavior
- Compare search results with old implementation

#### 11.2.3 Port `doctor`

**Current behavior** (`pkg/commands/doctor.go`):
- Uses `CollectTicketWorkspaces` and `CollectTicketScaffoldsWithoutIndex`
- Validates `index.md` frontmatter
- Walks docs within tickets and validates frontmatter
- Checks `RelatedFiles` existence using bespoke candidate list

**New implementation**:
```go
func (c *DoctorCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{})
    if err != nil {
        return fmt.Errorf("failed to discover workspace: %w", err)
    }
    
    // Query all docs including errors
    query := workspace.DocQuery{
        Scope: workspace.Scope{Kind: workspace.ScopeRepo},
        Options: workspace.DocQueryOptions{
            IncludeErrors: true,      // include invalid frontmatter docs
            IncludeDiagnostics: true, // get diagnostics
        },
    }
    
    res, err := ws.QueryDocs(ctx, query)
    if err != nil {
        return fmt.Errorf("query failed: %w", err)
    }
    
    // Emit diagnostics
    for _, diag := range res.Diagnostics {
        docmgr.RenderTaxonomy(ctx, diag)
    }
    
    // Validate RelatedFiles existence (using same normalization as index)
    for _, handle := range res.Docs {
        if handle.Doc == nil {
            continue  // skip invalid frontmatter (already in diagnostics)
        }
        for _, rf := range handle.Doc.RelatedFiles {
            // Use workspace's resolver (same as index) to check existence
            // ... validation logic ...
        }
    }
    
    // Ticket-level checks (can use QueryTickets API when available)
    // ...
    
    return nil
}
```

**Changes**:
- Replace `CollectTicketWorkspaces` with `QueryDocs(IncludeErrors=true)`
- Use workspace's normalization (same as index) for `RelatedFiles` validation
- Ticket-level validation can use a future `QueryTickets` API

**Validation**:
- Ensure all current diagnostics are still emitted
- Ensure `RelatedFiles` validation uses consistent normalization

#### 11.2.4 Port `relate`

**Current behavior** (`pkg/commands/relate.go`):
- Finds target doc (by `--doc` path or `--ticket` → `index.md`)
- Normalizes related file paths using `paths.Resolver`
- Updates frontmatter

**New implementation**:
```go
func (c *RelateCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{})
    if err != nil {
        return fmt.Errorf("failed to discover workspace: %w", err)
    }
    
    // Find target doc using QueryDocs
    var targetDoc *workspace.DocHandle
    if settings.Doc != "" {
        query := workspace.DocQuery{
            Scope: workspace.Scope{Kind: workspace.ScopeDoc, DocPath: settings.Doc},
        }
        res, err := ws.QueryDocs(ctx, query)
        if err != nil || len(res.Docs) == 0 {
            return fmt.Errorf("doc not found: %s", settings.Doc)
        }
        targetDoc = &res.Docs[0]
    } else if settings.Ticket != "" {
        query := workspace.DocQuery{
            Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: settings.Ticket},
            Filters: workspace.DocFilters{Ticket: settings.Ticket},
            Options: workspace.DocQueryOptions{IncludeBody: false},
        }
        // Filter to index.md (can add is_index filter later)
        // ...
    }
    
    // Normalize related file using workspace's resolver (same as index)
    resolver := ws.Resolver(targetDoc.Path)
    normalized := resolver.Normalize(settings.RelatedFile)
    
    // Update frontmatter
    // ...
    
    return nil
}
```

**Changes**:
- Use `QueryDocs` to find target doc instead of manual path resolution
- Use workspace's resolver (same normalization as index) for consistency
- Remove `findTicketDirectory` helper (replaced by `QueryDocs`)

**Validation**:
- Ensure normalization matches index normalization
- Test with various path formats

### 11.3 Phase 3: Cleanup

**Tasks**:
1. Remove duplicated walkers:
   - `findTicketDirectory` helper (replaced by `QueryDocs`)
   - Ad-hoc `filepath.Walk` loops in commands (replaced by `QueryDocs`)
2. Deprecate or remove inconsistent skip logic:
   - Commands no longer need custom skip predicates
   - All skip logic is centralized in `workspace` ingestion
3. Update tests:
   - Replace command-level tests with `Workspace` API tests
   - Commands become thin wrappers (minimal testing needed)

**Success criteria**:
- No commands use `filepath.Walk` directly (except for non-doc files)
- No commands implement custom skip logic
- All commands use `workspace.Workspace` API
- Test coverage maintained or improved

## 12) Edge cases and gotchas (for implementers)

### 12.1 Path normalization edge cases

**Case: Repo root detection fails**
- **What happens**: `paths.Resolver` falls back to doc-relative, config-relative, or docs-root-relative normalization
- **Impact**: `norm_repo_rel` will be empty; matching must fall back to `norm_abs` or `norm_clean`
- **Mitigation**: Store multiple normalized keys in `related_files` table; query compiler tries all of them

**Case: Relative paths with `../`**
- **What happens**: `paths.Resolver` resolves `../` against anchors, but if it escapes the anchor, normalization may fail
- **Impact**: `norm_repo_rel` might be empty; `norm_abs` might be the only reliable key
- **Mitigation**: Always store `norm_abs` as a fallback

**Case: Paths with `~` (home directory)**
- **What happens**: `paths.Resolver.expandHome()` expands `~` to actual home directory
- **Impact**: `~/workspace/file.go` becomes `/home/user/workspace/file.go` in `norm_abs`
- **Mitigation**: Store both original (`raw_path`) and normalized forms

### 12.2 Invalid frontmatter handling

**Case: YAML syntax error**
- **What happens**: `documents.ReadDocumentWithFrontmatter` returns an error; doc is indexed with `parse_ok=0`
- **Impact**: Doc appears in diagnostics but not in default query results
- **Mitigation**: Ensure `parse_err` contains enough context (file, line, column) for repair workflows

**Case: Valid YAML but invalid schema**
- **What happens**: YAML parses but doesn't match `models.Document` struct; doc is indexed with `parse_ok=1` but some fields are empty
- **Impact**: Doc appears in results but may have missing `Ticket`, `DocType`, etc.
- **Note**: This is different from parse errors; schema validation is deferred (see Q8)

### 12.3 Skip rules edge cases

**Case: Nested `archive/` directories**
- **What happens**: A doc at `ticket/archive/subdir/doc.md` is tagged `is_archived_path=1`
- **Impact**: Default queries exclude it; users must opt-in with `IncludeArchivedPath=true`
- **Rationale**: Archive is for historical artifacts; default queries focus on active docs

**Case: `index.md` in subdirectories**
- **What happens**: A doc at `ticket/design/index.md` is tagged `is_index=1` but is not a ticket index
- **Impact**: It's included in doc queries (per Decision 12); only ticket root `index.md` is special
- **Rationale**: `index.md` files can be useful navigation docs within ticket subdirectories

**Case: Control docs in subdirectories**
- **What happens**: A doc at `ticket/design/README.md` is **not** tagged `is_control_doc=1` (only root-level control docs are tagged)
- **Impact**: It appears in default queries (unlike root-level `README.md`)
- **Rationale**: Control docs are only special at ticket root; subdirectory `README.md` files are regular docs

### 12.4 Query compilation edge cases

**Case: Contradictory scope + filters**
- **What happens**: User specifies `ScopeTicket` with `TicketID="MEN-3475"` but also `Filters.Ticket="MEN-9999"`
- **Impact**: Query compiler detects contradiction and returns hard error (per Decision 16/D2)
- **Mitigation**: Validate query before compilation; return clear error message

**Case: Reverse lookup normalization failure**
- **What happens**: User queries `--file "nonexistent/path.go"` and normalization fails (no anchor matches)
- **Impact**: Query compiler emits diagnostic and applies fallback matching strategy
- **Mitigation**: Try multiple normalized keys; if all fail, return empty results with diagnostic explaining why

**Case: Empty query (no filters)**
- **What happens**: User calls `QueryDocs` with empty `DocFilters` and `ScopeRepo`
- **Impact**: Returns all docs in repo (subject to `IncludeArchivedPath`/`IncludeControlDocs` defaults)
- **Rationale**: Empty query is valid; useful for "list everything" commands

### 12.5 Performance considerations

**Case: Large workspace (10k+ docs)**
- **What happens**: Index build time increases linearly with doc count
- **Impact**: CLI startup time may be noticeable (100ms-1s for large workspaces)
- **Mitigation**: Index build is eager but happens once per CLI invocation; queries are fast (SQL)

**Case: Memory usage**
- **What happens**: Index stores doc metadata; if `IncludeBody=true`, full markdown content is stored
- **Impact**: Memory usage scales with workspace size
- **Mitigation**: `IncludeBody` defaults to `false`; only load body when content search is needed

**Case: Concurrent access**
- **What happens**: Multiple CLI invocations build separate in-memory indexes
- **Impact**: No shared state; each invocation is independent
- **Rationale**: Per-invocation index is simpler than shared persistent index (deferred to future)

## 13) Open questions (explicitly deferred)

- **Topic match semantics** (any vs all vs configurable) — deferred (Decision 13 section). Current implementation uses "any" (`IN` clause); future enhancement could support "all" (`GROUP BY` + `HAVING COUNT(DISTINCT topic) = ?`).
- **Directory reverse lookup semantics** for `RelatedDir` (match `related_files` vs doc `path` vs either) — deferred. Current design only matches `related_files`; could be extended to also match docs whose path is inside the directory.
- **DiscoverWorkspace failure/fallback behavior** for missing config/repo-root — parked. Current design requires `Root`, `ConfigDir`, `RepoRoot`; best-effort discovery may fail; policy TBD.
- **Richer boolean `Where Expr` in `DocQuery`** — postponed until needed. Current filters are AND-ed together; future enhancement could support `(A OR B) AND C` expressions.
- **SQLite FTS (Full-Text Search)** for content search — deferred. Current design does post-query content filtering; FTS would enable SQL-level content search.
- **Persistent index** (on-disk DB) — deferred. Current design is per-invocation in-memory; persistent index would enable faster startup and cross-invocation queries.


