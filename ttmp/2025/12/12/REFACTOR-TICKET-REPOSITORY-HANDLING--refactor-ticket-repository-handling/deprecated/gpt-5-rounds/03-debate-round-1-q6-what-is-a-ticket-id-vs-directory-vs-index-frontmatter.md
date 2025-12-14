---
Title: 'Debate Round 1 — Q6: What is a ticket (ID vs directory vs index frontmatter)?'
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
LastUpdated: 2025-12-12T14:56:15.955253467-05:00
---

# Debate Round 1 — Q6: What is a ticket (ID vs directory vs index frontmatter)?

## Goal

Debate **Question 6** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> What is the canonical meaning of “ticket” (ID vs directory vs index frontmatter)?

The purpose is to converge on a **clear, implementable definition** of ticket identity + conflict policy for the future repository lookup API.

## Context

This round is about **semantics + UX + structure**, not performance/security/backwards compatibility.

Relevant current behavior (high-level):

- Ticket discovery is **index.md-frontmatter-driven**.
- Many commands resolve “ticket ID → ticket dir” by scanning the docs root and matching `index.md` frontmatter `Ticket`.
- Directory names look like `TICKET--slug`, but are **not treated as authoritative** for identity in the core lookup helper.

## Pre-Debate Research (codebase evidence)

### Evidence A — Ticket discovery is “dir with index.md”, and ticket identity comes from frontmatter

- `internal/workspace/discovery.go`:
  - `CollectTicketWorkspaces(root, ...)` walks the docs root and treats any directory containing `index.md` as a “ticket workspace candidate”.
  - It calls `documents.ReadDocumentWithFrontmatter(index.md)` and stores the parsed `*models.Document` in `TicketWorkspace.Doc`.
  - If parsing fails, it still returns `TicketWorkspace{Path, FrontmatterErr}` (with `Doc == nil`), which some commands ignore and some (doctor) surface.

Implication: the “ticket” as surfaced to most commands is `ws.Doc.Ticket`, not `filepath.Base(ws.Path)`.

### Evidence B — The shared helper for “ticket id → directory” is frontmatter-only

- `pkg/commands/import_file.go`:

```go
func findTicketDirectory(root, ticket string) (string, error) {
  workspaces, err := workspace.CollectTicketWorkspaces(root, nil)
  ...
  for _, ws := range workspaces {
    if ws.Doc != nil && ws.Doc.Ticket == ticket {
      return ws.Path, nil
    }
  }
  return "", fmt.Errorf("ticket not found: %s", ticket)
}
```

Implication:
- If `index.md` exists but frontmatter is invalid → `ws.Doc == nil` → **treated as “ticket not found”** by most commands.
- If `index.md` is missing entirely → directory is not discovered by `CollectTicketWorkspaces` → also “ticket not found”.
- If two directories have index frontmatter with the same ticket ID → the first match “wins” (currently ordered by path).

### Evidence C — Ticket creation couples directory name and frontmatter, but lookup does not use the directory name

- `pkg/commands/create_ticket.go`:
  - directory path template includes `{{TICKET}}--{{SLUG}}`
  - created `index.md` sets frontmatter `Ticket: <settings.Ticket>`

This means directory names usually mirror the ticket ID, but the *actual* identity is carried in frontmatter.

### Evidence D — Some operations assume directory prefix reflects ticket id (rename-ticket)

- `pkg/commands/rename_ticket.go`:
  - locates ticket dir by ticket ID via `findTicketDirectory`
  - computes new directory base by replacing the leading prefix:
    - `if strings.HasPrefix(base, settings.Ticket) { remainder = strings.TrimPrefix(base, settings.Ticket) }`
    - `newBase := settings.NewTicket + remainder`

Implication: directory base naming is treated as an *expected convention* for rename/move UX, but ticket identity is still frontmatter-driven.

### Evidence E — Doctor treats some missing/invalid cases as directory-driven diagnostics

- `pkg/commands/doctor.go`:
  - `CollectTicketScaffoldsWithoutIndex` emits `missing_index` and uses `filepath.Base(missing)` as the “ticket” field in the output row.
  - Invalid frontmatter for index emits `invalid_frontmatter` and reports `ticket = filepath.Base(ticketPath)` for that case.
  - “multiple index.md files found” is checked inside a ticket directory.

Implication: for “broken tickets”, directory name is currently used as the best available label.

### What’s missing today (explicitly)

There is **no centralized, consistent conflict policy** for:
- directory base says one ticket ID but index frontmatter says another
- two workspaces with the same ticket ID
- “ticket exists but broken” surfaced to commands that need to act (repair, move, relate)

## Debate (Question 6)

### Opening Statements (Round 1)

#### Mara (Staff Engineer) — “Frontmatter is authoritative; directory is presentation”

The code already treats `index.md` frontmatter as the canonical source of truth for `Ticket` via `CollectTicketWorkspaces` and `findTicketDirectory`. We should formalize that: **a ticket is the set of docs whose `Ticket` field matches a ticket ID**, and the ticket workspace directory is a container discovered through `index.md`.

However, we must stop collapsing broken states into “not found”. The new API should return a `TicketHandle` with:
- `Dir`, `IndexPath`
- `TicketID` (from frontmatter if parseable, else inferred from dir name)
- `ParseErr` (for index/doc parse issues)
- `IdentityConfidence` (frontmatter / inferred)

Conflicts should produce structured diagnostics, not silent “pick the first”.

#### Jon (Senior Engineer) — “Directory name is a useful fallback; keep it simple”

The directory name convention (`TICKET--slug`) is clearly part of the workflow. I agree frontmatter should remain authoritative *when valid*, but we need a pragmatic fallback: **if index frontmatter can’t be parsed or is missing, infer ticket ID from directory base** (best-effort) so we can still run commands that repair or move things.

I’d prefer an API that gives:
- `TicketDir(ticketID)` strict (frontmatter-verified)
- `TicketDirsByGuess(ticketID)` best-effort (directory inference)
…so commands can choose strictness explicitly.

#### `workspace.ResolveRoot` — “Identity semantics should be consistent across roots”

Whatever “ticket identity” means, it must be stable across the root resolution chain. If the API sometimes infers from directory and sometimes reads from frontmatter, callers need a consistent contract that doesn’t vary depending on where the root came from (`.ttmp.yaml`, git, cwd).

I want a single `RepoContext`:
- `Root` (docs root)
- `ConfigDir`
- `RepoRoot`
- plus a “ticket identity policy” knob.

#### `workspace.CollectTicketWorkspaces` — “I already encode the current definition: index.md + frontmatter”

Today I define “ticket workspaces” by the presence of `index.md` and (ideally) parseable frontmatter. That’s why most commands don’t even see “missing index” tickets unless doctor reports them.

If we want directory inference, that’s fine, but make it explicit: either extend my return type (e.g., “scaffold without index”) and/or provide a separate discovery method for “ticket-like directories”.

#### `documents.WalkDocuments` — “Identity should be readable from files, not from filenames”

I’m built around: “parse frontmatter when possible; still visit all files”. Ticket identity should follow the same philosophy:
- frontmatter is the structured contract
- filenames/directories are helpful conventions but not authoritative

The new API should not force callers to re-implement “try parse, else keep error”.

#### `paths.Resolver` — “Identity affects path anchoring; pick one or you get mismatches”

Ticket identity decisions affect how we resolve doc paths and related file paths. If ticket selection can come from directory inference, then path anchors need to remain consistent: doc-path anchor should not depend on how ticket ID was derived.

I’m fine with “frontmatter authoritative + directory fallback”, but the API must return enough context to normalize paths consistently (e.g., always provide `DocPath` and `DocsRoot`).

#### `pkg/commands/*` (bloc) — “Give me predictable outcomes and actionable errors”

From a command author’s perspective, the biggest pain is that “ticket not found” can mean:
- truly no such ticket
- index missing
- index frontmatter invalid
- duplicate ticket IDs across directories

We need:
- an identity policy
- error types that let commands present actionable messages
- helpers for common operations (index doc path, ticket dir, enumerate docs)

### Rebuttals (Round 2)

#### Mara → Jon

I’m onboard with directory inference as a fallback, but not as a parallel “guess API” that every command can misuse. If we make inference easy, we’ll get inconsistent behavior across verbs again. Better: **one `TicketByID` that returns a handle + status**, and commands can choose strictness *by checking the handle’s state*.

#### Jon → Mara

Fair. The key is explicitness. If `TicketByID` returns a handle that can be `Status=Inferred` vs `Status=Verified`, I can live with that. But I want it to be obvious at the call site whether we’re okay operating on inferred tickets.

#### `CollectTicketWorkspaces` → everyone

If you want “ticket dir exists but index missing” in the main lookup flow, reuse `CollectTicketScaffoldsWithoutIndex` (doctor already does) or integrate it into the repository layer. Just don’t bolt directory inference into unrelated command codepaths.

#### `pkg/commands/*` → `CollectTicketWorkspaces`

We need a central function that doesn’t hide broken tickets. We’re okay if “list tickets” defaults to only valid tickets, but other commands (repair/move/doctor) need visibility into broken states with actionable errors.

#### `documents.WalkDocuments` → `pkg/commands/*`

If you standardize on handles that include `(doc, body, parseErr)`, you can avoid special-casing all these situations. It’s the same pattern everywhere: parse when possible; keep error otherwise; don’t silently drop.

### Moderator Summary

**Areas of agreement**
- Frontmatter `Ticket` is the **primary authoritative identity** when it parses.
- Directory naming (`TICKET--slug`) is a **strong convention**, useful for fallback labeling and repairs.
- The new API must **stop collapsing** missing/invalid index into “ticket not found”.

**Open design decisions**
- Do we represent ticket identity as:
  - strictly “frontmatter only”, and treat anything else as invalid?
  - “frontmatter preferred, directory inference permitted as `Status=Inferred`”?
- How do we handle duplicates (two directories with same ticket ID)?
  - error with “multiple matches”
  - pick one with a deterministic rule + emit diagnostic

**Proposed next step (for the next debate round)**
- Draft a `TicketHandle` / `TicketIdentityStatus` sketch and sanity-check it against 3 verbs:
  - `docmgr add`
  - `docmgr relate --ticket`
  - `docmgr doctor --ticket`

## Usage Examples

If you want to continue debating this question, the next round can be:

- “Round 3: propose 2–3 concrete identity policies and run them through real edge cases”

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/01-debate-candidates-repository-lookup-ticket-finding.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`
