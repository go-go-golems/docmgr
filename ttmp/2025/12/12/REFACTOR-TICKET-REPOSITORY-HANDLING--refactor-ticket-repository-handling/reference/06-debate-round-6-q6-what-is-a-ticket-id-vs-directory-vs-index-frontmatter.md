---
Title: 'Debate Round 6 — Q6: What is a ticket (ID vs directory vs index frontmatter)?'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/discovery.go
      Note: CollectTicketWorkspaces - frontmatter-driven discovery
    - Path: pkg/commands/doctor.go
      Note: Multiple index.md detection + exact match filtering
    - Path: pkg/commands/import_file.go
      Note: findTicketDirectory - exact frontmatter match
    - Path: pkg/commands/list_tickets.go
      Note: Substring vs exact match inconsistency
    - Path: pkg/commands/rename_ticket.go
      Note: Directory name used for rename operations
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T15:00:00-05:00
---


# Debate Round 6 — Q6: What is a ticket (ID vs directory vs index frontmatter)?

## Goal

Debate **Question 6** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> What is the canonical meaning of "ticket" (ID vs directory vs index frontmatter)?

**Prompt**: "When resolving a ticket, what is authoritative: directory naming, `index.md` frontmatter `Ticket`, or both?"

**Edge cases to decide**:
- directory name suggests ticket X but `index.md` says ticket Y
- multiple directories have `index.md` with the same ticket ID
- `index.md` missing: should we infer ticket ID from directory name to enable repairs?

**Acceptance criteria**:
- A conflict-resolution policy (including diagnostics the user sees)
- An explicit definition of "ticket identity" used across commands

## Context

This round is about **semantics + UX + structure**, not performance/security/backwards compatibility.

Relevant current behavior (high-level):
- Ticket discovery is **index.md-frontmatter-driven** (`CollectTicketWorkspaces`).
- Many commands resolve "ticket ID → ticket dir" by scanning the docs root and matching `index.md` frontmatter `Ticket`.
- Directory names follow pattern `TICKET--slug`, but are **not treated as authoritative** for identity in the core lookup helper (`findTicketDirectory`).
- Some commands use substring matching (`strings.Contains`) for filtering, others use exact match (`==`).

## Pre-Debate Research

### Evidence A — Current ticket discovery: frontmatter is authoritative

**Location**: `internal/workspace/discovery.go:26-70`

```go
func CollectTicketWorkspaces(root string, skipDir func(relPath, baseName string) bool) ([]TicketWorkspace, error) {
    // ...
    indexPath := filepath.Join(path, "index.md")
    if fi, err := os.Stat(indexPath); err == nil && !fi.IsDir() {
        doc, _, err := documents.ReadDocumentWithFrontmatter(indexPath)
        if err != nil {
            workspaces = append(workspaces, TicketWorkspace{Path: path, FrontmatterErr: err})
        } else {
            workspaces = append(workspaces, TicketWorkspace{Path: path, Doc: doc})
        }
        return fs.SkipDir
    }
}
```

**Findings**:
- Discovery treats any directory with `index.md` as a ticket workspace candidate.
- Ticket identity comes from `doc.Ticket` (parsed frontmatter), NOT from `filepath.Base(path)`.
- If frontmatter parsing fails, workspace is still recorded but with `FrontmatterErr` set and `Doc == nil`.
- Directory name is completely ignored for identity purposes.

### Evidence B — Ticket lookup helper: exact frontmatter match only

**Location**: `pkg/commands/import_file.go:112-123`

```go
func findTicketDirectory(root, ticket string) (string, error) {
    workspaces, err := workspace.CollectTicketWorkspaces(root, nil)
    if err != nil {
        return "", err
    }
    for _, ws := range workspaces {
        if ws.Doc != nil && ws.Doc.Ticket == ticket {  // EXACT MATCH
            return ws.Path, nil
        }
    }
    return "", fmt.Errorf("ticket not found: %s", ticket)
}
```

**Findings**:
- Uses exact match (`ws.Doc.Ticket == ticket`), not substring or directory-name matching.
- Only considers workspaces where `ws.Doc != nil` (valid frontmatter).
- Returns first match (no duplicate detection).
- Used by: `add`, `relate`, `import_file`, `meta_update`, `doc_move`, `ticket_move`, `ticket_close`, `tasks`, `renumber`, `rename_ticket`, `search`.

### Evidence C — Filtering inconsistency: exact vs substring

**Location**: `pkg/commands/list_tickets.go:188,294`

```go
// Exact match in doctor
if settings.Ticket != "" && doc.Ticket != settings.Ticket {
    continue
}

// Substring match in list_tickets
if settings.Ticket != "" && !strings.Contains(doc.Ticket, settings.Ticket) {
    continue
}
```

**Findings**:
- `doctor` uses exact match (`doc.Ticket != settings.Ticket`).
- `list_tickets` uses substring match (`strings.Contains(doc.Ticket, settings.Ticket)`).
- `list.go` also uses substring match.
- This inconsistency means `docmgr list tickets --ticket MEN` matches `MEN-3475`, but `docmgr doctor --ticket MEN` does not.

### Evidence D — Directory name used for rename/move operations

**Location**: `pkg/commands/rename_ticket.go:98-105`

```go
base := filepath.Base(oldDir)
remainder := ""
if strings.HasPrefix(base, settings.Ticket) {
    remainder = strings.TrimPrefix(base, settings.Ticket) // includes leading '-' if present
}
newBase := settings.NewTicket + remainder
newDir := filepath.Join(filepath.Dir(oldDir), newBase)
```

**Findings**:
- `rename_ticket` assumes directory name starts with the old ticket ID.
- Preserves the suffix (slug) when renaming.
- If directory name doesn't match ticket ID, rename still proceeds but may create unexpected directory names.

### Evidence E — Real-world directory patterns

**Sample from `ttmp/`**:
```
Dir: REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling
  Frontmatter Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING

Dir: FIX-FRONTMATTER-VALIDATION--fix-frontmatter-validation-improve-ticket-recognition-and-missing-frontmatter-handling
  Frontmatter Ticket: FIX-FRONTMATTER-VALIDATION
```

**Findings**:
- Directory names follow pattern `TICKET--slug` (double dash separator).
- Frontmatter `Ticket` field matches the directory prefix (before `--`).
- No observed conflicts in practice (directory name matches frontmatter).

### Evidence F — Missing index.md handling

**Location**: `internal/workspace/discovery.go:72-115`

```go
func CollectTicketScaffoldsWithoutIndex(root string, skipDir func(relPath, baseName string) bool) ([]string, error) {
    // Finds directories with scaffold markers but missing index.md
    if hasWorkspaceScaffold(path) {
        missing = append(missing, path)
    }
}
```

**Findings**:
- `doctor` uses this to detect "broken" tickets (scaffold exists but no `index.md`).
- No attempt to infer ticket ID from directory name.
- These directories are not included in `CollectTicketWorkspaces` results.

### Evidence G — Multiple index.md detection

**Location**: `pkg/commands/doctor.go:339-354`

```go
indexFiles := findIndexFiles(ticketPath, settings.IgnoreDirs, settings.IgnoreGlobs)
if len(indexFiles) > 1 {
    // Warning: multiple index.md files found
}
```

**Findings**:
- `doctor` detects multiple `index.md` files within a ticket directory.
- Treated as a warning, not an error.
- No detection of multiple directories with the same ticket ID in frontmatter.

## Opening Statements

### Mara (Staff Engineer) — "Unify semantics, reduce surprise"

**Position**: Frontmatter `Ticket` field is the **single source of truth** for ticket identity. Directory names are **derived metadata** that should match but are not authoritative.

**Proposed policy**:
1. **Ticket identity = `index.md` frontmatter `Ticket` field** (exact match, case-sensitive).
2. **Directory name is a hint, not authority**: When resolving `--ticket X`, scan all workspaces and match `doc.Ticket == X`. If directory name suggests X but frontmatter says Y, frontmatter wins.
3. **Conflict detection**: If multiple directories have `index.md` with `Ticket: X`, return all matches with a diagnostic. Commands should fail with a clear error listing conflicting paths.
4. **Missing index.md**: If directory has scaffold markers but no `index.md`, return a "broken workspace" state. Do NOT infer ticket ID from directory name (too error-prone).
5. **Filtering consistency**: All commands use exact match for `--ticket` filter. If substring matching is desired, add `--ticket-contains` flag.

**Rationale**:
- Frontmatter is user-editable and version-controlled; directory names are filesystem artifacts.
- Single source of truth prevents ambiguity.
- Explicit conflict detection prevents silent failures.
- Consistent filtering reduces cognitive load.

**Example API**:
```go
type TicketIdentity struct {
    ID      string   // From frontmatter
    Path    string   // Directory path
    Conflicts []string // Other paths with same ID
    State   TicketState // Valid, Broken, MissingIndex
}

func (r *Repository) FindTicket(id string) ([]TicketIdentity, error)
```

### Jon (Senior Engineer) — "Small API surface, easy to adopt"

**Position**: Frontmatter is primary, but directory name should be a **fallback** for broken states and a **validation signal** for conflicts.

**Proposed policy**:
1. **Primary**: Match `doc.Ticket == id` (exact).
2. **Fallback**: If no match and directory name matches pattern `^TICKET--`, infer ticket ID from directory name for "repair" operations.
3. **Conflict detection**: If multiple matches, return all with a warning. Let commands decide whether to fail or proceed.
4. **Missing index.md**: If scaffold exists, infer ticket ID from directory name (`filepath.Base(dir)` before `--`). Mark as "inferred" state.
5. **Filtering**: Default to exact match, but allow `--ticket-contains` for convenience.

**Rationale**:
- Enables repair workflows (`doctor` can suggest fixes for broken tickets).
- Directory names are usually reliable (created by `create-ticket`).
- Less strict than Mara's approach, easier migration path.
- Commands can opt into strictness.

**Example API**:
```go
type TicketMatch struct {
    Path      string
    TicketID  string  // From frontmatter or inferred
    Source    string  // "frontmatter" | "directory-name"
    Conflicts []string
}

func (r *Repository) FindTicket(id string, opts FindOptions) ([]TicketMatch, error)
```

### `workspace.CollectTicketWorkspaces` — "Tickets are index.md directories"

**Position**: My current behavior is correct: **directory structure + parseable frontmatter = ticket workspace**. Directory name is irrelevant for identity.

**Defense**:
- I already handle broken states: `TicketWorkspace{Path, FrontmatterErr}`.
- I don't check directory names because they're not reliable (users can rename directories).
- Multiple directories with same ticket ID? That's a caller problem—I return all matches, callers should dedupe.

**Proposed policy**:
1. Keep current behavior: return all directories with `index.md`.
2. Add optional validation: if caller wants, I can check for duplicate `doc.Ticket` values and return diagnostics.
3. Do NOT infer from directory names—that's outside my scope.

**Rationale**:
- I'm a discovery primitive, not a conflict resolver.
- Callers (commands) should handle business logic.
- Adding directory-name inference would make me more complex.

### `pkg/commands/*` (as a bloc) — "I need an API that's easy to call"

**Position**: Give us **one method** that handles all cases: exact match, conflicts, broken states, with clear errors.

**Pain points**:
- `findTicketDirectory` fails silently if multiple matches (returns first).
- No way to detect "directory name suggests X but frontmatter says Y".
- Inconsistent filtering (`==` vs `Contains`) causes confusion.

**Proposed policy**:
1. **Single lookup method**: `FindTicket(id string) (TicketHandle, error)`.
2. **TicketHandle** includes: `ID`, `Path`, `IndexDoc`, `State` (Valid|Broken|MissingIndex|Conflict).
3. **Errors are explicit**: `ErrTicketNotFound`, `ErrTicketConflict([]paths)`, `ErrTicketBroken(reason)`.
4. **Filtering**: Always exact match unless `--ticket-contains` flag.

**Example usage**:
```go
ticket, err := repo.FindTicket("MEN-3475")
if err != nil {
    if errors.Is(err, ErrTicketConflict) {
        // Show user: "Multiple tickets found: path1, path2"
    }
    return err
}
if ticket.State == Broken {
    // Show: "Ticket workspace exists but index.md is invalid: <reason>"
}
```

### `paths.Resolver` — "Normalize everything, compare apples-to-apples"

**Position**: I don't care about ticket identity, but if you're comparing paths, use my normalization.

**Observation**:
- Current code compares raw directory paths (`ws.Path`).
- If paths are relative vs absolute, comparisons can fail.
- RelatedFiles paths use my normalization, but ticket directory resolution doesn't.

**Suggestion**:
- When returning ticket paths, normalize them through me.
- When detecting conflicts, compare normalized paths.
- This ensures `ttmp/2025/12/12/MEN-3475--slug` and `./ttmp/2025/12/12/MEN-3475--slug` are treated as the same.

## Rebuttals

### Mara responds to Jon

**Jon's fallback approach is dangerous**: Inferring ticket ID from directory names enables silent failures. If a user renames a directory but forgets to update `index.md`, commands would operate on the wrong ticket. Better to fail fast with a clear error.

**Jon's "inferred" state adds complexity**: Now callers must check `Source` field and decide whether to trust it. This is exactly the kind of cognitive load I want to eliminate.

**Counter-proposal**: Keep directory-name inference as a **separate repair helper** (`SuggestTicketIDFromDirectory(dir string) string`), not in the core lookup API.

### Jon responds to Mara

**Mara's strictness blocks repair workflows**: If `index.md` is corrupted, `doctor` can't suggest fixes because it can't identify the ticket. My fallback enables "repair mode" where we can reconstruct ticket identity from directory structure.

**Mara's conflict detection is too strict**: If two tickets have the same ID (typo, copy-paste error), failing immediately prevents users from seeing the problem. Better to return all matches with a warning and let commands decide.

**Compromise**: Make inference opt-in via `FindOptions.InferFromDirectory bool`. Default to false (strict), but allow repair tools to opt in.

### `workspace.CollectTicketWorkspaces` responds to commands

**Commands are asking me to do too much**: I'm a discovery primitive. If you want conflict detection, deduplication, or directory-name inference, build that on top of me. Don't make me a monolith.

**Current design is correct**: I return raw data (`[]TicketWorkspace`), callers filter/dedupe. This keeps me simple and testable.

**If you want a higher-level API, create `internal/repository` that wraps me**: But don't change my semantics.

### `pkg/commands/*` responds to `workspace.CollectTicketWorkspaces`

**Your "caller should dedupe" approach causes bugs**: `findTicketDirectory` returns the first match silently. If there are duplicates, users don't know. We need conflict detection at the API level.

**Your broken-state handling is incomplete**: You return `FrontmatterErr`, but many commands ignore it. We need explicit "broken" vs "missing index" states.

**Proposal**: Keep `CollectTicketWorkspaces` as-is, but add `internal/repository` layer that:
- Wraps your results
- Detects conflicts
- Provides `FindTicket(id)` with clear errors
- Handles broken states explicitly

### Mara responds to `workspace.CollectTicketWorkspaces`

**I agree**: Keep discovery primitive, add repository layer. But the repository layer should enforce my strict policy (frontmatter-only, explicit conflicts).

### Jon responds to `workspace.CollectTicketWorkspaces`

**I agree**: Keep discovery primitive. But the repository layer should support my fallback approach (inference, opt-in).

## Moderator Summary

### Key Arguments

1. **Frontmatter is authoritative** (unanimous): All candidates agree that `index.md` frontmatter `Ticket` field is the primary source of ticket identity.

2. **Directory name role** (disagreement):
   - **Mara**: Directory name is derived metadata, not authoritative. Never infer from it.
   - **Jon**: Directory name is a fallback for broken states and repair workflows.
   - **`workspace.CollectTicketWorkspaces`**: Directory name is irrelevant to my scope.

3. **Conflict handling** (disagreement):
   - **Mara**: Fail fast with explicit error listing all conflicting paths.
   - **Jon**: Return all matches with warning, let commands decide.
   - **`workspace.CollectTicketWorkspaces`**: Return all matches, caller's problem.

4. **Missing index.md** (disagreement):
   - **Mara**: Return "broken" state, do NOT infer ticket ID.
   - **Jon**: Infer ticket ID from directory name, mark as "inferred".
   - **`workspace.CollectTicketWorkspaces`**: Already handled via `CollectTicketScaffoldsWithoutIndex`.

5. **Filtering consistency** (agreement): All agree that exact match should be default, with optional substring matching via flag.

### Tensions

1. **Strictness vs. Repair**: Mara's strict approach prevents silent failures but blocks repair workflows. Jon's fallback enables repairs but risks incorrect operations.

2. **API layering**: Should conflict detection live in discovery (`CollectTicketWorkspaces`) or in a repository wrapper? Commands want it in the API, discovery wants to stay primitive.

3. **Error handling**: Should conflicts be errors (fail fast) or warnings (return all matches)? Commands want explicit errors, but Jon argues warnings enable better UX.

### Interesting Ideas

1. **Opt-in inference**: Jon's `FindOptions.InferFromDirectory` provides a middle ground—strict by default, repair-friendly when opted in.

2. **Separate repair helpers**: Mara's suggestion to keep inference as a separate helper (`SuggestTicketIDFromDirectory`) keeps core API clean while enabling repair tools.

3. **Repository wrapper layer**: All agree that `internal/repository` should wrap `CollectTicketWorkspaces` to add conflict detection and explicit error types.

### Open Questions

1. **Default behavior**: Should `FindTicket(id)` fail on conflicts (Mara) or return all matches with warning (Jon)?

2. **Inference scope**: Should directory-name inference be:
   - Never (Mara)
   - Always as fallback (Jon)
   - Opt-in only (compromise)

3. **Conflict resolution UX**: When multiple tickets have same ID, should the API:
   - Return error immediately
   - Return all matches + diagnostic
   - Provide a "resolve conflict" helper method

4. **Missing index.md**: Should we:
   - Never infer ticket ID (Mara)
   - Always infer from directory name (Jon)
   - Infer only in repair mode (compromise)

### Next Steps

1. **Design the repository wrapper API** (`internal/repository`) that wraps `CollectTicketWorkspaces`.
2. **Decide on conflict policy**: Fail-fast vs. return-all-with-warning.
3. **Decide on inference policy**: Never vs. opt-in vs. always-fallback.
4. **Prototype the API** with 2-3 command call sites (`add`, `relate`, `doctor`).
5. **Test edge cases**: Multiple conflicts, broken frontmatter, missing index.md.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/gpt-5-rounds/03-debate-round-1-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md`

