---
Title: Exporting Glazed and docmgr docs into skills analysis
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
      Note: Skill format and packaging constraints
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/package_skill.py
      Note: Packaging flow to .skill
    - Path: ../../../../../../../../../../.codex/skills/.system/skill-creator/scripts/quick_validate.py
      Note: Frontmatter validation rules
    - Path: ../../../../../../../glazed/pkg/doc/topics/01-help-system.md
      Note: Glazed help doc metadata model
    - Path: ../../../../../../../glazed/pkg/help/help.go
      Note: HelpSystem markdown loader
    - Path: internal/documents/frontmatter.go
      Note: Frontmatter parsing behavior
    - Path: internal/workspace/sqlite_schema.go
      Note: Workspace index schema for selection
    - Path: pkg/doc/doc.go
      Note: Embedded doc loading for docmgr help docs
    - Path: pkg/models/document.go
      Note: docmgr document frontmatter schema
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-13T09:45:42-05:00
WhatFor: ""
WhenToUse: ""
---


# Analysis

## Goal and scope

This analysis examines ways to export Glazed documentation and docmgr documents into Agent Skills packages. It covers single-document conversion, multi-document bundling, and configuration-driven packaging via a skill.yaml DSL. It also considers overlaps between Glazed and docmgr, and the option to embed a skill subsection in frontmatter to declare export intent.

The scope is packaging and metadata mapping. It does not prescribe a final implementation, but it does identify constraints, candidate data formats, and risks that affect feasibility.

## Problem statement

Glazed and docmgr both curate Markdown documentation with structured frontmatter, but they do not ship as Agent Skills artifacts by default. Users may want to ship a skill that contains a curated subset of docs, packaged as SKILL.md plus references, and optionally scripts or assets. The problem is to define a deterministic, reproducible mapping from these doc ecosystems into skills, while remaining compatible with skill validators and consumer agents.

## Inputs and constraints

### Inputs from Glazed docs

- Glazed help docs are Markdown files with YAML frontmatter (Title, Slug, Short, Topics, Commands, Flags, SectionType) and content.
- Glazed docs are loaded via `help.LoadSectionFromMarkdown`, which parses frontmatter and treats the remainder as content.
- The docs are typically embedded and used via a HelpSystem with a SQLite store.

### Inputs from docmgr docs

- docmgr ticket docs are Markdown files with YAML frontmatter mapping to `models.Document` (Title, Ticket, DocType, Topics, Summary, RelatedFiles, etc).
- docmgr help docs are also Markdown files loaded into the same help system model as Glazed.
- docmgr workspace indexing and metadata normalization live in SQLite, which could be a query source for selection.

### Constraints from Agent Skills packages

- A skill is a directory containing at least `SKILL.md` (with YAML frontmatter) and optional `scripts/`, `references/`, and `assets/`.
- `SKILL.md` frontmatter requires `name` and `description`, and name must match directory and naming constraints.
- Skill instructions should be concise; long content belongs in references.
- Packaging should be reproducible: deterministic file names, stable ordering, and predictable transforms.

## Export surfaces and mapping opportunities

### Glazed help docs -> skill

- Each help doc is already structured for help UI consumption; the primary mapping is to place docs in `references/` and summarize in `SKILL.md`.
- The `Slug` field is a good candidate for filename and link targets.
- The `Short` field can seed skill section summaries or bullet lists.

### docmgr ticket docs -> skill

- docmgr docs include structured metadata that can drive selection (Ticket, DocType, Topics, Status).
- The `Summary`, `WhatFor`, and `WhenToUse` fields can seed skill descriptions.
- `RelatedFiles` can be translated into a skill section listing key code references.

### Shared help system model

- Both Glazed and docmgr help docs are loadable into the same HelpSystem model, which means a single exporter can operate on that model and ignore where docs originated.
- If the exporter operates on a HelpSystem store, it can unify Glazed and docmgr help docs in one pipeline.

## Strategy options

### Option 1: Single document -> skill

Convert one Markdown doc into a minimal skill package. The conversion is straightforward and low-risk, but yields many small skills.

Potential mapping:

- `SKILL.md` frontmatter
  - `name`: derived from `Slug` or from a sanitized Title.
  - `description`: derived from `Short`, `Summary`, or the first paragraph.
- `SKILL.md` body
  - Short overview and usage bullets.
  - Link to `references/<slug>.md`.
- `references/<slug>.md`
  - Original document content, optionally with frontmatter stripped.

Pros:

- Simple and deterministic.
- Aligns with skill discoverability (one doc per skill).

Cons:

- Generates many small skills, which can bloat skill lists.
- Cross-doc context is lost.

### Option 2: Bundle documents -> skill

Package a curated set of docs into a single skill. This yields fewer skills and supports thematic grouping.

Potential mapping:

- `SKILL.md` body includes an index of included docs with short summaries.
- `references/` contains each doc, with filenames based on Slug or a normalized Title.
- Optional `assets/` for shared diagrams or data files.

Pros:

- Better thematic grouping and fewer skills.
- Good for domain-specific playbooks or doc sets.

Cons:

- Requires selection logic and ordering rules.
- A single skill might become too large if not bounded.

### Option 3: skill.yaml DSL

Define a separate declarative config (skill.yaml) that describes how to build a skill from docs. The exporter reads this config and performs selection, transforms, and packaging.

Example DSL sketch:

```yaml
skill:
  name: docmgr-cli
  description: Reference docs and workflows for docmgr CLI usage.
  license: Proprietary
  compatibility: Requires docmgr source tree and local filesystem access.

sources:
  - type: docmgr-ticket
    ticket: DOCMGR-UX-001
    docTypes: [reference, playbook]
    include:
      - "reference/*.md"
    transform:
      strip-frontmatter: true
      add-origin-footer: true

  - type: glazed-help
    topics: [documentation, help]
    sectionTypes: [GeneralTopic, Tutorial]

output:
  references-dir: references
  include-index: true
  order-by: [topic, title]
```

Pros:

- Explicit and reproducible.
- Can be version-controlled and reviewed.
- Works for both Glazed and docmgr if the exporter supports multiple source types.

Cons:

- Adds another config artifact to maintain.
- Needs clear validation and error reporting.

### Option 4: Frontmatter skill subsection

Embed skill export directives in the document frontmatter itself. This is appealing because it keeps export intent near the doc content, but it interacts with docmgr parsing and rewriting.

Example frontmatter block:

```yaml
Skill:
  export: true
  name: docmgr-diagnostics
  section: Diagnostics
  include-in:
    - docmgr-core
```

Pros:

- Minimal extra files.
- Authors can opt-in directly in a document.

Cons:

- docmgr writes frontmatter from `models.Document`, which may drop unknown fields.
- Mixing export config into docs may lead to accidental leakage in shared docs.

## DSL or frontmatter overlap and unification

A hybrid approach is plausible:

- A top-level skill.yaml declares the skill package and selection rules.
- Individual docs can include a `Skill` block to override name, label, or exclude.
- If docmgr writes a doc and strips unknown fields, keep the `Skill` block in a separate sidecar (e.g., `doc.md.skill.yaml`) to avoid loss.

This balances central control (skill.yaml) with author-local hints (frontmatter or sidecar).

## Proposed data formats

### skill.yaml (export plan)

```yaml
skill:
  name: glazed-help-docs
  description: Help topics and tutorials for Glazed, packaged as an Agent Skill.

sources:
  - type: glazed-help
    topics: [help, documentation]
    sectionTypes: [GeneralTopic, Tutorial]

  - type: docmgr-help
    topics: [docmgr, documentation]

options:
  strip-frontmatter: true
  normalize-headings: true
  output-references: references
```

### Frontmatter skill subsection (doc opt-in)

```yaml
Skill:
  export: true
  skill-name: glazed-help-docs
  label: "Help System"
  order: 10
```

## Risks and open questions

- **Name matching**: The Agent Skills spec requires `name` to match directory. If names are derived from doc Slug, ensure the folder name matches or add a normalization step.
- **Frontmatter loss**: docmgr rewrites frontmatter using `models.Document`; unknown fields may be dropped. A sidecar file avoids this.
- **Doc size**: Large docs should land in `references/` to avoid huge SKILL.md bodies.
- **Doc provenance**: If doc content is merged, include source headers or a footer indicating origin.
- **Access control**: Some docs should not be exported to skills at all. A rule system or explicit allowlist is necessary.

## Potential overlap with docmgr

The docmgr workspace index already contains normalized metadata and doc body content. A skill exporter could:

- Query docmgr's SQLite index (topic filters, doc types, status) to select docs for export.
- Use docmgr's path normalization to map related files into skill references or an index section.
- Reuse docmgr's existing readme export patterns for a predictable ordering of docs.

This suggests a tight integration path: use docmgr as the selection engine and output a skill package using the skill-creator packaging rules.
