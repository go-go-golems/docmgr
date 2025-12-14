---
Title: 'How to Judge: Contestant #3 Norma the Navigator (Path Normalization)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/paths/resolver.go
      Note: |-
        Core path resolver (anchors, canonicalization, fuzzy matching)
        Primary normalization + matching logic
    - Path: internal/workspace/index_builder.go
      Note: Calls normalization during ingestion (writes related_files rows)
    - Path: internal/workspace/normalization.go
      Note: |-
        Workspace wrapper that persists normalized RelatedFiles keys
        Workspace normalization wrapper persisted to sqlite
    - Path: internal/workspace/query_docs_sql.go
      Note: |-
        Uses normalized forms in SQL matching / filtering
        Uses normalized keys for file/dir filtering
    - Path: test-scenarios/testing-doc-manager/14-path-normalization.sh
      Note: |-
        End-to-end scenario proving searches work across multiple path forms
        Scenario that exercises path normalization through CLI
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T20:52:59.367116869-05:00
---


# How to Judge: Contestant #3 Norma the Navigator (Path Normalization)

## Goal

Provide a **copy/paste judging procedure** for Contestant #3 (“Norma the Navigator”) so a jury can:

- run Norma’s show (real commands),
- observe whether normalization + matching behave correctly,
- decide ship / needs work with concrete evidence.

## Context

Norma exists because **humans refer to the same file in multiple ways**:

- absolute: `/home/manuel/repo/internal/workspace/index_builder.go`
- repo-relative: `internal/workspace/index_builder.go`
- docs-root-relative: `2025/12/12/TICKET--slug/index.md`
- doc-relative: `../../../../../backend/chat/api/register.go`
- “dirty” forms: `./pkg/../pkg/file.go`, `~/repo/pkg/file.go`, mixed separators

The Workspace refactor stores **multiple normalized keys** for each `RelatedFiles` entry so later operations (reverse lookup, “docs that reference file X”, “docs under dir Y”) can match reliably even when the query uses a different path form than what was originally stored.

Norma spans:

- `internal/paths.Resolver` (anchor search + normalization + fuzzy matching helpers)
- `internal/workspace/normalization.go` (what Workspace persists into sqlite)

Most importantly, Norma is exercised today by **scenario #14**:

- `test-scenarios/testing-doc-manager/14-path-normalization.sh`
- It relates the same file using doc-relative, docs-root-relative, and absolute forms, then proves `doc search --file` finds the same ticket doc across all those inputs (including a basename-only input).

## Quick Reference

### Norma’s core invariants (what must be true)

When normalizing a `RelatedFiles` raw path, we expect:

- **Multiple representations are recorded** (the “normalization envelope”):
  - `norm_canonical` (best-effort canonical key; preference order matters)
  - `norm_repo_rel`
  - `norm_docs_rel`
  - `norm_doc_rel` (may include `../`)
  - `norm_abs`
  - `norm_clean` (cleaned version of raw input; still useful fallback)
  - `anchor` (which base was used: repo/doc/config/docs-root/docs-parent)
- **Canonical preference order** is stable:
  - repo-relative preferred, then docs-root-relative, then doc-relative, then abs.
- **Matching is forgiving but not insane**:
  - exact intersection of representations is a match,
  - otherwise allow suffix matching (up to 3 segments),
  - otherwise allow conservative substring matching.

### What “good” looks like in the CLI

In scenario #14, the same doc must be found when the `--file` flag is provided as:

- doc-relative
- docs-root-relative
- absolute
- basename only (e.g., `register.go`)

If any of these fails, Norma is not doing her job.

### Commands (the judging kit)

#### (A) Fast “real system” proof (recommended): run scenario 14 inside the scenario harness

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr && \
go build -o /tmp/docmgr-local ./cmd/docmgr && \
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-norma
```

Then confirm in the output that scenario #14 ran and did not error.

#### (B) Focused run (only if you already have the mock workspace created)

```bash
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/14-path-normalization.sh /tmp/docmgr-scenario-norma
```

Note: this script expects `/tmp/docmgr-scenario-norma/acme-chat-app` to already exist.

#### (C) Unit-level sanity (optional, not sufficient alone)

If/when unit tests exist for `internal/paths`, run:

```bash
go test ./internal/paths -count=1 -v
```

### Failure triage checklist (what to inspect when something breaks)

- If doc search doesn’t match for doc-relative inputs:
  - check resolver anchors: `DocPath`, `DocsRoot`, `RepoRoot`, `ConfigDir`
- If only absolute paths work:
  - check `repoRoot` detection / `filepath.Rel` guarding in `Resolver.Normalize`
- If basename matching is too permissive/too strict:
  - inspect fuzzy matching rules (`Suffixes(3)` and `containsSubstring`)
- If sqlite has missing norm_* columns:
  - check `internal/workspace/normalizeRelatedFile` and ingestion insert fields

## Usage Examples

### “Judge Mode” checklist (copy/paste into deliberation)

After you run the scenario, record:

1. **Which forms were tested**
   - [ ] doc-relative
   - [ ] docs-root-relative
   - [ ] absolute
   - [ ] basename-only
2. **Observed behavior**
   - [ ] each `doc search --file ...` returns the expected ticket doc
3. **Robustness notes**
   - [ ] works on Linux paths with `/`
   - [ ] handles `..` segments (doc-relative deep traversal)
4. **Maintainability notes**
   - [ ] preference order is documented or obvious
   - [ ] matching strategy has guardrails (avoid false positives)

### Minimal manual reproduction (outside scenarios)

If you want to probe normalization via code rather than scripts:

- create a temporary docs root,
- relate a file with multiple path forms,
- run `doc search --file` for each form,
- confirm results are identical.

(In practice, scenario #14 already does this robustly.)

## Related

- Candidate roster: `reference/16-talent-show-candidates-code-performance-review.md`
- Judge panel: `reference/18-the-jury-panel-judge-personas-and-judging-criteria.md`
- Candidate #2 (ingestion) evidence: `analysis/05-code-review-contestant-2-ingrid-the-indexer-index-builder-initindex.md`
