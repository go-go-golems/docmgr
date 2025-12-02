---
Title: 'Bug Report: doc list --ticket fails in multi-repo setup'
Ticket: DOCMGR-BUG-001
Status: active
Topics:
    - bug
    - multi-repo
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-01T15:39:22.592622015-05:00
---

# Bug Report: doc list --ticket fails in multi-repo setup

## Goal

Document two related issues with `docmgr doc list`:

1. **Multi-repo root resolution**: `doc list --ticket` may fail to find tickets when run from a workspace root in a multi-repo setup
2. **Silent skipping of broken documents**: Documents with invalid frontmatter are silently skipped without warnings, making it appear as if no documents exist

## Context

### Setup Description

In a multi-repo workspace setup:

```
workspace/
├── .ttmp.yaml          # Points to project1/ttmp
├── project1/
│   └── ttmp/
│       └── YYYY/MM/DD/
│           └── SOME-TICKET/
│               └── ...
└── project2/
    └── ttmp/
```

Where `workspace/.ttmp.yaml` contains:
```yaml
root: project1/ttmp
```

### Bug Behavior

1. **From `workspace/`**: Running `docmgr doc list --ticket SOME-TICKET` returns no results
2. **From `workspace/project1/ttmp`**: Running `docmgr doc list --ticket SOME-TICKET` correctly finds and lists documents

This indicates that root resolution is not working correctly when running from the workspace root.

## Root Cause Analysis

### Code Flow

1. **Root Resolution** (`internal/workspace/config.go:ResolveRoot`):
   - Finds `.ttmp.yaml` at `workspace/.ttmp.yaml` via `FindTTMPConfigPath()`
   - Reads config with `root: project1/ttmp` (relative path)
   - Resolves to: `filepath.Join(filepath.Dir(cfgPath), cfg.Root)`
   - This should resolve to an absolute path like `/path/to/workspace/project1/ttmp`

2. **Document Listing** (`pkg/commands/list_docs.go`):
   - `RunIntoGlazeProcessor` method (line 125): Calls `workspace.ResolveRoot(settings.Root)`
   - Line 172: Calls `filepath.Walk(settings.Root, ...)` directly without ensuring the path is absolute
   - **Issue**: If `ResolveRoot` returns a relative path, `filepath.Walk` will walk relative to the current working directory, which may not be correct

3. **Comparison with `Run` method**:
   - The `Run` method (human-friendly output) at lines 393-398 does make the root absolute:
   ```go
   absRoot := settings.Root
   if !filepath.IsAbs(absRoot) {
       if cwd, err := os.Getwd(); err == nil {
           absRoot = filepath.Join(cwd, absRoot)
       }
   }
   ```
   - However, `RunIntoGlazeProcessor` (used for structured output) does NOT have this check

### The Problem

The `ResolveRoot` function in `internal/workspace/config.go` may return a relative path in some cases, or the path resolution logic may not always produce an absolute path. When `list_docs.go` uses this path directly in `filepath.Walk` without ensuring it's absolute, the walk happens relative to the current working directory, which can lead to incorrect behavior.

### Specific Issue

In `pkg/commands/list_docs.go:RunIntoGlazeProcessor`:
- Line 125: `settings.Root = workspace.ResolveRoot(settings.Root)`
- Line 172: `filepath.Walk(settings.Root, ...)` - uses root directly without ensuring it's absolute

The `Run` method (line 393-398) has code to make the root absolute, but `RunIntoGlazeProcessor` does not.

## Quick Reference

### Affected Code Locations

1. **`pkg/commands/list_docs.go`**:
   - `RunIntoGlazeProcessor` method (line 114-250): Missing absolute path check before `filepath.Walk`
   - `Run` method (line 255-551): Has absolute path check (lines 393-398) but `RunIntoGlazeProcessor` doesn't

2. **`internal/workspace/config.go`**:
   - `ResolveRoot` function (line 189-257): May return relative paths in some edge cases

### Expected Behavior

When running `docmgr doc list --ticket SOME-TICKET` from `workspace/`:
1. Find `.ttmp.yaml` at `workspace/.ttmp.yaml`
2. Resolve `root: project1/ttmp` to absolute path `/absolute/path/to/workspace/project1/ttmp`
3. Walk that absolute path to find all documents
4. Filter by ticket and return results

### Actual Behavior

The command fails to find documents, likely because:
- The resolved root path is relative or not properly resolved
- `filepath.Walk` is called with a relative path, causing it to walk from the wrong location

## Proposed Fixes

### Fix 1: Ensure Absolute Path in `RunIntoGlazeProcessor`

Add the same absolute path check that exists in `Run` method to `RunIntoGlazeProcessor`:

```go
// After line 125 in RunIntoGlazeProcessor
settings.Root = workspace.ResolveRoot(settings.Root)

// Ensure root is absolute
if !filepath.IsAbs(settings.Root) {
    if cwd, err := os.Getwd(); err == nil {
        settings.Root = filepath.Join(cwd, settings.Root)
    }
}
```

**Rationale**: Consistent with existing `Run` method, localized fix, ensures path is absolute before use.

### Fix 2: Show Warnings for Skipped Documents

Modify `list_docs.go` to collect and display warnings when documents are skipped:

**In `RunIntoGlazeProcessor`**:
```go
var skippedDocs []struct {
    path  string
    error string
}

// In the filepath.Walk callback:
doc, err := readDocumentFrontmatter(path)
if err != nil {
    skippedDocs = append(skippedDocs, struct{path, error string}{
        path: path,
        error: err.Error(),
    })
    docmgr.RenderTaxonomy(ctx, docmgrctx.NewListingSkip("list_docs", path, err.Error(), err))
    return nil
}

// After walk completes, print warnings if any:
if len(skippedDocs) > 0 {
    fmt.Fprintf(os.Stderr, "Warning: Skipped %d document(s) due to frontmatter parsing errors:\n", len(skippedDocs))
    for _, skipped := range skippedDocs {
        relPath, _ := filepath.Rel(settings.Root, skipped.path)
        fmt.Fprintf(os.Stderr, "  - %s: %s\n", relPath, skipped.error)
    }
    fmt.Fprintf(os.Stderr, "\n")
}
```

**In `Run` method**: Similar approach, but integrate warnings into the human-readable output format.

**Rationale**: Makes it immediately clear to users why documents aren't appearing, improving debuggability and user experience.

## Usage Examples

### Reproducing the Bug

```bash
# Setup
cd /path/to/workspace
echo "root: project1/ttmp" > .ttmp.yaml

# Create a ticket in project1/ttmp
cd project1/ttmp
docmgr ticket create-ticket --ticket SOME-TICKET --title "Test Ticket"
docmgr doc add --ticket SOME-TICKET --doc-type reference --title "Test Doc"

# From workspace root - FAILS
cd /path/to/workspace
docmgr doc list --ticket SOME-TICKET
# Returns: No documents found.

# From project1/ttmp - WORKS
cd /path/to/workspace/project1/ttmp
docmgr doc list --ticket SOME-TICKET
# Returns: Lists documents correctly
```

### Expected After Fix

```bash
# From workspace root - SHOULD WORK
cd /path/to/workspace
docmgr doc list --ticket SOME-TICKET
# Should return: Lists documents correctly
```

## Related

- `pkg/commands/list_docs.go` - Implementation of doc list command
- `internal/workspace/config.go` - Root resolution logic
- Similar issue may affect other commands that use `ResolveRoot` without ensuring absolute paths

## Additional Finding: Silent Skipping of Documents with Broken Frontmatter

During investigation of the real-world case (`INTEGRATE-MOMENTS-PERSISTENCE` ticket), a critical usability issue was discovered:

### Problem

Documents with broken or invalid frontmatter are **silently skipped** in `doc list` output, making it appear as if no documents exist when they actually do.

**Example**: Documents with legacy `RelatedFiles` format (scalar strings) fail to parse:
```yaml
RelatedFiles:
    - go-go-mento/go/pkg/webchat/turns_persistence.go
    - go-go-mento/go/pkg/persistence/turns/repo.go
```

**Error**: `yaml: unmarshal errors: cannot unmarshal !!str into models.RelatedFile`

**Current behavior**: Documents are silently skipped with no indication to the user:
- `doc list --ticket INTEGRATE-MOMENTS-PERSISTENCE` returns "No documents found."
- User has no way to know that documents exist but were skipped due to parsing errors
- Only `docmgr doctor` reveals the frontmatter parsing errors

**Impact**: 
- Confusing user experience - appears as if documents don't exist
- Difficult to debug why documents aren't showing up
- Users must run `doctor` separately to discover parsing issues

### Root Cause

In `pkg/commands/list_docs.go`, when `readDocumentFrontmatter` fails:

```go
doc, err := readDocumentFrontmatter(path)
if err != nil {
    docmgr.RenderTaxonomy(ctx, docmgrctx.NewListingSkip("list_docs", path, err.Error(), err))
    return nil  // Silently skip
}
```

The diagnostic taxonomy is rendered, but:
1. In human-readable output (`Run` method), these diagnostics may not be visible
2. In structured output (`RunIntoGlazeProcessor`), diagnostics are rendered but may be ignored
3. No clear warning message is shown to the user

### Required Behavior

When `doc list` encounters documents with broken frontmatter, it should:

1. **Show warnings** indicating which documents were skipped and why
2. **Continue listing** other valid documents
3. **Make it clear** that some documents exist but couldn't be parsed

### Proposed Solution

Add warning output to `doc list` when documents are skipped:

**For human-readable output** (`Run` method):
- Print warnings to stderr before the document listing
- Format: `Warning: Skipped document due to frontmatter error: <path> - <error>`

**For structured output** (`RunIntoGlazeProcessor`):
- Include skipped documents in output with a `skipped: true` field and `error` field
- Or render diagnostics more prominently

**Example output**:
```
Warning: Skipped 2 documents due to frontmatter parsing errors:
  - moments/ttmp/.../reference/persistence-port-analysis.md: yaml: unmarshal errors: ...
  - moments/ttmp/.../reference/web-port-analysis.md: yaml: unmarshal errors: ...

## Documents (0)

No documents found.
```

This would immediately alert users that documents exist but have parsing issues, making debugging much easier.
