---
Title: Diagnostics integration diary
Ticket: DOCMGR-ERROR-TAXONOMY
Status: active
Topics:
    - errors
    - ux
    - yaml
DocType: working-note
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/documents/frontmatter.go
      Note: Taxonomy wrapping
    - Path: pkg/commands/doctor.go
      Note: Refactored doctor to use adapter
    - Path: pkg/commands/template_validate.go
      Note: Taxonomy wrapping
    - Path: pkg/diagnostics/docmgr/adapter.go
      Note: Adapter created
    - Path: pkg/diagnostics/docmgrctx/frontmatter.go
      Note: Frontmatter taxonomy
    - Path: pkg/diagnostics/docmgrctx/templates.go
      Note: Template taxonomy
    - Path: pkg/diagnostics/docmgrrules/frontmatter_rules.go
      Note: Frontmatter rules
    - Path: test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh
      Note: Scenario script
ExternalSources: []
Summary: Daily log for diagnostics taxonomy integration.
LastUpdated: 2025-12-01T13:06:20-05:00
---



# Diagnostics integration diary

## Summary
- Track progress and decisions while integrating shared diagnostics taxonomy/rules across docmgr.

## Notes
- 2025-12-01: Added shared adapter task and diary; planning to move diagnostics rendering helpers out of doctor into pkg/diagnostics/docmgr for reuse across verbs.
- 2025-12-01: Added frontmatter and template taxonomies plus frontmatter rules; wrapped frontmatter parsing and template validate errors into taxonomy; doctor now uses shared adapter; ran diagnostics smoke script after refactor.
- 2025-12-01: Added listing and workspace taxonomies + rules; list_docs now emits taxonomy on parse skip; added workspace/staleness contexts; wired template/frontmatter parsing to taxonomy; tests and smoke still passing.

## Decisions
- Use shared adapter package to avoid command-specific helpers and to enable other verbs to render diagnostics.

## Next Steps
- Move rendering helpers to pkg/diagnostics/docmgr.
- Wire doctor to use adapter; prepare to wire other verbs after adapter lands.
