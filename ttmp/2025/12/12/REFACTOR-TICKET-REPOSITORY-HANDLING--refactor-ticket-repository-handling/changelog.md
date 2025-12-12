# Changelog

## 2025-12-12

- Initial workspace created


## 2025-12-12

Created analysis mapping ticket discovery + ticket/doc lookup logic; identified duplication/inconsistencies and proposed a central repository abstraction.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md — Analysis doc


## 2025-12-12

Debate Round 1 completed (Question 6): ticket identity semantics (frontmatter vs directory vs fallback) with research-backed arguments.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/03-debate-round-1-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md — Debate round doc


## 2025-12-12

Debate Round 2 (Q7) completed: modeled lookup scopes (repo/ticket/doc + reverse lookup) with research-backed arguments.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/04-debate-round-2-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md — Debate round doc


## 2025-12-12

Debate Round 3 (Q8) completed: proposed clean boundary between lookup context and vocabulary/policy validation.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/05-debate-round-3-q8-how-do-we-keep-vocabulary-config-concerns-from-leaking-everywhere.md — Debate round doc


## 2025-12-12

Decision recorded for Q7 debate: prefer single QueryDocs(ctx, scope, filters...) API entry point.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/04-debate-round-2-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md — Recorded API direction choice (A)


## 2025-12-12

Debate Round 4 (Q11) completed: QueryDocs(ctx, scope, filters...) design trade-offs (request/response shape, parse policy, reverse lookup modeling).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/06-debate-round-4-q11-design-querydocs-ctx-scope-filters.md — Debate round doc


## 2025-12-12

Debate Round 6 (Q6): Explored ticket identity semantics (frontmatter vs directory name). Key findings: frontmatter is authoritative, but conflict detection and missing index.md handling need policy decisions. Candidates agreed on repository wrapper layer but disagreed on strictness vs repair workflows.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/06-debate-round-6-q6-what-is-a-ticket-id-vs-directory-vs-index-frontmatter.md — Debate round document


## 2025-12-12

Debate Rounds 7-8: Explored scope modeling (Q7) and vocabulary/config boundaries (Q8). Round 7: Candidates disagreed on enum vs separate methods vs optional fields for scope API. Round 8: Consensus on vocabulary-agnostic repository, disagreement on validation layer design (separate Validator vs optional methods vs VocabularyService).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/07-debate-round-7-q7-how-should-we-model-scope-in-lookups-repo-vs-ticket-vs-doc.md — Round 7
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/08-debate-round-8-q8-how-do-we-keep-vocabulary-config-concerns-from-leaking-everywhere.md — Round 8


## 2025-12-12

Debate Round 9 (Q11): Designed QueryDocs API signature. Explored structured request/response vs minimal signature vs optional fields. Key decisions needed: DocHandle contract (agreed: path, doc, body, readErr), error policy (skip vs return-with-error), body loading (toggle vs always-empty), ordering API, Repository context holding. Candidates agreed on using WalkDocuments internally and paths.Resolver for reverse lookup.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/09-debate-round-9-q11-design-querydocs-ctx-scope-filters.md — Round 9


## 2025-12-12

Debate Round 10 (Q1): Defined Repository object responsibilities. Explored type name (Repository vs Workspace), state management (caching vs delegation), and method signatures. Key decisions needed: package location (new internal/repository vs extend workspace), caching policy (internal vs explicit vs none), and API complexity (structured request vs simple options). Candidates agreed on holding resolved context (root, configDir, repoRoot) and providing resolver factory.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/10-debate-round-10-q1-what-is-the-repository-object-and-what-does-it-own.md — Round 10


## 2025-12-12

Debate Round 11 (Q2): Explored broken and partial state representation. Current state: TicketWorkspace includes FrontmatterErr, commands handle inconsistently (skip vs fail vs report), WalkDocuments contract includes readErr. Key decisions needed: error representation (enum vs error types vs sentinel errors vs simple fields), handle vs error return, missing index detection, orphaned doc handling. Candidates agreed on distinguishing missing index vs invalid frontmatter, and matching WalkDocuments contract for DocHandle.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/11-debate-round-11-q2-how-should-we-represent-broken-and-partial-states.md — Round 11


## 2025-12-12

Debate Round 12 (Q3): Explored filter and enumeration semantics. Current state: ticket filter inconsistent (substring vs exact), skip rules inconsistent (string contains vs prefix vs ignore patterns), document enumeration differs (skip vs return-with-error). Key decisions needed: ticket filter semantics (exact vs substring vs configurable), skip rules (hardcoded vs configurable vs .docmgrignore), document enumeration policy (skip vs return-with-error, include vs exclude index.md). Candidates agreed on unifying skip rules and using WalkDocuments/CollectTicketWorkspaces as enumeration primitives.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/12-debate-round-12-q3-what-are-the-semantics-of-filters-and-enumeration.md — Round 12


## 2025-12-12

Design log started: Recording interactive Repository API decisions (workspace.Workspace). Decisions so far: extend workspace package; support both discovering + injected constructors.

### Related Files

- ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/13-design-log-repository-api.md — Design log


## 2025-12-12

Debate Round 13: Considered SQLite-backed in-memory index for lookup/reverse lookup. Key takeaway: keep API semantic; SQLite enables joins for reverse lookup + deterministic ordering, but normalization still required. Open questions: reverse lookup as Scope vs Filter; whether to add advanced Expr/Where; indexing lifecycle.

### Related Files

- ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/14-debate-round-13-sqlite-index-backend-influences-lookup-and-reverse-lookup.md — Round 13


## 2025-12-12

Design spec drafted: workspace.Workspace SQLite-backed repository lookup API (goals, API, skip rules, diagnostics contract, schema sketch, migration plan).

### Related Files

- ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md — Design spec


## 2025-12-12

Started implementation: created implementation diary; reviewed existing workspace config/discovery + path normalization + document walking contracts.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/config.go — Baseline for WorkspaceContext discovery.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/discovery.go — Baseline for ticket discovery semantics.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/15-diary.md — Implementation diary for this refactor.


## 2025-12-12

Implemented Workspace skeleton: WorkspaceContext + DiscoverWorkspace/NewWorkspaceFromContext + Resolver wiring (best-effort config load).

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/workspace.go — Workspace entry point + construction helpers.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/15-diary.md — Recorded implementation steps.


## 2025-12-12

Added integration-first testing plan; will use existing test-scenarios harness as baseline and extend as QueryDocs lands.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/run-all.sh — Integration scenario baseline for regression comparison.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/02-testing-strategy-integration-first.md — Testing plan for this refactor.


## 2025-12-12

Ran baseline integration scenario suite against system docmgr; scenario completed OK. This run is our pre-refactor behavior reference.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/run-all.sh — Baseline integration test harness (passed on 2025-12-12).
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/15-diary.md — Recorded the exact command + scenario root used for baseline run.


## 2025-12-12

Ran integration scenario suite against locally built refactor docmgr binary (DOCMGR_PATH=/tmp/docmgr-refactor-local-2025-12-12); scenario completed OK.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/run-all.sh — Integration harness (passed against local refactor binary).
- /tmp/docmgr-refactor-local-2025-12-12 — Local built binary under test.


## 2025-12-12

Implemented in-memory SQLite bootstrap + Workspace index schema (docs/doc_topics/related_files) per spec; added schema smoke test.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/sqlite_schema.go — Schema DDL + in-memory DB open.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/sqlite_schema_test.go — Schema creation smoke test.


## 2025-12-12

Implemented canonical ingest-time skip rules + path tagging helpers (skip .meta and _*/; tag archive/scripts/sources/control-docs/index). Added unit tests.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/skip_policy.go — Skip rules + tags used by ingestion.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/skip_policy_test.go — Skip/tagging unit tests.


## 2025-12-12

Implemented Workspace in-memory index ingestion (InitIndex): walk docs, parse frontmatter, store docs/topics/related_files with parse_ok/parse_err and path tags; added unit test.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/index_builder.go — Workspace index ingestion.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/index_builder_test.go — Ingestion smoke test.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/workspace.go — Workspace now owns DB handle and InitIndex.


## 2025-12-12

Added workspace export-sqlite command to export the in-memory index to a SQLite file. Exported DB includes a README table populated from embedded pkg/doc/*.md so the DB is self-describing. Added scenario smoke test and ran it successfully.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/sqlite_export.go — README table + VACUUM INTO export implementation.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/commands/workspace_export_sqlite.go — New CLI verb (classic Run) for exporting sqlite.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/pkg/doc/embedded_docs.go — Reads embedded docs for README table.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/test-scenarios/testing-doc-manager/19-export-sqlite.sh — Scenario smoke test for export-sqlite.


## 2025-12-12

Expanded RelatedFiles normalization: persist canonical + repo/docs/doc/abs + clean keys in sqlite for reliable reverse lookup; added helper + stronger ingestion test.

### Related Files

- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/index_builder.go — Persist norm_* columns.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/index_builder_test.go — Normalization assertions.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/normalization.go — Normalization helper.
- /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/internal/workspace/sqlite_schema.go — Schema columns + indexes.

