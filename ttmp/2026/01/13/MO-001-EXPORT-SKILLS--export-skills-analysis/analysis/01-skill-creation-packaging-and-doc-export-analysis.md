---
Title: Skill creation, packaging, and doc export analysis
Ticket: MO-001-EXPORT-SKILLS
Status: active
Topics:
    - documentation
    - tools
    - docmgr
    - glaze
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/SKILL.md
      Note: Skill creation and packaging guidance
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/init_skill.py
      Note: Initialization workflow and template generation
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/package_skill.py
      Note: Packaging logic and zip layout
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/quick_validate.py
      Note: Validation rules for SKILL.md frontmatter
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-installer/SKILL.md
      Note: Installation path conventions
    - Path: ../../../../../../../glazed/pkg/doc/topics/01-help-system.md
      Note: Help system model and metadata
    - Path: ../../../../../../../glazed/pkg/doc/topics/14-writing-help-entries.md
      Note: Frontmatter template and authoring rules
    - Path: ../../../../../../../glazed/pkg/help/help.go
      Note: Markdown loader and AddSection flow
    - Path: ../../../../../../../glazed/pkg/help/store/store.go
      Note: SQLite schema for help sections
    - Path: internal/documents/frontmatter.go
      Note: Frontmatter parsing and writing
    - Path: internal/workspace/sqlite_export.go
      Note: Workspace export to SQLite
    - Path: internal/workspace/sqlite_schema.go
      Note: Workspace index schema
    - Path: pkg/doc/doc.go
      Note: Embedded doc loading for docmgr
    - Path: pkg/doc/embedded_docs.go
      Note: Export of embedded docs
    - Path: pkg/models/document.go
      Note: Document frontmatter model
ExternalSources:
    - https://agentskills.io/home
    - https://agentskills.io/specification
    - https://cursor.com/docs/context/skills
Summary: ""
LastUpdated: 2026-01-13T09:43:13-05:00
WhatFor: ""
WhenToUse: ""
---


# Analysis

## Goal and scope

This analysis documents how skills are created and packaged in the local Codex skills directory, and how Glazed and docmgr documentation can be packaged for distribution or export. The intent is to provide a concrete, implementation-grounded view that references the exact scripts, data structures, and schemas used in the repository, with enough detail to build an exporter or to audit existing artifacts.

The analysis is strictly based on local sources. For skill creation and packaging, the authoritative reference is the skill-creator tool and its scripts in `/home/manuel/.codex/skills/.system/skill-creator/`. For Glazed docs, the reference is the help system documentation and code in `glazed/pkg/doc` and `glazed/pkg/help`. For docmgr documents, the reference is the document frontmatter model, workspace index schema, and SQLite export flow in the `docmgr/` tree.

## Source files and symbols reviewed

- `/home/manuel/.codex/skills/.system/skill-creator/SKILL.md` (skill creation workflow, packaging steps, and naming rules)
- `/home/manuel/.codex/skills/.system/skill-creator/scripts/init_skill.py` (skill initialization flow, `normalize_skill_name`, `init_skill`, `create_resource_dirs`)
- `/home/manuel/.codex/skills/.system/skill-creator/scripts/package_skill.py` (packaging flow, `package_skill`)
- `/home/manuel/.codex/skills/.system/skill-creator/scripts/quick_validate.py` (validation rules, `validate_skill`)
- `/home/manuel/.codex/skills/.system/skill-installer/SKILL.md` (installation path conventions for skills)
- `/home/manuel/.codex/skills/dist/docmgr.skill` (example packaged skill structure)
- `/home/manuel/workspaces/2026-01-13/install-skills/glazed/pkg/doc/topics/01-help-system.md` (help system model, section metadata)
- `/home/manuel/workspaces/2026-01-13/install-skills/glazed/pkg/doc/topics/14-writing-help-entries.md` (frontmatter format and embed-based loading)
- `/home/manuel/workspaces/2026-01-13/install-skills/glazed/pkg/doc/doc.go` (embed-based doc loading)
- `/home/manuel/workspaces/2026-01-13/install-skills/glazed/pkg/help/help.go` (markdown loading, `LoadSectionFromMarkdown`, `LoadSectionsFromFS`)
- `/home/manuel/workspaces/2026-01-13/install-skills/glazed/pkg/help/store/store.go` (SQLite schema for help sections)
- `/home/manuel/workspaces/2026-01-13/install-skills/docmgr/pkg/doc/doc.go` (embed-based doc loading for docmgr)
- `/home/manuel/workspaces/2026-01-13/install-skills/docmgr/pkg/doc/embedded_docs.go` (`ReadEmbeddedMarkdownDocs` for exporting embedded docs)
- `/home/manuel/workspaces/2026-01-13/install-skills/docmgr/internal/documents/frontmatter.go` (frontmatter parsing and writing)
- `/home/manuel/workspaces/2026-01-13/install-skills/docmgr/pkg/models/document.go` (`Document` schema and YAML frontmatter model)
- `/home/manuel/workspaces/2026-01-13/install-skills/docmgr/internal/workspace/sqlite_schema.go` (workspace index schema)
- `/home/manuel/workspaces/2026-01-13/install-skills/docmgr/internal/workspace/sqlite_export.go` (`ExportIndexToSQLiteFile`, README table population)

External references reviewed:

- `https://agentskills.io/home`
- `https://agentskills.io/specification`
- `https://cursor.com/docs/context/skills`

## Part A: Skill creation and packaging in ~/.codex/skills

### On-disk layout of ~/.codex/skills

The local skills directory contains both development-time skill folders and packaged artifacts. The observed top-level structure includes:

- `/home/manuel/.codex/skills/<skill-name>/` directories, each containing a `SKILL.md` file and optional `scripts/`, `references/`, and `assets/` resources.
- `/home/manuel/.codex/skills/.system/` for system skills such as `skill-creator` and `skill-installer`.
- `/home/manuel/.codex/skills/dist/` containing packaged `.skill` files created by the packager.
- `/home/manuel/.codex/skills/*.zip` which appear to be bundled copies of skills (for example `skills.zip` and `diary.zip`).

The packaging tooling produces `.skill` artifacts that are zip archives with a `.skill` extension. An example inspection of `/home/manuel/.codex/skills/dist/docmgr.skill` shows entries like `docmgr/SKILL.md` and `docmgr/references/docmgr.md`, which confirms that the package stores the skill directory name as the root folder inside the archive.

### Skill metadata format and constraints

Skills are defined by a single `SKILL.md` file at the skill root. Two sources describe the structure and constraints:

- The skill-creator guide (`/home/manuel/.codex/skills/.system/skill-creator/SKILL.md`) describes the expected directory structure, frontmatter expectations, and packaging process.
- The validator (`/home/manuel/.codex/skills/.system/skill-creator/scripts/quick_validate.py`) enforces concrete constraints on frontmatter keys and value shapes.

The validator accepts the following YAML frontmatter keys in `SKILL.md` (see `validate_skill`):

- `name` (required)
- `description` (required)
- `license` (optional)
- `allowed-tools` (optional)
- `metadata` (optional)

The skill-creator guidance emphasizes only `name` and `description` and discourages extra keys, but the validator explicitly allows the additional keys above. This means any packaging flow that uses `quick_validate.py` can accept `license`, `allowed-tools`, or `metadata` without failing validation.

The `name` field is validated as:

- Lowercase letters, digits, and hyphens only (`^[a-z0-9-]+$`).
- No leading or trailing hyphen, and no consecutive hyphens (`--` rejected).
- Maximum length 64 characters.

The `description` field is validated as:

- A string.
- No angle brackets `<` or `>`.
- Maximum length 1024 characters.

#### Data format: minimal SKILL.md frontmatter

```yaml
---
name: docmgr
description: Documentation management with the docmgr CLI and ticket workspaces. Use when working on ticket docs or docmgr CLI workflows.
---
```

#### Data format: full skill layout

```
<skill-name>/
  SKILL.md
  scripts/          # optional
  references/       # optional
  assets/           # optional
```

### Skill initialization workflow (init_skill.py)

The initializer script (`/home/manuel/.codex/skills/.system/skill-creator/scripts/init_skill.py`) encodes the deterministic creation flow. It includes a name normalization routine and a template for `SKILL.md`. It will optionally create resource directories and example files.

Key functions and flow:

- `normalize_skill_name` lowercases and converts any non-alphanumeric sequence to a single hyphen, trims leading or trailing hyphens, and collapses consecutive hyphens.
- `title_case_skill_name` converts hyphenated names into a human-readable Title Case display used for headings.
- `parse_resources` validates the `--resources` value against the allowed set `{scripts,references,assets}`, deduplicates, and exits on invalid values.
- `init_skill` creates the skill directory, writes `SKILL.md` from a template, and optionally creates resource directories (and optional example files via `create_resource_dirs`).

#### Pseudocode: init_skill.py (simplified)

```text
normalize_skill_name(raw_name):
  lower-case
  replace any non [a-z0-9] runs with "-"
  trim leading/trailing "-"
  collapse multiple "-"

init_skill(name, path, resources, examples):
  skill_dir = resolve(path) + "/" + name
  if skill_dir exists: error
  mkdir(skill_dir)
  write SKILL.md using SKILL_TEMPLATE
  for each resource in resources:
    mkdir(skill_dir/resource)
    if examples: write example file
  print next steps
```

### Skill packaging workflow (package_skill.py and quick_validate.py)

The packager (`/home/manuel/.codex/skills/.system/skill-creator/scripts/package_skill.py`) zips the entire skill directory and validates the `SKILL.md` frontmatter before packaging. Validation is delegated to `quick_validate.py`, which performs basic frontmatter and naming checks.

Key steps in `package_skill`:

1. Resolve the skill path and verify it exists and is a directory.
2. Verify `SKILL.md` exists at the root.
3. Run `validate_skill` and abort on errors.
4. Determine output directory (default is current directory, or user-specified).
5. Create `<skill-name>.skill` via `zipfile.ZipFile`.
6. Add all files recursively under the skill directory to the zip, with archive names relative to the skill directory's parent. This ensures the top-level folder inside the zip is `<skill-name>/`.

#### Pseudocode: package_skill.py (simplified)

```text
package_skill(skill_path, output_dir):
  skill_path = resolve(skill_path)
  assert skill_path is directory
  assert skill_path/SKILL.md exists
  (valid, message) = validate_skill(skill_path)
  if not valid: error
  output_dir = resolve(output_dir or cwd)
  skill_name = basename(skill_path)
  skill_filename = output_dir/skill_name + ".skill"

  open zipfile(skill_filename):
    for file in rglob(skill_path, "*"):
      if file is regular:
        arcname = file relative to skill_path.parent
        zip.write(file, arcname)
```

#### Data format: .skill file structure

The `.skill` file is a zip archive. Using `docmgr.skill` as a reference, the archive layout looks like:

```
<skill-name>.skill (zip)
  <skill-name>/SKILL.md
  <skill-name>/references/...
  <skill-name>/scripts/...
  <skill-name>/assets/...
```

### Packaging artifacts in ~/.codex/skills

The installed skills directory shows two patterns for packaged artifacts:

- `.skill` files in `/home/manuel/.codex/skills/dist/`, created by the packager script.
- `.zip` files in `/home/manuel/.codex/skills/`, which appear to be multi-skill bundles or snapshots (for example `skills.zip` and `diary.zip`). These are not produced by `package_skill.py`, but can be treated as generic zip bundles if needed.

### Skill installation behavior

The skill-installer guide (`/home/manuel/.codex/skills/.system/skill-installer/SKILL.md`) documents the installation workflow and confirms that installed skills are placed in `$CODEX_HOME/skills/<skill-name>`, which defaults to `~/.codex/skills`. This behavior is important for packaging flows, because it clarifies the canonical source folder that should be zipped when preparing a `.skill` artifact.

## Part B: Packaging Glazed documentation

### Documentation frontmatter format for help sections

Glazed documentation entries are Markdown files with YAML frontmatter, stored under `glazed/pkg/doc/topics`. The two documentation topics referenced here (`01-help-system.md` and `14-writing-help-entries.md`) show the expected frontmatter fields and provide guidance on consistent style.

From `glazed/pkg/doc/topics/01-help-system.md`, the frontmatter describes the section metadata used to populate the help system. From `glazed/pkg/doc/topics/14-writing-help-entries.md`, the frontmatter template is explicit about each field.

#### Data format: Glazed help section frontmatter

```yaml
---
Title: Help System
Slug: help-system
Short: Glazed provides a powerful, queryable help system for creating rich CLI documentation with sections, metadata, and programmatic access.
Topics:
- help
- documentation
Commands:
- help
Flags:
- flag
- topic
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---
```

### Loading flow: markdown into HelpSystem

The Glazed help system loads sections via `LoadSectionFromMarkdown` and `HelpSystem.LoadSectionsFromFS` in `glazed/pkg/help/help.go`.

- `LoadSectionFromMarkdown` uses `github.com/adrg/frontmatter` to parse frontmatter into a metadata map and returns the remaining Markdown as `Content`. It maps fields like `Title`, `Slug`, `Short`, `Topics`, `Commands`, `Flags`, `IsTopLevel`, `IsTemplate`, `ShowPerDefault`, and `Order` into a `Section` struct.
- `HelpSystem.LoadSectionsFromFS` walks a filesystem recursively, skips `README.md` and non-markdown files, reads `.md`, then calls `LoadSectionFromMarkdown` and `AddSection` for each section.
- `HelpSystem.AddSection` upserts into a SQLite-backed store.

#### Pseudocode: Glazed help loading

```text
LoadSectionsFromFS(fs, dir):
  entries = read dir
  for entry in entries:
    if entry is dir: recurse
    if entry is file:
      if not .md or is README.md: continue
      bytes = read file
      section = LoadSectionFromMarkdown(bytes)
      AddSection(section)

LoadSectionFromMarkdown(bytes):
  metadata, rest = frontmatter.Parse(bytes)
  section = new Section
  section.Title = metadata["Title"]
  section.Slug = metadata["Slug"]
  section.Short = metadata["Short"]
  section.SectionType = parse SectionType or default GeneralTopic
  section.Topics/Commands/Flags = list from metadata
  section.IsTopLevel, IsTemplate, ShowPerDefault
  section.Content = rest
  validate slug and title
```

### Data storage: SQLite schema for help sections

The help system stores sections in SQLite (`glazed/pkg/help/store/store.go`). The schema is created in `createTables` and includes a `sections` table with metadata fields and an optional FTS5 table. This makes packaging into a database straightforward.

#### Data format: help system schema (key fields)

```sql
CREATE TABLE sections (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  slug TEXT NOT NULL UNIQUE,
  section_type INTEGER NOT NULL,
  title TEXT NOT NULL,
  sub_title TEXT,
  short TEXT,
  content TEXT,
  topics TEXT,
  flags TEXT,
  commands TEXT,
  is_top_level BOOLEAN DEFAULT FALSE,
  is_template BOOLEAN DEFAULT FALSE,
  show_per_default BOOLEAN DEFAULT FALSE,
  order_num INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Packaging strategy for Glazed docs

There are two native packaging strategies already used in the codebase:

1. **Embed documentation into a binary** using `go:embed` and `HelpSystem.LoadSectionsFromFS`.
   - Implemented in `glazed/pkg/doc/doc.go` via `AddDocToHelpSystem`.
   - At build time, all `.md` files under `glazed/pkg/doc` are embedded into the binary.
   - At runtime, the embedded filesystem is walked and loaded into the SQLite store.

2. **Persist a SQLite help database** for distribution or offline use.
   - The store supports file-based SQLite via `store.New(dbPath)`.
   - The packaging flow is: initialize a `HelpSystem` with a file-based store, load the embedded or file-based docs, then ship the resulting database file.

#### Pseudocode: export Glazed docs to SQLite

```text
st = store.New("/path/to/glazed-docs.sqlite")
hs = help.NewHelpSystemWithStore(st)
err = doc.AddDocToHelpSystem(hs)  // uses embedded FS
st.Close()

# The SQLite file now contains the sections table populated with docs.
```

### Notes on section metadata coupling

The help system relies on consistent frontmatter values for filtering and query DSL operations. The following fields are directly queryable via predicates or DSL terms (see `glazed/pkg/help/dsl_bridge.go` and `glazed/pkg/help/query.go`):

- `SectionType` for type filters (`type:example`, `type:tutorial`).
- `Topics` for topic filters (`topic:database`).
- `Commands` and `Flags` for command and flag filters.
- `IsTopLevel` and `ShowPerDefault` to control visibility in default help output.

Packaging must preserve these fields exactly, because the query DSL depends on them. This is why the frontmatter format in `14-writing-help-entries.md` is critical.

## Part C: Packaging docmgr documents

### Two distinct doc sets in docmgr

docmgr has two related but distinct documentation surfaces:

1. **docmgr help docs** (used in `docmgr help`), stored as Markdown under `docmgr/pkg/doc` and embedded in the binary.
2. **docmgr ticket documents** (stored under `docmgr/ttmp`), where each document is a Markdown file with YAML frontmatter following the `Document` schema.

Both can be packaged, but the data models and targets differ.

### docmgr help docs: embed and export

The help docs use the same embed and load pattern as Glazed. `docmgr/pkg/doc/doc.go` declares an embedded filesystem and `AddDocToHelpSystem` calls `help.HelpSystem.LoadSectionsFromFS`. The downstream model and SQLite schema are the Glazed help system ones described above.

In addition, `docmgr/pkg/doc/embedded_docs.go` provides `ReadEmbeddedMarkdownDocs`, which returns a list of embedded Markdown files (`EmbeddedDoc{Name, Content}`) for use in export tooling. This is the hook that `docmgr/internal/workspace/sqlite_export.go` uses to package the help docs into a SQLite `README` table.

### docmgr ticket docs: frontmatter model and parsing

Ticket documents are Markdown files with YAML frontmatter that conforms to `docmgr/pkg/models.Document`. The frontmatter is parsed by `ReadDocumentWithFrontmatter` in `docmgr/internal/documents/frontmatter.go`, which uses YAML decoding and strict `---` delimiters.

#### Data format: docmgr document frontmatter

```yaml
---
Title: Skill creation, packaging, and doc export analysis
Ticket: MO-001-EXPORT-SKILLS
Status: active
Topics: [documentation, tools, docmgr, glaze]
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-13T09:16:54.702496782-05:00
WhatFor: ""
WhenToUse: ""
---
```

The `Document` struct includes additional fields like `RelatedFiles` (path and optional note) and `ExternalSources`, which can be used to link a document to source code or imported content. The docmgr parser also preprocesses YAML to quote risky scalars via `frontmatter.PreprocessYAML` to reduce parsing failures.

### docmgr workspace index schema

docmgr creates an in-memory SQLite index for fast search and filtering. The schema is defined in `docmgr/internal/workspace/sqlite_schema.go` and includes tables for documents, topics, owners, and related files. This schema is the basis for export and can be used to package docmgr documents into a portable SQLite file.

Key tables:

- `docs`: one row per markdown document with parsed frontmatter and optional body.
- `doc_topics`: many-to-many mapping of docs to topics.
- `doc_owners`: many-to-many mapping of docs to owners.
- `related_files`: normalized related file paths (canonical and fallback representations).

#### Data format: docmgr workspace schema (key fields)

```sql
CREATE TABLE docs (
  doc_id INTEGER PRIMARY KEY,
  path TEXT NOT NULL UNIQUE,
  ticket_id TEXT,
  doc_type TEXT,
  status TEXT,
  intent TEXT,
  title TEXT,
  last_updated TEXT,
  what_for TEXT,
  when_to_use TEXT,
  parse_ok INTEGER NOT NULL DEFAULT 1,
  parse_err TEXT,
  is_index INTEGER NOT NULL DEFAULT 0,
  is_archived_path INTEGER NOT NULL DEFAULT 0,
  is_scripts_path INTEGER NOT NULL DEFAULT 0,
  is_sources_path INTEGER NOT NULL DEFAULT 0,
  is_control_doc INTEGER NOT NULL DEFAULT 0,
  body TEXT
);

CREATE TABLE doc_topics (
  doc_id INTEGER NOT NULL,
  topic_lower TEXT NOT NULL,
  topic_original TEXT,
  PRIMARY KEY (doc_id, topic_lower)
);

CREATE TABLE related_files (
  rf_id INTEGER PRIMARY KEY,
  doc_id INTEGER NOT NULL,
  note TEXT,
  norm_canonical TEXT,
  norm_repo_rel TEXT,
  norm_docs_rel TEXT,
  norm_doc_rel TEXT,
  norm_abs TEXT,
  norm_clean TEXT,
  anchor TEXT,
  raw_path TEXT
);
```

This schema captures not only the frontmatter metadata, but also normalization of related file paths, which is critical for reliable `docmgr doc relate` searches.

### Exporting docmgr docs to SQLite

`docmgr/internal/workspace/sqlite_export.go` provides `ExportIndexToSQLiteFile`. This function creates a persistent SQLite file from the in-memory workspace index and injects help docs into a `README` table to make the export self-describing.

Export flow summary:

- Require a fully initialized workspace index (`InitIndex` must have been called).
- Verify output path and handle `--force` behavior.
- Create the `README` table if needed.
- Populate `README` with embedded docs from `ReadEmbeddedMarkdownDocs` and a synthetic `__about__.md` entry describing the export.
- Use `VACUUM INTO` to write a consistent single-file SQLite snapshot.

#### Pseudocode: docmgr export to SQLite

```text
ExportIndexToSQLiteFile(ctx, outPath, force):
  assert workspace index is initialized
  assert output parent directory exists
  if outPath exists and not force: error
  if outPath exists and force: remove

  ensureReadmeTable(db)
  populateReadmeTable(db)  // inserts __about__.md and embedded docs

  lit = sqliteQuoteStringLiteral(outPath)
  db.Exec("VACUUM INTO " + lit)
```

### Packaging strategy for docmgr ticket documents

Given the docmgr codebase, there are three practical packaging options:

1. **Export the workspace index to SQLite** via `ExportIndexToSQLiteFile`.
   - Produces a single SQLite file containing the docs metadata, optional body content, and a README table with embedded help docs.
   - Best for portable, queryable snapshots and programmatic consumption.

2. **Package the raw Markdown tree** under `docmgr/ttmp` into a zip or tar archive.
   - Preserves original files and frontmatter exactly.
   - Best for human inspection or for re-importing into another docmgr instance.

3. **Embed docs into a binary** using the help system infrastructure.
   - Applicable to docmgr help docs, not ticket docs.
   - Useful for CLI help packaging where the docs must be included at build time.

The first option is the most aligned with docmgr's internal architecture because it captures both metadata and normalized indexes. The second option is a faithful raw archive. The third option is specific to the help docs, not the ticket doc tree.

## Part D: External standard and client behavior (Agent Skills + Cursor)

The local skill tooling in `~/.codex/skills` aligns closely with the Agent Skills standard, but the official specification and client integrations introduce additional constraints and behavior expectations. This section captures those external expectations and highlights where they differ from the local validator so a packaging pipeline can target the strictest common denominator.

### Agent Skills official specification (agentskills.io/specification)

The official specification defines the normative SKILL.md frontmatter fields, constraints, and the expected folder layout. In addition to the `name` and `description` fields emphasized by local tooling, the standard recognizes fields that are not currently validated by `quick_validate.py` (notably `compatibility`). This is important if you intend to distribute skills to multiple clients that rely on the official spec.

Key requirements and constraints from the spec:

- **Directory structure**: A skill is a directory with `SKILL.md` at its root; optional folders include `scripts/`, `references/`, and `assets/`.
- **Frontmatter required fields**: `name` and `description`.
- **`name` constraints**:
  - Length 1-64 characters.
  - Lowercase alphanumeric characters and hyphens only.
  - Must not start or end with a hyphen, and must not contain consecutive hyphens.
  - **Must match the parent directory name** (this is explicit in the spec; the local validator does not enforce this).
- **`description` constraints**:
  - Length 1-1024 characters.
  - Should describe what the skill does and when to use it.
  - Should include keywords that help agents recognize relevance.
- **Optional fields**:
  - `license`: Short license identifier or reference to a bundled license file.
  - `compatibility`: Up to 500 characters describing environment requirements (intended product, packages, network access). This is part of the spec but **not allowed** by the local `quick_validate.py` rules.
  - `metadata`: Arbitrary string-to-string map for vendor-specific metadata.
  - `allowed-tools`: Space-delimited list of pre-approved tools; marked experimental.

#### Data format: spec-compliant SKILL.md frontmatter (with optional fields)

```yaml
---
name: pdf-processing
description: Extracts text and tables from PDF files, fills forms, and merges PDFs. Use when working with PDF documents or when the user mentions PDFs, forms, or document extraction.
license: Apache-2.0
compatibility: Requires pandoc and xelatex; expects local filesystem access.
metadata:
  author: example-org
  version: "1.0"
allowed-tools: Bash(git:*) Bash(jq:*) Read
---
```

#### Progressive disclosure and file references

The spec reiterates a progressive disclosure model similar to the local skill-creator guidance but with explicit limits:

- Metadata (`name`, `description`) is loaded at startup for all skills.
- The `SKILL.md` body is loaded when the skill is activated (recommended < 5000 tokens and < 500 lines).
- `scripts/`, `references/`, and `assets/` are loaded on demand.

For file references inside `SKILL.md`, the spec mandates:

- Use **relative paths** from the skill root (e.g., `references/REFERENCE.md`).
- Keep reference chains **one level deep** from `SKILL.md` (avoid long nested chains).

#### Validation tooling

The spec mentions `skills-ref validate ./my-skill` as a reference validator. This implies that if you intend to distribute skills to a broader ecosystem, you should validate with **both**:

- Local validator: `/home/manuel/.codex/skills/.system/skill-creator/scripts/quick_validate.py`
- Spec validator: `skills-ref validate ./my-skill`

Because the local validator does not permit `compatibility`, a packaging pipeline may need to support a "spec-compliant mode" that allows `compatibility` while still enforcing the core naming constraints.

### Cursor integration behavior (cursor.com/docs/context/skills)

Cursor's Agent Skills integration adds client-specific behaviors that affect distribution and discovery:

- **Automatic discovery on startup**: Cursor scans skill directories and makes them available to Agent; the agent decides relevance based on context.
- **Manual invocation**: Skills can be invoked by typing `/` in Agent chat and selecting the skill.
- **Discovery locations** (scope-aware):
  - `.cursor/skills/` (project-level)
  - `.claude/skills/` (project-level, Claude compatibility)
  - `~/.cursor/skills/` (user-level, global)
  - `~/.claude/skills/` (user-level, global)
- **Frontmatter fields in Cursor docs**:
  - `description` is required and is displayed in menus.
  - `name` is optional; if omitted, the parent folder name is used.

This creates a compatibility nuance:

- The official spec requires `name`, but Cursor will function if `name` is omitted.
- For maximum compatibility, **include `name` and ensure it matches the folder name**, satisfying the spec and still aligning with Cursor's fallback behavior.

Cursor also surfaces skills in Settings -> Rules (Agent Decides section) and supports installing skills from GitHub repositories via a "Remote Rule (Github)" workflow. This implies that packaging for Cursor should treat a GitHub repo path as the canonical distribution unit, consistent with the local `skill-installer` guidance.

## Summary of packaging paths and data contracts

- Skills are defined by `SKILL.md` frontmatter and optional resources, validated by `quick_validate.py`, and packaged into `.skill` zip files by `package_skill.py`.
- Glazed help docs are Markdown files with YAML frontmatter parsed by `LoadSectionFromMarkdown` and stored in an SQLite help store; they can be embedded into binaries or exported by using a file-backed store.
- docmgr help docs follow the same help system model as Glazed and can be embedded or exported via `ReadEmbeddedMarkdownDocs` into a `README` table in a workspace SQLite export.
- docmgr ticket docs are YAML-frontmatter Markdown files that map to the `Document` model and are indexed into a workspace SQLite schema; the canonical packaging flow is `ExportIndexToSQLiteFile`.
- The Agent Skills spec adds `compatibility` and stricter rules (name must match directory, 1-64 chars), plus optional `allowed-tools`; these are not enforced by the local validator today.
- Cursor discovers skills from `.cursor/skills/`, `.claude/skills/`, and user-level equivalents, requires `description`, and treats `name` as optional; packaging should still include `name` to remain spec-compliant.

## Suggested follow-up integrations (optional)

If a single unified packaging pipeline is desired, the minimal compatible flow is:

- Export Glazed help docs to SQLite using `store.New(dbPath)` plus `doc.AddDocToHelpSystem`.
- Export docmgr workspace docs to SQLite using `ExportIndexToSQLiteFile` (which already includes a README of embedded docs).
- Package skill directories into `.skill` files using `package_skill.py`.
- Bundle the resulting `.skill` files and the doc SQLite files into a single zip if a monolithic distribution artifact is needed.

This preserves the native data models and minimizes transformation loss.
