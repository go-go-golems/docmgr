# Tasks

## Completed setup

- [x] Create DOCMGR-IGNORE-001 ticket workspace.
- [x] Inspect current doctor, workspace indexing, document walking, and QueryDocs behavior.
- [x] Write intern-facing nested `.docmgrignore` and shared ignore policy design guide.
- [x] Revise design guide to use direct workspace-owned `go-gitignore` cutover instead of opt-in compatibility path.
- [x] Write investigation diary.
- [x] Upload initial design bundle to reMarkable.

## Implementation checklist

### Phase 1 — Dependency and package skeleton

- [x] Add `github.com/denormal/go-gitignore` to `go.mod` / `go.sum`.
- [x] Create `internal/ignore` package.
- [x] Define `LoadOptions`, `Matcher`, `Decision`, and source metadata types.
- [x] Add built-in ignore defaults (`.git/`, `node_modules/`, `.pnpm/`, `dist/`, `build/`, `coverage/`, `.venv/`, `__pycache__/`).
- [x] Add root/docs-root `.docmgrignore` loading.
- [x] Add table tests for root/docs-root loading and missing ignore files.

### Phase 2 — Matching semantics

- [x] Implement `Matcher.Match(path string, isDir bool) Decision` using `go-gitignore`.
- [x] Normalize absolute, repo-relative, and docs-root-relative candidate paths before matching.
- [x] Test `node_modules/` matches nested path segments and descendants.
- [x] Test `node_modules/` does not match substring directories like `my-node_modules-cache`.
- [x] Test `dist/`, `.git/`, `**/draft-*.md`, and anchored patterns used in docs.
- [x] Test built-in ignores independently from `.docmgrignore` files.

### Phase 3 — Workspace ownership

- [x] Add `ignore *ignore.Matcher` field to `internal/workspace.Workspace`.
- [x] Add `IgnoreMatcher() *ignore.Matcher` accessor.
- [x] Load the matcher during `DiscoverWorkspace` / `NewWorkspaceFromContext`.
- [x] Ensure missing `.docmgrignore` files are non-fatal.
- [x] Add workspace construction tests covering matcher presence and defaults.

### Phase 4 — Index-time pruning

- [x] Wire `Workspace.IgnoreMatcher()` into `Workspace.InitIndex` / `ingestWorkspaceDocs`.
- [x] Combine existing `DefaultIngestSkipDir` with matcher-based skip decisions.
- [x] Ensure ignored directories are pruned before `ReadDocumentWithFrontmatter`.
- [x] Add an index builder test where invalid Markdown in `scripts/node_modules` is not indexed.
- [x] Verify non-ignored invalid Markdown still appears with `IncludeErrors=true`.

### Phase 5 — Doctor hard cutover

- [x] Remove doctor-local `.docmgrignore` loading helpers (`loadDocmgrIgnore`, `matchesAnyGlob`, `normalizeIgnorePattern`, etc.).
- [x] Remove doctor post-filtering of `QueryDocs` results for ignored paths.
- [x] Use `ws.IgnoreMatcher()` for `FindTicketScaffoldsMissingIndex` skip callback.
- [x] Use `ws.IgnoreMatcher()` for duplicate `index.md` scan pruning.
- [x] Preserve `doctor --doc` explicit single-file validation even when a file is normally ignored.
- [x] Replace old doctor ignore regression test with workspace/doctor integration tests.

### Phase 6 — Nested `.docmgrignore`

- [x] Discover nested `.docmgrignore` files under docs root.
- [x] Use built-ins and root/docs-root rules while discovering nested ignore files to avoid walking ignored dependency trees.
- [x] Scope nested ignore files to their containing directory subtree.
- [x] Add tests for ticket-local and `scripts/`-local `.docmgrignore` files.
- [x] Add tests for parent + nested precedence, including negation if supported by `go-gitignore` composition.

### Phase 7 — CLI explanation command

- [x] Add `docmgr ignore explain <path>` command.
- [x] Render final decision, path, matched source class/file, and reason.
- [x] Add structured output fields for Glazed rendering.
- [x] Add command tests or scenario coverage.

### Phase 8 — Docs and scenarios

- [x] Update `pkg/doc/docmgr-how-to-setup.md` for `go-gitignore`-backed behavior and nested `.docmgrignore`.
- [x] Update `pkg/doc/docmgr-cli-guide.md` so ignore behavior is workspace-wide, not doctor-only.
- [x] Update `pkg/doc/docmgr-doctor-validation-workflow.md` to remove post-filter-only wording.
- [x] Update `pkg/doc/docmgr-codebase-architecture.md` with the ignore engine in workspace/indexing architecture.
- [x] Add a scenario test for ignored `scripts/node_modules` invalid Markdown.
- [x] Add a scenario test for nested `.docmgrignore`.

### Phase 9 — Validation and delivery

- [x] Run focused unit tests for `internal/ignore`, `internal/workspace`, and `pkg/commands`.
- [ ] Run full `go test ./...` if feasible.
- [x] Run relevant `test-scenarios/testing-doc-manager` smoke tests.
- [ ] Run `docmgr --root /home/manuel/code/wesen/go-go-golems/docmgr/ttmp doctor --ticket DOCMGR-IGNORE-001 --stale-after 30` unless intentionally skipped.
- [x] Update diary after each implementation phase.
- [x] Commit at appropriate implementation boundaries.
