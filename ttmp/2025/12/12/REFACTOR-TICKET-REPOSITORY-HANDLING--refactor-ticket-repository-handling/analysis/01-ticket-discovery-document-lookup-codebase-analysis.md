---
Title: 'Ticket discovery & document lookup: codebase analysis'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/documents/walk.go
      Note: Shared markdown walker
    - Path: internal/paths/resolver.go
      Note: Path normalization/matching
    - Path: internal/workspace/config.go
      Note: Docs root/config resolution (ResolveRoot
    - Path: internal/workspace/discovery.go
      Note: Ticket discovery (CollectTicketWorkspaces)
    - Path: pkg/commands/add.go
      Note: Doc creation within a ticket
    - Path: pkg/commands/doc_move.go
      Note: Doc path resolution + moving between tickets
    - Path: pkg/commands/doctor.go
      Note: Ticket/doc validation + RelatedFiles existence checks
    - Path: pkg/commands/import_file.go
      Note: findTicketDirectory helper (ticket id -> dir)
    - Path: pkg/commands/list_docs.go
      Note: Doc listing implementation
    - Path: pkg/commands/list_tickets.go
      Note: Ticket listing implementation
    - Path: pkg/commands/meta_update.go
      Note: Bulk doc updates + doc-type filtering
    - Path: pkg/commands/relate.go
      Note: RelatedFiles update + suggestion scanning
    - Path: pkg/commands/rename_ticket.go
      Note: Rename ticket id + update frontmatter
    - Path: pkg/commands/search.go
      Note: Global/ticket-scoped scanning + reverse-lookup
    - Path: pkg/commands/ticket_move.go
      Note: Ticket relocation based on template
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T14:31:19.164845751-05:00
---


## Goal

Centralize and de-duplicate “ticket repository” behavior:

- **Finding tickets** (ticket ID → ticket dir, listing tickets, filtering tickets)
- **Finding ticket documents** (enumerating docs, filtering docs, reading frontmatter, resolving doc paths)
- **Cross-cutting path resolution** (docs-root/config/repo anchors, related-file normalization/existence checks)

This document maps the current codepaths and highlights places where behavior diverges or is duplicated, to inform a future “TicketRepository” (or similar) abstraction.

## Definitions (as implemented today)

- **Docs root**: typically `ttmp/`, resolved via `workspace.ResolveRoot()` (`internal/workspace/config.go`).
- **Ticket workspace**: a directory under docs root that contains an `index.md` with parseable frontmatter.
- **Ticket identity**: primarily derived from `index.md` frontmatter `Ticket: ...` (not from directory name).
- **Document**: any `.md` with parseable frontmatter; often assumed to have `Ticket`, `DocType`, `Title`, etc.

## Core building blocks in the repo

### Root/config resolution

- **`internal/workspace/config.go`**
  - **`ResolveRoot(root string) string`**: fallback chain via explicit `--root`, `.ttmp.yaml`, git root, cwd.
  - **`FindTTMPConfigPath()`**, **`LoadWorkspaceConfig()`**, **`ResolveVocabularyPath()`**, **`FindRepositoryRoot()`**.

Impact: many commands do `settings.Root = workspace.ResolveRoot(settings.Root)`, but not all.

### Ticket discovery (directory → “this is a ticket”)

- **`internal/workspace/discovery.go`**
  - **`CollectTicketWorkspaces(root, skipDir)`**: `filepath.WalkDir` over docs root; for each directory:
    - ignores directories whose base name starts with `_`
    - if it finds `<dir>/index.md`, it tries `documents.ReadDocumentWithFrontmatter(index.md)`
    - returns a `TicketWorkspace{Path, Doc, FrontmatterErr}` and **`fs.SkipDir`** (stops descending into that ticket)
  - **`CollectTicketScaffoldsWithoutIndex(root, skipDir)`**: detects “ticket-like” dirs missing `index.md`.

Key nuance: `CollectTicketWorkspaces()` returns entries even when `index.md` exists but frontmatter parsing fails (`Doc == nil`, `FrontmatterErr != nil`).

### Ticket ID → ticket directory

- **`pkg/commands/import_file.go`**
  - **`findTicketDirectory(root, ticket string)`**:
    - calls `workspace.CollectTicketWorkspaces(root, nil)`
    - returns the first `ws.Path` whose `ws.Doc.Ticket == ticket`
    - otherwise returns `ticket not found`

This helper is used in many commands (`add`, `relate`, `meta update`, `rename-ticket`, `ticket move`, `doc move`, `search`, etc.).

Important consequence: **if a ticket directory exists but `index.md` is missing or has invalid frontmatter, most commands will report “ticket not found”** rather than “ticket exists but index/frontmatter is invalid/missing”.

### Document walking / frontmatter parsing

- **`internal/documents/walk.go`**
  - **`WalkDocuments(root, fn, opts...)`**: `filepath.WalkDir`, skips directories starting with `_` by default, calls `ReadDocumentWithFrontmatter` for every `.md`.

There is also a commands-level wrapper:

- **`pkg/commands/document_utils.go`**
  - **`readDocumentFrontmatter(path)`** → `documents.ReadDocumentWithFrontmatter(path)`

## How “finding tickets” works across commands

### Listing tickets

- **`pkg/commands/list_tickets.go`**
  - resolves root (`workspace.ResolveRoot`)
  - uses `workspace.CollectTicketWorkspaces`
  - filters:
    - `--ticket` uses **substring match**: `strings.Contains(doc.Ticket, settings.Ticket)`
    - `--status` uses equality
  - sorts by `LastUpdated` (from index frontmatter)

### Status / doctor (ticket-level scanning)

- **`pkg/commands/status.go`**: uses `workspace.CollectTicketWorkspaces` for summarizing tickets.
- **`pkg/commands/doctor.go`**:
  - uses `workspace.CollectTicketWorkspaces` with ignore support
  - separately calls `workspace.CollectTicketScaffoldsWithoutIndex` to emit `missing_index` errors
  - validates index frontmatter and then walks inside each ticket for per-file checks

### Rename/move ticket

- **`pkg/commands/rename_ticket.go`**
  - finds ticket dir via `findTicketDirectory` (ticket must be discoverable via valid index frontmatter)
  - updates ticket id across docs using `documents.WalkDocuments(oldDir, ...)`
  - renames directory

- **`pkg/commands/ticket_move.go`**
  - finds ticket dir via `findTicketDirectory`
  - renders destination path via `renderTicketPath` (also used by ticket creation)
  - renames directory and “touches” `LastUpdated` in index.md best-effort

## How “finding ticket documents” works across commands

### Listing docs (global)

- **`pkg/commands/list_docs.go`**
  - resolves root (`workspace.ResolveRoot`)
  - walks *the entire docs root* via `filepath.Walk`
  - considers any `.md` except `index.md` a “doc”
  - parses frontmatter (`readDocumentFrontmatter`) and applies filters
    - `--ticket` uses **strict equality**: `doc.Ticket == settings.Ticket`
    - `--topics` matches any topic, case-insensitive
  - emits diagnostics on parse error in glaze mode (human mode silently skips parse errors)

Notable divergence: this path does **not** use `internal/documents.WalkDocuments` and therefore does **not** inherit its default “skip `_` dirs” behavior.

### Search (global + per-ticket)

- **`pkg/commands/search.go`**
  - uses multiple `filepath.Walk(...)` passes for different modes:
    - global doc scanning under `settings.Root`
    - ticket-scoped scanning (via `findTicketDirectory`)
  - uses `paths.NewResolver(...)` to normalize/compare file paths for reverse-lookup (`--file`, `--dir`) and related-file features

### Add/import/meta/move flows (ticket-scoped)

- **`pkg/commands/add.go`**: `findTicketDirectory` → read `index.md` frontmatter → create doc under `<ticket>/<doc-type>/`.
- **`pkg/commands/import_file.go`**: `findTicketDirectory` → write into `<ticket>/sources/...` and update `index.md`.
- **`pkg/commands/meta_update.go`**:
  - `findTicketDirectory`
  - if `--doc-type` is used, it walks *all markdown files* under ticket and parses frontmatter to filter by `DocType`
- **`pkg/commands/doc_move.go`**:
  - resolves doc path via `resolveDocPath` (custom logic; not `paths.Resolver`)
  - reads source doc frontmatter to get source ticket
  - resolves source/dest ticket dirs via `findTicketDirectory`
  - computes relative path inside ticket to preserve subpath

### Related files behavior (relate vs doctor)

- **`pkg/commands/relate.go`**
  - determines target doc:
    - `--doc` uses the provided path
    - `--ticket` uses `findTicketDirectory` and targets `<ticket>/index.md`
  - normalizes related-file paths using `paths.NewResolver(...).Normalize(...)` (docs-root/doc-path/config-dir anchors)
  - suggestion mode scans docs with yet another `filepath.Walk` and parses frontmatter

- **`pkg/commands/doctor.go`**
  - validates `RelatedFiles` existence using a *separate*, bespoke “candidate list” strategy:
    - repo root, config dir, parent of config dir, cwd, etc.
  - does not reuse `paths.Resolver` / `paths.MatchPaths` / `paths.DirectoryMatch`

This creates a real risk of “relate canonicalizes to X” but “doctor checks existence against a different resolution set”.

## Duplications + inconsistencies worth refactoring

### 1) Multiple walkers with different skip semantics

We currently have at least three traversal patterns:

- `workspace.CollectTicketWorkspaces` (skips `_` dirs, stops at ticket root, index-only parsing)
- `documents.WalkDocuments` (skips `_` dirs, parses all `.md`)
- ad-hoc `filepath.Walk` in commands (often does **not** skip `_` dirs; error-handling differs)

Impact:
- commands can disagree on what “the set of docs” even is (especially with `ttmp/_guidelines`, `ttmp/_templates`, etc.)

### 2) Ticket filter semantics differ by command

- `list tickets --ticket X` uses **substring match**
- `list docs --ticket X` uses **exact match**

This affects scripting and user expectations.

### 3) `findTicketDirectory` is expensive and brittle

- Expensive: it walks the entire docs root and parses every `index.md` each time.
- Brittle: it only matches tickets whose `index.md` frontmatter parses successfully.

This brittleness propagates because many commands depend on it.

### 4) Root resolution is inconsistently applied

Most commands call `workspace.ResolveRoot`, but e.g. `pkg/commands/list.go` currently does not (it uses `settings.Root` directly).

### 5) Path resolution is duplicated across “related files” features

- `relate` uses `internal/paths.Resolver` for normalization and matching
- `doctor` uses bespoke “candidate absolute paths” logic
- `doc_move` uses `resolveDocPath` bespoke logic

## Suggested “central repository” abstraction (proposal)

### Design goals

- **Single source of truth** for ticket discovery and ticket ID → directory resolution.
- **Consistent traversal rules** (skip behavior, error handling, performance).
- **Explicitly represent “ticket exists but is broken”** (missing index, invalid frontmatter) instead of collapsing into “ticket not found”.
- **Cache** discovery results per invocation (avoid N× full-root walks).
- **Unify path normalization** (reuse `internal/paths.Resolver` across commands and validations).

### Shape (conceptual)

Introduce a new internal package (names TBD) such as `internal/repository` (or extend `internal/workspace`) with a stateful object:

- `type Repository struct { Root string; ConfigDir string; RepoRoot string; Resolver *paths.Resolver; ... cached tickets/docs ... }`

Return handles that preserve both “happy-path metadata” and “broken state”:

- `type TicketHandle struct { Dir string; IndexPath string; Doc *models.Document; FrontmatterErr error; TicketID string /* from Doc or inferred */ }`
- `type DocHandle struct { Path string; Doc *models.Document; Body string; FrontmatterErr error }`

Core methods (examples):

- `Tickets(ctx, filter) ([]TicketHandle, error)`
- `TicketByID(ctx, id) (TicketHandle, error)` (optionally: “strict” vs “best-effort” modes)
- `Docs(ctx, filter) ([]DocHandle, error)` (global or ticket-scoped)
- `DocsInTicket(ctx, ticketID) ([]DocHandle, error)`

And helpers used by multiple verbs:

- `ResolveDocPath(ctx, raw string) (string, error)` (replace `resolveDocPath` variants)
- `ResolveRelatedFile(ctx, docPath, rawRelatedFile string) paths.NormalizedPath`
- `RelatedFileExists(ctx, docPath, rfPath string) (bool, details...)` (reuse same resolver as `relate`)

### Immediate refactor candidates

- Move `findTicketDirectory` out of `pkg/commands/import_file.go` into the centralized repository layer and make it:
  - cached
  - able to return “ticket found but index invalid/missing” as a first-class error type
- Replace ad-hoc `filepath.Walk` doc scanning in:
  - `list_docs.go`, `meta_update.go`, `search.go`, `relate.go` (suggest scan)
  with a shared walk that has consistent skip + error rules (likely built on `documents.WalkDocuments`).
- Make `doctor` validate related files using the same `paths.Resolver` + anchors as `relate`, so “canonical path” means the same thing everywhere.

## Notes / next questions for the refactor

- Should ticket ID be inferred from directory name as a fallback when `index.md` is missing/invalid? (Likely yes, to improve UX and allow repair commands.)
- Should “underscore dirs” be globally excluded from doc enumeration, or only from ticket discovery? (Right now behavior varies.)
- Do we want a “strict” mode where invalid frontmatter is a hard error, vs “best-effort” mode where broken docs/tickets are returned with parse errors attached?

