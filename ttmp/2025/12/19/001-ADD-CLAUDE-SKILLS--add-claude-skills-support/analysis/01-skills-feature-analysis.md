---
Title: Skills Feature Analysis
Ticket: 001-ADD-CLAUDE-SKILLS
Status: active
Topics:
    - features
    - skills
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/doc/docmgr-codebase-architecture.md
      Note: Reference architecture documentation created during analysis
    - Path: pkg/doc/docmgr-how-to-add-cli-verbs.md
      Note: Implementation guide created during analysis
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T12:39:41.562234764-05:00
---


# Skills Feature Analysis

## Overview

This document analyzes the docmgr codebase to understand how to implement a "skills" feature. Skills are markdown documents that describe capabilities or knowledge areas, similar to documents but with additional preamble fields for "what this skill is for" and "when this skill should be used". Skills can be located in:
- Workspace root under `/skills` directory
- Ticket-specific directories under `<ticket>/skills`

The implementation requires:
1. `docmgr skill list` - Lists all skills with: what it's for, when to use, topics, and related paths. Supports filtering by ticket, topics, file paths, and directories.
2. `docmgr skill show <skill>` - Shows detailed information about a specific skill

## Codebase Architecture Analysis

### Command Structure

docmgr uses a hierarchical command structure built on:
- **Cobra** for CLI command parsing and registration
- **Glazed** framework for structured output (JSON/YAML/CSV) and dual-mode commands (human-friendly + scriptable)

#### Command Registration Pattern

Commands are organized in `cmd/docmgr/cmds/` with subdirectories for each command group:
- `doc/` - Document operations (`doc add`, `doc list`, `doc search`)
- `vocab/` - Vocabulary management (`vocab list`, `vocab add`)
- `ticket/` - Ticket operations
- `meta/` - Metadata operations

Each command group has an `Attach()` function that registers subcommands:

```go
// Pattern from cmd/docmgr/cmds/doc/doc.go
func Attach(root *cobra.Command) error {
    docCmd := &cobra.Command{
        Use:   "doc",
        Short: "Document workspace operations",
    }
    
    listCmd, err := newListCommand()
    // ... create other commands
    
    docCmd.AddCommand(listCmd, ...)
    root.AddCommand(docCmd)
    return nil
}
```

#### Command Implementation Pattern

Commands follow a dual-mode pattern:
1. **BareCommand** interface - Human-friendly output
2. **GlazeCommand** interface - Structured output for scripts

Example from `pkg/commands/vocab_list.go`:
- `Run()` method implements `BareCommand` - prints formatted text
- `RunIntoGlazeProcessor()` method implements `GlazeCommand` - outputs structured rows

Commands use a `Settings` struct with `glazed.parameter` tags for type-safe parameter access:

```go
type VocabListSettings struct {
    Category string `glazed.parameter:"category"`
    Root     string `glazed.parameter:"root"`
}
```

### Document Model

Documents are defined in `pkg/models/document.go`:

```go
type Document struct {
    Title           string
    Ticket          string
    Status          string
    Topics          []string
    DocType         string
    Intent          string
    Owners          []string
    RelatedFiles    RelatedFiles
    ExternalSources []string
    Summary         string
    LastUpdated     time.Time
}
```

**Key observations:**
- Documents use YAML frontmatter for metadata
- `RelatedFiles` is a structured list with `Path` and optional `Note`
- Frontmatter parsing is handled by `internal/documents/frontmatter.go`
- Documents support vocabulary validation for Topics, DocTypes, Intent, Status

### Workspace Discovery

Workspace discovery is handled by `internal/workspace/discovery.go` and `internal/workspace/workspace.go`:

**Key functions:**
- `workspace.DiscoverWorkspace()` - Discovers workspace root and configuration
- `workspace.InitIndex()` - Builds SQLite index of all documents
- `workspace.QueryDocs()` - Queries documents using filters (ticket, status, doc-type, topics, related files)

**Workspace resolution order:**
1. `--root` flag
2. `.ttmp.yaml` in current directory
3. `.ttmp.yaml` in parent directories (walk up)
4. `DOCMGR_ROOT` environment variable
5. Git repository root: `<git-root>/ttmp`
6. Default: `ttmp` in current directory

### Document Walking

`internal/documents/walk.go` provides `WalkDocuments()` function:
- Walks directory tree recursively
- Skips directories starting with `_` by default
- Invokes callback for each `.md` file found
- Reads frontmatter using `ReadDocumentWithFrontmatter()`

### Vocabulary System

Vocabulary is managed via `pkg/commands/vocabulary.go`:
- Loads from `vocabulary.yaml` (resolved via workspace config)
- Supports categories: `topics`, `docTypes`, `intent`, `status`
- Each entry has `slug` and `description`
- Vocabulary is used for validation and suggestions

### Related Files

`RelatedFiles` is a structured list stored in document frontmatter:
- Format: `- Path: path/to/file.go\n  Note: Description`
- Legacy format (scalar strings) still supported for backward compatibility
- Normalized and indexed in SQLite for fast queries
- Supports multiple path representations (repo-relative, docs-relative, absolute)

## Skills Implementation Design

### Skills as Documents

Skills should be treated as a special type of document with:
- Same frontmatter structure as regular documents
- Additional preamble fields:
  - `WhatFor` - What this skill is for
  - `WhenToUse` - When this skill should be used
- DocType: `skill` (should be added to vocabulary)
- Location: `/skills` directory (workspace root) or `<ticket>/skills`

### File Structure

```
ttmp/
  skills/
    01-skill-name.md
    02-another-skill.md
  <ticket>/
    skills/
      01-ticket-specific-skill.md
```

### Skill Document Format

```yaml
---
Title: Skill Name
Ticket: TICKET-ID (optional, for ticket-specific skills)
DocType: skill
Topics:
  - topic1
  - topic2
RelatedFiles:
  - Path: path/to/file.go
    Note: Why this file relates to the skill
WhatFor: What this skill is for (preamble field)
WhenToUse: When this skill should be used (preamble field)
---
```

### Implementation Plan

#### 1. Extend Document Model

Add preamble fields to `pkg/models/document.go`:
- `WhatFor string` - Optional field for skills
- `WhenToUse string` - Optional field for skills

**Note:** These should be optional fields, not required. Regular documents won't have them.

#### 2. Create Skill Commands

Create `cmd/docmgr/cmds/skill/` directory:
- `skill.go` - `Attach()` function
- `list.go` - `newListCommand()` for `skill list`
- `show.go` - `newShowCommand()` for `skill show`

#### 3. Skill List Command

`docmgr skill list` should:
- Query all skills using `workspace.QueryDocs()` with `DocType == "skill"`
- Output columns: `skill`, `what_for`, `when_to_use`, `topics`, `related_paths`, `path`
- Support filtering by:
  - `--ticket` - Filter by ticket ID
  - `--topics` - Filter by topics (OR logic, any topic matches)
  - `--file` - Filter by related file path (skills that reference this file)
  - `--dir` - Filter by directory (skills that reference files in this directory)
- Support structured output (`--with-glaze-output`)

**Implementation notes:**
- Use `workspace.QueryDocs()` with `DocFilters{DocType: "skill"}`
- Use `RelatedFile` filter for `--file` flag (same as `docmgr doc search`)
- Use `RelatedDir` filter for `--dir` flag (same as `docmgr doc search`)
- Skills are automatically indexed in SQLite (no separate discovery needed)

#### 4. Skill Show Command

`docmgr skill show <skill>` should:
- Find skill by name (slug or title match)
- Display full skill information:
  - Title
  - WhatFor
  - WhenToUse
  - Topics
  - RelatedFiles (with notes)
  - Full markdown body
- Handle ambiguity (multiple skills with same name)

### Integration Points

1. **Workspace Discovery**: Use existing `workspace.DiscoverWorkspace()` to find root
2. **Document Parsing**: Use `documents.ReadDocumentWithFrontmatter()` for parsing
3. **Vocabulary**: Add `skill` to `docTypes` vocabulary
4. **Command Pattern**: Follow `vocab/` command pattern (simple list/show operations)

### Open Questions

1. **Skill Naming**: How to identify skills uniquely? By filename slug? By title? Both?
2. **Skill Validation**: Should skills require `WhatFor` and `WhenToUse`? Or make them optional?
3. **Ticket Skills**: Should ticket-specific skills inherit ticket context automatically?
4. **Search Integration**: Should `docmgr doc search` include skills? Or keep them separate?

**Resolved:**
- **Indexing**: Skills are indexed in SQLite automatically (same as regular documents). No separate discovery needed - `skill list` uses `QueryDocs()` with `DocType == "skill"` filter.

### Next Steps

1. Create skill command structure (`cmd/docmgr/cmds/skill/`)
2. Implement `skill list` command with filtering (ticket, topics, file, dir)
3. Implement `skill show` command
4. Add `skill` to vocabulary
5. Test with sample skill documents
6. Update documentation
