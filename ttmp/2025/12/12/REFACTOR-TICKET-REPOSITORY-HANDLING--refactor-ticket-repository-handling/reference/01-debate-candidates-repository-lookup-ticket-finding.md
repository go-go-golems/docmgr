---
Title: Debate Candidates — Repository lookup & ticket finding
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
LastUpdated: 2025-12-12T14:46:53.158339875-05:00
---

# Debate Candidates — Repository lookup & ticket finding

## Goal

Provide a concrete, codebase-grounded set of debate candidates for designing a new **repository lookup / ticket finding API** (docs root, ticket workspace discovery, doc enumeration, workspace-specific helpers).

## Context

This debate is about centralizing spread/duplicated “repository lookup” logic currently found across:

- `internal/workspace/*` (root resolution + ticket discovery)
- `internal/documents/*` (frontmatter parsing + walking markdown files)
- `internal/paths/*` (path normalization/matching)
- `pkg/commands/*` (ad-hoc ticket/doc lookup and traversal)

Constraints for this debate:

- **Candidates**: 7 total → **5 code-entity personas** + **2 engineers**
- **Do NOT** make this about: performance, security, backwards compatibility

## Quick Reference

### Candidate lineup (7)

#### Engineers (2)

1) **Mara (Staff Engineer) — “Unify semantics, reduce surprise”**
   - **Bias**: Consistency across commands; single mental model; explicit error types.
   - **Cares about**:
     - One shared way to: resolve docs root, find a ticket, list docs, normalize related-file paths
     - Making “ticket exists but broken” a first-class state (missing/invalid `index.md`)
   - **Likely proposal**: Introduce a `Repository`/`TicketRepository` object used by all commands; standardize filters (exact vs contains), standardize traversal + skip rules.

2) **Jon (Senior Engineer) — “Small API surface, easy to adopt”**
   - **Bias**: Minimal viable abstraction that can be rolled out incrementally without boiling the ocean.
   - **Cares about**:
     - Getting 2–3 “core verbs” onto the new API first (e.g. `find ticket dir`, `list docs in ticket`)
     - Keeping usage ergonomic for `pkg/commands/*`
   - **Likely proposal**: Start with a small `RepoLookup` interface + adapters; keep walking/parsing close to call sites until confidence grows.

#### Code entities (5)

3) **`workspace.ResolveRoot` (from `internal/workspace/config.go`) — “I decide what ‘root’ means”**
   - **Represents**: Root/config discovery chain and the anchor points (`.ttmp.yaml`, git root, cwd).
   - **Wants**:
     - One place to compute root/config/vocab/repo-root and share it across commands.
     - A single “context object” (e.g. `RepoContext`) so commands don’t each rebuild configDir/repoRoot.
   - **Pushback**:
     - “Don’t leak my fallback chain in every command; inject it once.”

4) **`workspace.CollectTicketWorkspaces` (from `internal/workspace/discovery.go`) — “Tickets are index.md directories”**
   - **Represents**: Ticket discovery semantics (“ticket workspace = dir with parseable `index.md` frontmatter”).
   - **Wants**:
     - Centralized discovery with configurable skip rules (ignore `_` dirs, `.docmgrignore`, etc.).
     - A richer return type (valid ticket, invalid frontmatter, missing index scaffolds) so callers don’t collapse everything into “not found”.
   - **Pushback**:
     - “Stop re-walking the world via `findTicketDirectory`.”

5) **`documents.WalkDocuments` (from `internal/documents/walk.go`) — “One traversal, one parse contract”**
   - **Represents**: A shared filesystem walker for `.md` that returns `(doc, body, readErr)`.
   - **Wants**:
     - Commands to stop duplicating `filepath.Walk` loops and instead share a consistent traversal + parse behavior.
   - **Pushback**:
     - “If you want consistent doc enumeration, reuse me.”

6) **`paths.Resolver` (from `internal/paths/resolver.go`) — “Normalize everything, compare apples-to-apples”**
   - **Represents**: Path normalization/matching across anchors (repo, doc, config, docs root, docs parent).
   - **Wants**:
     - A single standard for “canonical path” used by:
       - relate (writes RelatedFiles)
       - doctor (validates existence)
       - search (reverse lookup)
   - **Pushback**:
     - “Stop having custom candidate lists; use my anchors consistently.”

7) **`pkg/commands/*` (as a bloc) — “I need an API that’s easy to call”**
   - **Represents**: CLI commands needing ticket/doc lookup (e.g. `add`, `relate`, `meta update`, `doc move`, `search`).
   - **Wants**:
     - A small set of utilities: `TicketDir(ticketID)`, `IndexDoc(ticketID)`, `DocsInTicket(ticketID, filter)`
     - Clear errors: not found vs invalid frontmatter vs outside workspace
   - **Pushback**:
     - “Don’t make me thread 8 parameters; give me a single configured object.”

## Usage Examples

### How to use these candidates in a debate round

- **Pre-debate research**: have each candidate “point at” their own file(s) and summarize current behavior.
- **Opening statements**: each candidate argues for what the new repository API should look like.
- **Rebuttals**: force them to answer:
  - “How does your design prevent duplication?”
  - “How does a command author use it in 3 lines?”
  - “How do we represent broken tickets/docs?”

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`
