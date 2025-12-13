---
Title: 'Code Review: Contestant #1 DJ Skippy (Skip Policy)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/skip_policy.go
      Note: Code under review
    - Path: internal/workspace/skip_policy_performance_test.go
      Note: Runtime behavior evidence (performance-stage tests)
    - Path: internal/workspace/skip_policy_test.go
      Note: Unit tests referenced
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Spec authority (§6)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T20:24:13.157175665-05:00
---


# Code Review: Contestant #1 — DJ Skippy (Skip Policy)

## Scope

This review covers the ingest-time skip policy and path tagging logic:

- `internal/workspace/skip_policy.go`
  - `DefaultIngestSkipDir` (directory skip predicate)
  - `ComputePathTags` (doc path tagging)
  - helpers: `containsPathSegment`, `hasSiblingIndex`, `isControlDocBase`

It also evaluates:

- spec adherence (Design Spec §6)
- runtime behavior as observed in:
  - `internal/workspace/skip_policy_test.go`
  - `internal/workspace/skip_policy_performance_test.go`

## Executive Summary

**Verdict:** ✅ Ship (minor documentation improvements recommended)  
**Overall quality:** High. The implementation is small, idiomatic, and has strong test coverage for the tricky cases.

### Strengths

- **Correctness**: Matches spec intent for skip rules and tagging; avoids substring false positives (`archive` vs `myarchive`).
- **Simplicity**: Minimal branching, clean helpers, no over-abstraction.
- **Test coverage**: Unit tests + performance-stage tests cover high-risk edge cases (control-doc detection and segment boundaries).

### Key Risks / Gaps

- **Assumption documentation**: Some behavior relies on external contracts (e.g. `filepath.WalkDir` provides non-nil `fs.DirEntry`) and conservative error behavior (`os.Stat` failures treated as “no sibling index”). These choices are fine, but should be documented to prevent future “simplifications”.
- **Spec traceability**: Comments cite Spec §6 generally, but do not consistently point to exact sub-sections/decisions (optional improvement).

## Spec Mapping (Design Spec §6)

### Directories (Spec §6.1)

- **Skip `.meta/` entirely**: ✅ Implemented by `name == ".meta"` in `DefaultIngestSkipDir`.
- **Skip underscore dirs (`_*/`) entirely**: ✅ Implemented by `strings.HasPrefix(name, "_")`.
- **Include `archive/` but tag** `is_archived_path=1`: ✅ Implemented by `containsPathSegment(...,"archive")` in `ComputePathTags`.
- **Include `scripts/` but tag** `is_scripts_path=1`: ✅ Implemented by `containsPathSegment(...,"scripts")`.
- **Include `sources/` but tag** `is_sources_path=1`: ✅ Implemented by `containsPathSegment(...,"sources")`.

### Control docs at ticket root (Spec §6.2)

- **Include** `README.md`, `tasks.md`, `changelog.md` but **tag** `is_control_doc=1` and default-hide: ✅ Implemented by `isControlDocBase` + `hasSiblingIndex`.
- **Ticket root detection**: Implementation interprets “ticket root” as “directory contains `index.md`” and uses sibling `index.md` as marker. This matches docmgr conventions and avoids false positives like `sources/README.md`.

## Runtime Behavior Evidence

### Unit tests (`skip_policy_test.go`)

Coverage includes:

- `.meta` and `_templates/_guidelines` skipped at directory level (`TestDefaultIngestSkipDir`).
- Control docs require sibling `index.md` (`TestComputePathTags_ControlDocsRequireSiblingIndex`).
- Segment checks for `archive/scripts/sources` are boundary-safe (`TestComputePathTags_PathSegments`).

### Performance-stage tests (`skip_policy_performance_test.go`)

Observed:

- Segment boundary behavior is explicitly tested:
  - `archive/…` tags archived
  - `myarchive/…` does not
  - `scripts/…` tags scripts
  - `scripts-old/…` does not
- Control docs tagged at root but not in nested directories
- Grand Finale simulates a realistic tree and prints a JSON report of decisions

This is unusually good for reviewability: it doubles as living documentation of intended semantics.

## Code Review Notes (Implementation)

### `DefaultIngestSkipDir`

**Good**
- Direct and correct.
- Matches Spec §6.1 precisely.

**Doc improvement (recommended)**
- Add a short comment clarifying that `d` is expected non-nil in the intended usage (WalkDir contract).

### `ComputePathTags`

**Good**
- Uses `filepath.ToSlash(filepath.Clean(docPath))` before segment checks.
- Uses boundary-safe `containsPathSegment` to avoid substring false positives.
- Uses `strings.EqualFold` to detect `index.md` (case-insensitive).
- Control doc detection avoids nested `README.md` by requiring sibling `index.md`.

**Doc improvement (recommended)**
- Add a comment explaining:
  - why sibling `index.md` is the ticket-root marker,
  - and why `os.Stat` errors are treated as “no sibling index” (conservative fallback).

### `containsPathSegment`

**Good**
- Correct for the bug class it prevents.

**Doc improvement (recommended)**
- Add a short “why” comment that explicitly calls out the `myarchive`/`archive` class of false positives.

## Recommendations / Action Items

### 1) Add “WHY” comments at decision points (recommended)

No behavior change; just harden the component against future refactors:

- In `DefaultIngestSkipDir`: clarify `filepath.WalkDir`/`fs.DirEntry` contract assumption
- In `hasSiblingIndex`: clarify conservative semantics of `os.Stat` errors
- In `containsPathSegment`: explain boundary matching rationale

### 2) Optional: increase spec traceability

If desired, change comments from “Spec §6” to specific “Spec §6.1 / §6.2” at the exact functions, and/or reference the specific decisions listed in the spec preface.

## Ship Checklist (for this component)

- [x] Matches spec intent for `.meta/` and `_*/` skip rules
- [x] Tags `archive/`, `scripts/`, `sources/` based on segment boundaries
- [x] Control docs tagged only at ticket root (sibling `index.md` marker)
- [x] Unit tests cover core logic
- [x] Performance-stage tests cover edge cases + provide a human-readable trace
- [ ] Add “WHY” comments to harden against future refactors (recommended but not blocking)

## Related

- Design spec: `design/01-workspace-sqlite-repository-api-design-spec.md` (§6)
- Baseline tests: `internal/workspace/skip_policy_test.go`
- Performance stage: `internal/workspace/skip_policy_performance_test.go`
- Talent show framework: `reference/16-talent-show-candidates-code-performance-review.md`
- Judge panel: `reference/18-the-jury-panel-judge-personas-and-judging-criteria.md`
- Jury deliberation transcript: `reference/19-jury-deliberation-contestant-1-dj-skippy-skip-policy.md`
