# Changelog

## 2026-06-08

- Initial workspace created


## 2026-06-08

Created intern-facing design guide for nested .docmgrignore and a shared ignore policy, based on current doctor/workspace/indexing evidence.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/design-doc/01-nested-docmgrignore-and-shared-ignore-policy-implementation-guide.md — Primary implementation guide
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/reference/01-investigation-diary.md — Chronological investigation diary


## 2026-06-08

Adjusted diary prompt formatting so the reMarkable PDF bundle renders cleanly without LaTeX interpreting literal backslash-n sequences.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/reference/01-investigation-diary.md — Prompt context formatting for PDF rendering


## 2026-06-08

Pivoted the ignore-system guide to a direct workspace-owned go-gitignore cutover and expanded tasks.md into phase-by-phase implementation checklist.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/design-doc/01-nested-docmgrignore-and-shared-ignore-policy-implementation-guide.md — Updated direct go-gitignore implementation plan
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/reference/01-investigation-diary.md — Recorded design pivot and implementation setup
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/tasks.md — Detailed phase-by-phase implementation checklist


## 2026-06-08

Implemented internal/ignore with go-gitignore-backed matching, built-in defaults, root/docs/nested .docmgrignore coverage, and focused tests.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/go.mod — Added go-gitignore dependency
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/ignore/ignore.go — New go-gitignore-backed matcher wrapper
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/ignore/ignore_test.go — Matcher semantics and nested .docmgrignore tests
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/tasks.md — Checked off completed package and matching tasks


## 2026-06-08

Made Workspace own the ignore matcher and pruned ignored directories during index ingestion before frontmatter parsing.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/internal/ignore/ignore.go — Root-path guard for go-gitignore Absolute
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/workspace/index_builder.go — Index-time pruning with DefaultIngestSkipDir plus ignore matcher
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/workspace/index_builder_test.go — Workspace matcher and ignored dependency pruning tests
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/workspace/workspace.go — Workspace-owned ignore matcher construction and accessor
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/tasks.md — Checked off workspace ownership and index pruning tasks


## 2026-06-08

Cut doctor over to workspace-owned .docmgrignore behavior, removed doctor-local ignore loading helpers, and kept CLI ignore flags as explicit command filters.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/commands/doctor.go — Doctor hard cutover to Workspace.IgnoreMatcher
- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/commands/doctor_test.go — Removed obsolete helper-level regression test
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/tasks.md — Checked off doctor hard-cutover tasks


## 2026-06-08

Added docmgr ignore explain command with structured output and fixed repo-relative docs-root path resolution.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/cmd/docmgr/cmds/ignorecmd/explain.go — Cobra wiring for ignore explain
- /home/manuel/code/wesen/go-go-golems/docmgr/cmd/docmgr/cmds/ignorecmd/ignore.go — Ignore command namespace
- /home/manuel/code/wesen/go-go-golems/docmgr/cmd/docmgr/cmds/root.go — Root command registration for ignore namespace
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/ignore/ignore.go — Repo-relative docs-root path resolution
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/ignore/ignore_test.go — Path resolution regression test
- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/commands/ignore_explain.go — New ignore decision explanation command


## 2026-06-08

Updated ignore docs and scenario coverage; added file-level WalkDocuments skip support after scenario exposed nested .docmgrignore file patterns still being parsed.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/internal/documents/walk.go — Added WithSkipFile traversal hook
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/workspace/index_builder.go — Applied ignore matcher to files before frontmatter parsing
- /home/manuel/code/wesen/go-go-golems/docmgr/internal/workspace/index_builder_test.go — Added file-level nested ignore pruning coverage
- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/doc/docmgr-cli-guide.md — Documented ignore explain and built-ins
- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/doc/docmgr-codebase-architecture.md — Added workspace ignore policy architecture note
- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/doc/docmgr-doctor-validation-workflow.md — Updated doctor validation ignore flow
- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/doc/docmgr-how-to-setup.md — Documented workspace-wide ignore behavior
- /home/manuel/code/wesen/go-go-golems/docmgr/test-scenarios/testing-doc-manager/21-ignore-policy.sh — New ignore policy scenario


## 2026-06-08

Final validation passed: go test ./... and local-binary docmgr doctor for DOCMGR-IGNORE-001 both succeeded; full scenario runner remains blocked by unrelated scenariolog dependency drift.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/reference/01-investigation-diary.md — Recorded final validation evidence and scenariolog caveat
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/tasks.md — Checked off final validation tasks


## 2026-06-08

Repaired scenariolog by aligning nested Glazed dependency with facade-package imports; full scenario suite now passes with sqlite_fts5 docmgr build.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/scenariolog/go.mod — Updated Glazed dependency to v1.0.5 and Go version required by that dependency
- /home/manuel/code/wesen/go-go-golems/docmgr/scenariolog/go.sum — Updated nested module dependency checksums
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/reference/01-investigation-diary.md — Recorded scenariolog repair and full scenario validation
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/tasks.md — Added and checked scenario harness repair tasks


## 2026-06-08

Addressed PR #40 review: duplicate-index filesystem walk now honors file-level ignore decisions before appending index.md files.

### Related Files

- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/commands/doctor.go — Applied file-level skip predicate in duplicate index scan
- /home/manuel/code/wesen/go-go-golems/docmgr/pkg/commands/doctor_test.go — Regression test for ignored duplicate index.md files
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/reference/01-investigation-diary.md — Recorded PR review fix
- /home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/06/08/DOCMGR-IGNORE-001--design-nested-docmgrignore-and-shared-ignore-policy/tasks.md — Added PR #40 review follow-up checklist

