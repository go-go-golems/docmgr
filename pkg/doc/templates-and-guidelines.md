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

# Templates and Guidelines

Consistent documentation structure makes reviews faster and decisions easier to understand months later. docmgr provides two complementary tools that work together: **document templates** scaffold the initial structure of document bodies, while **guidelines** provide writing guidance to help authors understand what "good" looks like for each document type. This separation allows teams to evolve their documentation standards independently—templates define the skeleton, guidelines define the quality standards.

## Core Concepts

### Document Templates

Document templates (`_templates/`) are markdown files that define the **body structure** of documents when they're created. They use simple variable substitution (like `{{TITLE}}`, `{{TICKET}}`) to scaffold sections, headings, and placeholder content. Templates are optional—if no template exists for a doc type, the document is created with only frontmatter and an empty body.

**Key characteristics:**
- **Scaffolding only**: Templates create initial structure, not final content
- **Variable substitution**: Use placeholders like `{{TITLE}}`, `{{TICKET}}`, `{{TOPICS}}`
- **Filesystem-based**: Stored in `ttmp/_templates/<doc-type>.md`
- **Team-customizable**: Edit templates to match your team's documentation style

### Guidelines

Guidelines (`_guidelines/`) are markdown files that explain **how to write** each document type effectively. They're displayed automatically after creating a document and can be viewed anytime with `docmgr doc guidelines`. Guidelines focus on quality standards, required elements, and best practices rather than structure.

**Key characteristics:**
- **Writing guidance**: Explain what makes a good document of this type
- **Quality standards**: Define required elements and review checklists
- **Best practices**: Share team-specific expectations and conventions
- **Non-enforcing**: Guidelines are suggestions, not requirements

## Where They Live

When you run `docmgr init`, docmgr scaffolds two directories in your docs root:

- `ttmp/_templates/` — One Markdown file per doc type (e.g., `design-doc.md`, `reference.md`, `playbook.md`)
- `ttmp/_guidelines/` — One Markdown file per doc type with best practices and section checklists

These directories are part of your repository, so your team can customize house style and keep it versioned alongside your documentation.

**Common doc types:**
- `index` — Ticket landing page
- `design-doc` — Architecture and design decisions
- `reference` — API contracts and quick reference
- `playbook` — Operational procedures and runbooks
- `code-review` — Code review summaries
- `analysis` — Research and analysis documents
- `til` — Today I Learned entries
- `working-note` — Free-form notes and meeting summaries

## How Document Templates Work

Document templates scaffold the body content of documents when you create them with `docmgr doc add`. The template system uses simple variable substitution—no complex logic, just placeholders that get replaced with actual values.

### Template Variables

Templates support these variables that are automatically replaced:

- `{{TITLE}}` — Document title (from `--title` flag)
- `{{TICKET}}` — Ticket identifier (from `--ticket` flag)
- `{{DATE}}` — Current timestamp in ISO format
- `{{TOPICS}}` — YAML-formatted topics array
- `{{OWNERS}}` — YAML-formatted owners array
- `{{SUMMARY}}` — Summary text (from `--summary` flag)
- `{{STATUS}}` — Document status (defaults to ticket status)

### Template Resolution

When you create a document:

1. **Check filesystem**: Look for `ttmp/_templates/<doc-type>.md`
2. **If found**: Extract body (skip frontmatter), substitute variables, use as document body
3. **If not found**: Create document with only frontmatter (empty body)

**Important:** Templates are loaded from the filesystem only. Embedded templates are only used for scaffolding via `docmgr init`, not during document creation.

### Example Template

Here's a typical `design-doc` template:

```markdown
# {{TITLE}}

## Executive Summary

<!-- Provide a high-level overview of the design proposal -->

## Problem Statement

<!-- Describe the problem this design addresses -->

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
```

When you run `docmgr doc add --doc-type design-doc --title "API Design"`, this template creates a document with all these sections pre-populated with your title.

## How Guidelines Work

Guidelines are displayed **after** creating a document to help authors understand how to write that type of document effectively. They're also available on-demand via `docmgr doc guidelines`.

### When Guidelines Are Shown

- **After document creation**: When you run `docmgr doc add`, guidelines for that doc type are automatically displayed
- **On demand**: Run `docmgr doc guidelines --doc-type <type>` to view guidelines anytime
- **Listing available types**: Run `docmgr doc guidelines --list` to see all doc types with guidelines

### Guideline Resolution

Guidelines are loaded from the filesystem only:

1. **Check filesystem**: Look for `ttmp/_guidelines/<doc-type>.md`
2. **If found**: Display the guideline content
3. **If not found**: Show a message that no guideline exists (command still succeeds)

**Important:** Embedded guidelines are only used for scaffolding via `docmgr init`, not during runtime.

### Example Guideline

Here's a typical `design-doc` guideline:

```markdown
# Guidelines: Design Documents

## Purpose
Design documents provide structured rationale and architecture notes. They document decisions, alternatives considered, and implementation plans.

## Required Elements
- **Executive Summary**: High-level overview (3–7 sentences)
- **Problem Statement**: Clear description of the problem
- **Proposed Solution**: Detailed solution description
- **Design Decisions**: Key decisions with rationale
- **Alternatives Considered**: Other approaches evaluated
- **Implementation Plan**: Steps to implement

## Best Practices
- Start with executive summary for quick scanning
- Document the "why" behind decisions, not just the "what"
- Include diagrams or code examples where helpful
- Keep open questions visible until resolved
- Link to related RFCs, tickets, or references
- Use clear section hierarchy for easy navigation
```

## Customizing Templates and Guidelines

Templates and guidelines are meant to be customized to match your team's needs. Edit them directly in your repository to evolve your documentation standards.

### Editing Templates

1. **Open the template file**: `ttmp/_templates/<doc-type>.md`
2. **Modify structure**: Add, remove, or reorder sections
3. **Update placeholders**: Use template variables where appropriate
4. **Commit changes**: Version control tracks your template evolution

**Example customization:**

```markdown
# {{TITLE}}

## Context
<!-- Add your team-specific context section -->

## Decision
<!-- Your team's decision format -->

## Impact
<!-- How this affects other systems -->
```

### Editing Guidelines

1. **Open the guideline file**: `ttmp/_guidelines/<doc-type>.md`
2. **Update standards**: Modify required elements and best practices
3. **Add team-specific guidance**: Include your team's conventions
4. **Commit changes**: Keep guidelines versioned with your docs

**Example customization:**

```markdown
# Guidelines: Design Documents

## Our Team's Standards
- Always include a diagram (use Mermaid syntax)
- Link to at least 3 related tickets
- Get 2+ approvals before marking complete

## Review Checklist
- [ ] Diagram is clear and accurate
- [ ] All alternatives have pros/cons listed
- [ ] Implementation plan has concrete milestones
```

### Best Practices for Customization

- **Start simple**: Begin with default templates/guidelines, then evolve based on actual needs
- **Keep focused**: Templates should scaffold structure, not prescribe every detail
- **Document changes**: When making significant changes, explain the rationale in a PR
- **Team consensus**: Templates and guidelines should reflect team agreement, not individual preferences
- **Regular review**: Periodically review templates/guidelines to ensure they're still useful

## Template Examples

### Minimal Template (Reference Doc)

```markdown
# {{TITLE}}

## Goal

<!-- What is the purpose of this reference document? -->

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->
```

### Structured Template (Playbook)

```markdown
# {{TITLE}}

## Purpose

<!-- What does this playbook accomplish? -->

## Environment Assumptions

<!-- What environment or setup is required? -->

## Commands

<!-- List of commands to execute -->

```bash
# Command sequence
```

## Exit Criteria

<!-- What indicates success or completion? -->

## Notes

<!-- Additional context or warnings -->
```

### Flexible Template (Working Note)

```markdown
# {{TITLE}}

## Summary

<!-- Brief summary for LLM ingestion -->

## Notes

<!-- Free-form notes, meeting summaries, or research findings -->

## Decisions

<!-- Any decisions made during this working session -->

## Next Steps

<!-- Actions or follow-ups -->
```

## Guideline Examples

### Design Document Guidelines

```markdown
# Guidelines: Design Documents

## Purpose
Design documents provide structured rationale and architecture notes. They document decisions, alternatives considered, and implementation plans.

## Required Elements
- **Executive Summary**: High-level overview (3–7 sentences)
- **Problem Statement**: Clear description of the problem
- **Proposed Solution**: Detailed solution description
- **Design Decisions**: Key decisions with rationale
- **Alternatives Considered**: Other approaches evaluated
- **Implementation Plan**: Steps to implement

## Best Practices
- Start with executive summary for quick scanning
- Document the "why" behind decisions, not just the "what"
- Include diagrams or code examples where helpful
- Keep open questions visible until resolved
- Link to related RFCs, tickets, or references
- Use clear section hierarchy for easy navigation
```

### Reference Document Guidelines

```markdown
# Guidelines: Reference Documents

## Purpose
Reference documents provide copy/paste-ready context, API contracts, prompt packs, or quick-look tables. They're designed for reuse in LLM prompts or as quick reference during development.

## Required Elements
- **Goal**: What this reference accomplishes
- **Context**: Background needed to use the reference
- **Quick Reference**: The actual reference content (code, tables, prompts)
- **Usage Examples**: How to use this reference

## Best Practices
- Focus on copy/paste-ready content
- Keep context minimal but sufficient
- Format for easy scanning (tables, code blocks)
- Include practical examples
- Update when APIs or processes change
- Link to related references or design docs
```

## Versioning and Process

Templates and guidelines are part of your repository, so they evolve alongside your documentation standards.

### Version Control

- **Store in repo**: Keep `_templates/` and `_guidelines/` in version control
- **Evolve via PRs**: Treat template/guideline changes as code changes
- **Document rationale**: Explain why templates/guidelines changed in PR descriptions
- **Team review**: Get team consensus before making significant changes

### Change Management

- **Start conservative**: Begin with minimal templates, add structure as needed
- **Incremental changes**: Evolve templates based on actual usage patterns
- **Remove unused sections**: If a template section is always deleted, remove it
- **Link in onboarding**: Reference templates/guidelines in team onboarding docs

## Frequently Asked Questions

**Q: Do templates enforce structure automatically?**
A: Yes, when a matching template exists under `ttmp/_templates/`, `docmgr doc add` will render the body with variable substitution. If no template exists, the document is created with only frontmatter.

**Q: What happens if I don't have a template for a doc type?**
A: The document is created with only frontmatter (empty body). You can add content manually or create a template for that doc type.

**Q: Can I use templates without frontmatter?**
A: Templates can include frontmatter (which is ignored) or just body content. The frontmatter in templates is only for documentation—the actual frontmatter is generated by docmgr.

**Q: How do I preview what a template will create?**
A: Create a test document: `docmgr doc add --ticket TEST --doc-type <type> --title "Test"` and check the generated file.

**Q: Any guidance for RelatedFiles?**
A: Prefer adding a short rationale note per file when possible. Use `docmgr doc relate --file-note "path:why it matters"` to capture the context.

**Q: How should we maintain changelogs?**
A: Capture decisions and progress in `changelog.md`. Use `docmgr changelog update` to append dated entries and optionally include related files with notes.

**Q: Can guidelines differ between teams?**
A: Yes. Start with shared defaults and layer team-specific files in `ttmp/_guidelines/`. Each team can customize their guidelines independently.

**Q: How do I preview guidelines?**
A: Run `docmgr doc guidelines --doc-type <type> --output markdown`.

**Q: Where do embedded templates come from?**
A: When you run `docmgr init`, templates and guidelines are scaffolded from embedded defaults. After that, only filesystem templates/guidelines are used. Embedded templates are not used during document creation.

**Q: Can I have different templates for the same doc type?**
A: No, each doc type has one template file. If you need variation, create a new doc type or customize the template to be flexible.

**Q: Do templates support conditional logic?**
A: No, templates use simple variable substitution only. For complex logic, create the document structure manually or use multiple doc types.

## Related Documentation

- **Verb Output Templates**: See `docmgr help verb-templates-and-schema` for information about postfix templates for command output
- **Vocabulary Management**: See `docmgr help how-to-setup` for managing doc types and topics
- **Document Creation**: See `docmgr help how-to-use` for complete workflow documentation
