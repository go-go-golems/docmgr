---
Title: 'Bug Report: doc relate fails on non-docmgr markdown files'
Ticket: FIX-FRONTMATTER-VALIDATION
Status: active
Topics:
    - frontmatter
    - validation
    - parsing
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-11T19:05:05.906732726-05:00
---

# Bug Report: doc relate fails on non-docmgr markdown files

## Goal

Document the bug where `docmgr doc relate --doc <path>` fails with a confusing error message when attempting to relate files to a markdown file that is not a docmgr document (lacks frontmatter or is outside a ticket workspace).

## Context

The `doc relate` command is designed to update the `RelatedFiles` field in docmgr document frontmatter. However, it attempts to parse frontmatter from any markdown file path provided via `--doc`, without validating:
1. Whether the file is within a recognized ticket workspace
2. Whether the file has frontmatter before attempting to parse it
3. Whether the file is actually a docmgr document

## Bug Summary

**Command**: `docmgr doc relate --doc <path> --file-note "file.go:reason"`

**Error**: 
```
Error: taxonomy: docmgr.frontmatter.parse/yaml_syntax: /path/to/file.md frontmatter delimiters '---' not found
```

**Root Cause**: The command attempts to parse frontmatter from any markdown file without first checking if it's a valid docmgr document.

## Reproduction Steps

1. Create a directory structure that matches docmgr date pattern but is NOT a ticket workspace:
   ```bash
   mkdir -p ttmp/2025/11/25/profile_editor
   ```

2. Create a plain markdown file (no frontmatter):
   ```bash
   cat > ttmp/2025/11/25/profile_editor/profile_editor_implementation_summary.md <<EOF
   # Profile Editor Implementation Summary
   
   ## Overview
   This document summarizes the implementation...
   EOF
   ```

3. Verify directory is NOT a ticket workspace:
   ```bash
   test -f ttmp/2025/11/25/profile_editor/index.md  # Returns false
   ```

4. Run `doc relate` command:
   ```bash
   docmgr doc relate \
     --doc ttmp/2025/11/25/profile_editor/profile_editor_implementation_summary.md \
     --file-note "backend/file.go:Reason"
   ```

5. **Observed Behavior**: Error occurs immediately
6. **Expected Behavior**: Clear error message indicating file is not a docmgr document or suggesting to use `--ticket` instead

## Error Details

**Error Message**: 
```
Error: taxonomy: docmgr.frontmatter.parse/yaml_syntax: /path/to/file.md frontmatter delimiters '---' not found
```

**Error Location**: 
- Function: `extractFrontmatter` in `internal/documents/frontmatter.go:141`
- Called from: `ReadDocumentWithFrontmatter` in `internal/documents/frontmatter.go:29`
- Called from: `RelateCommand.RunIntoGlazeProcessor` in `pkg/commands/relate.go:360`

**Error Type**: Taxonomy-wrapped error (`docmgr.frontmatter.parse/yaml_syntax`)

## Affected Scenarios

1. **Plain markdown files** (no frontmatter) in directories matching docmgr date pattern
2. **Files outside ticket workspaces** (directory lacks `index.md`)
3. **Legacy markdown files** that predate docmgr frontmatter requirements
4. **User confusion** when accidentally referencing wrong file path

## Impact

- **User Experience**: Confusing error message doesn't help users understand the root cause
- **Workflow Disruption**: Users cannot easily relate files to non-docmgr documents
- **False Positives**: Error suggests frontmatter parsing issue when file simply isn't a docmgr document

## Quick Reference

### Current Behavior
```bash
# Fails with frontmatter parsing error
docmgr doc relate --doc /path/to/plain.md --file-note "file.go:reason"
# Error: frontmatter delimiters '---' not found
```

### Expected Behavior
```bash
# Should provide clear guidance
docmgr doc relate --doc /path/to/plain.md --file-note "file.go:reason"
# Error: File is not a docmgr document (missing frontmatter): /path/to/plain.md
# Suggestion: Use --ticket to relate files to a ticket index, or ensure the file is within a ticket directory.
```

### Workaround
```bash
# Use --ticket instead of --doc for non-docmgr files
docmgr doc relate --ticket TICKET-ID --file-note "file.go:reason"
```

## Related Code Locations

- **Command Implementation**: `pkg/commands/relate.go:185-201` (path resolution), `360` (frontmatter parsing)
- **Frontmatter Parsing**: `internal/documents/frontmatter.go:23-61` (ReadDocumentWithFrontmatter), `121-153` (extractFrontmatter)
- **Ticket Discovery**: `internal/workspace/discovery.go:26-70` (CollectTicketWorkspaces)

## Related Documents

- See analysis document: `analysis/02-code-flow-analysis-frontmatter-validation-failure.md`
- See initial analysis: `analysis/01-frontmatter-validation-and-ticket-recognition-analysis.md`
