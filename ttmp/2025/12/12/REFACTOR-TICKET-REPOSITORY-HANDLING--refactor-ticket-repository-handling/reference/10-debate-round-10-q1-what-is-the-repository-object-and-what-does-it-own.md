---
Title: 'Debate Round 10 — Q1: What is the Repository object and what does it own?'
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
      Note: WalkDocuments - unified document walking with skip rules
    - Path: internal/paths/resolver.go
      Note: Path normalization needs context (DocsRoot
    - Path: internal/workspace/config.go
      Note: WorkspaceConfig exists but is not a repository object
    - Path: internal/workspace/discovery.go
      Note: CollectTicketWorkspaces - stateless but expensive (full walk each time)
    - Path: pkg/commands/import_file.go
      Note: findTicketDirectory - expensive ticket ID resolution (full walk each time)
    - Path: pkg/commands/list_docs.go
      Note: Document enumeration with inconsistent skip rules
    - Path: pkg/commands/search.go
      Note: Document enumeration with inconsistent skip rules
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T17:00:00-05:00
---


# Debate Round 10 — Q1: What is the Repository object and what does it own?

## Goal

Debate **Question 1** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> What is the "Repository" object and what does it own?

**Prompt**: "If we introduce `Repository` / `TicketRepository` / `Workspace` as a first-class object, what is the smallest coherent set of responsibilities it should own?"

**Must cover**:
- root/config discovery (or injected context?)
- ticket discovery + ticket-id → directory resolution
- document enumeration (global, per-ticket)
- path normalization for related files / reverse lookup

**Acceptance criteria**:
- A proposed type name + 5–10 method signatures
- The minimal state/config it carries (e.g. `Root`, `ConfigDir`, `RepoRoot`, `Resolver`)
- Clear division of labor between `workspace`, `documents`, `paths`, and the new API

## Context

This round is about **API boundaries + responsibilities**, not performance/security/backwards compatibility.

Inputs from previous rounds:
- **Q6**: Frontmatter is authoritative for ticket identity; broken states must be represented.
- **Q7**: Scope should be explicit (enum vs separate methods vs optional fields—still debated).
- **Q8**: Repository should be vocabulary-agnostic; validation belongs in separate layer.
- **Q11**: `QueryDocs` API design assumes a `Repository` object exists but doesn't define it.

Current state:
- Context resolution (`Root`, `ConfigDir`, `RepoRoot`) is duplicated per-command.
- Ticket discovery (`CollectTicketWorkspaces`) is stateless but expensive (full walk each time).
- Document enumeration is scattered across commands with inconsistent skip rules.
- Path normalization (`paths.Resolver`) is created per-query, not cached.

## Pre-Debate Research

### Evidence A — Context resolution is duplicated per-command

**Location**: `pkg/commands/search.go:221-235`, `pkg/commands/meta_update.go:163-177`

**Findings**:
- Every command calls:
  - `workspace.ResolveRoot()` to get docs root.
  - `workspace.FindTTMPConfigPath()` to get config path.
  - `workspace.ResolveVocabularyPath()` (sometimes) to get vocabulary path.
- Resolution happens per-command invocation, not centralized.
- Commands construct `paths.Resolver` with these values each time.

**Implication**: Repository should hold resolved context (root, configDir, repoRoot) and provide it to queries.

### Evidence B — Ticket discovery is stateless but expensive

**Location**: `internal/workspace/discovery.go:26-70`

**Findings**:
- `CollectTicketWorkspaces(root, skipDir)` walks entire docs root each time.
- Returns `[]TicketWorkspace` with path, doc, and frontmatter error.
- No caching; each call performs full filesystem walk.
- Used by multiple commands (`list_tickets`, `doctor`, `import_file`, etc.).

**Implication**: Repository should cache ticket discovery results per-invocation to avoid N× full-root walks.

### Evidence C — Document enumeration is scattered with inconsistent skip rules

**Location**: `pkg/commands/search.go:271-291`, `pkg/commands/list_docs.go:172-192`, `internal/documents/walk.go:31-57`

**Findings**:
- `search`: Uses `filepath.Walk`, skips `/_templates/` and `/_guidelines/` by string contains.
- `list_docs`: Uses `filepath.Walk`, skips `index.md` explicitly.
- `documents.WalkDocuments`: Skips directories starting with `_` by default.
- Skip rules are inconsistent across commands.

**Implication**: Repository should provide unified document enumeration with consistent skip rules.

### Evidence D — Path normalization is created per-query

**Location**: `pkg/commands/search.go:321-325`, `pkg/commands/relate.go:213-221`

**Findings**:
- Commands create `paths.NewResolver(ResolverOptions{...})` for each query.
- Resolver options include `DocsRoot`, `DocPath`, `ConfigDir`, `RepoRoot`.
- These values are resolved per-query, not cached.

**Implication**: Repository should hold a `paths.Resolver` (or provide resolver factory) to avoid per-query resolution.

### Evidence E — Ticket ID → directory resolution is expensive and brittle

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

- Calls `CollectTicketWorkspaces` (full walk) each time.
- Only matches tickets with valid frontmatter (brittle).
- Used by multiple commands (`import_file`, `rename_ticket`, `meta_update`, `add`, `doc_move`, `ticket_move`).

**Implication**: Repository should provide cached `TicketByID(id)` method.

### Evidence F — `paths.Resolver` needs context from multiple sources

**Location**: `internal/paths/resolver.go:21-73`

**Findings**:
- `ResolverOptions` requires:
  - `DocsRoot`: Resolved docs root.
  - `DocPath`: Current document path (for doc-relative resolution).
  - `ConfigDir`: Config directory (for config-relative resolution).
  - `RepoRoot`: Repository root (for repo-relative resolution).
- Resolver normalizes paths against multiple anchors (repo, doc, config, docs-root, docs-parent).

**Implication**: Repository should provide resolver factory that uses cached context, or hold resolver instances.

### Evidence G — `WorkspaceConfig` exists but is not a repository object

**Location**: `internal/workspace/config.go:66-74`

**Findings**:
- `WorkspaceConfig` struct holds:
  - `Root`: Docs root path.
  - `Defaults`: Default metadata values.
  - `Vocabulary`: Vocabulary file path.
- This is configuration data, not a repository API.
- Commands load config but don't use it as a repository object.

**Implication**: Repository should wrap `WorkspaceConfig` and add lookup methods.

## Opening Statements

### Mara (Staff Engineer) — "Repository holds context and provides unified lookup API"

**Position**: `Repository` should be a stateful object that holds resolved context and provides unified lookup methods.

**Proposed API**:
```go
package repository

type Repository struct {
    // Resolved context (set at construction)
    Root     string // Resolved docs root
    ConfigDir string // Resolved config directory
    RepoRoot  string // Resolved repository root (if any)
    
    // Cached state (lazy-loaded)
    tickets []TicketHandle // Cached ticket discovery results
    ticketsLoaded bool
    
    // Resolver factory (uses cached context)
    resolverFactory func(docPath string) *paths.Resolver
}

func NewRepository(ctx context.Context, opts RepositoryOptions) (*Repository, error) {
    // Resolve context once
    root, _ := workspace.ResolveRoot(opts.Root)
    configPath, _ := workspace.FindTTMPConfigPath()
    configDir := filepath.Dir(configPath)
    repoRoot, _ := workspace.FindRepositoryRoot()
    
    return &Repository{
        Root: root,
        ConfigDir: configDir,
        RepoRoot: repoRoot,
        resolverFactory: func(docPath string) *paths.Resolver {
            return paths.NewResolver(paths.ResolverOptions{
                DocsRoot: root,
                DocPath: docPath,
                ConfigDir: configDir,
                RepoRoot: repoRoot,
            })
        },
    }, nil
}

// Core methods (5-10 signatures)
func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) (QueryResult, error)
func (r *Repository) QueryTickets(ctx context.Context, filters TicketFilters) ([]TicketHandle, error)
func (r *Repository) TicketByID(ctx context.Context, id string) (TicketHandle, error)
func (r *Repository) ResolveDocPath(ctx context.Context, raw string) (string, error)
func (r *Repository) ResolveRelatedFile(ctx context.Context, docPath, rawPath string) (paths.NormalizedPath, error)
func (r *Repository) RelatedFileExists(ctx context.Context, docPath, rfPath string) (bool, error)
```

**Rationale**:
- Holds resolved context (root, configDir, repoRoot) to avoid per-query resolution.
- Caches ticket discovery to avoid N× full-root walks.
- Provides resolver factory that uses cached context.
- Unified API for all lookup operations.

**Division of labor**:
- `workspace`: Config loading, root resolution (stateless utilities).
- `documents`: Frontmatter parsing, document walking (stateless utilities).
- `paths`: Path normalization algorithms (stateless utilities).
- `repository`: Stateful repository object that composes these utilities.

**Example usage**:
```go
repo, _ := repository.NewRepository(ctx, repository.RepositoryOptions{})
ticket, _ := repo.TicketByID(ctx, "MEN-3475")
docs, _ := repo.QueryDocs(ctx, repository.QueryRequest{
    Scope: repository.Scope{Type: repository.ScopeTicket, TicketID: "MEN-3475"},
})
```

**Trade-offs**:
1. **Stateful vs stateless**: Stateful allows caching but requires careful lifecycle management (when to invalidate cache?).
2. **Context injection vs discovery**: Repository discovers context at construction (simpler) vs injecting context (more testable).

### Jon (Senior Engineer) — "Minimal Repository, delegate to existing packages"

**Position**: `Repository` should be minimal—just hold context and delegate to existing packages.

**Proposed API**:
```go
package repository

type Repository struct {
    Root     string
    ConfigDir string
    RepoRoot  string
}

func NewRepository(ctx context.Context) (*Repository, error) {
    root, _ := workspace.ResolveRoot("")
    configPath, _ := workspace.FindTTMPConfigPath()
    configDir := filepath.Dir(configPath)
    repoRoot, _ := workspace.FindRepositoryRoot()
    
    return &Repository{
        Root: root,
        ConfigDir: configDir,
        RepoRoot: repoRoot,
    }, nil
}

// Minimal methods (5-7 signatures)
func (r *Repository) QueryDocs(ctx context.Context, scope Scope, filters Filters) ([]DocHandle, error)
func (r *Repository) QueryTickets(ctx context.Context, filters TicketFilters) ([]TicketHandle, error)
func (r *Repository) TicketByID(ctx context.Context, id string) (TicketHandle, error)
func (r *Repository) Resolver(docPath string) *paths.Resolver
```

**Rationale**:
- Minimal state: only resolved context (root, configDir, repoRoot).
- No caching: delegate to `workspace.CollectTicketWorkspaces` (let callers cache if needed).
- Resolver factory: simple method that creates resolver with cached context.
- Delegate to existing packages: `QueryDocs` uses `documents.WalkDocuments`, `QueryTickets` uses `workspace.CollectTicketWorkspaces`.

**Division of labor**:
- `workspace`: Ticket discovery, config loading (stateless utilities).
- `documents`: Document walking, frontmatter parsing (stateless utilities).
- `paths`: Path normalization (stateless utilities).
- `repository`: Thin wrapper that holds context and composes utilities.

**Example usage**:
```go
repo, _ := repository.NewRepository(ctx)
ticket, _ := repo.TicketByID(ctx, "MEN-3475") // Calls workspace.CollectTicketWorkspaces internally
docs, _ := repo.QueryDocs(ctx, repository.Scope{Type: repository.ScopeTicket, TicketID: "MEN-3475"}, repository.Filters{})
```

**Trade-offs**:
1. **Caching vs simplicity**: No caching keeps Repository simple but may be slower (callers can cache if needed).
2. **Delegation vs abstraction**: Delegates to existing packages (less abstraction) vs provides unified API (more abstraction).

### `workspace` package — "I'm already doing this; extend me"

**Position**: Don't create a new `repository` package; extend `workspace` with repository methods.

**Proposed API**:
```go
package workspace

type Workspace struct {
    Root     string
    ConfigDir string
    RepoRoot  string
    Config    *WorkspaceConfig
}

func NewWorkspace(ctx context.Context) (*Workspace, error) {
    root, _ := ResolveRoot("")
    configPath, _ := FindTTMPConfigPath()
    configDir := filepath.Dir(configPath)
    repoRoot, _ := FindRepositoryRoot()
    config, _ := LoadWorkspaceConfig()
    
    return &Workspace{
        Root: root,
        ConfigDir: configDir,
        RepoRoot: repoRoot,
        Config: config,
    }, nil
}

// Extend existing functions as methods
func (w *Workspace) QueryDocs(ctx context.Context, req QueryRequest) (QueryResult, error)
func (w *Workspace) QueryTickets(ctx context.Context, filters TicketFilters) ([]TicketHandle, error)
func (w *Workspace) TicketByID(ctx context.Context, id string) (TicketHandle, error)
func (w *Workspace) Resolver(docPath string) *paths.Resolver
```

**Rationale**:
- `workspace` already provides config loading, root resolution, ticket discovery.
- Extending `workspace` avoids creating a new package.
- `Workspace` name is clearer than `Repository` (it's a documentation workspace).

**Division of labor**:
- `workspace`: Workspace object + config/root resolution + ticket discovery.
- `documents`: Document walking, frontmatter parsing (stateless utilities).
- `paths`: Path normalization (stateless utilities).
- Commands use `workspace.Workspace` object instead of calling functions directly.

**Example usage**:
```go
ws, _ := workspace.NewWorkspace(ctx)
ticket, _ := ws.TicketByID(ctx, "MEN-3475")
docs, _ := ws.QueryDocs(ctx, workspace.QueryRequest{...})
```

**Trade-offs**:
1. **Package name**: `workspace` is clearer than `repository` but may conflict with existing `workspace` functions.
2. **Extending vs new package**: Extending avoids new package but may make `workspace` package larger.

### `pkg/commands/*` (as a bloc) — "Give us a simple API we can call"

**Position**: Repository should be simple to use from commands; hide complexity.

**Proposed API**:
```go
package repository

type Repository struct {
    root     string
    configDir string
    repoRoot  string
    // Internal caching (hidden from callers)
    ticketsCache []TicketHandle
    ticketsCacheTime time.Time
}

func NewRepository(ctx context.Context) (*Repository, error) {
    // Resolve context once
    root, _ := workspace.ResolveRoot("")
    configPath, _ := workspace.FindTTMPConfigPath()
    configDir := filepath.Dir(configPath)
    repoRoot, _ := workspace.FindRepositoryRoot()
    
    return &Repository{
        root: root,
        configDir: configDir,
        repoRoot: repoRoot,
    }, nil
}

// Simple methods (commands can call directly)
func (r *Repository) QueryDocs(ctx context.Context, opts QueryOptions) ([]DocHandle, error)
func (r *Repository) QueryTickets(ctx context.Context, opts TicketQueryOptions) ([]TicketHandle, error)
func (r *Repository) TicketByID(ctx context.Context, id string) (TicketHandle, error)
func (r *Repository) ResolveDocPath(ctx context.Context, raw string) (string, error)
func (r *Repository) ResolveRelatedFile(ctx context.Context, docPath, rawPath string) (paths.NormalizedPath, error)
```

**Rationale**:
- Simple API: commands can call methods directly without constructing complex request structs.
- Hidden caching: Repository caches internally (commands don't need to manage cache).
- Context resolution: Repository resolves context once at construction (commands don't need to call `ResolveRoot`).

**Division of labor**:
- `workspace`: Config loading, root resolution (stateless utilities, used by Repository).
- `documents`: Document walking, frontmatter parsing (stateless utilities, used by Repository).
- `paths`: Path normalization (stateless utilities, used by Repository).
- `repository`: Stateful repository object that commands use directly.

**Example usage**:
```go
repo, _ := repository.NewRepository(ctx)
ticket, _ := repo.TicketByID(ctx, "MEN-3475")
docs, _ := repo.QueryDocs(ctx, repository.QueryOptions{
    TicketID: "MEN-3475",
    Filters: repository.Filters{DocType: "design-doc"},
})
```

**Trade-offs**:
1. **Simple vs flexible**: Simple API is easier to use but less flexible (can't control caching, etc.).
2. **Hidden complexity**: Hidden caching is convenient but may surprise callers (when is cache invalidated?).

### `paths.Resolver` — "I need context; Repository should provide it"

**Position**: Repository should provide resolver factory that uses cached context.

**Proposed API**:
```go
func (r *Repository) Resolver(docPath string) *paths.Resolver {
    return paths.NewResolver(paths.ResolverOptions{
        DocsRoot: r.Root,
        DocPath: docPath,
        ConfigDir: r.ConfigDir,
        RepoRoot: r.RepoRoot,
    })
}
```

**Rationale**:
- Resolver needs context (root, configDir, repoRoot) that Repository holds.
- Repository should provide resolver factory to avoid per-query context resolution.
- Commands can call `repo.Resolver(docPath)` instead of constructing resolver options.

**Suggestion**: Repository should provide resolver factory method, not hold resolver instances (resolvers are per-document).

## Rebuttals

### Mara responds to Jon

**Jon's minimal Repository doesn't solve the caching problem**: If Repository doesn't cache tickets, every `TicketByID` call walks the entire root. This is expensive.

**Jon's delegation is fine, but caching should be in Repository**: Let Repository cache ticket discovery; callers shouldn't need to manage cache.

**Counter-proposal**: Repository should cache tickets internally, but provide `InvalidateCache()` method for callers who need fresh data.

### Jon responds to Mara

**Mara's caching adds complexity**: When does cache get invalidated? What if filesystem changes? Caching is hard to get right.

**Mara's stateful Repository is harder to test**: Stateless utilities are easier to test; stateful Repository needs mock filesystem.

**Compromise**: Repository can cache tickets, but cache should be explicit (callers can opt out with `ForceRefresh` flag).

### `workspace` responds to Mara

**Mara's new `repository` package duplicates `workspace`**: `workspace` already provides config loading, root resolution, ticket discovery. Why create a new package?

**Suggestion**: Extend `workspace` with repository methods instead of creating a new package.

### `workspace` responds to Jon

**Jon's minimal Repository is fine, but why not extend `workspace`?**: If Repository is just a context holder, why not make it `workspace.Workspace`?

**Suggestion**: Extend `workspace` with `NewWorkspace()` constructor and methods.

### `pkg/commands/*` responds to all

**We prefer Mara's approach**: Caching is important for performance; commands shouldn't need to manage cache.

**We prefer simple API**: `QueryDocs(opts)` is easier than `QueryDocs(req)` with structured request.

**Preference**: Repository with caching + simple API (Mara's approach but with simpler method signatures).

### `paths.Resolver` responds to all

**I agree with resolver factory**: Repository should provide `Resolver(docPath)` method that uses cached context.

**Suggestion**: All candidates should include resolver factory method.

## Moderator Summary

### Key Arguments

1. **Type name** (disagreement):
   - **Mara**: `Repository` (new package `internal/repository`).
   - **Jon**: `Repository` (new package `internal/repository`).
   - **`workspace`**: `Workspace` (extend existing `workspace` package).
   - **`pkg/commands/*`**: `Repository` (new package, simple API).

2. **State** (agreement): All agree Repository should hold resolved context (root, configDir, repoRoot).

3. **Caching** (disagreement):
   - **Mara**: Repository should cache ticket discovery internally.
   - **Jon**: No caching (delegate to existing packages, let callers cache if needed).
   - **`pkg/commands/*`**: Hidden caching (convenient but may surprise callers).

4. **Method signatures** (disagreement):
   - **Mara**: Structured request/response (`QueryDocs(ctx, QueryRequest) (QueryResult, error)`).
   - **Jon**: Minimal signature (`QueryDocs(ctx, Scope, Filters) ([]DocHandle, error)`).
   - **`pkg/commands/*`**: Simple options (`QueryDocs(ctx, QueryOptions) ([]DocHandle, error)`).

5. **Division of labor** (agreement): All agree that Repository should compose existing packages (`workspace`, `documents`, `paths`), not replace them.

6. **Resolver factory** (agreement): All agree Repository should provide resolver factory method.

### Tensions

1. **New package vs extend existing**: Should Repository be a new `internal/repository` package (Mara, Jon) or extend `workspace` (`workspace` package)?

2. **Caching policy**: Should Repository cache tickets internally (Mara, `pkg/commands/*`) or delegate without caching (Jon)?

3. **API complexity**: Should Repository use structured request/response (Mara) or simple options (`pkg/commands/*`)?

4. **State management**: Should Repository be stateful with caching (Mara) or minimal with delegation (Jon)?

### Interesting Ideas

1. **Resolver factory**: All candidates agree Repository should provide `Resolver(docPath)` method that uses cached context.

2. **Cache invalidation**: Mara's suggestion to provide `InvalidateCache()` method gives callers control over cache freshness.

3. **Extend workspace**: `workspace` package's suggestion to extend existing package avoids creating a new package.

4. **Hidden caching**: `pkg/commands/*`'s suggestion to hide caching internally is convenient but may surprise callers.

### Open Questions

1. **Package location**: Should Repository be:
   - New package `internal/repository` (Mara, Jon)?
   - Extended `workspace` package (`workspace` package)?

2. **Caching policy**: Should Repository:
   - Cache tickets internally (Mara, `pkg/commands/*`)?
   - Delegate without caching (Jon)?
   - Provide explicit cache control (`InvalidateCache()` method)?

3. **Method signatures**: Should Repository use:
   - Structured request/response (`QueryDocs(ctx, QueryRequest) (QueryResult, error)`)?
   - Minimal signature (`QueryDocs(ctx, Scope, Filters) ([]DocHandle, error)`)?
   - Simple options (`QueryDocs(ctx, QueryOptions) ([]DocHandle, error)`)?

4. **State management**: Should Repository:
   - Hold cached tickets (stateful)?
   - Delegate to existing packages (stateless)?
   - Provide both (cached methods + uncached methods)?

5. **Resolver instances**: Should Repository:
   - Hold resolver instances (one per document)?
   - Provide resolver factory (create resolver on demand)?

### Next Steps

1. **Decide on package location**: New `internal/repository` vs extend `workspace`.
2. **Decide on caching policy**: Internal caching vs delegation vs explicit cache control.
3. **Design method signatures**: Structured request vs minimal signature vs simple options.
4. **Prototype Repository**: Implement with 2-3 command call sites (`search`, `list_tickets`, `doctor`).
5. **Test caching behavior**: Verify cache invalidation and performance improvements.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/06-debate-round-6-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/07-debate-round-7-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/08-debate-round-8-q8-how-do-we-keep-vocabulary-config-concerns-from-leaking-everywhere.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/09-debate-round-9-q11-design-querydocs-ctx-scope-filters.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`

