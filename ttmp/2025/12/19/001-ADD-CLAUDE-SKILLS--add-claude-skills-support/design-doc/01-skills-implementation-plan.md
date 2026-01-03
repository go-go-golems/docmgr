---
Title: 'Skills: Implementation Plan'
Ticket: 001-ADD-CLAUDE-SKILLS
Status: active
Topics:
    - features
    - skills
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T12:58:03.160824786-05:00
---

# Skills: Implementation Plan

## Executive Summary

This design adds a first-class **skills** feature to docmgr by treating skills as regular docmgr documents with a dedicated `DocType: skill`, plus two new optional frontmatter fields: `WhatFor` and `WhenToUse`. The CLI will expose `docmgr skill list` and `docmgr skill show`, where list supports filtering by ticket/topics and—critically—by **related file paths/directories** (`--file`, `--dir`) using the same query semantics as `docmgr doc search`.

The key implementation detail is that docmgr’s `workspace.QueryDocs()` currently returns `models.Document` instances **hydrated from the SQLite index**, not by re-reading the markdown file. Therefore, `WhatFor`/`WhenToUse` must be stored in the in-memory SQLite schema (and populated during ingest) so `skill list/show` can display them without any separate “discovery” pass.

## Problem Statement

We want docmgr to manage “skills” as structured documentation artifacts that can be listed, filtered, and inspected in a CLI-friendly way. In practice, skills should help a developer (or an LLM) quickly answer: “What does this skill do?” and “When should I use it?”, plus provide metadata-based discovery (topics) and code linkage (related paths).

Today, docmgr has strong building blocks—YAML frontmatter, `RelatedFiles`, workspace discovery, and an indexed query layer—but no dedicated skill verbs, no canonical place to store skill-specific preamble fields, and no “filter by code path” workflow tailored to skills.

## Proposed Solution

### 1) Represent skills as documents

Skills are markdown documents with YAML frontmatter and `DocType: skill`. They can live anywhere under the docs root (`ttmp/`), but conventionally:

- Workspace-level skills: `ttmp/skills/*.md`
- Ticket-local skills: `ttmp/YYYY/MM/DD/<TICKET>--<slug>/skills/*.md`

The content contract is:

```yaml
---
Title: "Skill: Foo"
Ticket: 001-ADD-CLAUDE-SKILLS   # optional for workspace-level skills
DocType: skill
Topics: [backend, tooling]
WhatFor: "Short paragraph: what this skill is for"
WhenToUse: "Short paragraph: when this skill should be used"
RelatedFiles:
  - Path: backend/foo/bar.go
    Note: "Why this code matters for this skill"
---
```

### 2) Add `docmgr skill list`

`docmgr skill list` lists skills and prints:

- `WhatFor`
- `WhenToUse`
- `Topics`
- `RelatedFiles` paths (and optionally notes)

It supports filtering by:

- `--ticket` (skill docs for a ticket)
- `--topics` (OR semantics, any topic matches)
- `--file` (skills referencing a specific related file path)
- `--dir` (skills referencing any related file within a directory)

### 3) Add `docmgr skill show <skill>`

`docmgr skill show <skill>` resolves a single skill and prints what it does (the preamble fields + metadata + full markdown body). Matching strategy (initial version):

- Exact title match OR case-insensitive contains match
- If ambiguous: show candidates and error (or show first + warn—TBD; see Open Questions)

### 4) No separate “skill discovery”

We will **not** implement a separate discovery system. Skills are indexed like any other markdown document in `ttmp/`, and list/show use `workspace.QueryDocs()` filtered by `DocType == "skill"`, plus `RelatedFile`/`RelatedDir` for path filtering.

## Design Decisions

### D1. Store `WhatFor`/`WhenToUse` in the SQLite index

`workspace.QueryDocs()` does not re-read markdown files; it reconstructs `models.Document` from indexed columns and then batch-hydrates topics + related files. Therefore, if we want `skill list` to show `WhatFor`/`WhenToUse` without doing an additional per-file read, those fields must be part of the `docs` table and populated by the index builder.

### D2. Reuse existing path filtering semantics (`--file`, `--dir`)

docmgr already supports reverse lookups via `DocFilters.RelatedFile` and `DocFilters.RelatedDir` (used by `docmgr doc search`). Skills need the same behavior: “show me skills relevant to code under `backend/chat/`”. Reusing the existing query layer keeps semantics consistent and reduces complexity.

### D3. Skills are “just documents” (DocType-guided)

Keeping skills as normal documents preserves the existing toolchain: frontmatter parsing, validation, indexing, and structured output. The only special handling is a couple of extra fields and a dedicated CLI surface.

## Alternatives Considered

### A1. Implement a separate “skill discovery” walker (rejected)

This would walk `ttmp/skills/` and `<ticket>/skills/` directories and parse skills directly. We rejected this because it duplicates the indexing system and would require separate filtering logic for `--file` / `--dir`. It also conflicts with docmgr’s direction of treating documents uniformly via `QueryDocs()`.

### A2. Keep `WhatFor`/`WhenToUse` out of the index and re-read files (rejected)

This would use `QueryDocs()` only to get candidate paths, then re-parse each file to extract `WhatFor`/`WhenToUse`. It’s feasible for small numbers of skills, but it creates inconsistent behavior (some fields indexed, others not) and complicates performance reasoning and future filtering/formatting.

### A3. Store the entire frontmatter blob in SQLite (deferred)

We could store the full YAML node or serialized frontmatter in SQLite. This would be more general, but it introduces a larger schema surface and complicates portability. For now, we only need two additional fields; adding columns is simpler.

## Implementation Plan

### 1. Data model: add skill fields to `models.Document`

Add two optional fields to `pkg/models/document.go`:

- `WhatFor string  \`yaml:"WhatFor" json:"whatFor"\``
- `WhenToUse string \`yaml:"WhenToUse" json:"whenToUse"\``

These fields are optional for all documents; only skills are expected to populate them.

- [ ] Add `WhatFor` and `WhenToUse` fields to `models.Document`
- [ ] Add/adjust any JSON/YAML tests for the model if needed

### 2. Workspace index: extend SQLite schema to store skill fields

Update `internal/workspace/sqlite_schema.go` to add two nullable TEXT columns to the `docs` table:

- `what_for TEXT`
- `when_to_use TEXT`

Also update any schema tests (`internal/workspace/sqlite_schema_test.go`) to validate these columns exist.

- [ ] Add `what_for` and `when_to_use` columns to the `docs` table DDL
- [ ] Update schema tests accordingly

### 3. Workspace ingest: populate new columns during indexing

Update `internal/workspace/index_builder.go` to insert the new fields when `parseOK == 1`:

- Extract from parsed `doc.WhatFor`, `doc.WhenToUse`
- Insert into `docs` table for each document

This ensures skills can be queried and displayed without re-reading files.

- [ ] Extend `INSERT INTO docs (...)` and args to include `what_for`, `when_to_use`
- [ ] Ensure parse-failed docs still insert with `NULL` for these fields

### 4. Query layer: hydrate new fields in `workspace.QueryDocs()`

Update `internal/workspace/query_docs_sql.go` and `internal/workspace/query_docs.go`:

- Include the new columns in SELECT
- Scan them into `sql.NullString`
- Set `doc.WhatFor` and `doc.WhenToUse` on the returned `*models.Document`

- [ ] Extend compiled SQL SELECT to include `d.what_for, d.when_to_use`
- [ ] Extend `rows.Scan(...)` and document hydration to set fields
- [ ] Add/adjust query tests if they assert column order/shape

### 5. CLI surface: add `docmgr skill` command tree

Add a new command group under `cmd/docmgr/cmds/skill/` and attach it in `cmd/docmgr/cmds/root.go`.

- [ ] Create `cmd/docmgr/cmds/skill/skill.go` with `Attach()` wiring
- [ ] Register `skill.Attach(rootCmd)` in `cmd/docmgr/cmds/root.go`

### 6. Implement `docmgr skill list` (with path filtering)

Implement as a dual-mode command in `pkg/commands/skill_list.go`:

**Flags:**
- `--root` (consistent with other verbs)
- `--ticket`
- `--topics` (string list)
- `--file` (string; maps to `DocFilters.RelatedFile`)
- `--dir` (string; maps to `DocFilters.RelatedDir`)

**Query shape:**

```go
res, err := ws.QueryDocs(ctx, workspace.DocQuery{
  Scope: scopeFromTicketFlag(settings.Ticket),
  Filters: workspace.DocFilters{
    DocType: "skill",
    Ticket: settings.Ticket,
    TopicsAny: settings.Topics,
    RelatedFile: maybeOne(settings.File),
    RelatedDir: maybeOne(settings.Dir),
  },
  Options: workspace.DocQueryOptions{ /* include control docs? likely false */ },
})
```

**Output columns (structured mode):**
- `skill` (Title)
- `what_for` (WhatFor)
- `when_to_use` (WhenToUse)
- `topics`
- `related_paths` (paths extracted from RelatedFiles)
- `path` (doc path)

- [ ] Implement `pkg/commands/skill_list.go` (GlazeCommand + BareCommand)
- [ ] Wire cobra command in `cmd/docmgr/cmds/skill/list.go` with completions for `--ticket`, `--topics`, `--root`
- [ ] Ensure `--file`/`--dir` behave the same as `docmgr doc search`

### 7. Implement `docmgr skill show <skill>`

Implement as a command in `pkg/commands/skill_show.go`. It should:

- Query `DocType == "skill"` (and optionally scope to ticket if flag added later)
- Match the requested skill string against Title (case-insensitive)
- Use `IncludeBody: true` to print markdown body as part of output

- [ ] Implement `pkg/commands/skill_show.go`
- [ ] Wire cobra command in `cmd/docmgr/cmds/skill/show.go`
- [ ] Decide ambiguity behavior (error with candidates vs first match + warning)

### 8. Vocabulary + templates

Add `skill` to `docTypes` vocabulary and (optionally) add a `_templates` entry for a skill scaffold.

- [ ] Add `skill` to `ttmp/vocabulary.yaml` docTypes
- [ ] Add a `_templates` skill template including `WhatFor` and `WhenToUse` (optional but recommended)

### 9. Tests + scenarios

Add tests to ensure:

- Index schema includes new columns
- QueryDocs hydrates `WhatFor`/`WhenToUse`
- Skill list filtering by `--file` and `--dir` works (at least one integration test)

- [ ] Add/adjust unit tests for schema/query hydration
- [ ] Add a minimal scenario test under `test-scenarios/` for `skill list --file/--dir`

### 10. Documentation

Update docs to mention the new skill verbs and the filter flags.

- [ ] Update `pkg/doc/docmgr-cli-guide.md` to include `docmgr skill list/show` usage
- [ ] Add examples for `--file` and `--dir` filtering

## Open Questions

1. **Skill matching rules**: Should `skill show <skill>` require exact match by title, support slug matching, or both?
2. **Ambiguity handling**: If multiple skills match, should we error and print candidates, or choose the best match?
3. **Ticket-local vs global skills**: Do we want a `--scope repo|ticket` flag, or is `--ticket` enough?
4. **Validation**: Should `WhatFor` and `WhenToUse` be required when `DocType == skill`, or remain optional?

## References

- Analysis: `ttmp/2025/12/19/001-ADD-CLAUDE-SKILLS--add-claude-skills-support/analysis/01-skills-feature-analysis.md`
- Workspace query hydration: `internal/workspace/query_docs.go`
- Workspace index ingest: `internal/workspace/index_builder.go`
- Workspace schema: `internal/workspace/sqlite_schema.go`
- Related file filtering semantics: `pkg/commands/search.go`
