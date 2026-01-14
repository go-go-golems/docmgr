---
Title: 'Brainstorm: packaging Glazed + docmgr docs as skills'
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
      Note: Skill structure guidance
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/package_skill.py
      Note: Packaging implementation
    - Path: ttmp/2026/01/13/MO-001-EXPORT-SKILLS--export-skills-analysis/analysis/02-exporting-glazed-and-docmgr-docs-into-skills-analysis.md
      Note: Analysis basis for brainstorm
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-13T09:46:45-05:00
WhatFor: ""
WhenToUse: ""
---


# Brainstorm: packaging Glazed + docmgr docs as skills

## Executive Summary

This brainstorm outlines design directions for exporting Glazed and docmgr documentation into Agent Skills packages. The core idea is to transform curated Markdown docs into skill artifacts (SKILL.md + references/), while preserving metadata, maintaining discoverability, and respecting the skill format constraints. We explore three primary approaches: single-doc skills, multi-doc bundle skills, and a declarative skill.yaml DSL, along with a hybrid plan that optionally embeds skill directives in doc frontmatter or sidecar files.

## Problem Statement

There is no built-in pipeline that converts Glazed or docmgr docs into Agent Skills. We need a design that:

- Selects the right docs reliably (by topic, type, or ticket).
- Produces spec-compliant skills with deterministic naming and packaging.
- Minimizes extra maintenance burden for doc authors.
- Scales from single documents to curated sets of docs.

## Brainstorm Inventory

### Idea A: Single document -> skill

Each Markdown doc becomes its own skill, with the full doc placed in `references/` and a short summary in `SKILL.md`.

- Strength: Simple, deterministic, easy to validate.
- Weakness: Skill list may become too large or noisy.

### Idea B: Curated bundle -> skill

A curated set of docs becomes a single skill. `SKILL.md` is the index, and all docs live in `references/`.

- Strength: Good thematic grouping, fewer skills.
- Weakness: Requires selection logic and ordering rules.

### Idea C: Declarative skill.yaml DSL

A configuration file describes the skill metadata, doc selection, transforms, and output layout. The exporter uses this DSL to generate the skill package.

- Strength: Reproducible and version-controlled.
- Weakness: Another artifact to maintain and validate.

### Idea D: Frontmatter "Skill" subsection

Documents include a `Skill:` block that declares how they should be exported (skill name, label, order).

- Strength: Keeps export intent close to content.
- Weakness: docmgr rewrites frontmatter and may drop unknown fields unless protected.

### Idea E: Sidecar skill metadata files

For each doc, optionally provide a `doc.md.skill.yaml` file describing export options (skill name, labels, inclusion).

- Strength: Avoids frontmatter loss; more explicit.
- Weakness: Extra files to manage; naming conventions required.

## Proposed Solution (initial direction)

Adopt a hybrid design:

- **Primary control**: a `skill.yaml` export plan that defines skill metadata and document selection.
- **Optional overrides**: doc-level hints via a `Skill:` frontmatter block or sidecar file.
- **Exporter pipeline**: a single tool that can read Glazed docs, docmgr docs, or docmgr's SQLite index, normalize content, and generate a skill directory that is packaged with the existing skill packager.

This keeps the core configuration centralized while allowing doc authors to opt in or refine naming and ordering.

## Design Decisions

1. **Use an explicit export plan (skill.yaml) as the canonical source of truth.**
   - Rationale: deterministic behavior and easier review in code reviews.

2. **Keep full doc content in `references/` by default.**
   - Rationale: SKILL.md should remain concise and focused; large content should not bloat activation context.

3. **Support both Glazed and docmgr inputs with a shared model.**
   - Rationale: they are already compatible via the help system model; a single exporter should reduce duplication.

4. **Permit but do not require frontmatter `Skill` directives.**
   - Rationale: avoids forcing doc authors to modify frontmatter while still enabling opt-in signals.

## Alternatives Considered

- **Pure frontmatter control only**: simple for authors but fragile when docmgr rewrites frontmatter and drops unknown fields.
- **Manual curation only**: reliable but not scalable and easy to drift from source docs.
- **Skill per ticket**: ties skills to ticket lifecycle; good for temporary knowledge but poor for long-term reuse.

## Implementation Plan

1. **Define the skill.yaml schema**
   - Required: skill name, description.
   - Optional: sources (docmgr/glazed), filters (topic, doc type), transforms, output layout.

2. **Implement a document collector**
   - Load Glazed or docmgr help docs via HelpSystem APIs.
   - Optionally load docmgr ticket docs from filesystem or docmgr SQLite index.

3. **Implement transformations**
   - Strip frontmatter.
   - Normalize headings and lists.
   - Add provenance footer or header.

4. **Generate skill output**
   - Write `SKILL.md` with a short overview and index list.
   - Write each doc into `references/` with stable filenames.

5. **Validate and package**
   - Run the local validator.
   - Optionally run spec-level validation.
   - Package with the existing skill packager.

## Open Questions

- Should the exporter always keep the original doc filename, or derive from Slug/Title?
- How should the exporter handle multiple docs with identical titles or slugs?
- Should docmgr tickets include an explicit opt-in to avoid exporting sensitive notes?
- Can the docmgr SQLite index be treated as the canonical selection layer, or should filesystem selection be preferred?
