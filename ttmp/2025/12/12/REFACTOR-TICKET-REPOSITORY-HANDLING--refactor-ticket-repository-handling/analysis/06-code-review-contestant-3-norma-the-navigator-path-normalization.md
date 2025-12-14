---
Title: 'Code Review: Contestant #3 Norma the Navigator (Path Normalization)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/paths/resolver.go
      Note: |-
        Primary code reviewed (normalization + matching)
        Code reviewed
    - Path: internal/workspace/index_builder.go
      Note: Calls normalization when ingesting RelatedFiles (writes related_files rows)
    - Path: internal/workspace/normalization.go
      Note: Workspace wrapper that persists normalized keys for RelatedFiles
    - Path: test-scenarios/testing-doc-manager/14-path-normalization.sh
      Note: |-
        Integration evidence (CLI path-form reconciliation)
        Evidence used in review
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Integration harness used for evidence capture
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/21-how-to-judge-contestant-3-norma-the-navigator-path-normalization.md
      Note: |-
        Judging rubric / how-to
        Judging rubric
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/22-jury-deliberation-contestant-3-norma-the-navigator-path-normalization.md
      Note: |-
        Jury deliberation transcript backing this review
        Deliberation basis
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T20:54:47.597756289-05:00
---


## Executive Summary

**Verdict:** ✅ **SHIP** (minor documentation follow-ups; consider future tightening of substring fallback)  
**Aggregate jury score:** **9.0/10**  
**Critical issues:** None found during scenario evidence run  
**Main risk:** potential false positives from substring-based matching fallback (currently last-resort, but worth documenting and possibly constraining later)

Norma’s normalization/matching logic is doing exactly what users need: different spellings of the same path (doc-relative, docs-root-relative, absolute, basename) can still find the correct documents. This is proven by scenario-level evidence (`14-path-normalization.sh`) and supported by the resolver’s multi-anchor strategy.

## Scope

**Reviewed code:**

- `internal/paths/resolver.go`
  - `Resolver.Normalize` (anchor order and existence-first selection)
  - `NormalizedPath.Representations`, `Suffixes`, `Best`, `Empty`
  - `MatchPaths` (intersection → suffix → substring)
  - `DirectoryMatch` (prefix match across representations)
- `internal/workspace/normalization.go`
  - `normalizeRelatedFile` (persistence envelope)

**Evidence executed:**

- Scenario suite `run-all.sh` including `14-path-normalization.sh` at root `/tmp/docmgr-scenario-norma`
- Ingestion sanity: `go test ./internal/workspace -run TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles -v` (asserts normalization envelope for a representative RelatedFiles entry)

## What Norma Does (high-level)

Norma is the system’s “path bilingual interpreter.” She:

1. Accepts a raw user-provided path string (may be abs/rel/doc-relative/dirty/tilde/etc).
2. Tries to resolve it against a set of anchors (repo, doc, config, docs-root, docs-parent).
3. Produces multiple representations (repo-relative, docs-relative, doc-relative, abs, canonical).
4. Enables best-effort matching for reverse lookup and directory filters.

## Runtime Evidence (what we ran)

### Integration-stage proof (high signal)

Commands:

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr
go build -o /tmp/docmgr-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-norma
```

What scenario #14 verifies:

- The same doc is discoverable by `doc search --file` across:
  - doc-relative input
  - docs-root-relative input
  - absolute input
  - basename-only input (`register.go`)

Observed excerpt (shortened from scenario output):

```text
... file=backend/chat/api/register.go note=Doc-relative path reference
... file=../../../../../backend/chat/api/register.go, ../backend/chat/api/register.go note=Doc-relative path reference (deep traversal) | Ttmp-relative path reference (shallower traversal)
... file=backend/chat/ws/manager.go note=Absolute path reference (host-specific)
```

### Unit-stage supporting evidence (sanity)

Command:

```bash
go test ./internal/workspace -run TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles -count=1 -v
```

What it proves (relevant to Norma):

- `related_files` rows contain canonical + fallback normalization keys and `anchor`.

## Strengths

- **Multi-anchor strategy is explicit and predictable**: repo → doc → config → docs-root → docs-parent.
- **Existence-first selection**: returns the first anchor that resolves to an existing path; otherwise returns deterministic fallback.
- **Multiple representations** reduce “path spelling” brittleness across commands.
- **Scenario-level evidence covers real user workflow** (doc relate + doc search).
- **Normalization logic is centralized** (one resolver to reason about rather than per-command hacks).

## Risks / Concerns

### 1) Substring matching fallback may cause false positives

`MatchPaths` tries substring intersection as the last resort. This enables basename-only convenience, but can match unintended targets in pathological cases.

Mitigations present today:

- substring is the **third** tier (after exact representation intersection and suffix matching),
- all values are normalized (lowercase + slash), reducing trivial mismatches.

Recommendation: document this trade-off explicitly and consider narrowing substring usage (see below).

### 2) Anchor order is a core semantic contract (but not explained inline)

The chosen anchor order materially affects canonicalization. Without “WHY” comments, a future refactor could reorder anchors and subtly change behavior.

## Recommendations (small, targeted)

### 1) Add “WHY” comments in `Resolver.Normalize` (recommended)

Add a short comment explaining the anchor order rationale (repo-relative stability, doc-relative support, config/docs fallbacks).

### 2) Document match tiers and substring intent in `MatchPaths` (recommended)

Add a comment documenting:

- tier 1: representations intersection (preferred),
- tier 2: suffix matching (up to 3 segments),
- tier 3: substring fallback (primarily to support basename-only convenience; with known risk).

### 3) Consider constraining substring fallback (optional follow-up)

Potential refinement:

- only allow substring fallback when the query looks like a basename (no slashes), or
- only allow substring fallback when the query length is above a threshold, or
- remove substring fallback and add explicit basename indexing (more work).

Not required to ship now, but worth tracking.

## Ship Checklist

- [x] Scenario #14 path normalization passes in full harness run
- [x] Ingestion test asserts normalization envelope fields exist
- [ ] Add “WHY” comments around anchor order and match tiers (recommended follow-up)
