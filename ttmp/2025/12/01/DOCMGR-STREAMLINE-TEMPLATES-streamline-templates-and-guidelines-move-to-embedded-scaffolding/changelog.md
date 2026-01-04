# Changelog

## 2025-12-01

- Initial workspace created


## 2025-12-01

Streamlined templates and guidelines system: moved templates/guidelines to embedded FS for scaffolding only, removed legacy string map fallback, moved verb templates to examples/, and created comprehensive documentation. Templates and guidelines now load from filesystem only at runtime; embedded versions are only used by docmgr init for initial scaffolding. If no template exists, documents are created with only frontmatter (empty body). Verb templates moved to examples/verb-templates/ as reference examples. Created verb-templates-and-schema.md and updated templates-and-guidelines.md with comprehensive guides following glazed style guide.


## 2025-12-01

Created comprehensive analysis report analyzing which templates add value vs slop. Identified useful templates (design-doc, reference, playbook, code-review), mixed templates needing streamlining (index, working-note, tutorial), and slop templates to remove (log, script, task-list). Report includes recommendations for template cleanup and next steps.

### Related Files

- docmgr/ttmp/2025/12/01/DOCMGR-STREAMLINE-TEMPLATES-streamline-templates-and-guidelines-move-to-embedded-scaffolding/analysis/01-template-analysis-useful-vs-slop.md — Comprehensive analysis of template usefulness


## 2025-12-01

Implemented embedded filesystem for templates and guidelines. Created internal/templates/embedded.go with LoadEmbeddedTemplate/LoadEmbeddedGuideline functions. Moved all templates from ttmp/_templates/ to internal/templates/embedded/_templates/ and guidelines to internal/templates/embedded/_guidelines/. Updated scaffold.go to use embedded FS for docmgr init scaffolding.

### Related Files

- docmgr/internal/templates/embedded.go — New embedded FS loading system
- docmgr/internal/templates/embedded/_guidelines/ — All guidelines moved here for scaffolding
- docmgr/internal/templates/embedded/_templates/ — All document templates moved here for scaffolding


## 2025-12-01

Updated runtime resolution to use filesystem-only loading. Modified LoadTemplate and LoadGuideline to check filesystem first, then return false (no legacy fallback). Updated add.go to create empty body when no template found. Updated guidelines_cmd.go to handle missing guidelines gracefully. Embedded templates/guidelines are now ONLY used for docmgr init scaffolding, never during document creation.

### Related Files

- docmgr/internal/templates/embedded.go — Updated LoadTemplate/LoadGuideline to filesystem-only
- docmgr/pkg/commands/add.go — Removed fallback content


## 2025-12-01

Removed legacy TemplateContent and GuidelineContent string maps from runtime resolution. These maps remain in code only for docmgr init scaffolding. Runtime document creation now uses filesystem templates only, or creates empty body if none exist. This simplifies the codebase and makes template behavior more predictable.

### Related Files

- docmgr/internal/templates/templates.go — TemplateContent kept only for scaffolding
- docmgr/pkg/commands/guidelines.go — GuidelineContent kept only for scaffolding


## 2025-12-01

Moved verb output templates from ttmp/templates/ to examples/verb-templates/ in codebase. Verb templates are now reference examples only, not automatically used. Users can copy them to ttmp/templates/ if they want to use them. This clarifies that verb templates are optional and separates them from document templates.

### Related Files

- docmgr/examples/verb-templates/ — Verb templates moved here as reference examples


## 2025-12-01

Created comprehensive documentation following glazed style guide. Created verb-templates-and-schema.md with complete guide to verb templates, template schema introspection, and best practices. Completely rewrote templates-and-guidelines.md with topic-focused introductions, detailed examples, and comprehensive FAQs. Both documents now follow glazed documentation standards.

### Related Files

- docmgr/pkg/doc/templates-and-guidelines.md — Complete rewrite following glazed style guide
- docmgr/pkg/doc/verb-templates-and-schema.md — New comprehensive guide for verb templates


## 2026-01-03

Auto-close: inactive >14d (last_updated 2025-12-01 20:01; tasks_open 0)

