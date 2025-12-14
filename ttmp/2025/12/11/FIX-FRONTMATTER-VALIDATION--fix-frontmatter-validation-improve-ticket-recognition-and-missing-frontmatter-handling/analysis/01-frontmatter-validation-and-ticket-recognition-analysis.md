---
Title: Frontmatter Validation and Ticket Recognition Analysis
Ticket: FIX-FRONTMATTER-VALIDATION
Status: active
Topics:
    - frontmatter
    - validation
    - parsing
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-11T14:55:52.570547458-05:00
---

# Frontmatter Validation and Ticket Recognition Analysis

## Problem Statement

When running `docmgr doc relate --doc <path>` on a markdown file that is not part of a docmgr ticket workspace (or lacks frontmatter), the command fails with:

```
Error: taxonomy: docmgr.frontmatter.parse/yaml_syntax: /path/to/file.md frontmatter delimiters '---' not found
```

This occurs even when the file path is outside a recognized docmgr ticket directory structure. The error suggests that docmgr is attempting to parse frontmatter from files that may not be docmgr documents at all.

## Root Cause Analysis

### 1. How Tickets Are Discovered

**Location**: `internal/workspace/discovery.go`

The `CollectTicketWorkspaces` function discovers tickets by:

1. Walking the docs root directory (`ttmp/` by default)
2. Looking for directories that contain an `index.md` file
3. Attempting to parse the `index.md` frontmatter
4. If parsing succeeds, the directory is considered a ticket workspace
5. If parsing fails, it's still recorded but with `FrontmatterErr` set

**Key Code**:
```go
indexPath := filepath.Join(path, "index.md")
if fi, err := os.Stat(indexPath); err == nil && !fi.IsDir() {
    doc, _, err := documents.ReadDocumentWithFrontmatter(indexPath)
    if err != nil {
        workspaces = append(workspaces, TicketWorkspace{Path: path, FrontmatterErr: err})
    } else {
        workspaces = append(workspaces, TicketWorkspace{Path: path, Doc: doc})
    }
    return fs.SkipDir
}
```

**Issues**:
- Discovery happens by directory structure (presence of `index.md`), not by validating the file is actually a docmgr document
- Directories starting with `_` are skipped, but other non-ticket directories with `index.md` are still processed
- No validation that the directory structure matches expected ticket patterns

### 2. How `doc relate` Resolves Document Paths

**Location**: `pkg/commands/relate.go` (lines 180-201)

When `--doc` flag is provided:

1. The path is taken directly from `settings.Doc` (line 186)
2. Path is converted to absolute (line 198)
3. No validation that the path is within a recognized ticket workspace
4. No check if the file has frontmatter before attempting to parse

**Key Code**:
```go
if settings.Doc != "" {
    targetDocPath = settings.Doc
} else {
    // ... ticket resolution logic
}

targetDocPath, err = filepath.Abs(targetDocPath)
if err != nil {
    return fmt.Errorf("failed to resolve document path: %w", err)
}

// Later, line 360:
doc, content, err := documents.ReadDocumentWithFrontmatter(targetDocPath)
if err != nil {
    return err
}
```

**Issues**:
- No validation that the document path is within a recognized ticket workspace
- No check if file exists before parsing
- No graceful handling for files without frontmatter

### 3. Frontmatter Parsing Behavior

**Location**: `internal/documents/frontmatter.go`

The `extractFrontmatter` function requires:

1. File must start with `---` on the first line
2. Must have a closing `---` delimiter
3. Returns error "frontmatter delimiters '---' not found" if not found

**Key Code** (lines 120-142):
```go
func extractFrontmatter(raw []byte) ([]byte, []byte, int, error) {
    lines := bytes.Split(raw, []byte("\n"))
    if len(lines) == 0 {
        return nil, nil, 0, fmt.Errorf("empty file")
    }

    start := -1
    end := -1
    for i, line := range lines {
        if i == 0 && bytes.Equal(bytes.TrimSpace(line), []byte("---")) {
            start = i
            continue
        }
        if start >= 0 && bytes.Equal(bytes.TrimSpace(line), []byte("---")) {
            end = i
            break
        }
    }

    if start != 0 || end <= start {
        return nil, nil, 0, fmt.Errorf("frontmatter delimiters '---' not found")
    }
    // ...
}
```

**Issues**:
- Strict requirement: frontmatter must start at line 0
- No tolerance for files without frontmatter (e.g., plain markdown files)
- Error message doesn't distinguish between "not a docmgr doc" vs "malformed frontmatter"

### 4. How Other Commands Handle Missing Frontmatter

**Search Command** (`pkg/commands/search.go`):
- Uses `readDocumentWithContent` which calls `ReadDocumentWithFrontmatter`
- On error, silently skips the file (line 290): `return nil // Skip files with invalid frontmatter`
- This is appropriate for search, but not for operations that require valid documents

**List Docs Command** (`pkg/commands/list_docs.go`):
- Uses `readDocumentFrontmatter` 
- On error, silently skips (line 338): `return nil`
- Skips `index.md` files (line 332)

**Doctor Command** (`pkg/commands/doctor.go`):
- Validates all markdown files within ticket workspaces
- Reports frontmatter errors as issues (lines 611-629)
- Only validates files within discovered ticket workspaces

## Key Findings

### Finding 1: Ticket Recognition is Based on Directory Structure, Not Content

Tickets are discovered by:
- Presence of `index.md` in a directory
- Directory name doesn't start with `_`
- Not explicitly excluded by skip functions

**Problem**: A directory with `index.md` that isn't a docmgr ticket (e.g., a regular markdown documentation directory) will be treated as a ticket workspace if it's under the docs root.

### Finding 2: `doc relate --doc` Doesn't Validate Document Context

When `--doc` is provided:
- Path is resolved to absolute
- No check if path is within a recognized ticket workspace
- No check if file has frontmatter before parsing
- No distinction between "not a docmgr doc" and "malformed docmgr doc"

**Problem**: Users can accidentally reference files outside docmgr ticket workspaces, leading to confusing errors.

### Finding 3: Frontmatter Parsing is Strict and Doesn't Handle Missing Frontmatter

The parser:
- Requires frontmatter to start at line 0
- Returns a generic error for missing frontmatter
- Doesn't distinguish between "not a docmgr document" and "malformed frontmatter"

**Problem**: Error messages don't help users understand whether:
- The file isn't a docmgr document (should use a different command)
- The file is a docmgr document but missing frontmatter (needs fixing)
- The file has frontmatter but it's malformed (needs fixing)

### Finding 4: Commands Have Inconsistent Error Handling

- **Search/List**: Silently skip files with invalid frontmatter
- **Relate**: Fail immediately on parse error
- **Doctor**: Report errors but continue processing

**Problem**: Inconsistent behavior makes it hard for users to understand what's happening.

## Recommendations

### Recommendation 1: Improve Ticket Recognition

**Option A: Validate Ticket Structure**
- Check for presence of expected ticket subdirectories (`.meta/`, `design/`, etc.)
- Validate that `index.md` contains required frontmatter fields (Ticket, DocType)
- Only consider directories as tickets if they pass validation

**Option B: Use Explicit Ticket Markers**
- Require a `.docmgr-ticket` marker file or metadata
- Only directories with this marker are considered tickets

**Option C: Validate Ticket ID Format**
- Ensure the Ticket field in frontmatter matches expected patterns
- Cross-reference directory name with Ticket field

**Preferred**: Option A + C combination - validate structure and Ticket field consistency.

### Recommendation 2: Add Document Path Validation in `doc relate`

Before parsing frontmatter:

1. **Check if file exists**
2. **Check if path is within a recognized ticket workspace**
   - Use `CollectTicketWorkspaces` to get list of valid ticket paths
   - Verify the document path is within one of these paths
3. **Provide better error messages**:
   - If file doesn't exist: "File not found: <path>"
   - If file is outside ticket workspace: "File is not within a recognized docmgr ticket workspace: <path>. Use --ticket to relate files to a ticket index, or ensure the file is within a ticket directory."
   - If file lacks frontmatter: "File does not appear to be a docmgr document (missing frontmatter): <path>"

### Recommendation 3: Improve Frontmatter Parsing Error Messages

Enhance `extractFrontmatter` to provide more specific errors:

- "File does not start with frontmatter delimiter '---'" (vs "frontmatter delimiters '---' not found")
- "Frontmatter block is not closed (missing closing '---')"
- "File appears to be plain markdown without docmgr frontmatter"

### Recommendation 4: Add Graceful Handling for Missing Frontmatter

For commands that operate on documents:

1. **Detect missing frontmatter early** (before YAML parsing)
2. **Provide actionable guidance**:
   - If file is in a ticket directory: "File is missing frontmatter. Run 'docmgr doc add' to create a proper docmgr document."
   - If file is outside ticket directory: "File is not a docmgr document. Use --ticket to relate files to a ticket index instead."

### Recommendation 5: Standardize Error Handling Across Commands

Create a shared utility for document validation that:
- Checks file existence
- Validates ticket workspace context
- Provides consistent error messages
- Can be used by all commands that operate on documents

## Implementation Strategy

### Phase 1: Immediate Fixes (High Priority)

1. **Add path validation in `doc relate`**
   - Check file exists
   - Validate path is within ticket workspace (when `--doc` is used)
   - Improve error messages

2. **Enhance frontmatter error messages**
   - More specific error messages in `extractFrontmatter`
   - Distinguish between "not a docmgr doc" and "malformed frontmatter"

### Phase 2: Improved Ticket Recognition (Medium Priority)

1. **Enhance `CollectTicketWorkspaces`**
   - Validate ticket structure (presence of `.meta/` or other markers)
   - Validate Ticket field matches expected format
   - Only return directories that pass validation

2. **Add ticket validation utility**
   - Shared function to validate if a directory is a valid ticket workspace
   - Used by discovery and other commands

### Phase 3: Consistent Error Handling (Lower Priority)

1. **Create document validation utility**
   - Shared validation logic
   - Consistent error messages
   - Used across all commands

2. **Update all commands to use shared utilities**
   - Refactor search, list, relate, etc. to use shared validation

## Related Code Locations

### Core Parsing
- `internal/documents/frontmatter.go` - Frontmatter extraction and parsing
- `internal/documents/walk.go` - Document walking utilities

### Ticket Discovery
- `internal/workspace/discovery.go` - `CollectTicketWorkspaces` function
- `pkg/commands/import_file.go` - `findTicketDirectory` function

### Command Implementations
- `pkg/commands/relate.go` - `doc relate` command (lines 158-569)
- `pkg/commands/search.go` - Search command (handles missing frontmatter gracefully)
- `pkg/commands/list_docs.go` - List docs command (skips invalid frontmatter)
- `pkg/commands/doctor.go` - Doctor command (reports frontmatter errors)

### Utilities
- `pkg/commands/document_utils.go` - `readDocumentFrontmatter` and `readDocumentWithContent` helpers

## Test Cases to Consider

1. **File outside ticket workspace**
   - `docmgr doc relate --doc /path/to/regular.md --file-note "file.go:reason"`
   - Expected: Clear error that file is not within a ticket workspace

2. **File without frontmatter in ticket workspace**
   - `docmgr doc relate --doc ttmp/.../ticket/plain.md --file-note "file.go:reason"`
   - Expected: Error suggesting to use `doc add` to create proper document

3. **File with malformed frontmatter**
   - `docmgr doc relate --doc ttmp/.../ticket/malformed.md --file-note "file.go:reason"`
   - Expected: Specific error about frontmatter syntax issues

4. **Non-existent file**
   - `docmgr doc relate --doc /nonexistent/file.md --file-note "file.go:reason"`
   - Expected: "File not found" error

5. **Directory with index.md but not a ticket**
   - Create directory structure with `index.md` but missing ticket metadata
   - Expected: Should not be recognized as ticket, or should be recognized but with clear validation errors

## Reproduction Case

### Original Error Scenario

**Command**:
```bash
cd /home/manuel/workspaces/2025-12-01/integrate-moments-persistence && \
docmgr doc relate \
  --doc moments/ttmp/2025/11/25/profile_editor/profile_editor_implementation_summary.md \
  --file-note "moments/backend/pkg/prompts/resolver.go:Profile editor resolver updates summary" \
  --file-note "moments/backend/pkg/promptutil/resolve.go:Profile editor prompt resolution behavior"
```

**Error**:
```
Error: taxonomy: docmgr.frontmatter.parse/yaml_syntax: /home/manuel/workspaces/2025-12-01/integrate-moments-persistence/moments/ttmp/2025/11/25/profile_editor/profile_editor_implementation_summary.md frontmatter delimiters '---' not found
```

### Directory Structure Analysis

**Location**: `/home/manuel/workspaces/2025-12-01/integrate-moments-persistence/moments/ttmp/2025/11/25/profile_editor/`

**Contents**:
- `profile_editing_functional_spec.md` (plain markdown, no frontmatter)
- `profile_editor_implementation_plan.md` (plain markdown, no frontmatter)
- `profile_editor_implementation_summary.md` (plain markdown, no frontmatter)
- **No `index.md` file** - directory is NOT a recognized ticket workspace

**File Content** (first 5 lines):
```markdown
# Profile Editor Implementation Summary

## Overview

This document summarizes the implementation of the Profile Editor feature...
```

**Key Observations**:
1. Directory structure matches docmgr date pattern (`ttmp/2025/11/25/`)
2. Directory does NOT contain `index.md` - not a ticket workspace
3. File is plain markdown without frontmatter (starts with `#` not `---`)
4. Command attempts to parse frontmatter anyway and fails

### Reproduced Error

**Test Setup**:
```bash
# Copied directory structure to test workspace
mkdir -p ttmp/2025/11/25/profile_editor
cp <original-file> ttmp/2025/11/25/profile_editor/profile_editor_implementation_summary.md

# Verified no index.md exists
test -f ttmp/2025/11/25/profile_editor/index.md  # Returns false

# Reproduced error
go run cmd/docmgr/main.go doc relate \
  --doc ttmp/2025/11/25/profile_editor/profile_editor_implementation_summary.md \
  --file-note "test.go:test"
```

**Result**: Same error reproduced:
```
Error: taxonomy: docmgr.frontmatter.parse/yaml_syntax: /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr/ttmp/2025/11/25/profile_editor/profile_editor_implementation_summary.md frontmatter delimiters '---' not found
```

**Conclusion**: The error occurs because:
1. `doc relate --doc` doesn't validate that the path is within a recognized ticket workspace
2. It attempts to parse frontmatter from any markdown file
3. The file doesn't have frontmatter (it's plain markdown)
4. The error message doesn't help the user understand the root cause

## Next Steps

1. Review and approve this analysis
2. Prioritize recommendations
3. Implement Phase 1 fixes (path validation and error messages)
4. Test with the reported error case ✅ (reproduced successfully)
5. Consider Phase 2 improvements based on feedback
