---
Title: 'Debate Round 8 — Q8: How do we keep vocabulary/config concerns from leaking everywhere?'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/config.go
      Note: ResolveVocabularyPath - config-aware resolution
    - Path: pkg/commands/doctor.go
      Note: Vocabulary loading and validation
    - Path: pkg/commands/meta_update.go
      Note: Vocabulary path loaded but not used
    - Path: pkg/commands/vocabulary.go
      Note: LoadVocabulary helper
    - Path: pkg/models/document.go
      Note: Vocabulary-agnostic model
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T15:20:00-05:00
---


# Debate Round 8 — Q8: How do we keep vocabulary/config concerns from leaking everywhere?

## Goal

Debate **Question 8** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> How do we keep vocabulary/config concerns from leaking everywhere?

**Prompt**: "Should repository lookup be aware of vocabulary/config at all, or should it expose a neutral model and let higher layers validate?"

**Acceptance criteria**:
- A clear boundary: what belongs to lookup vs validation vs UI/formatting
- Example: how `doctor` would validate topics/status/doc-type using the new API

## Context

This round is about **separation of concerns + API boundaries**, not performance/security/backwards compatibility.

Current state:
- `doctor` loads vocabulary and validates topics/status/doc-type against it.
- `add` uses vocabulary to validate doc-type before creating documents.
- `meta update` loads vocabulary path but doesn't validate (inconsistent).
- `list_docs`, `list_tickets`, `search` don't use vocabulary at all.

Vocabulary is loaded via `LoadVocabulary()` which resolves path through config, but this logic is duplicated across commands.

## Pre-Debate Research

### Evidence A — `doctor` loads vocabulary and validates extensively

**Location**: `pkg/commands/doctor.go:222-251`

**Findings**:
```go
// Load vocabulary for validation (best-effort)
vocab, _ := LoadVocabulary()
topicSet := map[string]struct{}{}
topicList := make([]string, 0, len(vocab.Topics))
for _, it := range vocab.Topics {
    topicSet[it.Slug] = struct{}{}
    topicList = append(topicList, it.Slug)
}
// ... similar for docTypeSet, intentSet, statusSet
```

- Vocabulary is loaded once per command invocation.
- Converted to sets/maps for O(1) lookup.
- Used to validate topics, doc-type, intent, status in frontmatter.
- Validation emits warnings (not errors) for unknown values.

**Implication**: Vocabulary is a validation concern, not a lookup concern.

### Evidence B — `add` validates doc-type against vocabulary

**Location**: `pkg/commands/add.go` (implicit, via vocabulary resolution)

**Findings**:
- `add` command requires `--doc-type` parameter.
- Doc-type is stored in frontmatter but not validated against vocabulary in `add.go` itself.
- Vocabulary path is resolved and displayed but not used for validation.

**Implication**: Validation is inconsistent—some commands validate, others don't.

### Evidence C — `LoadVocabulary` resolves path through config

**Location**: `pkg/commands/vocabulary.go:13-33`

**Findings**:
```go
func LoadVocabulary() (*models.Vocabulary, error) {
    if path, err := workspace.ResolveVocabularyPath(); err == nil {
        if _, err2 := os.Stat(path); err2 == nil {
            return loadVocabularyFromFile(path)
        }
    }
    // Not found, return empty vocabulary
    return &models.Vocabulary{...}, nil
}
```

- Vocabulary path resolution uses `workspace.ResolveVocabularyPath()` (config-aware).
- Returns empty vocabulary if not found (best-effort, doesn't fail).
- Used by multiple commands but each loads independently.

**Implication**: Vocabulary loading is duplicated across commands.

### Evidence D — `models.Document` has no vocabulary awareness

**Location**: `pkg/models/document.go:69-81`

**Findings**:
- `Document` struct has fields: `Topics []string`, `DocType string`, `Status string`, `Intent string`.
- No validation against vocabulary in the model itself.
- `Document.Validate()` only checks required fields (Title, Ticket, DocType), not vocabulary membership.

**Implication**: Document model is vocabulary-agnostic; validation happens at command layer.

### Evidence E — `meta_update` loads vocabulary path but doesn't validate

**Location**: `pkg/commands/meta_update.go:163-177`

**Findings**:
```go
cfgPath, _ := workspace.FindTTMPConfigPath()
vocabPath, _ := workspace.ResolveVocabularyPath()
// ... vocabulary path is stored in context but never used
```

- Vocabulary path is resolved and stored in `MetaUpdateContext`.
- Never loaded or used for validation.
- Inconsistent with `doctor` which validates vocabulary.

**Implication**: Vocabulary resolution is separated from vocabulary usage.

### Evidence F — `list_docs` and `search` don't use vocabulary at all

**Location**: `pkg/commands/list_docs.go`, `pkg/commands/search.go`

**Findings**:
- These commands enumerate documents but don't validate vocabulary.
- They filter by topics/doc-type/status but don't check if values are valid.
- No vocabulary loading or validation logic.

**Implication**: Vocabulary is optional—some commands use it, others don't.

## Opening Statements

### Mara (Staff Engineer) — "Unify semantics, reduce surprise"

**Position**: Repository lookup API should be **vocabulary-agnostic**. Vocabulary validation belongs in a separate validation layer.

**Proposed architecture**:
```go
// Repository layer: vocabulary-agnostic
type Repository struct {
    // No vocabulary field
}

func (r *Repository) QueryDocs(scope Scope, filters Filters) ([]DocHandle, error) {
    // Returns raw documents, no vocabulary validation
}

// Validation layer: vocabulary-aware
type Validator struct {
    vocab *models.Vocabulary
}

func (v *Validator) ValidateDoc(doc *models.Document) []ValidationIssue {
    // Checks vocabulary membership, required fields, etc.
}

func (v *Validator) ValidateTopics(topics []string) []string {
    // Returns list of unknown topics
}
```

**Rationale**:
- Clear separation: lookup finds documents, validation checks them.
- Commands can opt into validation (doctor) or skip it (list, search).
- Repository API stays simple and doesn't depend on vocabulary file.
- Validation can be tested independently.

**Example usage**:
```go
// Lookup (no vocabulary)
docs, _ := repo.QueryDocs(Scope{Type: ScopeRepo}, Filters{})

// Validation (vocabulary-aware)
validator := NewValidator(vocab)
for _, doc := range docs {
    issues := validator.ValidateDoc(doc)
    // Handle issues
}
```

### Jon (Senior Engineer) — "Small API surface, easy to adopt"

**Position**: Repository should **optionally** accept vocabulary for validation, but not require it.

**Proposed API**:
```go
type Repository struct {
    vocab *models.Vocabulary  // Optional, can be nil
}

func NewRepository(root string, opts RepositoryOptions) (*Repository, error) {
    // opts.Vocabulary is optional
}

func (r *Repository) QueryDocs(scope Scope, filters Filters) ([]DocHandle, error) {
    // If vocab is set, can optionally validate during enumeration
    // But validation is opt-in, not required
}
```

**Rationale**:
- Commands that need validation can pass vocabulary to repository.
- Commands that don't need it can skip vocabulary loading.
- Repository can provide convenience methods like `ValidateDoc()` if vocab is set.
- Simpler than separate validation layer for common cases.

**Example usage**:
```go
// Without vocabulary (fast, no validation)
repo, _ := NewRepository(root, RepositoryOptions{})

// With vocabulary (enables validation)
vocab, _ := LoadVocabulary()
repo, _ := NewRepository(root, RepositoryOptions{Vocabulary: vocab})
docs, _ := repo.QueryDocs(...)
issues := repo.ValidateDoc(doc)  // Convenience method
```

### `pkg/commands/*` (as a bloc) — "I need an API that's easy to call"

**Position**: Give us **vocabulary loading as a service** that we can inject, but keep it separate from repository lookup.

**Proposed API**:
```go
// Vocabulary service (separate from repository)
type VocabularyService struct {
    vocab *models.Vocabulary
}

func NewVocabularyService() (*VocabularyService, error) {
    vocab, _ := LoadVocabulary()
    return &VocabularyService{vocab: vocab}, nil
}

func (vs *VocabularyService) ValidateTopics(topics []string) []string {
    // Returns unknown topics
}

// Repository stays vocabulary-agnostic
type Repository struct {
    // No vocabulary
}

// Commands inject vocabulary service when needed
func (c *DoctorCommand) Run(...) {
    repo := NewRepository(root)
    vocabSvc := NewVocabularyService()
    
    docs, _ := repo.QueryDocs(...)
    for _, doc := range docs {
        unknown := vocabSvc.ValidateTopics(doc.Topics)
        // Handle validation
    }
}
```

**Rationale**:
- Commands control when to load/use vocabulary.
- Repository stays simple and fast.
- Vocabulary service can be shared across commands.
- Clear separation: repository = lookup, vocabulary service = validation.

### `workspace.ResolveVocabularyPath` — "I decide where vocabulary lives"

**Position**: Vocabulary path resolution belongs in workspace/config layer, not repository layer.

**Defense**:
- I already handle vocabulary path resolution through config fallback chain.
- Repository shouldn't need to know about `.ttmp.yaml` or vocabulary file locations.
- Commands should resolve vocabulary path, load vocabulary, then pass it to validation layer.

**Proposed policy**:
1. Keep vocabulary path resolution in `workspace` package.
2. Commands call `ResolveVocabularyPath()` and `LoadVocabulary()`.
3. Pass loaded vocabulary to validation layer (not repository).

**Rationale**:
- I'm a config primitive, not a validation primitive.
- Repository shouldn't depend on config resolution logic.
- Clear separation: config → vocabulary loading → validation.

### `models.Document` — "I'm vocabulary-agnostic by design"

**Position**: I should remain vocabulary-agnostic. Validation belongs outside the model.

**Defense**:
- My `Validate()` method only checks required fields, not vocabulary membership.
- This is intentional—vocabulary is a workspace concern, not a document concern.
- Documents can exist with unknown topics/status/doc-type; validation is optional.

**Proposed policy**:
1. Keep `Document.Validate()` as-is (required fields only).
2. Add separate `VocabularyValidator` that checks vocabulary membership.
3. Commands choose when to apply vocabulary validation.

**Rationale**:
- Documents are portable—they don't depend on workspace vocabulary.
- Validation is contextual (workspace-specific vocabulary).
- Separation enables testing and reuse.

## Rebuttals

### Mara responds to Jon

**Jon's optional vocabulary in repository is still coupling**: Repository shouldn't know about vocabulary at all. Even optional coupling makes the API more complex.

**Jon's convenience methods create confusion**: If `ValidateDoc()` is a convenience method, why is it on Repository? It should be on a Validator.

**Counter-proposal**: Keep repository vocabulary-agnostic, provide separate `Validator` type that commands can use.

### Jon responds to Mara

**Mara's separate validation layer adds complexity**: Now commands need to manage two objects (Repository + Validator) instead of one. My approach keeps it simple.

**Mara's separation is too strict**: If vocabulary validation is common (doctor, add), why force every command to wire up a validator? Repository can provide it as a convenience.

**Compromise**: Repository can have optional vocabulary, but validation methods are clearly marked as "validation helpers" (not core lookup).

### `pkg/commands/*` responds to Mara

**Mara's separate validator is fine, but vocabulary loading is duplicated**: Every command that needs validation calls `LoadVocabulary()`. Why not centralize this?

**Preference**: `pkg/commands/*`'s vocabulary service approach—centralize loading, keep repository simple.

### `pkg/commands/*` responds to Jon

**Jon's optional vocabulary in repository is acceptable**: As long as it's truly optional and doesn't affect lookup performance.

**Preference**: Either Jon's approach OR vocabulary service. Both are better than Mara's strict separation (too much wiring).

### `workspace.ResolveVocabularyPath` responds to all

**I agree with Mara**: Vocabulary path resolution belongs in workspace layer, not repository. Commands should resolve and load, then pass to validation.

**Suggestion**: Provide `workspace.LoadVocabulary()` helper that combines resolution + loading, so commands don't duplicate this logic.

### `models.Document` responds to all

**I agree with Mara**: Keep me vocabulary-agnostic. Validation belongs in a separate layer.

**Suggestion**: Add `Document.ValidateVocabulary(vocab *Vocabulary) []ValidationIssue` method if you want, but keep it optional and separate from `Validate()`.

## Moderator Summary

### Key Arguments

1. **Repository should be vocabulary-agnostic** (Mara, `pkg/commands/*`, `workspace.ResolveVocabularyPath`, `models.Document`): All agree that repository lookup shouldn't depend on vocabulary.

2. **Validation layer** (disagreement):
   - **Mara**: Separate `Validator` type (strict separation).
   - **Jon**: Optional vocabulary in repository with convenience methods.
   - **`pkg/commands/*`**: Separate `VocabularyService` (centralized loading).

3. **Vocabulary loading** (agreement): All agree that vocabulary loading should be centralized (not duplicated in every command).

4. **Document model** (agreement): `models.Document` should remain vocabulary-agnostic; validation is contextual.

### Tensions

1. **Separation vs convenience**: Mara's strict separation is cleaner but requires more wiring. Jon's optional vocabulary is convenient but couples repository to vocabulary.

2. **Where validation lives**: Should validation be:
   - Separate `Validator` type (Mara)
   - Optional methods on Repository (Jon)
   - Separate `VocabularyService` (`pkg/commands/*`)

3. **Vocabulary loading**: Should it be:
   - Command responsibility (Mara)
   - Repository option (Jon)
   - Centralized service (`pkg/commands/*`)

### Interesting Ideas

1. **`workspace.LoadVocabulary()` helper**: Centralize vocabulary loading so commands don't duplicate resolution + loading logic.

2. **Vocabulary service pattern**: `pkg/commands/*`'s suggestion to have a separate service that commands inject provides clear separation without strict layering.

3. **Optional validation methods**: Jon's suggestion to have validation as optional convenience methods (clearly marked) provides ergonomics without coupling.

### Open Questions

1. **Default behavior**: Should repository validation be:
   - Opt-in only (Mara: separate validator)
   - Opt-in via optional vocabulary (Jon)
   - Always available via service (`pkg/commands/*`)

2. **Performance**: Does vocabulary validation during enumeration slow down lookup?
   - If yes: validation should be separate (Mara)
   - If no: optional validation in repository is fine (Jon)

3. **Error handling**: Should vocabulary validation failures be:
   - Warnings (current doctor behavior)
   - Errors (stricter)
   - Configurable (per command)

4. **Vocabulary loading**: Should it be:
   - Lazy (load on first validation)
   - Eager (load at repository creation)
   - Command-controlled (current)

### Next Steps

1. **Design validation layer**: Decide on Validator vs VocabularyService vs optional methods.
2. **Centralize vocabulary loading**: Create `workspace.LoadVocabulary()` helper.
3. **Prototype validation API**: Test with `doctor` and `add` commands.
4. **Measure performance**: Verify that vocabulary validation doesn't slow down lookup.
5. **Document boundaries**: Clearly document what belongs to lookup vs validation vs UI.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/05-debate-round-3-q8-how-do-we-keep-vocabulary-config-concerns-from-leaking-everywhere.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`

