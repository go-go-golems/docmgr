---
Title: 'Doc Search: Implementation and API Guide'
Ticket: 001-ADD-DOCMGR-UI
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - tooling
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/doc/search.go
      Note: Cobra wiring for search + glaze dual-mode + completions
    - Path: internal/paths/resolver.go
      Note: Path normalization + fuzzy match utilities powering reverse lookup
    - Path: internal/workspace/index_builder.go
      Note: Builds the in-memory SQLite index used by search
    - Path: internal/workspace/query_docs.go
      Note: Workspace.QueryDocs API and hydration
    - Path: internal/workspace/query_docs_sql.go
      Note: SQL compilation for filters (topics
    - Path: internal/workspace/sqlite_schema.go
      Note: Index schema (docs
    - Path: pkg/commands/search.go
      Note: Primary implementation of doc search (flags
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T13:39:08.17084042-05:00
---


# Doc Search: Implementation and API Guide

## Goal

This document is an exhaustive guide to `docmgr`’s search functionality: how it works internally, what the CLI surface area is, what the output contracts look like, and how to extend search safely. It is written to be useful in two directions:

- Implementation → usage: you can read the code paths and then understand the CLI behavior you’re seeing.
- Usage → implementation: you can start from a CLI command and understand what code paths and data structures it exercises.

## Context

`docmgr` manages structured documentation stored as Markdown files with YAML frontmatter in a “ticket workspace” layout (date + ticket ID + slug). Search is a core workflow feature: it lets you find docs by full-text content, metadata (ticket/topics/doc-type/status), and reverse lookups from code files back to docs via `RelatedFiles`.

Search in `docmgr` is intentionally pragmatic:

- Metadata + reverse lookups are implemented as an in-memory SQLite index rebuilt per invocation.
- Content search is currently a post-filter substring scan over document bodies (not FTS).
- Path matching for `--file` and `--dir` is intentionally “best effort” to handle real-world path drift (repo-relative vs doc-relative vs absolute vs historical oddities).

This guide assumes you’re working in a repository that contains a `.ttmp.yaml` at/above your current working directory. In this repo, `.ttmp.yaml` points docmgr at `docmgr/ttmp` as the docs root.

### Running docmgr in this repo (important)

In this workspace, `go run` from the repo root fails due to `go.work` constraints. The simplest reliable way to run `docmgr` commands is to run from the `docmgr/` module and disable the workspace file:

```bash
cd docmgr
GOWORK=off go run ./cmd/docmgr status --summary-only
GOWORK=off go run ./cmd/docmgr doc search --query "Workspace.QueryDocs"
```

This does not change `docmgr` behavior; it only changes how Go resolves modules for building the CLI binary.

## Quick Reference

### Command names

- Canonical: `docmgr doc search`
- Alias: `docmgr search` (same command, different invocation path)

### What “search” can do

`docmgr doc search` supports these major modes:

1. **Content search (substring)**: `--query "term"` (optionally with metadata filters).
2. **Metadata filtering (index-backed)**: `--ticket`, `--topics`, `--doc-type`, `--status`.
3. **Reverse lookup (index-backed)**:
   - file → docs: `--file path/to/file.go`
   - dir → docs: `--dir path/to/dir/`
4. **External sources (post-filter)**: `--external-source "https://..."`.
5. **Date filtering (post-filter)**: `--since`, `--until`, `--created-since`, `--updated-since`.
6. **File suggestions (heuristics blend)**: `--files` (git history + git status + ripgrep + existing `RelatedFiles`).

### Output modes (two “personalities”)

The search command runs in *dual mode*:

- **Human (bare) mode**: default; prints a line per result.
- **Structured (glaze) mode**: `--with-glaze-output`; emits rows that can be rendered as `json`, `table`, `csv`, etc.

This means you can use the same verb interactively and in scripts.

### Structured output schema (glaze mode)

When you run with `--with-glaze-output`, `doc search` emits rows with these fields (column names):

| Field | Type (typical) | Meaning |
|---|---:|---|
| `ticket` | string | Ticket ID from frontmatter. |
| `title` | string | Document title from frontmatter. |
| `doc_type` | string | Document type (`DocType`). |
| `status` | string | Document status (`Status`). |
| `topics` | string | Topics joined as a single comma-separated string. |
| `path` | string | Path relative to docs root. |
| `snippet` | string | Extracted snippet (content context). |
| `file` | string (optional) | Only present for `--file` searches; comma-joined matching related file(s). |
| `file_note` | string (optional) | Only present for `--file` searches; notes for the matching related file(s), joined with ` | `. |

In `--files` suggestion mode, the emitted row schema changes to:

| Field | Meaning |
|---|---|
| `file` | Suggested file path. |
| `source` | Heuristic source (e.g., `related_files`, `git_history`, `git_modified`, `ripgrep`). |
| `reason` | Human-readable explanation. |

### Flags and semantics (CLI contract)

| Flag | Meaning | Implemented where | Notes |
|---|---|---|---|
| `--query <text>` | Full-text content search | `pkg/commands/search.go` | Case-insensitive substring match on the markdown body. |
| `--ticket <ID>` | Restrict to a ticket | `internal/workspace/query_docs_sql.go` | Implemented as a SQL `WHERE d.ticket_id = ?` scope/filter. |
| `--topics a,b,c` | Match any topic | `internal/workspace/query_docs_sql.go` | OR semantics (“TopicsAny”). |
| `--doc-type <type>` | Filter by DocType | `internal/workspace/query_docs_sql.go` | Exact match on frontmatter `DocType`. |
| `--status <status>` | Filter by Status | `internal/workspace/query_docs_sql.go` | Exact match on frontmatter `Status`. |
| `--file <path>` | Reverse lookup: docs referencing a file | `internal/workspace/query_docs_sql.go` | Best-effort matching across multiple normalized keys + basename-only suffix matching. |
| `--dir <dir>` | Reverse lookup: docs referencing a dir | `internal/workspace/query_docs_sql.go` | Prefix matching across normalized keys (`LIKE "<prefix>/%"`). |
| `--external-source <substr>` | Filter by external source URL | `pkg/commands/search.go` | Reads frontmatter again per candidate doc; substring match. |
| `--since`, `--until` | Filter by `LastUpdated` range | `pkg/commands/search.go` | Uses frontmatter `LastUpdated`. |
| `--updated-since` | Filter by `LastUpdated >= t` | `pkg/commands/search.go` | Same as `--since` but separately named. |
| `--created-since` | Filter by “created” time | `pkg/commands/search.go` | Uses filesystem `ModTime()` as a proxy. |
| `--files` | Suggest related files (not docs) | `pkg/commands/search.go` | Blends `RelatedFiles` + git + ripgrep + git status. |
| `--root <dir>` | Docs root | `internal/workspace/config.go` | If `--root` is the default (`ttmp`), `.ttmp.yaml` may override it. |
| `--print-template-schema` | Print schema of template data | `pkg/commands/search.go` | For template authoring; bypasses search execution. |
| `--schema-format json\|yaml` | Schema output format | `pkg/commands/search.go` | Affects only `--print-template-schema`. |

### A mental model in one diagram

```text
CLI (cobra)
  |
  v
SearchCommand (pkg/commands/search.go)
  |
  |-- Resolve docs root (.ttmp.yaml / --root / git / cwd)
  |
  |-- DiscoverWorkspace (internal/workspace)
  |     - Root, ConfigDir, RepoRoot
  |     - paths.Resolver anchors
  |
  |-- InitIndex (internal/workspace)
  |     - open in-memory SQLite
  |     - create schema (docs, doc_topics, related_files)
  |     - ingest documents (walk + parse frontmatter + store norms)
  |
  |-- QueryDocs (internal/workspace)
  |     - compile SQL from scope + filters
  |     - query index, hydrate topics + related_files
  |
  |-- Post-filters (command layer)
  |     - content substring match (--query)
  |     - external source match (--external-source)
  |     - date filters (--since/--until/--created-since/--updated-since)
  |
  `-- Output
        - human lines OR
        - glazed rows (--with-glaze-output)
        - optional postfix template (templates/doc/search.templ)
```

## Usage Examples

### 1) Content search (full-text substring)

Search across all docs for a term:

```bash
cd docmgr
GOWORK=off go run ./cmd/docmgr doc search --query "Workspace.QueryDocs"
```

Ticket-scoped content search:

```bash
GOWORK=off go run ./cmd/docmgr doc search --ticket REFACTOR-TICKET-REPOSITORY-HANDLING --query "Workspace.QueryDocs"
```

### 2) Metadata-only filtering (no content query)

Find “design-doc” docs tagged with any of the provided topics:

```bash
GOWORK=off go run ./cmd/docmgr doc search \
  --topics docmgr,cli \
  --doc-type design-doc
```

### 3) Reverse lookup: file → docs (`--file`)

Find docs that reference a particular code file:

```bash
GOWORK=off go run ./cmd/docmgr doc search \
  --file docmgr/pkg/commands/search.go \
  --with-glaze-output --output table
```

Basename-only queries are supported (suffix match):

```bash
GOWORK=off go run ./cmd/docmgr doc search \
  --file search.go \
  --with-glaze-output --output table
```

### 4) Reverse lookup: directory → docs (`--dir`)

Find docs that reference any file under a directory:

```bash
GOWORK=off go run ./cmd/docmgr doc search \
  --dir pkg/commands \
  --with-glaze-output --output table
```

### 5) Script-friendly output (structured mode)

Emit JSON rows (the default in glaze mode in this repo):

```bash
GOWORK=off go run ./cmd/docmgr doc search \
  --query "Workspace.QueryDocs" \
  --with-glaze-output --output json
```

If you specifically want **only one field per line** (e.g., just the doc paths), use the glazed “select” facility with template output:

```bash
GOWORK=off go run ./cmd/docmgr doc search \
  --query "Workspace.QueryDocs" \
  --with-glaze-output --output template --select path
```

### 6) “Which files should I relate?” suggestions (`--files`)

For a ticket, suggest candidate files by blending:

- existing `RelatedFiles` across docs (scoped),
- git history (recent commit activity),
- git status (modified/staged/untracked),
- ripgrep matches for the query/topics.

```bash
GOWORK=off go run ./cmd/docmgr doc search \
  --ticket 001-ADD-DOCMGR-UI \
  --topics docmgr \
  --files
```

### 7) Template authoring helper

Print the schema (shape) of the data that’s made available to verb postfix templates:

```bash
GOWORK=off go run ./cmd/docmgr doc search --print-template-schema --schema-format json
```

This does not run a search; it just prints a schema preview.

## Implementation Guide (Deep Dive)

### 0) The primary entry points (when you’re reading code)

If you want to understand search by grepping and reading, start here:

- CLI wiring and flags:
  - `docmgr/cmd/docmgr/cmds/doc/search.go`
  - `docmgr/pkg/commands/search.go`
- Indexing and query engine:
  - `docmgr/internal/workspace/index_builder.go`
  - `docmgr/internal/workspace/sqlite_schema.go`
  - `docmgr/internal/workspace/query_docs.go`
  - `docmgr/internal/workspace/query_docs_sql.go`
- Path normalization (crucial for `--file` / `--dir`):
  - `docmgr/internal/paths/resolver.go`

### 1) How the CLI works (cobra + glazed dual mode)

At the cobra layer, search is a standard cobra subcommand under `docmgr doc`:

- `docmgr/cmd/docmgr/cmds/root.go` attaches the `doc` command subtree and also adds an alias command named `search` that points at the same cobra command as `doc search`.
- `docmgr/cmd/docmgr/cmds/doc/search.go` builds a cobra command from the glazed command definition in `pkg/commands/search.go`:
  - `cli.WithDualMode(true)` enables two implementations on the same verb:
    - `Run(...)` (“bare” mode): human printing.
    - `RunIntoGlazeProcessor(...)` (“glaze” mode): row emission.
  - `cli.WithGlazeToggleFlag("with-glaze-output")` makes `--with-glaze-output` switch the command into glazed mode.

In this repo, glazed mode defaults to `--output json` due to `docmgr/cmd/docmgr/cmds/common/common.go` setting glazed defaults for all glazed commands.

**Autocompletion:** `docmgr/cmd/docmgr/cmds/doc/search.go` configures carapace completions for flags like `--ticket`, `--topics`, `--doc-type`, `--status`, etc.

### 2) Search is two systems: index-backed filters + post-filters

Conceptually, the search command breaks down into two parts:

1. **Index-backed “candidate selection”** via `Workspace.QueryDocs(...)`:
   - ticket scope
   - topics/doc-type/status filters
   - reverse lookup filters (`--file`, `--dir`)
   - plus visibility controls (include/exclude archive/scripts/control docs)
2. **Post-filtering** in the command layer:
   - content substring search (`--query`)
   - external source matching (`--external-source`)
   - date filters (`--since`, `--until`, `--created-since`, `--updated-since`)

This split is not accidental. The index exists to handle the heavy structural filtering (especially reverse lookup) in a predictable way. Content search is currently simple enough to do as a scan over candidate bodies.

### 3) Workspace discovery and root resolution

Before search can build an index, it needs to know:

- where docs live (the “docs root”),
- where the config lives (`.ttmp.yaml` directory),
- where the repo root is (for path canonicalization).

This is handled by `internal/workspace.DiscoverWorkspace(...)`:

- It resolves the docs root via `internal/workspace.ResolveRoot(...)`:
  - if you pass a non-default `--root` it uses it,
  - otherwise it uses `.ttmp.yaml` if found,
  - otherwise it falls back to `<git-root>/ttmp` and then `<cwd>/ttmp`.
- It computes `ConfigDir` from the `.ttmp.yaml` location (or uses heuristics).
- It computes `RepoRoot` via repository discovery.
- It constructs a `paths.Resolver` anchored on:
  - docs root (`DocsRoot`)
  - config directory (`ConfigDir`)
  - repository root (`RepoRoot`)

These anchors are what enable the reverse lookup matching to work across many path forms.

### 4) The index: in-memory SQLite, rebuilt per invocation

`Workspace.InitIndex(ctx, opts)` builds the full queryable state for a CLI invocation.

Key properties:

- The database is SQLite in memory (`file:docmgr_workspace_<n>?mode=memory&cache=shared`).
- A new unique in-memory DB is created per `Workspace` instance (to avoid cross-test leakage).
- The schema is created fresh and then populated by walking all documents under the root.

#### 4.1 Schema overview (what gets indexed)

The schema is defined in `internal/workspace/sqlite_schema.go`. Search-relevant tables:

- `docs`
  - one row per markdown file
  - stores path, ticket_id, doc_type, status, intent, title, last_updated
  - stores parse_ok + parse_err for diagnostics
  - stores path tags (`is_archived_path`, `is_scripts_path`, `is_control_doc`, …)
  - optionally stores the full body (`body`) when `IncludeBody=true`
- `doc_topics`
  - `(doc_id, topic_lower)` primary key
  - powers case-insensitive topic matching
- `related_files`
  - one row per `RelatedFiles` entry from frontmatter
  - stores both the raw string and multiple normalized representations:
    - `norm_canonical`, `norm_repo_rel`, `norm_docs_rel`, `norm_doc_rel`, `norm_abs`, `norm_clean`

This denormalization of “path representations” is what makes reverse lookup robust.

#### 4.2 Ingestion flow (how the tables get populated)

Ingestion happens in `internal/workspace/index_builder.go`:

Pseudocode (simplified but faithful):

```text
InitIndex(IncludeBody):
  db := openInMemorySQLite()
  createWorkspaceSchema(db)
  walk docsRoot:
    for each *.md:
      (doc, body, err) := ReadDocumentWithFrontmatter(path)
      if err:
        insert docs row with parse_ok=0, parse_err=err, ticket_id inferred from path
        continue
      insert docs row with ticket/doc_type/status/title/last_updated
      if IncludeBody: store body
      insert topics into doc_topics (lowercase + original)
      for each RelatedFiles entry:
        resolver anchored at doc path
        normalized := resolver.Normalize(raw_related_path)
        insert related_files row with norm_* and raw_path
  commit tx
```

Important details:

- **Skip rules**:
  - directories named `.meta` are skipped
  - directories starting with `_` are skipped
  - (also, `internal/documents.WalkDocuments` skips underscore dirs independently)
- **Parse errors are indexed**:
  - docs with invalid frontmatter still get a `docs` row (`parse_ok=0`),
  - and `ticket_id` is best-effort inferred from the path layout,
  - which allows `doctor`/repair flows and ticket scoping to “see” broken docs.

### 5) QueryDocs: the index-backed API that search builds on

`Workspace.QueryDocs(ctx, DocQuery)` is the primary internal API for asking questions like:

- “give me all docs in ticket X”
- “give me all docs tagged with topic Y”
- “give me docs that reference file Z”

#### 5.1 Public request/response types (API reference)

Defined in `internal/workspace/query_docs.go`:

```text
DocQuery:
  Scope:   Repo | Ticket(TicketID) | Doc(DocPath)
  Filters:
    Ticket, DocType, Status
    TopicsAny []string
    RelatedFile []string
    RelatedDir  []string
  Options:
    IncludeBody
    IncludeErrors
    IncludeDiagnostics
    IncludeArchivedPath / IncludeScriptsPath / IncludeControlDocs
    OrderBy (path | last_updated)
    Reverse (descending)

DocQueryResult:
  Docs []DocHandle
  Diagnostics []Taxonomy

DocHandle:
  Path (slash-cleaned)
  Doc (*models.Document)   // may be nil if parse failed
  Body (optional)
  ReadErr (optional)
```

#### 5.2 How SQL gets compiled

SQL compilation is implemented in `internal/workspace/query_docs_sql.go`:

- It builds a `WHERE` list and `Args` slice based on:
  - visibility flags (hide archive/scripts/control docs unless included),
  - parse_ok filter (unless `IncludeErrors=true`),
  - scope (repo/ticket/doc),
  - metadata filters,
  - topic OR clause,
  - related file/dir EXISTS subqueries,
  - ordering.

Topics are indexed as lowercase (`doc_topics.topic_lower`) to avoid case sensitivity issues for topic matching.

#### 5.3 Reverse lookup in SQL (file and directory matching)

Reverse lookup is the most subtle part of search, and it’s worth understanding precisely.

##### 5.3.1 How `RelatedFiles` paths are produced (why `doc relate` matters)

Reverse lookup quality is only as good as the `RelatedFiles` data stored in documents.

When you run `docmgr doc relate`, it **canonicalizes** both existing and newly added related file paths before writing them back into the document frontmatter. The canonicalization strategy is implemented in `docmgr/pkg/commands/relate.go` via `canonicalizeWithResolver(...)`, which uses a `paths.Resolver` anchored at the *target document’s* location.

In practice, this means:

- If the referenced file is inside the repo root, the stored path tends to become repo-relative (portable, stable).
- If the referenced file is outside the repo root, the stored path tends to become absolute (not portable, but unambiguous).
- If the resolver can’t confidently anchor a path, the stored value may remain a cleaned relative string.

This write-time normalization is the “first line of defense” for reverse lookup. The index and query layers still include fallbacks, but they work best when frontmatter is already mostly canonical.

##### 5.3.2 `paths.Resolver.Normalize`: anchors and canonical selection

The `paths.Resolver` is the “source of truth” for turning user strings into comparable path keys. It’s implemented in `docmgr/internal/paths/resolver.go`.

**Anchor search order (for relative inputs)**

When the input is not absolute, normalization tries a sequence of base directories (“anchors”), in order:

1. repo root (`AnchorRepo`)
2. current document directory (`AnchorDoc`)
3. config directory (`AnchorConfig`)
4. docs root (`AnchorDocsRoot`)
5. docs parent (`AnchorDocsParent`)

The resolver returns the first anchored resolution where the resulting path **exists** on disk. If none exist, it returns a fallback normalization based on the first viable anchor, and if even that fails it falls back to the cleaned input string.

**Canonical representation**

Once a candidate absolute path is chosen (or a fallback is selected), the resolver computes multiple representations and then chooses a canonical display/search key:

```text
Canonical = firstNonEmpty(RepoRelative, DocsRelative, DocRelative, Abs)
```

This is why repo-relative strings often “win” and become the stored `RelatedFiles` values after `doc relate`.

**Data stored per RelatedFiles entry** (index-time):

- raw: the exact string in frontmatter
- normalized keys: `norm_canonical`, `norm_repo_rel`, `norm_docs_rel`, `norm_doc_rel`, `norm_abs`, `norm_clean`

**Query-time inputs**:

- `--file <raw>` becomes a key set via `queryPathKeys(resolver, raw)`
- `--dir <raw>` becomes prefix patterns derived from the same key set

`queryPathKeys(...)` returns a de-duplicated list of comparable strings:

- `resolver.Normalize(raw)` yields:
  - canonical (prefers repo-relative, then docs-relative, then doc-relative, then abs)
  - repo-relative
  - docs-relative
  - doc-relative (may include `../`)
  - abs
  - original clean form
- the workspace layer also adds a cleaned path derived from `filepath.Clean(raw)` with slash normalization

**File match SQL strategy**:

- For each of the persisted columns (`norm_*` and `raw_path`):
  - match if `col IN (<keys>)`
  - additionally, if the user query is basename-only (no slashes), match if `col LIKE "%/<basename>"`

This is why `--file search.go` works without you needing to know the full path.

**Directory match SQL strategy**:

- For each persisted column:
  - match if `col LIKE "<prefix>/%"` for any computed prefix

**Case sensitivity note:** unlike topics (which are lowercased in the index), path keys are stored and compared largely as-is (with slash normalization and cleaning, but not forced to lowercase). Exact `IN (...)` matches are therefore case-sensitive. On most Linux repos this is fine because paths are consistently cased; if you have mixed-case paths, expect surprises.

#### 5.4 Why the command layer still “re-matches” for `--file`

Even though the SQL layer already filters candidates for `--file`, `pkg/commands/search.go` also computes:

- `matchedFiles`: which `RelatedFiles` entries matched the query
- `matchedNotes`: the corresponding notes

It does this by re-normalizing per document using a resolver anchored at that document path and applying:

- `paths.MatchPaths(queryNorm, relatedNorm)` (a fuzzy matcher), and
- explicit basename/suffix checks.

This is a **presentation feature**: it explains to the user *which* RelatedFiles entry caused a match and surfaces its note.

`paths.MatchPaths(...)` is intentionally fuzzy and is implemented in `docmgr/internal/paths/resolver.go`. At a high level it:

1. compares multiple normalized representations (canonical, repo-relative, docs-relative, doc-relative, abs, original clean),
2. tries suffix matching on up to the last 3 path segments (to catch “close enough” matches),
3. finally falls back to substring containment checks.

The “real” filtering for `--file` is still the SQL layer’s `EXISTS (SELECT … FROM related_files …)` clause; the fuzzy matcher exists to produce better *explanations* in output.

### 6) Post-filters: what is *not* indexed (yet)

After `QueryDocs` returns candidate docs, the search command applies additional filtering:

#### 6.1 Content search (`--query`)

Implementation: `pkg/commands/search.go`

- The document body is lowercased and checked with `strings.Contains(...)` against a lowercased query.
- There is no tokenization, stemming, ranking, or boolean logic.
- The snippet is derived by taking ~100 characters of context around the first match.

Practical implications:

- Queries are “dumb but predictable”.
- Performance depends on how many bodies you scan (which depends on how selective your index-backed filters are).

#### 6.2 External source filtering (`--external-source`)

Implementation: `pkg/commands/search.go` + `pkg/commands/document_utils.go`

- The command re-reads the frontmatter for each candidate doc and checks `ExternalSources`.
- Matching is substring-based, in both directions:
  - `externalSourceString contains query` OR `query contains externalSourceString`

This is robust when users paste slightly different URL fragments.

#### 6.3 Date filters (`--since`, `--until`, `--updated-since`, `--created-since`)

Implementation: `pkg/commands/search.go`

- `--since` / `--until` / `--updated-since` are compared against frontmatter `LastUpdated` (RFC3339 timestamp).
- `--created-since` uses `os.Stat(path).ModTime()` as a proxy “created time”.

Date parsing supports:

- absolute formats like `2025-01-01`, RFC3339, etc.
- relative strings like `2 weeks ago`, `last month`, `today`, `yesterday`.

### 7) File suggestion mode (`--files`): how it decides what to output

`--files` switches the command into a different behavior: it emits candidate file paths instead of documents.

In glazed mode, this is implemented in `SearchCommand.suggestFiles(...)`. In bare mode, the logic is similar but prints directly.

The heuristic sources are:

1. **Existing `RelatedFiles`** across documents (scoped by `--ticket` and optionally filtered by `--topics`).
2. **Git history**: recently modified code-like files from the last ~30 commits.
3. **Ripgrep**: `rg --files-with-matches` for the first search term (query or a topic) across common code types.
4. **Git status**: modified/staged/untracked code-like files.

The suggestion stream is not a ranked ML system; it’s a “multi-sensor” list. You are expected to apply judgment.

### 8) Postfix verb templates (optional output customization)

After printing human output (bare mode), `doc search` attempts to render a postfix template if present under the docs root.

Template lookup rules (`internal/templates/verb_output.go`):

- For `doc search`, it tries: `<root>/templates/doc/search.templ`
- As a fallback, it tries: `<root>/templates/search.templ`

The template receives a data map containing:

- `Verbs`, `Root`, `Now`, `Settings`
- plus verb-specific fields (for search: `Query`, `TotalResults`, `Results[...]`)

Use `--print-template-schema` to see the expected shape of this data.

## Extension Guide (How to Extend Search Without Breaking It)

This section is written as a “change playbook”: if you want to add features or change semantics, these are the seams you should touch.

### A) Adding a new index-backed filter (fast and queryable)

Example goals:

- filter by `Intent`
- filter by `Owners`
- filter by “has related files”

General approach:

1. **Decide where the data should live in the index**:
   - new column in `docs`, or
   - new table keyed by `doc_id` (like `doc_topics`), or
   - new normalized “side table” similar to `related_files`.
2. **Populate it during ingestion**:
   - `internal/workspace/index_builder.go`
3. **Add query compilation support**:
   - extend `DocFilters`
   - extend `compileDocQueryWithParseFilter(...)` in `internal/workspace/query_docs_sql.go`
4. **Expose it in the CLI**:
   - add fields to `SearchSettings` and flags in `pkg/commands/search.go`
   - pass the filter into `Workspace.QueryDocs(...)`
5. **Add/extend scenario tests**:
   - scripts under `docmgr/test-scenarios/` and/or Go tests in `internal/workspace/*_test.go`

### B) Adding a post-filter (simple, but can be slow)

Example goals:

- regex content match
- “title contains …”
- “external sources match …” (already a post-filter)

If you can implement it as a post-filter after `QueryDocs`, it’s often only a change in `pkg/commands/search.go`. The tradeoff is performance: you may be scanning more candidate docs and re-reading files.

### C) Adding true full-text search (SQLite FTS) (recommended future evolution)

If you want ranking, stemming-ish behavior, or faster content search at scale, the natural next step is SQLite FTS5.

High-level design:

1. In schema creation, add:
   - an FTS virtual table `docs_fts(doc_id UNINDEXED, body)` (or `content=docs` style)
2. During ingestion, populate `docs_fts` for parse_ok docs when `IncludeBody=true`.
3. In query compilation, if `--query` is present, join against `docs_fts`:
   - `WHERE docs_fts MATCH ?`
4. Optionally order by ranking (`bm25(docs_fts)`).

Pseudocode sketch:

```text
if query != "":
  sql += "JOIN docs_fts fts ON fts.doc_id = d.doc_id"
  where += "fts MATCH ?"
  args += query
  orderBy = "bm25(fts)"
```

This would move “content search” from post-filter to index-backed filter and make search much faster on large workspaces.

### D) Extending reverse lookup semantics (`--file` / `--dir`)

If you need to change how file/dir matching works, be very careful: this is the UX-critical surface.

Where semantics live today:

- Index-time normalization:
  - `internal/paths/resolver.go`
  - `internal/workspace/normalization.go`
  - `internal/workspace/index_builder.go`
- Query-time key generation:
  - `internal/workspace/query_docs.go` (`queryPathKeys`)
  - `internal/workspace/query_docs_sql.go` (`relatedFileExistsClause`, `relatedDirExistsClause`)
- Presentation-time extra matching for `--file`:
  - `pkg/commands/search.go` (`paths.MatchPaths`, basename fallback)

Concrete suggestions:

- If you tighten matching, add explicit compatibility notes and update scenario scripts that rely on fuzzy matching.
- If you loosen matching, watch for false positives (especially with basename-only queries).

### E) Improving `--created-since`

Today, “created” is approximated by file `ModTime()`. If you need real semantics, consider:

- adding a `CreatedAt` field in frontmatter, or
- deriving created time from git history (first commit touching the file), or
- storing a “first seen” timestamp in an on-disk cache (more complexity).

## Related

- Diary (research trail for this document): `docmgr/ttmp/2026/01/03/001-ADD-DOCMGR-UI--add-docmgr-web-ui/reference/01-diary.md`
- Search implementation: `docmgr/pkg/commands/search.go`
- Workspace index/query engine: `docmgr/internal/workspace/`

## Related

<!-- Link to related documents or resources -->
