---
Title: Path normalization design
Ticket: DOCMGR-PATH-NORMALIZE
Status: draft
Topics:
    - docmgr
    - search
    - developer-experience
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: docmgr/pkg/commands/relate.go
      Note: Needs to normalize all --file-note inputs before writing RelatedFiles
    - Path: docmgr/pkg/commands/search.go
      Note: --file/--dir filters must use canonicalized comparisons
    - Path: docmgr/internal/workspace/config.go
      Note: Provides roots (repo, docs, config) required for candidate resolution
ExternalSources: []
Summary: >
    Proposes a canonical path resolution layer that can normalize relate/search
    inputs, simplify absolute paths, and provide fuzzy matching so docmgr can
    reliably connect documents to code regardless of the user’s working
    directory.
LastUpdated: 2025-11-26
---

# Path normalization design

## Executive Summary

We will introduce a reusable path-resolution package that turns any user-supplied
path (absolute or relative) into a canonical representation anchored on the git
repository root. `docmgr relate` will call this layer before persisting
RelatedFiles, and `docmgr search --file/--dir` will normalize both stored data
and the query before comparing. This fixes today’s brittle substring matching,
enables fuzzy lookups, and prevents host-specific absolute paths from leaking
into frontmatter.

## Problem Statement

- Related files are stored exactly as typed. `../../foo/bar.go` and
  `pkg/foo/bar.go` become distinct entries even though they reference the same
  file.
- Reverse lookup (`docmgr search --file foo.go`) uses plain substring checks, so
  even simple variations in relative roots miss matches entirely.
- There is no shared logic for resolving a path relative to the git root,
  ticket workspace, `.ttmp.yaml`, or the parent of the docs root. Users cannot
  predict which variant will work.

## Proposed Solution

### 1. Path resolution layer

Create a new package `internal/paths` that exposes:

- `type ResolutionContext struct { RepoRoot, DocsRoot, ConfigDir, TTMPParent, DocDir string }`
- `type NormalizedPath struct { Abs, RepoRelative, DocsRelative, DocRelative string; Segments []string; Exists bool }`
- `func NewResolutionContext(docPath string, docsRoot string) (*ResolutionContext, error)`  
  - Repo root comes from `workspace.FindRepositoryRoot()`.  
  - Config dir is `filepath.Dir(.ttmp.yaml)`.  
  - TTMP parent is `filepath.Dir(docsRoot)`.
- `func (c *ResolutionContext) Normalize(raw string) (*NormalizedPath, error)`
  - Step 1: trim spaces, replace backslashes with `/`, collapse repeated `/`.
  - Step 2: if path is absolute, call `filepath.Clean`, attempt
    `filepath.EvalSymlinks`, and derive `RepoRelative` via `filepath.Rel(repoRoot, abs)`.
  - Step 3: if relative, build candidate bases in priority order:
    1. Directory containing the active document (if known)
    2. Git repo root
    3. `.ttmp.yaml` directory (config dir)
    4. Docs root
    5. Parent of docs root
  - Step 4: join each base with the relative input, call `filepath.Clean`, test
    `os.Stat`. The first existing file wins.
  - Step 5: If nothing exists, still produce a `NormalizedPath` anchored to repo
    root (best-effort) so we can store consistent relative strings even for files
    that aren’t committed yet.

### 2. Canonical storage rules

- RelatedFiles should store `NormalizedPath.RepoRelative` when available.  
  - If repo root is unavailable (non-git workspace), fallback to path relative to
    docs root.  
  - Keep the original user-provided string in the note if we want a breadcrumb.
- Absolute inputs are simplified to eliminate `/./`, `/../`, and tilde expansion.
- `--remove-files` should accept any equivalent path by normalizing each removal
  argument before matching against stored entries.

### 3. Search enhancements

- Normalize the `--file` argument using the same resolver.  
  - Compute `candidateKeys := {query.RepoRelative, query.Abs, query.DocRelative, last N segments}`.
- For each document:
  - Normalize every `RelatedFiles.Path` once per document (cache inside run).
  - Match when any of the following is true:
    - Exact equality on repo-relative strings.
    - One path is a suffix of the other when split into segments and the suffix
      length ≥ 2 segments.
    - Directory queries (`--dir`) match when the normalized repo-relative path
      has the query as a prefix.
- Surface the matched canonical path and note to the user so they can see what
  string is stored.

### 4. API surface changes

- Add a hidden flag `--path-debug` to dump the normalization decision tree when
  DOCMGR_DEBUG is set.
- Emit warnings when we detect absolute paths with user home prefixes to confirm
  that they are being converted.

## Design Decisions

- **Canonical form**: repo-relative with forward slashes (POSIX style) so tickets
  render consistently across platforms.
- **Existence checks**: we only block normalization if a path resolves to two
  different existing files. Otherwise we keep trying candidates until we get a
  hit or exhaust our list.
- **Caching**: `ResolutionContext` caches successful joins (`map[string]string`)
  keyed by `(rawInput, baseKey)` to avoid repeated `os.Stat` calls while walking
  hundreds of docs.
- **Separation of concerns**: `internal/paths` does not know about CLI flags—it
  just consumes already-discovered directories. Commands remain thin wrappers.

## Alternatives Considered

1. **Ask users to always supply repo-relative paths**  
   - Rejected because it breaks the “edit in context” workflow and still leaves
     search without fuzzy matching.
2. **Normalize only when writing**  
   - Without normalizing queries, reverse lookup would continue to miss entries
     created before this change.
3. **Store every variant**  
   - Leads to bloat, doesn’t help search, and complicates deduplication logic.

## Implementation Plan

1. `internal/paths` package  
   - [ ] Implement `ResolutionContext` discovery helpers (repo root, config dir,
     doc dir).  
   - [ ] Implement `Normalize` + helper to reduce multiple separators, expand `~`.
2. `docmgr relate` updates  
   - [ ] Normalize each `--file-note` path before manipulating the `current`
     map.  
   - [ ] Normalize `--remove-files` inputs and match against canonical keys.  
   - [ ] Add human-readable summary showing the stored canonical path.
3. `docmgr search` updates  
   - [ ] Normalize `settings.File` and `settings.Dir`.  
   - [ ] Normalize every stored `rf.Path` once per document and replace
     `strings.Contains/HasPrefix` with canonical comparisons + suffix fuzzy match.
4. Tests + tooling  
   - [ ] Unit tests for resolver covering every anchor scenario.  
   - [ ] Integration tests for relate/search to prove reverse lookup works when
     you link via doc-relative, repo-relative, or absolute inputs.  
   - [ ] Documentation updates (how-to guide, changelog).

## Open Questions

- Should we persist both canonical and raw inputs for audit purposes?
- How should we treat files that live outside the git repository (e.g., `/tmp`)?
  We may want to keep them as absolute paths but still clean them.
- Do we need to guard against symlink attacks where `../../` escapes the repo?

## References

- [Path handling analysis](../reference/01-path-handling-analysis.md)

