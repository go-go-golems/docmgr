---
Title: Code review guide (senior)
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - workspace
    - sqlite
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/query_docs.go
      Note: QueryDocs API implementation under review.
    - Path: internal/workspace/query_docs_sql.go
      Note: DocQuery→SQL compiler under review.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md
      Note: Design spec referenced by this review guide.
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/15-diary.md
      Note: Implementation diary referenced by this review guide.
ExternalSources: []
Summary: Guide for a senior engineer to critically review the Workspace+SQLite refactor implementation against the design spec.
LastUpdated: 0001-01-01T00:00:00Z
---


## Why this doc exists

This ticket is refactoring the “repository lookup” plumbing of `docmgr` to centralize doc/ticket discovery and querying behind a new `internal/workspace.Workspace` API, backed by an in-memory SQLite index.

This document is a **review map**: what changed, where to look, what spec constraints apply, and what to challenge.

Primary implementation narrative lives in:
- `reference/15-diary.md`
- `design/01-workspace-sqlite-repository-api-design-spec.md`

## Current integration status (important for scope)

As of now, the new backend is **not yet wired into normal verbs** like `doc list`, `doc search`, `doctor`, `relate`.

The only user-facing integration so far is:
- `docmgr workspace export-sqlite` (debugging verb)

The porting of normal verbs begins with Spec §11.2 (ticket Tasks [9–12]).

## The spec this work is implementing

High-signal spec anchors to review against:

- **Construction**: Spec **§5.1**
  - `DiscoverWorkspace` / `NewWorkspaceFromContext`
  - invariants: Root/RepoRoot/ConfigDir
- **Canonical skip rules**: Spec **§6**
  - skip `.meta/` + `_*/`
  - tag `archive/`, `scripts/`, `sources/`, control docs, `index.md`
- **Indexing / parse-state**: Spec **§7.1** + **§9.1–§9.2**
  - ingest invalid docs as `parse_ok=0` with `parse_err`
- **Path normalization & reverse lookup**: Spec **§7.3** + **§12.1**
  - use `paths.Resolver` and persist multiple comparable representations
- **QueryDocs API**: Spec **§5.2** + **§10.1–§10.4**
  - structured request/response, scope + filters + options
  - reverse lookup as filters compiled to SQL, not a new scope kind
  - contradictory queries are hard errors
- **Diagnostics**: Spec **§10.6 / §8**
  - planned, but not fully implemented yet (see “Known gaps”)

## What I implemented (by subsystem) + where to review

### 1) Workspace “front door”

- **Files**
  - `internal/workspace/workspace.go`
- **Goal**
  - Establish a single canonical entry point that owns anchors for normalization and (later) query/diagnostics.
- **Spec mapping**
  - Spec §5.1
- **Review focus**
  - Discovery semantics and best-effort behavior (there’s an open question in the spec about strictness).
  - Invariants: Root/ConfigDir/RepoRoot must be stable across commands.

### 2) SQLite schema + pragmas

- **Files**
  - `internal/workspace/sqlite_schema.go`
  - `internal/workspace/sqlite_schema_test.go`
- **Goal**
  - Minimal stable schema to support filtering + reverse lookup as SQL joins.
- **Spec mapping**
  - Spec §9.1–§9.2
- **Review focus**
  - Schema columns align with the query/compiler requirements (especially `related_files`).
  - Pragmas: correctness vs performance trade-offs.
  - Index coverage (ticket_id, parse_ok, norm_* lookup columns).

### 3) Canonical skip policy + path tags

- **Files**
  - `internal/workspace/skip_policy.go`
  - `internal/workspace/skip_policy_test.go`
- **Goal**
  - Define “what is indexed” once; tag categories so query-time defaults can hide noise without dropping data.
- **Spec mapping**
  - Spec §6
- **Review focus**
  - Directory skip correctness and potential false positives/negatives.
  - Tag correctness (especially `is_control_doc` policy).

### 4) Index builder / ingestion

- **Files**
  - `internal/workspace/index_builder.go`
  - `internal/workspace/index_builder_test.go`
- **Goal**
  - Walk docs once, parse frontmatter, store rows in SQLite.
  - Persist parse state (`parse_ok/parse_err`) instead of silently dropping broken docs.
- **Spec mapping**
  - Spec §6, §7.1, §9.2
- **Review focus**
  - Transactionality: ingest uses a single transaction.
  - Path stored in `docs.path`: absolute, cleaned, slash-normalized.
  - Frontmatter field mapping: `ticket/docType/status/intent/title/last_updated`.
  - Broken-state semantics: parse failures become `parse_ok=0`, with error surfaced later by query options.

### 5) RelatedFiles normalization envelope

- **Files**
  - `internal/workspace/normalization.go`
  - `internal/paths/resolver.go`
- **Goal**
  - Normalize `RelatedFiles` robustly and persist multiple keys so reverse lookup isn’t brittle.
- **Spec mapping**
  - Spec §7.3, §12.1
- **Review focus**
  - Correct anchor usage: ingestion uses a resolver with `DocPath` set (doc-relative normalization).
  - Persisted keys: `norm_canonical`, `norm_repo_rel`, `norm_docs_rel`, `norm_doc_rel`, `norm_abs`, `norm_clean`, plus `anchor/raw_path`.
  - “Canonical” meaning: should match `paths.NormalizedPath.Canonical` semantics (repo-rel preferred).

### 6) Query API (QueryDocs) + SQL compiler

- **Files**
  - `internal/workspace/query_docs.go`
  - `internal/workspace/query_docs_sql.go`
  - `internal/workspace/query_docs_test.go`
- **Goal**
  - Provide a single doc lookup API for future ports of `list docs`, `search`, etc.
  - Express reverse lookup via SQL `EXISTS (...)` against `related_files`.
- **Spec mapping**
  - Spec §5.2 and §10.1–§10.4
- **Review focus**
  - **Defaults**:
    - default-hide `is_archived_path`, `is_scripts_path`, `is_control_doc`
    - default-exclude parse errors unless `IncludeErrors=true`
  - **Scope semantics**:
    - `ScopeTicket` is `WHERE d.ticket_id=?`
    - `ScopeDoc` resolves to absolute and matches `d.path=?` (strict)
  - **Filter semantics**:
    - `TopicsAny`: OR semantics via `EXISTS` on `doc_topics`
    - `RelatedFile`: OR semantics against multiple persisted `related_files.*` columns (exact match set)
    - `RelatedDir`: OR semantics using `LIKE prefix/%` against multiple persisted columns
  - **Hard errors**:
    - contradictory `ScopeTicket` + `Filters.Ticket` mismatch => error (per spec)
  - **SQL hygiene**:
    - placeholders used for all values (no string concatenation of user input)
    - ensure no accidental Cartesian products (uses `EXISTS`, not joins)
  - **Hydration choices**:
    - Query returns `DocHandle` with `models.Document` reconstructed from stored columns.
    - Topics and RelatedFiles are hydrated best-effort with extra queries per doc (OK for now; review whether we should batch later).

### 7) Export debugging artifact (export-sqlite)

- **Files**
  - `internal/workspace/sqlite_export.go`
  - `internal/workspace/sqlite_export_test.go`
  - `pkg/commands/workspace_export_sqlite.go`
  - `cmd/docmgr/cmds/workspace/export_sqlite.go`
  - `test-scenarios/testing-doc-manager/19-export-sqlite.sh`
- **Goal**
  - Create a shareable SQLite snapshot using `VACUUM INTO`, with a self-describing `README` table.
- **Spec mapping**
  - Not a core spec requirement; supports developer UX and debugging.
- **Review focus**
  - File overwrite semantics (`--force`) and “no mkdir” policy.
  - Correctness of README population and determinism.

## Known gaps / intentionally deferred (review should call these out)

- **Diagnostics in QueryDocs** (Spec §10.6 / §8):
  - `DocQueryResult.Diagnostics` is currently returned as `nil`.
  - The code returns parse-error docs as `DocHandle{ReadErr: ...}` when `IncludeErrors=true`, but doesn’t yet emit structured taxonomy diagnostics when `IncludeErrors=false` and `IncludeDiagnostics=true`.
- **Full “order by” surface**:
  - Only `path` and `last_updated` are supported (enough for initial ports).
- **FTS/content search**:
  - Not implemented (spec defers FTS; content search stays as post-filter in `search` when ported).

## Senior review checklist (recommended flow)

### A) Sanity + invariants
- [ ] `Workspace` construction matches Spec §5.1; no command should reimplement discovery once ports begin.
- [ ] Docs root + repo root are always absolute/clean and consistent across ingestion and query.
- [ ] `Workspace.DB()` exposure is acceptable for tests/debugging (or should it be package-private later?).

### B) Schema correctness
- [ ] `related_files` columns cover all planned matching modes (exact vs prefix dir match).
- [ ] Indexes exist on the columns used in WHERE/EXISTS predicates.
- [ ] Pragmas are correct for in-memory usage and don’t break expected semantics.

### C) Ingestion correctness
- [ ] Skip rules are applied exactly once and match Spec §6.
- [ ] Parse errors are indexed with `parse_ok=0` + `parse_err` (Spec §7.1).
- [ ] Tags are computed consistently and stored on docs rows.
- [ ] Transactions: failure mid-ingest does not leave partial rows.

### D) Normalization & matching semantics
- [ ] Resolver anchors are set correctly for doc-relative normalization (`DocPath` is set).
- [ ] Persisted keys are sufficient for “same file written differently” cases (Spec §7.3 / §12.1).
- [ ] Dir matching with `LIKE 'prefix/%'` is correct and won’t produce false positives due to missing path-boundary checks beyond `/`.

### E) QueryDocs semantics
- [ ] Default behavior matches spec expectations:
  - default-hide noisy categories, default-exclude parse errors.
- [ ] Contradictory query detection is strict and matches Decision 16/D2.
- [ ] OR semantics for `RelatedFile[]` and `RelatedDir[]` are implemented as intended.
- [ ] The SQL compiler cannot produce invalid SQL for empty lists / empty strings.
- [ ] No SQL injection; all values are passed as parameters.

### F) Performance considerations (don’t over-optimize yet, but flag risks)
- [ ] `QueryDocs` hydrates topics/related files using per-doc queries (N+1); acceptable for now?
- [ ] `relatedDir` matching does multiple `LIKE` checks across multiple columns; acceptable for now?
- [ ] If not acceptable, propose a follow-up: batch hydration or build aggregated JSON in SQL (SQLite `group_concat`) later.

## Suggested “break it” test cases to run during review

- **Path variants**: same file referenced as repo-relative, absolute, and doc-relative (`../..`) should match `RelatedFile`.
- **Directory reverse lookup**: ensure `RelatedDir` doesn’t match `foo/barista` when searching for `foo/bar`.
- **Broken docs**: confirm `IncludeErrors=false` hides them, `IncludeErrors=true` returns them with `ReadErr != nil`.
- **Visibility tags**: ensure `IncludeControlDocs` flips `tasks.md` visibility as intended.


