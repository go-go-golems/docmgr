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

Document the bug where `docmgr doc list --ticket SOME-TICKET` fails to find tickets when run from a workspace root in a multi-repo setup, even though the ticket exists and can be found when running from the project's ttmp directory.

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

## Proposed Fix

### Option 1: Ensure Absolute Path in `RunIntoGlazeProcessor`

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

### Option 2: Ensure `ResolveRoot` Always Returns Absolute Path

Modify `ResolveRoot` in `internal/workspace/config.go` to always return an absolute path by using `filepath.Abs()` on the final resolved path.

### Recommendation

**Option 1** is preferred because:
- It's consistent with the existing `Run` method
- It's a localized fix that doesn't change the behavior of `ResolveRoot` (which may be used elsewhere)
- It ensures the path is absolute right before use

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
