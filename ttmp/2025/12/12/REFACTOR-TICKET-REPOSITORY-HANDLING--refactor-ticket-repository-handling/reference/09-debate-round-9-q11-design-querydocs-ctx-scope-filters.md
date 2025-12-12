---
Title: 'Debate Round 9 — Q11: Design QueryDocs(ctx, scope, filters...)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/documents/walk.go
      Note: WalkDocuments provides DocHandle contract (path
    - Path: internal/paths/resolver.go
      Note: Path normalization for reverse lookup (MatchPaths
    - Path: internal/workspace/config.go
      Note: Context resolution (ResolveRoot
    - Path: pkg/commands/doctor.go
      Note: Error handling (emit diagnostic
    - Path: pkg/commands/list_docs.go
      Note: Filter patterns duplicated
    - Path: pkg/commands/list_tickets.go
      Note: Ordering by LastUpdated (newest first)
    - Path: pkg/commands/search.go
      Note: Filter patterns duplicated
    - Path: pkg/models/document.go
      Note: Document model (no vocabulary awareness
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T16:00:00-05:00
---


# Debate Round 9 — Q11: Design QueryDocs(ctx, scope, filters...)

## Goal

Debate **Question 11** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> How should `QueryDocs(ctx, scope, filters...)` be designed?

**Prompt**: "What is the smallest, clean, and expressive design for `QueryDocs` that covers repo/ticket/doc scopes and supports reverse lookup and future extensions without turning into a grab-bag?"

**Must decide**:
- **`Scope` model** (repo, ticket-by-id, doc-by-path; file/dir reverse lookup as scope vs helper)
- **Filter model** (struct vs options; exact vs fuzzy; case handling)
- **Return types** (paths vs parsed docs vs handles including parse errors + bodies)
- **Parse/error behavior** (skip invalid docs by default vs return-with-error handle; how caller opts in/out)
- **Determinism** (stable ordering guarantees)
- **Context plumbing** (what `QueryDocs` needs from resolved root/configDir/repoRoot/path resolver)

**Acceptance criteria**:
- Proposed signatures for:
  - `QueryDocs`
  - core types: `Scope`, `Filters`, `DocHandle` (and optionally `TicketHandle`)
- 2 concrete call-site sketches from `pkg/commands/*`:
  - `search` (repo scan + reverse lookup)
  - `relate` or `doctor` (ticket scope / doc-only)
- A migration plan at the API level ("how do we adopt this without rewriting everything at once?")

## Context

This round is about **API design + semantics**, not performance/security/backwards compatibility.

Inputs from previous rounds:
- **Q6**: Frontmatter is authoritative for ticket identity; broken states must be represented.
- **Q7**: Scope should be explicit (enum vs separate methods vs optional fields—still debated).
- **Q8**: Repository should be vocabulary-agnostic; validation belongs in separate layer.

Current state:
- Commands use ad-hoc filtering (string comparisons, case-insensitive matching).
- Error handling is inconsistent (some skip invalid docs, others emit diagnostics).
- Ordering is ad-hoc (some commands sort, others don't).
- Context (root, configDir, repoRoot) is resolved per-command, not centralized.

## Pre-Debate Research

### Evidence A — `WalkDocuments` already provides a DocHandle contract

**Location**: `internal/documents/walk.go:11-13, 54-55`

**Findings**:
```go
type WalkDocumentFunc func(path string, doc *models.Document, body string, readErr error) error

// In WalkDocuments:
doc, body, readErr := ReadDocumentWithFrontmatter(path)
return fn(path, doc, body, readErr)
```

- Explicit contract: `doc` and `body` are nil when `readErr` is non-nil.
- Callback receives all three (path, doc, body, error) together.
- This is essentially the `DocHandle` shape we want `QueryDocs` to standardize.

**Implication**: `QueryDocs` should return `[]DocHandle` where `DocHandle` includes path, doc, body, and readErr.

### Evidence B — Commands skip invalid docs inconsistently

**Location**: `pkg/commands/search.go:288-291`, `pkg/commands/list_docs.go:188-192`

**Findings**:
- `search`: On parse error, silently skips (`return nil`).
- `list_docs`: On parse error, emits `ListingSkip` taxonomy in glaze mode, then continues.
- `doctor`: On parse error, emits diagnostic row and continues (doesn't skip).

**Implication**: Error handling policy is inconsistent. `QueryDocs` needs explicit policy (skip vs return-with-error).

### Evidence C — Filter patterns are duplicated across commands

**Location**: `pkg/commands/search.go:293-319`, `pkg/commands/list_docs.go:194-221`

**Findings**:
- Both implement identical filter logic:
  - `Ticket`: exact match (`doc.Ticket == settings.Ticket`)
  - `Status`: exact match (`doc.Status == settings.Status`)
  - `DocType`: exact match (`doc.DocType == settings.DocType`)
  - `Topics`: case-insensitive, any-match (`strings.EqualFold` + nested loops)
- Filter logic is duplicated verbatim.

**Implication**: Filters should be a shared type with consistent matching semantics.

### Evidence D — Reverse lookup uses inconsistent matching

**Location**: `pkg/commands/search.go:342-384`

**Findings**:
- `--file` reverse lookup:
  - Uses `paths.Resolver.Normalize()` and `paths.MatchPaths()` in one code path (correct).
  - Uses raw `strings.Contains()` in another code path (inconsistent).
- `--dir` reverse lookup:
  - Uses `paths.DirectoryMatch()` in one code path.
  - Uses raw `strings.HasPrefix()` in another code path.

**Implication**: Reverse lookup should use `paths.Resolver` consistently, not ad-hoc string operations.

### Evidence E — Ordering is ad-hoc and inconsistent

**Location**: `pkg/commands/list_tickets.go:197-302`, `pkg/commands/list_docs.go:402-422`

**Findings**:
- `list_tickets`: Sorts by `LastUpdated` (newest first) when `--sort` is used.
- `list_docs`: Groups by ticket, sorts tickets by `LastUpdated`, sorts docs within ticket by `LastUpdated`.
- `search`: No explicit sorting (relies on `filepath.Walk` order, which is filesystem-dependent).

**Implication**: `QueryDocs` should provide deterministic ordering (by default: path lexicographic, optionally by LastUpdated).

### Evidence F — Context resolution is duplicated per-command

**Location**: `pkg/commands/search.go:221-235`, `pkg/commands/meta_update.go:163-177`

**Findings**:
- Every command calls:
  - `workspace.ResolveRoot()` to get docs root.
  - `workspace.FindTTMPConfigPath()` to get config path.
  - `workspace.ResolveVocabularyPath()` (sometimes) to get vocabulary path.
- This resolution happens per-command invocation, not centralized.

**Implication**: `QueryDocs` should accept a `Repository` or `RepoContext` that holds resolved paths, not re-resolve them.

### Evidence G — Date filtering is implemented per-command

**Location**: `pkg/commands/search.go:400-420`

**Findings**:
- `search` implements date filtering:
  - `--since`, `--until`: filter by `LastUpdated`.
  - `--created-since`: filter by file modification time.
- Uses `time.Parse` with multiple formats.
- Other commands don't support date filtering.

**Implication**: Date filtering should be part of `Filters` type, not command-specific logic.

### Evidence H — External source filtering is ad-hoc string matching

**Location**: `pkg/commands/search.go:386-398`

**Findings**:
- `--external-source`: Uses `strings.Contains()` for bidirectional substring matching.
- No normalization or URL parsing.

**Implication**: External source filtering should be part of `Filters` type with consistent matching semantics.

## Opening Statements

### Mara (Staff Engineer) — "Structured request/response, composable and extensible"

**Position**: `QueryDocs` should use a structured request/response pattern to avoid parameter explosion and enable future extensions.

**Proposed API**:
```go
type QueryRequest struct {
    Scope   Scope
    Filters Filters
    Options QueryOptions
}

type QueryOptions struct {
    IncludeBody      bool // Include markdown body in DocHandle
    IncludeErrors    bool // Return handles for parse errors (vs skip)
    OrderBy          OrderBy // Path, LastUpdated, etc.
    Reverse          bool // Reverse order
    IgnorePatterns   []string // Additional ignore patterns (merged with defaults)
}

type QueryResult struct {
    Handles    []DocHandle
    Diagnostics []Diagnostic // Skipped files, parse errors (if IncludeErrors=false)
}

type DocHandle struct {
    Path    string
    Doc     *models.Document // nil if ReadErr != nil
    Body    string           // Empty if not requested or ReadErr != nil
    ReadErr error            // nil if parse succeeded
}

func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) (QueryResult, error)
```

**Rationale**:
- Single method signature covers all use cases.
- `QueryRequest` struct allows future extensions without breaking changes.
- `IncludeErrors` policy is explicit (caller chooses skip vs return-with-error).
- `OrderBy` provides deterministic ordering.
- `Diagnostics` captures skipped files when `IncludeErrors=false`.

**Example usage**:
```go
// Repo-wide search with filters
result, _ := repo.QueryDocs(ctx, QueryRequest{
    Scope: Scope{Type: ScopeRepo},
    Filters: Filters{
        Topics: []string{"api", "backend"},
        DocType: "design-doc",
    },
    Options: QueryOptions{
        IncludeBody: true,
        OrderBy: OrderByLastUpdated,
        Reverse: true,
    },
})

// Reverse lookup
result, _ := repo.QueryDocs(ctx, QueryRequest{
    Scope: Scope{Type: ScopeReverseFile, FilePath: "pkg/commands/add.go"},
    Options: QueryOptions{IncludeErrors: false},
})
```

**Call-site sketch (`search`)**:
```go
req := QueryRequest{
    Scope: resolveScope(settings), // --ticket, --file, --dir, or repo-wide
    Filters: Filters{
        Topics: settings.Topics,
        DocType: settings.DocType,
        Status: settings.Status,
        Ticket: settings.Ticket,
        ExternalSource: settings.ExternalSource,
        Since: parseDate(settings.Since),
        Until: parseDate(settings.Until),
    },
    Options: QueryOptions{
        IncludeBody: true, // For full-text search
        OrderBy: OrderByPath, // Deterministic but not sorted
    },
}
result, _ := repo.QueryDocs(ctx, req)
for _, handle := range result.Handles {
    if handle.ReadErr != nil {
        continue // Skip parse errors (IncludeErrors=false)
    }
    // Full-text search in handle.Body
    if matchesQuery(handle.Body, settings.Query) {
        emitResult(handle)
    }
}
```

**Trade-offs**:
1. **Verbosity**: More verbose than positional args, but clearer and more extensible.
2. **Zero values**: Empty `QueryRequest` must have sensible defaults (repo-wide, no filters, skip errors).

### Jon (Senior Engineer) — "Minimal signature, explicit DocHandle contract"

**Position**: Keep `QueryDocs` signature minimal; use `DocHandle` to represent all document states (valid, invalid, parse error).

**Proposed API**:
```go
type Scope struct {
    Type     ScopeType
    TicketID string // If Type == ScopeTicket
    DocPath  string // If Type == ScopeDoc
    FilePath string // If Type == ScopeReverseFile
    DirPath  string // If Type == ScopeReverseDir
}

type Filters struct {
    Ticket         string
    DocType        string
    Status         string
    Topics         []string // Any-match, case-insensitive
    ExternalSource string   // Substring match
    Since          time.Time // Filter by LastUpdated >= Since
    Until          time.Time // Filter by LastUpdated <= Until
}

type DocHandle struct {
    Path    string
    Doc     *models.Document // nil if ReadErr != nil
    Body    string           // Always empty (caller can read if needed)
    ReadErr error            // nil if parse succeeded
}

func (r *Repository) QueryDocs(ctx context.Context, scope Scope, filters Filters) ([]DocHandle, error)
```

**Rationale**:
- Minimal signature: `(ctx, scope, filters)` covers all cases.
- `DocHandle` always includes `ReadErr`; caller decides what to do with errors.
- No `IncludeErrors` toggle—always return handles (caller filters).
- No `IncludeBody` toggle—body is always empty (caller reads if needed for performance).
- Simpler than structured request/response.

**Example usage**:
```go
// Repo-wide search
handles, _ := repo.QueryDocs(ctx, Scope{Type: ScopeRepo}, Filters{
    Topics: []string{"api"},
    DocType: "design-doc",
})

// Filter out parse errors
validHandles := []DocHandle{}
for _, h := range handles {
    if h.ReadErr != nil {
        continue
    }
    validHandles = append(validHandles, h)
}
```

**Call-site sketch (`search`)**:
```go
scope := resolveScope(settings) // --ticket, --file, --dir, or repo-wide
filters := Filters{
    Ticket: settings.Ticket,
    Topics: settings.Topics,
    DocType: settings.DocType,
    Status: settings.Status,
    ExternalSource: settings.ExternalSource,
    Since: parseDate(settings.Since),
    Until: parseDate(settings.Until),
}
handles, _ := repo.QueryDocs(ctx, scope, filters)

// Full-text search (read body for matching handles)
for _, handle := range handles {
    if handle.ReadErr != nil {
        continue // Skip parse errors
    }
    body, _ := os.ReadFile(handle.Path) // Caller reads body
    if matchesQuery(string(body), settings.Query) {
        emitResult(handle, body)
    }
}
```

**Trade-offs**:
1. **Body loading**: Caller must read body separately (performance cost), but avoids loading bodies for filtered-out docs.
2. **Error handling**: Caller must filter out `ReadErr != nil` handles (more code), but explicit and flexible.

### `pkg/commands/*` (as a bloc) — "Give us optional fields, not enums"

**Position**: Use optional fields in `QueryOptions` struct, not enums. Commands can pass flags directly.

**Proposed API**:
```go
type QueryOptions struct {
    // Scope (mutually exclusive)
    TicketID    string // Narrow to ticket
    DocPath     string // Single doc (overrides TicketID)
    ReverseFile string // Find docs referencing file
    ReverseDir  string // Find docs in/referencing dir
    
    // Filters
    Filters Filters
    
    // Options
    IncludeBody   bool // Include body in DocHandle
    IncludeErrors bool // Return handles for parse errors
    OrderBy       string // "path", "lastUpdated", etc.
    Reverse       bool // Reverse order
}

type DocHandle struct {
    Path    string
    Doc     *models.Document
    Body    string
    ReadErr error
}

func (r *Repository) QueryDocs(ctx context.Context, opts QueryOptions) ([]DocHandle, error)
```

**Rationale**:
- Single method with optional fields matches command flag patterns.
- Commands can construct `QueryOptions` directly from flags.
- No enum construction required.
- Validation can enforce mutual exclusivity (e.g., `DocPath` and `TicketID` can't both be set).

**Example usage**:
```go
// Repo-wide
handles, _ := repo.QueryDocs(ctx, QueryOptions{
    Filters: Filters{Topics: []string{"api"}},
})

// Ticket + doc-type
handles, _ := repo.QueryDocs(ctx, QueryOptions{
    TicketID: "MEN-3475",
    Filters: Filters{DocType: "design-doc"},
})

// Reverse lookup
handles, _ := repo.QueryDocs(ctx, QueryOptions{
    ReverseFile: "pkg/commands/add.go",
})
```

**Call-site sketch (`doctor`)**:
```go
opts := QueryOptions{
    TicketID: settings.Ticket, // Empty if --all
    IncludeErrors: true, // Doctor wants to see parse errors
    OrderBy: "path", // Deterministic ordering
}
handles, _ := repo.QueryDocs(ctx, opts)

for _, handle := range handles {
    if handle.ReadErr != nil {
        // Emit diagnostic for parse error
        emitDiagnostic(handle.Path, handle.ReadErr)
        continue
    }
    // Validate handle.Doc
    validateDoc(handle.Doc)
}
```

**Trade-offs**:
1. **Mutual exclusivity**: Need validation to ensure scope fields aren't conflicting (e.g., `DocPath` and `TicketID`).
2. **Default behavior**: Empty `QueryOptions` must default to repo-wide scope (clear documentation needed).

### `documents.WalkDocuments` — "I'm the primitive; build QueryDocs on top of me"

**Position**: `QueryDocs` should use `WalkDocuments` internally, not re-implement traversal.

**Proposed API**:
```go
// Repository uses WalkDocuments internally
func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) ([]DocHandle, error) {
    handles := []DocHandle{}
    
    root := r.resolveScopeRoot(req.Scope)
    err := documents.WalkDocuments(root, func(path string, doc *models.Document, body string, readErr error) error {
        handle := DocHandle{
            Path: path,
            Doc: doc,
            Body: body, // If req.Options.IncludeBody
            ReadErr: readErr,
        }
        
        // Apply filters
        if !matchesFilters(handle, req.Filters) {
            return nil // Skip
        }
        
        // Apply error policy
        if readErr != nil && !req.Options.IncludeErrors {
            return nil // Skip errors
        }
        
        handles = append(handles, handle)
        return nil
    }, r.walkOptions(req.Options)...)
    
    // Apply ordering
    sortHandles(handles, req.Options.OrderBy, req.Options.Reverse)
    
    return handles, err
}
```

**Rationale**:
- `WalkDocuments` is the traversal primitive; `QueryDocs` composes it with filtering/ordering.
- No duplication of traversal logic.
- `WalkDocuments` contract (path, doc, body, readErr) maps directly to `DocHandle`.

**Suggestion**: Keep `WalkDocuments` as-is; `QueryDocs` is a convenience wrapper that adds filtering/ordering.

### `paths.Resolver` — "Normalize everything for reverse lookup"

**Position**: Reverse lookup scope must use `paths.Resolver` for consistent matching.

**Proposed API**:
```go
type Scope struct {
    Type     ScopeType
    FilePath string // If Type == ScopeReverseFile (will be normalized)
    DirPath  string // If Type == ScopeReverseDir (will be normalized)
}

// In QueryDocs implementation:
if scope.Type == ScopeReverseFile {
    resolver := paths.NewResolver(paths.ResolverOptions{
        DocsRoot: r.Root,
        ConfigDir: r.ConfigDir,
    })
    queryNorm := resolver.Normalize(scope.FilePath)
    
    // Match against normalized RelatedFiles paths
    for _, handle := range handles {
        for _, rf := range handle.Doc.RelatedFiles {
            rfNorm := resolver.Normalize(rf.Path)
            if paths.MatchPaths(queryNorm, rfNorm) {
                // Match
            }
        }
    }
}
```

**Rationale**:
- Current code has bugs due to inconsistent matching (sometimes `paths.MatchPaths`, sometimes `strings.Contains`).
- Reverse lookup must normalize paths through `paths.Resolver` for consistency.
- `QueryDocs` should accept a `Repository` that holds `paths.Resolver` (or creates it per-query).

**Suggestion**: Add `NormalizeQueryPath(path string) NormalizedPath` helper that all reverse lookup scopes use.

### `workspace.ResolveRoot` — "Repository holds resolved context"

**Position**: `QueryDocs` must accept a `Repository` that holds resolved paths, not re-resolve them.

**Proposed API**:
```go
type Repository struct {
    Root     string // Resolved docs root
    ConfigDir string // Resolved config directory
    RepoRoot  string // Resolved repo root (if any)
}

func NewRepository(ctx context.Context) (*Repository, error) {
    root, _ := workspace.ResolveRoot()
    configPath, _ := workspace.FindTTMPConfigPath()
    configDir := filepath.Dir(configPath)
    repoRoot, _ := workspace.FindRepoRoot(root)
    
    return &Repository{
        Root: root,
        ConfigDir: configDir,
        RepoRoot: repoRoot,
    }, nil
}

func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) (QueryResult, error) {
    // Use r.Root, r.ConfigDir, r.RepoRoot for path resolution
}
```

**Rationale**:
- Context resolution happens once at `Repository` creation, not per-query.
- `QueryDocs` uses resolved paths from `Repository` (no re-resolution).
- Commands create `Repository` once, reuse for multiple queries.

**Suggestion**: `Repository` should be the primary API entry point; `QueryDocs` is a method on it.

## Rebuttals

### Mara responds to Jon

**Jon's minimal signature is too rigid**: With only `(scope, filters)`, we can't add options (ordering, error policy) without breaking changes. Structured request allows extensions.

**Jon's body loading is inefficient**: Caller reading body separately means we can't optimize (e.g., skip body parsing for filtered-out docs). `IncludeBody` toggle allows optimization.

**Counter-proposal**: Use structured request, but make `QueryOptions` fields optional with sensible defaults.

### Jon responds to Mara

**Mara's structured request is over-engineered**: Most commands only need scope + filters. Why force them to construct `QueryRequest` when `(scope, filters)` is clearer?

**Mara's IncludeBody toggle adds complexity**: If body is needed, caller can read it. If not needed, why parse it? Always-empty body is simpler.

**Compromise**: Use `(scope, filters, opts...)` variadic options pattern:
```go
func (r *Repository) QueryDocs(ctx context.Context, scope Scope, filters Filters, opts ...QueryOption) ([]DocHandle, error)
```

### `pkg/commands/*` responds to Mara

**Mara's structured request is verbose**: Commands have to construct `QueryRequest` struct. Our optional fields approach matches command flags directly.

**Mara's QueryResult wrapper is unnecessary**: Just return `[]DocHandle`; diagnostics can be separate method if needed.

**Preference**: Our optional fields approach OR Jon's minimal signature. Both are simpler than structured request.

### `pkg/commands/*` responds to Jon

**Jon's body loading is acceptable**: Commands can read body when needed. Performance is acceptable for current use cases.

**Jon's error filtering is acceptable**: Commands already filter errors (doctor emits diagnostics, search skips). Explicit filtering is fine.

**Preference**: Either Jon's approach OR our optional fields. Both are better than Mara's structured request.

### `documents.WalkDocuments` responds to all

**I agree with composition**: `QueryDocs` should use me internally, not re-implement traversal. But keep my contract (path, doc, body, readErr) as-is.

**Suggestion**: `QueryDocs` composes me with filtering/ordering; no changes to my API.

### `paths.Resolver` responds to all

**Reverse lookup must normalize**: Whatever API shape you choose, reverse lookup must use `paths.Resolver` consistently. Current code has bugs due to inconsistent matching.

**Suggestion**: Add `NormalizeQueryPath()` helper that all reverse lookup scopes use.

### `workspace.ResolveRoot` responds to all

**Repository must hold context**: `QueryDocs` must accept a `Repository` that holds resolved paths. Don't re-resolve on every query.

**Suggestion**: `NewRepository(ctx)` resolves once; `QueryDocs` uses resolved paths.

## Moderator Summary

### Key Arguments

1. **API shape** (disagreement):
   - **Mara**: Structured request/response (`QueryDocs(ctx, QueryRequest) (QueryResult, error)`).
   - **Jon**: Minimal signature (`QueryDocs(ctx, Scope, Filters) ([]DocHandle, error)`).
   - **`pkg/commands/*`**: Optional fields (`QueryDocs(ctx, QueryOptions) ([]DocHandle, error)`).

2. **DocHandle contract** (agreement): All agree that `DocHandle` should include path, doc, body, and readErr (matching `WalkDocuments` contract).

3. **Error handling** (disagreement):
   - **Mara**: Explicit `IncludeErrors` toggle (skip vs return-with-error).
   - **Jon**: Always return handles; caller filters errors.
   - **`pkg/commands/*`**: Optional `IncludeErrors` field.

4. **Body loading** (disagreement):
   - **Mara**: `IncludeBody` toggle (optimize by skipping body parsing).
   - **Jon**: Always-empty body; caller reads if needed.
   - **`pkg/commands/*`**: Optional `IncludeBody` field.

5. **Ordering** (agreement): All agree that deterministic ordering is needed (by default: path lexicographic, optionally by LastUpdated).

6. **Context resolution** (agreement): All agree that `QueryDocs` should accept a `Repository` that holds resolved paths, not re-resolve them.

7. **Reverse lookup** (agreement): All agree that reverse lookup must use `paths.Resolver` for consistent matching.

8. **Composition** (agreement): All agree that `QueryDocs` should use `WalkDocuments` internally, not re-implement traversal.

### Tensions

1. **Structured vs minimal**: Mara's structured request is extensible but verbose. Jon's minimal signature is clear but rigid. `pkg/commands/*`'s optional fields are ergonomic but require validation.

2. **Error policy**: Should errors be skipped by default (Mara's `IncludeErrors=false`) or always returned (Jon's explicit filtering)?

3. **Body loading**: Should body be included by default (performance cost) or always-empty (caller reads separately)?

4. **Scope model**: Should scope be enum (Mara, Jon) or optional fields (`pkg/commands/*`)? (This was debated in Q7.)

### Interesting Ideas

1. **Variadic options pattern**: Jon's suggestion to use `opts ...QueryOption` provides extensibility without structured request:
   ```go
   func (r *Repository) QueryDocs(ctx context.Context, scope Scope, filters Filters, opts ...QueryOption) ([]DocHandle, error)
   ```

2. **QueryResult wrapper**: Mara's `QueryResult` with `Diagnostics` captures skipped files when `IncludeErrors=false`, but may be unnecessary if errors are always returned.

3. **NormalizeQueryPath helper**: `paths.Resolver`'s suggestion to add `NormalizeQueryPath()` ensures consistent reverse lookup matching.

### Open Questions

1. **Default behavior**: What should empty `QueryRequest`/`QueryOptions` default to?
   - Scope: repo-wide?
   - Filters: none?
   - IncludeErrors: false (skip) or true (return)?
   - IncludeBody: false (empty) or true (parse)?

2. **Ordering API**: How should ordering be specified?
   - Enum (`OrderByPath`, `OrderByLastUpdated`)?
   - String (`"path"`, `"lastUpdated"`)?
   - Separate method (`QueryDocs(...).OrderBy(...)`)?

3. **Migration strategy**: How do we adopt `QueryDocs` without rewriting all commands at once?
   - Option 1: Implement `QueryDocs`, migrate commands one-by-one.
   - Option 2: Implement `QueryDocs` alongside existing code, deprecate old patterns.
   - Option 3: Implement `QueryDocs` as wrapper around existing code, gradually refactor.

4. **Filter semantics**: Should filters be:
   - Exact match (current: `doc.Ticket == filter.Ticket`)?
   - Fuzzy match (substring, case-insensitive)?
   - Configurable (exact vs fuzzy per filter)?

5. **Reverse lookup scope**: Should reverse lookup be:
   - Scope type (`ScopeReverseFile`, `ScopeReverseDir`)?
   - Filter (`Filters{RelatedFile: "...", RelatedDir: "..."}`)?
   - Separate method (`FindDocsReferencingFile(...)`)?

### Next Steps

1. **Prototype the API** with 2-3 command call sites (`search`, `doctor`, `relate`).
2. **Decide on API shape**: Structured request vs minimal signature vs optional fields.
3. **Design DocHandle contract**: Finalize fields (path, doc, body, readErr).
4. **Design error policy**: Skip vs return-with-error (default behavior).
5. **Design ordering API**: Enum vs string vs separate method.
6. **Integrate paths.Resolver**: Ensure reverse lookup uses normalization consistently.
7. **Design Repository type**: How to hold resolved context (root, configDir, repoRoot).
8. **Create migration plan**: How to adopt `QueryDocs` incrementally.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/06-debate-round-6-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/07-debate-round-7-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/08-debate-round-8-q8-how-do-we-keep-vocabulary-config-concerns-from-leaking-everywhere.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`

