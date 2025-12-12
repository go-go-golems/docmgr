---
Title: 'Debate Round 4 — Q11: Design QueryDocs(ctx, scope, filters...)'
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
LastUpdated: 2025-12-12T15:11:16.378721158-05:00
---

# Debate Round 4 — Q11: Design QueryDocs(ctx, scope, filters...)

## Goal

Debate **Question 11** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> How should `QueryDocs(ctx, scope, filters...)` be designed?

We want a single API entry point that:

- covers **repo / ticket / doc** scope (per the Q7 decision)
- supports reverse lookup (file/dir) in a consistent way
- provides a crisp parse/error contract (no silent “skip unless you know”)
- yields deterministic output that commands can format consistently

## Context

This debate is about API design + semantics only (not performance/security/backcompat).

Inputs:

- You chose **A)**: a single `QueryDocs(ctx, scope, filters...)` direction (recorded in Round 2).
- Q6 surfaced that “broken states” must be represented (missing index / invalid frontmatter ≠ not found).

## Pre-Debate Research (codebase evidence)

### Evidence A — We already have a neutral doc-walk “handle contract”

- `internal/documents/walk.go`:
  - `WalkDocuments(root, fn)` calls `ReadDocumentWithFrontmatter(path)` and invokes:
    - `fn(path, doc, body, readErr)`
  - It’s explicitly documented that `doc` and `body` are nil when `readErr` is non-nil.

This is essentially the “DocHandle” shape we want `QueryDocs` to standardize.

### Evidence B — Command-level doc enumeration differs in skip + error behavior

- `pkg/commands/list_docs.go`:
  - walks `settings.Root` with `filepath.Walk`
  - skips `index.md`
  - on parse error:
    - glaze mode emits `ListingSkip` taxonomy
    - then silently continues (no result/handle for that doc)

- `pkg/commands/search.go`:
  - walks `settings.Root` with `filepath.Walk`
  - skips `/_templates/` and `/_guidelines/` by string contains
  - on invalid frontmatter: silently skips the file
  - reverse lookup uses `paths.Resolver` to normalize and match RelatedFiles

### Evidence C — We already have a path normalization layer that wants to be reused

- `internal/paths/resolver.go`:
  - `Resolver.Normalize(raw)` returns a `NormalizedPath` with multiple comparable representations
  - `MatchPaths` and `DirectoryMatch` implement fuzzy-ish comparisons

If `QueryDocs` supports reverse lookup (file/dir), it should be built on this resolver consistently.

### Evidence D — Required field validation is separate from parsing

- `pkg/models/document.go`:
  - `Document.Validate()` enforces required fields (Title/Ticket/DocType)

Implication: parse success does not imply schema validity; `QueryDocs` should not conflate parsing with “valid doc” unless explicitly asked.

## Proposed API design candidates (for the debate)

These are **strawman** shapes to debate (not implementation yet).

### Strawman 1 — Single method with structured request + structured response

- `QueryDocs(ctx, QueryRequest) (QueryResult, error)`

### Strawman 2 — Single method with `Scope` + `Filters` + options

- `QueryDocs(ctx, Scope, Filters, ...QueryOption) ([]DocHandle, error)`

## Debate (Question 11)

### Opening Statements (Round 1)

#### Mara (Staff Engineer) — “Make QueryDocs request/response structured; keep it composable”

We’re going to evolve this API quickly. A function signature with 8 positional args will rot immediately. I want:

- `QueryDocs(ctx, QueryRequest) (QueryResult, error)`

Where `QueryRequest` includes:
- `Scope`
- `Filters`
- `Include` toggles (include body? include frontmatter-only? include parse errors?)
- `Match` options (for reverse lookup: file/dir match using `paths.Resolver`)
- `Ignore` policy (underscore dirs, `_guidelines`, `_templates`, `.docmgrignore` merged list)

And `QueryResult` includes:
- ordered `[]DocHandle`
- optional `[]Diagnostic` (for skipped docs / parse errors if not returned as handles)

This makes adoption consistent and keeps room for future switches.

#### Jon (Senior Engineer) — “Keep it minimal: Scope + Filters + a couple options”

I agree on one entry point, but I worry about `QueryRequest` turning into a kitchen sink. My preference:

- `QueryDocs(ctx, scope, filters, opts...) ([]DocHandle, error)`

With `DocHandle` always including:
- `Path`
- `Doc` (may be nil)
- `Body` (optional)
- `ReadErr` (optional)

Then policy layers (doctor) can decide what to do with invalid/empty docs. Reverse lookup can be a `Scope` variant if we really want it unified, but I’d keep it as a filter:

- `Filters{RelatedFile: "...", RelatedDir: "..."}`

#### `workspace.ResolveRoot` — “QueryDocs must accept a resolved context, not re-discover it”

The new API must not re-run root/config discovery on every call. We should have a repository object (or `RepoContext`) created once; `QueryDocs` is a method on that:

- `repo.QueryDocs(ctx, req)`

That also cleanly provides `Root`, `ConfigDir`, `RepoRoot` for path normalization.

#### `workspace.CollectTicketWorkspaces` — “Ticket scope needs a clear input: ticket id vs ticket dir”

For `ScopeTicket`, we need to decide:
- does scope accept ticket ID and the repo resolves to dir?
- or does it accept ticket directory directly?

I recommend ticket ID, because commands use `--ticket`. But that means `QueryDocs` must define what happens when:
- ticket missing
- ticket exists but index missing/invalid
- multiple dirs match ticket ID

#### `documents.WalkDocuments` — “DocHandle is already defined by my callback signature”

Don’t invent a new pattern; standardize on it:

- parse errors are not exceptional; they are part of the result space
- callers can choose to drop them, but the API should make that explicit

I’d like `QueryDocs` to expose an “include invalid” policy rather than silently skipping.

#### `paths.Resolver` — “Reverse lookup belongs in QueryDocs, but only if matching is standardized”

Search currently builds a doc-scoped resolver and uses `MatchPaths` and `DirectoryMatch`. If reverse lookup is in `QueryDocs`, `QueryDocs` should:
- construct resolvers consistently using the repo’s context
- define exactly what is matched (canonical vs representations vs suffixes)

Otherwise we’ll end up with “search matches differently from doctor/relate”.

#### `pkg/commands/*` (bloc) — “We need a default behavior that matches human expectations”

We need defaults that make CLI output predictable:

- deterministic ordering
- clear handling of invalid frontmatter:
  - either list it as a result with an attached error, or explicitly report “skipped due to parse error”
  - but never silently vanish without a trace (unless caller asks for silent)

Also, we need `QueryDocs` to support common call sites in 2–3 lines.

### Rebuttals (Round 2)

#### Mara → Jon

If we put reverse lookup into `Filters`, we’ll get confusion between “filter” and “scope”. Reverse lookup can be either, but it needs a single definition. I’m okay with a hybrid: `ScopeRepo/Ticket/Doc` plus a `Match` section for reverse lookup within those scopes.

#### Jon → Mara

Agreed, but keep it simple: one `Scope` plus optional `Match` is fine. Let’s avoid building an entire query language prematurely.

#### `documents.WalkDocuments` → everyone

Please make parse behavior explicit:
- `ParseModeSkipInvalid`
- `ParseModeIncludeInvalidAsHandles`

Otherwise `QueryDocs` will be impossible to use consistently across commands.

### Moderator Summary (current synthesis)

**Likely shape**
- `QueryDocs` should take a **structured request** (even if small today) to avoid positional-arg growth.
- `QueryDocs` should return **DocHandles** that can represent parse errors (or alternatively return diagnostics for skipped docs, but still be explicit).
- Reverse lookup should be expressed as a **first-class part of the request** (either scope variant or match criteria), and should reuse `paths.Resolver`.

**Key open decisions to finalize**
- Request/response types:
  - `QueryDocs(ctx, QueryRequest) (QueryResult, error)` vs `QueryDocs(ctx, scope, filters, opts...) ([]DocHandle, error)`
- Default parse policy:
  - include invalid docs as handles vs skip and emit diagnostics
- Where reverse lookup lives:
  - `ScopeReverseLookupFile/Dir` vs `Match{File,Dir}` criteria within a scope

**Proposed next step**
- Draft the concrete types and 2 call-site sketches (search + relate).

## Usage Examples

Next round / next doc could be a “call-site sketch” document:

- Rewrite `pkg/commands/search.go` at the level of pseudo-code to use `repo.QueryDocs(...)` with:
  - repo scope
  - ticket scope
  - reverse lookup mode
- Rewrite `pkg/commands/relate.go` for:
  - doc-only target
  - ticket-index target

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/04-debate-round-2-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`
