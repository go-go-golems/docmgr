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
      Note: |-
        Refactored doctor to use adapter
        Workspace taxonomy wiring
    - Path: pkg/commands/list_docs.go
      Note: Listing taxonomy wiring
    - Path: pkg/commands/template_validate.go
      Note: Taxonomy wrapping
    - Path: pkg/diagnostics/docmgr/adapter.go
      Note: Adapter created
    - Path: pkg/diagnostics/docmgrctx/constructors.go
      Note: Constructors helper
    - Path: pkg/diagnostics/docmgrctx/frontmatter.go
      Note: Frontmatter taxonomy
    - Path: pkg/diagnostics/docmgrctx/listing.go
      Note: Listing taxonomy
    - Path: pkg/diagnostics/docmgrctx/templates.go
      Note: Template taxonomy
    - Path: pkg/diagnostics/docmgrctx/workspace.go
      Note: Workspace/staleness taxonomy
    - Path: pkg/diagnostics/docmgrrules/frontmatter_rules.go
      Note: Frontmatter rules
    - Path: pkg/diagnostics/docmgrrules/listing_rule.go
      Note: Listing rule
    - Path: pkg/diagnostics/docmgrrules/workspace_rule.go
      Note: Workspace rule
    - Path: test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh
      Note: |-
        Scenario script
        Expanded diagnostics smoke
        Expanded coverage
ExternalSources: []
Summary: Daily log for diagnostics taxonomy integration.
LastUpdated: 2025-12-01T17:40:00-05:00
---







# Diagnostics integration diary

## Summary
- Track progress and decisions while integrating shared diagnostics taxonomy/rules across docmgr.

## Notes
- 2025-12-01: Added shared adapter task and diary; planning to move diagnostics rendering helpers out of doctor into pkg/diagnostics/docmgr for reuse across verbs.
- 2025-12-01: Added frontmatter and template taxonomies plus frontmatter rules; wrapped frontmatter parsing and template validate errors into taxonomy; doctor now uses shared adapter; ran diagnostics smoke script after refactor.
- 2025-12-01: Added listing and workspace taxonomies + rules; list_docs now emits taxonomy on parse skip; added workspace/staleness contexts; wired template/frontmatter parsing to taxonomy; tests and smoke still passing.
- 2025-12-01: Wired missing_index and stale doctor findings to workspace taxonomy rendering; ensured meta_update/relate rely on taxonomy-wrapped frontmatter errors; refreshed playbook with newcomer guidance.
- 2025-12-01: Wired doctor to emit frontmatter schema taxonomies for missing required fields/status/topics and render parse errors from discovery/walkers; reran go test ./pkg/commands and ./pkg/diagnostics/... to confirm.
- 2025-12-01: Added diagnostics renderer collector + doctor --diagnostics-json flag, adapter unit test, and smoke script check for generated JSON output.
- 2025-12-01: Added frontmatter schema rule unit test, documented doctor --diagnostics-json in how-to-use, and marked remaining tasks complete.

## Decisions
- Use shared adapter package to avoid command-specific helpers and to enable other verbs to render diagnostics.

## Next Steps
- Move rendering helpers to pkg/diagnostics/docmgr.
- Wire doctor to use adapter; prepare to wire other verbs after adapter lands.
