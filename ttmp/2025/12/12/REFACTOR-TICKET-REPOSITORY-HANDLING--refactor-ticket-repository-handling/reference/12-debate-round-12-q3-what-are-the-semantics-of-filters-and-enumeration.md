---
Title: 'Debate Round 12 — Q3: What are the semantics of filters + enumeration?'
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
LastUpdated: 2025-12-12T19:00:00.000000000-05:00
---

# Debate Round 12 — Q3: What are the semantics of filters + enumeration?

## Goal

Debate **Question 3** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> What are the semantics of filters + enumeration?

**Prompt**: "What should be the canonical semantics for enumerating tickets/docs and applying filters, and how do we enforce that across commands?"

**Topics to decide**:
- exact vs substring match for `--ticket` filters
- skip rules for directories (e.g. `_guidelines`, `_templates`, `.meta`, `archive`, etc.)
- whether doc listing is 'all markdown' vs 'only parseable frontmatter docs'
- how to expose the same semantics to `list`, `search`, `doctor`, `meta update`

**Acceptance criteria**:
- A stated policy (1–2 pages) that can be implemented in one place
- Example results for at least 3 commands using the policy

## Context

This round is about **canonical semantics + policy enforcement**, not performance/security/backwards compatibility.

Inputs from previous rounds:
- **Q7**: Scope should be explicit (repo-wide vs ticket-only vs doc-only).
- **Q11**: `QueryDocs` API design includes filters but doesn't define filter semantics.
- **Q2**: Broken states should be represented (invalid frontmatter, missing index).

Current state:
- Ticket filter semantics differ: `list_tickets` uses substring match, `search`/`list_docs` use exact match.
- Skip rules are inconsistent: `search` uses string contains, `WalkDocuments` skips `_` dirs, `doctor` uses ignore patterns.
- Document enumeration differs: some commands skip invalid frontmatter, others return handles with errors.

## Pre-Debate Research

### Evidence A — Ticket filter semantics differ by command

**Location**: `pkg/commands/list_tickets.go:188, 294`, `pkg/commands/search.go:294`, `pkg/commands/list_docs.go:341`

**Findings**:
- `list_tickets`: Uses substring match (`strings.Contains(doc.Ticket, settings.Ticket)`).
- `search`: Uses exact match (`doc.Ticket == settings.Ticket`).
- `list_docs`: Uses exact match (`doc.Ticket == settings.Ticket`).

**Implication**: Ticket filter semantics are inconsistent; Repository API should standardize this.

### Evidence B — Skip rules are implemented inconsistently

**Location**: `pkg/commands/search.go:283`, `internal/documents/walk.go:42-43`, `pkg/commands/doctor.go:692-693`

**Findings**:
- `search`: Skips `/_templates/` and `/_guidelines/` using `strings.Contains(path, "/_templates/")`.
- `WalkDocuments`: Skips directories starting with `_` using `strings.HasPrefix(base, "_")`.
- `doctor`: Uses ignore patterns from `.docmgrignore` and `--ignore-dirs`/`--ignore-globs` flags.
- `CollectTicketWorkspaces`: Skips directories starting with `_` (same as `WalkDocuments`).

**Implication**: Skip rules should be unified; Repository API should provide consistent skip behavior.

### Evidence C — Document enumeration differs: skip vs return-with-error

**Location**: `pkg/commands/search.go:288-291`, `pkg/commands/list_docs.go:188-192`, `internal/documents/walk.go:54-55`

**Findings**:
- `search`: On parse error, silently skips (`return nil`).
- `list_docs`: On parse error, emits `ListingSkip` taxonomy then continues (skips).
- `WalkDocuments`: On parse error, calls callback with `readErr != nil` (returns-with-error).

**Implication**: Document enumeration policy should be explicit (skip vs return-with-error); Repository API should support both.

### Evidence D — Topics filtering uses case-insensitive any-match

**Location**: `pkg/commands/search.go:303-319`, `pkg/commands/list_docs.go:350-366`

**Findings**:
- Topics filtering uses `strings.EqualFold(strings.TrimSpace(filterTopic), strings.TrimSpace(docTopic))`.
- Any-match semantics: if any filter topic matches any doc topic, document is included.
- Case-insensitive matching.

**Implication**: Topics filtering semantics are consistent; Repository API should preserve this.

### Evidence E — `list_docs` explicitly skips `index.md`

**Location**: `pkg/commands/list_docs.go:183-186`

**Findings**:
```go
// Skip index.md files (those are tickets, use list tickets for those)
if info.Name() == "index.md" {
    return nil
}
```

- `list_docs` explicitly skips `index.md` files.
- Other commands (`search`, `WalkDocuments`) don't skip `index.md`.

**Implication**: Document enumeration should have explicit policy on `index.md` handling.

### Evidence F — `meta_update` skips invalid frontmatter when filtering by doc-type

**Location**: `pkg/commands/meta_update.go:352-356`

**Findings**:
```go
if docTypeFilter != "" {
    doc, _, err := documents.ReadDocumentWithFrontmatter(path)
    if err != nil {
        return nil // Skip files with invalid frontmatter
    }
    if doc.DocType != docTypeFilter {
        return nil
    }
}
```

- When filtering by doc-type, `meta_update` skips files with invalid frontmatter.
- Without doc-type filter, it doesn't check frontmatter validity.

**Implication**: Document enumeration policy should be consistent regardless of filters.

### Evidence G — `doctor` uses ignore patterns from `.docmgrignore`

**Location**: `pkg/commands/doctor.go:746-748`

**Findings**:
- `doctor` loads ignore patterns from `<repoRoot>/.docmgrignore`.
- Uses `--ignore-dirs` and `--ignore-globs` flags for additional ignore patterns.
- Other commands don't use `.docmgrignore`.

**Implication**: Ignore rules should be unified; Repository API should support `.docmgrignore` and command flags.

## Opening Statements

### Mara (Staff Engineer) — "Explicit policy with configurable options"

**Position**: Repository API should enforce canonical semantics but allow opt-in/opt-out for edge cases.

**Proposed Policy**:
```go
package repository

// EnumerationPolicy defines what gets enumerated
type EnumerationPolicy struct {
    // Skip rules
    SkipUnderscoreDirs bool // Default: true (skip _templates, _guidelines)
    SkipDotMeta        bool // Default: true (skip .meta/)
    SkipArchive        bool // Default: false (include archive/ by default)
    IgnorePatterns     []string // From .docmgrignore + command flags
    
    // Document inclusion
    IncludeIndexMD     bool // Default: false (exclude index.md from doc enumeration)
    IncludeInvalidDocs bool // Default: false (skip docs with invalid frontmatter)
    
    // Filter semantics
    TicketFilterMatch  MatchType // Default: MatchExact
    StatusFilterMatch  MatchType // Default: MatchExact
    DocTypeFilterMatch MatchType // Default: MatchExact
    TopicsFilterMatch  MatchType // Default: MatchAnyCaseInsensitive
}

type MatchType int
const (
    MatchExact MatchType = iota
    MatchSubstring
    MatchPrefix
    MatchCaseInsensitive
    MatchAnyCaseInsensitive // For topics: any filter topic matches any doc topic
)

// Repository uses policy for all queries
func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) (QueryResult, error) {
    policy := r.defaultPolicy()
    if req.Options.Policy != nil {
        policy = mergePolicy(policy, req.Options.Policy)
    }
    // Apply policy to enumeration and filtering
}
```

**Rationale**:
- Explicit policy makes semantics clear and testable.
- Configurable options allow commands to opt-in/opt-out (e.g., `doctor` might include invalid docs).
- Default policy provides sensible defaults (exact match for ticket, skip underscore dirs, exclude index.md).

**Example usage (`list_docs`)**:
```go
policy := EnumerationPolicy{
    SkipUnderscoreDirs: true,
    IncludeIndexMD: false, // Skip index.md
    IncludeInvalidDocs: false, // Skip invalid frontmatter
    TicketFilterMatch: MatchExact,
    TopicsFilterMatch: MatchAnyCaseInsensitive,
}
req := QueryRequest{
    Scope: Scope{Type: ScopeRepo},
    Filters: Filters{Ticket: settings.Ticket, Topics: settings.Topics},
    Options: QueryOptions{Policy: &policy},
}
result, _ := repo.QueryDocs(ctx, req)
```

**Example usage (`search`)**:
```go
policy := EnumerationPolicy{
    SkipUnderscoreDirs: true, // Skip _templates, _guidelines
    IncludeIndexMD: true, // Include index.md in search
    IncludeInvalidDocs: false, // Skip invalid frontmatter
    TicketFilterMatch: MatchExact,
}
req := QueryRequest{
    Scope: Scope{Type: ScopeRepo},
    Filters: Filters{Ticket: settings.Ticket},
    Options: QueryOptions{Policy: &policy},
}
result, _ := repo.QueryDocs(ctx, req)
```

**Example usage (`doctor`)**:
```go
policy := EnumerationPolicy{
    SkipUnderscoreDirs: true,
    IgnorePatterns: loadDocmgrIgnore(), // From .docmgrignore
    IncludeIndexMD: true, // Doctor checks index.md
    IncludeInvalidDocs: true, // Doctor reports invalid docs
    TicketFilterMatch: MatchExact,
}
req := QueryRequest{
    Scope: Scope{Type: ScopeTicket, TicketID: settings.Ticket},
    Options: QueryOptions{Policy: &policy, IncludeErrors: true},
}
result, _ := repo.QueryDocs(ctx, req)
```

**Trade-offs**:
1. **Explicit vs implicit**: Explicit policy is clearer but more verbose. Commands must construct policy structs.
2. **Configurability vs simplicity**: Configurable options are flexible but complex. Default policy should cover 90% of cases.

### Jon (Senior Engineer) — "Sensible defaults, minimal configuration"

**Position**: Repository API should have sensible defaults; only expose configuration for truly necessary cases.

**Proposed Policy**:
```go
package repository

// Default enumeration policy (hardcoded)
const (
    DefaultSkipUnderscoreDirs = true
    DefaultSkipDotMeta = true
    DefaultIncludeIndexMD = false // Exclude index.md from doc enumeration
    DefaultIncludeInvalidDocs = false // Skip invalid frontmatter
    DefaultTicketFilterMatch = MatchExact
    DefaultTopicsFilterMatch = MatchAnyCaseInsensitive
)

// Only expose configuration for edge cases
type QueryOptions struct {
    IncludeIndexMD     bool // Override default (false)
    IncludeInvalidDocs bool // Override default (false)
    IgnorePatterns     []string // Additional ignore patterns
}

func (r *Repository) QueryDocs(ctx context.Context, scope Scope, filters Filters, opts ...QueryOption) ([]DocHandle, error) {
    // Use defaults + opts overrides
}
```

**Rationale**:
- Sensible defaults cover most cases (exact match for ticket, skip underscore dirs, exclude index.md).
- Minimal configuration keeps API simple (only expose options that commands actually need).
- Hardcoded defaults are easier to reason about than policy structs.

**Example usage (`list_docs`)**:
```go
// Uses defaults: skip index.md, skip invalid docs, exact match
docs, _ := repo.QueryDocs(ctx, Scope{Type: ScopeRepo}, Filters{
    Ticket: settings.Ticket,
    Topics: settings.Topics,
})
```

**Example usage (`search`)**:
```go
// Override: include index.md
docs, _ := repo.QueryDocs(ctx, Scope{Type: ScopeRepo}, Filters{
    Ticket: settings.Ticket,
}, WithIncludeIndexMD(true))
```

**Example usage (`doctor`)**:
```go
// Override: include invalid docs, include index.md
docs, _ := repo.QueryDocs(ctx, Scope{Type: ScopeTicket, TicketID: settings.Ticket}, Filters{},
    WithIncludeIndexMD(true),
    WithIncludeInvalidDocs(true),
    WithIgnorePatterns(loadDocmgrIgnore()),
)
```

**Trade-offs**:
1. **Defaults vs configurability**: Hardcoded defaults are simpler but less flexible. Commands can't customize skip rules easily.
2. **Minimal vs explicit**: Minimal configuration is easier but less clear (what are the defaults?).

### `pkg/commands/*` (as a bloc) — "Match current behavior, make it consistent"

**Position**: Repository API should match current command behavior but make it consistent across commands.

**Proposed Policy**:
```go
package repository

// Match current behavior
const (
    // Skip rules: always skip _templates, _guidelines, .meta
    AlwaysSkipDirs = []string{"_templates", "_guidelines", ".meta"}
    
    // Filter semantics: exact match for ticket/status/doc-type, case-insensitive any-match for topics
    TicketFilterExact = true
    TopicsFilterCaseInsensitive = true
    TopicsFilterAnyMatch = true // Any filter topic matches any doc topic
)

// Document enumeration: skip index.md, skip invalid frontmatter (unless IncludeErrors=true)
func (r *Repository) QueryDocs(ctx context.Context, opts QueryOptions) ([]DocHandle, error) {
    // Always skip AlwaysSkipDirs
    // Skip index.md unless opts.IncludeIndexMD
    // Skip invalid docs unless opts.IncludeErrors
    // Use exact match for ticket/status/doc-type
    // Use case-insensitive any-match for topics
}
```

**Rationale**:
- Match current behavior to minimize migration effort.
- Make it consistent: all commands use same skip rules and filter semantics.
- Simple API: no policy structs, just consistent behavior.

**Example usage (`list_docs`)**:
```go
// Uses defaults: skip index.md, skip invalid docs, exact match
docs, _ := repo.QueryDocs(ctx, QueryOptions{
    Filters: Filters{Ticket: settings.Ticket, Topics: settings.Topics},
})
```

**Example usage (`search`)**:
```go
// Override: include index.md
docs, _ := repo.QueryDocs(ctx, QueryOptions{
    Filters: Filters{Ticket: settings.Ticket},
    IncludeIndexMD: true,
})
```

**Example usage (`doctor`)**:
```go
// Override: include invalid docs, include index.md
docs, _ := repo.QueryDocs(ctx, QueryOptions{
    TicketID: settings.Ticket,
    IncludeIndexMD: true,
    IncludeErrors: true, // Return handles with ReadErr
    IgnorePatterns: loadDocmgrIgnore(),
})
```

**Trade-offs**:
1. **Current vs ideal**: Matching current behavior is easier but may preserve inconsistencies (e.g., substring match for tickets).
2. **Consistency vs flexibility**: Consistent behavior is simpler but less flexible (can't customize skip rules per-command).

### `internal/documents/WalkDocuments` — "I'm the enumeration primitive; policy belongs above me"

**Position**: `WalkDocuments` is the enumeration primitive; skip rules and filter semantics belong in Repository layer.

**Defense**:
- I enumerate all markdown files and call callback with `readErr` (I don't skip).
- Skip rules (`_` dirs) are built into me, but additional skip rules should be configured via `WithSkipDir`.
- Filter semantics (exact vs substring) don't belong in me; Repository should apply filters after enumeration.

**Proposed Policy**:
```go
// WalkDocuments remains as-is (enumeration primitive)
func WalkDocuments(root string, fn WalkDocumentFunc, opts ...WalkOption) error

// Repository applies skip rules and filters
func (r *Repository) QueryDocs(ctx context.Context, req QueryRequest) (QueryResult, error) {
    // Configure WalkDocuments skip rules
    skipDirs := buildSkipRules(req.Options.Policy)
    
    // Enumerate via WalkDocuments
    err := documents.WalkDocuments(r.Root, func(path string, doc *models.Document, body string, readErr error) error {
        // Apply skip rules (index.md, invalid docs)
        if shouldSkip(path, doc, readErr, req.Options.Policy) {
            return nil
        }
        
        // Apply filters (ticket, topics, etc.)
        if !matchesFilters(doc, req.Filters, req.Options.Policy) {
            return nil
        }
        
        // Include in results
        handles = append(handles, DocHandle{Path: path, Doc: doc, Body: body, ReadErr: readErr})
        return nil
    }, documents.WithSkipDir(skipDirs))
}
```

**Rationale**:
- `WalkDocuments` is enumeration primitive (finds all markdown files).
- Repository applies skip rules and filters (policy layer).
- Clear separation: enumeration vs filtering.

**Suggestion**: Keep `WalkDocuments` as-is; Repository composes it with skip rules and filters.

### `workspace.CollectTicketWorkspaces` — "I already have skip rules"

**Position**: `CollectTicketWorkspaces` already implements skip rules (`_` dirs); Repository should reuse this.

**Defense**:
- I already skip directories starting with `_` (same as `WalkDocuments`).
- I accept `skipDir` predicate for additional skip rules.
- Repository should use me for ticket enumeration, not re-implement skip rules.

**Proposed Policy**:
```go
// Repository uses CollectTicketWorkspaces for ticket enumeration
func (r *Repository) QueryTickets(ctx context.Context, filters TicketFilters) ([]TicketHandle, error) {
    skipDir := buildSkipDirPredicate(filters.IgnorePatterns)
    workspaces, _ := workspace.CollectTicketWorkspaces(r.Root, skipDir)
    
    // Apply filters
    filtered := []TicketHandle{}
    for _, ws := range workspaces {
        if matchesTicketFilters(ws, filters) {
            filtered = append(filtered, convertToTicketHandle(ws))
        }
    }
    return filtered, nil
}
```

**Rationale**:
- Reuse existing `CollectTicketWorkspaces` (don't duplicate skip rules).
- Repository applies filters after enumeration (same pattern as `QueryDocs`).

**Suggestion**: Repository should use `CollectTicketWorkspaces` for ticket enumeration, `WalkDocuments` for doc enumeration.

## Rebuttals

### Mara responds to Jon

**Jon's hardcoded defaults are too rigid**: What if a command needs different skip rules? Hardcoded defaults don't allow customization.

**Jon's minimal configuration doesn't cover all cases**: What about `.docmgrignore`? What about custom skip patterns? Minimal configuration is insufficient.

**Counter-proposal**: Use policy struct with sensible defaults, but allow full customization when needed.

### Jon responds to Mara

**Mara's policy struct is too complex**: Most commands only need defaults. Why force them to construct policy structs?

**Mara's configurability adds complexity**: Every command must understand policy options. Simple defaults are easier.

**Compromise**: Use sensible defaults + minimal options (only expose what commands actually need).

### `pkg/commands/*` responds to Mara

**Mara's policy struct is too verbose**: Commands have to construct policy structs. Our simple options are easier.

**Mara's configurability is overkill**: Most commands just need consistent behavior, not full customization.

**Preference**: Our simple options approach OR Jon's minimal configuration. Both are simpler than policy structs.

### `pkg/commands/*` responds to Jon

**Jon's hardcoded defaults are acceptable**: As long as they match current behavior, hardcoded defaults are fine.

**Jon's minimal configuration is acceptable**: Only expose options that commands actually need.

**Preference**: Either our approach OR Jon's approach. Both are simpler than Mara's policy structs.

### `documents.WalkDocuments` responds to all

**I agree with separation**: I'm enumeration primitive; Repository applies skip rules and filters. This is correct.

**Suggestion**: Keep me as-is; Repository composes me with skip rules and filters.

### `workspace.CollectTicketWorkspaces` responds to all

**I agree with reuse**: Repository should use me for ticket enumeration, not re-implement skip rules.

**Suggestion**: Repository should use `CollectTicketWorkspaces` for tickets, `WalkDocuments` for docs.

## Moderator Summary

### Key Arguments

1. **Policy representation** (disagreement):
   - **Mara**: Explicit policy struct (`EnumerationPolicy`) with full configurability.
   - **Jon**: Sensible defaults + minimal options (only expose what's needed).
   - **`pkg/commands/*`**: Match current behavior, make it consistent (simple options).

2. **Skip rules** (agreement): All agree that skip rules should be unified (skip `_templates`, `_guidelines`, `.meta`).

3. **Filter semantics** (disagreement):
   - **Mara**: Configurable match types (exact, substring, prefix, case-insensitive).
   - **Jon**: Hardcoded defaults (exact match for ticket, case-insensitive any-match for topics).
   - **`pkg/commands/*`**: Match current behavior (exact match for ticket, case-insensitive any-match for topics).

4. **Document enumeration** (agreement): All agree that policy should be explicit (skip vs return-with-error, include vs exclude index.md).

5. **Enumeration primitives** (agreement): All agree that Repository should use `WalkDocuments` for doc enumeration and `CollectTicketWorkspaces` for ticket enumeration.

### Tensions

1. **Explicit vs implicit**: Mara's explicit policy is clearer but more verbose. Jon's defaults are simpler but less clear.

2. **Configurability vs simplicity**: Mara's full configurability is flexible but complex. Jon's minimal options are simpler but less flexible.

3. **Current vs ideal**: `pkg/commands/*` wants to match current behavior, but current behavior has inconsistencies (substring vs exact match).

4. **Policy location**: Should policy be in Repository API (Mara, Jon) or in enumeration primitives (`documents.WalkDocuments`, `workspace.CollectTicketWorkspaces`)?

### Interesting Ideas

1. **Policy struct with defaults**: Mara's suggestion to use policy struct with sensible defaults provides flexibility without forcing commands to specify everything.

2. **Minimal options**: Jon's suggestion to only expose options that commands actually need keeps API simple.

3. **Match current behavior**: `pkg/commands/*`'s suggestion to match current behavior minimizes migration effort but may preserve inconsistencies.

4. **Separation of concerns**: `documents.WalkDocuments`'s requirement that enumeration and filtering be separate ensures clear boundaries.

### Open Questions

1. **Ticket filter semantics**: Should ticket filter be:
   - Exact match (current `search`, `list_docs`)?
   - Substring match (current `list_tickets`)?
   - Configurable (Mara's approach)?

2. **Skip rules**: Should skip rules be:
   - Hardcoded (always skip `_templates`, `_guidelines`, `.meta`)?
   - Configurable (policy struct or options)?
   - From `.docmgrignore` (current `doctor`)?

3. **Document enumeration**: Should document enumeration:
   - Skip invalid frontmatter by default (current `search`, `list_docs`)?
   - Return handles with errors by default (current `WalkDocuments`)?
   - Be configurable (policy option)?

4. **Index.md handling**: Should `index.md` be:
   - Excluded from doc enumeration by default (current `list_docs`)?
   - Included by default (current `search`)?
   - Configurable (policy option)?

5. **Policy location**: Should policy be:
   - In Repository API (Mara, Jon)?
   - In enumeration primitives (`WalkDocuments`, `CollectTicketWorkspaces`)?
   - Separate policy layer?

### Next Steps

1. **Decide on ticket filter semantics**: Exact match vs substring match vs configurable.
2. **Design skip rules**: Hardcoded vs configurable vs `.docmgrignore`.
3. **Design document enumeration policy**: Skip vs return-with-error, include vs exclude index.md.
4. **Design filter semantics**: Match types (exact, substring, case-insensitive, any-match).
5. **Prototype policy**: Test with `list_docs`, `search`, `doctor` commands.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/07-debate-round-7-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/09-debate-round-9-q11-design-querydocs-ctx-scope-filters.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/11-debate-round-11-q2-how-should-we-represent-broken-and-partial-states.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`

