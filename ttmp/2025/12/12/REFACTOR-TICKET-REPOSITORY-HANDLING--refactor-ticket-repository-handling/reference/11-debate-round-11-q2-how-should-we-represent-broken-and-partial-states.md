---
Title: 'Debate Round 11 — Q2: How should we represent broken and partial states?'
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
      Note: WalkDocuments contract includes readErr parameter
    - Path: internal/workspace/discovery.go
      Note: CollectTicketScaffoldsWithoutIndex detects missing index.md separately
    - Path: pkg/commands/doctor.go
      Note: Emits diagnostic row and continues (distinguishes missing_index vs invalid_frontmatter)
    - Path: pkg/commands/import_file.go
      Note: findTicketDirectory collapses broken states to not found
    - Path: pkg/commands/list_docs.go
      Note: Emits ListingSkip taxonomy then continues
    - Path: pkg/commands/relate.go
      Note: Fails immediately on parse error
    - Path: pkg/commands/search.go
      Note: Silently skips files with invalid frontmatter
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T18:00:00-05:00
---


# Debate Round 11 — Q2: How should we represent broken and partial states?

## Goal

Debate **Question 2** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> How should we represent "broken" and "partial" states?

**Prompt**: "How should the new API represent these states so commands can act intelligently?"

**Cases**:
- ticket dir exists but `index.md` missing
- `index.md` exists but frontmatter invalid
- docs exist with invalid frontmatter
- doc is outside any recognized ticket workspace

**Acceptance criteria**:
- A small error taxonomy (types or sentinel errors) and/or result structs that include parse errors
- Example: how `docmgr relate`, `docmgr add`, `docmgr doctor` would respond to each case

## Context

This round is about **error representation + API contracts**, not performance/security/backwards compatibility.

Inputs from previous rounds:
- **Q6**: Frontmatter is authoritative for ticket identity; broken states must be represented.
- **Q11**: `QueryDocs` should return `DocHandle` with `ReadErr` field; error policy (skip vs return-with-error) is debated.
- **Q1**: Repository should provide unified lookup API.

Current state:
- `TicketWorkspace` already includes `FrontmatterErr` field.
- Commands handle broken states inconsistently (skip vs fail vs report).
- `WalkDocuments` contract includes `readErr` parameter.
- Missing `index.md` is detected separately via `CollectTicketScaffoldsWithoutIndex`.

## Pre-Debate Research

### Evidence A — `TicketWorkspace` already represents broken states

**Location**: `internal/workspace/discovery.go:14-19, 53-58`

**Findings**:
```go
type TicketWorkspace struct {
    Path           string
    Doc            *models.Document
    FrontmatterErr error
}

// In CollectTicketWorkspaces:
doc, _, err := documents.ReadDocumentWithFrontmatter(indexPath)
if err != nil {
    workspaces = append(workspaces, TicketWorkspace{Path: path, FrontmatterErr: err})
} else {
    workspaces = append(workspaces, TicketWorkspace{Path: path, Doc: doc})
}
```

- `TicketWorkspace` includes `FrontmatterErr` field for invalid frontmatter.
- When `FrontmatterErr != nil`, `Doc == nil`.
- This is already a "broken state" representation.

**Implication**: Repository API can use similar pattern (handle with error field).

### Evidence B — Commands handle broken states inconsistently

**Location**: `pkg/commands/search.go:288-291`, `pkg/commands/list_docs.go:188-192`, `pkg/commands/doctor.go:314-328`

**Findings**:
- `search`: On parse error, silently skips (`return nil`).
- `list_docs`: On parse error, emits `ListingSkip` taxonomy in glaze mode, then continues.
- `doctor`: On parse error, emits diagnostic row and continues (doesn't skip).
- `relate`: On parse error, fails immediately (returns error).

**Implication**: Commands need consistent error handling policy; Repository API should support multiple policies.

### Evidence C — `WalkDocuments` contract includes error parameter

**Location**: `internal/documents/walk.go:11-13, 54-55`

**Findings**:
```go
type WalkDocumentFunc func(path string, doc *models.Document, body string, readErr error) error

// In WalkDocuments:
doc, body, readErr := ReadDocumentWithFrontmatter(path)
return fn(path, doc, body, readErr)
```

- Explicit contract: `doc` and `body` are nil when `readErr` is non-nil.
- Callback receives error alongside path/doc/body.
- This is the "handle with error" pattern.

**Implication**: `DocHandle` should include `ReadErr` field (matches `WalkDocuments` contract).

### Evidence D — Missing `index.md` is detected separately

**Location**: `internal/workspace/discovery.go:72-115`, `pkg/commands/doctor.go:292-309`

**Findings**:
- `CollectTicketScaffoldsWithoutIndex` finds directories with scaffold markers but missing `index.md`.
- `doctor` calls this separately and emits `missing_index` diagnostics.
- Missing `index.md` is not represented in `TicketWorkspace` (only invalid frontmatter is).

**Implication**: Repository API should represent missing `index.md` as distinct from invalid frontmatter.

### Evidence E — `findTicketDirectory` collapses broken states to "not found"

**Location**: `pkg/commands/import_file.go:findTicketDirectory`

**Findings**:
```go
func findTicketDirectory(root, ticket string) (string, error) {
    workspaces, err := workspace.CollectTicketWorkspaces(root, nil)
    if err != nil {
        return "", err
    }
    for _, ws := range workspaces {
        if ws.Doc != nil && ws.Doc.Ticket == ticket {
            return ws.Path, nil
        }
    }
    return "", fmt.Errorf("ticket not found: %s", ticket)
}
```

- Only matches tickets with `ws.Doc != nil` (valid frontmatter).
- Tickets with `FrontmatterErr != nil` are ignored.
- Returns "ticket not found" even if ticket directory exists but frontmatter is invalid.

**Implication**: Repository API should distinguish "ticket not found" from "ticket exists but broken".

### Evidence F — `add` command fails if ticket index is invalid

**Location**: `pkg/commands/add.go:150-180` (implicit)

**Findings**:
- `add` uses `findTicketDirectory` to locate ticket.
- If ticket has invalid frontmatter, `findTicketDirectory` returns "not found".
- `add` fails with "ticket not found" even though ticket directory exists.

**Implication**: Commands like `add` need to handle broken tickets differently (repair vs fail).

### Evidence G — `doctor` distinguishes multiple broken state types

**Location**: `pkg/commands/doctor.go:296-328, 614-629`

**Findings**:
- `doctor` distinguishes:
  - `missing_index`: Directory has scaffold but no `index.md`.
  - `invalid_frontmatter`: `index.md` exists but frontmatter parse fails.
  - `invalid_frontmatter` (doc): Document has invalid frontmatter.
- Each type emits different diagnostic with different severity.

**Implication**: Repository API should distinguish these broken state types.

## Opening Statements

### Mara (Staff Engineer) — "Explicit error taxonomy with typed handles"

**Position**: Repository API should use typed handles with explicit error taxonomy.

**Proposed API**:
```go
package repository

// Error taxonomy
type BrokenState int

const (
    BrokenStateNone BrokenState = iota
    BrokenStateMissingIndex
    BrokenStateInvalidFrontmatter
    BrokenStateOrphanedDoc // Doc outside any ticket workspace
)

type TicketHandle struct {
    Dir         string
    IndexPath   string
    Doc         *models.Document // nil if BrokenState != BrokenStateNone
    BrokenState BrokenState
    ParseErr    error // Non-nil if BrokenState == BrokenStateInvalidFrontmatter
    TicketID    string // From Doc.Ticket or inferred from directory name
}

type DocHandle struct {
    Path        string
    Doc         *models.Document // nil if ReadErr != nil
    Body        string
    ReadErr     error // Non-nil if frontmatter parse failed
    BrokenState BrokenState // BrokenStateOrphanedDoc if outside ticket workspace
}

func (r *Repository) TicketByID(ctx context.Context, id string) (TicketHandle, error) {
    // Returns TicketHandle with BrokenState set if ticket exists but is broken
    // Returns error only if ticket truly not found
}

func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) (QueryResult, error) {
    // Returns DocHandle with ReadErr set for invalid frontmatter
    // Returns DocHandle with BrokenStateOrphanedDoc for docs outside ticket workspaces
}
```

**Rationale**:
- Explicit error taxonomy (`BrokenState`) makes broken states first-class.
- Typed handles (`TicketHandle`, `DocHandle`) carry both success and error state.
- Commands can inspect `BrokenState` to decide how to respond (repair vs fail vs skip).

**Example usage (`add`)**:
```go
ticket, err := repo.TicketByID(ctx, "MEN-3475")
if err != nil {
    return fmt.Errorf("ticket not found: %w", err)
}
if ticket.BrokenState == BrokenStateInvalidFrontmatter {
    return fmt.Errorf("ticket index has invalid frontmatter: %w", ticket.ParseErr)
}
if ticket.BrokenState == BrokenStateMissingIndex {
    // Could auto-repair: create index.md
    return fmt.Errorf("ticket missing index.md (use repair command)")
}
// Ticket is valid, proceed
```

**Example usage (`doctor`)**:
```go
tickets, _ := repo.QueryTickets(ctx, TicketFilters{})
for _, ticket := range tickets {
    switch ticket.BrokenState {
    case BrokenStateMissingIndex:
        emitDiagnostic("missing_index", ticket.Dir)
    case BrokenStateInvalidFrontmatter:
        emitDiagnostic("invalid_frontmatter", ticket.IndexPath, ticket.ParseErr)
    }
}
```

**Trade-offs**:
1. **Explicit vs implicit**: Explicit `BrokenState` enum is clearer but requires maintenance (new states need enum update).
2. **Handle vs error**: Handles with error fields are more flexible but require checking multiple fields.

### Jon (Senior Engineer) — "Simple error field, let callers classify"

**Position**: Use simple error fields in handles; let callers classify errors.

**Proposed API**:
```go
package repository

type TicketHandle struct {
    Dir       string
    IndexPath string
    Doc       *models.Document // nil if Err != nil
    Err       error            // Non-nil if ticket is broken
    TicketID  string          // From Doc.Ticket or inferred
}

type DocHandle struct {
    Path    string
    Doc     *models.Document // nil if ReadErr != nil
    Body    string
    ReadErr error // Non-nil if frontmatter parse failed
}

func (r *Repository) TicketByID(ctx context.Context, id string) (TicketHandle, error) {
    // Returns TicketHandle with Err set if ticket exists but is broken
    // Returns error only if ticket truly not found
}

func (r *Repository) QueryDocs(ctx context.Context, scope Scope, filters Filters) ([]DocHandle, error) {
    // Returns DocHandle with ReadErr set for invalid frontmatter
}
```

**Rationale**:
- Simple error fields (`Err`, `ReadErr`) are easier to use.
- Callers can classify errors using `errors.Is` or error type checks.
- No need to maintain error taxonomy enum.

**Example usage (`add`)**:
```go
ticket, err := repo.TicketByID(ctx, "MEN-3475")
if err != nil {
    return fmt.Errorf("ticket not found: %w", err)
}
if ticket.Err != nil {
    // Classify error
    if errors.Is(ticket.Err, documents.ErrMissingFrontmatter) {
        return fmt.Errorf("ticket missing index.md")
    }
    return fmt.Errorf("ticket index has invalid frontmatter: %w", ticket.Err)
}
// Ticket is valid, proceed
```

**Example usage (`doctor`)**:
```go
tickets, _ := repo.QueryTickets(ctx, TicketFilters{})
for _, ticket := range tickets {
    if ticket.Err != nil {
        // Classify error
        if errors.Is(ticket.Err, documents.ErrMissingFrontmatter) {
            emitDiagnostic("missing_index", ticket.Dir)
        } else {
            emitDiagnostic("invalid_frontmatter", ticket.IndexPath, ticket.Err)
        }
    }
}
```

**Trade-offs**:
1. **Simple vs explicit**: Simple error fields are easier but require callers to classify errors.
2. **Error classification**: Callers must use `errors.Is` or type checks (more code but more flexible).

### `pkg/commands/*` (as a bloc) — "Give us sentinel errors we can check"

**Position**: Use sentinel errors that commands can check with `errors.Is`.

**Proposed API**:
```go
package repository

var (
    ErrTicketNotFound      = errors.New("ticket not found")
    ErrTicketMissingIndex  = errors.New("ticket missing index.md")
    ErrTicketInvalidFrontmatter = errors.New("ticket index has invalid frontmatter")
    ErrDocInvalidFrontmatter    = errors.New("document has invalid frontmatter")
    ErrDocOrphaned              = errors.New("document outside ticket workspace")
)

type TicketHandle struct {
    Dir       string
    IndexPath string
    Doc       *models.Document
    TicketID  string
}

type DocHandle struct {
    Path    string
    Doc     *models.Document
    Body    string
}

func (r *Repository) TicketByID(ctx context.Context, id string) (TicketHandle, error) {
    // Returns ErrTicketNotFound if ticket doesn't exist
    // Returns ErrTicketMissingIndex if ticket dir exists but index.md missing
    // Returns ErrTicketInvalidFrontmatter if index.md exists but frontmatter invalid
    // Returns TicketHandle with Doc set if ticket is valid
}

func (r *Repository) QueryDocs(ctx context.Context, opts QueryOptions) ([]DocHandle, error) {
    // Returns []DocHandle (skips invalid docs unless IncludeErrors=true)
    // If IncludeErrors=true, returns handles with ErrDocInvalidFrontmatter wrapped
}
```

**Rationale**:
- Sentinel errors are easy to check with `errors.Is`.
- Commands can handle different error types explicitly.
- No need for error taxonomy enum or error fields in handles.

**Example usage (`add`)**:
```go
ticket, err := repo.TicketByID(ctx, "MEN-3475")
if errors.Is(err, ErrTicketNotFound) {
    return fmt.Errorf("ticket not found")
}
if errors.Is(err, ErrTicketMissingIndex) {
    return fmt.Errorf("ticket missing index.md (use repair command)")
}
if errors.Is(err, ErrTicketInvalidFrontmatter) {
    return fmt.Errorf("ticket index has invalid frontmatter: %w", err)
}
if err != nil {
    return err
}
// Ticket is valid, proceed
```

**Example usage (`doctor`)**:
```go
tickets, _ := repo.QueryTickets(ctx, TicketQueryOptions{})
for _, ticket := range tickets {
    // Tickets are always valid (broken tickets returned as errors)
}
// Check for broken tickets separately
for _, id := range ticketIDs {
    _, err := repo.TicketByID(ctx, id)
    if errors.Is(err, ErrTicketMissingIndex) {
        emitDiagnostic("missing_index", id)
    } else if errors.Is(err, ErrTicketInvalidFrontmatter) {
        emitDiagnostic("invalid_frontmatter", id, err)
    }
}
```

**Trade-offs**:
1. **Sentinel errors vs error fields**: Sentinel errors are simpler but require separate error returns (can't return handle + error together).
2. **Error handling**: Commands must check errors explicitly (more code but clearer intent).

### `workspace.CollectTicketWorkspaces` — "I already return broken states"

**Position**: `TicketWorkspace` already represents broken states; Repository should use similar pattern.

**Defense**:
- I already return `TicketWorkspace{Path, FrontmatterErr}` for broken tickets.
- Commands can check `ws.Doc == nil` and `ws.FrontmatterErr != nil`.
- This pattern works; Repository should extend it, not replace it.

**Proposed API**:
```go
// Repository uses TicketWorkspace internally
type TicketHandle struct {
    workspace.TicketWorkspace // Embed existing type
    TicketID string // Add inferred ticket ID
}

// Repository methods return TicketHandle (extends TicketWorkspace)
func (r *Repository) TicketByID(ctx context.Context, id string) (TicketHandle, error) {
    workspaces, _ := workspace.CollectTicketWorkspaces(r.Root, nil)
    for _, ws := range workspaces {
        // Match by ID (from Doc.Ticket or inferred)
        if matchesID(ws, id) {
            return TicketHandle{
                TicketWorkspace: ws,
                TicketID: inferTicketID(ws),
            }, nil
        }
    }
    return TicketHandle{}, ErrTicketNotFound
}
```

**Rationale**:
- Reuse existing `TicketWorkspace` type (don't duplicate).
- Commands already know how to handle `FrontmatterErr`.
- Repository extends existing pattern, doesn't replace it.

**Suggestion**: Repository should embed `TicketWorkspace` in `TicketHandle` to maintain compatibility.

### `documents.WalkDocuments` — "My contract already includes errors"

**Position**: `DocHandle` should match `WalkDocuments` contract (path, doc, body, readErr).

**Defense**:
- My contract already includes `readErr` parameter.
- `DocHandle` should match this contract exactly.
- No need for additional error taxonomy; `readErr` is sufficient.

**Proposed API**:
```go
type DocHandle struct {
    Path    string
    Doc     *models.Document // nil if ReadErr != nil
    Body    string           // Empty if ReadErr != nil
    ReadErr error            // nil if parse succeeded
}

// Repository uses WalkDocuments internally
func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) ([]DocHandle, error) {
    handles := []DocHandle{}
    err := documents.WalkDocuments(r.Root, func(path string, doc *models.Document, body string, readErr error) error {
        handles = append(handles, DocHandle{
            Path: path,
            Doc: doc,
            Body: body,
            ReadErr: readErr,
        })
        return nil
    })
    return handles, err
}
```

**Rationale**:
- `DocHandle` matches `WalkDocuments` contract exactly.
- No additional error taxonomy needed; `ReadErr` is sufficient.
- Commands can check `handle.ReadErr != nil` to detect broken docs.

**Suggestion**: Keep `DocHandle` simple; don't add `BrokenState` enum or sentinel errors.

## Rebuttals

### Mara responds to Jon

**Jon's simple error fields require classification**: Callers must use `errors.Is` or type checks to classify errors. This is more code and error-prone.

**Jon's approach doesn't distinguish missing index vs invalid frontmatter**: Both would be `Err != nil`, but callers need to know which case it is.

**Counter-proposal**: Use `BrokenState` enum + error field. Enum distinguishes cases, error provides details.

### Jon responds to Mara

**Mara's error taxonomy enum is over-engineered**: Most commands only need to know "is it broken?" not "what type of broken?". Simple error field is sufficient.

**Mara's enum requires maintenance**: New broken state types require enum update. Error types are more flexible.

**Compromise**: Use error types (not enum) + error field. Error types can be extended without enum update.

### `pkg/commands/*` responds to Mara

**Mara's error taxonomy enum is too complex**: Commands just need to check "is it broken?" Sentinel errors are simpler.

**Mara's handles with error fields are confusing**: Is `BrokenState != BrokenStateNone` the same as `ParseErr != nil`? Redundant fields are confusing.

**Preference**: Sentinel errors (our approach) OR simple error fields (Jon's approach). Both are simpler than enum.

### `pkg/commands/*` responds to Jon

**Jon's error classification is acceptable**: Commands can use `errors.Is` to check error types. This is fine.

**Jon's approach is better than enum**: Error types are more flexible than enum. We prefer sentinel errors, but error fields are acceptable.

**Preference**: Sentinel errors (our approach) OR error fields (Jon's approach). Both are acceptable.

### `workspace.CollectTicketWorkspaces` responds to all

**I agree with embedding**: Repository should embed `TicketWorkspace` in `TicketHandle` to maintain compatibility.

**Missing index is separate**: `CollectTicketScaffoldsWithoutIndex` handles missing index separately. Repository should use both functions.

**Suggestion**: Repository should:
- Use `CollectTicketWorkspaces` for tickets with `index.md` (valid or invalid frontmatter).
- Use `CollectTicketScaffoldsWithoutIndex` for tickets missing `index.md`.
- Combine results into `TicketHandle` with appropriate `BrokenState` or error.

### `documents.WalkDocuments` responds to all

**I agree with simple error field**: `DocHandle` should match my contract (path, doc, body, readErr). No need for additional error taxonomy.

**Orphaned docs are not my concern**: I don't know about ticket workspaces. Repository should handle orphaned doc detection separately.

**Suggestion**: Keep `DocHandle` simple (path, doc, body, readErr). Repository can add `BrokenStateOrphanedDoc` separately if needed.

## Moderator Summary

### Key Arguments

1. **Error representation** (disagreement):
   - **Mara**: Explicit error taxonomy enum (`BrokenState`) + error fields.
   - **Jon**: Simple error fields (`Err`, `ReadErr`); callers classify errors.
   - **`pkg/commands/*`**: Sentinel errors (`ErrTicketNotFound`, etc.).
   - **`workspace.CollectTicketWorkspaces`**: Embed existing `TicketWorkspace` type.

2. **Handle structure** (agreement): All agree handles should include error information (enum, field, or sentinel).

3. **Missing index vs invalid frontmatter** (agreement): All agree these should be distinguished (enum value, error type, or sentinel error).

4. **DocHandle contract** (agreement): All agree `DocHandle` should match `WalkDocuments` contract (path, doc, body, readErr).

5. **TicketHandle structure** (disagreement):
   - **Mara**: `TicketHandle` with `BrokenState` enum + `ParseErr` field.
   - **Jon**: `TicketHandle` with `Err` field.
   - **`pkg/commands/*`**: `TicketHandle` without error (errors returned separately).
   - **`workspace.CollectTicketWorkspaces`**: Embed `TicketWorkspace` in `TicketHandle`.

### Tensions

1. **Explicit vs simple**: Mara's explicit enum is clearer but more complex. Jon's simple error fields are easier but require classification.

2. **Handle vs error return**: Should broken states be in handle fields (`pkg/commands/*`'s sentinel errors return errors separately) or handle fields (Mara, Jon)?

3. **Error taxonomy**: Should errors be classified via enum (Mara), error types (Jon), or sentinel errors (`pkg/commands/*`)?

4. **Compatibility**: Should Repository reuse `TicketWorkspace` (`workspace.CollectTicketWorkspaces`) or define new types (Mara, Jon)?

### Interesting Ideas

1. **Embed TicketWorkspace**: `workspace.CollectTicketWorkspaces`'s suggestion to embed `TicketWorkspace` in `TicketHandle` maintains compatibility.

2. **Error types vs enum**: Jon's suggestion to use error types (not enum) provides flexibility without enum maintenance.

3. **Sentinel errors**: `pkg/commands/*`'s suggestion to use sentinel errors is simple and easy to check with `errors.Is`.

4. **Match WalkDocuments contract**: `documents.WalkDocuments`'s requirement that `DocHandle` match its contract ensures consistency.

### Open Questions

1. **Error representation**: Should broken states be:
   - Enum (`BrokenState`)?
   - Error types (`ErrTicketMissingIndex`, etc.)?
   - Sentinel errors (`ErrTicketNotFound`, etc.)?
   - Simple error fields (`Err`, `ReadErr`)?

2. **Handle vs error return**: Should broken states be:
   - In handle fields (`TicketHandle.Err`, `DocHandle.ReadErr`)?
   - Returned as errors (`TicketByID` returns error for broken tickets)?

3. **Missing index detection**: Should Repository:
   - Use `CollectTicketScaffoldsWithoutIndex` separately?
   - Combine with `CollectTicketWorkspaces` results?
   - Detect missing index internally?

4. **Orphaned docs**: Should Repository:
   - Detect orphaned docs (docs outside ticket workspaces)?
   - Return `BrokenStateOrphanedDoc`?
   - Skip orphaned docs by default?

5. **Compatibility**: Should Repository:
   - Reuse `TicketWorkspace` type (embed in `TicketHandle`)?
   - Define new types (`TicketHandle`, `DocHandle`)?

### Next Steps

1. **Decide on error representation**: Enum vs error types vs sentinel errors vs simple fields.
2. **Design handle structure**: Finalize `TicketHandle` and `DocHandle` fields.
3. **Design missing index handling**: How to detect and represent missing `index.md`.
4. **Design orphaned doc handling**: How to detect and represent docs outside ticket workspaces.
5. **Prototype error handling**: Test with `add`, `doctor`, `relate` commands.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/06-debate-round-6-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/09-debate-round-9-q11-design-querydocs-ctx-scope-filters.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/10-debate-round-10-q1-what-is-the-repository-object-and-what-does-it-own.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`

