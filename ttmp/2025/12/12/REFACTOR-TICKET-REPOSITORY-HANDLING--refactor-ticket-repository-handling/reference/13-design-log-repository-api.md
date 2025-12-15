---
Title: 'Design Log — Repository API (workspace.Workspace)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T15:45:13-05:00
---

# Design Log — Repository API (workspace.Workspace)

This document records **decisions** made during the interactive design session for a centralized repository/workspace lookup API.

## Decisions (chronological)

### 2025-12-12 — Decision 1: Package + type placement

- **Decision**: Extend `internal/workspace` with a stateful object (e.g. `workspace.Workspace`) as the primary entry point (instead of creating `internal/repository`).
- **Why**: Keeps “workspace context + discovery” co-located with existing resolution utilities while still allowing a clean API surface for lookup/query methods.
- **Related debate**: Q1 (Round 10) + Q11 (Round 9).

### 2025-12-12 — Decision 2: Construction style (discovery vs injection)

- **Decision**: Support **both**:
  - a **discovering constructor** (CLI-friendly; resolves root/config/repo root once), and
  - an **injected constructor** (test-friendly; accepts fully-specified context).
- **Why**: CLI ergonomics + unit-test determinism.
- **Related debate**: Q1 (Round 10) + Q11 (Round 9) + Q8 (Round 8).

### 2025-12-12 — Decision 3: Required context vs best-effort

- **Decision**: `workspace.Workspace` should have a **fully specified context** available:
  - required: `Root`, `ConfigDir`, `RepoRoot`.
- **And**: provide a helper/builder that can do **best-effort discovery** (root/config/repo-root resolution) so top-level commands don’t need to pass every detail.
- **Why**: internal APIs get strong invariants; CLI stays ergonomic.
- **Related debate**: Q1 (Round 10) + Q11 (Round 9).

### 2025-12-12 — Decision 4: Primary QueryDocs shape

- **Decision**: Use a **structured request/response** shape:
  - `QueryDocs(ctx, req DocQuery) (DocQueryResult, error)`
- **Why**: avoids “grab-bag” signatures, supports future extensions, keeps semantics centralized.
- **Related debate**: Q11 (Round 9).

### 2025-12-12 — Decision 5: Skip rules (archive/ handling)

- **Decision**: **Include** `archive/` in the index, but **tag** entries (e.g. `is_archived_path = true`) so default queries can exclude it without losing data.
- **Why**: preserves discoverability and reverse lookup for archived artifacts while keeping “normal” results clean.
- **Related debate**: Q3 (Round 12) + SQLite (Round 13).

### 2025-12-12 — Decision 6: Skip rules (.meta/ handling)

- **Decision**: **Skip `.meta/` entirely** at ingest time (not indexed / not queryable).
- **Why**: `.meta/` is primarily for non-doc artifacts (e.g. `sources.yaml`), and indexing it risks pulling in implementation metadata; prefer explicit exposure later if needed.
- **Related debate**: Q3 (Round 12) + SQLite (Round 13).

### 2025-12-12 — Decision 7: Skip rules (underscore dirs)

- **Decision**: **Skip all underscore dirs** (`_*/`) at ingest time.
- **Why**: underscore dirs are reserved for non-content support artifacts (templates/guidelines/etc.) and should not pollute the queryable corpus.
- **Related debate**: Q3 (Round 12).

### 2025-12-12 — Decision 8: Skip rules (`scripts/` handling)

- **Decision**: **Include** `scripts/` in the index, but **tag** entries (e.g. `is_scripts_path = true`) so defaults can hide it.
- **Why**: scripts often include operational notes and markdown; we want them discoverable but not always in default listings.
- **Related debate**: Q3 (Round 12).

### 2025-12-12 — Decision 9: Skip rules (`sources/` handling)

- **Decision**: **Include** `sources/` in the index and **tag** entries (e.g. `is_sources_path = true`), but **do not hide by default**.
- **Why**: imported sources can be relevant to queries/reverse lookup; tag supports future default-hiding without losing data.
- **Related debate**: Q3 (Round 12) + SQLite (Round 13).

### 2025-12-12 — Decision 10: Control docs at ticket root (`README.md`, `tasks.md`, `changelog.md`)

- **Decision**: **Include** them, but **tag** (e.g. `is_control_doc = true`) so defaults can hide them.
- **Why**: they’re useful, but often not part of the “core doc set” users want when listing/searching.
- **Related debate**: Q3 (Round 12) + SQLite (Round 13).

### 2025-12-12 — Decision 11: Invalid frontmatter docs in the index

- **Decision**: **Index** invalid-frontmatter docs as rows with `parse_ok = false` and `parse_err` populated, but **exclude from default queries** (unless explicitly requested).
- **Why**: supports doctor/repair workflows and diagnostics without polluting normal results.
- **Related debate**: Q2 (Round 11) + Q11 (Round 9).

### 2025-12-12 — Decision 12: `index.md` (ticket index docs) in doc queries by default

- **Decision**: **Include** `index.md` by default in doc queries.
- **Why**: ticket indices are central knowledge nodes; inclusion improves search and reverse lookup coverage.
- **Related debate**: Q3 (Round 12).

### 2025-12-12 — Decision 13: Ticket filter semantics

- **Decision**: Ticket filter is **exact match** (`ticket_id = ?`) by default.
- **Why**: deterministic and scriptable; avoids surprising partial matches.
- **Related debate**: Q3 (Round 12).

### 2025-12-12 — Deferred: Topic match semantics + dir reverse semantics

- **Deferred**:
  - **Topic match semantics (Q14)**: not decided yet; easy to support multiple behaviors given SQL backend.
  - **Directory reverse lookup semantics (Q15)**: not decided yet; focus remains on big-picture API.

### 2025-12-12 — Decision 14: SQLite index lifecycle (build timing)

- **Decision**: Build the in-memory SQLite index **eagerly** during workspace construction / discovery (workspace is query-ready).
- **Why**: commands and library call sites can assume a ready-to-query workspace; avoids “first query pays surprise cost”.
- **Related debate**: SQLite (Round 13) + Q11 (Round 9).

### 2025-12-12 — Decision 15: No refresh in the short term (per-invocation index)

- **Decision**: For now, the CLI builds a fresh **in-memory SQLite index once per invocation** and exits; no long-running workspace sessions and no `RefreshIndex()` API yet.
- **Why**: keeps the first implementation simple and deterministic; refresh/daemon behavior can be added later.
- **Related debate**: SQLite (Round 13).

### 2025-12-12 — Decision 16: QueryDocs diagnostics + error contract

- **Invalid-frontmatter rows (D1)**:
  - **Decision**: When `IncludeErrors=false`, **exclude** invalid-frontmatter docs from results **but** emit **Diagnostics** explaining what was skipped.
- **Contradictory / invalid queries (D2)**:
  - **Decision**: Return a **hard error** (fail-fast).
- **Reverse lookup normalization failure (D3)**:
  - **Decision**: Emit **Diagnostics** and apply a documented **fallback matching strategy** (e.g. try alternate normalized keys), rather than silently failing.
- **Why**: preserve strong correctness guarantees while keeping the UX explainable (avoid silent empty sets).
- **Related debate**: Q2 (Round 11) + Q11 (Round 9) + SQLite (Round 13).

### 2025-12-12 — Note / open decision: Taxonomy “message” + “suggestions” fields

- **Observation**: Some domains already carry suggestion-like data in the taxonomy **Context** payload.
  - Example: `docmgrctx.FrontmatterParseContext` contains `Problem`, `Snippet`, and `Fixes []string`.
- **Open decision**: Make “message” and “suggestions/hints” **uniformly available** for UIs without requiring domain-specific type assertions:
  - Option 1: extend `core.Taxonomy` with optional fields like:
    - `Message string`
    - `Suggestions []string` (or `Hints []string`)
  - Option 2: standardize via an interface on `ContextPayload` (e.g. optional `Hints()` / `Suggestions()`).
- **Goal**: enable `QueryDocs` diagnostics (and other commands) to present consistent guidance (what happened + what to do next) across domains.

## Meta

- **Operational note**: Do not run `docmgr doc relate` automatically; only run it when explicitly requested.

## Interaction write-up (post-hoc narrative)

This section summarizes the interactive design session decisions and the reasoning behind them, tied back to debate rounds.

### 1) Package placement: extend `workspace`

We started by choosing where the new centralized lookup API should live:
- **Option**: new `internal/repository` package (wrap `workspace/documents/paths`)
- **Option**: extend `internal/workspace` with a first-class object

**Decision**: extend `internal/workspace` (a `workspace.Workspace`-like object).

**Rationale** (from Q1/Q11 discussion):
- Keeps workspace resolution concerns (root/config/repo-root) co-located with the existing resolution utilities.
- Still allows a coherent “repository API” surface without scattering “find root / find config / find repo root” logic across commands.

**Debate references**:
- Q1 / Round 10 (object responsibilities + boundaries)
- Q11 / Round 9 (QueryDocs wants a resolved context holder)

### 2) Construction: support both “discover” and “inject”

We then chose how the object is constructed:
- **Discovering constructor**: CLI-friendly best-effort discovery (resolve root/config/repo-root once).
- **Injected constructor**: test-friendly fully specified context.

**Decision**: support **both** (discover + inject).

**Rationale**:
- CLI stays ergonomic.
- Tests and library use can be deterministic and explicit.

**Debate references**:
- Q1 / Round 10 (context plumbing)
- Q11 / Round 9 (avoid re-resolution per query)
- Q8 / Round 8 (keep vocabulary/validation separate; but workspace can still locate/configure paths)

### 3) Strong invariants internally; best-effort builder for top-level commands

We discussed what the core object should require vs what the helper can discover.

**Decision**:
- Core `workspace.Workspace` should have a **fully specified context** available, with required anchors:
  - `Root`, `ConfigDir`, `RepoRoot`.
- Provide a **best-effort builder/helper** that can discover those anchors so top-level commands don’t need to pass all details.

**Rationale**:
- Internal code (query compiler, normalization) benefits from strong invariants.
- Top-level UX stays simple.

**Debate references**:
- Q1 / Round 10 (state vs convenience)
- Q11 / Round 9 (context requirements)

We explicitly parked a follow-up question:
- how the best-effort builder should behave when config or repo-root discovery fails (strict error vs fallback vs mixed + warnings).

### 4) QueryDocs: structured request/response

We selected the “big picture” `QueryDocs` shape (Q11 topic).

**Decision**: structured request/response:
- `QueryDocs(ctx, DocQuery) (DocQueryResult, error)`

**Rationale**:
- We expect iteration (filters, ignore rules, error policy, ordering).
- A request struct avoids signature churn and avoids an unbounded options grab-bag.

**Debate references**:
- Q11 / Round 9 (structured req/resp vs minimal signature vs options struct)

### 5) SQLite backend insight: reverse lookup + correlations become joins, but semantics remain

You proposed implementing lookup/search by loading all workspace data into an **in-memory SQLite** database.

We then explored how this changes the reverse lookup and lookup structure:
- Reverse lookup becomes a **join** (`docs` ↔ `related_files`) instead of manual loops.
- Complex combinations (reverse lookup + topics + status + ticket) become straightforward SQL.
- Deterministic ordering becomes trivial (`ORDER BY ...`) everywhere.

**Key point**:
SQLite doesn’t remove the need for **canonical semantics** (skip rules, invalid-frontmatter handling, index.md inclusion, path normalization keys). It makes executing those semantics easier and more consistent.

This discussion was captured in:
- `reference/14-debate-round-13-sqlite-index-backend-influences-lookup-and-reverse-lookup.md`

### 6) Reverse lookup modeling: “reverse as filters” reasoning (Option B)

We discussed modeling reverse lookup in the API:
- as a **Scope** variant (ReverseFile/ReverseDir), vs
- as **filters** that compile into joins/WHERE constraints.

You asked whether “scope and filters can be unified” (i.e., even “scopes are filters”).
With SQLite, that’s absolutely possible: “scope” becomes a convenience macro that compiles into the same join/where pipeline.

We also discussed the key trade:
- Scope provides clarity and avoids certain ambiguous combinations by construction.
- Filters provide maximal composability, but require careful validation/diagnostics to avoid confusing “empty results”.

**Decision recorded from the interaction**:
- You indicated you were convinced by **Option B** in the reverse-lookup modeling discussion: treat reverse lookup as **constraints/filters compiled into SQL joins**, rather than forcing a distinct “mode” API surface.

**Implication**:
- The compiler must add joins implicitly when reverse constraints are present (e.g., `related_files` join).
- We should provide good diagnostics for mis-specified or contradictory constraints.

**Debate references**:
- Q7 / Round 7 (scope modeling)
- Q11 / Round 9 (QueryDocs request/response)
- SQLite round (Round 13)

### 7) Error/diagnostics philosophy (in progress)

We discussed that “contradictory constraints” returning empty results is technically acceptable, but often poor UX.
We started framing a policy question:
- should contradictory/unsupported query combinations produce:
  - an error, or
  - empty results + diagnostics, or
  - empty results silently.

This is **not fully decided** yet (unless explicitly recorded later); it remains an open policy choice.

## Current API sketch (pseudocode)

This is a **working sketch** (not final naming). It reflects the decisions above: **extend `workspace`**, require a **fully-specified context** for the core object, but provide a **best-effort builder** for CLI.

```go
package workspace

// Workspace is the new centralized “repository API” entry point.
// Invariants: ctx.Root, ctx.ConfigDir, ctx.RepoRoot are non-empty.
type Workspace struct {
    ctx WorkspaceContext

    // Optional caches (design TBD)
    // ticketsCache []TicketHandle
    // docsCache    ...
}

// Fully specified, test-friendly context.
type WorkspaceContext struct {
    Root      string // docs root (absolute)
    ConfigDir string // directory containing .ttmp.yaml (absolute)
    RepoRoot  string // repository root (absolute)

    // Best-effort, may be nil if config missing/malformed (policy TBD)
    Config *WorkspaceConfig

    // Optional: plumbed “extra context”
    // VocabularyPath string // but vocabulary/validation should stay separate (Q8)
}

// Core constructor: requires fully specified context.
func NewWorkspaceFromContext(ctx WorkspaceContext) (*Workspace, error)

// CLI-friendly builder: resolves Root/ConfigDir/RepoRoot best-effort, then calls NewWorkspaceFromContext.
type DiscoverOptions struct {
    RootOverride string // e.g. from --root (can be relative)
    // Future: ConfigOverride, CWD override for tests, etc.
}
func DiscoverWorkspace(ctx context.Context, opts DiscoverOptions) (*Workspace, error)

// Resolver factory (uses ctx anchors)
func (w *Workspace) Resolver(docPath string) *paths.Resolver

// Ticket discovery / lookup
func (w *Workspace) QueryTickets(ctx context.Context, q TicketQuery) ([]TicketHandle, error)
func (w *Workspace) TicketByID(ctx context.Context, id string) (TicketHandle, error)

// Document lookup
func (w *Workspace) QueryDocs(ctx context.Context, q DocQuery) (DocQueryResult, error)
```

Policy still TBD (see debate rounds):
- **Broken states**: should `TicketHandle` / `DocHandle` carry parse errors, missing index, etc.? (Q2 / Round 11)
- **Filters + enumeration**: exact vs substring, skip rules, index.md handling, invalid-frontmatter handling (Q3 / Round 12)
- **QueryDocs design**: request vs options vs minimal signature (Q11 / Round 9)

### QueryDocs request/response sketch (pseudocode)

```go
type DocQuery struct {
    Scope   Scope
    Filters DocFilters
    Options DocQueryOptions
}

type DocQueryResult struct {
    Docs        []DocHandle
    Diagnostics []Diagnostic // optional: only if requested / produced
}

// Mirrors documents.WalkDocuments contract.
type DocHandle struct {
    Path    string
    Doc     *models.Document // nil if ReadErr != nil
    Body    string           // optionally populated
    ReadErr error
}

type DocQueryOptions struct {
    IncludeBody    bool
    IncludeErrors  bool // include DocHandles even when ReadErr != nil
    OrderBy        OrderBy
    Reverse        bool
    // TODO: enumeration policy hooks (skip rules, index.md inclusion, ignore patterns)
}
```

## Open design decisions (next)

- **Parking lot (asked, not yet decided)**:
  - **Q4**: `DiscoverWorkspace(...)` behavior when `.ttmp.yaml` is not found and/or repo-root detection fails:
    - strict error vs fallback (derive from `Root`) vs mixed + warnings.

- **Context contents**: what fields are required vs optional (`Root`, `ConfigPath`, `ConfigDir`, `RepoRoot`, loaded config, resolver factory, caches).
- **Caching policy**: ticket discovery caching scope + invalidation.
- **Core API entry points**: final `QueryDocs` signature, plus `TicketByID`, `QueryTickets`, etc.
- **Canonical enumeration policy**: ticket filter semantics, skip rules, index.md inclusion, invalid-frontmatter handling.


