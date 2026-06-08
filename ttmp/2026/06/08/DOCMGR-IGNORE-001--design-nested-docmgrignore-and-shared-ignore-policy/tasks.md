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

- [ ] Add `github.com/denormal/go-gitignore` to `go.mod` / `go.sum`.
- [ ] Create `internal/ignore` package.
- [ ] Define `LoadOptions`, `Matcher`, `Decision`, and source metadata types.
- [ ] Add built-in ignore defaults (`.git/`, `node_modules/`, `.pnpm/`, `dist/`, `build/`, `coverage/`, `.venv/`, `__pycache__/`).
- [ ] Add root/docs-root `.docmgrignore` loading.
- [ ] Add table tests for root/docs-root loading and missing ignore files.

### Phase 2 — Matching semantics

- [ ] Implement `Matcher.Match(path string, isDir bool) Decision` using `go-gitignore`.
- [ ] Normalize absolute, repo-relative, and docs-root-relative candidate paths before matching.
- [ ] Test `node_modules/` matches nested path segments and descendants.
- [ ] Test `node_modules/` does not match substring directories like `my-node_modules-cache`.
- [ ] Test `dist/`, `.git/`, `**/draft-*.md`, and anchored patterns used in docs.
- [ ] Test built-in ignores independently from `.docmgrignore` files.

### Phase 3 — Workspace ownership

- [ ] Add `ignore *ignore.Matcher` field to `internal/workspace.Workspace`.
- [ ] Add `IgnoreMatcher() *ignore.Matcher` accessor.
- [ ] Load the matcher during `DiscoverWorkspace` / `NewWorkspaceFromContext`.
- [ ] Ensure missing `.docmgrignore` files are non-fatal.
- [ ] Add workspace construction tests covering matcher presence and defaults.

### Phase 4 — Index-time pruning

- [ ] Wire `Workspace.IgnoreMatcher()` into `Workspace.InitIndex` / `ingestWorkspaceDocs`.
- [ ] Combine existing `DefaultIngestSkipDir` with matcher-based skip decisions.
- [ ] Ensure ignored directories are pruned before `ReadDocumentWithFrontmatter`.
- [ ] Add an index builder test where invalid Markdown in `scripts/node_modules` is not indexed.
- [ ] Verify non-ignored invalid Markdown still appears with `IncludeErrors=true`.

### Phase 5 — Doctor hard cutover

- [ ] Remove doctor-local `.docmgrignore` loading helpers (`loadDocmgrIgnore`, `matchesAnyGlob`, `normalizeIgnorePattern`, etc.).
- [ ] Remove doctor post-filtering of `QueryDocs` results for ignored paths.
- [ ] Use `ws.IgnoreMatcher()` for `FindTicketScaffoldsMissingIndex` skip callback.
- [ ] Use `ws.IgnoreMatcher()` for duplicate `index.md` scan pruning.
- [ ] Preserve `doctor --doc` explicit single-file validation even when a file is normally ignored.
- [ ] Replace old doctor ignore regression test with workspace/doctor integration tests.

### Phase 6 — Nested `.docmgrignore`

- [ ] Discover nested `.docmgrignore` files under docs root.
- [ ] Use built-ins and root/docs-root rules while discovering nested ignore files to avoid walking ignored dependency trees.
- [ ] Scope nested ignore files to their containing directory subtree.
- [ ] Add tests for ticket-local and `scripts/`-local `.docmgrignore` files.
- [ ] Add tests for parent + nested precedence, including negation if supported by `go-gitignore` composition.

### Phase 7 — CLI explanation command

- [ ] Add `docmgr ignore explain <path>` command.
- [ ] Render final decision, path, matched source class/file, and reason.
- [ ] Add structured output fields for Glazed rendering.
- [ ] Add command tests or scenario coverage.

### Phase 8 — Docs and scenarios

- [ ] Update `pkg/doc/docmgr-how-to-setup.md` for `go-gitignore`-backed behavior and nested `.docmgrignore`.
- [ ] Update `pkg/doc/docmgr-cli-guide.md` so ignore behavior is workspace-wide, not doctor-only.
- [ ] Update `pkg/doc/docmgr-doctor-validation-workflow.md` to remove post-filter-only wording.
- [ ] Update `pkg/doc/docmgr-codebase-architecture.md` with the ignore engine in workspace/indexing architecture.
- [ ] Add a scenario test for ignored `scripts/node_modules` invalid Markdown.
- [ ] Add a scenario test for nested `.docmgrignore`.

### Phase 9 — Validation and delivery

- [ ] Run focused unit tests for `internal/ignore`, `internal/workspace`, and `pkg/commands`.
- [ ] Run full `go test ./...` if feasible.
- [ ] Run relevant `test-scenarios/testing-doc-manager` smoke tests.
- [ ] Run `docmgr --root /home/manuel/code/wesen/go-go-golems/docmgr/ttmp doctor --ticket DOCMGR-IGNORE-001 --stale-after 30` unless intentionally skipped.
- [ ] Update diary after each implementation phase.
- [ ] Commit at appropriate implementation boundaries.
