---
Title: 'Debate Round 7 — Q7: How should we model scope in lookups (repo vs ticket vs doc)?'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/commands/doctor.go
      Note: Explicit scope modes (--doc
    - Path: pkg/commands/list_docs.go
      Note: Repo-wide walk with ticket filtering
    - Path: pkg/commands/meta_update.go
      Note: Three distinct scopes (doc
    - Path: pkg/commands/relate.go
      Note: Scope determines search root for suggestions
    - Path: pkg/commands/search.go
      Note: Multiple scope modes (repo-wide
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T15:15:00-05:00
---


# Debate Round 7 — Q7: How should we model scope in lookups (repo vs ticket vs doc)?

## Goal

Debate **Question 7** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> How should we model "scope" in lookups (repo-wide vs ticket-only vs doc-only)?

**Prompt**: "What scopes should be built into the API, and how should a caller select them?"

**Acceptance criteria**:
- A small scoping model (enums/options) that avoids duplicating methods
- Example: one API call each for
  - all docs
  - docs within a ticket
  - docs related to a file/directory (reverse lookup)

## Context

This round is about **semantics + ergonomics + structure**, not performance/security/backwards compatibility.

Current commands already encode scope implicitly:
- `docmgr search` can act repo-wide, ticket-only (`--ticket`), or reverse-lookup-by-file/dir (`--file`, `--dir`).
- `docmgr relate` targets either a specific doc (`--doc`) or a ticket index (`--ticket`).
- `docmgr meta update` targets either a single doc (`--doc`), a ticket (`--ticket`), or ticket+doc-type (`--ticket --doc-type`).
- `docmgr doctor` has explicit modes (`--doc` single file, `--ticket`, `--all`).

## Pre-Debate Research

### Evidence A — `search` encodes multiple scopes in one command

**Location**: `pkg/commands/search.go:32-49, 271-556`

**Findings**:
- `search` uses flags to determine scope:
  - `--ticket`: filters to docs with matching ticket ID (but still walks repo-wide, then filters)
  - `--file`: reverse lookup (finds docs that reference a file path)
  - `--dir`: reverse lookup (finds docs in a directory or referencing files in it)
  - No `--ticket` flag: walks entire `settings.Root`
- Implementation pattern:
  ```go
  if settings.Ticket != "" {
      ticketDir, err = findTicketDirectory(settings.Root, settings.Ticket)
      // ... then walk ticketDir instead of settings.Root
  } else {
      ticketDir = settings.Root
  }
  err = filepath.Walk(ticketDir, ...)
  ```
- Reverse lookup (`--file`, `--dir`) still walks repo-wide but filters by `RelatedFiles` matching.

**Implication**: Scope is determined by conditional logic, not a first-class abstraction.

### Evidence B — `meta update` has three distinct scopes

**Location**: `pkg/commands/meta_update.go:179-200`

**Findings**:
- Scope determined by flag combination:
  - `--doc`: single file (`filesToUpdate = []string{settings.Doc}`)
  - `--ticket` only: ticket index (`filesToUpdate = []string{filepath.Join(ticketDir, "index.md")}`)
  - `--ticket --doc-type`: all docs of that type in ticket (`findMarkdownFiles(ticketDir, settings.DocType)`)
- Each scope requires different traversal logic.

**Implication**: Commands re-implement scope logic rather than delegating to a shared API.

### Evidence C — `relate` uses scope to determine search root for suggestions

**Location**: `pkg/commands/relate.go:213-221`

**Findings**:
- When `--suggest` is used, scope determines search root:
  ```go
  searchRoot := settings.Root
  if ticketDir != "" {
      searchRoot = ticketDir  // Narrow to ticket scope
  }
  ```
- Suggestion heuristics (git, ripgrep, RelatedFiles) operate within `searchRoot`.

**Implication**: Scope affects not just enumeration but also auxiliary operations (suggestions, validation).

### Evidence D — `doctor` has explicit scope modes

**Location**: `pkg/commands/doctor.go:268-285`

**Findings**:
- Three distinct modes:
  - `--doc`: single file validation (`validateSingleDoc`)
  - `--ticket`: ticket workspace validation (uses `CollectTicketWorkspaces` + filters)
  - `--all`: all tickets (uses `CollectTicketWorkspaces` without filter)
- Single-file mode bypasses ticket discovery entirely.

**Implication**: Scope determines which discovery primitives are used.

### Evidence E — Reverse lookup is ad-hoc string matching

**Location**: `pkg/commands/search.go:342-384, 1254-1288`

**Findings**:
- `--file` reverse lookup:
  - Walks repo-wide
  - For each doc, checks if any `RelatedFiles` path matches query (substring or contains)
  - Uses `paths.MatchPaths` for normalized comparison (in one code path) but also raw string contains (in another)
- `--dir` reverse lookup:
  - Checks if doc path is within directory (prefix match)
  - Also checks if any `RelatedFiles` path is within directory
- Inconsistent: some code paths use `paths.Resolver`, others use raw string operations.

**Implication**: Reverse lookup is not a first-class scope type; it's implemented as filtering.

### Evidence F — `list_docs` walks repo-wide but filters by ticket

**Location**: `pkg/commands/list_docs.go:172-244`

**Findings**:
- Always walks `settings.Root` repo-wide
- Filters by `--ticket` flag: `if settings.Ticket != "" && doc.Ticket != settings.Ticket { return nil }`
- Does NOT narrow traversal to ticket directory (less efficient but simpler).

**Implication**: Some commands optimize traversal (search narrows to ticketDir), others don't (list_docs walks everything).

## Opening Statements

### Mara (Staff Engineer) — "Unify semantics, reduce surprise"

**Position**: Scope should be a **first-class enum** passed to a single `QueryDocs` method. This eliminates conditional logic and makes scope explicit.

**Proposed API**:
```go
type Scope struct {
    Type ScopeType
    TicketID string  // If Type == ScopeTicket
    DocPath string   // If Type == ScopeDoc
    FilePath string  // If Type == ScopeReverseFile
    DirPath string   // If Type == ScopeReverseDir
}

type ScopeType int
const (
    ScopeRepo ScopeType = iota      // All docs in repo
    ScopeTicket                      // All docs in a ticket
    ScopeDoc                         // Single document
    ScopeReverseFile                 // Docs referencing a file
    ScopeReverseDir                  // Docs in/referencing a directory
)

func (r *Repository) QueryDocs(scope Scope, filters Filters) ([]DocHandle, error)
```

**Rationale**:
- Single method signature covers all use cases.
- Scope is explicit and type-safe.
- Commands can't accidentally mix scopes (e.g., `--ticket` + `--doc`).
- Reverse lookup becomes a first-class scope type, not ad-hoc filtering.

**Example usage**:
```go
// Repo-wide
docs, _ := repo.QueryDocs(Scope{Type: ScopeRepo}, Filters{Topic: "api"})

// Ticket scope
docs, _ := repo.QueryDocs(Scope{Type: ScopeTicket, TicketID: "MEN-3475"}, Filters{})

// Reverse lookup
docs, _ := repo.QueryDocs(Scope{Type: ScopeReverseFile, FilePath: "pkg/commands/add.go"}, Filters{})
```

### Jon (Senior Engineer) — "Small API surface, easy to adopt"

**Position**: Use **separate methods** for each scope type. Simpler to understand and adopt incrementally.

**Proposed API**:
```go
// Repo-wide
func (r *Repository) QueryAllDocs(filters Filters) ([]DocHandle, error)

// Ticket scope
func (r *Repository) QueryTicketDocs(ticketID string, filters Filters) ([]DocHandle, error)

// Single doc
func (r *Repository) GetDoc(path string) (DocHandle, error)

// Reverse lookup (helpers, not core scope)
func (r *Repository) FindDocsReferencingFile(filePath string, filters Filters) ([]DocHandle, error)
func (r *Repository) FindDocsInDirectory(dirPath string, filters Filters) ([]DocHandle, error)
```

**Rationale**:
- Each method has a clear, single purpose.
- Easier to document and test.
- Commands can adopt one method at a time.
- Reverse lookup is a helper, not a core scope (can be implemented via `QueryAllDocs` + filtering).

**Example usage**:
```go
// Repo-wide
docs, _ := repo.QueryAllDocs(Filters{Topic: "api"})

// Ticket scope
docs, _ := repo.QueryTicketDocs("MEN-3475", Filters{})

// Reverse lookup
docs, _ := repo.FindDocsReferencingFile("pkg/commands/add.go", Filters{})
```

### `pkg/commands/*` (as a bloc) — "I need an API that's easy to call"

**Position**: Give us **one method with optional scope parameters**. Most commands need flexibility to combine scopes (e.g., `--ticket` + `--doc-type`).

**Proposed API**:
```go
type QueryOptions struct {
    TicketID string      // Optional: narrow to ticket
    DocPath string       // Optional: single doc (overrides TicketID)
    ReverseFile string   // Optional: find docs referencing file
    ReverseDir string    // Optional: find docs in/referencing dir
    Filters Filters      // Metadata filters
}

func (r *Repository) QueryDocs(opts QueryOptions) ([]DocHandle, error)
```

**Rationale**:
- Single method covers all cases.
- Optional fields allow flexible combinations (`--ticket` + `--doc-type`).
- Commands can pass flags directly without constructing enums.
- Validation can enforce mutual exclusivity (e.g., `DocPath` and `TicketID` can't both be set).

**Example usage**:
```go
// Repo-wide
docs, _ := repo.QueryDocs(QueryOptions{Filters: Filters{Topic: "api"}})

// Ticket + doc-type filter
docs, _ := repo.QueryDocs(QueryOptions{
    TicketID: "MEN-3475",
    Filters: Filters{DocType: "design-doc"},
})

// Reverse lookup
docs, _ := repo.QueryDocs(QueryOptions{ReverseFile: "pkg/commands/add.go"})
```

### `documents.WalkDocuments` — "One traversal, one parse contract"

**Position**: Scope should determine **where I'm invoked**, not how I work. Keep traversal logic separate from scope selection.

**Proposed API**:
```go
// Scope determines root path, then I walk it
func (r *Repository) GetScopeRoot(scope Scope) (string, error)

// Then callers use me with that root
err := documents.WalkDocuments(root, func(path string, doc *Document, body string, err error) error {
    // Apply filters here
})
```

**Rationale**:
- I'm a traversal primitive, not a query engine.
- Scope is about "where to start walking", not "how to walk".
- Commands can compose scope resolution + traversal + filtering.

**Counter-proposal**: Repository API should return an iterator/channel, not a slice. Scope determines iterator source.

### `paths.Resolver` — "Normalize everything, compare apples-to-apples"

**Position**: Reverse lookup scope should use **my normalization** for matching. Don't implement ad-hoc string operations.

**Observation**:
- Current reverse lookup uses inconsistent matching (sometimes `paths.MatchPaths`, sometimes raw `strings.Contains`).
- Scope for reverse lookup should normalize the query path through me before matching.

**Suggestion**:
- Reverse lookup scope should accept a `NormalizedPath` (or resolve it internally).
- Matching should use `paths.MatchPaths` consistently.
- This ensures `pkg/commands/add.go` matches `./pkg/commands/add.go` and `backend/../pkg/commands/add.go`.

## Rebuttals

### Mara responds to Jon

**Jon's separate methods create duplication**: Each method needs to handle filters, error types, and result formatting. A single method with scope enum avoids this.

**Jon's reverse lookup as helpers is inconsistent**: If reverse lookup is "just filtering", why have separate methods? Either make it first-class (my approach) or keep it as filtering (no helpers).

**Counter-proposal**: Use a single method with scope enum, but provide convenience wrappers:
```go
func (r *Repository) QueryAllDocs(filters Filters) ([]DocHandle, error) {
    return r.QueryDocs(Scope{Type: ScopeRepo}, filters)
}
```

### Jon responds to Mara

**Mara's enum approach is over-engineered**: Most commands only need one scope type. Why force them to construct a `Scope` struct when a simple method call is clearer?

**Mara's single method becomes complex**: With 5 scope types, the method needs conditional logic anyway. Separate methods are simpler.

**Compromise**: Use separate methods, but make reverse lookup helpers that call `QueryAllDocs` internally (not separate core methods).

### `pkg/commands/*` responds to Mara

**Mara's enum approach is verbose**: Commands have to construct `Scope` structs. With optional fields (Jon's approach), we can pass flags directly.

**Mara's mutual exclusivity is unclear**: How do we know which scope fields are valid together? Optional fields with validation are clearer.

**Preference**: Jon's separate methods OR `pkg/commands/*`'s optional fields approach. Both are easier than enums.

### `pkg/commands/*` responds to Jon

**Jon's separate methods don't handle combinations**: `--ticket --doc-type` needs both ticket scope AND doc-type filter. Separate methods can't express this cleanly.

**Preference**: Our optional fields approach handles combinations naturally:
```go
repo.QueryDocs(QueryOptions{
    TicketID: "MEN-3475",
    Filters: Filters{DocType: "design-doc"},
})
```

### `documents.WalkDocuments` responds to commands

**Commands are asking me to do too much**: I'm a traversal primitive. If you want query semantics, build that on top of me.

**Current design is correct**: Commands should:
1. Resolve scope → root path
2. Call me with root path
3. Apply filters in callback

**If you want a query API, create `internal/repository` that wraps me**: But don't change my semantics.

### Mara responds to `documents.WalkDocuments`

**I agree**: Keep traversal primitive, add repository layer. But the repository layer should use my scope enum approach for consistency.

### `paths.Resolver` responds to all

**Reverse lookup must use normalization**: Current code has bugs because matching is inconsistent. Whatever scope model you choose, reverse lookup should normalize paths through me.

**Suggestion**: Add `NormalizeQueryPath(path string) NormalizedPath` helper that all reverse lookup scopes use.

## Moderator Summary

### Key Arguments

1. **Scope should be explicit** (unanimous): All candidates agree that scope should be a first-class concept, not implicit conditional logic.

2. **API shape** (disagreement):
   - **Mara**: Single method with scope enum (`QueryDocs(Scope, Filters)`).
   - **Jon**: Separate methods per scope type (`QueryAllDocs`, `QueryTicketDocs`, etc.).
   - **`pkg/commands/*`**: Single method with optional fields (`QueryDocs(QueryOptions)`).

3. **Reverse lookup** (disagreement):
   - **Mara**: First-class scope type (`ScopeReverseFile`, `ScopeReverseDir`).
   - **Jon**: Helper methods (not core scope).
   - **`paths.Resolver`**: Must use normalization consistently.

4. **Traversal vs query** (agreement): All agree that `documents.WalkDocuments` should remain a primitive, and scope resolution should live in repository layer.

### Tensions

1. **Simplicity vs flexibility**: Jon's separate methods are simpler but don't handle combinations well. `pkg/commands/*`'s optional fields handle combinations but are more complex.

2. **Enum vs optional fields**: Mara's enum is type-safe but verbose. Optional fields are ergonomic but require validation.

3. **Reverse lookup status**: Is it a core scope type (Mara) or a helper (Jon)? Current implementation treats it as filtering, not scope.

### Interesting Ideas

1. **Convenience wrappers**: Mara's suggestion to provide `QueryAllDocs()` that calls `QueryDocs(Scope{Type: ScopeRepo})` gives both simplicity and consistency.

2. **Iterator/channel API**: `documents.WalkDocuments`'s suggestion to return an iterator instead of a slice could enable lazy evaluation and better memory usage.

3. **NormalizedPath for reverse lookup**: `paths.Resolver`'s requirement that reverse lookup normalize paths ensures consistency.

### Open Questions

1. **Default scope**: Should `QueryDocs()` with no scope default to repo-wide, or require explicit scope?

2. **Combinations**: How should `--ticket --doc-type` be expressed?
   - Mara: `Scope{Type: ScopeTicket, TicketID: "X"}` + `Filters{DocType: "Y"}`
   - Jon: `QueryTicketDocs("X", Filters{DocType: "Y"})`
   - `pkg/commands/*`: `QueryOptions{TicketID: "X", Filters: Filters{DocType: "Y"}}`

3. **Reverse lookup implementation**: Should it:
   - Walk repo-wide and filter (current)
   - Use an index/cache of RelatedFiles
   - Be a separate traversal primitive

4. **Error handling**: Should scope resolution errors (e.g., ticket not found) be:
   - Returned immediately (fail-fast)
   - Included in result set as error handles

### Next Steps

1. **Prototype the API** with 2-3 command call sites (`search`, `meta update`, `doctor`).
2. **Decide on scope model**: Enum vs separate methods vs optional fields.
3. **Design reverse lookup**: First-class scope vs helper vs filtering.
4. **Integrate path normalization**: Ensure reverse lookup uses `paths.Resolver` consistently.
5. **Test combinations**: Verify `--ticket --doc-type` works correctly.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/04-debate-round-2-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`

