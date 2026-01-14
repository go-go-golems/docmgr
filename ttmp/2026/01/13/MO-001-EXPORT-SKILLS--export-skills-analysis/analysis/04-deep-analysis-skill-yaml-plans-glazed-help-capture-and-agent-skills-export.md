---
Title: 'Deep analysis: skill.yaml plans, glazed help capture, and Agent Skills export'
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
      Note: Agent Skills structure and packaging constraints
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/package_skill.py
      Note: Packaging to .skill
    - Path: internal/templates/embedded/_templates/skill.md
      Note: Skill template scaffold
    - Path: internal/workspace/sqlite_schema.go
      Note: Workspace schema with skill-related fields
    - Path: pkg/commands/skill_list.go
      Note: Current skill list implementation to be replaced
    - Path: pkg/commands/skill_show.go
      Note: Current skill show implementation to be replaced
    - Path: pkg/doc/how-to-write-skills.md
      Note: DocType skill contract and conventions
    - Path: pkg/doc/using-skills.md
      Note: User-facing skill UX that must be updated
    - Path: ttmp/_guidelines/skill.md
      Note: Guidelines for DocType skill documents
ExternalSources:
    - https://agentskills.io/specification
Summary: ""
LastUpdated: 2026-01-13T15:44:11-05:00
WhatFor: ""
WhenToUse: ""
---


# Analysis

## Goal and scope

This analysis describes the work required to keep the existing docmgr skill verbs (`skill list`, `skill show`) but switch their backing store from DocType skill documents to a new skill.yaml plan format. The new format should support explicit file references and a binary help capture source that runs `$binary help <topic>` to populate reference content. It must also add import and export commands to produce standards-compliant Agent Skills artifacts.

The focus is on requirements, data contracts, indexing needs, CLI behavior, and the steps to implement the change without breaking current docmgr workflows.

## Current docmgr skill system (baseline)

Docmgr "skills" today are documents with `DocType: skill`. The system is defined by:

- `docmgr/pkg/commands/skill_list.go` and `docmgr/pkg/commands/skill_show.go`.
- `docmgr/pkg/doc/how-to-write-skills.md` and `docmgr/ttmp/_guidelines/skill.md`.
- The document frontmatter model in `docmgr/pkg/models/document.go` (fields `WhatFor`, `WhenToUse`).

The verbs operate on the workspace index (SQLite) by querying `DocType: skill`. The listing output includes `WhatFor` and `WhenToUse` plus topics and related files; `skill show` prints the full markdown body for the chosen doc and filters by ticket status by default.

Key behavior today:

- Discovery is DocType-driven, not directory-driven.
- Skills can be workspace-level (`ttmp/skills/`) or ticket-scoped (`.../skill/`), but the path is a convention only.
- The index is built from markdown frontmatter and used for filtering and lookup.

## New requirement: keep verbs but swap backing format

You want to keep the verbs (list/show), but store skills as `skill.yaml` plans rather than DocType skill docs. The plans should live in:

- `ttmp/skills/` (workspace library).
- `ttmp/YYYY/MM/DD/<TICKET>--<slug>/skills/` when `--ticket` is used.

This introduces a fundamental change: the source of truth becomes configuration files, not docmgr docs. The verbs must therefore be reimplemented to parse `skill.yaml`, collect metadata, and optionally render a generated skill view.

## skill.yaml capabilities

The plan format must support at least two kinds of content sources:

1. Explicit file references (include file content in the skill output).
2. Help topics from a Glazed binary (run `$binary help <topic>` and capture stdout).

A plan can also include metadata needed by the verbs:

- Display name, description, what_for, when_to_use, topics.
- Ticket scoping and visibility rules (optional, but needed if the verbs should respect current active-ticket filtering).

### Candidate skill.yaml schema (draft)

```yaml
skill:
  name: glaze-help
  title: Glaze Help System
  description: Help topics for Glazed commands and help system usage.
  what_for: Provide reference help output and documentation for Glazed.
  when_to_use: Use when working with Glazed CLI help or help topic documentation.
  topics: [glaze, documentation, help]

sources:
  - type: file
    path: glazed/pkg/doc/topics/01-help-system.md
    strip-frontmatter: true
    output: references/glaze-help-system.md

  - type: binary-help
    binary: glaze
    topic: writing-help-entries
    output: references/glaze-help-writing-help-entries.md
    wrap: markdown

output:
  skill_dir_name: glaze-help
  skill_md:
    include_index: true
    index_title: Included references
```

Notes:

- `name` should be the Agent Skills-compatible slug (lowercase, hyphenated).
- `title` is a human-friendly label for `skill show` output.
- `what_for` and `when_to_use` can feed `skill list` output.
- `sources` is ordered, and outputs should be deterministic.

## Impact on docmgr verbs

### skill list

Required changes:

- Replace `DocType: skill` queries with filesystem discovery of `skill.yaml`.
- Add a new index layer or in-memory scanning of plan metadata.
- Preserve filters by `--ticket`, `--topics`, `--file`, `--dir`:
  - `--topics` matches `skill.topics`.
  - `--file` and `--dir` should match explicit `file` sources in the plan.
  - For `binary-help` sources, `--file` and `--dir` may not apply.

Data needed per plan:

- display title (prefer `skill.title`, fallback to `skill.name`).
- what_for, when_to_use for listing output.
- topics array for filtering.
- plan path (for load command display).

`skill list` must be updated to scan both `ttmp/skills/` (workspace) and `.../skills/` inside a ticket when `--ticket` is provided.

### skill show

`skill show` should load the plan and render a "resolved view":

- Show the skill metadata (title, what_for, when_to_use).
- Optionally show the rendered plan contents (the sources with file content and captured help output). This could be controlled by flags such as `--render` or `--resolve`.

Matching logic should remain similar to current behavior (title, slug, path). It should also support passing the plan path directly.

Ticket filtering should follow the same default rules: when no `--ticket` is set, hide plans belonging to non-active tickets. That implies plans must be tied to a ticket (from path) and the ticket index should be loaded to evaluate status.

## Binary help capture requirements

The `binary-help` source introduces new concerns:

- **Execution**: run `$binary help <topic>` and capture stdout.
- **Environment**: the binary must be on PATH (or referenced by absolute path in the plan).
- **Stability**: help output may change across versions. Captured output should be saved with optional version metadata.
- **Security**: running a binary may be unsafe in some contexts; plan execution should be explicit (not automatic in `skill list`).

Recommended approach:

- `skill show` should not execute help capture by default. Provide a `--resolve` flag to build a generated view and optionally cache output.
- `skill export` should execute help capture and store results in references.

## Export to Agent Skills (new command)

Add a `docmgr skill export` command that:

1. Loads a `skill.yaml` plan.
2. Resolves sources:
   - reads files into `references/`.
   - captures binary help output into `references/`.
3. Generates `SKILL.md`:
   - frontmatter uses `skill.name` and `skill.description`.
   - body includes a short overview and an index of references.
4. Packages as a `.skill` file using the existing skill packager (`package_skill.py`).

Key validation:

- `skill.name` must be Agent Skills-compliant.
- `skill.description` must be <= 1024 chars.
- `skill_dir_name` must match `skill.name` to satisfy spec.

### Example export flow

```text
# docmgr skill export <plan> --out dist
resolve sources
write skill-dir/SKILL.md
write skill-dir/references/*.md
run package_skill.py skill-dir dist
```

## Import from Agent Skills (new command)

Add a `docmgr skill import` command that:

- Accepts a `.skill` file or a skill directory.
- Parses `SKILL.md` frontmatter and body.
- Creates a `skill.yaml` plan with:
  - `skill.name`, `skill.title`, `skill.description` derived from `SKILL.md`.
  - A `file` source for each reference file (or optionally embed the SKILL.md body itself).
- Writes the plan to `ttmp/skills/<name>/skill.yaml` or a ticket `skills/` folder if `--ticket` is provided.

This gives docmgr users a way to ingest skills from the standard ecosystem without rewriting the content manually.

## Discovery and indexing changes

A new discovery layer is required because the workspace index currently indexes markdown documents, not YAML plans.

Options:

1. **Filesystem scan at runtime**: For each `skill list` call, scan `ttmp/skills/` and optional ticket skill folders, parse YAML, and filter in-memory.
2. **Extend workspace index**: Add a new table for skill plans (for example `skill_plans`) and index them on workspace init.

Given docmgr's current architecture, the pragmatic path is filesystem scan initially, then optional indexing if performance becomes a problem.

Potential index fields for `skill_plans`:

- plan_path
- skill_name
- title
- description
- what_for
- when_to_use
- topics (normalized)
- ticket_id (from path)
- sources (for file/dir filtering)

## Interactions with docmgr skills

DocType skill documents can still exist as workflow guides. The new plan format does not replace them; it adds a second concept:

- **DocType skill**: an instructional workflow doc (docmgr UX).
- **skill.yaml plan**: a packaging recipe for Agent Skills.

The verbs should be scoped to plans only, per your requirement, but docmgr can optionally expose a `docmgr skill doc` command if both need to coexist. The analysis assumes the verbs now operate on plan files exclusively.

## Risks and edge cases

- **Ambiguity**: A skill title may collide across plans. Matching should include plan path disambiguation like today.
- **Help topics not found**: `$binary help <topic>` failures must be captured and surfaced in export logs.
- **Non-deterministic output**: Help output may change; consider embedding version info or storing the binary path + version.
- **Large output**: Help output could be large; put all captured content in `references/`, not `SKILL.md`.
- **Security**: Running binaries as part of export should be explicit and logged; avoid running during `skill list`.

## Work required (summary)

- Define the `skill.yaml` schema and validation rules.
- Implement plan discovery in `ttmp/skills/` and ticket `skills/` folders.
- Rework `skill list/show` to operate on plans instead of DocType skill docs.
- Add `skill export` and `skill import` commands.
- Implement help capture runner and output normalization for `binary-help` sources.
- Add tests for plan parsing, discovery, and export packaging.
