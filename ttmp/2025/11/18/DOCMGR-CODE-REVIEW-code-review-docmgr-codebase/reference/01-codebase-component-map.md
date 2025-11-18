---
Title: Codebase Component Map
Ticket: DOCMGR-CODE-REVIEW
Status: active
Topics:
    - docmgr
    - code-review
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-18T09:45:28.984714452-05:00
---


# Codebase Component Map

## Goal

This document maps out the major components of the docmgr codebase, their symbols, files, and functionality. It serves as a reference for code review by identifying all areas that need to be reviewed, without performing the actual review.

## Context

docmgr is a Go CLI tool for managing documentation workspaces in LLM-assisted workflows. It uses:
- Cobra for CLI command structure
- Glazed framework for command definitions and output formatting
- YAML frontmatter for document metadata
- File system operations for workspace management

## Architecture Overview

The codebase follows a standard Go CLI structure:
- `cmd/docmgr/` - Main entry point
- `pkg/commands/` - Individual command implementations
- `pkg/models/` - Data models
- `pkg/doc/` - Embedded documentation
- `pkg/utils/` - Utility functions

## Component Map

### 1. Entry Point & CLI Setup

**File:** `docmgr/cmd/docmgr/main.go`

**Key Symbols:**
- `main()` - Entry point that sets up Cobra root command and registers all subcommands
- Uses Glazed help system integration via `help.NewHelpSystem()` and `doc.AddDocToHelpSystem()`
- Builds Cobra commands from Glazed command descriptions using `cli.BuildCobraCommand()`

**Functionality:**
- Registers all commands (init, configure, create-ticket, list, add, doctor, import, meta, vocab, search, guidelines, relate, renumber, layout-fix, status, changelog, tasks)
- Sets up dual-mode output (human-friendly and structured via Glaze) for certain commands
- Configures help system with embedded documentation

**Review Areas:**
- Command registration order and dependencies
- Error handling in command creation
- Help system integration
- Cobra/Glazed integration patterns

---

### 2. Data Models

**File:** `docmgr/pkg/models/document.go`

**Key Types:**
- `Document` - Core document metadata structure (Title, Ticket, Status, Topics, DocType, Intent, Owners, RelatedFiles, ExternalSources, Summary, LastUpdated)
- `Vocabulary` - Defines valid topics, docTypes, and intent values
- `VocabItem` - Individual vocabulary entry (slug, description)
- `ExternalSource` - Metadata about imported sources
- `TicketDirectory` - Ticket workspace representation
- `RelatedFile` - Single related file with optional note
- `RelatedFiles` - List of RelatedFile with custom YAML unmarshaling

**Key Functions:**
- `RelatedFile.UnmarshalYAML()` - Supports both scalar strings (legacy) and mapping nodes (with Path/Note)
- `RelatedFiles.UnmarshalYAML()` - Handles backward-compatible YAML decoding
- `RelatedFiles.MarshalYAML()` - Always marshals as sequence of objects

**Functionality:**
- Defines the core data structures used throughout the application
- Handles YAML serialization/deserialization with backward compatibility
- Supports both legacy (scalar strings) and new (structured objects) RelatedFiles formats

**Review Areas:**
- YAML unmarshaling logic and edge cases
- Backward compatibility handling
- Data validation (or lack thereof)
- Type safety and error handling

---

### 3. Configuration Management

**File:** `docmgr/pkg/commands/config.go`

**Key Types:**
- `TTMPConfig` - Repository-level configuration structure (Root, Defaults, FilenamePrefixPolicy, Vocabulary)

**Key Functions:**
- `FindTTMPConfigPath()` - Walks up from CWD to find `.ttmp.yaml`, respects `DOCMGR_CONFIG` env var
- `LoadTTMPConfig()` - Loads and parses `.ttmp.yaml`, normalizes relative paths
- `ResolveRoot()` - Resolves docs root path with fallback chain (config → git root → CWD)
- `FindGitRoot()` - Walks up to find `.git` directory, handles gitdir: files
- `FindRepositoryRoot()` - Finds repo root via git, go.mod, or doc/ markers
- `DetectMultipleTTMPRoots()` - Detects multiple ttmp directories in parent hierarchy
- `ResolveVocabularyPath()` - Resolves vocabulary.yaml path with priority chain

**Functionality:**
- Manages repository-level configuration via `.ttmp.yaml`
- Handles path resolution with multiple fallback strategies
- Supports both absolute and relative paths
- Integrates with Git repository detection

**Review Areas:**
- Path resolution logic and edge cases
- Error handling for missing configs
- Fallback chain correctness
- Git repository detection robustness
- Multiple root detection logic

---

### 4. Workspace Discovery & Management

**File:** `docmgr/pkg/commands/workspaces.go`

**Key Types:**
- `TicketWorkspace` - Represents a discovered workspace with path, document, and frontmatter error

**Key Functions:**
- `collectTicketWorkspaces()` - Walks docs root, finds directories with `index.md`, skips `_`-prefixed dirs
- `collectTicketScaffoldsWithoutIndex()` - Finds directories with workspace markers but missing `index.md`
- `hasWorkspaceScaffold()` - Checks for workspace structure markers (design, reference, playbooks, scripts, sources, various, archive, .meta)

**Functionality:**
- Discovers ticket workspaces by walking file system
- Identifies incomplete workspaces (scaffolds without index.md)
- Filters out special directories (those starting with `_`)

**Review Areas:**
- File system traversal efficiency
- Error handling in walk operations
- Skip logic correctness
- Workspace detection heuristics

---

### 5. Command: Init

**File:** `docmgr/pkg/commands/init.go`

**Key Types:**
- `InitCommand` - Command for initializing docs root
- `InitSettings` - Command parameters (Root, SeedVocabulary, Force)

**Key Functions:**
- `NewInitCommand()` - Creates init command with parameter definitions
- `InitCommand.RunIntoGlazeProcessor()` - Executes init logic
- `seedDefaultVocabulary()` - Seeds default vocabulary entries

**Functionality:**
- Creates `ttmp/` directory structure
- Scaffolds templates and guidelines if missing
- Optionally seeds default vocabulary
- Creates vocabulary.yaml with default topics/docTypes/intent

**Review Areas:**
- Directory creation and error handling
- Template scaffolding logic
- Default vocabulary seeding
- Force flag behavior

---

### 6. Command: Configure

**File:** `docmgr/pkg/commands/configure.go`

**Key Types:**
- `ConfigureCommand` - Command for creating/updating `.ttmp.yaml`
- `ConfigureSettings` - Parameters (Root, Owners, Intent, Vocabulary, FilenamePrefixPolicy)

**Key Functions:**
- `NewConfigureCommand()` - Creates configure command
- `ConfigureCommand.RunIntoGlazeProcessor()` - Writes `.ttmp.yaml` configuration

**Functionality:**
- Creates or updates repository-level `.ttmp.yaml` configuration
- Sets defaults for owners, intent, vocabulary path, filename prefix policy

**Review Areas:**
- YAML file writing and formatting
- Configuration validation
- Merge vs overwrite behavior

---

### 7. Command: Create Ticket

**File:** `docmgr/pkg/commands/create_ticket.go`

**Key Types:**
- `CreateTicketCommand` - Command for creating ticket workspaces
- `CreateTicketSettings` - Parameters (Ticket, Title, Topics, Root, Force, PathTemplate)
- `DefaultTicketPathTemplate` - Default path template constant

**Key Functions:**
- `NewCreateTicketCommand()` - Creates command
- `CreateTicketCommand.RunIntoGlazeProcessor()` - Creates ticket directory structure

**Functionality:**
- Creates ticket workspace directory with configurable path template
- Scaffolds standard directory structure (design/, reference/, playbooks/, etc.)
- Creates `index.md`, `tasks.md`, `changelog.md` with frontmatter
- Supports path template placeholders ({{YYYY}}, {{MM}}, {{DD}}, {{DATE}}, {{TICKET}}, {{SLUG}}, {{TITLE}})

**Review Areas:**
- Path template parsing and substitution
- Directory structure creation
- Frontmatter generation
- Template variable handling
- Force flag behavior

---

### 8. Command: Add Document

**File:** `docmgr/pkg/commands/add.go`

**Key Types:**
- `AddCommand` - Command for adding documents to workspaces
- `AddSettings` - Parameters (Ticket, DocType, Title, Root, Topics, Owners, Status, Intent, ExternalSources, Summary, RelatedFiles)

**Key Functions:**
- `NewAddCommand()` - Creates command
- `AddCommand.RunIntoGlazeProcessor()` - Creates new document file

**Functionality:**
- Creates new document in appropriate subdirectory (based on doc-type)
- Generates frontmatter from parameters and ticket defaults
- Applies numeric prefix to filename
- Uses template system for document body

**Review Areas:**
- Document creation logic
- Frontmatter generation and merging
- Filename prefix application
- Template selection and rendering
- Subdirectory creation

---

### 9. Command: List (Tickets & Docs)

**Files:** 
- `docmgr/pkg/commands/list.go` - Base list functionality
- `docmgr/pkg/commands/list_tickets.go` - List tickets command
- `docmgr/pkg/commands/list_docs.go` - List docs command

**Key Types:**
- `ListCommand` - Base list command
- `ListTicketsCommand` - List tickets command
- `ListDocsCommand` - List docs command
- Various Settings structs for each command

**Key Functions:**
- `readDocumentFrontmatter()` - Reads and parses YAML frontmatter from markdown files
- `ListTicketsCommand.RunIntoGlazeProcessor()` - Lists ticket workspaces
- `ListDocsCommand.RunIntoGlazeProcessor()` - Lists documents within tickets

**Functionality:**
- Lists ticket workspaces with metadata
- Lists documents with filtering by ticket, doc-type, status
- Supports structured output via Glaze
- Reads frontmatter from markdown files

**Review Areas:**
- Frontmatter parsing robustness
- Error handling for malformed YAML
- Filtering logic
- Output formatting
- Performance with large workspaces

---

### 10. Command: Search

**File:** `docmgr/pkg/commands/search.go`

**Key Types:**
- `SearchCommand` - Command for searching documents
- `SearchSettings` - Parameters (Query, Ticket, DocType, Status, File, UpdatedSince, UpdatedBefore, WithFiles, Suggest)

**Key Functions:**
- `NewSearchCommand()` - Creates command
- `SearchCommand.RunIntoGlazeProcessor()` - Executes search
- `suggestFiles()` - Suggests files based on search terms
- `suggestFilesFromGit()` - Uses git to find relevant files
- `suggestFilesFromGitStatus()` - Suggests from git status (modified/added files)
- `suggestFilesFromRipgrep()` - Uses ripgrep for file search
- `suggestFilesFromGrep()` - Fallback grep-based search
- `readDocumentWithContent()` - Reads document with full content
- `extractSnippet()` - Extracts relevant snippet from content
- `parseDate()` - Parses date strings for filtering

**Functionality:**
- Full-text search across document content and metadata
- File-based search (finds docs related to specific files)
- Date-based filtering
- File suggestion using multiple strategies (git, ripgrep, grep)
- Snippet extraction with context

**Review Areas:**
- Search algorithm efficiency
- File suggestion heuristics
- Date parsing robustness
- Snippet extraction quality
- Multiple search strategy fallbacks
- Error handling for missing tools (ripgrep, git)

---

### 11. Command: Doctor (Validation)

**File:** `docmgr/pkg/commands/doctor.go`

**Key Types:**
- `DoctorCommand` - Command for validating workspaces
- `DoctorSettings` - Parameters (Root, Ticket, All, IgnoreDirs, IgnoreGlobs, StaleAfterDays, FailOn)

**Key Functions:**
- `NewDoctorCommand()` - Creates command
- `DoctorCommand.RunIntoGlazeProcessor()` - Validates workspaces

**Functionality:**
- Validates document workspaces for issues
- Checks for missing frontmatter, invalid metadata, broken structure
- Validates vocabulary entries
- Checks for stale documents
- Validates RelatedFiles existence
- Respects `.docmgrignore` file for exclusions
- Supports severity levels (none, warning, error)

**Review Areas:**
- Validation rule completeness
- Error categorization
- Ignore file parsing (.docmgrignore)
- Staleness calculation
- File existence checking
- Performance with large workspaces

---

### 12. Command: Meta Update

**File:** `docmgr/pkg/commands/meta_update.go`

**Key Types:**
- `MetaUpdateCommand` - Command for updating document metadata
- `MetaUpdateSettings` - Parameters (Doc, Ticket, DocType, Field, Value, Root)

**Key Functions:**
- `NewMetaUpdateCommand()` - Creates command
- `MetaUpdateCommand.RunIntoGlazeProcessor()` - Updates metadata
- `updateDocumentField()` - Updates specific field in frontmatter
- `findMarkdownFiles()` - Finds markdown files matching criteria

**Functionality:**
- Updates YAML frontmatter fields in documents
- Supports bulk updates (by ticket, doc-type)
- Validates YAML syntax
- Updates LastUpdated timestamp automatically

**Review Areas:**
- YAML parsing and writing correctness
- Field update logic
- Bulk update efficiency
- Error handling for invalid YAML
- Timestamp update logic

---

### 13. Command: Relate

**File:** `docmgr/pkg/commands/relate.go`

**Key Types:**
- `RelateCommand` - Command for relating files to docs/tickets
- `RelateSettings` - Parameters (Ticket, Doc, FileNote, RemoveFiles, Root)

**Key Functions:**
- `NewRelateCommand()` - Creates command
- `RelateCommand.RunIntoGlazeProcessor()` - Manages file relationships
- `appendNote()` - Appends notes to existing file entries

**Functionality:**
- Adds file relationships to document frontmatter
- Supports file notes (required)
- Can remove files from RelatedFiles
- Merges notes when file already exists
- Updates frontmatter with structured RelatedFiles format

**Review Areas:**
- File note merging logic
- Duplicate handling
- Frontmatter update correctness
- Note appending behavior

---

### 14. Command: Tasks Management

**File:** `docmgr/pkg/commands/tasks.go`

**Key Types:**
- `parsedTask` - Represents a parsed task from tasks.md
- `TasksListCommand`, `TasksAddCommand`, `TasksCheckCommand`, `TasksUncheckCommand`, `TasksEditCommand`, `TasksRemoveCommand` - Various task commands
- Corresponding Settings structs

**Key Functions:**
- `loadTasksFile()` - Loads and parses tasks.md
- `parseTasksFromLines()` - Parses markdown checkbox tasks
- `formatTaskLine()` - Formats task line with checkbox
- Various command Run methods for each operation

**Functionality:**
- Manages checkbox tasks in `tasks.md` files
- Supports list, add, check, uncheck, edit, remove operations
- Parses markdown checkbox syntax (`- [ ]` and `- [x]`)
- Preserves file structure and formatting

**Review Areas:**
- Task parsing robustness
- Markdown formatting preservation
- Task ID assignment and tracking
- File editing correctness
- Edge cases (empty files, malformed tasks)

---

### 15. Command: Changelog

**File:** `docmgr/pkg/commands/changelog.go`

**Key Types:**
- `ChangelogUpdateCommand` - Command for updating changelog
- `ChangelogUpdateSettings` - Parameters (Ticket, Entry, FileNote, Suggest, ApplySuggestions, Query, Root)

**Key Functions:**
- `NewChangelogUpdateCommand()` - Creates command
- `ChangelogUpdateCommand.RunIntoGlazeProcessor()` - Updates changelog

**Functionality:**
- Appends entries to `changelog.md`
- Supports file notes with entries
- Can suggest entries based on git history or search query
- Formats entries with timestamps

**Review Areas:**
- Entry formatting
- Timestamp handling
- File note integration
- Suggestion algorithm
- Git integration for suggestions

---

### 16. Command: Vocabulary Management

**Files:**
- `docmgr/pkg/commands/vocabulary.go` - Vocabulary loading/saving
- `docmgr/pkg/commands/vocab_list.go` - List vocabulary command
- `docmgr/pkg/commands/vocab_add.go` - Add vocabulary entry command

**Key Types:**
- `VocabListCommand`, `VocabAddCommand` - Commands
- Corresponding Settings structs

**Key Functions:**
- `LoadVocabulary()` - Loads vocabulary.yaml
- `loadVocabularyFromFile()` - Loads from specific file
- `SaveVocabulary()` - Saves vocabulary to file
- Command Run methods for list/add operations

**Functionality:**
- Manages vocabulary.yaml file
- Lists vocabulary entries by category
- Adds new vocabulary entries (topics, docTypes, intent)
- Validates vocabulary structure

**Review Areas:**
- Vocabulary file parsing
- Entry validation
- Duplicate handling
- File writing correctness

---

### 17. Command: Guidelines

**Files:**
- `docmgr/pkg/commands/guidelines.go` - Guidelines content
- `docmgr/pkg/commands/guidelines_cmd.go` - Guidelines command

**Key Types:**
- `GuidelinesCommand` - Command for displaying guidelines
- `GuidelinesSettings` - Parameters (DocType, List)

**Key Functions:**
- `NewGuidelinesCommand()` - Creates command
- `GuidelinesCommand.RunIntoGlazeProcessor()` - Displays guidelines

**Functionality:**
- Displays guidelines for document types
- Lists available document types
- Loads guidelines from embedded or file system

**Review Areas:**
- Guidelines loading logic
- Document type discovery
- Content formatting

---

### 18. Command: Templates

**File:** `docmgr/pkg/commands/templates.go`

**Key Variables:**
- `TemplateContent` - Map of template content by doc-type

**Key Functions:**
- `GetTemplate()` - Gets template for doc-type
- `loadTemplate()` - Loads template from file system
- `extractFrontmatterAndBody()` - Separates frontmatter from body
- `renderTemplateBody()` - Renders template body with document data

**Functionality:**
- Manages document templates
- Supports embedded templates and file system templates
- Renders templates with document metadata
- Separates frontmatter from body content

**Review Areas:**
- Template rendering logic
- Variable substitution
- Template loading priority
- Error handling for missing templates

---

### 19. Command: Renumber

**File:** `docmgr/pkg/commands/renumber.go`

**Key Types:**
- `RenumberCommand` - Command for resequencing numeric prefixes
- `RenumberSettings` - Parameters (Ticket, Root)

**Key Functions:**
- `NewRenumberCommand()` - Creates command
- `RenumberCommand.RunIntoGlazeProcessor()` - Renumbers files
- `updateTicketReferences()` - Updates internal links after renumbering

**Functionality:**
- Resequences numeric prefixes in filenames (01-, 02-, etc.)
- Updates internal markdown links to reflect new numbering
- Maintains order based on current prefixes

**Review Areas:**
- Prefix parsing and generation
- Link update logic
- Order preservation
- Edge cases (missing prefixes, gaps)

---

### 20. Command: Layout Fix

**File:** `docmgr/pkg/commands/layout_fix.go`

**Key Types:**
- `LayoutFixCommand` - Command for fixing directory layout
- `LayoutFixSettings` - Parameters (Root, DryRun)

**Key Functions:**
- `NewLayoutFixCommand()` - Creates command
- `LayoutFixCommand.RunIntoGlazeProcessor()` - Moves files to subdirectories

**Functionality:**
- Moves documents into appropriate subdirectories based on doc-type
- Updates links in documents
- Supports dry-run mode

**Review Areas:**
- File movement logic
- Link update correctness
- Directory structure validation
- Dry-run accuracy

---

### 21. Command: Status

**File:** `docmgr/pkg/commands/status.go`

**Key Types:**
- `StatusCommand` - Command for workspace status
- `StatusSettings` - Parameters (Root, SummaryOnly, StaleAfter)

**Key Functions:**
- `NewStatusCommand()` - Creates command
- `StatusCommand.RunIntoGlazeProcessor()` - Displays status

**Functionality:**
- Shows overall workspace status
- Counts tickets and documents
- Identifies stale documents
- Provides summary or detailed view

**Review Areas:**
- Status calculation accuracy
- Staleness detection
- Performance with large workspaces
- Summary vs detailed output

---

### 22. Command: Import File

**File:** `docmgr/pkg/commands/import_file.go`

**Key Types:**
- `ImportFileCommand` - Command for importing external files
- `ImportFileSettings` - Parameters (Ticket, File, Root)

**Key Functions:**
- `NewImportFileCommand()` - Creates command
- `ImportFileCommand.RunIntoGlazeProcessor()` - Imports files
- `findTicketDirectory()` - Finds ticket directory

**Functionality:**
- Imports external files into ticket workspace
- Copies files to sources/ directory
- Updates document metadata with import information

**Review Areas:**
- File copying logic
- Metadata update
- Duplicate handling
- Error handling for missing files

---

### 23. Filename Prefix Management

**File:** `docmgr/pkg/commands/filename_prefix.go`

**Key Variables:**
- `numericPrefixRe` - Regex for numeric prefixes

**Key Functions:**
- `hasNumericPrefix()` - Checks if filename has numeric prefix
- `stripNumericPrefix()` - Removes prefix, returns base name and prefix info
- `nextPrefixForDir()` - Calculates next prefix for directory
- `buildPrefixedDocPath()` - Builds path with appropriate prefix

**Functionality:**
- Manages numeric prefixes (01-, 02-, etc.) for document filenames
- Supports 2-digit and 3-digit prefixes
- Calculates next available prefix
- Maintains ordering

**Review Areas:**
- Prefix calculation logic
- Edge cases (99 → 100 transition)
- Directory scanning efficiency
- Prefix format validation

---

### 24. Scaffolding

**File:** `docmgr/pkg/commands/scaffold.go`

**Key Functions:**
- `writeFileIfNotExists()` - Writes file only if it doesn't exist
- `scaffoldTemplatesAndGuidelines()` - Scaffolds template and guideline files

**Functionality:**
- Creates template and guideline files in workspace
- Prevents overwriting existing files
- Sets up standard directory structure

**Review Areas:**
- File creation logic
- Directory structure creation
- Force flag handling

---

### 25. Utilities

**File:** `docmgr/pkg/utils/slug.go`

**Key Functions:**
- `Slugify()` - Converts string to URL-friendly slug

**Functionality:**
- Generates slugs from titles/names
- Used for directory and filename generation

**Review Areas:**
- Slug generation correctness
- Edge cases (special characters, unicode)
- Collision handling

---

### 26. Embedded Documentation

**File:** `docmgr/pkg/doc/doc.go`

**Key Variables:**
- `docFS` - Embedded filesystem containing documentation

**Key Functions:**
- `AddDocToHelpSystem()` - Loads embedded docs into help system

**Functionality:**
- Embeds markdown documentation files
- Integrates with Glazed help system
- Provides built-in help content

**Review Areas:**
- Documentation completeness
- Help system integration
- File embedding correctness

---

### 27. Constants

**File:** `docmgr/pkg/commands/constants.go`

**Key Content:**
- Various constants used across commands

**Functionality:**
- Centralizes shared constants

**Review Areas:**
- Constant definitions
- Usage consistency

---

## Cross-Cutting Concerns

### Error Handling
- Review error wrapping and context propagation
- Check error messages for clarity
- Verify error handling consistency across commands

### File System Operations
- Review file path handling (absolute vs relative)
- Check for race conditions in file operations
- Verify directory creation and cleanup

### YAML Processing
- Review YAML parsing robustness
- Check for edge cases in frontmatter parsing
- Verify YAML writing correctness

### Git Integration
- Review git command execution
- Check error handling for missing git
- Verify git repository detection

### Performance
- Review file system traversal efficiency
- Check for unnecessary file reads
- Verify caching strategies (if any)

### Testing
- Review test coverage (relate_test.go exists)
- Check for missing test files
- Verify test quality

## Files Requiring Review

### Core Infrastructure
- `cmd/docmgr/main.go` - Entry point and command registration
- `pkg/models/document.go` - Data models and YAML handling
- `pkg/commands/config.go` - Configuration management
- `pkg/commands/workspaces.go` - Workspace discovery

### Command Implementations
All files in `pkg/commands/`:
- `add.go`, `changelog.go`, `configure.go`, `create_ticket.go`
- `doctor.go`, `filename_prefix.go`, `guidelines_cmd.go`, `guidelines.go`
- `import_file.go`, `init.go`, `layout_fix.go`, `list_docs.go`
- `list_tickets.go`, `list.go`, `meta_update.go`, `relate.go`
- `renumber.go`, `scaffold.go`, `search.go`, `status.go`
- `tasks.go`, `templates.go`, `vocab_add.go`, `vocab_list.go`
- `vocabulary.go`

### Supporting Code
- `pkg/utils/slug.go` - Utility functions
- `pkg/doc/doc.go` - Embedded documentation
- `pkg/commands/constants.go` - Shared constants

## Related

- See ticket index.md for overall code review goals
- See tasks.md for specific review tasks
