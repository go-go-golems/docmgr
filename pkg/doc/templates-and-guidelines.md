---
Title: Templates and Guidelines
Slug: templates-and-guidelines
Short: How _templates/ and _guidelines/ work, how to customize them, and how to use them effectively.
Topics:
- docmgr
- templates
- guidelines
- writing
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why Templates and Guidelines Exist

Consistent documentation makes reviews faster and decisions easier to understand months later. `docmgr` provides two complementary tools:

- Templates: Ready-made starting points for documents
- Guidelines: Human-readable guardrails that explain what “good” looks like

Use templates to bootstrap content quickly. Use guidelines to shape what you write and to align contributors around shared expectations.

## Where They Live

When you run `docmgr init`, `docmgr` scaffolds two root-level folders:

- `ttmp/_templates/` — One Markdown file per doc type (for example, `design-doc.md`, `reference.md`, `playbook.md`)
- `ttmp/_guidelines/` — One Markdown file per doc type with best practices and section checklists

Common doc types include: `index`, `design-doc`, `reference`, `playbook`, `til`, `analysis`, `misc`, and `code-review`.

These folders are part of your repository so your team can customize house style and keep it versioned.

## How They’re Used in the CLI

- `docmgr guidelines --doc-type <type> --output markdown`
  - Shows the guideline text for a given doc type
  - If `ttmp/_guidelines/<type>.md` exists, it is used; otherwise the embedded default is shown

- `docmgr add --ticket <ticket> --doc-type <type> --title <title>`
  - Creates a new file with frontmatter
  - If `ttmp/_templates/<type>.md` exists, the body is rendered from the template with variable substitution (for example, `{{TITLE}}`, `{{TICKET}}`, `{{TOPICS}}`, `{{OWNERS}}`)

Tip: Until auto-templating ships, open the template side-by-side and copy the relevant sections as you draft.

## Customizing Templates and Guidelines

1) Edit files under `ttmp/_templates/` and `ttmp/_guidelines/` directly
2) Keep sections short and focused; follow a consistent heading structure
3) Prefer actionable checklists and examples to long prose
4) Capture team-specific expectations (naming, diagrams, API table formats)

Example `ttmp/_templates/design-doc.md`:

```markdown
# {{ .Title }}

## Executive Summary
One paragraph that explains the problem and the proposed approach.

## Problem Statement
What user or system pain are we addressing? Include constraints and goals.

## Proposed Solution
Key components, trade-offs, and rationale. Include diagrams if helpful.

## Alternatives Considered
Briefly list the options you evaluated and why they were not chosen.

## Implementation Plan
Phases, owners, and risks. Link to issues and PRs as they are created.
```

Example `ttmp/_guidelines/design-doc.md` (excerpt):

```markdown
# Guidelines: Design Documents

## Purpose
Explain the “why” behind decisions; enable future readers to re-derive context.

## Required Elements
- Executive summary (3–7 sentences)
- Key decisions with rationale
- Alternatives considered
- Implementation plan with milestones

## Review Checklist
- Scope fits a single decision or small set of related decisions
- Trade-offs are explicit
- Owners and impacted teams are listed
```

## Versioning and Process

- Store templates and guidelines in the repo (under `ttmp/`) and evolve them via PRs
- Treat large template changes as design changes: explain the rationale in the PR
- Link your guidelines in onboarding docs and reference them during reviews

## Frequently Asked Questions

Q: Do templates enforce structure automatically?
A: Yes, when a matching template exists under `ttmp/_templates/`, `docmgr add` will render the body with variable substitution.

Q: Any guidance for RelatedFiles?
A: Prefer adding a short rationale note per file when possible. Use `docmgr relate --file-note "path:why it matters"` to capture the context.

Q: How should we maintain changelogs?
A: Capture decisions and progress in `changelog.md`. Use `docmgr changelog update` to append dated entries and optionally include related files with notes.

Q: Can guidelines differ between teams?
A: Yes. Start with shared defaults and layer team-specific files in `ttmp/_guidelines/`.

Q: How do I preview guidelines?
A: Run `docmgr guidelines --doc-type <type> --output markdown`.


