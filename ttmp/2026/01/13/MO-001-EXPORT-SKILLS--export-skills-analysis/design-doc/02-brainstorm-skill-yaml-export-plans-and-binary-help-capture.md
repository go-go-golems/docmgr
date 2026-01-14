---
Title: 'Brainstorm: skill.yaml export plans and binary help capture'
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
      Note: Skill package structure guidance
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/package_skill.py
      Note: Packaging mechanics
    - Path: ttmp/2026/01/13/MO-001-EXPORT-SKILLS--export-skills-analysis/analysis/03-docmgr-skill-system-analysis-and-interaction-with-skill-exports.md
      Note: Analysis of current docmgr skill system
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-13T09:59:10-05:00
WhatFor: ""
WhenToUse: ""
---


# Brainstorm: skill.yaml export plans and binary help capture

## Executive Summary

This design brainstorm proposes a `skills/` folder containing `skill.yaml` export plans that build Agent Skills artifacts from existing docs and CLI help output. Each plan declares skill metadata plus a set of sources, including explicit files and help topics from a Glazed-based binary (captured via `$binary help <topic>`). The exporter assembles the content into a skill directory (SKILL.md + references/) and packages it using the existing skill packager. The approach can reuse docmgr's skill documents as inputs without replacing docmgr's skill discovery system.

## Problem Statement

We need a reproducible way to turn curated documentation and runtime help output into Agent Skills packages. The solution must:

- Allow explicit selection of source docs and help topics.
- Capture the current help output of a Glazed binary and store it as references.
- Remain compatible with Agent Skills validation rules.
- Avoid breaking docmgr's existing skill system (DocType: skill).

## Proposed Solution

Introduce a `skills/` directory containing `skill.yaml` export plans. Each plan declares:

- Skill metadata (`name`, `description`, optional `license`, `compatibility`).
- Inputs (file paths, docmgr docs by query, and binary help topics).
- Output mapping (SKILL.md index, references/ filenames, and ordering).

The exporter reads the plan, resolves inputs, captures help output from the binary, and writes a ready-to-package skill directory. It then invokes the existing packager to generate a `.skill` artifact.

### Example skill.yaml

```yaml
skill:
  name: glaze-help
  description: Help topics and reference docs for the Glazed CLI. Use when working with Glazed commands or help system topics.
  license: Proprietary

sources:
  - type: binary-help
    binary: glaze
    topics:
      - help-system
      - writing-help-entries
    output:
      dir: references
      filename-template: "glaze-help-{topic}.md"
      wrap-in-codeblock: false

  - type: file
    path: glazed/pkg/doc/topics/01-help-system.md
    strip-frontmatter: true

output:
  skill-md:
    intro: "This skill collects Glazed help system documentation and CLI help output."
    index:
      - label: Help System
        ref: references/glaze-help-help-system.md
      - label: Writing Help Entries
        ref: references/glaze-help-writing-help-entries.md
```

## Design Decisions

1. **Use `skill.yaml` as the primary export plan**
   - Centralized and reviewable configuration.
   - Avoids relying on doc frontmatter that docmgr might rewrite.

2. **Support a `binary-help` source type**
   - Explicitly capture the output of `$binary help <topic>`.
   - Output stored as markdown in `references/` (or optionally wrapped in code blocks).

3. **Keep docmgr's DocType skill system intact**
   - Docmgr skills remain workflow documents.
   - Exporter can optionally include docmgr skill docs as source files when desired.

4. **Stable filenames from templates**
   - Avoid collisions and allow deterministic packaging.

## Interaction with docmgr skill system

- docmgr skills are already curated workflow docs. They can be included as `file` sources in `skill.yaml`, or selected via a docmgr query if the exporter integrates with docmgr's workspace index.
- The exporter should not replace `docmgr skill list/show`; those commands are for interactive workflow guidance.
- If a tighter integration is needed, a new doc type (for example `skill-plan`) could be indexed, but that is optional and should not block initial export functionality.

## Alternatives Considered

- **Replace docmgr skills with Agent Skills**: risky because docmgr skills serve a different purpose (workflow discipline) and are used in docmgr CLI UX.
- **Embed skill directives in doc frontmatter**: fragile with docmgr frontmatter rewrite; leads to hidden, mixed concerns.
- **Manual skill packaging only**: low tooling value, high ongoing cost.

## Implementation Plan

1. Define the `skill.yaml` schema (skill metadata, sources, output layout).
2. Implement a resolver for source types:
   - `file`: read file contents, optionally strip frontmatter.
   - `docmgr-query`: use docmgr index to select docs.
   - `binary-help`: run `$binary help <topic>` and capture stdout.
3. Create a generator that writes:
   - `SKILL.md` with a concise index and usage guidance.
   - `references/` files per source.
4. Validate with the existing skill validator and package with `package_skill.py`.
5. Add a small CLI wrapper (for example `docmgr skill export --plan skills/<name>/skill.yaml`).

## Open Questions

- Should the exporter call `glaze help <topic>` directly, or should it parse `glaze help --list` to discover topics?
- What should be the default format for captured help output (raw markdown vs fenced code block)?
- Do we want to store captured help output alongside the skill plan for caching and reproducibility?
- How should the exporter handle help output localization or versioning if the binary changes?
