---
Title: 'Design: skill.yaml discovery, skill verbs, and Agent Skills import/export'
Ticket: MO-001-EXPORT-SKILLS
Status: active
Topics:
    - documentation
    - tools
    - docmgr
    - glaze
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/SKILL.md
      Note: Agent Skills format for export
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/package_skill.py
      Note: Export packaging implementation
    - Path: pkg/commands/skill_list.go
      Note: Verb behavior changes required
    - Path: pkg/commands/skill_show.go
      Note: Verb behavior changes required
    - Path: pkg/doc/using-skills.md
      Note: User-facing skill UX
ExternalSources:
    - https://agentskills.io/specification
Summary: ""
LastUpdated: 2026-01-13T15:44:11-05:00
WhatFor: ""
WhenToUse: ""
---


# Design: skill.yaml discovery, skill verbs, and Agent Skills import/export

## Executive Summary

We will keep the `docmgr skill` verbs but change their backing data model to a new `skill.yaml` plan format. Plans are discovered in `ttmp/skills/` (workspace) and `ttmp/YYYY/MM/DD/<TICKET>--<slug>/skills/` (ticket) when `--ticket` is provided. Each plan can reference explicit files and help topics captured from Glazed binaries via `$binary help <topic>`. New `docmgr skill export` and `docmgr skill import` commands will bridge plans to the Agent Skills standard (`.skill` packages). The existing DocType skill documents remain as workflow docs but are no longer the source of `skill list/show` results.

## Problem Statement

Docmgr skills are currently DocType skill documents. The new requirement is to make skill verbs operate on a `skill.yaml` plan format that can reference files and capture help topics from a Glazed binary. We also need to export and import the standard Agent Skills packages, while keeping docmgr's core workflows stable and avoiding accidental execution of binaries during listing operations.

## Proposed Solution

Introduce a `skill.yaml` plan schema and a discovery layer that scans `ttmp/skills/` and ticket `skills/` folders. The plan becomes the single source of truth for `skill list` and `skill show`. We add two new commands:

- `docmgr skill export`: resolve a plan into a skill directory and package it as `.skill`.
- `docmgr skill import`: ingest a `.skill` artifact or directory and generate a plan plus reference files.

The `binary-help` source type runs `$binary help <topic>` and writes the output as a reference file during export or explicit resolve actions. Listing operations never execute binaries.

## Design Decisions

1. **Plans over DocType skill docs**
   - Rationale: explicit and reproducible source selection, with a clear path to Agent Skills packaging.

2. **Discovery paths are fixed**
   - Workspace plans: `ttmp/skills/`.
   - Ticket plans: `ttmp/YYYY/MM/DD/<TICKET>--<slug>/skills/` when `--ticket` is set.
   - Rationale: simple mental model and predictable scoping.

3. **Binary execution only during export/resolve**
   - Rationale: listing should be safe, fast, and deterministic.

4. **Agent Skills compliance enforced at export time**
   - Rationale: keep plan editing flexible, enforce constraints only when packaging.

## skill.yaml schema (v1)

```yaml
skill:
  name: glaze-help
  title: Glaze Help System
  description: Help topics and references for Glazed. Use when working with Glazed CLI help.
  what_for: Provide help output and docs for Glazed commands and help topics.
  when_to_use: Use when referencing Glazed CLI help or help topics.
  topics: [glaze, documentation, help]
  license: Proprietary
  compatibility: Requires glaze binary in PATH.

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

Schema notes:

- `skill.name` must be Agent Skills-compliant; `skill_dir_name` defaults to `skill.name`.
- `title` is for display in `skill list/show` and can differ from name.
- `what_for` and `when_to_use` are required for listing parity with current verbs.
- `sources` defines the ordered content collection.

## Discovery and indexing behavior

### Discovery

- If `--ticket` is not provided:
  - Only scan `ttmp/skills/`.
- If `--ticket` is provided:
  - Scan `ttmp/skills/` (optional) and the ticket `skills/` folder.
  - The ticket folder is derived from the ticket index (same resolution used by `docmgr ticket`).

### Matching and filtering

`skill list` filters:

- `--topics`: match `skill.topics`.
- `--file` and `--dir`: match explicit `file` sources only (not `binary-help`).
- Ticket status filtering mirrors current behavior: when `--ticket` is absent, hide plans belonging to non-active tickets (use ticket index status).

`skill show` matching:

- Title match: `skill.title` and `skill.name`.
- Slug match: `skill.name` and plan folder name.
- Path match: full or relative plan path.

## Verb behavior changes

### skill list

- Replaces DocType skill query with plan discovery.
- Outputs: skill title, what_for, when_to_use, topics, referenced files, plan path, load command.
- Does not run any binary commands.

### skill show

- Displays plan metadata and sources.
- Optional `--resolve` flag:
  - Generates a resolved view including file contents and captured help output.
  - This can either be printed or written to a temporary directory.

### skill export

- Inputs: plan path or plan name.
- Actions:
  - Resolve sources (read files, run binary help).
  - Create skill directory with `SKILL.md` + `references/`.
  - Validate and package to `.skill`.

### skill import

- Inputs: `.skill` file or directory.
- Actions:
  - Extract skill into a temp directory if needed.
  - Parse `SKILL.md` and references.
  - Generate `skill.yaml` plan and save references under a plan folder.

## Export and import details

### Export packaging

- Use existing packager script to create `.skill`.
- Ensure `skill_dir_name` matches `skill.name`.
- Produce `SKILL.md` with:
  - Required frontmatter (`name`, `description`).
  - Body: short overview + reference index.

### Import mapping

- `SKILL.md` frontmatter -> `skill.name`, `skill.description`.
- `SKILL.md` body -> optional `skill_md.intro` in plan.
- `references/` files -> `file` sources in plan.

## Design Decisions (tradeoffs)

- **Keep DocType skill docs**: We do not delete or repurpose them; they remain workflow docs.
- **Plan-only verbs**: Simplifies UX but may confuse users who expect DocType skills. Mitigate via help text updates and migration notes.
- **Explicit ticket discovery**: Avoids scanning all tickets by default, preserving current performance and scoping behavior.

## Alternatives Considered

- **Overload DocType skill docs**: Adds frontmatter fields for plan sources. Rejected due to frontmatter rewrite constraints.
- **Replace verbs with new commands**: Would reduce confusion but breaks established UX and scripts.
- **Always resolve binaries during list**: Too slow and unsafe.

## Implementation Plan

1. Add YAML plan parser and schema validation.
2. Implement plan discovery (workspace + ticket paths).
3. Update `skill list` and `skill show` to use plan metadata.
4. Add `binary-help` source resolver with safe execution.
5. Implement `skill export` and `skill import` commands.
6. Update docs: `using-skills.md`, `how-to-write-skills.md`, and guidelines to describe plan-based skills.
7. Add tests for discovery, filtering, export, import, and binary help capture.

## Open Questions

- Should plan discovery include both workspace and ticket plans when `--ticket` is set, or only ticket plans?
- Do we need a caching policy for binary help output to avoid repeated execution?
- How should we handle plan versioning and help output drift?
- Should `skill show --resolve` write artifacts to disk or stream to stdout?
