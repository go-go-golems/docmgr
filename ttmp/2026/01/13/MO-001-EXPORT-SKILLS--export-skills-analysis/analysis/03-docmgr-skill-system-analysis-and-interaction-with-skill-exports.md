---
Title: docmgr skill system analysis and interaction with skill exports
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
    - Path: internal/templates/embedded/_templates/skill.md
      Note: Skill template scaffold
    - Path: internal/workspace/sqlite_schema.go
      Note: Skill-related fields in workspace schema
    - Path: pkg/commands/skill_list.go
      Note: Skill discovery and filtering logic
    - Path: pkg/commands/skill_show.go
      Note: Skill matching and rendering behavior
    - Path: pkg/doc/how-to-write-skills.md
      Note: DocType skill frontmatter contract
    - Path: pkg/doc/using-skills.md
      Note: Skill list/show user workflow
    - Path: pkg/models/document.go
      Note: WhatFor/WhenToUse fields
    - Path: ttmp/_guidelines/skill.md
      Note: Skill document guidelines
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-13T09:58:26-05:00
WhatFor: ""
WhenToUse: ""
---


# Analysis

## Goal and scope

This analysis describes docmgr's current skill system, how it works, and how it would interact with a new export pipeline that packages Glazed/docmgr docs into Agent Skills. It focuses on the docmgr CLI behavior, frontmatter contract, indexing, and current skill discovery rules, then maps those concepts to potential export flows.

## What the docmgr "skill" system is

In docmgr, a "skill" is not an Agent Skills package. It is a docmgr document with `DocType: skill` that defines a disciplined workflow for LLMs and humans. The goal is to encode process guidance (TDD, systematic debugging, brainstorming) as structured markdown that is discoverable and loadable via `docmgr skill list` and `docmgr skill show`.

The authoritative contract is documented in `docmgr/pkg/doc/how-to-write-skills.md` and `docmgr/ttmp/_guidelines/skill.md`. The required frontmatter fields include `Title`, `DocType: skill`, `Topics`, `WhatFor`, and `WhenToUse`. The body is a structured workflow that typically includes an overview, iron law, process steps, and verification checklist.

Key properties:

- **DocType-based discovery**: skills are identified by `DocType: skill` in frontmatter, not by directory name.
- **Workflow emphasis**: `WhatFor` and `WhenToUse` are explicit fields designed to help LLMs decide when to apply the skill.
- **Disciplined structure**: a recommended template is scaffolded by `docmgr doc add --doc-type skill` (see `docmgr/internal/templates/embedded/_templates/skill.md`).

## How docmgr skill discovery works

### Skill list

`docmgr skill list` is implemented by `docmgr/pkg/commands/skill_list.go`. It:

- Discovers the workspace and builds the in-memory index (`workspace.DiscoverWorkspace`, `InitIndex`).
- Queries for docs with `DocType: skill` and optional filters for topics, related files, and ticket scope.
- Filters out skills from non-active tickets by default (only status `active` is included unless `--ticket` is specified).
- Emits a load command for each skill, which includes the title or path depending on ambiguity.

The output includes `skill`, `what_for`, `when_to_use`, `topics`, `related_paths`, `path`, and a `load_command`. This is effectively a catalog of workflow skills for the current workspace.

### Skill show

`docmgr skill show` is implemented by `docmgr/pkg/commands/skill_show.go`. It:

- Loads doc bodies (`IncludeBody: true`) and performs a multi-strategy match against title, slug, and path.
- Applies default filtering to hide skills from completed/archived tickets unless `--ticket` is provided.
- Prints the skill's body (full markdown), along with WhatFor/WhenToUse and ticket context if applicable.

This means the docmgr skill system is a doc discovery and rendering layer, not a packaging format. It surfaces structured workflow documents rather than producing an Agent Skills package.

## Storage conventions

The docmgr docs emphasize conventions:

- **Workspace-level skills** typically live under `ttmp/skills/`.
- **Ticket-scoped skills** are created under `ttmp/YYYY/MM/DD/<TICKET>--<slug>/skill/` (singular) by `docmgr doc add --doc-type skill`.
- The directory name is a convention only; discovery uses frontmatter `DocType: skill`.

This matters for exports because a skill exporter should rely on metadata (DocType, Topics, WhatFor, WhenToUse) rather than directory location.

## Data model alignment with Agent Skills

Docmgr skills and Agent Skills share a similar purpose (teaching an agent how to work) but differ in data contract:

- **docmgr skill document**:
  - Frontmatter: `Title`, `DocType: skill`, `WhatFor`, `WhenToUse`, `Topics`, `RelatedFiles`, etc.
  - Body: workflow content.

- **Agent Skills package**:
  - Frontmatter: `name`, `description` (plus optional `license`, `compatibility`, etc).
  - Body: instructions.
  - Optional `references/` for large content.

Mapping implications:

- `Title` -> `name` (normalized to the Agent Skills naming constraints).
- `WhatFor` + `WhenToUse` -> `description` (condensed to <= 1024 chars).
- doc body -> `SKILL.md` body (or `references/` if body is long).
- `Topics` -> potential keywords in description or a skill-local index section.

The docmgr skill system is therefore a plausible source for generating Agent Skills, but it is not directly compatible without transformation.

## Interaction with a new export pipeline

### Reuse opportunities

- **Selection engine**: docmgr's workspace index already supports filtering by DocType, Topics, and RelatedFiles. An export pipeline can query `DocType: skill` (or other doc types) to select source documents.
- **Discovery UX**: `docmgr skill list` already provides canonical names and load commands; this can be reused as the selection UI for export.
- **Workflow content**: docmgr skill bodies are structured for agent consumption and may map well to Agent Skills SKILL.md bodies.

### Gaps and friction

- **Frontmatter mismatch**: docmgr frontmatter does not include `name` and `description` fields required by Agent Skills; these must be derived.
- **Unknown frontmatter fields**: docmgr rewrites frontmatter via `models.Document`. Any new `Skill:` block for export would require schema changes or a sidecar file, otherwise it will be dropped on rewrite.
- **Different semantics**: docmgr skills are about "how to work"; the new exporter also needs to package "what exists" docs (help pages). Mixing these may require separate skill types or distinct packaging rules.

## Implications for the requested skill.yaml export plans

The requested `skills/` folder with `skill.yaml` metadata is conceptually separate from docmgr's DocType skill system. A minimal integration path is:

- Treat `skill.yaml` as an export plan, not a docmgr skill.
- Allow the plan to reference docmgr skills (DocType skill) as inputs, but do not conflate the two.
- Use docmgr's index to resolve file paths and metadata, then generate Agent Skills artifacts via the plan.

If deeper integration is desired, docmgr could be extended with a new doc type (for example `skill-plan`) that points to or embeds `skill.yaml` and is indexed like other docs.

## Summary

- docmgr skills are structured workflow documents (`DocType: skill`) that are discovered and loaded via `docmgr skill list/show`.
- The system relies on `WhatFor` and `WhenToUse` for discovery, and filters by ticket status to reduce noise.
- This skill system is not the Agent Skills packaging format but can serve as a rich input source for conversion.
- A `skill.yaml` export plan should be treated as a separate layer that can optionally reference docmgr skills as inputs.
