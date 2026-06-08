---
Title: Nested docmgrignore and shared ignore policy implementation guide
Ticket: DOCMGR-IGNORE-001
Status: active
Topics:
    - docmgr
    - cli
    - testing
    - diagnostics
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/documents/walk.go
      Note: Lowest-level Markdown traversal hook where ignored directories should be pruned
    - Path: internal/workspace/discovery.go
      Note: Missing-index filesystem walk that needs the shared ignore matcher
    - Path: internal/workspace/index_builder.go
      Note: Workspace indexing ingest loop that currently parses Markdown before doctor post-filtering
    - Path: internal/workspace/query_docs.go
      Note: QueryDocs API and visibility options that should remain separate from ignore semantics
    - Path: internal/workspace/query_docs_sql.go
      Note: SQL visibility filters for archive/scripts/control docs
    - Path: internal/workspace/skip_policy.go
      Note: Canonical ingest skip policy and path-tag behavior to preserve
    - Path: pkg/commands/doctor.go
      Note: Current doctor-specific .docmgrignore loading
    - Path: pkg/doc/docmgr-doctor-validation-workflow.md
      Note: Existing documentation of doctor's post-filter compatibility behavior
    - Path: pkg/doc/docmgr-how-to-setup.md
      Note: User-facing .docmgrignore examples and promised pattern behavior
ExternalSources: []
Summary: Design a shared ignore engine for docmgr so .docmgrignore works consistently across discovery, indexing, doctor, search, and list commands.
LastUpdated: 2026-06-08T14:43:26.749782024-04:00
WhatFor: Guide a new intern through the current docmgr ignore behavior, why nested dependency folders leak into validation, and how to implement a robust shared ignore policy.
WhenToUse: Use before changing .docmgrignore handling, workspace indexing, doctor validation, document walking, or QueryDocs visibility behavior.
---


# Nested docmgrignore and shared ignore policy implementation guide

## Executive summary

`docmgr` already has several pieces of an ignore system, but those pieces live at different layers and do not yet form one coherent contract. The immediate symptom is that a ticket-local `scripts/node_modules/` directory can be indexed and validated even when the docs root contains `.docmgrignore` patterns such as `node_modules/` or `**/node_modules/`. The deeper issue is architectural: workspace indexing uses a canonical ingest skip policy, `doctor` applies `.docmgrignore` as a post-filter, missing-index detection uses a separate walk with a callback, and commands that query documents rely on `QueryDocs` visibility flags rather than a shared ignore resolver.

This design proposes a proper ignore subsystem centered on a new internal package, tentatively `internal/ignore`, built on `github.com/denormal/go-gitignore` rather than a handwritten glob engine. The package should load `.docmgrignore` files from the repository root, docs root, and nested directories; compile them into a reusable workspace-owned matcher; expose explanation/debug information where the dependency allows; and integrate with workspace discovery before markdown files are parsed. The intended end state is that ignored files are pruned before indexing, `doctor`, `doc list`, `doc search`, HTTP indexing, and future commands share one policy, and users can ask `docmgr ignore explain <path>` to see exactly why a file is included or excluded.

The implementation should be a direct correctness cutover, not a compatibility bridge. Workspace discovery should load the ignore policy once, store it on `Workspace`, and every workspace traversal should use it. Doctor-local `.docmgrignore` helpers and post-filter behavior should be removed instead of preserved. This is simpler than an opt-in transition because ignored paths have one meaning everywhere: they are not parsed, indexed, queried, or validated during workspace scans.

## Problem statement and scope

### User-visible problem

Users reasonably expect `.docmgrignore` to behave like a lightweight `.gitignore` for documentation validation. Existing docs explicitly show patterns such as `node_modules/`, `dist/`, `coverage/`, `**/draft-*.md`, and `**/scratch-*.md` as common patterns for `ttmp/.docmgrignore` (`pkg/doc/docmgr-how-to-setup.md:502-531`). The CLI guide says `doctor` respects a `.docmgrignore` file at the repository root or docs root (`pkg/doc/docmgr-cli-guide.md:426-428`). However, the current implementation only loads those two files inside `doctor` and then applies patterns as a post-filter over already-indexed documents (`pkg/commands/doctor.go:218-235`, `pkg/commands/doctor.go:375-382`).

That means a dependency directory can be scanned before the ignore filter runs. If the dependency contains arbitrary Markdown files without docmgr frontmatter, the index records parse errors and `doctor` can report noisy YAML/frontmatter failures. This is especially visible when ticket scripts use local npm dependencies:

```text
ttmp/YYYY/MM/DD/TICKET--slug/scripts/node_modules/.pnpm/playwright-core@.../README.md
```

The correct user outcome is simple: if `.docmgrignore` contains `node_modules/`, no command should parse, index, validate, list, search, or diagnose files below any `node_modules` directory unless a future explicit override says otherwise.

### Engineering problem

The codebase currently has three overlapping but incomplete concepts:

- **Canonical ingest skip policy:** `internal/workspace/skip_policy.go` defines hard-coded ingest-time skips for `.meta/` and underscore directories. This is fast because it can prune during walking (`internal/workspace/skip_policy.go:21-35`).
- **Document walking:** `internal/documents/walk.go` recursively walks markdown files and supports a `WithSkipDir` hook (`internal/documents/walk.go:15-27`, `internal/documents/walk.go:29-56`). The workspace index builder currently calls `WalkDocuments` without a custom ignore resolver (`internal/workspace/index_builder.go:139-185`).
- **Doctor compatibility filter:** `doctor` loads `.docmgrignore` from repository root and docs root, appends those patterns to `IgnoreGlobs`, and filters `QueryDocs` results after indexing (`pkg/commands/doctor.go:218-235`, `pkg/commands/doctor.go:370-382`). It also has separate helper functions for `shouldSkipDoctorDoc`, `findIndexFiles`, `matchesAnyGlob`, and `loadDocmgrIgnore` (`pkg/commands/doctor.go:820-965`).

The missing abstraction is a workspace-level ignore resolver that can be shared by all traversal and query entry points.

### In scope

This ticket covers the design for:

- `.docmgrignore` pattern semantics.
- Nested `.docmgrignore` discovery and precedence.
- A shared Go API for ignore loading, matching, directory pruning, and explanation.
- Integration with workspace indexing, `doctor`, missing-index detection, list/search behavior, HTTP index refresh, and tests.
- Migration from doctor-only helper functions to a reusable internal package.

### Out of scope

This ticket does not require implementing a full clone of Git's ignore engine on day one. It also does not require changing document frontmatter parsing, diagnostics taxonomy rendering, task/changelog commands, or the ticket layout itself except where those systems consume document discovery results.

## Current-state architecture

### High-level map

The relevant docmgr flow today looks like this:

```text
CLI command
  |
  |-- workspace.DiscoverWorkspace(root)
  |     resolves docs root, config dir, repo root
  |
  |-- Workspace.InitIndex()
  |     opens in-memory SQLite
  |     calls documents.WalkDocuments(root)
  |       skips _* directories by default
  |       reads every *.md via ReadDocumentWithFrontmatter
  |     inserts docs rows, including parse errors
  |
  |-- Workspace.QueryDocs(query)
  |     applies scope, filters, visibility flags, parse-error options
  |
  |-- command-specific post-processing
        doctor applies .docmgrignore after QueryDocs
```

Important files:

- `internal/workspace/workspace.go`: workspace construction and root/config/repo context.
- `internal/workspace/index_builder.go`: SQLite index lifecycle and ingest loop.
- `internal/documents/walk.go`: recursive Markdown walking and directory skip hook.
- `internal/workspace/query_docs.go`: public document query API and `DocQueryOptions`.
- `internal/workspace/query_docs_sql.go`: SQL visibility filters for archive/scripts/control docs.
- `internal/workspace/skip_policy.go`: current hard-coded ingest skip policy and path tags.
- `pkg/commands/doctor.go`: validation command, `.docmgrignore` loading, post-filtering, and duplicate-index scans.

### Workspace discovery

`DiscoverWorkspace` accepts a root override, resolves the docs root, loads `.ttmp.yaml` best-effort, resolves the config directory, and requires a repository root (`internal/workspace/workspace.go:37-90`). This matters because ignore files may live in either the repository root or docs root today, and nested support needs both roots to compute relative paths safely.

API reference:

```go
type DiscoverOptions struct {
    RootOverride string
}

func DiscoverWorkspace(ctx context.Context, opts DiscoverOptions) (*Workspace, error)
```

The resulting `WorkspaceContext` carries:

```go
type WorkspaceContext struct {
    Root      string // docs root, usually ttmp
    ConfigDir string // directory containing .ttmp.yaml or parent of root
    RepoRoot  string // git repository root
    Config    *WorkspaceConfig
}
```

### Document walking

`WalkDocuments` is the lowest-level recursive scanner for managed Markdown files. It uses `filepath.WalkDir`, skips underscore-prefixed directories by default, optionally accepts a `WithSkipDir` predicate, filters to `.md`, and calls `ReadDocumentWithFrontmatter` (`internal/documents/walk.go:29-56`).

Pseudocode for the current walk:

```go
func WalkDocuments(root string, fn WalkDocumentFunc, opts ...WalkOption) error {
    cfg := apply(opts)
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if d.IsDir() {
            if strings.HasPrefix(d.Name(), "_") { return fs.SkipDir }
            if cfg.skipDir != nil && cfg.skipDir(path, d) { return fs.SkipDir }
            return nil
        }
        if ext(path) != ".md" { return nil }
        doc, body, readErr := ReadDocumentWithFrontmatter(path)
        return fn(path, doc, body, readErr)
    })
}
```

This is the best integration point for directory pruning. If `node_modules` is skipped here, no downstream command sees dependency Markdown files unless it bypasses `WalkDocuments`.

### Workspace indexing

`Workspace.InitIndex` rebuilds an in-memory SQLite index per CLI invocation. It opens SQLite, creates schema, optionally creates FTS, and calls `ingestWorkspaceDocs` (`internal/workspace/index_builder.go:24-73`). The ingest loop calls `documents.WalkDocuments(wctx.Root, ...)` and inserts each Markdown file into the `docs` table (`internal/workspace/index_builder.go:75-185`).

The ingest loop deliberately stores parse-error documents with `parse_ok = 0` and a fallback ticket ID inferred from path (`internal/workspace/index_builder.go:159-170`). This is useful for repair workflows, but it is dangerous when ignored dependency files are allowed into the index: dependency Markdown files become repair findings even though the user asked to ignore them.

### QueryDocs visibility

`QueryDocs` requires an initialized index (`internal/workspace/query_docs.go:17-26`), compiles SQL filters, scans rows, and returns `DocHandle` objects. The query options include `IncludeErrors`, `IncludeArchivedPath`, `IncludeControlDocs`, `IncludeScriptsPath`, and `IncludeDiagnostics` (`internal/workspace/query_docs.go:331-342`). The SQL compiler hides archive, scripts, and control docs unless explicitly included (`internal/workspace/query_docs_sql.go:31-40`).

The key distinction is:

- **Visibility flags** decide which indexed documents are returned.
- **Ignore rules** should decide which filesystem paths are never indexed in the first place.

Do not implement `.docmgrignore` only as another `QueryDocs` visibility filter. That would still parse ignored files during indexing, waste work, and leak parse diagnostics into index initialization paths.

### Doctor command

`doctor` is currently the most complete consumer of ignore patterns, but it owns too much ignore-specific code. It loads patterns from repository root and docs root (`pkg/commands/doctor.go:218-235`). It builds a `skipFn` for missing-index detection (`pkg/commands/doctor.go:275-287`). It initializes the workspace index (`pkg/commands/doctor.go:309-316`). It queries all relevant docs, including parse errors, scripts, archive paths, and control docs (`pkg/commands/doctor.go:357-368`). It post-filters the `QueryDocs` result using `shouldSkipDoctorDoc` (`pkg/commands/doctor.go:375-382`). It separately scans for duplicate `index.md` files with `findIndexFiles` (`pkg/commands/doctor.go:859-890`).

The result is inconsistent timing:

```text
missing-index detection: skip callback during filesystem walk
workspace index: no .docmgrignore pruning before parse
QueryDocs: returns indexed parse errors
Doctor post-filter: attempts to remove ignored docs after indexing
Duplicate index scan: has its own ignore matching
```

### Current documentation contract

The docs promise more than the implementation robustly provides:

- `pkg/doc/docmgr-how-to-setup.md:506-529` lists common `.docmgrignore` patterns including `node_modules/`, `dist/`, `coverage/`, `**/draft-*.md`, and `**/scratch-*.md`.
- `pkg/doc/docmgr-how-to-setup.md:531` says doctor automatically respects `.docmgrignore` from both repository root and docs root.
- `pkg/doc/docmgr-cli-guide.md:426-428` says `doctor` respects `.docmgrignore` at repository root or docs root.
- `pkg/doc/docmgr-doctor-validation-workflow.md:51-53` accurately describes the current implementation as a compatibility post-filter over `QueryDocs` results.

The future implementation should either fully support the documented patterns or update the docs to state the exact subset. This guide recommends supporting the documented subset and adding tests.

## Gap analysis

### Gap 1: Ignore behavior is command-specific

Only `doctor` loads `.docmgrignore` today. Other commands that rely on workspace indexing, such as list/search/status/export flows, can still index ignored files unless they independently add ignore support. A shared ignore engine should live below commands, not inside `pkg/commands/doctor.go`.

### Gap 2: Ignored files are filtered too late

`doctor` applies `.docmgrignore` after `Workspace.QueryDocs`. Because `InitIndex` has already parsed Markdown files, ignored dependency Markdown can still create parse-error rows. This is why `node_modules` can surface as YAML/frontmatter noise.

### Gap 3: Pattern semantics are unclear

The docs show gitignore-like patterns, but the current helper historically used `filepath.Match`, which does not implement full gitignore semantics. A minimal patch can make simple directory names match path segments, but that is not enough for nested `.docmgrignore`, negation, anchored patterns, or reliable `**` semantics.

### Gap 4: Nested `.docmgrignore` files are absent

Current loading only checks the repository root and docs root. There is no mechanism for a ticket to contain its own `.docmgrignore`, for a `scripts/` directory to ignore its local package-manager artifacts, or for nested ignore files to override parent rules.

### Gap 5: No explanation/debug interface

When ignore behavior surprises users, there is no command that answers: which ignore file was loaded, which pattern matched, and why was this path included or excluded? That makes bugs hard to distinguish from user pattern mistakes.

## Proposed architecture

### New package: `internal/ignore`

Create a small package that owns loading, matching, and explaining docmgr ignore behavior. The implementation should use `github.com/denormal/go-gitignore` as the underlying gitignore-compatible matcher rather than reimplementing `*`, `**`, directory-only patterns, anchoring, and negation from scratch. Docmgr-specific code should focus on locating `.docmgrignore` files, composing root/docs/nested matchers, adding built-in excludes, normalizing paths, and producing decisions that commands can consume.

Suggested files:

```text
internal/ignore/
  ignore.go          # public types and Match API
  loader.go          # discover .docmgrignore files
  pattern.go         # lightweight source/line metadata for patterns where needed
  matcher.go         # go-gitignore-backed matching and docmgr path normalization
  explain.go         # decision trace objects
  ignore_test.go     # table tests for semantics
  loader_test.go     # root/docs/nested loading tests
```

The package should be internal because it is an implementation detail of the CLI and workspace system. If later the HTTP API or external packages need it, promote only stable types.

### Core types

API sketch:

```go
package ignore

type SourceKind string

const (
    SourceRepositoryRoot SourceKind = "repository-root"
    SourceDocsRoot       SourceKind = "docs-root"
    SourceNested         SourceKind = "nested"
    SourceBuiltin        SourceKind = "builtin"
    SourceCLI            SourceKind = "cli"
)

type Pattern struct {
    Raw           string
    Normalized    string
    SourcePath    string
    SourceKind    SourceKind
    Line          int
    Negated       bool
    DirectoryOnly bool
    Anchored      bool
}

type Decision struct {
    Path        string
    IsDir       bool
    Ignored     bool
    Matched     *Pattern
    Trace       []TraceStep
}

type TraceStep struct {
    Pattern Pattern
    Matched bool
    Effect  string // "ignore", "include", "no-match"
    Reason  string
}

type Matcher struct {
    repoRoot string
    docsRoot string
    patterns []Pattern
    nestedEnabled bool
}

func Load(ctx context.Context, opts LoadOptions) (*Matcher, error)
func (m *Matcher) Match(path string, isDir bool) Decision
func (m *Matcher) SkipDir(path string, d fs.DirEntry) bool
func (m *Matcher) Explain(path string, isDir bool) Decision
```

### Load options

```go
type LoadOptions struct {
    RepoRoot string
    DocsRoot string

    // Command-line compatibility with --ignore-dir and --ignore-glob.
    IgnoreDirs  []string
    IgnoreGlobs []string

    // Defaults true after rollout; start behind an option if needed.
    IncludeBuiltin bool
    IncludeNested  bool
}
```

The loader should read ignore files in deterministic order:

1. Built-in patterns, if enabled.
2. Repository-root `.docmgrignore`, if present.
3. Docs-root `.docmgrignore`, if present and distinct from repository root.
4. Nested `.docmgrignore` files, if enabled.
5. CLI `--ignore-dir` and `--ignore-glob` compatibility patterns.

Later patterns should be able to override earlier patterns if negation is supported. If negation is not implemented in phase one, reject or warn on `!pattern` instead of silently misinterpreting it.

### Matching semantics

The recommended matcher is `github.com/denormal/go-gitignore`, the same dependency already used by `clay/pkg/filefilter`. Docmgr should avoid inventing its own gitignore dialect. The supported semantics should therefore follow the dependency's gitignore behavior for normal `.gitignore` features, while docmgr adds clear path normalization, built-in default ignores, nested `.docmgrignore` source loading, and tests for the patterns documented in docmgr help.

Patterns docmgr must test and support:

| Pattern | Meaning |
| --- | --- |
| `node_modules/` | Ignore any directory segment named `node_modules` and all descendants. |
| `dist/` | Ignore any directory segment named `dist` and all descendants. |
| `.git/` | Ignore any `.git` directory and descendants. |
| `2024-*/` | Ignore matching directory names at any segment unless anchored. |
| `ttmp/*/design-doc/index.md` | Match docs-root-relative paths where `*` spans one segment. |
| `**/draft-*.md` | Match `draft-*.md` at any depth. |
| `/archive/` | Anchored to the ignore file's directory. |
| `!keep.md` | Re-include a path previously ignored, if negation is implemented. |

Important implementation notes:

- Normalize all candidate paths to slash form with `filepath.ToSlash(filepath.Clean(path))`.
- Prefer paths relative to the matcher base directory when calling `go-gitignore`; keep absolute paths only as fallback/debug metadata.
- Add tests proving `node_modules/` matches path segments and descendants, not substrings such as `my-node_modules-cache/`.
- Let `go-gitignore` handle `*`, `**`, anchoring, directory-only patterns, and negation whenever possible.
- Directory match should ignore all descendants without needing every file to match separately.
- If `go-gitignore` cannot expose enough source/line detail for full tracing, `docmgr ignore explain` can initially report the matching ignore file/matcher class and final decision, with richer line-level traces as a follow-up.

Pseudocode:

```go
func (m *Matcher) Match(path string, isDir bool) Decision {
    candidate := normalizeCandidate(path, m.repoRoot, m.docsRoot)
    ignored := false
    var matched *Pattern
    var trace []TraceStep

    for _, p := range m.patternsForPath(candidate) {
        ok, reason := matchPattern(p, candidate, isDir)
        if !ok {
            trace = append(trace, TraceStep{Pattern: p, Matched: false, Effect: "no-match", Reason: reason})
            continue
        }
        if p.Negated {
            ignored = false
            matched = &p
            trace = append(trace, TraceStep{Pattern: p, Matched: true, Effect: "include", Reason: reason})
        } else {
            ignored = true
            matched = &p
            trace = append(trace, TraceStep{Pattern: p, Matched: true, Effect: "ignore", Reason: reason})
        }
    }

    return Decision{Path: candidate.DisplayPath, IsDir: isDir, Ignored: ignored, Matched: matched, Trace: trace}
}
```

### Nested `.docmgrignore` model

Nested ignore files should be scoped to their own directory subtree. For example:

```text
ttmp/.docmgrignore
  node_modules/

ttmp/2026/06/08/TICKET--slug/scripts/.docmgrignore
  screenshots/
  !screenshots/keep.md
```

For a candidate path under `.../scripts/screenshots/a.png`, the matcher should apply:

1. built-ins,
2. repo-root rules,
3. docs-root rules,
4. ticket-level rules, if any parent directory has `.docmgrignore`,
5. scripts-level rules.

Implementation strategy:

- During `Load`, walk only the docs tree for files named `.docmgrignore`.
- Use existing skip rules while discovering ignore files to avoid scanning `.git`, `node_modules`, `.meta`, and underscore directories.
- Store nested patterns with `BaseDir` equal to the directory containing that ignore file.
- At match time, a nested pattern applies only if the candidate path is inside `BaseDir`.
- For anchored nested patterns (`/foo`), anchor to `BaseDir`, not docs root.
- For unanchored nested patterns, match within the relative path below `BaseDir`.

Data sketch:

```go
type Pattern struct {
    Raw        string
    SourcePath string
    BaseDir    string // directory containing the .docmgrignore
    Line       int
    Negated    bool
    Anchored   bool
    DirectoryOnly bool
}
```

### Built-in ignores

Docmgr should have conservative built-in ignores for directories that are almost never managed docs:

```text
.git/
node_modules/
.pnpm/
dist/
build/
coverage/
.venv/
__pycache__/
.meta/
```

Caveat: `.meta/` and `_*/` already have canonical ingest behavior. Keep those hard skips, but document them as built-ins. If negation support is added, do not allow negating `.meta/` or `_templates/` during indexing unless a future design explicitly permits it. These are control directories and should remain outside document indexing.

### Integrating with workspace ownership and indexing

Use the simpler direct design: `Workspace` owns the ignore matcher. Do not add an opt-in `BuildIndexOptions.IgnoreMatcher` transition. Once the matcher exists, workspace scans should always respect it. This is a hard correctness cutover: ignored paths are outside docmgr's managed document universe unless the user explicitly validates a single file.

Extend `Workspace`:

```go
type Workspace struct {
    ctx      WorkspaceContext
    resolver *paths.Resolver
    db       *sql.DB
    ignore   *ignore.Matcher
}

func (w *Workspace) IgnoreMatcher() *ignore.Matcher {
    return w.ignore
}
```

Load the matcher in `DiscoverWorkspace` or `NewWorkspaceFromContext` after `WorkspaceContext` is known:

```go
matcher, err := ignore.Load(ctx, ignore.LoadOptions{
    RepoRoot: ctx.RepoRoot,
    DocsRoot: ctx.Root,
    IncludeBuiltin: true,
    IncludeNested: true,
})
if err != nil {
    return nil, err
}
```

Then use it in `ingestWorkspaceDocs` through `documents.WithSkipDir`:

```go
walkErr := documents.WalkDocuments(
    wctx.Root,
    ingestFn,
    documents.WithSkipDir(func(path string, d fs.DirEntry) bool {
        if workspace.DefaultIngestSkipDir(path, d) { return true }
        if matcher != nil && matcher.Match(path, true).Ignored { return true }
        return false
    }),
)
```

This requires `ingestWorkspaceDocs` to receive the matcher, either by changing its signature or by passing the full `Workspace` instead of only `WorkspaceContext`. Prefer the smallest clean signature change that keeps tests readable.

### Integrating with `doctor`

`doctor` should stop owning ignore parsing entirely. It should use `ws.IgnoreMatcher()` for every filesystem skip decision and remove the current helper functions once replacement tests pass. There should be no doctor-only compatibility post-filter after the cutover.

New doctor flow:

```go
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
if err != nil { return err }

matcher := ws.IgnoreMatcher()

missingIndexDirs, err := workspace.FindTicketScaffoldsMissingIndex(ctx, settings.Root,
    func(relPath, base string) bool {
        if containsString(settings.IgnoreDirs, base) { return true } // CLI compatibility if retained
        if matcher == nil { return false }
        return matcher.Match(filepath.Join(settings.Root, relPath), true).Ignored
    })

if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
    return err
}

qr, err := ws.QueryDocs(ctx, query)
```

After this change:

- ignored paths are not indexed, so `QueryDocs` never sees them;
- `shouldSkipDoctorDoc`, `matchesAnyGlob`, `normalizeIgnorePattern`, and `loadDocmgrIgnore` can be deleted from `pkg/commands/doctor.go`;
- duplicate `index.md` scanning should use the workspace matcher while walking;
- `--ignore-dir` / `--ignore-glob` can either be folded into the workspace matcher as command overrides or removed in a later CLI cleanup. For this implementation, keep the flags by passing command overrides into the matcher or by applying them only to doctor-specific extra scans.

### Integrating with `QueryDocs`

Do not add ignore patterns directly to SQL. `QueryDocs` should query documents that were intentionally indexed. If a command wants to include normally hidden scripts or archive paths, it can use `IncludeScriptsPath` or `IncludeArchivedPath`; ignored paths should not exist in the index.

The one useful addition is metadata for debugging:

```go
type BuildIndexStats struct {
    IndexedDocs int
    SkippedDirs []ignore.Decision
    SkippedFiles []ignore.Decision
}
```

This can power diagnostics or a future `docmgr status --explain-ignores` mode.

### New CLI: `docmgr ignore explain`

Add a command that explains ignore behavior:

```bash
docmgr ignore explain ttmp/2026/06/08/TICKET--slug/scripts/node_modules/playwright/README.md
```

Example output:

```text
Path: ttmp/2026/06/08/TICKET--slug/scripts/node_modules/playwright/README.md
Decision: ignored

Matched pattern:
  source: ttmp/.docmgrignore:6
  pattern: node_modules/
  reason: directory-only pattern matched path segment "node_modules"

Trace:
  builtin:.git/                         no-match
  ttmp/.docmgrignore:6 node_modules/    ignore
```

This command should support structured output through Glazed, but the human text is important for debugging.

Command placement:

```text
pkg/commands/ignore.go
cmd/docmgr/cmds/root.go or command registration file
```

Possible fields:

- `path`
- `ignored`
- `matched_pattern`
- `source_path`
- `source_line`
- `reason`
- `trace_json`

## Decision records

### Decision: Move ignore handling below commands

- **Context:** `doctor` currently owns `.docmgrignore` loading and matching, while indexing and other commands can still see ignored files.
- **Options considered:** Keep doctor-only filtering; add ad-hoc filters to each command; create a shared ignore package used during discovery/indexing.
- **Decision:** Create `internal/ignore` and integrate it with workspace indexing and filesystem walks.
- **Rationale:** Ignore behavior is a workspace concern. Applying it before parsing prevents dependency Markdown from becoming diagnostics noise and keeps command behavior consistent.
- **Consequences:** The workspace index becomes more faithful to user intent, but tests that expected ignored files to appear in query results may need updating.
- **Status:** proposed

### Decision: Support a documented gitignore-like subset instead of full Git parity initially

- **Context:** Users expect `.docmgrignore` to resemble `.gitignore`, but implementing every Git edge case is unnecessary for docmgr's current workflows.
- **Options considered:** Use `filepath.Match`; implement complete Git ignore semantics by hand; implement a docmgr-specific subset; use the existing `github.com/denormal/go-gitignore` dependency pattern from clay.
- **Decision:** Use `github.com/denormal/go-gitignore` for matching and keep docmgr-specific code limited to loading, composition, built-ins, nested source discovery, and explanation.
- **Rationale:** This avoids a custom ignore dialect and should correctly handle normal gitignore features with less code. Clay already uses this dependency successfully in `pkg/filefilter`.
- **Consequences:** The implementation depends on the library's semantics and diagnostics. If line-level explanation is not exposed, docmgr may need lightweight metadata tracking around loaded files.
- **Status:** proposed

### Decision: Prune ignored directories during indexing

- **Context:** Post-filtering after `QueryDocs` is too late because files have already been parsed and indexed.
- **Options considered:** Filter only `QueryDocs`; filter doctor rows; prune during `WalkDocuments`; combine pruning and post-filter safety.
- **Decision:** Prune ignored directories in `WalkDocuments` during `Workspace.InitIndex` and remove doctor-local post-filter compatibility behavior.
- **Rationale:** Directory pruning is faster, avoids false parse errors, and matches user expectations for dependency/build directories.
- **Consequences:** Ignored docs disappear from doctor/list/search/export results, which is the desired hard cutover and should be documented.
- **Status:** proposed

### Decision: Add `ignore explain` before adding more pattern features

- **Context:** Ignore systems are hard to debug because users cannot see which rule won.
- **Options considered:** Rely on tests and docs; add verbose logs; add a first-class explain command.
- **Decision:** Add `docmgr ignore explain <path>` as part of the implementation.
- **Rationale:** Explanation output reduces support burden and makes future nested/negation behavior auditable.
- **Consequences:** The matcher API must retain source/line/reason metadata, not just return a boolean.
- **Status:** proposed

## Implementation guide for a new intern

### Phase 0: Reproduce and understand the failure

Goal: understand why dependency Markdown files become doctor findings.

1. Create or use a ticket with a local npm dependency under `scripts/node_modules`.
2. Add `node_modules/` to `ttmp/.docmgrignore`.
3. Run:

```bash
docmgr --root /path/to/repo/ttmp doctor --ticket TICKET-ID --stale-after 30
```

4. Observe whether `node_modules` Markdown files appear as `invalid_frontmatter` findings.
5. Read these files:
   - `pkg/commands/doctor.go`
   - `internal/workspace/index_builder.go`
   - `internal/documents/walk.go`
   - `internal/workspace/skip_policy.go`
   - `internal/workspace/query_docs.go`

Expected learning: `doctor` loads ignore patterns, but the index builder parses files before doctor post-filtering.

### Phase 1: Add `internal/ignore` using `go-gitignore`

Create `internal/ignore` and add `github.com/denormal/go-gitignore` to `go.mod`. The package should wrap the dependency with docmgr-specific path normalization and decision types.

First API:

```go
matcher, err := ignore.Load(ctx, ignore.LoadOptions{
    RepoRoot: repoRoot,
    DocsRoot: docsRoot,
    IncludeBuiltin: true,
    IncludeNested: true,
})

decision := matcher.Match(absPath, isDir)
if decision.Ignored { ... }
```

Minimum tests:

- `node_modules/` matches `ticket/scripts/node_modules/playwright/README.md`.
- `node_modules/` does not match `ticket/scripts/my-node_modules-cache/README.md`.
- `dist/` matches a nested `dist` directory.
- `.git/` matches nested `.git` paths.
- `**/draft-*.md` matches documented recursive draft patterns.
- nested `.docmgrignore` patterns apply only to their own subtree.

Use clay as prior art, especially `/home/manuel/code/wesen/go-go-golems/clay/pkg/filefilter/layer.go`, but do not copy clay's substring-based directory exclusion. Docmgr should rely on `go-gitignore` and segment-safe tests.

### Phase 2: Make `Workspace` own the matcher

Extend `Workspace` so every command receives the same ignore policy after discovery.

Implementation sketch:

```go
type Workspace struct {
    ctx      WorkspaceContext
    resolver *paths.Resolver
    db       *sql.DB
    ignore   *ignore.Matcher
}

func (w *Workspace) IgnoreMatcher() *ignore.Matcher {
    return w.ignore
}
```

Load the matcher inside `NewWorkspaceFromContext` or directly in `DiscoverWorkspace` after `WorkspaceContext` is known. Prefer loading in `NewWorkspaceFromContext` if tests construct workspaces through that path and should get production behavior. If tests need custom contexts without ignore files, missing `.docmgrignore` files must be non-fatal.

### Phase 3: Prune ignored paths during indexing

Modify `Workspace.InitIndex` / `ingestWorkspaceDocs` so `documents.WalkDocuments` receives a skip predicate that combines the existing canonical hard skips with the workspace matcher.

Pseudocode:

```go
walkErr := documents.WalkDocuments(
    w.ctx.Root,
    ingestFn,
    documents.WithSkipDir(func(path string, d fs.DirEntry) bool {
        if workspace.DefaultIngestSkipDir(path, d) {
            return true
        }
        if w.ignore != nil && w.ignore.Match(path, true).Ignored {
            return true
        }
        return false
    }),
)
```

Add an index builder test that creates ignored invalid Markdown below `scripts/node_modules` and verifies it is not present even with `IncludeErrors=true` and `IncludeScriptsPath=true`.

### Phase 4: Replace doctor-local ignore logic

Inside `doctor`, remove the command-owned `.docmgrignore` implementation. The command should call `ws.IgnoreMatcher()` and use that matcher for every extra filesystem walk.

Remove or supersede these helpers from `pkg/commands/doctor.go` once tests pass:

- `loadDocmgrIgnore`
- `matchesAnyGlob`
- `matchesSimplePathSegmentPattern`
- `normalizeIgnorePattern`
- `shouldSkipDoctorDoc`

Doctor should not post-filter `QueryDocs` results for ignored paths because ignored paths should never enter the index. The remaining doctor-specific skip callback for `FindTicketScaffoldsMissingIndex` and duplicate `index.md` detection should delegate to `ws.IgnoreMatcher()`.

### Phase 5: Preserve explicit single-file validation

`docmgr doctor --doc path/to/ignored.md` is different from a workspace scan. If the user explicitly names a file, validate it even if the matcher says it is normally ignored. Optionally print or return a note in the future, but do not silently skip the file.

### Phase 6: Add nested `.docmgrignore` loading

Nested loading should happen inside `internal/ignore.Load`. Use built-ins plus root/docs-root rules while discovering nested ignore files so dependency directories are not walked unnecessarily.

Algorithm:

```go
func loadNested(ctx context.Context, docsRoot string, baseMatcher *Matcher) ([]sourceMatcher, error) {
    var matchers []sourceMatcher
    filepath.WalkDir(docsRoot, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if ctx.Err() != nil { return ctx.Err() }
        if d.IsDir() {
            if baseMatcher.Match(path, true).Ignored { return fs.SkipDir }
            return nil
        }
        if d.Name() != ".docmgrignore" { return nil }
        m := loadOneGitignore(filepath.Dir(path), path)
        matchers = append(matchers, m)
        return nil
    })
    return matchers, nil
}
```

Each nested matcher should apply only to paths under the directory containing that `.docmgrignore`.

### Phase 7: Add `docmgr ignore explain`

Add a CLI command once `Decision` is stable enough to be useful. The first version can report final ignored/included status and which matcher source matched if exact pattern line tracing is expensive.

Example human output:

```text
Decision: ignored
Path: ttmp/.../scripts/node_modules/playwright/README.md
Source: ttmp/.docmgrignore
Reason: gitignore matcher ignored this path
```

### Phase 8: Update docs and scenario tests

Update these docs:

- `pkg/doc/docmgr-how-to-setup.md`: document `go-gitignore`-backed behavior and nested `.docmgrignore`.
- `pkg/doc/docmgr-cli-guide.md`: mention workspace-wide ignore behavior, not doctor-only behavior.
- `pkg/doc/docmgr-doctor-validation-workflow.md`: remove the post-filter-only description.
- `pkg/doc/docmgr-codebase-architecture.md`: add ignore engine to workspace/indexing architecture.

Update scenarios:

- `test-scenarios/testing-doc-manager/01-create-mock-codebase.sh`: keep `.docmgrignore` seed.
- Add a new scenario step that creates nested `scripts/node_modules` with invalid Markdown and asserts doctor stays clean.
- Add `ignore explain` smoke coverage once the command exists.

## Test strategy

### Unit tests

Add focused tests in `internal/ignore`:

- parser skips comments/blank lines;
- parser preserves source path/line number;
- simple directory patterns match path segments;
- substring false positives do not match;
- `*` matches one segment;
- `**` matches zero or more segments;
- anchored patterns are anchored to the ignore file base directory;
- nested patterns apply only under their base directory;
- negation, if implemented, overrides earlier ignore decisions.

### Workspace integration tests

Add tests in `internal/workspace/index_builder_test.go`:

- ignored directories are not indexed;
- ignored invalid Markdown does not create parse-error docs;
- non-ignored invalid Markdown still appears when `IncludeErrors=true`;
- scripts visibility still works for non-ignored scripts when `IncludeScriptsPath=true`.

### Doctor tests

Add tests in `pkg/commands/doctor_test.go`:

- root `.docmgrignore` suppresses nested dependency Markdown;
- docs-root `.docmgrignore` suppresses nested dependency Markdown;
- nested ticket `.docmgrignore` suppresses ticket-local generated files;
- `--ignore-glob` and `--ignore-dir` still work;
- `doctor --doc` remains exact-file validation and should not silently ignore the explicitly requested file unless the product decision says otherwise.

Recommended product decision for `doctor --doc`: explicit single-file validation should validate the file even if ignored, but it can print a note that the file is normally ignored in workspace scans.

### Scenario tests

Add a shell scenario:

```bash
mkdir -p ttmp/2026/06/08/DEMO--ignore/scripts/node_modules/pkg
cat > ttmp/2026/06/08/DEMO--ignore/index.md <<'EOF'
---
Title: Demo
Ticket: DEMO
Status: active
Topics: [docmgr]
DocType: index
Intent: ticket-specific
---
# Demo
EOF
cat > ttmp/2026/06/08/DEMO--ignore/scripts/node_modules/pkg/README.md <<'EOF'
# Package README without frontmatter
EOF
cat > ttmp/.docmgrignore <<'EOF'
node_modules/
EOF

docmgr doctor --ticket DEMO --fail-on error
```

Expected: no invalid-frontmatter finding for `node_modules/pkg/README.md`.

## Migration plan

1. Add `github.com/denormal/go-gitignore` to docmgr and create `internal/ignore` with tests.
2. Make `Workspace` load and own the ignore matcher during discovery/construction.
3. Prune ignored directories/files during `Workspace.InitIndex` before frontmatter parsing.
4. Replace doctor helper functions with `ws.IgnoreMatcher()` and remove post-filter compatibility behavior.
5. Ensure list/search/status/export inherit ignore behavior through the workspace index.
6. Add nested `.docmgrignore` loading.
7. Add `docmgr ignore explain`.
8. Update docs and scenario tests.

## Risks and mitigations

### Risk: Existing users rely on ignored files appearing in search

If `.docmgrignore` starts pruning during indexing, ignored files disappear from search/export results. That is likely correct, but it is a behavior change.

Mitigation:

- Document the change.
- Add `ignore explain`.
- Consider a temporary `--no-ignore` debug flag for `workspace export sqlite` or `doc search` if needed.

### Risk: Nested ignore loading walks too much

If nested loading recursively scans dependency folders before it knows they are ignored, it can be slow.

Mitigation:

- Load built-ins/root/docs-root patterns first.
- Use those patterns to prune nested ignore discovery.
- Always hard-skip `.git`, `.meta`, and underscore directories.

### Risk: Pattern semantics diverge from Git in surprising ways

Users may expect advanced `.gitignore` behavior.

Mitigation:

- Document the supported subset.
- Warn on unsupported syntax.
- Keep matcher decisions explainable.
- Consider a third-party gitignore library if the subset grows too complex.

### Risk: Negation complicates directory pruning

Git's negation rules have a sharp edge: if a parent directory is ignored, a later negation cannot re-include a child unless the parent path remains traversable. A naive pruning implementation can make `!keep.md` impossible.

Mitigation:

- Either postpone negation, or implement a conservative `MayContainReinclude(dir)` check before pruning.
- For phase one, it is acceptable to reject negation with a warning if the docs do not promise it.

## Alternatives considered

### Keep the minimal matcher patch only

A small patch can make `node_modules/` match nested path segments. This fixes the immediate case, but it leaves ignore behavior in `doctor`, keeps post-filter timing, and does not help nested ignore files or other commands.

### Use only `--ignore-dir node_modules`

This avoids pattern semantics but requires users to remember command-line flags and does not honor `.docmgrignore` as documented. It is useful as a compatibility input, not as the primary system.

### Import a mature gitignore library

A mature library would reduce semantic bugs. The downside is an extra dependency and the need to adapt explanation output. This remains a good option if implementing `**`, anchoring, and negation becomes time-consuming.

### Store ignored paths in SQLite and filter in SQL

This would make debugging possible, but it still parses ignored Markdown and stores rows for files users asked to exclude. It is useful for optional diagnostics, not as the main behavior.

## Open questions

1. Should negation (`!pattern`) be implemented in the first release, or should docmgr warn that it is unsupported?
2. Should `doctor --doc ignored.md` validate the file anyway because it was explicitly requested? This guide recommends yes.
3. Should built-in ignores be always-on, or should `.docmgrignore` be the only source of user-visible exclusions?
4. Should `docmgr ignore explain` show all loaded patterns by default, or only matching patterns unless `--verbose` is set?
5. Should the HTTP API expose ignore decisions in index refresh responses?

## API and file reference checklist

Start code review in this order:

1. `internal/ignore/*`: new matcher, parser, loader, explanation API.
2. `internal/documents/walk.go`: verify `WithSkipDir` remains generic and simple.
3. `internal/workspace/index_builder.go`: verify ignored dirs are pruned before `ReadDocumentWithFrontmatter`.
4. `internal/workspace/skip_policy.go`: verify hard skips and path tags still do distinct jobs.
5. `pkg/commands/doctor.go`: verify doctor no longer owns pattern semantics.
6. `pkg/commands/ignore.go`: verify `ignore explain` is wired and useful.
7. `pkg/doc/*.md`: verify user docs match actual pattern behavior.
8. `test-scenarios/testing-doc-manager/*`: verify end-to-end behavior catches regressions.

## References

- `internal/workspace/workspace.go:37-90` — workspace root/config/repo discovery.
- `internal/documents/walk.go:15-56` — document walking and skip hook.
- `internal/workspace/index_builder.go:24-73` — index initialization lifecycle.
- `internal/workspace/index_builder.go:139-185` — current ingest loop that parses Markdown files.
- `internal/workspace/skip_policy.go:21-35` — current hard-coded ingest skip policy.
- `internal/workspace/query_docs.go:17-26` — `QueryDocs` requires initialized index.
- `internal/workspace/query_docs.go:331-342` — query visibility and diagnostics options.
- `internal/workspace/query_docs_sql.go:31-40` — SQL visibility filters for archive/scripts/control docs.
- `internal/workspace/discovery.go:12-18` — missing-index detection purpose.
- `internal/workspace/discovery.go:20-54` — missing-index filesystem walk and skip callback.
- `pkg/commands/doctor.go:218-235` — current root/docs-root `.docmgrignore` loading.
- `pkg/commands/doctor.go:357-382` — current doctor query and post-filter flow.
- `pkg/commands/doctor.go:820-965` — current doctor-local ignore helper functions.
- `pkg/doc/docmgr-how-to-setup.md:502-531` — documented `.docmgrignore` examples and contract.
- `pkg/doc/docmgr-cli-guide.md:426-428` — CLI guide statement about ignore files.
- `pkg/doc/docmgr-doctor-validation-workflow.md:45-53` — current post-filter architecture description.


## 2026-06-08 update: simplified implementation direction

After reviewing the complexity tradeoff, the preferred implementation is now the direct workspace-owned `go-gitignore` cutover. The earlier opt-in `BuildIndexOptions.IgnoreMatcher` migration path is intentionally superseded. The final implementation should be simpler because it removes dual behavior: workspace discovery loads ignore policy once, workspace indexing always respects it, and commands stop carrying their own ignore parsing logic.

The only compatibility behavior to keep initially is explicit single-file validation: `docmgr doctor --doc path/to/ignored.md` should still validate that file because the user named it directly. Workspace scans, ticket scans, list/search/export, and missing-index/duplicate-index walks should all treat ignored paths as absent.
