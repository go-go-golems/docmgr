---
Title: Debate Questions — Repository lookup API design
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
LastUpdated: 2025-12-12T14:46:53.302712297-05:00
---

# Debate Questions — Repository lookup API design

## Goal

Define the debate questions (3) for designing a new centralized API around repository lookup, ticket finding, document lookup, and workspace-specific functionality.

## Context

This debate is scoped to **API/UX/design semantics**, not:

- performance
- security
- backwards compatibility

Grounding files for context (non-exhaustive):

- `internal/workspace/config.go` (root/config/vocab/repo-root resolution)
- `internal/workspace/discovery.go` (ticket discovery)
- `internal/documents/walk.go` (markdown walking + frontmatter parsing contract)
- `internal/paths/resolver.go` (path normalization/matching)
- `pkg/commands/*` (current command-specific lookup logic)

## Quick Reference

## Question 1 — What is the “Repository” object and what does it own?

Design a new API surface that centralizes lookup without becoming a grab-bag.

- **Prompt**: “If we introduce `Repository` / `TicketRepository` / `Workspace` as a first-class object, what is the smallest coherent set of responsibilities it should own?”
- **Must cover**:
  - root/config discovery (or injected context?)
  - ticket discovery + ticket-id → directory resolution
  - document enumeration (global, per-ticket)
  - path normalization for related files / reverse lookup
- **Forbid**: arguments about perf/security/backcompat.
- **Acceptance criteria** (what a good answer includes):
  - A proposed type name + 5–10 method signatures
  - The minimal state/config it carries (e.g. `Root`, `ConfigDir`, `RepoRoot`, `Resolver`)
  - Clear division of labor between `workspace`, `documents`, `paths`, and the new API

## Question 2 — How should we represent “broken” and “partial” states?

Many current call sites collapse to “ticket not found” when the workspace exists but is broken.

- **Prompt**: “How should the new API represent these states so commands can act intelligently?”
- **Cases**:
  - ticket dir exists but `index.md` missing
  - `index.md` exists but frontmatter invalid
  - docs exist with invalid frontmatter
  - doc is outside any recognized ticket workspace
- **Acceptance criteria**:
  - A small error taxonomy (types or sentinel errors) and/or result structs that include parse errors
  - Example: how `docmgr relate`, `docmgr add`, `docmgr doctor` would respond to each case

## Question 3 — What are the semantics of filters + enumeration?

Currently “what counts as a doc/ticket” and filter behavior differs by command.

- **Prompt**: “What should be the canonical semantics for enumerating tickets/docs and applying filters, and how do we enforce that across commands?”
- **Topics to decide**:
  - exact vs substring match for `--ticket` filters
  - skip rules for directories (e.g. `_guidelines`, `_templates`, `.meta`, `archive`, etc.)
  - whether doc listing is ‘all markdown’ vs ‘only parseable frontmatter docs’
  - how to expose the same semantics to `list`, `search`, `doctor`, `meta update`
- **Acceptance criteria**:
  - A stated policy (1–2 pages) that can be implemented in one place
  - Example results for at least 3 commands using the policy

## Question 4 — Where should “command convenience” live (API vs helper layer)?

- **Prompt**: “Which behaviors should be first-class methods on the repository API, and which should stay as thin command helpers?”
- **Examples to classify**:
  - “ticket index doc path” lookup
  - “doc path input normalization” (accept abs / relative-to-root / relative-to-cwd)
  - “doc-type directory mapping” (e.g. `design-doc` → `design/` vs `design-doc/`)
  - “ticket-scoped default root” when invoked from inside a ticket directory
- **Acceptance criteria**:
  - A clear layering proposal (e.g. `internal/repository` vs `pkg/commands/helpers`)
  - A rule of thumb for future features (“belongs in repo API if …”)

## Question 5 — What should the new API return: paths, documents, or handles?

- **Prompt**: “Should repo lookup methods return raw paths, parsed frontmatter docs, doc bodies, or typed handles that carry both?”
- **Acceptance criteria**:
  - A small set of return types (e.g. `TicketHandle`, `DocHandle`)
  - Explicit treatment of “parse error but still return the file”
  - Concrete example: how `docmgr list docs` and `docmgr doctor` would consume the return type

## Question 6 — What is the canonical meaning of “ticket” (ID vs directory vs index frontmatter)?

- **Prompt**: “When resolving a ticket, what is authoritative: directory naming, `index.md` frontmatter `Ticket`, or both?”
- **Edge cases to decide**:
  - directory name suggests ticket X but `index.md` says ticket Y
  - multiple directories have `index.md` with the same ticket ID
  - `index.md` missing: should we infer ticket ID from directory name to enable repairs?
- **Acceptance criteria**:
  - A conflict-resolution policy (including diagnostics the user sees)
  - An explicit definition of “ticket identity” used across commands

## Question 7 — How should we model “scope” in lookups (repo-wide vs ticket-only vs doc-only)?

- **Prompt**: “What scopes should be built into the API, and how should a caller select them?”
- **Acceptance criteria**:
  - A small scoping model (enums/options) that avoids duplicating methods
  - Example: one API call each for
    - all docs
    - docs within a ticket
    - docs related to a file/directory (reverse lookup)

## Question 8 — How do we keep vocabulary/config concerns from leaking everywhere?

- **Prompt**: “Should repository lookup be aware of vocabulary/config at all, or should it expose a neutral model and let higher layers validate?”
- **Acceptance criteria**:
  - A clear boundary: what belongs to lookup vs validation vs UI/formatting
  - Example: how `doctor` would validate topics/status/doc-type using the new API

## Question 9 — How should ignore rules be unified across features?

- **Prompt**: “What should be the single, canonical ignore mechanism for ticket/doc scans (underscore dirs, `.docmgrignore`, command flags)?”
- **Acceptance criteria**:
  - A single ignore representation (rules + precedence) usable by all scanners
  - Example: how it impacts `list docs`, `doctor`, and `search`

## Question 10 — What is the extension mechanism for workspace-specific functionality?

- **Prompt**: “How does the new repository API support workspace-specific helpers without becoming a monolith?”
- **Examples**:
  - “external sources index” operations (`.meta/sources.yaml` and index updates)
  - “numeric prefix policy” helpers
  - “ticket scaffold structure” checks
- **Acceptance criteria**:
  - A plugin/extension story (interfaces, sub-services, or feature modules)
  - A way to keep core lookup stable while allowing new capabilities

## Question 11 — How should `QueryDocs(ctx, scope, filters...)` be designed?

You’ve selected **one primary API entry point** for doc/ticket lookup; now we need to design it precisely.

- **Prompt**: “What is the smallest, clean, and expressive design for `QueryDocs` that covers repo/ticket/doc scopes and supports reverse lookup and future extensions without turning into a grab-bag?”
- **Must decide**:
  - **`Scope` model** (repo, ticket-by-id, doc-by-path; file/dir reverse lookup as scope vs helper)
  - **Filter model** (struct vs options; exact vs fuzzy; case handling)
  - **Return types** (paths vs parsed docs vs handles including parse errors + bodies)
  - **Parse/error behavior** (skip invalid docs by default vs return-with-error handle; how caller opts in/out)
  - **Determinism** (stable ordering guarantees)
  - **Context plumbing** (what `QueryDocs` needs from resolved root/configDir/repoRoot/path resolver)
- **Acceptance criteria**:
  - Proposed signatures for:
    - `QueryDocs`
    - core types: `Scope`, `Filters`, `DocHandle` (and optionally `TicketHandle`)
  - 2 concrete call-site sketches from `pkg/commands/*`:
    - `search` (repo scan + reverse lookup)
    - `relate` or `doctor` (ticket scope / doc-only)
  - A migration plan at the API level (“how do we adopt this without rewriting everything at once?”)

## Usage Examples

### Suggested moderator format (copy/paste)

1) Restate the question.
2) Ask each candidate for:
   - their proposed API shape
   - 1 concrete example call site in `pkg/commands/*`
   - 2 explicit trade-offs (NOT perf/security/backcompat)
3) Ask rebuttals focused on inconsistencies and adoption burden.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`
