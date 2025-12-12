---
Title: 'Debate Round 2 — Q7: How should we model scope in lookups (repo vs ticket vs doc)?'
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
LastUpdated: 2025-12-12T15:04:32.587234182-05:00
---

# Debate Round 2 — Q7: How should we model scope in lookups (repo vs ticket vs doc)?

## Goal

Debate **Question 7** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> How should we model “scope” in lookups (repo-wide vs ticket-only vs doc-only)?

We want an API design that expresses scope clearly and prevents commands from re-implementing bespoke traversal rules.

## Context

Scope already exists implicitly in command flags:

- `docmgr search` can act repo-wide, ticket-only, or reverse-lookup-by-file/dir.
- `docmgr relate` targets either a specific doc (`--doc`) or a ticket index (`--ticket`).
- `docmgr meta update` targets either a single doc (`--doc`), a ticket (`--ticket`), or ticket+doc-type (`--ticket --doc-type`).
- `docmgr doctor` has explicit modes (`--doc` single file, `--ticket`, `--all`).

This round is about **semantics + ergonomics + structure**, not performance/security/backcompat.

## Pre-Debate Research (codebase evidence)

### Evidence A — `search` encodes multiple scopes in one command

- `pkg/commands/search.go` flags include:
  - `--ticket` (filter docs by ticket frontmatter)
  - `--file` and `--dir` (reverse lookup: find docs that reference a file/dir)
  - `--external-source`
  - `--files` (a mode switch: suggest files instead of searching docs)
- In the implementation, `search`:
  - uses `filepath.Walk(settings.Root, ...)` to scan docs repo-wide
  - (in other branches) resolves `ticketDir` via `findTicketDirectory` and scans within the ticket
  - uses `paths.NewResolver` for matching related-file paths to `--file` and `--dir`
  - hard-codes a skip for `/_templates/` and `/_guidelines/`

### Evidence B — “Doc-only vs Ticket-index” is a repeated pattern

- `pkg/commands/relate.go`:
  - requires **either** `--doc` **or** `--ticket`
  - if `--ticket` is used, it targets `<ticketDir>/index.md` (after resolving ticketDir)

- `pkg/commands/meta_update.go`:
  - `--doc` updates that file
  - `--ticket` updates index by default
  - `--ticket --doc-type` updates all docs of that type under the ticket

### Evidence C — `doctor` already has a crisp scope model

- `pkg/commands/doctor.go`:
  - `--doc`: single-file mode (overrides `--ticket/--all`)
  - otherwise scans discovered tickets and optionally filters to `--ticket`
  - `--all` exists, but in practice it just means “scan all tickets under root” (since scanning is root-based anyway)

### What this suggests

Scope exists but is expressed inconsistently:
- sometimes as “mode switches” (`--files`, `--doc overrides everything`)
- sometimes as “target selection” (`--doc` vs `--ticket`)
- sometimes as “filter” (`--ticket` as metadata filter)

## Debate (Question 7)

### Opening Statements (Round 1)

#### Mara (Staff Engineer) — “Scopes should be explicit types”

We should model scope explicitly in the repository API so call sites don’t invent their own traversal contracts. I want something like:

- `ScopeRepo` (all docs under root)
- `ScopeTicket(ticketID)`
- `ScopeDoc(docPath)`
- `ScopeReverseLookupFile(path)`
- `ScopeReverseLookupDir(path)`

Then expose *one* entry point (conceptually):

- `QueryDocs(ctx, scope, filters...)`

And a few helpers that build those scopes. This preserves a single policy for skipping, parsing, and matching.

#### Jon (Senior Engineer) — “Three scopes are enough; reverse lookup is a separate operation”

I think we should keep it simpler:

- repo-wide (`AllDocs`)
- ticket-wide (`DocsInTicket`)
- single doc (`DocByPath`)

Reverse lookup (`docs that reference file/dir`) is a separate index-like operation, and should be a dedicated method:

- `DocsReferencingFile(path)` / `DocsReferencingDir(path)`

This keeps the primary API understandable.

#### `workspace.ResolveRoot` — “Scope must not re-resolve root differently per call”

Whatever scope model we pick, it must be anchored on the same resolved root/config context. Today each command does its own root resolution + configDir derivation; the new API should do that once and then scope operations just operate on that stable context.

#### `workspace.CollectTicketWorkspaces` — “Ticket scope should start from ticket discovery”

Ticket scope should be a first-class thing: “ticket = directory discovered from index.md”. If a caller asks for `ScopeTicket(MEN-1234)`, the repository should resolve it to a ticket directory and then all ticket scanning should be relative to that directory.

But: we must decide whether ticket scope includes:
- only parseable docs?
- all `.md` with parse errors returned?

#### `documents.WalkDocuments` — “Scope is just a root + parse contract”

Scope can be expressed by “what root to walk”:

- repo scope: walk docs root
- ticket scope: walk ticket directory
- doc scope: read a single file

The real unification is the parse contract: always return `(doc, body, readErr)` and never silently skip errors unless caller requests it.

#### `paths.Resolver` — “Reverse lookup scope needs path normalization baked in”

Reverse lookup isn’t just “another scope”; it’s a different matching problem. It requires canonicalizing:

- the query (`--file`, `--dir`)
- each related file path in docs (using anchors)

So, if we include reverse lookup in `QueryDocs`, the repository must own/construct a consistent `paths.Resolver` per doc, and the caller shouldn’t have to care.

#### `pkg/commands/*` (bloc) — “We want scope to be obvious at the CLI layer”

We want a repository API where the scope aligns with flags:

- `--doc` → doc scope
- `--ticket` → ticket scope
- no target → repo scope

And reverse lookup should be a standard facility used by both `search` and (potentially) future tooling.

### Rebuttals (Round 2)

#### Mara → Jon

If we keep reverse lookup “separate”, we risk two implementations again: `search` will do one thing, `relate` suggestions will do another. I’m okay if reverse lookup is separate methods, but they must share the same underlying scanning primitives and normalization policy.

#### Jon → Mara

Agree; my intent is API clarity. We can still share internals. Exposing reverse lookup as explicit methods makes it clearer what the behavior is.

#### `documents.WalkDocuments` → everyone

Please don’t conflate scope modeling with “filter semantics.” Scope is *where you look*; filters are *what you match*. You can keep scope small and still support rich filters.

### Moderator Summary

**Consensus direction**
- Use a small explicit set of scopes:
  - repo-wide, ticket-wide, doc-only
- Provide reverse lookup as explicit operations, but implemented on shared scanning primitives.

**Key API decision to draft next**
- One of:
  - `QueryDocs(ctx, scope, filters...)`
  - or a small method set (`Docs()`, `DocsInTicket()`, `DocByPath()`, `DocsReferencingFile()`, `DocsReferencingDir()`)

**Concrete follow-up (actionable)**
- Draft 5 method signatures and re-implement one command end-to-end in thought-experiment form:
  - `search` (because it touches both repo scope and reverse lookup)

### Decision (manuel)

- **Selected**: **A)** use a single entry point: **`QueryDocs(ctx, scope, filters...)`**.

### Implications / Next steps

- **Define `Scope`**: minimal variants we need immediately:
  - repo-wide
  - ticket-wide (by ticket ID)
  - doc-only (by doc path)
  - reverse-lookup (file/dir) can be either:
    - additional `Scope` variants, **or**
    - separate helpers that still call `QueryDocs` internally
- **Define parse behavior**:
  - does `QueryDocs` return docs-with-parse-errors as handles, or skip by default?
- **Pick one “golden” call site**:
  - sketch `search` rewritten in terms of `QueryDocs` (repo scope + ticket scope + reverse-lookup).

## Usage Examples

If we continue with a next round, we should pick a “golden” command and rewrite its behavior using the proposed scope model (no code yet, just call-site shape).

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/03-debate-round-1-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`
