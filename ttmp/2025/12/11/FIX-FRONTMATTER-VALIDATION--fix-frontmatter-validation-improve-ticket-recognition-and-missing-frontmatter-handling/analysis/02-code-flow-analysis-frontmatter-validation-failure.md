---
Title: 'Code Flow Analysis: Frontmatter Validation Failure'
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
LastUpdated: 2025-12-11T19:05:05.906732726-05:00
---

# Code Flow Analysis: Frontmatter Validation Failure

## Overview

This document traces the exact code execution path that leads to the frontmatter validation error when `docmgr doc relate --doc <path>` is called on a non-docmgr markdown file. It identifies the specific lines of code where validation is missing and where errors occur.

## Execution Flow

### Step 1: Command Entry Point

**File**: `pkg/commands/relate.go`  
**Function**: `RelateCommand.RunIntoGlazeProcessor`  
**Line**: 158

```go
func (c *RelateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
```

**What happens**: Command handler receives parsed command-line arguments.

### Step 2: Settings Initialization

**File**: `pkg/commands/relate.go`  
**Lines**: 163-169

```go
settings := &RelateSettings{}
if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
	return fmt.Errorf("failed to parse settings: %w", err)
}

// Apply config root if present
settings.Root = workspace.ResolveRoot(settings.Root)
```

**What happens**: 
- Settings struct is populated from command-line flags
- Root directory is resolved (defaults to `ttmp/`)

**Validation**: ✅ None needed here

### Step 3: Document Path Resolution

**File**: `pkg/commands/relate.go`  
**Lines**: 180-201

```go
// Resolve target document path
var targetDocPath string
var ticketDir string
var err error

if settings.Doc != "" {
	targetDocPath = settings.Doc  // ⚠️ LINE 186: Path taken directly, no validation
} else {
	if settings.Ticket == "" {
		return fmt.Errorf("must specify either --doc or --ticket")
	}
	ticketDir, err = findTicketDirectory(settings.Root, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}
	targetDocPath = filepath.Join(ticketDir, "index.md")
}

targetDocPath, err = filepath.Abs(targetDocPath)  // ⚠️ LINE 198: Only converts to absolute, no validation
if err != nil {
	return fmt.Errorf("failed to resolve document path: %w", err)
}
```

**What happens**:
- If `--doc` is provided, path is taken directly from `settings.Doc` (line 186)
- Path is converted to absolute (line 198)
- **No validation** that:
  - File exists
  - Path is within a recognized ticket workspace
  - File has frontmatter

**Missing Validation**: ❌ File existence check, ticket workspace validation, frontmatter detection

### Step 4: Path Resolver Creation

**File**: `pkg/commands/relate.go`  
**Lines**: 203-207

```go
resolver := paths.NewResolver(paths.ResolverOptions{
	DocsRoot:  settings.Root,
	DocPath:   targetDocPath,
	ConfigDir: configDir,
})
```

**What happens**: Creates a path resolver for normalizing file paths in RelatedFiles.

**Validation**: ✅ None needed here (resolver doesn't validate document)

### Step 5: Suggestion Collection (Optional)

**File**: `pkg/commands/relate.go`  
**Lines**: 209-357

**What happens**: If `--suggest` flag is provided, collects file suggestions. This section handles errors gracefully by skipping files with invalid frontmatter (line 243-244).

**Note**: The suggestion collection code shows the correct pattern - it skips files with frontmatter errors:
```go
doc, err := readDocumentFrontmatter(path)
if err != nil {
	return nil  // Skip files with invalid frontmatter
}
```

**Validation**: ✅ Handles errors gracefully (but only for suggestions, not for target document)

### Step 6: Frontmatter Parsing (ERROR OCCURS HERE)

**File**: `pkg/commands/relate.go`  
**Line**: 360

```go
// Read the target document
doc, content, err := documents.ReadDocumentWithFrontmatter(targetDocPath)  // ⚠️ LINE 360: No pre-validation
if err != nil {
	return err  // ⚠️ LINE 361-362: Error returned immediately, no context
}
```

**What happens**:
- Calls `ReadDocumentWithFrontmatter` without any pre-validation
- If error occurs, returns it immediately
- No distinction between "file not found", "not a docmgr doc", or "malformed frontmatter"

**Missing Validation**: ❌ Pre-check for file existence, frontmatter presence, ticket workspace context

### Step 7: Frontmatter Reading

**File**: `internal/documents/frontmatter.go`  
**Function**: `ReadDocumentWithFrontmatter`  
**Lines**: 23-61

```go
func ReadDocumentWithFrontmatter(path string) (*models.Document, string, error) {
	raw, err := os.ReadFile(path)  // ⚠️ LINE 24: File read, but error only if file doesn't exist
	if err != nil {
		return nil, "", err
	}

	fm, body, fmStartLine, err := extractFrontmatter(raw)  // ⚠️ LINE 29: Calls extractFrontmatter
	if err != nil {
		tax := docmgrctx.NewFrontmatterParseTaxonomy(path, 0, 0, "", err.Error(), err)
		return nil, "", core.WrapWithCause(err, tax)  // ⚠️ LINE 31-32: Wraps error in taxonomy
	}
	// ... rest of parsing
}
```

**What happens**:
- Reads file content (fails only if file doesn't exist)
- Calls `extractFrontmatter` to find frontmatter delimiters
- If `extractFrontmatter` fails, wraps error in taxonomy and returns

**Missing Validation**: ❌ No check if file has frontmatter before calling `extractFrontmatter`

### Step 8: Frontmatter Extraction (ERROR GENERATED HERE)

**File**: `internal/documents/frontmatter.go`  
**Function**: `extractFrontmatter`  
**Lines**: 121-153

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

	if start != 0 || end <= start {  // ⚠️ LINE 140: Error condition
		return nil, nil, 0, fmt.Errorf("frontmatter delimiters '---' not found")  // ⚠️ LINE 141: Generic error
	}
	// ... rest of extraction
}
```

**What happens**:
- Scans file for `---` delimiters
- Requires `---` at line 0 (start)
- Requires closing `---` after start
- If not found, returns generic error: `"frontmatter delimiters '---' not found"`

**Problem**: 
- Generic error message doesn't distinguish between:
  - File without frontmatter (not a docmgr doc)
  - File with malformed frontmatter (missing closing delimiter)
  - File with frontmatter starting at wrong line

**Error Message**: `"frontmatter delimiters '---' not found"`

### Step 9: Error Propagation

**File**: `internal/documents/frontmatter.go`  
**Lines**: 29-32

```go
fm, body, fmStartLine, err := extractFrontmatter(raw)
if err != nil {
	tax := docmgrctx.NewFrontmatterParseTaxonomy(path, 0, 0, "", err.Error(), err)
	return nil, "", core.WrapWithCause(err, tax)
}
```

**What happens**:
- Error from `extractFrontmatter` is wrapped in `FrontmatterParseTaxonomy`
- Taxonomy includes path but no line/column info (both set to 0)
- Error message from `extractFrontmatter` is preserved

**Taxonomy Type**: `docmgr.frontmatter.parse/yaml_syntax`

### Step 10: Error Return to Command

**File**: `pkg/commands/relate.go`  
**Lines**: 360-362

```go
doc, content, err := documents.ReadDocumentWithFrontmatter(targetDocPath)
if err != nil {
	return err  // ⚠️ Error returned as-is, no additional context
}
```

**What happens**:
- Error is returned immediately
- No additional context is added
- User sees taxonomy error message

**Final Error Output**:
```
Error: taxonomy: docmgr.frontmatter.parse/yaml_syntax: /path/to/file.md frontmatter delimiters '---' not found
```

## Code Flow Diagram

```
RelateCommand.RunIntoGlazeProcessor (relate.go:158)
  │
  ├─> Parse settings (relate.go:163-169)
  │
  ├─> Resolve document path (relate.go:185-201)
  │   │
  │   ├─> if --doc: take path directly ❌ NO VALIDATION
  │   └─> Convert to absolute path ❌ NO VALIDATION
  │
  ├─> Create path resolver (relate.go:203-207)
  │
  ├─> [Optional] Collect suggestions (relate.go:209-357)
  │   └─> Skips files with errors ✅ HANDLES ERRORS GRACEFULLY
  │
  └─> Read document frontmatter (relate.go:360) ❌ NO PRE-VALIDATION
      │
      └─> ReadDocumentWithFrontmatter (frontmatter.go:23)
          │
          ├─> Read file (frontmatter.go:24) ✅ Checks file existence
          │
          └─> extractFrontmatter (frontmatter.go:29)
              │
              └─> Scan for --- delimiters (frontmatter.go:121-153)
                  │
                  └─> ERROR: "frontmatter delimiters '---' not found" (frontmatter.go:141)
                      │
                      └─> Wrap in taxonomy (frontmatter.go:31-32)
                          │
                          └─> Return error (relate.go:361-362)
                              │
                              └─> User sees: "Error: taxonomy: docmgr.frontmatter.parse/yaml_syntax: ..."
```

## Missing Validation Points

### 1. Path Validation (relate.go:185-201)

**Location**: After path resolution, before frontmatter parsing

**Missing Checks**:
- [ ] File exists (`os.Stat`)
- [ ] Path is within a recognized ticket workspace
- [ ] File has frontmatter (quick peek at first line)

**Current Behavior**: Path is taken directly and converted to absolute, then immediately used for parsing.

### 2. Frontmatter Detection (frontmatter.go:23-29)

**Location**: Before calling `extractFrontmatter`

**Missing Checks**:
- [ ] Quick check if file starts with `---`
- [ ] Distinguish between "no frontmatter" vs "malformed frontmatter"

**Current Behavior**: Always calls `extractFrontmatter`, which returns generic error if frontmatter not found.

### 3. Ticket Workspace Validation (relate.go:185-201)

**Location**: When `--doc` is provided

**Missing Checks**:
- [ ] Verify path is within a discovered ticket workspace
- [ ] Use `CollectTicketWorkspaces` to get valid ticket paths
- [ ] Check if document path is within any ticket workspace

**Current Behavior**: No validation that document is within a ticket workspace.

## Comparison with Other Commands

### Search Command (search.go:288-291)

```go
doc, content, err := readDocumentWithContent(path)
if err != nil {
	return nil // Skip files with invalid frontmatter ✅ GRACEFUL HANDLING
}
```

**Behavior**: Silently skips files with invalid frontmatter (appropriate for search).

### List Docs Command (list_docs.go:336-339)

```go
doc, err := readDocumentFrontmatter(path)
if err != nil {
	return nil ✅ GRACEFUL HANDLING
}
```

**Behavior**: Silently skips files with invalid frontmatter (appropriate for listing).

### Doctor Command (doctor.go:614-629)

```go
_, err := readDocumentFrontmatter(path)
if err != nil {
	hasIssues = true
	row := types.NewRow(...) ✅ REPORTS AS ISSUE
	// ... emit diagnostic
}
```

**Behavior**: Reports frontmatter errors as issues (appropriate for validation).

### Relate Command (relate.go:360-362)

```go
doc, content, err := documents.ReadDocumentWithFrontmatter(targetDocPath)
if err != nil {
	return err ❌ FAILS IMMEDIATELY
}
```

**Behavior**: Fails immediately with no context (inappropriate - should validate first).

## Root Cause Summary

1. **No pre-validation**: `doc relate --doc` doesn't validate the document path before parsing
2. **Generic error messages**: `extractFrontmatter` returns generic error that doesn't distinguish between "not a docmgr doc" and "malformed frontmatter"
3. **No ticket workspace check**: Command doesn't verify document is within a recognized ticket workspace
4. **Inconsistent error handling**: Other commands handle missing frontmatter gracefully, but `relate` fails immediately

## Recommended Fixes

### Fix 1: Add Path Validation (relate.go:185-201)

```go
if settings.Doc != "" {
	targetDocPath = settings.Doc
	targetDocPath, err = filepath.Abs(targetDocPath)
	if err != nil {
		return fmt.Errorf("failed to resolve document path: %w", err)
	}
	
	// NEW: Validate file exists
	if _, err := os.Stat(targetDocPath); err != nil {
		return fmt.Errorf("file not found: %s", targetDocPath)
	}
	
	// NEW: Validate path is within a ticket workspace
	workspaces, err := workspace.CollectTicketWorkspaces(settings.Root, nil)
	if err == nil {
		isInWorkspace := false
		for _, ws := range workspaces {
			if strings.HasPrefix(targetDocPath, ws.Path) {
				isInWorkspace = true
				break
			}
		}
		if !isInWorkspace {
			return fmt.Errorf("file is not within a recognized docmgr ticket workspace: %s\nSuggestion: Use --ticket to relate files to a ticket index, or ensure the file is within a ticket directory.", targetDocPath)
		}
	}
}
```

### Fix 2: Improve Frontmatter Error Messages (frontmatter.go:121-153)

```go
func extractFrontmatter(raw []byte) ([]byte, []byte, int, error) {
	lines := bytes.Split(raw, []byte("\n"))
	if len(lines) == 0 {
		return nil, nil, 0, fmt.Errorf("empty file")
	}

	// Check first line
	firstLine := bytes.TrimSpace(lines[0])
	if !bytes.Equal(firstLine, []byte("---")) {
		return nil, nil, 0, fmt.Errorf("file does not start with frontmatter delimiter '---' (file appears to be plain markdown without docmgr frontmatter)")
	}

	start := 0
	end := -1
	for i := 1; i < len(lines); i++ {
		if bytes.Equal(bytes.TrimSpace(lines[i]), []byte("---")) {
			end = i
			break
		}
	}

	if end <= start {
		return nil, nil, 0, fmt.Errorf("frontmatter block is not closed (missing closing '---' delimiter)")
	}
	// ... rest of extraction
}
```

### Fix 3: Add Frontmatter Detection Helper

```go
// HasFrontmatter checks if a file likely has frontmatter by peeking at the first line
func HasFrontmatter(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return false, nil
	}
	
	firstLine := bytes.TrimSpace(scanner.Bytes())
	return bytes.Equal(firstLine, []byte("---")), nil
}
```

## Related Files

- `pkg/commands/relate.go` - Command implementation (lines 158-569)
- `internal/documents/frontmatter.go` - Frontmatter parsing (lines 21-210)
- `internal/workspace/discovery.go` - Ticket workspace discovery (lines 26-70)
- `pkg/commands/document_utils.go` - Document reading utilities (lines 8-15)

## Test Cases

1. **File outside ticket workspace** - Should error with clear message
2. **File without frontmatter** - Should error with suggestion to use `--ticket`
3. **File with malformed frontmatter** - Should error with specific frontmatter issue
4. **Non-existent file** - Should error with "file not found"
5. **Valid docmgr document** - Should work as expected
