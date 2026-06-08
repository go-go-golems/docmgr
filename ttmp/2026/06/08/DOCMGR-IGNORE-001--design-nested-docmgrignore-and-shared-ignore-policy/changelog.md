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

