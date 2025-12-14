---
Title: "Docs update guide (pkg/doc) — REFACTOR-TICKET-REPOSITORY-HANDLING"
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
    - docs
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/doc/docmgr-cli-guide.md
      Note: Add `workspace export-sqlite` command + QueryDocs/normalization notes
    - Path: pkg/doc/docmgr-how-to-use.md
      Note: Update search/relate/doctor/list semantics to match workspace.QueryDocs + normalization tiers
    - Path: pkg/doc/docmgr-doctor-validation-workflow.md
      Note: Algorithmic walkthrough currently describes pre-refactor discovery; update to index-backed flow
    - Path: pkg/commands/workspace_export_sqlite.go
      Note: New CLI surface to document (export workspace index to sqlite)
    - Path: pkg/commands/search.go
      Note: Search now discovers workspace + builds index + uses QueryDocs for metadata + reverse lookup
    - Path: pkg/commands/list_docs.go
      Note: List docs now uses Workspace.QueryDocs (index-backed), emits diagnostics
    - Path: pkg/commands/doctor.go
      Note: Doctor now builds one workspace index + uses QueryDocs and post-filters ignore globs
    - Path: pkg/commands/relate.go
      Note: Relate now resolves ticket/doc via QueryDocs and normalizes paths via paths.Resolver
    - Path: internal/workspace/workspace.go
      Note: Workspace entry point (DiscoverWorkspace/NewWorkspaceFromContext/InitIndex/QueryDocs)
    - Path: internal/workspace/index_builder.go
      Note: Ingestion pipeline (skip policy + frontmatter parse + store docs/topics/related_files)
    - Path: internal/workspace/query_docs.go
      Note: QueryDocs semantics + diagnostics emission + hydration
    - Path: internal/workspace/skip_policy.go
      Note: Canonical ingest-time skip rules/tagging (archive/scripts/control docs, .meta, _*/ etc)
    - Path: internal/workspace/normalization.go
      Note: RelatedFiles normalization + fallback matching keys
    - Path: internal/paths/resolver.go
      Note: Path normalization used for reverse lookup and relate operations
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-14T00:00:00Z
---

# Docs update guide (pkg/doc) — REFACTOR-TICKET-REPOSITORY-HANDLING

This is a short guide for updating the embedded documentation in `pkg/doc/` after the “ticket repository handling” refactor landed (Workspace + index-backed QueryDocs).

## What changed (user-visible)

- **Index-backed operations**: `doc search`, `list docs`, `doctor`, and `doc relate` now run through:
  - `workspace.DiscoverWorkspace(...)`
  - `Workspace.InitIndex(...)`
  - `Workspace.QueryDocs(...)`
- **Reverse lookup is more consistent**: matching for `--file` / `--dir` is driven by `paths.Resolver` normalization plus a small set of compatibility fallbacks (e.g., basename/suffix match for `register.go` style queries).
- **New command**: `docmgr workspace export-sqlite` exports the in-memory workspace index to a SQLite file. The exported DB includes a README table populated from embedded `pkg/doc/*.md`, so keeping these docs current matters.

## Primary sources (in this ticket)

- **Changelog**: `changelog.md` (lists the implemented features and points to the relevant code files)
- **Behavior parity**: `analysis/11-comparison-suite-report-system-vs-local.md` (scenario-level parity for the common subset; notes output diffs due to timestamps/order)
- **Design spec**: `design/01-workspace-sqlite-repository-api-design-spec.md` (contracts/semantics)

## Documentation files to update (what + how)

### 1) `pkg/doc/docmgr-cli-guide.md`

- **Add a new section** documenting `docmgr workspace export-sqlite`.
- **Include examples** (these match the implementation in `pkg/commands/workspace_export_sqlite.go`):
  - `docmgr workspace export-sqlite --out /tmp/docmgr-index.sqlite`
  - `docmgr workspace export-sqlite --out /tmp/docmgr-index.sqlite --force`
  - `docmgr workspace export-sqlite --out /tmp/docmgr-index.sqlite --include-body`
  - `docmgr workspace export-sqlite --root ttmp --out /tmp/docmgr-index.sqlite`
- **Explain what’s inside** (tables + README embedded docs) and why it’s useful (debugging/sharing a snapshot of workspace state).

### 2) `pkg/doc/docmgr-how-to-use.md`

- **Search section**: add a short “How reverse lookup matches paths” note:
  - `docmgr doc search --file` uses path normalization (doc/repo/root-aware), then falls back to basename/suffix matching for convenience.
  - Recommend **using repo-relative paths** where possible, but mention that absolute paths can work (depending on resolver context).
- **Relate section**: add one line that relate uses the same normalization pipeline as the index, so reverse lookup matches what relate writes.
- **Optional (small)**: add a “Performance/architecture” callout that these commands build an ephemeral in-memory index for consistent behavior.

### 3) `pkg/doc/docmgr-doctor-validation-workflow.md` (IMPORTANT: likely stale)

This doc is an algorithmic walkthrough and currently describes **pre-refactor discovery**.

- **Update “Input Discovery”** to reflect current behavior:
  - `doctor` discovers workspace via `workspace.DiscoverWorkspace`, builds an index (`InitIndex`), then queries via `QueryDocs`.
  - The canonical skip policy is applied at ingestion time (`internal/workspace/skip_policy.go`).
  - `.docmgrignore` and `--ignore-glob` are still supported; `doctor` applies them as a **post-filter** over QueryDocs results to preserve legacy behavior.
- **Add code pointers**: `pkg/commands/doctor.go`, `internal/workspace/index_builder.go`, `internal/workspace/query_docs.go`, `internal/workspace/skip_policy.go`.

### 4) `pkg/doc/docmgr-diagnostics-and-rules.md` (likely small update)

- **Add a note** that QueryDocs can emit diagnostics for:
  - parse-error documents (when diagnostics are requested)
  - normalization fallback matches
- Point at `pkg/diagnostics/docmgrctx/query_docs.go` as the taxonomy source.

### 5) `pkg/doc/docmgr-ci-automation.md` (optional)

- Add an optional debugging pattern: export the workspace index as a CI artifact via `docmgr workspace export-sqlite`.
- If desired, mention that the exported DB includes embedded docs (README table), which can help offline triage.

## Quick “doc update checklist” (for this refactor)

- [x] Add `workspace export-sqlite` to `docmgr-cli-guide.md`
- [x] Fix `docmgr-doctor-validation-workflow.md` to match index-backed doctor flow
- [x] Add a brief reverse-lookup normalization note to `docmgr-how-to-use.md`
- [x] Update diagnostics/CI docs with QueryDocs diagnostics + export-sqlite usage


