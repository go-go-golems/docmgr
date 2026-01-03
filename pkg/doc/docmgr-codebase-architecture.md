---
Title: docmgr Codebase Architecture
Slug: codebase-architecture
Short: Architectural overview of workspace discovery, ticket management, document parsing, and frontmatter handling in docmgr.
Topics:
- docmgr
- architecture
- implementation
- workspace
- documents
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: GeneralTopic
---

# docmgr Codebase Architecture

## Overview

docmgr's architecture centers on a workspace abstraction that discovers documentation roots, builds in-memory SQLite indexes for fast queries, and provides a unified API for document operations. The system separates concerns cleanly: workspace discovery handles root resolution, document parsing extracts metadata from YAML frontmatter, and the query system enables efficient filtering and searching. This design allows commands to focus on business logic while the workspace handles the complexity of file system traversal, path normalization, and indexing.

**This guide covers:** Workspace discovery and indexing, ticket workspace structure, document model and frontmatter parsing, and how these components integrate to support docmgr's commands.

**Intended audience:** Developers extending docmgr or implementing new features that interact with the workspace, documents, or tickets.

## Workspace Discovery and Resolution

Workspace discovery resolves the documentation root directory through a six-level fallback chain, ensuring commands work consistently whether run from the repository root, a subdirectory, or with explicit configuration. The discovery process also identifies the configuration directory and repository root, creating a complete context for path resolution and vocabulary loading.

**Why this matters:** When you run `docmgr doc list` from anywhere in your repository, it needs to find where your documentation lives. Instead of requiring you to always specify `--root ttmp`, docmgr tries multiple strategies automatically. This makes commands work intuitively whether you're in the repo root, a subdirectory, or even a nested project folder.

**Resolution order:**

The system tries each method in order until it finds a valid docs root:

1. **`--root` flag**: Explicit command-line argument (highest priority)
   - Example: `docmgr doc list --root /path/to/docs`
   - Use when: You need to override the default location

2. **`.ttmp.yaml` in current directory**: Configuration file with `root:` field
   - Example: `.ttmp.yaml` contains `root: custom-docs`
   - Use when: Your project uses a non-standard docs directory name

3. **`.ttmp.yaml` in parent directories**: Walks up the directory tree
   - Example: Running from `project/src/` finds `.ttmp.yaml` in `project/`
   - Use when: Working in subdirectories of a configured project

4. **`DOCMGR_ROOT` environment variable**: System-wide or session setting
   - Example: `export DOCMGR_ROOT=/shared/docs`
   - Use when: You have a shared documentation location across projects

5. **Git repository root**: Automatically finds `<git-root>/ttmp`
   - Example: Running from anywhere in a git repo finds `ttmp/` at repo root
   - Use when: Working in a standard git repository structure

6. **Default fallback**: `ttmp` in current directory
   - Example: Creates `ttmp/` if it doesn't exist
   - Use when: No other configuration is found

**Visual flow:**

```
Command Execution
    │
    ├─→ Try --root flag
    │   └─→ Found? Use it ✓
    │
    ├─→ Try .ttmp.yaml (current dir)
    │   └─→ Found? Use root: field ✓
    │
    ├─→ Walk up tree for .ttmp.yaml
    │   └─→ Found? Use root: field ✓
    │
    ├─→ Check DOCMGR_ROOT env var
    │   └─→ Set? Use it ✓
    │
    ├─→ Find git root, check for ttmp/
    │   └─→ Found? Use it ✓
    │
    └─→ Default to ./ttmp
        └─→ Use current directory ✓
```

**Implementation:**

The `DiscoverWorkspace()` function encapsulates this logic:

```go
// internal/workspace/workspace.go
func DiscoverWorkspace(ctx context.Context, opts DiscoverOptions) (*Workspace, error) {
    root := opts.RootOverride
    if root == "" {
        root = "ttmp"
    }
    root = ResolveRoot(root)  // Applies fallback chain
    
    // Best-effort config load
    cfg, _ := LoadWorkspaceConfig()
    
    // Resolve config directory and repository root
    configDir := resolveConfigDir(cfg)
    repoRoot, _ := FindRepositoryRoot()
    
    return NewWorkspaceFromContext(WorkspaceContext{
        Root:      root,
        ConfigDir: configDir,
        RepoRoot:  repoRoot,
        Config:    cfg,
    })
}
```

**WorkspaceContext** captures the resolved environment:

The `WorkspaceContext` struct holds all the information needed to work with documents:

- **`Root`**: Absolute path to docs root (typically `ttmp/`)
  - Example: `/home/user/project/ttmp`
  - Used for: Finding all documentation files

- **`ConfigDir`**: Directory containing `.ttmp.yaml` (usually repo root)
  - Example: `/home/user/project`
  - Used for: Loading vocabulary and configuration

- **`RepoRoot`**: Git repository root (for path normalization)
  - Example: `/home/user/project`
  - Used for: Converting relative paths to repo-relative paths

- **`Config`**: Parsed workspace configuration (may be nil)
  - Contains: Custom root paths, vocabulary location, etc.
  - Used for: Overriding defaults and customizing behavior

**Usage in commands:**

Every command that needs to access documents follows this pattern:

```go
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
    RootOverride: settings.Root,
})
if err != nil {
    return fmt.Errorf("failed to discover workspace: %w", err)
}
// ws.Context().Root now contains the resolved absolute path
```

**Key points for developers:**

- Always use `DiscoverWorkspace()` instead of manually resolving paths
- The resolved `Root` is always an absolute path, making it safe to use
- If discovery fails, commands should return clear error messages
- The workspace context is immutable once created (thread-safe)

## SQLite Indexing System

docmgr builds an in-memory SQLite index on each CLI invocation to enable fast queries across all documents. The index stores document metadata, topics, and related files with normalized paths, allowing efficient filtering by ticket, status, doc-type, topics, and file paths. This design trades a small startup cost for query performance and eliminates the need for persistent index files.

**Why this matters:** Searching through hundreds of markdown files on every command would be slow. Instead, docmgr builds a fast SQLite database in memory that lets you query documents instantly. Think of it like a library catalog: instead of walking through every shelf, you look up books in a card catalog.

**Index lifecycle:**

The index is created fresh on every command invocation. Here's what happens:

1. **Command starts**: User runs `docmgr doc list`
2. **Index initialization**: `ws.InitIndex()` is called
3. **Index building**: All documents are scanned and indexed
4. **Query execution**: Commands query the index for fast results
5. **Command ends**: Index is discarded (in-memory, no cleanup needed)

**Why rebuild every time?**

- **No staleness**: Index always matches current files (no cache invalidation)
- **Simplicity**: No need to manage persistent index files or updates
- **Trade-off**: Small startup cost (~100-500ms for large repos) for correctness

```go
// Initialize index (rebuilds from scratch)
if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{
    IncludeBody: false,  // Set true for full-text search
}); err != nil {
    return fmt.Errorf("failed to initialize workspace index: %w", err)
}
```

**Index schema** (`internal/workspace/sqlite_schema.go`):

The SQLite database has three main tables that work together:

**1. `docs` table** - Core document metadata:
- Stores: path, ticket, doc_type, status, intent, title, last_updated
- Purpose: Fast lookup of document properties
- Indexes: On ticket, doc_type, status for fast filtering

**2. `doc_topics` table** - Many-to-many relationship:
- Stores: doc_id → topic mapping
- Purpose: Documents can have multiple topics, topics can appear in many documents
- Example: A document about "API Design" might have topics: `[api, architecture, backend]`

**3. `related_files` table** - File references with normalized paths:
- Stores: doc_id, file paths in multiple formats, optional notes
- Purpose: Enable reverse lookups (find docs referencing a file)
- Normalization: Same file can be referenced as absolute, relative, or repo-relative path

**Visual schema:**

```
┌─────────────────┐
│      docs       │
├─────────────────┤
│ doc_id (PK)     │
│ path            │
│ ticket          │
│ doc_type        │
│ status          │
│ title           │
│ last_updated    │
└────────┬────────┘
         │
         │ 1:N
         ├─────────────────┐
         │                 │
         ▼                 ▼
┌──────────────┐  ┌──────────────────┐
│ doc_topics   │  │ related_files    │
├──────────────┤  ├──────────────────┤
│ doc_id (FK)  │  │ doc_id (FK)      │
│ topic        │  │ norm_repo_rel     │
└──────────────┘  │ norm_docs_rel    │
                  │ norm_abs         │
                  │ note             │
                  └──────────────────┘
```

**Indexing process:**

When `InitIndex()` is called, here's the step-by-step process:

1. **Walk documents**: `documents.WalkDocuments()` traverses the docs root
   - Recursively visits every `.md` file
   - Skips directories starting with `_` (like `_templates/`)
   - Can be customized with skip functions

2. **Parse frontmatter**: Each `.md` file is parsed via `ReadDocumentWithFrontmatter()`
   - Extracts YAML frontmatter block
   - Parses into `Document` struct
   - Handles parse errors gracefully (still indexes with error metadata)

3. **Extract metadata**: Ticket, doc-type, topics, related files are extracted
   - Required fields: Title, Ticket, DocType
   - Optional fields: Status, Topics, RelatedFiles, etc.
   - Missing fields don't block indexing

4. **Normalize paths**: Related files are normalized to multiple representations
   - Converts absolute paths to repo-relative
   - Handles both absolute and relative path formats
   - Stores multiple representations for flexible matching

5. **Insert into SQLite**: Documents, topics, and related files are inserted
   - Uses transactions for atomicity
   - Foreign keys ensure referential integrity
   - Indexes are created for fast queries

**Performance considerations:**

- **Memory usage**: In-memory SQLite is efficient (~1-10MB for typical repos)
- **Build time**: Scales linearly with number of documents (~1-5ms per document)
- **Query time**: Sub-millisecond for most queries (thanks to indexes)

**Path normalization** (`internal/paths/normalization.go`):

One of the trickiest parts of indexing is handling file paths. The same file might be referenced in different ways:

- Absolute: `/home/user/project/backend/api/user.go`
- Repo-relative: `backend/api/user.go`
- Docs-relative: `../../backend/api/user.go`

**Solution**: Store multiple normalized representations for each file:

- **`norm_repo_rel`**: Repository-relative path (preferred canonical key)
  - Example: `backend/api/user.go`
  - Used for: Most queries and matching

- **`norm_docs_rel`**: Docs-root relative path
  - Example: `../../backend/api/user.go` (if doc is in `ttmp/2025/12/19/`)
  - Used for: Fallback matching

- **`norm_abs`**: Absolute path
  - Example: `/home/user/project/backend/api/user.go`
  - Used for: Final fallback and display

- **`norm_canonical`**: Best-effort canonical key (prefers repo_rel)
  - Example: `backend/api/user.go`
  - Used for: Primary matching key

**Why multiple representations?**

This allows queries to match files regardless of how they were referenced in frontmatter. When a user searches for `backend/api/user.go`, the query can match:
- Documents that reference it as absolute path
- Documents that reference it as relative path
- Documents that reference it as repo-relative path

**Example:**

```yaml
# Document A (in ttmp/2025/12/19/TICKET-001/)
RelatedFiles:
  - Path: /home/user/project/backend/api/user.go  # Absolute

# Document B (in ttmp/2025/12/20/TICKET-002/)
RelatedFiles:
  - Path: ../../backend/api/user.go  # Relative

# Document C (in ttmp/2025/12/21/TICKET-003/)
RelatedFiles:
  - Path: backend/api/user.go  # Repo-relative
```

All three documents will match a query for `backend/api/user.go` because the normalization stores all representations.

## Ticket Workspace Structure

Ticket workspaces are date-organized directories containing an index document, standard subdirectories for document types, and metadata files. The structure provides a consistent layout while remaining flexible enough to accommodate different documentation needs.

**Why this matters:** Every ticket gets its own workspace directory with a predictable structure. This makes it easy to find related documents, keeps things organized chronologically, and provides a consistent experience across all tickets. Think of it like a filing cabinet: each ticket is a folder with labeled sections.

**Directory structure:**

Tickets are organized by date (YYYY/MM/DD) and then by ticket ID:

```
ttmp/
  YYYY/                    # Year (e.g., 2025)
    MM/                    # Month (e.g., 12)
      DD/                  # Day (e.g., 19)
        <TICKET>--<slug>/  # Ticket directory
          index.md              # Ticket overview (required)
          tasks.md              # Task list
          changelog.md          # Change history
          design/               # Design documents
          reference/            # Reference docs
          playbooks/            # Operational procedures
          scripts/              # Utility scripts
          sources/              # Source code references
          various/              # Miscellaneous docs
          archive/              # Archived documents
          .meta/                # Metadata files
```

**Why date-based organization?**

- **Chronological browsing**: Easy to see what was worked on when
- **Natural grouping**: Related tickets from the same time period are nearby
- **Scalability**: Prevents any single directory from getting too large
- **Git-friendly**: Works well with git's directory structure

**Directory purposes:**

- **`index.md`**: Required ticket overview document
  - Contains: Ticket summary, goals, context
  - Used by: `docmgr ticket list` to find tickets

- **`tasks.md`**: Task tracking
  - Contains: Checklist of work items
  - Used by: `docmgr task` commands

- **`changelog.md`**: Change history
  - Contains: Timeline of what changed and why
  - Used by: Understanding ticket evolution

- **`design/`**: Design documents
  - Contains: Architecture decisions, design specs
  - Used by: Long-term reference

- **`reference/`**: Reference documentation
  - Contains: API contracts, data schemas, how-tos
  - Used by: Quick lookup during development

- **`playbooks/`**: Operational procedures
  - Contains: Test procedures, runbooks, checklists
  - Used by: QA and operations teams

- **`scripts/`**: Utility scripts
  - Contains: Helper scripts, automation
  - Used by: Development and testing

- **`sources/`**: Source code references
  - Contains: Code snippets, examples
  - Used by: Code documentation

- **`various/`**: Miscellaneous documents
  - Contains: Anything that doesn't fit other categories
  - Used by: Flexible storage

- **`archive/`**: Archived documents
  - Contains: Old or superseded documents
  - Used by: Historical reference

- **`.meta/`**: Metadata files
  - Contains: Internal metadata (not user-facing)
  - Used by: docmgr internals

**Ticket discovery:**

Tickets are discovered by querying for documents with `DocType == "index"`. This is efficient because:

- Only one query needed to find all tickets
- Index document (`index.md`) is required, so every ticket has one
- Can filter by ticket ID, status, topics, etc.

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: ticketID},
    Filters: workspace.DocFilters{DocType: "index"},
    Options: workspace.DocQueryOptions{
        IncludeErrors: false,
        IncludeArchivedPath: true,
        IncludeScriptsPath: true,
        IncludeControlDocs: true,
    },
})
```

**How ticket discovery works:**

1. **Query index documents**: Find all docs with `DocType == "index"`
2. **Extract ticket ID**: From `index.md` frontmatter (`Ticket: MEN-3475`)
3. **Filter by scope**: If `ScopeTicket` is used, filter to specific ticket
4. **Return results**: List of ticket index documents with metadata

**Path tags** (`internal/workspace/index_builder.go`):

During indexing, documents are automatically tagged based on their location. These tags enable filtering and special handling:

- **`IsIndex`**: Document is a ticket index (`index.md`)
  - Used for: Finding tickets, ticket-specific queries
  - Example: `ttmp/2025/12/19/MEN-3475--feature/index.md`

- **`IsArchivedPath`**: Document is in `archive/` subdirectory
  - Used for: Filtering out archived docs from normal queries
  - Example: `ttmp/2025/12/19/MEN-3475--feature/archive/old-design.md`

- **`IsScriptsPath`**: Document is in `scripts/` subdirectory
  - Used for: Separating scripts from documentation
  - Example: `ttmp/2025/12/19/MEN-3475--feature/scripts/setup.sh`

- **`IsSourcesPath`**: Document is in `sources/` subdirectory
  - Used for: Identifying source code references
  - Example: `ttmp/2025/12/19/MEN-3475--feature/sources/example.go`

- **`IsControlDoc`**: Document is a control file (tasks.md, changelog.md, etc.)
  - Used for: Special handling of metadata files
  - Example: `ttmp/2025/12/19/MEN-3475--feature/tasks.md`

**Why tags matter:**

Tags allow queries to include or exclude specific document types:

```go
Options: workspace.DocQueryOptions{
    IncludeArchivedPath: false,  // Don't show archived docs
    IncludeScriptsPath:   false,  // Don't show scripts
    IncludeControlDocs:    true,   // Include tasks.md, changelog.md
}
```

This gives fine-grained control over what documents appear in results.

## Document Model and Frontmatter

Documents are markdown files with YAML frontmatter containing structured metadata. The frontmatter parsing system handles both legacy and current formats, provides error diagnostics, and supports preprocessing to reduce parse failures.

**Why this matters:** Every document needs metadata (ticket, topics, status) to be searchable and organized. The frontmatter system extracts this metadata automatically, validates it, and makes it queryable. Think of frontmatter like a library card catalog entry: it tells you what the document is about without reading the whole thing.

**Document model** (`pkg/models/document.go`):

The `Document` struct represents all the metadata that can be stored in a document's frontmatter:

```go
type Document struct {
    Title           string       // Document title (required)
    Ticket          string       // Ticket ID (required)
    Status          string       // Workflow status (draft, active, review, etc.)
    Topics          []string     // List of topics (api, backend, etc.)
    DocType         string       // Document type (required: design-doc, reference, etc.)
    Intent          string       // Longevity intent (long-term, short-term, throwaway)
    Owners          []string     // List of owner usernames
    RelatedFiles    RelatedFiles // Code files related to this document
    ExternalSources []string     // External references (URLs, etc.)
    Summary         string       // Brief summary of the document
    LastUpdated     time.Time    // When the document was last modified
}
```

**Field purposes:**

- **Required fields** (validation fails if missing):
  - `Title`: Human-readable document name
  - `Ticket`: Which ticket this document belongs to
  - `DocType`: What kind of document this is

- **Optional fields** (can be empty):
  - `Status`: Current workflow state
  - `Topics`: Categories for filtering and discovery
  - `Owners`: Who is responsible for this document
  - `RelatedFiles`: Links to code files
  - `Summary`: Brief description
  - `LastUpdated`: Modification timestamp

**Frontmatter format:**

Frontmatter appears at the top of every markdown file, between `---` delimiters:

```yaml
---
Title: API Design for User Service
Ticket: MEN-3475
DocType: design-doc
Topics: [api, architecture]
Owners: [alice, bob]
Status: active
Intent: long-term
RelatedFiles:
  - Path: backend/api/user.go
    Note: Main API implementation
Summary: Design document for user service API
LastUpdated: 2025-12-19T10:00:00Z
---

# Document Content

Markdown body content here...
```

**Visual structure:**

```
┌─────────────────────────────────────┐
│  ---                                │
│  Title: API Design                 │  ← YAML Frontmatter
│  Ticket: MEN-3475                   │     (Metadata)
│  Topics: [api, architecture]        │
│  ...                                 │
│  ---                                │
├─────────────────────────────────────┤
│                                     │
│  # Document Content                 │  ← Markdown Body
│                                     │     (Content)
│  This is the actual content...      │
│                                     │
└─────────────────────────────────────┘
```

**Frontmatter parsing** (`internal/documents/frontmatter.go`):

The parsing pipeline handles three stages, each with error handling:

**Stage 1: Extraction** - Manual scanning for `---` delimiters

The parser scans the file line-by-line to find the frontmatter block:

- Looks for first line that equals `---` (trimmed)
- Looks for second `---` delimiter
- Extracts everything between them as frontmatter
- Everything after becomes the body

**Why manual scanning?** Faster than regex, gives precise line numbers for error reporting.

**Stage 2: Preprocessing** - Quote risky scalars

Before parsing YAML, the system quotes values that might cause parse errors:

- Values starting with special chars: `@`, `` ` ``, `#`, `&`, `*`, `!`, `|`, `>`
- Values containing colons: `: ` or trailing `:`
- Values with inline comments: ` #`
- Values with tabs or template markers

**Example:**

```yaml
# Before preprocessing
Title: API: Design & Implementation

# After preprocessing
Title: 'API: Design & Implementation'
```

**Why preprocessing?** Reduces parse failures by 80-90% for common edge cases.

**Stage 3: Decoding** - YAML decoder parses into Document struct

The preprocessed YAML is decoded into a `Document` struct:

- Uses `yaml.NewDecoder()` for parsing
- Handles type conversion automatically
- Validates required fields
- Returns structured error if parsing fails

```go
func ReadDocumentWithFrontmatter(path string) (*models.Document, string, error) {
    raw, _ := os.ReadFile(path)
    
    // Extract frontmatter block
    fm, body, fmStartLine, _ := extractFrontmatter(raw)
    
    // Preprocess to quote risky values
    fm = frontmatter.PreprocessYAML(fm)
    
    // Decode YAML
    var node yaml.Node
    dec := yaml.NewDecoder(bytes.NewReader(fm))
    if err := dec.Decode(&node); err != nil {
        // Extract line/col, build snippet, wrap in taxonomy
        return nil, "", wrapParseError(err, path, fmStartLine)
    }
    
    // Decode into Document struct
    var doc models.Document
    node.Decode(&doc)
    
    return &doc, string(body), nil
}
```

**Error handling:**

When parsing fails, errors are wrapped in a diagnostics taxonomy that includes:

- **File path**: Which file had the error
- **Line/column numbers**: Exact location of the problem
- **Code snippet**: 3 lines of context around the error
- **Problem description**: User-friendly explanation
- **Suggested fixes**: Auto-generated fixes when possible

**Example error output:**

```
YAML/frontmatter syntax error
File: ttmp/2025/12/19/MEN-3475/index.md
Line: 5, Column: 12

  3 | Topics: [api, backend]
  4 | DocType: design-doc
> 5 | Status: active: review
    |            ^
  6 | Summary: Design document

Problem: Unexpected colon in scalar value
Suggestion: Quote the value: Status: 'active: review'
```

**RelatedFiles format:**

The `RelatedFiles` field supports both legacy and current formats for backward compatibility:

**Legacy format** (still supported):

```yaml
RelatedFiles:
  - backend/api/user.go
  - frontend/components/User.tsx
```

**Current format** (preferred):

```yaml
RelatedFiles:
  - Path: backend/api/user.go
    Note: Main API implementation
  - Path: frontend/components/User.tsx
    Note: Frontend component consuming the API
```

**Why both formats?**

- **Backward compatibility**: Old documents still work
- **Gradual migration**: Teams can migrate over time
- **Flexibility**: Simple cases don't need notes

The `RelatedFiles` type (`pkg/models/document.go`) handles both formats automatically via custom `UnmarshalYAML()` that detects the format and converts appropriately.

## Document Walking

Document walking provides a callback-based API for traversing the documentation tree. It's used by indexing, validation, and discovery operations that need to process all markdown files.

**Why this matters:** Many operations need to visit every document in the workspace. Instead of each command implementing its own file traversal logic, `WalkDocuments()` provides a consistent, tested way to iterate through all markdown files. Think of it like a `for` loop over all documents.

**WalkDocuments function** (`internal/documents/walk.go`):

The function takes a root directory and a callback function that gets called for each `.md` file:

```go
func WalkDocuments(root string, fn WalkDocumentFunc, opts ...WalkOption) error {
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        // Skip directories starting with "_"
        if d.IsDir() && strings.HasPrefix(d.Name(), "_") {
            return fs.SkipDir
        }
        
        // Process .md files
        if strings.ToLower(filepath.Ext(d.Name())) == ".md" {
            doc, body, readErr := ReadDocumentWithFrontmatter(path)
            return fn(path, doc, body, readErr)
        }
        
        return nil
    })
}
```

**How it works:**

1. **Recursive traversal**: Uses `filepath.WalkDir()` to visit every file
2. **Skip hidden directories**: Automatically skips directories starting with `_` (like `_templates/`)
3. **Filter markdown files**: Only processes files with `.md` extension
4. **Parse each file**: Calls `ReadDocumentWithFrontmatter()` automatically
5. **Invoke callback**: Calls your function with parsed document and body

**Callback signature:**

```go
type WalkDocumentFunc func(
    path string,              // Full path to the file
    doc *models.Document,     // Parsed document (nil if parse failed)
    body string,              // Markdown body content
    readErr error,            // Parse error (nil if successful)
) error
```

**Usage:**

Here's how you'd use it to process all documents:

```go
err := documents.WalkDocuments(root, func(path string, doc *models.Document, body string, readErr error) error {
    if readErr != nil {
        // Handle parse errors (skip, log, etc.)
        fmt.Printf("Failed to parse %s: %v\n", path, readErr)
        return nil  // Continue walking
    }
    
    // Process valid document
    fmt.Printf("Found document: %s (ticket: %s)\n", doc.Title, doc.Ticket)
    processDocument(doc, body)
    return nil
}, documents.WithSkipDir(func(path string, d fs.DirEntry) bool {
    // Custom skip logic
    return strings.Contains(path, "node_modules")
}))
```

**Custom skip logic:**

You can customize which directories to skip:

```go
documents.WithSkipDir(func(path string, d fs.DirEntry) bool {
    // Skip node_modules, .git, vendor, etc.
    skipDirs := []string{"node_modules", ".git", "vendor", "build"}
    for _, skip := range skipDirs {
        if strings.Contains(path, skip) {
            return true
        }
    }
    return false
})
```

**Common use cases:**

- **Indexing**: Walk all documents and insert into SQLite
- **Validation**: Check all documents for errors or missing fields
- **Statistics**: Count documents by type, topic, or ticket
- **Migration**: Update document format across the workspace

**Performance:**

- **Efficient**: Uses Go's optimized `filepath.WalkDir()`
- **Memory-friendly**: Processes one document at a time (streaming)
- **Fast**: Typically processes 100-1000 documents per second

## Query System

The query system provides a flexible API for filtering documents by ticket, status, doc-type, topics, and related files. Queries use the SQLite index for performance and support scoping to repository-wide or ticket-specific searches.

**Why this matters:** Once documents are indexed, you need a way to find them. The query system lets you filter by any combination of metadata fields, making it easy to find exactly what you're looking for. Think of it like a database query: you specify what you want, and it returns matching documents instantly.

**Query structure** (`internal/workspace/query_docs.go`):

A query consists of three parts: scope, filters, and options:

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{
        Kind: workspace.ScopeRepo,  // or ScopeTicket with TicketID
    },
    Filters: workspace.DocFilters{
        Ticket:    "MEN-3475",
        Status:    "active",
        DocType:   "design-doc",
        TopicsAny: []string{"api", "backend"},
        RelatedFile: []string{"backend/api/user.go"},
    },
    Options: workspace.DocQueryOptions{
        IncludeBody:         false,
        IncludeErrors:       false,
        IncludeArchivedPath: true,
        IncludeScriptsPath:  true,
        IncludeControlDocs:  true,
        OrderBy:             workspace.OrderByLastUpdated,
        Reverse:             true,
    },
})
```

**Query components:**

**1. Scope** - Where to search:

- **`ScopeRepo`**: Search entire repository (all tickets)
  - Use when: Finding documents across all tickets
  - Example: "List all design docs"

- **`ScopeTicket`**: Search within a specific ticket
  - Use when: Finding documents for one ticket
  - Example: "List all docs for MEN-3475"

**2. Filters** - What to match:

- **`Ticket`**: Exact ticket ID match
- **`Status`**: Exact status match (draft, active, review, etc.)
- **`DocType`**: Exact doc type match (design-doc, reference, etc.)
- **`TopicsAny`**: Match if document has ANY of these topics (OR logic)
- **`RelatedFile`**: Match if document references this file

**3. Options** - How to return results:

- **`IncludeBody`**: Include full markdown body (increases memory)
- **`IncludeErrors`**: Include documents that failed to parse
- **`IncludeArchivedPath`**: Include documents in `archive/` directories
- **`IncludeScriptsPath`**: Include documents in `scripts/` directories
- **`IncludeControlDocs`**: Include control files (tasks.md, changelog.md)
- **`OrderBy`**: Sort by path, last_updated, etc.
- **`Reverse`**: Reverse sort order (newest first, etc.)

**Query result:**

The query returns a `DocQueryResult` containing a list of `DocHandle` objects:

```go
type DocQueryResult struct {
    Docs []DocHandle  // Documents matching query
}

type DocHandle struct {
    Path    string            // File path
    Doc     *models.Document  // Parsed document (nil if parse failed)
    Body    string            // Markdown body (only if IncludeBody=true)
    Error   error             // Parse error (if any)
}
```

**Important:** Always check `h.Doc == nil` before accessing document fields, as parse errors can occur.

**Common query patterns:**

**1. List all tickets:**

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Filters: workspace.DocFilters{DocType: "index"},
    Options: workspace.DocQueryOptions{
        OrderBy: workspace.OrderByLastUpdated,
        Reverse: true,  // Newest first
    },
})
```

**2. Find ticket documents:**

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{
        Kind:     workspace.ScopeTicket,
        TicketID: "MEN-3475",
    },
    Filters: workspace.DocFilters{},
})
```

**3. Search by file:**

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Filters: workspace.DocFilters{
        RelatedFile: []string{"backend/api/user.go"},
    },
})
```

**4. Filter by topics (OR logic):**

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
    Scope: workspace.Scope{Kind: workspace.ScopeRepo},
    Filters: workspace.DocFilters{
        TopicsAny: []string{"api", "backend"},  // Matches if doc has api OR backend
    },
})
```

**Query performance:**

- **Fast**: Uses SQLite indexes for sub-millisecond queries
- **Scalable**: Handles thousands of documents efficiently
- **Memory-efficient**: Only loads requested data (body optional)

## Integration Points

Commands integrate with the workspace system through a consistent pattern:

1. **Discover workspace**: `workspace.DiscoverWorkspace()` resolves root and context
2. **Initialize index**: `ws.InitIndex()` builds SQLite index (if querying needed)
3. **Query or walk**: Use `QueryDocs()` for filtered access or `WalkDocuments()` for discovery
4. **Process results**: Handle documents, errors, and metadata

**Why this pattern matters:** Every command that works with documents follows the same steps. This consistency makes commands predictable, testable, and easy to understand. Once you learn the pattern, you can read any command's code and understand what it does.

**Example command pattern:**

Here's a complete example showing how a list command integrates with the workspace:

```go
func (c *ListDocsCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Step 1: Parse command settings
    settings := &ListDocsSettings{}
    parsedLayers.InitializeStruct(layers.DefaultSlug, settings)
    
    // Step 2: Discover workspace
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
        RootOverride: settings.Root,
    })
    if err != nil {
        return err
    }
    
    // Step 3: Initialize index (needed for queries)
    if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{
        IncludeBody: false,  // Don't need body for listing
    }); err != nil {
        return err
    }
    
    // Step 4: Query documents
    res, err := ws.QueryDocs(ctx, workspace.DocQuery{
        Scope: workspace.Scope{Kind: workspace.ScopeRepo},
        Filters: workspace.DocFilters{
            Ticket:  settings.Ticket,
            DocType: settings.DocType,
            TopicsAny: settings.Topics,
        },
    })
    if err != nil {
        return err
    }
    
    // Step 5: Output results
    for _, h := range res.Docs {
        if h.Doc == nil {
            continue  // Skip parse errors
        }
        row := types.NewRow(
            types.MRP("ticket", h.Doc.Ticket),
            types.MRP("title", h.Doc.Title),
            // ... more fields
        )
        gp.AddRow(ctx, row)
    }
    
    return nil
}
```

**Visual flow:**

```
Command Execution
    │
    ├─→ Parse Settings
    │   └─→ Extract flags/args
    │
    ├─→ Discover Workspace
    │   └─→ Resolve docs root
    │
    ├─→ Initialize Index
    │   └─→ Build SQLite index
    │
    ├─→ Query Documents
    │   └─→ Filter by criteria
    │
    └─→ Process Results
        ├─→ Handle errors
        ├─→ Format output
        └─→ Return
```

**When to use QueryDocs vs WalkDocuments:**

- **Use `QueryDocs()`** when:
  - You need filtering (by ticket, status, topics, etc.)
  - You need fast lookups
  - You're working with indexed data

- **Use `WalkDocuments()`** when:
  - You need to process every document
  - You're building an index
  - You're doing validation or migration
  - You don't need filtering

**Error handling best practices:**

- **Always check `h.Doc == nil`**: Parse errors can occur
- **Continue on non-fatal errors**: Don't stop processing if one document fails
- **Return clear errors**: Use `fmt.Errorf()` with context
- **Log parse errors**: Help users understand what went wrong

## Key Design Decisions

This section explains the "why" behind major architectural choices. Understanding these decisions helps you work with the codebase effectively and make informed choices when extending it.

**In-memory SQLite index:**

**Decision:** Rebuild the index from scratch on every CLI invocation instead of using a persistent index file.

**Rationale:**

- **No staleness**: Index always matches current files (no cache invalidation needed)
- **Simplicity**: No need to manage index files, updates, or cleanup
- **Correctness**: Can't have bugs from stale indexes

**Trade-offs:**

- **Startup cost**: Small delay (~100-500ms for large repos) to build index
- **Memory usage**: Index lives in memory during command execution
- **No persistence**: Can't reuse index across commands

**When this matters:** For most commands, the startup cost is negligible compared to the query performance benefits. For very large repositories (10,000+ documents), consider caching strategies.

**Path normalization:**

**Decision:** Store multiple normalized path representations for each related file instead of a single canonical path.

**Rationale:**

- **Flexible matching**: Queries work regardless of how paths were written
- **User-friendly**: Users don't need to worry about path format
- **Reliable**: Handles edge cases (absolute vs relative paths)

**Trade-offs:**

- **Storage overhead**: Each file takes ~4x storage (multiple representations)
- **Complexity**: More code to maintain normalization logic
- **Query complexity**: Need to check multiple fields

**When this matters:** The storage overhead is minimal (~few KB per document), but the flexibility is crucial for user experience. Without this, users would constantly struggle with path format mismatches.

**Best-effort parsing:**

**Decision:** Continue indexing documents even when frontmatter parsing fails, storing error metadata instead of skipping them.

**Rationale:**

- **Diagnostics**: Can report all parse errors, not just the first one
- **Repair workflows**: Can fix documents programmatically
- **Partial functionality**: Even broken docs can be found by path

**Trade-offs:**

- **Complexity**: Need to handle `doc == nil` cases everywhere
- **Error handling**: More error paths to test
- **User confusion**: Broken docs appear in some queries

**When this matters:** This enables the `docmgr doctor` command to report all issues at once, making it much easier to fix documentation problems in bulk.

**Vocabulary-guided validation:**

**Decision:** Validate topics, doc-types, intent, and status against a vocabulary file, but only warn (not error) for unknown values.

**Rationale:**

- **Consistency**: Encourages teams to use standard terms
- **Flexibility**: Doesn't break when new terms are added
- **Evolution**: Vocabulary can grow over time

**Trade-offs:**

- **No enforcement**: Teams can still use invalid terms
- **Warnings only**: Might miss typos or mistakes
- **Vocabulary management**: Need to maintain vocabulary.yaml

**When this matters:** This balance between guidance and flexibility allows teams to adopt docmgr gradually while still encouraging best practices. Strict validation would be too rigid for real-world use.
